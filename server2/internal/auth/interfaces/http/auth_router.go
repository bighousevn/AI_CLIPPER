package http

import (
	"ai-clipper/server2/internal/auth/application"
	"ai-clipper/server2/internal/middleware"

	"github.com/gin-gonic/gin"
)

// NewAuthRouter creates and registers all routes for the authentication service.
func NewAuthRouter(router *gin.Engine, controller *AuthController, tokenGenerator application.TokenGenerator) {
	v1 := router.Group("/api/v1")
	{
		auth := v1.Group("/auth")
		{
			auth.POST("/register", controller.Register)
			auth.POST("/login", controller.Login)
			auth.POST("/refresh-token", controller.RefreshToken)
			auth.POST("/forgot-password", controller.ForgotPassword)
			auth.POST("/reset-password", controller.ResetPassword)
			auth.GET("/verify-email", controller.VerifyEmail)
			auth.GET("/logout", controller.Logout)
		}

		authenticated := v1.Group("/")
		authenticated.Use(middleware.AuthMiddleware(tokenGenerator))
		{
			authenticated.GET("/users/me", controller.GetProfile)
			authenticated.POST("/users/me/password", controller.ChangePassword)
		}
	}
}
