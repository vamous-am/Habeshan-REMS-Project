package common

import (
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitDB() *gorm.DB {
	
	// 1. Fetch variables from your .env file
    user := os.Getenv("POSTGRES_USER")
    password := os.Getenv("POSTGRES_PASSWORD")
    dbName := os.Getenv("POSTGRES_DB")
    port := os.Getenv("DB_PORT")
  
  // Default host to localhost for local development
  host := "localhost" 

    // 2. Validate BEFORE building the DSN — this is the real guard,
	// replacing the old dead "dsn == ''" check which could never fire.
	if user == "" || password == "" || dbName == "" || port == "" {
		log.Fatal("one or more required env vars are missing (POSTGRES_USER, POSTGRES_PASSWORD, POSTGRES_DB, DB_PORT) — copy .env.example to .env and fill it in")
	}

  // 3. Format into a standard connection URL
    dsn := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=disable", 
    user, password, host, port, dbName,
  )

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn), 
	})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("failed to get underlying sql.DB: %v", err)
	}

	// Pool settings — fine defaults for a small team's staging/dev DB.
	// Revisit only if someone hits real connection limits.
	sqlDB.SetMaxOpenConns(20)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)

	if err := sqlDB.Ping(); err != nil {
		log.Fatalf("database ping failed: %v", err)
	}

	fmt.Println("✅ Connected to database")
	DB = db
	return db
}