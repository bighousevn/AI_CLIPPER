package http

import (
	"ai-clipper/server2/internal/auth/application"
	"ai-clipper/server2/internal/middleware"
	paymentApp "ai-clipper/server2/internal/payment/application"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// PaymentController handles payment HTTP requests
type PaymentController struct {
	useCase *paymentApp.PaymentUseCase
}

// NewPaymentController creates a new PaymentController
func NewPaymentController(useCase *paymentApp.PaymentUseCase) *PaymentController {
	return &PaymentController{useCase: useCase}
}

// Checkout creates a checkout session
// @Summary Create Checkout Session
// @Description Creates a Stripe checkout session for buying credits
// @Tags Payment
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body paymentApp.CheckoutRequest true "Checkout Request"
// @Success 200 {object} paymentApp.CheckoutResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/payment/checkout [post]
func (ctrl *PaymentController) Checkout(c *gin.Context) {
	var req paymentApp.CheckoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userIDVal, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID := userIDVal.(uuid.UUID)

	res, err := ctrl.useCase.CreateCheckoutSession(c.Request.Context(), userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

// RegisterRoutes registers payment routes
func RegisterRoutes(router *gin.Engine, ctrl *PaymentController, tokenGen application.TokenGenerator) {
	group := router.Group("/api/v1/payment")
	group.Use(middleware.AuthMiddleware(tokenGen))
	{
		group.POST("/checkout", ctrl.Checkout)
	}
}
