package dashboard

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/habeshan-rems/backend/internal/common"
	"gorm.io/gorm"
)

// AttendanceReport handles GET /api/v1/reports/attendance
// Filterable by user_id, date range — FR-DASH-02
func AttendanceReport(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		orgID := c.Locals("org_id").(uuid.UUID)

		fromStr := c.Query("from")
		toStr := c.Query("to")
		userIDStr := c.Query("user_id")

		from, _ := time.Parse("2006-01-02", fromStr)
		to, _ := time.Parse("2006-01-02", toStr)
		if to.IsZero() {
			to = time.Now()
		}
		if from.IsZero() {
			from = to.AddDate(0, -1, 0)
		}

		query := db.Table("attendance_logs").
			Where("org_id = ? AND clock_in >= ? AND clock_in <= ?", orgID, from, to).
			Where("sync_status = 'SYNCED_VERIFIED'")

		if userIDStr != "" {
			uid, err := uuid.Parse(userIDStr)
			if err == nil {
				query = query.Where("user_id = ?", uid)
			}
		}

		var rows []map[string]interface{}
		if err := query.Find(&rows).Error; err != nil {
			return common.HandleError(c, err)
		}

		return common.OK(c, fiber.Map{
			"from":    from.Format("2006-01-02"),
			"to":      to.Format("2006-01-02"),
			"records": rows,
		})
	}
}

// TaskReport handles GET /api/v1/reports/tasks
// Filterable by user_id, status, date range — FR-DASH-03
func TaskReport(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		orgID := c.Locals("org_id").(uuid.UUID)

		fromStr := c.Query("from")
		toStr := c.Query("to")
		userIDStr := c.Query("user_id")
		status := c.Query("status")

		from, _ := time.Parse("2006-01-02", fromStr)
		to, _ := time.Parse("2006-01-02", toStr)
		if to.IsZero() {
			to = time.Now()
		}
		if from.IsZero() {
			from = to.AddDate(0, -1, 0)
		}

		query := db.Table("task_time_logs").
			Where("org_id = ? AND started_at >= ? AND started_at <= ?", orgID, from, to)

		if userIDStr != "" {
			uid, err := uuid.Parse(userIDStr)
			if err == nil {
				query = query.Where("user_id = ?", uid)
			}
		}
		if status != "" {
			query = query.Joins("JOIN tasks ON tasks.id = task_time_logs.task_id").
				Where("tasks.status = ?", status)
		}

		var rows []map[string]interface{}
		if err := query.Find(&rows).Error; err != nil {
			return common.HandleError(c, err)
		}

		return common.OK(c, fiber.Map{
			"from":    from.Format("2006-01-02"),
			"to":      to.Format("2006-01-02"),
			"records": rows,
		})
	}
}
