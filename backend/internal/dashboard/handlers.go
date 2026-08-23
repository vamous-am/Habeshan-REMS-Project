package dashboard

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/habeshan-rems/backend/internal/common"
	"gorm.io/gorm"
)

type DashboardService struct {
	db *gorm.DB
}

func NewDashboardService(db *gorm.DB) *DashboardService {
	return &DashboardService{db: db}
}

func getOrgIDFromCtx(c *fiber.Ctx) (uuid.UUID, error) {
	if val, ok := c.Locals("org_id").(string); ok && val != "" {
		return uuid.Parse(val)
	}
	if val, ok := c.Locals("org_id").(uuid.UUID); ok {
		return val, nil
	}
	if val := c.Get("X-Org-ID"); val != "" {
		return uuid.Parse(val)
	}
	return uuid.Nil, common.ErrUnauthorized
}

// GetManagerDashboard handles GET /api/v1/dashboard/manager
// Now wired to live data — FR-DASH-01
func GetManagerDashboardLive(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		orgID, err := getOrgIDFromCtx(c)
		if err != nil {
			return common.Fail(c, fiber.StatusUnauthorized, "unauthorized")
		}

		// Attendance today
		var present int64
		var total int64
		db.Table("attendance_logs").
			Where("org_id = ? AND DATE(clock_in) = CURRENT_DATE AND sync_status = 'SYNCED_VERIFIED'", orgID).
			Count(&present)
		db.Table("users").Where("org_id = ? AND status = 'active'", orgID).Count(&total)

		// Task progress
		type TaskCount struct {
			Status string
			Count  int64
		}
		var taskCounts []TaskCount
		db.Table("tasks").
			Select("status, COUNT(*) as count").
			Where("org_id = ?", orgID).
			Group("status").
			Scan(&taskCounts)

		taskProgress := fiber.Map{}
		for _, tc := range taskCounts {
			taskProgress[tc.Status] = tc.Count
		}

		// Pending approvals
		var pendingApprovals int64
		db.Table("timesheets").
			Where("org_id = ? AND status = 'submitted'", orgID).
			Count(&pendingApprovals)

		return common.OK(c, fiber.Map{
			"team_attendance_today": fiber.Map{
				"present": present,
				"absent":  total - present,
				"total":   total,
			},
			"task_progress":     taskProgress,
			"pending_approvals": pendingApprovals,
		})
	}
}

// GetLeaderboard returns placeholder leaderboard data until ranking is implemented.
func GetLeaderboard(c *fiber.Ctx) error {
	return common.OK(c, fiber.Map{
		"opted_out": false,
		"entries":   []fiber.Map{},
	})
}
