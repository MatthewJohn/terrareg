package module

import (
	"context"
	"database/sql"

	moduleCmd "github.com/matthewjohn/terrareg/terrareg-go/internal/application/command/module"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/repository"
	"gorm.io/gorm"
)

// ModuleProviderRedirectRepositoryImpl implements both module provider redirect repository interfaces
type ModuleProviderRedirectRepositoryImpl struct {
	db *gorm.DB
}

// NewModuleProviderRedirectRepository creates a new module provider redirect repository
func NewModuleProviderRedirectRepository(db *gorm.DB) *ModuleProviderRedirectRepositoryImpl {
	return &ModuleProviderRedirectRepositoryImpl{
		db: db,
	}
}

// Create creates a new module provider redirect (implements command interface)
func (r *ModuleProviderRedirectRepositoryImpl) Create(ctx context.Context, req moduleCmd.CreateModuleProviderRedirectRequest) error {
	redirect := &repository.ModuleProviderRedirect{
		ModuleProviderID:   req.ToModuleProviderID,
		NamespaceID:        0, // Will need to fetch this from the module provider
		Module:             req.FromModule,
		Provider:           req.FromProvider,
	}

	// First, get the namespace ID from the module provider
	var namespaceID sql.NullInt64
	err := r.db.WithContext(ctx).Table("module_provider").
		Select("namespace_id").
		Where("id = ?", req.ToModuleProviderID).
		Scan(&namespaceID).Error
	if err != nil {
		return err
	}

	if !namespaceID.Valid {
		return sql.ErrNoRows
	}

	redirect.NamespaceID = int(namespaceID.Int64)

	// Check if redirect already exists
	var existingCount int64
	err = r.db.WithContext(ctx).Table("module_provider_redirect").
		Where("namespace_id = ? AND module = ? AND provider = ?", redirect.NamespaceID, redirect.Module, redirect.Provider).
		Count(&existingCount).Error
	if err != nil {
		return err
	}

	if existingCount > 0 {
		// Redirect already exists, update it
		return r.db.WithContext(ctx).Table("module_provider_redirect").
			Where("namespace_id = ? AND module = ? AND provider = ?", redirect.NamespaceID, redirect.Module, redirect.Provider).
			Update("module_provider_id", req.ToModuleProviderID).Error
	}

	// Create new redirect
	return r.db.WithContext(ctx).Table("module_provider_redirect").
		Create(redirect).Error
}

// GetAll retrieves all module provider redirects (implements command interface)
func (r *ModuleProviderRedirectRepositoryImpl) GetAll(ctx context.Context) ([]*moduleCmd.ModuleProviderRedirect, error) {
	type redirectResult struct {
		ID               int
		ModuleProviderID int
		NamespaceName    string
		FromModule       string
		FromProvider     string
	}

	var results []redirectResult

	// Get redirects with module provider details
	err := r.db.WithContext(ctx).
		Table("module_provider_redirect mpr").
		Select(`
			mpr.id,
			mpr.module_provider_id,
			ns.name as namespace_name,
			mp.module as from_module,
			mp.provider as from_provider
		`).
		Joins("LEFT JOIN module_provider mp ON mpr.module_provider_id = mp.id").
		Joins("LEFT JOIN namespace ns ON mp.namespace_id = ns.id").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	// Convert to expected format
	var redirects []*moduleCmd.ModuleProviderRedirect
	for _, result := range results {
		redirects = append(redirects, &moduleCmd.ModuleProviderRedirect{
			ID:               result.ID,
			FromNamespace:    result.NamespaceName,
			FromModule:       result.FromModule,
			FromProvider:     result.FromProvider,
			ToModuleProviderID: result.ModuleProviderID,
		})
	}

	return redirects, nil
}

// GetByFrom retrieves a redirect by the from fields (implements command interface)
func (r *ModuleProviderRedirectRepositoryImpl) GetByFrom(ctx context.Context, namespace, moduleName, providerName string) (*moduleCmd.ModuleProviderRedirect, error) {
	var redirect struct {
		ID               int
		ModuleProviderID int
		NamespaceID      int
		Module           string
		Provider         string
		NamespaceName    string
	}

	err := r.db.WithContext(ctx).
		Table("module_provider_redirect mpr").
		Select(`
			mpr.id,
			mpr.module_provider_id,
			mpr.namespace_id,
			mp.module as module,
			mp.provider as provider,
			ns.name as namespace_name
		`).
		Joins("LEFT JOIN module_provider mp ON mpr.module_provider_id = mp.id").
		Joins("LEFT JOIN namespace ns ON mp.namespace_id = ns.id").
		Where("ns.name = ? AND mpr.module = ? AND mpr.provider = ?", namespace, moduleName, providerName).
		Scan(&redirect).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	return &moduleCmd.ModuleProviderRedirect{
		ID:               redirect.ID,
		FromNamespace:    redirect.NamespaceName,
		FromModule:       redirect.Module,
		FromProvider:     redirect.Provider,
		ToModuleProviderID: redirect.ModuleProviderID,
	}, nil
}

// Delete deletes a module provider redirect (implements command interface)
func (r *ModuleProviderRedirectRepositoryImpl) Delete(ctx context.Context, namespace, module, provider string) error {
	return r.db.WithContext(ctx).
		Table("module_provider_redirect").
		Joins("INNER JOIN namespace ns ON module_provider_redirect.namespace_id = ns.id").
		Where("ns.name = ? AND module_provider_redirect.module = ? AND module_provider_redirect.provider = ?", namespace, module, provider).
		Delete(&repository.ModuleProviderRedirect{}).Error
}

// GetByOriginalDetails retrieves a module provider by original details (implements domain interface)
func (r *ModuleProviderRedirectRepositoryImpl) GetByOriginalDetails(ctx context.Context, namespace, moduleName, providerName string, caseInsensitive bool) (*model.ModuleProvider, error) {
	// First try to find a redirect
	var redirect repository.ModuleProviderRedirect
	query := r.db.WithContext(ctx).
		Table("module_provider_redirect").
		Joins("INNER JOIN namespace ns ON module_provider_redirect.namespace_id = ns.id")

	if caseInsensitive {
		query = query.Where("LOWER(ns.name) = LOWER(?) AND LOWER(module_provider_redirect.module) = LOWER(?) AND LOWER(module_provider_redirect.provider) = LOWER(?)", namespace, moduleName, providerName)
	} else {
		query = query.Where("ns.name = ? AND module_provider_redirect.module = ? AND module_provider_redirect.provider = ?", namespace, moduleName, providerName)
	}

	err := query.First(&redirect).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// No redirect found, try to find the module provider directly
			return r.findModuleProviderDirectly(ctx, namespace, moduleName, providerName, caseInsensitive)
		}
		return nil, err
	}

	// Found redirect, now get the target module provider
	var targetProvider model.ModuleProvider
	err = r.db.WithContext(ctx).
		Preload("Namespace").
		Preload("LatestVersion").
		First(&targetProvider, redirect.ModuleProviderID).Error
	if err != nil {
		return nil, err
	}

	return &targetProvider, nil
}

// GetByModuleProvider retrieves redirects for a module provider (implements domain interface)
func (r *ModuleProviderRedirectRepositoryImpl) GetByModuleProvider(ctx context.Context, moduleProviderID int) ([]*repository.ModuleProviderRedirect, error) {
	var redirects []*repository.ModuleProviderRedirect
	err := r.db.WithContext(ctx).
		Where("module_provider_id = ?", moduleProviderID).
		Find(&redirects).Error
	return redirects, err
}

// Helper method to find module provider directly
func (r *ModuleProviderRedirectRepositoryImpl) findModuleProviderDirectly(ctx context.Context, namespace, moduleName, providerName string, caseInsensitive bool) (*model.ModuleProvider, error) {
	var mp model.ModuleProvider
	query := r.db.WithContext(ctx).
		Preload("Namespace").
		Preload("LatestVersion").
		Joins("INNER JOIN namespace ns ON module_provider.namespace_id = ns.id")

	if caseInsensitive {
		query = query.Where("LOWER(ns.name) = LOWER(?) AND LOWER(module_provider.module) = LOWER(?) AND LOWER(module_provider.provider) = LOWER(?)", namespace, moduleName, providerName)
	} else {
		query = query.Where("ns.name = ? AND module_provider.module = ? AND module_provider.provider = ?", namespace, moduleName, providerName)
	}

	err := query.First(&mp).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	return &mp, nil
}