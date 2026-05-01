package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth/repository"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockSessionRepositoryForManagement is a mock SessionRepository for testing session management
type mockSessionRepositoryForManagement struct {
	repository.SessionRepository
	sessions      map[string]*auth.Session
	createError   error
	validateError error
	deleteError   error
	refreshError  error
	mu            mockMutex
}

func newMockSessionRepositoryForManagement() *mockSessionRepositoryForManagement {
	return &mockSessionRepositoryForManagement{
		sessions: make(map[string]*auth.Session),
	}
}

func (m *mockSessionRepositoryForManagement) Create(ctx context.Context, session *auth.Session) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.createError != nil {
		return m.createError
	}
	m.sessions[session.ID] = session
	return nil
}

func (m *mockSessionRepositoryForManagement) FindByID(ctx context.Context, sessionID string) (*auth.Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.validateError != nil {
		return nil, m.validateError
	}
	session, exists := m.sessions[sessionID]
	if !exists {
		return nil, nil
	}
	// Check expiry
	if session.IsExpired() {
		return nil, nil
	}
	return session, nil
}

func (m *mockSessionRepositoryForManagement) Delete(ctx context.Context, sessionID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.deleteError != nil {
		return m.deleteError
	}
	delete(m.sessions, sessionID)
	return nil
}

func (m *mockSessionRepositoryForManagement) CleanupExpired(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for id, session := range m.sessions {
		if session.IsExpired() {
			delete(m.sessions, id)
		}
	}
	return nil
}

func (m *mockSessionRepositoryForManagement) UpdateProviderSourceAuth(ctx context.Context, sessionID string, data []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.refreshError != nil {
		return m.refreshError
	}
	if session, exists := m.sessions[sessionID]; exists {
		session.Expiry = time.Now().Add(1 * time.Hour)
	}
	return nil
}

func (m *mockSessionRepositoryForManagement) reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sessions = make(map[string]*auth.Session)
	m.createError = nil
	m.validateError = nil
	m.deleteError = nil
	m.refreshError = nil
}

// newTestInfraConfig creates a test infrastructure config for cookie service
func newTestInfraConfig() *config.InfrastructureConfig {
	return &config.InfrastructureConfig{
		SecretKey:         "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
		SessionCookieName: "terrareg_session",
	}
}

// mockMutex provides simple locking for tests
type mockMutex struct{}

func (m *mockMutex) Lock()   {}
func (m *mockMutex) Unlock() {}

// TestNewSessionManagementService tests the constructor
func TestNewSessionManagementService(t *testing.T) {
	t.Run("returns nil when cookie service is nil", func(t *testing.T) {
		sessionService := NewSessionService(nil, nil)
		service := NewSessionManagementService(sessionService, nil)

		assert.Nil(t, service, "Service should be nil when cookieService is nil")
	})

	t.Run("creates service when cookie service is provided", func(t *testing.T) {
		sessionService := NewSessionService(nil, nil)
		cookieService := newTestCookieService(t, newTestInfraConfig())

		service := NewSessionManagementService(sessionService, cookieService)

		assert.NotNil(t, service)
		assert.Equal(t, sessionService, service.sessionService)
		assert.Equal(t, cookieService, service.cookieService)
		assert.Nil(t, service.auditLogger)
	})
}

// TestSessionManagementService_SetAuditLogger tests setting the audit logger
func TestSessionManagementService_SetAuditLogger(t *testing.T) {
	sessionService := NewSessionService(nil, nil)
	cookieService := newTestCookieService(t, newTestInfraConfig())
	service := NewSessionManagementService(sessionService, cookieService)

	auditLogger := &AuditLogger{}
	service.SetAuditLogger(auditLogger)

	assert.Equal(t, auditLogger, service.auditLogger)
}

// TestSessionManagementService_IsAvailable tests availability check
func TestSessionManagementService_IsAvailable(t *testing.T) {
	sessionService := NewSessionService(nil, nil)
	cookieService := newTestCookieService(t, newTestInfraConfig())
	service := NewSessionManagementService(sessionService, cookieService)

	assert.True(t, service.IsAvailable())
}

// TestSessionManagementService_CreateSessionAndCookie tests session and cookie creation
func TestSessionManagementService_CreateSessionAndCookie(t *testing.T) {
	t.Run("successfully creates session and cookie", func(t *testing.T) {
		mockRepo := newMockSessionRepositoryForManagement()
		sessionService := NewSessionService(mockRepo, nil)
		cookieService := newTestCookieService(t, newTestInfraConfig())
		service := NewSessionManagementService(sessionService, cookieService)

		ctx := context.Background()
		w := httptest.NewRecorder()

		err := service.CreateSessionAndCookie(ctx, w, auth.AuthMethodGitHub, "testuser", false, []string{"group1"}, nil, nil, nil)

		require.NoError(t, err)

		// Check cookie was set
		cookies := w.Result().Cookies()
		assert.NotEmpty(t, cookies, "Cookie should be set")
		assert.Equal(t, "terrareg_session", cookies[0].Name)
		assert.True(t, cookies[0].Secure)
		assert.True(t, cookies[0].HttpOnly)

		// Verify session was created in repository
		assert.Len(t, mockRepo.sessions, 1)
	})

	t.Run("creates session with custom TTL", func(t *testing.T) {
		mockRepo := newMockSessionRepositoryForManagement()
		sessionService := NewSessionService(mockRepo, nil)
		cookieService := newTestCookieService(t, newTestInfraConfig())
		service := NewSessionManagementService(sessionService, cookieService)

		customTTL := 2 * time.Hour

		ctx := context.Background()
		w := httptest.NewRecorder()

		err := service.CreateSessionAndCookie(ctx, w, auth.AuthMethodGitHub, "testuser", false, nil, nil, nil, &customTTL)

		require.NoError(t, err)

		// Check cookie MaxAge matches custom TTL
		cookies := w.Result().Cookies()
		assert.NotEmpty(t, cookies)
		assert.Equal(t, int(customTTL.Seconds()), cookies[0].MaxAge)
	})

	t.Run("handles session creation failure", func(t *testing.T) {
		mockRepo := newMockSessionRepositoryForManagement()
		mockRepo.createError = assert.AnError
		sessionService := NewSessionService(mockRepo, nil)
		cookieService := newTestCookieService(t, newTestInfraConfig())
		service := NewSessionManagementService(sessionService, cookieService)

		ctx := context.Background()
		w := httptest.NewRecorder()

		err := service.CreateSessionAndCookie(ctx, w, auth.AuthMethodGitHub, "testuser", false, nil, nil, nil, nil)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create session")
	})
}

// TestSessionManagementService_ValidateSessionCookie tests session cookie validation
func TestSessionManagementService_ValidateSessionCookie(t *testing.T) {
	t.Run("successfully validates session cookie", func(t *testing.T) {
		mockRepo := newMockSessionRepositoryForManagement()
		sessionService := NewSessionService(mockRepo, nil)
		cookieService := newTestCookieService(t, newTestInfraConfig())
		service := NewSessionManagementService(sessionService, cookieService)

		// Create a valid session first
		session := &auth.Session{
			ID:                 "test-session-123",
			Expiry:             time.Now().Add(1 * time.Hour),
			ProviderSourceAuth: []byte(`{}`),
		}
		mockRepo.sessions[session.ID] = session

		// Create session data and encrypt it
		sessionData := &auth.SessionData{
			SessionID: session.ID,
		}
		encryptedCookie, _ := cookieService.EncryptSession(sessionData)

		ctx := context.Background()

		validatedSession, err := service.ValidateSessionCookie(ctx, encryptedCookie)

		require.NoError(t, err)
		assert.NotNil(t, validatedSession)
		assert.Equal(t, "test-session-123", validatedSession.ID)
	})

	t.Run("handles invalid cookie", func(t *testing.T) {
		mockRepo := newMockSessionRepositoryForManagement()
		sessionService := NewSessionService(mockRepo, nil)
		cookieService := newTestCookieService(t, newTestInfraConfig())
		service := NewSessionManagementService(sessionService, cookieService)

		ctx := context.Background()

		session, err := service.ValidateSessionCookie(ctx, "invalid-cookie-value")

		assert.Error(t, err)
		assert.Nil(t, session)
		assert.Contains(t, err.Error(), "invalid session cookie")
	})

	t.Run("handles expired session", func(t *testing.T) {
		mockRepo := newMockSessionRepositoryForManagement()
		sessionService := NewSessionService(mockRepo, nil)
		cookieService := newTestCookieService(t, newTestInfraConfig())
		service := NewSessionManagementService(sessionService, cookieService)

		// Create an expired session
		session := &auth.Session{
			ID:                 "expired-session",
			Expiry:             time.Now().Add(-1 * time.Hour),
			ProviderSourceAuth: []byte(`{}`),
		}
		mockRepo.sessions[session.ID] = session

		// Create session data and encrypt it
		sessionData := &auth.SessionData{
			SessionID: session.ID,
		}
		encryptedCookie, _ := cookieService.EncryptSession(sessionData)

		ctx := context.Background()

		validatedSession, err := service.ValidateSessionCookie(ctx, encryptedCookie)

		assert.Error(t, err)
		assert.Nil(t, validatedSession)
		assert.Contains(t, err.Error(), "session validation failed")
	})
}

// TestSessionManagementService_ClearSessionAndCookie tests clearing session and cookie
func TestSessionManagementService_ClearSessionAndCookie(t *testing.T) {
	t.Run("successfully clears session and cookie", func(t *testing.T) {
		mockRepo := newMockSessionRepositoryForManagement()
		sessionService := NewSessionService(mockRepo, nil)
		cookieService := newTestCookieService(t, newTestInfraConfig())
		service := NewSessionManagementService(sessionService, cookieService)

		// Create a valid session
		session := &auth.Session{
			ID:                 "test-session-123",
			Expiry:             time.Now().Add(1 * time.Hour),
			ProviderSourceAuth: []byte(`{}`),
		}
		mockRepo.sessions[session.ID] = session

		// Create encrypted cookie
		sessionData := &auth.SessionData{
			SessionID: session.ID,
		}
		encryptedCookie, _ := cookieService.EncryptSession(sessionData)

		ctx := context.Background()
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)
		r.AddCookie(&http.Cookie{Name: "terrareg_session", Value: encryptedCookie})

		err := service.ClearSessionAndCookie(ctx, w, r)

		require.NoError(t, err)

		// Check cookie was cleared
		cookies := w.Result().Cookies()
		assert.NotEmpty(t, cookies)
		// Cleared cookie should have MaxAge < 0
		assert.Less(t, cookies[0].MaxAge, 0)

		// Verify session was deleted from repository
		_, exists := mockRepo.sessions[session.ID]
		assert.False(t, exists)
	})

	t.Run("handles missing cookie gracefully", func(t *testing.T) {
		mockRepo := newMockSessionRepositoryForManagement()
		sessionService := NewSessionService(mockRepo, nil)
		cookieService := newTestCookieService(t, newTestInfraConfig())
		service := NewSessionManagementService(sessionService, cookieService)

		ctx := context.Background()
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)
		// No cookie added

		err := service.ClearSessionAndCookie(ctx, w, r)

		require.NoError(t, err)

		// Cookie should still be cleared
		cookies := w.Result().Cookies()
		assert.NotEmpty(t, cookies)
		assert.Less(t, cookies[0].MaxAge, 0)
	})
}

// TestSessionManagementService_RefreshSessionAndCookie tests refreshing session and cookie
func TestSessionManagementService_RefreshSessionAndCookie(t *testing.T) {
	t.Run("successfully refreshes session and cookie", func(t *testing.T) {
		mockRepo := newMockSessionRepositoryForManagement()
		sessionService := NewSessionService(mockRepo, nil)
		cookieService := newTestCookieService(t, newTestInfraConfig())
		service := NewSessionManagementService(sessionService, cookieService)

		// Create a valid session
		session := &auth.Session{
			ID:                 "test-session-123",
			Expiry:             time.Now().Add(30 * time.Minute), // Will expire soon
			ProviderSourceAuth: []byte(`{}`),
		}
		mockRepo.sessions[session.ID] = session

		// Create encrypted cookie
		sessionData := &auth.SessionData{
			SessionID: session.ID,
		}
		encryptedCookie, _ := cookieService.EncryptSession(sessionData)

		newTTL := 2 * time.Hour

		ctx := context.Background()
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)
		r.AddCookie(&http.Cookie{Name: "terrareg_session", Value: encryptedCookie})

		err := service.RefreshSessionAndCookie(ctx, w, r, newTTL)

		require.NoError(t, err)

		// Check cookie was updated
		cookies := w.Result().Cookies()
		assert.NotEmpty(t, cookies)
		assert.Equal(t, int(newTTL.Seconds()), cookies[0].MaxAge)
	})

	t.Run("handles missing cookie", func(t *testing.T) {
		mockRepo := newMockSessionRepositoryForManagement()
		sessionService := NewSessionService(mockRepo, nil)
		cookieService := newTestCookieService(t, newTestInfraConfig())
		service := NewSessionManagementService(sessionService, cookieService)

		ctx := context.Background()
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)
		// No cookie added

		err := service.RefreshSessionAndCookie(ctx, w, r, 1*time.Hour)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no session cookie found")
	})

	t.Run("handles invalid cookie", func(t *testing.T) {
		mockRepo := newMockSessionRepositoryForManagement()
		sessionService := NewSessionService(mockRepo, nil)
		cookieService := newTestCookieService(t, newTestInfraConfig())
		service := NewSessionManagementService(sessionService, cookieService)

		ctx := context.Background()
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)
		r.AddCookie(&http.Cookie{Name: "terrareg_session", Value: "invalid-cookie"})

		err := service.RefreshSessionAndCookie(ctx, w, r, 1*time.Hour)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid session cookie")
	})
}

// TestSessionManagementService_GetSessionFromCookie tests getting session from cookie
func TestSessionManagementService_GetSessionFromCookie(t *testing.T) {
	t.Run("successfully extracts and validates session", func(t *testing.T) {
		mockRepo := newMockSessionRepositoryForManagement()
		sessionService := NewSessionService(mockRepo, nil)
		cookieService := newTestCookieService(t, newTestInfraConfig())
		service := NewSessionManagementService(sessionService, cookieService)

		// Create a valid session
		session := &auth.Session{
			ID:                 "test-session-123",
			Expiry:             time.Now().Add(1 * time.Hour),
			ProviderSourceAuth: []byte(`{}`),
		}
		mockRepo.sessions[session.ID] = session

		// Create encrypted cookie
		sessionData := &auth.SessionData{
			SessionID: session.ID,
			Username:  "testuser",
		}
		encryptedCookie, _ := cookieService.EncryptSession(sessionData)

		ctx := context.Background()
		r := httptest.NewRequest("GET", "/", nil)
		r.AddCookie(&http.Cookie{Name: "terrareg_session", Value: encryptedCookie})

		retrievedSession, retrievedData, err := service.GetSessionFromCookie(ctx, r)

		require.NoError(t, err)
		assert.Equal(t, "test-session-123", retrievedSession.ID)
		assert.Equal(t, "testuser", retrievedData.Username)
	})

	t.Run("handles missing cookie", func(t *testing.T) {
		mockRepo := newMockSessionRepositoryForManagement()
		sessionService := NewSessionService(mockRepo, nil)
		cookieService := newTestCookieService(t, newTestInfraConfig())
		service := NewSessionManagementService(sessionService, cookieService)

		ctx := context.Background()
		r := httptest.NewRequest("GET", "/", nil)
		// No cookie added

		session, sessionData, err := service.GetSessionFromCookie(ctx, r)

		assert.Error(t, err)
		assert.Nil(t, session)
		assert.Nil(t, sessionData)
		assert.Contains(t, err.Error(), "no session cookie found")
	})
}

// TestSessionManagementService_SetCookieForExistingSession tests setting cookie for existing session
func TestSessionManagementService_SetCookieForExistingSession(t *testing.T) {
	t.Run("successfully sets cookie for existing session", func(t *testing.T) {
		mockRepo := newMockSessionRepositoryForManagement()
		sessionService := NewSessionService(mockRepo, nil)
		cookieService := newTestCookieService(t, newTestInfraConfig())
		service := NewSessionManagementService(sessionService, cookieService)

		// Create a valid session
		session := &auth.Session{
			ID:                 "existing-session-123",
			Expiry:             time.Now().Add(1 * time.Hour),
			ProviderSourceAuth: []byte(`{}`),
		}
		mockRepo.sessions[session.ID] = session

		ctx := context.Background()
		w := httptest.NewRecorder()

		err := service.SetCookieForExistingSession(ctx, w, "existing-session-123", "testuser", "github")

		require.NoError(t, err)

		// Check cookie was set
		cookies := w.Result().Cookies()
		assert.NotEmpty(t, cookies)
		assert.Equal(t, "terrareg_session", cookies[0].Name)
		assert.True(t, cookies[0].Secure)
		assert.True(t, cookies[0].HttpOnly)
	})

	t.Run("handles non-existent session", func(t *testing.T) {
		mockRepo := newMockSessionRepositoryForManagement()
		sessionService := NewSessionService(mockRepo, nil)
		cookieService := newTestCookieService(t, newTestInfraConfig())
		service := NewSessionManagementService(sessionService, cookieService)

		ctx := context.Background()
		w := httptest.NewRecorder()

		err := service.SetCookieForExistingSession(ctx, w, "non-existent-session", "testuser", "github")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get session")
	})
}

// TestSessionManagementService_EdgeCases tests edge cases
func TestSessionManagementService_EdgeCases(t *testing.T) {
	t.Run("handles empty session ID in ClearSessionAndCookie", func(t *testing.T) {
		mockRepo := newMockSessionRepositoryForManagement()
		sessionService := NewSessionService(mockRepo, nil)
		cookieService := newTestCookieService(t, newTestInfraConfig())
		service := NewSessionManagementService(sessionService, cookieService)

		// Create encrypted cookie with empty session ID
		sessionData := &auth.SessionData{
			SessionID: "",
		}
		encryptedCookie, _ := cookieService.EncryptSession(sessionData)

		ctx := context.Background()
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)
		r.AddCookie(&http.Cookie{Name: "terrareg_session", Value: encryptedCookie})

		// Should not error, just skip deletion
		err := service.ClearSessionAndCookie(ctx, w, r)

		require.NoError(t, err)
	})

	t.Run("handles nil provider data in CreateSessionAndCookie", func(t *testing.T) {
		mockRepo := newMockSessionRepositoryForManagement()
		sessionService := NewSessionService(mockRepo, nil)
		cookieService := newTestCookieService(t, newTestInfraConfig())
		service := NewSessionManagementService(sessionService, cookieService)

		ctx := context.Background()
		w := httptest.NewRecorder()

		err := service.CreateSessionAndCookie(ctx, w, auth.AuthMethodGitHub, "testuser", false, nil, nil, nil, nil)

		require.NoError(t, err)
	})
}
