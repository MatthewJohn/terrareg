package testutils

import (
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/config/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/config"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
)

// TestLogger is a no-op logger for testing
var TestLogger = zerolog.Nop()

// SetupTestDatabase creates an in-memory SQLite database for testing
func SetupTestDatabase(t *testing.T) *sqldb.Database {
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

	return db
}

// GetTestLogger returns a test logger
func GetTestLogger() zerolog.Logger {
	return TestLogger
}

// CreateTestDomainConfig creates a test domain configuration
func CreateTestDomainConfig(t *testing.T) *model.DomainConfig {
	return &model.DomainConfig{
		TrustedNamespaces:        []string{"test"},
		VerifiedModuleNamespaces: []string{"verified"},
		AllowModuleHosting:       model.ModuleHostingModeAllow,
		SecretKeySet:             true,
		OpenIDConnectEnabled:     true,
		SAMLEnabled:              true,
		AdminLoginEnabled:        true,
	}
}

// CreateTestInfraConfig creates a test infrastructure configuration
func CreateTestInfraConfig(t *testing.T) *config.InfrastructureConfig {
	return &config.InfrastructureConfig{
		ListenPort:    5000,
		PublicURL:     "http://localhost:5000",
		DomainName:    "localhost",
		Debug:         true,
		DatabaseURL:   "sqlite:///:memory:",
		DataDirectory: "/tmp/terrareg-test",
		SecretKey:     "test-secret-key-that-is-32-chars-long",
	}
}

// CleanupTestDatabase closes the test database
func CleanupTestDatabase(t *testing.T, db *sqldb.Database) {
	if db != nil {
		db.Close()
	}
}

// CreateNamespace creates a test namespace in the database
func CreateNamespace(t *testing.T, db *sqldb.Database, name string) sqldb.NamespaceDB {
	displayName := name + " Display"
	namespace := sqldb.NamespaceDB{
		Namespace:     name,
		DisplayName:   &displayName,
		NamespaceType: sqldb.NamespaceTypeNone,
	}

	err := db.DB.Create(&namespace).Error
	require.NoError(t, err)

	return namespace
}

// CreateModuleProvider creates a test module provider in the database
func CreateModuleProvider(t *testing.T, db *sqldb.Database, namespaceID int, moduleName, providerName string) sqldb.ModuleProviderDB {
	moduleProvider := sqldb.ModuleProviderDB{
		NamespaceID:           namespaceID,
		Module:                moduleName,
		Provider:              providerName,
		Verified:              nil, // false
		GitProviderID:         nil,
		RepoBaseURLTemplate:   nil,
		RepoCloneURLTemplate:  nil,
		RepoBrowseURLTemplate: nil,
		GitTagFormat:          nil,
		GitPath:               nil,
		ArchiveGitPath:        false,
		LatestVersionID:       nil,
	}

	err := db.DB.Create(&moduleProvider).Error
	require.NoError(t, err)

	return moduleProvider
}

// CreateModuleVersion creates a test module version in the database
func CreateModuleVersion(t *testing.T, db *sqldb.Database, moduleProviderID int, version string) sqldb.ModuleVersionDB {
	moduleVersion := sqldb.ModuleVersionDB{
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

	err := db.DB.Create(&moduleVersion).Error
	require.NoError(t, err)

	return moduleVersion
}

// CreateModuleDetails creates test module details in the database
func CreateModuleDetails(t *testing.T, db *sqldb.Database, readmeContent string) sqldb.ModuleDetailsDB {
	moduleDetails := sqldb.ModuleDetailsDB{
		ReadmeContent:    []byte(readmeContent),
		TerraformDocs:    []byte("{}"),
		Tfsec:            []byte("{}"),
		Infracost:        []byte("{}"),
		TerraformGraph:   []byte("{}"),
		TerraformModules: []byte("{}"),
		TerraformVersion: []byte("1.0.0"),
	}

	err := db.DB.Create(&moduleDetails).Error
	require.NoError(t, err)

	return moduleDetails
}

// CreateSubmodule creates a test submodule in the database
func CreateSubmodule(t *testing.T, db *sqldb.Database, moduleVersionID int, path, name, submoduleType string, moduleDetailsID *int) sqldb.SubmoduleDB {
	submodule := sqldb.SubmoduleDB{
		ParentModuleVersion: moduleVersionID,
		Path:                path,
		Name:                &name,
		Type:                &submoduleType,
		ModuleDetailsID:     moduleDetailsID,
	}

	err := db.DB.Create(&submodule).Error
	require.NoError(t, err)

	return submodule
}

// CreateExampleFile creates a test example file in the database
func CreateExampleFile(t *testing.T, db *sqldb.Database, submoduleID int, path, content string) sqldb.ExampleFileDB {
	exampleFile := sqldb.ExampleFileDB{
		SubmoduleID: submoduleID,
		Path:        path,
		Content:     []byte(content),
	}

	err := db.DB.Create(&exampleFile).Error
	require.NoError(t, err)

	return exampleFile
}
