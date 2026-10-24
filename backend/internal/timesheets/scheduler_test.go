package timesheets

import (
	"context"
	"testing"
	"time"
)

func TestCalculateWeeklyPeriod(t *testing.T) {
	// Case 1: Wednesday midday (2026-10-21 14:30:00 in Addis Ababa)
	wednesday := time.Date(2026, 10, 21, 14, 30, 0, 0, AddisLocation)
	start, end, sleep := CalculateWeeklyPeriod(wednesday)

	expectedStart := time.Date(2026, 10, 19, 0, 0, 0, 0, AddisLocation) // Monday
	expectedEnd := time.Date(2026, 10, 26, 0, 0, 0, 0, AddisLocation)   // Next Monday

	if !start.Equal(expectedStart) {
		t.Errorf("expected start %v, got %v", expectedStart, start)
	}
	if !end.Equal(expectedEnd) {
		t.Errorf("expected end %v, got %v", expectedEnd, end)
	}
	if sleep != expectedEnd.Sub(wednesday) {
		t.Errorf("expected sleep %v, got %v", expectedEnd.Sub(wednesday), sleep)
	}

	// Case 2: Sunday night (2026-10-25 23:59:00 in Addis Ababa)
	sunday := time.Date(2026, 10, 25, 23, 59, 0, 0, AddisLocation)
	startSun, endSun, sleepSun := CalculateWeeklyPeriod(sunday)

	if !startSun.Equal(expectedStart) {
		t.Errorf("Sunday must belong to the current week starting %v, got %v", expectedStart, startSun)
	}
	if !endSun.Equal(expectedEnd) {
		t.Errorf("Sunday must end on next Monday %v, got %v", expectedEnd, endSun)
	}
	if sleepSun != 1*time.Minute {
		t.Errorf("expected sleep 1m, got %v", sleepSun)
	}

	// Case 3: Monday morning (start of week)
	monday := time.Date(2026, 10, 19, 0, 0, 0, 0, AddisLocation)
	startMon, endMon, sleepMon := CalculateWeeklyPeriod(monday)

	if !startMon.Equal(expectedStart) || !endMon.Equal(expectedEnd) {
		t.Errorf("Monday start mismatch")
	}
	if sleepMon != 7*24*time.Hour {
		t.Errorf("expected full week sleep duration, got %v", sleepMon)
	}
}

func TestStartScheduler_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		// Run with nil db; cancellation should exit before any db queries
		StartScheduler(ctx, nil)
		close(done)
	}()

	// Cancel context immediately
	cancel()

	select {
	case <-done:
		// Succeeded in launching and terminating gracefully
	case <-time.After(2 * time.Second):
		t.Fatal("StartScheduler did not stop promptly on context cancellation")
	}
}
