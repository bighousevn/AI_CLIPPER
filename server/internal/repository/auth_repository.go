package repository

import (
	"bighousevn/be/internal/models"

	"gorm.io/gorm"
)

type AuthRepository interface {
	GetUserByEmail(email string) (*models.User, error)
	CreateUser(user *models.User) error
	UpdateUser(user *models.User) error
	GetUserByID(id uint) (*models.User, error)
	GetUserByPasswordResetToken(token string) (*models.User, error)
	GetUserByEmailVerificationToken(token string) (*models.User, error)
	GetUserByRefreshToken(token string) (*models.User, error)
}

type authRepository struct {
	db *gorm.DB
}

func NewAuthRepository(db *gorm.DB) AuthRepository {
	return &authRepository{db: db}
}
func (r *authRepository) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *authRepository) GetUserByID(id uint) (*models.User, error) {
	var user models.User
	if err := r.db.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *authRepository) GetUserByPasswordResetToken(token string) (*models.User, error) {
	var user models.User
	if err := r.db.Where("password_reset_token = ?", token).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *authRepository) GetUserByEmailVerificationToken(token string) (*models.User, error) {
	var user models.User
	if err := r.db.Where("email_verification_token = ?", token).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *authRepository) GetUserByRefreshToken(token string) (*models.User, error) {
	var user models.User
	if err := r.db.Where("refresh_token = ?", token).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *authRepository) CreateUser(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *authRepository) UpdateUser(user *models.User) error {
	return r.db.Save(user).Error
}
