package http

import (
	userDomain "ai-clipper/server2/internal/auth/domain/user"
	"ai-clipper/server2/internal/payment/infrastructure"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v83"
)

// WebhookController handles Stripe webhooks
type WebhookController struct {
	stripeService *infrastructure.StripeService
	userRepo      userDomain.UserRepository
}

// NewWebhookController creates a new WebhookController
func NewWebhookController(stripeService *infrastructure.StripeService, userRepo userDomain.UserRepository) *WebhookController {
	return &WebhookController{
		stripeService: stripeService,
		userRepo:      userRepo,
	}
}

// HandleWebhook handles incoming Stripe webhook events
// @Summary Handle Stripe Webhook
// @Description Handles stripe webhook events to update user credits
// @Tags Payment
// @Accept json
// @Produce json
// @Router /api/v1/webhook/stripe [post]
func (ctrl *WebhookController) HandleWebhook(c *gin.Context) {
	const MaxBodyBytes = int64(65536)
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, MaxBodyBytes)
	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Printf("Error reading request body: %v", err)
		c.Status(http.StatusServiceUnavailable)
		return
	}

	signature := c.GetHeader("stripe-signature")
	event, err := ctrl.stripeService.ConstructEvent(payload, signature)
	if err != nil {
		log.Printf("Error verifying webhook signature: %v", err)
		c.Status(http.StatusBadRequest)
		return
	}

	if event.Type == "checkout.session.completed" {
		var session stripe.CheckoutSession
		err := json.Unmarshal(event.Data.Raw, &session)
		if err != nil {
			log.Printf("Error unmarshaling checkout session: %v", err)
			c.Status(http.StatusBadRequest)
			return
		}

		if session.Customer == nil {
			log.Printf("Checkout session missing customer ID")
			c.Status(http.StatusBadRequest)
			return
		}

		customerID := session.Customer.ID
		amountTotal := session.AmountTotal

		// Find user
		user, err := ctrl.userRepo.FindByStripeCustomerID(c.Request.Context(), customerID)
		if err != nil {
			log.Printf("Error finding user by stripe customer ID %s: %v", customerID, err)
			c.Status(http.StatusInternalServerError)
			return
		}
		if user == nil {
			log.Printf("User not found for stripe customer ID %s", customerID)
			// Return 200 to acknowledge receipt even if user not found to prevent retries
			c.Status(http.StatusOK)
			return
		}

		// Calculate credits
		var creditsToAdd int
		switch amountTotal {
		case 2000: // $20.00
			creditsToAdd = 20
		case 4799: // $47.99
			creditsToAdd = 50
		case 8999: // $89.99
			creditsToAdd = 100
		default:
			log.Printf("Warning: Unknown amount %d for user %s. No credits added.", amountTotal, user.ID)
			creditsToAdd = 0
		}

		if creditsToAdd > 0 {
			user.Credits += creditsToAdd
			if err := ctrl.userRepo.Save(c.Request.Context(), user); err != nil {
				log.Printf("Error saving user credits: %v", err)
				c.Status(http.StatusInternalServerError)
				return
			}
			log.Printf("Added %d credits to user %s (Total: %d)", creditsToAdd, user.ID, user.Credits)
		}
	}

	c.Status(http.StatusOK)
}

// RegisterWebhookRoutes registers webhook routes
func RegisterWebhookRoutes(router *gin.Engine, ctrl *WebhookController) {
	group := router.Group("/api/v1/webhook")
	{
		group.POST("/stripe", ctrl.HandleWebhook)
	}
}
