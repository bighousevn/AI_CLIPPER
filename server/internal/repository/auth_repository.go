package repository

import (
	"bighousevn/be/internal/models"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	supabase "github.com/supabase-community/supabase-go"
)

// AuthRepository defines the interface for authentication-related database operations.
type AuthRepository interface {
	GetUserByEmail(email string) (*models.User, error)
	CreateUser(user *models.User) error
	UpdateUser(user *models.User) error
	GetUserByID(id uuid.UUID) (*models.User, error)
	GetUserByField(field, value string) (*models.User, error)
}

type authRepository struct {
	client *supabase.Client
}

// NewAuthRepository creates a new instance of AuthRepository.
func NewAuthRepository(client *supabase.Client) AuthRepository {
	return &authRepository{client: client}
}

// GetUserByEmail retrieves a user by their email address.
func (r *authRepository) GetUserByEmail(email string) (*models.User, error) {
	return r.GetUserByField("email", email)
}

// GetUserByID retrieves a user by their ID.
func (r *authRepository) GetUserByID(id uuid.UUID) (*models.User, error) {
	return r.GetUserByField("id", id.String())
}

// GetUserByField retrieves a single user by a specific field and value.
func (r *authRepository) GetUserByField(field, value string) (*models.User, error) {
	var results []models.User

	fmt.Printf("DEBUG REPO: Querying user by field '%s' with value '%s'\n", field, value)

	// Execute returns (data []byte, count int64, error)
	data, count, err := r.client.From("users").Select("*", "exact", false).Eq(field, value).Execute()
	if err != nil {
		fmt.Printf("ERROR REPO: Query failed: %v\n", err)
		return nil, fmt.Errorf("error querying user by %s: %w", field, err)
	}

	fmt.Printf("DEBUG REPO: Query returned %d rows, data length: %d\n", count, len(data))
	fmt.Printf("DEBUG REPO: Raw data: %s\n", string(data))

	// Unmarshal the JSON response
	if err := json.Unmarshal(data, &results); err != nil {
		fmt.Printf("ERROR REPO: Unmarshal failed: %v\n", err)
		return nil, fmt.Errorf("error unmarshaling user data: %w", err)
	}

	fmt.Printf("DEBUG REPO: Found %d users\n", len(results))
	if len(results) == 0 {
		return nil, nil // Or a specific "not found" error
	}
	return &results[0], nil
}

// CreateUser creates a new user in the database.
func (r *authRepository) CreateUser(user *models.User) error {
	fmt.Printf("DEBUG REPO: Creating user - Email: %s, Token: %v, Expires: %v\n",
		user.Email, user.EmailVerificationToken, user.EmailVerificationExpires)

	// Execute returns (data []byte, count int64, error)
	data, count, err := r.client.From("users").Insert(user, false, "", "", "").Execute()
	if err != nil {
		fmt.Printf("ERROR REPO: Failed to create user: %v\n", err)
		return fmt.Errorf("error creating user: %w", err)
	}

	fmt.Printf("DEBUG REPO: Insert returned - Count: %d, Data: %s\n", count, string(data))
	return nil
}

// UpdateUser updates an existing user's information in the database.
func (r *authRepository) UpdateUser(user *models.User) error {
	// Execute returns (data []byte, count int64, error)
	_, _, err := r.client.From("users").Update(user, "", "").Eq("id", user.ID.String()).Execute()
	if err != nil {
		return fmt.Errorf("error updating user: %w", err)
	}
	return nil
}
