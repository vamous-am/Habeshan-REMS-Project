package attendance

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/habeshan-rems/backend/internal/common"
	"gorm.io/gorm"
)

var (
	ErrActiveClockInExists = errors.New("user already has an active clock-in session")
)

type Service interface {
	ClockIn(orgID, userID uuid.UUID, req ClockInRequest) (*AttendanceLog, error)
}

type service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) Service {
	return &service{db: db}
}

func (s *service) ClockIn(orgID, userID uuid.UUID, req ClockInRequest) (*AttendanceLog, error) {
	// 1. Ensure user does not already have an open session (clock_out IS NULL)
	var existing AttendanceLog
	err := s.db.Where("org_id = ? AND user_id = ? AND clock_out IS NULL", orgID, userID).First(&existing).Error
	if err == nil {
		return nil, ErrActiveClockInExists
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// 2. Default to UTC server time if clock_in is omitted
	clockInTime := req.ClockIn
	if clockInTime.IsZero() {
		clockInTime = time.Now().UTC()
	}

	// 3. Create the log entry with SYNCED_VERIFIED status for direct online requests
	log := AttendanceLog{
		BaseModel: common.BaseModel{
			OrgID: orgID,
		},
		UserID:     userID,
		ClockIn:    clockInTime,
		ClockOut:   nil,
		SyncStatus: SyncStatusSyncedVerified,
		DeviceHash: req.DeviceHash,
		RecordUUID: req.RecordUUID,
	}

	if err := s.db.Create(&log).Error; err != nil {
		return nil, err
	}

	return &log, nil
}
