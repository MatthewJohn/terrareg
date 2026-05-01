package selenium

import (
	"context"
	cryptorand "crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	mathrand "math/rand"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	analyticsQuery "github.com/matthewjohn/terrareg/terrareg-go/internal/application/query/analytics"
	domainConfigService "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/config/service"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/container"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/logging"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/version"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/handler/terrareg"
)

// testDbMutex protects database file access across parallel tests
// Python tests run sequentially, but Go runs tests in parallel by default
// This mutex ensures only one test uses the database at a time
var testDbMutex sync.Mutex

// testCounter generates unique IDs for test database files
// Must be atomic to handle parallel subtests safely
var testCounter int64

// generateTestSigningKey generates a test RSA signing key and saves it to a file.
// This is required for Terraform OIDC tests.
func generateTestSigningKey(t *testing.T) string {
	// Generate RSA private key
	privateKey, err := rsa.GenerateKey(cryptorand.Reader, 2048)
	require.NoError(t, err, "Failed to generate RSA key")

	// Encode private key to PEM format
	privateKeyPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}
	privateKeyBytes := pem.EncodeToMemory(privateKeyPEM)

	// Create a temporary file for the signing key
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "signing_key.pem")

	err = os.WriteFile(keyPath, privateKeyBytes, 0600)
	require.NoError(t, err, "Failed to write signing key file")

	// Note: Go's testing framework will clean up tmpDir automatically
	// To inspect signing key on failure, check the temporary directory path

	return keyPath
}

// TestServer manages a test instance of the Terrareg server.
// This is the Go equivalent of Python's test.selenium.SeleniumTest server setup.
// Python reference: /app/test/selenium/__init__.py - SeleniumTest._setup_server()
type TestServer struct {
	t               *testing.T
	container       *container.Container
	httpServer      *http.Server
	port            int
	baseURL         string
	db              *sqldb.Database
	configOverrides map[string]string
	logger          logging.Logger
	serverCtx       context.Context
	serverCancel    context.CancelFunc
	serverWg        sync.WaitGroup
	originalWd      string                // Original working directory to restore on shutdown
	testDataSetup   func(*sqldb.Database) // Optional test data setup function
}

// TestServerOption is a function that configures the test server after container creation.
// This pattern allows tests to customize the server (e.g., mock repositories) without
// modifying the core server setup logic.
type TestServerOption func(*TestServer)

// NewTestServer creates and starts a new test Terrareg server.
// This is the Go equivalent of Python's setup_class method.
// Python reference: /app/test/selenium/__init__.py - SeleniumTest.setup_class()
func NewTestServer(t *testing.T, configOverrides map[string]string, opts ...TestServerOption) *TestServer {
	ts := &TestServer{
		t:               t,
		configOverrides: configOverrides,
		logger:          logging.NewTestLogger(t),
	}

	// Generate unique database file name for this test to ensure isolation
	// This prevents test pollution when tests run sequentially
	// Use atomic increment to handle parallel subtests safely
	counter := atomic.AddInt64(&testCounter, 1)
	dbFileName := fmt.Sprintf("temp-selenium-%d.db", counter)

	// Generate test signing key for Terraform OIDC
	signingKeyPath := generateTestSigningKey(t)
	if ts.configOverrides == nil {
		ts.configOverrides = make(map[string]string)
	}
	ts.configOverrides["TERRAFORM_OIDC_IDP_SIGNING_KEY_PATH"] = signingKeyPath

	// Override DATABASE_URL with unique database file
	// Use relative path (two slashes) for sqlite:// so it's created in current directory
	// NOT three slashes (sqlite:///) which would create in root directory
	ts.configOverrides["DATABASE_URL"] = fmt.Sprintf("sqlite://%s", dbFileName)

	ts.setup()

	// Apply options AFTER setup (when container exists)
	for _, opt := range opts {
		opt(ts)
	}

	// Note: Database file cleanup is now handled by the bootstrap function
	// which deletes old database files BEFORE creating the new database

	// Call test data setup AFTER setup but before returning
	// This is done here because testDataSetup needs the database to exist,
	// but it should happen before the server handles requests
	if ts.testDataSetup != nil {
		ts.testDataSetup(ts.db)
	}

	return ts
}

// setup initializes the test server following the bootstrap pattern from cmd/server/main.go
func (ts *TestServer) setup() {
	// Change to terrareg-go directory so relative paths work correctly
	// This ensures ./static finds the correct static files directory
	wd, err := os.Getwd()
	if err != nil {
		ts.t.Fatalf("Failed to get working directory: %v", err)
	}
	// Save original directory to restore later
	ts.originalWd = wd

	// Change to the directory containing this test file (terrareg-go/test/selenium)
	// Then go up two levels to reach terrareg-go root
	testDir, _ := os.Getwd()
	if filepath.Base(testDir) == "selenium" {
		// Already in test/selenium, go up two levels
		err = os.Chdir("../..")
	} else {
		// Try to find terrareg-go directory
		err = os.Chdir("/app/terrareg-go")
	}
	if err != nil {
		ts.t.Fatalf("Failed to chdir to terrareg-go: %v", err)
	}

	// Set default config values if not provided
	// Python reference: /app/test/selenium/__init__.py - _get_database_path() returns 'temp-selenium.db'
	defaults := map[string]string{
		"LISTEN_PORT":                          "5000", // Valid port (will be overridden to random port after bootstrap)
		"PUBLIC_URL":                           "http://127.0.0.1:5000",
		"DATABASE_URL":                         "sqlite:///temp-selenium.db", // File-based DB like Python tests
		"SECRET_KEY":                           "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
		"ADMIN_AUTHENTICATION_TOKEN":           "test-admin-token",
		"ALLOW_MODULE_HOSTING":                 "true",
		"DEBUG":                                "true",
		"SESSION_COOKIE_NAME":                  "terrareg_session",
		"SESSION_EXPIRY_MINS":                  "60",
		"ADMIN_SESSION_EXPIRY_MINS":            "60",
		"SESSION_REFRESH_MINS":                 "5",
		"TRUSTED_NAMESPACE_LABEL":              "Trusted",
		"CONTRIBUTED_NAMESPACE_LABEL":          "Contributed",
		"VERIFIED_MODULE_LABEL":                "Verified",
		"ALLOW_CUSTOM_GIT_URL_MODULE_PROVIDER": "true",
		"ALLOW_CUSTOM_GIT_URL_MODULE_VERSION":  "true",
		"AUTO_CREATE_NAMESPACE":                "true",
		"AUTO_CREATE_MODULE_PROVIDER":          "true",
		"DISABLE_ANALYTICS":                    "false",
		"AUTO_PUBLISH_MODULE_VERSIONS":         "true",
		"MODULE_VERSION_REINDEX_MODE":          "legacy",
		"PRODUCT":                              "terraform",
		"DEFAULT_TERRAFORM_VERSION":            "1.5.7",
		"MANAGE_TERRAFORM_RC_FILE":             "false",
		"MODULES_DIRECTORY":                    "modules",
		"EXAMPLES_DIRECTORY":                   "examples",
		"PROVIDER_SOURCES":                     "[]",
		"PROVIDER_CATEGORIES":                  `[{"id": 1, "name": "Example Category", "slug": "example-category", "user-selectable": true}]`,
		"GITHUB_URL":                           "https://github.com",
		"GITHUB_API_URL":                       "https://api.github.com",
		"GITHUB_LOGIN_TEXT":                    "Login with Github",
		"OPENID_CONNECT_LOGIN_TEXT":            "Login using OpenID Connect",
		"SAML2_LOGIN_TEXT":                     "Login using SAML",
		"INFRACOST_TLS_INSECURE_SKIP_VERIFY":   "false",
		"ALLOW_UNIDENTIFIED_DOWNLOADS":         "false",
		// Terraform OIDC settings (will be overridden with generated key)
		"TERRAFORM_OIDC_IDP_SUBJECT_ID_HASH_SALT": "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
		"TERRAFORM_OIDC_IDP_SESSION_EXPIRY":       "3600",
	}

	// Merge user overrides with defaults
	for k, v := range ts.configOverrides {
		defaults[k] = v
	}
	ts.configOverrides = defaults

	// Set environment variables (equivalent to Python's unittest.mock.patch)
	for k, v := range ts.configOverrides {
		os.Setenv(k, v)
	}

	// Bootstrap the application (following cmd/server/main.go pattern)
	ts.bootstrap()

	// Find an available port in 20000-21000 range (like Python tests)
	port := ts.findAvailablePort()
	ts.port = port
	ts.baseURL = fmt.Sprintf("http://127.0.0.1:%d", port)

	// Create HTTP server with the chi router from container.Server
	// Override the ListenPort in the container's InfraConfig
	ts.container.InfraConfig.ListenPort = port
	ts.container.InfraConfig.PublicURL = ts.baseURL

	// Get the router from the container's Server
	router := ts.container.Server.GetRouter()

	// Create our own http.Server to control the port
	ts.serverCtx, ts.serverCancel = context.WithCancel(context.Background())
	ts.httpServer = &http.Server{
		Addr:         fmt.Sprintf("127.0.0.1:%d", port),
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start server in background
	ts.serverWg.Add(1)
	go func() {
		defer ts.serverWg.Done()
		ts.t.Logf("Test server listening on %s", ts.baseURL)
		if err := ts.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			ts.t.Logf("Server error: %v", err)
		}
	}()

	// Wait for server to be ready
	ts.waitForServer()
}

// bootstrap bootstraps the application following the pattern in cmd/server/main.go
func (ts *TestServer) bootstrap() {
	// Lock mutex to prevent parallel tests from interfering with database
	// Python tests run sequentially, so this emulates that behavior
	testDbMutex.Lock()

	// Extract database file path from DATABASE_URL config and delete old database files
	// This must happen BEFORE creating the new database
	if dbURL, ok := ts.configOverrides["DATABASE_URL"]; ok {
		// Parse sqlite:// prefix to get the actual file path
		if len(dbURL) > 9 && dbURL[:9] == "sqlite://" {
			dbPath := dbURL[9:] // Remove "sqlite://" prefix
			os.Remove(dbPath)
			os.Remove(dbPath + "-wal")
			os.Remove(dbPath + "-shm")
		}
	}

	// Register cleanup function to unlock mutex after test completes
	ts.t.Cleanup(func() {
		testDbMutex.Unlock()
	})

	// Load configuration using the new configuration service
	versionReader := version.NewVersionReader()
	configService := domainConfigService.NewConfigurationService(
		domainConfigService.ConfigurationServiceOptions{},
		versionReader,
	)

	domainConfig, infraConfig, err := configService.LoadConfiguration()
	require.NoError(ts.t, err, "Failed to load configuration")

	// Initialize database
	ts.db, err = sqldb.NewDatabase(infraConfig.DatabaseURL, infraConfig.Debug)
	require.NoError(ts.t, err, "Failed to connect to database")

	// Run auto-migration for all models (from cmd/server/main.go)
	err = ts.autoMigrate()
	require.NoError(ts.t, err, "Failed to auto-migrate database")

	// Initialize dependency injection container with new configuration architecture
	ts.container, err = container.NewContainer(
		domainConfig,
		infraConfig,
		configService,
		ts.logger,
		ts.db,
	)
	require.NoError(ts.t, err, "Failed to create container")
}

// autoMigrate runs GORM auto-migration for all models
func (ts *TestServer) autoMigrate() error {
	return ts.db.AutoMigrate()
}

// findAvailablePort finds an available port in the range 20000-21000 (like Python tests).
// Python reference: /app/test/selenium/__init__.py - _setup_server() method
func (ts *TestServer) findAvailablePort() int {
	// Try ports in range 20000-21000 (matching Python's range)
	for i := 0; i < 100; i++ {
		port := 20000 + mathrand.Intn(1000)
		listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
		if err == nil {
			listener.Close()
			return port
		}
	}
	// If all ports in range are busy, use system-assigned port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(ts.t, err, "Failed to find any available port")
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()
	return port
}

// waitForServer waits for the server to be ready to accept connections.
func (ts *TestServer) waitForServer() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			ts.t.Fatal("Server did not start within timeout")
		case <-time.After(100 * time.Millisecond):
			client := &http.Client{Timeout: 1 * time.Second}
			resp, err := client.Get(ts.baseURL + "/")
			if err == nil {
				resp.Body.Close()
				return
			}
		}
	}
}

// GetURL returns the full URL for a given path.
// Python reference: /app/test/selenium/__init__.py - get_url()
func (ts *TestServer) GetURL(path string) string {
	return ts.baseURL + path
}

// GetPort returns the port the server is listening on.
func (ts *TestServer) GetPort() int {
	return ts.port
}

// GetContainer returns the DI container (useful for setting up test data).
func (ts *TestServer) GetContainer() *container.Container {
	return ts.container
}

// WithMockAnalytics replaces the AnalyticsRepo with a mock that returns the specified download count.
// This matches Python's behavior of mocking only get_total_downloads while keeping other queries real.
// Python reference: /app/test/selenium/test_homepage.py - mock.patch('get_total_downloads', return_value=2005)
func WithMockAnalytics(totalDownloads int) TestServerOption {
	return func(ts *TestServer) {
		// Create mock repository, wrapping the real one for non-mocked methods
		realRepo := ts.container.AnalyticsRepo
		ts.container.AnalyticsRepo = NewMockAnalyticsRepository(totalDownloads, realRepo)

		// Recreate GlobalStatsQuery with the mocked repo
		ts.container.GlobalStatsQuery = analyticsQuery.NewGlobalStatsQuery(
			ts.container.NamespaceRepo,
			ts.container.ModuleProviderRepo,
			ts.container.AnalyticsRepo, // Now uses the mock
		)

		// Recreate AnalyticsHandler with the mocked query
		newAnalyticsHandler := terrareg.NewAnalyticsHandler(
			ts.container.GlobalStatsQuery,
			ts.container.GlobalUsageStatsQuery,
			ts.container.GetDownloadSummaryQuery,
			ts.container.RecordModuleDownloadCmd,
			ts.container.GetMostRecentlyPublishedQuery,
			ts.container.GetMostDownloadedThisWeekQuery,
			ts.container.GetTokenVersionsQuery,
		)

		// Update both container and server AnalyticsHandler references
		ts.container.AnalyticsHandler = newAnalyticsHandler
		ts.container.Server.AnalyticsHandler = newAnalyticsHandler
	}
}

// GetDB returns the database connection (useful for setting up test data).
func (ts *TestServer) GetDB() *sqldb.Database {
	return ts.db
}

// Shutdown stops the test server and cleans up resources.
// Python reference: /app/test/selenium/__init__.py - _teardown_server()
func (ts *TestServer) Shutdown() {
	// Shutdown HTTP server
	if ts.httpServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		ts.httpServer.Shutdown(ctx)
	}

	// Cancel server context
	if ts.serverCancel != nil {
		ts.serverCancel()
	}

	// Wait for server goroutines to finish
	done := make(chan struct{})
	go func() {
		ts.serverWg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Server stopped successfully
	case <-time.After(10 * time.Second):
		ts.t.Log("Warning: Server did not stop within timeout")
	}

	// Close database
	if ts.db != nil {
		ts.db.Close()
	}

	// Restore original working directory
	if ts.originalWd != "" {
		os.Chdir(ts.originalWd)
	}
}
