package api

import (
	"bighousevn/be/db"
	"bighousevn/be/models"
	"bighousevn/be/utils"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetProfile gets the currently logged-in user's profile
func GetProfile(c *gin.Context) {
	// The user is attached to the context by the AuthMiddleware
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found in context"})
		return
	}

	// Type assertion
	userModel, ok := user.(*models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not assert user type from context"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":       userModel.ID,
		"username": userModel.Username,
		"email":    userModel.Email,
	})
}

// ChangePassword handles password changes for authenticated users.
func ChangePassword(c *gin.Context) {
	var req models.ChangePasswordRequest

	// The user is attached to the context by the AuthMiddleware
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

	// Check if the old password is correct
	if !utils.CheckPasswordHash(req.OldPassword, user.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid old password"})
		return
	}

	// Hash the new password
	newHashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash new password"})
		return
	}

	// Update password
	user.Password = newHashedPassword
	if err := db.UpdateUser(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
}

func Register(c *gin.Context) {
	var req models.RegisterRequest


	if err := c.ShouldBindJSON(&req); err != nil {

		errorResponse := utils.HandleValidationErrors(err)
		c.JSON(http.StatusBadRequest, errorResponse)
		return
	}

	// Check if user with the same email already exists
	_, err := db.GetUserByEmail(req.Email)
	if err == nil {
	
		c.JSON(http.StatusConflict, gin.H{"error": "User with this email already exists"})
		return
	}

	
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error while checking for user"})
		return
	}


	// Hash the password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Generate email verification token
	verificationToken, err := utils.GenerateRandomToken(32)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate verification token"})
		return
	}
	hashedVerificationToken, err := utils.HashPassword(verificationToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash verification token"})
		return
	}

	verificationExpires := time.Now().Add(time.Hour * 24) // Token valid for 24 hours

	// Create a new user model
	newUser := models.User{
		Username: req.Username,
		Email:    req.Email,
		Password: hashedPassword, // This is the hash
		EmailVerificationToken: &hashedVerificationToken,
		EmailVerificationExpires: &verificationExpires,
	}

	// Save the new user to the database
	if err := db.CreateUser(&newUser); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// Send verification email
	verificationLink := fmt.Sprintf("http://localhost:3000/verify-email?token=%s", verificationToken)
	emailBody := fmt.Sprintf("<h1>Welcome!</h1><p>Please verify your email by clicking this link: <a href=\"%s\">%s</a></p>", verificationLink, verificationLink)
	go func() {
		if err := utils.SendEmail(newUser.Email, "Verify Your Email", emailBody); err != nil {
			log.Printf("Failed to send verification email to %s: %v", newUser.Email, err)
		}
	}()

	c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully. Please check your email to verify your account.", "user_id": newUser.ID})
}


func Login(c *gin.Context) {
	var req models.LoginRequest

	// Bind and validate the request body
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse := utils.HandleValidationErrors(err)
		c.JSON(http.StatusBadRequest, errorResponse)
		return
	}

	// Find the user by email
	user, err := db.GetUserByEmail(req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Check if the password is correct
	if !utils.CheckPasswordHash(req.Password, user.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	// Generate JWT
	token, err := utils.GenerateJWT(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Login successful", "token": token})
}

func ForgotPassword(c *gin.Context) {
	var req models.ForgotPasswordRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse := utils.HandleValidationErrors(err)
		c.JSON(http.StatusBadRequest, errorResponse)
		return
	}

	user, err := db.GetUserByEmail(req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusOK, gin.H{"message": "If an account with that email exists, a password reset link has been sent."})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Generate a password reset token
	token, err := utils.GenerateRandomToken(32)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Set token and expiry. Store the raw token.
	expires := time.Now().Add(time.Hour * 1) // Token valid for 1 hour
	user.PasswordResetToken = &token
	user.PasswordResetExpires = &expires

	if err := db.UpdateUser(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save reset token"})
		return
	}

	// Send password reset email
	resetLink := fmt.Sprintf("http://localhost:3000/reset-password?token=%s", token)
	emailBody := fmt.Sprintf("<h1>Password Reset</h1><p>Please reset your password by clicking this link: <a href=\"%s\">%s</a></p>", resetLink, resetLink)
	go func() {
		if err := utils.SendEmail(user.Email, "Reset Your Password", emailBody); err != nil {
			log.Printf("Failed to send password reset email to %s: %v", user.Email, err)
		}
	}()

	c.JSON(http.StatusOK, gin.H{"message": "If an account with that email exists, a password reset link has been sent."})
}

func ResetPassword(c *gin.Context) {
	var req models.ResetPasswordRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse := utils.HandleValidationErrors(err)
		c.JSON(http.StatusBadRequest, errorResponse)
		return
	}

	// Find user by the raw token from the request
	user, err := db.GetUserByPasswordResetToken(req.Token)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or expired password reset token"})
		return
	}

	// Check if the token has expired
	if user.PasswordResetExpires == nil || time.Now().After(*user.PasswordResetExpires) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or expired password reset token"})
		return
	}

	// Hash the new password
	newHashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash new password"})
		return
	}

	// Update password and clear reset fields
	user.Password = newHashedPassword
	user.PasswordResetToken = nil
	user.PasswordResetExpires = nil

	if err := db.UpdateUser(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password has been reset successfully."})
}

func VerifyEmail(c *gin.Context) {
	var req models.VerifyEmailRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse := utils.HandleValidationErrors(err)
		c.JSON(http.StatusBadRequest, errorResponse)
		return
	}

	// Hash the token from the request to match the one in the DB
	hashedToken, err := utils.HashPassword(req.Token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash token"})
		return
	}

	user, err := db.GetUserByEmailVerificationToken(hashedToken)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or expired email verification token"})
		return
	}

	// Check if the token has expired
	if user.EmailVerificationExpires == nil || time.Now().After(*user.EmailVerificationExpires) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or expired email verification token"})
		return
	}

	// Mark email as verified and clear verification fields
	user.IsEmailVerified = true
	user.EmailVerificationToken = nil
	user.EmailVerificationExpires = nil

	if err := db.UpdateUser(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify email"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Email has been verified successfully."})
}

