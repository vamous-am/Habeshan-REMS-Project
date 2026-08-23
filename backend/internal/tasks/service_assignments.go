package tasks

// service_assignments.go — FR-TASK-02 (assign / unassign)

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/habeshan-rems/backend/internal/auth"
	"gorm.io/gorm"
)

// AssignTask — FR-TASK-02
// Assigns one or more employees to a task in one bulk write.
// Duplicate assignments are silently skipped (ON CONFLICT DO NOTHING).
func (s *taskService) AssignTask(taskID uuid.UUID, userIDs []uuid.UUID, callerID, orgID uuid.UUID) error {
	return s.AssignTaskByIdentifiers(taskID, userIDs, nil, nil, callerID, orgID)
}

// AssignTaskByIdentifiers resolves user IDs, email addresses, or names to users in the org
// and bulk assigns them to the specified task.
func (s *taskService) AssignTaskByIdentifiers(
	taskID uuid.UUID,
	userIDs []uuid.UUID,
	emails []string,
	identifiers []string,
	callerID, orgID uuid.UUID,
) error {
	caller, err := s.resolveUser(callerID)
	if err != nil {
		return err
	}
	if err := requireOrgMatch(caller, orgID); err != nil {
		return err
	}
	if err := requireRole(caller, auth.RoleAdmin, auth.RoleManager); err != nil {
		return fmt.Errorf("%w: only admins and managers can assign tasks", err)
	}

	if _, err := s.taskRepo.GetTaskByIDForOrg(taskID, orgID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		return ErrInternal
	}

	targetMap := make(map[uuid.UUID]bool)

	// Add direct userIDs
	for _, uid := range userIDs {
		if uid != uuid.Nil {
			targetMap[uid] = true
		}
	}

	// Resolve email strings
	for _, email := range emails {
		if email == "" {
			continue
		}
		u, err := s.userRepo.FindUserByIdentifier(orgID, email)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("%w: user with email %q not found in organization", ErrBadRequest, email)
			}
			return ErrInternal
		}
		targetMap[u.ID.ID] = true
	}

	// Resolve arbitrary identifiers (UUID string, email, or name)
	for _, ident := range identifiers {
		if ident == "" {
			continue
		}
		u, err := s.userRepo.FindUserByIdentifier(orgID, ident)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("%w: user %q not found in organization", ErrBadRequest, ident)
			}
			return ErrInternal
		}
		targetMap[u.ID.ID] = true
	}

	if len(targetMap) == 0 {
		return fmt.Errorf("%w: at least one valid user identifier or email is required", ErrBadRequest)
	}

	assignments := make([]TaskAssignment, 0, len(targetMap))
	for uid := range targetMap {
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
// Removes a single employee from a task by user ID.
func (s *taskService) UnassignTask(taskID, userID uuid.UUID, callerID, orgID uuid.UUID) error {
	caller, err := s.resolveUser(callerID)
	if err != nil {
		return err
	}
	if err := requireOrgMatch(caller, orgID); err != nil {
		return err
	}
	if err := requireRole(caller, auth.RoleAdmin, auth.RoleManager); err != nil {
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

// UnassignTaskByIdentifier unassigns by either user UUID string or email address.
func (s *taskService) UnassignTaskByIdentifier(taskID uuid.UUID, identifier string, callerID, orgID uuid.UUID) error {
	u, err := s.userRepo.FindUserByIdentifier(orgID, identifier)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("%w: user %q not found", ErrNotFound, identifier)
		}
		return ErrInternal
	}
	return s.UnassignTask(taskID, u.ID.ID, callerID, orgID)
}

// GetAssignableUsers returns all active users in the org for manager/admin assignment.
func (s *taskService) GetAssignableUsers(callerID, orgID uuid.UUID) ([]auth.User, error) {
	caller, err := s.resolveUser(callerID)
	if err != nil {
		return nil, err
	}
	if err := requireOrgMatch(caller, orgID); err != nil {
		return nil, err
	}
	return s.userRepo.GetUsersByOrgID(orgID)
}

// GetUsersDetails converts a slice of user UUIDs into public AssignedUserDTOs.
func (s *taskService) GetUsersDetails(userIDs []uuid.UUID) ([]AssignedUserDTO, error) {
	if len(userIDs) == 0 {
		return []AssignedUserDTO{}, nil
	}
	users, err := s.userRepo.GetUsersByIDs(userIDs)
	if err != nil {
		return nil, err
	}
	result := make([]AssignedUserDTO, len(users))
	for i, u := range users {
		result[i] = AssignedUserDTO{
			ID:       u.ID.ID,
			FullName: u.FullName,
			Email:    u.Email,
			Role:     string(u.Role),
		}
	}
	return result, nil
}
