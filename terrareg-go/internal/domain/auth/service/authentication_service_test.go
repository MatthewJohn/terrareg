package service

import (
	"testing"
	"time"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/config"
	"github.com/matthewjohn/terrareg/terrareg-go/test/testutils/mocks"
	"github.com/stretchr/testify/assert"
)

// TestNewAuthenticationService_NilSessionService verifies that NewAuthenticationService
// returns an error when sessionService is nil
func TestNewAuthenticationService_NilSessionService(t *testing.T) {
	infraCfg := &config.InfrastructureConfig{
		PublicURL: "http://localhost:3000",
		SecretKey: "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef", // 64 hex chars = 32 bytes
	}
	cookieService := NewCookieService(infraCfg)
	mockAuditService := new(mocks.MockAuthenticationAuditService)

	// Test with nil sessionService
	authService, err := NewAuthenticationService(nil, cookieService, mockAuditService)
	assert.Error(t, err)
	assert.Nil(t, authService)
	assert.Contains(t, err.Error(), "sessionService cannot be nil")
}

// TestNewAuthenticationService_NilCookieService verifies that NewAuthenticationService
// returns an error when cookieService is nil
func TestNewAuthenticationService_NilCookieService(t *testing.T) {
	sessionConfig := &SessionDatabaseConfig{
		DefaultTTL:      24 * time.Hour,
		MaxTTL:          30 * 24 * time.Hour,
		CleanupInterval: 1 * time.Hour,
	}
	sessionService := NewSessionService(nil, sessionConfig)
	mockAuditService := new(mocks.MockAuthenticationAuditService)

	// Test with nil cookieService
	authService, err := NewAuthenticationService(sessionService, nil, mockAuditService)
	assert.Error(t, err)
	assert.Nil(t, authService)
	assert.Contains(t, err.Error(), "cookieService cannot be nil")
}

// TestNewAuthenticationService_NilAuditService verifies that NewAuthenticationService
// returns an error when authAuditService is nil
func TestNewAuthenticationService_NilAuditService(t *testing.T) {
	sessionConfig := &SessionDatabaseConfig{
		DefaultTTL:      24 * time.Hour,
		MaxTTL:          30 * 24 * time.Hour,
		CleanupInterval: 1 * time.Hour,
	}
	sessionService := NewSessionService(nil, sessionConfig)

	infraCfg := &config.InfrastructureConfig{
		PublicURL: "http://localhost:3000",
		SecretKey: "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
	}
	cookieService := NewCookieService(infraCfg)

	// Test with nil audit service
	authService, err := NewAuthenticationService(sessionService, cookieService, nil)
	assert.Error(t, err)
	assert.Nil(t, authService)
	assert.Contains(t, err.Error(), "authAuditService cannot be nil")
}

// TestNewAuthenticationService_AllNil verifies that NewAuthenticationService
// returns an error when all dependencies are nil
func TestNewAuthenticationService_AllNil(t *testing.T) {
	// Test with all nil
	authService, err := NewAuthenticationService(nil, nil, nil)
	assert.Error(t, err)
	assert.Nil(t, authService)
	assert.Contains(t, err.Error(), "sessionService cannot be nil")
}

// TestNewAuthenticationService_ValidDependencies verifies that NewAuthenticationService
// returns a valid service when all dependencies are provided
func TestNewAuthenticationService_ValidDependencies(t *testing.T) {
	sessionConfig := &SessionDatabaseConfig{
		DefaultTTL:      24 * time.Hour,
		MaxTTL:          30 * 24 * time.Hour,
		CleanupInterval: 1 * time.Hour,
	}
	sessionService := NewSessionService(nil, sessionConfig)

	infraCfg := &config.InfrastructureConfig{
		PublicURL: "http://localhost:3000",
		SecretKey: "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef", // 64 hex chars = 32 bytes
	}
	cookieService := NewCookieService(infraCfg)
	mockAuditService := new(mocks.MockAuthenticationAuditService)

	// Test with valid dependencies
	authService, err := NewAuthenticationService(sessionService, cookieService, mockAuditService)
	assert.NoError(t, err)
	assert.NotNil(t, authService)
}
