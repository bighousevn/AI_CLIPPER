package file

import "github.com/google/uuid"

type FileRepository interface {
	Save(file *File) error
	FindByID(id uuid.UUID) (*File, error)
	FindByUserID(userID uuid.UUID) ([]*File, error)
	Delete(id uuid.UUID) error
	UpdateStatus(id uuid.UUID, status string, clipCount int) error
}
