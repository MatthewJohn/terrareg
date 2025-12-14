package gpgkey

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	gpgkeyModel "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/gpgkey/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/gpgkey/repository"
)

// gpgKeyRepositoryImpl implements the GPGKeyRepository interface using GORM
type gpgKeyRepositoryImpl struct {
	db *gorm.DB
}

// NewGPGKeyRepository creates a new GPG key repository
func NewGPGKeyRepository(db *gorm.DB) repository.GPGKeyRepository {
	return &gpgKeyRepositoryImpl{
		db: db,
	}
}

// FindByID finds a GPG key by its ID
func (r *gpgKeyRepositoryImpl) FindByID(ctx context.Context, id int) (*gpgkeyModel.GPGKey, error) {
	var gpgKeyDB GPGKeyDBModel

	err := r.db.WithContext(ctx).
		Preload("Namespace").
		First(&gpgKeyDB, id).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // Not found is not an error
		}
		return nil, fmt.Errorf("failed to find GPG key by ID: %w", err)
	}

	return r.dbModelToDomain(&gpgKeyDB), nil
}

// FindByKeyID finds a GPG key by its key ID
func (r *gpgKeyRepositoryImpl) FindByKeyID(ctx context.Context, keyID string) (*gpgkeyModel.GPGKey, error) {
	var gpgKeyDB GPGKeyDBModel

	err := r.db.WithContext(ctx).
		Preload("Namespace").
		Where("key_id = ?", keyID).
		First(&gpgKeyDB).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // Not found is not an error
		}
		return nil, fmt.Errorf("failed to find GPG key by key ID: %w", err)
	}

	return r.dbModelToDomain(&gpgKeyDB), nil
}

// FindByFingerprint finds a GPG key by its fingerprint
func (r *gpgKeyRepositoryImpl) FindByFingerprint(ctx context.Context, fingerprint string) (*gpgkeyModel.GPGKey, error) {
	var gpgKeyDB GPGKeyDBModel

	err := r.db.WithContext(ctx).
		Preload("Namespace").
		Where("fingerprint = ?", fingerprint).
		First(&gpgKeyDB).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // Not found is not an error
		}
		return nil, fmt.Errorf("failed to find GPG key by fingerprint: %w", err)
	}

	return r.dbModelToDomain(&gpgKeyDB), nil
}

// FindByNamespace finds all GPG keys for a namespace
func (r *gpgKeyRepositoryImpl) FindByNamespace(ctx context.Context, namespaceName string) ([]*gpgkeyModel.GPGKey, error) {
	var gpgKeyDBs []GPGKeyDBModel

	err := r.db.WithContext(ctx).
		Preload("Namespace").
		Joins("JOIN namespace ON gpg_key.namespace_id = namespace.id").
		Where("namespace.name = ?", namespaceName).
		Order("gpg_key.created_at DESC").
		Find(&gpgKeyDBs).Error

	if err != nil {
		return nil, fmt.Errorf("failed to find GPG keys by namespace: %w", err)
	}

	return r.dbModelsToDomain(gpgKeyDBs), nil
}

// FindByNamespaceAndKeyID finds a GPG key by namespace and key ID
func (r *gpgKeyRepositoryImpl) FindByNamespaceAndKeyID(ctx context.Context, namespaceName, keyID string) (*gpgkeyModel.GPGKey, error) {
	var gpgKeyDB GPGKeyDBModel

	err := r.db.WithContext(ctx).
		Preload("Namespace").
		Joins("JOIN namespace ON gpg_key.namespace_id = namespace.id").
		Where("namespace.name = ? AND gpg_key.key_id = ?", namespaceName, keyID).
		First(&gpgKeyDB).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // Not found is not an error
		}
		return nil, fmt.Errorf("failed to find GPG key by namespace and key ID: %w", err)
	}

	return r.dbModelToDomain(&gpgKeyDB), nil
}

// FindMultipleByNamespaces finds all GPG keys for multiple namespaces
func (r *gpgKeyRepositoryImpl) FindMultipleByNamespaces(ctx context.Context, namespaceNames []string) ([]*gpgkeyModel.GPGKey, error) {
	if len(namespaceNames) == 0 {
		return []*gpgkeyModel.GPGKey{}, nil
	}

	var gpgKeyDBs []GPGKeyDBModel

	err := r.db.WithContext(ctx).
		Preload("Namespace").
		Joins("JOIN namespace ON gpg_key.namespace_id = namespace.id").
		Where("namespace.name IN ?", namespaceNames).
		Order("gpg_key.created_at DESC").
		Find(&gpgKeyDBs).Error

	if err != nil {
		return nil, fmt.Errorf("failed to find GPG keys by multiple namespaces: %w", err)
	}

	return r.dbModelsToDomain(gpgKeyDBs), nil
}

// Save saves a GPG key (creates if new, updates if existing)
func (r *gpgKeyRepositoryImpl) Save(ctx context.Context, gpgKey *gpgkeyModel.GPGKey) error {
	gpgKeyDB := r.domainToDBModel(gpgKey)

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if gpgKey.ID() == 0 {
			// Create new GPG key
			if err := tx.Create(gpgKeyDB).Error; err != nil {
				return fmt.Errorf("failed to create GPG key: %w", err)
			}
			gpgKey.SetID(gpgKeyDB.ID)
		} else {
			// Update existing GPG key
			if err := tx.Save(gpgKeyDB).Error; err != nil {
				return fmt.Errorf("failed to update GPG key: %w", err)
			}
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to save GPG key: %w", err)
	}

	return nil
}

// Delete deletes a GPG key by its ID
func (r *gpgKeyRepositoryImpl) Delete(ctx context.Context, id int) error {
	result := r.db.WithContext(ctx).Delete(&GPGKeyDBModel{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete GPG key: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("GPG key not found")
	}
	return nil
}

// DeleteByNamespaceAndKeyID deletes a GPG key by namespace and key ID
func (r *gpgKeyRepositoryImpl) DeleteByNamespaceAndKeyID(ctx context.Context, namespaceName, keyID string) error {
	result := r.db.WithContext(ctx).
		Joins("JOIN namespace ON gpg_key.namespace_id = namespace.id").
		Where("namespace.name = ? AND gpg_key.key_id = ?", namespaceName, keyID).
		Delete(&GPGKeyDBModel{})

	if result.Error != nil {
		return fmt.Errorf("failed to delete GPG key: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("GPG key not found")
	}
	return nil
}

// ExistsByFingerprint checks if a GPG key with the given fingerprint exists
func (r *gpgKeyRepositoryImpl) ExistsByFingerprint(ctx context.Context, fingerprint string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&GPGKeyDBModel{}).
		Where("fingerprint = ?", fingerprint).
		Count(&count).Error

	if err != nil {
		return false, fmt.Errorf("failed to check if GPG key exists by fingerprint: %w", err)
	}

	return count > 0, nil
}

// IsInUse checks if a GPG key is in use by any provider versions
func (r *gpgKeyRepositoryImpl) IsInUse(ctx context.Context, keyID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Table("provider_versions").
		Joins("JOIN gpg_key ON provider_versions.gpg_key_id = gpg_key.id").
		Where("gpg_key.key_id = ?", keyID).
		Count(&count).Error

	if err != nil {
		return false, fmt.Errorf("failed to check if GPG key is in use: %w", err)
	}

	return count > 0, nil
}

// GetUsedByVersionCount returns the number of provider versions using this GPG key
func (r *gpgKeyRepositoryImpl) GetUsedByVersionCount(ctx context.Context, keyID string) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Table("provider_versions").
		Joins("JOIN gpg_key ON provider_versions.gpg_key_id = gpg_key.id").
		Where("gpg_key.key_id = ?", keyID).
		Count(&count).Error

	if err != nil {
		return 0, fmt.Errorf("failed to get version count for GPG key: %w", err)
	}

	return int(count), nil
}

// FindAll finds all GPG keys (admin function)
func (r *gpgKeyRepositoryImpl) FindAll(ctx context.Context) ([]*gpgkeyModel.GPGKey, error) {
	var gpgKeyDBs []GPGKeyDBModel

	err := r.db.WithContext(ctx).
		Preload("Namespace").
		Order("gpg_key.created_at DESC").
		Find(&gpgKeyDBs).Error

	if err != nil {
		return nil, fmt.Errorf("failed to find all GPG keys: %w", err)
	}

	return r.dbModelsToDomain(gpgKeyDBs), nil
}

// Helper methods to convert between domain and DB models

func (r *gpgKeyRepositoryImpl) dbModelToDomain(gpgKeyDB *GPGKeyDBModel) *gpgkeyModel.GPGKey {
	gpgKey, _ := gpgkeyModel.NewGPGKey(
		gpgKeyDB.NamespaceID,
		gpgKeyDB.ASCIILArmor,
		gpgKeyDB.KeyID,
		gpgKeyDB.Fingerprint,
	)

	gpgKey.SetID(gpgKeyDB.ID)
	gpgKey.SetSource(gpgKeyDB.Source)
	gpgKey.SetSourceURL(gpgKeyDB.SourceURL)
	gpgKey.SetTrustSignature(gpgKeyDB.TrustSignature)

	// Set namespace entity
	namespace := gpgkeyModel.NewNamespace(gpgKeyDB.Namespace.ID, gpgKeyDB.Namespace.Name)
	gpgKey.SetNamespace(namespace)

	return gpgKey
}

func (r *gpgKeyRepositoryImpl) domainToDBModel(gpgKey *gpgkeyModel.GPGKey) *GPGKeyDBModel {
	gpgKeyDB := &GPGKeyDBModel{
		ID:             gpgKey.ID(),
		NamespaceID:    gpgKey.NamespaceID(),
		ASCIILArmor:    gpgKey.ASCIIArmor(),
		KeyID:          gpgKey.KeyID(),
		Fingerprint:    gpgKey.Fingerprint(),
		Source:         gpgKey.Source(),
		SourceURL:      gpgKey.SourceURL(),
		TrustSignature: gpgKey.TrustSignature(),
		CreatedAt:      gpgKey.CreatedAt(),
		UpdatedAt:      gpgKey.UpdatedAt(),
	}

	return gpgKeyDB
}

func (r *gpgKeyRepositoryImpl) dbModelsToDomain(gpgKeyDBs []GPGKeyDBModel) []*gpgkeyModel.GPGKey {
	gpgKeys := make([]*gpgkeyModel.GPGKey, len(gpgKeyDBs))
	for i, gpgKeyDB := range gpgKeyDBs {
		gpgKeys[i] = r.dbModelToDomain(&gpgKeyDB)
	}
	return gpgKeys
}
