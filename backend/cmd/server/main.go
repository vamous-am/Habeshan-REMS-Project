package main

import (
	"log"
	"github.com/habeshan-rems/backend/internal/common"
	"github.com/joho/godotenv"
	
	// each feature package gets imported here as it's built
)

func main() {
	if err:= godotenv.Load(); err != nil {
		log.Println("no .env file found, relying on real environment variables")
	}
	db := common.InitDB()
	_ = db // each feature's router/handler setup will use this

	// app := fiber.New()
	// ... register middleware, routes per feature, app.Listen(":8080")
}