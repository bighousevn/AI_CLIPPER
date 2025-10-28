package domain

import (
	"context"
	"time"
)

// User is a pure domain entity, representing a user.
// It has no dependencies on the database or web framework.
type User struct {
	ID                       int64
	Username                 string
	Email                    string
	Password                 string // This is the hashed password
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
	DeletedAt                *time.Time // Use a pointer to time for nullable DeletedAt
}

// UserRepository defines the interface for user data persistence.
// It uses context for cancellation and deadlines.
type UserRepository interface {
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByID(ctx context.Context, id int64) (*User, error)
	FindByPasswordResetToken(ctx context.Context, token string) (*User, error)
	FindByEmailVerificationToken(ctx context.Context, token string) (*User, error)
	FindByRefreshToken(ctx context.Context, token string) (*User, error)
	Create(ctx context.Context, user *User) error
	Update(ctx context.Context, user *User) error
}
