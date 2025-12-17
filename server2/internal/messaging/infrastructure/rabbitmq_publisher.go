package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"ai-clipper/server2/internal/file/domain/file"
	"ai-clipper/server2/internal/messaging/domain"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	QueueVideoProcessing   = "video_processing"
	QueueEmailNotification = "email_notification"
	QueueStatusUpdate      = "status_update"
)

// RabbitMQPublisher implements MessagePublisher interface
type RabbitMQPublisher struct {
	client  *RabbitMQClient
	channel *amqp.Channel
}

// NewRabbitMQPublisher creates a new RabbitMQ publisher
func NewRabbitMQPublisher(client *RabbitMQClient) (*RabbitMQPublisher, error) {
	// Create dedicated channel for publisher
	channel, err := client.NewChannel()
	if err != nil {
		return nil, fmt.Errorf("failed to create publisher channel: %w", err)
	}

	publisher := &RabbitMQPublisher{
		client:  client,
		channel: channel,
	}

	// Declare queues
	if err := client.DeclareQueue(channel, QueueVideoProcessing); err != nil {
		channel.Close()
		return nil, err
	}

	if err := client.DeclareQueue(channel, QueueEmailNotification); err != nil {
		channel.Close()
		return nil, err
	}

	if err := client.DeclareQueue(channel, QueueStatusUpdate); err != nil {
		channel.Close()
		return nil, err
	}

	return publisher, nil
}

// PublishVideoProcessing publishes a video processing request
func (p *RabbitMQPublisher) PublishVideoProcessing(fileID, userID, filePath string, config file.VideoConfig) error {
	message := domain.VideoProcessingMessage{
		FileID:   fileID,
		UserID:   userID,
		FilePath: filePath,
		Config:   config,
	}

	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	return p.publish(QueueVideoProcessing, body)
}

// PublishEmailNotification publishes an email notification request
func (p *RabbitMQPublisher) PublishEmailNotification(emailType, to, username, content string) error {
	message := domain.EmailNotificationMessage{
		Type:     emailType,
		To:       to,
		Username: username,
		Content:  content,
	}

	msgBody, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	return p.publish(QueueEmailNotification, msgBody)
}

// PublishStatusUpdate publishes a status update message
func (p *RabbitMQPublisher) PublishStatusUpdate(fileID, userID, status string, clipCount int) error {
	message := domain.StatusUpdateMessage{
		FileID:    fileID,
		UserID:    userID,
		Status:    status,
		ClipCount: clipCount,
	}

	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	return p.publish(QueueStatusUpdate, body)
}

// publish sends a message to a queue
func (p *RabbitMQPublisher) publish(queueName string, body []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := p.channel.PublishWithContext(
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

// Close closes the publisher channel
func (p *RabbitMQPublisher) Close() error {
	if p.channel != nil {
		return p.channel.Close()
	}
	return nil
}
