package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// CORS handles Cross-Origin Resource Sharing by injecting appropriate headers.
// It accepts an allowedOrigin wildcard or specific domain (e.g. http://localhost:3000).
func CORS(allowedOrigin string) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		
		// If the Origin header is present and matches the configured allowed origin
		if origin != "" && (allowedOrigin == "*" || origin == allowedOrigin) {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
			c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
			c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")
		}

		// Handle preflight OPTIONS requests immediately
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
