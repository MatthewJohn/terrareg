package testutils

import (
	"testing"

	analyticsCmd "github.com/matthewjohn/terrareg/terrareg-go/internal/application/command/analytics"
	namespaceCmd "github.com/matthewjohn/terrareg/terrareg-go/internal/application/command/namespace"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/application/query/analytics"
	auditQuery "github.com/matthewjohn/terrareg/terrareg-go/internal/application/query/audit"
	moduleQuery "github.com/matthewjohn/terrareg/terrareg-go/internal/application/query/module"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/application/query/namespace"
	auditService "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/audit/service"
	moduleService "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/service"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
	auditRepo "github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb/audit"
	"github.com/stretchr/testify/require"
)

// TestApplicationServices holds all common application services (queries/commands)
type TestApplicationServices struct {
	// Module Queries
	ListNamespaces *moduleQuery.ListNamespacesQuery
	ListModules    *moduleQuery.ListModulesQuery

	// Namespace Queries
	NamespaceDetails *namespace.NamespaceDetailsQuery

	// Namespace Commands
	CreateNamespace *namespaceCmd.CreateNamespaceCommand
	UpdateNamespace *namespaceCmd.UpdateNamespaceCommand
	DeleteNamespace *namespaceCmd.DeleteNamespaceCommand

	// Analytics Queries
	GlobalStats               *analytics.GlobalStatsQuery
	GlobalUsageStats          *analytics.GlobalUsageStatsQuery
	GetDownloadSummary        *analytics.GetDownloadSummaryQuery
	GetMostRecentlyPublished  *analytics.GetMostRecentlyPublishedQuery
	GetMostDownloadedThisWeek *analytics.GetMostDownloadedThisWeekQuery
	GetTokenVersions          *analytics.GetTokenVersionsQuery

	// Analytics Commands
	RecordModuleDownload *analyticsCmd.RecordModuleDownloadCommand

	// Audit Queries
	GetAuditHistory *auditQuery.GetAuditHistoryQuery
}

// CreateTestApplicationServices creates all common application services with consistent config
func CreateTestApplicationServices(t *testing.T, db *sqldb.Database, repos *TestRepositories, opts ...ConfigOption) *TestApplicationServices {
	cfg := CreateTestDomainConfigWith(t, opts...)

	// Create module/namespace services and queries
	listNamespacesQuery := moduleQuery.NewListNamespacesQuery(repos.Namespace)
	namespaceSvc := moduleService.NewNamespaceService(cfg)
	namespaceDetailsQuery := namespace.NewNamespaceDetailsQuery(repos.Namespace, namespaceSvc)

	listModulesQuery := moduleQuery.NewListModulesQuery(repos.ModuleProvider)

	// Create analytics queries
	globalStatsQuery := analytics.NewGlobalStatsQuery(repos.Namespace, repos.ModuleProvider, repos.Analytics)
	globalUsageStatsQuery := analytics.NewGlobalUsageStatsQuery(repos.ModuleProvider, repos.Analytics)
	getDownloadSummaryQuery := analytics.NewGetDownloadSummaryQuery(repos.Analytics)
	getMostRecentlyPublishedQuery := analytics.NewGetMostRecentlyPublishedQuery(repos.Analytics)
	getMostDownloadedThisWeekQuery := analytics.NewGetMostDownloadedThisWeekQuery(repos.Analytics)
	getTokenVersionsQuery := analytics.NewGetTokenVersionsQuery(repos.Analytics)

	// Create analytics commands
	recordModuleDownloadCmd := analyticsCmd.NewRecordModuleDownloadCommand(repos.ModuleProvider, repos.Analytics)

	// Create audit service and query
	auditHistoryRepo, err := auditRepo.NewAuditHistoryRepository(db.DB)
	require.NoError(t, err, "Failed to create AuditHistoryRepository")
	auditSvc := auditService.NewAuditService(auditHistoryRepo)
	getAuditHistoryQuery := auditQuery.NewGetAuditHistoryQuery(auditSvc)

	// Create namespace audit service
	namespaceAuditService := auditService.NewNamespaceAuditService(auditHistoryRepo)

	// Create namespace commands
	createNamespaceCmd := namespaceCmd.NewCreateNamespaceCommand(repos.Namespace, namespaceAuditService)
	updateNamespaceCmd := namespaceCmd.NewUpdateNamespaceCommand(repos.Namespace, namespaceAuditService)
	deleteNamespaceCmd := namespaceCmd.NewDeleteNamespaceCommand(repos.Namespace, repos.ModuleProvider, repos.Provider, namespaceAuditService)

	return &TestApplicationServices{
		// Module Queries
		ListNamespaces:   listNamespacesQuery,
		ListModules:      listModulesQuery,
		NamespaceDetails: namespaceDetailsQuery,

		// Namespace Commands
		CreateNamespace: createNamespaceCmd,
		UpdateNamespace: updateNamespaceCmd,
		DeleteNamespace: deleteNamespaceCmd,

		// Analytics Queries
		GlobalStats:               globalStatsQuery,
		GlobalUsageStats:          globalUsageStatsQuery,
		GetDownloadSummary:        getDownloadSummaryQuery,
		GetMostRecentlyPublished:  getMostRecentlyPublishedQuery,
		GetMostDownloadedThisWeek: getMostDownloadedThisWeekQuery,
		GetTokenVersions:          getTokenVersionsQuery,

		// Analytics Commands
		RecordModuleDownload: recordModuleDownloadCmd,

		// Audit Queries
		GetAuditHistory: getAuditHistoryQuery,
	}
}

// CreateNamespaceService creates a NamespaceService with test config
func CreateNamespaceService(t *testing.T, opts ...ConfigOption) *moduleService.NamespaceService {
	cfg := CreateTestDomainConfigWith(t, opts...)
	return moduleService.NewNamespaceService(cfg)
}
