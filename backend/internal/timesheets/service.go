package timesheets

import (
	"time"

	"github.com/google/uuid"
	"github.com/habeshan-rems/backend/internal/attendance"
	"github.com/habeshan-rems/backend/internal/tasks"
	"gorm.io/gorm"
)

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

// GenerateDraftTimesheets is called by the scheduler once per period.
// For each user in the org, it:
//  1. Pulls all attendance_logs for the period
//  2. Pulls all task_time_logs for the period
//  3. Sums total hours from both
//  4. Creates a draft Timesheet row if one doesn't already exist
//
// FR-TS-01
func (s *Service) GenerateDraftTimesheets(orgID uuid.UUID, periodStart, periodEnd time.Time) error {
	// 1. Find all unique user IDs that have attendance logs in this period
	var userIDs []uuid.UUID
	if err := s.db.
		Model(&attendance.AttendanceLog{}).
		Where("org_id = ? AND clock_in >= ? AND clock_in < ?", orgID, periodStart, periodEnd).
		Where("sync_status = ?", attendance.SyncStatusSyncedVerified).
		Distinct("user_id").
		Pluck("user_id", &userIDs).Error; err != nil {
		return err
	}

	for _, userID := range userIDs {
		if err := s.generateForUser(orgID, userID, periodStart, periodEnd); err != nil {
			// Log and continue — one user failing shouldn't block others
			continue
		}
	}
	return nil
}

func (s *Service) generateForUser(orgID, userID uuid.UUID, periodStart, periodEnd time.Time) error {
	// Skip if a timesheet already exists for this user+period
	var existing Timesheet
	result := s.db.Where(
		"org_id = ? AND user_id = ? AND period_start = ? AND period_end = ?",
		orgID, userID, periodStart, periodEnd,
	).First(&existing)
	if result.Error == nil {
		return nil // already exists, skip
	}

	// Sum attendance hours for the period
	var attendanceHours float64
	s.db.Model(&attendance.AttendanceLog{}).
		Where("org_id = ? AND user_id = ? AND clock_in >= ? AND clock_in < ?", orgID, userID, periodStart, periodEnd).
		Where("sync_status = ? AND total_hours IS NOT NULL", attendance.SyncStatusSyncedVerified).
		Select("COALESCE(SUM(total_hours), 0)").
		Scan(&attendanceHours)

	// Sum task timer hours for the period (duration_minutes → hours)
	var taskMinutes int
	s.db.Model(&tasks.TaskTimeLog{}).
		Where("user_id = ? AND started_at >= ? AND started_at < ?", userID, periodStart, periodEnd).
		Where("sync_status = ? AND duration_minutes IS NOT NULL", tasks.SyncSyncedVerified).
		Select("COALESCE(SUM(duration_minutes), 0)").
		Scan(&taskMinutes)

	taskHours := float64(taskMinutes) / 60.0
	totalHours := attendanceHours + taskHours

	// Create the draft timesheet
	ts := Timesheet{
		UserID:      userID,
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
		TotalHours:  totalHours,
		Status:      "draft",
	}
	ts.OrgID = orgID

	return s.db.Create(&ts).Error
}
