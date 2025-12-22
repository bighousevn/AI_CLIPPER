package application

import (
	userDomain "ai-clipper/server2/internal/auth/domain/user"
	"ai-clipper/server2/internal/file/domain/clip"
	"ai-clipper/server2/internal/file/domain/file"
	messagedomain "ai-clipper/server2/internal/messaging/domain"
	"context"
	"errors"
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

// FileUseCase handles file upload business logic
type FileUseCase struct {
	fileRepo         file.FileRepository
	clipRepo         clip.ClipRepository
	userRepo         userDomain.UserRepository
	storageService   StorageService
	modalService     ModalService
	messagePublisher messagedomain.MessagePublisher
}

// NewFileUseCase creates a new FileUseCase
func NewFileUseCase(
	fileRepo file.FileRepository,
	clipRepo clip.ClipRepository,
	userRepo userDomain.UserRepository,
	storageService StorageService,
	modalService ModalService,
	messagePublisher messagedomain.MessagePublisher,
) *FileUseCase {
	return &FileUseCase{
		fileRepo:         fileRepo,
		clipRepo:         clipRepo,
		userRepo:         userRepo,
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

	// Sanitize filename: replace spaces with underscores to avoid issues in processing
	dto.Header.Filename = strings.ReplaceAll(dto.Header.Filename, " ", "_")

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
		if delErr := uc.storageService.Delete(filePath); delErr != nil {
			log.Printf("Failed to delete file after validation failure: %v", delErr)
		}
		return nil, errors.New("file size exceeds maximum allowed limit")
	}

	// Save file metadata to database
	if err := uc.fileRepo.Save(fileEntity); err != nil {
		log.Printf("Failed to save file metadata: %v", err)
		// Clean up uploaded file
		if delErr := uc.storageService.Delete(filePath); delErr != nil {
			log.Printf("Failed to delete file after save failure: %v", delErr)
		}
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
	ctx := context.Background()

	// Parse IDs
	fid, err := uuid.Parse(fileID)
	if err != nil {
		return fmt.Errorf("invalid file ID: %w", err)
	}

	uid, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	// 0. Check Credits
	user, err := uc.userRepo.FindByID(ctx, uid)
	if err != nil {
		return fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return errors.New("user not found")
	}

	if user.Credits < 1 {
		log.Printf("User %s has insufficient credits (%d). Marking file as no_credit.", uid, user.Credits)
		if updateErr := uc.fileRepo.UpdateStatus(fid, "no_credit", 0); updateErr != nil {
			log.Printf("Failed to update status to no_credit: %v", updateErr)
		}
		if pubErr := uc.messagePublisher.PublishStatusUpdate(fileID, userID, "no_credit", 0); pubErr != nil {
			log.Printf("Failed to publish no_credit status: %v", pubErr)
		}
		return errors.New("insufficient credits")
	}

	// Deduct credit
	user.Credits--
	if err := uc.userRepo.Save(ctx, user); err != nil {
		return fmt.Errorf("failed to deduct credit: %w", err)
	}
	log.Printf("Deducted 1 credit from user %s. Remaining: %d", uid, user.Credits)

	// Refund helper with Retry logic
	refundCredit := func() error {
		var lastErr error
		maxRetries := 3
		for i := 0; i < maxRetries; i++ {
			log.Printf("Refunding credit to user %s (Attempt %d/%d)", uid, i+1, maxRetries)

			// Reload user to ensure we have latest version (optimistic locking safety)
			currentUser, err := uc.userRepo.FindByID(ctx, uid)
			if err != nil {
				lastErr = err
				time.Sleep(500 * time.Millisecond)
				continue
			}

			currentUser.Credits++
			if err := uc.userRepo.Save(ctx, currentUser); err != nil {
				lastErr = err
				log.Printf("Failed to refund attempt %d: %v", i+1, err)
				time.Sleep(500 * time.Millisecond)
				continue
			}

			log.Printf("Successfully refunded credit to user %s", uid)
			return nil
		}

		log.Printf("CRITICAL: Failed to refund credit to user %s after %d attempts: %v", uid, maxRetries, lastErr)
		return fmt.Errorf("failed to refund credit after retries: %w", lastErr)
	}

	// 1. Get file info to get path
	fileEntity, err := uc.fileRepo.FindByID(fid)
	if err != nil {
		if refErr := refundCredit(); refErr != nil {
			return fmt.Errorf("failed to find file: %v, and refund failed: %w", err, refErr)
		}
		return fmt.Errorf("failed to find file: %w", err)
	}
	if fileEntity == nil {
		if refErr := refundCredit(); refErr != nil {
			return fmt.Errorf("file not found, and refund failed: %w", refErr)
		}
		return errors.New("file not found")
	}

	// 2. Update status to PROCESSING
	if err := uc.fileRepo.UpdateStatus(fid, "processing", 0); err != nil {
		log.Printf("Failed to update status to processing: %v", err)
		// Continue processing even if status update fails
	} else {
		// Notify frontend about processing start via SSE
		if pubErr := uc.messagePublisher.PublishStatusUpdate(
			fileEntity.ID.String(),
			fileEntity.UserID.String(),
			"processing",
			0,
		); pubErr != nil {
			log.Printf("Warning: Failed to publish processing status: %v", pubErr)
		}
	}

	// 3. Call Modal Service (BLOCKING until finished)
	log.Printf("Calling Modal for file: %s with prompt: %s", fileEntity.FilePath, config.Prompt)
	err = uc.modalService.SendVideoToModal(fileEntity.FilePath, config)
	if err != nil {
		log.Printf("Modal processing failed for file %s: %v", fid, err)

		// Refund credit on failure
		refErr := refundCredit()

		// Update status to FAILED
		if updateErr := uc.fileRepo.UpdateStatus(fid, "failed", 0); updateErr != nil {
			log.Printf("Failed to update status to failed: %v", updateErr)
		}

		// Notify failure via SSE
		if pubErr := uc.messagePublisher.PublishStatusUpdate(
			fileEntity.ID.String(),
			fileEntity.UserID.String(),
			"failed",
			0,
		); pubErr != nil {
			log.Printf("Warning: Failed to publish failed status: %v", pubErr)
		}

		// CLEANUP: Delete original file
		if delErr := uc.storageService.Delete(fileEntity.FilePath); delErr != nil {
			log.Printf("Cleanup warning: Failed to delete original file %s: %v", fileEntity.FilePath, delErr)
		}

		// CLEANUP: Delete any partial clips
		userFolder := ""
		if len(fileEntity.FilePath) > 0 {
			// Extract "user-id" from "user-id/filename"
			parts := strings.Split(fileEntity.FilePath, "/")
			if len(parts) > 0 {
				userFolder = parts[0]
			}
		}

		if userFolder != "" {
			clipsFolder := userFolder + "/clips"
			// List and delete
			if partialClips, listErr := uc.storageService.ListFiles(clipsFolder); listErr == nil {

				fileUUID := strings.TrimSuffix(filepath.Base(fileEntity.FilePath), filepath.Ext(fileEntity.FilePath))

				for _, pClip := range partialClips {
					if strings.Contains(filepath.Base(pClip), fileUUID) {
						if delErr := uc.storageService.Delete(pClip); delErr != nil {
							log.Printf("Cleanup warning: Failed to delete partial clip %s: %v", pClip, delErr)
						} else {
							log.Printf("Cleanup: Deleted partial clip %s", pClip)
						}
					}
				}
			}
		}

		if refErr != nil {
			return fmt.Errorf("modal processing failed: %v, and refund failed: %w", err, refErr)
		}
		return fmt.Errorf("modal processing failed: %w", err)
	}

	// 4. Processing Done: List clips from Storage
	userFolder := ""
	fileIdentifier := "" // e.g. "uuid-filename"

	if len(fileEntity.FilePath) > 0 {
		// user-id/uuid-filename.mp4
		dir, file := filepath.Split(fileEntity.FilePath)
		userFolder = filepath.Clean(dir) // "user-id"
		ext := filepath.Ext(file)
		fileIdentifier = file[:len(file)-len(ext)]
	}

	if userFolder == "" || fileIdentifier == "" {
		if refErr := refundCredit(); refErr != nil {
			return fmt.Errorf("failed to extract user folder, and refund failed: %w", refErr)
		}
		if updateErr := uc.fileRepo.UpdateStatus(fid, "failed", 0); updateErr != nil {
			log.Printf("Failed to update status to failed: %v", updateErr)
		}
		return fmt.Errorf("failed to extract user folder or identifier from path: %s", fileEntity.FilePath)
	}

	clipsFolder := userFolder + "/clips"
	clipsFolder = filepath.ToSlash(clipsFolder) // Ensure forward slashes for storage API

	// Add retry logic for listing files (Eventual Consistency)
	var clipPaths []string
	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		allClips, err := uc.storageService.ListFiles(clipsFolder)
		if err == nil {
			// Filter clips belonging to this file
			for _, p := range allClips {
				// Check if clip path contains the fileIdentifier
				// Clip path format: user-id/clips/uuid-filename_clip_0.mp4
				if strings.Contains(filepath.Base(p), fileIdentifier) {
					clipPaths = append(clipPaths, p)
				}
			}

			if len(clipPaths) > 0 {
				break
			}
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
		clipEntity := clip.NewClip(uid, fid, fileEntity.FileName, clipPath)
		if err := uc.clipRepo.Save(clipEntity); err != nil {
			log.Printf("Failed to save clip %s: %v", clipPath, err)
			continue
		}
		savedCount++
	}

	// 6. Update Status to SUCCESS or FAILED if no clips
	finalStatus := "success"
	var processingErr error
	if savedCount == 0 {
		finalStatus = "failed"
		log.Printf("Video processing completed but no clips were found/generated. Marking as failed and refunding.")
		if refErr := refundCredit(); refErr != nil {
			processingErr = fmt.Errorf("no clips generated and refund failed: %w", refErr)
		}
	} else {
		log.Printf("Video processing completed successfully. Saved %d clips.", savedCount)
	}

	if err := uc.fileRepo.UpdateStatus(fid, finalStatus, savedCount); err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	// Phase 4 hook: Notify completion via RabbitMQ
	if err := uc.messagePublisher.PublishStatusUpdate(
		fileEntity.ID.String(),
		fileEntity.UserID.String(),
		finalStatus,
		savedCount,
	); err != nil {
		log.Printf("Warning: Failed to publish status update: %v", err)
	}

	return processingErr
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
		SourceName:     clipEntity.SourceName,
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
			SourceName:     c.SourceName,
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
