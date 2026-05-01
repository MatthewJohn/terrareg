package provider

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	providerquery "github.com/matthewjohn/terrareg/terrareg-go/internal/application/query/provider"
	providerdomainrepo "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider/repository"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
	sqldbprovider "github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb/provider"
	testutils "github.com/matthewjohn/terrareg/terrareg-go/test/integration/testutils"
)

func TestDebugSeleniumProviderSearchData(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()

	// Setup repository
	providerRepo := sqldbprovider.NewProviderRepository(db.DB)
	searchQuery := providerquery.NewSearchProvidersQuery(providerRepo)

	// Setup test data using the selenium test data setup
	testutils.SetupComprehensiveProviderSearchTestData(t, db)

	// Check what providers were created
	var providers []sqldb.ProviderDB
	err := db.DB.Find(&providers).Error
	require.NoError(t, err)
	t.Logf("Total providers in DB: %d", len(providers))
	for _, p := range providers {
		t.Logf("  Provider: %s (NS ID: %d, LatestVersionID: %v)", p.Name, p.NamespaceID, p.LatestVersionID)
	}

	// Now search for "mixed"
	result, err := searchQuery.Execute(ctx, providerdomainrepo.ProviderSearchQuery{
		Query:  "mixed",
		Offset: 0,
		Limit:  10,
	})
	require.NoError(t, err)

	t.Logf("Search results - Total: %d, Returned: %d", result.TotalCount, len(result.Providers))
	for i, p := range result.Providers {
		t.Logf("  %d. %s (ID: %d, Namespace: %s)", i+1, p.Name(), p.ID(), p.Namespace().Name())
	}

	t.Logf("VersionData entries: %d", len(result.VersionData))
	for providerID, versionData := range result.VersionData {
		t.Logf("  Provider %d: Version=%s, VersionID=%d", providerID, versionData.Version, versionData.VersionID)
	}
}
