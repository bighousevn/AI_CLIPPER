package models

import (
	"time"

	"gorm.io/gorm"
)

// User represents a user in the database, with GORM tags.
type User struct {
	ID                       int64          `json:"id" gorm:"primaryKey"`
	Username                 string         `json:"username" gorm:"unique"`
	Email                    string         `json:"email" gorm:"unique"`
	Password                 string         `json:"-" gorm:"column:password_hash"`
	Credits                  int            `json:"credits" gorm:"default:10"`
	StripeCustomerID         *string        `json:"stripe_customer_id" gorm:"unique"`
	RefreshToken             *string        `json:"-" gorm:"unique"`
	PasswordResetToken       *string        `json:"-" gorm:"uniqueIndex"`
	PasswordResetExpires     *time.Time     `json:"-"`
	IsEmailVerified          bool           `json:"is_email_verified" gorm:"default:false"`
	EmailVerificationToken   *string        `json:"-" gorm:"uniqueIndex"`
	EmailVerificationExpires *time.Time     `json:"-"`
	CreatedAt                time.Time      `json:"created_at"`
	UpdatedAt                time.Time      `json:"updated_at"`
	DeletedAt                gorm.DeletedAt `gorm:"index"`
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
