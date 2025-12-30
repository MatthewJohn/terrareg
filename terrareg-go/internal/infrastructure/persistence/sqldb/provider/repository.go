package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider/repository"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
)

// ProviderRepository implements the provider repository interface using GORM
// Uses main sqldb models (ProviderDB, ProviderVersionDB, etc.) for DDD consistency
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
	var models []sqldb.ProviderDB
	var count int64

	// Get total count
	if err := r.db.WithContext(ctx).Model(&sqldb.ProviderDB{}).Count(&count).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count providers: %w", err)
	}

	// Get providers with pagination
	// Note: provider table has no created_at, so order by id
	if err := r.db.WithContext(ctx).
		Preload("Namespace").
		Preload("ProviderCategory").
		Preload("Repository").
		Preload("LatestVersion").
		Offset(offset).
		Limit(limit).
		Order("id DESC").
		Find(&models).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to find providers: %w", err)
	}

	// Convert to domain entities
	providers := make([]*provider.Provider, len(models))
	for i, model := range models {
		prov := toDomainProvider(&model)
		providers[i] = prov
	}

	return providers, int(count), nil
}

// Search searches for providers by query
func (r *ProviderRepository) Search(ctx context.Context, query string, offset, limit int) ([]*provider.Provider, int, error) {
	var models []sqldb.ProviderDB
	var count int64

	// Build search query
	// Filter to only include providers with a latest version (matching Python behavior)
	db := r.db.WithContext(ctx).Model(&sqldb.ProviderDB{}).Where("latest_version_id IS NOT NULL AND latest_version_id > 0")
	if query != "" {
		searchPattern := "%" + strings.ToLower(query) + "%"
		db = db.Where("LOWER(name) LIKE ? OR LOWER(description) LIKE ?", searchPattern, searchPattern)
	}

	// Get total count
	if err := db.Count(&count).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count search results: %w", err)
	}

	// Get providers with pagination
	queryBuilder := db.Preload("Namespace").
		Preload("ProviderCategory").
		Preload("Repository").
		Preload("LatestVersion").
		Offset(offset)

	// Only apply limit if it's greater than 0 (0 means no limit, matching module repository behavior)
	if limit > 0 {
		queryBuilder = queryBuilder.Limit(limit)
	}

	if err := queryBuilder.
		Order("id DESC").
		Find(&models).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to search providers: %w", err)
	}

	// Convert to domain entities
	providers := make([]*provider.Provider, len(models))
	for i, model := range models {
		prov := toDomainProvider(&model)
		providers[i] = prov
	}

	return providers, int(count), nil
}

// FindByNamespaceAndName retrieves a provider by namespace and name
func (r *ProviderRepository) FindByNamespaceAndName(ctx context.Context, namespace, providerName string) (*provider.Provider, error) {
	var model sqldb.ProviderDB

	// Build query - use correct table name "provider"
	db := r.db.WithContext(ctx).
		Preload("Namespace").
		Preload("ProviderCategory").
		Preload("Repository").
		Preload("LatestVersion")
	if namespace != "" {
		// Join with namespace to filter by namespace name
		// Note: namespace table uses 'namespace' as column name, not 'name'
		db = db.Joins("JOIN namespace ON namespace.id = provider.namespace_id").
			Where("namespace.namespace = ? AND provider.name = ?", namespace, providerName)
	} else {
		db = db.Where("provider.name = ?", providerName)
	}

	if err := db.First(&model).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find provider: %w", err)
	}

	// Convert to domain entity
	return toDomainProvider(&model), nil
}

// FindByID retrieves a provider by its ID
func (r *ProviderRepository) FindByID(ctx context.Context, providerID int) (*provider.Provider, error) {
	var model sqldb.ProviderDB

	if err := r.db.WithContext(ctx).
		Preload("Namespace").
		Preload("ProviderCategory").
		Preload("Repository").
		Preload("LatestVersion").
		First(&model, providerID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find provider: %w", err)
	}

	// Convert to domain entity
	return toDomainProvider(&model), nil
}

// FindVersionsByProvider retrieves all versions for a provider
func (r *ProviderRepository) FindVersionsByProvider(ctx context.Context, providerID int) ([]*provider.ProviderVersion, error) {
	var models []sqldb.ProviderVersionDB

	if err := r.db.WithContext(ctx).
		Preload("GPGKey").
		Where("provider_id = ?", providerID).
		Order("id DESC").
		Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to find provider versions: %w", err)
	}

	// Convert to domain entities
	versions := make([]*provider.ProviderVersion, len(models))
	for i, model := range models {
		version := toDomainProviderVersion(&model)
		versions[i] = version
	}

	return versions, nil
}

// FindVersionByProviderAndVersion retrieves a specific version
func (r *ProviderRepository) FindVersionByProviderAndVersion(ctx context.Context, providerID int, version string) (*provider.ProviderVersion, error) {
	var model sqldb.ProviderVersionDB

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
	return toDomainProviderVersion(&model), nil
}

// FindVersionByID retrieves a version by its ID
func (r *ProviderRepository) FindVersionByID(ctx context.Context, versionID int) (*provider.ProviderVersion, error) {
	var model sqldb.ProviderVersionDB

	if err := r.db.WithContext(ctx).
		Preload("GPGKey").
		First(&model, versionID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find provider version: %w", err)
	}

	// Convert to domain entity
	return toDomainProviderVersion(&model), nil
}

// FindBinariesByVersion retrieves all binaries for a provider version
func (r *ProviderRepository) FindBinariesByVersion(ctx context.Context, versionID int) ([]*provider.ProviderBinary, error) {
	var models []sqldb.ProviderVersionBinaryDB

	if err := r.db.WithContext(ctx).
		Where("provider_version_id = ?", versionID).
		Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to find provider binaries: %w", err)
	}

	// Convert to domain entities
	binaries := make([]*provider.ProviderBinary, len(models))
	for i, model := range models {
		binary := toDomainProviderBinary(&model)
		binaries[i] = binary
	}

	return binaries, nil
}

// FindBinaryByPlatform retrieves a specific binary for a platform
func (r *ProviderRepository) FindBinaryByPlatform(ctx context.Context, versionID int, os, arch string) (*provider.ProviderBinary, error) {
	var model sqldb.ProviderVersionBinaryDB

	if err := r.db.WithContext(ctx).
		Where("provider_version_id = ? AND operating_system = ? AND architecture = ?", versionID, os, arch).
		First(&model).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find provider binary: %w", err)
	}

	// Convert to domain entity
	return toDomainProviderBinary(&model), nil
}

// FindGPGKeysByProvider retrieves all GPG keys for a provider
// NOTE: GPG keys belong to namespaces, not providers. This method is kept for compatibility
// but should be deprecated. Use namespace-based GPG key lookup instead.
func (r *ProviderRepository) FindGPGKeysByProvider(ctx context.Context, providerID int) ([]*provider.GPGKey, error) {
	// First get the provider to find its namespace
	var providerModel sqldb.ProviderDB
	if err := r.db.WithContext(ctx).First(&providerModel, providerID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find provider: %w", err)
	}

	// Now get GPG keys for the namespace
	var models []sqldb.GPGKeyDB
	if err := r.db.WithContext(ctx).
		Where("namespace_id = ?", providerModel.NamespaceID).
		Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to find GPG keys: %w", err)
	}

	// Convert to domain entities
	keys := make([]*provider.GPGKey, len(models))
	for i, model := range models {
		key := toDomainGPGKey(&model)
		keys[i] = key
	}

	return keys, nil
}

// FindGPGKeyByKeyID retrieves a GPG key by its key identifier
func (r *ProviderRepository) FindGPGKeyByKeyID(ctx context.Context, keyID string) (*provider.GPGKey, error) {
	var model sqldb.GPGKeyDB

	if err := r.db.WithContext(ctx).
		Where("key_id = ?", keyID).
		First(&model).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find GPG key: %w", err)
	}

	// Convert to domain entity
	return toDomainGPGKey(&model), nil
}

// Save persists a provider aggregate to the database
func (r *ProviderRepository) Save(ctx context.Context, prov *provider.Provider) error {
	model := toDBProvider(prov)

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
	model := toDBProviderVersion(version)

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
	model := toDBProviderBinary(binary)

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
	model := toDBGPGKey(gpgKey)

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
	if err := r.db.WithContext(ctx).Delete(&sqldb.ProviderVersionDB{}, versionID).Error; err != nil {
		return fmt.Errorf("failed to delete provider version: %w", err)
	}
	return nil
}

// DeleteBinary removes a provider binary
func (r *ProviderRepository) DeleteBinary(ctx context.Context, binaryID int) error {
	if err := r.db.WithContext(ctx).Delete(&sqldb.ProviderVersionBinaryDB{}, binaryID).Error; err != nil {
		return fmt.Errorf("failed to delete provider binary: %w", err)
	}
	return nil
}

// DeleteGPGKey removes a GPG key
func (r *ProviderRepository) DeleteGPGKey(ctx context.Context, keyID int) error {
	if err := r.db.WithContext(ctx).Delete(&sqldb.GPGKeyDB{}, keyID).Error; err != nil {
		return fmt.Errorf("failed to delete GPG key: %w", err)
	}
	return nil
}

// SetLatestVersion updates the latest version for a provider
func (r *ProviderRepository) SetLatestVersion(ctx context.Context, providerID, versionID int) error {
	if err := r.db.WithContext(ctx).
		Model(&sqldb.ProviderDB{}).
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
		Model(&sqldb.ProviderVersionDB{}).
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

// Mapper functions (domain → DB)
// Following the same pattern as the module package

func toDBProvider(p *provider.Provider) *sqldb.ProviderDB {
	return &sqldb.ProviderDB{
		ID:                        p.ID(),
		NamespaceID:               p.NamespaceID(),
		Name:                      p.Name(),
		Description:               p.Description(),
		Tier:                      sqldb.ProviderTier(p.Tier()),
		DefaultProviderSourceAuth: p.UseProviderSourceAuth(),
		ProviderCategoryID:        p.CategoryID(),
		RepositoryID:              p.RepositoryID(),
		LatestVersionID:           p.LatestVersionID(),
	}
}

func toDBProviderVersion(v *provider.ProviderVersion) *sqldb.ProviderVersionDB {
	// Serialize protocol versions to JSON bytes
	protocolVersionsJSON, _ := json.Marshal(v.ProtocolVersions())

	return &sqldb.ProviderVersionDB{
		ID:               v.ID(),
		ProviderID:       v.ProviderID(),
		Version:          v.Version(),
		GitTag:           v.GitTag(),
		Beta:             v.Beta(),
		PublishedAt:      v.PublishedAt(),
		GPGKeyID:         v.GPGKeyID(), // Non-nullable in correct schema
		ExtractionVersion: nil,
		ProtocolVersions: protocolVersionsJSON, // []byte, not string
	}
}

func toDBProviderBinary(b *provider.ProviderBinary) *sqldb.ProviderVersionBinaryDB {
	return &sqldb.ProviderVersionBinaryDB{
		ID:                b.ID(),
		ProviderVersionID: b.VersionID(),
		Name:              b.FileName(),
		OperatingSystem:   sqldb.ProviderBinaryOperatingSystemType(b.OperatingSystem()),
		Architecture:      sqldb.ProviderBinaryArchitectureType(b.Architecture()),
		Checksum:          b.FileHash(), // Domain uses FileHash, DB uses Checksum
	}
}

func toDBGPGKey(k *provider.GPGKey) *sqldb.GPGKeyDB {
	// Handle nullable field conversions
	var keyID *string
	if keyIDVal := k.KeyID(); keyIDVal != "" {
		keyID = &keyIDVal
	}

	// Convert time.Time to *time.Time
	var createdAt, updatedAt *time.Time
	if createdVal := k.CreatedAt(); !createdVal.IsZero() {
		createdAt = &createdVal
	}
	if updatedVal := k.UpdatedAt(); !updatedVal.IsZero() {
		updatedAt = &updatedVal
	}

	return &sqldb.GPGKeyDB{
		ID:         k.ID(),
		NamespaceID: 0, // Will be set by caller (GPG keys belong to namespace, not provider)
		ASCIIArmor: []byte(k.AsciiArmor()),
		KeyID:      keyID,
		Fingerprint: nil, // Not in domain model
		Source:     nil, // Not in domain model
		SourceURL:  nil, // Not in domain model
		CreatedAt:  createdAt,
		UpdatedAt:  updatedAt,
	}
}

// Mapper functions (DB → domain)

func toDomainProvider(db *sqldb.ProviderDB) *provider.Provider {
	return provider.ReconstructProvider(
		db.ID,
		db.NamespaceID,
		db.Name,
		db.Description,
		string(db.Tier),
		db.ProviderCategoryID, // Field name: ProviderCategoryID
		db.RepositoryID,
		db.LatestVersionID,
		db.DefaultProviderSourceAuth, // Field name: DefaultProviderSourceAuth
	)
}

func toDomainProviderVersion(db *sqldb.ProviderVersionDB) *provider.ProviderVersion {
	// Parse protocol versions from JSON bytes
	var protocolVersions []string
	if len(db.ProtocolVersions) > 0 {
		if err := json.Unmarshal(db.ProtocolVersions, &protocolVersions); err != nil {
			// Fallback to default protocol
			protocolVersions = []string{"5.0"}
		}
	} else {
		protocolVersions = []string{"5.0"}
	}

	return provider.ReconstructProviderVersion(
		db.ID,
		db.ProviderID,
		db.Version,
		db.GitTag,
		db.Beta,
		db.PublishedAt,
		db.GPGKeyID, // Non-nullable in correct schema
		protocolVersions,
	)
}

func toDomainProviderBinary(db *sqldb.ProviderVersionBinaryDB) *provider.ProviderBinary {
	return provider.ReconstructProviderBinary(
		db.ID,
		db.ProviderVersionID,
		string(db.OperatingSystem),
		string(db.Architecture),
		db.Name,       // DB field: Name, domain: FileName
		0,            // FileSize - not stored in DB
		db.Checksum,   // DB field: Checksum, domain: FileHash
		"",           // DownloadURL - not stored in DB
	)
}

func toDomainGPGKey(db *sqldb.GPGKeyDB) *provider.GPGKey {
	// Handle nullable to non-nullable conversions
	var keyID string
	if db.KeyID != nil {
		keyID = *db.KeyID
	}

	var createdAt, updatedAt time.Time
	if db.CreatedAt != nil {
		createdAt = *db.CreatedAt
	}
	if db.UpdatedAt != nil {
		updatedAt = *db.UpdatedAt
	}

	return provider.ReconstructGPGKey(
		db.ID,
		"",      // KeyText - not in DB
		string(db.ASCIIArmor),
		keyID,
		nil,     // TrustSignature - not in DB
		createdAt,
		updatedAt,
	)
}
