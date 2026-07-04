package database

import (
	"fmt"
	"log"
	"time"

	"github.com/pubudulakmal/auth-api/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Connect opens a GORM PostgreSQL connection using the provided DSN and
// auto-migrates all registered models.
// It retries the connection with exponential backoff for up to ~60 seconds
// to survive race conditions where the DB container starts after the API
// (common on Railway and other container platforms).
func Connect(dsn string) (*gorm.DB, error) {
	if dsn == "" {
		return nil, fmt.Errorf("DATABASE_URL is not set — add a PostgreSQL service in Railway and link it to this service")
	}

	var db *gorm.DB
	var err error

	maxAttempts := 8
	backoff := 2 * time.Second

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})
		if err == nil {
			// Verify the connection is actually alive
			sqlDB, pingErr := db.DB()
			if pingErr == nil {
				pingErr = sqlDB.Ping()
			}
			if pingErr == nil {
				break
			}
			err = pingErr
		}

		if attempt == maxAttempts {
			return nil, fmt.Errorf("could not connect to database after %d attempts: %v", maxAttempts, err)
		}

		log.Printf("DB not ready (attempt %d/%d): %v — retrying in %s...", attempt, maxAttempts, err, backoff)
		time.Sleep(backoff)
		if backoff < 16*time.Second {
			backoff *= 2
		}
	}

	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS pgcrypto").Error; err != nil {
		// pgcrypto is required for gen_random_uuid(); log but don't fatal —
		// the extension may already exist or the user may lack superuser rights.
		log.Printf("Warning: could not enable pgcrypto extension: %v", err)
	}

	if err := db.AutoMigrate(&models.User{}, &models.RefreshToken{}); err != nil {
		return nil, fmt.Errorf("auto-migration failed: %v", err)
	}

	fmt.Println("Database connected and migrations applied successfully")
	return db, nil
}
