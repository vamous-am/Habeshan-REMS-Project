package auth

import (
	"github.com/gofiber/fiber/v2"
	"github.com/habeshan-rems/backend/internal/common"
)

// Handler exposes HTTP handlers for auth endpoints.
// It only handles request/response — business logic lives in the Service.
type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// Register handles POST /api/v1/auth/register
func (h *Handler) Register(c *fiber.Ctx) error {
	var req RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return common.Fail(c, fiber.StatusBadRequest, "invalid request body")
	}

	if req.OrgName == "" || req.FullName == "" || req.Email == "" || req.Password == "" {
		return common.Fail(c, fiber.StatusBadRequest, "org_name, full_name, email and password are required")
	}
	if len(req.Password) < 8 {
		return common.Fail(c, fiber.StatusBadRequest, "password must be at least 8 characters")
	}

	resp, err := h.svc.Register(req)
	if err != nil {
		return common.HandleError(c, err)
	}

	return common.Created(c, resp)
}

// Lookup handles POST /api/v1/auth/lookup
// Step 1 of the two-step login flow — returns which orgs have this email.
func (h *Handler) Lookup(c *fiber.Ctx) error {
	var req LookupRequest
	if err := c.BodyParser(&req); err != nil {
		return common.Fail(c, fiber.StatusBadRequest, "invalid request body")
	}
	if req.Email == "" {
		return common.Fail(c, fiber.StatusBadRequest, "email is required")
	}

	resp, err := h.svc.Lookup(req)
	if err != nil {
		return common.HandleError(c, err)
	}

	return common.OK(c, resp)
}

// Login handles POST /api/v1/auth/login
func (h *Handler) Login(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return common.Fail(c, fiber.StatusBadRequest, "invalid request body")
	}

	if req.Email == "" || req.Password == "" || req.OrgID == "" {
		return common.Fail(c, fiber.StatusBadRequest, "email, password and org_id are required")
	}

	resp, err := h.svc.Login(req)
	if err != nil {
		return common.HandleError(c, err)
	}

	return common.OK(c, resp)
}
// ForgotPassword handles POST /api/v1/auth/forgot-password
// ⚠️ MVP/demo: returns the reset token directly in the response.
func (h *Handler) ForgotPassword(c *fiber.Ctx) error {
	var req ForgotPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return common.Fail(c, fiber.StatusBadRequest, "invalid request body")
	}
	if req.Email == "" || req.OrgID == "" {
		return common.Fail(c, fiber.StatusBadRequest, "email and org_id are required")
	}

	resp, err := h.svc.ForgotPassword(req)
	if err != nil {
		return common.HandleError(c, err)
	}

	return common.OK(c, resp)
}

// ResetPassword handles POST /api/v1/auth/reset-password
func (h *Handler) ResetPassword(c *fiber.Ctx) error {
	var req ResetPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return common.Fail(c, fiber.StatusBadRequest, "invalid request body")
	}
	if req.ResetToken == "" || req.NewPassword == "" {
		return common.Fail(c, fiber.StatusBadRequest, "reset_token and new_password are required")
	}
	if len(req.NewPassword) < 8 {
		return common.Fail(c, fiber.StatusBadRequest, "password must be at least 8 characters")
	}

	if err := h.svc.ResetPassword(req); err != nil {
		return common.HandleError(c, err)
	}

	return common.OK(c, fiber.Map{"message": "password reset successful"})
}

// Logout handles POST /api/v1/auth/logout

// There is no server-side session or token store right now, so this is a
// no-op: the client is responsible for deleting its stored token. This
// endpoint exists so the client always has something to call and the API
func (h *Handler) Logout(c *fiber.Ctx) error {
	return common.OK(c, fiber.Map{"message": "logged out"})
}