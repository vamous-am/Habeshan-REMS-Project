package tasks

// service_assignments.go — FR-TASK-02 (assign / unassign)

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AssignTask — FR-TASK-02
// Assigns one or more employees to a task in one bulk write.
// Duplicate assignments are silently skipped (ON CONFLICT DO NOTHING).
func (s *taskService) AssignTask(taskID uuid.UUID, userIDs []uuid.UUID, callerID, orgID uuid.UUID) error {
	if len(userIDs) == 0 {
		return fmt.Errorf("%w: at least one user ID is required", ErrBadRequest)
	}

	caller, err := s.resolveUser(callerID)
	if err != nil {
		return err
	}
	if err := requireOrgMatch(caller, orgID); err != nil {
		return err
	}
	if err := requireRole(caller, RoleAdmin, RoleManager); err != nil {
		return fmt.Errorf("%w: only admins and managers can assign tasks", err)
	}

	if _, err := s.taskRepo.GetTaskByIDForOrg(taskID, orgID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		return ErrInternal
	}

	assignments := make([]TaskAssignment, 0, len(userIDs))
	for _, uid := range userIDs {
		u, err := s.userRepo.GetUserByID(uid)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("%w: user %s not found", ErrBadRequest, uid)
			}
			return ErrInternal
		}
		if u.OrgID != orgID {
			return fmt.Errorf("%w: user %s does not belong to this organisation", ErrForbidden, uid)
		}
		assignments = append(assignments, TaskAssignment{TaskID: taskID, UserID: uid})
	}

	if err := s.assignRepo.BulkAssignTask(assignments); err != nil {
		return ErrInternal
	}
	return nil
}

// UnassignTask — FR-TASK-02
// Removes a single employee from a task.
func (s *taskService) UnassignTask(taskID, userID uuid.UUID, callerID, orgID uuid.UUID) error {
	caller, err := s.resolveUser(callerID)
	if err != nil {
		return err
	}
	if err := requireOrgMatch(caller, orgID); err != nil {
		return err
	}
	if err := requireRole(caller, RoleAdmin, RoleManager); err != nil {
		return fmt.Errorf("%w: only admins and managers can unassign tasks", err)
	}

	if _, err := s.taskRepo.GetTaskByIDForOrg(taskID, orgID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		return ErrInternal
	}

	if err := s.assignRepo.UnassignTask(taskID, userID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		return ErrInternal
	}
	return nil
}
