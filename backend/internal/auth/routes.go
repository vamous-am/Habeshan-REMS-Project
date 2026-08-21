package auth

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// RegisterRoutes mounts the auth endpoints under /api/v1/auth.
// Called from main.go after the database is connected.
func RegisterRoutes(app *fiber.App, db *gorm.DB) {
	svc := NewService(db)
	h := NewHandler(svc)

	auth := app.Group("/api/v1/auth")
	auth.Post("/lookup", h.Lookup)   // step 1: which orgs have this email?
	auth.Post("/register", h.Register) // FR-AUTH-08
	auth.Post("/login", h.Login)       // FR-AUTH-01/02/03
	auth.Post("/logout", h.Logout)     // FR-AUTH-06
}