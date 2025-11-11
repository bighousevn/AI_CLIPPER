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

	// GetSignedURL returns a signed URL for downloading a file (expires in 1 hour)
	GetSignedURL(filePath string, expiresIn int) (string, error)

	// Download downloads a file from storage and returns the file bytes
	Download(filePath string) ([]byte, error)

	// ListFiles lists all files in a folder
	ListFiles(folderPath string) ([]string, error)
}
