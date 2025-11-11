package http

import (
	"ai-clipper/server2/internal/auth/application"
	"ai-clipper/server2/internal/middleware"

	"github.com/gin-gonic/gin"
)

// NewFileRouter creates and registers all routes for file operations
func NewFileRouter(
	router *gin.Engine,
	fileController *FileController,
	tokenGenerator application.TokenGenerator,
) {
	v1 := router.Group("/api/v1")
	{
		authenticated := v1.Group("/")
		authenticated.Use(middleware.AuthMiddleware(tokenGenerator))
		{
			// File upload
			authenticated.POST("/upload", fileController.UploadFile)

			// Get user's files
			authenticated.GET("/files/me", fileController.GetMyFiles)

			// Process video
			authenticated.POST("/files/:file_id/process", fileController.ProcessVideo)

			// Get clip by ID
			authenticated.GET("/clips/:clip_id", fileController.GetClip)

			// Get user's clips
			authenticated.GET("/clips/me", fileController.GetUserClips)

			// Download clip by ID
			authenticated.GET("/clips/:clip_id/download", fileController.DownloadClip)

			// Test endpoint to list clips from storage (for debugging)
			authenticated.GET("/test/list-clips", fileController.TestListClips)
		}
	}
}
