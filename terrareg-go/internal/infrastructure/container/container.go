package container

import (
	"github.com/rs/zerolog"
	"gorm.io/gorm"

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

	// Handlers
	NamespaceHandler *terrareg.NamespaceHandler

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

	// Initialize handlers
	c.NamespaceHandler = terrareg.NewNamespaceHandler(c.ListNamespacesQuery)

	// Initialize template renderer
	templateRenderer, err := template.NewRenderer(cfg)
	if err != nil {
		return nil, err
	}
	c.TemplateRenderer = templateRenderer

	// Initialize HTTP server
	c.Server = http.NewServer(cfg, logger, c.NamespaceHandler, c.TemplateRenderer)

	return c, nil
}

// GetDB returns the database instance
func (c *Container) GetDB() *gorm.DB {
	return c.DB.DB
}
