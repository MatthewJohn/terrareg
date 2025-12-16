package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// StateStorageService handles secure storage and validation of OAuth/OIDC state parameters
type StateStorageService struct {
	sessionService *SessionService
}

// StateInfo contains information stored for OAuth/OIDC flows
type StateInfo struct {
	State       string    `json:"state"`
	RedirectURL string    `json:"redirect_url"`
	CreatedAt   time.Time `json:"created_at"`
	ExpiresAt   time.Time `json:"expires_at"`
	AuthMethod  string    `json:"auth_method"` // "oidc", "github", etc.
}

// NewStateStorageService creates a new state storage service
func NewStateStorageService(sessionService *SessionService) *StateStorageService {
	return &StateStorageService{
		sessionService: sessionService,
	}
}

// GenerateAndStoreState generates a secure state parameter and stores it with the given information
func (s *StateStorageService) GenerateAndStoreState(ctx context.Context, redirectURL, authMethod string) (string, error) {
	// Generate secure random state
	state, err := s.generateSecureState()
	if err != nil {
		return "", fmt.Errorf("failed to generate secure state: %w", err)
	}

	// Create state info
	stateInfo := &StateInfo{
		State:       state,
		RedirectURL: redirectURL,
		CreatedAt:   time.Now(),
		ExpiresAt:   time.Now().Add(10 * time.Minute), // Short TTL for state
		AuthMethod:  authMethod,
	}

	// Serialize state info
	stateData, err := json.Marshal(stateInfo)
	if err != nil {
		return "", fmt.Errorf("failed to serialize state info: %w", err)
	}

	// Store in session with short TTL
	ttl := 10 * time.Minute
	session, err := s.sessionService.CreateSession(ctx, "state_storage", stateData, &ttl)
	if err != nil {
		return "", fmt.Errorf("failed to store state: %w", err)
	}

	// Return a combination of session ID and state for additional security
	combinedState := fmt.Sprintf("%s:%s", session.ID, state)
	return base64.URLEncoding.EncodeToString([]byte(combinedState)), nil
}

// ValidateAndConsumeState validates a state parameter and returns the stored information
// It also removes the state from storage to prevent replay attacks
func (s *StateStorageService) ValidateAndConsumeState(ctx context.Context, encodedState string) (*StateInfo, error) {
	// Decode the combined state
	decoded, err := base64.URLEncoding.DecodeString(encodedState)
	if err != nil {
		return nil, fmt.Errorf("invalid state format: %w", err)
	}

	// Split session ID and state
	parts := strings.SplitN(string(decoded), ":", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid state format")
	}

	sessionID, state := parts[0], parts[1]

	// Retrieve session
	session, err := s.sessionService.GetSession(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("state not found or expired: %w", err)
	}

	// Check if session is for state storage
	if session == nil || len(session.ProviderSourceAuth) == 0 {
		return nil, fmt.Errorf("invalid state session")
	}

	// Deserialize state info
	var stateInfo StateInfo
	if err := json.Unmarshal(session.ProviderSourceAuth, &stateInfo); err != nil {
		return nil, fmt.Errorf("failed to deserialize state info: %w", err)
	}

	// Validate the state parameter
	if stateInfo.State != state {
		return nil, fmt.Errorf("state mismatch - possible CSRF attack")
	}

	// Check if state has expired
	if time.Now().After(stateInfo.ExpiresAt) {
		// Clean up expired session
		s.sessionService.DeleteSession(ctx, sessionID)
		return nil, fmt.Errorf("state has expired")
	}

	// Delete the session to prevent replay attacks
	if err := s.sessionService.DeleteSession(ctx, sessionID); err != nil {
		// Log error but don't fail the request
		fmt.Printf("Warning: failed to delete state session: %v\n", err)
	}

	return &stateInfo, nil
}

// generateSecureState generates a cryptographically secure random state string
func (s *StateStorageService) generateSecureState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

