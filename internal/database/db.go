package database

import (
	"fmt"
	"log"

	"github.com/pubudulakmal/auth-api/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Connect opens a GORM PostgreSQL connection using the provided DSN and
// auto-migrates all registered models.
func Connect(dsn string) *gorm.DB {
	if dsn == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS pgcrypto").Error; err != nil {
		// pgcrypto is required for gen_random_uuid(); log but don't fatal —
		// the extension may already exist or the user may lack superuser rights.
		log.Printf("Warning: could not enable pgcrypto extension: %v", err)
	}

	if err := db.AutoMigrate(&models.User{}, &models.RefreshToken{}); err != nil {
		log.Fatalf("Auto-migration failed: %v", err)
	}

	fmt.Println("Database connected and migrations applied successfully")
	return db
}
