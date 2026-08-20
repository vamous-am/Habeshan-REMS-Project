package middleware

import (
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/habeshan-rems/backend/internal/auth"
	"github.com/habeshan-rems/backend/internal/common"
)

// Locals keys set by JWTAuth for downstream handlers and RequireRole.
const (
	LocalUserID = "user_id"
	LocalOrgID  = "org_id"
	LocalRole   = "role"
)

// JWTAuth validates the Bearer token against JWT_SECRET using the same
// auth.Claims shape auth.Service signs (FR-AUTH-04). On success it stores
// user_id, org_id, and role in fiber.Locals under the constants above.
func JWTAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		header := c.Get("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			return common.Fail(c, fiber.StatusUnauthorized, "missing or malformed token")
		}

		tokenString := strings.TrimPrefix(header, "Bearer ")
		claims := &auth.Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fiber.ErrUnauthorized
			}
			return []byte(os.Getenv("JWT_SECRET")), nil
		})
		if err != nil || !token.Valid {
			return common.Fail(c, fiber.StatusUnauthorized, "invalid or expired token")
		}

		c.Locals(LocalUserID, claims.UserID)
		c.Locals(LocalOrgID, claims.OrgID)
		c.Locals(LocalRole, claims.Role)
		return c.Next()
	}
}