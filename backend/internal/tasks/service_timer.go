package tasks

// service_timer.go — FR-TASK-05/06/07/08/10 (timer operations + history)

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// StartTimer — FR-TASK-05 / FR-TASK-10
// Opens a new timer segment. Idempotent on record_uuid.
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

	// Only assigned employees (or managers/admins) can start a timer.
	assigned, err := s.assignRepo.IsTaskAssignedToUser(task.ID.ID, callerID)
	if err != nil {
		return TaskTimeLog{}, ErrInternal
	}
	if !assigned && caller.Role == RoleEmployee {
		return TaskTimeLog{}, fmt.Errorf("%w: you are not assigned to this task", ErrForbidden)
	}

	if existing, hit, err := s.idempotentLookup(req.RecordUUID); err != nil {
		return TaskTimeLog{}, err
	} else if hit {
		return existing, nil
	}

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
		RecordUUID: req.RecordUUID,
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
// Closes the active segment and stamps the mandatory pause reason.
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

	if existing, hit, err := s.idempotentLookup(req.RecordUUID); err != nil {
		return TaskTimeLog{}, err
	} else if hit {
		return existing, nil
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

	reason := string(req.PauseReason)
	now := req.PausedAt
	active.StoppedAt = &now
	active.DurationMinutes = durationMinutes(active.StartedAt, now)
	active.PauseReason = &reason
	active.RecordUUID = req.RecordUUID
	active.SyncStatus = req.SyncStatus
	active.DeviceHash = req.DeviceHash

	if err := s.timeLogRepo.UpdateTimeLog(&active); err != nil {
		return TaskTimeLog{}, ErrInternal
	}
	_ = s.taskRepo.UpdateTaskStatus(req.TaskID, StatusPaused)
	return active, nil
}

// ResumeTimer — FR-TASK-05 / FR-TASK-10
// Opens a new segment after a pause.
func (s *taskService) ResumeTimer(req TimerResumeRequest, callerID, orgID uuid.UUID) (TaskTimeLog, error) {
	caller, err := s.resolveUser(callerID)
	if err != nil {
		return TaskTimeLog{}, err
	}
	if err := requireOrgMatch(caller, orgID); err != nil {
		return TaskTimeLog{}, err
	}

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
		RecordUUID: req.RecordUUID,
	}

	if err := s.timeLogRepo.CreateTimeLog(&entry); err != nil {
		return TaskTimeLog{}, ErrInternal
	}
	_ = s.taskRepo.UpdateTaskStatus(task.ID.ID, StatusInProgress)
	return entry, nil
}

// StopTimer — FR-TASK-05 / FR-TASK-10
// Closes the active segment with no pause reason.
func (s *taskService) StopTimer(req TimerStopRequest, callerID, orgID uuid.UUID) (TaskTimeLog, error) {
	caller, err := s.resolveUser(callerID)
	if err != nil {
		return TaskTimeLog{}, err
	}
	if err := requireOrgMatch(caller, orgID); err != nil {
		return TaskTimeLog{}, err
	}

	if existing, hit, err := s.idempotentLookup(req.RecordUUID); err != nil {
		return TaskTimeLog{}, err
	} else if hit {
		return existing, nil
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

	now := req.StoppedAt
	active.StoppedAt = &now
	active.DurationMinutes = durationMinutes(active.StartedAt, now)
	active.RecordUUID = req.RecordUUID
	active.SyncStatus = req.SyncStatus
	active.DeviceHash = req.DeviceHash

	if err := s.timeLogRepo.UpdateTimeLog(&active); err != nil {
		return TaskTimeLog{}, ErrInternal
	}
	return active, nil
}

// GetTimerHistory — FR-TASK-07 / FR-TASK-08
// Returns the full audit log for a task.
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
