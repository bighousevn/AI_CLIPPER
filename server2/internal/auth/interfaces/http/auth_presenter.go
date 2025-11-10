package http

import (
	"github.com/gin-gonic/gin"
)

// AuthPresenter handles the presentation logic for authentication responses.
type AuthPresenter struct{}

func NewAuthPresenter() *AuthPresenter {
	return &AuthPresenter{}
}

// RenderSuccess formats and sends a successful JSON response.
func (p *AuthPresenter) RenderSuccess(c *gin.Context, status int, data interface{}) {
	c.JSON(status, data)
}

// RenderError formats and sends an error JSON response.
func (p *AuthPresenter) RenderError(c *gin.Context, status int, err error) {
	c.JSON(status, gin.H{"error": err.Error()})
}
