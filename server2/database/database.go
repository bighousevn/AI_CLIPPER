package database

import (
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// User model maps to the 'users' table in Supabase.
type User struct {
	ID                       uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Username                 *string   `gorm:"unique"`
	Email                    string    `gorm:"not null;unique"`
	PasswordHash             *string
	Credits                  int64     `gorm:"default:10"`
	StripeCustomerID         *string   `gorm:"unique"`
	RefreshToken             *string   `gorm:"unique"`
	PasswordResetToken       *string
	PasswordResetExpires     *time.Time
	IsEmailVerified          bool `gorm:"default:false"`
	EmailVerificationToken   *string
	EmailVerificationExpires *time.Time
	CreatedAt                *time.Time
	UpdatedAt                *time.Time
	DeletedAt                gorm.DeletedAt `gorm:"index"`
}

// TableName specifies the table name for the User model.
func (User) TableName() string {
	return "users"
}

// UploadedFile model maps to the 'uploaded_files' table.
type UploadedFile struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID      uuid.UUID `gorm:"type:uuid;not null"`
	User        User      `gorm:"foreignKey:UserID"`
	DisplayName string    `gorm:"not null;default:''"`
	Status      string    `gorm:"not null;default:'queue'"`
	Uploaded    bool      `gorm:"not null"`
	CreatedAt   time.Time `gorm:"not null;default:now()"`
	UpdatedAt   *time.Time
}

// TableName specifies the table name for the UploadedFile model.
func (UploadedFile) TableName() string {
	return "uploaded_files"
}

// Clip model maps to the 'clips' table.
type Clip struct {
	ID             uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID         uuid.UUID `gorm:"type:uuid;not null"`
	User           User      `gorm:"foreignKey:UserID"`
	UploadedFileID uuid.UUID `gorm:"type:uuid;not null"`
	UploadedFile   UploadedFile `gorm:"foreignKey:UploadedFileID"`
	CreatedAt      time.Time `gorm:"not null;default:now()"`
	UpdatedAt      *time.Time
}

// TableName specifies the table name for the Clip model.
func (Clip) TableName() string {
	return "clips"
}

// InitDatabase initializes the database connection.
func InitDatabase(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		PrepareStmt: false, // Disable prepared statement caching to avoid conflicts with Supabase connection pooling
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	log.Println("Database connection established.")
	return db, nil
}
