package api

import (
	"bighousevn/be/models"
	"bighousevn/be/services"
	"bighousevn/be/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthController struct {
	authService services.AuthService
}

func NewAuthController(authService services.AuthService) *AuthController {
	return &AuthController{authService: authService}
}

func (ctrl *AuthController) Register(c *gin.Context) {
	var req models.RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse := utils.HandleValidationErrors(err)
		c.JSON(http.StatusBadRequest, errorResponse)
		return
	}

	user, err := ctrl.authService.Register(&req)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully. Please check your email to verify your account.", "user_id": user.ID})
}

func (ctrl *AuthController) Login(c *gin.Context) {
	var req models.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse := utils.HandleValidationErrors(err)
		c.JSON(http.StatusBadRequest, errorResponse)
		return
	}

	accessToken, refreshToken, err := ctrl.authService.Login(&req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.SetCookie("refresh_token", refreshToken, 3600*24*30, "/", "localhost", false, true)

	c.JSON(http.StatusOK, gin.H{"message": "Login successful", "access_token": accessToken})
}

func (ctrl *AuthController) RefreshToken(c *gin.Context) {
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Refresh token not found"})
		return
	}

	accessToken, newRefreshToken, err := ctrl.authService.RefreshToken(refreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.SetCookie("refresh_token", newRefreshToken, 3600*24*30, "/", "localhost", false, true)

	c.JSON(http.StatusOK, gin.H{"message": "Tokens refreshed successfully", "access_token": accessToken})
}

func (ctrl *AuthController) ForgotPassword(c *gin.Context) {
	var req models.ForgotPasswordRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse := utils.HandleValidationErrors(err)
		c.JSON(http.StatusBadRequest, errorResponse)
		return
	}

	if err := ctrl.authService.ForgotPassword(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "If an account with that email exists, a password reset link has been sent."})
}

func (ctrl *AuthController) ResetPassword(c *gin.Context) {
	var req models.ResetPasswordRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse := utils.HandleValidationErrors(err)
		c.JSON(http.StatusBadRequest, errorResponse)
		return
	}

	if err := ctrl.authService.ResetPassword(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password has been reset successfully."})
}

func (ctrl *AuthController) VerifyEmail(c *gin.Context) {
	var req models.VerifyEmailRequest

	if err := c.ShouldBindQuery(&req); err != nil {
		errorResponse := utils.HandleValidationErrors(err)
		c.JSON(http.StatusBadRequest, errorResponse)
		return
	}

	if err := ctrl.authService.VerifyEmail(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Email has been verified successfully."})
}

func (ctrl *AuthController) GetProfile(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found in context"})
		return
	}

	userModel, ok := user.(*models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not assert user type from context"})
		return
	}

	// Return user profile without sensitive information
	c.JSON(http.StatusOK, gin.H{
		"id":                 userModel.ID,
		"username":           userModel.Username,
		"email":              userModel.Email,
		"credits":            userModel.Credits,
		"stripe_customer_id": userModel.StripeCustomerID,
	})
}

func (ctrl *AuthController) ChangePassword(c *gin.Context) {
	var req models.ChangePasswordRequest

	userCtx, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found in context"})
		return
	}

	user, ok := userCtx.(*models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not assert user type from context"})
		return
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse := utils.HandleValidationErrors(err)
		c.JSON(http.StatusBadRequest, errorResponse)
		return
	}

	if err := ctrl.authService.ChangePassword(user, &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
}
