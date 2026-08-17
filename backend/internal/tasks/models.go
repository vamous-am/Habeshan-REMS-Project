package tasks

import "time"

// Task with Priority and Status ENUMs
type Priority string

const (
	HIGH_PRIORITY   Priority = "high"
	MEDIUM_PRIORITY Priority = "medium"
	LOW_PRIORITY    Priority = "low"
)

type Status string

const (
	TO_DO       Status = "to_do"
	IN_PROGRESS Status = "in_progress"
	PAUSED      Status = "paused"
	BLOCKED     Status = "blocked"
	COMPLETED   Status = "completed"
)

type Task struct {
	TaskID      uint       `json:"task_id" gorm:"primaryKey"`
	Title       string     `json:"title" gorm:"type:varchar(150);not null"`
	Description string     `json:"description" gorm:"type:text"`
	OrgID       uint       `json:"org_id" gorm:"not null"`
	CreatedBy   uint       `json:"created_by" gorm:"not null"`
	Priority    Priority   `json:"priority" gorm:"default:medium"`
	Status      Status     `json:"status" gorm:"default:to_do;not null"`
	DueDate     *time.Time `json:"due_date"`
	CreatedAt   time.Time  `json:"created_at" gorm:"autoCreateTime"`

	Organization Organization `gorm:"foreignKey:OrgID;references:ID"`
	Creator      User         `gorm:"foreingKey:UserID;references:ID"`
}

// TaskAssignment
type TaskAssignment struct {
	TaskID         uint       `json:"task_id" gorm:"primaryKey"`
	AssignedUserID uint       `json:"assigned_user_id" gorm:"primaryKey"`
	AssignedAt     *time.Time `json:"assigned_at" gorm:"autoCreateTime"`

	Task Task `gorm:"foreignKey:TaskID;references:TaskID"`
}

// TaskTimeLog and SyncStatus ENUM
type SyncStatus string

const (
	OFFLINE_LOGGED    SyncStatus = "offline_logged"
	PENDING_SYNC      SyncStatus = "pending_sync"
	SYNCED_VERIFIED   SyncStatus = "synced_verified"
	REJECTED_TAMPERED SyncStatus = "rejected_tampered"
)

type TaskTimeLog struct {
	LogID           uint       `json:"log_id" gorm:"primaryKey"`
	TaskID          uint       `json:"task_id" gorm:"not null"`
	UserID          uint       `json:"user_id" gorm:"not null"`
	StartTime       time.Time  `json:"start_time" gorm:"not null"`
	EndTime         *time.Time `json:"end_time"`
	DurationMinutes uint       `json:"duration_minutes"`
	PauseReason     string     `json:"pause_reason" gorm:"type:varchar(50)"`
	SyncStatus      SyncStatus `json:"sync_status" gorm:"default:pending_sync;not null"`
	DeviceHash      string     `json:"device_hash" gorm:"type:varchar(128);not null"`

	Task Task `gorm:"foreignKey:TaskID;references:TaskID"`
	User User `gorm:"foreingKey:UserID;references:UserID"`
}
