package unit

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/config"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/config/model"
	moduleCmd "github.com/matthewjohn/terrareg/terrareg-go/internal/application/command/module"
	moduleQuery "github.com/matthewjohn/terrareg/terrareg-go/internal/application/query/module"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/handler/terrareg"
	httputils "github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/utils"
)

// MockUploadModuleVersionCommand is a mock implementation of UploadModuleVersionCommand
type MockUploadModuleVersionCommand struct {
	mock.Mock
}

func (m *MockUploadModuleVersionCommand) Execute(ctx context.Context, req moduleCmd.UploadModuleVersionRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func TestModuleVersionUpload_AllowModuleHosting(t *testing.T) {
	tests := []struct {
		name                 string
		allowModuleHosting   model.ModuleHostingMode
		expectedStatusCode   int
		expectedErrorMessage string
		expectCommandCall    bool
	}{
		{
			name:                 "Allow mode - upload should succeed",
			allowModuleHosting:   model.ModuleHostingModeAllow,
			expectedStatusCode:   http.StatusOK,
			expectCommandCall:    true,
		},
		{
			name:                 "Enforce mode - upload should succeed",
			allowModuleHosting:   model.ModuleHostingModeEnforce,
			expectedStatusCode:   http.StatusOK,
			expectCommandCall:    true,
		},
		{
			name:                 "Disallow mode - upload should be blocked",
			allowModuleHosting:   model.ModuleHostingModeDisallow,
			expectedStatusCode:   http.StatusBadRequest,
			expectedErrorMessage: "Module upload is disabled.",
			expectCommandCall:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mocks
			mockUploadCmd := new(MockUploadModuleVersionCommand)

			// Create config with the specific ALLOW_MODULE_HOSTING mode
			cfg := &config.Config{
				AllowModuleHosting: tt.allowModuleHosting,
			}

			// Create handler
			handler := terrareg.NewModuleHandler(
				nil, // listModulesQuery
				nil, // searchModulesQuery
				nil, // getModuleProviderQuery
				nil, // listModuleProvidersQuery
				nil, // getModuleVersionQuery
				nil, // getModuleDownloadQuery
				nil, // getModuleProviderSettingsQuery
				nil, // getSubmodulesQuery
				nil, // getExamplesQuery
				nil, // createModuleProviderCmd
				nil, // publishModuleVersionCmd
				nil, // updateModuleProviderSettingsCmd
				nil, // deleteModuleProviderCmd
				mockUploadCmd,
				nil, // importModuleVersionCmd
				cfg,
			)

			// Create test request with multipart form data
			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)
			_ = writer.WriteField("test", "field")
			writer.Close()

			req := httptest.NewRequest("POST", "/v1/terrareg/modules/testns/testmod/testprov/1.0.0/upload", body)
			req.Header.Set("Content-Type", writer.FormDataContentType())

			// Mock chi URL parameters
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("namespace", "testns")
			rctx.URLParams.Add("name", "testmod")
			rctx.URLParams.Add("provider", "testprov")
			rctx.URLParams.Add("version", "1.0.0")
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			// Create response recorder
			w := httptest.NewRecorder()

			// Set up mock expectations
			if tt.expectCommandCall {
				mockUploadCmd.On("Execute", mock.Anything, mock.AnythingOfType("module.UploadModuleVersionRequest")).Return(nil).Once()
			}

			// Execute handler
			handler.HandleModuleVersionUpload(w, req)

			// Check response
			assert.Equal(t, tt.expectedStatusCode, w.Code)

			if tt.expectedErrorMessage != "" {
				var response map[string]interface{}
				err := httputils.DecodeJSON(w.Body, &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], tt.expectedErrorMessage)
			}

			// Verify mock expectations
			mockUploadCmd.AssertExpectations(t)
		})
	}
}

func TestModuleDownload_AllowModuleHosting(t *testing.T) {
	tests := []struct {
		name                 string
		allowModuleHosting   model.ModuleHostingMode
		expectedStatusCode   int
		expectedErrorMessage string
		expectQueryCall      bool
	}{
		{
			name:               "Allow mode - download should proceed",
			allowModuleHosting: model.ModuleHostingModeAllow,
			expectQueryCall:    true,
		},
		{
			name:               "Enforce mode - download should proceed",
			allowModuleHosting: model.ModuleHostingModeEnforce,
			expectQueryCall:    true,
		},
		{
			name:                 "Disallow mode - download should be blocked",
			allowModuleHosting:   model.ModuleHostingModeDisallow,
			expectedStatusCode:   http.StatusInternalServerError,
			expectedErrorMessage: "Module hosting is disabled",
			expectQueryCall:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mocks
			mockDownloadQuery := new(MockGetModuleDownloadQuery)

			// Create config with the specific ALLOW_MODULE_HOSTING mode
			cfg := &config.Config{
				AllowModuleHosting: tt.allowModuleHosting,
			}

			// Create handler
			handler := terrareg.NewModuleHandler(
				nil, // listModulesQuery
				nil, // searchModulesQuery
				nil, // getModuleProviderQuery
				nil, // listModuleProvidersQuery
				nil, // getModuleVersionQuery
				mockDownloadQuery,
				nil, // getModuleProviderSettingsQuery
				nil, // getSubmodulesQuery
				nil, // getExamplesQuery
				nil, // createModuleProviderCmd
				nil, // publishModuleVersionCmd
				nil, // updateModuleProviderSettingsCmd
				nil, // deleteModuleProviderCmd
				nil, // uploadModuleVersionCmd
				nil, // importModuleVersionCmd
				cfg,
			)

			// Create test request
			req := httptest.NewRequest("GET", "/v1/modules/testns/testmod/testprov/1.0.0/download", nil)

			// Mock chi URL parameters
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("namespace", "testns")
			rctx.URLParams.Add("name", "testmod")
			rctx.URLParams.Add("provider", "testprov")
			rctx.URLParams.Add("version", "1.0.0")
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			// Create response recorder
			w := httptest.NewRecorder()

			// Set up mock expectations
			if tt.expectQueryCall {
				// For this test, we'll have the query return an error since we haven't set up the full mock
				// The important thing is that the query gets called when hosting is allowed
				mockDownloadQuery.On("Execute", mock.Anything, "testns", "testmod", "testprov", "1.0.0").Return(nil, assert.AnError).Once()
			}

			// Execute handler
			handler.HandleModuleDownload(w, req)

			// Check response for disallowed case
			if tt.expectedErrorMessage != "" {
				assert.Equal(t, tt.expectedStatusCode, w.Code)
				body := w.Body.String()
				assert.Contains(t, body, tt.expectedErrorMessage)
			}

			// Verify mock expectations
			mockDownloadQuery.AssertExpectations(t)
		})
	}
}

// MockGetModuleDownloadQuery is a mock implementation of GetModuleDownloadQuery
type MockGetModuleDownloadQuery struct {
	mock.Mock
}

func (m *MockGetModuleDownloadQuery) Execute(ctx context.Context, namespace, name, provider, version string) (*moduleQuery.ModuleDownloadInfo, error) {
	args := m.Called(ctx, namespace, name, provider, version)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*moduleQuery.ModuleDownloadInfo), args.Error(1)
}