package tasks

import (
	"time"

	"github.com/google/uuid"
	"github.com/habeshan-rems/backend/internal/common"
	"github.com/habeshan-rems/backend/internal/auth"
)

// ─── Priority ────────────────────────────────────────────────────────────────

// Priority is the task urgency level (FR-TASK-01).
// Only three values are permitted — do not add a fourth.
type Priority string

const (
	PriorityHigh   Priority = "high"
	PriorityMedium Priority = "medium"
	PriorityLow    Priority = "low"
)

func (p Priority) IsValid() bool {
	switch p {
	case PriorityHigh, PriorityMedium, PriorityLow:
		return true
	}
	return false
}

// ─── Status ───────────────────────────────────────────────────────────────────

// Status is the lifecycle state of a task (FR-TASK-04).
// Exactly five values — the authoritative transition rules live in the service layer.
type Status string

const (
	StatusToDo       Status = "to_do"
	StatusInProgress Status = "in_progress"
	StatusPaused     Status = "paused"
	StatusBlocked    Status = "blocked"
	StatusCompleted  Status = "completed"
)

func (s Status) IsValid() bool {
	switch s {
	case StatusToDo, StatusInProgress, StatusPaused, StatusBlocked, StatusCompleted:
		return true
	}
	return false
}

// ─── SyncStatus ───────────────────────────────────────────────────────────────

// SyncStatus mirrors the attendance_logs sync state machine (FR-TASK-10).
// The four values and their meaning are identical to attendance_logs.sync_status.
type SyncStatus string

const (
	SyncOfflineLogged    SyncStatus = "offline_logged"
	SyncPendingSync      SyncStatus = "pending_sync"
	SyncSyncedVerified   SyncStatus = "synced_verified"
	SyncRejectedTampered SyncStatus = "rejected_tampered"
)

func (s SyncStatus) IsValid() bool {
	switch s {
	case SyncOfflineLogged, SyncPendingSync, SyncSyncedVerified, SyncRejectedTampered:
		return true
	}
	return false
}

// ─── PauseReason ─────────────────────────────────────────────────────────────

// PauseReason is the mandatory reason when a timer is paused (FR-TASK-06).
// A pause without a valid reason is rejected by the service layer.
type PauseReason string

const (
	PauseReasonPowerOutage         PauseReason = "Power Outage"
	PauseReasonPersonalBreak       PauseReason = "Personal Break"
	PauseReasonCommuteTravel       PauseReason = "Commute/Travel"
	PauseReasonWaitingOnDependency PauseReason = "Waiting on Dependency"
	PauseReasonOther               PauseReason = "Other"
)

func (r PauseReason) IsValid() bool {
	switch r {
	case PauseReasonPowerOutage,
		PauseReasonPersonalBreak,
		PauseReasonCommuteTravel,
		PauseReasonWaitingOnDependency,
		PauseReasonOther:
		return true
	}
	return false
}

// ─── Task ─────────────────────────────────────────────────────────────────────

// Task is the central aggregate.
// Embeds common.ID (uuid PK), common.TenantScoped (org_id), common.Timestamps.
// No soft-delete: only organizations and users carry deleted_at per schema contract.
type Task struct {
	common.ID
	common.TenantScoped
	common.Timestamps

	Title       string    `gorm:"type:varchar(150);not null"                 json:"title"`
	Description *string   `gorm:"type:text"                                  json:"description"`
	CreatedBy   uuid.UUID `gorm:"type:uuid;not null;index"                   json:"created_by"`
	Priority    Priority  `gorm:"type:varchar(10);not null;default:'medium'" json:"priority"`
	Status      Status    `gorm:"type:varchar(20);not null;default:'to_do'"  json:"status"`
	// DueDate is a plain calendar date (no time-of-day), nullable (FR-TASK-01).
	// Stored as DATE in Postgres; *time.Time so nil ↔ NULL round-trips cleanly.
	DueDate *time.Time `gorm:"type:date" json:"due_date"`

	// GORM associations — preload only, not stored columns.
	Creator      auth.User             `gorm:"foreignKey:CreatedBy;references:ID" json:"-"`
	Organization auth.Organization     `gorm:"foreignKey:OrgID;references:ID"     json:"-"`
	Assignments  []TaskAssignment `gorm:"foreignKey:TaskID;references:ID"    json:"-"`
}

// ─── TaskAssignment ───────────────────────────────────────────────────────────

// TaskAssignment is the many-to-many junction table between tasks and employees.
// It has a composite PK (task_id, user_id) — no surrogate UUID, no org_id
// (transitively scoped through tasks).  Schema contract §task_assignments.
type TaskAssignment struct {
	TaskID     uuid.UUID `gorm:"type:uuid;primaryKey"              json:"task_id"`
	UserID     uuid.UUID `gorm:"type:uuid;primaryKey"              json:"user_id"`
	AssignedAt time.Time `gorm:"autoCreateTime"                    json:"assigned_at"`

	// GORM associations — preload only.
	Task Task            `gorm:"foreignKey:TaskID;references:ID" json:"-"`
	User auth.User `gorm:"foreignKey:UserID;references:ID" json:"-"`
}

// ─── TaskTimeLog ──────────────────────────────────────────────────────────────

// TaskTimeLog records a single timer segment (start → pause/stop).
// Each start, resume, pause, and stop produces a new row — history is never
// overwritten, keeping the log auditable (FR-TASK-08).
//
// Embeds common.ID (uuid PK) and common.Timestamps.
// No OrgID — org isolation is reached transitively through task_id → tasks.org_id,
// which avoids the redundant column while keeping queries correct.
//
// record_uuid — client-generated idempotency key (FR-TASK-10 / FR-ATT-08).
//
//	UNIQUE constraint prevents duplicate sync submissions.
//
// device_hash — tamper-evident fingerprint of the submitting device (FR-TASK-10).
// sync_status — offline sync state machine, same values as attendance_logs.
type TaskTimeLog struct {
	common.ID
	common.Timestamps

	TaskID          uuid.UUID  `gorm:"type:uuid;not null;index"                         json:"task_id"`
	UserID          uuid.UUID  `gorm:"type:uuid;not null;index"                         json:"user_id"`
	StartedAt       time.Time  `gorm:"not null"                                         json:"started_at"`
	StoppedAt       *time.Time `gorm:"default:null"                                     json:"stopped_at"`
	DurationMinutes *int       `gorm:"default:null"                                     json:"duration_minutes"`
	// PauseReason is mandatory when the timer is paused; NULL for all other rows.
	PauseReason *string    `gorm:"type:varchar(50);default:null"                    json:"pause_reason,omitempty"`
	SyncStatus  SyncStatus `gorm:"type:varchar(30);not null;default:'pending_sync'" json:"sync_status"`
	DeviceHash  string     `gorm:"type:varchar(128);not null"                       json:"device_hash"`
	RecordUUID  uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex"                   json:"record_uuid"`

	// GORM associations — preload only.
	Task Task `gorm:"foreignKey:TaskID;references:ID" json:"-"`
	User auth.User `gorm:"foreignKey:UserID;references:ID" json:"-"`
}
