package attendance

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// RegisterRoutes initializes dependencies and mounts HTTP endpoints onto the Fiber app
func RegisterRoutes(app *fiber.App, db *gorm.DB) {
	service := NewService(db)
	handler := NewHandler(service)

	api := app.Group("/api/v1/attendance")
	api.Post("/clock-in", handler.ClockIn)
	api.Post("/clock-out", handler.ClockOut)
}