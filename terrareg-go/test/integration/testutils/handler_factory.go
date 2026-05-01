package testutils

import (
	"testing"

	moduleQuery "github.com/matthewjohn/terrareg/terrareg-go/internal/application/query/module"
	namespaceService "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/service"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/url/service"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
	analyticsRepo "github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb/analytics"
	v1 "github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/handler/terraform/v1"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/handler/terrareg"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/presenter"
	"github.com/stretchr/testify/require"
)

// CreateNamespaceHandler creates a fully configured NamespaceHandler
func CreateNamespaceHandler(t *testing.T, db *sqldb.Database, opts ...ConfigOption) *terrareg.NamespaceHandler {
	repos := CreateTestRepositories(t, db, opts...)
	services := CreateTestApplicationServices(t, db, repos, opts...)

	handler, err := terrareg.NewNamespaceHandler(
		services.ListNamespaces,
		services.CreateNamespace,
		services.UpdateNamespace,
		services.DeleteNamespace,
		services.NamespaceDetails,
	)
	require.NoError(t, err, "Failed to create NamespaceHandler")
	return handler
}

// CreateModuleListHandler creates a fully configured ModuleListHandler (v1 Terraform)
func CreateModuleListHandler(t *testing.T, db *sqldb.Database, opts ...ConfigOption) *v1.ModuleListHandler {
	repos := CreateTestRepositories(t, db, opts...)
	services := CreateTestApplicationServices(t, db, repos, opts...)

	return v1.NewModuleListHandler(services.ListModules)
}

// CreateAnalyticsHandler creates a fully configured AnalyticsHandler
func CreateAnalyticsHandler(t *testing.T, db *sqldb.Database, opts ...ConfigOption) *terrareg.AnalyticsHandler {
	repos := CreateTestRepositories(t, db, opts...)
	services := CreateTestApplicationServices(t, db, repos, opts...)

	return terrareg.NewAnalyticsHandler(
		services.GlobalStats,
		services.GlobalUsageStats,
		services.GetDownloadSummary,
		services.RecordModuleDownload,
		services.GetMostRecentlyPublished,
		services.GetMostDownloadedThisWeek,
		services.GetTokenVersions,
	)
}

// CreateAuditHandler creates a fully configured AuditHandler
func CreateAuditHandler(t *testing.T, db *sqldb.Database, opts ...ConfigOption) *terrareg.AuditHandler {
	repos := CreateTestRepositories(t, db, opts...)
	services := CreateTestApplicationServices(t, db, repos, opts...)

	handler := terrareg.NewAuditHandler(services.GetAuditHistory)
	return handler
}

// CreateModuleVersionsHandler creates a fully configured ModuleHandler for testing module versions endpoint
// Uses NewModuleReadHandlerForTesting which only requires the queries needed for read operations
// Python reference: /app/test/unit/terrareg/server/test_api_module_versions.py
func CreateModuleVersionsHandler(t *testing.T, db *sqldb.Database, opts ...ConfigOption) *terrareg.ModuleHandler {
	repos := CreateTestRepositories(t, db, opts...)

	// Create the queries needed for read operations
	// These match the parameters expected by NewModuleReadHandlerForTesting
	listModulesQuery := moduleQuery.NewListModulesQuery(repos.ModuleProvider)

	searchModulesQuery, err := moduleQuery.NewSearchModulesQuery(repos.ModuleProvider)
	require.NoError(t, err, "Failed to create SearchModulesQuery")

	getModuleProviderQuery := moduleQuery.NewGetModuleProviderQuery(repos.ModuleProvider)
	listModuleProvidersQuery := moduleQuery.NewListModuleProvidersQuery(repos.ModuleProvider)

	// Use the simpler testing constructor that only requires read operation dependencies
	return terrareg.NewModuleReadHandlerForTesting(
		listModulesQuery,
		searchModulesQuery,
		getModuleProviderQuery,
		listModuleProvidersQuery,
		nil, // analyticsRepo - not needed for module versions endpoint
		nil, // versionPresenter - not needed for module versions endpoint
	)
}

// CreateTerraregModuleDetailsHandler creates a fully configured ModuleHandler for testing Terrareg module provider details endpoint
// Uses NewModuleReadHandlerForTesting which only requires the queries needed for read operations
// Python reference: /app/test/unit/terrareg/server/test_api_terrareg_module_provider_details.py
func CreateTerraregModuleDetailsHandler(t *testing.T, db *sqldb.Database, opts ...ConfigOption) *terrareg.ModuleHandler {
	repos := CreateTestRepositories(t, db, opts...)

	// Create the queries needed for read operations
	listModulesQuery := moduleQuery.NewListModulesQuery(repos.ModuleProvider)

	searchModulesQuery, err := moduleQuery.NewSearchModulesQuery(repos.ModuleProvider)
	require.NoError(t, err, "Failed to create SearchModulesQuery")

	getModuleProviderQuery := moduleQuery.NewGetModuleProviderQuery(repos.ModuleProvider)
	listModuleProvidersQuery := moduleQuery.NewListModuleProvidersQuery(repos.ModuleProvider)

	// Create namespace service and analytics repository (needed for versionPresenter)
	domainConfig := CreateTestDomainConfig(t)
	namespaceSvc := namespaceService.NewNamespaceService(domainConfig)

	analyticsRepository, err := analyticsRepo.NewAnalyticsRepository(db.DB, repos.Namespace, namespaceSvc)
	require.NoError(t, err, "Failed to create AnalyticsRepository")

	// Create URLService and versionPresenter for full Terrareg response
	infraConfig := CreateTestInfraConfig(t)
	urlService, err := service.NewURLService(infraConfig)
	require.NoError(t, err)
	versionPresenter := presenter.NewModuleVersionPresenter(namespaceSvc, analyticsRepository, urlService)

	// Use the simpler testing constructor that only requires read operation dependencies
	return terrareg.NewModuleReadHandlerForTesting(
		listModulesQuery,
		searchModulesQuery,
		getModuleProviderQuery,
		listModuleProvidersQuery,
		analyticsRepository,
		versionPresenter, // Pass versionPresenter for full Terrareg response
	)
}

// CreateModuleVersionDetailsHandler creates a fully configured ModuleHandler for testing module version details endpoint
// Python reference: /app/test/unit/terrareg/server/test_api_module_version_details.py
func CreateModuleVersionDetailsHandler(t *testing.T, db *sqldb.Database, opts ...ConfigOption) *terrareg.ModuleHandler {
	repos := CreateTestRepositories(t, db, opts...)

	getModuleProviderQuery := moduleQuery.NewGetModuleProviderQuery(repos.ModuleProvider)

	getModuleVersionQuery, err := moduleQuery.NewGetModuleVersionQuery(repos.ModuleProvider)
	require.NoError(t, err, "Failed to create GetModuleVersionQuery")

	// Create namespace service and analytics repository (needed for versionPresenter)
	domainConfig := CreateTestDomainConfig(t)
	namespaceSvc := namespaceService.NewNamespaceService(domainConfig)

	analyticsRepository, err := analyticsRepo.NewAnalyticsRepository(db.DB, repos.Namespace, namespaceSvc)
	require.NoError(t, err, "Failed to create AnalyticsRepository")

	// Create URLService and versionPresenter for full Terrareg response
	infraConfig := CreateTestInfraConfig(t)
	urlService, err := service.NewURLService(infraConfig)
	require.NoError(t, err)
	versionPresenter := presenter.NewModuleVersionPresenter(namespaceSvc, analyticsRepository, urlService)

	// Create handler with all dependencies - using the specialized constructor
	return terrareg.NewModuleVersionDetailsHandlerForTesting(
		getModuleVersionQuery,
		getModuleProviderQuery,
		analyticsRepository,
		versionPresenter,
	)
}
