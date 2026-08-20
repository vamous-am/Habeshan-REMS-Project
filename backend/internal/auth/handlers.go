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

// Login handles POST /api/v1/auth/login
func (h *Handler) Login(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return common.Fail(c, fiber.StatusBadRequest, "invalid request body")
	}

	if req.Email == "" || req.Password == "" {
		return common.Fail(c, fiber.StatusBadRequest, "email and password are required")
	}

	resp, err := h.svc.Login(req)
	if err != nil {
		return common.HandleError(c, err)
	}

	return common.OK(c, resp)
}