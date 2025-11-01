package models

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user in the database.
type User struct {
	ID                       *uuid.UUID `json:"id,omitempty"`
	Username                 string     `json:"username,omitempty"`
	Email                    string     `json:"email"`
	PasswordHash             string     `json:"password_hash"`
	Credits                  int        `json:"credits,omitempty"`
	StripeCustomerID         *string    `json:"stripe_customer_id,omitempty"`
	RefreshToken             *string    `json:"refresh_token"`
	PasswordResetToken       *string    `json:"password_reset_token"`
	PasswordResetExpires     *time.Time `json:"password_reset_expires"`
	IsEmailVerified          bool       `json:"is_email_verified"`
	EmailVerificationToken   *string    `json:"email_verification_token"`
	EmailVerificationExpires *time.Time `json:"email_verification_expires"`
	CreatedAt                time.Time  `json:"created_at,omitempty"`
	UpdatedAt                time.Time  `json:"updated_at,omitempty"`
	DeletedAt                *time.Time `json:"deleted_at"`
}

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6,max=100"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type ResetPasswordRequest struct {
	Token    string `json:"token" binding:"required"`
	Password string `json:"password" binding:"required,min=6,max=100"`
}

type VerifyEmailRequest struct {
	Token string `form:"token" binding:"required"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6,max=100"`
}
