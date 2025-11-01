package utils

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"os"

	"github.com/google/uuid"
	storage_go "github.com/supabase-community/storage-go"
)

const (
	// 500MB
	maxVideoSize = 50 << 20
	bucketName   = "uploaded_files"
)

func UploadFileToSupabase(storageClient *storage_go.Client, file multipart.File, header *multipart.FileHeader, userID uuid.UUID) (string, error) {
	// Validate user ID
	if userID == uuid.Nil {
		return "", fmt.Errorf("invalid user ID")
	}

	// Validate file size
	if header.Size > maxVideoSize {
		return "", fmt.Errorf("file size exceeds the limit of 50MB")
	}

	// Read file content
	data, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	// Define the path for the file in the bucket (use / for cloud storage, not filepath.Join)
	fileName := fmt.Sprintf("%s-%s", userID.String(), header.Filename)
	destinationPath := fmt.Sprintf("user-%s/%s", userID.String(), fileName)

	// Upload the file to Supabase Storage
	_, err = storageClient.UploadFile(bucketName, destinationPath, bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("failed to upload file to supabase: %w", err)
	}

	// Manually construct the public URL
	supabaseURL := os.Getenv("SUPABASE_URL")
	publicURL := fmt.Sprintf("%s/storage/v1/object/public/%s/%s", supabaseURL, bucketName, destinationPath)

	return publicURL, nil
}
