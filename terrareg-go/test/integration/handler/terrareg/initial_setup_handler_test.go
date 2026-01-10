package terrareg_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	moduleRepo "github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb/module"
	urlService "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/url/service"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/application/query/setup"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/handler/terrareg"
	"github.com/matthewjohn/terrareg/terrareg-go/test/integration/testutils"
)

// TestInitialSetupHandler_HandleInitialSetup_NoNamespaces tests when no namespaces exist
func TestInitialSetupHandler_HandleInitialSetup_NoNamespaces(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create handler with empty database
	namespaceRepository := moduleRepo.NewNamespaceRepository(db.DB)
	moduleProviderRepository := moduleRepo.NewModuleProviderRepository(db.DB, namespaceRepository, nil)
	moduleVersionRepository := moduleRepo.NewModuleVersionRepository(db.DB)
	infraConfig := testutils.CreateTestInfraConfig(t)
	domainConfig := testutils.CreateTestDomainConfig(t)
	urlSvc := urlService.NewURLService(infraConfig)
	getInitialSetupQuery := setup.NewGetInitialSetupQuery(namespaceRepository, moduleProviderRepository, moduleVersionRepository, urlSvc, domainConfig)
	handler := terrareg.NewInitialSetupHandler(getInitialSetupQuery)

	req := httptest.NewRequest("GET", "/v1/terrareg/initial_setup", nil)
	w := httptest.NewRecorder()

	handler.HandleInitialSetup(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	response := testutils.GetJSONBody(t, w)

	assert.Equal(t, false, response["namespace_created"])
	assert.Equal(t, false, response["module_created"])
	assert.Equal(t, false, response["version_indexed"])
	assert.Equal(t, false, response["version_published"])
	assert.Equal(t, false, response["module_configured_with_git"])
}

// TestInitialSetupHandler_HandleInitialSetup_NamespaceOnly tests when only namespace exists
func TestInitialSetupHandler_HandleInitialSetup_NamespaceOnly(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create namespace only
	testutils.CreateNamespace(t, db, "test-namespace")

	namespaceRepository := moduleRepo.NewNamespaceRepository(db.DB)
	moduleProviderRepository := moduleRepo.NewModuleProviderRepository(db.DB, namespaceRepository, nil)
	moduleVersionRepository := moduleRepo.NewModuleVersionRepository(db.DB)
	domainConfig := testutils.CreateTestDomainConfig(t)
	infraConfig := testutils.CreateTestInfraConfig(t)
	urlSvc := urlService.NewURLService(infraConfig)
	getInitialSetupQuery := setup.NewGetInitialSetupQuery(namespaceRepository, moduleProviderRepository, moduleVersionRepository, urlSvc, domainConfig)
	handler := terrareg.NewInitialSetupHandler(getInitialSetupQuery)

	req := httptest.NewRequest("GET", "/v1/terrareg/initial_setup", nil)
	w := httptest.NewRecorder()

	handler.HandleInitialSetup(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	response := testutils.GetJSONBody(t, w)

	assert.Equal(t, true, response["namespace_created"])
	assert.Equal(t, false, response["module_created"])
	assert.Equal(t, false, response["version_indexed"])
	assert.Equal(t, false, response["version_published"])
	assert.Equal(t, false, response["module_configured_with_git"])
}

// TestInitialSetupHandler_HandleInitialSetup_WithModule tests when module exists
func TestInitialSetupHandler_HandleInitialSetup_WithModule(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create namespace and module provider
	namespace := testutils.CreateNamespace(t, db, "test-namespace")
	testutils.CreateModuleProvider(t, db, namespace.ID, "test-module", "aws")

	namespaceRepository := moduleRepo.NewNamespaceRepository(db.DB)
	moduleProviderRepository := moduleRepo.NewModuleProviderRepository(db.DB, namespaceRepository, nil)
	moduleVersionRepository := moduleRepo.NewModuleVersionRepository(db.DB)
	domainConfig := testutils.CreateTestDomainConfig(t)
	infraConfig := testutils.CreateTestInfraConfig(t)
	urlSvc := urlService.NewURLService(infraConfig)
	getInitialSetupQuery := setup.NewGetInitialSetupQuery(namespaceRepository, moduleProviderRepository, moduleVersionRepository, urlSvc, domainConfig)
	handler := terrareg.NewInitialSetupHandler(getInitialSetupQuery)

	req := httptest.NewRequest("GET", "/v1/terrareg/initial_setup", nil)
	w := httptest.NewRecorder()

	handler.HandleInitialSetup(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	response := testutils.GetJSONBody(t, w)

	assert.Equal(t, true, response["namespace_created"])
	assert.Equal(t, true, response["module_created"])
	assert.Equal(t, false, response["version_indexed"])
	assert.Equal(t, false, response["version_published"])
	assert.Equal(t, false, response["module_configured_with_git"])
	// Should have URLs set
	assert.Contains(t, response, "module_view_url")
	assert.Contains(t, response, "module_publish_endpoint")
}

// TestInitialSetupHandler_HandleInitialSetup_WithVersion tests when version exists
func TestInitialSetupHandler_HandleInitialSetup_WithVersion(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create namespace, module provider, and version
	namespace := testutils.CreateNamespace(t, db, "test-namespace")
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "test-module", "aws")
	testutils.CreateModuleVersion(t, db, moduleProvider.ID, "1.0.0")

	namespaceRepository := moduleRepo.NewNamespaceRepository(db.DB)
	moduleProviderRepository := moduleRepo.NewModuleProviderRepository(db.DB, namespaceRepository, nil)
	moduleVersionRepository := moduleRepo.NewModuleVersionRepository(db.DB)
	domainConfig := testutils.CreateTestDomainConfig(t)
	infraConfig := testutils.CreateTestInfraConfig(t)
	urlSvc := urlService.NewURLService(infraConfig)
	getInitialSetupQuery := setup.NewGetInitialSetupQuery(namespaceRepository, moduleProviderRepository, moduleVersionRepository, urlSvc, domainConfig)
	handler := terrareg.NewInitialSetupHandler(getInitialSetupQuery)

	req := httptest.NewRequest("GET", "/v1/terrareg/initial_setup", nil)
	w := httptest.NewRecorder()

	handler.HandleInitialSetup(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	response := testutils.GetJSONBody(t, w)

	assert.Equal(t, true, response["namespace_created"])
	assert.Equal(t, true, response["module_created"])
	assert.Equal(t, true, response["version_indexed"])
	assert.Equal(t, false, response["version_published"])
	assert.Equal(t, false, response["module_configured_with_git"])
}

// TestInitialSetupHandler_HandleInitialSetup_PublishedVersion tests when published version exists
func TestInitialSetupHandler_HandleInitialSetup_PublishedVersion(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create namespace, module provider, and published version
	namespace := testutils.CreateNamespace(t, db, "test-namespace")
	moduleProvider := testutils.CreateModuleProvider(t, db, namespace.ID, "test-module", "aws")
	testutils.CreatePublishedModuleVersion(t, db, moduleProvider.ID, "1.0.0")

	namespaceRepository := moduleRepo.NewNamespaceRepository(db.DB)
	moduleProviderRepository := moduleRepo.NewModuleProviderRepository(db.DB, namespaceRepository, nil)
	moduleVersionRepository := moduleRepo.NewModuleVersionRepository(db.DB)
	domainConfig := testutils.CreateTestDomainConfig(t)
	infraConfig := testutils.CreateTestInfraConfig(t)
	urlSvc := urlService.NewURLService(infraConfig)
	getInitialSetupQuery := setup.NewGetInitialSetupQuery(namespaceRepository, moduleProviderRepository, moduleVersionRepository, urlSvc, domainConfig)
	handler := terrareg.NewInitialSetupHandler(getInitialSetupQuery)

	req := httptest.NewRequest("GET", "/v1/terrareg/initial_setup", nil)
	w := httptest.NewRecorder()

	handler.HandleInitialSetup(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	response := testutils.GetJSONBody(t, w)

	assert.Equal(t, true, response["namespace_created"])
	assert.Equal(t, true, response["module_created"])
	assert.Equal(t, true, response["version_indexed"])
	assert.Equal(t, true, response["version_published"])
	assert.Equal(t, false, response["module_configured_with_git"])
}

// TestInitialSetupHandler_HandleInitialSetup_WithGit tests when module has git configuration
func TestInitialSetupHandler_HandleInitialSetup_WithGit(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create namespace and module provider with git configuration
	namespace := testutils.CreateNamespace(t, db, "test-namespace")
	gitCloneURL := "https://github.com/example/repo.git"
	moduleProvider := testutils.CreateModuleProviderWithGit(t, db, namespace.ID, "test-module", "aws", &gitCloneURL)

	testutils.CreatePublishedModuleVersion(t, db, moduleProvider.ID, "1.0.0")

	namespaceRepository := moduleRepo.NewNamespaceRepository(db.DB)
	moduleProviderRepository := moduleRepo.NewModuleProviderRepository(db.DB, namespaceRepository, nil)
	moduleVersionRepository := moduleRepo.NewModuleVersionRepository(db.DB)
	domainConfig := testutils.CreateTestDomainConfig(t)
	infraConfig := testutils.CreateTestInfraConfig(t)
	urlSvc := urlService.NewURLService(infraConfig)
	getInitialSetupQuery := setup.NewGetInitialSetupQuery(namespaceRepository, moduleProviderRepository, moduleVersionRepository, urlSvc, domainConfig)
	handler := terrareg.NewInitialSetupHandler(getInitialSetupQuery)

	req := httptest.NewRequest("GET", "/v1/terrareg/initial_setup", nil)
	w := httptest.NewRecorder()

	handler.HandleInitialSetup(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	response := testutils.GetJSONBody(t, w)

	assert.Equal(t, true, response["namespace_created"])
	assert.Equal(t, true, response["module_created"])
	assert.Equal(t, true, response["version_indexed"])
	assert.Equal(t, true, response["version_published"])
	assert.Equal(t, true, response["module_configured_with_git"])
}

// TestInitialSetupHandler_HandleInitialSetup_MethodNotAllowed tests that POST is rejected
func TestInitialSetupHandler_HandleInitialSetup_MethodNotAllowed(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	namespaceRepository := moduleRepo.NewNamespaceRepository(db.DB)
	moduleProviderRepository := moduleRepo.NewModuleProviderRepository(db.DB, namespaceRepository, nil)
	moduleVersionRepository := moduleRepo.NewModuleVersionRepository(db.DB)
	domainConfig := testutils.CreateTestDomainConfig(t)
	infraConfig := testutils.CreateTestInfraConfig(t)
	urlSvc := urlService.NewURLService(infraConfig)
	getInitialSetupQuery := setup.NewGetInitialSetupQuery(namespaceRepository, moduleProviderRepository, moduleVersionRepository, urlSvc, domainConfig)
	handler := terrareg.NewInitialSetupHandler(getInitialSetupQuery)

	req := httptest.NewRequest("POST", "/v1/terrareg/initial_setup", nil)
	w := httptest.NewRecorder()

	handler.HandleInitialSetup(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

// TestInitialSetupHandler_HandleInitialSetup_CompleteSetup tests fully configured setup
func TestInitialSetupHandler_HandleInitialSetup_CompleteSetup(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Create complete setup with git configuration
	namespace := testutils.CreateNamespace(t, db, "test-namespace")
	gitCloneURL := "https://github.com/example/repo.git"
	moduleProvider := testutils.CreateModuleProviderWithGit(t, db, namespace.ID, "test-module", "aws", &gitCloneURL)
	testutils.CreatePublishedModuleVersion(t, db, moduleProvider.ID, "1.0.0")

	namespaceRepository := moduleRepo.NewNamespaceRepository(db.DB)
	moduleProviderRepository := moduleRepo.NewModuleProviderRepository(db.DB, namespaceRepository, nil)
	moduleVersionRepository := moduleRepo.NewModuleVersionRepository(db.DB)
	domainConfig := testutils.CreateTestDomainConfig(t)
	infraConfig := testutils.CreateTestInfraConfig(t)
	urlSvc := urlService.NewURLService(infraConfig)
	getInitialSetupQuery := setup.NewGetInitialSetupQuery(namespaceRepository, moduleProviderRepository, moduleVersionRepository, urlSvc, domainConfig)
	handler := terrareg.NewInitialSetupHandler(getInitialSetupQuery)

	req := httptest.NewRequest("GET", "/v1/terrareg/initial_setup", nil)
	w := httptest.NewRecorder()

	handler.HandleInitialSetup(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	response := testutils.GetJSONBody(t, w)

	// All flags should be true
	assert.Equal(t, true, response["namespace_created"])
	assert.Equal(t, true, response["module_created"])
	assert.Equal(t, true, response["version_indexed"])
	assert.Equal(t, true, response["version_published"])
	assert.Equal(t, true, response["module_configured_with_git"])

	// All URLs should be set
	assert.Contains(t, response, "module_view_url")
	assert.Contains(t, response, "module_upload_endpoint")
	assert.Contains(t, response, "module_publish_endpoint")
}
