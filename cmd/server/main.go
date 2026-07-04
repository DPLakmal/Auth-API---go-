package main

import (
	"fmt"
	"log"

	"github.com/pubudulakmal/auth-api/internal/config"
	"github.com/pubudulakmal/auth-api/internal/database"
	"github.com/pubudulakmal/auth-api/internal/handlers"
	"github.com/pubudulakmal/auth-api/internal/repository"
	"github.com/pubudulakmal/auth-api/internal/router"
	"github.com/pubudulakmal/auth-api/internal/services"
)

func main() {
	// 1. Load configuration
	cfg := config.Load()

	if cfg.JWTSecret == "" {
		log.Fatal("JWT_SECRET must be set")
	}

	// 2. Connect to database and run migrations
	db := database.Connect(cfg.DatabaseURL)

	// 3. Wire up repositories
	userRepo := repository.NewUserRepository(db)
	tokenRepo := repository.NewTokenRepository(db)

	// 4. Wire up services
	jwtService := services.NewJWTService(cfg.JWTSecret, cfg.AccessTokenTTL, cfg.RefreshTokenTTL)
	authService := services.NewAuthService(userRepo, tokenRepo, jwtService)

	// 5. Wire up handlers
	authHandler := handlers.NewAuthHandler(authService)
	userHandler := handlers.NewUserHandler(userRepo)

	// 6. Set up router
	r := router.Setup(authHandler, userHandler, jwtService)

	// 7. Start server
	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("Server starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
