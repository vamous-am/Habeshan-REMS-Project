package timesheets

import (
	"encoding/csv"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ExportCSV handles GET /api/v1/timesheets/export?format=csv&user_id=...
// FR-TS-07
func ExportCSV(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userIDStr := c.Query("user_id")
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("valid user_id required")
		}

		var timesheets []Timesheet
		if err := db.Where("user_id = ?", userID).
			Order("period_start DESC").
			Find(&timesheets).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("failed to fetch timesheets")
		}

		var sb strings.Builder
		w := csv.NewWriter(&sb)

		// Header row
		w.Write([]string{"Period Start", "Period End", "Total Hours", "Status", "Rejection Reason"})

		for _, ts := range timesheets {
			reason := ""
			if ts.RejectionReason != nil {
				reason = *ts.RejectionReason
			}
			w.Write([]string{
				ts.PeriodStart.Format("2006-01-02"),
				ts.PeriodEnd.Format("2006-01-02"),
				fmt.Sprintf("%.2f", ts.TotalHours),
				ts.Status,
				reason,
			})
		}
		w.Flush()

		c.Set("Content-Type", "text/csv")
		c.Set("Content-Disposition", "attachment; filename=timesheets.csv")
		return c.SendString(sb.String())
	}
}
