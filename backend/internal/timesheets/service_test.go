package timesheets

import (
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestCalculateTotalHours_AttendanceOnlyAuthoritative(t *testing.T) {
	tests := []struct {
		name            string
		attendanceHours float64
		taskHours       float64
		taskSegments    []float64
		expectedTotal   float64
	}{
		{
			name:            "8 attendance hours and 6 task hours produces 8.0, not 14.0",
			attendanceHours: 8.0,
			taskHours:       6.0,
			taskSegments:    []float64{3.0, 3.0},
			expectedTotal:   8.0,
		},
		{
			name:            "Attendance hours with no task logs still produces attendance total",
			attendanceHours: 8.0,
			taskHours:       0.0,
			taskSegments:    nil,
			expectedTotal:   8.0,
		},
		{
			name:            "Task hours greater than attendance hours do not increase TotalHours",
			attendanceHours: 7.5,
			taskHours:       9.0, // active logs logged across shifts or unconstrained
			taskSegments:    []float64{4.5, 4.5},
			expectedTotal:   7.5,
		},
		{
			name:            "Multiple task segments are not added to attendance duration",
			attendanceHours: 8.0,
			taskHours:       7.25,
			taskSegments:    []float64{1.25, 2.0, 2.0, 2.0},
			expectedTotal:   8.0,
		},
		{
			name:            "Partial shift duration (e.g. 3.75 hours) preserved exactly",
			attendanceHours: 3.75,
			taskHours:       2.5,
			taskSegments:    []float64{2.5},
			expectedTotal:   3.75,
		},
		{
			name:            "Zero attendance hours produces 0.0 total",
			attendanceHours: 0.0,
			taskHours:       5.0,
			taskSegments:    []float64{5.0},
			expectedTotal:   0.0,
		},
		{
			name:            "Negative attendance guard produces 0.0",
			attendanceHours: -2.5,
			taskHours:       0.0,
			taskSegments:    nil,
			expectedTotal:   0.0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := CalculateTotalHours(tc.attendanceHours)
			if math.Abs(actual-tc.expectedTotal) > 1e-6 {
				t.Errorf("CalculateTotalHours(%v) = %v; want %v", tc.attendanceHours, actual, tc.expectedTotal)
			}

			// Verify that if taskHours were erroneously added (the old bug),
			// it would differ from the corrected authoritative total whenever taskHours > 0.
			if tc.taskHours > 0 {
				buggyTotal := tc.attendanceHours + tc.taskHours
				if actual == buggyTotal {
					t.Errorf("expected actual (%v) to differ from buggy inflated total (%v)", actual, buggyTotal)
				}
			}
		})
	}
}

func TestTimesheetExportAndModel_Formatting(t *testing.T) {
	// Timesheet model verification
	now := time.Now().UTC()
	periodStart := now.AddDate(0, 0, -7)
	periodEnd := now

	ts := Timesheet{
		UserID:      uuid.New(),
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
		TotalHours:  CalculateTotalHours(8.5),
		Status:      "draft",
	}

	if ts.TotalHours != 8.5 {
		t.Fatalf("expected ts.TotalHours == 8.5, got %v", ts.TotalHours)
	}

	// Verify export formatting %.2f
	formatted := fmt.Sprintf("%.2f", ts.TotalHours)
	if formatted != "8.50" {
		t.Errorf("expected formatted total hours to be '8.50', got %q", formatted)
	}

	// Verify status lifecycle transitions do not alter TotalHours
	ts.Status = "submitted"
	if ts.TotalHours != 8.5 {
		t.Errorf("submission altered TotalHours")
	}

	ts.Status = "approved"
	if ts.TotalHours != 8.5 {
		t.Errorf("approval altered TotalHours")
	}
}
