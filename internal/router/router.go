package router

import (
	"github.com/gin-gonic/gin"
	"github.com/pubudulakmal/auth-api/internal/handlers"
	"github.com/pubudulakmal/auth-api/internal/middleware"
	"github.com/pubudulakmal/auth-api/internal/services"
)

// Setup registers all routes on the provided Gin engine and returns it.
func Setup(
	authHandler *handlers.AuthHandler,
	userHandler *handlers.UserHandler,
	jwtService *services.JWTService,
	allowedOrigin string,
) *gin.Engine {
	r := gin.Default()

	// Apply CORS middleware
	r.Use(middleware.CORS(allowedOrigin))

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	v1 := r.Group("/api/v1")
	{
		// Public auth routes
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.Refresh)
		}

		// Protected routes — JWT middleware applied
		protected := v1.Group("")
		protected.Use(middleware.JWTAuth(jwtService))
		{
			protected.POST("/auth/logout", authHandler.Logout)

			users := protected.Group("/users")
			{
				users.GET("/me", userHandler.GetMe)
			}
		}
	}

	return r
}
