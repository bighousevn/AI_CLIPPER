package infrastructure

import (
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// RabbitMQClient manages the connection to RabbitMQ
type RabbitMQClient struct {
	conn *amqp.Connection
	url  string
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

	log.Println("Successfully connected to RabbitMQ")
	return nil
}

// NewChannel creates a new channel from the connection
func (c *RabbitMQClient) NewChannel() (*amqp.Channel, error) {
	if c.conn == nil || c.conn.IsClosed() {
		return nil, fmt.Errorf("connection is not open")
	}

	channel, err := c.conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to create channel: %w", err)
	}

	return channel, nil
}

// DeclareQueue declares a queue (idempotent operation)
func (c *RabbitMQClient) DeclareQueue(channel *amqp.Channel, queueName string) error {
	_, err := channel.QueueDeclare(
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

// Close closes the connection
func (c *RabbitMQClient) Close() error {
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
