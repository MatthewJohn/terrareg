package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/config"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/config/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/container"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/handler/terrareg"
)

// AllowModuleHostingIntegrationTestSuite tests ALLOW_MODULE_HOSTING functionality end-to-end
type AllowModuleHostingIntegrationTestSuite struct {
	suite.Suite
	server   *http.Server
	db       *sqldb.Database
	container *container.Container
}

func (suite *AllowModuleHostingIntegrationTestSuite) SetupSuite() {
	// Use in-memory database for testing
	db, err := sqldb.NewDatabase(":memory:", true)
	suite.Require().NoError(err)
	suite.db = db

	// Create test config
	cfg := &config.Config{
		ListenPort:                5000,
		PublicURL:                "http://localhost:5000",
		DatabaseURL:               ":memory:",
		Debug:                    true,
		AllowModuleHosting:       model.ModuleHostingModeAllow,
		TrustedNamespaces:        []string{},
	}

	// Create logger
	logger := zerolog.Nop()

	// Create container with test dependencies
	suite.container, err = container.NewContainer(cfg, logger, db)
	suite.Require().NoError(err)

	// Create server
	suite.server = http.NewServer(suite.container)
}

func (suite *AllowModuleHostingIntegrationTestSuite) TearDownSuite() {
	if suite.db != nil {
		suite.db.Close()
	}
}

func (suite *AllowModuleHostingIntegrationTestSuite) TestUploadEndpoint_AllowModuleHostingModes() {
	tests := []struct {
		name               string
		allowModuleHosting model.ModuleHostingMode
		expectedStatus     int
		setupFunc          func(cfg *config.Config)
	}{
		{
			name:               "ALLOW mode - upload should succeed",
			allowModuleHosting: model.ModuleHostingModeAllow,
			expectedStatus:     http.StatusOK,
			setupFunc: func(cfg *config.Config) {
				cfg.AllowModuleHosting = model.ModuleHostingModeAllow
			},
		},
		{
			name:               "ENFORCE mode - upload should succeed",
			allowModuleHosting: model.ModuleHostingModeEnforce,
			expectedStatus:     http.StatusOK,
			setupFunc: func(cfg *config.Config) {
				cfg.AllowModuleHosting = model.ModuleHostingModeEnforce
			},
		},
		{
			name:               "DISALLOW mode - upload should be blocked",
			allowModuleHosting: model.ModuleHostingModeDisallow,
			expectedStatus:     http.StatusBadRequest,
			setupFunc: func(cfg *config.Config) {
				cfg.AllowModuleHosting = model.ModuleHostingModeDisallow
			},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			// Update config in container
			tt.setupFunc(suite.container.Config)

			// Create test module provider first (required for upload)
			namespace := "testns"
			module := "testmod"
			provider := "testprov"

			// Create namespace
			createNamespaceReq := httptest.NewRequest("POST", "/v1/terrareg/namespaces", bytes.NewBufferString(`{"name":"testns"}`))
			createNamespaceReq.Header.Set("Content-Type", "application/json")
			createNamespaceResp := httptest.NewRecorder()
			suite.server.Router.ServeHTTP(createNamespaceResp, createNamespaceReq)

			// Create module provider
			createModuleReq := httptest.NewRequest("POST", "/v1/terrareg/modules/testns/testmod/testprov/create", nil)
			createModuleResp := httptest.NewRecorder()
			suite.server.Router.ServeHTTP(createModuleResp, createModuleReq)
			suite.Equal(http.StatusCreated, createModuleResp.Code, "Failed to create module provider")

			// Create multipart form for upload
			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)
			part, err := writer.CreateFormFile("file", "test-module.zip")
			suite.Require().NoError(err)
			part.Write([]byte("dummy zip content"))
			writer.Close()

			// Create upload request
			uploadReq := httptest.NewRequest("POST", "/v1/terrareg/modules/testns/testmod/testprov/1.0.0/upload", body)
			uploadReq.Header.Set("Content-Type", writer.FormDataContentType())
			uploadResp := httptest.NewRecorder()

			// Execute request
			suite.server.Router.ServeHTTP(uploadResp, uploadReq)

			// Check response
			suite.Equal(tt.expectedStatus, uploadResp.Code)

			if tt.expectedStatus == http.StatusBadRequest {
				var response map[string]interface{}
				err = json.Unmarshal(uploadResp.Body.Bytes(), &response)
				suite.NoError(err)
				suite.Contains(response["error"], "Module upload is disabled.")
			} else if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err = json.Unmarshal(uploadResp.Body.Bytes(), &response)
				suite.NoError(err)
				suite.Contains(response["message"], "uploaded successfully")
			}
		})
	}
}

func (suite *AllowModuleHostingIntegrationTestSuite) TestSourceDownloadEndpoint_AllowModuleHostingModes() {
	tests := []struct {
		name               string
		allowModuleHosting model.ModuleHostingMode
		expectedStatus     int
		setupFunc          func(cfg *config.Config)
	}{
		{
			name:               "ALLOW mode - download should proceed",
			allowModuleHosting: model.ModuleHostingModeAllow,
			expectedStatus:     http.StatusNotImplemented, // Our implementation returns 501 for now
			setupFunc: func(cfg *config.Config) {
				cfg.AllowModuleHosting = model.ModuleHostingModeAllow
			},
		},
		{
			name:               "ENFORCE mode - download should proceed",
			allowModuleHosting: model.ModuleHostingModeEnforce,
			expectedStatus:     http.StatusNotImplemented,
			setupFunc: func(cfg *config.Config) {
				cfg.AllowModuleHosting = model.ModuleHostingModeEnforce
			},
		},
		{
			name:               "DISALLOW mode - download should be blocked",
			allowModuleHosting: model.ModuleHostingModeDisallow,
			expectedStatus:     http.StatusInternalServerError,
			setupFunc: func(cfg *config.Config) {
				cfg.AllowModuleHosting = model.ModuleHostingModeDisallow
			},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			// Update config in container
			tt.setupFunc(suite.container.Config)

			// Create source download request
			downloadReq := httptest.NewRequest("GET", "/v1/terrareg/modules/testns/testmod/testprov/1.0.0/source.zip", nil)
			downloadResp := httptest.NewRecorder()

			// Execute request
			suite.server.Router.ServeHTTP(downloadResp, downloadReq)

			// Check response
			suite.Equal(tt.expectedStatus, downloadResp.Code)

			if tt.expectedStatus == http.StatusInternalServerError {
				var response map[string]interface{}
				err := json.Unmarshal(downloadResp.Body.Bytes(), &response)
				suite.NoError(err)
				suite.Contains(response["message"], "Module hosting is disabled")
			} else if tt.expectedStatus == http.StatusNotImplemented {
				var response map[string]interface{}
				err := json.Unmarshal(downloadResp.Body.Bytes(), &response)
				suite.NoError(err)
				suite.Contains(response["message"], "Source download not yet implemented")
			}
		})
	}
}

func (suite *AllowModuleHostingIntegrationTestSuite) TestTerraformDownloadEndpoint_AllowModuleHostingModes() {
	tests := []struct {
		name               string
		allowModuleHosting model.ModuleHostingMode
		expectedStatus     int
		setupFunc          func(cfg *config.Config)
	}{
		{
			name:               "ALLOW mode - should proceed",
			allowModuleHosting: model.ModuleHostingModeAllow,
			expectedStatus:     http.StatusNotFound, // Will return 404 because module doesn't exist, but that's expected
			setupFunc: func(cfg *config.Config) {
				cfg.AllowModuleHosting = model.ModuleHostingModeAllow
			},
		},
		{
			name:               "ENFORCE mode - should proceed",
			allowModuleHosting: model.ModuleHostingModeEnforce,
			expectedStatus:     http.StatusNotFound,
			setupFunc: func(cfg *config.Config) {
				cfg.AllowModuleHosting = model.ModuleHostingModeEnforce
			},
		},
		{
			name:               "DISALLOW mode - should be blocked",
			allowModuleHosting: model.ModuleHostingModeDisallow,
			expectedStatus:     http.StatusInternalServerError,
			setupFunc: func(cfg *config.Config) {
				cfg.AllowModuleHosting = model.ModuleHostingModeDisallow
			},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			// Update config in container
			tt.setupFunc(suite.container.Config)

			// Create Terraform download request
			downloadReq := httptest.NewRequest("GET", "/v1/modules/testns/testmod/testprov/1.0.0/download", nil)
			downloadResp := httptest.NewRecorder()

			// Execute request
			suite.server.Router.ServeHTTP(downloadResp, downloadReq)

			// Check response
			suite.Equal(tt.expectedStatus, downloadResp.Code)

			if tt.expectedStatus == http.StatusInternalServerError {
				body := downloadResp.Body.String()
				suite.Contains(body, "Module hosting is disabled")
			}
		})
	}
}

func TestAllowModuleHostingIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(AllowModuleHostingIntegrationTestSuite))
}