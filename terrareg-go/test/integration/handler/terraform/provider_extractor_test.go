package terraform_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
	"github.com/matthewjohn/terrareg/terrareg-go/test/integration/testutils"
)

func TestProviderExtractor_DocumentationExtraction(t *testing.T) {
	ctx := context.Background()
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	namespace := testutils.CreateNamespace(t, db, "test-doc-extract", nil)

	// CreateProvider requires: (*testing.T, *sqldb.Database, namespaceID int, name string, description *string, tier sqldb.ProviderTier, categoryID *int)
	description := "Test provider for documentation extraction"
	provider := testutils.CreateProvider(t, db, namespace.ID, "test-provider", &description, sqldb.ProviderTierCommunity, nil)

	category := testutils.CreateProviderCategory(t, db, "Test Category", "test-category", true)

	// CreateProviderVersion requires: (*testing.T, *sqldb.Database, providerID int, version string, gpgKeyID int, beta bool, publishedAt *time.Time)
	providerVersion := testutils.CreateProviderVersion(t, db, provider.ID, "1.0.0", 0, true, nil)

	// Create test provider directory with docs
	tempDir := t.TempDir()
	sourceDir := filepath.Join(tempDir, "provider-docs")
	providerDocsDir := filepath.Join(sourceDir, "docs")
	require.NoError(t, os.MkdirAll(providerDocsDir, 0755))

	// Create overview doc
	overviewContent := `---
subcategory: "testprovider/v1"
page_title: "Test Provider Overview"
description: "A test provider overview for documentation extraction testing"
---

# Test Provider Overview

This provider is used for testing documentation extraction in terrareg.
`
	require.NoError(t, os.WriteFile(filepath.Join(providerDocsDir, "index.md"), []byte(overviewContent), 0644))

	// Create resource subdirectory and doc
	resourceDir := filepath.Join(providerDocsDir, "resources")
	require.NoError(t, os.MkdirAll(resourceDir, 0755))
	resourceContent := `---
subcategory: "testprovider/v1"
page_title: "Test Provider: test_example"
description: "Creates a test example resource for documentation extraction testing"
---

# test_example

Manages a test example resource for documentation extraction testing.
`
	require.NoError(t, os.WriteFile(filepath.Join(resourceDir, "test_example.md"), []byte(resourceContent), 0644))

	// Create datasource subdirectory and doc
	datasourceDir := filepath.Join(providerDocsDir, "data-sources")
	require.NoError(t, os.MkdirAll(datasourceDir, 0755))
	datasourceContent := `---
subcategory: "testprovider/v1"
page_title: "Test Provider: test_example"
description: "Creates a test example data source for documentation extraction testing"
---

# test_example

Manages a test example data source for documentation extraction testing.
`
	require.NoError(t, os.WriteFile(filepath.Join(datasourceDir, "test_example.md"), []byte(datasourceContent), 0644))

	// Create guide subdirectory and doc
	guideDir := filepath.Join(providerDocsDir, "guides")
	require.NoError(t, os.MkdirAll(guideDir, 0755))
	guideContent := `---
subcategory: "testprovider/v1"
page_title: "Test Provider Guide"
description: "A guide example for documentation extraction testing"
---

# Test Provider Guide

This is an example guide for testing documentation extraction in terrareg.
`
	require.NoError(t, os.WriteFile(filepath.Join(guideDir, "example_guide.md"), []byte(guideContent), 0644))

	// TODO: Implement full provider extraction flow
	// The test is currently a placeholder as the extraction orchestrator
	// needs to be implemented to tie together all the individual services.
	//
	// Components implemented:
	// - GPG verification service (provider_extraction_gpg_service.go)
	// - Source extraction service (provider_source_extraction_service.go)
	// - Binary processing service (provider_binary_processing_service.go)
	// - Documentation service (provider_documentation_service.go)
	//
	// Remaining work:
	// - Create extraction orchestrator that uses all services
	// - Wire up in container
	// - Complete integration tests

	_ = ctx
	_ = sourceDir
	_ = providerVersion
	_ = provider
	_ = category
	_ = db
}
