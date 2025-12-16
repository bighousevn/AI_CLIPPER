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
	Status     string
	ClipCount  int
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// NewFile creates a new File entity
func NewFile(userID uuid.UUID, fileName, filePath string, fileSize int64, mimeType string) *File {
	now := time.Now()
	return &File{
		ID:         uuid.New(),
		UserID:     userID,
		FileName:   fileName,
		FilePath:   filePath,
		FileSize:   fileSize,
		MimeType:   mimeType,
		UploadedAt: now,
		Status:     "uploaded",
		ClipCount:  0,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

// IsValidSize checks if the file size is within acceptable limits (e.g., 100MB)
func (f *File) IsValidSize() bool {
	maxSize := int64(50 * 1024 * 1024) // 50MB
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
