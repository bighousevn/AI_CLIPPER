package infrastructure

import (
	"fmt"
	"log"

	"ai-clipper/server2/internal/messaging/domain"

	amqp "github.com/rabbitmq/amqp091-go"
)

// RabbitMQConsumer implements MessageConsumer interface
type RabbitMQConsumer struct {
	client           *RabbitMQClient
	videoChannel     *amqp.Channel
	emailChannel     *amqp.Channel
	statusChannel    *amqp.Channel
	videoQoSPrefetch int
	emailQoSPrefetch int
}

// NewRabbitMQConsumer creates a new RabbitMQ consumer with separate channels
func NewRabbitMQConsumer(client *RabbitMQClient) (*RabbitMQConsumer, error) {
	// Create separate channel for video processing
	videoChannel, err := client.NewChannel()
	if err != nil {
		return nil, fmt.Errorf("failed to create video channel: %w", err)
	}

	// Create separate channel for email notification
	emailChannel, err := client.NewChannel()
	if err != nil {
		videoChannel.Close()
		return nil, fmt.Errorf("failed to create email channel: %w", err)
	}

	// Create separate channel for status updates
	statusChannel, err := client.NewChannel()
	if err != nil {
		videoChannel.Close()
		emailChannel.Close()
		return nil, fmt.Errorf("failed to create status channel: %w", err)
	}

	consumer := &RabbitMQConsumer{
		client:           client,
		videoChannel:     videoChannel,
		emailChannel:     emailChannel,
		statusChannel:    statusChannel,
		videoQoSPrefetch: 30, // Modal allows 30 concurrent instances
		emailQoSPrefetch: 3,  // For local testing
	}

	// Declare queues
	if err := client.DeclareQueue(videoChannel, QueueVideoProcessing); err != nil {
		videoChannel.Close()
		emailChannel.Close()
		statusChannel.Close()
		return nil, err
	}

	if err := client.DeclareQueue(emailChannel, QueueEmailNotification); err != nil {
		videoChannel.Close()
		emailChannel.Close()
		statusChannel.Close()
		return nil, err
	}

	if err := client.DeclareQueue(statusChannel, QueueStatusUpdate); err != nil {
		videoChannel.Close()
		emailChannel.Close()
		statusChannel.Close()
		return nil, err
	}

	log.Printf("RabbitMQ Consumer created with QoS: video=%d, email=%d", consumer.videoQoSPrefetch, consumer.emailQoSPrefetch)
	return consumer, nil
}

// ConsumeVideoProcessing starts consuming video processing messages
func (c *RabbitMQConsumer) ConsumeVideoProcessing(handler domain.MessageHandler) error {
	return c.consume(c.videoChannel, QueueVideoProcessing, c.videoQoSPrefetch, handler)
}

// ConsumeEmailNotification starts consuming email notification messages
func (c *RabbitMQConsumer) ConsumeEmailNotification(handler domain.MessageHandler) error {
	return c.consume(c.emailChannel, QueueEmailNotification, c.emailQoSPrefetch, handler)
}

// ConsumeStatusUpdate starts consuming status update messages
func (c *RabbitMQConsumer) ConsumeStatusUpdate(handler domain.MessageHandler) error {
	// Status updates are fast and lightweight, prefetch 100 is fine
	return c.consume(c.statusChannel, QueueStatusUpdate, 100, handler)
}

// consume starts consuming messages from a queue
func (c *RabbitMQConsumer) consume(channel *amqp.Channel, queueName string, prefetchCount int, handler domain.MessageHandler) error {
	// Set QoS (prefetch count based on queue type)
	err := channel.Qos(
		prefetchCount, // prefetch count (30 for video, 3 for email)
		0,             // prefetch size
		false,         // global
	)
	if err != nil {
		return fmt.Errorf("failed to set QoS: %w", err)
	}

	log.Printf("Set QoS prefetch=%d for queue: %s", prefetchCount, queueName)

	msgs, err := channel.Consume(
		queueName, // queue
		"",        // consumer tag (auto-generated)
		false,     // auto-ack (manual ack for reliability)
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer for %s: %w", queueName, err)
	}

	log.Printf("Started consuming messages from queue: %s", queueName)

	// Process messages
	go func() {
		for msg := range msgs {
			// Spawn a goroutine for each message to process concurrently
			// The number of concurrent goroutines is limited by QoS prefetch count
			go func(d amqp.Delivery) {
				log.Printf("Received message from %s: %s", queueName, string(d.Body))

				// Call handler
				if err := handler(d.Body); err != nil {
					log.Printf("Error processing message: %v", err)
					// Don't requeue on error - send to DLQ or drop
					if nackErr := d.Nack(false, false); nackErr != nil {
						log.Printf("Failed to nack message: %v", nackErr)
					}
					log.Printf("Message rejected and NOT requeued (dropped)")
				} else {
					// Acknowledge successful processing
					if ackErr := d.Ack(false); ackErr != nil {
						log.Printf("Failed to ack message: %v", ackErr)
					}
					log.Printf("Message processed successfully")
				}
			}(msg)
		}
	}()

	return nil
}

// Close closes the consumer channels
func (c *RabbitMQConsumer) Close() error {
	if c.videoChannel != nil {
		if err := c.videoChannel.Close(); err != nil {
			log.Printf("Error closing video channel: %v", err)
		}
	}

	if c.emailChannel != nil {
		if err := c.emailChannel.Close(); err != nil {
			log.Printf("Error closing email channel: %v", err)
		}
	}

	if c.statusChannel != nil {
		if err := c.statusChannel.Close(); err != nil {
			log.Printf("Error closing status channel: %v", err)
		}
	}

	log.Println("RabbitMQ consumer channels closed")
	return nil
}
