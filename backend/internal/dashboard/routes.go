package dashboard

import (
	"github.com/gofiber/fiber/v2"
	"github.com/habeshan-rems/backend/internal/middleware"
	"github.com/habeshan-rems/backend/internal/timesheets"
	"gorm.io/gorm"
)

// RegisterRoutes wires this feature's endpoints onto the app.
func RegisterRoutes(app *fiber.App, db *gorm.DB) {
	d := app.Group("/api/v1", middleware.JWTAuth())
	d.Get("/dashboard/manager", GetManagerDashboardLive(db))
	d.Get("/dashboard/leaderboard", GetLeaderboard)
	d.Get("/reports/attendance", AttendanceReport(db))
	d.Get("/reports/tasks", TaskReport(db))
	d.Get("/timesheets/export", timesheets.ExportCSV(db))
}
