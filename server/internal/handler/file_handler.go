package handler

import (
	"bighousevn/be/internal/database"
	utils "bighousevn/be/internal/file"
	"bighousevn/be/internal/models"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// FileHandler handles file-related requests
type FileHandler struct{}

// NewFileHandler creates a new FileHandler
func NewFileHandler() *FileHandler {
	return &FileHandler{}
}

// UploadFile handles the file upload process
func (h *FileHandler) UploadFile(c *gin.Context) {
	log.Printf("DEBUG UPLOAD: Starting file upload")

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		log.Printf("ERROR UPLOAD: Failed to get file from form: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "File is required"})
		return
	}
	defer file.Close()

	log.Printf("DEBUG UPLOAD: File received: %s, size: %d", header.Filename, header.Size)

	// Lấy user từ context, được set bởi AuthMiddleware
	userInterface, exists := c.Get("user")
	if !exists {
		log.Printf("ERROR UPLOAD: User not found in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	user, ok := userInterface.(*models.User)
	if !ok {
		log.Printf("ERROR UPLOAD: Failed to cast user from context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user data"})
		return
	}

	log.Printf("DEBUG UPLOAD: User authenticated: %s (ID: %s)", user.Email, user.ID)

	// Upload file to Supabase
	filePath, err := utils.UploadFileToSupabase(database.StorageClient, file, header, *user.ID)
	if err != nil {
		log.Printf("ERROR UPLOAD: Failed to upload to Supabase: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Printf("DEBUG UPLOAD: File uploaded successfully to: %s", filePath)
	c.JSON(http.StatusOK, gin.H{"message": "File uploaded successfully", "path": filePath})
}
