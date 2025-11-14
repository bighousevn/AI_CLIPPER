package domain

// MessageHandler is a function that processes a message
type MessageHandler func(body []byte) error

// MessageConsumer defines the interface for consuming messages from a message broker
// This interface belongs to the domain layer and is implementation-agnostic
type MessageConsumer interface {
	// ConsumeVideoProcessing starts consuming video processing messages
	ConsumeVideoProcessing(handler MessageHandler) error

	// ConsumeEmailNotification starts consuming email notification messages
	ConsumeEmailNotification(handler MessageHandler) error

	// Close closes the consumer connection
	Close() error
}
