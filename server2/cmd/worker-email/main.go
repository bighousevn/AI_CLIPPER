package main

import (
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"

	authInfra "ai-clipper/server2/internal/auth/infrastructure"
	"ai-clipper/server2/internal/messaging/domain"
	"ai-clipper/server2/internal/messaging/infrastructure"

	"github.com/joho/godotenv"
)

func main() {
	log.Println("Starting RabbitMQ Email Worker...")

	// 1. Load Environment Variables
	if err := godotenv.Load("../../.env"); err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}

	// 2. Initialize Email Service
	emailService := authInfra.NewSMTPEmailService()

	// 3. Initialize RabbitMQ
	rabbitmqURL := os.Getenv("RABBITMQ_URL")
	if rabbitmqURL == "" {
		rabbitmqURL = "amqp://admin:admin123@localhost:5672/"
	}

	rabbitmqClient, err := infrastructure.NewRabbitMQClient(rabbitmqURL)
	if err != nil {
		log.Fatalf("Failed to create RabbitMQ client: %v", err)
	}
	defer rabbitmqClient.Close()

	consumer, err := infrastructure.NewRabbitMQConsumer(rabbitmqClient)
	if err != nil {
		log.Fatalf("Failed to create RabbitMQ consumer: %v", err)
	}
	defer consumer.Close()

	// 4. Start consuming email messages
	err = consumer.ConsumeEmailNotification(func(body []byte) error {
		var msg domain.EmailNotificationMessage
		if err := json.Unmarshal(body, &msg); err != nil {
			log.Printf("Failed to unmarshal email message: %v", err)
			return err
		}

		log.Printf("Received email task: Type=%s, To=%s", msg.Type, msg.To)

		var sendErr error
		switch msg.Type {
		case "VERIFY":
			// Content is the token
			sendErr = emailService.SendVerificationEmail(msg.To, msg.Username, msg.Content)
		case "RESET":
			// Content is the token
			sendErr = emailService.SendPasswordResetEmail(msg.To, msg.Username, msg.Content)
		default:
			log.Printf("Unknown email type: %s", msg.Type)
			return nil // Ack message even if type unknown to remove from queue
		}

		if sendErr != nil {
			log.Printf("Failed to send email: %v", sendErr)
			// Returning error here will cause the message to be Nack-ed (and potentially requeued or DLQ)
			// depending on consumer implementation. Currently RabbitMQConsumer drops msg on error.
			return sendErr
		}

		log.Printf("Email sent successfully to %s", msg.To)
		return nil
	})

	if err != nil {
		log.Fatalf("Failed to start email notification consumer: %v", err)
	}

	log.Println("Email Worker is running. Press Ctrl+C to exit.")

	// 5. Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down Email Worker...")
}
