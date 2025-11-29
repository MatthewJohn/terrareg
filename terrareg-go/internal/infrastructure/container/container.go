package container

import (
	"github.com/rs/zerolog"
	"gorm.io/gorm"

	analyticsCmd "github.com/terrareg/terrareg/internal/application/command/analytics"
	moduleCmd "github.com/terrareg/terrareg/internal/application/command/module"
	"github.com/terrareg/terrareg/internal/application/command/namespace"
	analyticsQuery "github.com/terrareg/terrareg/internal/application/query/analytics"
	"github.com/terrareg/terrareg/internal/application/query/module"
	"github.com/terrareg/terrareg/internal/config"
	moduleRepo "github.com/terrareg/terrareg/internal/domain/module/repository"
	"github.com/terrareg/terrareg/internal/infrastructure/persistence/sqldb"
	analyticsPersistence "github.com/terrareg/terrareg/internal/infrastructure/persistence/sqldb/analytics"
	modulePersistence "github.com/terrareg/terrareg/internal/infrastructure/persistence/sqldb/module"
	"github.com/terrareg/terrareg/internal/interfaces/http"
	"github.com/terrareg/terrareg/internal/interfaces/http/handler/terrareg"
	"github.com/terrareg/terrareg/internal/interfaces/http/template"
)

// Container holds all application dependencies
type Container struct {
	Config *config.Config
	Logger zerolog.Logger
	DB     *sqldb.Database

	// Repositories
	NamespaceRepo      moduleRepo.NamespaceRepository
	ModuleProviderRepo moduleRepo.ModuleProviderRepository
	AnalyticsRepo      analyticsCmd.AnalyticsRepository

	// Commands
	CreateNamespaceCmd              *namespace.CreateNamespaceCommand
	CreateModuleProviderCmd         *moduleCmd.CreateModuleProviderCommand
	PublishModuleVersionCmd         *moduleCmd.PublishModuleVersionCommand
	UpdateModuleProviderSettingsCmd *moduleCmd.UpdateModuleProviderSettingsCommand
	DeleteModuleProviderCmd         *moduleCmd.DeleteModuleProviderCommand
	RecordModuleDownloadCmd         *analyticsCmd.RecordModuleDownloadCommand

	// Queries
	ListNamespacesQuery            *module.ListNamespacesQuery
	ListModulesQuery               *module.ListModulesQuery
	SearchModulesQuery             *module.SearchModulesQuery
	GetModuleProviderQuery         *module.GetModuleProviderQuery
	ListModuleProvidersQuery       *module.ListModuleProvidersQuery
	GetModuleVersionQuery          *module.GetModuleVersionQuery
	GetModuleDownloadQuery         *module.GetModuleDownloadQuery
	GetModuleProviderSettingsQuery *module.GetModuleProviderSettingsQuery
	GlobalStatsQuery               *analyticsQuery.GlobalStatsQuery
	GetDownloadSummaryQuery        *analyticsQuery.GetDownloadSummaryQuery
	GetMostRecentlyPublishedQuery  *analyticsQuery.GetMostRecentlyPublishedQuery
	GetMostDownloadedThisWeekQuery *analyticsQuery.GetMostDownloadedThisWeekQuery

	// Handlers
	NamespaceHandler *terrareg.NamespaceHandler
	ModuleHandler    *terrareg.ModuleHandler
	AnalyticsHandler *terrareg.AnalyticsHandler

	// Template renderer
	TemplateRenderer *template.Renderer

	// HTTP Server
	Server *http.Server
}

// NewContainer creates and initializes the dependency injection container
func NewContainer(cfg *config.Config, logger zerolog.Logger, db *sqldb.Database) (*Container, error) {
	c := &Container{
		Config: cfg,
		Logger: logger,
		DB:     db,
	}

	// Initialize repositories
	c.NamespaceRepo = modulePersistence.NewNamespaceRepository(db.DB)
	c.ModuleProviderRepo = modulePersistence.NewModuleProviderRepository(db.DB, c.NamespaceRepo)
	c.AnalyticsRepo = analyticsPersistence.NewAnalyticsRepository(db.DB)

	// Initialize commands
	c.CreateNamespaceCmd = namespace.NewCreateNamespaceCommand(c.NamespaceRepo)
	c.CreateModuleProviderCmd = moduleCmd.NewCreateModuleProviderCommand(c.NamespaceRepo, c.ModuleProviderRepo)
	c.PublishModuleVersionCmd = moduleCmd.NewPublishModuleVersionCommand(c.ModuleProviderRepo)
	c.UpdateModuleProviderSettingsCmd = moduleCmd.NewUpdateModuleProviderSettingsCommand(c.ModuleProviderRepo)
	c.DeleteModuleProviderCmd = moduleCmd.NewDeleteModuleProviderCommand(c.ModuleProviderRepo)
	c.RecordModuleDownloadCmd = analyticsCmd.NewRecordModuleDownloadCommand(c.ModuleProviderRepo, c.AnalyticsRepo)

	// Initialize queries
	c.ListNamespacesQuery = module.NewListNamespacesQuery(c.NamespaceRepo)
	c.ListModulesQuery = module.NewListModulesQuery(c.ModuleProviderRepo)
	c.SearchModulesQuery = module.NewSearchModulesQuery(c.ModuleProviderRepo)
	c.GetModuleProviderQuery = module.NewGetModuleProviderQuery(c.ModuleProviderRepo)
	c.ListModuleProvidersQuery = module.NewListModuleProvidersQuery(c.ModuleProviderRepo)
	c.GetModuleVersionQuery = module.NewGetModuleVersionQuery(c.ModuleProviderRepo)
	c.GetModuleDownloadQuery = module.NewGetModuleDownloadQuery(c.ModuleProviderRepo)
	c.GetModuleProviderSettingsQuery = module.NewGetModuleProviderSettingsQuery(c.ModuleProviderRepo)
	c.GlobalStatsQuery = analyticsQuery.NewGlobalStatsQuery(c.NamespaceRepo, c.ModuleProviderRepo)
	c.GetDownloadSummaryQuery = analyticsQuery.NewGetDownloadSummaryQuery(c.AnalyticsRepo)
	c.GetMostRecentlyPublishedQuery = analyticsQuery.NewGetMostRecentlyPublishedQuery(c.AnalyticsRepo)
	c.GetMostDownloadedThisWeekQuery = analyticsQuery.NewGetMostDownloadedThisWeekQuery(c.AnalyticsRepo)

	// Initialize handlers
	c.NamespaceHandler = terrareg.NewNamespaceHandler(c.ListNamespacesQuery, c.CreateNamespaceCmd)
	c.ModuleHandler = terrareg.NewModuleHandler(
		c.ListModulesQuery,
		c.SearchModulesQuery,
		c.GetModuleProviderQuery,
		c.ListModuleProvidersQuery,
		c.GetModuleVersionQuery,
		c.GetModuleDownloadQuery,
		c.GetModuleProviderSettingsQuery,
		c.CreateModuleProviderCmd,
		c.PublishModuleVersionCmd,
		c.UpdateModuleProviderSettingsCmd,
		c.DeleteModuleProviderCmd,
	)
	c.AnalyticsHandler = terrareg.NewAnalyticsHandler(
		c.GlobalStatsQuery,
		c.GetDownloadSummaryQuery,
		c.RecordModuleDownloadCmd,
		c.GetMostRecentlyPublishedQuery,
		c.GetMostDownloadedThisWeekQuery,
	)

	// Initialize template renderer
	templateRenderer, err := template.NewRenderer(cfg)
	if err != nil {
		return nil, err
	}
	c.TemplateRenderer = templateRenderer

	// Initialize HTTP server
	c.Server = http.NewServer(cfg, logger, c.NamespaceHandler, c.ModuleHandler, c.AnalyticsHandler, c.TemplateRenderer)

	return c, nil
}

// GetDB returns the database instance
func (c *Container) GetDB() *gorm.DB {
	return c.DB.DB
}
