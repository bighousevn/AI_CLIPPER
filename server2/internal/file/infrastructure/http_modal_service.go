package infrastructure

import (
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

// ModalResponse represents the response from Modal API
type ModalResponse struct {
	ClipPaths []string `json:"clip_paths"`
	Message   string   `json:"message"`
}

// ProcessVideo calls Modal endpoint to process video and returns clip paths
func (s *HTTPModalService) ProcessVideo(storagePath string) ([]string, error) {
	requestBody := map[string]string{
		"storage_path": storagePath,
	}
	log.Printf("Processing video at storage path: %s", storagePath)

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", s.modalURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.modalToken))

	log.Printf("Sending request to Modal API: %s", s.modalURL)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call Modal API: %w", err)
	}
	defer resp.Body.Close()

	log.Printf("Modal API response status: %d", resp.StatusCode)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return nil, fmt.Errorf("Modal API returned error status: %d", resp.StatusCode)
	}

	// Parse response to get clip paths
	var modalResp ModalResponse
	if err := json.NewDecoder(resp.Body).Decode(&modalResp); err != nil {
		log.Printf("Failed to decode Modal response: %v", err)
		return nil, fmt.Errorf("failed to decode Modal response: %w", err)
	}

	log.Printf("Modal API call successful, received %d clips for storage path: %s", len(modalResp.ClipPaths), storagePath)
	return modalResp.ClipPaths, nil
}
