package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"

	"github.com/habeshan-rems/backend/internal/admin"
	"github.com/habeshan-rems/backend/internal/attendance"
	"github.com/habeshan-rems/backend/internal/auth"
	"github.com/habeshan-rems/backend/internal/common"
	"github.com/habeshan-rems/backend/internal/dashboard"
	"github.com/habeshan-rems/backend/internal/notifications"
	"github.com/habeshan-rems/backend/internal/timesheets"
)

func main() {
	// 1. Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, relying on environment variables")
	}

	// 2. Initialize Database connection
	db := common.InitDB()

	// 3. Initialize Fiber App
	app := fiber.New(fiber.Config{
		AppName: "Habeshan REMS API v1",
	})

	// 4. Configure CORS middleware (for Vite React frontend)
	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:5173, http://localhost:3000",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET, POST, HEAD, PUT, DELETE, PATCH",
	}))

	// 5. Health Check Endpoint
	app.Get("/api/v1/health", func(c *fiber.Ctx) error {
		return common.OK(c, fiber.Map{"status": "ok"})
	})

	// 6.Register feature routes
    attendance.RegisterRoutes(app, db)
	auth.RegisterRoutes(app, db)
	dashboard.RegisterRoutes(app, db)
	notifications.RegisterRoutes(app, db)
	timesheets.RegisterRoutes(app, db)
	admin.RegisterRoutes(app, db)

	// 7. Start HTTP Server
	port := getEnv("PORT", "8080")
	log.Printf("🚀 Server listening on port :%s", port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("❌ Server failed to start: %v", err)
	}
}

// Helper function to read environment variables with fallback
func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
