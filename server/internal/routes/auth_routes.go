package routes

import (
	"bighousevn/be/internal/handler"
	"bighousevn/be/internal/handler/middleware"
	"bighousevn/be/internal/repository"

	"github.com/gin-gonic/gin"
)

func SetupAuthRoutes(r *gin.Engine, authController *handler.AuthController, authRepo repository.AuthRepository) {
	v1 := r.Group("/api/v1")
	{
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authController.Register)
			auth.POST("/login", authController.Login)
			auth.POST("/refresh-token", authController.RefreshToken)
			auth.POST("/forgot-password", authController.ForgotPassword)
			auth.POST("/reset-password", authController.ResetPassword)
			auth.GET("/verify-email", authController.VerifyEmail)
			auth.GET("/logout", authController.Logout)
		}

		authenticated := v1.Group("/")
		authenticated.Use(middleware.AuthMiddleware(authRepo))
		{
			authenticated.GET("/users/me", authController.GetProfile)
			authenticated.POST("/users/me/password", authController.ChangePassword)

		}
	}
}