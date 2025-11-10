package application

// ModalService defines the interface for calling Modal endpoints
type ModalService interface {
	ProcessVideo(storagePath string) ([]string, error) // Returns list of clip paths
}
