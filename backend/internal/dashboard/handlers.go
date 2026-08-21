package dashboard

import "github.com/gofiber/fiber/v2"

// GetManagerDashboard returns mock/static data for now — Task 2 is just
// the shell so frontend has something to build against. Real aggregation
// (FR-DASH-01) comes later, once attendance + task-timer data exists.
func GetManagerDashboard(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": "success",
		"data": fiber.Map{
			"team_attendance_today": fiber.Map{
				"present": 8,
				"absent":  2,
				"total":   10,
			},
			"task_progress": fiber.Map{
				"to_do":       5,
				"in_progress": 7,
				"completed":   12,
			},
			"pending_approvals": 3,
		},
	})
}
