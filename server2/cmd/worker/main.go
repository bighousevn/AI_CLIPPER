package main

import (
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"

	"ai-clipper/server2/database"
	fileApp "ai-clipper/server2/internal/file/application"
	fileInfra "ai-clipper/server2/internal/file/infrastructure"
	"ai-clipper/server2/internal/messaging/domain"
	"ai-clipper/server2/internal/messaging/infrastructure"

	"github.com/joho/godotenv"
	storage "github.com/supabase-community/storage-go"
)

func main() {
	log.Println("Starting RabbitMQ worker...")

	// 1. Load Environment Variables
	if err := godotenv.Load("../../.env"); err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}

	// 2. Database Connection
	dsn := os.Getenv("SUPABASE_URLL")
	if dsn == "" {
		log.Fatal("SUPABASE_URLL environment variable not set")
	}
	db, err := database.InitDatabase(dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// 3. Initialize dependencies
	fileRepo := fileInfra.NewGormFileRepository(db)
	clipRepo := fileInfra.NewGormClipRepository(db)

	// Supabase Storage
	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_SERVICE_ROLE_KEY")
	if supabaseURL == "" || supabaseKey == "" {
		log.Fatal("SUPABASE_URL or SUPABASE_SERVICE_ROLE_KEY not set")
	}
	storageClient := storage.NewClient(supabaseURL+"/storage/v1", supabaseKey, nil)
	storageBucket := os.Getenv("SUPABASE_STORAGE_BUCKET")
	if storageBucket == "" {
		storageBucket = "uploaded_files"
	}
	storageService := fileInfra.NewSupabaseStorageService(storageClient, storageBucket)

	// Modal Service
	modalURL := os.Getenv("MODAL_URL")
	modalToken := os.Getenv("MODAL_TOKEN")
	if modalURL == "" || modalToken == "" {
		log.Fatal("MODAL_URL or MODAL_TOKEN not set")
	}
	modalService := fileInfra.NewHTTPModalService(modalURL, modalToken)

	// 4. Initialize RabbitMQ
	rabbitmqURL := os.Getenv("RABBITMQ_URL")
	if rabbitmqURL == "" {
		rabbitmqURL = "amqp://admin:admin123@localhost:5672/" // default matches docker-compose
	}

	rabbitmqClient, err := infrastructure.NewRabbitMQClient(rabbitmqURL)
	if err != nil {
		log.Fatalf("Failed to create RabbitMQ client: %v", err)
	}
	defer rabbitmqClient.Close()

	// Initialize Publisher for worker (to send status updates later)
	rabbitmqPublisher, err := infrastructure.NewRabbitMQPublisher(rabbitmqClient)
	if err != nil {
		log.Fatalf("Failed to create RabbitMQ publisher: %v", err)
	}
	defer rabbitmqPublisher.Close()

	// File UseCase (used by message handlers)
	fileUseCase := fileApp.NewFileUseCase(fileRepo, clipRepo, storageService, modalService, rabbitmqPublisher)

	consumer, err := infrastructure.NewRabbitMQConsumer(rabbitmqClient)
	if err != nil {
		log.Fatalf("Failed to create RabbitMQ consumer: %v", err)
	}

	// 5. Start consuming messages
	// Video Processing Consumer
	err = consumer.ConsumeVideoProcessing(func(body []byte) error {
		var msg domain.VideoProcessingMessage
		if err := json.Unmarshal(body, &msg); err != nil {
			log.Printf("Failed to unmarshal video processing message: %v", err)
			return err
		}

		log.Printf("Processing video for file: %s, user: %s", msg.FileID, msg.UserID)

		// Call the actual processing logic
		if err := fileUseCase.ProcessVideo(msg.FileID, msg.UserID, msg.Config); err != nil {
			log.Printf("Failed to process video: %v", err)
			return err
		}

		log.Printf("Successfully processed video for file: %s", msg.FileID)
		return nil
	})
	if err != nil {
		log.Fatalf("Failed to start video processing consumer: %v", err)
	}

	// Email Notification Consumer (future use)
	err = consumer.ConsumeEmailNotification(func(body []byte) error {
		var msg domain.EmailNotificationMessage
		if err := json.Unmarshal(body, &msg); err != nil {
			log.Printf("Failed to unmarshal email notification: %v", err)
			return err
		}

		log.Printf("Sending email to: %s, subject: %s", msg.To, msg.Subject)
		// TODO: Implement email sending logic
		return nil
	})
	if err != nil {
		log.Fatalf("Failed to start email notification consumer: %v", err)
	}

	log.Println("Worker is running. Press Ctrl+C to exit.")

	// 6. Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down worker...")
}
