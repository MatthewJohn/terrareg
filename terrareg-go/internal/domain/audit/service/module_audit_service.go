package service

import (
	"context"
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

// LogModuleProviderUpdateGitTagFormat logs when git tag format is updated
func (s *ModuleAuditService) LogModuleProviderUpdateGitTagFormat(ctx context.Context, username, namespace, module, provider, oldValue, newValue string) error {
	objectID := fmt.Sprintf("%s/%s/%s", namespace, module, provider)
	audit := model.NewAuditHistory(
		username,
		model.AuditActionModuleProviderUpdateGitTagFormat,
		"module_provider",
		objectID,
		&oldValue,
		&newValue,
	)
	return s.auditRepo.Create(ctx, audit)
}

// LogModuleProviderUpdateGitProvider logs when git provider is updated
func (s *ModuleAuditService) LogModuleProviderUpdateGitProvider(ctx context.Context, username, namespace, module, provider, oldValue, newValue string) error {
	objectID := fmt.Sprintf("%s/%s/%s", namespace, module, provider)
	audit := model.NewAuditHistory(
		username,
		model.AuditActionModuleProviderUpdateGitProvider,
		"module_provider",
		objectID,
		&oldValue,
		&newValue,
	)
	return s.auditRepo.Create(ctx, audit)
}

// LogModuleProviderUpdateGitPath logs when git path is updated
func (s *ModuleAuditService) LogModuleProviderUpdateGitPath(ctx context.Context, username, namespace, module, provider, oldValue, newValue string) error {
	objectID := fmt.Sprintf("%s/%s/%s", namespace, module, provider)
	audit := model.NewAuditHistory(
		username,
		model.AuditActionModuleProviderUpdateGitPath,
		"module_provider",
		objectID,
		&oldValue,
		&newValue,
	)
	return s.auditRepo.Create(ctx, audit)
}

// LogModuleProviderUpdateArchiveGitPath logs when archive git path setting is updated
func (s *ModuleAuditService) LogModuleProviderUpdateArchiveGitPath(ctx context.Context, username, namespace, module, provider string, oldValue, newValue bool) error {
	objectID := fmt.Sprintf("%s/%s/%s", namespace, module, provider)
	oldValueStr := fmt.Sprintf("%v", oldValue)
	newValueStr := fmt.Sprintf("%v", newValue)
	audit := model.NewAuditHistory(
		username,
		model.AuditActionModuleProviderUpdateArchiveGitPath,
		"module_provider",
		objectID,
		&oldValueStr,
		&newValueStr,
	)
	return s.auditRepo.Create(ctx, audit)
}

// LogModuleProviderUpdateGitCustomBaseURL logs when custom git base URL is updated
func (s *ModuleAuditService) LogModuleProviderUpdateGitCustomBaseURL(ctx context.Context, username, namespace, module, provider, oldValue, newValue string) error {
	objectID := fmt.Sprintf("%s/%s/%s", namespace, module, provider)
	audit := model.NewAuditHistory(
		username,
		model.AuditActionModuleProviderUpdateGitCustomBaseURL,
		"module_provider",
		objectID,
		&oldValue,
		&newValue,
	)
	return s.auditRepo.Create(ctx, audit)
}

// LogModuleProviderUpdateGitCustomCloneURL logs when custom git clone URL is updated
func (s *ModuleAuditService) LogModuleProviderUpdateGitCustomCloneURL(ctx context.Context, username, namespace, module, provider, oldValue, newValue string) error {
	objectID := fmt.Sprintf("%s/%s/%s", namespace, module, provider)
	audit := model.NewAuditHistory(
		username,
		model.AuditActionModuleProviderUpdateGitCustomCloneURL,
		"module_provider",
		objectID,
		&oldValue,
		&newValue,
	)
	return s.auditRepo.Create(ctx, audit)
}

// LogModuleProviderUpdateGitCustomBrowseURL logs when custom git browse URL is updated
func (s *ModuleAuditService) LogModuleProviderUpdateGitCustomBrowseURL(ctx context.Context, username, namespace, module, provider, oldValue, newValue string) error {
	objectID := fmt.Sprintf("%s/%s/%s", namespace, module, provider)
	audit := model.NewAuditHistory(
		username,
		model.AuditActionModuleProviderUpdateGitCustomBrowseURL,
		"module_provider",
		objectID,
		&oldValue,
		&newValue,
	)
	return s.auditRepo.Create(ctx, audit)
}

// LogModuleProviderUpdateVerified logs when verified status is updated
func (s *ModuleAuditService) LogModuleProviderUpdateVerified(ctx context.Context, username, namespace, module, provider string, oldValue, newValue bool) error {
	objectID := fmt.Sprintf("%s/%s/%s", namespace, module, provider)
	oldValueStr := fmt.Sprintf("%v", oldValue)
	newValueStr := fmt.Sprintf("%v", newValue)
	audit := model.NewAuditHistory(
		username,
		model.AuditActionModuleProviderUpdateVerified,
		"module_provider",
		objectID,
		&oldValueStr,
		&newValueStr,
	)
	return s.auditRepo.Create(ctx, audit)
}

// LogModuleProviderUpdateNamespace logs when namespace is updated
func (s *ModuleAuditService) LogModuleProviderUpdateNamespace(ctx context.Context, username, namespace, module, provider, oldValue, newValue string) error {
	objectID := fmt.Sprintf("%s/%s/%s", namespace, module, provider)
	audit := model.NewAuditHistory(
		username,
		model.AuditActionModuleProviderUpdateNamespace,
		"module_provider",
		objectID,
		&oldValue,
		&newValue,
	)
	return s.auditRepo.Create(ctx, audit)
}

// LogModuleProviderUpdateModuleName logs when module name is updated
func (s *ModuleAuditService) LogModuleProviderUpdateModuleName(ctx context.Context, username, namespace, module, provider, oldValue, newValue string) error {
	objectID := fmt.Sprintf("%s/%s/%s", namespace, module, provider)
	audit := model.NewAuditHistory(
		username,
		model.AuditActionModuleProviderUpdateModuleName,
		"module_provider",
		objectID,
		&oldValue,
		&newValue,
	)
	return s.auditRepo.Create(ctx, audit)
}

// LogModuleProviderUpdateProviderName logs when provider name is updated
func (s *ModuleAuditService) LogModuleProviderUpdateProviderName(ctx context.Context, username, namespace, module, provider, oldValue, newValue string) error {
	objectID := fmt.Sprintf("%s/%s/%s", namespace, module, provider)
	audit := model.NewAuditHistory(
		username,
		model.AuditActionModuleProviderUpdateProviderName,
		"module_provider",
		objectID,
		&oldValue,
		&newValue,
	)
	return s.auditRepo.Create(ctx, audit)
}

// LogModuleProviderRedirectDelete logs when a module provider redirect is deleted
func (s *ModuleAuditService) LogModuleProviderRedirectDelete(ctx context.Context, username, namespace, module, provider string) error {
	objectID := fmt.Sprintf("%s/%s/%s", namespace, module, provider)
	audit := model.NewAuditHistory(
		username,
		model.AuditActionModuleProviderRedirectDelete,
		"module_provider_redirect",
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
