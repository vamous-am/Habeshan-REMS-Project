package timesheets

import (
	"errors"
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

// GenerateDraftTimesheets — FR-TS-01
func (s *Service) GenerateDraftTimesheets(orgID uuid.UUID, periodStart, periodEnd time.Time) error {
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
			continue
		}
	}
	return nil
}

func (s *Service) generateForUser(orgID, userID uuid.UUID, periodStart, periodEnd time.Time) error {
	var existing Timesheet
	result := s.db.Where(
		"org_id = ? AND user_id = ? AND period_start = ? AND period_end = ?",
		orgID, userID, periodStart, periodEnd,
	).First(&existing)
	if result.Error == nil {
		return nil
	}

	var attendanceHours float64
	s.db.Model(&attendance.AttendanceLog{}).
		Where("org_id = ? AND user_id = ? AND clock_in >= ? AND clock_in < ?", orgID, userID, periodStart, periodEnd).
		Where("sync_status = ? AND total_hours IS NOT NULL", attendance.SyncStatusSyncedVerified).
		Select("COALESCE(SUM(total_hours), 0)").
		Scan(&attendanceHours)

	var taskMinutes int
	s.db.Model(&tasks.TaskTimeLog{}).
		Where("user_id = ? AND started_at >= ? AND started_at < ?", userID, periodStart, periodEnd).
		Where("sync_status = ? AND duration_minutes IS NOT NULL", tasks.SyncSyncedVerified).
		Select("COALESCE(SUM(duration_minutes), 0)").
		Scan(&taskMinutes)

	totalHours := attendanceHours + float64(taskMinutes)/60.0

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

// Submit transitions a timesheet from draft → submitted — FR-TS-02/03
func (s *Service) Submit(timesheetID uuid.UUID) error {
	var ts Timesheet
	if err := s.db.First(&ts, "id = ?", timesheetID).Error; err != nil {
		return errors.New("timesheet not found")
	}

	if ts.Status != "draft" && ts.Status != "rejected" {
		return errors.New("only draft or rejected timesheets can be submitted")
	}

	return s.db.Model(&ts).Updates(map[string]any{
		"status": "submitted",
	}).Error
}

// Approve transitions a timesheet from submitted → approved — FR-TS-04/05
func (s *Service) Approve(timesheetID, reviewerID uuid.UUID) error {
	var ts Timesheet
	if err := s.db.First(&ts, "id = ?", timesheetID).Error; err != nil {
		return errors.New("timesheet not found")
	}

	if ts.Status != "submitted" {
		return errors.New("only submitted timesheets can be approved")
	}

	return s.db.Model(&ts).Updates(map[string]any{
		"status":      "approved",
		"reviewed_by": reviewerID,
	}).Error
}

// Reject transitions a timesheet from submitted → rejected — FR-TS-04/06
// Rejection reason is mandatory.
func (s *Service) Reject(timesheetID, reviewerID uuid.UUID, reason string) error {
	var ts Timesheet
	if err := s.db.First(&ts, "id = ?", timesheetID).Error; err != nil {
		return errors.New("timesheet not found")
	}

	if ts.Status != "submitted" {
		return errors.New("only submitted timesheets can be rejected")
	}

	return s.db.Model(&ts).Updates(map[string]any{
		"status":           "rejected",
		"reviewed_by":      reviewerID,
		"rejection_reason": reason,
	}).Error
}
