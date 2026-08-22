package tasks

import (
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/habeshan-rems/backend/internal/common"
	"github.com/habeshan-rems/backend/internal/auth"
	"gorm.io/gorm"
)

// Re-export common sentinel errors so callers inside this package don't need
// an import alias for common.
var (
	ErrNotFound     = common.ErrNotFound
	ErrConflict     = common.ErrConflict
	ErrInternal     = common.ErrInternal
	ErrBadRequest   = common.ErrBadRequest
	ErrUnauthorized = common.ErrUnauthorized
	ErrForbidden    = common.ErrForbidden
)

// ─── TaskService interface ────────────────────────────────────────────────────

type TaskService interface {
	// FR-TASK-01
	CreateTask(req CreateTaskRequest, callerID uuid.UUID) (Task, error)

	// FR-TASK-03
	GetMyTasks(callerID, orgID uuid.UUID) ([]Task, error)
	GetTaskByID(taskID uuid.UUID, callerID, orgID uuid.UUID) (Task, error)

	// FR-TASK-02
	AssignTask(taskID uuid.UUID, userIDs []uuid.UUID, callerID, orgID uuid.UUID) error
	UnassignTask(taskID, userID uuid.UUID, callerID, orgID uuid.UUID) error

	// FR-TASK-04
	UpdateTaskStatus(taskID uuid.UUID, newStatus Status, callerID, orgID uuid.UUID) error

	// FR-TASK-05 / FR-TASK-06 / FR-TASK-08
	StartTimer(req TimerStartRequest, callerID, orgID uuid.UUID) (TaskTimeLog, error)
	PauseTimer(req TimerPauseRequest, callerID, orgID uuid.UUID) (TaskTimeLog, error)
	ResumeTimer(req TimerResumeRequest, callerID, orgID uuid.UUID) (TaskTimeLog, error)
	StopTimer(req TimerStopRequest, callerID, orgID uuid.UUID) (TaskTimeLog, error)

	// FR-TASK-07 / FR-TASK-08
	GetTimerHistory(taskID uuid.UUID, callerID, orgID uuid.UUID) ([]TaskTimeLog, error)

	// FR-TASK-09
	GetTaskStatusCounts(callerID, orgID uuid.UUID) (map[Status]int64, error)
	GetOverdueTasks(callerID, orgID uuid.UUID) ([]Task, error)
}

// ─── taskService struct & constructor ────────────────────────────────────────

type taskService struct {
	taskRepo    TaskRepository
	assignRepo  TaskAssignmentRepository
	timeLogRepo TaskTimeLogRepository
	userRepo    auth.UserRepository
	orgRepo     auth.OrganizationRepository
}

func NewService(
	taskRepo TaskRepository,
	assignRepo TaskAssignmentRepository,
	timeLogRepo TaskTimeLogRepository,
	userRepo auth.UserRepository,
	orgRepo auth.OrganizationRepository,
) TaskService {
	return &taskService{
		taskRepo:    taskRepo,
		assignRepo:  assignRepo,
		timeLogRepo: timeLogRepo,
		userRepo:    userRepo,
		orgRepo:     orgRepo,
	}
}

// ─── shared helpers ───────────────────────────────────────────────────────────

// resolveUser fetches the caller and maps a not-found into ErrUnauthorized.
func (s *taskService) resolveUser(userID uuid.UUID) (auth.User, error) {
	u, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return auth.User{}, fmt.Errorf("%w: user not found", ErrUnauthorized)
		}
		return auth.User{}, ErrInternal
	}
	return u, nil
}

// requireOrgMatch returns ErrForbidden when the user is not in orgID.
func requireOrgMatch(user auth.User, orgID uuid.UUID) error {
	if user.OrgID != orgID {
		return ErrForbidden
	}
	return nil
}

// requireRole returns ErrForbidden unless the caller holds one of the given roles.
func requireRole(user auth.User, roles ...auth.Role) error {
	for _, r := range roles {
		if user.Role == r {
			return nil
		}
	}
	return ErrForbidden
}

// canSeeTask checks whether a caller is allowed to read a task.
//   - admins / managers : any task in their org
//   - employees         : only tasks assigned to them
func (s *taskService) canSeeTask(caller auth.User, task Task, orgID uuid.UUID) error {
	if caller.OrgID != orgID || task.OrgID != orgID {
		return ErrForbidden
	}
	if caller.Role == auth.RoleAdmin || caller.Role == auth.RoleManager {
		return nil
	}
	if caller.Role == auth.RoleEmployee {
		assigned, err := s.assignRepo.IsTaskAssignedToUser(task.ID.ID, caller.ID.ID)
		if err != nil {
			return ErrInternal
		}
		if !assigned {
			return ErrForbidden
		}
		return nil
	}
	return ErrForbidden
}

// durationMinutes returns floor(stop−start) in minutes as a nullable *int.
func durationMinutes(start, stop time.Time) *int {
	d := int(math.Floor(stop.Sub(start).Minutes()))
	if d < 0 {
		d = 0
	}
	return &d
}

// idempotentLookup checks whether a record_uuid was already processed.
// Returns (log, nil) on a hit, (zero, nil) on a miss, (zero, err) on DB error.
func (s *taskService) idempotentLookup(recordUUID uuid.UUID) (TaskTimeLog, bool, error) {
	existing, err := s.timeLogRepo.FindByRecordUUID(recordUUID)
	if err == nil {
		return existing, true, nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return TaskTimeLog{}, false, nil
	}
	return TaskTimeLog{}, false, ErrInternal
}
