package attendance

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const MaxOnlineClockDrift = 5 * time.Minute

var (
	ErrActiveClockInExists = errors.New("active clock-in session already exists")
	ErrNoActiveClockIn     = errors.New("no active clock-in session found")
	ErrTamperedHash        = errors.New("device hash signature mismatch: record rejected as tampered")
	ErrInvalidTimestamp    = errors.New("invalid timestamp format: RFC3339 required")
	ErrTimestampDrift      = errors.New("timestamp outside allowed online window")
	ErrDuplicateRecordUUID = errors.New("duplicate record uuid")
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

func constantTimeCompare(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

// VerifyDeviceHash checks whether deviceHash matches either the canonical pipe format
// or legacy colon format for the given record fields using constant-time comparison.
// Note: This provides tamper/integrity verification across offline and online paths,
// but does not provide hardware/device authentication without a per-device secret.
func VerifyDeviceHash(recordUUID, userID, actionType, rawTimestamp, deviceHash string) bool {
	if recordUUID == "" || userID == "" || actionType == "" || rawTimestamp == "" || deviceHash == "" {
		return false
	}
	rawPipe := fmt.Sprintf("%s|%s|%s|%s", recordUUID, userID, actionType, rawTimestamp)
	hashPipe := sha256.Sum256([]byte(rawPipe))
	expectedPipe := hex.EncodeToString(hashPipe[:])

	rawColon := fmt.Sprintf("%s:%s:%s:%s", recordUUID, userID, actionType, rawTimestamp)
	hashColon := sha256.Sum256([]byte(rawColon))
	expectedColon := hex.EncodeToString(hashColon[:])

	return constantTimeCompare(deviceHash, expectedPipe) || constantTimeCompare(deviceHash, expectedColon)
}

// ParseTimestamp parses an RFC3339 or RFC3339Nano timestamp string into UTC.
func ParseTimestamp(tsStr string) (time.Time, error) {
	if ts, err := time.Parse(time.RFC3339Nano, tsStr); err == nil {
		return ts.UTC(), nil
	}
	ts, err := time.Parse(time.RFC3339, tsStr)
	if err != nil {
		return time.Time{}, err
	}
	return ts.UTC(), nil
}

// ValidateOnlineDrift returns true if clientTime is within maxDrift of serverNow.
func ValidateOnlineDrift(clientTime, serverNow time.Time, maxDrift time.Duration) bool {
	diff := serverNow.Sub(clientTime)
	if diff < 0 {
		diff = -diff
	}
	return diff <= maxDrift
}

func (s *service) ClockIn(orgID, userID uuid.UUID, req ClockInRequest) (*AttendanceLog, error) {
	now := time.Now().UTC()

	// Ensure no active verified clock-in session exists
	var active AttendanceLog
	err := s.db.Where("org_id = ? AND user_id = ? AND clock_out IS NULL AND sync_status = ?",
		orgID, userID, SyncStatusSyncedVerified).First(&active).Error
	if err == nil {
		return nil, ErrActiveClockInExists
	}

	// Check if record_uuid already exists to prevent duplicate key crashes
	var existing AttendanceLog
	if err := s.db.Where("record_uuid = ?", req.RecordUUID).First(&existing).Error; err == nil {
		return nil, ErrDuplicateRecordUUID
	}

	// Helper to persist rejected audit record without leaving an open active session
	persistRejectedAudit := func(eventTime time.Time) {
		tamperedLog := AttendanceLog{
			UserID:     userID,
			ClockIn:    eventTime,
			ClockOut:   &eventTime,
			SyncStatus: SyncStatusRejectedTampered,
			DeviceHash: req.DeviceHash,
			RecordUUID: req.RecordUUID,
		}
		tamperedLog.OrgID = orgID
		_ = s.db.Create(&tamperedLog)
	}

	// Parse timestamp (RFC3339 / RFC3339Nano)
	parsedTs, err := ParseTimestamp(req.Timestamp)
	if err != nil {
		persistRejectedAudit(now)
		return nil, ErrInvalidTimestamp
	}

	// Validate online drift window (max 5 minutes)
	if !ValidateOnlineDrift(parsedTs, now, MaxOnlineClockDrift) {
		persistRejectedAudit(now)
		return nil, ErrTimestampDrift
	}

	// Verify cryptographic device hash
	if !VerifyDeviceHash(req.RecordUUID.String(), userID.String(), "CLOCK_IN", req.Timestamp, req.DeviceHash) {
		persistRejectedAudit(parsedTs)
		return nil, ErrTamperedHash
	}

	attLog := AttendanceLog{
		UserID:     userID,
		ClockIn:    parsedTs,
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
	now := time.Now().UTC()

	// Locate active verified clock-in session
	var active AttendanceLog
	err := s.db.Where("org_id = ? AND user_id = ? AND clock_out IS NULL AND sync_status = ?",
		orgID, userID, SyncStatusSyncedVerified).First(&active).Error
	if err != nil {
		return nil, ErrNoActiveClockIn
	}

	// Helper to persist rejected audit record without overwriting active session or violating unique constraints
	persistRejectedAudit := func(eventTime time.Time) {
		var existing AttendanceLog
		if err := s.db.Where("record_uuid = ?", req.RecordUUID).First(&existing).Error; err != nil {
			tamperedLog := AttendanceLog{
				UserID:     userID,
				ClockIn:    eventTime,
				ClockOut:   &eventTime,
				SyncStatus: SyncStatusRejectedTampered,
				DeviceHash: req.DeviceHash,
				RecordUUID: req.RecordUUID,
			}
			tamperedLog.OrgID = orgID
			_ = s.db.Create(&tamperedLog)
		}
	}

	// Parse timestamp (RFC3339 / RFC3339Nano)
	parsedTs, err := ParseTimestamp(req.Timestamp)
	if err != nil {
		persistRejectedAudit(now)
		return nil, ErrInvalidTimestamp
	}

	// Validate online drift window (max 5 minutes)
	if !ValidateOnlineDrift(parsedTs, now, MaxOnlineClockDrift) {
		persistRejectedAudit(now)
		return nil, ErrTimestampDrift
	}

	// Verify cryptographic device hash
	if !VerifyDeviceHash(req.RecordUUID.String(), userID.String(), "CLOCK_OUT", req.Timestamp, req.DeviceHash) {
		persistRejectedAudit(parsedTs)
		return nil, ErrTamperedHash
	}

	// Valid clock-out: update active verified session
	duration := parsedTs.Sub(active.ClockIn).Hours()
	if duration < 0 {
		duration = 0
	}

	active.ClockOut = &parsedTs
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

		if !VerifyDeviceHash(rec.RecordUUID, rec.UserID, rec.ActionType, rec.Timestamp, rec.DeviceHash) {
			nowUTC := time.Now().UTC()
			tamperedLog := AttendanceLog{
				UserID:     parsedUserID,
				ClockIn:    nowUTC,
				ClockOut:   &nowUTC,
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

		ts, err := ParseTimestamp(rec.Timestamp)
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
			if err := s.db.Where("org_id = ? AND user_id = ? AND clock_out IS NULL AND sync_status = ?",
				orgID, parsedUserID, SyncStatusSyncedVerified).
				Order("clock_in desc").
				First(&active).Error; err == nil {

				duration := ts.Sub(active.ClockIn).Hours()
				if duration < 0 {
					duration = 0
				}
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

func (s *service) GetSelfHistory(orgID, userID uuid.UUID, query AttendanceHistoryQuery) (*AttendanceHistoryResponse, error) {
	base := s.db.Where("org_id = ? AND user_id = ?", orgID, userID)
	dbQuery, page, limit := s.buildQuery(base, query)
	return s.fetchPaginated(dbQuery, page, limit)
}

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

func (s *service) GetOrgHistory(orgID uuid.UUID, query AttendanceHistoryQuery) (*AttendanceHistoryResponse, error) {
	base := s.db.Where("org_id = ?", orgID)
	if query.UserID != nil && *query.UserID != uuid.Nil {
		base = base.Where("user_id = ?", *query.UserID)
	}
	dbQuery, page, limit := s.buildQuery(base, query)
	return s.fetchPaginated(dbQuery, page, limit)
}