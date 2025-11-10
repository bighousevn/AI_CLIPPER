package user

import (
	"time"

	"github.com/google/uuid"
)

// User is the core entity for a user in our system.
type User struct {
	ID                       uuid.UUID
	Username                 string
	Email                    string
	PasswordHash             string
	Credits                  int
	StripeCustomerID         *string
	RefreshToken             *string
	PasswordResetToken       *string
	PasswordResetExpires     *time.Time
	IsEmailVerified          bool
	EmailVerificationToken   *string
	EmailVerificationExpires *time.Time
	CreatedAt                time.Time
	UpdatedAt                time.Time
	DeletedAt                *time.Time
}