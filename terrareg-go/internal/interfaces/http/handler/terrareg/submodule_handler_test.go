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
)

// MockGetSubmoduleDetailsQuery is a mock for GetSubmoduleDetailsQuery
type MockGetSubmoduleDetailsQuery struct {
	mock.Mock
}

func (m *MockGetSubmoduleDetailsQuery) Execute(ctx context.Context, namespace, moduleName, provider, version, submodulePath string) (*module.SubmoduleDetails, error) {
	args := m.Called(ctx, namespace, moduleName, provider, version, submodulePath)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*module.SubmoduleDetails), args.Error(1)
}

// MockGetSubmoduleReadmeHTMLQuery is a mock for GetSubmoduleReadmeHTMLQuery
type MockGetSubmoduleReadmeHTMLQuery struct {
	mock.Mock
}

func (m *MockGetSubmoduleReadmeHTMLQuery) Execute(ctx context.Context, namespace, moduleName, provider, version, submodulePath string) (string, error) {
	args := m.Called(ctx, namespace, moduleName, provider, version, submodulePath)
	if args.String(0) == "" && args.Error(1) != nil {
		return "", args.Error(1)
	}
	return args.String(0), args.Error(1)
}

func TestSubmoduleHandler_HandleSubmoduleDetails(t *testing.T) {
	tests := []struct {
		name               string
		method             string
		url                string
		expectedStatus     int
		setupMocks         func(*MockGetSubmoduleDetailsQuery, *MockGetSubmoduleReadmeHTMLQuery)
		expectedBodyContains string
	}{
		{
			name:           "invalid method",
			method:         "POST",
			url:            "/modules/test/mod/provider/1.0.0/submodules/details/submod",
			expectedStatus: http.StatusMethodNotAllowed,
			setupMocks:     func(*MockGetSubmoduleDetailsQuery, *MockGetSubmoduleReadmeHTMLQuery) {},
		},
		{
			name:           "missing namespace parameter",
			method:         "GET",
			url:            "/modules//mod/provider/1.0.0/submodules/details/submod",
			expectedStatus: http.StatusBadRequest,
			setupMocks:     func(*MockGetSubmoduleDetailsQuery, *MockGetSubmoduleReadmeHTMLQuery) {},
			expectedBodyContains: "Missing required path parameters",
		},
		{
			name:           "missing module parameter",
			method:         "GET",
			url:            "/modules/test//provider/1.0.0/submodules/details/submod",
			expectedStatus: http.StatusBadRequest,
			setupMocks:     func(*MockGetSubmoduleDetailsQuery, *MockGetSubmoduleReadmeHTMLQuery) {},
			expectedBodyContains: "Missing required path parameters",
		},
		{
			name:           "missing provider parameter",
			method:         "GET",
			url:            "/modules/test/mod//1.0.0/submodules/details/submod",
			expectedStatus: http.StatusBadRequest,
			setupMocks:     func(*MockGetSubmoduleDetailsQuery, *MockGetSubmoduleReadmeHTMLQuery) {},
			expectedBodyContains: "Missing required path parameters",
		},
		{
			name:           "missing version parameter",
			method:         "GET",
			url:            "/modules/test/mod/provider//submodules/details/submod",
			expectedStatus: http.StatusBadRequest,
			setupMocks:     func(*MockGetSubmoduleDetailsQuery, *MockGetSubmoduleReadmeHTMLQuery) {},
			expectedBodyContains: "Missing required path parameters",
		},
		{
			name:           "missing submodule parameter",
			method:         "GET",
			url:            "/modules/test/mod/provider/1.0.0/submodules/details/",
			expectedStatus: http.StatusBadRequest,
			setupMocks:     func(*MockGetSubmoduleDetailsQuery, *MockGetSubmoduleReadmeHTMLQuery) {},
			expectedBodyContains: "Missing required path parameters",
		},
		{
			name:           "module provider not found",
			method:         "GET",
			url:            "/modules/test/mod/provider/1.0.0/submodules/details/submod",
			expectedStatus: http.StatusNotFound,
			setupMocks: func(mockDetails *MockGetSubmoduleDetailsQuery, mockReadme *MockGetSubmoduleReadmeHTMLQuery) {
				mockDetails.On("Execute", mock.Anything, "test", "mod", "provider", "1.0.0", "submod").
					Return(nil, assert.AnError).
					Once()
			},
			expectedBodyContains: "module provider not found",
		},
		{
			name:           "successful response",
			method:         "GET",
			url:            "/modules/test/mod/provider/1.0.0/submodules/details/submod",
			expectedStatus: http.StatusOK,
			setupMocks: func(mockDetails *MockGetSubmoduleDetailsQuery, mockReadme *MockGetSubmoduleReadmeHTMLQuery) {
				submoduleDetails := &module.SubmoduleDetails{
					Path:        "submod",
					Description: "Test Submodule",
					Readme:      "# Test Submodule\n\nThis is a test submodule",
					Files: []module.SubmoduleFile{
						{
							Path:     "main.tf",
							Content:  "terraform {}",
							IsBinary: false,
						},
					},
				}
				mockDetails.On("Execute", mock.Anything, "test", "mod", "provider", "1.0.0", "submod").
					Return(submoduleDetails, nil).
					Once()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockDetails := &MockGetSubmoduleDetailsQuery{}
			mockReadme := &MockGetSubmoduleReadmeHTMLQuery{}
			tt.setupMocks(mockDetails, mockReadme)

			handler := NewSubmoduleHandler(mockDetails, mockReadme)

			// Create request
			req := httptest.NewRequest(tt.method, tt.url, nil)

			// Set up chi context
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("namespace", "test")
			rctx.URLParams.Add("name", "mod")
			rctx.URLParams.Add("provider", "provider")
			rctx.URLParams.Add("version", "1.0.0")
			rctx.URLParams.Add("submodule", "submod")
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

			mockDetails.AssertExpectations(t)
		})
	}
}

func TestSubmoduleHandler_HandleSubmoduleReadmeHTML(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		url            string
		expectedStatus int
		setupMocks     func(*MockGetSubmoduleDetailsQuery, *MockGetSubmoduleReadmeHTMLQuery)
		checkHeaders   bool
	}{
		{
			name:           "invalid method",
			method:         "POST",
			url:            "/modules/test/mod/provider/1.0.0/submodules/readme_html/submod",
			expectedStatus: http.StatusMethodNotAllowed,
			setupMocks:     func(*MockGetSubmoduleDetailsQuery, *MockGetSubmoduleReadmeHTMLQuery) {},
		},
		{
			name:           "missing parameters",
			method:         "GET",
			url:            "/modules/test/mod/provider/1.0.0/submodules/readme_html/",
			expectedStatus: http.StatusBadRequest,
			setupMocks:     func(*MockGetSubmoduleDetailsQuery, *MockGetSubmoduleReadmeHTMLQuery) {},
		},
		{
			name:           "no readme content",
			method:         "GET",
			url:            "/modules/test/mod/provider/1.0.0/submodules/readme_html/submod",
			expectedStatus: http.StatusOK,
			setupMocks: func(mockDetails *MockGetSubmoduleDetailsQuery, mockReadme *MockGetSubmoduleReadmeHTMLQuery) {
				mockReadme.On("Execute", mock.Anything, "test", "mod", "provider", "1.0.0", "submod").
					Return("", assert.AnError).
					Once()
			},
			checkHeaders: true,
		},
		{
			name:           "successful response",
			method:         "GET",
			url:            "/modules/test/mod/provider/1.0.0/submodules/readme_html/submod",
			expectedStatus: http.StatusOK,
			setupMocks: func(mockDetails *MockGetSubmoduleDetailsQuery, mockReadme *MockGetSubmoduleReadmeHTMLQuery) {
				readmeHTML := "<h1>Test Submodule</h1><p>This is a test submodule</p>"
				mockReadme.On("Execute", mock.Anything, "test", "mod", "provider", "1.0.0", "submod").
					Return(readmeHTML, nil).
					Once()
			},
			checkHeaders: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockDetails := &MockGetSubmoduleDetailsQuery{}
			mockReadme := &MockGetSubmoduleReadmeHTMLQuery{}
			tt.setupMocks(mockDetails, mockReadme)

			handler := NewSubmoduleHandler(mockDetails, mockReadme)

			// Create request
			req := httptest.NewRequest(tt.method, tt.url, nil)

			// Set up chi context
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("namespace", "test")
			rctx.URLParams.Add("name", "mod")
			rctx.URLParams.Add("provider", "provider")
			rctx.URLParams.Add("version", "1.0.0")
			rctx.URLParams.Add("submodule", "submod")
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

			mockReadme.AssertExpectations(t)
		})
	}
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}
