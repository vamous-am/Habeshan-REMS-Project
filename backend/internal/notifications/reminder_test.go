package notifications

import (
	"context"
	"testing"
	"time"
)

func TestCalculateNext930Am(t *testing.T) {
	// Case 1: Early morning before 9:30 AM (08:00 AM Addis Ababa)
	early := time.Date(2026, 10, 24, 8, 0, 0, 0, AddisLocation)
	nextRun, sleep := CalculateNext930Am(early)

	expectedRun := time.Date(2026, 10, 24, 9, 30, 0, 0, AddisLocation)
	if !nextRun.Equal(expectedRun) {
		t.Errorf("expected run %v, got %v", expectedRun, nextRun)
	}
	expectedSleep := 1*time.Hour + 30*time.Minute
	if sleep != expectedSleep {
		t.Errorf("expected sleep %v, got %v", expectedSleep, sleep)
	}

	// Case 2: Late morning after 9:30 AM (10:00 AM Addis Ababa)
	late := time.Date(2026, 10, 24, 10, 0, 0, 0, AddisLocation)
	nextRunLate, sleepLate := CalculateNext930Am(late)

	expectedRunLate := time.Date(2026, 10, 25, 9, 30, 0, 0, AddisLocation) // Tomorrow 9:30 AM
	if !nextRunLate.Equal(expectedRunLate) {
		t.Errorf("expected run %v, got %v", expectedRunLate, nextRunLate)
	}
	expectedSleepLate := 23 * time.Hour + 30 * time.Minute
	if sleepLate != expectedSleepLate {
		t.Errorf("expected sleep %v, got %v", expectedSleepLate, sleepLate)
	}
}

func TestCalculateDayBoundsUTC(t *testing.T) {
	// 2026-10-24 15:00 Addis Ababa
	now := time.Date(2026, 10, 24, 15, 0, 0, 0, AddisLocation)
	startUTC, endUTC := CalculateDayBoundsUTC(now)

	// Midnight Addis Ababa (2026-10-24 00:00:00 +0300) in UTC is 2026-10-23 21:00:00 UTC
	expectedStartUTC := time.Date(2026, 10, 23, 21, 0, 0, 0, time.UTC)
	expectedEndUTC := time.Date(2026, 10, 24, 21, 0, 0, 0, time.UTC)

	if !startUTC.Equal(expectedStartUTC) {
		t.Errorf("expected startUTC %v, got %v", expectedStartUTC, startUTC)
	}
	if !endUTC.Equal(expectedEndUTC) {
		t.Errorf("expected endUTC %v, got %v", expectedEndUTC, endUTC)
	}
	if endUTC.Sub(startUTC) != 24*time.Hour {
		t.Errorf("day bounds must span exactly 24 hours")
	}
}

func TestStartClockInReminderScheduler_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		StartClockInReminderScheduler(ctx, nil)
		close(done)
	}()

	// Cancel context immediately
	cancel()

	select {
	case <-done:
		// Successfully launched and terminated
	case <-time.After(2 * time.Second):
		t.Fatal("StartClockInReminderScheduler did not stop promptly on cancellation")
	}
}
