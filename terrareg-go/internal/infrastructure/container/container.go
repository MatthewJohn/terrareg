package container

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog"
	"gorm.io/gorm"

	analyticsCmd "github.com/matthewjohn/terrareg/terrareg-go/internal/application/command/analytics"
	authCmd "github.com/matthewjohn/terrareg/terrareg-go/internal/application/command/auth"
	gpgkeyCmd "github.com/matthewjohn/terrareg/terrareg-go/internal/application/command/gpgkey"
	moduleCmd "github.com/matthewjohn/terrareg/terrareg-go/internal/application/command/module"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/application/command/namespace"
	providerCmd "github.com/matthewjohn/terrareg/terrareg-go/internal/application/command/provider"
	analyticsQuery "github.com/matthewjohn/terrareg/terrareg-go/internal/application/query/analytics"
	auditQuery "github.com/matthewjohn/terrareg/terrareg-go/internal/application/query/audit"
	authQuery "github.com/matthewjohn/terrareg/terrareg-go/internal/application/query/auth"
	configQuery "github.com/matthewjohn/terrareg/terrareg-go/internal/application/query/config"
	gpgkeyQuery "github.com/matthewjohn/terrareg/terrareg-go/internal/application/query/gpgkey"
	graphQuery "github.com/matthewjohn/terrareg/terrareg-go/internal/application/query/graph"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/application/query/module"
	moduleQuery "github.com/matthewjohn/terrareg/terrareg-go/internal/application/query/module"
	namespaceQuery "github.com/matthewjohn/terrareg/terrareg-go/internal/application/query/namespace"
	providerQuery "github.com/matthewjohn/terrareg/terrareg-go/internal/application/query/provider"
	providerLogoQuery "github.com/matthewjohn/terrareg/terrareg-go/internal/application/query/provider_logo"
	setupQuery "github.com/matthewjohn/terrareg/terrareg-go/internal/application/query/setup"
	terraformCmd "github.com/matthewjohn/terrareg/terrareg-go/internal/application/terraform"
	auditRepo "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/audit/repository"
	auditservice "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/audit/service"
	authRepo "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth/repository"
	authservice "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth/service"
	domainConfig "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/config/model"
	configService "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/config/service"
	gitService "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/git/service"
	gpgkeyRepo "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/gpgkey/repository"
	gpgkeyService "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/gpgkey/service"
	graphRepo "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/graph/repository"
	graphService "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/graph/service"
	moduleModel "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/model"
	moduleRepo "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/repository"
	moduleService "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/service" // Alias for the new module service
	providerRepo "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider/repository"
	sharedService "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared/service"
	storageModel "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/storage/model"
	storageService "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/storage/service"
	urlservice "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/url/service"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/config"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/git"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb/transaction"
	storageInfrastructure "github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/storage"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/version"

	providerLogoRepository "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider_logo/repository"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/parser"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
	analyticsPersistence "github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb/analytics"
	auditPersistence "github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb/audit"
	authPersistence "github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb/auth"
	gpgkeyPersistence "github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb/gpgkey"
	graphPersistence "github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb/graph"
	modulePersistence "github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb/module"
	sqldbprovider "github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb/provider"
	providerLogoRepo "github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/provider_logo"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http"
	terraformHandler "github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/handler/terraform"
	v1 "github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/handler/terraform/v1"
	v2 "github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/handler/terraform/v2"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/handler/terrareg"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/handler/webhook"
	terrareg_middleware "github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/middleware"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/template"
)

// Container holds all application dependencies
type Container struct {
	// Configuration architecture
	DomainConfig  *domainConfig.DomainConfig
	InfraConfig   *config.InfrastructureConfig
	ConfigService *configService.ConfigurationService

	Logger zerolog.Logger
	DB     *sqldb.Database

	// Repositories
	NamespaceRepo                     moduleRepo.NamespaceRepository
	ModuleProviderRepo                moduleRepo.ModuleProviderRepository
	ModuleVersionRepo                 moduleRepo.ModuleVersionRepository
	ModuleVersionFileRepo             moduleModel.ModuleVersionFileRepository
	ModuleProviderRedirectRepo        *modulePersistence.ModuleProviderRedirectRepositoryImpl
	AnalyticsRepo                     analyticsCmd.AnalyticsRepository
	ProviderRepo                      providerRepo.ProviderRepository
	ProviderLogoRepo                  providerLogoRepository.ProviderLogoRepository
	SessionRepo                       authRepo.SessionRepository
	UserGroupRepo                     authRepo.UserGroupRepository
	TerraformIdpAuthorizationCodeRepo authRepo.TerraformIdpAuthorizationCodeRepository
	TerraformIdpAccessTokenRepo       authRepo.TerraformIdpAccessTokenRepository
	TerraformIdpSubjectIdentifierRepo authRepo.TerraformIdpSubjectIdentifierRepository

	// Audit
	AuditHistoryRepo auditRepo.AuditHistoryRepository

	// Graph
	GraphRepo    graphRepo.DependencyGraphRepository
	GraphService *graphService.GraphService

	// GPG Keys
	GPGKeyRepo    gpgkeyRepo.GPGKeyRepository
	GPGKeyService *gpgkeyService.GPGKeyService

	// Infrastructure Services
	GitClient              gitService.GitClient
	DomainStorageService   storageService.StorageService // Core domain storage (8 methods)
	ModuleStorageService   moduleService.StorageService  // Module-specific storage with filesystem ops
	PathBuilder            storageService.PathBuilder
	TempDirManager         storageService.TemporaryDirectoryManager
	StorageWorkflowService storageService.StorageWorkflowService
	GitService             gitService.GitService
	ModuleIndexingService  moduleService.ModuleIndexingService
	ModuleParser           moduleService.ModuleParser

	// Domain Services
	ModuleImporterService *moduleService.ModuleImporterService
	NamespaceService      *moduleService.NamespaceService
	SecurityService       *moduleService.SecurityService
	ModuleFileService     *moduleService.ModuleFileService
	WebhookService        *moduleService.WebhookService
	MarkdownService       *sharedService.MarkdownService

	// Transaction Services
	SavepointHelper                     *transaction.SavepointHelper
	SecurityScanningService             *moduleService.SecurityScanningService
	ModuleCreationWrapper               *moduleService.ModuleCreationWrapperService
	ArchiveGenerationTransactionService *moduleService.ArchiveGenerationTransactionService
	MetadataProcessingService           *moduleService.MetadataProcessingService
	TerraformExecutorService            *moduleService.TerraformExecutorService
	TransactionProcessingOrchestrator   *moduleService.TransactionProcessingOrchestrator
	AuthFactory                         *authservice.AuthFactory
	SessionService                      *authservice.SessionService
	CookieService                       *authservice.CookieService
	AuthenticationService               *authservice.AuthenticationService
	SessionCleanupService               *authservice.SessionCleanupService
	TerraformIdpService                 *authservice.TerraformIdpService
	OIDCService                         *authservice.OIDCService
	SAMLService                         *authservice.SAMLService
	StateStorageService                 *authservice.StateStorageService
	URLService                          *urlservice.URLService

	// Commands
	CreateNamespaceCmd              *namespace.CreateNamespaceCommand
	UpdateNamespaceCmd              *namespace.UpdateNamespaceCommand
	DeleteNamespaceCmd              *namespace.DeleteNamespaceCommand
	CreateModuleProviderCmd         *moduleCmd.CreateModuleProviderCommand
	PublishModuleVersionCmd         *moduleCmd.PublishModuleVersionCommand
	UpdateModuleProviderSettingsCmd *moduleCmd.UpdateModuleProviderSettingsCommand
	DeleteModuleProviderCmd         *moduleCmd.DeleteModuleProviderCommand
	UploadModuleVersionCmd          *moduleCmd.UploadModuleVersionCommand
	ImportModuleVersionCmd          *moduleCmd.ImportModuleVersionCommand
	GetModuleVersionFileCmd         *moduleCmd.GetModuleVersionFileQuery
	DeleteModuleVersionCmd          *moduleCmd.DeleteModuleVersionCommand
	GenerateModuleSourceCmd         *moduleCmd.GenerateModuleSourceCommand
	GetVariableTemplateQuery        *moduleCmd.GetVariableTemplateQuery
	RecordModuleDownloadCmd         *analyticsCmd.RecordModuleDownloadCommand
	AdminLoginCmd                   *authCmd.AdminLoginCommand
	CreateModuleProviderRedirectCmd *moduleCmd.CreateModuleProviderRedirectCommand
	DeleteModuleProviderRedirectCmd *moduleCmd.DeleteModuleProviderRedirectCommand

	// Authentication Commands
	OidcLoginCmd    *authCmd.OidcLoginCommand
	OidcCallbackCmd *authCmd.OidcCallbackCommand
	SamlLoginCmd    *authCmd.SamlLoginCommand
	SamlMetadataCmd *authCmd.SamlMetadataCommand
	GithubOAuthCmd  *authCmd.GithubOAuthCommand

	// Terraform Authentication Commands
	AuthenticateOIDCTokenCmd *terraformCmd.AuthenticateOIDCTokenCommand
	ValidateTokenCmd         *terraformCmd.ValidateTokenCommand
	GetUserCmd               *terraformCmd.GetUserCommand

	// Provider Commands
	CreateOrUpdateProviderCmd *providerCmd.CreateOrUpdateProviderCommand
	PublishProviderVersionCmd *providerCmd.PublishProviderVersionCommand
	ManageGPGKeyCmd           *providerCmd.ManageGPGKeyCommand
	GetProviderDownloadQuery  *providerCmd.GetProviderDownloadQuery

	// GPG Key Commands
	ManageGPGKeyCmd2 *gpgkeyCmd.ManageGPGKeyCommand

	// Audit Queries
	GetAuditHistoryQuery *auditQuery.GetAuditHistoryQuery

	// Graph Queries
	GetModuleDependencyGraphQuery *graphQuery.GetModuleDependencyGraphQuery

	// Provider Queries
	GetProviderVersionQuery  *providerQuery.GetProviderVersionQuery
	GetNamespaceGPGKeysQuery *providerQuery.GetNamespaceGPGKeysQuery

	// GPG Key Queries
	GetNamespaceGPGKeysQuery2        *gpgkeyQuery.GetNamespaceGPGKeysQuery
	GetMultipleNamespaceGPGKeysQuery *gpgkeyQuery.GetMultipleNamespaceGPGKeysQuery
	GetGPGKeyQuery                   *gpgkeyQuery.GetGPGKeyQuery

	// Provider Logo Queries
	GetProviderLogosQuery *providerLogoQuery.GetAllProviderLogosQuery

	// Queries
	ListNamespacesQuery             *module.ListNamespacesQuery
	NamespaceDetailsQuery           *namespaceQuery.NamespaceDetailsQuery
	ListModulesQuery                *module.ListModulesQuery
	SearchModulesQuery              *module.SearchModulesQuery
	GetModuleProviderQuery          *module.GetModuleProviderQuery
	ListModuleProvidersQuery        *module.ListModuleProvidersQuery
	GetModuleVersionQuery           *module.GetModuleVersionQuery
	GetModuleProviderRedirectsQuery *moduleQuery.GetModuleProviderRedirectsQuery
	GetModuleDownloadQuery          *module.GetModuleDownloadQuery
	GetModuleProviderSettingsQuery  *module.GetModuleProviderSettingsQuery
	ListModuleVersionsQuery         *module.ListModuleVersionsQuery
	GetSubmodulesQuery              *module.GetSubmodulesQuery
	GetExamplesQuery                *module.GetExamplesQuery
	GetIntegrationsQuery            *module.GetIntegrationsQuery
	GetReadmeHTMLQuery              *module.GetReadmeHTMLQuery
	GetSubmoduleDetailsQuery        *module.GetSubmoduleDetailsQuery
	GetSubmoduleReadmeHTMLQuery     *module.GetSubmoduleReadmeHTMLQuery
	GetExampleDetailsQuery          *module.GetExampleDetailsQuery
	GetExampleReadmeHTMLQuery       *module.GetExampleReadmeHTMLQuery
	GetExampleFileListQuery         *module.GetExampleFileListQuery
	GetExampleFileQuery             *module.GetExampleFileQuery
	GlobalStatsQuery                *analyticsQuery.GlobalStatsQuery
	GlobalUsageStatsQuery           *analyticsQuery.GlobalUsageStatsQuery
	GetDownloadSummaryQuery         *analyticsQuery.GetDownloadSummaryQuery
	GetMostRecentlyPublishedQuery   *analyticsQuery.GetMostRecentlyPublishedQuery
	GetMostDownloadedThisWeekQuery  *analyticsQuery.GetMostDownloadedThisWeekQuery
	GetTokenVersionsQuery           *analyticsQuery.GetTokenVersionsQuery
	ListProvidersQuery              *providerQuery.ListProvidersQuery
	SearchProvidersQuery            *providerQuery.SearchProvidersQuery
	GetProviderQuery                *providerQuery.GetProviderQuery
	GetProviderVersionsQuery        *providerQuery.GetProviderVersionsQuery
	CheckSessionQuery               *authQuery.CheckSessionQuery
	IsAuthenticatedQuery            *authQuery.IsAuthenticatedQuery

	// Config queries
	GetConfigQuery  *configQuery.GetConfigQuery
	GetVersionQuery *configQuery.GetVersionQuery

	// Handlers
	NamespaceHandler           *terrareg.NamespaceHandler
	ModuleHandler              *terrareg.ModuleHandler
	SubmoduleHandler           *terrareg.SubmoduleHandler
	ExampleHandler             *terrareg.ExampleHandler
	AnalyticsHandler           *terrareg.AnalyticsHandler
	ProviderHandler            *terrareg.ProviderHandler
	ProviderLogosHandler       *terrareg.ProviderLogosHandler
	AuthHandler                *terrareg.AuthHandler
	AuditHandler               *terrareg.AuditHandler
	GraphHandler               *terrareg.GraphHandler
	ModuleWebhookHandler       *webhook.ModuleWebhookHandler
	TerraformV1ModuleHandler   *v1.TerraformV1ModuleHandler // New V1 Terraform Module Handler
	TerraformV2ProviderHandler *v2.TerraformV2ProviderHandler
	TerraformV2CategoryHandler *v2.TerraformV2CategoryHandler
	TerraformV2GPGHandler      *v2.TerraformV2GPGHandler

	// Terraform Authentication Handlers
	TerraformAuthHandler        *terraformHandler.TerraformAuthHandler
	TerraformIDPHandler         *terraformHandler.TerraformIDPHandler
	TerraformStaticTokenHandler *terraformHandler.TerraformStaticTokenHandler

	// Initial Setup
	GetInitialSetupQuery *setupQuery.GetInitialSetupQuery
	InitialSetupHandler  *terrareg.InitialSetupHandler

	// Middleware
	AuthMiddleware    *terrareg_middleware.AuthMiddleware
	SessionMiddleware *terrareg_middleware.SessionMiddleware

	// Template renderer
	TemplateRenderer *template.Renderer

	// Config/Version Handlers
	ConfigHandler  *terrareg.ConfigHandler
	VersionHandler *terrareg.VersionHandler

	// Search Filters
	SearchFiltersQuery   *moduleQuery.SearchFiltersQuery
	SearchFiltersHandler *terrareg.SearchFiltersHandler

	// HTTP Server
	Server *http.Server
}

// NewContainer creates and initializes the dependency injection container
func NewContainer(
	domainConfig *domainConfig.DomainConfig,
	infraConfig *config.InfrastructureConfig,
	configService *configService.ConfigurationService,
	logger zerolog.Logger,
	db *sqldb.Database,
) (*Container, error) {
	c := &Container{
		DomainConfig:  domainConfig,
		InfraConfig:   infraConfig,
		ConfigService: configService,
		Logger:        logger,
		DB:            db,
	}

	// Initialize repositories
	c.NamespaceRepo = modulePersistence.NewNamespaceRepository(db.DB)
	c.ModuleProviderRepo = modulePersistence.NewModuleProviderRepository(db.DB, c.NamespaceRepo, domainConfig) // Uses DomainConfig for TrustedNamespaces
	c.ModuleVersionRepo = modulePersistence.NewModuleVersionRepository(db.DB)
	c.ModuleVersionFileRepo = modulePersistence.NewModuleVersionFileRepository(db.DB)
	c.ModuleProviderRedirectRepo = modulePersistence.NewModuleProviderRedirectRepository(db.DB)
	c.ProviderRepo = sqldbprovider.NewProviderRepository(db.DB)
	c.ProviderLogoRepo = providerLogoRepo.NewProviderLogoRepository()
	c.SessionRepo = authPersistence.NewSessionRepository(db.DB)
	c.UserGroupRepo = authPersistence.NewUserGroupRepository(db.DB)
	c.TerraformIdpAuthorizationCodeRepo = authPersistence.NewTerraformIdpAuthorizationCodeRepository(db.DB)
	c.TerraformIdpAccessTokenRepo = authPersistence.NewTerraformIdpAccessTokenRepository(db.DB)
	c.TerraformIdpSubjectIdentifierRepo = authPersistence.NewTerraformIdpSubjectIdentifierRepository(db.DB)
	c.AuditHistoryRepo = auditPersistence.NewAuditHistoryRepository(db.DB)
	c.GraphRepo = graphPersistence.NewDependencyGraphRepository(db.DB)
	c.GPGKeyRepo = gpgkeyPersistence.NewGPGKeyRepository(db.DB)

	c.URLService = urlservice.NewURLService(infraConfig)

	// Initialize infrastructure services
	c.GitClient = git.NewGitClientImpl()

	// Initialize consolidated storage services
	pathConfig := storageService.GetDefaultPathConfig(infraConfig.DataDirectory)
	c.PathBuilder = storageService.NewPathBuilderService(pathConfig)

	// Initialize domain storage service (8 core methods)
	domainStorageService, err := storageInfrastructure.NewLocalStorageService(infraConfig.DataDirectory, c.PathBuilder)
	if err != nil {
		return nil, fmt.Errorf("failed to create domain storage service: %w", err)
	}
	c.DomainStorageService = domainStorageService

	// Initialize temporary directory manager
	tempDirManager, err := storageInfrastructure.NewTemporaryDirectoryManager()
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary directory manager: %w", err)
	}
	c.TempDirManager = tempDirManager

	// Initialize storage workflow service - convert pathConfig type
	pathConfigModel := &storageModel.StoragePathConfig{
		BasePath:      pathConfig.BasePath,
		ModulesPath:   pathConfig.ModulesPath,
		ProvidersPath: pathConfig.ProvidersPath,
		UploadPath:    pathConfig.UploadPath,
		TempPath:      pathConfig.TempPath,
	}
	c.StorageWorkflowService = storageService.NewStorageWorkflowServiceImpl(
		c.DomainStorageService,
		c.PathBuilder,
		tempDirManager,
		pathConfigModel,
	)

	// Initialize git service for workflow
	c.GitService = gitService.NewGitService()

	// Create adapter for module-specific storage operations (CopyDir, ExtractArchive, etc.)
	moduleStorageAdapter := storageInfrastructure.NewModuleStorageAdapter(c.DomainStorageService, c.PathBuilder)
	c.ModuleStorageService = moduleStorageAdapter

	// Initialize module indexing service (pending ModuleProcessorService and ArchiveGenerationService)
	// TODO: Implement ModuleProcessorService and ArchiveGenerationService
	// c.ModuleIndexingService = moduleService.NewModuleIndexingServiceImpl(
	//     c.GitService,
	//     c.StorageWorkflowService,
	//     c.ModuleProcessorService, // TODO: Create this service
	//     c.ArchiveGenerationService, // TODO: Create this service
	//     c.Logger,
	// )

	c.ModuleParser = parser.NewModuleParserImpl(c.ModuleStorageService)

	// Initialize foundation transaction services
	savepointHelper := transaction.NewSavepointHelper(db.DB)
	c.SavepointHelper = savepointHelper

	// Initialize combined security scanning service with transaction support
	securityScanningService := moduleService.NewSecurityScanningService(
		c.ModuleFileService,
		c.ModuleVersionRepo,
		savepointHelper,
	)
	c.SecurityScanningService = securityScanningService

	// Initialize module creation wrapper for atomic module creation
	moduleCreationWrapper := moduleService.NewModuleCreationWrapperService(
		c.ModuleVersionRepo,
		savepointHelper,
	)
	c.ModuleCreationWrapper = moduleCreationWrapper

	// Initialize archive generation transaction service
	archiveGenService := moduleService.NewArchiveGenerationTransactionService(savepointHelper)
	c.ArchiveGenerationTransactionService = archiveGenService

	// Initialize metadata processing service
	metadataService := moduleService.NewMetadataProcessingService(savepointHelper)
	c.MetadataProcessingService = metadataService

	// Initialize terraform executor service with tfswitch config
	tfswitchConfig := &moduleService.TfswitchConfig{
		DefaultTerraformVersion: "1.5.7",              // TODO: configure from domain config
		TerraformProduct:        "terraform",          // TODO: configure from domain config
		ArchiveMirror:           "",                   // TODO: configure from infra config
		BinaryPath:              "/app/bin/terraform", // TODO: configure from infra config
	}
	terraformExecutorService := moduleService.NewTerraformExecutorService(
		savepointHelper,
		"bin/terraform", // TODO: configure terraform binary path
		30*time.Second,  // TODO: configure lock timeout
		tfswitchConfig,
	)
	c.TerraformExecutorService = terraformExecutorService

	// Initialize transaction processing orchestrator (simplified for now)
	processingOrchestrator := moduleService.NewTransactionProcessingOrchestrator(
		nil, // archiveService - TODO: implement ArchiveProcessor
		terraformExecutorService,
		metadataService,
		securityScanningService,
		nil, // fileContentService - TODO: implement FileStorage/FileProcessing services
		archiveGenService,
		moduleCreationWrapper,
		savepointHelper,
		c.DomainConfig,
		c.Logger,
		c.ModuleVersionRepo,
		c.ModuleProviderRepo,
	)
	c.TransactionProcessingOrchestrator = processingOrchestrator

	// Initialize updated module importer service with transaction support
	c.ModuleImporterService = moduleService.NewModuleImporterService(
		processingOrchestrator,
		moduleCreationWrapper,
		savepointHelper,
		c.ModuleProviderRepo,
		c.GitClient,
		c.ModuleStorageService,
		c.ModuleParser,
		domainConfig,
		infraConfig,
		logger,
	)

	// Initialize existing domain services
	c.NamespaceService = moduleService.NewNamespaceService(domainConfig) // Uses DomainConfig for business logic
	c.AnalyticsRepo = analyticsPersistence.NewAnalyticsRepository(db.DB, c.NamespaceRepo, c.NamespaceService)

	// Initialize security and module file services (legacy - can be gradually phased out)
	c.SecurityService = moduleService.NewSecurityService()
	c.MarkdownService = sharedService.NewMarkdownService()
	c.ModuleFileService = moduleService.NewModuleFileService(
		c.ModuleProviderRepo,
		c.ModuleVersionFileRepo,
		c.NamespaceService,
		c.SecurityService,
	)

	// Initialize webhook service (updated to use transaction-aware services)
	c.WebhookService = moduleService.NewWebhookService(
		c.ModuleImporterService,
		c.ModuleProviderRepo,
		infraConfig,
		savepointHelper,
		moduleCreationWrapper,
	)

	// Initialize auth services
	// Use the refactored SessionService (pure database operations)
	sessionService := authservice.NewSessionService(c.SessionRepo, authservice.DefaultSessionDatabaseConfig())
	c.SessionService = sessionService

	// Initialize cookie service (pure cookie operations)
	cookieService := authservice.NewCookieService(infraConfig) // Uses InfrastructureConfig for auth settings
	c.CookieService = cookieService

	// Initialize authentication service (orchestrates session and cookie operations)
	c.AuthenticationService = authservice.NewAuthenticationService(sessionService, cookieService)

	// Initialize session cleanup service
	c.SessionCleanupService = authservice.NewSessionCleanupService(
		sessionService,
		logger,
		authservice.DefaultSessionDatabaseConfig().CleanupInterval,
	)

	// Initialize audit service
	auditService := auditservice.NewAuditService(c.AuditHistoryRepo)

	// Initialize graph service
	c.GraphService = graphService.NewGraphService(c.GraphRepo)

	// Initialize GPG key service
	c.GPGKeyService = gpgkeyService.NewGPGKeyService(c.GPGKeyRepo, c.NamespaceRepo)

	c.TerraformIdpService = authservice.NewTerraformIdpService(
		c.TerraformIdpAuthorizationCodeRepo,
		c.TerraformIdpAccessTokenRepo,
		c.TerraformIdpSubjectIdentifierRepo,
		infraConfig.TerraformOidcIdpSigningKeyPath,
		infraConfig.PublicURL,
	)

	// Initialize AuthFactory after TerraformIdpService is created
	c.AuthFactory = authservice.NewAuthFactory(c.SessionRepo, c.UserGroupRepo, infraConfig, c.TerraformIdpService, &c.Logger)

	// Initialize OIDC and SAML services (these may return nil if not configured)
	ctx := context.Background()
	c.OIDCService, _ = authservice.NewOIDCService(ctx, infraConfig)
	c.SAMLService, _ = authservice.NewSAMLService(infraConfig)

	// Initialize StateStorageService
	c.StateStorageService = authservice.NewStateStorageService(c.SessionService)

	// Initialize commands
	c.CreateNamespaceCmd = namespace.NewCreateNamespaceCommand(c.NamespaceRepo)
	c.UpdateNamespaceCmd = namespace.NewUpdateNamespaceCommand(c.NamespaceRepo)
	c.DeleteNamespaceCmd = namespace.NewDeleteNamespaceCommand(c.NamespaceRepo)
	c.CreateModuleProviderCmd = moduleCmd.NewCreateModuleProviderCommand(c.NamespaceRepo, c.ModuleProviderRepo)
	c.PublishModuleVersionCmd = moduleCmd.NewPublishModuleVersionCommand(c.ModuleProviderRepo)
	c.UpdateModuleProviderSettingsCmd = moduleCmd.NewUpdateModuleProviderSettingsCommand(c.ModuleProviderRepo)
	c.DeleteModuleProviderCmd = moduleCmd.NewDeleteModuleProviderCommand(c.ModuleProviderRepo)
	c.UploadModuleVersionCmd = moduleCmd.NewUploadModuleVersionCommand(c.ModuleProviderRepo, c.ModuleParser, c.ModuleStorageService, infraConfig) // Uses InfrastructureConfig for file operations
	c.ImportModuleVersionCmd = moduleCmd.NewImportModuleVersionCommand(c.ModuleImporterService)
	c.RecordModuleDownloadCmd = analyticsCmd.NewRecordModuleDownloadCommand(c.ModuleProviderRepo, c.AnalyticsRepo)
	c.GetModuleVersionFileCmd = moduleCmd.NewGetModuleVersionFileQuery(c.ModuleFileService)
	c.DeleteModuleVersionCmd = moduleCmd.NewDeleteModuleVersionCommand(c.ModuleProviderRepo, c.ModuleVersionRepo)
	c.GenerateModuleSourceCmd = moduleCmd.NewGenerateModuleSourceCommand(c.ModuleProviderRepo, c.ModuleFileService)
	c.GetVariableTemplateQuery = moduleCmd.NewGetVariableTemplateQuery(c.ModuleProviderRepo, c.ModuleFileService)

	// Initialize admin login command
	c.AdminLoginCmd = authCmd.NewAdminLoginCommand(c.AuthFactory, c.SessionService, infraConfig) // Uses InfrastructureConfig for auth settings

	// Initialize redirect commands and queries
	c.CreateModuleProviderRedirectCmd = moduleCmd.NewCreateModuleProviderRedirectCommand(c.ModuleProviderRedirectRepo)
	c.DeleteModuleProviderRedirectCmd = moduleCmd.NewDeleteModuleProviderRedirectCommand(c.ModuleProviderRedirectRepo)
	c.GetModuleProviderRedirectsQuery = moduleQuery.NewGetModuleProviderRedirectsQuery(c.ModuleProviderRedirectRepo)

	// Initialize authentication commands
	c.OidcLoginCmd = authCmd.NewOidcLoginCommand(c.AuthFactory, c.SessionService, infraConfig, c.OIDCService)
	c.OidcCallbackCmd = authCmd.NewOidcCallbackCommand(c.AuthFactory, c.SessionService, infraConfig, c.OIDCService, c.OidcLoginCmd)
	c.SamlLoginCmd = authCmd.NewSamlLoginCommand(c.AuthFactory, c.SessionService, infraConfig, c.SAMLService)
	c.SamlMetadataCmd = authCmd.NewSamlMetadataCommand(c.AuthFactory, c.SessionService, infraConfig)
	c.GithubOAuthCmd = authCmd.NewGithubOAuthCommand(c.AuthFactory, c.SessionService, infraConfig)

	// Initialize Terraform authentication commands
	c.AuthenticateOIDCTokenCmd = terraformCmd.NewAuthenticateOIDCTokenCommand(c.AuthFactory)
	c.ValidateTokenCmd = terraformCmd.NewValidateTokenCommand(c.AuthFactory)
	c.GetUserCmd = terraformCmd.NewGetUserCommand(c.AuthFactory)

	// Initialize queries
	c.ListNamespacesQuery = module.NewListNamespacesQuery(c.NamespaceRepo)
	c.NamespaceDetailsQuery = namespaceQuery.NewNamespaceDetailsQuery(c.NamespaceRepo, c.NamespaceService)
	c.ListModulesQuery = module.NewListModulesQuery(c.ModuleProviderRepo)
	c.SearchModulesQuery = module.NewSearchModulesQuery(c.ModuleProviderRepo)
	c.GetModuleProviderQuery = module.NewGetModuleProviderQuery(c.ModuleProviderRepo)
	c.ListModuleProvidersQuery = moduleQuery.NewListModuleProvidersQuery(c.ModuleProviderRepo)
	c.GetModuleVersionQuery = moduleQuery.NewGetModuleVersionQuery(c.ModuleProviderRepo)
	c.ListModuleVersionsQuery = moduleQuery.NewListModuleVersionsQuery(c.ModuleProviderRepo) // New query
	c.GetModuleDownloadQuery = moduleQuery.NewGetModuleDownloadQuery(c.ModuleProviderRepo)
	c.GetModuleProviderSettingsQuery = moduleQuery.NewGetModuleProviderSettingsQuery(c.ModuleProviderRepo)
	c.GetSubmodulesQuery = moduleQuery.NewGetSubmodulesQuery(c.ModuleProviderRepo) // Uses database instead of filesystem
	c.GetExamplesQuery = moduleQuery.NewGetExamplesQuery(c.ModuleProviderRepo)     // Uses database instead of filesystem
	c.GetIntegrationsQuery = moduleQuery.NewGetIntegrationsQuery(c.ModuleProviderRepo)
	c.GetReadmeHTMLQuery = moduleQuery.NewGetReadmeHTMLQuery(c.ModuleProviderRepo, c.MarkdownService)
	c.GetSubmoduleDetailsQuery = moduleQuery.NewGetSubmoduleDetailsQuery(c.ModuleProviderRepo, c.ModuleVersionRepo)
	c.GetSubmoduleReadmeHTMLQuery = moduleQuery.NewGetSubmoduleReadmeHTMLQuery(c.ModuleProviderRepo, c.ModuleVersionRepo)
	c.GetExampleDetailsQuery = moduleQuery.NewGetExampleDetailsQuery(c.ModuleProviderRepo, c.ModuleVersionRepo)
	c.GetExampleReadmeHTMLQuery = moduleQuery.NewGetExampleReadmeHTMLQuery(c.ModuleProviderRepo, c.ModuleVersionRepo)
	c.GetExampleFileListQuery = moduleQuery.NewGetExampleFileListQuery(c.ModuleProviderRepo, c.ModuleVersionRepo)
	c.GetExampleFileQuery = moduleQuery.NewGetExampleFileQuery(c.ModuleProviderRepo, c.ModuleVersionRepo)
	c.GlobalStatsQuery = analyticsQuery.NewGlobalStatsQuery(c.NamespaceRepo, c.ModuleProviderRepo, c.AnalyticsRepo)
	c.GlobalUsageStatsQuery = analyticsQuery.NewGlobalUsageStatsQuery(c.ModuleProviderRepo, c.AnalyticsRepo)
	c.GetDownloadSummaryQuery = analyticsQuery.NewGetDownloadSummaryQuery(c.AnalyticsRepo)
	c.GetMostRecentlyPublishedQuery = analyticsQuery.NewGetMostRecentlyPublishedQuery(c.AnalyticsRepo)
	c.GetMostDownloadedThisWeekQuery = analyticsQuery.NewGetMostDownloadedThisWeekQuery(c.AnalyticsRepo)
	c.GetTokenVersionsQuery = analyticsQuery.NewGetTokenVersionsQuery(c.AnalyticsRepo)
	c.ListProvidersQuery = providerQuery.NewListProvidersQuery(c.ProviderRepo)
	c.SearchProvidersQuery = providerQuery.NewSearchProvidersQuery(c.ProviderRepo)
	c.GetProviderQuery = providerQuery.NewGetProviderQuery(c.ProviderRepo)
	c.GetProviderVersionsQuery = providerQuery.NewGetProviderVersionsQuery(c.ProviderRepo)
	c.GetProviderVersionQuery = providerQuery.NewGetProviderVersionQuery(c.ProviderRepo)
	c.GetNamespaceGPGKeysQuery = providerQuery.NewGetNamespaceGPGKeysQuery(c.NamespaceRepo)
	c.GetProviderLogosQuery = providerLogoQuery.NewGetAllProviderLogosQuery(c.ProviderLogoRepo)
	c.CreateOrUpdateProviderCmd = providerCmd.NewCreateOrUpdateProviderCommand(c.ProviderRepo, c.NamespaceRepo)
	c.PublishProviderVersionCmd = providerCmd.NewPublishProviderVersionCommand(c.ProviderRepo, c.NamespaceRepo)
	c.ManageGPGKeyCmd = providerCmd.NewManageGPGKeyCommand(c.ProviderRepo, c.NamespaceRepo)
	c.GetProviderDownloadQuery = providerCmd.NewGetProviderDownloadQuery(c.ProviderRepo, c.NamespaceRepo, c.AnalyticsRepo, c.GPGKeyRepo)
	c.CheckSessionQuery = authQuery.NewCheckSessionQuery(c.SessionRepo)
	c.IsAuthenticatedQuery = authQuery.NewIsAuthenticatedQuery()
	c.GetAuditHistoryQuery = auditQuery.NewGetAuditHistoryQuery(auditService)

	// Graph queries
	c.GetModuleDependencyGraphQuery = graphQuery.NewGetModuleDependencyGraphQuery(c.GraphService)

	// GPG key commands and queries
	c.ManageGPGKeyCmd2 = gpgkeyCmd.NewManageGPGKeyCommand(c.GPGKeyService)
	c.GetNamespaceGPGKeysQuery2 = gpgkeyQuery.NewGetNamespaceGPGKeysQuery(c.GPGKeyService)
	c.GetMultipleNamespaceGPGKeysQuery = gpgkeyQuery.NewGetMultipleNamespaceGPGKeysQuery(c.GPGKeyService)
	c.GetGPGKeyQuery = gpgkeyQuery.NewGetGPGKeyQuery(c.GPGKeyService)

	// Initialize config repository and queries
	versionReader := version.NewVersionReader()
	configRepository := config.NewConfigRepositoryImpl(versionReader, domainConfig)
	c.GetConfigQuery = configQuery.NewGetConfigQuery(configRepository)
	c.GetVersionQuery = configQuery.NewGetVersionQuery(configRepository)

	// Initialize initial setup query
	c.GetInitialSetupQuery = setupQuery.NewGetInitialSetupQuery(
		c.NamespaceRepo,
		c.ModuleProviderRepo,
		c.ModuleVersionRepo,
		c.URLService,
		domainConfig, // Uses DomainConfig for business logic
	)

	// Initialize config/version handlers
	c.ConfigHandler = terrareg.NewConfigHandler(c.GetConfigQuery)
	c.VersionHandler = terrareg.NewVersionHandler(c.GetVersionQuery)

	// Initialize initial setup handler
	c.InitialSetupHandler = terrareg.NewInitialSetupHandler(c.GetInitialSetupQuery)

	// Initialize handlers
	c.NamespaceHandler = terrareg.NewNamespaceHandler(c.ListNamespacesQuery, c.CreateNamespaceCmd, c.UpdateNamespaceCmd, c.DeleteNamespaceCmd, c.NamespaceDetailsQuery)
	c.ModuleHandler = terrareg.NewModuleHandler(
		c.ListModulesQuery,
		c.SearchModulesQuery,
		c.GetModuleProviderQuery,
		c.ListModuleProvidersQuery,
		c.GetModuleVersionQuery,
		c.GetModuleDownloadQuery,
		c.GetModuleProviderSettingsQuery,
		c.GetReadmeHTMLQuery,
		c.GetSubmodulesQuery,
		c.GetExamplesQuery,
		c.GetIntegrationsQuery,
		c.CreateModuleProviderCmd,
		c.PublishModuleVersionCmd,
		c.UpdateModuleProviderSettingsCmd,
		c.DeleteModuleProviderCmd,
		c.UploadModuleVersionCmd,
		c.ImportModuleVersionCmd,
		c.GetModuleVersionFileCmd,
		c.DeleteModuleVersionCmd,
		c.GenerateModuleSourceCmd,
		c.GetVariableTemplateQuery,
		c.CreateModuleProviderRedirectCmd,
		c.DeleteModuleProviderRedirectCmd,
		c.GetModuleProviderRedirectsQuery,
		domainConfig,
		c.NamespaceService,
		c.AnalyticsRepo,
	)
	c.AnalyticsHandler = terrareg.NewAnalyticsHandler(
		c.GlobalStatsQuery,
		c.GlobalUsageStatsQuery,
		c.GetDownloadSummaryQuery,
		c.RecordModuleDownloadCmd,
		c.GetMostRecentlyPublishedQuery,
		c.GetMostDownloadedThisWeekQuery,
		c.GetTokenVersionsQuery,
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
		c.GetProviderDownloadQuery,
	)
	c.ProviderLogosHandler = terrareg.NewProviderLogosHandler(c.GetProviderLogosQuery)
	c.AuthHandler = terrareg.NewAuthHandler(
		c.AdminLoginCmd,
		c.CheckSessionQuery,
		c.IsAuthenticatedQuery,
		c.OidcLoginCmd,
		c.OidcCallbackCmd,
		c.SamlLoginCmd,
		c.SamlMetadataCmd,
		c.GithubOAuthCmd,
		c.AuthenticationService,
		c.StateStorageService,
		infraConfig,
	)
	c.AuditHandler = terrareg.NewAuditHandler(c.GetAuditHistoryQuery)
	c.GraphHandler = terrareg.NewGraphHandler(c.GetModuleDependencyGraphQuery)

	// Initialize webhook handler with upload API keys for signature validation
	var uploadAPIKeys []string
	if infraConfig.UploadApiKeys != nil {
		uploadAPIKeys = infraConfig.UploadApiKeys
	}
	c.ModuleWebhookHandler = webhook.NewModuleWebhookHandler(c.WebhookService, uploadAPIKeys)
	c.TerraformV1ModuleHandler = v1.NewTerraformV1ModuleHandler(c.ListModulesQuery, c.SearchModulesQuery, c.GetModuleProviderQuery, c.ListModuleVersionsQuery, c.GetModuleDownloadQuery, c.GetModuleVersionQuery) // Instantiate the new handler

	// Initialize submodule and example handlers
	c.SubmoduleHandler = terrareg.NewSubmoduleHandler(c.GetSubmoduleDetailsQuery, c.GetSubmoduleReadmeHTMLQuery)
	c.ExampleHandler = terrareg.NewExampleHandler(c.GetExampleDetailsQuery, c.GetExampleReadmeHTMLQuery, c.GetExampleFileListQuery, c.GetExampleFileQuery)

	// Initialize v2 handlers
	c.TerraformV2ProviderHandler = v2.NewTerraformV2ProviderHandler(c.GetProviderQuery, c.GetProviderVersionsQuery, c.GetProviderVersionQuery, c.ListProvidersQuery, c.GetProviderDownloadQuery)
	c.TerraformV2CategoryHandler = v2.NewTerraformV2CategoryHandler(providerQuery.NewListUserSelectableProviderCategoriesQuery(nil)) // TODO: Add proper category repo
	c.TerraformV2GPGHandler = v2.NewTerraformV2GPGHandler(
		c.ManageGPGKeyCmd2,
		c.GetNamespaceGPGKeysQuery2,
		c.GetMultipleNamespaceGPGKeysQuery,
		c.GetGPGKeyQuery,
	)

	// Initialize Terraform authentication handlers
	c.TerraformAuthHandler = terraformHandler.NewTerraformAuthHandler(
		c.AuthenticateOIDCTokenCmd,
		c.ValidateTokenCmd,
		c.GetUserCmd,
	)
	c.TerraformIDPHandler = terraformHandler.NewTerraformIDPHandler(c.TerraformIdpService)
	c.TerraformStaticTokenHandler = terraformHandler.NewTerraformStaticTokenHandler()

	// Initialize middleware
	c.AuthMiddleware = terrareg_middleware.NewAuthMiddleware(domainConfig, c.AuthFactory) // Uses DomainConfig for auth settings
	c.SessionMiddleware = terrareg_middleware.NewSessionMiddleware(c.AuthenticationService, c.Logger)

	// Initialize template renderer
	templateRenderer, err := template.NewRenderer(domainConfig, infraConfig)
	if err != nil {
		return nil, err
	}
	c.TemplateRenderer = templateRenderer

	// Initialize search filters
	c.SearchFiltersQuery = moduleQuery.NewSearchFiltersQuery(c.ModuleProviderRepo, domainConfig) // Uses DomainConfig for filtering logic
	c.SearchFiltersHandler = terrareg.NewSearchFiltersHandler(c.SearchFiltersQuery)

	// Initialize HTTP server
	c.Server = http.NewServer(
		infraConfig,
		domainConfig,
		logger,
		c.NamespaceHandler,
		c.ModuleHandler,
		c.SubmoduleHandler,
		c.ExampleHandler,
		c.AnalyticsHandler,
		c.ProviderHandler,
		c.AuthHandler,
		c.AuditHandler,
		c.InitialSetupHandler,
		c.AuthMiddleware,
		c.TemplateRenderer,
		c.SessionMiddleware,
		c.TerraformV1ModuleHandler, // Pass the new handler to the server constructor
		c.TerraformV2ProviderHandler,
		c.TerraformV2CategoryHandler,
		c.TerraformV2GPGHandler,
		c.TerraformIDPHandler,
		c.TerraformStaticTokenHandler,
		c.ConfigHandler,
		c.VersionHandler,
		c.ProviderLogosHandler,
		c.SearchFiltersHandler,
		c.ModuleWebhookHandler, // Add webhook handler
		c.GraphHandler,
	)

	return c, nil
}

// GetDB returns the database instance
func (c *Container) GetDB() *gorm.DB {
	return c.DB.DB
}
