package tasks

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ─── TaskTimeLogRepository interface ─────────────────────────────────────────

// TaskTimeLogRepository defines all database operations for task_time_logs.
// Column names match the schema contract exactly: started_at, stopped_at,
// duration_minutes, pause_reason, sync_status, device_hash, record_uuid.
type TaskTimeLogRepository interface {
	// CreateTimeLog inserts a new timer segment row.
	CreateTimeLog(log *TaskTimeLog) error

	// UpsertTimeLog inserts or updates a timer segment using record_uuid as the
	// idempotency key (FR-TASK-10 / FR-ATT-08 pattern).  A second submission
	// with the same record_uuid is ignored (ON CONFLICT DO NOTHING).
	UpsertTimeLog(log *TaskTimeLog) error

	// GetActiveTimeLog returns the open timer segment for (taskID, userID) —
	// the row where stopped_at IS NULL.  Returns gorm.ErrRecordNotFound when
	// no open segment exists.
	GetActiveTimeLog(taskID, userID uuid.UUID) (TaskTimeLog, error)

	// GetTimeLogsByTaskID returns the full timer history for a task ordered
	// by started_at ASC (auditable history, FR-TASK-08).
	GetTimeLogsByTaskID(taskID uuid.UUID) ([]TaskTimeLog, error)

	// GetTimeLogsByTaskAndUser returns the timer history for a specific
	// (task, user) pair, ordered by started_at ASC.
	GetTimeLogsByTaskAndUser(taskID, userID uuid.UUID) ([]TaskTimeLog, error)

	// FindByRecordUUID looks up a log entry by its client-generated idempotency
	// key.  Used to detect duplicate sync submissions (FR-TASK-10).
	FindByRecordUUID(recordUUID uuid.UUID) (TaskTimeLog, error)

	// UpdateTimeLog persists changes to an existing timer segment (e.g. setting
	// stopped_at and duration_minutes on stop/pause).
	UpdateTimeLog(log *TaskTimeLog) error
}

// ─── taskTimeLogRepository implementation ────────────────────────────────────

type taskTimeLogRepository struct {
	db *gorm.DB
}

func NewTaskTimeLogRepository(db *gorm.DB) TaskTimeLogRepository {
	return &taskTimeLogRepository{db: db}
}

// ── writes ────────────────────────────────────────────────────────────────────

func (r *taskTimeLogRepository) CreateTimeLog(log *TaskTimeLog) error {
	return r.db.Create(log).Error
}

// UpsertTimeLog uses ON CONFLICT (record_uuid) DO NOTHING so that replaying
// the same sync payload is safe and idempotent.  This mirrors the attendance
// offline-sync pattern (FR-TASK-10 / FR-ATT-08).
func (r *taskTimeLogRepository) UpsertTimeLog(log *TaskTimeLog) error {
	return r.db.
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "record_uuid"}},
			DoNothing: true,
		}).
		Create(log).Error
}

func (r *taskTimeLogRepository) UpdateTimeLog(log *TaskTimeLog) error {
	return r.db.Save(log).Error
}

// ── reads ─────────────────────────────────────────────────────────────────────

// GetActiveTimeLog finds the open timer segment (stopped_at IS NULL).
// A task can have at most one open segment per user at any given moment;
// the service layer enforces this invariant before calling CreateTimeLog.
func (r *taskTimeLogRepository) GetActiveTimeLog(taskID, userID uuid.UUID) (TaskTimeLog, error) {
	var log TaskTimeLog
	err := r.db.
		Where("task_id = ? AND user_id = ? AND stopped_at IS NULL", taskID, userID).
		First(&log).Error
	if err != nil {
		return TaskTimeLog{}, err
	}
	return log, nil
}

// GetTimeLogsByTaskID returns all timer segments for a task in chronological
// order.  Includes segments from all users assigned to the task.
func (r *taskTimeLogRepository) GetTimeLogsByTaskID(taskID uuid.UUID) ([]TaskTimeLog, error) {
	var logs []TaskTimeLog
	if err := r.db.
		Where("task_id = ?", taskID).
		Order("started_at ASC").
		Find(&logs).Error; err != nil {
		return nil, err
	}
	return logs, nil
}

// GetTimeLogsByTaskAndUser returns the timer history for one (task, user) pair.
func (r *taskTimeLogRepository) GetTimeLogsByTaskAndUser(taskID, userID uuid.UUID) ([]TaskTimeLog, error) {
	var logs []TaskTimeLog
	if err := r.db.
		Where("task_id = ? AND user_id = ?", taskID, userID).
		Order("started_at ASC").
		Find(&logs).Error; err != nil {
		return nil, err
	}
	return logs, nil
}

// FindByRecordUUID looks up a log entry by its unique idempotency key.
func (r *taskTimeLogRepository) FindByRecordUUID(recordUUID uuid.UUID) (TaskTimeLog, error) {
	var log TaskTimeLog
	err := r.db.
		Where("record_uuid = ?", recordUUID).
		First(&log).Error
	if err != nil {
		return TaskTimeLog{}, err
	}
	return log, nil
}
