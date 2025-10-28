package application

import (
	userDomain "ai-clipper/server2/internal/auth/domain/user"
	"context"
	"errors"
	"time"
)

type IAuthUseCase interface {
	Register(ctx context.Context, req RegisterRequest) (*RegisterResponse, error)
}

type AuthUseCase struct {
	userRepo userDomain.UserRepository
	hasher   PasswordHasher
}

func NewAuthUseCase(userRepo userDomain.UserRepository, hasher PasswordHasher) IAuthUseCase {
	return &AuthUseCase{
		userRepo: userRepo,
		hasher:   hasher,
	}
}

func (uc *AuthUseCase) Register(ctx context.Context, req RegisterRequest) (*RegisterResponse, error) {

	existingUser, err := uc.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {

	}
	if existingUser != nil {
		return nil, errors.New("user with this email already exists")
	}

	// 2. Hash the password
	hashedPassword, err := uc.hasher.Hash(req.Password)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	// 3. Create a new user entity
	newUser := &userDomain.User{
		Username:        req.Username,
		Email:           req.Email,
		PasswordHash:    hashedPassword,
		Credits:         10, // Default credits
		IsEmailVerified: false,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// 4. Save the new user to the database
	if err := uc.userRepo.Save(ctx, newUser); err != nil {
		return nil, errors.New("failed to save user")
	}

	// 5. Map the new user to a response DTO
	response := &RegisterResponse{
		ID:       newUser.ID,
		Username: newUser.Username,
		Email:    newUser.Email,
	}

	return response, nil
}
