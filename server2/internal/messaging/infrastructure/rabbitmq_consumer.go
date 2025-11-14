package infrastructure

import (
	"fmt"
	"log"

	"ai-clipper/server2/internal/messaging/domain"
)

// RabbitMQConsumer implements MessageConsumer interface
type RabbitMQConsumer struct {
	client *RabbitMQClient
}

// NewRabbitMQConsumer creates a new RabbitMQ consumer
func NewRabbitMQConsumer(client *RabbitMQClient) (*RabbitMQConsumer, error) {
	consumer := &RabbitMQConsumer{
		client: client,
	}

	// Declare queues
	if err := client.DeclareQueue(QueueVideoProcessing); err != nil {
		return nil, err
	}

	if err := client.DeclareQueue(QueueEmailNotification); err != nil {
		return nil, err
	}

	return consumer, nil
}

// ConsumeVideoProcessing starts consuming video processing messages
func (c *RabbitMQConsumer) ConsumeVideoProcessing(handler domain.MessageHandler) error {
	return c.consume(QueueVideoProcessing, handler)
}

// ConsumeEmailNotification starts consuming email notification messages
func (c *RabbitMQConsumer) ConsumeEmailNotification(handler domain.MessageHandler) error {
	return c.consume(QueueEmailNotification, handler)
}

// consume starts consuming messages from a queue
func (c *RabbitMQConsumer) consume(queueName string, handler domain.MessageHandler) error {
	// Set QoS (prefetch 1 message at a time for fair distribution)
	err := c.client.GetChannel().Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		return fmt.Errorf("failed to set QoS: %w", err)
	}

	msgs, err := c.client.GetChannel().Consume(
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
			log.Printf("Received message from %s: %s", queueName, string(msg.Body))

			// Call handler
			if err := handler(msg.Body); err != nil {
				log.Printf("Error processing message: %v", err)
				// Don't requeue on error - send to DLQ or drop
				// This prevents infinite requeue loops for invalid messages
				msg.Nack(false, false) // false = don't requeue
				log.Printf("Message rejected and NOT requeued (dropped)")
			} else {
				// Acknowledge successful processing
				msg.Ack(false)
				log.Printf("Message processed successfully")
			}
		}
	}()

	return nil
}

// Close closes the consumer
func (c *RabbitMQConsumer) Close() error {
	return nil // Connection managed by client
}
