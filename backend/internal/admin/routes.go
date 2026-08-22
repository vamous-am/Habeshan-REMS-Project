package admin

import (
	"github.com/gofiber/fiber/v2"
	"github.com/habeshan-rems/backend/internal/middleware"
	"gorm.io/gorm"
)

// RegisterRoutes wires every admin endpoint behind JWTAuth + RequireRole("admin").
// No route in this group is reachable by manager or employee tokens.
func RegisterRoutes(app *fiber.App, db *gorm.DB) {
	svc := NewService(db)
	h := NewHandler(svc)

	adminGroup := app.Group("/api/v1/admin", middleware.JWTAuth(), middleware.RequireRole("admin"))

	// User CRUD (FR-ADMIN-01)
	adminGroup.Get("/users", h.ListUsers)
	adminGroup.Get("/users/:id", h.GetUser)
	adminGroup.Patch("/users/:id", h.UpdateUser)
	adminGroup.Delete("/users/:id", h.DeleteUser)

	// Role assignment (FR-ADMIN-02)
	adminGroup.Patch("/users/:id/role", h.UpdateUserRole)

	// Teams (FR-ADMIN-03)
	adminGroup.Post("/teams", h.CreateTeam)
	adminGroup.Get("/teams", h.ListTeams)
	adminGroup.Post("/teams/:id/members", h.AddTeamMember)
	adminGroup.Delete("/teams/:id/members/:userId", h.RemoveTeamMember)

	// Org settings (FR-ADMIN-04)
	adminGroup.Get("/org", h.GetOrgSettings)
	adminGroup.Patch("/org", h.UpdateOrgSettings)
}