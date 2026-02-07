package terrareg_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	moduleQuery "github.com/matthewjohn/terrareg/terrareg-go/internal/application/query/module"
	moduleRepo "github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb/module"
	"github.com/matthewjohn/terrareg/terrareg-go/test/integration/testutils"
)

// TestGetIntegrationsQuery_UsesFrontendID verifies all integration URLs use FrontendID format
// instead of numeric database IDs.
// Python reference: /app/test/selenium/test_module_provider.py:1234-1291
func TestGetIntegrationsQuery_UsesFrontendID(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create test data with distinct names (module != provider)
	namespace := testutils.CreateNamespace(t, db, "mycompany", nil)
	_ = testutils.CreateModuleProvider(t, db, namespace.ID, "vpc-prod", "aws")

	// Create query
	namespaceRepository := moduleRepo.NewNamespaceRepository(db.DB)
	domainConfig := testutils.CreateTestDomainConfig(t)
	moduleProviderRepository, err := moduleRepo.NewModuleProviderRepository(
		db.DB, namespaceRepository, domainConfig,
	)
	require.NoError(t, err)

	getIntegrationsQuery := moduleQuery.NewGetIntegrationsQuery(moduleProviderRepository)

	// Execute query
	ctx := context.Background()
	integrations, err := getIntegrationsQuery.Execute(ctx, "mycompany", "vpc-prod", "aws")
	require.NoError(t, err)

	// Verify integrations exist
	assert.GreaterOrEqual(t, len(integrations), 5, "Expected at least 5 integrations")

	// Build expected FrontendID: namespace/module/provider
	expectedFrontendID := "mycompany/vpc-prod/aws"

	// Helper to verify URL contains FrontendID, not numeric ID
	assertURLContainsFrontendID := func(integration moduleQuery.Integration, expectedEndpoint string) {
		url := integration.URL

		// Verify URL contains FrontendID format
		expectedPrefix := "/v1/terrareg/modules/" + expectedFrontendID
		assert.Contains(t, url, expectedPrefix,
			"URL should contain FrontendID '%s', got: %s", expectedPrefix, url)

		// Verify URL does NOT contain numeric ID (like /123/)
		assert.NotContains(t, url, "/v1/terrareg/modules/[0-9]+/",
			"URL should not contain numeric ID pattern, got: %s", url)

		// Verify exact endpoint pattern
		assert.Contains(t, url, expectedEndpoint,
			"URL should contain endpoint '%s', got: %s", expectedEndpoint, url)
	}

	// Build integration map by description for easier testing
	integrationsMap := make(map[string]moduleQuery.Integration)
	for _, integration := range integrations {
		integrationsMap[integration.Description] = integration
	}

	// Test Import integration
	importInt, ok := integrationsMap["Trigger module version import"]
	require.True(t, ok, "Import integration missing")
	assertURLContainsFrontendID(importInt, "${version}/import")
	assert.Equal(t, "POST", *importInt.Method)
	assert.Nil(t, importInt.ComingSoon)

	// Test Upload integration
	uploadInt, ok := integrationsMap["Create module version using source archive"]
	require.True(t, ok, "Upload integration missing")
	assertURLContainsFrontendID(uploadInt, "${version}/upload")
	assert.Equal(t, "POST", *uploadInt.Method)
	assert.Nil(t, uploadInt.ComingSoon)

	// Test Publish integration
	publishInt, ok := integrationsMap["Mark a module version as published"]
	require.True(t, ok, "Publish integration missing")
	assertURLContainsFrontendID(publishInt, "${version}/publish")
	assert.Equal(t, "POST", *publishInt.Method)
	assert.Nil(t, publishInt.ComingSoon)

	// Test GitHub webhook
	githubInt, ok := integrationsMap["GitHub hook trigger"]
	require.True(t, ok, "GitHub webhook integration missing")
	assertURLContainsFrontendID(githubInt, "hooks/github")
	assert.Nil(t, githubInt.Method, "Webhook should have no method")

	// Test Bitbucket webhook
	bitbucketInt, ok := integrationsMap["Bitbucket hook trigger"]
	require.True(t, ok, "Bitbucket webhook integration missing")
	assertURLContainsFrontendID(bitbucketInt, "hooks/bitbucket")
	assert.Nil(t, bitbucketInt.Method)

	// Test GitLab webhook (coming soon)
	gitlabInt, ok := integrationsMap["Gitlab hook trigger"]
	require.True(t, ok, "GitLab webhook integration missing")
	assertURLContainsFrontendID(gitlabInt, "hooks/gitlab")
	assert.Nil(t, gitlabInt.Method)
	require.NotNil(t, gitlabInt.ComingSoon, "GitLab should have coming_soon field")
	assert.True(t, *gitlabInt.ComingSoon, "GitLab should be marked as coming_soon")
}

// TestGetIntegrationsQuery_URLStructure verifies complete URL structure
func TestGetIntegrationsQuery_URLStructure(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	namespace := testutils.CreateNamespace(t, db, "mycompany", nil)
	_ = testutils.CreateModuleProvider(t, db, namespace.ID, "network", "google")

	namespaceRepository := moduleRepo.NewNamespaceRepository(db.DB)
	domainConfig := testutils.CreateTestDomainConfig(t)
	moduleProviderRepository, err := moduleRepo.NewModuleProviderRepository(
		db.DB, namespaceRepository, domainConfig,
	)
	require.NoError(t, err)

	getIntegrationsQuery := moduleQuery.NewGetIntegrationsQuery(moduleProviderRepository)

	ctx := context.Background()
	integrations, err := getIntegrationsQuery.Execute(ctx, "mycompany", "network", "google")
	require.NoError(t, err)

	// Build expected FrontendID
	expectedFrontendID := "mycompany/network/google"
	expectedPrefix := "/v1/terrareg/modules/" + expectedFrontendID

	// Verify all URLs follow expected pattern
	for _, integration := range integrations {
		url := integration.URL
		require.True(t, len(url) > 0, "URL field should not be empty")

		// All URLs should start with the expected prefix
		assert.True(t, len(url) >= len(expectedPrefix) && url[:len(expectedPrefix)] == expectedPrefix,
			"URL should start with '%s', got: %s", expectedPrefix, url)

		// URLs should NOT contain numeric IDs
		assert.NotContains(t, url, "/v1/terrareg/modules/[0-9]+/",
			"URL should not contain numeric ID pattern, got: %s", url)
	}
}

// TestGetIntegrationsQuery_GitLabComingSoon verifies GitLab integration is marked coming soon
func TestGetIntegrationsQuery_GitLabComingSoon(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	namespace := testutils.CreateNamespace(t, db, "gruntwork-io", nil)
	_ = testutils.CreateModuleProvider(t, db, namespace.ID, "security", "aws")

	namespaceRepository := moduleRepo.NewNamespaceRepository(db.DB)
	domainConfig := testutils.CreateTestDomainConfig(t)
	moduleProviderRepository, err := moduleRepo.NewModuleProviderRepository(
		db.DB, namespaceRepository, domainConfig,
	)
	require.NoError(t, err)

	getIntegrationsQuery := moduleQuery.NewGetIntegrationsQuery(moduleProviderRepository)

	ctx := context.Background()
	integrations, err := getIntegrationsQuery.Execute(ctx, "gruntwork-io", "security", "aws")
	require.NoError(t, err)

	// Find GitLab integration
	var gitlabIntegration *moduleQuery.Integration
	for i := range integrations {
		if integrations[i].Description == "Gitlab hook trigger" {
			gitlabIntegration = &integrations[i]
			break
		}
	}

	// Verify GitLab integration exists and has coming_soon: true
	require.NotNil(t, gitlabIntegration, "GitLab integration should exist")
	require.NotNil(t, gitlabIntegration.ComingSoon, "GitLab should have coming_soon field")
	assert.True(t, *gitlabIntegration.ComingSoon, "GitLab should be marked as coming_soon")

	// Verify other webhooks don't have coming_soon
	for _, integration := range integrations {
		if integration.Description != "Gitlab hook trigger" {
			if integration.ComingSoon != nil && *integration.ComingSoon {
				t.Errorf("Only GitLab should have coming_soon set, got: true (description: %s)", integration.Description)
			}
		}
	}
}
