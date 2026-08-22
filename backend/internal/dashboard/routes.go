package dashboard

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// RegisterRoutes wires this feature's endpoints onto the app.
func RegisterRoutes(app *fiber.App, db *gorm.DB) {
	app.Get("/api/v1/dashboard/manager", GetManagerDashboardLive(db))
	app.Get("/api/v1/dashboard/leaderboard", GetLeaderboard)
	app.Get("/api/v1/reports/attendance", AttendanceReport(db))
	app.Get("/api/v1/reports/tasks", TaskReport(db))
}
