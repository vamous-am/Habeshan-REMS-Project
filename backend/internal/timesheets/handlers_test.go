package timesheets

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/habeshan-rems/backend/internal/middleware"
)

// TestGetCallerContext_StringLocals is the regression test for the 401 bug
// found 2026-11-05: JWTAuth stores user_id and org_id as strings (see
// auth.Claims and middleware/jwt.go), but getCallerContext used to assert
// them as uuid.UUID, so every timesheet request failed with 401 even for a
// signed-in admin. This test feeds the function the string form, exactly
// as a real login does, plus the uuid form and the failure cases.
func TestGetCallerContext_StringLocals(t *testing.T) {
	userID := uuid.New()
	orgID := uuid.New()

	cases := []struct {
		name       string
		userVal    interface{}
		orgVal     interface{}
		roleVal    interface{}
		wantStatus int
		wantBody   string
	}{
		{
			name:       "string IDs as JWTAuth stores them",
			userVal:    userID.String(),
			orgVal:     orgID.String(),
			roleVal:    "admin",
			wantStatus: http.StatusOK,
			wantBody:   userID.String() + "/" + orgID.String() + "/admin",
		},
		{
			name:       "uuid IDs still accepted",
			userVal:    userID,
			orgVal:     orgID,
			roleVal:    "manager",
			wantStatus: http.StatusOK,
			wantBody:   userID.String() + "/" + orgID.String() + "/manager",
		},
		{
			name:       "missing IDs rejected",
			userVal:    nil,
			orgVal:     nil,
			roleVal:    "admin",
			wantStatus: http.StatusUnauthorized,
			wantBody:   "unauthorized",
		},
		{
			name:       "malformed string ID rejected",
			userVal:    "not-a-uuid",
			orgVal:     orgID.String(),
			roleVal:    "admin",
			wantStatus: http.StatusUnauthorized,
			wantBody:   "unauthorized",
		},
		{
			name:       "missing role rejected",
			userVal:    userID.String(),
			orgVal:     orgID.String(),
			roleVal:    "",
			wantStatus: http.StatusUnauthorized,
			wantBody:   "unauthorized",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			app := fiber.New()
			app.Get("/test", func(c *fiber.Ctx) error {
				if tc.userVal != nil {
					c.Locals(middleware.LocalUserID, tc.userVal)
				}
				if tc.orgVal != nil {
					c.Locals(middleware.LocalOrgID, tc.orgVal)
				}
				if tc.roleVal != nil {
					c.Locals(middleware.LocalRole, tc.roleVal)
				}
				uid, oid, role, err := getCallerContext(c)
				if err != nil {
					return c.Status(http.StatusUnauthorized).SendString("unauthorized")
				}
				return c.SendString(uid.String() + "/" + oid.String() + "/" + string(role))
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("app.Test failed: %v", err)
			}
			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)
			if resp.StatusCode != tc.wantStatus {
				t.Fatalf("status = %d, want %d (body %q)", resp.StatusCode, tc.wantStatus, string(body))
			}
			if string(body) != tc.wantBody {
				t.Fatalf("body = %q, want %q", string(body), tc.wantBody)
			}
		})
	}
}
