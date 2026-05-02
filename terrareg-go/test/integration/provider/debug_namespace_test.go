package provider

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	providerquery "github.com/matthewjohn/terrareg/terrareg-go/internal/application/query/provider"
	providerdomainrepo "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider/repository"
	sqldbprovider "github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb/provider"
	testutils "github.com/matthewjohn/terrareg/terrareg-go/test/integration/testutils"
)

func TestDebugProviderNamespace(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	ctx := context.Background()

	// Setup test data
	testutils.SetupComprehensiveProviderSearchTestData(t, db)

	// Setup repository
	providerRepo := sqldbprovider.NewProviderRepository(db.DB)
	searchQuery := providerquery.NewSearchProvidersQuery(providerRepo)

	// Execute search
	result, err := searchQuery.Execute(ctx, providerdomainrepo.ProviderSearchQuery{
		Query:  "mixed",
		Offset: 0,
		Limit:  10,
	})
	require.NoError(t, err)

	t.Logf("Total providers found: %d", result.TotalCount)
	t.Logf("Providers returned: %d", len(result.Providers))

	for i, p := range result.Providers {
		namespace := p.Namespace()
		if namespace == nil {
			t.Logf("  %d. Provider %s (%d): Namespace is nil!", i+1, p.Name(), p.ID())
		} else {
			t.Logf("  %d. Provider %s (%d): Namespace = %s (ID=%d)", i+1, p.Name(), p.ID(), namespace.Name(), namespace.ID())
		}
	}
}
