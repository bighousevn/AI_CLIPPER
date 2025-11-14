package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"ai-clipper/server2/internal/messaging/domain"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	QueueVideoProcessing   = "video_processing"
	QueueEmailNotification = "email_notification"
)

// RabbitMQPublisher implements MessagePublisher interface
type RabbitMQPublisher struct {
	client *RabbitMQClient
}

// NewRabbitMQPublisher creates a new RabbitMQ publisher
func NewRabbitMQPublisher(client *RabbitMQClient) (*RabbitMQPublisher, error) {
	publisher := &RabbitMQPublisher{
		client: client,
	}

	// Declare queues
	if err := client.DeclareQueue(QueueVideoProcessing); err != nil {
		return nil, err
	}

	if err := client.DeclareQueue(QueueEmailNotification); err != nil {
		return nil, err
	}

	return publisher, nil
}

// PublishVideoProcessing publishes a video processing request
func (p *RabbitMQPublisher) PublishVideoProcessing(fileID, userID, filePath string) error {
	message := domain.VideoProcessingMessage{
		FileID:   fileID,
		UserID:   userID,
		FilePath: filePath,
	}

	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	return p.publish(QueueVideoProcessing, body)
}

// PublishEmailNotification publishes an email notification request
func (p *RabbitMQPublisher) PublishEmailNotification(to, subject, body string) error {
	message := domain.EmailNotificationMessage{
		To:      to,
		Subject: subject,
		Body:    body,
	}

	msgBody, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	return p.publish(QueueEmailNotification, msgBody)
}

// publish sends a message to a queue
func (p *RabbitMQPublisher) publish(queueName string, body []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := p.client.GetChannel().PublishWithContext(
		ctx,
		"",        // exchange (default)
		queueName, // routing key (queue name)
		false,     // mandatory
		false,     // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent, // persistent messages
			Timestamp:    time.Now(),
		},
	)

	if err != nil {
		return fmt.Errorf("failed to publish message to %s: %w", queueName, err)
	}

	return nil
}

// Close closes the publisher
func (p *RabbitMQPublisher) Close() error {
	return nil // Connection managed by client
}
