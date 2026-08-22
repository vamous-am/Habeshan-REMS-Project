package tasks

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ─── TaskAssignmentRepository interface ───────────────────────────────────────

// TaskAssignmentRepository defines all database operations for task_assignments.
// Column names match the schema contract: task_id, user_id, assigned_at.
type TaskAssignmentRepository interface {
	// AssignTask inserts a single assignment row.  The caller must check for
	// duplicates before calling, or handle the returned conflict error.
	AssignTask(assignment *TaskAssignment) error

	// BulkAssignTask inserts multiple assignment rows in one statement.
	// Rows whose (task_id, user_id) pair already exists are skipped via
	// ON CONFLICT DO NOTHING, making the operation idempotent.
	BulkAssignTask(assignments []TaskAssignment) error

	// UnassignTask removes the (taskID, userID) row.
	// Returns gorm.ErrRecordNotFound when the row does not exist.
	UnassignTask(taskID, userID uuid.UUID) error

	// GetAssignmentsByTaskID returns all assignment rows for a task.
	GetAssignmentsByTaskID(taskID uuid.UUID) ([]TaskAssignment, error)

	// GetAssignedUserIDs returns just the user_id values for a task.
	// Use this when only the slice of IDs is needed to avoid loading User rows.
	GetAssignedUserIDs(taskID uuid.UUID) ([]uuid.UUID, error)

	// IsTaskAssignedToUser checks whether a (taskID, userID) pair exists.
	IsTaskAssignedToUser(taskID, userID uuid.UUID) (bool, error)
}

// ─── taskAssignmentRepository implementation ──────────────────────────────────

type taskAssignmentRepository struct {
	db *gorm.DB
}

func NewTaskAssignmentRepository(db *gorm.DB) TaskAssignmentRepository {
	return &taskAssignmentRepository{db: db}
}

func (r *taskAssignmentRepository) AssignTask(assignment *TaskAssignment) error {
	return r.db.Create(assignment).Error
}

// BulkAssignTask inserts all rows with ON CONFLICT (task_id, user_id) DO NOTHING
// so duplicate assignments are silently skipped in one round-trip.
func (r *taskAssignmentRepository) BulkAssignTask(assignments []TaskAssignment) error {
	if len(assignments) == 0 {
		return nil
	}
	return r.db.
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(&assignments).Error
}

func (r *taskAssignmentRepository) UnassignTask(taskID, userID uuid.UUID) error {
	result := r.db.
		Where("task_id = ? AND user_id = ?", taskID, userID).
		Delete(&TaskAssignment{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *taskAssignmentRepository) GetAssignmentsByTaskID(taskID uuid.UUID) ([]TaskAssignment, error) {
	var assignments []TaskAssignment
	if err := r.db.
		Where("task_id = ?", taskID).
		Find(&assignments).Error; err != nil {
		return nil, err
	}
	return assignments, nil
}

// GetAssignedUserIDs returns only the user_id column — one scan, no extra JOIN.
func (r *taskAssignmentRepository) GetAssignedUserIDs(taskID uuid.UUID) ([]uuid.UUID, error) {
	var userIDs []uuid.UUID
	if err := r.db.
		Model(&TaskAssignment{}).
		Select("user_id").
		Where("task_id = ?", taskID).
		Scan(&userIDs).Error; err != nil {
		return nil, err
	}
	return userIDs, nil
}

func (r *taskAssignmentRepository) IsTaskAssignedToUser(taskID, userID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.
		Model(&TaskAssignment{}).
		Where("task_id = ? AND user_id = ?", taskID, userID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
