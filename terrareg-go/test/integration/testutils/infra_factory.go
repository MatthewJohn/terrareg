package testutils

import (
	"testing"

	analyticsCmd "github.com/matthewjohn/terrareg/terrareg-go/internal/application/command/analytics"
	authRepository "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth/repository"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/config/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/repository"
	moduleService "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/service"
	providerrepo "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider/repository"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
	sqldbAnalytics "github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb/analytics"
	sqldbAuth "github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb/auth"
	sqldbModule "github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb/module"
	sqldbProvider "github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb/provider"
	"github.com/stretchr/testify/require"
)

// ConfigOption modifies DomainConfig during test setup
type ConfigOption func(*model.DomainConfig)

// WithTrustedNamespaces sets trusted namespaces for tests
func WithTrustedNamespaces(namespaces ...string) ConfigOption {
	return func(cfg *model.DomainConfig) {
		cfg.TrustedNamespaces = namespaces
	}
}

// WithVerifiedNamespaces sets verified module namespaces for tests
func WithVerifiedNamespaces(namespaces ...string) ConfigOption {
	return func(cfg *model.DomainConfig) {
		cfg.VerifiedModuleNamespaces = namespaces
	}
}

// WithModuleHostingMode sets module hosting mode for tests
func WithModuleHostingMode(mode model.ModuleHostingMode) ConfigOption {
	return func(cfg *model.DomainConfig) {
		cfg.AllowModuleHosting = mode
	}
}

// WithSecretKeySet sets whether secret key is configured
func WithSecretKeySet(set bool) ConfigOption {
	return func(cfg *model.DomainConfig) {
		cfg.SecretKeySet = set
	}
}

// WithOpenIDConnectEnabled sets whether OpenID Connect is enabled
func WithOpenIDConnectEnabled(enabled bool) ConfigOption {
	return func(cfg *model.DomainConfig) {
		cfg.OpenIDConnectEnabled = enabled
	}
}

// WithSAMLEnabled sets whether SAML is enabled
func WithSAMLEnabled(enabled bool) ConfigOption {
	return func(cfg *model.DomainConfig) {
		cfg.SAMLEnabled = enabled
	}
}

// WithAdminLoginEnabled sets whether admin login is enabled
func WithAdminLoginEnabled(enabled bool) ConfigOption {
	return func(cfg *model.DomainConfig) {
		cfg.AdminLoginEnabled = enabled
	}
}

// CreateTestDomainConfigWith creates a test DomainConfig with optional overrides
func CreateTestDomainConfigWith(t *testing.T, opts ...ConfigOption) *model.DomainConfig {
	cfg := CreateTestDomainConfig(t) // Start with base config
	for _, opt := range opts {
		opt(cfg)
	}
	return cfg
}

// CreateModuleProviderRepository creates a ModuleProviderRepository with proper config
func CreateModuleProviderRepository(t *testing.T, db *sqldb.Database, opts ...ConfigOption) repository.ModuleProviderRepository {
	cfg := CreateTestDomainConfigWith(t, opts...)
	nsRepo := sqldbModule.NewNamespaceRepository(db.DB)
	repo, err := sqldbModule.NewModuleProviderRepository(db.DB, nsRepo, cfg)
	require.NoError(t, err, "Failed to create ModuleProviderRepository")
	return repo
}

// CreateNamespaceRepository creates a NamespaceRepository
func CreateNamespaceRepository(t *testing.T, db *sqldb.Database) repository.NamespaceRepository {
	return sqldbModule.NewNamespaceRepository(db.DB)
}

// CreateModuleVersionRepository creates a ModuleVersionRepository
func CreateModuleVersionRepository(t *testing.T, db *sqldb.Database) repository.ModuleVersionRepository {
	repo, err := sqldbModule.NewModuleVersionRepository(db.DB)
	require.NoError(t, err, "Failed to create ModuleVersionRepository")
	return repo
}

// CreateUserGroupRepository creates a UserGroupRepository
func CreateUserGroupRepository(t *testing.T, db *sqldb.Database) authRepository.UserGroupRepository {
	repo, err := sqldbAuth.NewUserGroupRepository(db.DB)
	require.NoError(t, err, "Failed to create UserGroupRepository")
	return repo
}

// TestRepositories holds all common repositories with consistent config
type TestRepositories struct {
	Namespace      repository.NamespaceRepository
	ModuleProvider repository.ModuleProviderRepository
	ModuleVersion  repository.ModuleVersionRepository
	UserGroup      authRepository.UserGroupRepository
	Analytics      analyticsCmd.AnalyticsRepository
	Provider       providerrepo.ProviderRepository
}

// CreateTestRepositories creates all common repositories with consistent config
func CreateTestRepositories(t *testing.T, db *sqldb.Database, opts ...ConfigOption) *TestRepositories {
	cfg := CreateTestDomainConfigWith(t, opts...)

	// Create namespace repository (no config needed)
	nsRepo := sqldbModule.NewNamespaceRepository(db.DB)

	// Create module provider repository with config
	mpRepo, err := sqldbModule.NewModuleProviderRepository(db.DB, nsRepo, cfg)
	require.NoError(t, err, "Failed to create ModuleProviderRepository")

	// Create module version repository
	mvRepo, err := sqldbModule.NewModuleVersionRepository(db.DB)
	require.NoError(t, err, "Failed to create ModuleVersionRepository")

	// Create user group repository
	ugRepo, err := sqldbAuth.NewUserGroupRepository(db.DB)
	require.NoError(t, err, "Failed to create UserGroupRepository")

	// Create namespace service (needed for analytics repository)
	namespaceSvc := moduleService.NewNamespaceService(cfg)

	// Create analytics repository
	analyticsRepoImpl, err := sqldbAnalytics.NewAnalyticsRepository(db.DB, nsRepo, namespaceSvc)
	require.NoError(t, err, "Failed to create AnalyticsRepository")

	// Create provider repository
	providerRepo := sqldbProvider.NewProviderRepository(db.DB)

	return &TestRepositories{
		Namespace:      nsRepo,
		ModuleProvider: mpRepo,
		ModuleVersion:  mvRepo,
		UserGroup:      ugRepo,
		Analytics:      analyticsRepoImpl,
		Provider:       providerRepo,
	}
}
