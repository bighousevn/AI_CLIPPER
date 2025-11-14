package httputil

// HTTPError represents an error response.
type HTTPError struct {
	Error string `json:"error" example:"error message"`
}
