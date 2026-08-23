package timesheets

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/habeshan-rems/backend/internal/common"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// ListTimesheets handles GET /api/v1/timesheets
func (h *Handler) ListTimesheets(c *fiber.Ctx) error {
	userIDStr := c.Query("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return common.Fail(c, fiber.StatusBadRequest, "valid user_id query param required")
	}

	var timesheets []Timesheet
	if err := h.service.db.
		Where("user_id = ?", userID).
		Order("period_start DESC").
		Find(&timesheets).Error; err != nil {
		return common.HandleError(c, err)
	}

	return common.OK(c, timesheets)
}

// GetTimesheet handles GET /api/v1/timesheets/:id
func (h *Handler) GetTimesheet(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return common.Fail(c, fiber.StatusBadRequest, "invalid timesheet id")
	}

	var ts Timesheet
	if err := h.service.db.First(&ts, "id = ?", id).Error; err != nil {
		return common.HandleError(c, common.ErrNotFound)
	}

	return common.OK(c, ts)
}

// SubmitTimesheet handles PUT /api/v1/timesheets/:id/submit
// Transitions DRAFT → SUBMITTED — FR-TS-02/03
func (h *Handler) SubmitTimesheet(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return common.Fail(c, fiber.StatusBadRequest, "invalid timesheet id")
	}

	if err := h.service.Submit(id); err != nil {
		return common.HandleError(c, err)
	}

	return common.OK(c, fiber.Map{"message": "timesheet submitted"})
}

// ApproveTimesheet handles PUT /api/v1/timesheets/:id/approve
// Transitions SUBMITTED → APPROVED — FR-TS-04/05
func (h *Handler) ApproveTimesheet(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return common.Fail(c, fiber.StatusBadRequest, "invalid timesheet id")
	}

	// Pull reviewer ID from mock locals (swap for JWT claims later)
	reviewerID := c.Locals("user_id").(uuid.UUID)

	if err := h.service.Approve(id, reviewerID); err != nil {
		return common.HandleError(c, err)
	}

	return common.OK(c, fiber.Map{"message": "timesheet approved"})
}

// RejectTimesheet handles PUT /api/v1/timesheets/:id/reject
// Transitions SUBMITTED → REJECTED, reason required — FR-TS-04/06
func (h *Handler) RejectTimesheet(c *fiber.Ctx) error {
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

	reviewerID := c.Locals("user_id").(uuid.UUID)

	if err := h.service.Reject(id, reviewerID, body.Reason); err != nil {
		return common.HandleError(c, err)
	}

	return common.OK(c, fiber.Map{"message": "timesheet rejected"})
}
