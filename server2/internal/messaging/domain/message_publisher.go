package domain

// MessagePublisher defines the interface for publishing messages to a message broker
// This interface belongs to the domain layer and is implementation-agnostic
type MessagePublisher interface {
	// PublishVideoProcessing publishes a video processing request
	PublishVideoProcessing(fileID, userID, filePath string) error

	// PublishEmailNotification publishes an email notification request
	PublishEmailNotification(to, subject, body string) error

	// Close closes the publisher connection
	Close() error
}

// VideoProcessingMessage represents a video processing request message
type VideoProcessingMessage struct {
	FileID   string `json:"file_id"`
	UserID   string `json:"user_id"`
	FilePath string `json:"file_path"`
}

// EmailNotificationMessage represents an email notification message
type EmailNotificationMessage struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}
