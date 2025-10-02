package services

import (
	"bighousevn/be/models"
	"bighousevn/be/repository"
	"bighousevn/be/utils"
	"errors"
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
)

type AuthService interface {
	Register(req *models.RegisterRequest) (*models.User, error)
	Login(req *models.LoginRequest) (string, error)
	ForgotPassword(req *models.ForgotPasswordRequest) error
	ResetPassword(req *models.ResetPasswordRequest) error
	VerifyEmail(req *models.VerifyEmailRequest) error
	ChangePassword(user *models.User, req *models.ChangePasswordRequest) error
}

type authService struct {
	authRepo repository.AuthRepository
}

func NewAuthService(authRepo repository.AuthRepository) AuthService {
	return &authService{authRepo: authRepo}
}

func (s *authService) Register(req *models.RegisterRequest) (*models.User, error) {
	_, err := s.authRepo.GetUserByEmail(req.Email)
	if err == nil {
		return nil, errors.New("user with this email already exists")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("database error while checking for user")
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	verificationToken, err := utils.GenerateRandomToken(32)
	if err != nil {
		return nil, errors.New("failed to generate verification token")
	}
	hashedVerificationToken, err := utils.HashPassword(verificationToken)
	if err != nil {
		return nil, errors.New("failed to hash verification token")
	}

	verificationExpires := time.Now().Add(time.Hour * 24)

	newUser := models.User{
		Username:               req.Username,
		Email:                  req.Email,
		Password:               hashedPassword,
		EmailVerificationToken: &hashedVerificationToken,
		EmailVerificationExpires: &verificationExpires,
	}

	if err := s.authRepo.CreateUser(&newUser); err != nil {
		return nil, errors.New("failed to create user")
	}

	verificationLink := fmt.Sprintf("http://localhost:3000/verify-email?token=%s", verificationToken)
	emailBody := fmt.Sprintf("<h1>Welcome!</h1><p>Please verify your email by clicking this link: <a href=\"%s\">%s</a></p>", verificationLink, verificationLink)
	go func() {
		if err := utils.SendEmail(newUser.Email, "Verify Your Email", emailBody); err != nil {
			log.Printf("Failed to send verification email to %s: %v", newUser.Email, err)
		}
	}()

	return &newUser, nil
}

func (s *authService) Login(req *models.LoginRequest) (string, error) {
	user, err := s.authRepo.GetUserByEmail(req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", errors.New("invalid email or password")
		}
		return "", errors.New("database error")
	}

	if !utils.CheckPasswordHash(req.Password, user.Password) {
		return "", errors.New("invalid email or password")
	}

	token, err := utils.GenerateJWT(user.ID)
	if err != nil {
		return "", errors.New("failed to generate token")
	}

	return token, nil
}

func (s *authService) ForgotPassword(req *models.ForgotPasswordRequest) error {
	user, err := s.authRepo.GetUserByEmail(req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil // Don't reveal if user exists
		}
		return errors.New("database error")
	}

	token, err := utils.GenerateRandomToken(32)
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
		if err := utils.SendEmail(user.Email, "Reset Your Password", emailBody); err != nil {
			log.Printf("Failed to send password reset email to %s: %v", user.Email, err)
		}
	}()

	return nil
}

func (s *authService) ResetPassword(req *models.ResetPasswordRequest) error {
	user, err := s.authRepo.GetUserByPasswordResetToken(req.Token)
	if err != nil {
		return errors.New("invalid or expired password reset token")
	}

	if user.PasswordResetExpires == nil || time.Now().After(*user.PasswordResetExpires) {
		return errors.New("invalid or expired password reset token")
	}

	newHashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return errors.New("failed to hash new password")
	}

	user.Password = newHashedPassword
	user.PasswordResetToken = nil
	user.PasswordResetExpires = nil

	if err := s.authRepo.UpdateUser(user); err != nil {
		return errors.New("failed to update password")
	}

	return nil
}

func (s *authService) VerifyEmail(req *models.VerifyEmailRequest) error {
	hashedToken, err := utils.HashPassword(req.Token)
	if err != nil {
		return errors.New("failed to hash token")
	}

	user, err := s.authRepo.GetUserByEmailVerificationToken(hashedToken)
	if err != nil {
		return errors.New("invalid or expired email verification token")
	}

	if user.EmailVerificationExpires == nil || time.Now().After(*user.EmailVerificationExpires) {
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
	if !utils.CheckPasswordHash(req.OldPassword, user.Password) {
		return errors.New("invalid old password")
	}

	newHashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return errors.New("failed to hash new password")
	}

	user.Password = newHashedPassword
	if err := s.authRepo.UpdateUser(user); err != nil {
		return errors.New("failed to update password")
	}

	return nil
}