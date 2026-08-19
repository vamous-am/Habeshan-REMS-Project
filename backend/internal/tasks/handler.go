package tasks

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/habeshan-rems/backend/internal/common"
)

// ─── auth context helpers ─────────────────────────────────────────────────────
//
// Until the real auth middleware (Dev 1) is wired in, the caller's identity is
// read from request headers that the middleware will eventually set.
//
// Header contract (to be set by JWT middleware):
//   X-User-ID  — authenticated user's UUID
//   X-Org-ID   — authenticated user's organisation UUID
//
// This keeps the handler code stable: when the middleware is added it only
// needs to populate these two headers; no handler changes are required.

func callerFromCtx(c *fiber.Ctx) (callerID, orgID uuid.UUID, err error) {
	rawUser := c.Get("X-User-ID")
	rawOrg := c.Get("X-Org-ID")

	if rawUser == "" || rawOrg == "" {
		return uuid.Nil, uuid.Nil, common.ErrUnauthorized
	}

	callerID, err = uuid.Parse(rawUser)
	if err != nil {
		return uuid.Nil, uuid.Nil, common.ErrUnauthorized
	}

	orgID, err = uuid.Parse(rawOrg)
	if err != nil {
		return uuid.Nil, uuid.Nil, common.ErrUnauthorized
	}

	return callerID, orgID, nil
}

// parseTaskID parses ":id" from the route parameter.
func parseTaskID(c *fiber.Ctx) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return uuid.Nil, common.ErrBadRequest
	}
	return id, nil
}

// ─── Handler ──────────────────────────────────────────────────────────────────

// Handler holds all HTTP handlers for the tasks feature.
// It depends on TaskService and TaskAssignmentRepository (for assignment
// queries that the service exposes through the repo layer).
type Handler struct {
	svc        TaskService
	assignRepo TaskAssignmentRepository
}

func NewHandler(svc TaskService, assignRepo TaskAssignmentRepository) *Handler {
	return &Handler{svc: svc, assignRepo: assignRepo}
}

// ─── Task CRUD ────────────────────────────────────────────────────────────────

// CreateTask  POST /tasks
// FR-TASK-01: manager or admin creates a task.
func (h *Handler) CreateTask(c *fiber.Ctx) error {
	callerID, orgID, err := callerFromCtx(c)
	if err != nil {
		return common.HandleError(c, err)
	}

	var req CreateTaskRequest
	if err := c.BodyParser(&req); err != nil {
		return common.Fail(c, fiber.StatusBadRequest, "invalid request body")
	}
	req.OrgID = orgID // always taken from auth context, never trusted from body

	task, err := h.svc.CreateTask(req, callerID)
	if err != nil {
		return common.HandleError(c, err)
	}

	assignedTo, _ := h.assignRepo.GetAssignedUserIDs(task.ID.ID)
	return common.Created(c, TaskToResponse(task, assignedTo))
}

// GetMyTasks  GET /tasks
// FR-TASK-03: visibility-scoped task list for the authenticated caller.
func (h *Handler) GetMyTasks(c *fiber.Ctx) error {
	callerID, orgID, err := callerFromCtx(c)
	if err != nil {
		return common.HandleError(c, err)
	}

	tasks, err := h.svc.GetMyTasks(callerID, orgID)
	if err != nil {
		return common.HandleError(c, err)
	}

	resp := make([]TaskResponse, 0, len(tasks))
	for _, t := range tasks {
		resp = append(resp, TaskToResponse(t, nil))
	}
	return common.OK(c, resp)
}

// GetTaskByID  GET /tasks/:id
// Returns a single task if the caller is permitted to see it.
func (h *Handler) GetTaskByID(c *fiber.Ctx) error {
	callerID, orgID, err := callerFromCtx(c)
	if err != nil {
		return common.HandleError(c, err)
	}

	taskID, err := parseTaskID(c)
	if err != nil {
		return common.HandleError(c, err)
	}

	task, err := h.svc.GetTaskByID(taskID, callerID, orgID)
	if err != nil {
		return common.HandleError(c, err)
	}

	assignedTo, _ := h.assignRepo.GetAssignedUserIDs(task.ID.ID)
	return common.OK(c, TaskToResponse(task, assignedTo))
}

// ─── Assignment ───────────────────────────────────────────────────────────────

// AssignTask  POST /tasks/:id/assignments
// FR-TASK-02: assign one or more employees to a task.
func (h *Handler) AssignTask(c *fiber.Ctx) error {
	callerID, orgID, err := callerFromCtx(c)
	if err != nil {
		return common.HandleError(c, err)
	}

	taskID, err := parseTaskID(c)
	if err != nil {
		return common.HandleError(c, err)
	}

	var req AssignTaskRequest
	if err := c.BodyParser(&req); err != nil {
		return common.Fail(c, fiber.StatusBadRequest, "invalid request body")
	}

	if err := h.svc.AssignTask(taskID, req.UserIDs, callerID, orgID); err != nil {
		return common.HandleError(c, err)
	}

	assignedTo, _ := h.assignRepo.GetAssignedUserIDs(taskID)
	return common.OK(c, fiber.Map{"task_id": taskID, "assigned_to": assignedTo})
}

// UnassignTask  DELETE /tasks/:id/assignments/:userID
// FR-TASK-02: remove an employee from a task.
func (h *Handler) UnassignTask(c *fiber.Ctx) error {
	callerID, orgID, err := callerFromCtx(c)
	if err != nil {
		return common.HandleError(c, err)
	}

	taskID, err := parseTaskID(c)
	if err != nil {
		return common.HandleError(c, err)
	}

	targetUserID, err := uuid.Parse(c.Params("userID"))
	if err != nil {
		return common.Fail(c, fiber.StatusBadRequest, "invalid user id")
	}

	if err := h.svc.UnassignTask(taskID, targetUserID, callerID, orgID); err != nil {
		return common.HandleError(c, err)
	}

	return common.OK(c, fiber.Map{"message": "assignment removed"})
}

// ─── Status ───────────────────────────────────────────────────────────────────

// ChangeStatus  PATCH /tasks/:id/status
// FR-TASK-04: update task status with transition validation.
func (h *Handler) ChangeStatus(c *fiber.Ctx) error {
	callerID, orgID, err := callerFromCtx(c)
	if err != nil {
		return common.HandleError(c, err)
	}

	taskID, err := parseTaskID(c)
	if err != nil {
		return common.HandleError(c, err)
	}

	var req ChangeStatusRequest
	if err := c.BodyParser(&req); err != nil {
		return common.Fail(c, fiber.StatusBadRequest, "invalid request body")
	}

	if err := h.svc.UpdateTaskStatus(taskID, req.Status, callerID, orgID); err != nil {
		return common.HandleError(c, err)
	}

	return common.OK(c, fiber.Map{"task_id": taskID, "status": req.Status})
}

// ─── Timer ────────────────────────────────────────────────────────────────────

// StartTimer  POST /tasks/:id/timer/start
// FR-TASK-05: open a new timer segment.
func (h *Handler) StartTimer(c *fiber.Ctx) error {
	callerID, orgID, err := callerFromCtx(c)
	if err != nil {
		return common.HandleError(c, err)
	}

	taskID, err := parseTaskID(c)
	if err != nil {
		return common.HandleError(c, err)
	}

	var req TimerStartRequest
	if err := c.BodyParser(&req); err != nil {
		return common.Fail(c, fiber.StatusBadRequest, "invalid request body")
	}
	req.TaskID = taskID

	if req.DeviceHash == "" {
		return common.Fail(c, fiber.StatusBadRequest, "device_hash is required")
	}
	if req.RecordUUID == uuid.Nil {
		return common.Fail(c, fiber.StatusBadRequest, "record_uuid is required")
	}
	if req.SyncStatus == "" {
		req.SyncStatus = SyncPendingSync
	}

	log, err := h.svc.StartTimer(req, callerID, orgID)
	if err != nil {
		return common.HandleError(c, err)
	}

	return common.Created(c, TimeLogToResponse(log))
}

// PauseTimer  POST /tasks/:id/timer/pause
// FR-TASK-05 / FR-TASK-06: pause with mandatory reason.
func (h *Handler) PauseTimer(c *fiber.Ctx) error {
	callerID, orgID, err := callerFromCtx(c)
	if err != nil {
		return common.HandleError(c, err)
	}

	taskID, err := parseTaskID(c)
	if err != nil {
		return common.HandleError(c, err)
	}

	var req TimerPauseRequest
	if err := c.BodyParser(&req); err != nil {
		return common.Fail(c, fiber.StatusBadRequest, "invalid request body")
	}
	req.TaskID = taskID

	if req.DeviceHash == "" {
		return common.Fail(c, fiber.StatusBadRequest, "device_hash is required")
	}
	if req.RecordUUID == uuid.Nil {
		return common.Fail(c, fiber.StatusBadRequest, "record_uuid is required")
	}
	if req.SyncStatus == "" {
		req.SyncStatus = SyncPendingSync
	}

	log, err := h.svc.PauseTimer(req, callerID, orgID)
	if err != nil {
		return common.HandleError(c, err)
	}

	return common.OK(c, TimeLogToResponse(log))
}

// ResumeTimer  POST /tasks/:id/timer/resume
// FR-TASK-05: open a new segment after a pause.
func (h *Handler) ResumeTimer(c *fiber.Ctx) error {
	callerID, orgID, err := callerFromCtx(c)
	if err != nil {
		return common.HandleError(c, err)
	}

	taskID, err := parseTaskID(c)
	if err != nil {
		return common.HandleError(c, err)
	}

	var req TimerResumeRequest
	if err := c.BodyParser(&req); err != nil {
		return common.Fail(c, fiber.StatusBadRequest, "invalid request body")
	}
	req.TaskID = taskID

	if req.DeviceHash == "" {
		return common.Fail(c, fiber.StatusBadRequest, "device_hash is required")
	}
	if req.RecordUUID == uuid.Nil {
		return common.Fail(c, fiber.StatusBadRequest, "record_uuid is required")
	}
	if req.SyncStatus == "" {
		req.SyncStatus = SyncPendingSync
	}

	log, err := h.svc.ResumeTimer(req, callerID, orgID)
	if err != nil {
		return common.HandleError(c, err)
	}

	return common.Created(c, TimeLogToResponse(log))
}

// StopTimer  POST /tasks/:id/timer/stop
// FR-TASK-05: close the active segment.
func (h *Handler) StopTimer(c *fiber.Ctx) error {
	callerID, orgID, err := callerFromCtx(c)
	if err != nil {
		return common.HandleError(c, err)
	}

	taskID, err := parseTaskID(c)
	if err != nil {
		return common.HandleError(c, err)
	}

	var req TimerStopRequest
	if err := c.BodyParser(&req); err != nil {
		return common.Fail(c, fiber.StatusBadRequest, "invalid request body")
	}
	req.TaskID = taskID

	if req.DeviceHash == "" {
		return common.Fail(c, fiber.StatusBadRequest, "device_hash is required")
	}
	if req.RecordUUID == uuid.Nil {
		return common.Fail(c, fiber.StatusBadRequest, "record_uuid is required")
	}
	if req.SyncStatus == "" {
		req.SyncStatus = SyncPendingSync
	}

	log, err := h.svc.StopTimer(req, callerID, orgID)
	if err != nil {
		return common.HandleError(c, err)
	}

	return common.OK(c, TimeLogToResponse(log))
}

// ─── Timer history ────────────────────────────────────────────────────────────

// GetTimerHistory  GET /tasks/:id/timer
// FR-TASK-07 / FR-TASK-08: returns timer log with pause-reason redaction.
func (h *Handler) GetTimerHistory(c *fiber.Ctx) error {
	callerID, orgID, err := callerFromCtx(c)
	if err != nil {
		return common.HandleError(c, err)
	}

	taskID, err := parseTaskID(c)
	if err != nil {
		return common.HandleError(c, err)
	}

	logs, err := h.svc.GetTimerHistory(taskID, callerID, orgID)
	if err != nil {
		return common.HandleError(c, err)
	}

	resp := make([]TimeLogResponse, 0, len(logs))
	for _, l := range logs {
		resp = append(resp, TimeLogToResponse(l))
	}
	return common.OK(c, resp)
}

// ─── Aggregates ───────────────────────────────────────────────────────────────

// GetStatusCounts  GET /tasks/stats/status-counts
// FR-TASK-09: task counts grouped by status.
func (h *Handler) GetStatusCounts(c *fiber.Ctx) error {
	callerID, orgID, err := callerFromCtx(c)
	if err != nil {
		return common.HandleError(c, err)
	}

	counts, err := h.svc.GetTaskStatusCounts(callerID, orgID)
	if err != nil {
		return common.HandleError(c, err)
	}

	return common.OK(c, StatusCountsToResponse(counts))
}

// GetOverdueTasks  GET /tasks/stats/overdue
// FR-TASK-09: list of overdue tasks.
func (h *Handler) GetOverdueTasks(c *fiber.Ctx) error {
	callerID, orgID, err := callerFromCtx(c)
	if err != nil {
		return common.HandleError(c, err)
	}

	tasks, err := h.svc.GetOverdueTasks(callerID, orgID)
	if err != nil {
		return common.HandleError(c, err)
	}

	resp := make([]TaskResponse, 0, len(tasks))
	for _, t := range tasks {
		resp = append(resp, TaskToResponse(t, nil))
	}
	return common.OK(c, resp)
}

// ─── compile-time guard ───────────────────────────────────────────────────────

// Ensure HandleError is reachable even if only used via common.HandleError.
var _ = errors.New
