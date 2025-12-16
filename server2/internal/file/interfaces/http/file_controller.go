package http

import (
	"ai-clipper/server2/internal/file/application"
	domainFile "ai-clipper/server2/internal/file/domain/file"
	"encoding/json"
	"fmt"
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

// UploadFile godoc
// @Summary Upload a file
// @Description Upload a video file for processing. The file should be sent as multipart/form-data.
// @Tags files
// @Accept  multipart/form-data
// @Produce  json
// @Param   file formData file true "The file to upload"
// @Success 200 {object} application.FileResponseDTO "File uploaded successfully"
// @Failure 400 {object} httputil.HTTPError "File is required"
// @Failure 401 {object} httputil.HTTPError "User not authenticated"
// @Failure 500 {object} httputil.HTTPError "Internal server error"
// @Security ApiKeyAuth
// @Router /upload [post]
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

	configStrs := c.Request.MultipartForm.Value["config"]
	if len(configStrs) == 0 {
		log.Printf("Fail to receive config")
		ctrl.presenter.RespondError(c, http.StatusBadRequest, "Config is required")
		return
	}
	var config domainFile.VideoConfig
	if err := json.Unmarshal([]byte(configStrs[0]), &config); err != nil {
		log.Printf("Error parsing config JSON: %v", err)
		ctrl.presenter.RespondError(c, http.StatusBadRequest, "Config is invalid")
		return
	}

	log.Printf("Received config: %+v", config)

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
		Config: config,
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

// GetMyFiles godoc
// @Summary Get user's uploaded files
// @Description Retrieves a list of all files uploaded by the currently authenticated user.
// @Tags files
// @Produce  json
// @Success 200 {array} application.FileResponseDTO "Successfully retrieved files"
// @Failure 401 {object} httputil.HTTPError "User not authenticated"
// @Failure 500 {object} httputil.HTTPError "Internal server error"
// @Security ApiKeyAuth
// @Router /files/me [get]
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

// ProcessVideo godoc
// @Summary Process a video file
// @Description Starts a video processing job for a previously uploaded file.
// @Tags files
// @Produce  json
// @Param   file_id path string true "The ID of the file to process"
// @Success 202 {object} object "message: Video processing started successfully, file_id: {file_id}"
// @Failure 400 {object} httputil.HTTPError "file_id is required"
// @Failure 401 {object} httputil.HTTPError "User not authenticated"
// @Failure 500 {object} httputil.HTTPError "Internal server error"
// @Security ApiKeyAuth
// @Router /files/{file_id}/process [post]
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
	if err := ctrl.fileUseCase.ProcessVideo(fileID, userID, domainFile.VideoConfig{}); err != nil {
		log.Printf("Failed to process video: %v", err)
		ctrl.presenter.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}

	ctrl.presenter.RespondSuccess(c, http.StatusAccepted, gin.H{
		"message": "Video processing started successfully",
		"file_id": fileID,
	})
}

// GetClip godoc
// @Summary Get a clip by ID
// @Description Retrieves the details of a specific clip by its ID.
// @Tags clips
// @Produce  json
// @Param   clip_id path string true "The ID of the clip"
// @Success 200 {object} application.ClipResponseDTO "Successfully retrieved clip"
// @Failure 400 {object} httputil.HTTPError "clip_id is required"
// @Failure 401 {object} httputil.HTTPError "User not authenticated"
// @Failure 403 {object} httputil.HTTPError "User is not authorized to view this clip"
// @Failure 404 {object} httputil.HTTPError "Clip not found"
// @Security ApiKeyAuth
// @Router /clips/{clip_id} [get]
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

// GetUserClips godoc
// @Summary Get user's clips
// @Description Retrieves a list of all clips generated by the currently authenticated user.
// @Tags clips
// @Produce  json
// @Success 200 {array} application.ClipResponseDTO "Successfully retrieved clips"
// @Failure 401 {object} httputil.HTTPError "User not authenticated"
// @Failure 500 {object} httputil.HTTPError "Internal server error"
// @Security ApiKeyAuth
// @Router /clips/me [get]
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

// DownloadClip godoc
// @Summary Download a clip
// @Description Downloads the video file for a specific clip.
// @Tags clips
// @Produce  video/mp4
// @Param   clip_id path string true "The ID of the clip to download"
// @Success 200 {file} file "The clip video file"
// @Failure 400 {object} httputil.HTTPError "clip_id is required"
// @Failure 401 {object} httputil.HTTPError "User not authenticated"
// @Failure 403 {object} httputil.HTTPError "User is not authorized to download this clip"
// @Failure 404 {object} httputil.HTTPError "Clip not found"
// @Security ApiKeyAuth
// @Router /clips/{clip_id}/download [get]
func (ctrl *FileController) DownloadClip(c *gin.Context) {
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

	// Execute use case to download clip
	fileBytes, filePath, err := ctrl.fileUseCase.DownloadClip(clipID, userID)
	if err != nil {
		if err.Error() == "unauthorized: clip does not belong to user" {
			ctrl.presenter.RespondError(c, http.StatusForbidden, err.Error())
			return
		}
		ctrl.presenter.RespondError(c, http.StatusNotFound, err.Error())
		return
	}

	// Extract filename from path (e.g., "user-xxx/clips/clip_0.mp4" -> "clip_0.mp4")
	filename := "clip.mp4"
	if len(filePath) > 0 {
		parts := []rune(filePath)
		lastSlash := -1
		for i := len(parts) - 1; i >= 0; i-- {
			if parts[i] == '/' {
				lastSlash = i
				break
			}
		}
		if lastSlash >= 0 && lastSlash < len(parts)-1 {
			filename = string(parts[lastSlash+1:])
		}
	}

	// Set headers for file download
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Type", "video/mp4")
	c.Header("Content-Length", fmt.Sprintf("%d", len(fileBytes)))

	// Write file bytes to response
	c.Data(http.StatusOK, "video/mp4", fileBytes)
}

// TestListClips godoc
// @Summary [Test] List clips from storage
// @Description A test endpoint to list clips directly from a storage folder. Requires authentication.
// @Tags test
// @Produce  json
// @Param   folder query string true "The folder path in storage to list clips from"
// @Success 200 {object} object "A list of clips in the folder"
// @Failure 400 {object} httputil.HTTPError "folder query parameter is required"
// @Failure 500 {object} httputil.HTTPError "Internal server error"
// @Security ApiKeyAuth
// @Router /test/list-clips [get]
func (ctrl *FileController) TestListClips(c *gin.Context) {
	// Get folder path from query param
	folderPath := c.Query("folder")
	if folderPath == "" {
		ctrl.presenter.RespondError(c, http.StatusBadRequest, "folder query parameter is required")
		return
	}

	log.Printf("Testing list clips from folder: %s", folderPath)

	// Call use case to test listing clips
	clips, err := ctrl.fileUseCase.TestListClipsFromStorage(folderPath)
	if err != nil {
		ctrl.presenter.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}

	ctrl.presenter.RespondSuccess(c, http.StatusOK, gin.H{
		"folder":      folderPath,
		"clips_count": len(clips),
		"clips":       clips,
	})
}
