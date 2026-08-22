package main

import (
	"log"

	"github.com/habeshan-rems/backend/internal/common"
	"github.com/habeshan-rems/backend/migrations"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, relying on real environment variables")
	}

	db := common.InitDB()

	if err := migrations.RunAll(db); err != nil {
		log.Fatalf("migration failed: %v", err)
	}

	log.Println("✅ migrations complete")
}
