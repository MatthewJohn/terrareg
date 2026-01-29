package testutils

import (
	"fmt"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/config/model"
	configService "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/config/service"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/config"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/container"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/version"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/handler/webhook"
)

// testWriter is an io.Writer that writes to testing.T.Log()
type testWriter struct {
	t *testing.T
}

// Write implements io.Writer by writing to t.Log()
func (tw *testWriter) Write(p []byte) (n int, err error) {
	tw.t.Log(string(p))
	return len(p), nil
}

// TestLogger is a no-op logger for testing (used when no *testing.T is available)
var TestLogger = zerolog.Nop()

// GetTestLogger returns a logger that outputs to testing.T.Log()
// This integrates with Go's testing framework and shows logs with `go test -v`
func GetTestLogger(t *testing.T) zerolog.Logger {
	return zerolog.New(&testWriter{t: t}).With().Timestamp().Logger()
}

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
	return CreateTestInfraConfigWithPublicURL(t, "http://localhost:5000")
}

// CreateTestInfraConfigWithPublicURL creates a test infrastructure configuration with a custom PUBLIC_URL
func CreateTestInfraConfigWithPublicURL(t *testing.T, publicURL string) *config.InfrastructureConfig {
	return &config.InfrastructureConfig{
		ListenPort:                  5000,
		PublicURL:                   publicURL,
		DomainName:                  "localhost",
		Debug:                       true,
		DatabaseURL:                 "sqlite:///:memory:",
		DataDirectory:               "/tmp/terrareg-test",
		SecretKey:                   "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
		SessionCookieName:           "terrareg_session",
		AdminAuthenticationToken:    "test-admin-api-key",
		UploadApiKeys:               []string{"test-upload-key"},
		AdminSessionExpiryMins:      60, // 1 hour for admin sessions
		TerraformLockTimeoutSeconds: 1800, // 30 minutes default (required for terraform operations)
		// Terraform configuration for tests - prevents tfswitch from trying to prompt interactively
		TerraformDefaultVersion: "1.5.7", // Use a specific version to avoid interactive prompts
		TerraformProduct:        "terraform",
	}
}

// CleanupTestDatabase closes the test database
func CleanupTestDatabase(t *testing.T, db *sqldb.Database) {
	if db != nil {
		db.Close()
	}
}

// CreateNamespace creates a test namespace in the database
// displayName is optional - pass nil or empty string for no display name
func CreateNamespace(t *testing.T, db *sqldb.Database, name string, displayName *string) sqldb.NamespaceDB {
	namespace := sqldb.NamespaceDB{
		Namespace:     name,
		DisplayName:   displayName,
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

	err := db.DB.Save(&moduleProvider).Error
	require.NoError(t, err)

	return moduleProvider
}

// CreateModuleProviderWithGit creates a test module provider with git configuration
func CreateModuleProviderWithGit(t *testing.T, db *sqldb.Database, namespaceID int, moduleName, providerName string, gitCloneURL *string) sqldb.ModuleProviderDB {
	moduleProvider := sqldb.ModuleProviderDB{
		NamespaceID:           namespaceID,
		Module:                moduleName,
		Provider:              providerName,
		Verified:              nil, // false
		GitProviderID:         nil,
		RepoBaseURLTemplate:   nil,
		RepoCloneURLTemplate:  gitCloneURL,
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
	now := time.Now()
	moduleVersion := sqldb.ModuleVersionDB{
		ModuleProviderID:      moduleProviderID,
		Version:               version,
		Beta:                  false,
		Internal:              false,
		Published:             &published,
		PublishedAt:           &now,
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

// CreatePublishedBetaModuleVersion creates a published test module version in the database with Beta=true
// Note: This does NOT set the version as the latest version, as beta versions should not be considered for latest
func CreatePublishedBetaModuleVersion(t *testing.T, db *sqldb.Database, moduleProviderID int, version string) sqldb.ModuleVersionDB {
	published := true
	now := time.Now()
	moduleVersion := sqldb.ModuleVersionDB{
		ModuleProviderID:      moduleProviderID,
		Version:               version,
		Beta:                  true,
		Internal:              false,
		Published:             &published,
		PublishedAt:           &now,
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

	// Note: We do NOT set this as the latest version because beta versions should not be considered for latest

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

// CreateFullyPopulatedModuleVersion creates a fully populated published module version.
// Python reference: /app/test/unit/terrareg/test_data.py - fullypopulated module
// This populates all optional fields for comprehensive testing, matching the Python test data pattern.
func CreateFullyPopulatedModuleVersion(t *testing.T, db *sqldb.Database, moduleProviderID int, version string) sqldb.ModuleVersionDB {
	t.Helper()

	// Populate all fields similar to Python's fullypopulated test data
	owner := "This is the owner of the module"
	description := "This is a test module version for tests."
	repoBaseURL := "https://link-to.com/source-code-here"
	repoCloneURL := "ssh://mp-clone-url.com/namespace-module-provider"
	repoBrowseURL := "https://mp-browse-url.com/namespace-module-provider/browse/{tag}/{path}suffix"
	gitPath := "modules/test"
	gitSHA := "abc123def456"

	published := true
	now := time.Now()

	moduleVersion := sqldb.ModuleVersionDB{
		ModuleProviderID:      moduleProviderID,
		Version:               version,
		Beta:                  false,
		Internal:              false,
		Published:             &published,
		PublishedAt:           &now,
		Owner:                 &owner,
		Description:           &description,
		RepoBaseURLTemplate:   &repoBaseURL,
		RepoCloneURLTemplate:  &repoCloneURL,
		RepoBrowseURLTemplate: &repoBrowseURL,
		GitPath:               &gitPath,
		GitSHA:                &gitSHA,
		ArchiveGitPath:        false,
		VariableTemplate:      nil,
		ExtractionVersion:     nil,
		ModuleDetailsID:       nil,
	}

	err := db.DB.Create(&moduleVersion).Error
	require.NoError(t, err, "Failed to create fully populated module version")

	// Update the module provider to set this as the latest version
	err = db.DB.Model(&sqldb.ModuleProviderDB{}).
		Where("id = ?", moduleProviderID).
		Update("latest_version_id", moduleVersion.ID).Error
	require.NoError(t, err, "Failed to set latest version")

	return moduleVersion
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

// CreateRepository creates a test repository in the database
// Python reference: /app/test/unit/terrareg/test_data.py - repository creation
func CreateRepository(t *testing.T, db *sqldb.Database, providerID *string, owner, name, description, cloneURL, logoURL, providerSourceName string) sqldb.RepositoryDB {
	t.Helper()

	repo := sqldb.RepositoryDB{
		ProviderID:         providerID,
		Owner:              &owner,
		Name:               &name,
		Description:        []byte(description),
		CloneURL:           &cloneURL,
		LogoURL:            &logoURL,
		ProviderSourceName: providerSourceName,
	}

	err := db.DB.Create(&repo).Error
	require.NoError(t, err)

	return repo
}

// CreateProviderVersionWithRepository creates a test provider with repository and version
// This is a comprehensive helper that creates all necessary entities for a fully populated provider
// matching Python's test_data.py structure
// Python reference: /app/test/unit/terrareg/server/test_api_provider_list.py - test_data
func CreateProviderVersionWithRepository(t *testing.T, db *sqldb.Database, namespaceID int, providerName, version, gitTag string, description *string, tier sqldb.ProviderTier, gpgKeyID int, publishedAt *time.Time) (sqldb.ProviderDB, sqldb.RepositoryDB, sqldb.ProviderVersionDB) {
	t.Helper()

	// Create provider
	provider := CreateProvider(t, db, namespaceID, providerName, description, tier, nil)

	// Create repository linked to the provider
	providerID := fmt.Sprintf("%d", provider.ID)
	owner := fmt.Sprintf("namespace-%d", namespaceID)
	repoName := fmt.Sprintf("terraform-provider-%s", providerName)
	repoDescription := "Test Provider Description"
	if description != nil && *description != "" {
		repoDescription = *description
	}
	cloneURL := fmt.Sprintf("https://github.example.com/%s/%s.git", owner, repoName)
	logoURL := fmt.Sprintf("https://github.example.com/%s/logo.png", owner)
	repository := CreateRepository(t, db, &providerID, owner, repoName, repoDescription, cloneURL, logoURL, "Test Github Autogenerate")

	// Update provider to link to repository
	err := db.DB.Model(&provider).Update("repository_id", repository.ID).Error
	require.NoError(t, err)

	// Create provider version with git tag
	gitTagPtr := &gitTag
	providerVersion := sqldb.ProviderVersionDB{
		ProviderID:  provider.ID,
		Version:     version,
		GPGKeyID:    gpgKeyID,
		GitTag:      gitTagPtr,
		Beta:        false,
		PublishedAt: publishedAt,
	}
	err = db.DB.Create(&providerVersion).Error
	require.NoError(t, err)

	// Set as latest version
	SetProviderLatestVersion(t, db, provider.ID, providerVersion.ID)

	return provider, repository, providerVersion
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

	cont, err := container.NewContainer(domainConfig, infraConfig, cfgService, GetTestLogger(t), db)
	require.NoError(t, err)

	return cont
}

// InfraConfigOption is a function that modifies the infrastructure configuration
type InfraConfigOption func(*config.InfrastructureConfig)

// CreateTestContainerWithConfig creates a test container with custom infrastructure configuration
// This allows tests to override specific config values (e.g., for mock auth servers)
func CreateTestContainerWithConfig(t *testing.T, db *sqldb.Database, opts ...InfraConfigOption) *container.Container {
	domainConfig := CreateTestDomainConfig(t)
	infraConfig := CreateTestInfraConfig(t)

	// Apply custom configuration options
	for _, opt := range opts {
		opt(infraConfig)
	}

	// Create a simple config service for testing
	versionReader := version.NewVersionReader()
	cfgService := configService.NewConfigurationService(configService.ConfigurationServiceOptions{}, versionReader)

	cont, err := container.NewContainer(domainConfig, infraConfig, cfgService, GetTestLogger(t), db)
	require.NoError(t, err)

	return cont
}

// WithOIDCConfig sets OIDC configuration for testing
func WithOIDCConfig(issuer, clientID, clientSecret string) InfraConfigOption {
	return func(cfg *config.InfrastructureConfig) {
		cfg.OpenIDConnectIssuer = issuer
		cfg.OpenIDConnectClientID = clientID
		cfg.OpenIDConnectClientSecret = clientSecret
		cfg.OpenIDConnectScopes = []string{"openid", "profile", "email"}
	}
}

// WithSAMLConfig sets SAML configuration for testing
func WithSAMLConfig(entityID, metadataURL string) InfraConfigOption {
	return func(cfg *config.InfrastructureConfig) {
		cfg.SAML2EntityID = entityID
		cfg.SAML2IDPMetadataURL = metadataURL
	}
}

// WithPublicURL sets the public URL for testing
func WithPublicURL(publicURL string) InfraConfigOption {
	return func(cfg *config.InfrastructureConfig) {
		cfg.PublicURL = publicURL
	}
}

// CreateTestWebhookHandler creates a test webhook handler with proper dependencies
// It uses the container to wire all dependencies together
func CreateTestWebhookHandler(t *testing.T, db *sqldb.Database, uploadAPIKeys []string) *webhook.ModuleWebhookHandler {
	cont := CreateTestContainer(t, db)

	// Create a new webhook handler with the provided upload API keys
	return webhook.NewModuleWebhookHandler(cont.WebhookService, uploadAPIKeys)
}

// GetNamespace retrieves a namespace by name from the database
// This is used by selenium tests to get namespace information for setting up permissions
func GetNamespace(t *testing.T, db *sqldb.Database, name string) sqldb.NamespaceDB {
	var namespace sqldb.NamespaceDB
	err := db.DB.Where("namespace = ?", name).First(&namespace).Error
	require.NoError(t, err, "Namespace should exist: %s", name)
	return namespace
}
