package infrastructure

import (
	"ai-clipper/server2/internal/file/domain/clip"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// GormClipRepository implements ClipRepository using GORM
type GormClipRepository struct {
	db *gorm.DB
}

// NewGormClipRepository creates a new GORM clip repository
func NewGormClipRepository(db *gorm.DB) *GormClipRepository {
	return &GormClipRepository{db: db}
}

// ClipModel represents the database model for clips
type ClipModel struct {
	ID             uuid.UUID  `gorm:"type:uuid;primaryKey"`
	UserID         uuid.UUID  `gorm:"type:uuid;not null;index"`
	UploadedFileID uuid.UUID  `gorm:"type:uuid;not null;index"`
	FilePath       string     `gorm:"type:varchar(500);not null"`
	CreatedAt      time.Time  `gorm:"not null"`
	DeletedAt      *time.Time `gorm:"index"` // For soft delete
}

// TableName specifies the table name for ClipModel
func (ClipModel) TableName() string {
	return "clips"
}

// Save persists a clip to the database
func (r *GormClipRepository) Save(c *clip.Clip) error {
	model := &ClipModel{
		ID:             c.ID,
		UserID:         c.UserID,
		UploadedFileID: c.UploadedFileID,
		FilePath:       c.FilePath,
		CreatedAt:      c.CreatedAt,
		DeletedAt:      c.DeletedAt,
	}

	if err := r.db.Create(model).Error; err != nil {
		return fmt.Errorf("failed to save clip: %w", err)
	}

	return nil
}

// FindByID retrieves a clip by its ID (excluding soft deleted)
func (r *GormClipRepository) FindByID(id uuid.UUID) (*clip.Clip, error) {
	var model ClipModel
	if err := r.db.Where("id = ? AND deleted_at IS NULL", id).First(&model).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find clip: %w", err)
	}

	return r.toDomain(&model), nil
}

// FindByUserID retrieves all clips for a user (excluding soft deleted)
func (r *GormClipRepository) FindByUserID(userID uuid.UUID) ([]*clip.Clip, error) {
	var models []ClipModel
	if err := r.db.Where("user_id = ? AND deleted_at IS NULL", userID).Order("created_at DESC").Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to find clips: %w", err)
	}

	clips := make([]*clip.Clip, len(models))
	for i, model := range models {
		clips[i] = r.toDomain(&model)
	}

	return clips, nil
}

// FindByUploadFileID retrieves all clips for an uploaded file (excluding soft deleted)
func (r *GormClipRepository) FindByUploadFileID(uploadFileID uuid.UUID) ([]*clip.Clip, error) {
	var models []ClipModel
	if err := r.db.Where("uploaded_file_id = ? AND deleted_at IS NULL", uploadFileID).Order("created_at DESC").Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to find clips: %w", err)
	}

	clips := make([]*clip.Clip, len(models))
	for i, model := range models {
		clips[i] = r.toDomain(&model)
	}

	return clips, nil
}

// SoftDelete marks a clip as deleted
func (r *GormClipRepository) SoftDelete(id uuid.UUID) error {
	now := time.Now()
	if err := r.db.Model(&ClipModel{}).Where("id = ?", id).Update("deleted_at", now).Error; err != nil {
		return fmt.Errorf("failed to soft delete clip: %w", err)
	}
	return nil
}

// toDomain converts ClipModel to domain Clip entity
func (r *GormClipRepository) toDomain(model *ClipModel) *clip.Clip {
	return &clip.Clip{
		ID:             model.ID,
		UserID:         model.UserID,
		UploadedFileID: model.UploadedFileID,
		FilePath:       model.FilePath,
		CreatedAt:      model.CreatedAt,
		DeletedAt:      model.DeletedAt,
	}
}
