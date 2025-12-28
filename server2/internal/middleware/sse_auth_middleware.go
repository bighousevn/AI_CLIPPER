package middleware

import (
	"ai-clipper/server2/internal/auth/application"
	"net/http"

	"github.com/gin-gonic/gin"
)

// SSEAuthMiddleware creates a middleware to protect SSE routes with JWT authentication from cookies
func SSEAuthMiddleware(tokenGenerator application.TokenGenerator) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Try to get token from cookie
		tokenString, err := c.Cookie("refresh_token")
		if err != nil {
			tokenString = c.Query("refresh_token")
		}

		if tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Refresh token is missing in cookie or query"})
			return
		}

		// Validate the refresh token
		userID, err := tokenGenerator.ValidateRefreshToken(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			return
		}

		c.Set("userID", userID)
		c.Next()
	}
}
