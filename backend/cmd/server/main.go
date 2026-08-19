package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/google/uuid"
	"github.com/habeshan-rems/backend/internal/attendance"
	"github.com/habeshan-rems/backend/internal/common"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, relying on real environment variables")
	}

	db := common.InitDB()

	// Automatically create the attendance_logs table in PostgreSQL
	if err := db.AutoMigrate(&attendance.AttendanceLog{}); err != nil {
		log.Fatalf("❌ Schema migration failed: %v", err)
	}
	log.Println("✅ Attendance database table ready")

	app := fiber.New()

	// Enable CORS for Vite frontend
	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:5173, http://localhost:3000",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET, POST, HEAD, PUT, DELETE, PATCH",
	}))

	// Temporary mock auth middleware until Dev 1 finishes JWT setup
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("org_id", uuid.MustParse("11111111-1111-1111-1111-111111111111"))
		c.Locals("user_id", uuid.MustParse("22222222-2222-2222-2222-222222222222"))
		return c.Next()
	})

	attendance.RegisterRoutes(app, db)

	log.Println("🚀 Server listening on http://localhost:8080")
	if err := app.Listen(":8080"); err != nil {
		log.Fatalf("❌ Server failed to start: %v", err)
	}
}