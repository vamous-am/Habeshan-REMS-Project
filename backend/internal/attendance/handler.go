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

// GetSelfHistory handles GET /attendance/me (Task 15)
func (h *Handler) GetSelfHistory(c *fiber.Ctx) error {
	var query AttendanceHistoryQuery
	if err := c.QueryParser(&query); err != nil {
		return common.Fail(c, fiber.StatusBadRequest, "Invalid query parameters")
	}

	var orgID, userID uuid.UUID
	if val := c.Locals("org_id"); val != nil {
		orgID, _ = val.(uuid.UUID)
	}
	if val := c.Locals("user_id"); val != nil {
		userID, _ = val.(uuid.UUID)
	}

	res, err := h.service.GetSelfHistory(orgID, userID, query)
	if err != nil {
		return common.Fail(c, fiber.StatusInternalServerError, "Failed to fetch attendance history")
	}
	return c.Status(fiber.StatusOK).JSON(res)
}

// GetTeamHistory handles GET /attendance/team (Task 15)
func (h *Handler) GetTeamHistory(c *fiber.Ctx) error {
	var query AttendanceHistoryQuery
	if err := c.QueryParser(&query); err != nil {
		return common.Fail(c, fiber.StatusBadRequest, "Invalid query parameters")
	}

	var orgID, managerID uuid.UUID
	if val := c.Locals("org_id"); val != nil {
		orgID, _ = val.(uuid.UUID)
	}
	if val := c.Locals("user_id"); val != nil {
		managerID, _ = val.(uuid.UUID)
	}

	res, err := h.service.GetTeamHistory(orgID, managerID, query)
	if err != nil {
		return common.Fail(c, fiber.StatusInternalServerError, "Failed to fetch team attendance history")
	}
	return c.Status(fiber.StatusOK).JSON(res)
}

// GetOrgHistory handles GET /attendance/org (Task 15)
func (h *Handler) GetOrgHistory(c *fiber.Ctx) error {
	var query AttendanceHistoryQuery
	if err := c.QueryParser(&query); err != nil {
		return common.Fail(c, fiber.StatusBadRequest, "Invalid query parameters")
	}

	var orgID uuid.UUID
	if val := c.Locals("org_id"); val != nil {
		orgID, _ = val.(uuid.UUID)
	}

	res, err := h.service.GetOrgHistory(orgID, query)
	if err != nil {
		return common.Fail(c, fiber.StatusInternalServerError, "Failed to fetch org attendance history")
	}
	return c.Status(fiber.StatusOK).JSON(res)
}
