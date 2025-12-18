package application

import (
	userDomain "ai-clipper/server2/internal/auth/domain/user"
	"ai-clipper/server2/internal/payment/infrastructure"
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/google/uuid"
)

// PaymentUseCase handles payment business logic
type PaymentUseCase struct {
	stripeService *infrastructure.StripeService
	userRepo      userDomain.UserRepository
}

// NewPaymentUseCase creates a new PaymentUseCase
func NewPaymentUseCase(stripeService *infrastructure.StripeService, userRepo userDomain.UserRepository) *PaymentUseCase {
	return &PaymentUseCase{
		stripeService: stripeService,
		userRepo:      userRepo,
	}
}

// CreateCheckoutSession creates a checkout session for a specific credit pack
func (uc *PaymentUseCase) CreateCheckoutSession(ctx context.Context, userID uuid.UUID, req CheckoutRequest) (*CheckoutResponse, error) {
	// 1. Get User to retrieve Stripe Customer ID
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil || user == nil {
		return nil, errors.New("user not found")
	}

	if user.StripeCustomerID == nil {
		// Attempt to create customer if missing (self-healing)
		customerID, err := uc.stripeService.CreateCustomer(user.Email, user.Username)
		if err != nil {
			return nil, fmt.Errorf("failed to create stripe customer: %w", err)
		}
		user.StripeCustomerID = &customerID
		if err := uc.userRepo.Save(ctx, user); err != nil {
			return nil, fmt.Errorf("failed to update user with stripe id: %w", err)
		}
	}

	// 2. Determine Price ID
	var priceID string
	switch req.Pack {
	case "small":
		priceID = os.Getenv("STRIPE_SMALL_CREDIT_PACK")
	case "medium":
		priceID = os.Getenv("STRIPE_MEDIUM_CREDIT_PACK")
	case "large":
		priceID = os.Getenv("STRIPE_LARGE_CREDIT_PACK")
	default:
		return nil, errors.New("invalid pack")
	}

	if priceID == "" {
		return nil, errors.New("price id not configured")
	}

	// 3. Create Checkout Session
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:3000"
	}
	successURL := baseURL + "/dashboard/billing?status=success"
	cancelURL := baseURL + "/dashboard/billing?status=cancel"

	checkoutURL, err := uc.stripeService.CreateCheckoutSession(*user.StripeCustomerID, priceID, successURL, cancelURL)
	if err != nil {
		return nil, err
	}

	return &CheckoutResponse{CheckoutURL: checkoutURL}, nil
}
