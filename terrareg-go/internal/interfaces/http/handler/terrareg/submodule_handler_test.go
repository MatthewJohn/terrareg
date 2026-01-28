package terrareg_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/application/query/module"
	modulemodel "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/repository"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared/types"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/url/service"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/config"
	terrareg "github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/handler/terrareg"
)

// MockModuleProviderRepository is a minimal mock for testing
type MockModuleProviderRepository struct {
	mock.Mock
}

func (m *MockModuleProviderRepository) FindByNamespaceModuleProvider(ctx context.Context, namespace types.NamespaceName, moduleName types.ModuleName, provider types.ModuleProviderName) (*modulemodel.ModuleProvider, error) {
	args := m.Called(ctx, namespace, moduleName, provider)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*modulemodel.ModuleProvider), args.Error(1)
}

// Implement other required interface methods minimally
func (m *MockModuleProviderRepository) Save(ctx context.Context, moduleProvider *modulemodel.ModuleProvider) error {
	return nil
}
func (m *MockModuleProviderRepository) Create(ctx context.Context, moduleProvider *modulemodel.ModuleProvider) error {
	return nil
}
func (m *MockModuleProviderRepository) Delete(ctx context.Context, id int) error { return nil }
func (m *MockModuleProviderRepository) FindByID(ctx context.Context, id int) (*modulemodel.ModuleProvider, error) {
	return nil, nil
}
func (m *MockModuleProviderRepository) FindByNamespace(ctx context.Context, namespace types.NamespaceName) ([]*modulemodel.ModuleProvider, error) {
	return nil, nil
}
func (m *MockModuleProviderRepository) Search(ctx context.Context, query repository.ModuleSearchQuery) (*repository.ModuleSearchResult, error) {
	return nil, nil
}
func (m *MockModuleProviderRepository) Exists(ctx context.Context, namespace types.NamespaceName, moduleName types.ModuleName, provider types.ModuleProviderName) (bool, error) {
	return false, nil
}

// MockModuleVersionRepository is a minimal mock for testing
type MockModuleVersionRepository struct {
	mock.Mock
}

func (m *MockModuleVersionRepository) FindByModuleProviderAndVersion(ctx context.Context, moduleProviderID int, version types.ModuleVersion) (*modulemodel.ModuleVersion, error) {
	args := m.Called(ctx, moduleProviderID, version)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*modulemodel.ModuleVersion), args.Error(1)
}

// Implement other required interface methods minimally
func (m *MockModuleVersionRepository) FindByModuleProvider(ctx context.Context, moduleProviderID int, includeBeta, includeUnpublished bool) ([]*modulemodel.ModuleVersion, error) {
	return nil, nil
}
func (m *MockModuleVersionRepository) Save(ctx context.Context, moduleVersion *modulemodel.ModuleVersion) (*modulemodel.ModuleVersion, error) {
	return nil, nil
}
func (m *MockModuleVersionRepository) FindByID(ctx context.Context, id int) (*modulemodel.ModuleVersion, error) {
	return nil, nil
}
func (m *MockModuleVersionRepository) Delete(ctx context.Context, id int) error { return nil }
func (m *MockModuleVersionRepository) Exists(ctx context.Context, moduleProviderID int, version types.ModuleVersion) (bool, error) {
	return false, nil
}
func (m *MockModuleVersionRepository) UpdateModuleDetailsID(ctx context.Context, moduleVersionID int, moduleDetailsID int) error {
	return nil
}

func TestSubmoduleHandler_HandleSubmoduleDetails(t *testing.T) {
	tests := []struct {
		name                 string
		method               string
		url                  string
		expectedStatus       int
		setupMocks           func(*MockModuleProviderRepository, *MockModuleVersionRepository)
		expectedBodyContains string
	}{
		{
			name:           "module provider not found",
			method:         "GET",
			url:            "/modules/test/mod/provider/1.0.0/submodules/details/submod",
			expectedStatus: http.StatusNotFound,
			setupMocks: func(mockProviderRepo *MockModuleProviderRepository, mockVersionRepo *MockModuleVersionRepository) {
				// Return (nil, nil) - query will convert to ErrModuleProviderNotFound
				mockProviderRepo.On("FindByNamespaceModuleProvider", mock.Anything, types.NamespaceName("test"), types.ModuleName("mod"), types.ModuleProviderName("provider")).
					Return(nil, nil).
					Once()
			},
			expectedBodyContains: "module provider not found",
		},
		{
			name:           "successful response",
			method:         "GET",
			url:            "/modules/test/mod/provider/1.0.0/submodules/details/submod",
			expectedStatus: http.StatusOK,
			setupMocks: func(mockProviderRepo *MockModuleProviderRepository, mockVersionRepo *MockModuleVersionRepository) {
				// Create a minimal module provider mock
				moduleProvider := &modulemodel.ModuleProvider{}
				mockProviderRepo.On("FindByNamespaceModuleProvider", mock.Anything, types.NamespaceName("test"), types.ModuleName("mod"), types.ModuleProviderName("provider")).
					Return(moduleProvider, nil).
					Once()

				// Create a published module version with a submodule
				moduleVersion, _ := modulemodel.NewModuleVersion("1.0.0", nil, false)
				moduleVersion.Publish()
				// Add submodule with nil details (will return empty specs)
				submodule := modulemodel.NewSubmodule("submod", nil, nil, nil)
				moduleVersion.AddSubmodule(submodule)
				mockVersionRepo.On("FindByModuleProviderAndVersion", mock.Anything, mock.AnythingOfType("int"), "1.0.0").
					Return(moduleVersion, nil).
					Once()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockProviderRepo := &MockModuleProviderRepository{}
			mockVersionRepo := &MockModuleVersionRepository{}
			tt.setupMocks(mockProviderRepo, mockVersionRepo)
			urlService := service.NewURLService(&config.InfrastructureConfig{
				PublicURL: "http://localhost:5000",
			})

			detailsQuery := module.NewGetSubmoduleDetailsQuery(mockProviderRepo, mockVersionRepo, urlService)
			readmeQuery := module.NewGetSubmoduleReadmeHTMLQuery(mockProviderRepo, mockVersionRepo)
			handler := terrareg.NewSubmoduleHandler(detailsQuery, readmeQuery)

			// Create request
			req := httptest.NewRequest(tt.method, tt.url, nil)

			// Set up chi context
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("namespace", "test")
			rctx.URLParams.Add("name", "mod")
			rctx.URLParams.Add("provider", "provider")
			rctx.URLParams.Add("version", "1.0.0")
			rctx.URLParams.Add("*", "submod")
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			// Create response recorder
			w := httptest.NewRecorder()

			// Act
			handler.HandleSubmoduleDetails(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedBodyContains != "" {
				assert.Contains(t, w.Body.String(), tt.expectedBodyContains)
			}

			if tt.name == "successful response" {
				assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
			}

			mockProviderRepo.AssertExpectations(t)
			mockVersionRepo.AssertExpectations(t)
		})
	}
}

func TestSubmoduleHandler_HandleSubmoduleReadmeHTML(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		url            string
		expectedStatus int
		setupMocks     func(*MockModuleProviderRepository, *MockModuleVersionRepository)
		checkHeaders   bool
	}{
		{
			name:           "successful response",
			method:         "GET",
			url:            "/modules/test/mod/provider/1.0.0/submodules/readme_html/submod",
			expectedStatus: http.StatusOK,
			setupMocks: func(mockProviderRepo *MockModuleProviderRepository, mockVersionRepo *MockModuleVersionRepository) {
				moduleProvider := &modulemodel.ModuleProvider{}
				mockProviderRepo.On("FindByNamespaceModuleProvider", mock.Anything, "test", "mod", "provider").
					Return(moduleProvider, nil).
					Once()
				// Create a published module version with a submodule
				moduleVersion, _ := modulemodel.NewModuleVersion("1.0.0", nil, false)
				moduleVersion.Publish()
				// Add submodule with nil details (will return empty specs)
				submodule := modulemodel.NewSubmodule("submod", nil, nil, nil)
				moduleVersion.AddSubmodule(submodule)
				mockVersionRepo.On("FindByModuleProviderAndVersion", mock.Anything, mock.AnythingOfType("int"), "1.0.0").
					Return(moduleVersion, nil).
					Once()
			},
			checkHeaders: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockProviderRepo := &MockModuleProviderRepository{}
			mockVersionRepo := &MockModuleVersionRepository{}
			tt.setupMocks(mockProviderRepo, mockVersionRepo)
			urlService := service.NewURLService(&config.InfrastructureConfig{
				PublicURL: "http://localhost:5000",
			})

			detailsQuery := module.NewGetSubmoduleDetailsQuery(mockProviderRepo, mockVersionRepo, urlService)
			readmeQuery := module.NewGetSubmoduleReadmeHTMLQuery(mockProviderRepo, mockVersionRepo)
			handler := terrareg.NewSubmoduleHandler(detailsQuery, readmeQuery)

			// Create request
			req := httptest.NewRequest(tt.method, tt.url, nil)

			// Set up chi context
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("namespace", "test")
			rctx.URLParams.Add("name", "mod")
			rctx.URLParams.Add("provider", "provider")
			rctx.URLParams.Add("version", "1.0.0")
			rctx.URLParams.Add("*", "submod")
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			// Create response recorder
			w := httptest.NewRecorder()

			// Act
			handler.HandleSubmoduleReadmeHTML(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.checkHeaders {
				assert.Equal(t, "text/html; charset=utf-8", w.Header().Get("Content-Type"))
			}

			mockProviderRepo.AssertExpectations(t)
			mockVersionRepo.AssertExpectations(t)
		})
	}
}
