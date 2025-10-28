package domain

import "context"

type UserRepository interface {
	FindByID(ctx context.Context, id int64) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByPasswordResetToken(ctx context.Context, token string) (*User, error)
	FindByEmailVerificationToken(ctx context.Context, token string) (*User, error)
	FindByRefreshToken(ctx context.Context, token string) (*User, error)
	Save(ctx context.Context, user *User) error // Handles both Create and Update
	Delete(ctx context.Context, id int64) error // Optional: for hard deletes if needed
}
