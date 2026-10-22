package timesheets

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/habeshan-rems/backend/internal/auth"
	"github.com/habeshan-rems/backend/internal/common"
	"github.com/habeshan-rems/backend/internal/middleware"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// getCallerContext extracts (userID, orgID, role) from JWT locals set by JWTAuth.
// Returns ErrUnauthorized if any field is missing without relying on client-supplied headers.
func getCallerContext(c *fiber.Ctx) (userID uuid.UUID, orgID uuid.UUID, role auth.Role, err error) {
	userIDVal, ok := c.Locals(middleware.LocalUserID).(uuid.UUID)
	if !ok || userIDVal == uuid.Nil {
		return uuid.Nil, uuid.Nil, "", common.ErrUnauthorized
	}
	orgIDVal, ok := c.Locals(middleware.LocalOrgID).(uuid.UUID)
	if !ok || orgIDVal == uuid.Nil {
		return uuid.Nil, uuid.Nil, "", common.ErrUnauthorized
	}
	roleVal, ok := c.Locals(middleware.LocalRole).(string)
	if !ok || roleVal == "" {
		return uuid.Nil, uuid.Nil, "", common.ErrUnauthorized
	}
	return userIDVal, orgIDVal, auth.Role(roleVal), nil
}

// ListTimesheets handles GET /api/v1/timesheets
// Access rules:
//   - Employee: may only list their own timesheets (user_id forced to callerID)
//   - Manager:  may list timesheets for any user in a team they manage
//   - Admin:    may list timesheets for any user in their org
//
// The ?user_id= query param is honored only for Manager/Admin callers.
func (h *Handler) ListTimesheets(c *fiber.Ctx) error {
	callerID, orgID, role, err := getCallerContext(c)
	if err != nil {
		return common.HandleError(c, err)
	}

	var targetUserID uuid.UUID

	switch role {
	case auth.RoleEmployee:
		// Employees may only see their own timesheets regardless of query param
		targetUserID = callerID

	case auth.RoleManager, auth.RoleAdmin:
		userIDStr := c.Query("user_id")
		if userIDStr != "" {
			parsed, parseErr := uuid.Parse(userIDStr)
			if parseErr != nil {
				return common.Fail(c, fiber.StatusBadRequest, "invalid user_id query param")
			}
			// For manager, verify the target employee is in one of their teams
			if role == auth.RoleManager {
				if scopeErr := h.service.assertManagerOwnsEmployee(orgID, callerID, parsed); scopeErr != nil {
					return common.HandleError(c, scopeErr)
				}
			}
			targetUserID = parsed
		} else {
			// No filter: admin sees all in org; manager sees all in their teams (handled by query below)
			targetUserID = uuid.Nil
		}
	}

	query := h.service.db.Where("org_id = ?", orgID).Order("period_start DESC")

	if targetUserID != uuid.Nil {
		query = query.Where("user_id = ?", targetUserID)
	} else if role == auth.RoleManager {
		// Manager with no user_id filter: scope to employees in their teams
		query = query.Where(
			"user_id IN (SELECT tm.user_id FROM team_members tm JOIN teams t ON t.id = tm.team_id WHERE t.manager_id = ? AND t.org_id = ?)",
			callerID, orgID,
		)
	}

	var timesheets []Timesheet
	if err := query.Find(&timesheets).Error; err != nil {
		return common.HandleError(c, err)
	}

	return common.OK(c, timesheets)
}

// GetTimesheet handles GET /api/v1/timesheets/:id
// Scoped to the caller's org to prevent cross-org data access.
func (h *Handler) GetTimesheet(c *fiber.Ctx) error {
	_, orgID, _, err := getCallerContext(c)
	if err != nil {
		return common.HandleError(c, err)
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return common.Fail(c, fiber.StatusBadRequest, "invalid timesheet id")
	}

	var ts Timesheet
	// Filter by org_id to ensure tenant isolation
	if err := h.service.db.First(&ts, "id = ? AND org_id = ?", id, orgID).Error; err != nil {
		return common.HandleError(c, common.ErrNotFound)
	}

	return common.OK(c, ts)
}

// SubmitTimesheet handles PUT /api/v1/timesheets/:id/submit
// Restricted to role=employee by routes.go. Ownership verified in service.Submit.
// Manager/admin submit-on-behalf is prohibited.
func (h *Handler) SubmitTimesheet(c *fiber.Ctx) error {
	callerID, _, _, err := getCallerContext(c)
	if err != nil {
		return common.HandleError(c, err)
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return common.Fail(c, fiber.StatusBadRequest, "invalid timesheet id")
	}

	if err := h.service.Submit(id, callerID); err != nil {
		return common.HandleError(c, err)
	}

	return common.OK(c, fiber.Map{"message": "timesheet submitted"})
}

// ApproveTimesheet handles PUT /api/v1/timesheets/:id/approve
// Restricted to role=manager|admin by routes.go.
// Manager callers are team-scoped inside service.Approve.
func (h *Handler) ApproveTimesheet(c *fiber.Ctx) error {
	reviewerID, orgID, role, err := getCallerContext(c)
	if err != nil {
		return common.HandleError(c, err)
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return common.Fail(c, fiber.StatusBadRequest, "invalid timesheet id")
	}

	if err := h.service.Approve(id, reviewerID, role, orgID); err != nil {
		return common.HandleError(c, err)
	}

	return common.OK(c, fiber.Map{"message": "timesheet approved"})
}

// RejectTimesheet handles PUT /api/v1/timesheets/:id/reject
// Restricted to role=manager|admin by routes.go.
// Manager callers are team-scoped inside service.Reject.
func (h *Handler) RejectTimesheet(c *fiber.Ctx) error {
	reviewerID, orgID, role, err := getCallerContext(c)
	if err != nil {
		return common.HandleError(c, err)
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return common.Fail(c, fiber.StatusBadRequest, "invalid timesheet id")
	}

	var body struct {
		Reason string `json:"reason"`
	}
	if err := c.BodyParser(&body); err != nil {
		return common.Fail(c, fiber.StatusBadRequest, "invalid request body")
	}
	if body.Reason == "" {
		return common.Fail(c, fiber.StatusBadRequest, "rejection reason is required")
	}

	if err := h.service.Reject(id, reviewerID, role, orgID, body.Reason); err != nil {
		return common.HandleError(c, err)
	}

	return common.OK(c, fiber.Map{"message": "timesheet rejected"})
}
