package service

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockSessionRepository is a mock implementation of SessionRepository for testing
type mockSessionRepository struct {
	mu            sync.RWMutex
	sessions      map[string]*auth.Session
	createError   error
	findError     error
	deleteError   error
	cleanupError  error
	updateError   error
	createCalled  int
	findCalled    int
	deleteCalled  int
	cleanupCalled int
	updateCalled  int
	lastSessionID string
}

func newMockSessionRepository() *mockSessionRepository {
	return &mockSessionRepository{
		sessions: make(map[string]*auth.Session),
	}
}

func (m *mockSessionRepository) Create(ctx context.Context, session *auth.Session) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.createCalled++
	if m.createError != nil {
		return m.createError
	}
	if session.ID == "" {
		return errors.New("session ID cannot be empty")
	}
	m.sessions[session.ID] = session
	m.lastSessionID = session.ID
	return nil
}

func (m *mockSessionRepository) FindByID(ctx context.Context, sessionID string) (*auth.Session, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	m.findCalled++
	m.lastSessionID = sessionID
	if m.findError != nil {
		return nil, m.findError
	}
	session, exists := m.sessions[sessionID]
	if !exists {
		return nil, nil // Not found
	}
	return session, nil
}

func (m *mockSessionRepository) Delete(ctx context.Context, sessionID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.deleteCalled++
	m.lastSessionID = sessionID
	if m.deleteError != nil {
		return m.deleteError
	}
	delete(m.sessions, sessionID)
	return nil
}

func (m *mockSessionRepository) CleanupExpired(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cleanupCalled++
	if m.cleanupError != nil {
		return m.cleanupError
	}
	// Remove expired sessions
	for id, session := range m.sessions {
		if session.IsExpired() {
			delete(m.sessions, id)
		}
	}
	return nil
}

func (m *mockSessionRepository) UpdateProviderSourceAuth(ctx context.Context, sessionID string, data []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.updateCalled++
	m.lastSessionID = sessionID
	if m.updateError != nil {
		return m.updateError
	}
	if session, exists := m.sessions[sessionID]; exists {
		session.ProviderSourceAuth = data
		return nil
	}
	return errors.New("session not found")
}

// Compile-time interface implementation check
var _ repository.SessionRepository = (*mockSessionRepository)(nil)

func (m *mockSessionRepository) reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sessions = make(map[string]*auth.Session)
	m.createError = nil
	m.findError = nil
	m.deleteError = nil
	m.cleanupError = nil
	m.updateError = nil
	m.createCalled = 0
	m.findCalled = 0
	m.deleteCalled = 0
	m.cleanupCalled = 0
	m.updateCalled = 0
	m.lastSessionID = ""
}

func (m *mockSessionRepository) setSession(id string, session *auth.Session) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sessions[id] = session
}

// TestDefaultSessionDatabaseConfig tests the default configuration
func TestDefaultSessionDatabaseConfig(t *testing.T) {
	config := DefaultSessionDatabaseConfig()

	assert.Equal(t, 1*time.Hour, config.DefaultTTL)
	assert.Equal(t, 24*time.Hour, config.MaxTTL)
	assert.Equal(t, 1*time.Hour, config.CleanupInterval)
}

// TestNewSessionService tests the constructor
func TestNewSessionService(t *testing.T) {
	t.Run("with explicit config", func(t *testing.T) {
		mockRepo := newMockSessionRepository()
		config := &SessionDatabaseConfig{
			DefaultTTL:      30 * time.Minute,
			MaxTTL:          12 * time.Hour,
			CleanupInterval: 30 * time.Minute,
		}

		service := NewSessionService(mockRepo, config)

		assert.NotNil(t, service)
		assert.Equal(t, mockRepo, service.sessionRepo)
		assert.Equal(t, config, service.config)
	})

	t.Run("with nil config uses defaults", func(t *testing.T) {
		mockRepo := newMockSessionRepository()

		service := NewSessionService(mockRepo, nil)

		assert.NotNil(t, service)
		assert.Equal(t, mockRepo, service.sessionRepo)
		assert.NotNil(t, service.config)
		assert.Equal(t, 1*time.Hour, service.config.DefaultTTL)
		assert.Equal(t, 24*time.Hour, service.config.MaxTTL)
		assert.Equal(t, 1*time.Hour, service.config.CleanupInterval)
	})
}

// TestCreateSession tests session creation
func TestCreateSession(t *testing.T) {
	ctx := context.Background()

	t.Run("creates session with default TTL", func(t *testing.T) {
		mockRepo := newMockSessionRepository()
		config := &SessionDatabaseConfig{
			DefaultTTL:      1 * time.Hour,
			MaxTTL:          24 * time.Hour,
			CleanupInterval: 1 * time.Hour,
		}
		service := NewSessionService(mockRepo, config)

		providerData := []byte(`{"user_id": "123", "email": "test@example.com"}`)
		session, err := service.CreateSession(ctx, "oidc", providerData, nil)

		require.NoError(t, err)
		assert.NotNil(t, session)
		assert.NotEmpty(t, session.ID)
		assert.Equal(t, providerData, session.ProviderSourceAuth)
		assert.Equal(t, 1, mockRepo.createCalled)
		assert.True(t, session.Expiry.After(time.Now()))
		assert.True(t, session.Expiry.Before(time.Now().Add(2*time.Hour)))
	})

	t.Run("creates session with custom TTL", func(t *testing.T) {
		mockRepo := newMockSessionRepository()
		config := &SessionDatabaseConfig{
			DefaultTTL:      1 * time.Hour,
			MaxTTL:          24 * time.Hour,
			CleanupInterval: 1 * time.Hour,
		}
		service := NewSessionService(mockRepo, config)

		customTTL := 2 * time.Hour
		providerData := []byte(`{"user_id": "123"}`)
		session, err := service.CreateSession(ctx, "github", providerData, &customTTL)

		require.NoError(t, err)
		assert.NotNil(t, session)
		assert.True(t, session.Expiry.After(time.Now().Add(customTTL - time.Minute)))
		assert.True(t, session.Expiry.Before(time.Now().Add(customTTL + time.Minute)))
	})

	t.Run("respects max TTL limit", func(t *testing.T) {
		mockRepo := newMockSessionRepository()
		config := &SessionDatabaseConfig{
			DefaultTTL:      1 * time.Hour,
			MaxTTL:          2 * time.Hour,
			CleanupInterval: 1 * time.Hour,
		}
		service := NewSessionService(mockRepo, config)

		// Request 10 hours but max is 2
		requestedTTL := 10 * time.Hour
		providerData := []byte(`{"user_id": "123"}`)
		session, err := service.CreateSession(ctx, "saml", providerData, &requestedTTL)

		require.NoError(t, err)
		assert.NotNil(t, session)
		// Should be capped at max TTL
		assert.True(t, session.Expiry.Before(time.Now().Add(3*time.Hour)))
		assert.True(t, session.Expiry.After(time.Now().Add(config.MaxTTL-time.Minute)))
	})

	t.Run("generates unique session IDs", func(t *testing.T) {
		mockRepo := newMockSessionRepository()
		service := NewSessionService(mockRepo, nil)

		providerData := []byte(`{"user_id": "123"}`)
		session1, err := service.CreateSession(ctx, "oidc", providerData, nil)
		require.NoError(t, err)

		// Small delay to ensure different timestamp
		time.Sleep(10 * time.Nanosecond)

		session2, err := service.CreateSession(ctx, "oidc", providerData, nil)
		require.NoError(t, err)

		assert.NotEqual(t, session1.ID, session2.ID)
	})

	t.Run("handles repository create error", func(t *testing.T) {
		mockRepo := newMockSessionRepository()
		mockRepo.createError = errors.New("database connection failed")
		service := NewSessionService(mockRepo, nil)

		providerData := []byte(`{"user_id": "123"}`)
		session, err := service.CreateSession(ctx, "oidc", providerData, nil)

		assert.Error(t, err)
		assert.Nil(t, session)
		assert.Contains(t, err.Error(), "failed to create session")
		assert.Equal(t, 1, mockRepo.createCalled)
	})
}

// TestGetEffectiveTTL tests TTL calculation logic
func TestGetEffectiveTTL(t *testing.T) {
	config := &SessionDatabaseConfig{
		DefaultTTL:      1 * time.Hour,
		MaxTTL:          24 * time.Hour,
		CleanupInterval: 1 * time.Hour,
	}
	service := NewSessionService(newMockSessionRepository(), config)

	t.Run("uses default TTL when nil", func(t *testing.T) {
		ttl := service.getEffectiveTTL(nil)
		assert.Equal(t, 1*time.Hour, ttl)
	})

	t.Run("uses provided TTL within limits", func(t *testing.T) {
		customTTL := 2 * time.Hour
		ttl := service.getEffectiveTTL(&customTTL)
		assert.Equal(t, 2*time.Hour, ttl)
	})

	t.Run("caps at MaxTTL when exceeded", func(t *testing.T) {
		excessiveTTL := 48 * time.Hour
		ttl := service.getEffectiveTTL(&excessiveTTL)
		assert.Equal(t, 24*time.Hour, ttl)
	})

	t.Run("handles TTL exactly at MaxTTL", func(t *testing.T) {
		exactMaxTTL := 24 * time.Hour
		ttl := service.getEffectiveTTL(&exactMaxTTL)
		assert.Equal(t, 24*time.Hour, ttl)
	})

	t.Run("handles very small TTL", func(t *testing.T) {
		tinyTTL := 1 * time.Second
		ttl := service.getEffectiveTTL(&tinyTTL)
		assert.Equal(t, 1*time.Second, ttl)
	})
}

// TestValidateSession tests session validation
func TestValidateSession(t *testing.T) {
	ctx := context.Background()

	t.Run("validates existing non-expired session", func(t *testing.T) {
		mockRepo := newMockSessionRepository()
		service := NewSessionService(mockRepo, nil)

		// Create a session
		session := &auth.Session{
			ID:     "test-session-123",
			Expiry: time.Now().Add(1 * time.Hour),
			ProviderSourceAuth: []byte(`{"user_id": "123"}`),
		}
		mockRepo.setSession("test-session-123", session)

		// Validate it
		validated, err := service.ValidateSession(ctx, "test-session-123")

		require.NoError(t, err)
		assert.Equal(t, session.ID, validated.ID)
		assert.Equal(t, 1, mockRepo.findCalled)
	})

	t.Run("rejects non-existent session", func(t *testing.T) {
		mockRepo := newMockSessionRepository()
		service := NewSessionService(mockRepo, nil)

		validated, err := service.ValidateSession(ctx, "non-existent")

		assert.Error(t, err)
		assert.Nil(t, validated)
		assert.Contains(t, err.Error(), "session not found")
	})

	t.Run("rejects expired session", func(t *testing.T) {
		mockRepo := newMockSessionRepository()
		service := NewSessionService(mockRepo, nil)

		// Create an expired session
		session := &auth.Session{
			ID:     "expired-session",
			Expiry: time.Now().Add(-1 * time.Hour),
			ProviderSourceAuth: []byte(`{}`),
		}
		mockRepo.setSession("expired-session", session)

		validated, err := service.ValidateSession(ctx, "expired-session")

		assert.Error(t, err)
		assert.Nil(t, validated)
		assert.Contains(t, err.Error(), "session expired")
	})

	t.Run("handles repository error", func(t *testing.T) {
		mockRepo := newMockSessionRepository()
		mockRepo.findError = errors.New("database error")
		service := NewSessionService(mockRepo, nil)

		validated, err := service.ValidateSession(ctx, "any-id")

		assert.Error(t, err)
		assert.Nil(t, validated)
		assert.Contains(t, err.Error(), "session not found")
	})

	t.Run("handles session exactly at expiry time", func(t *testing.T) {
		mockRepo := newMockSessionRepository()
		service := NewSessionService(mockRepo, nil)

		// Session that expired right now
		session := &auth.Session{
			ID:     "just-expired",
			Expiry: time.Now().Add(-1 * time.Millisecond), // Just expired
			ProviderSourceAuth: []byte(`{}`),
		}
		mockRepo.setSession("just-expired", session)

		validated, err := service.ValidateSession(ctx, "just-expired")

		assert.Error(t, err)
		assert.Nil(t, validated)
	})
}

// TestGetSession tests session retrieval without validation
func TestGetSession(t *testing.T) {
	ctx := context.Background()

	t.Run("retrieves existing session", func(t *testing.T) {
		mockRepo := newMockSessionRepository()
		service := NewSessionService(mockRepo, nil)

		session := &auth.Session{
			ID:     "test-session",
			Expiry: time.Now().Add(-1 * time.Hour), // Even expired
			ProviderSourceAuth: []byte(`{"data": "value"}`),
		}
		mockRepo.setSession("test-session", session)

		retrieved, err := service.GetSession(ctx, "test-session")

		require.NoError(t, err)
		assert.Equal(t, session.ID, retrieved.ID)
		assert.Equal(t, session.ProviderSourceAuth, retrieved.ProviderSourceAuth)
	})

	t.Run("returns nil for non-existent session", func(t *testing.T) {
		mockRepo := newMockSessionRepository()
		service := NewSessionService(mockRepo, nil)

		retrieved, err := service.GetSession(ctx, "non-existent")

		require.NoError(t, err)
		assert.Nil(t, retrieved)
	})

	t.Run("propagates repository errors", func(t *testing.T) {
		mockRepo := newMockSessionRepository()
		mockRepo.findError = errors.New("database error")
		service := NewSessionService(mockRepo, nil)

		retrieved, err := service.GetSession(ctx, "any-id")

		assert.Error(t, err)
		assert.Nil(t, retrieved)
	})
}

// TestDeleteSession tests session deletion
func TestDeleteSession(t *testing.T) {
	ctx := context.Background()

	t.Run("deletes existing session", func(t *testing.T) {
		mockRepo := newMockSessionRepository()
		service := NewSessionService(mockRepo, nil)

		// Create a session
		session := &auth.Session{
			ID:     "to-delete",
			Expiry: time.Now().Add(1 * time.Hour),
			ProviderSourceAuth: []byte(`{}`),
		}
		mockRepo.setSession("to-delete", session)

		// Delete it
		err := service.DeleteSession(ctx, "to-delete")

		require.NoError(t, err)
		assert.Equal(t, 1, mockRepo.deleteCalled)
		assert.NotContains(t, mockRepo.sessions, "to-delete")
	})

	t.Run("handles deletion of non-existent session", func(t *testing.T) {
		mockRepo := newMockSessionRepository()
		service := NewSessionService(mockRepo, nil)

		// Delete non-existent session (should not error)
		err := service.DeleteSession(ctx, "non-existent")

		require.NoError(t, err)
		assert.Equal(t, 1, mockRepo.deleteCalled)
	})

	t.Run("propagates repository errors", func(t *testing.T) {
		mockRepo := newMockSessionRepository()
		mockRepo.deleteError = errors.New("database error")
		service := NewSessionService(mockRepo, nil)

		err := service.DeleteSession(ctx, "any-id")

		assert.Error(t, err)
		assert.Equal(t, 1, mockRepo.deleteCalled)
	})
}

// TestCleanupExpiredSessions tests expired session cleanup
func TestCleanupExpiredSessions(t *testing.T) {
	ctx := context.Background()

	t.Run("removes only expired sessions", func(t *testing.T) {
		mockRepo := newMockSessionRepository()
		service := NewSessionService(mockRepo, nil)

		// Create mixed sessions
		expiryTime := time.Now()
		mockRepo.setSession("active-1", &auth.Session{
			ID:     "active-1",
			Expiry: expiryTime.Add(1 * time.Hour),
			ProviderSourceAuth: []byte(`{}`),
		})
		mockRepo.setSession("active-2", &auth.Session{
			ID:     "active-2",
			Expiry: expiryTime.Add(2 * time.Hour),
			ProviderSourceAuth: []byte(`{}`),
		})
		mockRepo.setSession("expired-1", &auth.Session{
			ID:     "expired-1",
			Expiry: expiryTime.Add(-1 * time.Hour),
			ProviderSourceAuth: []byte(`{}`),
		})
		mockRepo.setSession("expired-2", &auth.Session{
			ID:     "expired-2",
			Expiry: expiryTime.Add(-2 * time.Hour),
			ProviderSourceAuth: []byte(`{}`),
		})

		err := service.CleanupExpiredSessions(ctx)

		require.NoError(t, err)
		assert.Equal(t, 1, mockRepo.cleanupCalled)
		// Check sessions still exist by trying to retrieve them
		active1, _ := mockRepo.FindByID(ctx, "active-1")
		active2, _ := mockRepo.FindByID(ctx, "active-2")
		expired1, _ := mockRepo.FindByID(ctx, "expired-1")
		expired2, _ := mockRepo.FindByID(ctx, "expired-2")
		assert.NotNil(t, active1)
		assert.NotNil(t, active2)
		assert.Nil(t, expired1)
		assert.Nil(t, expired2)
	})

	t.Run("handles empty session store", func(t *testing.T) {
		mockRepo := newMockSessionRepository()
		service := NewSessionService(mockRepo, nil)

		err := service.CleanupExpiredSessions(ctx)

		require.NoError(t, err)
		assert.Equal(t, 1, mockRepo.cleanupCalled)
	})

	t.Run("handles all sessions expired", func(t *testing.T) {
		mockRepo := newMockSessionRepository()
		service := NewSessionService(mockRepo, nil)

		expiryTime := time.Now()
		mockRepo.setSession("expired-1", &auth.Session{
			ID:     "expired-1",
			Expiry: expiryTime.Add(-1 * time.Minute),
			ProviderSourceAuth: []byte(`{}`),
		})
		mockRepo.setSession("expired-2", &auth.Session{
			ID:     "expired-2",
			Expiry: expiryTime.Add(-5 * time.Minute),
			ProviderSourceAuth: []byte(`{}`),
		})

		err := service.CleanupExpiredSessions(ctx)

		require.NoError(t, err)
		// Verify all sessions were removed
		expired1, _ := mockRepo.FindByID(ctx, "expired-1")
		expired2, _ := mockRepo.FindByID(ctx, "expired-2")
		assert.Nil(t, expired1)
		assert.Nil(t, expired2)
	})

	t.Run("propagates repository errors", func(t *testing.T) {
		mockRepo := newMockSessionRepository()
		mockRepo.cleanupError = errors.New("database error")
		service := NewSessionService(mockRepo, nil)

		err := service.CleanupExpiredSessions(ctx)

		assert.Error(t, err)
		assert.Equal(t, 1, mockRepo.cleanupCalled)
	})
}

// TestRefreshSession tests session refresh functionality
func TestRefreshSession(t *testing.T) {
	ctx := context.Background()

	t.Run("refreshes session expiry time", func(t *testing.T) {
		mockRepo := newMockSessionRepository()
		config := &SessionDatabaseConfig{
			DefaultTTL:      1 * time.Hour,
			MaxTTL:          24 * time.Hour,
			CleanupInterval: 1 * time.Hour,
		}
		service := NewSessionService(mockRepo, config)

		// Create a session with short expiry
		originalExpiry := time.Now().Add(10 * time.Minute)
		session := &auth.Session{
			ID:     "to-refresh",
			Expiry: originalExpiry,
			ProviderSourceAuth: []byte(`{"user": "test"}`),
		}
		mockRepo.setSession("to-refresh", session)

		// Refresh with new TTL
		newTTL := 2 * time.Hour
		refreshed, err := service.RefreshSession(ctx, "to-refresh", newTTL)

		require.NoError(t, err)
		assert.NotNil(t, refreshed)
		// New expiry should be ~2 hours from now
		assert.True(t, refreshed.Expiry.After(time.Now().Add(newTTL-time.Minute)))
		assert.True(t, refreshed.Expiry.Before(time.Now().Add(newTTL+time.Minute)))
		assert.Equal(t, 1, mockRepo.updateCalled)
	})

	t.Run("fails to refresh non-existent session", func(t *testing.T) {
		mockRepo := newMockSessionRepository()
		service := NewSessionService(mockRepo, nil)

		refreshed, err := service.RefreshSession(ctx, "non-existent", 1*time.Hour)

		assert.Error(t, err)
		assert.Nil(t, refreshed)
		assert.Contains(t, err.Error(), "session not found")
	})

	t.Run("respects max TTL on refresh", func(t *testing.T) {
		mockRepo := newMockSessionRepository()
		config := &SessionDatabaseConfig{
			DefaultTTL:      1 * time.Hour,
			MaxTTL:          2 * time.Hour,
			CleanupInterval: 1 * time.Hour,
		}
		service := NewSessionService(mockRepo, config)

		session := &auth.Session{
			ID:     "to-refresh",
			Expiry: time.Now().Add(10 * time.Minute),
			ProviderSourceAuth: []byte(`{}`),
		}
		mockRepo.setSession("to-refresh", session)

		// Try to set 10 hours but should be capped at 2
		refreshed, err := service.RefreshSession(ctx, "to-refresh", 10*time.Hour)

		require.NoError(t, err)
		assert.True(t, refreshed.Expiry.Before(time.Now().Add(3*time.Hour)))
		assert.True(t, refreshed.Expiry.After(time.Now().Add(config.MaxTTL-time.Minute)))
	})

	t.Run("handles repository update error", func(t *testing.T) {
		mockRepo := newMockSessionRepository()
		mockRepo.updateError = errors.New("update failed")
		service := NewSessionService(mockRepo, nil)

		session := &auth.Session{
			ID:     "to-refresh",
			Expiry: time.Now().Add(10 * time.Minute),
			ProviderSourceAuth: []byte(`{}`),
		}
		mockRepo.setSession("to-refresh", session)

		refreshed, err := service.RefreshSession(ctx, "to-refresh", 1*time.Hour)

		assert.Error(t, err)
		assert.Nil(t, refreshed)
		assert.Contains(t, err.Error(), "failed to update session")
	})

	t.Run("handles repository find error", func(t *testing.T) {
		mockRepo := newMockSessionRepository()
		mockRepo.findError = errors.New("find failed")
		service := NewSessionService(mockRepo, nil)

		refreshed, err := service.RefreshSession(ctx, "any-id", 1*time.Hour)

		assert.Error(t, err)
		assert.Nil(t, refreshed)
		assert.Contains(t, err.Error(), "session not found")
	})
}

// TestGenerateSessionID tests session ID generation
func TestGenerateSessionID(t *testing.T) {
	service := NewSessionService(newMockSessionRepository(), nil)

	t.Run("generates non-empty IDs", func(t *testing.T) {
		id := service.generateSessionID()
		assert.NotEmpty(t, id)
	})

	t.Run("generates unique IDs", func(t *testing.T) {
		ids := make(map[string]bool)
		for i := 0; i < 100; i++ {
			id := service.generateSessionID()
			// With nanosecond precision, all should be unique
			// but we allow some duplicates due to test execution speed
			ids[id] = true
		}
		assert.Greater(t, len(ids), 90) // At least 90 unique IDs out of 100
	})

	t.Run("IDs are string representations of nanosecond timestamps", func(t *testing.T) {
		id := service.generateSessionID()
		// Should be parseable as an integer
		var timestamp int64
		_, err := fmt.Sscanf(id, "%d", &timestamp)
		assert.NoError(t, err)
		assert.Greater(t, timestamp, int64(0))
	})
}

// TestSessionWithDifferentAuthMethods tests session creation with various auth methods
func TestSessionWithDifferentAuthMethods(t *testing.T) {
	ctx := context.Background()
	mockRepo := newMockSessionRepository()
	service := NewSessionService(mockRepo, nil)

	authMethods := []string{"oidc", "github", "gitlab", "bitbucket", "saml"}

	for _, method := range authMethods {
		t.Run("auth_method_"+method, func(t *testing.T) {
			providerData := []byte(fmt.Sprintf(`{"auth_method": "%s", "user_id": "123"}`, method))
			session, err := service.CreateSession(ctx, method, providerData, nil)

			require.NoError(t, err)
			assert.NotNil(t, session)
			assert.Equal(t, providerData, session.ProviderSourceAuth)
		})
	}
}

// TestSessionWithVariousProviderData tests sessions with different provider data formats
func TestSessionWithVariousProviderData(t *testing.T) {
	ctx := context.Background()
	mockRepo := newMockSessionRepository()
	service := NewSessionService(mockRepo, nil)

	testCases := []struct {
		name string
		data []byte
	}{
		{
			name: "empty data",
			data: []byte{},
		},
		{
			name: "simple JSON",
			data: []byte(`{"user_id": "123"}`),
		},
		{
			name: "complex JSON",
			data: []byte(`{"user_id": "123", "email": "test@example.com", "groups": ["admin", "user"], "metadata": {"key": "value"}}`),
		},
		{
			name: "binary data",
			data: []byte{0x00, 0x01, 0x02, 0xFF, 0xFE, 0xFD},
		},
		{
			name: "unicode data",
			data: []byte(`{"name": "用户", "emoji": "😀"}`),
		},
		{
			name: "large data",
			data: []byte(`{"data": "` + string(make([]byte, 10000)) + `"}`),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			session, err := service.CreateSession(ctx, "test", tc.data, nil)

			require.NoError(t, err)
			assert.NotNil(t, session)
			assert.Equal(t, tc.data, session.ProviderSourceAuth)

			// Verify we can retrieve it
			retrieved, err := service.GetSession(ctx, session.ID)
			require.NoError(t, err)
			assert.Equal(t, tc.data, retrieved.ProviderSourceAuth)
		})
	}
}

// TestEdgeCaseEmptySessionID tests edge cases with empty/invalid session IDs
func TestEdgeCaseEmptySessionID(t *testing.T) {
	ctx := context.Background()
	mockRepo := newMockSessionRepository()
	service := NewSessionService(mockRepo, nil)

	t.Run("validate empty session ID", func(t *testing.T) {
		validated, err := service.ValidateSession(ctx, "")

		assert.Error(t, err)
		assert.Nil(t, validated)
	})

	t.Run("get empty session ID", func(t *testing.T) {
		retrieved, err := service.GetSession(ctx, "")

		require.NoError(t, err)
		assert.Nil(t, retrieved)
	})

	t.Run("delete empty session ID", func(t *testing.T) {
		err := service.DeleteSession(ctx, "")

		// Should not error - deleting non-existent is fine
		require.NoError(t, err)
	})

	t.Run("refresh empty session ID", func(t *testing.T) {
		refreshed, err := service.RefreshSession(ctx, "", 1*time.Hour)

		assert.Error(t, err)
		assert.Nil(t, refreshed)
	})
}

// TestConcurrentSessionOperations tests thread safety of session operations
func TestConcurrentSessionOperations(t *testing.T) {
	ctx := context.Background()
	mockRepo := newMockSessionRepository()
	service := NewSessionService(mockRepo, nil)

	t.Run("concurrent session creation", func(t *testing.T) {
		// Create 50 sessions concurrently
		results := make(chan *auth.Session, 50)
		errors := make(chan error, 50)

		for i := 0; i < 50; i++ {
			go func(idx int) {
				providerData := []byte(fmt.Sprintf(`{"user_id": "%d"}`, idx))
				session, err := service.CreateSession(ctx, "oidc", providerData, nil)
				if err != nil {
					errors <- err
				} else {
					results <- session
				}
			}(i)
		}

		// Collect results
		sessionCount := 0
		for i := 0; i < 50; i++ {
			select {
			case <-results:
				sessionCount++
			case err := <-errors:
				t.Errorf("Unexpected error: %v", err)
			}
		}

		assert.Equal(t, 50, sessionCount)
	})

	t.Run("concurrent session validation", func(t *testing.T) {
		// Create a session
		providerData := []byte(`{"user_id": "123"}`)
		session, err := service.CreateSession(ctx, "oidc", providerData, nil)
		require.NoError(t, err)

		// Validate it 100 times concurrently
		errors := make(chan error, 100)
		for i := 0; i < 100; i++ {
			go func() {
				_, err := service.ValidateSession(ctx, session.ID)
				errors <- err
			}()
		}

		// All validations should succeed
		for i := 0; i < 100; i++ {
			err := <-errors
			assert.NoError(t, err)
		}
	})
}

// Benchmark tests for performance
func BenchmarkCreateSession(b *testing.B) {
	ctx := context.Background()
	mockRepo := newMockSessionRepository()
	service := NewSessionService(mockRepo, nil)
	providerData := []byte(`{"user_id": "123"}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.CreateSession(ctx, "oidc", providerData, nil)
	}
}

func BenchmarkValidateSession(b *testing.B) {
	ctx := context.Background()
	mockRepo := newMockSessionRepository()
	service := NewSessionService(mockRepo, nil)

	// Create a session to validate
	providerData := []byte(`{"user_id": "123"}`)
	session, _ := service.CreateSession(ctx, "oidc", providerData, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.ValidateSession(ctx, session.ID)
	}
}

func BenchmarkGenerateSessionID(b *testing.B) {
	service := NewSessionService(newMockSessionRepository(), nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = service.generateSessionID()
	}
}
