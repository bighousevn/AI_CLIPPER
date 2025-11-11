package infrastructure

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"

	"github.com/google/uuid"
	storage "github.com/supabase-community/storage-go"
)

// SupabaseStorageService implements StorageService using Supabase Storage
type SupabaseStorageService struct {
	client     *storage.Client
	bucketName string
}

// NewSupabaseStorageService creates a new Supabase storage service
func NewSupabaseStorageService(client *storage.Client, bucketName string) *SupabaseStorageService {
	return &SupabaseStorageService{
		client:     client,
		bucketName: bucketName,
	}
}

// Upload uploads a file to Supabase Storage
func (s *SupabaseStorageService) Upload(file multipart.File, header *multipart.FileHeader, userID string) (string, error) {
	// Read file content
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	// Reset file pointer
	file.Seek(0, 0)

	// Create file path: user-{userID}/{uuid}-{filename}
	fileID := uuid.New().String()
	filePath := fmt.Sprintf("user-%s/%s-%s", userID, fileID, header.Filename)

	// Upload to Supabase
	_, err = s.client.UploadFile(s.bucketName, filePath, bytes.NewReader(fileBytes))
	if err != nil {
		return "", fmt.Errorf("failed to upload to Supabase: %w", err)
	}

	return filePath, nil
}

// Delete deletes a file from Supabase Storage
func (s *SupabaseStorageService) Delete(filePath string) error {
	_, err := s.client.RemoveFile(s.bucketName, []string{filePath})
	if err != nil {
		return fmt.Errorf("failed to delete file from Supabase: %w", err)
	}
	return nil
}

// GetURL returns the public URL of a file
func (s *SupabaseStorageService) GetURL(filePath string) (string, error) {
	resp := s.client.GetPublicUrl(s.bucketName, filePath)
	return resp.SignedURL, nil
}

// GetSignedURL returns a signed URL for downloading a file with expiration
func (s *SupabaseStorageService) GetSignedURL(filePath string, expiresIn int) (string, error) {
	resp, err := s.client.CreateSignedUrl(s.bucketName, filePath, expiresIn)
	if err != nil {
		return "", fmt.Errorf("failed to create signed URL: %w", err)
	}
	return resp.SignedURL, nil
}

// Download downloads a file from Supabase Storage
func (s *SupabaseStorageService) Download(filePath string) ([]byte, error) {
	fileBytes, err := s.client.DownloadFile(s.bucketName, filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to download file from Supabase: %w", err)
	}
	return fileBytes, nil
}

// ListFiles lists all files in a folder
func (s *SupabaseStorageService) ListFiles(folderPath string) ([]string, error) {
	// List files in the folder
	files, err := s.client.ListFiles(s.bucketName, folderPath, storage.FileSearchOptions{
		Limit:  100,
		Offset: 0,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list files from Supabase: %w", err)
	}

	// Extract file paths (only actual files, not directories)
	var filePaths []string
	for _, file := range files {
		// Skip if it's a directory or empty name
		if file.Name == "" {
			continue
		}

		// Skip Supabase placeholder files
		if file.Name == ".emptyFolderPlaceholder" {
			continue
		}

		// Check if it's a file (has ID means it's a real file)
		// Directories in Supabase don't have ID
		if file.Id == "" {
			continue
		}

		// Build full path: folderPath/filename
		fullPath := fmt.Sprintf("%s/%s", folderPath, file.Name)
		filePaths = append(filePaths, fullPath)
	}

	return filePaths, nil
}
