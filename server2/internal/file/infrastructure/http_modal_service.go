package infrastructure

import (
	"ai-clipper/server2/internal/file/domain/file"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// HTTPModalService implements ModalService using HTTP client
type HTTPModalService struct {
	modalURL   string
	modalToken string
	httpClient *http.Client
}

// NewHTTPModalService creates a new HTTP Modal service
func NewHTTPModalService(modalURL, modalToken string) *HTTPModalService {
	return &HTTPModalService{
		modalURL:   modalURL,
		modalToken: modalToken,
		httpClient: &http.Client{
			Timeout: 10 * time.Minute, // Video processing can take long time
		},
	}
}

// SendVideoToModal calls Modal endpoint to trigger video processing
func (s *HTTPModalService) SendVideoToModal(storagePath string, config file.VideoConfig) error {
	requestBody := map[string]any{
		"storage_path": storagePath,
		"config":       config,
	}
	log.Printf("Sending video to Modal API for storage path: %s", storagePath)

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", s.modalURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.modalToken))

	log.Printf("Sending request to Modal API: %s", s.modalURL)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to call Modal API: %w", err)
	}
	defer resp.Body.Close()

	log.Printf("Modal API response status: %d", resp.StatusCode)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("Modal API returned error status: %d", resp.StatusCode)
	}

	log.Printf("Modal API call successful for storage path: %s", storagePath)
	return nil
}
