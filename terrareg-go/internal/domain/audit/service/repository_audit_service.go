package service

import (
	"context"

	auditRepo "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/audit/repository"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/audit/model"
)

// RepositoryAuditService handles audit logging for repository operations
// Python reference: /app/terrareg/repository_model.py - repository audit event creation
type RepositoryAuditService struct {
	auditRepo auditRepo.AuditHistoryRepository
}

// NewRepositoryAuditService creates a new RepositoryAuditService
func NewRepositoryAuditService(auditRepo auditRepo.AuditHistoryRepository) *RepositoryAuditService {
	return &RepositoryAuditService{
		auditRepo: auditRepo,
	}
}

// LogRepositoryCreate logs repository creation audit event
// Python reference: /app/terrareg/repository_model.py:40 - AuditAction.REPOSITORY_CREATE
func (s *RepositoryAuditService) LogRepositoryCreate(ctx context.Context, repoName string) error {
	username := getUsernameFromContext(ctx)
	audit := model.NewAuditHistory(
		username,
		model.AuditActionRepositoryCreate,
		"Repository",
		repoName,
		nil,
		nil,
	)
	return s.auditRepo.Create(ctx, audit)
}

// LogRepositoryUpdate logs repository update audit event
// Python reference: /app/terrareg/repository_model.py - implied by update operations
func (s *RepositoryAuditService) LogRepositoryUpdate(ctx context.Context, repoName string, oldSettings, newSettings string) error {
	username := getUsernameFromContext(ctx)
	audit := model.NewAuditHistory(
		username,
		model.AuditActionRepositoryUpdate,
		"Repository",
		repoName,
		&oldSettings,
		&newSettings,
	)
	return s.auditRepo.Create(ctx, audit)
}

// LogRepositoryDelete logs repository deletion audit event
// Python reference: implied by delete operations
func (s *RepositoryAuditService) LogRepositoryDelete(ctx context.Context, repoName string) error {
	username := getUsernameFromContext(ctx)
	audit := model.NewAuditHistory(
		username,
		model.AuditActionRepositoryDelete,
		"Repository",
		repoName,
		nil,
		nil,
	)
	return s.auditRepo.Create(ctx, audit)
}
