package admin

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/habeshan-rems/backend/internal/common"
	"github.com/habeshan-rems/backend/internal/middleware"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// orgIDFromCtx pulls org_id out of the JWT claims stored by JWTAuth.
// Every handler in this file uses this — it's the enforcement point for
// "an admin can only ever act within their own org."
func orgIDFromCtx(c *fiber.Ctx) (uuid.UUID, error) {
	raw, _ := c.Locals(middleware.LocalOrgID).(string)
	return uuid.Parse(raw)
}

// ── User CRUD ────────────────────────────────────────────────────────────────

func (h *Handler) ListUsers(c *fiber.Ctx) error {
	orgID, err := orgIDFromCtx(c)
	if err != nil {
		return common.HandleError(c, common.ErrUnauthorized)
	}
	users, err := h.svc.ListUsers(orgID)
	if err != nil {
		return common.HandleError(c, err)
	}
	return common.OK(c, users)
}

func (h *Handler) GetUser(c *fiber.Ctx) error {
	orgID, err := orgIDFromCtx(c)
	if err != nil {
		return common.HandleError(c, common.ErrUnauthorized)
	}
	userID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return common.HandleError(c, common.ErrBadRequest)
	}
	user, err := h.svc.GetUser(orgID, userID)
	if err != nil {
		return common.HandleError(c, err)
	}
	return common.OK(c, user)
}

func (h *Handler) UpdateUser(c *fiber.Ctx) error {
	orgID, err := orgIDFromCtx(c)
	if err != nil {
		return common.HandleError(c, common.ErrUnauthorized)
	}
	userID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return common.HandleError(c, common.ErrBadRequest)
	}
	var req UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return common.HandleError(c, common.ErrBadRequest)
	}
	user, err := h.svc.UpdateUser(orgID, userID, req)
	if err != nil {
		return common.HandleError(c, err)
	}
	return common.OK(c, user)
}

func (h *Handler) DeleteUser(c *fiber.Ctx) error {
	orgID, err := orgIDFromCtx(c)
	if err != nil {
		return common.HandleError(c, common.ErrUnauthorized)
	}
	userID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return common.HandleError(c, common.ErrBadRequest)
	}
	if err := h.svc.DeleteUser(orgID, userID); err != nil {
		return common.HandleError(c, err)
	}
	return common.OK(c, fiber.Map{"deleted": true})
}

// ── Role assignment ──────────────────────────────────────────────────────────

func (h *Handler) UpdateUserRole(c *fiber.Ctx) error {
	orgID, err := orgIDFromCtx(c)
	if err != nil {
		return common.HandleError(c, common.ErrUnauthorized)
	}
	userID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return common.HandleError(c, common.ErrBadRequest)
	}
	var req UpdateUserRoleRequest
	if err := c.BodyParser(&req); err != nil || req.Role == "" {
		return common.HandleError(c, common.ErrBadRequest)
	}
	user, err := h.svc.UpdateUserRole(orgID, userID, req.Role)
	if err != nil {
		return common.HandleError(c, err)
	}
	return common.OK(c, user)
}

// ── Teams ────────────────────────────────────────────────────────────────────

func (h *Handler) CreateTeam(c *fiber.Ctx) error {
	orgID, err := orgIDFromCtx(c)
	if err != nil {
		return common.HandleError(c, common.ErrUnauthorized)
	}
	var req CreateTeamRequest
	if err := c.BodyParser(&req); err != nil || req.Name == "" || req.ManagerID == "" {
		return common.HandleError(c, common.ErrBadRequest)
	}
	team, err := h.svc.CreateTeam(orgID, req)
	if err != nil {
		return common.HandleError(c, err)
	}
	return common.Created(c, team)
}

func (h *Handler) ListTeams(c *fiber.Ctx) error {
	orgID, err := orgIDFromCtx(c)
	if err != nil {
		return common.HandleError(c, common.ErrUnauthorized)
	}
	teams, err := h.svc.ListTeams(orgID)
	if err != nil {
		return common.HandleError(c, err)
	}
	return common.OK(c, teams)
}

func (h *Handler) AddTeamMember(c *fiber.Ctx) error {
	orgID, err := orgIDFromCtx(c)
	if err != nil {
		return common.HandleError(c, common.ErrUnauthorized)
	}
	teamID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return common.HandleError(c, common.ErrBadRequest)
	}
	var req AddTeamMemberRequest
	if err := c.BodyParser(&req); err != nil || req.UserID == "" {
		return common.HandleError(c, common.ErrBadRequest)
	}
	if err := h.svc.AddTeamMember(orgID, teamID, req); err != nil {
		return common.HandleError(c, err)
	}
	return common.OK(c, fiber.Map{"added": true})
}

func (h *Handler) RemoveTeamMember(c *fiber.Ctx) error {
	orgID, err := orgIDFromCtx(c)
	if err != nil {
		return common.HandleError(c, common.ErrUnauthorized)
	}
	teamID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return common.HandleError(c, common.ErrBadRequest)
	}
	userID, err := uuid.Parse(c.Params("userId"))
	if err != nil {
		return common.HandleError(c, common.ErrBadRequest)
	}
	if err := h.svc.RemoveTeamMember(orgID, teamID, userID); err != nil {
		return common.HandleError(c, err)
	}
	return common.OK(c, fiber.Map{"removed": true})
}

// ── Org settings ─────────────────────────────────────────────────────────────

func (h *Handler) GetOrgSettings(c *fiber.Ctx) error {
	orgID, err := orgIDFromCtx(c)
	if err != nil {
		return common.HandleError(c, common.ErrUnauthorized)
	}
	org, err := h.svc.GetOrgSettings(orgID)
	if err != nil {
		return common.HandleError(c, err)
	}
	return common.OK(c, org)
}

func (h *Handler) UpdateOrgSettings(c *fiber.Ctx) error {
	orgID, err := orgIDFromCtx(c)
	if err != nil {
		return common.HandleError(c, common.ErrUnauthorized)
	}
	var req UpdateOrgRequest
	if err := c.BodyParser(&req); err != nil {
		return common.HandleError(c, common.ErrBadRequest)
	}
	org, err := h.svc.UpdateOrgSettings(orgID, req)
	if err != nil {
		return common.HandleError(c, err)
	}
	return common.OK(c, org)
}