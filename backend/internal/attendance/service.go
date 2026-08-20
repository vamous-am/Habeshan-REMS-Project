package attendance

import (
	"bytes"
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrActiveClockInExists = errors.New("active clock-in session already exists")
	ErrNoActiveClockIn     = errors.New("no active clock-in session found")
)

type UserExportFilter struct {
	UserID    *uuid.UUID
	StartDate *time.Time
	EndDate   *time.Time
}

type Service interface {
	ClockIn(orgID, userID uuid.UUID, req ClockInRequest) (*AttendanceLog, error)
	ClockOut(orgID, userID uuid.UUID, req ClockOutRequest) (*AttendanceLog, error)
	SyncBatch(orgID uuid.UUID, reqs []SyncRecordRequest) (BatchSyncResponse, error)
	ExportLogs(orgID uuid.UUID, filter UserExportFilter) ([]byte, error)
}

type service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) Service {
	return &service{db: db}
}

func (s *service) ClockIn(orgID, userID uuid.UUID, req ClockInRequest) (*AttendanceLog, error) {
	var active AttendanceLog
	err := s.db.Where("org_id = ? AND user_id = ? AND clock_out IS NULL", orgID, userID).First(&active).Error
	if err == nil {
		return nil, ErrActiveClockInExists
	}

	attLog := AttendanceLog{
		UserID:     userID,
		ClockIn:    time.Now().UTC(),
		SyncStatus: SyncStatusSyncedVerified,
		DeviceHash: req.DeviceHash,
		RecordUUID: req.RecordUUID,
	}
	attLog.OrgID = orgID

	if err := s.db.Create(&attLog).Error; err != nil {
		return nil, err
	}

	return &attLog, nil
}

func (s *service) ClockOut(orgID, userID uuid.UUID, req ClockOutRequest) (*AttendanceLog, error) {
	var active AttendanceLog
	err := s.db.Where("org_id = ? AND user_id = ? AND clock_out IS NULL", orgID, userID).First(&active).Error
	if err != nil {
		return nil, ErrNoActiveClockIn
	}

	now := time.Now().UTC()
	duration := now.Sub(active.ClockIn).Hours()

	active.ClockOut = &now
	active.TotalHours = &duration
	active.SyncStatus = SyncStatusSyncedVerified

	if err := s.db.Save(&active).Error; err != nil {
		return nil, err
	}

	return &active, nil
}

// ComputeRecordHash generates SHA-256 checksum matching both pipe '|' and colon ':' client formats
func ComputeRecordHash(rec SyncRecordRequest) string {
	rawPipe := fmt.Sprintf("%s|%s|%s|%s", rec.RecordUUID, rec.UserID, rec.ActionType, rec.Timestamp)
	hashPipe := sha256.Sum256([]byte(rawPipe))
	return hex.EncodeToString(hashPipe[:])
}

func (s *service) SyncBatch(orgID uuid.UUID, records []SyncRecordRequest) (BatchSyncResponse, error) {
	var results []SyncResult

	for _, rec := range records {
		parsedRecordUUID, err := uuid.Parse(rec.RecordUUID)
		if err != nil {
			results = append(results, SyncResult{
				RecordUUID: rec.RecordUUID,
				Status:     string(SyncStatusRejectedTampered),
				Message:    "Invalid UUID format",
			})
			continue
		}

		parsedUserID, err := uuid.Parse(rec.UserID)
		if err != nil {
			results = append(results, SyncResult{
				RecordUUID: rec.RecordUUID,
				Status:     string(SyncStatusRejectedTampered),
				Message:    "Invalid User ID format",
			})
			continue
		}

		// Task 10: Check DB for existing record_uuid (Idempotent Deduplication)
		var existing AttendanceLog
		if err := s.db.Where("record_uuid = ?", parsedRecordUUID).First(&existing).Error; err == nil {
			results = append(results, SyncResult{
				RecordUUID: rec.RecordUUID,
				Status:     "ALREADY_SYNCED",
				Message:    "Record UUID already exists in database",
			})
			continue
		}

		// Task 11: Cryptographic Hash Tamper Verification
		expectedHashPipe := ComputeRecordHash(rec)
		rawColon := fmt.Sprintf("%s:%s:%s:%s", rec.RecordUUID, rec.UserID, rec.ActionType, rec.Timestamp)
		hashColon := sha256.Sum256([]byte(rawColon))
		expectedHashColon := hex.EncodeToString(hashColon[:])

		if rec.DeviceHash != expectedHashPipe && rec.DeviceHash != expectedHashColon {
			tamperedLog := AttendanceLog{
				UserID:     parsedUserID,
				ClockIn:    time.Now().UTC(),
				SyncStatus: SyncStatusRejectedTampered,
				DeviceHash: rec.DeviceHash,
				RecordUUID: parsedRecordUUID,
			}
			tamperedLog.OrgID = orgID
			s.db.Create(&tamperedLog)

			results = append(results, SyncResult{
				RecordUUID: rec.RecordUUID,
				Status:     string(SyncStatusRejectedTampered),
				Message:    "Device hash signature mismatch",
			})
			continue
		}

		ts, err := time.Parse(time.RFC3339, rec.Timestamp)
		if err != nil {
			ts = time.Now().UTC()
		}

		if rec.ActionType == "CLOCK_IN" {
			logEntry := AttendanceLog{
				UserID:     parsedUserID,
				ClockIn:    ts,
				SyncStatus: SyncStatusSyncedVerified,
				DeviceHash: rec.DeviceHash,
				RecordUUID: parsedRecordUUID,
			}
			logEntry.OrgID = orgID
			if err := s.db.Create(&logEntry).Error; err != nil {
				results = append(results, SyncResult{
					RecordUUID: rec.RecordUUID,
					Status:     "ERROR",
					Message:    err.Error(),
				})
				continue
			}
		} else if rec.ActionType == "CLOCK_OUT" {
			var active AttendanceLog
			if err := s.db.Where("org_id = ? AND user_id = ? AND clock_out IS NULL", orgID, parsedUserID).
				Order("clock_in desc").
				First(&active).Error; err == nil {

				duration := ts.Sub(active.ClockIn).Hours()
				active.ClockOut = &ts
				active.TotalHours = &duration
				active.SyncStatus = SyncStatusSyncedVerified
				s.db.Save(&active)
			} else {
				logEntry := AttendanceLog{
					UserID:     parsedUserID,
					ClockIn:    ts,
					ClockOut:   &ts,
					SyncStatus: SyncStatusSyncedVerified,
					DeviceHash: rec.DeviceHash,
					RecordUUID: parsedRecordUUID,
				}
				logEntry.OrgID = orgID
				s.db.Create(&logEntry)
			}
		}

		results = append(results, SyncResult{
			RecordUUID: rec.RecordUUID,
			Status:     string(SyncStatusSyncedVerified),
		})
	}

	return BatchSyncResponse{
		Processed: len(results),
		Results:   results,
	}, nil
}

// ExportLogs handles Task 13: Querying database and formatting CSV binary stream
func (s *service) ExportLogs(orgID uuid.UUID, filter UserExportFilter) ([]byte, error) {
	var logs []AttendanceLog
	query := s.db.Where("org_id = ?", orgID)

	if filter.UserID != nil {
		query = query.Where("user_id = ?", *filter.UserID)
	}
	if filter.StartDate != nil {
		query = query.Where("clock_in >= ?", *filter.StartDate)
	}
	if filter.EndDate != nil {
		query = query.Where("clock_in <= ?", *filter.EndDate)
	}

	if err := query.Order("clock_in DESC").Find(&logs).Error; err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	writer := csv.NewWriter(buf)

	// CSV Header
	header := []string{"Record UUID", "User ID", "Clock In", "Clock Out", "Total Hours", "Sync Status"}
	if err := writer.Write(header); err != nil {
		return nil, err
	}

	for _, l := range logs {
		clockOutStr := "N/A"
		if l.ClockOut != nil {
			clockOutStr = l.ClockOut.Format(time.RFC3339)
		}

		totalHrsStr := "0.00"
		if l.TotalHours != nil {
			totalHrsStr = fmt.Sprintf("%.2f", *l.TotalHours)
		}

		row := []string{
			l.RecordUUID.String(),
			l.UserID.String(),
			l.ClockIn.Format(time.RFC3339),
			clockOutStr,
			totalHrsStr,
			string(l.SyncStatus),
		}

		if err := writer.Write(row); err != nil {
			return nil, err
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}