package tasks

// service_aggregates.go — FR-TASK-09 (status counts, overdue tasks)

import (
	"time"

	"github.com/google/uuid"
)

// GetTaskStatusCounts — FR-TASK-09
// Admins get org-wide counts; managers get counts scoped to their team members.
func (s *taskService) GetTaskStatusCounts(callerID, orgID uuid.UUID) (map[Status]int64, error) {
	caller, err := s.resolveUser(callerID)
	if err != nil {
		return nil, err
	}
	if err := requireOrgMatch(caller, orgID); err != nil {
		return nil, err
	}
	if err := requireRole(caller, RoleAdmin, RoleManager); err != nil {
		return nil, err
	}

	if caller.Role == RoleAdmin {
		return s.taskRepo.GetTaskStatusCounts(orgID)
	}

	memberIDs, err := s.userRepo.GetTeamMembersForManager(callerID, orgID)
	if err != nil {
		return nil, ErrInternal
	}
	return s.taskRepo.GetTaskStatusCountsForManager(orgID, memberIDs)
}

// GetOverdueTasks — FR-TASK-09
// Admins get all overdue tasks in the org; managers get overdue tasks assigned
// to their team members.
func (s *taskService) GetOverdueTasks(callerID, orgID uuid.UUID) ([]Task, error) {
	caller, err := s.resolveUser(callerID)
	if err != nil {
		return nil, err
	}
	if err := requireOrgMatch(caller, orgID); err != nil {
		return nil, err
	}
	if err := requireRole(caller, RoleAdmin, RoleManager); err != nil {
		return nil, err
	}

	if caller.Role == RoleAdmin {
		return s.taskRepo.GetOverdueTasks(orgID)
	}

	memberIDs, err := s.userRepo.GetTeamMembersForManager(callerID, orgID)
	if err != nil {
		return nil, ErrInternal
	}

	tasks, err := s.taskRepo.GetTasksByAssignedUsers(orgID, memberIDs)
	if err != nil {
		return nil, ErrInternal
	}

	now := time.Now()
	overdue := make([]Task, 0, len(tasks))
	for _, t := range tasks {
		if t.Status != StatusCompleted && t.DueDate != nil && t.DueDate.Before(now) {
			overdue = append(overdue, t)
		}
	}
	return overdue, nil
