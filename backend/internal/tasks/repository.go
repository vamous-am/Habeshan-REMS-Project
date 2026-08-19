package tasks

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ─── TaskRepository interface ─────────────────────────────────────────────────

// TaskRepository defines all database operations for the tasks table.
type TaskRepository interface {
	// CreateTask inserts a new task row.
	CreateTask(task *Task) error

	// GetTaskByID fetches a single task by PK.  Returns common.ErrNotFound if
	// missing.
	GetTaskByID(taskID uuid.UUID) (Task, error)

	// GetTaskByIDForOrg fetches a task only if it belongs to orgID.
	GetTaskByIDForOrg(taskID, orgID uuid.UUID) (Task, error)

	// GetTaskForUser returns a task only when the given user appears in
	// task_assignments.  Used for employee-level access checks.
	GetTaskForUser(taskID, userID uuid.UUID) (Task, error)

	// GetTasksByUserID returns all tasks assigned to a specific employee via
	// task_assignments (FR-TASK-03: employees see only their own tasks).
	GetTasksByUserID(userID uuid.UUID) ([]Task, error)

	// GetTasksByOrgID returns all tasks for an organisation.
	// Used by admins (FR-TASK-03: org-level visibility).
	GetTasksByOrgID(orgID uuid.UUID) ([]Task, error)

	// GetTasksByCreator returns tasks created by a specific user within an org.
	// Used as a fallback for managers who need to see tasks they created.
	GetTasksByCreator(createdBy, orgID uuid.UUID) ([]Task, error)

	// GetTasksByAssignedUsers returns all tasks within orgID that are assigned
	// to any user in userIDs.  Used for manager visibility (FR-TASK-03): the
	// service resolves the manager's team members and passes the slice here,
	// producing a single JOIN query instead of N individual look-ups.
	GetTasksByAssignedUsers(orgID uuid.UUID, userIDs []uuid.UUID) ([]Task, error)

	// UpdateTaskStatus changes only the status column of a task.
	// The caller (service) is responsible for validating the transition.
	UpdateTaskStatus(taskID uuid.UUID, status Status) error

	// GetOverdueTasks returns tasks within orgID whose due_date is before now
	// and whose status is not completed (FR-TASK-09).
	GetOverdueTasks(orgID uuid.UUID) ([]Task, error)

	// GetTaskStatusCounts returns the count of tasks in each status bucket for
	// an organisation (FR-TASK-09).  Key is the Status string value.
	GetTaskStatusCounts(orgID uuid.UUID) (map[Status]int64, error)

	// GetOverdueTaskCountForManager returns overdue task count scoped to the
	// employees managed by managerID within orgID (FR-TASK-09).
	GetOverdueTaskCountForManager(orgID uuid.UUID, memberIDs []uuid.UUID) (int64, error)

	// GetTaskStatusCountsForManager returns status counts scoped to the given
	// member set (FR-TASK-09).
	GetTaskStatusCountsForManager(orgID uuid.UUID, memberIDs []uuid.UUID) (map[Status]int64, error)
}

// ─── taskRepository implementation ───────────────────────────────────────────

type taskRepository struct {
	db *gorm.DB
}

func NewTaskRepository(db *gorm.DB) TaskRepository {
	return &taskRepository{db: db}
}

// ── writes ────────────────────────────────────────────────────────────────────

func (r *taskRepository) CreateTask(task *Task) error {
	return r.db.Create(task).Error
}

func (r *taskRepository) UpdateTaskStatus(taskID uuid.UUID, status Status) error {
	result := r.db.
		Model(&Task{}).
		Where("id = ?", taskID).
		Update("status", status)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// ── single-row reads ──────────────────────────────────────────────────────────

func (r *taskRepository) GetTaskByID(taskID uuid.UUID) (Task, error) {
	var task Task
	err := r.db.Where("id = ?", taskID).First(&task).Error
	if err != nil {
		return Task{}, err
	}
	return task, nil
}

func (r *taskRepository) GetTaskByIDForOrg(taskID, orgID uuid.UUID) (Task, error) {
	var task Task
	err := r.db.
		Where("id = ? AND org_id = ?", taskID, orgID).
		First(&task).Error
	if err != nil {
		return Task{}, err
	}
	return task, nil
}

// GetTaskForUser returns a task only when userID appears in task_assignments
// for that task.  Uses EXISTS to avoid pulling the full assignment rows.
func (r *taskRepository) GetTaskForUser(taskID, userID uuid.UUID) (Task, error) {
	var task Task
	err := r.db.
		Where("id = ?", taskID).
		Where(
			"EXISTS (SELECT 1 FROM task_assignments WHERE task_id = tasks.id AND user_id = ?)",
			userID,
		).
		First(&task).Error
	if err != nil {
		return Task{}, err
	}
	return task, nil
}

// ── list reads ────────────────────────────────────────────────────────────────

// GetTasksByUserID returns every task assigned to userID via task_assignments.
// Single JOIN — no N+1.
func (r *taskRepository) GetTasksByUserID(userID uuid.UUID) ([]Task, error) {
	var tasks []Task
	err := r.db.
		Joins("JOIN task_assignments ON task_assignments.task_id = tasks.id").
		Where("task_assignments.user_id = ?", userID).
		Find(&tasks).Error
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

func (r *taskRepository) GetTasksByOrgID(orgID uuid.UUID) ([]Task, error) {
	var tasks []Task
	err := r.db.Where("org_id = ?", orgID).Find(&tasks).Error
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

func (r *taskRepository) GetTasksByCreator(createdBy, orgID uuid.UUID) ([]Task, error) {
	var tasks []Task
	err := r.db.
		Where("created_by = ? AND org_id = ?", createdBy, orgID).
		Find(&tasks).Error
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

// GetTasksByAssignedUsers fetches all tasks within orgID that are assigned to
// any user in userIDs.  Single JOIN + IN clause — no N+1.
func (r *taskRepository) GetTasksByAssignedUsers(orgID uuid.UUID, userIDs []uuid.UUID) ([]Task, error) {
	if len(userIDs) == 0 {
		return []Task{}, nil
	}
	var tasks []Task
	err := r.db.
		Joins("JOIN task_assignments ON task_assignments.task_id = tasks.id").
		Where("tasks.org_id = ? AND task_assignments.user_id IN ?", orgID, userIDs).
		Distinct("tasks.*").
		Find(&tasks).Error
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

// ── aggregates ────────────────────────────────────────────────────────────────

// GetOverdueTasks returns non-completed tasks whose due_date < today.
func (r *taskRepository) GetOverdueTasks(orgID uuid.UUID) ([]Task, error) {
	var tasks []Task
	err := r.db.
		Where("org_id = ? AND due_date < ? AND status != ?", orgID, time.Now(), StatusCompleted).
		Find(&tasks).Error
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

// statusCountRow is used to scan the GROUP BY result from Postgres.
type statusCountRow struct {
	Status Status
	Count  int64
}

// GetTaskStatusCounts returns a map of status → count for all tasks in an org.
func (r *taskRepository) GetTaskStatusCounts(orgID uuid.UUID) (map[Status]int64, error) {
	var rows []statusCountRow
	err := r.db.
		Model(&Task{}).
		Select("status, COUNT(*) as count").
		Where("org_id = ?", orgID).
		Group("status").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return statusRowsToMap(rows), nil
}

// GetOverdueTaskCountForManager returns count of overdue, non-completed tasks
// assigned to any of the given memberIDs within orgID.
func (r *taskRepository) GetOverdueTaskCountForManager(orgID uuid.UUID, memberIDs []uuid.UUID) (int64, error) {
	if len(memberIDs) == 0 {
		return 0, nil
	}
	var count int64
	err := r.db.
		Model(&Task{}).
		Joins("JOIN task_assignments ON task_assignments.task_id = tasks.id").
		Where(
			"tasks.org_id = ? AND tasks.due_date < ? AND tasks.status != ? AND task_assignments.user_id IN ?",
			orgID, time.Now(), StatusCompleted, memberIDs,
		).
		Distinct("tasks.id").
		Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}

// GetTaskStatusCountsForManager returns status counts scoped to a set of member IDs.
func (r *taskRepository) GetTaskStatusCountsForManager(orgID uuid.UUID, memberIDs []uuid.UUID) (map[Status]int64, error) {
	if len(memberIDs) == 0 {
		return map[Status]int64{}, nil
	}
	var rows []statusCountRow
	err := r.db.
		Model(&Task{}).
		Select("tasks.status, COUNT(DISTINCT tasks.id) as count").
		Joins("JOIN task_assignments ON task_assignments.task_id = tasks.id").
		Where("tasks.org_id = ? AND task_assignments.user_id IN ?", orgID, memberIDs).
		Group("tasks.status").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return statusRowsToMap(rows), nil
}

// statusRowsToMap converts the scanned rows to a map, backfilling zeroes for
// statuses that had no tasks so callers can always read any key safely.
func statusRowsToMap(rows []statusCountRow) map[Status]int64 {
	result := map[Status]int64{
		StatusToDo:       0,
		StatusInProgress: 0,
		StatusPaused:     0,
		StatusBlocked:    0,
		StatusCompleted:  0,
	}
	for _, row := range rows {
		result[row.Status] = row.Count
	}
	return result
}
