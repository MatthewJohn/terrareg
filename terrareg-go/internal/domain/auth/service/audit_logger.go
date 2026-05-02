package service

import (
	"context"
	"time"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/audit/model"
	auditRepo "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/audit/repository"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/logging"
)

// AuthEvent represents an authentication event for audit logging
type AuthEvent struct {
	Timestamp time.Time `json:"timestamp"`
	Provider  string    `json:"provider"`
	Username  string    `json:"username,omitempty"`
	Success   bool      `json:"success"`
	IPAddress string    `json:"ip_address"`
	UserAgent string    `json:"user_agent"`
	SessionID string    `json:"session_id,omitempty"`
	Error     string    `json:"error,omitempty"`
	Endpoint  string    `json:"endpoint,omitempty"`
	Method    string    `json:"method,omitempty"`
	Namespace string    `json:"namespace,omitempty"`
	Action    string    `json:"action,omitempty"`
}

// AuditLogger handles comprehensive audit logging for authentication events
// Python reference: /app/terrareg/server/api/github/github_login_callback.py:65 - USER_LOGIN audit
type AuditLogger struct {
	logger    logging.Logger
	auditRepo auditRepo.AuditHistoryRepository
}

// NewAuditLogger creates a new audit logger
func NewAuditLogger(baseLogger logging.Logger, auditRepo auditRepo.AuditHistoryRepository) *AuditLogger {
	return &AuditLogger{
		logger:    baseLogger,
		auditRepo: auditRepo,
	}
}

// LogAuthEvent logs an authentication event
func (a *AuditLogger) LogAuthEvent(ctx context.Context, event AuthEvent) {
	// Ensure timestamp is set
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now().UTC()
	}

	logEvent := a.logger.Info()
	if !event.Success {
		logEvent = a.logger.Warn()
	}

	// Build log entry with all fields
	logEntry := logEvent.
		Time("timestamp", event.Timestamp).
		Str("provider", event.Provider).
		Bool("success", event.Success).
		Str("ip_address", event.IPAddress).
		Str("user_agent", event.UserAgent)

	// Add optional fields if they exist
	if event.Username != "" {
		logEntry = logEntry.Str("username", event.Username)
	}
	if event.SessionID != "" {
		logEntry = logEntry.Str("session_id", event.SessionID)
	}
	if event.Error != "" {
		logEntry = logEntry.Str("error", event.Error)
	}
	if event.Endpoint != "" {
		logEntry = logEntry.Str("endpoint", event.Endpoint)
	}
	if event.Method != "" {
		logEntry = logEntry.Str("method", event.Method)
	}
	if event.Namespace != "" {
		logEntry = logEntry.Str("namespace", event.Namespace)
	}
	if event.Action != "" {
		logEntry = logEntry.Str("action", event.Action)
	}

	// Log the event
	logEntry.Msg("authentication_event")
}

// LogLoginAttemptDirect creates a database audit entry for a login event
// This is a convenience method for direct audit logging without the full log event
// Python reference: /app/terrareg/server/api/github/github_login_callback.py:65 - USER_LOGIN audit
// Note: Python uses old_value=None, new_value=None for login events
func (a *AuditLogger) LogLoginAttemptDirect(ctx context.Context, provider, username, sessionID string) {
	if username == "" {
		return
	}

	// Create database audit entry for successful logins
	// Python reference: /app/terrareg/server/api/github/github_login_callback.py:65
	// Note: Python uses old_value=None, new_value=None for login events
	audit := model.NewAuditHistory(
		username,
		model.AuditActionUserLogin,
		"User",
		username,
		nil, // old_value is None (matching Python)
		nil, // new_value is None (matching Python)
	)
	// Create audit entry synchronously
	_ = a.auditRepo.Create(ctx, audit)
}

// LogLoginAttempt logs a login attempt
// Python reference: /app/terrareg/server/api/github/github_login_callback.py:65 - USER_LOGIN audit
func (a *AuditLogger) LogLoginAttempt(ctx context.Context, provider, username, ipAddress, userAgent, sessionID string, success bool, errorMsg string) {
	event := AuthEvent{
		Timestamp: time.Now().UTC(),
		Provider:  provider,
		Username:  username,
		Success:   success,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		SessionID: sessionID,
		Action:    "login",
	}

	if !success && errorMsg != "" {
		event.Error = errorMsg
	}

	a.LogAuthEvent(ctx, event)

	// Create database audit entry for successful logins
	// Python reference: /app/terrareg/server/api/github/github_login_callback.py:65 - USER_LOGIN audit
	// Note: Python uses old_value=None, new_value=None for login events
	if success && username != "" {
		audit := model.NewAuditHistory(
			username,
			model.AuditActionUserLogin,
			"User",
			username,
			nil, // old_value is None (matching Python)
			nil, // new_value is None (matching Python)
		)
		// Create audit entry synchronously
		_ = a.auditRepo.Create(ctx, audit)
	}
}

// LogLogoutAttempt logs a logout attempt
func (a *AuditLogger) LogLogoutAttempt(ctx context.Context, provider, username, ipAddress, userAgent, sessionID string, success bool, errorMsg string) {
	event := AuthEvent{
		Timestamp: time.Now().UTC(),
		Provider:  provider,
		Username:  username,
		Success:   success,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		SessionID: sessionID,
		Action:    "logout",
	}

	if !success && errorMsg != "" {
		event.Error = errorMsg
	}

	a.LogAuthEvent(ctx, event)
}

// LogAPIAccess logs API access with authentication context
func (a *AuditLogger) LogAPIAccess(ctx context.Context, provider, username, ipAddress, userAgent, endpoint, method, sessionID string, success bool, errorMsg string) {
	event := AuthEvent{
		Timestamp: time.Now().UTC(),
		Provider:  provider,
		Username:  username,
		Success:   success,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Endpoint:  endpoint,
		Method:    method,
		SessionID: sessionID,
		Action:    "api_access",
	}

	if !success && errorMsg != "" {
		event.Error = errorMsg
	}

	a.LogAuthEvent(ctx, event)
}

// LogNamespaceAccess logs namespace access attempts
func (a *AuditLogger) LogNamespaceAccess(ctx context.Context, provider, username, ipAddress, userAgent, namespace, action string, success bool, errorMsg string) {
	event := AuthEvent{
		Timestamp: time.Now().UTC(),
		Provider:  provider,
		Username:  username,
		Success:   success,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Namespace: namespace,
		Action:    action,
	}

	if !success && errorMsg != "" {
		event.Error = errorMsg
	}

	a.LogAuthEvent(ctx, event)
}

// LogSecurityEvent logs security-related events
func (a *AuditLogger) LogSecurityEvent(ctx context.Context, eventType, description, ipAddress, userAgent, username string) {
	event := AuthEvent{
		Timestamp: time.Now().UTC(),
		Provider:  "security",
		Username:  username,
		Success:   false, // Security events are typically negative
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Error:     description,
		Action:    eventType,
	}

	// Use error level for security events
	a.logger.Error().
		Time("timestamp", event.Timestamp).
		Str("event_type", eventType).
		Str("description", description).
		Str("ip_address", ipAddress).
		Str("user_agent", userAgent).
		Str("username", username).
		Msg("security_event")
}
