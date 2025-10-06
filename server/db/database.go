package db

import (
	"bighousevn/be/models"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
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

	// Auto-migrate the User model
	err = DB.AutoMigrate(&models.User{})
	if err != nil {
		var pgErr *pgconn.PgError
		// Check if the error is a PostgreSQL error with code 42P07 (relation already exists).
		// If it is, we can safely ignore it.
		// Otherwise, it's a different error and we should return it.
		if !errors.As(err, &pgErr) || pgErr.Code != "42P07" {
			return err
		}
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