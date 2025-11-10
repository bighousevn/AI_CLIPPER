package application

import (
	"time"

	"github.com/google/uuid"
)

// TokenGenerator defines the interface for creating and validating tokens.
type TokenGenerator interface {
	GenerateAccessToken(userID uuid.UUID) (string, error)
	GenerateRefreshToken(userID uuid.UUID) (string, error)
	ValidateAccessToken(tokenString string) (uuid.UUID, error)
	ValidateRefreshToken(tokenString string) (uuid.UUID, error)
	GenerateRandomToken(length int) (string, error)
}

// TokenClaims represents the claims in a token.
type TokenClaims interface {
	GetUserID() uuid.UUID
	ExpiresAt() time.Time
}
