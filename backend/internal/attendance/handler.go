package attendance

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/habeshan-rems/backend/internal/common"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// ClockIn handles POST /attendance/clock-in
func (h *Handler) ClockIn(c *fiber.Ctx) error {
	var req ClockInRequest
	if err := c.BodyParser(&req); err != nil {
		return common.Fail(c, fiber.StatusBadRequest, "Invalid request body")
	}

	if req.RecordUUID == uuid.Nil || req.DeviceHash == "" {
		return common.Fail(c, fiber.StatusBadRequest, "record_uuid and device_hash are required")
	}

	// Safely retrieve org_id and user_id injected by Dev 1's JWT middleware
	var orgID, userID uuid.UUID
	if val := c.Locals("org_id"); val != nil {
		orgID, _ = val.(uuid.UUID)
	}
	if val := c.Locals("user_id"); val != nil {
		userID, _ = val.(uuid.UUID)
	}

	log, err := h.service.ClockIn(orgID, userID, req)
	if err != nil {
		if err == ErrActiveClockInExists {
			return common.Fail(c, fiber.StatusConflict, err.Error())
		}
		return common.Fail(c, fiber.StatusInternalServerError, "Failed to record clock-in")
	}

	resp := ClockInResponse{
		ID:         log.ID,
		UserID:     log.UserID,
		ClockIn:    log.ClockIn,
		SyncStatus: log.SyncStatus,
		RecordUUID: log.RecordUUID,
	}

	return common.Created(c, resp)
}
