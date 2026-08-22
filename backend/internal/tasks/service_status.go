package tasks

// service_status.go — FR-TASK-04 (status transitions)
//
// The authoritative transition graph lives here and nowhere else.
// To change the rules, edit allowedTransitions — no other file needs updating.

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// allowedTransitions maps each status to the statuses it may transition into.
//
// Current policy (conservative default — update when authoritative rules arrive):
//
//	to_do       → in_progress, blocked
//	in_progress → paused, blocked, completed
//	paused      → in_progress, blocked
//	blocked     → in_progress, to_do
//	completed   → (terminal)
var allowedTransitions = map[Status][]Status{
	StatusToDo:       {StatusInProgress, StatusBlocked},
	StatusInProgress: {StatusPaused, StatusBlocked, StatusCompleted},
	StatusPaused:     {StatusInProgress, StatusBlocked},
	StatusBlocked:    {StatusInProgress, StatusToDo},
	StatusCompleted:  {},
}

// isValidTransition returns true when from→to is in the transition table.
func isValidTransition(from, to Status) bool {
	for _, allowed := range allowedTransitions[from] {
		if allowed == to {
			return true
		}
	}
	return false
}

// UpdateTaskStatus — FR-TASK-04
func (s *taskService) UpdateTaskStatus(taskID uuid.UUID, newStatus Status, callerID, orgID uuid.UUID) error {
	if !newStatus.IsValid() {
		return fmt.Errorf("%w: unknown status %q", ErrBadRequest, newStatus)
	}

	caller, err := s.resolveUser(callerID)
	if err != nil {
		return err
	}
	if err := requireOrgMatch(caller, orgID); err != nil {
		return err
	}

	task, err := s.taskRepo.GetTaskByIDForOrg(taskID, orgID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		return ErrInternal
	}

	if err := s.canSeeTask(caller, task, orgID); err != nil {
		return err
	}
	if task.Status == newStatus {
		return nil // no-op
	}
	if !isValidTransition(task.Status, newStatus) {
		return fmt.Errorf("%w: cannot transition from %q to %q", ErrBadRequest, task.Status, newStatus)
	}

	if err := s.taskRepo.UpdateTaskStatus(taskID, newStatus); err != nil {
		return ErrInternal
	}
	return nil
}
