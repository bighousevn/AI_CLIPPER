package infrastructure

import (
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// RabbitMQClient manages the connection to RabbitMQ
type RabbitMQClient struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	url     string
}

// NewRabbitMQClient creates a new RabbitMQ client
func NewRabbitMQClient(url string) (*RabbitMQClient, error) {
	client := &RabbitMQClient{
		url: url,
	}

	if err := client.connect(); err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	return client, nil
}

// connect establishes connection to RabbitMQ with retry logic
func (c *RabbitMQClient) connect() error {
	var err error

	// Retry connection up to 5 times
	for i := 0; i < 5; i++ {
		c.conn, err = amqp.Dial(c.url)
		if err == nil {
			break
		}

		log.Printf("Failed to connect to RabbitMQ (attempt %d/5): %v", i+1, err)
		time.Sleep(time.Second * 2)
	}

	if err != nil {
		return fmt.Errorf("failed to dial RabbitMQ after retries: %w", err)
	}

	c.channel, err = c.conn.Channel()
	if err != nil {
		c.conn.Close()
		return fmt.Errorf("failed to open channel: %w", err)
	}

	log.Println("Successfully connected to RabbitMQ")
	return nil
}

// DeclareQueue declares a queue (idempotent operation)
func (c *RabbitMQClient) DeclareQueue(queueName string) error {
	_, err := c.channel.QueueDeclare(
		queueName, // name
		true,      // durable (survive broker restart)
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue %s: %w", queueName, err)
	}

	return nil
}

// GetChannel returns the AMQP channel
func (c *RabbitMQClient) GetChannel() *amqp.Channel {
	return c.channel
}

// Close closes the connection and channel
func (c *RabbitMQClient) Close() error {
	if c.channel != nil {
		if err := c.channel.Close(); err != nil {
			log.Printf("Error closing channel: %v", err)
		}
	}

	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			log.Printf("Error closing connection: %v", err)
		}
	}

	log.Println("RabbitMQ connection closed")
	return nil
}

// IsConnected checks if the connection is alive
func (c *RabbitMQClient) IsConnected() bool {
	return c.conn != nil && !c.conn.IsClosed()
}
