package service

import (
	"context"

	auditRepo "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/audit/repository"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/audit/model"
)

// GpgKeyAuditServiceInterface defines the interface for GPG key audit operations
// This allows for proper mocking in tests while keeping the implementation in GpgKeyAuditService
type GpgKeyAuditServiceInterface interface {
	LogGpgKeyCreate(ctx context.Context, keyID, namespace string) error
	LogGpgKeyDelete(ctx context.Context, keyID, namespace string) error
}

// GpgKeyAuditService handles audit logging for GPG key operations
// It implements GpgKeyAuditServiceInterface
// Python reference: /app/terrareg/models.py - GPG key audit event creation
type GpgKeyAuditService struct {
	auditRepo auditRepo.AuditHistoryRepository
}

// Ensure GpgKeyAuditService implements the interface at compile time
var _ GpgKeyAuditServiceInterface = (*GpgKeyAuditService)(nil)

// NewGpgKeyAuditService creates a new GpgKeyAuditService
func NewGpgKeyAuditService(auditRepo auditRepo.AuditHistoryRepository) *GpgKeyAuditService {
	return &GpgKeyAuditService{
		auditRepo: auditRepo,
	}
}

// LogGpgKeyCreate logs GPG key creation audit event
// Python reference: /app/terrareg/models.py:4056 - AuditAction.GPG_KEY_CREATE
func (s *GpgKeyAuditService) LogGpgKeyCreate(ctx context.Context, keyID, namespace string) error {
	username := getUsernameFromContext(ctx)
	objectID := namespace + "/" + keyID
	audit := model.NewAuditHistory(
		username,
		model.AuditActionGpgKeyCreate,
		"GpgKey",
		objectID,
		nil,
		nil,
	)
	return s.auditRepo.Create(ctx, audit)
}

// LogGpgKeyDelete logs GPG key deletion audit event
// Python reference: /app/terrareg/models.py:4178 - AuditAction.GPG_KEY_DELETE
func (s *GpgKeyAuditService) LogGpgKeyDelete(ctx context.Context, keyID, namespace string) error {
	username := getUsernameFromContext(ctx)
	objectID := namespace + "/" + keyID
	audit := model.NewAuditHistory(
		username,
		model.AuditActionGpgKeyDelete,
		"GpgKey",
		objectID,
		nil,
		nil,
	)
	return s.auditRepo.Create(ctx, audit)
}
