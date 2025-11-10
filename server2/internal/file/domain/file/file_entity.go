package file

import (
	"time"

	"github.com/google/uuid"
)

// File represents a file entity in the domain
type File struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	FileName   string
	FilePath   string
	FileSize   int64
	MimeType   string
	UploadedAt time.Time
}

// NewFile creates a new File entity
func NewFile(userID uuid.UUID, fileName, filePath string, fileSize int64, mimeType string) *File {
	return &File{
		ID:         uuid.New(),
		UserID:     userID,
		FileName:   fileName,
		FilePath:   filePath,
		FileSize:   fileSize,
		MimeType:   mimeType,
		UploadedAt: time.Now(),
	}
}

// IsValidSize checks if the file size is within acceptable limits (e.g., 100MB)
func (f *File) IsValidSize() bool {
	maxSize := int64(100 * 1024 * 1024) // 100MB
	return f.FileSize > 0 && f.FileSize <= maxSize
}

// IsVideo checks if the file is a video
func (f *File) IsVideo() bool {
	videoTypes := []string{
		"video/mp4",
		"video/mpeg",
		"video/quicktime",
		"video/x-msvideo",
		"video/x-matroska",
	}

	for _, vt := range videoTypes {
		if f.MimeType == vt {
			return true
		}
	}
	return false
}
