package application

import "ai-clipper/server2/internal/file/domain/file"

// ModalService defines the interface for calling Modal endpoints
type ModalService interface {
	SendVideoToModal(storagePath string, config file.VideoConfig) error // Sends video to Modal API for processing
}
