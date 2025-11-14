package http

import (
	"ai-clipper/server2/internal/auth/application"
	_ "ai-clipper/server2/internal/httputil"
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

// Register godoc
// @Summary Register a new user
// @Description Register a new user with email and password
// @Tags auth
// @Accept  json
// @Produce  json
// @Param   user     body       application.RegisterRequest     true  "User registration info"
// @Success 201 {object} application.RegisterResponse "Successfully registered"
// @Failure 400 {object} httputil.HTTPError "Invalid request body"
// @Failure 409 {object} httputil.HTTPError "User with this email already exists"
// @Router /auth/register [post]
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

// Login godoc
// @Summary Login a user
// @Description Login a user with email and password
// @Tags auth
// @Accept  json
// @Produce  json
// @Param   user     body       application.LoginRequest     true  "User login info"
// @Success 200 {object} application.LoginResponse "Successfully logged in"
// @Failure 400 {object} httputil.HTTPError "Invalid request body"
// @Failure 401 {object} httputil.HTTPError "Invalid credentials"
// @Router /auth/login [post]
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

// RefreshToken godoc
// @Summary Refresh access token
// @Description Refresh access token using refresh token
// @Tags auth
// @Produce  json
// @Success 200 {object} application.RefreshTokenResponse "Successfully refreshed token"
// @Failure 401 {object} httputil.HTTPError "Invalid refresh token"
// @Router /auth/refresh-token [post]
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

// Logout godoc
// @Summary Logout a user
// @Description Logout a user
// @Tags auth
// @Success 200 {object} object "Successfully logged out"
// @Failure 401 {object} httputil.HTTPError "Invalid refresh token"
// @Router /auth/logout [get]
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

// ForgotPassword godoc
// @Summary Forgot password
// @Description Forgot password
// @Tags auth
// @Accept  json
// @Produce  json
// @Param   email     body       application.ForgotPasswordRequest     true  "Email"
// @Success 200 {object} object "If an account with that email exists, a password reset link has been sent."
// @Failure 400 {object} httputil.HTTPError "Invalid request body"
// @Failure 500 {object} httputil.HTTPError "Internal server error"
// @Router /auth/forgot-password [post]
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

// ResetPassword godoc
// @Summary Reset password
// @Description Reset password
// @Tags auth
// @Accept  json
// @Produce  json
// @Param   token     body       application.ResetPasswordRequest     true  "Token"
// @Success 200 {object} object "Password has been reset successfully."
// @Failure 400 {object} httputil.HTTPError "Invalid request body"
// @Router /auth/reset-password [post]
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

// VerifyEmail godoc
// @Summary Verify email
// @Description Verify email
// @Tags auth
// @Produce  json
// @Param   token     query       string     true  "Token"
// @Success 200 {object} object "Email has been verified successfully."
// @Failure 400 {object} httputil.HTTPError "Invalid token"
// @Router /auth/verify-email [get]
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

// GetProfile godoc
// @Summary Get user profile
// @Description Get user profile
// @Tags users
// @Produce  json
// @Success 200 {object} application.UserProfileResponse "Successfully retrieved profile"
// @Failure 401 {object} httputil.HTTPError "User not found in context"
// @Failure 404 {object} httputil.HTTPError "User not found"
// @Router /users/me [get]
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

// ChangePassword godoc
// @Summary Change password
// @Description Change password
// @Tags users
// @Accept  json
// @Produce  json
// @Param   password     body       application.ChangePasswordRequest     true  "Password"
// @Success 200 {object} object "Password changed successfully"
// @Failure 400 {object} httputil.HTTPError "Invalid request body"
// @Failure 401 {object} httputil.HTTPError "User not found in context"
// @Failure 500 {object} httputil.HTTPError "Internal server error"
// @Router /users/me/password [post]
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