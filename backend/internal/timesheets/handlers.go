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
// Returns all timesheets for the requesting user.
// TODO: replace user_id query param with JWT claims once Dev 1's middleware is wired
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
