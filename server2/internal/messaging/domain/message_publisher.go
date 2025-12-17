package domain

import "ai-clipper/server2/internal/file/domain/file"

type MessagePublisher interface {
	PublishVideoProcessing(fileID, userID, filePath string, config file.VideoConfig) error
	PublishEmailNotification(emailType, to, username, content string) error
	PublishStatusUpdate(fileID, userID, status string, clipCount int) error
	Close() error
}

type VideoProcessingMessage struct {
	FileID   string           `json:"file_id"`
	UserID   string           `json:"user_id"`
	FilePath string           `json:"file_path"`
	Config   file.VideoConfig `json:"config"`
}

type StatusUpdateMessage struct {
	FileID    string `json:"file_id"`
	UserID    string `json:"user_id"`
	Status    string `json:"status"`
	ClipCount int    `json:"clip_count"`
}

type EmailNotificationMessage struct {
	Type     string `json:"type"`
	To       string `json:"to"`
	Username string `json:"username"`
	Content  string `json:"content"`
}
