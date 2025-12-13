package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider/repository"
)

// ProviderRepository implements the provider repository interface using GORM
type ProviderRepository struct {
	db *gorm.DB
}

// NewProviderRepository creates a new provider repository
func NewProviderRepository(db *gorm.DB) repository.ProviderRepository {
	return &ProviderRepository{
		db: db,
	}
}

// FindAll retrieves all providers with pagination
func (r *ProviderRepository) FindAll(ctx context.Context, offset, limit int) ([]*provider.Provider, int, error) {
	var models []ProviderModel
	var count int64

	// Get total count
	if err := r.db.WithContext(ctx).Model(&ProviderModel{}).Count(&count).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count providers: %w", err)
	}

	// Get providers with pagination
	if err := r.db.WithContext(ctx).
		Preload("Namespace").
		Preload("Category").
		Preload("LatestVersion").
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&models).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to find providers: %w", err)
	}

	// Convert to domain entities
	providers := make([]*provider.Provider, len(models))
	for i, model := range models {
		prov, err := r.modelToDomain(&model)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to convert model to domain: %w", err)
		}
		providers[i] = prov
	}

	return providers, int(count), nil
}

// Search searches for providers by query
func (r *ProviderRepository) Search(ctx context.Context, query string, offset, limit int) ([]*provider.Provider, int, error) {
	var models []ProviderModel
	var count int64

	// Build search query
	db := r.db.WithContext(ctx).Model(&ProviderModel{})
	if query != "" {
		searchPattern := "%" + strings.ToLower(query) + "%"
		db = db.Where("LOWER(name) LIKE ? OR LOWER(description) LIKE ?", searchPattern, searchPattern)
	}

	// Get total count
	if err := db.Count(&count).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count search results: %w", err)
	}

	// Get providers with pagination
	if err := db.Preload("Namespace").
		Preload("Category").
		Preload("LatestVersion").
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&models).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to search providers: %w", err)
	}

	// Convert to domain entities
	providers := make([]*provider.Provider, len(models))
	for i, model := range models {
		prov, err := r.modelToDomain(&model)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to convert model to domain: %w", err)
		}
		providers[i] = prov
	}

	return providers, int(count), nil
}

// FindByNamespaceAndName retrieves a provider by namespace and name
func (r *ProviderRepository) FindByNamespaceAndName(ctx context.Context, namespace, providerName string) (*provider.Provider, error) {
	var model ProviderModel

	// Build query
	db := r.db.WithContext(ctx).Preload("Namespace").Preload("Category").Preload("LatestVersion")
	if namespace != "" {
		// Join with namespace to filter by namespace name
		db = db.Joins("JOIN namespaces ON namespaces.id = providers.namespace_id").
			Where("namespaces.name = ? AND providers.name = ?", namespace, providerName)
	} else {
		db = db.Where("providers.name = ?", providerName)
	}

	if err := db.First(&model).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find provider: %w", err)
	}

	// Convert to domain entity
	return r.modelToDomain(&model)
}

// FindByID retrieves a provider by its ID
func (r *ProviderRepository) FindByID(ctx context.Context, providerID int) (*provider.Provider, error) {
	var model ProviderModel

	if err := r.db.WithContext(ctx).
		Preload("Namespace").
		Preload("Category").
		Preload("LatestVersion").
		First(&model, providerID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find provider: %w", err)
	}

	// Convert to domain entity
	return r.modelToDomain(&model)
}

// FindVersionsByProvider retrieves all versions for a provider
func (r *ProviderRepository) FindVersionsByProvider(ctx context.Context, providerID int) ([]*provider.ProviderVersion, error) {
	var models []ProviderVersionModel

	if err := r.db.WithContext(ctx).
		Preload("GPGKey").
		Where("provider_id = ?", providerID).
		Order("created_at DESC").
		Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to find provider versions: %w", err)
	}

	// Convert to domain entities
	versions := make([]*provider.ProviderVersion, len(models))
	for i, model := range models {
		version, err := r.versionModelToDomain(&model)
		if err != nil {
			return nil, fmt.Errorf("failed to convert version model to domain: %w", err)
		}
		versions[i] = version
	}

	return versions, nil
}

// FindVersionByProviderAndVersion retrieves a specific version
func (r *ProviderRepository) FindVersionByProviderAndVersion(ctx context.Context, providerID int, version string) (*provider.ProviderVersion, error) {
	var model ProviderVersionModel

	if err := r.db.WithContext(ctx).
		Preload("GPGKey").
		Where("provider_id = ? AND version = ?", providerID, version).
		First(&model).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find provider version: %w", err)
	}

	// Convert to domain entity
	return r.versionModelToDomain(&model)
}

// FindVersionByID retrieves a version by its ID
func (r *ProviderRepository) FindVersionByID(ctx context.Context, versionID int) (*provider.ProviderVersion, error) {
	var model ProviderVersionModel

	if err := r.db.WithContext(ctx).
		Preload("GPGKey").
		First(&model, versionID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find provider version: %w", err)
	}

	// Convert to domain entity
	return r.versionModelToDomain(&model)
}

// FindBinariesByVersion retrieves all binaries for a provider version
func (r *ProviderRepository) FindBinariesByVersion(ctx context.Context, versionID int) ([]*provider.ProviderBinary, error) {
	var models []ProviderBinaryModel

	if err := r.db.WithContext(ctx).
		Where("version_id = ?", versionID).
		Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to find provider binaries: %w", err)
	}

	// Convert to domain entities
	binaries := make([]*provider.ProviderBinary, len(models))
	for i, model := range models {
		binary := r.binaryModelToDomain(&model)
		binaries[i] = binary
	}

	return binaries, nil
}

// FindBinaryByPlatform retrieves a specific binary for a platform
func (r *ProviderRepository) FindBinaryByPlatform(ctx context.Context, versionID int, os, arch string) (*provider.ProviderBinary, error) {
	var model ProviderBinaryModel

	if err := r.db.WithContext(ctx).
		Where("version_id = ? AND operating_system = ? AND architecture = ?", versionID, os, arch).
		First(&model).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find provider binary: %w", err)
	}

	// Convert to domain entity
	return r.binaryModelToDomain(&model), nil
}

// FindGPGKeysByProvider retrieves all GPG keys for a provider
func (r *ProviderRepository) FindGPGKeysByProvider(ctx context.Context, providerID int) ([]*provider.GPGKey, error) {
	var models []GPGKeyModel

	if err := r.db.WithContext(ctx).
		Where("provider_id = ?", providerID).
		Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to find GPG keys: %w", err)
	}

	// Convert to domain entities
	keys := make([]*provider.GPGKey, len(models))
	for i, model := range models {
		key := r.gpgKeyModelToDomain(&model)
		keys[i] = key
	}

	return keys, nil
}

// FindGPGKeyByKeyID retrieves a GPG key by its key identifier
func (r *ProviderRepository) FindGPGKeyByKeyID(ctx context.Context, keyID string) (*provider.GPGKey, error) {
	var model GPGKeyModel

	if err := r.db.WithContext(ctx).
		Where("key_id = ?", keyID).
		First(&model).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find GPG key: %w", err)
	}

	// Convert to domain entity
	return r.gpgKeyModelToDomain(&model), nil
}

// Save persists a provider aggregate to the database
func (r *ProviderRepository) Save(ctx context.Context, prov *provider.Provider) error {
	model := r.domainToModel(prov)

	// Handle create or update
	if prov.ID() == 0 {
		// Create new provider
		if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
			return fmt.Errorf("failed to create provider: %w", err)
		}
		prov.SetID(model.ID)
	} else {
		// Update existing provider
		if err := r.db.WithContext(ctx).Save(model).Error; err != nil {
			return fmt.Errorf("failed to update provider: %w", err)
		}
	}

	return nil
}

// SaveVersion persists a provider version
func (r *ProviderRepository) SaveVersion(ctx context.Context, version *provider.ProviderVersion) error {
	model := r.versionDomainToModel(version)

	// Handle create or update
	if version.ID() == 0 {
		// Create new version
		if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
			return fmt.Errorf("failed to create provider version: %w", err)
		}
		version.SetID(model.ID)
	} else {
		// Update existing version
		if err := r.db.WithContext(ctx).Save(model).Error; err != nil {
			return fmt.Errorf("failed to update provider version: %w", err)
		}
	}

	return nil
}

// SaveBinary persists a provider binary
func (r *ProviderRepository) SaveBinary(ctx context.Context, binary *provider.ProviderBinary) error {
	model := r.binaryDomainToModel(binary)

	// Handle create or update
	if binary.ID() == 0 {
		// Create new binary
		if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
			return fmt.Errorf("failed to create provider binary: %w", err)
		}
		binary.SetID(model.ID)
	} else {
		// Update existing binary
		if err := r.db.WithContext(ctx).Save(model).Error; err != nil {
			return fmt.Errorf("failed to update provider binary: %w", err)
		}
	}

	return nil
}

// SaveGPGKey persists a GPG key
func (r *ProviderRepository) SaveGPGKey(ctx context.Context, gpgKey *provider.GPGKey) error {
	model := r.gpgKeyDomainToModel(gpgKey)

	// Handle create or update
	if gpgKey.ID() == 0 {
		// Create new GPG key
		if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
			return fmt.Errorf("failed to create GPG key: %w", err)
		}
		gpgKey.SetID(model.ID)
	} else {
		// Update existing GPG key
		if err := r.db.WithContext(ctx).Save(model).Error; err != nil {
			return fmt.Errorf("failed to update GPG key: %w", err)
		}
	}

	return nil
}

// DeleteVersion removes a provider version
func (r *ProviderRepository) DeleteVersion(ctx context.Context, versionID int) error {
	if err := r.db.WithContext(ctx).Delete(&ProviderVersionModel{}, versionID).Error; err != nil {
		return fmt.Errorf("failed to delete provider version: %w", err)
	}
	return nil
}

// DeleteBinary removes a provider binary
func (r *ProviderRepository) DeleteBinary(ctx context.Context, binaryID int) error {
	if err := r.db.WithContext(ctx).Delete(&ProviderBinaryModel{}, binaryID).Error; err != nil {
		return fmt.Errorf("failed to delete provider binary: %w", err)
	}
	return nil
}

// DeleteGPGKey removes a GPG key
func (r *ProviderRepository) DeleteGPGKey(ctx context.Context, keyID int) error {
	if err := r.db.WithContext(ctx).Delete(&GPGKeyModel{}, keyID).Error; err != nil {
		return fmt.Errorf("failed to delete GPG key: %w", err)
	}
	return nil
}

// SetLatestVersion updates the latest version for a provider
func (r *ProviderRepository) SetLatestVersion(ctx context.Context, providerID, versionID int) error {
	if err := r.db.WithContext(ctx).
		Model(&ProviderModel{}).
		Where("id = ?", providerID).
		Update("latest_version_id", versionID).Error; err != nil {
		return fmt.Errorf("failed to set latest version: %w", err)
	}
	return nil
}

// GetProviderVersionCount returns the number of versions for a provider
func (r *ProviderRepository) GetProviderVersionCount(ctx context.Context, providerID int) (int, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&ProviderVersionModel{}).
		Where("provider_id = ?", providerID).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count provider versions: %w", err)
	}
	return int(count), nil
}

// GetBinaryDownloadCount returns the download count for a binary
func (r *ProviderRepository) GetBinaryDownloadCount(ctx context.Context, binaryID int) (int64, error) {
	// Placeholder implementation - would need download tracking table
	return 0, nil
}

// Helper methods for converting between domain entities and database models

func (r *ProviderRepository) modelToDomain(model *ProviderModel) (*provider.Provider, error) {
	return provider.ReconstructProvider(
		model.ID,
		model.NamespaceID,
		model.Name,
		model.Description,
		model.Tier,
		model.CategoryID,
		model.RepositoryID,
		model.LatestVersionID,
		model.UseProviderSourceAuth,
	), nil
}

func (r *ProviderRepository) domainToModel(prov *provider.Provider) *ProviderModel {
	return &ProviderModel{
		ID:                    prov.ID(),
		NamespaceID:           prov.NamespaceID(),
		Name:                  prov.Name(),
		Description:           prov.Description(),
		Tier:                  prov.Tier(),
		CategoryID:            prov.CategoryID(),
		RepositoryID:          prov.RepositoryID(),
		LatestVersionID:       prov.LatestVersionID(),
		UseProviderSourceAuth: prov.UseProviderSourceAuth(),
	}
}

func (r *ProviderRepository) versionModelToDomain(model *ProviderVersionModel) (*provider.ProviderVersion, error) {
	// Parse protocol versions from JSON
	var protocolVersions []string
	if model.ProtocolVersions != "" {
		if err := json.Unmarshal([]byte(model.ProtocolVersions), &protocolVersions); err != nil {
			// Fallback to default protocol
			protocolVersions = []string{"5.0"}
		}
	}

	return provider.ReconstructProviderVersion(
		model.ID,
		model.ProviderID,
		model.Version,
		model.GitTag,
		model.Beta,
		model.PublishedAt,
		model.GPGKeyID,
		protocolVersions,
	), nil
}

func (r *ProviderRepository) versionDomainToModel(version *provider.ProviderVersion) *ProviderVersionModel {
	// Serialize protocol versions to JSON
	protocolVersionsJSON, _ := json.Marshal(version.ProtocolVersions())

	return &ProviderVersionModel{
		ID:               version.ID(),
		ProviderID:       version.ProviderID(),
		Version:          version.Version(),
		GitTag:           version.GitTag(),
		Beta:             version.Beta(),
		PublishedAt:      version.PublishedAt(),
		GPGKeyID:         version.GPGKeyID(),
		ProtocolVersions: string(protocolVersionsJSON),
	}
}

func (r *ProviderRepository) binaryModelToDomain(model *ProviderBinaryModel) *provider.ProviderBinary {
	return provider.ReconstructProviderBinary(
		model.ID,
		model.VersionID,
		model.OperatingSystem,
		model.Architecture,
		model.FileName,
		model.FileSize,
		model.FileHash,
		model.DownloadURL,
	)
}

func (r *ProviderRepository) binaryDomainToModel(binary *provider.ProviderBinary) *ProviderBinaryModel {
	return &ProviderBinaryModel{
		ID:              binary.ID(),
		VersionID:       binary.VersionID(),
		OperatingSystem: binary.OperatingSystem(),
		Architecture:    binary.Architecture(),
		FileName:        binary.FileName(),
		FileSize:        binary.FileSize(),
		FileHash:        binary.FileHash(),
		DownloadURL:     binary.DownloadURL(),
	}
}

func (r *ProviderRepository) gpgKeyModelToDomain(model *GPGKeyModel) *provider.GPGKey {
	return provider.ReconstructGPGKey(
		model.ID,
		model.KeyText,
		model.AsciiArmor,
		model.KeyID,
		model.TrustSignature,
		model.CreatedAt,
		model.UpdatedAt,
	)
}

func (r *ProviderRepository) gpgKeyDomainToModel(gpgKey *provider.GPGKey) *GPGKeyModel {
	return &GPGKeyModel{
		ID:             gpgKey.ID(),
		ProviderID:     0, // Will be set when saving
		KeyText:        gpgKey.KeyText(),
		AsciiArmor:     gpgKey.AsciiArmor(),
		KeyID:          gpgKey.KeyID(),
		TrustSignature: gpgKey.TrustSignature(),
	}
}