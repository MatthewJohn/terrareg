package repository

import (
	"context"

	gpgkeyModel "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/gpgkey/model"
)

// GPGKeyRepository defines the interface for GPG key persistence operations
type GPGKeyRepository interface {
	// FindByID finds a GPG key by its ID
	FindByID(ctx context.Context, id int) (*gpgkeyModel.GPGKey, error)

	// FindByKeyID finds a GPG key by its key ID
	FindByKeyID(ctx context.Context, keyID string) (*gpgkeyModel.GPGKey, error)

	// FindByFingerprint finds a GPG key by its fingerprint
	FindByFingerprint(ctx context.Context, fingerprint string) (*gpgkeyModel.GPGKey, error)

	// FindByNamespace finds all GPG keys for a namespace
	FindByNamespace(ctx context.Context, namespaceName string) ([]*gpgkeyModel.GPGKey, error)

	// FindByNamespaceAndKeyID finds a GPG key by namespace and key ID
	FindByNamespaceAndKeyID(ctx context.Context, namespaceName, keyID string) (*gpgkeyModel.GPGKey, error)

	// FindMultipleByNamespaces finds all GPG keys for multiple namespaces
	FindMultipleByNamespaces(ctx context.Context, namespaceNames []string) ([]*gpgkeyModel.GPGKey, error)

	// Save saves a GPG key (creates if new, updates if existing)
	Save(ctx context.Context, gpgKey *gpgkeyModel.GPGKey) error

	// Delete deletes a GPG key by its ID
	Delete(ctx context.Context, id int) error

	// DeleteByNamespaceAndKeyID deletes a GPG key by namespace and key ID
	DeleteByNamespaceAndKeyID(ctx context.Context, namespaceName, keyID string) error

	// ExistsByFingerprint checks if a GPG key with the given fingerprint exists
	ExistsByFingerprint(ctx context.Context, fingerprint string) (bool, error)

	// IsInUse checks if a GPG key is in use by any provider versions
	IsInUse(ctx context.Context, keyID string) (bool, error)

	// GetUsedByVersionCount returns the number of provider versions using this GPG key
	GetUsedByVersionCount(ctx context.Context, keyID string) (int, error)

	// FindAll finds all GPG keys (admin function)
	FindAll(ctx context.Context) ([]*gpgkeyModel.GPGKey, error)
}
