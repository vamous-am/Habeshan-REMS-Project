package tasks

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"github.com/habeshan-rems/backend/internal/middleware"
)

// RegisterRoutes wires all Tasks HTTP endpoints to the Fiber app.
//
// Route layout (all under /api/v1 — see Group call below):
//
//	POST   /api/v1/tasks                          — create task (admin/manager)
//	GET    /api/v1/tasks                          — list tasks (visibility-scoped)
//	GET    /api/v1/tasks/stats/status-counts      — aggregate counts by status (admin/manager)
//	GET    /api/v1/tasks/stats/overdue            — overdue task list (admin/manager)
//	GET    /api/v1/tasks/:id                      — single task detail
//	PATCH  /api/v1/tasks/:id/status               — change task status
//	POST   /api/v1/tasks/:id/assignments          — assign employees to task
//	DELETE /api/v1/tasks/:id/assignments/:userID  — remove employee from task
//	POST   /api/v1/tasks/:id/timer/start          — start timer
//	POST   /api/v1/tasks/:id/timer/pause          — pause timer (requires reason)
//	POST   /api/v1/tasks/:id/timer/resume         — resume timer
//	POST   /api/v1/tasks/:id/timer/stop           — stop timer
//	GET    /api/v1/tasks/:id/timer                — timer history (with pause-reason redaction)
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

	// All API routes live under /api/v1 to match the health-check convention
	// in cmd/server/main.go (/api/v1/health) and the frontend
	// VITE_API_BASE_URL default.
	tasks := app.Group("/api/v1/tasks")

	// ── collection routes ─────────────────────────────────────────────────────
	tasks.Post("/", middleware.JWTAuth(), middleware.RequireRole("admin", "manager"), h.CreateTask)
	tasks.Get("/", middleware.JWTAuth(), h.GetMyTasks)

	// ── aggregate stats (must come before /:id) ───────────────────────────────
	stats := tasks.Group("/stats")
	stats.Get("/status-counts", middleware.JWTAuth(), h.GetStatusCounts)
	stats.Get("/overdue", middleware.JWTAuth(), h.GetOverdueTasks)

	// ── single-task routes ────────────────────────────────────────────────────
	tasks.Get("/:id", middleware.JWTAuth(), h.GetTaskByID)
	tasks.Patch("/:id/status", middleware.JWTAuth(), h.ChangeStatus)

	// ── assignment routes ─────────────────────────────────────────────────────
	tasks.Post("/:id/assignments", middleware.JWTAuth(), middleware.RequireRole("admin", "manager"), h.AssignTask)
	tasks.Delete("/:id/assignments/:userID", middleware.JWTAuth(), middleware.RequireRole("admin", "manager"), h.UnassignTask)

	// ── timer routes ──────────────────────────────────────────────────────────
	tasks.Post("/:id/timer/start", middleware.JWTAuth(), h.StartTimer)
	tasks.Post("/:id/timer/pause", middleware.JWTAuth(), h.PauseTimer)
	tasks.Post("/:id/timer/resume", middleware.JWTAuth(), h.ResumeTimer)
	tasks.Post("/:id/timer/stop", middleware.JWTAuth(), h.StopTimer)
	tasks.Get("/:id/timer", middleware.JWTAuth(), h.GetTimerHistory)
}
