package timesheets

import (
	"github.com/gofiber/fiber/v2"
	"github.com/habeshan-rems/backend/internal/middleware"
	"gorm.io/gorm"
)

func RegisterRoutes(app *fiber.App, db *gorm.DB) {
	service := NewService(db)
	handler := NewHandler(service)

	ts := app.Group("/api/v1/timesheets", middleware.JWTAuth())

	// Employee + Manager + Admin: list and get
	ts.Get("/", handler.ListTimesheets)
	ts.Get("/:id", handler.GetTimesheet)

	// Only the owning employee may submit their timesheet
	ts.Put("/:id/submit", middleware.RequireRole("employee"), handler.SubmitTimesheet)

	// Managers (team-scoped) and admins (org-wide) may approve or reject timesheets
	ts.Put("/:id/approve", middleware.RequireRole("manager", "admin"), handler.ApproveTimesheet)
	ts.Put("/:id/reject", middleware.RequireRole("manager", "admin"), handler.RejectTimesheet)
}
