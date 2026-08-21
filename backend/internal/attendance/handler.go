package attendance

import (
	"log"

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

func (h *Handler) ClockIn(c *fiber.Ctx) error {
	var req ClockInRequest
	if err := c.BodyParser(&req); err != nil {
		return common.Fail(c, fiber.StatusBadRequest, "Invalid request body")
	}
	if req.RecordUUID == uuid.Nil || req.DeviceHash == "" {
		return common.Fail(c, fiber.StatusBadRequest, "record_uuid and device_hash are required")
	}
	var orgID, userID uuid.UUID
	if val := c.Locals("org_id"); val != nil {
		orgID, _ = val.(uuid.UUID)
	}
	if val := c.Locals("user_id"); val != nil {
		userID, _ = val.(uuid.UUID)
	}

	res, err := h.service.ClockIn(orgID, userID, req)
	if err != nil {
		return common.Fail(c, fiber.StatusBadRequest, err.Error())
	}
	return c.Status(fiber.StatusOK).JSON(res)
}

func (h *Handler) ClockOut(c *fiber.Ctx) error {
	var req ClockOutRequest
	if err := c.BodyParser(&req); err != nil {
		return common.Fail(c, fiber.StatusBadRequest, "Invalid request body")
	}
	if req.RecordUUID == uuid.Nil || req.DeviceHash == "" {
		return common.Fail(c, fiber.StatusBadRequest, "record_uuid and device_hash are required")
	}
	var orgID, userID uuid.UUID
	if val := c.Locals("org_id"); val != nil {
		orgID, _ = val.(uuid.UUID)
	}
	if val := c.Locals("user_id"); val != nil {
		userID, _ = val.(uuid.UUID)
	}

	res, err := h.service.ClockOut(orgID, userID, req)
	if err != nil {
		return common.Fail(c, fiber.StatusBadRequest, err.Error())
	}
	return c.Status(fiber.StatusOK).JSON(res)
}

func (h *Handler) SyncBatch(c *fiber.Ctx) error {
	var batchReq BatchSyncRequest

	if err := c.BodyParser(&batchReq); err != nil || len(batchReq.Records) == 0 {
		var singleRecord SyncRecordRequest
		// Fixed compiler type mismatch check (RecordUUID is a string field)
		if errSingle := c.BodyParser(&singleRecord); errSingle == nil && singleRecord.RecordUUID != "" {
			batchReq.Records = []SyncRecordRequest{singleRecord}
		} else {
			return common.Fail(c, fiber.StatusBadRequest, "Invalid or empty sync batch payload")
		}
	}

	var orgID uuid.UUID
	if val := c.Locals("org_id"); val != nil {
		orgID, _ = val.(uuid.UUID)
	}

	res, err := h.service.SyncBatch(orgID, batchReq.Records)
	if err != nil {
		log.Println("❌ BATCH SYNC ERROR:", err)
		return common.Fail(c, fiber.StatusInternalServerError, "Batch sync processing failed")
	}

	return c.Status(fiber.StatusOK).JSON(res)
}