package testutils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
)

// SetupComprehensiveModuleSearchTestData creates comprehensive module search test data
// matching Python's integration_test_data.py.
// This creates modules in "modulesearch-contributed" and "modulesearch-trusted" namespaces.
func SetupComprehensiveModuleSearchTestData(t *testing.T, db *sqldb.Database) {
	t.Helper()

	published := true

	// ===== modulesearch-contributed namespace (untrusted, not in TRUSTED_NAMESPACES) =====
	modulesearchContributedNs := CreateNamespace(t, db, "modulesearch-contributed", nil)

	// mixedsearch-result/aws (one version)
	provider1 := CreateModuleProvider(t, db, modulesearchContributedNs.ID, "mixedsearch-result", "aws")
	createVersion(t, db, provider1.ID, "1.0.0", &published, nil, "TestOwner1", "")

	// mixedsearch-result-multiversion/aws (multiple versions)
	provider2 := CreateModuleProvider(t, db, modulesearchContributedNs.ID, "mixedsearch-result-multiversion", "aws")
	createVersion(t, db, provider2.ID, "1.2.3", &published, nil, "", "")
	createVersion(t, db, provider2.ID, "2.0.0", &published, nil, "", "")

	// ===== modulesearch-trusted namespace (trusted, in TRUSTED_NAMESPACES) =====
	modulesearchTrustedNs := CreateNamespace(t, db, "modulesearch-trusted", nil)

	// mixedsearch-trusted-result/aws (one version)
	provider3 := CreateModuleProviderWithVerified(t, db, modulesearchTrustedNs.ID, "mixedsearch-trusted-result", "aws", true)
	createVersion(t, db, provider3.ID, "1.0.0", &published, nil, "", "")

	// mixedsearch-trusted-second-result/datadog (one version)
	provider4 := CreateModuleProviderWithVerified(t, db, modulesearchTrustedNs.ID, "mixedsearch-trusted-second-result", "datadog", true)
	createVersion(t, db, provider4.ID, "5.2.1", &published, nil, "", "")

	// mixedsearch-trusted-result-multiversion/null (multiple versions)
	provider5 := CreateModuleProviderWithVerified(t, db, modulesearchTrustedNs.ID, "mixedsearch-trusted-result-multiversion", "null", true)
	createVersion(t, db, provider5.ID, "1.2.3", &published, nil, "", "")
	createVersion(t, db, provider5.ID, "2.0.0", &published, nil, "", "")

	// mixedsearch-trusted-result-verified/gcp (verified, different provider)
	provider6 := CreateModuleProviderWithVerified(t, db, modulesearchTrustedNs.ID, "mixedsearch-trusted-result-verified", "gcp", true)
	createVersion(t, db, provider6.ID, "1.0.0", &published, nil, "", "")
}


// createVersion is a helper to create a module version with common attributes
// It automatically sets the version as the latest version on the module provider
func createVersion(t *testing.T, db *sqldb.Database, moduleProviderID int, version string,
	published, beta *bool, owner, description string) {
	t.Helper()

	moduleVersion := sqldb.ModuleVersionDB{
		ModuleProviderID: moduleProviderID,
		Version:          version,
		Beta:             false,
		Internal:         false,
		Published:        published,
	}

	if owner != "" {
		moduleVersion.Owner = &owner
	}
	if description != "" {
		moduleVersion.Description = &description
	}
	if beta != nil {
		moduleVersion.Beta = *beta
	}

	err := db.DB.Create(&moduleVersion).Error
	require.NoError(t, err)

	// Set this version as the latest version for the module provider
	// Only non-beta versions should be set as latest (matching Python behavior)
	// Beta versions should not overwrite the latest_version_id
	if !moduleVersion.Beta {
		err = db.DB.Model(&sqldb.ModuleProviderDB{}).
			Where("id = ?", moduleProviderID).
			Update("latest_version_id", moduleVersion.ID).Error
		require.NoError(t, err)
	}
}

// CreateModuleProviderWithVerified creates a module provider with specified verified status
func CreateModuleProviderWithVerified(t *testing.T, db *sqldb.Database, namespaceID int, moduleName, providerName string, verified bool) sqldb.ModuleProviderDB {
	t.Helper()

	moduleProvider := sqldb.ModuleProviderDB{
		NamespaceID: namespaceID,
		Module:      moduleName,
		Provider:    providerName,
		Verified:    &verified,
	}

	err := db.DB.Create(&moduleProvider).Error
	require.NoError(t, err)

	return moduleProvider
}

// SetupComprehensiveProviderSearchTestData creates comprehensive provider search test data
// matching Python's integration_test_data.py provider search data.
// Python reference: /app/test/selenium/test_data.py - 'contributed-providersearch' and 'providersearch-trusted'
func SetupComprehensiveProviderSearchTestData(t *testing.T, db *sqldb.Database) {
	t.Helper()

	// Create provider categories (matching Python's integration_provider_categories)
	createProviderCategory(t, db, "Visible Monitoring", "visible-monitoring", true)
	createProviderCategory(t, db, "Second Visible Cloud", "second-visible-cloud", true)

	// Get category IDs
	visibleMonitoringCat := getProviderCategoryBySlug(t, db, "visible-monitoring")
	secondVisibleCloudCat := getProviderCategoryBySlug(t, db, "second-visible-cloud")

	// ===== contributed-providersearch namespace (matching Python test_data.py) =====
	contributedNs := CreateNamespace(t, db, "contributed-providersearch", nil)

	// Create GPG key in namespace
	gpgKeyContributed := createGPGKeyInNamespace(t, db, contributedNs.ID, "D7AA1BEFF16FA788760E54F5591EF84DC5EDCD68")

	// mixedsearch-result (one version)
	provider1 := CreateProvider(t, db, contributedNs.ID, "mixedsearch-result",
		stringPtr("Test Multiple Versions"), sqldb.ProviderTierCommunity, &visibleMonitoringCat)
	createProviderVersion(t, db, provider1.ID, "1.0.0", gpgKeyContributed.ID, false)

	// mixedsearch-result-multiversion (multiple versions)
	provider2 := CreateProvider(t, db, contributedNs.ID, "mixedsearch-result-multiversion",
		stringPtr("Test Multiple Versions"), sqldb.ProviderTierCommunity, &visibleMonitoringCat)
	createProviderVersion(t, db, provider2.ID, "1.2.3", gpgKeyContributed.ID, false)
	createProviderVersion(t, db, provider2.ID, "2.0.0", gpgKeyContributed.ID, false)

	// ===== providersearch-trusted namespace (matching Python test_data.py) =====
	trustedNs := CreateNamespace(t, db, "providersearch-trusted", nil)

	// Create GPG key in namespace
	gpgKeyTrusted := createGPGKeyInNamespace(t, db, trustedNs.ID, "E8B4C3C6FE51E8FC1AFFCC6DEA2F2F9F9989A6E5")

	// mixedsearch-trusted-result (one version)
	provider3 := CreateProvider(t, db, trustedNs.ID, "mixedsearch-trusted-result",
		stringPtr("Test Multiple Versions"), sqldb.ProviderTierCommunity, &visibleMonitoringCat)
	createProviderVersion(t, db, provider3.ID, "1.0.0", gpgKeyTrusted.ID, false)

	// mixedsearch-trusted-result-multiversion (multiple versions)
	provider4 := CreateProvider(t, db, trustedNs.ID, "mixedsearch-trusted-result-multiversion",
		stringPtr("Test Multiple Versions"), sqldb.ProviderTierCommunity, &secondVisibleCloudCat)
	createProviderVersion(t, db, provider4.ID, "1.2.3", gpgKeyTrusted.ID, false)
	createProviderVersion(t, db, provider4.ID, "2.0.0", gpgKeyTrusted.ID, false)

	// mixedsearch-trusted-second-result (one version)
	provider5 := CreateProvider(t, db, trustedNs.ID, "mixedsearch-trusted-second-result",
		stringPtr("Test Multiple Versions"), sqldb.ProviderTierCommunity, &visibleMonitoringCat)
	createProviderVersion(t, db, provider5.ID, "1.0.0", gpgKeyTrusted.ID, false)
}

// createGPGKeyInNamespace creates a GPG key directly in a namespace (not linked to a provider)
func createGPGKeyInNamespace(t *testing.T, db *sqldb.Database, namespaceID int, fingerprint string) sqldb.GPGKeyDB {
	t.Helper()

	asciiArmor := []byte("-----BEGIN PGP PUBLIC KEY BLOCK-----\n\nTest ASCII armor for " + fingerprint + "\n-----END PGP PUBLIC KEY BLOCK-----")
	source := "test-source"
	keyID := &fingerprint

	gpgKey := sqldb.GPGKeyDB{
		NamespaceID: namespaceID,
		ASCIIArmor:  asciiArmor,
		KeyID:       keyID,
		Fingerprint: keyID,
		Source:      &source,
	}

	err := db.DB.Create(&gpgKey).Error
	require.NoError(t, err)

	return gpgKey
}

// CreateProviderWithGPGKey creates a provider with GPG key ID (instead of creating a new GPG key)
func CreateProviderWithGPGKey(t *testing.T, db *sqldb.Database, namespaceID int, name string, description *string, tier sqldb.ProviderTier, categoryID *int, gpgKeyID int) sqldb.ProviderDB {
	t.Helper()

	provider := sqldb.ProviderDB{
		NamespaceID:        namespaceID,
		Name:               name,
		Description:        description,
		Tier:               tier,
		ProviderCategoryID: categoryID,
	}

	err := db.DB.Create(&provider).Error
	require.NoError(t, err)

	return provider
}

// createProviderVersion is a helper to create a provider version with GPG key
// It automatically sets the version as the latest version on the provider
func createProviderVersion(t *testing.T, db *sqldb.Database, providerID int, version string, gpgKeyID int, beta bool) {
	t.Helper()

	gitTag := "v" + version
	publishedAt := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	providerVersion := sqldb.ProviderVersionDB{
		ProviderID:  providerID,
		Version:     version,
		GitTag:      &gitTag,
		GPGKeyID:    gpgKeyID,
		PublishedAt: &publishedAt,
		Beta:        beta,
	}

	err := db.DB.Create(&providerVersion).Error
	require.NoError(t, err)

	// Set this version as the latest version for the provider
	// Only non-beta versions should be set as latest (matching Python behavior)
	// Beta versions should not overwrite the latest_version_id
	if !providerVersion.Beta {
		err = db.DB.Model(&sqldb.ProviderDB{}).
			Where("id = ?", providerID).
			Update("latest_version_id", providerVersion.ID).Error
		require.NoError(t, err)
	}
}

// createProviderCategory creates a provider category
func createProviderCategory(t *testing.T, db *sqldb.Database, name, slug string, userSelectable bool) sqldb.ProviderCategoryDB {
	t.Helper()

	// First, try to find existing category by slug
	var existingCategory sqldb.ProviderCategoryDB
	err := db.DB.Where("slug = ?", slug).First(&existingCategory).Error
	if err == nil {
		// Category already exists, return it
		return existingCategory
	}

	// Category doesn't exist, create it
	namePtr := &name
	category := sqldb.ProviderCategoryDB{
		Name:           namePtr,
		Slug:           slug,
		UserSelectable: userSelectable,
	}

	err = db.DB.Create(&category).Error
	require.NoError(t, err)

	return category
}

// getProviderCategoryBySlug gets a provider category by slug
func getProviderCategoryBySlug(t *testing.T, db *sqldb.Database, slug string) int {
	t.Helper()

	var category sqldb.ProviderCategoryDB
	err := db.DB.Where("slug = ?", slug).First(&category).Error
	require.NoError(t, err)
	return category.ID
}

// stringPtr returns a pointer to a string
func stringPtr(s string) *string {
	return &s
}

// SetupFullyPopulatedModule creates the fullypopulated module from Python's test_data.py
// Python reference: /app/test/unit/terrareg/test_data.py - moduledetails/fullypopulated
// This creates a module with all fields populated for comprehensive testing.
//
// Returns: (namespace, moduleProvider, moduleVersions)
func SetupFullyPopulatedModule(t *testing.T, db *sqldb.Database) (sqldb.NamespaceDB, sqldb.ModuleProviderDB, []sqldb.ModuleVersionDB) {
	t.Helper()

	// Create namespace
	namespace := CreateNamespace(t, db, "moduledetails", nil)

	// Create module provider with all git configuration
	moduleProvider := CreateModuleProvider(t, db, namespace.ID, "fullypopulated", "testprovider")

	// Update module provider with all git configuration
	repoBaseURL := "https://mp-base-url.com/{namespace}/{module}-{provider}"
	repoBrowseURL := "https://mp-browse-url.com/{namespace}/{module}-{provider}/browse/{tag}/{path}suffix"
	repoCloneURL := "ssh://mp-clone-url.com/{namespace}/{module}-{provider}"

	err := db.DB.Model(&moduleProvider).Updates(map[string]interface{}{
		"repo_base_url_template":   &repoBaseURL,
		"repo_browse_url_template": &repoBrowseURL,
		"repo_clone_url_template":  &repoCloneURL,
	}).Error
	require.NoError(t, err)

	published := true
	beta := true
	internal := false
	publishedAt := time.Date(2022, 1, 5, 22, 53, 12, 0, time.UTC)

	// Create all versions from Python test_data.py
	versions := []struct {
		version     string
		published   *bool
		beta        *bool
		internal    *bool
		owner       *string
		description *string
		repoBaseURL *string
		readme      *string
		varTemplate *string
	}{
		// Older version
		{"1.2.0", &published, nil, nil, nil, nil, nil, nil, nil},
		// Newer unpublished version
		{"1.6.0", nil, nil, nil, nil, nil, nil, nil, nil},
		// Newer published beta version
		{"1.6.1-beta", &published, &beta, nil, nil, nil, nil, nil, nil},
		// Unpublished and beta version
		{"1.0.0-beta", nil, &beta, nil, nil, nil, nil, nil, nil},
		// Main fully populated version
		{
			"1.5.0",
			&published,
			nil,
			&internal,
			stringPtr("This is the owner of the module"),
			stringPtr("This is a test module version for tests."),
			stringPtr("https://link-to.com/source-code-here"),
			stringPtr("# This is an exaple README!"),
			stringPtr(`[
				{"name": "name_of_application", "type": "text", "quote_value": true, "additional_help": "Provide the name of the application"},
				{"name": "variable_template_with_markdown", "type": "text", "quote_value": true, "additional_help": "This **is** some _markdown_"},
				{"name": "variable_template_with_html", "type": "text", "quote_value": true, "additional_help": "This <b>is</b> some <i>html</i>"}
			]`),
		},
	}

	createdVersions := make([]sqldb.ModuleVersionDB, 0, len(versions))
	for _, v := range versions {
		moduleVersion := sqldb.ModuleVersionDB{
			ModuleProviderID: moduleProvider.ID,
			Version:          v.version,
			Beta:             false,
			Internal:         false,
			Published:        v.published,
		}

		if v.beta != nil {
			moduleVersion.Beta = *v.beta
		}
		if v.internal != nil {
			moduleVersion.Internal = *v.internal
		}
		if v.owner != nil {
			moduleVersion.Owner = v.owner
		}
		if v.description != nil {
			moduleVersion.Description = v.description
		}
		if v.repoBaseURL != nil {
			moduleVersion.RepoBaseURLTemplate = v.repoBaseURL
		}

		// Set published_at for main version
		if v.version == "1.5.0" {
			moduleVersion.PublishedAt = &publishedAt
		}

		// Create module details if readme or variable template is provided
		if v.readme != nil || v.varTemplate != nil {
			variableTemplate := []byte{}
			if v.varTemplate != nil {
				variableTemplate = []byte(*v.varTemplate)
			}

			readmeContent := []byte{}
			if v.readme != nil {
				readmeContent = []byte(*v.readme)
			}

			moduleDetails := sqldb.ModuleDetailsDB{
				ReadmeContent:    readmeContent,
				TerraformDocs:    []byte("{}"),
				Tfsec:            []byte("{}"),
				Infracost:        []byte("{}"),
				TerraformGraph:   []byte("{}"),
				TerraformModules: []byte("{}"),
				TerraformVersion: []byte("1.0.0"),
			}

			// Store variable template in module version instead
			if v.varTemplate != nil {
				moduleVersion.VariableTemplate = variableTemplate
			}

			err := db.DB.Create(&moduleDetails).Error
			require.NoError(t, err)
			moduleVersion.ModuleDetailsID = &moduleDetails.ID
		}

		err = db.DB.Create(&moduleVersion).Error
		require.NoError(t, err)
		createdVersions = append(createdVersions, moduleVersion)

		// Set version 1.5.0 as the latest version
		if v.version == "1.5.0" {
			err = db.DB.Model(&moduleProvider).Update("latest_version_id", moduleVersion.ID).Error
			require.NoError(t, err)
		}
	}

	return namespace, moduleProvider, createdVersions
}

// SetupTestNamespaceFromPython creates all test modules from Python's testnamespace
// Python reference: /app/test/unit/terrareg/test_data.py - testnamespace
// This includes: testmodulename, lonelymodule, mock-module, unverifiedmodule, internalmodule, etc.
func SetupTestNamespaceFromPython(t *testing.T, db *sqldb.Database) sqldb.NamespaceDB {
	t.Helper()

	namespace := CreateNamespace(t, db, "testnamespace", nil)

	published := true
	beta := true
	internal := true

	// testmodulename/testprovider (ID: 1, Latest: 2.4.1, Verified: true)
	provider1 := CreateModuleProviderWithVerified(t, db, namespace.ID, "testmodulename", "testprovider", true)
	createVersion(t, db, provider1.ID, "2.4.1", &published, nil, "", "")
	createVersion(t, db, provider1.ID, "1.0.0", &published, nil, "", "")

	// lonelymodule/testprovider (ID: 2, Latest: 1.0.0, Verified: true)
	provider2 := CreateModuleProviderWithVerified(t, db, namespace.ID, "lonelymodule", "testprovider", true)
	createVersion(t, db, provider2.ID, "1.0.0", &published, nil, "", "")

	// mock-module/testprovider (ID: 3, Verified: true, Latest: 1.2.3)
	provider3 := CreateModuleProviderWithVerified(t, db, namespace.ID, "mock-module", "testprovider", true)
	createVersion(t, db, provider3.ID, "1.2.3", &published, nil, "", "")

	// unverifiedmodule/testprovider (ID: 16, Verified: false, Latest: 1.2.3)
	provider4 := CreateModuleProviderWithVerified(t, db, namespace.ID, "unverifiedmodule", "testprovider", false)
	createVersion(t, db, provider4.ID, "1.2.3", &published, nil, "", "")

	// internalmodule/testprovider (ID: 17, Verified: false, Latest: 5.2.0, Internal: true)
	provider5 := CreateModuleProviderWithVerified(t, db, namespace.ID, "internalmodule", "testprovider", false)
	createVersionWithInternal(t, db, provider5.ID, "5.2.0", &published, &internal, "", "")

	// modulenorepourl/testprovider (ID: 5, Latest: 2.2.4)
	provider6 := CreateModuleProvider(t, db, namespace.ID, "modulenorepourl", "testprovider")
	createVersion(t, db, provider6.ID, "2.2.4", &published, nil, "", "")

	// onlybeta/testprovider (ID: 18, Latest: 2.2.4-beta, Beta: true)
	provider7 := CreateModuleProvider(t, db, namespace.ID, "onlybeta", "testprovider")
	createVersion(t, db, provider7.ID, "2.2.4-beta", &published, &beta, "", "")

	// modulewithrepourl/testprovider (ID: 6, Latest: 2.1.0, has repo_clone_url_template)
	provider8 := CreateModuleProvider(t, db, namespace.ID, "modulewithrepourl", "testprovider")
	repoCloneURL := "https://github.com/test/test.git"
	err := db.DB.Model(&provider8).Update("repo_clone_url_template", &repoCloneURL).Error
	require.NoError(t, err)
	createVersion(t, db, provider8.ID, "2.1.0", nil, nil, "", "")

	// modulenotpublished/testprovider (ID: 15, Latest: None, all versions unpublished)
	// Also has git configuration templates
	provider9 := CreateModuleProvider(t, db, namespace.ID, "modulenotpublished", "testprovider")
	repoBase := "https://custom-localhost.com/{namespace}/{module}-{provider}"
	repoBrowse := "https://custom-localhost.com/{namespace}/{module}-{provider}/browse/{tag}/{path}"
	repoClone := "ssh://custom-localhost.com/{namespace}/{module}-{provider}"
	err = db.DB.Model(&provider9).Updates(map[string]interface{}{
		"repo_base_url_template":   &repoBase,
		"repo_browse_url_template": &repoBrowse,
		"repo_clone_url_template":  &repoClone,
	}).Error
	require.NoError(t, err)
	createVersion(t, db, provider9.ID, "10.2.1", nil, nil, "", "")

	// withsecurityissues/testprovider (ID: 20, Latest: 1.0.0, has tfsec data)
	provider10 := CreateModuleProvider(t, db, namespace.ID, "withsecurityissues", "testprovider")
	createVersionWithTfsec(t, db, provider10.ID, "1.0.0", &published, nil, "", "")

	// wrongversionorder/testprovider - tests version sorting
	provider11 := CreateModuleProvider(t, db, namespace.ID, "wrongversionorder", "testprovider")
	createVersion(t, db, provider11.ID, "1.5.4", &published, nil, "", "")
	createVersion(t, db, provider11.ID, "2.1.0", &published, nil, "", "")
	createVersion(t, db, provider11.ID, "0.1.1", &published, nil, "", "")
	createVersion(t, db, provider11.ID, "10.23.0", &published, nil, "", "")
	createVersion(t, db, provider11.ID, "0.1.10", &published, nil, "", "")
	createVersion(t, db, provider11.ID, "0.0.9", &published, nil, "", "")
	createVersion(t, db, provider11.ID, "0.1.09", &published, nil, "", "")
	createVersion(t, db, provider11.ID, "0.1.8", &published, nil, "", "")
	createVersion(t, db, provider11.ID, "23.2.3-beta", &published, &beta, "", "")
	createVersion(t, db, provider11.ID, "5.21.2", nil, nil, "", "") // unpublished

	// noversions/testprovider - module with no versions
	_ = CreateModuleProvider(t, db, namespace.ID, "noversions", "testprovider")

	// onlyunpublished/testprovider - module with only unpublished versions
	provider13 := CreateModuleProvider(t, db, namespace.ID, "onlyunpublished", "testprovider")
	createVersion(t, db, provider13.ID, "0.1.8", nil, nil, "", "")

	return namespace
}

// createVersionWithInternal creates a version with internal flag
func createVersionWithInternal(t *testing.T, db *sqldb.Database, moduleProviderID int, version string,
	published, internal *bool, owner, description string) {
	t.Helper()

	moduleVersion := sqldb.ModuleVersionDB{
		ModuleProviderID: moduleProviderID,
		Version:          version,
		Beta:             false,
		Internal:         false,
		Published:        published,
	}

	if internal != nil {
		moduleVersion.Internal = *internal
	}
	if owner != "" {
		moduleVersion.Owner = &owner
	}
	if description != "" {
		moduleVersion.Description = &description
	}

	err := db.DB.Create(&moduleVersion).Error
	require.NoError(t, err)

	// Set this version as the latest version for the module provider
	err = db.DB.Model(&sqldb.ModuleProviderDB{}).
		Where("id = ?", moduleProviderID).
		Update("latest_version_id", moduleVersion.ID).Error
	require.NoError(t, err)
}

// createVersionWithTfsec creates a version with Tfsec security data
// Python reference: /app/test/unit/terrareg/test_data.py - withsecurityissues
func createVersionWithTfsec(t *testing.T, db *sqldb.Database, moduleProviderID int, version string,
	published, beta *bool, owner, description string) {
	t.Helper()

	// Create module details with Tfsec data
	// Includes all fields from tfsec JSON output
	// Python reference: /app/test/selenium/test_data.py withsecurityissues test data
	tfsecJSON := `{
		"results": [
			{
				"description": "Secret explicitly uses the default key.",
				"impact": "Using AWS managed keys reduces the flexibility and control over the encryption key",
				"links": [
					"https://aquasecurity.github.io/tfsec/v1.26.0/checks/aws/ssm/secret-use-customer-key/"
				],
				"location": {
					"end_line": 4,
					"filename": "main.tf",
					"start_line": 2
				},
				"long_id": "aws-ssm-secret-use-customer-key",
				"resolution": "Use customer managed keys",
				"resource": "aws_ssm_parameter.default_key",
				"rule_description": "SSM Parameter secrets should use customer managed keys",
				"rule_id": "AVD-AWS-0098",
				"rule_provider": "aws",
				"rule_service": "ssm",
				"severity": "LOW",
				"status": 0,
				"warning": false
			},
			{
				"description": "Some security issue 2.",
				"impact": "Entire project is compromised",
				"links": [
					"https://example.com/security-issue-2/"
				],
				"location": {
					"end_line": 10,
					"filename": "main.tf",
					"start_line": 6
				},
				"long_id": "bad-code-security-issue",
				"resolution": "Fix the security issue",
				"resource": "bad_resource.example",
				"rule_description": "This is a bad security issue",
				"rule_id": "DDG-ANC-001",
				"rule_provider": "bad",
				"rule_service": "code",
				"severity": "HIGH",
				"status": 0,
				"warning": false
			}
		]
	}`

	moduleDetails := sqldb.ModuleDetailsDB{
		ReadmeContent:    []byte{},
		TerraformDocs:    []byte("{}"),
		Tfsec:            []byte(tfsecJSON),
		Infracost:        []byte("{}"),
		TerraformGraph:   []byte("{}"),
		TerraformModules: []byte("{}"),
		TerraformVersion: []byte("1.0.0"),
	}

	err := db.DB.Create(&moduleDetails).Error
	require.NoError(t, err)

	moduleVersion := sqldb.ModuleVersionDB{
		ModuleProviderID: moduleProviderID,
		Version:          version,
		Beta:             false,
		Internal:         false,
		Published:        published,
		ModuleDetailsID:  &moduleDetails.ID,
	}

	if beta != nil {
		moduleVersion.Beta = *beta
	}
	if owner != "" {
		moduleVersion.Owner = &owner
	}
	if description != "" {
		moduleVersion.Description = &description
	}

	err = db.DB.Create(&moduleVersion).Error
	require.NoError(t, err)

	// Set this version as the latest version for the module provider
	// Only non-beta versions should be set as latest (matching Python behavior)
	// Beta versions should not overwrite the latest_version_id
	if !moduleVersion.Beta {
		err = db.DB.Model(&sqldb.ModuleProviderDB{}).
			Where("id = ?", moduleProviderID).
			Update("latest_version_id", moduleVersion.ID).Error
		require.NoError(t, err)
	}
}

// CreateModuleVersionWithSecurityIssues creates a module version with tfsec security data.
// Returns the created module version for use in tests.
func CreateModuleVersionWithSecurityIssues(t *testing.T, db *sqldb.Database, moduleProviderID int, version string, published *bool) sqldb.ModuleVersionDB {
	t.Helper()
	createVersionWithTfsec(t, db, moduleProviderID, version, published, nil, "", "")

	// Find and return the created module version
	var moduleVersion sqldb.ModuleVersionDB
	err := db.DB.Where("module_provider_id = ? AND version = ?", moduleProviderID, version).First(&moduleVersion).Error
	require.NoError(t, err, "Failed to find created module version")
	return moduleVersion
}
