package attendance

import (
	"fmt"
	"errors"
	"log"
	"time"

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

	attLog, err := h.service.ClockIn(orgID, userID, req)
	if err != nil {
		log.Println("❌ CLOCK-IN ERROR DETAILS:", err)

		if errors.Is(err, ErrActiveClockInExists) {
			return common.Fail(c, fiber.StatusConflict, err.Error())
		}
		return common.Fail(c, fiber.StatusInternalServerError, "Failed to record clock-in")
	}

	resp := AttendanceResponse{
		ID:         attLog.ID.ID,
		UserID:     attLog.UserID,
		ClockIn:    attLog.ClockIn,
		ClockOut:   attLog.ClockOut,
		TotalHours: attLog.TotalHours,
		SyncStatus: attLog.SyncStatus,
		DeviceHash: attLog.DeviceHash,
		RecordUUID: attLog.RecordUUID,
	}

	return common.Created(c, resp)
}

func (h *Handler) ClockOut(c *fiber.Ctx) error {
	var req ClockOutRequest

	if err := c.BodyParser(&req); err != nil {
		return common.Fail(c, fiber.StatusBadRequest, "Invalid request body")
	}

	var orgID, userID uuid.UUID
	if val := c.Locals("org_id"); val != nil {
		orgID, _ = val.(uuid.UUID)
	}
	if val := c.Locals("user_id"); val != nil {
		userID, _ = val.(uuid.UUID)
	}

	attLog, err := h.service.ClockOut(orgID, userID, req)
	if err != nil {
		log.Println("❌ CLOCK-OUT ERROR DETAILS:", err)

		if errors.Is(err, ErrNoActiveClockIn) {
			return common.Fail(c, fiber.StatusNotFound, err.Error())
		}
		return common.Fail(c, fiber.StatusInternalServerError, "Failed to record clock-out")
	}

	resp := AttendanceResponse{
		ID:         attLog.ID.ID,
		UserID:     attLog.UserID,
		ClockIn:    attLog.ClockIn,
		ClockOut:   attLog.ClockOut,
		TotalHours: attLog.TotalHours,
		SyncStatus: attLog.SyncStatus,
		DeviceHash: attLog.DeviceHash,
		RecordUUID: attLog.RecordUUID,
	}

	return common.Created(c, resp)
}

// SyncBatch handles POST /attendance/sync
func (h *Handler) SyncBatch(c *fiber.Ctx) error {
	var batchReq BatchSyncRequest

	// Support both batch array payload `{ records: [...] }` and single object fallback
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
		log.Println("❌ BATCH SYNC ERROR DETAILS:", err)
		return common.Fail(c, fiber.StatusInternalServerError, "Batch sync processing failed")
	}

	return c.Status(fiber.StatusOK).JSON(res)
}

// ExportLogs handles Task 13: GET /attendance/export
func (h *Handler) ExportLogs(c *fiber.Ctx) error {
	var orgID uuid.UUID
	if val := c.Locals("org_id"); val != nil {
		orgID, _ = val.(uuid.UUID)
	}

	// Filter query options
	var filter UserExportFilter
	if targetUser := c.Query("user_id"); targetUser != "" {
		if uUUID, err := uuid.Parse(targetUser); err == nil {
			filter.UserID = &uUUID
		}
	}
	if start := c.Query("start_date"); start != "" {
		if stTime, err := time.Parse("2006-01-02", start); err == nil {
			filter.StartDate = &stTime
		}
	}
	if end := c.Query("end_date"); end != "" {
		if endTime, err := time.Parse("2006-01-02", end); err == nil {
			// Extend to end-of-day boundary
			endTime = endTime.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			filter.EndDate = &endTime
		}
	}

	csvData, err := h.service.ExportLogs(orgID, filter)
	if err != nil {
		log.Println("❌ EXPORT ATTENDANCE ERROR:", err)
		return common.Fail(c, fiber.StatusInternalServerError, "Failed to generate attendance report")
	}

	fileName := fmt.Sprintf("attendance_report_%s.csv", time.Now().Format("20060102_150405"))

	c.Set("Content-Type", "text/csv")
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileName))
	return c.Status(fiber.StatusOK).Send(csvData)
}