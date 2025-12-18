package application

// CheckoutRequest represents the request body for checkout
type CheckoutRequest struct {
	Pack string `json:"pack" binding:"required,oneof=small medium large"`
}

// CheckoutResponse represents the response containing the checkout URL
type CheckoutResponse struct {
	CheckoutURL string `json:"checkout_url"`
}
