package infrastructure

import (
	"ai-clipper/server2/internal/file/domain/file"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// GormFileRepository implements FileRepository using GORM
type GormFileRepository struct {
	db *gorm.DB
}

// NewGormFileRepository creates a new GORM file repository
func NewGormFileRepository(db *gorm.DB) *GormFileRepository {
	return &GormFileRepository{db: db}
}

// FileModel represents the database model for files
type FileModel struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID      uuid.UUID  `gorm:"type:uuid;not null;index"`
	DisplayName string     `gorm:"not null;default:''"`         // Original filename
	Status      string     `gorm:"not null;default:'uploaded'"` // Status: uploaded, processing, completed
	Uploaded    bool       `gorm:"not null;default:true"`       // Upload completion flag
	FilePath    string     `gorm:"type:varchar(500)"`           // Storage path
	FileSize    int64      `gorm:"default:0"`                   // File size in bytes
	MimeType    string     `gorm:"type:varchar(100)"`           // MIME type
	ClipCount   int        `gorm:"default:0"`                   // Number of generated clips
	CreatedAt   *time.Time `gorm:"not null;default:now()"`
	UpdatedAt   *time.Time
}

// TableName specifies the table name for FileModel
func (FileModel) TableName() string {
	return "uploaded_files"
}

// Save persists a file metadata to the database
func (r *GormFileRepository) Save(f *file.File) error {
	now := time.Now()
	model := &FileModel{
		ID:          f.ID,
		UserID:      f.UserID,
		DisplayName: f.FileName, // Map FileName to DisplayName
		Status:      f.Status,   // Use status from entity
		Uploaded:    true,       // Mark as uploaded
		FilePath:    f.FilePath,
		FileSize:    f.FileSize,
		MimeType:    f.MimeType,
		CreatedAt:   &now,
		UpdatedAt:   &now,
	}

	// Debug log
	fmt.Printf("Attempting to save file: ID=%s, UserID=%s, Table=%s\n", model.ID, model.UserID, model.TableName())

	if err := r.db.Create(model).Error; err != nil {
		fmt.Printf("Database error details: %v\n", err)
		return fmt.Errorf("failed to save file: %w", err)
	}

	return nil
}

// FindByID retrieves a file by its ID
func (r *GormFileRepository) FindByID(id uuid.UUID) (*file.File, error) {
	var model FileModel
	if err := r.db.Where("id = ?", id).First(&model).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find file: %w", err)
	}

	return r.toDomain(&model), nil
}

// FindByUserID retrieves all files uploaded by a user
func (r *GormFileRepository) FindByUserID(userID uuid.UUID) ([]*file.File, error) {
	var models []FileModel
	if err := r.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to find files: %w", err)
	}

	files := make([]*file.File, len(models))
	for i, model := range models {
		files[i] = r.toDomain(&model)
	}

	return files, nil
}

// Delete removes a file metadata from the database
func (r *GormFileRepository) Delete(id uuid.UUID) error {
	if err := r.db.Delete(&FileModel{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

// UpdateStatus updates the processing status and clip count of a file
func (r *GormFileRepository) UpdateStatus(id uuid.UUID, status string, clipCount int) error {
	result := r.db.Model(&FileModel{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":     status,
		"clip_count": clipCount,
		"updated_at": time.Now(),
	})

	if result.Error != nil {
		return fmt.Errorf("failed to update file status: %w", result.Error)
	}

	return nil
}

// toDomain converts FileModel to domain File entity
func (r *GormFileRepository) toDomain(model *FileModel) *file.File {
	return &file.File{
		ID:       model.ID,
		UserID:   model.UserID,
		FileName: model.DisplayName, // Map DisplayName back to FileName
		FilePath:  model.FilePath,
		FileSize:  model.FileSize,
		MimeType:  model.MimeType,
		Status:    model.Status,
		ClipCount: model.ClipCount,
		CreatedAt: *model.CreatedAt,
		UpdatedAt: *model.UpdatedAt,
	}
}
