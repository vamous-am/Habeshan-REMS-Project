package tasks

// service_tasks.go — FR-TASK-01 (create), FR-TASK-03 (visibility, detail)

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CreateTask — FR-TASK-01
// Only admins and managers may create tasks.
func (s *taskService) CreateTask(req CreateTaskRequest, callerID uuid.UUID) (Task, error) {
	caller, err := s.resolveUser(callerID)
	if err != nil {
		return Task{}, err
	}
	if err := requireOrgMatch(caller, req.OrgID); err != nil {
		return Task{}, err
	}
	if err := requireRole(caller, RoleAdmin, RoleManager); err != nil {
		return Task{}, fmt.Errorf("%w: only admins and managers can create tasks", err)
	}

	// Confirm the organisation exists — prevents tasks being created under a
	// deleted or non-existent org even if the caller's user row still has that
	// org_id (e.g. soft-deleted org, stale token).
	if _, err := s.orgRepo.GetOrganizationByID(req.OrgID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return Task{}, fmt.Errorf("%w: organisation not found", ErrNotFound)
		}
		return Task{}, ErrInternal
	}

	if req.Title == "" {
		return Task{}, fmt.Errorf("%w: title is required", ErrBadRequest)
	}

	priority := req.Priority
	if priority == "" {
		priority = PriorityMedium
	} else if !priority.IsValid() {
		return Task{}, fmt.Errorf("%w: invalid priority", ErrBadRequest)
	}

	task := Task{
		Title:     req.Title,
		CreatedBy: callerID,
		Priority:  priority,
		Status:    StatusToDo,
		DueDate:   req.DueDate,
	}
	task.OrgID = req.OrgID
	if req.Description != "" {
		task.Description = &req.Description
	}

	if err := s.taskRepo.CreateTask(&task); err != nil {
		return Task{}, ErrInternal
	}
	return task, nil
}

// GetMyTasks — FR-TASK-03
// Returns the visibility-scoped task list based on the caller's role.
func (s *taskService) GetMyTasks(callerID, orgID uuid.UUID) ([]Task, error) {
	caller, err := s.resolveUser(callerID)
	if err != nil {
		return nil, err
	}
	if err := requireOrgMatch(caller, orgID); err != nil {
		return nil, err
	}

	switch caller.Role {
	case RoleAdmin:
		return s.taskRepo.GetTasksByOrgID(orgID)

	case RoleManager:
		return s.managerTasks(callerID, orgID)

	case RoleEmployee:
		return s.taskRepo.GetTasksByUserID(callerID)
	}

	return nil, ErrForbidden
}

// managerTasks merges team-assigned tasks with tasks the manager created,
// deduplicating by ID so a task assigned to a team member and created by the
// manager doesn't appear twice.
func (s *taskService) managerTasks(managerID, orgID uuid.UUID) ([]Task, error) {
	memberIDs, err := s.userRepo.GetTeamMembersForManager(managerID, orgID)
	if err != nil {
		return nil, ErrInternal
	}

	var tasks []Task
	if len(memberIDs) > 0 {
		tasks, err = s.taskRepo.GetTasksByAssignedUsers(orgID, memberIDs)
		if err != nil {
			return nil, ErrInternal
		}
	}

	created, err := s.taskRepo.GetTasksByCreator(managerID, orgID)
	if err != nil {
		return nil, ErrInternal
	}

	seen := make(map[uuid.UUID]struct{}, len(tasks))
	for _, t := range tasks {
		seen[t.ID.ID] = struct{}{}
	}
	for _, t := range created {
		if _, exists := seen[t.ID.ID]; !exists {
			tasks = append(tasks, t)
		}
	}
	return tasks, nil
}

// GetTaskByID returns a single task after a visibility check.
func (s *taskService) GetTaskByID(taskID uuid.UUID, callerID, orgID uuid.UUID) (Task, error) {
	caller, err := s.resolveUser(callerID)
	if err != nil {
		return Task{}, err
	}
	task, err := s.taskRepo.GetTaskByIDForOrg(taskID, orgID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return Task{}, ErrNotFound
		}
		return Task{}, ErrInternal
	}
	if err := s.canSeeTask(caller, task, orgID); err != nil {
		return Task{}, err
	}
	return task, nil
}
