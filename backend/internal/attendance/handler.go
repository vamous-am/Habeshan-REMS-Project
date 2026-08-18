package attendance

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/habeshan-rems/backend/internal/common"
)

type Handler struct {
	service Service
}

// NewHandler initializes a new attendance HTTP handler
func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// ClockIn handles POST /attendance/clock-in
func (h *Handler) ClockIn(c *fiber.Ctx) error {
	var req ClockInRequest

	// 1. Parse JSON request body into the DTO
	if err := c.BodyParser(&req); err != nil {
		return common.Fail(c, fiber.StatusBadRequest, "Invalid request body")
	}

	// 2. Validate required payload fields
	if req.RecordUUID == uuid.Nil || req.DeviceHash == "" {
		return common.Fail(c, fiber.StatusBadRequest, "record_uuid and device_hash are required")
	}

	// 3. Extract org_id and user_id injected into c.Locals by Dev 1's JWT middleware
	var orgID, userID uuid.UUID
	if val := c.Locals("org_id"); val != nil {
		orgID, _ = val.(uuid.UUID)
	}
	if val := c.Locals("user_id"); val != nil {
		userID, _ = val.(uuid.UUID)
	}

	// 4. Invoke service business logic
	log, err := h.service.ClockIn(orgID, userID, req)
	if err != nil {
		if err == ErrActiveClockInExists {
			return common.Fail(c, fiber.StatusConflict, err.Error())
		}
		return common.Fail(c, fiber.StatusInternalServerError, "Failed to record clock-in")
	}

	// 5. Format response payload
	resp := AttendanceResponse{
		ID:         log.ID,
		UserID:     log.UserID,
		ClockIn:    log.ClockIn,
		ClockOut:   log.ClockOut,
		TotalHours: log.TotalHours,
		SyncStatus: log.SyncStatus,
		DeviceHash: log.DeviceHash,
		RecordUUID: log.RecordUUID,
	}

	return common.Created(c, resp)
}
