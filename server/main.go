package main

import (
	"bighousevn/be/api"
	"bighousevn/be/db"
	"bighousevn/be/middleware"
	"bighousevn/be/repository"
	"bighousevn/be/services"
	"bighousevn/be/utils"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file, proceeding with environment variables")
	}

	// Initialize Database
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is not set in .env file")
	}
	log.Println("Initializing database connection...")
	if err := db.InitDB(dbURL); err != nil {
		log.Fatalf("could not connect to the database: %v", err)
	}
	log.Println("Database connection established successfully.")

	if err := utils.RegisterValidators(); err != nil {
		panic(err)
	}
	r := gin.Default()
	r.SetTrustedProxies(nil)
	// CORS configuration
	feURL := os.Getenv("FE_URL")
	if feURL == "" {
		feURL = "http://localhost:3000" // Default for local development
	}
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{feURL},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Initialize Repository, Service, and Controller
	authRepo := repository.NewAuthRepository(db.DB)
	authService := services.NewAuthService(authRepo)
	authController := api.NewAuthController(authService)

	v1 := r.Group("/api/v1")
	{
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authController.Register)
			auth.POST("/login", authController.Login)
			auth.POST("/refresh-token", authController.RefreshToken)
			auth.POST("/forgot-password", authController.ForgotPassword)
			auth.POST("/reset-password", authController.ResetPassword)
			auth.GET("/verify-email", authController.VerifyEmail)
		}

		authenticated := v1.Group("/")
		authenticated.Use(middleware.AuthMiddleware(authRepo))
		{
			authenticated.GET("/users/me", authController.GetProfile)
			authenticated.POST("/users/me/password", authController.ChangePassword)
		}
		v1.GET("/ping", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "pong",
			})
		})
	}

	// --- Graceful Shutdown Logic ---

	// Create a server object to have more control
	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	// Goroutine to start the server
	go func() {
		log.Println("Server starting on port 8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Goroutine to wait for a shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Close the database connection FIRST
	db.CloseDB()
	log.Println("Database connection closed.")

	// Wait a bit to ensure DB cleanup is complete
	time.Sleep(2 * time.Second)

	// Create a context with a timeout for the server to shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown: ", err)
	}

	log.Println("Server exiting.")
}
