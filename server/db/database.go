package db

import (
	"bighousevn/be/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB // Global variable to hold the GORM database connection

// InitDB initializes the database connection using GORM.
func InitDB(dataSourceName string) error {
	var err error
	DB, err = gorm.Open(postgres.Open(dataSourceName), &gorm.Config{
		PrepareStmt: true,
	})
	if err != nil {
		return err
	}

	return nil
}

// CloseDB closes the database connection pool.
func CloseDB() {
	sqlDB, err := DB.DB()
	sqlDB.Exec("DEALLOCATE ALL")
	if err == nil && sqlDB != nil {
		sqlDB.Close()
	}
}

// GetUserByID retrieves a user from the database by their ID.
func GetUserByID(id int64) (*models.User, error) {
	var user models.User
	result := DB.First(&user, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

// GetUserByEmail retrieves a user from the database by email using GORM.
func GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	// GORM returns gorm.ErrRecordNotFound if no record is found
	result := DB.Where("email = ?", email).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

// GetUserByPasswordResetToken retrieves a user by their password reset token.
func GetUserByPasswordResetToken(token string) (*models.User, error) {
	var user models.User
	result := DB.Where("password_reset_token = ?", token).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

// GetUserByEmailVerificationToken retrieves a user by their email verification token.
func GetUserByEmailVerificationToken(token string) (*models.User, error) {
	var user models.User
	result := DB.Where("email_verification_token = ?", token).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

// CreateUser adds a new user to the database using GORM.
func CreateUser(user *models.User) error {
	// The Create function will insert the new record and update the user object with the ID.
	result := DB.Create(user)
	return result.Error
}

// UpdateUser saves the changes to a user model.
func UpdateUser(user *models.User) error {
	result := DB.Save(user) // GORM's Save updates all fields based on the primary key
	return result.Error
}
