package terrareg

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/application/query/module"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http"
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

func (m *MockGetSubmoduleReadmeHTMLQuery) Execute(ctx context.Context, namespace, moduleName, provider, version, submodulePath string) (*module.SubmoduleReadmeHTMLResponse, error) {
	args := m.Called(ctx, namespace, moduleName, provider, version, submodulePath)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*module.SubmoduleReadmeHTMLResponse), args.Error(1)
}

func TestSubmoduleHandler_HandleSubmoduleDetails(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		url            string
		expectedStatus int
		setupMocks     func(*MockGetSubmoduleDetailsQuery, *MockGetSubmoduleReadmeHTMLQuery)
		expectedBody   string
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
			expectedBody:   "Missing required path parameters",
		},
		{
			name:           "missing module parameter",
			method:         "GET",
			url:            "/modules/test//provider/1.0.0/submodules/details/submod",
			expectedStatus: http.StatusBadRequest,
			setupMocks:     func(*MockGetSubmoduleDetailsQuery, *MockGetSubmoduleReadmeHTMLQuery) {},
			expectedBody:   "Missing required path parameters",
		},
		{
			name:           "missing provider parameter",
			method:         "GET",
			url:            "/modules/test/mod//1.0.0/submodules/details/submod",
			expectedStatus: http.StatusBadRequest,
			setupMocks:     func(*MockGetSubmoduleDetailsQuery, *MockGetSubmoduleReadmeHTMLQuery) {},
			expectedBody:   "Missing required path parameters",
		},
		{
			name:           "missing version parameter",
			method:         "GET",
			url:            "/modules/test/mod/provider//submodules/details/submod",
			expectedStatus: http.StatusBadRequest,
			setupMocks:     func(*MockGetSubmoduleDetailsQuery, *MockGetSubmoduleReadmeHTMLQuery) {},
			expectedBody:   "Missing required path parameters",
		},
		{
			name:           "missing submodule parameter",
			method:         "GET",
			url:            "/modules/test/mod/provider/1.0.0/submodules/details/",
			expectedStatus: http.StatusBadRequest,
			setupMocks:     func(*MockGetSubmoduleDetailsQuery, *MockGetSubmoduleReadmeHTMLQuery) {},
			expectedBody:   "Missing required path parameters",
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
			expectedBody: "module provider not found",
		},
		{
			name:           "successful response",
			method:         "GET",
			url:            "/modules/test/mod/provider/1.0.0/submodules/details/submod",
			expectedStatus: http.StatusOK,
			setupMocks: func(mockDetails *MockGetSubmoduleDetailsQuery, mockReadme *MockGetSubmoduleReadmeHTMLQuery) {
				submoduleDetails := &module.SubmoduleDetails{
					Path:  "submod",
					Name:  stringPtr("Test Submodule"),
					Type:  stringPtr("module"),
					Empty: false,
					Inputs: []module.Input{
						{
							Name:        "var1",
							Type:        "string",
							Description: "A test variable",
							Required:    true,
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

			if tt.expectedBody != "" {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				if err == nil {
					// If it's JSON, check the error message
					if errMsg, ok := response["error"].(string); ok {
						assert.Contains(t, errMsg, tt.expectedBody)
					}
				} else {
					// If not JSON, check the raw body
					assert.Contains(t, w.Body.String(), tt.expectedBody)
				}
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
					Return(nil, assert.AnError).
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
				readmeResponse := &module.SubmoduleReadmeHTMLResponse{
					HTML: "<h1>Test Submodule</h1><p>This is a test submodule</p>",
				}
				mockReadme.On("Execute", mock.Anything, "test", "mod", "provider", "1.0.0", "submod").
					Return(readmeResponse, nil).
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

			mockDetails.AssertExpectations(t)
			mockReadme.AssertExpectations(t)
		})
	}
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}

// Test that the handler properly integrates with the HTTP package
func TestSubmoduleHandler_Integration(t *testing.T) {
	// Arrange
	mockDetails := &MockGetSubmoduleDetailsQuery{}
	mockReadme := &MockGetSubmoduleReadmeHTMLQuery{}

	handler := NewSubmoduleHandler(mockDetails, mockReadme)

	// Test that RespondJSON works properly
	t.Run("RespondJSON integration", func(t *testing.T) {
		submoduleDetails := &module.SubmoduleDetails{
			Path:  "test",
			Empty: true,
		}

		mockDetails.On("Execute", mock.Anything, "test", "mod", "provider", "1.0.0", "test").
			Return(submoduleDetails, nil).
			Once()

		req := httptest.NewRequest("GET", "/modules/test/mod/provider/1.0.0/submodules/details/test", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("namespace", "test")
		rctx.URLParams.Add("name", "mod")
		rctx.URLParams.Add("provider", "provider")
		rctx.URLParams.Add("version", "1.0.0")
		rctx.URLParams.Add("submodule", "test")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		// Act
		handler.HandleSubmoduleDetails(w, req)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var response module.SubmoduleDetails
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "test", response.Path)
		assert.True(t, response.Empty)

		mockDetails.AssertExpectations(t)
	})
}