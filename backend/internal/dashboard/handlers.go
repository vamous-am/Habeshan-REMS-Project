package dashboard

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/habeshan-rems/backend/internal/common"
	"gorm.io/gorm"
)

// leaderboardRow is the raw result of the leaderboard aggregation query.
type leaderboardRow struct {
	FullName            string  `gorm:"column:full_name"`
	UserID              string  `gorm:"column:user_id"`
	CompletedCount      int64   `gorm:"column:completed_count"`
	TotalCount          int64   `gorm:"column:total_count"`
	CompletionPct       float64 `gorm:"column:completion_pct"`
}

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

// GetLeaderboard implements FR-DASH-04: top-3 performers ranked by task-completion
// percentage, calculated across all historically assigned tasks in the org.
//
// Metric formula (computed in SQL):
//
//	completion_pct = completed_count / NULLIF(total_count, 0) * 100
//
// Tie-breaking (deterministic):
//
//	completion_pct DESC, completed_count DESC, full_name ASC
//
// Scope:
//   - Only tasks in the caller's org (via tasks.org_id).
//   - Only users who have ≥ 1 assigned task (zero-assignment users excluded).
//   - User rows with deleted_at IS NOT NULL are excluded.
//   - All historical completed tasks (no rolling window — MVP assumption, documented).
//
// The assignment is many-to-many via task_assignments; a task completed counts
// towards every user it was assigned to.
func GetLeaderboard(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		orgID, err := getOrgIDFromCtx(c)
		if err != nil {
			return common.Fail(c, fiber.StatusUnauthorized, "unauthorized")
		}

		var rows []leaderboardRow
		err = db.Raw(`
			SELECT
				u.full_name                                                   AS full_name,
				CAST(ta.user_id AS TEXT)                                      AS user_id,
				COUNT(DISTINCT CASE WHEN t.status = 'completed' THEN t.id END) AS completed_count,
				COUNT(DISTINCT t.id)                                          AS total_count,
				CAST(
					COUNT(DISTINCT CASE WHEN t.status = 'completed' THEN t.id END)
					* 100.0
					/ NULLIF(COUNT(DISTINCT t.id), 0)
				AS FLOAT)                                                      AS completion_pct
			FROM task_assignments ta
			JOIN tasks t ON t.id = ta.task_id
			JOIN users u ON u.id = ta.user_id
			WHERE t.org_id = ?
			  AND u.deleted_at IS NULL
			GROUP BY ta.user_id, u.full_name
			HAVING COUNT(DISTINCT t.id) > 0
			ORDER BY completion_pct DESC,
			         completed_count DESC,
			         u.full_name ASC
			LIMIT 3
		`, orgID.String()).Scan(&rows).Error
		if err != nil {
			return common.Fail(c, fiber.StatusInternalServerError, "failed to load leaderboard")
		}

		// Assign sequential ranks based on sorted position.
		type entry struct {
			Rank                 int     `json:"rank"`
			EmployeeName         string  `json:"employee_name"`
			CompletedCount       int64   `json:"completed_count"`
			TotalCount           int64   `json:"total_count"`
			CompletionPercentage float64 `json:"completion_percentage"`
		}
		entries := make([]entry, 0, len(rows))
		for i, r := range rows {
			entries = append(entries, entry{
				Rank:                 i + 1,
				EmployeeName:         r.FullName,
				CompletedCount:       r.CompletedCount,
				TotalCount:           r.TotalCount,
				CompletionPercentage: r.CompletionPct,
			})
		}

		return common.OK(c, fiber.Map{
			"entries": entries,
		})
	}
}
