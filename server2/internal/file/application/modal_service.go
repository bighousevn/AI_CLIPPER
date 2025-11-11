package application

// ModalService defines the interface for calling Modal endpoints
type ModalService interface {
	ProcessVideo(storagePath string) error // Triggers video processing
}
