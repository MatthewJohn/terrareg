package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/config/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/config/service"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/config"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/container"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/version"
	httpServer "github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http"
	"github.com/matthewjohn/terrareg/terrareg-go/test/integration/testutils"
)

func TestSubmodulesExamples_EndToEndIntegration(t *testing.T) {
	// Setup test database
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Setup configs
	domainCfg := &model.DomainConfig{
		AllowModuleHosting:       model.ModuleHostingModeAllow,
		TrustedNamespaces:        []string{},
		VerifiedModuleNamespaces: []string{},
		TrustedNamespaceLabel:     "Trusted",
		ContributedNamespaceLabel: "Contributed",
		VerifiedModuleLabel:       "Verified",
		AnalyticsTokenPhrase:      "",
		AnalyticsTokenDescription: "",
		ExampleAnalyticsToken:     "my-tf-application",
		DisableAnalytics:          false,
		UploadAPIKeysEnabled:      false,
		PublishAPIKeysEnabled:     false,
		DisableTerraregExclusiveLabels: false,
		AllowCustomGitURLModuleProvider: true,
		AllowCustomGitURLModuleVersion:  true,
		SecretKeySet:              false,
		OpenIDConnectEnabled:      false,
		OpenIDConnectLoginText:    "Login with OpenID",
		SAMLEnabled:               false,
		SAMLLoginText:             "Login with SAML",
		AdminLoginEnabled:         false,
		AdditionalModuleTabs:      []string{},
		AutoCreateNamespace:       true,
		AutoCreateModuleProvider:  true,
		DefaultUiDetailsView:      model.DefaultUiInputOutputViewTable,
		TerraformExampleVersionTemplate: "",
		TerraformExampleVersionTemplatePreMajor: "",
		ProviderSources:           make(map[string]model.ProviderSourceConfig),
	}

	infraCfg := &config.InfrastructureConfig{
		ListenPort:  5000,
		PublicURL:   "http://localhost:5000",
		DatabaseURL: ":memory:",
		Debug:       true,
		DomainName:                 "",
		DataDirectory:              "./data",
		UploadDirectory:            "./data/upload",
		GitProviderConfig:          "",
		SAML2IDPMetadataURL:        "",
		SAML2IssuerEntityID:        "",
		OpenIDConnectClientID:      "",
		OpenIDConnectClientSecret:  "",
		OpenIDConnectIssuer:        "",
		AdminAuthenticationToken:   "",
		SecretKey:                  "",
		AllowProviderHosting:       true,
		AllowCustomGitProvider:     true,
		EnableAccessControls:       false,
		EnableSecurityScanning:     true,
		ApplicationName:            "Terrareg",
		LogoURL:                    "/static/images/logo.png",
		SiteWarning:                "",
		SessionExpiry:              0,
		AdminSessionExpiryMins:     60,
		SessionCookieName:          "terrareg_session",
		SessionRefreshAge:          0,
		InfracostAPIKey:            "",
		InfracostPricingAPIEndpoint: "",
		SentryDSN:                  "",
		SentryTracesSampleRate:     0.1,
	}

	// Create config service
	versionReader := version.NewVersionReader()
	configService := service.NewConfigurationService(service.ConfigurationServiceOptions{}, versionReader)

	// Create container with test dependencies
	container, err := container.NewContainer(domainCfg, infraCfg, configService, testutils.TestLogger, db)
	require.NoError(t, err)

	// Create server
	server := httpServer.NewServer(container)

	// Create test data in database
	namespace := testutils.CreateNamespace(t, db, "testns")
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "testmod", "testprov")
	moduleVersion := testutils.CreateModuleVersion(t, db, moduleProvider.ID, "1.0.0")
	moduleDetails := testutils.CreateModuleDetails(t, db, "# Test README")

	// Update module version with details
	err = db.DB.Model(&sqldb.ModuleVersionDB{}).Where("id = ?", moduleVersion.ID).Update("module_details_id", moduleDetails.ID).Error
	require.NoError(t, err)

	// Create submodules
	testutils.CreateSubmodule(t, db, moduleVersion.ID, "submodule1", "Submodule One", "", moduleDetails.ID)
	testutils.CreateSubmodule(t, db, moduleVersion.ID, "submodule2", "Submodule Two", "local", moduleDetails.ID)

	// Create examples (examples are also submodules with type="example")
	exampleDetails := testutils.CreateModuleDetails(t, db, "# Example README")
	example1 := testutils.CreateSubmodule(t, db, moduleVersion.ID, "example1", "Example One", "example", exampleDetails.ID)
	testutils.CreateExampleFile(t, db, example1.ID, "main.tf", "resource \"null_resource\" \"example\" {}")
	testutils.CreateExampleFile(t, db, example1.ID, "README.md", "# Example")

	example2 := testutils.CreateSubmodule(t, db, moduleVersion.ID, "example2", "Example Two", "example", exampleDetails.ID)
	testutils.CreateExampleFile(t, db, example2.ID, "main.tf", "resource \"local_file\" \"example\" {}")

	t.Run("Submodules endpoint returns database data", func(t *testing.T) {
		// Create request
		req := httptest.NewRequest("GET", "/v1/terrareg/modules/testns/testmod/testprov/1.0.0/submodules", nil)

		// Setup chi router context
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("namespace", "testns")
		rctx.URLParams.Add("name", "testmod")
		rctx.URLParams.Add("provider", "testprov")
		rctx.URLParams.Add("version", "1.0.0")

		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		// Create response recorder
		w := httptest.NewRecorder()

		// Execute request
		server.Router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code)

		// Response should be a direct array
		var response []interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Len(t, response, 2) // Should have 2 submodules (not examples)

		// Verify submodule data
		submodule1 := response[0].(map[string]interface{})
		assert.Equal(t, "submodule1", submodule1["path"])
		assert.Equal(t, "/modules/testns/testmod/testprov/1.0.0/submodule/submodule1", submodule1["href"])

		submodule2 := response[1].(map[string]interface{})
		assert.Equal(t, "submodule2", submodule2["path"])
		assert.Equal(t, "/modules/testns/testmod/testprov/1.0.0/submodule/submodule2", submodule2["href"])
	})

	t.Run("Examples endpoint returns database data", func(t *testing.T) {
		// Create request
		req := httptest.NewRequest("GET", "/v1/terrareg/modules/testns/testmod/testprov/1.0.0/examples", nil)

		// Setup chi router context
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("namespace", "testns")
		rctx.URLParams.Add("name", "testmod")
		rctx.URLParams.Add("provider", "testprov")
		rctx.URLParams.Add("version", "1.0.0")

		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		// Create response recorder
		w := httptest.NewRecorder()

		// Execute request
		server.Router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code)

		// Response should be a direct array
		var response []interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Len(t, response, 2) // Should have 2 examples

		// Verify example data
		example1 := response[0].(map[string]interface{})
		assert.Equal(t, "example1", example1["path"])
		assert.Equal(t, "/modules/testns/testmod/testprov/1.0.0/example/example1", example1["href"])

		example2 := response[1].(map[string]interface{})
		assert.Equal(t, "example2", example2["path"])
		assert.Equal(t, "/modules/testns/testmod/testprov/1.0.0/example/example2", example2["href"])
	})

	t.Run("Empty submodules returns empty array", func(t *testing.T) {
		// Create a module version without submodules
		namespace2 := testutils.CreateNamespace(t, db, "testns2")
		moduleProvider2 := testutils.CreateModuleProvider(t, db, namespace2.ID, "testmod2", "testprov2")
		moduleVersion2 := testutils.CreateModuleVersion(t, db, moduleProvider2.ID, "1.0.0")

		// Create request
		req := httptest.NewRequest("GET", "/v1/terrareg/modules/testns2/testmod2/testprov2/1.0.0/submodules", nil)

		// Setup chi router context
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("namespace", "testns2")
		rctx.URLParams.Add("name", "testmod2")
		rctx.URLParams.Add("provider", "testprov2")
		rctx.URLParams.Add("version", "1.0.0")

		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		// Create response recorder
		w := httptest.NewRecorder()

		// Execute request
		server.Router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code)

		// Response should be an empty array
		var response []interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Len(t, response, 0) // Empty array, not null
	})

	t.Run("Non-existent module returns 404", func(t *testing.T) {
		// Create request for non-existent module
		req := httptest.NewRequest("GET", "/v1/terrareg/modules/nonexistent/nonmod/nonprov/1.0.0/submodules", nil)

		// Setup chi router context
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("namespace", "nonexistent")
		rctx.URLParams.Add("name", "nonmod")
		rctx.URLParams.Add("provider", "nonprov")
		rctx.URLParams.Add("version", "1.0.0")

		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		// Create response recorder
		w := httptest.NewRecorder()

		// Execute request
		server.Router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestSubmodulesExamples_WithUnpublishedVersion(t *testing.T) {
	// Setup test database
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Setup minimal configs for this test
	domainCfg := &model.DomainConfig{
		AllowModuleHosting:  model.ModuleHostingModeAllow,
		TrustedNamespaces:   []string{},
		// Add other required fields with minimal defaults
		TrustedNamespaceLabel:     "Trusted",
		ContributedNamespaceLabel: "Contributed",
		VerifiedModuleLabel:       "Verified",
		AnalyticsTokenPhrase:      "",
		AnalyticsTokenDescription: "",
		ExampleAnalyticsToken:     "my-tf-application",
		DisableAnalytics:          false,
		UploadAPIKeysEnabled:      false,
		PublishAPIKeysEnabled:     false,
		DisableTerraregExclusiveLabels: false,
		AllowCustomGitURLModuleProvider: true,
		AllowCustomGitURLModuleVersion:  true,
		SecretKeySet:              false,
		OpenIDConnectEnabled:      false,
		OpenIDConnectLoginText:    "Login with OpenID",
		SAMLEnabled:               false,
		SAMLLoginText:             "Login with SAML",
		AdminLoginEnabled:         false,
		AdditionalModuleTabs:      []string{},
		AutoCreateNamespace:       true,
		AutoCreateModuleProvider:  true,
		DefaultUiDetailsView:      model.DefaultUiInputOutputViewTable,
		TerraformExampleVersionTemplate: "",
		TerraformExampleVersionTemplatePreMajor: "",
		ProviderSources:           make(map[string]model.ProviderSourceConfig),
	}

	infraCfg := &config.InfrastructureConfig{
		ListenPort:    5000,
		PublicURL:     "http://localhost:5000",
		DatabaseURL:   ":memory:",
		Debug:         true,
		DataDirectory: "./data",
		// Add minimal required infrastructure config
		ApplicationName:    "Terrareg",
		SessionCookieName:  "terrareg_session",
	}

	// Create config service
	versionReader := version.NewVersionReader()
	configService := service.NewConfigurationService(service.ConfigurationServiceOptions{}, versionReader)

	// Create container with test dependencies
	container, err := container.NewContainer(domainCfg, infraCfg, configService, testutils.TestLogger, db)
	require.NoError(t, err)

	// Create server
	server := httpServer.NewServer(container)

	// Create test data in database - UNPUBLISHED version
	namespace := testutils.CreateNamespace(t, db, "testns")
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "testmod", "testprov")
	moduleVersion := testutils.CreateModuleVersion(t, db, moduleProvider.ID, "1.0.0") // Unpublished by default

	// Create submodules for unpublished version
	testutils.CreateSubmodule(t, db, moduleVersion.ID, "submodule1", "Submodule One", "", nil)

	// Test that unpublished versions still return submodules (they should be accessible)
	req := httptest.NewRequest("GET", "/v1/terrareg/modules/testns/testmod/testprov/1.0.0/submodules", nil)

	// Setup chi router context
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("namespace", "testns")
	rctx.URLParams.Add("name", "testmod")
	rctx.URLParams.Add("provider", "testprov")
	rctx.URLParams.Add("version", "1.0.0")

	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	// Create response recorder
	w := httptest.NewRecorder()

	// Execute request
	server.Router.ServeHTTP(w, req)

	// Assert - should return 200 with submodules, even for unpublished versions
	assert.Equal(t, http.StatusOK, w.Code)

	var response []interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response, 1) // Should still return submodule for unpublished version
}