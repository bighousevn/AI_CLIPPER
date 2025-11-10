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
