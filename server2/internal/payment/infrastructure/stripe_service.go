package infrastructure

import (
	"fmt"
	"os"

	"github.com/stripe/stripe-go/v83"
	"github.com/stripe/stripe-go/v83/checkout/session"
	"github.com/stripe/stripe-go/v83/customer"
	"github.com/stripe/stripe-go/v83/webhook"
)

// StripeService handles interaction with Stripe API
type StripeService struct {
	secretKey     string
	webhookSecret string
	baseURL       string
}

// NewStripeService creates a new Stripe service
func NewStripeService() *StripeService {
	secretKey := os.Getenv("STRIPE_SECRET_KEY")
	stripe.Key = secretKey

	return &StripeService{
		secretKey:     secretKey,
		webhookSecret: os.Getenv("STRIPE_WEBHOOK_SECRET"),
		baseURL:       os.Getenv("BASE_URL"),
	}
}

// CreateCustomer creates a new Stripe customer
func (s *StripeService) CreateCustomer(email, name string) (string, error) {
	params := &stripe.CustomerParams{
		Email: stripe.String(email),
		Name:  stripe.String(name),
	}
	c, err := customer.New(params)
	if err != nil {
		return "", fmt.Errorf("failed to create stripe customer: %w", err)
	}
	return c.ID, nil
}

// CreateCheckoutSession creates a new checkout session for buying credits
func (s *StripeService) CreateCheckoutSession(customerID, priceID string, successURL, cancelURL string) (string, error) {
	params := &stripe.CheckoutSessionParams{
		Customer: stripe.String(customerID),
		Mode:     stripe.String(string(stripe.CheckoutSessionModePayment)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(priceID),
				Quantity: stripe.Int64(1),
			},
		},
		SuccessURL: stripe.String(successURL),
		CancelURL:  stripe.String(cancelURL),
		// Metadata is useful for webhook verification if needed,
		// but Customer ID is usually enough to link back to user.
	}

	sess, err := session.New(params)
	if err != nil {
		return "", fmt.Errorf("failed to create checkout session: %w", err)
	}

	return sess.URL, nil
}

// ConstructEvent verifies and parses the webhook event
func (s *StripeService) ConstructEvent(payload []byte, signature string) (stripe.Event, error) {
	return webhook.ConstructEvent(payload, signature, s.webhookSecret)
}
