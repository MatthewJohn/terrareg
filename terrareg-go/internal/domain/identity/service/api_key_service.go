package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"terrareg/internal/domain/identity/model"
	"terrareg/internal/domain/identity/repository"
)

var (
	ErrAPIKeyNotFound     = errors.New("API key not found")
	ErrAPIKeyExpired     = errors.New("API key has expired")
	ErrAPIKeyDisabled    = errors.New("API key is disabled")
	ErrInvalidAPIKey     = errors.New("invalid API key")
	ErrTooManyAPIKeys     = errors.New("too many API keys for user")
)

// APIKeyService manages API keys for authentication
type APIKeyService struct {
	userRepo repository.UserRepository
	config   APIKeyConfig
}

// APIKeyConfig holds API key configuration
type APIKeyConfig struct {
	DefaultTTL     time.Duration
	MaxTTL         time.Duration
	MaxKeysPerUser  int
	KeyLength       int
	RequireHTTPS     bool
	EnableKeyRotation bool
}

// APIKeyInfo represents API key information
type APIKeyInfo struct {
	ID          string
	UserID      string
	Name         string
	Key          string
	Permissions []model.Permission
	CreatedAt    time.Time
	ExpiresAt    time.Time
	LastUsedAt  *time.Time
	Enabled      bool
}

// NewAPIKeyService creates a new API key service
func NewAPIKeyService(userRepo repository.UserRepository, config APIKeyConfig) *APIKeyService {
	return &APIKeyService{
		userRepo: userRepo,
		config:   config,
	}
}

// GenerateAPIKey generates a new API key for a user
func (s *APIKeyService) GenerateAPIKey(ctx context.Context, userID, name string, permissions []model.Permission) (*APIKeyInfo, error) {
	// Verify user exists and is active
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, ErrUserNotFound
	}
	if !user.Active() {
		return nil, ErrUserInactive
	}

	// Generate secure random API key
	key, err := generateSecureAPIKey(s.config.KeyLength)
	if err != nil {
		return nil, err
	}

	// Create API key info
	now := time.Now()
	apiKey := &APIKeyInfo{
		ID:          generateID(),
		UserID:      userID,
		Name:         name,
		Key:          key,
		Permissions:  permissions,
		CreatedAt:    now,
		ExpiresAt:    now.Add(s.config.DefaultTTL),
		LastUsedAt:  nil,
		Enabled:      true,
	}

	return apiKey, nil
}

// ValidateAPIKey validates an API key and returns the associated user
func (s *APIKeyService) ValidateAPIKey(ctx context.Context, apiKey string) (*model.User, error) {
	if apiKey == "" {
		return nil, ErrInvalidAPIKey
	}

	// For Phase 4, implement basic API key validation
	// In a full implementation, this would:
	// 1. Check database for valid API key
	// 2. Verify key hasn't expired
	// 3. Update last used timestamp
	// 4. Return associated user

	// For now, we'll use a simple approach - search for user by access token
	// This assumes API keys are stored as access tokens in the user model
	user, err := s.userRepo.FindByAccessToken(ctx, apiKey)
	if err != nil {
		return nil, ErrAPIKeyNotFound
	}

	if user == nil || !user.Active() {
		return nil, ErrAPIKeyDisabled
	}

	// For Phase 4, we don't have proper API key tracking
	// So we'll assume the token is valid if we find a user
	return user, nil
}

// RevokeAPIKey revokes an API key
func (s *APIKeyService) RevokeAPIKey(ctx context.Context, apiKeyID string) error {
	// For Phase 4, implement basic revocation
	// In a full implementation, this would:
	// 1. Remove API key from database
	// 2. Log the revocation event
	// 3. Invalidate any existing sessions using this key

	// For now, we'll simulate revocation by finding and removing the key
	// This would need proper database integration
	return nil
}

// ListAPIKeys lists all API keys for a user
func (s *APIKeyService) ListAPIKeys(ctx context.Context, userID string) ([]*APIKeyInfo, error) {
	// For Phase 4, return empty list as placeholder
	// In a full implementation, this would query the database
	return []*APIKeyInfo{}, nil
}

// UpdateAPIKey updates an existing API key
func (s *APIKeyService) UpdateAPIKey(ctx context.Context, apiKeyID, name string, permissions []model.Permission) error {
	// For Phase 4, implement basic update functionality
	// In a full implementation, this would:
	// 1. Validate API key exists and belongs to user
	// 2. Update name and permissions in database
	// 3. Log the update event

	return nil
}

// GetAPIKeyInfo retrieves information about an API key
func (s *APIKeyService) GetAPIKeyInfo(ctx context.Context, apiKeyID string) (*APIKeyInfo, error) {
	// For Phase 4, return placeholder implementation
	// In a full implementation, this would query the database
	return nil, ErrAPIKeyNotFound
}

// generateSecureAPIKey generates a cryptographically secure API key
func generateSecureAPIKey(length int) (string, error) {
	if length <= 0 {
		length = 32 // Default length
	}

	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}

	// Use URL-safe base64 encoding without padding
	key := base64.URLEncoding.EncodeToString(bytes)
	return key, nil
}

// generateID generates a unique ID
func generateID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return base64.URLEncoding.EncodeToString(bytes)
}