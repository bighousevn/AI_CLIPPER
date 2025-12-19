package clip

import (
	"time"

	"github.com/google/uuid"
)

// Clip represents a video clip entity
type Clip struct {
	ID             uuid.UUID
	UserID         uuid.UUID
	UploadedFileID uuid.UUID
	SourceName     string // Name of the original uploaded video
	FilePath       string // Path to clip in Supabase Storage
	CreatedAt      time.Time
	DeletedAt      *time.Time // For soft delete
}

// NewClip creates a new clip entity
func NewClip(userID, uploadedFileID uuid.UUID, sourceName, filePath string) *Clip {
	return &Clip{
		ID:             uuid.New(),
		UserID:         userID,
		UploadedFileID: uploadedFileID,
		SourceName:     sourceName,
		FilePath:       filePath,
		CreatedAt:      time.Now(),
	}
}

// IsDeleted checks if the clip is soft deleted
func (c *Clip) IsDeleted() bool {
	return c.DeletedAt != nil
}

// SoftDelete marks the clip as deleted
func (c *Clip) SoftDelete() {
	now := time.Now()
	c.DeletedAt = &now
}

// ClipRepository defines the interface for clip persistence
type ClipRepository interface {
	Save(clip *Clip) error
	FindByID(id uuid.UUID) (*Clip, error)
	FindByUserID(userID uuid.UUID) ([]*Clip, error)
	FindByUploadFileID(uploadFileID uuid.UUID) ([]*Clip, error)
	SoftDelete(id uuid.UUID) error
}
