package application

import (
	"ai-clipper/server2/internal/file/domain/clip"
	"ai-clipper/server2/internal/file/domain/file"
	"errors"
	"fmt"
	"log"

	"github.com/google/uuid"
)

// FileUseCase handles file upload business logic
type FileUseCase struct {
	fileRepo       file.FileRepository
	clipRepo       clip.ClipRepository
	storageService StorageService
	modalService   ModalService
}

// NewFileUseCase creates a new FileUseCase
func NewFileUseCase(
	fileRepo file.FileRepository,
	clipRepo clip.ClipRepository,
	storageService StorageService,
	modalService ModalService,
) *FileUseCase {
	return &FileUseCase{
		fileRepo:       fileRepo,
		clipRepo:       clipRepo,
		storageService: storageService,
		modalService:   modalService,
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

	return &FileResponseDTO{
		ID:       fileEntity.ID.String(),
		FileName: fileEntity.FileName,
		FilePath: fileEntity.FilePath,
		FileSize: fileEntity.FileSize,
		Message:  "File uploaded successfully",
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
			ID:       f.ID.String(),
			FileName: f.FileName,
			FilePath: f.FilePath,
			FileSize: f.FileSize,
		}
	}

	return responses, nil
}

// ProcessVideo triggers video processing via Modal
func (uc *FileUseCase) ProcessVideo(fileID, userID string) error {
	log.Printf("Starting video processing for file: %s, user: %s", fileID, userID)

	// Parse IDs
	fid, err := uuid.Parse(fileID)
	if err != nil {
		return errors.New("invalid file ID")
	}

	uid, err := uuid.Parse(userID)
	if err != nil {
		return errors.New("invalid user ID")
	}

	// Get file from database
	fileEntity, err := uc.fileRepo.FindByID(fid)
	if err != nil {
		return err
	}

	if fileEntity == nil {
		return errors.New("file not found")
	}

	// Verify ownership
	if fileEntity.UserID != uid {
		return errors.New("unauthorized: file does not belong to user")
	}

	// Call Modal service asynchronously
	go func() {
		err := uc.modalService.ProcessVideo(fileEntity.FilePath)
		if err != nil {
			log.Printf("Modal processing failed for file %s: %v", fileEntity.ID, err)
			return
		}

		// Extract user folder from file path (e.g., "user-xxx/uuid-video.mp4" -> "user-xxx")
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
			log.Printf("Failed to extract user folder from file path: %s", fileEntity.FilePath)
			return
		}

		// Build clips folder path: user-xxx/clips
		clipsFolder := userFolder + "/clips"

		// List all clips from storage
		clipPaths, err := uc.storageService.ListFiles(clipsFolder)
		if err != nil {
			log.Printf("Failed to list clips from storage: %v", err)
			return
		}

		// Save clips to database
		savedCount := 0
		for _, clipPath := range clipPaths {
			clipEntity := clip.NewClip(uid, fid, clipPath)
			if err := uc.clipRepo.Save(clipEntity); err != nil {
				log.Printf("Failed to save clip %s: %v", clipPath, err)
				continue
			}
			savedCount++
		}

		log.Printf("Video processing complete: saved %d clips for file %s", savedCount, fileEntity.ID)
	}()

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
