package container

import (
	"github.com/rs/zerolog"
	"gorm.io/gorm"

	analyticsCmd "github.com/matthewjohn/terrareg/terrareg-go/internal/application/command/analytics"
	authCmd "github.com/matthewjohn/terrareg/terrareg-go/internal/application/command/auth"
	moduleCmd "github.com/matthewjohn/terrareg/terrareg-go/internal/application/command/module"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/application/command/namespace"
	providerCmd "github.com/matthewjohn/terrareg/terrareg-go/internal/application/command/provider"
	analyticsQuery "github.com/matthewjohn/terrareg/terrareg-go/internal/application/query/analytics"
	authQuery "github.com/matthewjohn/terrareg/terrareg-go/internal/application/query/auth"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/application/query/module"
	moduleQuery "github.com/matthewjohn/terrareg/terrareg-go/internal/application/query/module"
	providerQuery "github.com/matthewjohn/terrareg/terrareg-go/internal/application/query/provider"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/config"
	authRepo "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth/repository"
	gitService "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/git/service"
	moduleRepo "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/repository"
	moduleService "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/service" // Alias for the new module service
	providerRepo "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider/repository"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/git"

	providerRepository "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider/repository"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/parser"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
	analyticsPersistence "github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb/analytics"
	authPersistence "github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb/auth"
	modulePersistence "github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb/module"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/storage"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http"
	v1 "github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/handler/terraform/v1"
	v2 "github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/handler/terraform/v2"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/handler/terrareg"
	terrareg_middleware "github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/middleware"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/template"
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
	ProviderRepo       providerRepo.ProviderRepository
	SessionRepo        authRepo.SessionRepository

	// Infrastructure Services
	GitClient      gitService.GitClient
	StorageService moduleService.StorageService
	ModuleParser   moduleService.ModuleParser

	// Domain Services
	ModuleImporterService *moduleService.ModuleImporterService

	// Commands
	CreateNamespaceCmd              *namespace.CreateNamespaceCommand
	CreateModuleProviderCmd         *moduleCmd.CreateModuleProviderCommand
	PublishModuleVersionCmd         *moduleCmd.PublishModuleVersionCommand
	UpdateModuleProviderSettingsCmd *moduleCmd.UpdateModuleProviderSettingsCommand
	DeleteModuleProviderCmd         *moduleCmd.DeleteModuleProviderCommand
	UploadModuleVersionCmd          *moduleCmd.UploadModuleVersionCommand
	ImportModuleVersionCmd          *moduleCmd.ImportModuleVersionCommand
	RecordModuleDownloadCmd         *analyticsCmd.RecordModuleDownloadCommand
	CreateAdminSessionCmd           *authCmd.CreateAdminSessionCommand

	// Provider Commands
	CreateOrUpdateProviderCmd *providerCmd.CreateOrUpdateProviderCommand
	PublishProviderVersionCmd *providerCmd.PublishProviderVersionCommand
	ManageGPGKeyCmd           *providerCmd.ManageGPGKeyCommand

	// Provider Queries
	GetProviderVersionQuery *providerQuery.GetProviderVersionQuery

	// Queries
	ListNamespacesQuery            *module.ListNamespacesQuery
	ListModulesQuery               *module.ListModulesQuery
	SearchModulesQuery             *module.SearchModulesQuery
	GetModuleProviderQuery         *module.GetModuleProviderQuery
	ListModuleProvidersQuery       *module.ListModuleProvidersQuery
	GetModuleVersionQuery          *module.GetModuleVersionQuery
	GetModuleDownloadQuery         *module.GetModuleDownloadQuery
	GetModuleProviderSettingsQuery *module.GetModuleProviderSettingsQuery
	ListModuleVersionsQuery        *module.ListModuleVersionsQuery
	GetSubmodulesQuery             *module.GetSubmodulesQuery
	GetExamplesQuery               *module.GetExamplesQuery
	GlobalStatsQuery               *analyticsQuery.GlobalStatsQuery
	GetDownloadSummaryQuery        *analyticsQuery.GetDownloadSummaryQuery
	GetMostRecentlyPublishedQuery  *analyticsQuery.GetMostRecentlyPublishedQuery
	GetMostDownloadedThisWeekQuery *analyticsQuery.GetMostDownloadedThisWeekQuery
	ListProvidersQuery             *providerQuery.ListProvidersQuery
	SearchProvidersQuery           *providerQuery.SearchProvidersQuery
	GetProviderQuery               *providerQuery.GetProviderQuery
	GetProviderVersionsQuery       *providerQuery.GetProviderVersionsQuery
	CheckSessionQuery              *authQuery.CheckSessionQuery

	// Handlers
	NamespaceHandler         *terrareg.NamespaceHandler
	ModuleHandler            *terrareg.ModuleHandler
	AnalyticsHandler         *terrareg.AnalyticsHandler
	ProviderHandler          *terrareg.ProviderHandler
	AuthHandler              *terrareg.AuthHandler
	TerraformV1ModuleHandler *v1.TerraformV1ModuleHandler // New V1 Terraform Module Handler
	TerraformV2ProviderHandler *v2.TerraformV2ProviderHandler
	TerraformV2CategoryHandler *v2.TerraformV2CategoryHandler
	TerraformV2GPGHandler *v2.TerraformV2GPGHandler

	// Middleware
	AuthMiddleware *terrareg_middleware.AuthMiddleware

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
	c.ProviderRepo = providerRepository.NewProviderRepository()
	c.SessionRepo = authPersistence.NewSessionRepository(db.DB)

	// Initialize infrastructure services
	c.GitClient = git.NewGitClientImpl()
	c.StorageService = storage.NewLocalStorage()
	c.ModuleParser = parser.NewModuleParserImpl(c.StorageService)

	// Initialize domain services
	c.ModuleImporterService = moduleService.NewModuleImporterService(
		c.ModuleProviderRepo,
		c.GitClient,
		c.StorageService,
		c.ModuleParser,
		cfg,
	)

	// Initialize commands
	c.CreateNamespaceCmd = namespace.NewCreateNamespaceCommand(c.NamespaceRepo)
	c.CreateModuleProviderCmd = moduleCmd.NewCreateModuleProviderCommand(c.NamespaceRepo, c.ModuleProviderRepo)
	c.PublishModuleVersionCmd = moduleCmd.NewPublishModuleVersionCommand(c.ModuleProviderRepo)
	c.UpdateModuleProviderSettingsCmd = moduleCmd.NewUpdateModuleProviderSettingsCommand(c.ModuleProviderRepo)
	c.DeleteModuleProviderCmd = moduleCmd.NewDeleteModuleProviderCommand(c.ModuleProviderRepo)
	c.UploadModuleVersionCmd = moduleCmd.NewUploadModuleVersionCommand(c.ModuleProviderRepo, c.ModuleParser, c.StorageService, cfg)
	c.ImportModuleVersionCmd = moduleCmd.NewImportModuleVersionCommand(c.ModuleImporterService)
	c.RecordModuleDownloadCmd = analyticsCmd.NewRecordModuleDownloadCommand(c.ModuleProviderRepo, c.AnalyticsRepo)
	c.CreateAdminSessionCmd = authCmd.NewCreateAdminSessionCommand(c.SessionRepo, cfg)

	// Initialize queries
	c.ListNamespacesQuery = module.NewListNamespacesQuery(c.NamespaceRepo)
	c.ListModulesQuery = module.NewListModulesQuery(c.ModuleProviderRepo)
	c.SearchModulesQuery = module.NewSearchModulesQuery(c.ModuleProviderRepo)
	c.GetModuleProviderQuery = module.NewGetModuleProviderQuery(c.ModuleProviderRepo)
	c.ListModuleProvidersQuery = moduleQuery.NewListModuleProvidersQuery(c.ModuleProviderRepo)
	c.GetModuleVersionQuery = moduleQuery.NewGetModuleVersionQuery(c.ModuleProviderRepo)
	c.ListModuleVersionsQuery = moduleQuery.NewListModuleVersionsQuery(c.ModuleProviderRepo) // New query
	c.GetModuleDownloadQuery = moduleQuery.NewGetModuleDownloadQuery(c.ModuleProviderRepo)
	c.GetModuleProviderSettingsQuery = moduleQuery.NewGetModuleProviderSettingsQuery(c.ModuleProviderRepo)
	c.GetSubmodulesQuery = moduleQuery.NewGetSubmodulesQuery(c.ModuleProviderRepo, c.ModuleParser, cfg)
	c.GetExamplesQuery = moduleQuery.NewGetExamplesQuery(c.ModuleProviderRepo, c.ModuleParser, cfg)
	c.GlobalStatsQuery = analyticsQuery.NewGlobalStatsQuery(c.NamespaceRepo, c.ModuleProviderRepo)
	c.GetDownloadSummaryQuery = analyticsQuery.NewGetDownloadSummaryQuery(c.AnalyticsRepo)
	c.GetMostRecentlyPublishedQuery = analyticsQuery.NewGetMostRecentlyPublishedQuery(c.AnalyticsRepo)
	c.GetMostDownloadedThisWeekQuery = analyticsQuery.NewGetMostDownloadedThisWeekQuery(c.AnalyticsRepo)
	c.ListProvidersQuery = providerQuery.NewListProvidersQuery(c.ProviderRepo)
	c.SearchProvidersQuery = providerQuery.NewSearchProvidersQuery(c.ProviderRepo)
	c.GetProviderQuery = providerQuery.NewGetProviderQuery(c.ProviderRepo)
	c.GetProviderVersionsQuery = providerQuery.NewGetProviderVersionsQuery(c.ProviderRepo)
	c.GetProviderVersionQuery = providerQuery.NewGetProviderVersionQuery(c.ProviderRepo)
	c.GetProviderVersionQuery = providerQuery.NewGetProviderVersionQuery(c.ProviderRepo)
	c.CreateOrUpdateProviderCmd = providerCmd.NewCreateOrUpdateProviderCommand(c.ProviderRepo, c.NamespaceRepo)
	c.PublishProviderVersionCmd = providerCmd.NewPublishProviderVersionCommand(c.ProviderRepo, c.NamespaceRepo)
	c.ManageGPGKeyCmd = providerCmd.NewManageGPGKeyCommand(c.ProviderRepo, c.NamespaceRepo)
	c.CheckSessionQuery = authQuery.NewCheckSessionQuery(c.SessionRepo)

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
		c.GetSubmodulesQuery,
		c.GetExamplesQuery,
		c.CreateModuleProviderCmd,
		c.PublishModuleVersionCmd,
		c.UpdateModuleProviderSettingsCmd,
		c.DeleteModuleProviderCmd,
		c.UploadModuleVersionCmd,
		c.ImportModuleVersionCmd,
	)
	c.AnalyticsHandler = terrareg.NewAnalyticsHandler(
		c.GlobalStatsQuery,
		c.GetDownloadSummaryQuery,
		c.RecordModuleDownloadCmd,
		c.GetMostRecentlyPublishedQuery,
		c.GetMostDownloadedThisWeekQuery,
	)
	c.ProviderHandler = terrareg.NewProviderHandler(
		c.ListProvidersQuery,
		c.SearchProvidersQuery,
		c.GetProviderQuery,
		c.GetProviderVersionsQuery,
		c.GetProviderVersionQuery,
		c.CreateOrUpdateProviderCmd,
		c.PublishProviderVersionCmd,
		c.ManageGPGKeyCmd,
	)
	c.AuthHandler = terrareg.NewAuthHandler(
		c.CreateAdminSessionCmd,
		c.CheckSessionQuery,
		cfg,
	)
	c.TerraformV1ModuleHandler = v1.NewTerraformV1ModuleHandler(c.ListModulesQuery, c.SearchModulesQuery, c.GetModuleProviderQuery, c.ListModuleVersionsQuery, c.GetModuleDownloadQuery, c.GetModuleVersionQuery) // Instantiate the new handler

	// Initialize v2 handlers
	c.TerraformV2ProviderHandler = v2.NewTerraformV2ProviderHandler(c.GetProviderQuery, c.GetProviderVersionsQuery, c.GetProviderVersionQuery, c.ListProvidersQuery)
	c.TerraformV2CategoryHandler = v2.NewTerraformV2CategoryHandler(providerQuery.NewListUserSelectableProviderCategoriesQuery(nil)) // TODO: Add proper category repo
	c.TerraformV2GPGHandler = v2.NewTerraformV2GPGHandler()

	// Initialize middleware
	c.AuthMiddleware = terrareg_middleware.NewAuthMiddleware(cfg, c.CheckSessionQuery)

	// Initialize template renderer
	templateRenderer, err := template.NewRenderer(cfg)
	if err != nil {
		return nil, err
	}
	c.TemplateRenderer = templateRenderer

	// Initialize HTTP server
	c.Server = http.NewServer(
		cfg,
		logger,
		c.NamespaceHandler,
		c.ModuleHandler,
		c.AnalyticsHandler,
		c.ProviderHandler,
		c.AuthHandler,
		c.AuthMiddleware,
		c.TemplateRenderer,
		c.TerraformV1ModuleHandler, // Pass the new handler to the server constructor
		c.TerraformV2ProviderHandler,
		c.TerraformV2CategoryHandler,
		c.TerraformV2GPGHandler,
	)

	return c, nil
}

// GetDB returns the database instance
func (c *Container) GetDB() *gorm.DB {
	return c.DB.DB
}
