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
	ts.Get("/", handler.ListTimesheets)
	ts.Get("/:id", handler.GetTimesheet)
	ts.Put("/:id/submit", handler.SubmitTimesheet)
	ts.Put("/:id/approve", handler.ApproveTimesheet)
	ts.Put("/:id/reject", handler.RejectTimesheet)
}
