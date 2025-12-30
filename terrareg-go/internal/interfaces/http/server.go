package http

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/config/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/config"
	terraformHandler "github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/handler/terraform"
	tfv1ModuleHandler "github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/handler/terraform/v1" // New import
	tfv2ProviderHandler "github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/handler/terraform/v2"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/handler/terrareg"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/handler/webhook"
	http_middleware "github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/middleware"
	terrareg_middleware "github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/middleware"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/template"
)

// Server represents the HTTP server
type Server struct {
	router                      *chi.Mux
	infraConfig                 *config.InfrastructureConfig
	domainConfig                *model.DomainConfig
	logger                      zerolog.Logger
	namespaceHandler            *terrareg.NamespaceHandler
	moduleHandler               *terrareg.ModuleHandler
	submoduleHandler            *terrareg.SubmoduleHandler
	exampleHandler              *terrareg.ExampleHandler
	analyticsHandler            *terrareg.AnalyticsHandler
	providerHandler             *terrareg.ProviderHandler
	authHandler                 *terrareg.AuthHandler
	auditHandler                *terrareg.AuditHandler
	initialSetupHandler         *terrareg.InitialSetupHandler
	authMiddleware              *terrareg_middleware.AuthMiddleware
	templateRenderer            *template.Renderer
	sessionMiddleware           *terrareg_middleware.SessionMiddleware
	terraformV1ModuleHandler    *tfv1ModuleHandler.TerraformV1ModuleHandler // New field
	terraformV2ProviderHandler  *tfv2ProviderHandler.TerraformV2ProviderHandler
	terraformV2CategoryHandler  *tfv2ProviderHandler.TerraformV2CategoryHandler
	terraformV2GPGHandler       *tfv2ProviderHandler.TerraformV2GPGHandler
	terraformIDPHandler         *terraformHandler.TerraformIDPHandler
	terraformStaticTokenHandler *terraformHandler.TerraformStaticTokenHandler
	configHandler               *terrareg.ConfigHandler
	versionHandler              *terrareg.VersionHandler
	providerLogosHandler        *terrareg.ProviderLogosHandler
	searchFiltersHandler        *terrareg.SearchFiltersHandler
	moduleWebhookHandler        *webhook.ModuleWebhookHandler
	graphHandler                *terrareg.GraphHandler
	rateLimiter                 *http_middleware.RateLimiterMiddleware
}

// NewServer creates a new HTTP server
func NewServer(
	infraConfig *config.InfrastructureConfig,
	domainConfig *model.DomainConfig,
	logger zerolog.Logger,
	namespaceHandler *terrareg.NamespaceHandler,
	moduleHandler *terrareg.ModuleHandler,
	submoduleHandler *terrareg.SubmoduleHandler,
	exampleHandler *terrareg.ExampleHandler,
	analyticsHandler *terrareg.AnalyticsHandler,
	providerHandler *terrareg.ProviderHandler,
	authHandler *terrareg.AuthHandler,
	auditHandler *terrareg.AuditHandler,
	initialSetupHandler *terrareg.InitialSetupHandler,
	authMiddleware *terrareg_middleware.AuthMiddleware,
	templateRenderer *template.Renderer,
	sessionMiddleware *terrareg_middleware.SessionMiddleware,
	terraformV1ModuleHandler *tfv1ModuleHandler.TerraformV1ModuleHandler, // New parameter
	terraformV2ProviderHandler *tfv2ProviderHandler.TerraformV2ProviderHandler,
	terraformV2CategoryHandler *tfv2ProviderHandler.TerraformV2CategoryHandler,
	terraformV2GPGHandler *tfv2ProviderHandler.TerraformV2GPGHandler,
	terraformIDPHandler *terraformHandler.TerraformIDPHandler,
	terraformStaticTokenHandler *terraformHandler.TerraformStaticTokenHandler,
	configHandler *terrareg.ConfigHandler,
	versionHandler *terrareg.VersionHandler,
	providerLogosHandler *terrareg.ProviderLogosHandler,
	searchFiltersHandler *terrareg.SearchFiltersHandler,
	moduleWebhookHandler *webhook.ModuleWebhookHandler,
	graphHandler *terrareg.GraphHandler,
) *Server {
	s := &Server{
		router:                      chi.NewRouter(),
		infraConfig:                 infraConfig,
		domainConfig:                domainConfig,
		logger:                      logger,
		namespaceHandler:            namespaceHandler,
		moduleHandler:               moduleHandler,
		submoduleHandler:            submoduleHandler,
		exampleHandler:              exampleHandler,
		analyticsHandler:            analyticsHandler,
		providerHandler:             providerHandler,
		authHandler:                 authHandler,
		auditHandler:                auditHandler,
		initialSetupHandler:         initialSetupHandler,
		authMiddleware:              authMiddleware,
		templateRenderer:            templateRenderer,
		sessionMiddleware:           sessionMiddleware,
		terraformV1ModuleHandler:    terraformV1ModuleHandler, // Assign new handler
		terraformV2ProviderHandler:  terraformV2ProviderHandler,
		terraformV2CategoryHandler:  terraformV2CategoryHandler,
		terraformV2GPGHandler:       terraformV2GPGHandler,
		terraformIDPHandler:         terraformIDPHandler,
		terraformStaticTokenHandler: terraformStaticTokenHandler,
		configHandler:               configHandler,
		versionHandler:              versionHandler,
		providerLogosHandler:        providerLogosHandler,
		searchFiltersHandler:        searchFiltersHandler,
		moduleWebhookHandler:        moduleWebhookHandler,
		graphHandler:                graphHandler,
	}

	s.setupMiddleware()
	s.setupRoutes()

	return s
}

// setupMiddleware configures all HTTP middleware
func (s *Server) setupMiddleware() {
	// Standard middleware
	s.router.Use(middleware.RequestID)
	s.router.Use(middleware.RealIP)
	s.router.Use(terrareg_middleware.NewLogger(s.logger))
	s.router.Use(middleware.Recoverer)
	s.router.Use(middleware.Compress(5))

	// Security headers middleware for all routes
	s.router.Use(http_middleware.SecurityHeaders)

	// Initialize rate limiter (10 requests per second, burst of 5)
	s.rateLimiter = http_middleware.NewRateLimiterMiddleware(10, 5)

	// Session middleware for session management
	s.router.Use(s.sessionMiddleware.Session)

	// No global timeout middleware - apply route-specific timeouts only

	// CORS if needed
	// s.router.Use(cors.Handler(cors.Options{
	//     AllowedOrigins:   []string{"*"},
	//     AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
	//     AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
	//     AllowCredentials: true,
	// }))
}

// setupRoutes configures all HTTP routes
func (s *Server) setupRoutes() {
	// Well-known endpoints
	s.router.Get("/.well-known/terraform.json", s.handleTerraformWellKnown)

	// Terraform OIDC Identity Provider endpoints
	s.router.Get("/.well-known/openid-configuration", s.terraformIDPHandler.HandleOpenIDConfiguration)
	s.router.Get("/.well-known/jwks.json", s.terraformIDPHandler.HandleJWKS)
	s.router.Route("/oauth2", func(r chi.Router) {
		r.Get("/auth", s.terraformIDPHandler.HandleAuth)
		r.Post("/token", s.terraformIDPHandler.HandleToken)
		r.Get("/userinfo", s.terraformIDPHandler.HandleUserInfo)
	})

	// Terraform static token validation endpoints
	s.router.Get("/terraform/validate-token", s.terraformStaticTokenHandler.HandleValidateToken)
	s.router.Get("/terraform/auth-status", s.terraformStaticTokenHandler.HandleAuthStatus)

	// Metrics endpoint
	s.router.Get("/metrics", s.handleMetrics)

	// Terraform Registry API v1
	s.router.Route("/v1", func(r chi.Router) {
		// Modules
		r.Get("/modules", s.terraformV1ModuleHandler.HandleModuleList)          // Use the new handler
		r.Get("/modules/search", s.terraformV1ModuleHandler.HandleModuleSearch) // Use the new handler
		r.Get("/modules/{namespace}", s.handleNamespaceModules)
		r.Get("/modules/{namespace}/{name}", s.handleModuleDetails)
		r.Get("/modules/{namespace}/{name}/{provider}/downloads/summary", s.handleModuleDownloadsSummary)                   // Must come before general provider route
		r.Get("/modules/{namespace}/{name}/{provider}", s.terraformV1ModuleHandler.HandleModuleProviderDetails)             // Use the new handler
		r.Get("/modules/{namespace}/{name}/{provider}/versions", s.terraformV1ModuleHandler.HandleModuleVersions)           // Use the new handler
		r.Get("/modules/{namespace}/{name}/{provider}/download", s.terraformV1ModuleHandler.HandleModuleDownload)           // Use the new handler
		r.Get("/modules/{namespace}/{name}/{provider}/{version}", s.terraformV1ModuleHandler.HandleModuleVersionDetails)    // Use the new handler
		r.Get("/modules/{namespace}/{name}/{provider}/{version}/download", s.terraformV1ModuleHandler.HandleModuleDownload) // Use the new handler

		// Providers
		r.Route("/providers", func(r chi.Router) {
			r.Get("/", s.handleProviderList)
			r.Get("/search", s.handleProviderSearch)
			r.Get("/{namespace}", s.handleNamespaceProviders)
			r.Get("/{namespace}/{provider}", s.handleProviderDetails)
			r.Get("/{namespace}/{provider}/{version}", s.handleProviderDetails)
			r.Get("/{namespace}/{provider}/versions", s.handleProviderVersions)
			r.Get("/{namespace}/{provider}/{version}/download/{os}/{arch}", s.handleProviderDownload)
		})

		// Terrareg Custom API
		r.Route("/terrareg", func(r chi.Router) {
			r.With(s.authMiddleware.OptionalAuth).Get("/config", s.handleConfig)
			r.With(s.authMiddleware.OptionalAuth).Get("/git_providers", s.handleGitProviders)
			r.With(s.authMiddleware.OptionalAuth).Get("/health", s.handleHealth)
			r.With(s.authMiddleware.OptionalAuth).Get("/version", s.handleVersion)

			// Analytics
			r.Route("/analytics", func(r chi.Router) {
				r.With(s.authMiddleware.OptionalAuth).Get("/global/stats_summary", s.handleGlobalStatsSummary)
				r.With(s.authMiddleware.OptionalAuth).Get("/global/usage_stats", s.handleGlobalUsageStats)
				r.With(s.authMiddleware.OptionalAuth).Get("/global/most_recently_published_module_version", s.handleMostRecentlyPublished)
				r.With(s.authMiddleware.OptionalAuth).Get("/global/most_downloaded_module_provider_this_week", s.handleMostDownloadedThisWeek)
				r.With(s.authMiddleware.OptionalAuth).Get("/{namespace}/{name}/{provider}/{version}", s.handleModuleVersionAnalytics)
				r.With(s.authMiddleware.OptionalAuth).Get("/{namespace}/{name}/{provider}/token_versions", s.handleAnalyticsTokenVersions)
			})

			// Initial setup
			r.With(s.authMiddleware.OptionalAuth).Get("/initial_setup", s.handleInitialSetup)
			r.With(s.authMiddleware.OptionalAuth).Post("/initial_setup", s.handleInitialSetupPost)

			// Namespaces
			r.With(s.authMiddleware.OptionalAuth).Get("/namespaces", s.handleNamespaceList)
			r.With(s.authMiddleware.RequireAuth).Post("/namespaces", s.handleNamespaceCreate)
			r.With(s.authMiddleware.OptionalAuth).Get("/namespaces/{namespace}", s.handleNamespaceGet)
			r.With(s.authMiddleware.RequireNamespacePermission("FULL", "{namespace}")).Post("/namespaces/{namespace}", s.handleNamespaceUpdate)
			r.With(s.authMiddleware.RequireNamespacePermission("FULL", "{namespace}")).Delete("/namespaces/{namespace}", s.handleNamespaceDelete)

			// Modules
			r.With(s.authMiddleware.OptionalAuth).Get("/modules/{namespace}", s.handleTerraregNamespaceModules)
			r.With(s.authMiddleware.OptionalAuth).Get("/modules/{namespace}/{name}", s.handleTerraregModuleProviders)
			r.With(s.authMiddleware.OptionalAuth).Get("/modules/{namespace}/{name}/{provider}/versions", s.handleTerraregModuleProviderVersions)
			r.With(s.authMiddleware.OptionalAuth).Get("/modules/{namespace}/{name}/{provider}", s.handleTerraregModuleProviderDetails)
			r.With(s.authMiddleware.RequireNamespacePermission("FULL", "{namespace}")).Post("/modules/{namespace}/{name}/{provider}/create", s.handleModuleProviderCreate)
			r.With(s.authMiddleware.RequireNamespacePermission("FULL", "{namespace}")).Delete("/modules/{namespace}/{name}/{provider}/delete", s.handleModuleProviderDelete)
			r.With(s.authMiddleware.OptionalAuth).Get("/modules/{namespace}/{name}/{provider}/settings", s.handleModuleProviderSettings)
			r.With(s.authMiddleware.RequireNamespacePermission("MODIFY", "{namespace}")).Post("/modules/{namespace}/{name}/{provider}/settings", s.handleModuleProviderSettingsUpdate)
			r.With(s.authMiddleware.OptionalAuth).Get("/modules/{namespace}/{name}/{provider}/integrations", s.handleModuleProviderIntegrations)
			r.With(s.authMiddleware.OptionalAuth).Get("/modules/{namespace}/{name}/{provider}/redirects", s.handleModuleProviderRedirects)
			r.With(s.authMiddleware.RequireNamespacePermission("FULL", "{namespace}")).Put("/modules/{namespace}/{name}/{provider}/redirects", s.handleModuleProviderRedirectCreate)
			r.With(s.authMiddleware.RequireNamespacePermission("FULL", "{namespace}")).Delete("/modules/{namespace}/{name}/{provider}/redirects/{redirect_id}", s.handleModuleProviderRedirectDelete)

			// Module webhooks (matching Python implementation)
			r.Post("/modules/{namespace}/{name}/{provider}/hooks/github", s.moduleWebhookHandler.HandleModuleWebhook)
			r.Post("/modules/{namespace}/{name}/{provider}/hooks/bitbucket", s.moduleWebhookHandler.HandleModuleWebhook)
			r.Post("/modules/{namespace}/{name}/{provider}/hooks/gitlab", s.moduleWebhookHandler.HandleModuleWebhook)

			// Module versions
			r.With(s.authMiddleware.OptionalAuth).Get("/modules/{namespace}/{name}/{provider}/{version}", s.handleTerraregModuleVersionDetails)
			r.With(
				s.authMiddleware.RequireUploadPermission("{namespace}"),
				middleware.Timeout(time.Duration(s.infraConfig.ModuleIndexingTimeoutSeconds)*time.Second),
			).Post("/modules/{namespace}/{name}/{provider}/{version}/upload", s.handleModuleVersionUpload)
			r.With(
				middleware.Timeout(time.Duration(s.infraConfig.ModuleIndexingTimeoutSeconds)*time.Second),
				s.authMiddleware.RequireAuth,
			).Post("/modules/{namespace}/{name}/{provider}/{version}/import", s.handleModuleVersionCreate)
			r.With(
				s.authMiddleware.RequireUploadPermission("{namespace}"),
				middleware.Timeout(time.Duration(s.infraConfig.ModuleIndexingTimeoutSeconds)*time.Second),
			).Post("/modules/{namespace}/{name}/{provider}/import", s.handleModuleVersionImport)
			r.With(s.authMiddleware.RequireAuth).Post("/modules/{namespace}/{name}/{provider}/{version}/publish", s.handleModuleVersionPublish)
			r.With(s.authMiddleware.RequireNamespacePermission("FULL", "{namespace}")).Delete("/modules/{namespace}/{name}/{provider}/{version}/delete", s.handleModuleVersionDelete)
			r.With(s.authMiddleware.OptionalAuth).Get("/modules/{namespace}/{name}/{provider}/{version}/readme_html", s.handleModuleVersionReadmeHTML)
			r.With(s.authMiddleware.OptionalAuth).Get("/modules/{namespace}/{name}/{provider}/{version}/variable_template", s.handleModuleVersionVariableTemplate)
			r.With(s.authMiddleware.OptionalAuth).Get("/modules/{namespace}/{name}/{provider}/{version}/files/{path}", s.handleModuleVersionFile)
			r.With(s.authMiddleware.OptionalAuth).Get("/modules/{namespace}/{name}/{provider}/{version}/source.zip", s.handleModuleVersionSourceDownload)

			// Submodules
			r.With(s.authMiddleware.OptionalAuth).Get("/modules/{namespace}/{name}/{provider}/{version}/submodules", s.handleModuleVersionSubmodules)
			r.With(s.authMiddleware.OptionalAuth).Get("/modules/{namespace}/{name}/{provider}/{version}/submodules/details/{submodule:.*}", s.handleSubmoduleDetails)
			r.With(s.authMiddleware.OptionalAuth).Get("/modules/{namespace}/{name}/{provider}/{version}/submodules/readme_html/{submodule:.*}", s.handleSubmoduleReadmeHTML)

			// Examples
			r.With(s.authMiddleware.OptionalAuth).Get("/modules/{namespace}/{name}/{provider}/{version}/examples", s.handleModuleVersionExamples)
			r.With(s.authMiddleware.OptionalAuth).Get("/modules/{namespace}/{name}/{provider}/{version}/examples/details/{example:.*}", s.handleExampleDetails)
			r.With(s.authMiddleware.OptionalAuth).Get("/modules/{namespace}/{name}/{provider}/{version}/examples/readme_html/{example:.*}", s.handleExampleReadmeHTML)
			r.With(s.authMiddleware.OptionalAuth).Get("/modules/{namespace}/{name}/{provider}/{version}/examples/filelist/{example:.*}", s.handleExampleFileList)
			r.With(s.authMiddleware.OptionalAuth).Get("/modules/{namespace}/{name}/{provider}/{version}/examples/file/{file:.*}", s.handleExampleFile)

			// Graph
			r.With(s.authMiddleware.OptionalAuth).Get("/modules/{namespace}/{name}/{provider}/{version}/graph/data", s.handleGraphData)

			// Providers
			r.With(s.authMiddleware.OptionalAuth).Get("/providers/{namespace}", s.handleTerraregNamespaceProviders)
			r.With(s.authMiddleware.OptionalAuth).Get("/providers/{namespace}/{provider}/integrations", s.handleProviderIntegrations)
			r.With(s.authMiddleware.OptionalAuth).Get("/provider_logos", s.handleProviderLogos)

			// Search filters
			r.With(s.authMiddleware.OptionalAuth).Get("/search_filters", s.handleModuleSearchFilters)
			r.With(s.authMiddleware.OptionalAuth).Get("/modules/search/filters", s.handleModuleSearchFilters)
			r.With(s.authMiddleware.OptionalAuth).Get("/providers/search/filters", s.handleProviderSearchFilters)

			// Audit
			r.With(s.authMiddleware.RequireAdmin).Get("/audit-history", s.handleAuditHistory)

			// User groups
			r.With(s.authMiddleware.RequireAdmin).Get("/user-groups", s.handleUserGroupList)
			r.With(s.authMiddleware.RequireAdmin).Post("/user-groups", s.handleUserGroupCreate)
			r.With(s.authMiddleware.RequireAdmin).Get("/user-groups/{group}", s.handleUserGroupDetails)
			r.With(s.authMiddleware.RequireAdmin).Delete("/user-groups/{group}", s.handleUserGroupDelete)
			r.With(s.authMiddleware.RequireAdmin).Get("/user-groups/{group}/permissions/{namespace}", s.handleUserGroupNamespacePermissions)
			r.With(s.authMiddleware.RequireAdmin).Post("/user-groups/{group}/permissions/{namespace}", s.handleUserGroupNamespacePermissionsUpdate)

			// Auth
			r.Route("/auth", func(r chi.Router) {
				r.Post("/admin/login", s.handleAdminLogin)
				r.With(s.sessionMiddleware.Session).Get("/admin/is_authenticated", s.handleIsAuthenticated)
				r.Post("/logout", s.handleLogout)
			})
		})
	})

	// Terraform Registry API v2
	s.router.Route("/v2", func(r chi.Router) {
		// Provider endpoints
		r.With(s.authMiddleware.OptionalAuth).Get("/providers/{namespace}/{provider}", s.terraformV2ProviderHandler.HandleProviderDetails)
		r.With(s.authMiddleware.OptionalAuth).Get("/providers/{namespace}/{provider}/versions", s.terraformV2ProviderHandler.HandleProviderVersions)
		r.With(s.authMiddleware.OptionalAuth).Get("/providers/{namespace}/{provider}/{version}", s.terraformV2ProviderHandler.HandleProviderVersion)
		r.With(s.authMiddleware.OptionalAuth).Get("/providers/{namespace}/{provider}/{version}/download/{os}/{arch}", s.terraformV2ProviderHandler.HandleProviderDownload)
		r.With(s.authMiddleware.OptionalAuth).Get("/providers/{provider_id}/downloads/summary", s.terraformV2ProviderHandler.HandleProviderDownloadsSummary)

		// Provider docs (placeholder - can be implemented later)
		r.With(s.authMiddleware.OptionalAuth).Get("/provider-docs", s.handleV2ProviderDocs)
		r.With(s.authMiddleware.OptionalAuth).Get("/provider-docs/{doc_id}", s.handleV2ProviderDoc)

		// GPG keys
		r.With(s.authMiddleware.OptionalAuth).Get("/gpg-keys", s.terraformV2GPGHandler.HandleListGPGKeys)
		r.With(s.authMiddleware.RequireAdmin).Post("/gpg-keys", s.terraformV2GPGHandler.HandleCreateGPGKey)
		r.With(s.authMiddleware.OptionalAuth).Get("/gpg-keys/{namespace}/{key_id}", s.terraformV2GPGHandler.HandleGetGPGKey)

		// Categories
		r.With(s.authMiddleware.OptionalAuth).Get("/categories", s.terraformV2CategoryHandler.HandleListCategories)
	})

	// Authentication endpoints (matching Python URLs)
	// Apply security headers and rate limiting to auth endpoints
	authMiddlewareChain := func(next http.Handler) http.Handler {
		return s.rateLimiter.RateLimitAuth()(http_middleware.AuthSecurityHeaders(next))
	}

	s.router.With(authMiddlewareChain).Get("/openid/login", s.handleOIDCLogin)
	s.router.With(authMiddlewareChain).Get("/openid/callback", s.handleOIDCCallback)
	s.router.With(authMiddlewareChain).Get("/saml/login", s.handleSAMLLogin)
	s.router.With(authMiddlewareChain).Get("/saml/metadata", s.handleSAMLMetadata)
	s.router.With(authMiddlewareChain).Post("/saml/acs", s.handleSAMLACS)
	s.router.With(authMiddlewareChain).Get("/github/oauth", s.handleGitHubOAuth)

	// Provider source endpoints (GitHub, GitLab, etc.)
	s.router.With(s.authMiddleware.OptionalAuth).Get("/{provider_source}/login", s.handleProviderSourceLogin)
	s.router.With(s.authMiddleware.OptionalAuth).Get("/{provider_source}/callback", s.handleProviderSourceCallback)
	s.router.With(s.authMiddleware.OptionalAuth).Get("/{provider_source}/auth/status", s.handleProviderSourceAuthStatus)
	s.router.With(s.authMiddleware.RequireAuth).Get("/{provider_source}/organizations", s.handleProviderSourceOrganizations)
	s.router.With(s.authMiddleware.RequireAuth).Get("/{provider_source}/repositories", s.handleProviderSourceRepositories)
	s.router.With(s.authMiddleware.RequireAuth).Post("/{provider_source}/refresh-namespace", s.handleProviderSourceRefreshNamespace)
	s.router.With(s.authMiddleware.RequireAuth).Post("/{provider_source}/repositories/{repo_id}/publish-provider", s.handleProviderSourcePublishProvider)

	// Webhooks with extended timeout
	s.router.With(
		s.authMiddleware.OptionalAuth,
		middleware.Timeout(time.Duration(s.infraConfig.ModuleIndexingTimeoutSeconds)*time.Second),
	).Post("/v1/terrareg/modules/{namespace}/{name}/{provider}/hooks/github", s.handleGitHubWebhook)
	s.router.With(
		s.authMiddleware.OptionalAuth,
		middleware.Timeout(time.Duration(s.infraConfig.ModuleIndexingTimeoutSeconds)*time.Second),
	).Post("/v1/terrareg/modules/{namespace}/{name}/{provider}/hooks/bitbucket", s.handleBitBucketWebhook)

	// Test route without auth to test timeout middleware
	s.router.With(
		middleware.Timeout(30*time.Minute), // 30 minutes
	).Post("/v1/terrareg/test-timeout", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Timeout test successful"))
	})

	// Initial Setup API
	s.router.With(s.authMiddleware.OptionalAuth).Get("/v1/terrareg/initial_setup", s.handleInitialSetup)

	// Static files
	fileServer := http.FileServer(http.Dir("./static"))
	s.router.Handle("/static/*", http.StripPrefix("/static/", fileServer))

	// HTML views
	s.router.Get("/", s.handleIndex)
	s.router.Get("/login", s.handleLoginPage)
	s.router.Get("/logout", s.handleLogout)
	s.router.Get("/initial-setup", s.handleInitialSetupPage)
	s.router.Get("/create-namespace", s.handleCreateNamespacePage)
	s.router.Get("/edit-namespace/{namespace}", s.handleEditNamespacePage)
	s.router.Get("/create-module", s.handleCreateModulePage)
	s.router.Get("/create-provider", s.handleCreateProviderPage)
	s.router.Get("/user-groups", s.handleUserGroupsPage)
	s.router.Get("/audit-history", s.handleAuditHistoryPage)

	// Search routes
	s.router.Get("/search", s.handleSearchPage)
	s.router.Get("/search/modules", s.handleModuleSearchPage)
	s.router.Get("/search/providers", s.handleProviderSearchPage)
	s.router.Get("/modules/search", s.handleModuleSearchPage) // Legacy

	// Module pages
	s.router.Get("/modules", s.handleModulesPage)
	s.router.Get("/modules/{namespace}", s.handleNamespacePage)
	s.router.Get("/modules/{namespace}/{name}", s.handleModulePage)
	s.router.Get("/modules/{namespace}/{name}/{provider}", s.handleModuleProviderPage)
	s.router.Get("/modules/{namespace}/{name}/{provider}/{version}", s.handleModuleProviderPage)
	s.router.Get("/modules/{namespace}/{name}/{provider}/{version}/submodule/{submodule:.*}", s.handleSubmodulePage)
	s.router.Get("/modules/{namespace}/{name}/{provider}/{version}/example/{example:.*}", s.handleExamplePage)
	s.router.Get("/modules/{namespace}/{name}/{provider}/{version}/graph", s.handleGraphPage)
	s.router.Get("/modules/{namespace}/{name}/{provider}/{version}/graph/submodule/{submodule:.*}", s.handleGraphPage)
	s.router.Get("/modules/{namespace}/{name}/{provider}/{version}/graph/example/{example:.*}", s.handleGraphPage)

	// Provider pages
	s.router.Get("/providers", s.handleProvidersPage)
	s.router.Get("/providers/{namespace}", s.handleNamespacePage)
	s.router.Get("/providers/{namespace}/{provider}", s.handleProviderPage)
	s.router.Get("/providers/{namespace}/{provider}/latest", s.handleProviderPage)
	s.router.Get("/providers/{namespace}/{provider}/{version}", s.handleProviderPage)
	s.router.Get("/providers/{namespace}/{provider}/{version}/docs", s.handleProviderPage)
	s.router.Get("/providers/{namespace}/{provider}/{version}/docs/{category}/{slug}", s.handleProviderPage)
}

// Start starts the HTTP server with SSL/TLS and server type configuration
func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.infraConfig.ListenPort)

	// Determine if SSL/TLS should be enabled
	sslEnabled := s.infraConfig.SSLCertPrivateKey != "" && s.infraConfig.SSLCertPublicKey != ""

	if sslEnabled {
		s.logger.Info().
			Str("addr", addr).
			Bool("ssl_enabled", true).
			Msg("Starting HTTPS server")

		return s.startHTTPS(addr)
	} else {
		s.logger.Info().
			Str("addr", addr).
			Bool("ssl_enabled", false).
			Msg("Starting HTTP server")

		return s.startHTTP(addr)
	}
}

// startHTTP starts the HTTP server for non-SSL configuration
func (s *Server) startHTTP(addr string) error {
	// Use server type configuration from Python terrareg
	switch s.infraConfig.ServerType {
	case model.ServerTypeWaitress:
		// For Waitress compatibility, we could add custom server implementation
		// For now, use built-in HTTP server
		s.logger.Warn().
			Str("server_type", string(s.infraConfig.ServerType)).
			Msg("Waitress server type not implemented, using built-in HTTP server")
		fallthrough
	case model.ServerTypeBuiltin:
		fallthrough
	default:
		// Default built-in server
		// Use a longer timeout than standard requests to accommodate module processing
		// The StandardRequestTimeoutSeconds is for regular requests, but server-level timeout
		// should be longer to allow route-level timeouts to work properly
		serverTimeout := max(
			s.infraConfig.StandardRequestTimeoutSeconds,
			300, // 5 minutes minimum to accommodate route-level timeouts
		)
		server := &http.Server{
			Addr:         addr,
			Handler:      s.router,
			ReadTimeout:  time.Duration(serverTimeout) * time.Second,
			WriteTimeout: time.Duration(serverTimeout) * time.Second,
			IdleTimeout:  120 * time.Second,
		}

		return server.ListenAndServe()
	}
}

// startHTTPS starts the HTTPS server with SSL/TLS configuration
func (s *Server) startHTTPS(addr string) error {
	// Load SSL certificates
	cert, err := tls.LoadX509KeyPair(s.infraConfig.SSLCertPublicKey, s.infraConfig.SSLCertPrivateKey)
	if err != nil {
		return fmt.Errorf("failed to load SSL certificates: %w", err)
	}

	// Configure TLS
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
		ServerName:   s.infraConfig.DomainName, // Use domain name from config
	}

	// Create HTTPS server
	// Use a longer timeout than standard requests to accommodate module processing
	// The StandardRequestTimeoutSeconds is for regular requests, but server-level timeout
	// should be longer to allow route-level timeouts to work properly
	serverTimeout := max(
		s.infraConfig.StandardRequestTimeoutSeconds,
		300, // 5 minutes minimum to accommodate route-level timeouts
	)
	server := &http.Server{
		Addr:         addr,
		Handler:      s.router,
		TLSConfig:    tlsConfig,
		ReadTimeout:  time.Duration(serverTimeout) * time.Second,
		WriteTimeout: time.Duration(serverTimeout) * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	s.logger.Info().
		Str("domain", s.infraConfig.DomainName).
		Str("public_url", s.infraConfig.PublicURL).
		Msg("HTTPS server configured with SSL certificates")

	return server.ListenAndServeTLS("", "") // Certificates already loaded in TLSConfig
}

// Router returns the chi router (useful for testing)
func (s *Server) Router() *chi.Mux {
	return s.router
}

// Placeholder handlers - these will be implemented in separate files
func (s *Server) handleTerraformWellKnown(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"modules.v1":   "/v1/modules/",
		"providers.v1": "/v1/providers/",
	})
}

func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	// Prometheus metrics endpoint - to be implemented
	w.Write([]byte("# Metrics endpoint\n"))
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status": "ok",
	})
}

func (s *Server) handleVersion(w http.ResponseWriter, r *http.Request) {
	s.versionHandler.HandleVersion(w, r)
}

// Helper functions
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	// JSON encoding to be implemented
}

func respondError(w http.ResponseWriter, err error, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	// Error response to be implemented
}

// All other handlers are stubs for now - they will be implemented in Phase 2+
// This allows the server to compile and run

func (s *Server) handleNamespaceModules(w http.ResponseWriter, r *http.Request) {
	s.moduleHandler.HandleNamespaceModules(w, r)
}
func (s *Server) handleModuleDetails(w http.ResponseWriter, r *http.Request) {
	s.moduleHandler.HandleModuleDetails(w, r)
}
func (s *Server) handleModuleDownloadsSummary(w http.ResponseWriter, r *http.Request) {
	s.analyticsHandler.HandleModuleDownloadsSummary(w, r)
}
func (s *Server) handleProviderList(w http.ResponseWriter, r *http.Request) {
	s.providerHandler.HandleProviderList(w, r)
}

func (s *Server) handleProviderSearch(w http.ResponseWriter, r *http.Request) {
	s.providerHandler.HandleProviderSearch(w, r)
}

func (s *Server) handleNamespaceProviders(w http.ResponseWriter, r *http.Request) {
	s.providerHandler.HandleNamespaceProviders(w, r)
}

func (s *Server) handleProviderDetails(w http.ResponseWriter, r *http.Request) {
	s.providerHandler.HandleProviderDetails(w, r)
}

func (s *Server) handleProviderVersions(w http.ResponseWriter, r *http.Request) {
	s.providerHandler.HandleProviderVersions(w, r)
}

func (s *Server) handleProviderDownload(w http.ResponseWriter, r *http.Request) {
	// Delegate to the provider handler following DDD principles
	s.providerHandler.HandleProviderDownload(w, r)
}
func (s *Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	s.configHandler.HandleConfig(w, r)
}
func (s *Server) handleGitProviders(w http.ResponseWriter, r *http.Request) {
	s.providerHandler.HandleProviderList(w, r)
}
func (s *Server) handleGlobalStatsSummary(w http.ResponseWriter, r *http.Request) {
	s.analyticsHandler.HandleGlobalStatsSummary(w, r)
}
func (s *Server) handleGlobalUsageStats(w http.ResponseWriter, r *http.Request) {
	s.analyticsHandler.HandleGlobalUsageStats(w, r)
}
func (s *Server) handleMostRecentlyPublished(w http.ResponseWriter, r *http.Request) {
	s.analyticsHandler.HandleMostRecentlyPublished(w, r)
}
func (s *Server) handleMostDownloadedThisWeek(w http.ResponseWriter, r *http.Request) {
	s.analyticsHandler.HandleMostDownloadedThisWeek(w, r)
}
func (s *Server) handleModuleVersionAnalytics(w http.ResponseWriter, r *http.Request) {
	s.analyticsHandler.HandleModuleDownloadsSummary(w, r)
}
func (s *Server) handleAnalyticsTokenVersions(w http.ResponseWriter, r *http.Request) {
	s.analyticsHandler.HandleTokenVersions(w, r)
}
func (s *Server) handleInitialSetup(w http.ResponseWriter, r *http.Request) {
	s.initialSetupHandler.HandleInitialSetup(w, r)
}
func (s *Server) handleInitialSetupPost(w http.ResponseWriter, r *http.Request) {
	// For now, use the same GET handler for POST
	s.initialSetupHandler.HandleInitialSetup(w, r)
}
func (s *Server) handleNamespaceList(w http.ResponseWriter, r *http.Request) {
	s.namespaceHandler.HandleNamespaceList(w, r)
}
func (s *Server) handleNamespaceCreate(w http.ResponseWriter, r *http.Request) {
	s.namespaceHandler.HandleNamespaceCreate(w, r)
}
func (s *Server) handleNamespaceGet(w http.ResponseWriter, r *http.Request) {
	s.namespaceHandler.HandleNamespaceDetails(w, r)
}
func (s *Server) handleNamespaceUpdate(w http.ResponseWriter, r *http.Request) {
	s.namespaceHandler.HandleNamespaceUpdate(w, r)
}
func (s *Server) handleNamespaceDelete(w http.ResponseWriter, r *http.Request) {
	s.namespaceHandler.HandleNamespaceDelete(w, r)
}
func (s *Server) handleTerraregNamespaceModules(w http.ResponseWriter, r *http.Request) {
	s.moduleHandler.HandleNamespaceModules(w, r)
}
func (s *Server) handleTerraregModuleProviders(w http.ResponseWriter, r *http.Request) {
	s.moduleHandler.HandleModuleProviderDetails(w, r)
}
func (s *Server) handleTerraregModuleProviderDetails(w http.ResponseWriter, r *http.Request) {
	s.moduleHandler.HandleTerraregModuleProviderDetails(w, r)
}
func (s *Server) handleTerraregModuleVersionDetails(w http.ResponseWriter, r *http.Request) {
	s.moduleHandler.HandleModuleVersionDetails(w, r)
}
func (s *Server) handleTerraregModuleProviderVersions(w http.ResponseWriter, r *http.Request) {
	s.moduleHandler.HandleTerraregModuleProviderVersions(w, r)
}
func (s *Server) handleModuleProviderCreate(w http.ResponseWriter, r *http.Request) {
	s.moduleHandler.HandleModuleProviderCreate(w, r)
}
func (s *Server) handleModuleProviderDelete(w http.ResponseWriter, r *http.Request) {
	s.moduleHandler.HandleModuleProviderDelete(w, r)
}
func (s *Server) handleModuleProviderSettings(w http.ResponseWriter, r *http.Request) {
	s.moduleHandler.HandleModuleProviderSettingsGet(w, r)
}
func (s *Server) handleModuleProviderSettingsUpdate(w http.ResponseWriter, r *http.Request) {
	s.moduleHandler.HandleModuleProviderSettingsUpdate(w, r)
}
func (s *Server) handleModuleProviderIntegrations(w http.ResponseWriter, r *http.Request) {
	s.moduleHandler.HandleGetIntegrations(w, r)
}
func (s *Server) handleModuleProviderRedirects(w http.ResponseWriter, r *http.Request) {
	s.moduleHandler.HandleModuleProviderRedirectsGet(w, r)
}
func (s *Server) handleModuleProviderRedirectCreate(w http.ResponseWriter, r *http.Request) {
	s.moduleHandler.HandleModuleProviderRedirectCreate(w, r)
}
func (s *Server) handleModuleProviderRedirectDelete(w http.ResponseWriter, r *http.Request) {
	s.moduleHandler.HandleModuleProviderRedirectDelete(w, r)
}
func (s *Server) handleModuleVersionUpload(w http.ResponseWriter, r *http.Request) {
	s.moduleHandler.HandleModuleVersionUpload(w, r)
}
func (s *Server) handleModuleVersionCreate(w http.ResponseWriter, r *http.Request) {
	// Delegate to the module handler following DDD principles
	// This is the deprecated endpoint that requires version in URL
	s.moduleHandler.HandleModuleVersionCreate(w, r)
}
func (s *Server) handleModuleVersionImport(w http.ResponseWriter, r *http.Request) {
	s.moduleHandler.HandleModuleVersionImport(w, r)
}
func (s *Server) handleModuleVersionPublish(w http.ResponseWriter, r *http.Request) {
	s.moduleHandler.HandleModuleVersionPublish(w, r)
}
func (s *Server) handleModuleVersionDelete(w http.ResponseWriter, r *http.Request) {
	// Delegate to the module handler following DDD principles
	// This deletes a specific version, not the entire provider
	s.moduleHandler.HandleModuleVersionDelete(w, r)
}
func (s *Server) handleModuleVersionReadmeHTML(w http.ResponseWriter, r *http.Request) {
	// Delegate to the module handler following DDD principles
	s.moduleHandler.HandleModuleVersionReadmeHTML(w, r)
}
func (s *Server) handleModuleVersionVariableTemplate(w http.ResponseWriter, r *http.Request) {
	// Delegate to the module handler following DDD principles
	s.moduleHandler.HandleModuleVersionVariableTemplate(w, r)
}
func (s *Server) handleModuleVersionFile(w http.ResponseWriter, r *http.Request) {
	// Delegate to the module handler following DDD principles
	s.moduleHandler.HandleModuleFile(w, r)
}
func (s *Server) handleModuleVersionSourceDownload(w http.ResponseWriter, r *http.Request) {
	// Check if module hosting is disallowed
	if s.domainConfig.AllowModuleHosting == model.ModuleHostingModeDisallow {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message": "Module hosting is disabled"}`))
		return
	}

	// Get path parameters
	namespace := chi.URLParam(r, "namespace")
	name := chi.URLParam(r, "name")
	provider := chi.URLParam(r, "provider")
	version := chi.URLParam(r, "version")

	if namespace == "" || name == "" || provider == "" || version == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message": "Missing required path parameters"}`))
		return
	}

	// Delegate to the module handler following DDD principles
	s.moduleHandler.HandleModuleVersionSourceDownload(w, r)
}
func (s *Server) handleModuleVersionSubmodules(w http.ResponseWriter, r *http.Request) {
	s.moduleHandler.HandleGetSubmodules(w, r)
}
func (s *Server) handleSubmoduleDetails(w http.ResponseWriter, r *http.Request) {
	s.submoduleHandler.HandleSubmoduleDetails(w, r)
}
func (s *Server) handleSubmoduleReadmeHTML(w http.ResponseWriter, r *http.Request) {
	s.submoduleHandler.HandleSubmoduleReadmeHTML(w, r)
}
func (s *Server) handleModuleVersionExamples(w http.ResponseWriter, r *http.Request) {
	s.moduleHandler.HandleGetExamples(w, r)
}
func (s *Server) handleExampleDetails(w http.ResponseWriter, r *http.Request) {
	s.exampleHandler.HandleExampleDetails(w, r)
}
func (s *Server) handleExampleReadmeHTML(w http.ResponseWriter, r *http.Request) {
	s.exampleHandler.HandleExampleReadmeHTML(w, r)
}
func (s *Server) handleExampleFileList(w http.ResponseWriter, r *http.Request) {
	s.exampleHandler.HandleExampleFileList(w, r)
}
func (s *Server) handleExampleFile(w http.ResponseWriter, r *http.Request) {
	s.exampleHandler.HandleExampleFile(w, r)
}
func (s *Server) handleGraphData(w http.ResponseWriter, r *http.Request) {
	// For now, return a placeholder response
	s.graphHandler.HandleModuleDependencyGraph(w, r)
}
func (s *Server) handleTerraregNamespaceProviders(w http.ResponseWriter, r *http.Request) {
	s.moduleHandler.HandleModuleProviderDetails(w, r)
}
func (s *Server) handleProviderIntegrations(w http.ResponseWriter, r *http.Request) {
	// Provider integrations not yet implemented
	respondJSON(w, http.StatusNotImplemented, map[string]interface{}{
		"message": "Provider integrations not yet implemented",
	})
}
func (s *Server) handleProviderLogos(w http.ResponseWriter, r *http.Request) {
	s.providerLogosHandler.HandleGetProviderLogos(w, r)
}
func (s *Server) handleModuleSearchFilters(w http.ResponseWriter, r *http.Request) {
	s.searchFiltersHandler.HandleModuleSearchFilters(w, r)
}

func (s *Server) handleProviderSearchFilters(w http.ResponseWriter, r *http.Request) {
	// Provider search filters not yet implemented
	respondJSON(w, http.StatusNotImplemented, map[string]interface{}{
		"message": "Provider search filters not yet implemented",
	})
}
func (s *Server) handleAuditHistory(w http.ResponseWriter, r *http.Request) {
	s.auditHandler.HandleAuditHistoryGet(w, r)
}
func (s *Server) handleUserGroupList(w http.ResponseWriter, r *http.Request) {
	// For now, return a placeholder response
	respondJSON(w, http.StatusNotImplemented, map[string]interface{}{
		"message": "User group list not yet implemented",
	})
}
func (s *Server) handleUserGroupCreate(w http.ResponseWriter, r *http.Request) {
	// For now, return a placeholder response
	respondJSON(w, http.StatusNotImplemented, map[string]interface{}{
		"message": "User group creation not yet implemented",
	})
}
func (s *Server) handleUserGroupDetails(w http.ResponseWriter, r *http.Request) {
	// For now, return a placeholder response
	respondJSON(w, http.StatusNotImplemented, map[string]interface{}{
		"message": "User group details not yet implemented",
	})
}
func (s *Server) handleUserGroupDelete(w http.ResponseWriter, r *http.Request) {
	// For now, return a placeholder response
	respondJSON(w, http.StatusNotImplemented, map[string]interface{}{
		"message": "User group deletion not yet implemented",
	})
}
func (s *Server) handleUserGroupNamespacePermissions(w http.ResponseWriter, r *http.Request) {
	// For now, return a placeholder response
	respondJSON(w, http.StatusNotImplemented, map[string]interface{}{
		"message": "User group namespace permissions not yet implemented",
	})
}
func (s *Server) handleUserGroupNamespacePermissionsUpdate(w http.ResponseWriter, r *http.Request) {
	// For now, return a placeholder response
	respondJSON(w, http.StatusNotImplemented, map[string]interface{}{
		"message": "User group namespace permissions update not yet implemented",
	})
}
func (s *Server) handleAdminLogin(w http.ResponseWriter, r *http.Request) {
	s.authHandler.HandleAdminLogin(w, r)
}
func (s *Server) handleIsAuthenticated(w http.ResponseWriter, r *http.Request) {
	s.authHandler.HandleIsAuthenticated(w, r)
}
func (s *Server) handleV2ProviderDetails(w http.ResponseWriter, r *http.Request) {
	s.terraformV2ProviderHandler.HandleProviderDetails(w, r)
}
func (s *Server) handleV2ProviderDownloadsSummary(w http.ResponseWriter, r *http.Request) {
	s.terraformV2ProviderHandler.HandleProviderDownloadsSummary(w, r)
}
func (s *Server) handleV2ProviderDocs(w http.ResponseWriter, r *http.Request) {
	// Provider docs not yet implemented
	respondJSON(w, http.StatusNotImplemented, map[string]interface{}{
		"message": "Provider docs not yet implemented",
	})
}
func (s *Server) handleV2ProviderDoc(w http.ResponseWriter, r *http.Request) {
	// Provider doc not yet implemented
	respondJSON(w, http.StatusNotImplemented, map[string]interface{}{
		"message": "Provider doc not yet implemented",
	})
}
func (s *Server) handleV2GPGKeys(w http.ResponseWriter, r *http.Request) {
	s.terraformV2GPGHandler.HandleListGPGKeys(w, r)
}
func (s *Server) handleV2GPGKeyCreate(w http.ResponseWriter, r *http.Request) {
	s.terraformV2GPGHandler.HandleCreateGPGKey(w, r)
}
func (s *Server) handleV2GPGKey(w http.ResponseWriter, r *http.Request) {
	s.terraformV2GPGHandler.HandleGetGPGKey(w, r)
}
func (s *Server) handleV2Categories(w http.ResponseWriter, r *http.Request) {
	s.terraformV2CategoryHandler.HandleListCategories(w, r)
}
func (s *Server) handleOIDCLogin(w http.ResponseWriter, r *http.Request) {
	s.authHandler.HandleOIDCLogin(w, r)
}
func (s *Server) handleOIDCCallback(w http.ResponseWriter, r *http.Request) {
	s.authHandler.HandleOIDCCallback(w, r)
}
func (s *Server) handleSAMLLogin(w http.ResponseWriter, r *http.Request) {
	s.authHandler.HandleSAMLLogin(w, r)
}
func (s *Server) handleSAMLMetadata(w http.ResponseWriter, r *http.Request) {
	s.authHandler.HandleSAMLMetadata(w, r)
}
func (s *Server) handleSAMLACS(w http.ResponseWriter, r *http.Request) {
	s.authHandler.HandleSAMLACS(w, r)
}
func (s *Server) handleGitHubOAuth(w http.ResponseWriter, r *http.Request) {
	s.authHandler.HandleGitHubOAuth(w, r)
}
func (s *Server) handleProviderSourceLogin(w http.ResponseWriter, r *http.Request) {
	providerSource := chi.URLParam(r, "provider_source")

	// Handle GitHub OAuth
	if providerSource == "github" {
		s.authHandler.HandleGitHubOAuth(w, r)
		return
	}

	// Other provider sources not yet implemented
	respondJSON(w, http.StatusNotImplemented, map[string]interface{}{
		"message": "Provider source login not yet implemented",
	})
}
func (s *Server) handleProviderSourceCallback(w http.ResponseWriter, r *http.Request) {
	// Provider source callback not yet implemented
	respondJSON(w, http.StatusNotImplemented, map[string]interface{}{
		"message": "Provider source callback not yet implemented",
	})
}
func (s *Server) handleProviderSourceAuthStatus(w http.ResponseWriter, r *http.Request) {
	// Provider source auth status not yet implemented
	respondJSON(w, http.StatusNotImplemented, map[string]interface{}{
		"message": "Provider source auth status not yet implemented",
	})
}
func (s *Server) handleProviderSourceOrganizations(w http.ResponseWriter, r *http.Request) {
	// Provider source organizations not yet implemented
	respondJSON(w, http.StatusNotImplemented, map[string]interface{}{
		"message": "Provider source organizations not yet implemented",
	})
}
func (s *Server) handleProviderSourceRepositories(w http.ResponseWriter, r *http.Request) {
	// Provider source repositories not yet implemented
	respondJSON(w, http.StatusNotImplemented, map[string]interface{}{
		"message": "Provider source repositories not yet implemented",
	})
}
func (s *Server) handleProviderSourceRefreshNamespace(w http.ResponseWriter, r *http.Request) {
	// Provider source refresh namespace not yet implemented
	respondJSON(w, http.StatusNotImplemented, map[string]interface{}{
		"message": "Provider source refresh namespace not yet implemented",
	})
}
func (s *Server) handleProviderSourcePublishProvider(w http.ResponseWriter, r *http.Request) {
	// Provider source publish provider not yet implemented
	respondJSON(w, http.StatusNotImplemented, map[string]interface{}{
		"message": "Provider source publish provider not yet implemented",
	})
}
func (s *Server) handleGitHubWebhook(w http.ResponseWriter, r *http.Request) {
	// GitHub webhook not yet implemented
	respondJSON(w, http.StatusNotImplemented, map[string]interface{}{
		"message": "GitHub webhook not yet implemented",
	})
}
func (s *Server) handleBitBucketWebhook(w http.ResponseWriter, r *http.Request) {
	// BitBucket webhook not yet implemented
	respondJSON(w, http.StatusNotImplemented, map[string]interface{}{
		"message": "BitBucket webhook not yet implemented",
	})
}
func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	// Render the index template using the template renderer
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := s.templateRenderer.RenderWithRequest(r.Context(), w, "index.html", map[string]interface{}{
		"TEMPLATE_NAME": "index.html",
	}, r)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		s.logger.Error().Err(err).Msg("Failed to render index template")
	}
}
func (s *Server) handleLoginPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := s.templateRenderer.RenderWithRequest(r.Context(), w, "login.html", map[string]interface{}{
		"TEMPLATE_NAME": "login.html",
	}, r)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		s.logger.Error().Err(err).Msg("Failed to render login template")
	}
}
func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	// Clear session cookies directly since we don't have a centralized cookie clearing method
	// This matches the pattern of setting cookies with MaxAge=-1 to clear them
	http.SetCookie(w, &http.Cookie{
		Name:     "terrareg_session",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "is_admin_authenticated",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
func (s *Server) handleInitialSetupPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := s.templateRenderer.RenderWithRequest(r.Context(), w, "initial_setup.html", map[string]interface{}{
		"TEMPLATE_NAME": "initial_setup.html",
	}, r)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		s.logger.Error().Err(err).Msg("Failed to render initial_setup template")
	}
}
func (s *Server) handleCreateNamespacePage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := s.templateRenderer.RenderWithRequest(r.Context(), w, "create_namespace.html", map[string]interface{}{
		"TEMPLATE_NAME": "create_namespace.html",
	}, r)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		s.logger.Error().Err(err).Msg("Failed to render create_namespace template")
	}
}
func (s *Server) handleEditNamespacePage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := s.templateRenderer.RenderWithRequest(r.Context(), w, "namespace.html", map[string]interface{}{
		"TEMPLATE_NAME": "edit_namespace.html",
	}, r)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		s.logger.Error().Err(err).Msg("Failed to render namespace template")
	}
}
func (s *Server) handleCreateModulePage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := s.templateRenderer.RenderWithRequest(r.Context(), w, "create_module_provider.html", map[string]interface{}{
		"TEMPLATE_NAME": "create_module_provider.html",
	}, r)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		s.logger.Error().Err(err).Msg("Failed to render create_module_provider template")
	}
}
func (s *Server) handleCreateProviderPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := s.templateRenderer.RenderWithRequest(r.Context(), w, "create_provider.html", map[string]interface{}{
		"TEMPLATE_NAME": "create_provider.html",
	}, r)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		s.logger.Error().Err(err).Msg("Failed to render create_provider template")
	}
}
func (s *Server) handleUserGroupsPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := s.templateRenderer.RenderWithRequest(r.Context(), w, "user_groups.html", map[string]interface{}{
		"TEMPLATE_NAME": "user_groups.html",
	}, r)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		s.logger.Error().Err(err).Msg("Failed to render user_groups template")
	}
}
func (s *Server) handleAuditHistoryPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := s.templateRenderer.RenderWithRequest(r.Context(), w, "audit_history.html", map[string]interface{}{
		"TEMPLATE_NAME": "audit_history.html",
	}, r)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		s.logger.Error().Err(err).Msg("Failed to render audit_history template")
	}
}
func (s *Server) handleSearchPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := s.templateRenderer.RenderWithRequest(r.Context(), w, "search.html", map[string]interface{}{
		"TEMPLATE_NAME": "search.html",
	}, r)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		s.logger.Error().Err(err).Msg("Failed to render search template")
	}
}
func (s *Server) handleModuleSearchPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := s.templateRenderer.RenderWithRequest(r.Context(), w, "module_search.html", map[string]interface{}{
		"TEMPLATE_NAME": "module_search.html",
	}, r)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		s.logger.Error().Err(err).Msg("Failed to render module_search template")
	}
}
func (s *Server) handleProviderSearchPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := s.templateRenderer.RenderWithRequest(r.Context(), w, "provider_search.html", map[string]interface{}{
		"TEMPLATE_NAME": "provider_search.html",
	}, r)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		s.logger.Error().Err(err).Msg("Failed to render provider_search template")
	}
}
func (s *Server) handleModulesPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := s.templateRenderer.RenderWithRequest(r.Context(), w, "module.html", map[string]interface{}{
		"TEMPLATE_NAME": "module.html",
	}, r)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		s.logger.Error().Err(err).Msg("Failed to render module template")
	}
}
func (s *Server) handleNamespacePage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := s.templateRenderer.RenderWithRequest(r.Context(), w, "namespace.html", map[string]interface{}{
		"TEMPLATE_NAME": "namespace.html",
	}, r)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		s.logger.Error().Err(err).Msg("Failed to render namespace template")
	}
}
func (s *Server) handleModulePage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := s.templateRenderer.RenderWithRequest(r.Context(), w, "module.html", map[string]interface{}{
		"TEMPLATE_NAME": "module.html",
	}, r)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		s.logger.Error().Err(err).Msg("Failed to render module template")
	}
}
func (s *Server) handleModuleProviderPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := s.templateRenderer.RenderWithRequest(r.Context(), w, "module_provider.html", map[string]interface{}{
		"TEMPLATE_NAME": "module_provider.html",
	}, r)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		s.logger.Error().Err(err).Msg("Failed to render module_provider template")
	}
}
func (s *Server) handleSubmodulePage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := s.templateRenderer.RenderWithRequest(r.Context(), w, "submodule.html", map[string]interface{}{
		"TEMPLATE_NAME": "submodule.html",
	}, r)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		s.logger.Error().Err(err).Msg("Failed to render submodule template")
	}
}
func (s *Server) handleExamplePage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := s.templateRenderer.RenderWithRequest(r.Context(), w, "example.html", map[string]interface{}{
		"TEMPLATE_NAME": "example.html",
	}, r)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		s.logger.Error().Err(err).Msg("Failed to render example template")
	}
}
func (s *Server) handleGraphPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Extract path parameters
	namespace := chi.URLParam(r, "namespace")
	moduleName := chi.URLParam(r, "name")
	provider := chi.URLParam(r, "provider")
	version := chi.URLParam(r, "version")

	// Create template data with graph_data_url for the template
	templateData := map[string]interface{}{
		"TEMPLATE_NAME":  "graph.html",
		"graph_data_url": fmt.Sprintf("/v1/terrareg/modules/%s/%s/%s/graph/data", namespace, moduleName, provider, version),
		"namespace":      namespace,
		"module":         moduleName,
		"provider":       provider,
		"version":        version,
	}

	err := s.templateRenderer.RenderWithRequest(r.Context(), w, "graph.html", templateData, r)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		s.logger.Error().Err(err).Msg("Failed to render graph template")
	}
}
func (s *Server) handleProvidersPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := s.templateRenderer.RenderWithRequest(r.Context(), w, "provider.html", map[string]interface{}{
		"TEMPLATE_NAME": "provider.html",
	}, r)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		s.logger.Error().Err(err).Msg("Failed to render provider template")
	}
}
func (s *Server) handleProviderPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := s.templateRenderer.RenderWithRequest(r.Context(), w, "provider.html", map[string]interface{}{
		"TEMPLATE_NAME": "provider.html",
	}, r)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		s.logger.Error().Err(err).Msg("Failed to render provider template")
	}
}

// GetRouter returns the HTTP router for testing purposes
func (s *Server) GetRouter() *chi.Mux {
	return s.router
}
