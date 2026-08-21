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
	GetSelfHistory(orgID, userID uuid.UUID, query AttendanceHistoryQuery) (*AttendanceHistoryResponse, error)
	GetTeamHistory(orgID, managerID uuid.UUID, query AttendanceHistoryQuery) (*AttendanceHistoryResponse, error)
	GetOrgHistory(orgID uuid.UUID, query AttendanceHistoryQuery) (*AttendanceHistoryResponse, error)
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

	now := time.Now().UTC()
	attLog := AttendanceLog{
		UserID:     userID,
		ClockIn:    now,
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

		var existing AttendanceLog
		if err := s.db.Where("record_uuid = ?", parsedRecordUUID).First(&existing).Error; err == nil {
			results = append(results, SyncResult{
				RecordUUID: rec.RecordUUID,
				Status:     "ALREADY_SYNCED",
				Message:    "Record UUID already exists in database",
			})
			continue
		}

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

func (s *service) buildQuery(base *gorm.DB, query AttendanceHistoryQuery) (*gorm.DB, int, int) {
	page := query.Page
	if page <= 0 {
		page = 1
	}
	limit := query.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	if query.StartDate != nil && *query.StartDate != "" {
		base = base.Where("clock_in >= ?", *query.StartDate)
	}
	if query.EndDate != nil && *query.EndDate != "" {
		base = base.Where("clock_in <= ?", *query.EndDate)
	}

	return base, page, limit
}

func (s *service) fetchPaginated(dbQuery *gorm.DB, page, limit int) (*AttendanceHistoryResponse, error) {
	var total int64
	if err := dbQuery.Model(&AttendanceLog{}).Count(&total).Error; err != nil {
		return nil, err
	}

	var logs []AttendanceLog
	offset := (page - 1) * limit
	if err := dbQuery.Order("clock_in desc").Limit(limit).Offset(offset).Find(&logs).Error; err != nil {
		return nil, err
	}

	var responses []AttendanceResponse
	for _, l := range logs {
		responses = append(responses, AttendanceResponse{
			ID:         l.ID.ID,
			UserID:     l.UserID,
			ClockIn:    l.ClockIn,
			ClockOut:   l.ClockOut,
			TotalHours: l.TotalHours,
			SyncStatus: l.SyncStatus,
			DeviceHash: l.DeviceHash,
			RecordUUID: l.RecordUUID,
		})
	}

	return &AttendanceHistoryResponse{
		Total: int(total),
		Page:  page,
		Limit: limit,
		Data:  responses,
	}, nil
}

// GetSelfHistory fetches self-scoped attendance logs (Task 15)
func (s *service) GetSelfHistory(orgID, userID uuid.UUID, query AttendanceHistoryQuery) (*AttendanceHistoryResponse, error) {
	base := s.db.Where("org_id = ? AND user_id = ?", orgID, userID)
	dbQuery, page, limit := s.buildQuery(base, query)
	return s.fetchPaginated(dbQuery, page, limit)
}

// GetTeamHistory fetches team-scoped attendance logs (Task 15)
func (s *service) GetTeamHistory(orgID, managerID uuid.UUID, query AttendanceHistoryQuery) (*AttendanceHistoryResponse, error) {
	base := s.db.Where("org_id = ?", orgID)
	if query.TeamID != nil && *query.TeamID != uuid.Nil {
		base = base.Where("team_id = ?", *query.TeamID)
	}
	if query.UserID != nil && *query.UserID != uuid.Nil {
		base = base.Where("user_id = ?", *query.UserID)
	}
	dbQuery, page, limit := s.buildQuery(base, query)
	return s.fetchPaginated(dbQuery, page, limit)
}

// GetOrgHistory fetches organization-wide attendance logs (Task 15)
func (s *service) GetOrgHistory(orgID uuid.UUID, query AttendanceHistoryQuery) (*AttendanceHistoryResponse, error) {
	base := s.db.Where("org_id = ?", orgID)
	if query.UserID != nil && *query.UserID != uuid.Nil {
		base = base.Where("user_id = ?", *query.UserID)
	}
	dbQuery, page, limit := s.buildQuery(base, query)
	return s.fetchPaginated(dbQuery, page, limit)
}
