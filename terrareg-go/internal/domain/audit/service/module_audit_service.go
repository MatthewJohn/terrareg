package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/audit/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/audit/repository"
)

// ModuleAuditService handles audit logging for module operations
type ModuleAuditService struct {
	auditRepo repository.AuditHistoryRepository
}

// NewModuleAuditService creates a new module audit service
func NewModuleAuditService(auditRepo repository.AuditHistoryRepository) *ModuleAuditService {
	return &ModuleAuditService{
		auditRepo: auditRepo,
	}
}

// LogModuleProviderCreate logs when a module provider is created
func (s *ModuleAuditService) LogModuleProviderCreate(ctx context.Context, username, namespace, module, provider string) error {
	objectID := fmt.Sprintf("%s/%s/%s", namespace, module, provider)
	audit := model.NewAuditHistory(
		username,
		model.AuditActionModuleProviderCreate,
		"module_provider",
		objectID,
		nil,
		nil,
	)
	return s.auditRepo.Create(ctx, audit)
}

// LogModuleProviderUpdate logs when a module provider is updated
func (s *ModuleAuditService) LogModuleProviderUpdate(ctx context.Context, username, namespace, module, provider string, oldValues, newValues map[string]string) error {
	objectID := fmt.Sprintf("%s/%s/%s", namespace, module, provider)

	oldValueJSON, _ := json.Marshal(oldValues)
	newValueJSON, _ := json.Marshal(newValues)

	oldValueStr := string(oldValueJSON)
	newValueStr := string(newValueJSON)

	audit := model.NewAuditHistory(
		username,
		model.AuditActionModuleProviderUpdate,
		"module_provider",
		objectID,
		&oldValueStr,
		&newValueStr,
	)
	return s.auditRepo.Create(ctx, audit)
}

// LogModuleProviderDelete logs when a module provider is deleted
func (s *ModuleAuditService) LogModuleProviderDelete(ctx context.Context, username, namespace, module, provider string) error {
	objectID := fmt.Sprintf("%s/%s/%s", namespace, module, provider)
	audit := model.NewAuditHistory(
		username,
		model.AuditActionModuleProviderDelete,
		"module_provider",
		objectID,
		nil,
		nil,
	)
	return s.auditRepo.Create(ctx, audit)
}

// LogModuleVersionCreate logs when a module version is created
func (s *ModuleAuditService) LogModuleVersionCreate(ctx context.Context, username, namespace, module, provider, version string) error {
	objectID := fmt.Sprintf("%s/%s/%s/%s", namespace, module, provider, version)
	audit := model.NewAuditHistory(
		username,
		model.AuditActionModuleVersionCreate,
		"module_version",
		objectID,
		nil,
		nil,
	)
	return s.auditRepo.Create(ctx, audit)
}

// LogModuleVersionPublish logs when a module version is published
func (s *ModuleAuditService) LogModuleVersionPublish(ctx context.Context, username, namespace, module, provider, version string) error {
	objectID := fmt.Sprintf("%s/%s/%s/%s", namespace, module, provider, version)
	audit := model.NewAuditHistory(
		username,
		model.AuditActionModuleVersionPublish,
		"module_version",
		objectID,
		nil,
		nil,
	)
	return s.auditRepo.Create(ctx, audit)
}

// LogModuleVersionDelete logs when a module version is deleted
func (s *ModuleAuditService) LogModuleVersionDelete(ctx context.Context, username, namespace, module, provider, version string) error {
	objectID := fmt.Sprintf("%s/%s/%s/%s", namespace, module, provider, version)
	audit := model.NewAuditHistory(
		username,
		model.AuditActionModuleVersionDelete,
		"module_version",
		objectID,
		nil,
		nil,
	)
	return s.auditRepo.Create(ctx, audit)
}

// LogModuleVersionIndex logs when a module version is indexed
func (s *ModuleAuditService) LogModuleVersionIndex(ctx context.Context, username, namespace, module, provider, version string) error {
	objectID := fmt.Sprintf("%s/%s/%s/%s", namespace, module, provider, version)
	audit := model.NewAuditHistory(
		username,
		model.AuditActionModuleVersionIndex,
		"module_version",
		objectID,
		nil,
		nil,
	)
	return s.auditRepo.Create(ctx, audit)
}