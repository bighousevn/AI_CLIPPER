package sse

import (
	"log"
	"sync"
)

// Event represents an SSE event
type Event struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// Manager manages SSE connections
type Manager struct {
	// clients maps UserID to a list of channels
	clients map[string][]chan Event
	mu      sync.RWMutex
}

// NewManager creates a new SSE manager
func NewManager() *Manager {
	return &Manager{
		clients: make(map[string][]chan Event),
	}
}

// Subscribe adds a new client connection for a user
func (m *Manager) Subscribe(userID string) chan Event {
	m.mu.Lock()
	defer m.mu.Unlock()

	ch := make(chan Event, 10) // Buffer 10 events
	m.clients[userID] = append(m.clients[userID], ch)

	log.Printf("User %s subscribed to SSE (Total clients: %d)", userID, len(m.clients[userID]))
	return ch
}

// Unsubscribe removes a client connection
func (m *Manager) Unsubscribe(userID string, ch chan Event) {
	m.mu.Lock()
	defer m.mu.Unlock()

	clients := m.clients[userID]
	for i, clientCh := range clients {
		if clientCh == ch {
			// Remove channel from slice
			m.clients[userID] = append(clients[:i], clients[i+1:]...)
			close(clientCh)
			log.Printf("User %s unsubscribed from SSE", userID)
			break
		}
	}

	// Clean up map entry if empty
	if len(m.clients[userID]) == 0 {
		delete(m.clients, userID)
	}
}

// SendToUser sends an event to all connected clients of a user
func (m *Manager) SendToUser(userID string, eventType string, data interface{}) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	clients, ok := m.clients[userID]
	if !ok {
		// No clients connected for this user
		return
	}

	event := Event{
		Type: eventType,
		Data: data,
	}

	for _, ch := range clients {
		select {
		case ch <- event:
			// Sent successfully
		default:
			// Channel blocked/full, skip to avoid blocking sender
			log.Printf("Warning: SSE channel full for user %s, dropping event", userID)
		}
	}
}

// SendJsonToUser is a helper to send raw JSON data (for easier integration)
func (m *Manager) SendJsonToUser(userID string, eventType string, jsonData interface{}) {
	m.SendToUser(userID, eventType, jsonData)
}
