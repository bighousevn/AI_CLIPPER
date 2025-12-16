package application

import (
	"ai-clipper/server2/internal/file/domain/clip"
	"ai-clipper/server2/internal/file/domain/file"
	messagedomain "ai-clipper/server2/internal/messaging/domain"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
)

// FileUseCase handles file upload business logic
type FileUseCase struct {
	fileRepo         file.FileRepository
	clipRepo         clip.ClipRepository
	storageService   StorageService
	modalService     ModalService
	messagePublisher messagedomain.MessagePublisher
}

// NewFileUseCase creates a new FileUseCase
func NewFileUseCase(
	fileRepo file.FileRepository,
	clipRepo clip.ClipRepository,
	storageService StorageService,
	modalService ModalService,
	messagePublisher messagedomain.MessagePublisher,
) *FileUseCase {
	return &FileUseCase{
		fileRepo:         fileRepo,
		clipRepo:         clipRepo,
		storageService:   storageService,
		modalService:     modalService,
		messagePublisher: messagePublisher,
	}
}

// UploadFile handles the file upload use case
func (uc *FileUseCase) UploadFile(dto FileUploadDTO) (*FileResponseDTO, error) {
	log.Printf("Starting file upload for user: %s", dto.UserID)

	// Parse user ID
	userID, err := uuid.Parse(dto.UserID)
	if err != nil {
		log.Printf("Invalid user ID: %v", err)
		return nil, errors.New("invalid user ID")
	}

	// Upload file to storage
	filePath, err := uc.storageService.Upload(dto.File, dto.Header, dto.UserID)
	if err != nil {
		log.Printf("Failed to upload file to storage: %v", err)
		return nil, err
	}

	// Create file entity
	fileEntity := file.NewFile(
		userID,
		dto.Header.Filename,
		filePath,
		dto.Header.Size,
		dto.Header.Header.Get("Content-Type"),
	)

	// Validate file
	if !fileEntity.IsValidSize() {
		log.Printf("File size exceeds limit: %d bytes", fileEntity.FileSize)
		// Clean up uploaded file
		uc.storageService.Delete(filePath)
		return nil, errors.New("file size exceeds maximum allowed limit")
	}

	// Save file metadata to database
	if err := uc.fileRepo.Save(fileEntity); err != nil {
		log.Printf("Failed to save file metadata: %v", err)
		// Clean up uploaded file
		uc.storageService.Delete(filePath)
		return nil, err
	}

	log.Printf("File uploaded successfully: %s", fileEntity.ID)

	// Publish message to RabbitMQ for async processing with retry logic
	var publishErr error
	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		publishErr = uc.messagePublisher.PublishVideoProcessing(
			fileEntity.ID.String(),
			fileEntity.UserID.String(),
			fileEntity.FilePath,
			dto.Config,
		)
		if publishErr == nil {
			break
		}
		log.Printf("Failed to publish message (attempt %d/%d): %v", i+1, maxRetries, publishErr)
		time.Sleep(500 * time.Millisecond)
	}

	finalStatus := fileEntity.Status
	finalUpdatedAt := fileEntity.UpdatedAt

	if publishErr != nil {
		log.Printf("CRITICAL: Failed to publish video processing message after retries: %v", publishErr)
		
		// Cleanup: Delete file metadata from DB
		if err := uc.fileRepo.Delete(fileEntity.ID); err != nil {
			log.Printf("Failed to delete file metadata during cleanup: %v", err)
		}
		
		// Cleanup: Delete file from storage
		if err := uc.storageService.Delete(fileEntity.FilePath); err != nil {
			log.Printf("Failed to delete file from storage during cleanup: %v", err)
		}

		return nil, fmt.Errorf("failed to queue video for processing: %w", publishErr)
	}
	
	// Success path
	// Update status to "queued"
	if updateErr := uc.fileRepo.UpdateStatus(fileEntity.ID, "queued", 0); updateErr != nil {
		log.Printf("Failed to update status to queued: %v", updateErr)
	} else {
		finalStatus = "queued"
		finalUpdatedAt = time.Now()
		log.Printf("Video processing message published and status updated to queued for file: %s", fileEntity.ID)
	}

	return &FileResponseDTO{
		ID:        fileEntity.ID.String(),
		FileName:  fileEntity.FileName,
		FilePath:  fileEntity.FilePath,
		FileSize:  fileEntity.FileSize,
		Message:   "File uploaded and processing started",
		Status:    finalStatus,
		ClipCount: fileEntity.ClipCount,
		CreatedAt: fileEntity.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: finalUpdatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

// GetUserFiles retrieves all files uploaded by a user
func (uc *FileUseCase) GetUserFiles(userID string) ([]*FileResponseDTO, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("invalid user ID")
	}

	files, err := uc.fileRepo.FindByUserID(uid)
	if err != nil {
		return nil, err
	}

	responses := make([]*FileResponseDTO, len(files))
	for i, f := range files {
		responses[i] = &FileResponseDTO{
			ID:        f.ID.String(),
			FileName:  f.FileName,
			FilePath:  f.FilePath,
			FileSize:  f.FileSize,
			Status:    f.Status,
			CreatedAt: f.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt: f.UpdatedAt.Format("2006-01-02 15:04:05"),
			ClipCount: f.ClipCount,
		}
	}

	return responses, nil
}

// ProcessVideo handles the actual video processing logic (called by Worker)
func (uc *FileUseCase) ProcessVideo(fileID, userID string, config file.VideoConfig) error {
	log.Printf("Worker starting processing for file: %s, user: %s", fileID, userID)

	// Parse IDs
	fid, err := uuid.Parse(fileID)
	if err != nil {
		return fmt.Errorf("invalid file ID: %w", err)
	}

	uid, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	// 1. Get file info to get path
	fileEntity, err := uc.fileRepo.FindByID(fid)
	if err != nil {
		return fmt.Errorf("failed to find file: %w", err)
	}
	if fileEntity == nil {
		return errors.New("file not found")
	}

	// 2. Update status to PROCESSING
	if err := uc.fileRepo.UpdateStatus(fid, "processing", 0); err != nil {
		log.Printf("Failed to update status to processing: %v", err)
		// Continue processing even if status update fails
	}

	// 3. Call Modal Service (BLOCKING until finished)
	log.Printf("Calling Modal for file: %s with prompt: %s", fileEntity.FilePath, config.Prompt)
	err = uc.modalService.SendVideoToModal(fileEntity.FilePath, config)
	if err != nil {
		// Update status to FAILED
		uc.fileRepo.UpdateStatus(fid, "failed", 0)
		return fmt.Errorf("modal processing failed: %w", err)
	}

	// 4. Processing Done: List clips from Storage
	userFolder := ""
	if len(fileEntity.FilePath) > 0 {
		parts := []rune(fileEntity.FilePath)
		for i, ch := range parts {
			if ch == '/' {
				userFolder = string(parts[:i])
				break
			}
		}
	}

	if userFolder == "" {
		uc.fileRepo.UpdateStatus(fid, "failed", 0)
		return fmt.Errorf("failed to extract user folder from path: %s", fileEntity.FilePath)
	}

	clipsFolder := userFolder + "/clips"
	
	// Add retry logic for listing files (Eventual Consistency)
	var clipPaths []string
	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		clipPaths, err = uc.storageService.ListFiles(clipsFolder)
		if err == nil && len(clipPaths) > 0 {
			break
		}
		if i < maxRetries-1 {
			log.Printf("Retry listing clips (%d/%d)...", i+1, maxRetries)
			time.Sleep(2 * time.Second)
		}
	}

	if err != nil {
		log.Printf("Warning: Failed to list clips after retries: %v", err)
	}

	// 5. Save Clips to DB
	savedCount := 0
	for _, clipPath := range clipPaths {
		clipEntity := clip.NewClip(uid, fid, clipPath)
		if err := uc.clipRepo.Save(clipEntity); err != nil {
			log.Printf("Failed to save clip %s: %v", clipPath, err)
			continue
		}
		savedCount++
	}

	// 6. Update Status to SUCCESS
	finalStatus := "success"
	// if savedCount == 0 { finalStatus = "success_no_clips" } // Optional

	if err := uc.fileRepo.UpdateStatus(fid, finalStatus, savedCount); err != nil {
		return fmt.Errorf("failed to update success status: %w", err)
	}

	log.Printf("Video processing completed successfully. Saved %d clips.", savedCount)

	// Phase 4 hook: Notify completion via RabbitMQ
	if err := uc.messagePublisher.PublishStatusUpdate(
		fileEntity.ID.String(),
		fileEntity.UserID.String(),
		finalStatus,
		savedCount,
	); err != nil {
		log.Printf("Warning: Failed to publish status update: %v", err)
	}

	return nil
}

// GetClip retrieves a clip by ID with user authorization check
func (uc *FileUseCase) GetClip(clipID, userID string) (*ClipResponseDTO, error) {
	cid, err := uuid.Parse(clipID)
	if err != nil {
		return nil, errors.New("invalid clip ID")
	}

	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("invalid user ID")
	}

	clipEntity, err := uc.clipRepo.FindByID(cid)
	if err != nil {
		return nil, err
	}

	if clipEntity == nil {
		return nil, errors.New("clip not found")
	}

	// Verify ownership
	if clipEntity.UserID != uid {
		return nil, errors.New("unauthorized: clip does not belong to user")
	}

	// Generate signed URL for download (expires in 1 hour = 3600 seconds)
	downloadURL, err := uc.storageService.GetSignedURL(clipEntity.FilePath, 3600)
	if err != nil {
		return nil, fmt.Errorf("failed to generate download URL: %w", err)
	}

	return &ClipResponseDTO{
		ID:             clipEntity.ID.String(),
		UploadedFileID: clipEntity.UploadedFileID.String(),
		FilePath:       clipEntity.FilePath,
		DownloadURL:    downloadURL,
		CreatedAt:      clipEntity.CreatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

// DownloadClip downloads a clip file with authorization check
func (uc *FileUseCase) DownloadClip(clipID, userID string) ([]byte, string, error) {
	cid, err := uuid.Parse(clipID)
	if err != nil {
		return nil, "", errors.New("invalid clip ID")
	}

	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, "", errors.New("invalid user ID")
	}

	clipEntity, err := uc.clipRepo.FindByID(cid)
	if err != nil {
		return nil, "", err
	}

	if clipEntity == nil {
		return nil, "", errors.New("clip not found")
	}

	// Verify ownership
	if clipEntity.UserID != uid {
		return nil, "", errors.New("unauthorized: clip does not belong to user")
	}

	// Download file from storage
	fileBytes, err := uc.storageService.Download(clipEntity.FilePath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to download clip: %w", err)
	}

	return fileBytes, clipEntity.FilePath, nil
}

// GetUserClips retrieves all clips for a user
func (uc *FileUseCase) GetUserClips(userID string) ([]*ClipResponseDTO, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("invalid user ID")
	}

	clips, err := uc.clipRepo.FindByUserID(uid)
	if err != nil {
		return nil, err
	}

	responses := make([]*ClipResponseDTO, len(clips))
	for i, c := range clips {
		// Generate signed URL for each clip (expires in 1 hour)
		downloadURL, err := uc.storageService.GetSignedURL(c.FilePath, 3600)
		if err != nil {
			log.Printf("Failed to generate download URL for clip %s: %v", c.ID, err)
			downloadURL = "" // Set empty if failed
		}

		responses[i] = &ClipResponseDTO{
			ID:             c.ID.String(),
			UploadedFileID: c.UploadedFileID.String(),
			FilePath:       c.FilePath,
			DownloadURL:    downloadURL,
			CreatedAt:      c.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}

	return responses, nil
}

// TestListClipsFromStorage is a test method to list clips from storage folder
func (uc *FileUseCase) TestListClipsFromStorage(folderPath string) ([]string, error) {
	log.Printf("Testing list clips from storage folder: %s", folderPath)

	clipPaths, err := uc.storageService.ListFiles(folderPath)
	if err != nil {
		return nil, err
	}

	log.Printf("Found %d clips in folder %s", len(clipPaths), folderPath)
	for i, path := range clipPaths {
		log.Printf("  [%d] %s", i+1, path)
	}

	return clipPaths, nil
}
