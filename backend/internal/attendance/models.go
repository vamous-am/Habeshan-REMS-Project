package attendance

import (
	"time"

	"github.com/google/uuid"
	"github.com/habeshan-rems/backend/internal/common"
)

// SyncStatus defines the valid states for offline sync verification
type SyncStatus string

const (
	SyncStatusOfflineLogged    SyncStatus = "OFFLINE_LOGGED"
	SyncStatusPendingSync      SyncStatus = "PENDING_SYNC"
	SyncStatusSyncedVerified   SyncStatus = "SYNCED_VERIFIED"
	SyncStatusRejectedTampered SyncStatus = "REJECTED_TAMPERED"
)

// AttendanceLog represents an employee clock-in/out session in PostgreSQL
type AttendanceLog struct {
	common.BaseModel
	UserID     uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	ClockIn    time.Time  `gorm:"not null" json:"clock_in"`
	ClockOut   *time.Time `gorm:"default:null" json:"clock_out"`
	TotalHours *float64   `gorm:"type:numeric(6,2);default:null" json:"total_hours"`
	SyncStatus SyncStatus `gorm:"type:varchar(20);not null;default:'PENDING_SYNC'" json:"sync_status"`
	DeviceHash string     `gorm:"type:varchar(128);not null" json:"device_hash"`
	RecordUUID uuid.UUID  `gorm:"type:uuid;uniqueIndex;not null" json:"record_uuid"`
}

func (AttendanceLog) TableName() string {
	return "attendance_logs"
}
