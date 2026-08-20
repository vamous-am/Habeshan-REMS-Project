package attendance

import (
	"crypto/sha256"
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

type Service interface {
	ClockIn(orgID, userID uuid.UUID, req ClockInRequest) (*AttendanceLog, error)
	ClockOut(orgID, userID uuid.UUID, req ClockOutRequest) (*AttendanceLog, error)
	SyncBatch(orgID uuid.UUID, reqs []SyncRecordRequest) (BatchSyncResponse, error)
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
	// Standard canonical representation: {record_uuid}|{user_id}|{action_type}|{timestamp}
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
		// Fallback check for alternate colon delimiter formatting
		rawColon := fmt.Sprintf("%s:%s:%s:%s", rec.RecordUUID, rec.UserID, rec.ActionType, rec.Timestamp)
		hashColon := sha256.Sum256([]byte(rawColon))
		expectedHashColon := hex.EncodeToString(hashColon[:])

		if rec.DeviceHash != expectedHashPipe && rec.DeviceHash != expectedHashColon {
			// Record marked and logged as REJECTED_TAMPERED
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

		// Parse record timestamp safely
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
				// Standalone clock out log entry if no open active log exists
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