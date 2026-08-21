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

// GetLeaderboard returns mock leaderboard data for now.
//
// Task 07 is only the leaderboard shell.
// Real ranking data comes later once attendance and task-timer
// aggregation is available.
func GetLeaderboard(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": "success",
		"data": fiber.Map{
			"opted_out": false,
			"entries": []fiber.Map{
				{
					"rank":                  1,
					"employee_name":         "Employee A",
					"completion_percentage": 94,
				},
				{
					"rank":                  2,
					"employee_name":         "Employee B",
					"completion_percentage": 87,
				},
				{
					"rank":                  3,
					"employee_name":         "Employee C",
					"completion_percentage": 81,
				},
			},
		},
	})
}
