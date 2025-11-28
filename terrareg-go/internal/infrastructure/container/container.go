package container

import (
	"github.com/rs/zerolog"
	"gorm.io/gorm"

	"github.com/terrareg/terrareg/internal/application/query/analytics"
	"github.com/terrareg/terrareg/internal/application/query/module"
	"github.com/terrareg/terrareg/internal/config"
	moduleRepo "github.com/terrareg/terrareg/internal/domain/module/repository"
	"github.com/terrareg/terrareg/internal/infrastructure/persistence/sqldb"
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

	// Queries
	ListNamespacesQuery *module.ListNamespacesQuery
	ListModulesQuery    *module.ListModulesQuery
	SearchModulesQuery  *module.SearchModulesQuery
	GlobalStatsQuery    *analytics.GlobalStatsQuery

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

	// Initialize queries
	c.ListNamespacesQuery = module.NewListNamespacesQuery(c.NamespaceRepo)
	c.ListModulesQuery = module.NewListModulesQuery(c.ModuleProviderRepo)
	c.SearchModulesQuery = module.NewSearchModulesQuery(c.ModuleProviderRepo)
	c.GlobalStatsQuery = analytics.NewGlobalStatsQuery(c.NamespaceRepo, c.ModuleProviderRepo)

	// Initialize handlers
	c.NamespaceHandler = terrareg.NewNamespaceHandler(c.ListNamespacesQuery)
	c.ModuleHandler = terrareg.NewModuleHandler(c.ListModulesQuery, c.SearchModulesQuery)
	c.AnalyticsHandler = terrareg.NewAnalyticsHandler(c.GlobalStatsQuery)

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
