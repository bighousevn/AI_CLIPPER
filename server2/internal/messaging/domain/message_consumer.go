package domain

type MessageHandler func(body []byte) error

type MessageConsumer interface {
	ConsumeVideoProcessing(handler MessageHandler) error
	ConsumeEmailNotification(handler MessageHandler) error
	Close() error
}
