package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/habeshan-rems/backend/internal/auth"
	"github.com/habeshan-rems/backend/internal/common"
	"github.com/habeshan-rems/backend/internal/dashboard"
	"github.com/habeshan-rems/backend/internal/notifications"
	"github.com/habeshan-rems/backend/internal/admin"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, relying on real environment variables")
	}

	db := common.InitDB()

	app := fiber.New(fiber.Config{
		AppName: "Habeshan REMS API v1",
	})

	app.Get("/api/v1/health", func(c *fiber.Ctx) error {
		return common.OK(c, fiber.Map{"status": "ok"})
	})

	// Register feature routes
	auth.RegisterRoutes(app, db)
	dashboard.RegisterRoutes(app)
	notifications.RegisterRoutes(app, db)
	admin.RegisterRoutes(app, db)

	port := getEnv("PORT", "8080")
	log.Printf("🚀 server listening on :%s", port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("server failed to start: %v", err)
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
