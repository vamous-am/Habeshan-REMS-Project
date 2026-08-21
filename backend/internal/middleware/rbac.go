package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/habeshan-rems/backend/internal/common"
)

// RequireRole restricts a route to the given roles. Run it after JWTAuth —
// it reads the role JWTAuth stores in fiber.Locals (FR-AUTH-05).
//
// Usage:
//   admin := api.Group("/admin", middleware.JWTAuth(), middleware.RequireRole("admin"))
func RequireRole(roles ...string) fiber.Handler {
	allowed := make(map[string]struct{}, len(roles))
	for _, r := range roles {
		allowed[r] = struct{}{}
	}
	return func(c *fiber.Ctx) error {
		role, ok := c.Locals(LocalRole).(string)
		if !ok || role == "" {
			// Misconfigured route: RequireRole used without JWTAuth first
			return common.HandleError(c, common.ErrUnauthorized)
		}
		if _, permitted := allowed[role]; !permitted {
			return common.HandleError(c, common.ErrForbidden)
		}
		return c.Next()
	}
}