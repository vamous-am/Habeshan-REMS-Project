package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"

	"github.com/habeshan-rems/backend/internal/admin"
	"github.com/habeshan-rems/backend/internal/attendance"
	"github.com/habeshan-rems/backend/internal/auth"
	"github.com/habeshan-rems/backend/internal/common"
	"github.com/habeshan-rems/backend/internal/dashboard"
	"github.com/habeshan-rems/backend/internal/notifications"
	"github.com/habeshan-rems/backend/internal/tasks"
	"github.com/habeshan-rems/backend/internal/timesheets"
	"github.com/habeshan-rems/backend/migrations"
)

func main() {
	// 1. Root application context with cancellation on SIGINT/SIGTERM
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// 2. Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, relying on environment variables")
	}

	// 3. Initialize Database connection and run migrations
	db := common.InitDB()
	if err := migrations.RunAll(db); err != nil {
		log.Fatalf("❌ migrations failed: %v", err)
	}

	// 4. Start background schedulers (timesheets & clock-in reminders)
	timesheets.StartScheduler(ctx, db)
	notifications.StartClockInReminderScheduler(ctx, db)

	// 5. Initialize Fiber App
	app := fiber.New(fiber.Config{
		AppName: "Habeshan REMS API v1",
	})
	app.Use(func(c *fiber.Ctx) error {
		start := time.Now()

		err := c.Next()

		slog.Info("HTTP Request",
			slog.String("method", c.Method()),
			slog.String("path", c.Path()),
			slog.Int("status", c.Response().StatusCode()),
			slog.Duration("latency", time.Since(start)),
		)
		return err
	})

	// 6. Configure CORS middleware (for Vite React frontend)
	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:5173, http://localhost:3000",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization, X-User-ID, X-Org-ID",
		AllowMethods: "GET, POST, HEAD, PUT, DELETE, PATCH",
	}))

	// 7. Health Check Endpoint
	app.Get("/api/v1/health", func(c *fiber.Ctx) error {
		return common.OK(c, fiber.Map{"status": "ok"})
	})

	// 8. Register feature routes
	attendance.RegisterRoutes(app, db)
	auth.RegisterRoutes(app, db)
	dashboard.RegisterRoutes(app, db)
	notifications.RegisterRoutes(app, db)
	tasks.RegisterRoutes(app, db)
	timesheets.RegisterRoutes(app, db)
	admin.RegisterRoutes(app, db)

	// 9. Start HTTP Server with graceful shutdown
	port := getEnv("PORT", "8080")
	serverErr := make(chan error, 1)

	go func() {
		log.Printf("🚀 Server listening on port :%s", port)
		if err := app.Listen(":" + port); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	select {
	case err := <-serverErr:
		log.Printf("❌ Server error: %v", err)
	case <-ctx.Done():
		log.Println("🛑 Shutting down server gracefully...")
		if err := app.ShutdownWithTimeout(5 * time.Second); err != nil {
			log.Printf("⚠️ Server shutdown error: %v", err)
		}
	}
}

// Helper function to read environment variables with fallback
func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
