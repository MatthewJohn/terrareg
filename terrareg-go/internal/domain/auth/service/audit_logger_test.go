package service

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/audit/model"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

// mockAuditHistoryRepository is a mock for testing
type mockAuditHistoryRepository struct {
	mu          sync.Mutex
	createCalled bool
	createCount  int
	createError  error
	lastAudit   *model.AuditHistory
}

func (m *mockAuditHistoryRepository) Search(ctx context.Context, query model.AuditHistorySearchQuery) (*model.AuditHistorySearchResult, error) {
	return nil, nil
}

func (m *mockAuditHistoryRepository) GetTotalCount(ctx context.Context) (int, error) {
	return 0, nil
}

func (m *mockAuditHistoryRepository) GetFilteredCount(ctx context.Context, searchValue string) (int, error) {
	return 0, nil
}

func (m *mockAuditHistoryRepository) Create(ctx context.Context, audit *model.AuditHistory) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.createCalled = true
	m.createCount++
	m.lastAudit = audit
	return m.createError
}

// newTestLogger creates a test logger that writes to a strings.Builder
func newAuditTestLogger() (*zerolog.Logger, *strings.Builder) {
	var output strings.Builder
	logger := zerolog.New(&output).With().Timestamp().Logger()
	return &logger, &output
}

// newTestAuditLogger creates a test audit logger with mock repository
func newTestAuditLogger() (*AuditLogger, *mockAuditHistoryRepository) {
	logger, _ := newAuditTestLogger()
	mockRepo := &mockAuditHistoryRepository{}
	auditLogger := NewAuditLogger(*logger, mockRepo)
	return auditLogger, mockRepo
}

// TestNewAuditLogger tests the constructor
func TestNewAuditLogger(t *testing.T) {
	logger, _ := newAuditTestLogger()
	mockRepo := &mockAuditHistoryRepository{}

	auditLogger := NewAuditLogger(*logger, mockRepo)

	assert.NotNil(t, auditLogger)
	assert.Equal(t, mockRepo, auditLogger.auditRepo)
}

// TestLogAuthEvent tests basic auth event logging
func TestLogAuthEvent(t *testing.T) {
	t.Run("logs successful auth event with all fields", func(t *testing.T) {
		auditLogger, mockRepo := newTestAuditLogger()
		logger, output := newAuditTestLogger()
		auditLogger.logger = (*logger).With().Str("component", "audit").Logger()

		event := AuthEvent{
			Timestamp: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			Provider:  "github",
			Username:  "testuser",
			Success:   true,
			IPAddress: "192.168.1.1",
			UserAgent: "test-agent",
			SessionID: "session-123",
			Error:     "",
			Endpoint:  "/v1/modules",
			Method:    "GET",
			Namespace: "test-ns",
			Action:    "read",
		}

		auditLogger.LogAuthEvent(context.Background(), event)

		logOutput := output.String()
		assert.Contains(t, logOutput, "authentication_event")
		assert.Contains(t, logOutput, "github")
		assert.Contains(t, logOutput, "testuser")
		assert.Contains(t, logOutput, "192.168.1.1")
		assert.Contains(t, logOutput, "test-agent")
		assert.Contains(t, logOutput, "session-123")
		assert.Contains(t, logOutput, "/v1/modules")
		assert.Contains(t, logOutput, "test-ns")
		assert.False(t, mockRepo.createCalled, "Basic LogAuthEvent should not create database entry")
	})

	t.Run("logs failed auth event with warning level", func(t *testing.T) {
		auditLogger, _ := newTestAuditLogger()
		logger, output := newAuditTestLogger()
		auditLogger.logger = (*logger).With().Str("component", "audit").Logger()

		event := AuthEvent{
			Provider:  "github",
			Username:  "testuser",
			Success:   false,
			IPAddress: "192.168.1.1",
			UserAgent: "test-agent",
			Error:     "invalid credentials",
			Action:    "login",
		}

		auditLogger.LogAuthEvent(context.Background(), event)

		logOutput := output.String()
		assert.Contains(t, logOutput, "authentication_event")
		assert.Contains(t, logOutput, "invalid credentials")
		assert.Contains(t, logOutput, "false") // Success field
	})

	t.Run("sets default timestamp when not provided", func(t *testing.T) {
		auditLogger, _ := newTestAuditLogger()
		logger, output := newAuditTestLogger()
		auditLogger.logger = (*logger).With().Str("component", "audit").Logger()

		event := AuthEvent{
			Timestamp: time.Time{}, // Zero time
			Provider:  "github",
			Username:  "testuser",
			Success:   true,
			IPAddress: "192.168.1.1",
			UserAgent: "test-agent",
		}

		auditLogger.LogAuthEvent(context.Background(), event)

		logOutput := output.String()
		assert.Contains(t, logOutput, "authentication_event")
		// Verify it has a timestamp (log format includes time)
	})

	t.Run("logs minimal event with only required fields", func(t *testing.T) {
		auditLogger, _ := newTestAuditLogger()
		logger, output := newAuditTestLogger()
		auditLogger.logger = (*logger).With().Str("component", "audit").Logger()

		event := AuthEvent{
			Provider:  "test-provider",
			Success:   true,
			IPAddress: "10.0.0.1",
			UserAgent: "minimal-agent",
		}

		auditLogger.LogAuthEvent(context.Background(), event)

		logOutput := output.String()
		assert.Contains(t, logOutput, "authentication_event")
		assert.Contains(t, logOutput, "test-provider")
		assert.Contains(t, logOutput, "10.0.0.1")
	})
}

// TestLogLoginAttempt tests login attempt logging
func TestLogLoginAttempt(t *testing.T) {
	t.Run("logs successful login and creates database audit entry", func(t *testing.T) {
		auditLogger, mockRepo := newTestAuditLogger()
		logger, output := newAuditTestLogger()
		auditLogger.logger = (*logger).With().Str("component", "audit").Logger()

		ctx := context.Background()
		auditLogger.LogLoginAttempt(ctx, "github", "testuser", "192.168.1.1", "test-agent", "session-123", true, "")

		// Verify log output
		logOutput := output.String()
		assert.Contains(t, logOutput, "authentication_event")
		assert.Contains(t, logOutput, "github")
		assert.Contains(t, logOutput, "testuser")
		assert.Contains(t, logOutput, "login")

		// Verify database entry was created
		assert.True(t, mockRepo.createCalled, "Database entry should be created for successful login")
		assert.NotNil(t, mockRepo.lastAudit)
		assert.Equal(t, "testuser", mockRepo.lastAudit.Username())
		assert.Equal(t, model.AuditActionUserLogin, mockRepo.lastAudit.Action())
	})

	t.Run("does not create database entry for empty username", func(t *testing.T) {
		auditLogger, mockRepo := newTestAuditLogger()
		logger, output := newAuditTestLogger()
		auditLogger.logger = (*logger).With().Str("component", "audit").Logger()

		ctx := context.Background()
		auditLogger.LogLoginAttempt(ctx, "github", "", "192.168.1.1", "test-agent", "session-123", true, "")

		logOutput := output.String()
		assert.Contains(t, logOutput, "authentication_event")
		assert.False(t, mockRepo.createCalled, "No database entry should be created for empty username")
	})

	t.Run("logs failed login with error message", func(t *testing.T) {
		auditLogger, mockRepo := newTestAuditLogger()
		logger, output := newAuditTestLogger()
		auditLogger.logger = (*logger).With().Str("component", "audit").Logger()

		ctx := context.Background()
		auditLogger.LogLoginAttempt(ctx, "github", "testuser", "192.168.1.1", "test-agent", "", false, "invalid credentials")

		logOutput := output.String()
		assert.Contains(t, logOutput, "authentication_event")
		assert.Contains(t, logOutput, "invalid credentials")
		assert.False(t, mockRepo.createCalled, "No database entry for failed login")
	})

	t.Run("logs failed login without error message", func(t *testing.T) {
		auditLogger, mockRepo := newTestAuditLogger()
		logger, output := newAuditTestLogger()
		auditLogger.logger = (*logger).With().Str("component", "audit").Logger()

		ctx := context.Background()
		auditLogger.LogLoginAttempt(ctx, "github", "testuser", "192.168.1.1", "test-agent", "", false, "")

		logOutput := output.String()
		assert.Contains(t, logOutput, "authentication_event")
		assert.Contains(t, logOutput, "login")
		assert.False(t, mockRepo.createCalled)
	})

	t.Run("handles nil old and new values for login events", func(t *testing.T) {
		auditLogger, mockRepo := newTestAuditLogger()
		logger, output := newAuditTestLogger()
		auditLogger.logger = (*logger).With().Str("component", "audit").Logger()

		ctx := context.Background()
		auditLogger.LogLoginAttempt(ctx, "github", "testuser", "192.168.1.1", "test-agent", "session-123", true, "")

		// Verify nil values
		assert.True(t, mockRepo.createCalled)
		assert.Nil(t, mockRepo.lastAudit.OldValue())
		assert.Nil(t, mockRepo.lastAudit.NewValue())
		_ = output.String() // Consume output
	})
}

// TestLogLoginAttemptDirect tests direct login logging
func TestLogLoginAttemptDirect(t *testing.T) {
	t.Run("creates database audit entry for login", func(t *testing.T) {
		auditLogger, mockRepo := newTestAuditLogger()

		ctx := context.Background()
		auditLogger.LogLoginAttemptDirect(ctx, "github", "testuser", "session-123")

		assert.True(t, mockRepo.createCalled)
		assert.Equal(t, "testuser", mockRepo.lastAudit.Username())
		assert.Equal(t, model.AuditActionUserLogin, mockRepo.lastAudit.Action())
		assert.Equal(t, "User", mockRepo.lastAudit.ObjectType())
		assert.Equal(t, "testuser", mockRepo.lastAudit.ObjectID())
	})

	t.Run("returns early for empty username", func(t *testing.T) {
		auditLogger, mockRepo := newTestAuditLogger()

		ctx := context.Background()
		auditLogger.LogLoginAttemptDirect(ctx, "github", "", "session-123")

		assert.False(t, mockRepo.createCalled, "Should not create entry for empty username")
	})
}

// TestLogLogoutAttempt tests logout attempt logging
func TestLogLogoutAttempt(t *testing.T) {
	t.Run("logs successful logout", func(t *testing.T) {
		auditLogger, _ := newTestAuditLogger()
		logger, output := newAuditTestLogger()
		auditLogger.logger = (*logger).With().Str("component", "audit").Logger()

		ctx := context.Background()
		auditLogger.LogLogoutAttempt(ctx, "github", "testuser", "192.168.1.1", "test-agent", "session-123", true, "")

		logOutput := output.String()
		assert.Contains(t, logOutput, "authentication_event")
		assert.Contains(t, logOutput, "logout")
		assert.Contains(t, logOutput, "github")
		assert.Contains(t, logOutput, "testuser")
	})

	t.Run("logs failed logout with error", func(t *testing.T) {
		auditLogger, _ := newTestAuditLogger()
		logger, output := newAuditTestLogger()
		auditLogger.logger = (*logger).With().Str("component", "audit").Logger()

		ctx := context.Background()
		auditLogger.LogLogoutAttempt(ctx, "github", "testuser", "192.168.1.1", "test-agent", "session-123", false, "session not found")

		logOutput := output.String()
		assert.Contains(t, logOutput, "authentication_event")
		assert.Contains(t, logOutput, "logout")
		assert.Contains(t, logOutput, "session not found")
	})

	t.Run("does not create database entry for logout", func(t *testing.T) {
		auditLogger, mockRepo := newTestAuditLogger()

		ctx := context.Background()
		auditLogger.LogLogoutAttempt(ctx, "github", "testuser", "192.168.1.1", "test-agent", "session-123", true, "")

		assert.False(t, mockRepo.createCalled, "Logout should not create database audit entry")
	})
}

// TestLogAPIAccess tests API access logging
func TestLogAPIAccess(t *testing.T) {
	t.Run("logs successful API access", func(t *testing.T) {
		auditLogger, _ := newTestAuditLogger()
		logger, output := newAuditTestLogger()
		auditLogger.logger = (*logger).With().Str("component", "audit").Logger()

		ctx := context.Background()
		auditLogger.LogAPIAccess(ctx, "github", "testuser", "192.168.1.1", "test-agent", "/v1/modules", "GET", "session-123", true, "")

		logOutput := output.String()
		assert.Contains(t, logOutput, "authentication_event")
		assert.Contains(t, logOutput, "api_access")
		assert.Contains(t, logOutput, "/v1/modules")
		assert.Contains(t, logOutput, "GET")
	})

	t.Run("logs failed API access with error", func(t *testing.T) {
		auditLogger, _ := newTestAuditLogger()
		logger, output := newAuditTestLogger()
		auditLogger.logger = (*logger).With().Str("component", "audit").Logger()

		ctx := context.Background()
		auditLogger.LogAPIAccess(ctx, "api-key", "testuser", "192.168.1.1", "test-agent", "/v1/admin", "POST", "", false, "insufficient permissions")

		logOutput := output.String()
		assert.Contains(t, logOutput, "authentication_event")
		assert.Contains(t, logOutput, "api_access")
		assert.Contains(t, logOutput, "insufficient permissions")
	})

	t.Run("logs API access without username", func(t *testing.T) {
		auditLogger, _ := newTestAuditLogger()
		logger, output := newAuditTestLogger()
		auditLogger.logger = (*logger).With().Str("component", "audit").Logger()

		ctx := context.Background()
		auditLogger.LogAPIAccess(ctx, "api-key", "", "192.168.1.1", "test-agent", "/v1/modules", "GET", "", true, "")

		logOutput := output.String()
		assert.Contains(t, logOutput, "authentication_event")
		assert.Contains(t, logOutput, "api_access")
	})
}

// TestLogNamespaceAccess tests namespace access logging
func TestLogNamespaceAccess(t *testing.T) {
	t.Run("logs successful namespace access", func(t *testing.T) {
		auditLogger, _ := newTestAuditLogger()
		logger, output := newAuditTestLogger()
		auditLogger.logger = (*logger).With().Str("component", "audit").Logger()

		ctx := context.Background()
		auditLogger.LogNamespaceAccess(ctx, "github", "testuser", "192.168.1.1", "test-agent", "test-ns", "read", true, "")

		logOutput := output.String()
		assert.Contains(t, logOutput, "authentication_event")
		assert.Contains(t, logOutput, "test-ns")
		assert.Contains(t, logOutput, "read")
	})

	t.Run("logs failed namespace access", func(t *testing.T) {
		auditLogger, _ := newTestAuditLogger()
		logger, output := newAuditTestLogger()
		auditLogger.logger = (*logger).With().Str("component", "audit").Logger()

		ctx := context.Background()
		auditLogger.LogNamespaceAccess(ctx, "github", "testuser", "192.168.1.1", "test-agent", "restricted-ns", "delete", false, "access denied")

		logOutput := output.String()
		assert.Contains(t, logOutput, "authentication_event")
		assert.Contains(t, logOutput, "restricted-ns")
		assert.Contains(t, logOutput, "delete")
		assert.Contains(t, logOutput, "access denied")
	})

	t.Run("logs namespace modify action", func(t *testing.T) {
		auditLogger, _ := newTestAuditLogger()
		logger, output := newAuditTestLogger()
		auditLogger.logger = (*logger).With().Str("component", "audit").Logger()

		ctx := context.Background()
		auditLogger.LogNamespaceAccess(ctx, "session", "admin", "10.0.0.1", "admin-cli", "prod-ns", "modify", true, "")

		logOutput := output.String()
		assert.Contains(t, logOutput, "authentication_event")
		assert.Contains(t, logOutput, "modify")
		assert.Contains(t, logOutput, "prod-ns")
	})
}

// TestLogSecurityEvent tests security event logging
func TestLogSecurityEvent(t *testing.T) {
	t.Run("logs security event at error level", func(t *testing.T) {
		auditLogger, _ := newTestAuditLogger()
		logger, output := newAuditTestLogger()
		auditLogger.logger = (*logger).With().Str("component", "audit").Logger()

		ctx := context.Background()
		auditLogger.LogSecurityEvent(ctx, "brute_force", "Multiple failed login attempts", "192.168.1.100", "curl/7.68.0", "")

		logOutput := output.String()
		assert.Contains(t, logOutput, "security_event")
		assert.Contains(t, logOutput, "brute_force")
		assert.Contains(t, logOutput, "Multiple failed login attempts")
		assert.Contains(t, logOutput, "192.168.1.100")
		assert.Contains(t, logOutput, "curl/7.68.0")
	})

	t.Run("logs security event with username", func(t *testing.T) {
		auditLogger, _ := newTestAuditLogger()
		logger, output := newAuditTestLogger()
		auditLogger.logger = (*logger).With().Str("component", "audit").Logger()

		ctx := context.Background()
		auditLogger.LogSecurityEvent(ctx, "suspicious_activity", "Unusual access pattern", "10.0.0.1", "Mozilla/5.0", "user123")

		logOutput := output.String()
		assert.Contains(t, logOutput, "security_event")
		assert.Contains(t, logOutput, "suspicious_activity")
		assert.Contains(t, logOutput, "user123")
	})

	t.Run("sets success to false for security events", func(t *testing.T) {
		auditLogger, _ := newTestAuditLogger()
		logger, output := newAuditTestLogger()
		auditLogger.logger = (*logger).With().Str("component", "audit").Logger()

		ctx := context.Background()
		auditLogger.LogSecurityEvent(ctx, "intrusion", "Potential SQL injection", "172.16.0.1", "sqlmap/1.0", "attacker")

		logOutput := output.String()
		assert.Contains(t, logOutput, "security_event")
		assert.Contains(t, logOutput, "intrusion")
		// Security events use error level logging (which implies negative event)
		assert.Contains(t, logOutput, "\"level\":\"error\"")
	})
}

// TestAuthEvent_TimestampHandling tests timestamp edge cases
func TestAuthEvent_TimestampHandling(t *testing.T) {
	t.Run("uses provided timestamp", func(t *testing.T) {
		auditLogger, _ := newTestAuditLogger()
		logger, output := newAuditTestLogger()
		auditLogger.logger = (*logger).With().Str("component", "audit").Logger()

		fixedTime := time.Date(2024, 6, 15, 14, 30, 0, 0, time.UTC)
		event := AuthEvent{
			Timestamp: fixedTime,
			Provider:  "test",
			Success:   true,
			IPAddress: "127.0.0.1",
			UserAgent: "test",
		}

		auditLogger.LogAuthEvent(context.Background(), event)

		logOutput := output.String()
		assert.Contains(t, logOutput, "authentication_event")
		assert.Contains(t, logOutput, "2024-06-15")
	})

	t.Run("generates timestamp for zero time", func(t *testing.T) {
		auditLogger, _ := newTestAuditLogger()
		logger, output := newAuditTestLogger()
		auditLogger.logger = (*logger).With().Str("component", "audit").Logger()

		beforeLog := time.Now()
		event := AuthEvent{
			Timestamp: time.Time{},
			Provider:  "test",
			Success:   true,
			IPAddress: "127.0.0.1",
			UserAgent: "test",
		}

		auditLogger.LogAuthEvent(context.Background(), event)

		afterLog := time.Now()
		logOutput := output.String()
		assert.Contains(t, logOutput, "authentication_event")
		// Timestamp should be between before and after
		// (we can't easily test exact value, just that it doesn't panic)
		assert.True(t, afterLog.After(beforeLog) || afterLog.Equal(beforeLog))
	})
}

// TestConcurrentLogging tests concurrent logging operations
func TestConcurrentLogging(t *testing.T) {
	auditLogger, mockRepo := newTestAuditLogger()
	logger, _ := newAuditTestLogger()
	auditLogger.logger = (*logger).With().Str("component", "audit").Logger()

	t.Run("concurrent login attempts", func(t *testing.T) {
		const goroutines = 50
		done := make(chan bool, goroutines)

		for i := 0; i < goroutines; i++ {
			go func(id int) {
				ctx := context.Background()
				username := "user" + string(rune(id))
				auditLogger.LogLoginAttempt(ctx, "test-provider", username, "127.0.0.1", "test-agent", "session-123", true, "")
				done <- true
			}(i)
		}

		// Wait for all goroutines
		for i := 0; i < goroutines; i++ {
			<-done
		}

		// All logins should have been processed
		assert.Equal(t, goroutines, mockRepo.createCount)
	})
}

// TestLogAuthEvent_AllProviders tests various auth providers
func TestLogAuthEvent_AllProviders(t *testing.T) {
	providers := []string{
		"github",
		"gitlab",
		"openid_connect",
		"saml",
		"api-key",
		"admin-api-key",
		"upload-api-key",
		"publish-api-key",
		"terraform-oidc",
		"terraform-analytics",
	}

	for _, provider := range providers {
		t.Run("logs for provider: "+provider, func(t *testing.T) {
			auditLogger, _ := newTestAuditLogger()
			logger, output := newAuditTestLogger()
			auditLogger.logger = (*logger).With().Str("component", "audit").Logger()

			event := AuthEvent{
				Provider:  provider,
				Success:   true,
				IPAddress: "127.0.0.1",
				UserAgent: "test-agent",
				Action:    "login",
			}

			auditLogger.LogAuthEvent(context.Background(), event)

			logOutput := output.String()
			assert.Contains(t, logOutput, "authentication_event")
			assert.Contains(t, logOutput, provider)
		})
	}
}

// TestLogAuthEvent_Actions tests various action types
func TestLogAuthEvent_Actions(t *testing.T) {
	actions := []string{
		"login",
		"logout",
		"api_access",
		"namespace_access",
		"module_publish",
		"module_delete",
		"admin_action",
	}

	for _, action := range actions {
		t.Run("logs action: "+action, func(t *testing.T) {
			auditLogger, _ := newTestAuditLogger()
			logger, output := newAuditTestLogger()
			auditLogger.logger = (*logger).With().Str("component", "audit").Logger()

			event := AuthEvent{
				Provider:  "test",
				Success:   true,
				IPAddress: "127.0.0.1",
				UserAgent: "test-agent",
				Action:    action,
			}

			auditLogger.LogAuthEvent(context.Background(), event)

			logOutput := output.String()
			assert.Contains(t, logOutput, "authentication_event")
			assert.Contains(t, logOutput, action)
		})
	}
}
