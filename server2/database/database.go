package database

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
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

// RunMigrations applies database migrations.
func RunMigrations(dsn string) error {
	// Get the current working directory
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	// Assuming the app is run from 'server2/cmd/server', we need to go up two levels
	// to find 'database/migrations'.
	// Path: [wd]/../../database/migrations
	migrationsPath := filepath.Join(wd, "..", "..", "database", "migrations")
	
	// Convert backslashes to forward slashes for the file:// URL scheme on Windows
	migrationsPath = filepath.ToSlash(filepath.Clean(migrationsPath))

	log.Printf("Looking for migrations in: %s", migrationsPath)

	m, err := migrate.New(
		"file://"+migrationsPath,
		dsn,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	if err := m.Up(); err != nil {
		if err == migrate.ErrNoChange {
			log.Println("Database is up to date.")
			return nil
		}
		
		// Handle "Dirty database" error
		if err.Error() == "Dirty database version 1. Fix and force version." {
			log.Println("Database is dirty. Forcing version 1 and retrying...")
			if forceErr := m.Force(1); forceErr != nil {
				return fmt.Errorf("failed to force version 1: %w", forceErr)
			}
			// Retry Up after forcing
			if upErr := m.Up(); upErr != nil && upErr != migrate.ErrNoChange {
				return fmt.Errorf("failed to retry migrate up after forcing: %w", upErr)
			}
			log.Println("Database migrations applied successfully after forcing.")
			return nil
		}

		return fmt.Errorf("failed to run migrate up: %w", err)
	}

	log.Println("Database migrations applied successfully.")
	return nil
}
