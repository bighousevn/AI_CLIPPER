package main

import (
	"ai-clipper/server2/database"
	authApp "ai-clipper/server2/internal/auth/application"
	authInfra "ai-clipper/server2/internal/auth/infrastructure"
	authHttp "ai-clipper/server2/internal/auth/interfaces/http"
	fileApp "ai-clipper/server2/internal/file/application"
	fileInfra "ai-clipper/server2/internal/file/infrastructure"
	fileHttp "ai-clipper/server2/internal/file/interfaces/http"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	storage "github.com/supabase-community/storage-go"
)

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
	db, err := database.InitDatabase(dsn) // Use the new database package
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// 3. Dependency Injection (Wiring the components)
	log.Println("Initializing dependencies...")

	// Auth Module - Infrastructure
	userRepo := authInfra.NewGormUserRepository(db)
	passwordHasher := authInfra.NewBcryptPasswordHasher()

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET environment variable not set")
	}
	jwtConfig := authInfra.JWTConfig{
		AccessSecret:  jwtSecret,
		RefreshSecret: os.Getenv("JWT_REFRESH_SECRET"),
		AccessExpiry:  time.Hour * 1,       // 1 hour
		RefreshExpiry: time.Hour * 24 * 30, // 30 days
	}
	tokenGenerator := authInfra.NewJWTTokenGenerator(jwtConfig)

	// Auth Module - Application
	authUseCase := authApp.NewAuthUseCase(userRepo, passwordHasher, tokenGenerator)

	// Auth Module - Interfaces
	authPresenter := authHttp.NewAuthPresenter()
	authController := authHttp.NewAuthController(authUseCase, authPresenter)

	// File Module - Infrastructure
	fileRepo := fileInfra.NewGormFileRepository(db)

	// Initialize Supabase Storage Client
	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_SERVICE_ROLE_KEY")
	if supabaseURL == "" || supabaseKey == "" {
		log.Fatal("SUPABASE_URL or SUPABASE_SERVICE_ROLE_KEY not set")
	}
	storageClient := storage.NewClient(supabaseURL+"/storage/v1", supabaseKey, nil)
	storageBucket := os.Getenv("SUPABASE_STORAGE_BUCKET")
	if storageBucket == "" {
		storageBucket = "uploaded_files" // default bucket name
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
	fileUseCase := fileApp.NewFileUseCase(fileRepo, clipRepo, storageService, modalService)

	// File Module - Interfaces
	filePresenter := fileHttp.NewFilePresenter()
	fileController := fileHttp.NewFileController(fileUseCase, filePresenter)

	// 5. Initialize Gin Engine
	log.Println("Initializing web server...")
	router := gin.Default()

	// 6. Register Routes
	authHttp.NewAuthRouter(router, authController, tokenGenerator)
	fileHttp.NewFileRouter(router, fileController, tokenGenerator)

	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

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
