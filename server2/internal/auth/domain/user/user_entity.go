package domain

import "time"

// User is the core entity for a user in our system.
// It contains only business logic and data, with no dependencies on external frameworks.
type User struct {
	ID                       int64
	Username                 string
	Email                    string
	PasswordHash             string // Renamed for clarity
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
	DeletedAt                *time.Time // Use nullable time for soft deletes in domain
}