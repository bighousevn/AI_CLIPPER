package application

import (
	userDomain "ai-clipper/server2/internal/auth/domain/user"
	"context"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
)

// IAuthUseCase defines the interface for all authentication and user-related business logic.
type IAuthUseCase interface {
	Register(ctx context.Context, req RegisterRequest) (*RegisterResponse, error)
	Login(ctx context.Context, req LoginRequest) (*LoginResponse, error)
	RefreshToken(ctx context.Context, refreshToken string) (*RefreshTokenResponse, error)
	Logout(ctx context.Context, refreshToken string) error
	ForgotPassword(ctx context.Context, req ForgotPasswordRequest) error
	ResetPassword(ctx context.Context, req ResetPasswordRequest) error
	VerifyEmail(ctx context.Context, token string) error
	ChangePassword(ctx context.Context, userID uuid.UUID, req ChangePasswordRequest) error
	GetUserProfile(ctx context.Context, userID uuid.UUID) (*UserProfileResponse, error)
}

// AuthUseCase is the concrete implementation of IAuthUseCase.
type AuthUseCase struct {
	userRepo       userDomain.UserRepository
	hasher         PasswordHasher
	tokenGenerator TokenGenerator
	emailSender    EmailSender
}

// NewAuthUseCase creates a new instance of AuthUseCase.
func NewAuthUseCase(userRepo userDomain.UserRepository, hasher PasswordHasher, tokenGenerator TokenGenerator, emailSender EmailSender) IAuthUseCase {
	return &AuthUseCase{
		userRepo:       userRepo,
		hasher:         hasher,
		tokenGenerator: tokenGenerator,
		emailSender:    emailSender,
	}
}

func (uc *AuthUseCase) Register(ctx context.Context, req RegisterRequest) (*RegisterResponse, error) {
	email := strings.ToLower(req.Email)
	existingUser, err := uc.userRepo.FindByEmail(ctx, email)
	if err != nil {
		// We assume a specific "not found" error is not returned, so any error is a server error.
		log.Printf("ERROR: Failed to find user by email: %v", err)
		return nil, errors.New("server error")
	}

	if existingUser != nil && existingUser.IsEmailVerified {
		return nil, errors.New("user with this email already exists")
	}

	hashedPassword, err := uc.hasher.Hash(req.Password)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	verificationToken, err := uc.tokenGenerator.GenerateRandomToken(32)
	if err != nil {
		return nil, errors.New("failed to generate verification token")
	}
	verificationExpires := time.Now().UTC().Add(time.Hour * 24)

	var userToSave *userDomain.User
	if existingUser != nil { // User exists but is not verified
		userToSave = existingUser
		userToSave.Username = req.Username
		userToSave.PasswordHash = hashedPassword
		userToSave.EmailVerificationToken = &verificationToken
		userToSave.EmailVerificationExpires = &verificationExpires
	} else { // New user
		userToSave = &userDomain.User{
			ID:                       uuid.New(),
			Username:                 req.Username,
			Email:                    email,
			PasswordHash:             hashedPassword,
			Credits:                  10, // Default credits
			IsEmailVerified:          false,
			EmailVerificationToken:   &verificationToken,
			EmailVerificationExpires: &verificationExpires,
		}
	}

	if err := uc.userRepo.Save(ctx, userToSave); err != nil {
		log.Printf("ERROR: Failed to save user: %v", err)
		return nil, errors.New("failed to save user")
	}

	// Send verification email
	if err := uc.emailSender.SendVerificationEmail(userToSave.Email, userToSave.Username, verificationToken); err != nil {
		log.Printf("WARNING: Failed to send verification email to %s: %v", userToSave.Email, err)
		// Don't fail registration if email fails, user can request resend later
	} else {
		log.Printf("Verification email sent successfully to %s", userToSave.Email)
	}

	response := &RegisterResponse{
		ID:       userToSave.ID,
		Username: userToSave.Username,
		Email:    userToSave.Email,
	}

	return response, nil
}

func (uc *AuthUseCase) Login(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	user, err := uc.userRepo.FindByEmail(ctx, strings.ToLower(req.Email))
	if err != nil || user == nil {
		return nil, errors.New("invalid email or password")
	}

	if !uc.hasher.Check(req.Password, user.PasswordHash) {
		return nil, errors.New("invalid email or password")
	}

	if !user.IsEmailVerified {
		return nil, errors.New("email not verified")
	}

	accessToken, err := uc.tokenGenerator.GenerateAccessToken(user.ID)
	if err != nil {
		return nil, errors.New("failed to generate access token")
	}

	refreshToken, err := uc.tokenGenerator.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, errors.New("failed to generate refresh token")
	}

	user.RefreshToken = &refreshToken
	if err := uc.userRepo.Save(ctx, user); err != nil {
		return nil, errors.New("failed to save refresh token")
	}

	return &LoginResponse{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}

func (uc *AuthUseCase) RefreshToken(ctx context.Context, refreshToken string) (*RefreshTokenResponse, error) {
	userID, err := uc.tokenGenerator.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil || user == nil {
		return nil, errors.New("user not found")
	}

	if user.RefreshToken == nil || *user.RefreshToken != refreshToken {
		// Token reuse detected, clear the token for security
		user.RefreshToken = nil
		uc.userRepo.Save(ctx, user) // Try to save, but deny refresh anyway
		return nil, errors.New("refresh token mismatch or already used")
	}

	newAccessToken, err := uc.tokenGenerator.GenerateAccessToken(user.ID)
	if err != nil {
		return nil, errors.New("failed to generate new access token")
	}

	newRefreshToken, err := uc.tokenGenerator.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, errors.New("failed to generate new refresh token")
	}

	user.RefreshToken = &newRefreshToken
	if err := uc.userRepo.Save(ctx, user); err != nil {
		return nil, errors.New("failed to save new refresh token")
	}

	return &RefreshTokenResponse{AccessToken: newAccessToken, RefreshToken: newRefreshToken}, nil
}

func (uc *AuthUseCase) Logout(ctx context.Context, refreshToken string) error {
	userID, err := uc.tokenGenerator.ValidateRefreshToken(refreshToken)
	if err != nil {
		return errors.New("invalid refresh token")
	}

	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil || user == nil {
		return nil // User not found, nothing to do
	}

	user.RefreshToken = nil
	if err := uc.userRepo.Save(ctx, user); err != nil {
		return errors.New("failed to logout")
	}

	return nil
}

func (uc *AuthUseCase) ForgotPassword(ctx context.Context, req ForgotPasswordRequest) error {
	user, err := uc.userRepo.FindByEmail(ctx, strings.ToLower(req.Email))
	if err != nil || user == nil {
		return nil // Don't reveal if user exists
	}

	token, err := uc.tokenGenerator.GenerateRandomToken(32)
	if err != nil {
		return errors.New("failed to generate token")
	}

	expires := time.Now().UTC().Add(time.Hour * 1)
	user.PasswordResetToken = &token
	user.PasswordResetExpires = &expires

	if err := uc.userRepo.Save(ctx, user); err != nil {
		return errors.New("failed to save reset token")
	}

	// Send password reset email
	if err := uc.emailSender.SendPasswordResetEmail(user.Email, user.Username, token); err != nil {
		log.Printf("WARNING: Failed to send password reset email to %s: %v", user.Email, err)
		// Don't fail the request if email fails
	} else {
		log.Printf("Password reset email sent successfully to %s", user.Email)
	}

	return nil
}

func (uc *AuthUseCase) ResetPassword(ctx context.Context, req ResetPasswordRequest) error {
	user, err := uc.userRepo.FindByPasswordResetToken(ctx, req.Token)
	if err != nil || user == nil {
		return errors.New("invalid or expired password reset token")
	}

	if user.PasswordResetExpires == nil || time.Now().UTC().After(*user.PasswordResetExpires) {
		return errors.New("invalid or expired password reset token")
	}

	newHashedPassword, err := uc.hasher.Hash(req.Password)
	if err != nil {
		return errors.New("failed to hash new password")
	}

	user.PasswordHash = newHashedPassword
	user.PasswordResetToken = nil
	user.PasswordResetExpires = nil

	if err := uc.userRepo.Save(ctx, user); err != nil {
		return errors.New("failed to update password")
	}

	return nil
}

func (uc *AuthUseCase) VerifyEmail(ctx context.Context, token string) error {
	user, err := uc.userRepo.FindByEmailVerificationToken(ctx, token)
	if err != nil || user == nil {
		return errors.New("invalid or expired email verification token")
	}

	if user.EmailVerificationExpires == nil || time.Now().UTC().After(*user.EmailVerificationExpires) {
		return errors.New("invalid or expired email verification token")
	}

	user.IsEmailVerified = true
	user.EmailVerificationToken = nil
	user.EmailVerificationExpires = nil

	if err := uc.userRepo.Save(ctx, user); err != nil {
		return errors.New("failed to verify email")
	}

	return nil
}

func (uc *AuthUseCase) ChangePassword(ctx context.Context, userID uuid.UUID, req ChangePasswordRequest) error {
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil || user == nil {
		return errors.New("user not found")
	}

	if !uc.hasher.Check(req.OldPassword, user.PasswordHash) {
		return errors.New("invalid old password")
	}

	newHashedPassword, err := uc.hasher.Hash(req.NewPassword)
	if err != nil {
		return errors.New("failed to hash new password")
	}

	user.PasswordHash = newHashedPassword
	if err := uc.userRepo.Save(ctx, user); err != nil {
		return errors.New("failed to update password")
	}

	return nil
}

func (uc *AuthUseCase) GetUserProfile(ctx context.Context, userID uuid.UUID) (*UserProfileResponse, error) {
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil || user == nil {
		return nil, errors.New("user not found")
	}

	return &UserProfileResponse{
		ID:               user.ID,
		Username:         user.Username,
		Email:            user.Email,
		Credits:          user.Credits,
		StripeCustomerID: user.StripeCustomerID,
	}, nil
}
