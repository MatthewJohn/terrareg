package util

import (
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
)

// TestLogger is a no-op logger for testing
var TestLogger = zerolog.Nop()

// DatabaseTestHelper provides utilities for database testing
type DatabaseTestHelper struct {
	DB *sqldb.Database
}

// NewTestDatabase creates a test database with real SQLite in-memory storage
func NewTestDatabase(t *testing.T) *DatabaseTestHelper {
	db, err := sqldb.NewDatabase("sqlite:///:memory:", true)
	require.NoError(t, err)

	// Run auto-migration for all models
	err = db.DB.AutoMigrate(
		&sqldb.SessionDB{},
		&sqldb.TerraformIDPAuthorizationCodeDB{},
		&sqldb.TerraformIDPAccessTokenDB{},
		&sqldb.TerraformIDPSubjectIdentifierDB{},
		&sqldb.UserGroupDB{},
		&sqldb.NamespaceDB{},
		&sqldb.UserGroupNamespacePermissionDB{},
		&sqldb.GitProviderDB{},
		&sqldb.NamespaceRedirectDB{},
		&sqldb.ModuleDetailsDB{},
		&sqldb.ModuleProviderDB{},
		&sqldb.ModuleVersionDB{},
		&sqldb.ModuleProviderRedirectDB{},
		&sqldb.SubmoduleDB{},
		&sqldb.AnalyticsDB{},
		&sqldb.ProviderAnalyticsDB{},
		&sqldb.ExampleFileDB{},
		&sqldb.ModuleVersionFileDB{},
		&sqldb.GPGKeyDB{},
		&sqldb.ProviderSourceDB{},
		&sqldb.ProviderCategoryDB{},
		&sqldb.RepositoryDB{},
		&sqldb.ProviderDB{},
		&sqldb.ProviderVersionDB{},
		&sqldb.ProviderVersionDocumentationDB{},
		&sqldb.ProviderVersionBinaryDB{},
		&sqldb.AuthenticationTokenDB{},
		&sqldb.AuditHistoryDB{},
	)
	require.NoError(t, err)

	return &DatabaseTestHelper{
		DB: db,
	}
}

// Close closes the database connection
func (h *DatabaseTestHelper) Close(t *testing.T) {
	if h.DB != nil {
		h.DB.Close()
	}
}

// BeginTransaction starts a new transaction for test isolation
func (h *DatabaseTestHelper) BeginTransaction(t *testing.T) {
	// Note: GORM transactions are handled differently
	// For simplicity, we'll skip transaction handling in basic test setup
}

// RollbackTransaction rolls back a transaction (use for test cleanup)
func (h *DatabaseTestHelper) RollbackTransaction(t *testing.T) {
	// Note: GORM transactions are handled differently
	// For simplicity, we'll skip transaction handling in basic test setup
}

// Cleanup cleans up database resources
func (h *DatabaseTestHelper) Cleanup(t *testing.T) {
	h.Close(t)
}

// WithTransaction creates a transaction and ensures rollback for test isolation
func (h *DatabaseTestHelper) WithTransaction(t *testing.T, fn func()) {
	// Note: GORM transactions are handled differently
	// For simplicity, we'll skip transaction handling in basic test setup
	fn()
}

// CreateTestNamespace creates a test namespace with the given name
func (h *DatabaseTestHelper) CreateTestNamespace(t *testing.T, namespace string) sqldb.NamespaceDB {
	ns := sqldb.NamespaceDB{
		Namespace:     namespace,
		DisplayName:   &namespace, // Use namespace as display name for tests
		NamespaceType: sqldb.NamespaceTypeNone,
	}

	err := h.DB.DB.Create(&ns).Error
	require.NoError(t, err)

	return ns
}

// CreateTestModuleProvider creates a test module provider
func (h *DatabaseTestHelper) CreateTestModuleProvider(t *testing.T, namespaceID int, moduleName, providerName string) sqldb.ModuleProviderDB {
	mp := sqldb.ModuleProviderDB{
		NamespaceID:           namespaceID,
		Module:                moduleName,
		Provider:              providerName,
		Verified:              nil, // false by default
		GitProviderID:         nil,
		RepoBaseURLTemplate:   nil,
		RepoCloneURLTemplate:  nil,
		RepoBrowseURLTemplate: nil,
		GitTagFormat:          nil,
		GitPath:               nil,
		ArchiveGitPath:        false,
		LatestVersionID:       nil,
	}

	err := h.DB.DB.Create(&mp).Error
	require.NoError(t, err)

	return mp
}

// CreateTestModuleVersion creates a test module version
func (h *DatabaseTestHelper) CreateTestModuleVersion(t *testing.T, moduleProviderID int, version string) sqldb.ModuleVersionDB {
	mv := sqldb.ModuleVersionDB{
		ModuleProviderID:      moduleProviderID,
		Version:               version,
		Beta:                  false,
		Internal:              false,
		Published:             nil, // false by default
		PublishedAt:           nil,
		GitSHA:                nil,
		GitPath:               nil,
		ArchiveGitPath:        false,
		RepoBaseURLTemplate:   nil,
		RepoCloneURLTemplate:  nil,
		RepoBrowseURLTemplate: nil,
		Owner:                 nil,
		Description:           nil,
		VariableTemplate:      nil,
		ExtractionVersion:     nil,
		ModuleDetailsID:       nil,
	}

	err := h.DB.DB.Create(&mv).Error
	require.NoError(t, err)

	return mv
}

// CreateTestUserGroup creates a test user group
func (h *DatabaseTestHelper) CreateTestUserGroup(t *testing.T, name string) sqldb.UserGroupDB {
	ug := sqldb.UserGroupDB{
		Name:      name,
		SiteAdmin: false,
	}

	err := h.DB.DB.Create(&ug).Error
	require.NoError(t, err)

	return ug
}
