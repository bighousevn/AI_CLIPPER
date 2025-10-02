package db

import (
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
