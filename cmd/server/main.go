package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"sync"
	"syscall"

	"github.com/pubudulakmal/auth-api/internal/config"
	"github.com/pubudulakmal/auth-api/internal/database"
	"github.com/pubudulakmal/auth-api/internal/handlers"
	"github.com/pubudulakmal/auth-api/internal/repository"
	"github.com/pubudulakmal/auth-api/internal/router"
	"github.com/pubudulakmal/auth-api/internal/services"
)

// swappableHandler is a thread-safe http.Handler whose underlying handler
// can be replaced at runtime without restarting the server.
type swappableHandler struct {
	mu sync.RWMutex
	h  http.Handler
}

func (s *swappableHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	h := s.h
	s.mu.RUnlock()
	h.ServeHTTP(w, r)
}

func (s *swappableHandler) swap(h http.Handler) {
	s.mu.Lock()
	s.h = h
	s.mu.Unlock()
}

func main() {
	// 1. Load configuration
	cfg := config.Load()

	// 2. Boot a minimal HTTP server immediately so Railway's health check
	//    gets a 200 while we are still connecting to the database.
	//    The handler will be swapped to the full Gin router once the DB is ready.
	startupMux := http.NewServeMux()
	startupMux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"status":"starting"}`)
	})

	handler := &swappableHandler{h: startupMux}

	addr := fmt.Sprintf(":%s", cfg.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: handler,
	}

	go func() {
		log.Printf("HTTP server listening on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// 3. Validate variables and connect to database
	var initErr error

	if cfg.JWTSecret == "" {
		initErr = fmt.Errorf("JWT_SECRET environment variable is missing")
	}

	if initErr == nil {
		// This blocks and retries for ~60s
		db, err := database.Connect(cfg.DatabaseURL)
		if err != nil {
			initErr = fmt.Errorf("Database connection failed: %v", err)
		} else {
			// 4. Wire up repositories
			userRepo := repository.NewUserRepository(db)
			tokenRepo := repository.NewTokenRepository(db)
		
			// 5. Wire up services
			jwtService := services.NewJWTService(cfg.JWTSecret, cfg.AccessTokenTTL, cfg.RefreshTokenTTL)
			authService := services.NewAuthService(userRepo, tokenRepo, jwtService)
		
			// 6. Wire up handlers
			authHandler := handlers.NewAuthHandler(authService)
			userHandler := handlers.NewUserHandler(userRepo)
		
			// 7. Build the full Gin router and atomically swap it in
			fullRouter := router.Setup(authHandler, userHandler, jwtService)
			handler.swap(fullRouter)
		
			log.Println("Application fully initialized — all API routes are active")
		}
	} 
	
	if initErr != nil {
		// If initialization failed (e.g. missing JWT_SECRET), we swap in an error handler
		// so the server stays alive and Railway healthcheck can see what's wrong!
		errMux := http.NewServeMux()
		errMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/health" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				fmt.Fprintf(w, `{"status":"error","message":%q}`, initErr.Error())
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Server Configuration Error: %v\n", initErr)
		})
		handler.swap(errMux)
		log.Printf("Application failed to initialize: %v", initErr)
	}

	// 8. Block until OS interrupt (graceful shutdown)
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()

	log.Println("Shutting down gracefully…")
	if err := srv.Shutdown(context.Background()); err != nil {
		log.Printf("Shutdown error: %v", err)
	}
}
