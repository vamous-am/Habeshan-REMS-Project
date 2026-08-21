package attendance

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func RegisterRoutes(app *fiber.App, db *gorm.DB) {
	service := NewService(db)
	handler := NewHandler(service)

	api := app.Group("/api/v1/attendance")
	api.Post("/clock-in", handler.ClockIn)
	api.Post("/clock-out", handler.ClockOut)
	api.Post("/sync", handler.SyncBatch)

	// Scoped history views (Task 15)
	api.Get("/me", handler.GetSelfHistory)
	api.Get("/team", handler.GetTeamHistory)
	api.Get("/org", handler.GetOrgHistory)
}
