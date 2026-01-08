package testutils

import (
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/config/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/config"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/handler/webhook"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/container"
	configService "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/config/service"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/version"
)

// TestLogger is a no-op logger for testing
var TestLogger = zerolog.Nop()

// SetupTestDatabase creates an in-memory SQLite database for testing
func SetupTestDatabase(t *testing.T) *sqldb.Database {
	// Use file::memory:?cache=shared to create a unique in-memory database per connection
	// This prevents test data from leaking between tests
	db, err := sqldb.NewDatabase("file::memory:?cache=shared", true)
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
		// Temporarily exclude AuthenticationTokenDB due to enum issues
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
		ListenPort:                5000,
		PublicURL:                 "http://localhost:5000",
		DomainName:                "localhost",
		Debug:                     true,
		DatabaseURL:               "sqlite:///:memory:",
		DataDirectory:             "/tmp/terrareg-test",
		SecretKey:                 "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
		SessionCookieName:         "terrareg_session",
		AdminAuthenticationToken:  "test-admin-api-key",
		UploadApiKeys:             []string{"test-upload-key"},
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

// CreatePublishedModuleVersion creates a published test module version in the database
// It automatically sets the version as the latest version on the module provider
func CreatePublishedModuleVersion(t *testing.T, db *sqldb.Database, moduleProviderID int, version string) sqldb.ModuleVersionDB {
	published := true
	moduleVersion := sqldb.ModuleVersionDB{
		ModuleProviderID:      moduleProviderID,
		Version:               version,
		Beta:                  false,
		Internal:              false,
		Published:             &published,
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

	// Set this version as the latest version for the module provider
	// This is required for the search query to find the module
	err = db.DB.Model(&sqldb.ModuleProviderDB{}).
		Where("id = ?", moduleProviderID).
		Update("latest_version_id", moduleVersion.ID).Error
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

// CreateProviderCategory creates a test provider category in the database
func CreateProviderCategory(t *testing.T, db *sqldb.Database, name, slug string, userSelectable bool) sqldb.ProviderCategoryDB {
	category := sqldb.ProviderCategoryDB{
		Name:           &name,
		Slug:           slug,
		UserSelectable: userSelectable,
	}

	err := db.DB.Create(&category).Error
	require.NoError(t, err)

	return category
}

// CreateProvider creates a test provider in the database
func CreateProvider(t *testing.T, db *sqldb.Database, namespaceID int, name string, description *string, tier sqldb.ProviderTier, categoryID *int) sqldb.ProviderDB {
	provider := sqldb.ProviderDB{
		NamespaceID:               namespaceID,
		Name:                      name,
		Description:               description,
		Tier:                      tier,
		DefaultProviderSourceAuth: false,
		ProviderCategoryID:        categoryID,
		RepositoryID:              nil,
		LatestVersionID:           nil,
	}

	err := db.DB.Create(&provider).Error
	require.NoError(t, err)

	return provider
}

// CreateProviderVersion creates a test provider version in the database
func CreateProviderVersion(t *testing.T, db *sqldb.Database, providerID int, version string, gpgKeyID int, beta bool, publishedAt *time.Time) sqldb.ProviderVersionDB {
	providerVersion := sqldb.ProviderVersionDB{
		ProviderID:        providerID,
		Version:           version,
		GPGKeyID:          gpgKeyID,
		GitTag:            nil,
		Beta:              beta,
		PublishedAt:       publishedAt,
		ExtractionVersion: nil,
	}

	err := db.DB.Create(&providerVersion).Error
	require.NoError(t, err)

	return providerVersion
}

// CreateProviderVersionBinary creates a test provider version binary in the database
func CreateProviderVersionBinary(t *testing.T, db *sqldb.Database, providerVersionID int, name string, os sqldb.ProviderBinaryOperatingSystemType, arch sqldb.ProviderBinaryArchitectureType, checksum string) sqldb.ProviderVersionBinaryDB {
	binary := sqldb.ProviderVersionBinaryDB{
		ProviderVersionID: providerVersionID,
		Name:              name,
		OperatingSystem:   os,
		Architecture:      arch,
		Checksum:          checksum,
	}

	err := db.DB.Create(&binary).Error
	require.NoError(t, err)

	return binary
}

// SetProviderLatestVersion sets the latest version for a provider
func SetProviderLatestVersion(t *testing.T, db *sqldb.Database, providerID, latestVersionID int) {
	err := db.DB.Model(&sqldb.ProviderDB{}).Where("id = ?", providerID).Update("latest_version_id", latestVersionID).Error
	require.NoError(t, err)
}

// CreateGPGKeyWithNamespace creates a test GPG key in the database linked to a namespace
// This is the preferred method as GPG keys belong to namespaces, not providers
func CreateGPGKeyWithNamespace(t *testing.T, db *sqldb.Database, name string, namespaceID int, keyID string) sqldb.GPGKeyDB {
	asciiArmor := []byte("-----BEGIN PGP PUBLIC KEY BLOCK-----\n\nTest ASCII armor\n-----END PGP PUBLIC KEY BLOCK-----")
	fingerprint := &keyID

	gpgKey := sqldb.GPGKeyDB{
		NamespaceID: namespaceID,
		ASCIIArmor:  asciiArmor,
		KeyID:       &keyID,
		Fingerprint: fingerprint,
		Source:      &name,
		SourceURL:   nil,
		CreatedAt:   nil,
		UpdatedAt:   nil,
	}

	err := db.DB.Create(&gpgKey).Error
	require.NoError(t, err)

	return gpgKey
}

// CreateProviderVersionDocumentation creates a test provider version documentation in the database
func CreateProviderVersionDocumentation(t *testing.T, db *sqldb.Database, providerVersionID int, name, slug string, docType sqldb.ProviderDocumentationType) sqldb.ProviderVersionDocumentationDB {
	doc := sqldb.ProviderVersionDocumentationDB{
		ProviderVersionID: providerVersionID,
		Name:              name,
		Slug:              slug,
		Title:             nil,
		Description:       nil,
		Language:          "hcl",
		Subcategory:       nil,
		Filename:          "docs/" + name,
		DocumentationType: docType,
		Content:           []byte("# Test Documentation\n\nSome content here."),
	}

	err := db.DB.Create(&doc).Error
	require.NoError(t, err)

	return doc
}

// CreateGPGKey creates a test GPG key in the database (linked to namespace)
// Deprecated: Use CreateGPGKeyWithNamespace for clarity
func CreateGPGKey(t *testing.T, db *sqldb.Database, name string, providerID int, keyID string) sqldb.GPGKeyDB {
	// Provider versions use namespace GPG keys
	// For testing, we need to get the provider's namespace first
	var provider sqldb.ProviderDB
	err := db.DB.First(&provider, providerID).Error
	require.NoError(t, err)

	asciiArmor := []byte("-----BEGIN PGP PUBLIC KEY BLOCK-----\n\nTest ASCII armor\n-----END PGP PUBLIC KEY BLOCK-----")
	fingerprint := &keyID

	gpgKey := sqldb.GPGKeyDB{
		NamespaceID: provider.NamespaceID,
		ASCIIArmor:  asciiArmor,
		KeyID:       &keyID,
		Fingerprint: fingerprint,
		Source:      &name,
		SourceURL:   nil,
		CreatedAt:   nil,
		UpdatedAt:   nil,
	}

	err = db.DB.Create(&gpgKey).Error
	require.NoError(t, err)

	return gpgKey
}

// CreateTestContainer creates a test container with all dependencies wired together
// This is the preferred way to set up integration tests that need handlers or services
func CreateTestContainer(t *testing.T, db *sqldb.Database) *container.Container {
	domainConfig := CreateTestDomainConfig(t)
	infraConfig := CreateTestInfraConfig(t)

	// Create a simple config service for testing
	versionReader := version.NewVersionReader()
	cfgService := configService.NewConfigurationService(configService.ConfigurationServiceOptions{}, versionReader)

	cont, err := container.NewContainer(domainConfig, infraConfig, cfgService, GetTestLogger(), db)
	require.NoError(t, err)

	return cont
}

// CreateTestWebhookHandler creates a test webhook handler with proper dependencies
// It uses the container to wire all dependencies together
func CreateTestWebhookHandler(t *testing.T, db *sqldb.Database, uploadAPIKeys []string) *webhook.ModuleWebhookHandler {
	cont := CreateTestContainer(t, db)

	// Create a new webhook handler with the provided upload API keys
	return webhook.NewModuleWebhookHandler(cont.WebhookService, uploadAPIKeys)
}
