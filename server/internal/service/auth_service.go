package services

import (
	"bighousevn/be/internal/auth"
	"bighousevn/be/internal/email"
	"bighousevn/be/internal/models"
	"bighousevn/be/internal/repository"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
)

// AuthService defines the interface for authentication-related business logic.
type AuthService interface {
	Register(req *models.RegisterRequest) (*models.User, error)
	Login(req *models.LoginRequest) (string, string, error)
	RefreshToken(refreshToken string) (string, string, error)
	Logout(refreshToken string) error
	ForgotPassword(req *models.ForgotPasswordRequest) error
	ResetPassword(req *models.ResetPasswordRequest) error
	VerifyEmail(req *models.VerifyEmailRequest) error
	ChangePassword(user *models.User, req *models.ChangePasswordRequest) error
}

type authService struct {
	authRepo repository.AuthRepository
}

// NewAuthService creates a new instance of AuthService.
func NewAuthService(authRepo repository.AuthRepository) AuthService {
	return &authService{authRepo: authRepo}
}

func (s *authService) Register(req *models.RegisterRequest) (*models.User, error) {
	log.Printf("DEBUG: Received registration request for email: %s", req.Email)
	req.Email = strings.ToLower(req.Email)
	log.Printf("DEBUG: Normalized email to: %s", req.Email)

	user, err := s.authRepo.GetUserByEmail(req.Email)
	if err != nil {
		log.Printf("ERROR: Database error while checking for user: %v", err)
		return nil, errors.New("database error while checking for user")
	}

	if user != nil {
		log.Printf("DEBUG: User with email %s found. Is verified: %t", req.Email, user.IsEmailVerified)
	} else {
		log.Println("DEBUG: No user found with this email, proceeding to create a new user.")
	}

	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	verificationToken, err := auth.GenerateRandomToken(32)
	if err != nil {
		return nil, errors.New("failed to generate verification token")
	}
	// Use UTC and extend to 24 hours for better UX
	verificationExpires := time.Now().UTC().Add(time.Hour * 24)
	log.Printf("DEBUG: Verification token expires at: %v", verificationExpires)

	if user != nil {
		if user.IsEmailVerified {
			return nil, errors.New("user with this email already exists")
		}

		// User is not verified, overwrite data
		log.Printf("DEBUG: Updating existing unverified user")
		user.Username = req.Username
		user.PasswordHash = hashedPassword
		user.EmailVerificationToken = &verificationToken
		user.EmailVerificationExpires = &verificationExpires
		log.Printf("DEBUG: Setting EmailVerificationExpires to: %v", verificationExpires)

		if err := s.authRepo.UpdateUser(user); err != nil {
			log.Printf("ERROR: Failed to update unverified user: %v", err)
			return nil, errors.New("failed to update user")
		}
	} else {
		// User does not exist, create new user
		log.Printf("DEBUG: Creating new user with verification token: %s", verificationToken)
		newUser := &models.User{
			Username:                 req.Username,
			Email:                    req.Email,
			PasswordHash:             hashedPassword,
			EmailVerificationToken:   &verificationToken,
			EmailVerificationExpires: &verificationExpires,
		}

		log.Printf("DEBUG: About to create user in database with token: %s, expires: %v", *newUser.EmailVerificationToken, *newUser.EmailVerificationExpires)
		if err := s.authRepo.CreateUser(newUser); err != nil {
			log.Printf("ERROR: Failed to create new user, potential 409 Conflict: %v", err)
			return nil, errors.New("failed to create user")
		}
		log.Printf("DEBUG: User created successfully")

		// Query user back from database to get the Supabase-generated UUID
		createdUser, err := s.authRepo.GetUserByEmail(req.Email)
		if err != nil || createdUser == nil {
			log.Printf("ERROR: Failed to retrieve created user: %v", err)
			return nil, errors.New("failed to retrieve created user")
		}
		log.Printf("DEBUG: Retrieved user from DB with ID: %s", createdUser.ID)
		user = createdUser
	}

	verificationLink := fmt.Sprintf("http://localhost:3000/verify?token=%s", verificationToken)
	emailBody := fmt.Sprintf("<h1>Welcome!</h1><p>Please verify your email by clicking this link: <a href=\"%s\">%s</a></p>", verificationLink, verificationLink)
	go func() {
		if err := email.SendEmail(user.Email, "Verify Your Email", emailBody); err != nil {
			log.Printf("Failed to send verification email to %s: %v", user.Email, err)
		}
	}()

	return user, nil
}

func (s *authService) Login(req *models.LoginRequest) (string, string, error) {
	user, err := s.authRepo.GetUserByEmail(req.Email)
	if err != nil {
		return "", "", errors.New("database error")
	}
	if user == nil {
		return "", "", errors.New("invalid email or password")
	}

	if !auth.CheckPasswordHash(req.Password, user.PasswordHash) {
		return "", "", errors.New("invalid email or password")
	}

	accessToken, err := auth.GenerateJWT(*user.ID)
	if err != nil {
		return "", "", errors.New("failed to generate access token")
	}

	refreshToken, err := auth.GenerateRefreshToken(*user.ID)
	if err != nil {
		return "", "", errors.New("failed to generate refresh token")
	}

	user.RefreshToken = &refreshToken
	if err := s.authRepo.UpdateUser(user); err != nil {
		return "", "", errors.New("failed to save refresh token")
	}

	return accessToken, refreshToken, nil
}

func (s *authService) RefreshToken(refreshToken string) (string, string, error) {
	claims, err := auth.ValidateRefreshToken(refreshToken)
	if err != nil {
		return "", "", errors.New("invalid refresh token")
	}

	userID, err := uuid.Parse(claims["user_id"].(string))
	if err != nil {
		return "", "", errors.New("invalid user id in refresh token")
	}

	user, err := s.authRepo.GetUserByID(userID)
	if err != nil || user == nil {
		return "", "", errors.New("user not found")
	}

	if user.RefreshToken == nil || *user.RefreshToken != refreshToken {
		// Token reuse detected, clear the token for security
		if user != nil {
			user.RefreshToken = nil
			if err := s.authRepo.UpdateUser(user); err != nil {
				log.Printf("ERROR: Failed to clear refresh token for user %s: %v", user.ID, err)
				// Even if this fails, we should still deny the refresh token
			}
		}
		return "", "", errors.New("refresh token mismatch")
	}
	newAccessToken, err := auth.GenerateJWT(*user.ID)
	if err != nil {
		return "", "", errors.New("failed to generate new access token")
	}

	newRefreshToken, err := auth.GenerateRefreshToken(*user.ID)
	if err != nil {
		return "", "", errors.New("failed to generate new refresh token")
	}

	user.RefreshToken = &newRefreshToken
	if err := s.authRepo.UpdateUser(user); err != nil {
		return "", "", errors.New("failed to save new refresh token")
	}

	return newAccessToken, newRefreshToken, nil
}

func (s *authService) Logout(refreshToken string) error {
	claims, err := auth.ValidateRefreshToken(refreshToken)
	if err != nil {
		return errors.New("invalid refresh token")
	}

	userID, err := uuid.Parse(claims["user_id"].(string))
	if err != nil {
		return errors.New("invalid user id in refresh token")
	}

	user, err := s.authRepo.GetUserByID(userID)
	if err != nil || user == nil {
		return errors.New("user not found")
	}

	user.RefreshToken = nil
	if err := s.authRepo.UpdateUser(user); err != nil {
		return errors.New("failed to logout")
	}

	return nil
}

func (s *authService) ForgotPassword(req *models.ForgotPasswordRequest) error {
	user, err := s.authRepo.GetUserByEmail(req.Email)
	if err != nil {
		return errors.New("database error")
	}
	if user == nil {
		return nil // Don't reveal if user exists
	}

	token, err := auth.GenerateRandomToken(32)
	if err != nil {
		return errors.New("failed to generate token")
	}

	expires := time.Now().Add(time.Hour * 1)
	user.PasswordResetToken = &token
	user.PasswordResetExpires = &expires

	if err := s.authRepo.UpdateUser(user); err != nil {
		return errors.New("failed to save reset token")
	}

	resetLink := fmt.Sprintf("http://localhost:3000/reset-password?token=%s", token)
	emailBody := fmt.Sprintf("<h1>Password Reset</h1><p>Please reset your password by clicking this link: <a href=\"%s\">%s</a></p>", resetLink, resetLink)
	go func() {
		if err := email.SendEmail(user.Email, "Reset Your Password", emailBody); err != nil {
			log.Printf("Failed to send password reset email to %s: %v", user.Email, err)
		}
	}()

	return nil
}

func (s *authService) ResetPassword(req *models.ResetPasswordRequest) error {
	user, err := s.authRepo.GetUserByField("password_reset_token", req.Token)
	if err != nil || user == nil {
		return errors.New("invalid or expired password reset token")
	}

	if user.PasswordResetExpires == nil || time.Now().After(*user.PasswordResetExpires) {
		return errors.New("invalid or expired password reset token")
	}

	newHashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		return errors.New("failed to hash new password")
	}

	user.PasswordHash = newHashedPassword
	user.PasswordResetToken = nil
	user.PasswordResetExpires = nil

	if err := s.authRepo.UpdateUser(user); err != nil {
		return errors.New("failed to update password")
	}

	return nil
}

func (s *authService) VerifyEmail(req *models.VerifyEmailRequest) error {
	log.Printf("DEBUG: Attempting to verify email with token: %s", req.Token)
	user, err := s.authRepo.GetUserByField("email_verification_token", req.Token)
	if err != nil || user == nil {
		log.Printf("ERROR: User not found for verification token: %v", err)
		return errors.New("invalid or expired email verification token")
	}

	log.Printf("DEBUG: Found user: %s", user.Email)
	log.Printf("DEBUG: Current time (UTC): %v", time.Now().UTC())
	log.Printf("DEBUG: Current time (Local): %v", time.Now())
	log.Printf("DEBUG: Token expires at (from DB): %v", user.EmailVerificationExpires)

	if user.EmailVerificationExpires == nil {
		log.Printf("ERROR: EmailVerificationExpires is nil")
		return errors.New("invalid or expired email verification token")
	}

	currentTime := time.Now().UTC()
	expiresTime := *user.EmailVerificationExpires

	log.Printf("DEBUG: Time comparison - Current: %v, Expires: %v", currentTime, expiresTime)
	log.Printf("DEBUG: Is expired? %v", currentTime.After(expiresTime))

	if currentTime.After(expiresTime) {
		log.Printf("ERROR: Verification token expired")
		return errors.New("invalid or expired email verification token")
	}

	user.IsEmailVerified = true
	user.EmailVerificationToken = nil
	user.EmailVerificationExpires = nil

	if err := s.authRepo.UpdateUser(user); err != nil {
		return errors.New("failed to verify email")
	}

	return nil
}

func (s *authService) ChangePassword(user *models.User, req *models.ChangePasswordRequest) error {
	if !auth.CheckPasswordHash(req.OldPassword, user.PasswordHash) {
		return errors.New("invalid old password")
	}

	newHashedPassword, err := auth.HashPassword(req.NewPassword)
	if err != nil {
		return errors.New("failed to hash new password")
	}

	user.PasswordHash = newHashedPassword
	if err := s.authRepo.UpdateUser(user); err != nil {
		return errors.New("failed to update password")
	}

	return nil
}
