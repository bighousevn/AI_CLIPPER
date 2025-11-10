package file

import "github.com/google/uuid"

// FileRepository defines the interface for file data access
type FileRepository interface {
	// Save persists a file metadata to the database
	Save(file *File) error

	// FindByID retrieves a file by its ID
	FindByID(id uuid.UUID) (*File, error)

	// FindByUserID retrieves all files uploaded by a user
	FindByUserID(userID uuid.UUID) ([]*File, error)

	// Delete removes a file metadata from the database
	Delete(id uuid.UUID) error
}
