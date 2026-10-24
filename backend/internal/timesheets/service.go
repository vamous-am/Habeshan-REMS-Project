package timesheets

import (
	"errors"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/habeshan-rems/backend/internal/attendance"
	"github.com/habeshan-rems/backend/internal/auth"
	"github.com/habeshan-rems/backend/internal/common"
	"github.com/habeshan-rems/backend/internal/notifications"
	"gorm.io/gorm"
)

// CalculateTotalHours returns the authoritative total hours for a timesheet period.
// Total hours represent the verified attendance shift duration; task hours provide an
// activity breakdown rather than being added on top.
func CalculateTotalHours(attendanceHours float64) float64 {
	if attendanceHours < 0 {
		return 0
	}
	return attendanceHours
}

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

// GenerateDraftTimesheets creates draft timesheet entries for verified attendance logs in a period.
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

	// Total hours represent the attendance shift; task time is an activity breakdown,
	// not an additive duration component.
	totalHours := CalculateTotalHours(attendanceHours)

	ts := Timesheet{
		UserID:      userID,
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
		TotalHours:  totalHours,
		Status:      "draft",
	}
	ts.OrgID = orgID

	if err := s.db.Create(&ts).Error; err != nil {
		// Treat unique constraint violation as idempotent success if two generation workers raced
		if strings.Contains(strings.ToLower(err.Error()), "unique") || strings.Contains(err.Error(), "idx_timesheets_org_user_period") {
			return nil
		}
		return err
	}
	return nil
}

// Submit transitions a timesheet from draft/rejected → submitted.
// Caller must be the owner of the timesheet.
// Manager/admin submit-on-behalf is prohibited.
func (s *Service) Submit(timesheetID, callerID uuid.UUID) error {
	var ts Timesheet
	if err := s.db.First(&ts, "id = ?", timesheetID).Error; err != nil {
		return errors.New("timesheet not found")
	}

	// Ownership check: only the owning employee may submit or resubmit
	if ts.UserID != callerID {
		return common.ErrForbidden
	}

	if ts.Status != "draft" && ts.Status != "rejected" {
		return errors.New("only draft or rejected timesheets can be submitted")
	}

	if err := s.db.Model(&ts).Updates(map[string]any{
		"status": "submitted",
	}).Error; err != nil {
		return err
	}

	sender := notifications.NewTelegramSender(s.db)
	if err := sender.NotifyTimesheetSubmitted(ts.OrgID, ts.UserID); err != nil {
		log.Printf("⚠️ timesheet %s submitted, but manager notification could not be routed: %v", ts.ID.ID, err)
	}
	return nil
}

// Approve transitions a timesheet from submitted → approved.
// Manager callers are scoped to teams they manage.
// Admin callers are unrestricted within their organization.
func (s *Service) Approve(timesheetID, reviewerID uuid.UUID, reviewerRole auth.Role, orgID uuid.UUID) error {
	var ts Timesheet
	if err := s.db.First(&ts, "id = ?", timesheetID).Error; err != nil {
		return errors.New("timesheet not found")
	}

	if ts.Status != "submitted" {
		return errors.New("only submitted timesheets can be approved")
	}

	// Team-scope check: managers can only approve their own team members
	if reviewerRole == auth.RoleManager {
		if err := s.assertManagerOwnsEmployee(orgID, reviewerID, ts.UserID); err != nil {
			return err
		}
	}

	if err := s.db.Model(&ts).Updates(map[string]any{
		"status":      "approved",
		"reviewed_by": reviewerID,
	}).Error; err != nil {
		return err
	}

	sender := notifications.NewTelegramSender(s.db)
	sender.NotifyTimesheetApproved(ts.OrgID, ts.UserID)
	return nil
}

// Reject transitions a timesheet from submitted → rejected.
// Manager callers are scoped to teams they manage.
// Admin callers are unrestricted within their organization.
func (s *Service) Reject(timesheetID, reviewerID uuid.UUID, reviewerRole auth.Role, orgID uuid.UUID, reason string) error {
	var ts Timesheet
	if err := s.db.First(&ts, "id = ?", timesheetID).Error; err != nil {
		return errors.New("timesheet not found")
	}

	if ts.Status != "submitted" {
		return errors.New("only submitted timesheets can be rejected")
	}

	// Team-scope check: managers can only reject their own team members
	if reviewerRole == auth.RoleManager {
		if err := s.assertManagerOwnsEmployee(orgID, reviewerID, ts.UserID); err != nil {
			return err
		}
	}

	if err := s.db.Model(&ts).Updates(map[string]any{
		"status":           "rejected",
		"reviewed_by":      reviewerID,
		"rejection_reason": reason,
	}).Error; err != nil {
		return err
	}

	sender := notifications.NewTelegramSender(s.db)
	sender.NotifyTimesheetRejected(ts.OrgID, ts.UserID, reason)
	return nil
}

// assertManagerOwnsEmployee verifies that employeeID belongs to at least one
// team managed by managerID within the org.
// Returns ErrForbidden if no such membership exists.
func (s *Service) assertManagerOwnsEmployee(orgID, managerID, employeeID uuid.UUID) error {
	var count int64
	s.db.Table("team_members tm").
		Joins("JOIN teams t ON t.id = tm.team_id").
		Where("tm.user_id = ? AND t.manager_id = ? AND t.org_id = ?", employeeID, managerID, orgID).
		Count(&count)
	if count == 0 {
		return common.ErrForbidden
	}
	return nil
}
