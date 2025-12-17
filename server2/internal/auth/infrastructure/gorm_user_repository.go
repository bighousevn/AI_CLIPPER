package infrastructure

import (
	userDomain "ai-clipper/server2/internal/auth/domain/user"
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type User struct {
	gorm.Model
	ID                       uuid.UUID `gorm:"type:uuid;primary_key;"`
	Username                 string
	Email                    string `gorm:"uniqueIndex"`
	PasswordHash             string
	Credits                  int
	StripeCustomerID         *string
	RefreshToken             *string
	PasswordResetToken       *string `gorm:"index"`
	PasswordResetExpires     *time.Time
	IsEmailVerified          bool
	EmailVerificationToken   *string `gorm:"index"`
	EmailVerificationExpires *time.Time
}

func toDomain(gormUser *User) *userDomain.User {
	return &userDomain.User{
		ID:                       gormUser.ID,
		Username:                 gormUser.Username,
		Email:                    gormUser.Email,
		PasswordHash:             gormUser.PasswordHash,
		Credits:                  gormUser.Credits,
		StripeCustomerID:         gormUser.StripeCustomerID,
		RefreshToken:             gormUser.RefreshToken,
		PasswordResetToken:       gormUser.PasswordResetToken,
		PasswordResetExpires:     gormUser.PasswordResetExpires,
		IsEmailVerified:          gormUser.IsEmailVerified,
		EmailVerificationToken:   gormUser.EmailVerificationToken,
		EmailVerificationExpires: gormUser.EmailVerificationExpires,
		CreatedAt:                gormUser.CreatedAt,
		UpdatedAt:                gormUser.UpdatedAt,
		DeletedAt:                &gormUser.DeletedAt.Time,
	}
}

func fromDomain(domainUser *userDomain.User) *User {
	return &User{
		ID:                       domainUser.ID,
		Username:                 domainUser.Username,
		Email:                    domainUser.Email,
		PasswordHash:             domainUser.PasswordHash,
		Credits:                  domainUser.Credits,
		StripeCustomerID:         domainUser.StripeCustomerID,
		RefreshToken:             domainUser.RefreshToken,
		PasswordResetToken:       domainUser.PasswordResetToken,
		PasswordResetExpires:     domainUser.PasswordResetExpires,
		IsEmailVerified:          domainUser.IsEmailVerified,
		EmailVerificationToken:   domainUser.EmailVerificationToken,
		EmailVerificationExpires: domainUser.EmailVerificationExpires,
	}
}

type GormUserRepository struct {
	db *gorm.DB
}

func NewGormUserRepository(db *gorm.DB) userDomain.UserRepository {
	return &GormUserRepository{db: db}
}

func (r *GormUserRepository) findByField(ctx context.Context, field string, value interface{}) (*userDomain.User, error) {
	var gormUser User
	if err := r.db.WithContext(ctx).Where(field+" = ?", value).First(&gormUser).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Return nil, nil for not found, so the use case can handle it
		}
		return nil, err
	}
	return toDomain(&gormUser), nil
}

func (r *GormUserRepository) FindByID(ctx context.Context, id uuid.UUID) (*userDomain.User, error) {
	return r.findByField(ctx, "id", id)
}

func (r *GormUserRepository) FindByEmail(ctx context.Context, email string) (*userDomain.User, error) {
	return r.findByField(ctx, "email", email)
}

func (r *GormUserRepository) FindByPasswordResetToken(ctx context.Context, token string) (*userDomain.User, error) {
	return r.findByField(ctx, "password_reset_token", token)
}

func (r *GormUserRepository) FindByEmailVerificationToken(ctx context.Context, token string) (*userDomain.User, error) {
	return r.findByField(ctx, "email_verification_token", token)
}

func (r *GormUserRepository) FindByRefreshToken(ctx context.Context, token string) (*userDomain.User, error) {
	return r.findByField(ctx, "refresh_token", token)
}

func (r *GormUserRepository) FindByStripeCustomerID(ctx context.Context, stripeCustomerID string) (*userDomain.User, error) {
	return r.findByField(ctx, "stripe_customer_id", stripeCustomerID)
}

func (r *GormUserRepository) Save(ctx context.Context, user *userDomain.User) error {
	gormUser := fromDomain(user);

	// Use Clauses(clause.OnConflict) to handle both create and update (upsert)
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns(allColumnsExceptID()),
	}).Create(gormUser).Error
}

func (r *GormUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&User{}, id).Error
}

// Helper function to get all column names for the upsert operation
func allColumnsExceptID() []string {
	return []string{
		"username", "email", "password_hash", "credits", "stripe_customer_id",
		"refresh_token", "password_reset_token", "password_reset_expires",
		"is_email_verified", "email_verification_token", "email_verification_expires",
		"created_at", "updated_at", "deleted_at",
	}
}