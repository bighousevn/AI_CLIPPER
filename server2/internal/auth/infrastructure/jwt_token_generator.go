package infrastructure

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// JWTConfig holds the configuration for JWT generation.
type JWTConfig struct {
	AccessSecret  string
	RefreshSecret string
	AccessExpiry  time.Duration
	RefreshExpiry time.Duration
}

// JWTTokenGenerator is a concrete implementation of the TokenGenerator using JWT.
type JWTTokenGenerator struct {
	config JWTConfig
}

func NewJWTTokenGenerator(config JWTConfig) *JWTTokenGenerator {
	return &JWTTokenGenerator{config: config}
}

func (g *JWTTokenGenerator) GenerateAccessToken(userID uuid.UUID) (string, error) {
	return g.generateToken(userID, g.config.AccessExpiry, g.config.AccessSecret)
}

func (g *JWTTokenGenerator) GenerateRefreshToken(userID uuid.UUID) (string, error) {
	return g.generateToken(userID, g.config.RefreshExpiry, g.config.RefreshSecret)
}

func (g *JWTTokenGenerator) generateToken(userID uuid.UUID, expiry time.Duration, secret string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID.String(),
		"exp":     time.Now().Add(expiry).Unix(),
		"iat":     time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// validateToken is a helper method to validate JWT tokens with a given secret
func (g *JWTTokenGenerator) validateToken(tokenString string, secret string) (uuid.UUID, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})

	if err != nil {
		return uuid.Nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userIDStr, ok := claims["user_id"].(string)
		if !ok {
			return uuid.Nil, errors.New("invalid user_id in token")
		}
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			return uuid.Nil, errors.New("cannot parse user_id from token")
		}
		return userID, nil
	}
	return uuid.Nil, errors.New("invalid token")
}

func (g *JWTTokenGenerator) ValidateAccessToken(tokenString string) (uuid.UUID, error) {
	return g.validateToken(tokenString, g.config.AccessSecret)
}

func (g *JWTTokenGenerator) ValidateRefreshToken(tokenString string) (uuid.UUID, error) {
	return g.validateToken(tokenString, g.config.RefreshSecret)
}

func (g *JWTTokenGenerator) GenerateRandomToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
