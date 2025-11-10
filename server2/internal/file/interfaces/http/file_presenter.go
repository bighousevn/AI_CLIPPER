package http

import "github.com/gin-gonic/gin"

// FilePresenter formats HTTP responses for file operations
type FilePresenter struct{}

// NewFilePresenter creates a new FilePresenter
func NewFilePresenter() *FilePresenter {
	return &FilePresenter{}
}

// RespondSuccess sends a success response
func (p *FilePresenter) RespondSuccess(c *gin.Context, statusCode int, data interface{}) {
	c.JSON(statusCode, gin.H{
		"success": true,
		"data":    data,
	})
}

// RespondError sends an error response
func (p *FilePresenter) RespondError(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, gin.H{
		"success": false,
		"error":   message,
	})
}
