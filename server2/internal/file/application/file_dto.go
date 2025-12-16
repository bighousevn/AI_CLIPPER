package application

import (
	"ai-clipper/server2/internal/file/domain/file"
	"mime/multipart"
)

// FileUploadDTO represents the data required to upload a file
type FileUploadDTO struct {
	File   multipart.File
	Header *multipart.FileHeader
	UserID string
	Config file.VideoConfig
}

// FileResponseDTO represents the response for a file
type FileResponseDTO struct {
	ID        string `json:"id"`
	FileName  string `json:"file_name"`
	FilePath  string `json:"file_path"`
	FileSize  int64  `json:"file_size"`
	Message   string `json:"message,omitempty"`
	Status    string `json:"status,omitempty"`
	ClipCount int    `json:"clip_count,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
}

// ClipResponseDTO represents the response for a clip
type ClipResponseDTO struct {
	ID             string `json:"id"`
	UploadedFileID string `json:"uploaded_file_id"`
	FilePath       string `json:"file_path"`
	DownloadURL    string `json:"download_url"`
	CreatedAt      string `json:"created_at"`
}
