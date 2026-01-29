package selenium

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
	integrationTestUtils "github.com/matthewjohn/terrareg/terrareg-go/test/integration/testutils"
)

const (
	// DefaultTotalDownloads matches the Python mock value
	// Python reference: /app/test/selenium/test_homepage.py - mock.patch('get_total_downloads', return_value=2005)
	DefaultTotalDownloads = 2005
)

// HomepageTotalDownloads returns the expected download count for homepage tests
func HomepageTotalDownloads() int {
	return DefaultTotalDownloads
}

// UpdateModuleProviderVerified updates a module provider's verified status.
// This is equivalent to Python's provider.update_attributes(verified=True).
// Python reference: /app/test/selenium/test_homepage.py - provider.update_attributes(verified=True)
func UpdateModuleProviderVerified(t *testing.T, db *sqldb.Database, moduleProviderID int, verified bool) {
	verifiedPtr := &verified
	err := db.DB.Model(&sqldb.ModuleProviderDB{}).
		Where("id = ?", moduleProviderID).
		Update("verified", verifiedPtr).Error
	require.NoError(t, err, "Failed to update module provider verified status")
}

// UpdateModuleVersionPublishedAt updates a module version's published_at timestamp.
// This is equivalent to Python's module_version.update_attributes(published_at=datetime.now()).
// Python reference: /app/test/selenium/test_homepage.py - module_version.update_attributes(published_at=datetime.now())
func UpdateModuleVersionPublishedAt(t *testing.T, db *sqldb.Database, moduleVersionID int, publishedAt time.Time) {
	err := db.DB.Model(&sqldb.ModuleVersionDB{}).
		Where("id = ?", moduleVersionID).
		Update("published_at", publishedAt).Error
	require.NoError(t, err, "Failed to update module version published_at")
}

// GetNamespaceByName retrieves a namespace by name from the database.
// This is equivalent to Python's Namespace(name='...').
// Python reference: /app/test/selenium/test_homepage.py - Namespace('mostrecent')
func GetNamespaceByName(t *testing.T, db *sqldb.Database, name string) sqldb.NamespaceDB {
	var namespace sqldb.NamespaceDB
	err := db.DB.Where("namespace = ?", name).First(&namespace).Error
	require.NoError(t, err, "Failed to find namespace: %s", name)
	return namespace
}

// GetModuleProvider retrieves a module provider by namespace, module, and provider names.
// This is equivalent to Python's ModuleProvider(module=Module(namespace=Namespace(name='...'), name='...'), name='...').
// Python reference: /app/test/selenium/test_homepage.py - ModuleProvider lookup
func GetModuleProvider(t *testing.T, db *sqldb.Database, namespaceName, moduleName, providerName string) sqldb.ModuleProviderDB {
	var moduleProvider sqldb.ModuleProviderDB
	err := db.DB.Joins("JOIN namespace_db ON namespace_db.id = module_provider_db.namespace_id").
		Where("namespace_db.namespace = ?", namespaceName).
		Where("module_provider_db.module = ?", moduleName).
		Where("module_provider_db.provider = ?", providerName).
		First(&moduleProvider).Error
	require.NoError(t, err, "Failed to find module provider: %s/%s/%s", namespaceName, moduleName, providerName)
	return moduleProvider
}

// GetModuleVersion retrieves a module version by provider and version.
// This is equivalent to Python's ModuleVersion(module_provider=..., version='...').
// Python reference: /app/test/selenium/test_homepage.py - ModuleVersion lookup
func GetModuleVersion(t *testing.T, db *sqldb.Database, moduleProviderID int, version string) sqldb.ModuleVersionDB {
	var moduleVersion sqldb.ModuleVersionDB
	err := db.DB.Where("module_provider_id = ?", moduleProviderID).
		Where("version = ?", version).
		First(&moduleVersion).Error
	require.NoError(t, err, "Failed to find module version: %d/%s", moduleProviderID, version)
	return moduleVersion
}

// SetupHomepageTestData creates test data for homepage tests.
// This creates the modules and versions needed for the homepage to display properly.
// Python reference: /app/test/selenium/test_homepage.py - TestHomePage data setup
func SetupHomepageTestData(t *testing.T, db *sqldb.Database) {
	// Create "mostrecent" namespace and module for latest module version tests
	mostRecentNs := integrationTestUtils.CreateNamespace(t, db, "mostrecent", nil)
	mostRecentMp := integrationTestUtils.CreateModuleProvider(t, db, mostRecentNs.ID, "modulename", "providername")
	_ = integrationTestUtils.CreatePublishedModuleVersion(t, db, mostRecentMp.ID, "1.2.3")
	_ = integrationTestUtils.CreateModuleDetails(t, db, "# Test Module\n\nThis is a test module for homepage display.")

	// Create "trustednamespace" for trusted module tests
	trustedNs := integrationTestUtils.CreateNamespace(t, db, "trustednamespace", nil)
	trustedMp := integrationTestUtils.CreateModuleProvider(t, db, trustedNs.ID, "secondlatestmodule", "aws")
	_ = integrationTestUtils.CreatePublishedModuleVersion(t, db, trustedMp.ID, "4.4.1")
	_ = integrationTestUtils.CreateModuleDetails(t, db, "# Trusted Module\n\nThis is a trusted module.")
}

// SetupSearchTestData creates test data for search tests.
// This creates multiple modules with different attributes for search testing.
// Python reference: /app/test/selenium/test_module_search.py - search test data
func SetupSearchTestData(t *testing.T, db *sqldb.Database) {
	// Create namespaces
	ns1 := integrationTestUtils.CreateNamespace(t, db, "modulesearch", nil)
	_ = integrationTestUtils.CreateNamespace(t, db, "mixedsearch", nil)

	// Create module providers for module search
	mp1 := integrationTestUtils.CreateModuleProvider(t, db, ns1.ID, "modulesearch-trusted", "testprovider")
	mp2 := integrationTestUtils.CreateModuleProvider(t, db, ns1.ID, "modulesearch-result", "testprovider")
	mp3 := integrationTestUtils.CreateModuleProvider(t, db, ns1.ID, "othermodule", "testprovider")

	// Create published versions
	_ = integrationTestUtils.CreatePublishedModuleVersion(t, db, mp1.ID, "1.0.0")
	_ = integrationTestUtils.CreatePublishedModuleVersion(t, db, mp2.ID, "1.0.0")
	_ = integrationTestUtils.CreatePublishedModuleVersion(t, db, mp3.ID, "1.0.0")

	// Create module details
	integrationTestUtils.CreateModuleDetails(t, db, "# Trusted Module")
	integrationTestUtils.CreateModuleDetails(t, db, "# Search Result Module")
	integrationTestUtils.CreateModuleDetails(t, db, "# Other Module")
}

// SetupNamespaceTestData creates test data for namespace page tests.
// This creates a namespace with multiple modules and providers.
// Python reference: /app/test/selenium/test_namespace.py - namespace test data
func SetupNamespaceTestData(t *testing.T, db *sqldb.Database) {
	// Create namespace with various module types
	namespace := integrationTestUtils.CreateNamespace(t, db, "testnamespace", nil)

	// Create a standard module provider
	moduleProvider := integrationTestUtils.CreateModuleProvider(t, db, namespace.ID, "testmodule", "testprovider")

	// Create a published version
	_ = integrationTestUtils.CreatePublishedModuleVersion(t, db, moduleProvider.ID, "1.0.0")

	// Create module details
	_ = integrationTestUtils.CreateModuleDetails(t, db, "# Test Module\n\nModule description here.")
}

// SetupModuleProviderTestData creates test data for module provider page tests.
// This creates all modules, versions, and details needed by module_provider_test.go
// Python reference: /app/test/selenium/test_module_provider.py - module provider test data
// Python reference: /app/test/selenium/test_data.py - integration_test_data['moduledetails']
func SetupModuleProviderTestData(t *testing.T, db *sqldb.Database) {
	// Create namespace
	namespace := integrationTestUtils.CreateNamespace(t, db, "moduledetails", nil)

	// Create "fullypopulated" module provider with multiple versions
	fullyPopulatedMp := integrationTestUtils.CreateModuleProvider(t, db, namespace.ID, "fullypopulated", "testprovider")
	_ = integrationTestUtils.CreatePublishedModuleVersion(t, db, fullyPopulatedMp.ID, "1.5.0")
	_ = integrationTestUtils.CreatePublishedModuleVersion(t, db, fullyPopulatedMp.ID, "1.4.0")
	_ = integrationTestUtils.CreatePublishedModuleVersion(t, db, fullyPopulatedMp.ID, "1.2.0")
	_ = integrationTestUtils.CreatePublishedModuleVersion(t, db, fullyPopulatedMp.ID, "1.0.0")

	// Create beta and unpublished versions for fullypopulated
	_ = integrationTestUtils.CreatePublishedBetaModuleVersion(t, db, fullyPopulatedMp.ID, "1.7.0-beta") // Beta
	_ = integrationTestUtils.CreateModuleVersion(t, db, fullyPopulatedMp.ID, "1.6.0")                   // Unpublished
	_ = integrationTestUtils.CreateModuleVersion(t, db, fullyPopulatedMp.ID, "1.6.1-beta")              // Unpublished Beta
	_ = integrationTestUtils.CreateModuleVersion(t, db, fullyPopulatedMp.ID, "1.0.0-beta")              // Unpublished beta

	// Create module details with full content for fullypopulated
	readmeContent := `# Fully Populated Module

This is a test module version for tests.

## Features

- Feature 1
- Feature 2

## Usage

` + "```hcl\n" + `module "example" {
  source = "moduledetails/fullypopulated/testprovider"
  version = "1.5.0"
}
` + "```\n"
	moduleDetails := integrationTestUtils.CreateModuleDetails(t, db, readmeContent)

	// Update module versions to reference module details
	db.DB.Model(&sqldb.ModuleVersionDB{}).
		Where("module_provider_id = ?", fullyPopulatedMp.ID).
		Update("module_details_id", moduleDetails.ID)

	// Create "noversion" module provider - has no versions
	_ = integrationTestUtils.CreateModuleProvider(t, db, namespace.ID, "noversion", "testprovider")

	// Create "withsecurityissues" module provider for security tests
	securityMp := integrationTestUtils.CreateModuleProvider(t, db, namespace.ID, "withsecurityissues", "testprovider")
	_ = integrationTestUtils.CreatePublishedModuleVersion(t, db, securityMp.ID, "1.1.0")
	_ = integrationTestUtils.CreatePublishedModuleVersion(t, db, securityMp.ID, "1.0.0")

	securityDetails := integrationTestUtils.CreateModuleDetails(t, db, "# Security Module\n\nThis module has security issues.")
	db.DB.Model(&sqldb.ModuleVersionDB{}).
		Where("module_provider_id = ?", securityMp.ID).
		Update("module_details_id", securityDetails.ID)

	// Create example records for fullypopulated module
	// Python reference: /app/test/selenium/test_data.py - integration_test_data['moduledetails']['modules']['fullypopulated']['testprovider']['versions']['1.5.0']['examples']
	var moduleVersion1_5_0 sqldb.ModuleVersionDB
	err := db.DB.Where("module_provider_id = ? AND version = ?", fullyPopulatedMp.ID, "1.5.0").First(&moduleVersion1_5_0).Error
	require.NoError(t, err, "Failed to find module version 1.5.0")

	// Create example submodule for "examples/test-example"
	// Python: examples are stored in the submodule table with type='example'
	exampleType := "example"
	exampleDetails := integrationTestUtils.CreateModuleDetails(t, db, "# Example 1 README\n\nThis is a test example.")
	exampleSubmodule := integrationTestUtils.CreateSubmodule(t, db, moduleVersion1_5_0.ID, "examples/test-example", "", exampleType, &exampleDetails.ID)

	// Create example files for the example
	// Python reference: example_files in test_data.py
	_ = integrationTestUtils.CreateExampleFile(t, db, exampleSubmodule.ID, "examples/test-example/data.tf", "# This contains data objects")
	_ = integrationTestUtils.CreateExampleFile(t, db, exampleSubmodule.ID, "examples/test-example/variables.tf", `variable "test" {
  description = "test variable"
  type = string
}`)
	_ = integrationTestUtils.CreateExampleFile(t, db, exampleSubmodule.ID, "examples/test-example/main.tf", `# Call root module
module "root" {
  source = "../../"
}`)

	// Create a submodule for "modules/example-submodule1" (used in tests)
	submoduleDetails := integrationTestUtils.CreateModuleDetails(t, db, "# Submodule README\n\nThis is a test submodule.")
	submoduleType := "submodule"
	_ = integrationTestUtils.CreateSubmodule(t, db, moduleVersion1_5_0.ID, "modules/example-submodule1", "", submoduleType, &submoduleDetails.ID)
}

// SetupAuditHistoryTestData creates audit history test data.
// Python reference: /app/test/selenium/test_audit_history.py - TestAuditHistory._setup_test_audit_data
func SetupAuditHistoryTestData(t *testing.T, db *sqldb.Database) {
	// Clear existing audit history
	db.DB.Exec("DELETE FROM audit_history")

	// Create test namespace to prevent redirect to initial-setup
	integrationTestUtils.CreateNamespace(t, db, "test-namespace", nil)

	// Create test audit history entries matching the Python test data
	// Python reference: /app/test/selenium/test_audit_history.py - _AUDIT_DATA

	// User group entries
	// Format: username, action, object_type, object_id, old_value, new_value, timestamp
	userGroupDeleteTime := time.Date(2099, 1, 2, 9, 2, 0, 0, time.UTC)
	createAuditEntry(t, db, "useradmin", "user_group_delete", "UserGroup", "test-user-group", nil, nil, userGroupDeleteTime)

	userGroupCreateTime := time.Date(2099, 1, 2, 9, 1, 0, 0, time.UTC)
	createAuditEntry(t, db, "useradmin", "user_group_create", "UserGroup", "test-user-group", nil, nil, userGroupCreateTime)

	// Namespace entries
	namespaceCreateTime := time.Date(2020, 11, 27, 19, 14, 0, 0, time.UTC)
	createAuditEntry(t, db, "test-event-admin", "namespace_create", "Namespace", "test-namespace", nil, nil, namespaceCreateTime)

	// Module provider entries
	moduleProviderCreateTime := time.Date(2020, 11, 27, 19, 14, 10, 0, time.UTC)
	createAuditEntry(t, db, "test-event-admin", "module_provider_create", "ModuleProvider", "test-namespace/test-module/provider", nil, nil, moduleProviderCreateTime)

	oldCloneURL := "old-git-clone-url"
	newCloneURL := "new-git-clone-url"
	moduleProviderUpdateTime := time.Date(2020, 11, 28, 8, 47, 20, 0, time.UTC)
	createAuditEntry(t, db, "namespaceadmin", "module_provider_update_git_custom_clone_url", "ModuleProvider", "test-namespace/test-module/provider", &oldCloneURL, &newCloneURL, moduleProviderUpdateTime)

	// Module version entries
	moduleVersionIndex2Time := time.Date(2021, 12, 28, 19, 15, 10, 0, time.UTC)
	createAuditEntry(t, db, "namespaceowner", "module_version_index", "ModuleVersion", "test-namespace/test-module/provider/2.0.1", nil, nil, moduleVersionIndex2Time)

	moduleVersionIndex1Time := time.Date(2021, 12, 28, 19, 16, 23, 0, time.UTC)
	createAuditEntry(t, db, "namespaceowner", "module_version_index", "ModuleVersion", "test-namespace/test-module/provider/2.0.1", nil, nil, moduleVersionIndex1Time)

	moduleVersionPublishTime := time.Date(2021, 12, 29, 19, 23, 31, 0, time.UTC)
	createAuditEntry(t, db, "namespaceowner", "module_version_publish", "ModuleVersion", "test-namespace/test-module/provider/2.0.1", nil, nil, moduleVersionPublishTime)

	moduleVersionDeleteTime := time.Date(2021, 12, 29, 20, 12, 23, 0, time.UTC)
	createAuditEntry(t, db, "namespaceowner", "module_version_delete", "ModuleVersion", "test-namespace/test-module/provider/2.0.1", nil, nil, moduleVersionDeleteTime)

	// User login entries (testuser1-9)
	for i := 1; i <= 9; i++ {
		username := fmt.Sprintf("testuser%d", i)
		objectID := fmt.Sprintf("testuser%d", i)
		loginTime := time.Date(2020, 1, 2, 9, 49+i-1, 20+(i-1), 0, time.UTC)
		createAuditEntry(t, db, username, "user_login", "User", objectID, nil, nil, loginTime)
	}
}

// createAuditEntry creates a single audit history entry
func createAuditEntry(t *testing.T, db *sqldb.Database, username, action, objectType, objectID string, oldValue, newValue *string, timestamp time.Time) {
	auditEntry := sqldb.AuditHistoryDB{
		Timestamp:  &timestamp,
		Username:   &username,
		Action:     sqldb.AuditAction(action),
		ObjectType: &objectType,
		ObjectID:   &objectID,
		OldValue:   oldValue,
		NewValue:   newValue,
	}
	err := db.DB.Create(&auditEntry).Error
	require.NoError(t, err, "Failed to create audit entry")
}

// SetupLoginTestData creates minimal test data for login tests.
// Login tests typically don't need much module data.
// Python reference: /app/test/selenium/test_login.py - login test data
func SetupLoginTestData(t *testing.T, db *sqldb.Database) {
	// Login tests typically don't need any module data
	// Just creating a namespace for basic testing
	_ = integrationTestUtils.CreateNamespace(t, db, "login-test", nil)
}
