package service

import (
	"context"

	auditRepo "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/audit/repository"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/audit/model"
)

// ProviderAuditServiceInterface defines the interface for provider audit operations
// This allows for proper mocking in tests while keeping the implementation in ProviderAuditService
type ProviderAuditServiceInterface interface {
	LogProviderCreate(ctx context.Context, providerName, namespace string) error
	LogProviderDelete(ctx context.Context, providerName, namespace string) error
	LogProviderVersionIndex(ctx context.Context, providerName, namespace, version string) error
	LogProviderVersionDelete(ctx context.Context, providerName, namespace, version string) error
}

// ProviderAuditService handles audit logging for provider operations
// It implements ProviderAuditServiceInterface
// Python reference: /app/terrareg/provider_model.py - provider audit event creation
type ProviderAuditService struct {
	auditRepo auditRepo.AuditHistoryRepository
}

// Ensure ProviderAuditService implements the interface at compile time
var _ ProviderAuditServiceInterface = (*ProviderAuditService)(nil)

// NewProviderAuditService creates a new ProviderAuditService
func NewProviderAuditService(auditRepo auditRepo.AuditHistoryRepository) *ProviderAuditService {
	return &ProviderAuditService{
		auditRepo: auditRepo,
	}
}

// LogProviderCreate logs provider creation audit event
// Python reference: /app/terrareg/models.py:4229 - AuditAction.PROVIDER_CREATE
func (s *ProviderAuditService) LogProviderCreate(ctx context.Context, providerName, namespace string) error {
	username := getUsernameFromContext(ctx)
	objectID := namespace + "/" + providerName
	audit := model.NewAuditHistory(
		username,
		model.AuditActionProviderCreate,
		"Provider",
		objectID,
		nil,
		nil,
	)
	return s.auditRepo.Create(ctx, audit)
}

// LogProviderDelete logs provider deletion audit event
// Python reference: /app/terrareg/provider_model.py:85 - AuditAction.PROVIDER_DELETE
func (s *ProviderAuditService) LogProviderDelete(ctx context.Context, providerName, namespace string) error {
	username := getUsernameFromContext(ctx)
	objectID := namespace + "/" + providerName
	audit := model.NewAuditHistory(
		username,
		model.AuditActionProviderDelete,
		"Provider",
		objectID,
		nil,
		nil,
	)
	return s.auditRepo.Create(ctx, audit)
}

// LogProviderVersionIndex logs provider version index audit event
// Python reference: implied by module version index pattern
func (s *ProviderAuditService) LogProviderVersionIndex(ctx context.Context, providerName, namespace, version string) error {
	username := getUsernameFromContext(ctx)
	objectID := namespace + "/" + providerName + "/" + version
	audit := model.NewAuditHistory(
		username,
		model.AuditActionProviderVersionIndex,
		"ProviderVersion",
		objectID,
		nil,
		nil,
	)
	return s.auditRepo.Create(ctx, audit)
}

// LogProviderVersionDelete logs provider version deletion audit event
// Python reference: /app/terrareg/provider_version_model.py:209 - AuditAction.PROVIDER_VERSION_DELETE
func (s *ProviderAuditService) LogProviderVersionDelete(ctx context.Context, providerName, namespace, version string) error {
	username := getUsernameFromContext(ctx)
	objectID := namespace + "/" + providerName + "/" + version
	audit := model.NewAuditHistory(
		username,
		model.AuditActionProviderVersionDelete,
		"ProviderVersion",
		objectID,
		nil,
		nil,
	)
	return s.auditRepo.Create(ctx, audit)
}
