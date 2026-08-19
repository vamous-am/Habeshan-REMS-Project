package attendance

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/habeshan-rems/backend/internal/common"
	"gorm.io/gorm"
)

var (
	// ErrActiveClockInExists is returned when a user attempts to clock in without closing an open session
	ErrActiveClockInExists = errors.New("user already has an active clock-in session")
	// ErrNoActiveClockIn is returned when a user attempts to clock out without an active open session
	ErrNoActiveClockIn = errors.New("no active clock-in session found")
)

type Service interface {
	ClockIn(orgID, userID uuid.UUID, req ClockInRequest) (*AttendanceLog, error)
	ClockOut(orgID, userID uuid.UUID, req ClockOutRequest) (*AttendanceLog, error)
}

type service struct {
	db *gorm.DB
}

// NewService constructs a new attendance service instance
func NewService(db *gorm.DB) Service {
	return &service{db: db}
}

// ClockIn enforces business rules and creates an attendance record
func (s *service) ClockIn(orgID, userID uuid.UUID, req ClockInRequest) (*AttendanceLog, error) {
	// 1. Check for an active open session (clock_out IS NULL)
	var existing AttendanceLog
	err := s.db.Where("org_id = ? AND user_id = ? AND clock_out IS NULL", orgID, userID).First(&existing).Error
	if err == nil {
		return nil, ErrActiveClockInExists
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// 2. Use request timestamp or fallback to current UTC time
	clockInTime := req.Timestamp
	if clockInTime.IsZero() {
		clockInTime = time.Now().UTC()
	}

	// 3. Construct the record with SYNCED_VERIFIED status for online clock-in
	log := AttendanceLog{
		BaseModel: common.BaseModel{
			TenantScoped: common.TenantScoped{
				OrgID: orgID,
			},
		},
		UserID:     userID,
		ClockIn:    clockInTime,
		ClockOut:   nil,
		SyncStatus: SyncStatusSyncedVerified,
		DeviceHash: req.DeviceHash,
		RecordUUID: req.RecordUUID,
	}

	// 4. Save to PostgreSQL via GORM
	if err := s.db.Create(&log).Error; err != nil {
		return nil, err
	}

	return &log, nil
}

// ClockOut finds the active open session, updates clock_out timestamp, and calculates total_hours
func (s *service) ClockOut(orgID, userID uuid.UUID, req ClockOutRequest) (*AttendanceLog, error) {
	// 1. Find active open session (clock_out IS NULL)
	var log AttendanceLog
	err := s.db.Where("org_id = ? AND user_id = ? AND clock_out IS NULL", orgID, userID).First(&log).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNoActiveClockIn
	} else if err != nil {
		return nil, err
	}

	// 2. Use request timestamp or fallback to current UTC time
	clockOutTime := req.Timestamp
	if clockOutTime.IsZero() {
		clockOutTime = time.Now().UTC()
	}

	// 3. Calculate total duration in hours
	duration := clockOutTime.Sub(log.ClockIn).Hours()
	if duration < 0 {
		duration = 0.0
	}

	log.ClockOut = &clockOutTime
	log.TotalHours = &duration

	// 4. Save updated record
	if err := s.db.Save(&log).Error; err != nil {
		return nil, err
	}

	return &log, nil
}