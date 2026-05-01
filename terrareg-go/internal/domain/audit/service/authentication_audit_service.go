package service

import (
	"context"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/audit/model"
	auditRepo "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/audit/repository"
)

// AuthenticationAuditServiceInterface defines the interface for authentication audit operations
// This allows for proper mocking in tests while keeping the implementation in AuthenticationAuditService
type AuthenticationAuditServiceInterface interface {
	LogUserLogin(ctx context.Context, username, authMethod string) error
}

// AuthenticationAuditService handles audit logging for authentication events
// It implements AuthenticationAuditServiceInterface
// Python reference: /app/terrareg/server/api/github/github_login_callback.py:65
type AuthenticationAuditService struct {
	auditRepo auditRepo.AuditHistoryRepository
}

// Ensure AuthenticationAuditService implements the interface at compile time
var _ AuthenticationAuditServiceInterface = (*AuthenticationAuditService)(nil)

// NewAuthenticationAuditService creates a new AuthenticationAuditService
func NewAuthenticationAuditService(auditRepo auditRepo.AuditHistoryRepository) *AuthenticationAuditService {
	return &AuthenticationAuditService{
		auditRepo: auditRepo,
	}
}

// LogUserLogin logs user login audit event
// Python reference: AuditAction.USER_LOGIN (audit_action.py:40)
func (s *AuthenticationAuditService) LogUserLogin(ctx context.Context, username, authMethod string) error {
	// Extract username from context (falls back to parameter)
	// The auth method parameter is used for context but not stored in audit log
	// to match Python implementation (which only stores username)
	extractedUsername := getUsernameFromContext(ctx)
	if extractedUsername != "Built-in admin" {
		username = extractedUsername
	}

	audit := model.NewAuditHistory(
		username,
		model.AuditActionUserLogin,
		"User",
		username,
		nil,
		nil,
	)
	return s.auditRepo.Create(ctx, audit)
}
