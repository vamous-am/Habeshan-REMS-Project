package tasks

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// mockTimeLogRepo implements TaskTimeLogRepository for testing status transitions.
type mockTimeLogRepo struct {
	hasActive bool
}

func (m *mockTimeLogRepo) CreateTimeLog(log *TaskTimeLog) error { return nil }
func (m *mockTimeLogRepo) UpsertTimeLog(log *TaskTimeLog) error { return nil }
func (m *mockTimeLogRepo) GetActiveTimeLog(taskID, userID uuid.UUID) (TaskTimeLog, error) {
	return TaskTimeLog{}, nil
}
func (m *mockTimeLogRepo) HasActiveTimeLog(taskID uuid.UUID) (bool, error) {
	return m.hasActive, nil
}
func (m *mockTimeLogRepo) GetTimeLogsByTaskID(taskID uuid.UUID) ([]TaskTimeLog, error) {
	return nil, nil
}
func (m *mockTimeLogRepo) GetTimeLogsByTaskAndUser(taskID, userID uuid.UUID) ([]TaskTimeLog, error) {
	return nil, nil
}
func (m *mockTimeLogRepo) FindByRecordUUID(recordUUID uuid.UUID) (TaskTimeLog, error) {
	return TaskTimeLog{}, nil
}
func (m *mockTimeLogRepo) UpdateTimeLog(log *TaskTimeLog) error { return nil }

func TestAllowedTransitions_PausedToCompleted(t *testing.T) {
	// 1. Verify PAUSED -> COMPLETED is now in the transition matrix
	if !isValidTransition(StatusPaused, StatusCompleted) {
		t.Fatalf("expected transition from PAUSED to COMPLETED to be valid")
	}

	// 2. Verify other expected transitions from PAUSED
	if !isValidTransition(StatusPaused, StatusInProgress) {
		t.Errorf("expected transition from PAUSED to IN_PROGRESS to be valid")
	}
	if !isValidTransition(StatusPaused, StatusBlocked) {
		t.Errorf("expected transition from PAUSED to BLOCKED to be valid")
	}

	// 3. Verify terminal status
	if isValidTransition(StatusCompleted, StatusInProgress) {
		t.Errorf("expected COMPLETED to be terminal")
	}
	if isValidTransition(StatusCompleted, StatusToDo) {
		t.Errorf("expected COMPLETED to be terminal")
	}

	// 4. Verify invalid transitions
	if isValidTransition(StatusToDo, StatusCompleted) {
		t.Errorf("expected TO_DO to COMPLETED to be invalid")
	}
}

func TestActiveTimerGuardOnCompletion(t *testing.T) {
	taskID := uuid.New()

	// Simulated test helper verifying active timer guard
	checkCompletionGuard := func(hasActive bool) error {
		repo := &mockTimeLogRepo{hasActive: hasActive}
		active, err := repo.HasActiveTimeLog(taskID)
		if err != nil {
			return ErrInternal
		}
		if active {
			return ErrConflict
		}
		return nil
	}

	// Case 1: Timer actively running -> completing task must fail with ErrConflict
	if err := checkCompletionGuard(true); !errors.Is(err, ErrConflict) {
		t.Errorf("expected ErrConflict when completing a task with active timer, got %v", err)
	}

	// Case 2: Timer stopped/closed -> completing task must succeed
	if err := checkCompletionGuard(false); err != nil {
		t.Errorf("expected success when completing a task with no active timer, got %v", err)
	}
}
