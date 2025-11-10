package http

import (
	"ai-clipper/server2/internal/auth/application"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AuthController handles the business logic of authentication requests.
type AuthController struct {
	useCase   application.IAuthUseCase
	presenter *AuthPresenter
}

func NewAuthController(useCase application.IAuthUseCase, presenter *AuthPresenter) *AuthController {
	return &AuthController{
		useCase:   useCase,
		presenter: presenter,
	}
}

func (h *AuthController) Register(c *gin.Context) {
	var req application.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.presenter.RenderError(c, http.StatusBadRequest, err)
		return
	}
	res, err := h.useCase.Register(c.Request.Context(), req)
	if err != nil {
		h.presenter.RenderError(c, http.StatusConflict, err)
		return
	}
	h.presenter.RenderSuccess(c, http.StatusCreated, res)
}

func (h *AuthController) Login(c *gin.Context) {
	var req application.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.presenter.RenderError(c, http.StatusBadRequest, err)
		return
	}
	res, err := h.useCase.Login(c.Request.Context(), req)
	if err != nil {
		h.presenter.RenderError(c, http.StatusUnauthorized, err)
		return
	}
	c.SetCookie("refresh_token", res.RefreshToken, 3600*24*30, "/", "", false, true)
	h.presenter.RenderSuccess(c, http.StatusOK, gin.H{"access_token": res.AccessToken})
}

func (h *AuthController) RefreshToken(c *gin.Context) {
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		h.presenter.RenderError(c, http.StatusUnauthorized, err)
		return
	}
	res, err := h.useCase.RefreshToken(c.Request.Context(), refreshToken)
	if err != nil {
		h.presenter.RenderError(c, http.StatusUnauthorized, err)
		return
	}
	c.SetCookie("refresh_token", res.RefreshToken, 3600*24*30, "/", "", false, true)
	h.presenter.RenderSuccess(c, http.StatusOK, gin.H{"access_token": res.AccessToken})
}

func (h *AuthController) Logout(c *gin.Context) {
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		h.presenter.RenderError(c, http.StatusUnauthorized, err)
		return
	}
	if err := h.useCase.Logout(c.Request.Context(), refreshToken); err != nil {
		h.presenter.RenderError(c, http.StatusInternalServerError, err)
		return
	}
	c.SetCookie("refresh_token", "", -1, "/", "", false, true)
	h.presenter.RenderSuccess(c, http.StatusOK, gin.H{"message": "logout successful"})
}

func (h *AuthController) ForgotPassword(c *gin.Context) {
	var req application.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.presenter.RenderError(c, http.StatusBadRequest, err)
		return
	}
	if err := h.useCase.ForgotPassword(c.Request.Context(), req); err != nil {
		h.presenter.RenderError(c, http.StatusInternalServerError, err)
		return
	}
	h.presenter.RenderSuccess(c, http.StatusOK, gin.H{"message": "If an account with that email exists, a password reset link has been sent."})
}

func (h *AuthController) ResetPassword(c *gin.Context) {
	var req application.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.presenter.RenderError(c, http.StatusBadRequest, err)
		return
	}
	if err := h.useCase.ResetPassword(c.Request.Context(), req); err != nil {
		h.presenter.RenderError(c, http.StatusBadRequest, err)
		return
	}
	h.presenter.RenderSuccess(c, http.StatusOK, gin.H{"message": "Password has been reset successfully."})
}

func (h *AuthController) VerifyEmail(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		h.presenter.RenderError(c, http.StatusBadRequest, errors.New("token is required"))
		return
	}
	if err := h.useCase.VerifyEmail(c.Request.Context(), token); err != nil {
		h.presenter.RenderError(c, http.StatusBadRequest, err)
		return
	}
	h.presenter.RenderSuccess(c, http.StatusOK, gin.H{"message": "Email has been verified successfully."})
}

func (h *AuthController) GetProfile(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		h.presenter.RenderError(c, http.StatusUnauthorized, errors.New("user not found in context"))
		return
	}
	profile, err := h.useCase.GetUserProfile(c.Request.Context(), userID.(uuid.UUID))
	if err != nil {
		h.presenter.RenderError(c, http.StatusNotFound, err)
		return
	}
	h.presenter.RenderSuccess(c, http.StatusOK, profile)
}

func (h *AuthController) ChangePassword(c *gin.Context) {
	var req application.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.presenter.RenderError(c, http.StatusBadRequest, err)
		return
	}
	userID, exists := c.Get("userID")
	if !exists {
		h.presenter.RenderError(c, http.StatusUnauthorized, errors.New("user not found in context"))
		return
	}
	if err := h.useCase.ChangePassword(c.Request.Context(), userID.(uuid.UUID), req); err != nil {
		h.presenter.RenderError(c, http.StatusInternalServerError, err)
		return
	}
	h.presenter.RenderSuccess(c, http.StatusOK, gin.H{"message": "Password changed successfully"})
}
