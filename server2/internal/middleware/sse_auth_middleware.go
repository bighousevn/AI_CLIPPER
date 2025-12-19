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
		tokenString, err := c.Cookie("access_token")
		if err != nil {
			// If not in cookie, try query parameter as a fallback (useful for some SSE clients)
			tokenString = c.Query("access_token")
		}

		if tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Access token is missing in cookie or query"})
			return
		}

		// Validate the access token
		userID, err := tokenGenerator.ValidateAccessToken(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			return
		}

		c.Set("userID", userID)
		c.Next()
	}
}
