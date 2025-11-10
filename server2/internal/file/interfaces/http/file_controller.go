package http

import (
	"ai-clipper/server2/internal/file/application"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// FileController handles HTTP requests for file operations
type FileController struct {
	fileUseCase *application.FileUseCase
	presenter   *FilePresenter
}

// NewFileController creates a new FileController
func NewFileController(fileUseCase *application.FileUseCase, presenter *FilePresenter) *FileController {
	return &FileController{
		fileUseCase: fileUseCase,
		presenter:   presenter,
	}
}

// UploadFile handles file upload requests
func (ctrl *FileController) UploadFile(c *gin.Context) {

	if err := c.Request.ParseMultipartForm(100 << 20); err != nil {
		log.Printf("Failed to parse multipart form: %v", err)
	} else {
		if c.Request.MultipartForm != nil {
			log.Printf("Available file fields: %v", c.Request.MultipartForm.File)
			log.Printf("Available value fields: %v", c.Request.MultipartForm.Value)
		}
	}

	// Get file from form
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		log.Printf("Failed to get file from form with key 'file': %v", err)
		ctrl.presenter.RespondError(c, http.StatusBadRequest, "File is required")
		return
	}
	defer file.Close()

	log.Printf("File received: %s, size: %d", header.Filename, header.Size)

	// Get authenticated user from context
	userIDInterface, exists := c.Get("userID")
	if !exists {
		log.Printf("User ID not found in context")
		ctrl.presenter.RespondError(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	userID, ok := userIDInterface.(string)
	if !ok {
		// Try UUID type
		if uuidVal, ok := userIDInterface.(interface{ String() string }); ok {
			userID = uuidVal.String()
		} else {
			log.Printf("Failed to cast user ID from context")
			ctrl.presenter.RespondError(c, http.StatusUnauthorized, "Invalid user data")
			return
		}
	}

	log.Printf("User authenticated, ID: %s", userID)

	// Create DTO
	dto := application.FileUploadDTO{
		File:   file,
		Header: header,
		UserID: userID,
	}

	// Execute use case
	response, err := ctrl.fileUseCase.UploadFile(dto)
	if err != nil {
		log.Printf("Failed to upload file: %v", err)
		ctrl.presenter.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}

	log.Printf("File uploaded successfully: %s", response.ID)
	ctrl.presenter.RespondSuccess(c, http.StatusOK, response)
}

// GetMyFiles handles requests to get all files uploaded by the authenticated user
func (ctrl *FileController) GetMyFiles(c *gin.Context) {
	// Get authenticated user from context
	userIDInterface, exists := c.Get("userID")
	if !exists {
		ctrl.presenter.RespondError(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	userID, ok := userIDInterface.(string)
	if !ok {
		// Try UUID type
		if uuidVal, ok := userIDInterface.(interface{ String() string }); ok {
			userID = uuidVal.String()
		} else {
			ctrl.presenter.RespondError(c, http.StatusUnauthorized, "Invalid user data")
			return
		}
	}

	// Execute use case
	files, err := ctrl.fileUseCase.GetUserFiles(userID)
	if err != nil {
		ctrl.presenter.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}

	ctrl.presenter.RespondSuccess(c, http.StatusOK, files)
}

// ProcessVideo handles video processing requests
func (ctrl *FileController) ProcessVideo(c *gin.Context) {
	fileID := c.Param("file_id")
	if fileID == "" {
		ctrl.presenter.RespondError(c, http.StatusBadRequest, "file_id is required")
		return
	}

	// Get authenticated user from context
	userIDInterface, exists := c.Get("userID")
	if !exists {
		ctrl.presenter.RespondError(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	userID, ok := userIDInterface.(string)
	if !ok {
		if uuidVal, ok := userIDInterface.(interface{ String() string }); ok {
			userID = uuidVal.String()
		} else {
			ctrl.presenter.RespondError(c, http.StatusUnauthorized, "Invalid user data")
			return
		}
	}

	log.Printf("Processing video request: user=%s, fileID=%s", userID, fileID)

	// Execute use case
	if err := ctrl.fileUseCase.ProcessVideo(fileID, userID); err != nil {
		log.Printf("Failed to process video: %v", err)
		ctrl.presenter.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}

	ctrl.presenter.RespondSuccess(c, http.StatusAccepted, gin.H{
		"message": "Video processing started successfully",
		"file_id": fileID,
	})
}

// GetClip handles get clip by ID requests
func (ctrl *FileController) GetClip(c *gin.Context) {
	clipID := c.Param("clip_id")
	if clipID == "" {
		ctrl.presenter.RespondError(c, http.StatusBadRequest, "clip_id is required")
		return
	}

	// Get authenticated user from context
	userIDInterface, exists := c.Get("userID")
	if !exists {
		ctrl.presenter.RespondError(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	userID, ok := userIDInterface.(string)
	if !ok {
		if uuidVal, ok := userIDInterface.(interface{ String() string }); ok {
			userID = uuidVal.String()
		} else {
			ctrl.presenter.RespondError(c, http.StatusUnauthorized, "Invalid user data")
			return
		}
	}

	// Execute use case
	clip, err := ctrl.fileUseCase.GetClip(clipID, userID)
	if err != nil {
		if err.Error() == "unauthorized: clip does not belong to user" {
			ctrl.presenter.RespondError(c, http.StatusForbidden, err.Error())
			return
		}
		ctrl.presenter.RespondError(c, http.StatusNotFound, err.Error())
		return
	}

	ctrl.presenter.RespondSuccess(c, http.StatusOK, clip)
}

// GetUserClips handles get all user clips requests
func (ctrl *FileController) GetUserClips(c *gin.Context) {
	// Get authenticated user from context
	userIDInterface, exists := c.Get("userID")
	if !exists {
		ctrl.presenter.RespondError(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	userID, ok := userIDInterface.(string)
	if !ok {
		if uuidVal, ok := userIDInterface.(interface{ String() string }); ok {
			userID = uuidVal.String()
		} else {
			ctrl.presenter.RespondError(c, http.StatusUnauthorized, "Invalid user data")
			return
		}
	}

	// Execute use case
	clips, err := ctrl.fileUseCase.GetUserClips(userID)
	if err != nil {
		ctrl.presenter.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}

	ctrl.presenter.RespondSuccess(c, http.StatusOK, clips)
}
