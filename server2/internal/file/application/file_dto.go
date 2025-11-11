package application

import "mime/multipart"

// FileUploadDTO represents the data required to upload a file
type FileUploadDTO struct {
	File   multipart.File
	Header *multipart.FileHeader
	UserID string
}

// FileResponseDTO represents the response for a file
type FileResponseDTO struct {
	ID       string `json:"id"`
	FileName string `json:"file_name"`
	FilePath string `json:"file_path"`
	FileSize int64  `json:"file_size"`
	Message  string `json:"message,omitempty"`
}

// ClipResponseDTO represents the response for a clip
type ClipResponseDTO struct {
	ID             string `json:"id"`
	UploadedFileID string `json:"uploaded_file_id"`
	FilePath       string `json:"file_path"`
	DownloadURL    string `json:"download_url"`
	CreatedAt      string `json:"created_at"`
}
