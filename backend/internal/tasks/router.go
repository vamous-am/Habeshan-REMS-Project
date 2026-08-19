package tasks

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// RegisterRoutes wires all Tasks HTTP endpoints to the Fiber app.
//
// Route layout:
//
//	POST   /tasks                          — create task (admin/manager)
//	GET    /tasks                          — list tasks (visibility-scoped)
//	GET    /tasks/stats/status-counts      — aggregate counts by status (admin/manager)
//	GET    /tasks/stats/overdue            — overdue task list (admin/manager)
//	GET    /tasks/:id                      — single task detail
//	PATCH  /tasks/:id/status               — change task status
//	POST   /tasks/:id/assignments          — assign employees to task
//	DELETE /tasks/:id/assignments/:userID  — remove employee from task
//	POST   /tasks/:id/timer/start          — start timer
//	POST   /tasks/:id/timer/pause          — pause timer (requires reason)
//	POST   /tasks/:id/timer/resume         — resume timer
//	POST   /tasks/:id/timer/stop           — stop timer
//	GET    /tasks/:id/timer                — timer history (with pause-reason redaction)
//
// The /stats/* routes are registered before /:id so that Fiber's router does not
// misinterpret "stats" as a task UUID parameter.
func RegisterRoutes(app *fiber.App, db *gorm.DB) {
	// Build dependency graph.
	taskRepo := NewTaskRepository(db)
	assignRepo := NewTaskAssignmentRepository(db)
	timeLogRepo := NewTaskTimeLogRepository(db)
	userRepo := NewUserRepository(db)
	orgRepo := NewOrganizationRepository(db)

	svc := NewService(taskRepo, assignRepo, timeLogRepo, userRepo, orgRepo)
	h := NewHandler(svc, assignRepo)

	tasks := app.Group("/tasks")

	// ── collection routes ─────────────────────────────────────────────────────
	tasks.Post("/", h.CreateTask)
	tasks.Get("/", h.GetMyTasks)

	// ── aggregate stats (must come before /:id) ───────────────────────────────
	stats := tasks.Group("/stats")
	stats.Get("/status-counts", h.GetStatusCounts)
	stats.Get("/overdue", h.GetOverdueTasks)

	// ── single-task routes ────────────────────────────────────────────────────
	tasks.Get("/:id", h.GetTaskByID)
	tasks.Patch("/:id/status", h.ChangeStatus)

	// ── assignment routes ─────────────────────────────────────────────────────
	tasks.Post("/:id/assignments", h.AssignTask)
	tasks.Delete("/:id/assignments/:userID", h.UnassignTask)

	// ── timer routes ──────────────────────────────────────────────────────────
	tasks.Post("/:id/timer/start", h.StartTimer)
	tasks.Post("/:id/timer/pause", h.PauseTimer)
	tasks.Post("/:id/timer/resume", h.ResumeTimer) // was incorrectly mapped to PauseTask
	tasks.Post("/:id/timer/stop", h.StopTimer)
	tasks.Get("/:id/timer", h.GetTimerHistory)
}
