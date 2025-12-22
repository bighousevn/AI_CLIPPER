package main

import (
	"ai-clipper/server2/database"
	authApp "ai-clipper/server2/internal/auth/application"
	authInfra "ai-clipper/server2/internal/auth/infrastructure"
	authHttp "ai-clipper/server2/internal/auth/interfaces/http"
	fileApp "ai-clipper/server2/internal/file/application"
	fileInfra "ai-clipper/server2/internal/file/infrastructure"
	fileHttp "ai-clipper/server2/internal/file/interfaces/http"
	messagedomain "ai-clipper/server2/internal/messaging/domain"
	msgInfra "ai-clipper/server2/internal/messaging/infrastructure"
	"ai-clipper/server2/internal/middleware"
	paymentApp "ai-clipper/server2/internal/payment/application"
	paymentInfra "ai-clipper/server2/internal/payment/infrastructure"
	paymentHttp "ai-clipper/server2/internal/payment/interfaces/http"
	"ai-clipper/server2/internal/sse"
	sseHttp "ai-clipper/server2/internal/sse/interfaces/http"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	storage "github.com/supabase-community/storage-go"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "ai-clipper/server2/docs"
)

// @title AI Clipper API
// @version 1.0
// @description This is the API for the AI Clipper application.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /
func main() {
	// 1. Load Environment Variables
	if err := godotenv.Load("../../.env"); err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}

	// 2. Database Connection
	dsn := os.Getenv("SUPABASE_URLL")
	if dsn == "" {
		log.Fatal("SUPABASE_URLL environment variable not set")
	}

	// Run Database Migrations
	log.Println("Running database migrations...")
	if err := database.RunMigrations(dsn); err != nil {
		log.Fatalf("Failed to run database migrations: %v", err)
	}

	db, err := database.InitDatabase(dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// 3. Dependency Injection (Wiring the components)
	log.Println("Initializing dependencies...")

	// Initialize Stripe Service
	stripeService := paymentInfra.NewStripeService()

	// Auth Module - Infrastructure
	userRepo := authInfra.NewGormUserRepository(db)
	passwordHasher := authInfra.NewBcryptPasswordHasher()

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET environment variable not set")
	}
	jwtConfig := authInfra.JWTConfig{
		AccessSecret:  jwtSecret,
		RefreshSecret: os.Getenv("REFRESH_JWT_SECRET"),
		AccessExpiry:  time.Hour * 1,       // 1 hour
		RefreshExpiry: time.Hour * 24 * 30, // 30 days
	}
	tokenGenerator := authInfra.NewJWTTokenGenerator(jwtConfig)

	// Initialize RabbitMQ
	rabbitmqURL := os.Getenv("RABBITMQ_URL")
	if rabbitmqURL == "" {
		rabbitmqURL = "amqp://admin:admin123@localhost:5672/" // Default local
	}
	rabbitmqClient, err := msgInfra.NewRabbitMQClient(rabbitmqURL)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer rabbitmqClient.Close()

	rabbitmqPublisher, err := msgInfra.NewRabbitMQPublisher(rabbitmqClient)
	if err != nil {
		log.Fatalf("Failed to create RabbitMQ publisher: %v", err)
	}
	defer rabbitmqPublisher.Close()

	// Auth Module - Application
	authUseCase := authApp.NewAuthUseCase(userRepo, passwordHasher, tokenGenerator, rabbitmqPublisher, stripeService)

	// Auth Module - Interfaces
	authPresenter := authHttp.NewAuthPresenter()
	authController := authHttp.NewAuthController(authUseCase, authPresenter)

	// Payment Module - Application & Interfaces
	paymentUseCase := paymentApp.NewPaymentUseCase(stripeService, userRepo)
	paymentController := paymentHttp.NewPaymentController(paymentUseCase)
	webhookController := paymentHttp.NewWebhookController(stripeService, userRepo)

	// File Module - Infrastructure
	fileRepo := fileInfra.NewGormFileRepository(db)

	// Initialize Supabase Storage Client
	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_SERVICE_ROLE_KEY")
	if supabaseURL == "" || supabaseKey == "" {
		log.Fatal("SUPABASE_URL or SUPABASE_SERVICE_ROLE_KEY not set")
	}
	storageClient := storage.NewClient(supabaseURL+"/storage/v1", supabaseKey, nil)
	storageBucket := os.Getenv("SUPABASE_BUCKET_NAME")
	if storageBucket == "" {
		storageBucket = "uploaded_files"
	}
	storageService := fileInfra.NewSupabaseStorageService(storageClient, storageBucket)

	// Modal Service for video processing
	modalURL := os.Getenv("MODAL_URL")
	modalToken := os.Getenv("MODAL_TOKEN")
	if modalURL == "" || modalToken == "" {
		log.Fatal("MODAL_URL or MODAL_TOKEN not set")
	}
	modalService := fileInfra.NewHTTPModalService(modalURL, modalToken)

	// Clip Repository
	clipRepo := fileInfra.NewGormClipRepository(db)

	// File Module - Application
	fileUseCase := fileApp.NewFileUseCase(fileRepo, clipRepo, userRepo, storageService, modalService, rabbitmqPublisher)

	// File Module - Interfaces
	filePresenter := fileHttp.NewFilePresenter()
	fileController := fileHttp.NewFileController(fileUseCase, filePresenter)

	// --- SSE Setup ---
	sseManager := sse.NewManager()

	// Setup RabbitMQ Consumer for Status Updates
	rabbitmqConsumer, err := msgInfra.NewRabbitMQConsumer(rabbitmqClient)
	if err != nil {
		log.Fatalf("Failed to create RabbitMQ consumer: %v", err)
	}

	// Start listening for status updates
	err = rabbitmqConsumer.ConsumeStatusUpdate(func(body []byte) error {
		var msg messagedomain.StatusUpdateMessage
		if err := json.Unmarshal(body, &msg); err != nil {
			log.Printf("Failed to unmarshal status update: %v", err)
			return err
		}

		log.Printf("Received status update for user %s: %s", msg.UserID, msg.Status)

		// Push to SSE Manager
		sseManager.SendToUser(msg.UserID, "video_status", msg)
		return nil
	})
	if err != nil {
		log.Fatalf("Failed to start status update consumer: %v", err)
	}
	defer rabbitmqConsumer.Close()

	// 5. Initialize Gin Engine and Cors configuration
	log.Println("Initializing web server...")
	router := gin.Default()

	// Attach HTTP logger middleware
	router.Use(middleware.LoggerMiddleware())

	if err := router.SetTrustedProxies(nil); err != nil {
		panic(err)
	}
	// CORS configuration

	feURL := os.Getenv("FE_URL")
	if feURL == "" {
		feURL = "http://localhost:3000"
	}
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{feURL},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	router.Use(middleware.RateLimitingMiddleware())

	authHttp.NewAuthRouter(router, authController, tokenGenerator)
	fileHttp.NewFileRouter(router, fileController, tokenGenerator)
	paymentHttp.RegisterRoutes(router, paymentController, tokenGenerator)
	paymentHttp.RegisterWebhookRoutes(router, webhookController)

	// SSE Endpoint
	sseController := sseHttp.NewSSEController(sseManager)
	router.GET("/api/v1/events", middleware.SSEAuthMiddleware(tokenGenerator), sseController.StreamEvents)

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	go middleware.CleanupClients()
	// 7. Start Server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	serverAddr := fmt.Sprintf(":%s", port)
	log.Printf("Server starting on %s", serverAddr)
	if err := router.Run(serverAddr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
