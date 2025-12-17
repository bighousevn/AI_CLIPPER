package user

import (
	"context"

	"github.com/google/uuid"
)

type UserRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByPasswordResetToken(ctx context.Context, token string) (*User, error)
	FindByEmailVerificationToken(ctx context.Context, token string) (*User, error)
	FindByRefreshToken(ctx context.Context, token string) (*User, error)
	FindByStripeCustomerID(ctx context.Context, stripeCustomerID string) (*User, error)
	Save(ctx context.Context, user *User) error // Handles both Create and Update
	Delete(ctx context.Context, id uuid.UUID) error // Optional: for hard deletes if needed
}
