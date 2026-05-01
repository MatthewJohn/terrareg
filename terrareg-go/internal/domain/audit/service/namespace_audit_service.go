package service

import (
	"context"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/audit/model"
	auditRepo "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/audit/repository"
	types "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared/types"
)

// NamespaceAuditService handles audit logging for namespace operations
// Python reference: /app/terrareg/models.py - namespace audit event creation
type NamespaceAuditService struct {
	auditRepo auditRepo.AuditHistoryRepository
}

// NewNamespaceAuditService creates a new NamespaceAuditService
func NewNamespaceAuditService(auditRepo auditRepo.AuditHistoryRepository) *NamespaceAuditService {
	return &NamespaceAuditService{
		auditRepo: auditRepo,
	}
}

// LogNamespaceCreate logs namespace creation audit event
// Python reference: /app/terrareg/models.py:1144 - AuditAction.NAMESPACE_CREATE
func (s *NamespaceAuditService) LogNamespaceCreate(ctx context.Context, namespace types.NamespaceName) error {
	username := getUsernameFromContext(ctx)
	audit := model.NewAuditHistory(
		username,
		model.AuditActionNamespaceCreate,
		"Namespace",
		string(namespace),
		nil,
		nil,
	)
	return s.auditRepo.Create(ctx, audit)
}

// LogNamespaceModifyName logs namespace name modification audit event
// Python reference: /app/terrareg/models.py:890 - AuditAction.NAMESPACE_MODIFY_NAME
func (s *NamespaceAuditService) LogNamespaceModifyName(ctx context.Context, namespace types.NamespaceName, oldName, newName string) error {
	username := getUsernameFromContext(ctx)
	audit := model.NewAuditHistory(
		username,
		model.AuditActionNamespaceModifyName,
		"Namespace",
		string(namespace),
		&oldName,
		&newName,
	)
	return s.auditRepo.Create(ctx, audit)
}

// LogNamespaceModifyDisplayName logs namespace display name modification audit event
// Python reference: /app/terrareg/models.py:1118 - AuditAction.NAMESPACE_MODIFY_DISPLAY_NAME
func (s *NamespaceAuditService) LogNamespaceModifyDisplayName(ctx context.Context, namespace types.NamespaceName, oldDisplayName, newDisplayName *string) error {
	username := getUsernameFromContext(ctx)
	audit := model.NewAuditHistory(
		username,
		model.AuditActionNamespaceModifyDisplayName,
		"Namespace",
		string(namespace),
		oldDisplayName,
		newDisplayName,
	)
	return s.auditRepo.Create(ctx, audit)
}

// LogNamespaceDelete logs namespace deletion audit event
// Python reference: /app/terrareg/models.py:460 - AuditAction.NAMESPACE_DELETE
func (s *NamespaceAuditService) LogNamespaceDelete(ctx context.Context, namespace types.NamespaceName) error {
	username := getUsernameFromContext(ctx)
	audit := model.NewAuditHistory(
		username,
		model.AuditActionNamespaceDelete,
		"Namespace",
		string(namespace),
		nil,
		nil,
	)
	return s.auditRepo.Create(ctx, audit)
}
