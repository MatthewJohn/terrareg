package service

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/ProtonMail/go-crypto/openpgp/armor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/application/query"
	gpgkeyModel "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/gpgkey/model"
	gpgkeyRepo "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/gpgkey/repository"
	moduleModel "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/model"
	moduleRepo "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/repository"
	types "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared/types"
)

// Test constants
const (
	testNamespace = "test-namespace"
)

// Test GPG data generated at init
var (
	testKeyID          string
	testFingerprint    string
	testValidGPGKey    string
	testValidSignature string
)

func init() {
	// Generate a test GPG key and signature
	entity, err := openpgp.NewEntity("Test User", "test", "test@example.com", nil)
	if err != nil {
		panic(err)
	}

	testKeyID = entity.PrimaryKey.KeyIdString()
	testFingerprint = entity.PrimaryKey.KeyIdString()

	// Generate armored public key
	var pubKeyBuf bytes.Buffer
	encoder, err := armor.Encode(&pubKeyBuf, "PGP PUBLIC KEY BLOCK", nil)
	if err != nil {
		panic(err)
	}
	err = entity.SerializePrivate(encoder, nil)
	if err != nil {
		panic(err)
	}
	err = encoder.Close()
	if err != nil {
		panic(err)
	}
	testValidGPGKey = pubKeyBuf.String()

	// Generate a valid signature
	data := []byte("test data")
	var sigBuf bytes.Buffer
	sigEncoder, err := armor.Encode(&sigBuf, "PGP SIGNATURE", nil)
	if err != nil {
		panic(err)
	}
	err = openpgp.DetachSign(sigEncoder, entity, bytes.NewReader(data), nil)
	if err != nil {
		panic(err)
	}
	err = sigEncoder.Close()
	if err != nil {
		panic(err)
	}
	testValidSignature = sigBuf.String()
}

// mockGPGKeyRepository is a mock implementation of GPGKeyRepository
type mockGPGKeyRepository struct {
	saveFunc                      func(ctx context.Context, gpgKey *gpgkeyModel.GPGKey) error
	findByKeyIDFunc               func(ctx context.Context, keyID string) (*gpgkeyModel.GPGKey, error)
	findByNamespaceAndKeyIDFunc   func(ctx context.Context, namespaceName, keyID string) (*gpgkeyModel.GPGKey, error)
	findByNamespaceFunc           func(ctx context.Context, namespaceName string) ([]*gpgkeyModel.GPGKey, error)
	findMultipleByNamespacesFunc  func(ctx context.Context, namespaceNames []string) ([]*gpgkeyModel.GPGKey, error)
	existsByFingerprintFunc       func(ctx context.Context, fingerprint string) (bool, error)
	isInUseFunc                   func(ctx context.Context, keyID string) (bool, error)
	deleteByNamespaceAndKeyIDFunc func(ctx context.Context, namespaceName, keyID string) error
}

func (m *mockGPGKeyRepository) Save(ctx context.Context, gpgKey *gpgkeyModel.GPGKey) error {
	if m.saveFunc != nil {
		return m.saveFunc(ctx, gpgKey)
	}
	return nil
}

func (m *mockGPGKeyRepository) FindByID(ctx context.Context, id int) (*gpgkeyModel.GPGKey, error) {
	return nil, nil
}

func (m *mockGPGKeyRepository) FindByKeyID(ctx context.Context, keyID string) (*gpgkeyModel.GPGKey, error) {
	if m.findByKeyIDFunc != nil {
		return m.findByKeyIDFunc(ctx, keyID)
	}
	return nil, nil
}

func (m *mockGPGKeyRepository) FindByFingerprint(ctx context.Context, fingerprint string) (*gpgkeyModel.GPGKey, error) {
	return nil, nil
}

func (m *mockGPGKeyRepository) FindByNamespace(ctx context.Context, namespaceName string) ([]*gpgkeyModel.GPGKey, error) {
	if m.findByNamespaceFunc != nil {
		return m.findByNamespaceFunc(ctx, namespaceName)
	}
	return []*gpgkeyModel.GPGKey{}, nil
}

func (m *mockGPGKeyRepository) FindByNamespaceAndKeyID(ctx context.Context, namespaceName, keyID string) (*gpgkeyModel.GPGKey, error) {
	if m.findByNamespaceAndKeyIDFunc != nil {
		return m.findByNamespaceAndKeyIDFunc(ctx, namespaceName, keyID)
	}
	return nil, nil
}

func (m *mockGPGKeyRepository) FindMultipleByNamespaces(ctx context.Context, namespaceNames []string) ([]*gpgkeyModel.GPGKey, error) {
	if m.findMultipleByNamespacesFunc != nil {
		return m.findMultipleByNamespacesFunc(ctx, namespaceNames)
	}
	return []*gpgkeyModel.GPGKey{}, nil
}

func (m *mockGPGKeyRepository) Delete(ctx context.Context, id int) error {
	return nil
}

func (m *mockGPGKeyRepository) DeleteByNamespaceAndKeyID(ctx context.Context, namespaceName, keyID string) error {
	if m.deleteByNamespaceAndKeyIDFunc != nil {
		return m.deleteByNamespaceAndKeyIDFunc(ctx, namespaceName, keyID)
	}
	return nil
}

func (m *mockGPGKeyRepository) ExistsByFingerprint(ctx context.Context, fingerprint string) (bool, error) {
	if m.existsByFingerprintFunc != nil {
		return m.existsByFingerprintFunc(ctx, fingerprint)
	}
	return false, nil
}

func (m *mockGPGKeyRepository) IsInUse(ctx context.Context, keyID string) (bool, error) {
	if m.isInUseFunc != nil {
		return m.isInUseFunc(ctx, keyID)
	}
	return false, nil
}

func (m *mockGPGKeyRepository) GetUsedByVersionCount(ctx context.Context, keyID string) (int, error) {
	return 0, nil
}

func (m *mockGPGKeyRepository) FindAll(ctx context.Context) ([]*gpgkeyModel.GPGKey, error) {
	return []*gpgkeyModel.GPGKey{}, nil
}

var _ gpgkeyRepo.GPGKeyRepository = (*mockGPGKeyRepository)(nil)

// mockNamespaceRepository is a mock implementation of NamespaceRepository
type mockNamespaceRepository struct {
	findByNameFunc func(ctx context.Context, name types.NamespaceName) (*moduleModel.Namespace, error)
}

func (m *mockNamespaceRepository) Save(ctx context.Context, namespace *moduleModel.Namespace) error {
	return nil
}

func (m *mockNamespaceRepository) FindByID(ctx context.Context, id int) (*moduleModel.Namespace, error) {
	return moduleModel.ReconstructNamespace(id, types.NamespaceName(testNamespace), nil, moduleModel.NamespaceTypeNone), nil
}

func (m *mockNamespaceRepository) FindByName(ctx context.Context, name types.NamespaceName) (*moduleModel.Namespace, error) {
	if m.findByNameFunc != nil {
		return m.findByNameFunc(ctx, name)
	}
	return moduleModel.ReconstructNamespace(1, name, nil, moduleModel.NamespaceTypeNone), nil
}

func (m *mockNamespaceRepository) List(ctx context.Context, opts *query.ListOptions) ([]*moduleModel.Namespace, int, error) {
	return []*moduleModel.Namespace{
		moduleModel.ReconstructNamespace(1, types.NamespaceName(testNamespace), nil, moduleModel.NamespaceTypeNone),
	}, 1, nil
}

func (m *mockNamespaceRepository) Delete(ctx context.Context, id int) error {
	return nil
}

func (m *mockNamespaceRepository) Exists(ctx context.Context, name types.NamespaceName) (bool, error) {
	return true, nil
}

var _ moduleRepo.NamespaceRepository = (*mockNamespaceRepository)(nil)

// createTestGPGKey creates a test GPG key entity
func createTestGPGKey(namespaceID int) *gpgkeyModel.GPGKey {
	key, err := gpgkeyModel.NewGPGKey(namespaceID, testValidGPGKey, testKeyID, testFingerprint)
	if err != nil {
		panic(err)
	}
	key.SetID(1)
	key.SetNamespace(gpgkeyModel.NewNamespace(namespaceID, types.NamespaceName(testNamespace)))
	return key
}

func TestNewGPGKeyService(t *testing.T) {
	gpgRepo := &mockGPGKeyRepository{}
	nsRepo := &mockNamespaceRepository{}

	service := NewGPGKeyService(gpgRepo, nsRepo)

	assert.NotNil(t, service)
}

func TestGPGKeyService_CreateGPGKey(t *testing.T) {
	tests := []struct {
		name       string
		namespace  string
		asciiArmor string
		trustSig   *string
		source     *string
		sourceURL  *string
		setupRepo  func(*mockGPGKeyRepository)
		setupNs    func(*mockNamespaceRepository)
		wantErr    error
	}{
		{
			name:       "creates key successfully",
			namespace:  testNamespace,
			asciiArmor: testValidGPGKey,
			setupNs: func(ns *mockNamespaceRepository) {
				ns.findByNameFunc = func(ctx context.Context, name types.NamespaceName) (*moduleModel.Namespace, error) {
					return moduleModel.ReconstructNamespace(1, name, nil, moduleModel.NamespaceTypeNone), nil
				}
			},
			setupRepo: func(repo *mockGPGKeyRepository) {
				repo.existsByFingerprintFunc = func(ctx context.Context, fingerprint string) (bool, error) {
					return false, nil
				}
			},
			wantErr: nil,
		},
		{
			name:       "fails with namespace not found",
			namespace:  "nonexistent",
			asciiArmor: testValidGPGKey,
			setupNs: func(ns *mockNamespaceRepository) {
				ns.findByNameFunc = func(ctx context.Context, name types.NamespaceName) (*moduleModel.Namespace, error) {
					return nil, nil
				}
			},
			wantErr: errors.New("namespace 'nonexistent' does not exist"),
		},
		{
			name:       "fails with invalid ASCII armor",
			namespace:  testNamespace,
			asciiArmor: "not a valid gpg key",
			setupNs: func(ns *mockNamespaceRepository) {
				ns.findByNameFunc = func(ctx context.Context, name types.NamespaceName) (*moduleModel.Namespace, error) {
					return moduleModel.ReconstructNamespace(1, name, nil, moduleModel.NamespaceTypeNone), nil
				}
			},
			wantErr: gpgkeyModel.ErrInvalidASCIIArmor,
		},
		{
			name:       "fails with duplicate fingerprint",
			namespace:  testNamespace,
			asciiArmor: testValidGPGKey,
			setupNs: func(ns *mockNamespaceRepository) {
				ns.findByNameFunc = func(ctx context.Context, name types.NamespaceName) (*moduleModel.Namespace, error) {
					return moduleModel.ReconstructNamespace(1, name, nil, moduleModel.NamespaceTypeNone), nil
				}
			},
			setupRepo: func(repo *mockGPGKeyRepository) {
				repo.existsByFingerprintFunc = func(ctx context.Context, fingerprint string) (bool, error) {
					return true, nil
				}
			},
			wantErr: gpgkeyModel.ErrDuplicateFingerprint,
		},
		{
			name:       "creates key with optional fields",
			namespace:  testNamespace,
			asciiArmor: testValidGPGKey,
			trustSig:   stringPtr("trust-sig"),
			source:     stringPtr("manual"),
			sourceURL:  stringPtr("https://example.com/key"),
			setupNs: func(ns *mockNamespaceRepository) {
				ns.findByNameFunc = func(ctx context.Context, name types.NamespaceName) (*moduleModel.Namespace, error) {
					return moduleModel.ReconstructNamespace(1, name, nil, moduleModel.NamespaceTypeNone), nil
				}
			},
			setupRepo: func(repo *mockGPGKeyRepository) {
				repo.existsByFingerprintFunc = func(ctx context.Context, fingerprint string) (bool, error) {
					return false, nil
				}
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gpgRepo := &mockGPGKeyRepository{}
			nsRepo := &mockNamespaceRepository{}
			if tt.setupRepo != nil {
				tt.setupRepo(gpgRepo)
			}
			if tt.setupNs != nil {
				tt.setupNs(nsRepo)
			}
			service := NewGPGKeyService(gpgRepo, nsRepo)

			req := CreateGPGKeyRequest{
				Namespace:      tt.namespace,
				ASCIILArmor:    tt.asciiArmor,
				TrustSignature: tt.trustSig,
				Source:         tt.source,
				SourceURL:      tt.sourceURL,
			}

			key, err := service.CreateGPGKey(context.Background(), req)

			if tt.wantErr != nil {
				require.Error(t, err)
				if errors.Is(tt.wantErr, gpgkeyModel.ErrInvalidASCIIArmor) || errors.Is(tt.wantErr, gpgkeyModel.ErrDuplicateFingerprint) {
					assert.ErrorIs(t, err, tt.wantErr)
				} else {
					assert.Contains(t, err.Error(), tt.wantErr.Error())
				}
				assert.Nil(t, key)
			} else {
				require.NoError(t, err)
				require.NotNil(t, key)
				assert.Equal(t, tt.asciiArmor, key.ASCIIArmor())
				assert.NotEmpty(t, key.KeyID())
				assert.NotEmpty(t, key.Fingerprint())
				if tt.trustSig != nil {
					assert.Equal(t, tt.trustSig, key.TrustSignature())
				}
				if tt.source != nil {
					assert.Equal(t, *tt.source, key.Source())
				}
				if tt.sourceURL != nil {
					assert.Equal(t, tt.sourceURL, key.SourceURL())
				}
			}
		})
	}
}

func TestGPGKeyService_GetNamespaceGPGKeys(t *testing.T) {
	gpgRepo := &mockGPGKeyRepository{
		findByNamespaceFunc: func(ctx context.Context, namespaceName string) ([]*gpgkeyModel.GPGKey, error) {
			key1 := createTestGPGKey(1)
			key2 := createTestGPGKey(1)
			return []*gpgkeyModel.GPGKey{key1, key2}, nil
		},
	}
	nsRepo := &mockNamespaceRepository{
		findByNameFunc: func(ctx context.Context, name types.NamespaceName) (*moduleModel.Namespace, error) {
			return moduleModel.ReconstructNamespace(1, name, nil, moduleModel.NamespaceTypeNone), nil
		},
	}
	service := NewGPGKeyService(gpgRepo, nsRepo)

	keys, err := service.GetNamespaceGPGKeys(context.Background(), testNamespace)

	require.NoError(t, err)
	assert.Len(t, keys, 2)
}

func TestGPGKeyService_GetNamespaceGPGKeys_NamespaceNotFound(t *testing.T) {
	gpgRepo := &mockGPGKeyRepository{}
	nsRepo := &mockNamespaceRepository{
		findByNameFunc: func(ctx context.Context, name types.NamespaceName) (*moduleModel.Namespace, error) {
			return nil, nil
		},
	}
	service := NewGPGKeyService(gpgRepo, nsRepo)

	_, err := service.GetNamespaceGPGKeys(context.Background(), "nonexistent")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "does not exist")
}

func TestGPGKeyService_GetGPGKey(t *testing.T) {
	tests := []struct {
		name     string
		keyID    string
		setup    func(*mockGPGKeyRepository)
		wantErr  error
	}{
		{
			name:  "gets key successfully",
			keyID: testKeyID,
			setup: func(repo *mockGPGKeyRepository) {
				repo.findByNamespaceAndKeyIDFunc = func(ctx context.Context, namespaceName, keyID string) (*gpgkeyModel.GPGKey, error) {
					return createTestGPGKey(1), nil
				}
			},
			wantErr: nil,
		},
		{
			name:  "fails when key not found",
			keyID: "nonexistent",
			setup: func(repo *mockGPGKeyRepository) {
				repo.findByNamespaceAndKeyIDFunc = func(ctx context.Context, namespaceName, keyID string) (*gpgkeyModel.GPGKey, error) {
					return nil, nil
				}
			},
			wantErr: gpgkeyModel.ErrGPGKeyNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gpgRepo := &mockGPGKeyRepository{}
			tt.setup(gpgRepo)
			nsRepo := &mockNamespaceRepository{}
			service := NewGPGKeyService(gpgRepo, nsRepo)

			key, err := service.GetGPGKey(context.Background(), testNamespace, tt.keyID)

			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, key)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, key)
				assert.Equal(t, tt.keyID, key.KeyID())
			}
		})
	}
}

func TestGPGKeyService_DeleteGPGKey(t *testing.T) {
	tests := []struct {
		name    string
		keyID   string
		setup   func(*mockGPGKeyRepository)
		wantErr error
	}{
		{
			name:  "deletes key successfully",
			keyID: testKeyID,
			setup: func(repo *mockGPGKeyRepository) {
				repo.findByNamespaceAndKeyIDFunc = func(ctx context.Context, namespaceName, keyID string) (*gpgkeyModel.GPGKey, error) {
					return createTestGPGKey(1), nil
				}
				repo.isInUseFunc = func(ctx context.Context, keyID string) (bool, error) {
					return false, nil
				}
				repo.deleteByNamespaceAndKeyIDFunc = func(ctx context.Context, namespaceName, keyID string) error {
					return nil
				}
			},
			wantErr: nil,
		},
		{
			name:  "fails when key not found",
			keyID: "nonexistent",
			setup: func(repo *mockGPGKeyRepository) {
				repo.findByNamespaceAndKeyIDFunc = func(ctx context.Context, namespaceName, keyID string) (*gpgkeyModel.GPGKey, error) {
					return nil, nil
				}
			},
			wantErr: gpgkeyModel.ErrGPGKeyNotFound,
		},
		{
			name:  "fails when key is in use",
			keyID: testKeyID,
			setup: func(repo *mockGPGKeyRepository) {
				repo.findByNamespaceAndKeyIDFunc = func(ctx context.Context, namespaceName, keyID string) (*gpgkeyModel.GPGKey, error) {
					return createTestGPGKey(1), nil
				}
				repo.isInUseFunc = func(ctx context.Context, keyID string) (bool, error) {
					return true, nil
				}
			},
			wantErr: gpgkeyModel.ErrGPGKeyInUse,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gpgRepo := &mockGPGKeyRepository{}
			tt.setup(gpgRepo)
			nsRepo := &mockNamespaceRepository{}
			service := NewGPGKeyService(gpgRepo, nsRepo)

			err := service.DeleteGPGKey(context.Background(), testNamespace, tt.keyID)

			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestGPGKeyService_VerifySignature(t *testing.T) {
	tests := []struct {
		name   string
		keyID  string
		sig    string
		data   string
		setup  func(*mockGPGKeyRepository)
		valid  bool
		wantErr bool
	}{
		{
			name:  "verifies valid signature",
			keyID: testKeyID,
			sig:   testValidSignature,
			data:  "test data",
			setup: func(repo *mockGPGKeyRepository) {
				repo.findByKeyIDFunc = func(ctx context.Context, keyID string) (*gpgkeyModel.GPGKey, error) {
					key, err := gpgkeyModel.NewGPGKey(1, testValidGPGKey, testKeyID, testFingerprint)
					require.NoError(t, err)
					return key, nil
				}
			},
			valid:   true,
			wantErr: false,
		},
		{
			name:  "returns false for invalid signature",
			keyID: testKeyID,
			sig:   testValidSignature,
			data:  "different data",
			setup: func(repo *mockGPGKeyRepository) {
				repo.findByKeyIDFunc = func(ctx context.Context, keyID string) (*gpgkeyModel.GPGKey, error) {
					key, err := gpgkeyModel.NewGPGKey(1, testValidGPGKey, testKeyID, testFingerprint)
					require.NoError(t, err)
					return key, nil
				}
			},
			valid:   false,
			wantErr: false,
		},
		{
			name:  "fails when key not found",
			keyID: "nonexistent",
			sig:   testValidSignature,
			data:  "test data",
			setup: func(repo *mockGPGKeyRepository) {
				repo.findByKeyIDFunc = func(ctx context.Context, keyID string) (*gpgkeyModel.GPGKey, error) {
					return nil, nil
				}
			},
			valid:   false,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gpgRepo := &mockGPGKeyRepository{}
			tt.setup(gpgRepo)
			nsRepo := &mockNamespaceRepository{}
			service := NewGPGKeyService(gpgRepo, nsRepo)

			valid, err := service.VerifySignature(context.Background(), tt.keyID, tt.sig, tt.data)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.valid, valid)
			}
		})
	}
}

func TestGPGKeyService_GetMultipleNamespaceGPGKeys(t *testing.T) {
	gpgRepo := &mockGPGKeyRepository{
		findMultipleByNamespacesFunc: func(ctx context.Context, namespaceNames []string) ([]*gpgkeyModel.GPGKey, error) {
			key1 := createTestGPGKey(1)
			key2 := createTestGPGKey(2)
			return []*gpgkeyModel.GPGKey{key1, key2}, nil
		},
	}
	nsRepo := &mockNamespaceRepository{}
	service := NewGPGKeyService(gpgRepo, nsRepo)

	keys, err := service.GetMultipleNamespaceGPGKeys(context.Background(), []string{"ns1", "ns2"})

	require.NoError(t, err)
	assert.Len(t, keys, 2)
}

func TestGPGKeyService_GetMultipleNamespaceGPGKeys_EmptyList(t *testing.T) {
	gpgRepo := &mockGPGKeyRepository{}
	nsRepo := &mockNamespaceRepository{}
	service := NewGPGKeyService(gpgRepo, nsRepo)

	keys, err := service.GetMultipleNamespaceGPGKeys(context.Background(), []string{})

	require.NoError(t, err)
	assert.Empty(t, keys)
}

func stringPtr(s string) *string {
	return &s
}
