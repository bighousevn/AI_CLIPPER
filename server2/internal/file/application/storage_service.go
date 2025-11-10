package application

import "mime/multipart"

// StorageService defines the interface for file storage operations
type StorageService interface {
	// Upload uploads a file to the storage and returns the file path
	Upload(file multipart.File, header *multipart.FileHeader, userID string) (string, error)

	// Delete deletes a file from the storage
	Delete(filePath string) error

	// GetURL returns the public URL of a file
	GetURL(filePath string) (string, error)
}
