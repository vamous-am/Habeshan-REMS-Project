package attendance

import (
	"github.com/gofiber/fiber/v2"
	"github.com/habeshan-rems/backend/internal/middleware"
	"gorm.io/gorm"
)

func RegisterRoutes(app *fiber.App, db *gorm.DB) {
	service := NewService(db)
	handler := NewHandler(service)

	// Secure all attendance routes with Dev 1 JWT Authentication
	api := app.Group("/api/v1/attendance", middleware.JWTAuth())

	// Core Clock & Sync Endpoints (Accessible by all authenticated users)
	api.Post("/clock-in", handler.ClockIn)
	api.Post("/clock-out", handler.ClockOut)
	api.Post("/sync", handler.SyncBatch)

	// Scoped history views (Task 15)
	api.Get("/me", handler.GetSelfHistory)
	api.Get("/team", middleware.RequireRole("manager", "admin", "MANAGER", "ADMIN"), handler.GetTeamHistory)
	api.Get("/org", middleware.RequireRole("admin", "ADMIN"), handler.GetOrgHistory)
}