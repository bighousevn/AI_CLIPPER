package http

import (
	"ai-clipper/server2/internal/sse"
	"io"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// SSEController handles Server-Sent Events endpoints
type SSEController struct {
	sseManager *sse.Manager
}

// NewSSEController creates a new SSE controller
func NewSSEController(manager *sse.Manager) *SSEController {
	return &SSEController{
		sseManager: manager,
	}
}

// StreamEvents handles the SSE connection
// @Summary Stream real-time events (SSE)
// @Description Establishes a Server-Sent Events connection to receive real-time updates
// @Tags events
// @Produce text/event-stream
// @Security BearerAuth
// @Router /api/v1/events [get]
func (ctrl *SSEController) StreamEvents(c *gin.Context) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		c.Status(401)
		return
	}
	userID := userIDVal.(uuid.UUID).String()

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Transfer-Encoding", "chunked")

	clientChan := ctrl.sseManager.Subscribe(userID)
	defer ctrl.sseManager.Unsubscribe(userID, clientChan)

	c.Stream(func(w io.Writer) bool {
		// Wait for event or context close
		select {
		case event := <-clientChan:
			c.SSEvent(event.Type, event.Data)
			return true
		case <-c.Request.Context().Done():
			return false
		}
	})
}
