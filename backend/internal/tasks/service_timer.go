package tasks

// service_timer.go — FR-TASK-05/06/07/08/10 (timer operations + history)
//
// Timer segment model:
//   Each row in task_time_logs represents ONE open or closed segment.
//   - START  → INSERT a new row (stopped_at = NULL).
//   - PAUSE  → UPDATE the open row: set stopped_at, duration_minutes, pause_reason.
//              record_uuid on the row is NEVER mutated — it is the immutable
//              identity of that segment's creation event.
//              Idempotency for a repeated pause call is handled by checking
//              whether stopped_at is already populated on the active row.
//   - RESUME → INSERT a new open row (same as start).
//   - STOP   → UPDATE the open row: set stopped_at, duration_minutes only.
//              Same immutability rule as pause.

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// StartTimer — FR-TASK-05 / FR-TASK-10
// Inserts a new open timer segment. Idempotent on record_uuid.
func (s *taskService) StartTimer(req TimerStartRequest, callerID, orgID uuid.UUID) (TaskTimeLog, error) {
	caller, err := s.resolveUser(callerID)
	if err != nil {
		return TaskTimeLog{}, err
	}
	if err := requireOrgMatch(caller, orgID); err != nil {
		return TaskTimeLog{}, err
	}

	task, err := s.taskRepo.GetTaskByIDForOrg(req.TaskID, orgID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return TaskTimeLog{}, ErrNotFound
		}
		return TaskTimeLog{}, ErrInternal
	}

	// Employees must be assigned; managers/admins may start without assignment.
	assigned, err := s.assignRepo.IsTaskAssignedToUser(task.ID.ID, callerID)
	if err != nil {
		return TaskTimeLog{}, ErrInternal
	}
	if !assigned && caller.Role == RoleEmployee {
		return TaskTimeLog{}, fmt.Errorf("%w: you are not assigned to this task", ErrForbidden)
	}

	// Idempotency: same record_uuid means this start was already processed.
	if existing, hit, err := s.idempotentLookup(req.RecordUUID); err != nil {
		return TaskTimeLog{}, err
	} else if hit {
		return existing, nil
	}

	// Reject if there is already an open segment for this user on this task.
	_, activeErr := s.timeLogRepo.GetActiveTimeLog(task.ID.ID, callerID)
	if activeErr == nil {
		return TaskTimeLog{}, fmt.Errorf("%w: timer is already running for this task", ErrConflict)
	}
	if !errors.Is(activeErr, gorm.ErrRecordNotFound) {
		return TaskTimeLog{}, ErrInternal
	}

	entry := TaskTimeLog{
		TaskID:     task.ID.ID,
		UserID:     callerID,
		StartedAt:  req.StartedAt,
		SyncStatus: req.SyncStatus,
		DeviceHash: req.DeviceHash,
		RecordUUID: req.RecordUUID, // immutable after insert
	}

	if err := s.timeLogRepo.CreateTimeLog(&entry); err != nil {
		return TaskTimeLog{}, ErrInternal
	}
	if task.Status == StatusToDo {
		_ = s.taskRepo.UpdateTaskStatus(task.ID.ID, StatusInProgress)
	}
	return entry, nil
}

// PauseTimer — FR-TASK-05 / FR-TASK-06 / FR-TASK-10
// Closes the open segment by updating stopped_at, duration_minutes, and
// pause_reason.  record_uuid on the existing row is never touched.
//
// Idempotency: if stopped_at is already set on the active row this pause was
// already committed — return the closed row without error.
func (s *taskService) PauseTimer(req TimerPauseRequest, callerID, orgID uuid.UUID) (TaskTimeLog, error) {
	caller, err := s.resolveUser(callerID)
	if err != nil {
		return TaskTimeLog{}, err
	}
	if err := requireOrgMatch(caller, orgID); err != nil {
		return TaskTimeLog{}, err
	}
	if !req.PauseReason.IsValid() {
		return TaskTimeLog{}, fmt.Errorf("%w: a valid pause reason is required", ErrBadRequest)
	}

	if _, err := s.taskRepo.GetTaskByIDForOrg(req.TaskID, orgID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return TaskTimeLog{}, ErrNotFound
		}
		return TaskTimeLog{}, ErrInternal
	}

	active, err := s.timeLogRepo.GetActiveTimeLog(req.TaskID, callerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return TaskTimeLog{}, fmt.Errorf("%w: no running timer to pause", ErrConflict)
		}
		return TaskTimeLog{}, ErrInternal
	}

	// Idempotency: segment already closed — this pause was already applied.
	if active.StoppedAt != nil {
		return active, nil
	}

	reason := string(req.PauseReason)
	now := req.PausedAt
	// Only update the three mutable close-fields. record_uuid stays unchanged.
	active.StoppedAt = &now
	active.DurationMinutes = durationMinutes(active.StartedAt, now)
	active.PauseReason = &reason

	if err := s.timeLogRepo.UpdateTimeLog(&active); err != nil {
		return TaskTimeLog{}, ErrInternal
	}
	_ = s.taskRepo.UpdateTaskStatus(req.TaskID, StatusPaused)
	return active, nil
}

// ResumeTimer — FR-TASK-05 / FR-TASK-10
// Inserts a new open segment after a pause. Idempotent on record_uuid.
func (s *taskService) ResumeTimer(req TimerResumeRequest, callerID, orgID uuid.UUID) (TaskTimeLog, error) {
	caller, err := s.resolveUser(callerID)
	if err != nil {
		return TaskTimeLog{}, err
	}
	if err := requireOrgMatch(caller, orgID); err != nil {
		return TaskTimeLog{}, err
	}

	// Idempotency: same record_uuid means this resume was already processed.
	if existing, hit, err := s.idempotentLookup(req.RecordUUID); err != nil {
		return TaskTimeLog{}, err
	} else if hit {
		return existing, nil
	}

	task, err := s.taskRepo.GetTaskByIDForOrg(req.TaskID, orgID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return TaskTimeLog{}, ErrNotFound
		}
		return TaskTimeLog{}, ErrInternal
	}
	if task.Status != StatusPaused {
		return TaskTimeLog{}, fmt.Errorf("%w: task is not paused", ErrConflict)
	}

	// There must be no open segment; the previous one was closed by PauseTimer.
	_, activeErr := s.timeLogRepo.GetActiveTimeLog(task.ID.ID, callerID)
	if activeErr == nil {
		return TaskTimeLog{}, fmt.Errorf("%w: a timer segment is already open", ErrConflict)
	}
	if !errors.Is(activeErr, gorm.ErrRecordNotFound) {
		return TaskTimeLog{}, ErrInternal
	}

	entry := TaskTimeLog{
		TaskID:     task.ID.ID,
		UserID:     callerID,
		StartedAt:  req.ResumedAt,
		SyncStatus: req.SyncStatus,
		DeviceHash: req.DeviceHash,
		RecordUUID: req.RecordUUID, // immutable after insert
	}

	if err := s.timeLogRepo.CreateTimeLog(&entry); err != nil {
		return TaskTimeLog{}, ErrInternal
	}
	_ = s.taskRepo.UpdateTaskStatus(task.ID.ID, StatusInProgress)
	return entry, nil
}

// StopTimer — FR-TASK-05 / FR-TASK-10
// Closes the open segment by updating stopped_at and duration_minutes only.
// record_uuid on the existing row is never touched.
//
// Idempotency: if stopped_at is already set the stop was already committed —
// return the closed row without error.
func (s *taskService) StopTimer(req TimerStopRequest, callerID, orgID uuid.UUID) (TaskTimeLog, error) {
	caller, err := s.resolveUser(callerID)
	if err != nil {
		return TaskTimeLog{}, err
	}
	if err := requireOrgMatch(caller, orgID); err != nil {
		return TaskTimeLog{}, err
	}

	if _, err := s.taskRepo.GetTaskByIDForOrg(req.TaskID, orgID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return TaskTimeLog{}, ErrNotFound
		}
		return TaskTimeLog{}, ErrInternal
	}

	active, err := s.timeLogRepo.GetActiveTimeLog(req.TaskID, callerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return TaskTimeLog{}, fmt.Errorf("%w: no running timer to stop", ErrConflict)
		}
		return TaskTimeLog{}, ErrInternal
	}

	// Idempotency: segment already closed — this stop was already applied.
	if active.StoppedAt != nil {
		return active, nil
	}

	now := req.StoppedAt
	// Only update the two mutable close-fields. record_uuid stays unchanged.
	active.StoppedAt = &now
	active.DurationMinutes = durationMinutes(active.StartedAt, now)

	if err := s.timeLogRepo.UpdateTimeLog(&active); err != nil {
		return TaskTimeLog{}, ErrInternal
	}
	return active, nil
}

// GetTimerHistory — FR-TASK-07 / FR-TASK-08
// Returns the full audit log for a task in chronological order.
// Pause reasons are redacted for peer employees (FR-TASK-07).
func (s *taskService) GetTimerHistory(taskID uuid.UUID, callerID, orgID uuid.UUID) ([]TaskTimeLog, error) {
	caller, err := s.resolveUser(callerID)
	if err != nil {
		return nil, err
	}

	task, err := s.taskRepo.GetTaskByIDForOrg(taskID, orgID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, ErrInternal
	}
	if err := s.canSeeTask(caller, task, orgID); err != nil {
		return nil, err
	}

	logs, err := s.timeLogRepo.GetTimeLogsByTaskID(taskID)
	if err != nil {
		return nil, ErrInternal
	}

	// Redact pause reasons from other employees' rows (FR-TASK-07).
	if caller.Role != RoleAdmin && caller.Role != RoleManager {
		for i := range logs {
			if logs[i].UserID != callerID {
				logs[i].PauseReason = nil
			}
		}
	}
	return logs, nil
}
