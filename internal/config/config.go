package config

import (
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all application configuration loaded from environment variables.
type Config struct {
	DatabaseURL      string
	JWTSecret        string
	AccessTokenTTL   time.Duration
	RefreshTokenTTL  time.Duration
	Port             string
}

// Load reads configuration from a .env file (if present) and environment variables.
func Load() *Config {
	// Load .env file — ignore error in production where env vars are set directly
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, reading from environment")
	}

	accessTTL, err := time.ParseDuration(getEnv("ACCESS_TOKEN_TTL", "15m"))
	if err != nil {
		log.Fatalf("Invalid ACCESS_TOKEN_TTL: %v", err)
	}

	refreshTTL, err := time.ParseDuration(getEnv("REFRESH_TOKEN_TTL", "168h"))
	if err != nil {
		log.Fatalf("Invalid REFRESH_TOKEN_TTL: %v", err)
	}

	return &Config{
		DatabaseURL:     getEnv("DATABASE_URL", ""),
		JWTSecret:       getEnv("JWT_SECRET", ""),
		AccessTokenTTL:  accessTTL,
		RefreshTokenTTL: refreshTTL,
		Port:            getEnv("PORT", "8080"),
	}
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}
