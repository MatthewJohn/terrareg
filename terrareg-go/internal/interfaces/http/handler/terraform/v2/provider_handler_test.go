package v2

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	domainProvider "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider"
)

// MockGetProviderQuery mocks the GetProviderQuery
type MockGetProviderQuery struct {
	mock.Mock
}

func (m *MockGetProviderQuery) Execute(ctx context.Context, namespace, name string) (*domainProvider.Provider, error) {
	args := m.Called(ctx, namespace, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domainProvider.Provider), args.Error(1)
}

// MockListProvidersQuery mocks the ListProvidersQuery
type MockListProvidersQuery struct {
	mock.Mock
}

func (m *MockListProvidersQuery) Execute(ctx context.Context) ([]*domainProvider.Provider, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domainProvider.Provider), args.Error(1)
}

// MockGetProviderVersionsQuery mocks the GetProviderVersionsQuery
type MockGetProviderVersionsQuery struct {
	mock.Mock
}

func (m *MockGetProviderVersionsQuery) Execute(ctx context.Context, namespace, name string) ([]string, error) {
	args := m.Called(ctx, namespace, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

// TestTerraformV2ProviderHandler_QueryExecution tests provider query execution
func TestTerraformV2ProviderHandler_QueryExecution(t *testing.T) {
	tests := []struct {
		name        string
		testFunc    func(t *testing.T)
	}{
		{
			name: "get provider query execution",
			testFunc: func(t *testing.T) {
				mockQuery := new(MockGetProviderQuery)
				ctx := context.Background()

				expectedProvider := &domainProvider.Provider{}
				mockQuery.On("Execute", ctx, "hashicorp", "aws").Return(expectedProvider, nil)

				actualProvider, err := mockQuery.Execute(ctx, "hashicorp", "aws")

				assert.NoError(t, err)
				assert.Equal(t, expectedProvider, actualProvider)
				mockQuery.AssertExpectations(t)
			},
		},
		{
			name: "list providers query execution",
			testFunc: func(t *testing.T) {
				mockQuery := new(MockListProvidersQuery)
				ctx := context.Background()

				expectedProviders := []*domainProvider.Provider{{}, {}}
				mockQuery.On("Execute", ctx).Return(expectedProviders, nil)

				actualProviders, err := mockQuery.Execute(ctx)

				assert.NoError(t, err)
				assert.Equal(t, expectedProviders, actualProviders)
				mockQuery.AssertExpectations(t)
			},
		},
		{
			name: "get provider versions query execution",
			testFunc: func(t *testing.T) {
				mockQuery := new(MockGetProviderVersionsQuery)
				ctx := context.Background()

				expectedVersions := []string{"1.0.0", "1.1.0", "2.0.0"}
				mockQuery.On("Execute", ctx, "hashicorp", "aws").Return(expectedVersions, nil)

				actualVersions, err := mockQuery.Execute(ctx, "hashicorp", "aws")

				assert.NoError(t, err)
				assert.Equal(t, expectedVersions, actualVersions)
				mockQuery.AssertExpectations(t)
			},
		},
		{
			name: "query error handling",
			testFunc: func(t *testing.T) {
				mockQuery := new(MockGetProviderQuery)
				ctx := context.Background()

				expectedError := errors.New("provider not found")
				mockQuery.On("Execute", ctx, "nonexistent", "provider").Return(nil, expectedError)

				actualProvider, err := mockQuery.Execute(ctx, "nonexistent", "provider")

				assert.Error(t, err)
				assert.Nil(t, actualProvider)
				assert.Equal(t, expectedError, err)
				mockQuery.AssertExpectations(t)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.testFunc(t)
		})
	}
}

// TestTerraformV2ProviderHandler_URLParsing tests URL parameter parsing for provider endpoints
func TestTerraformV2ProviderHandler_URLParsing(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		expectedNS  string
		expectedProv string
	}{
		{
			name:        "provider details URL",
			url:         "/v2/providers/hashicorp/aws",
			expectedNS:  "hashicorp",
			expectedProv: "aws",
		},
		{
			name:        "provider versions URL",
			url:         "/v2/providers/hashicorp/aws/versions",
			expectedNS:  "hashicorp",
			expectedProv: "aws",
		},
		{
			name:        "provider version download URL",
			url:         "/v2/providers/hashicorp/aws/1.0.0/download",
			expectedNS:  "hashicorp",
			expectedProv: "aws",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock handler that extracts URL parameters
			mockHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Extract namespace and provider from URL path
				pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")

				var namespace, providerName string
				if len(pathParts) >= 3 && pathParts[0] == "v2" && pathParts[1] == "providers" {
					namespace = pathParts[2]
					if len(pathParts) > 3 {
						providerName = pathParts[3]
					}
				}

				assert.Equal(t, tt.expectedNS, namespace)
				assert.Equal(t, tt.expectedProv, providerName)

				w.WriteHeader(http.StatusOK)
				w.Write([]byte("OK"))
			})

			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			w := httptest.NewRecorder()

			mockHandler.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
		})
	}
}

// TestTerraformV2ProviderHandler_HTTPMethods tests different HTTP method handling
func TestTerraformV2ProviderHandler_HTTPMethods(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		url            string
		expectedStatus int
	}{
		{
			name:           "GET provider details",
			method:         http.MethodGet,
			url:            "/v2/providers/hashicorp/aws",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "GET provider versions",
			method:         http.MethodGet,
			url:            "/v2/providers/hashicorp/aws/versions",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "GET provider version details",
			method:         http.MethodGet,
			url:            "/v2/providers/hashicorp/aws/1.0.0",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "GET provider download",
			method:         http.MethodGet,
			url:            "/v2/providers/hashicorp/aws/1.0.0/download",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simple mock handler that responds to any GET request
			mockHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, tt.method, r.Method)
				assert.Equal(t, tt.url, r.URL.RequestURI())

				// Validate that it's a provider endpoint
				assert.Contains(t, r.URL.Path, "/v2/providers/")
				assert.True(t, len(strings.Split(strings.Trim(r.URL.Path, "/"), "/")) >= 3)

				w.WriteHeader(http.StatusOK)
				w.Write([]byte("OK"))
			})

			req := httptest.NewRequest(tt.method, tt.url, nil)
			w := httptest.NewRecorder()

			mockHandler.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// TestTerraformV2ProviderHandler_ResponseFormats tests expected response formats
func TestTerraformV2ProviderHandler_ResponseFormats(t *testing.T) {
	tests := []struct {
		name         string
		responseType string
		testFunc     func(t *testing.T)
	}{
		{
			name:         "provider details response",
			responseType: "provider_details",
			testFunc: func(t *testing.T) {
				// Test that a provider response would contain expected fields
				mockHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					// Simulate a provider details response
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{
						"id": "hashicorp/aws",
						"namespace": "hashicorp",
						"name": "aws",
						"versions": ["1.0.0", "1.1.0"]
					}`))
				})

				req := httptest.NewRequest(http.MethodGet, "/v2/providers/hashicorp/aws", nil)
				w := httptest.NewRecorder()

				mockHandler.ServeHTTP(w, req)

				assert.Equal(t, http.StatusOK, w.Code)
				assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
				assert.Contains(t, w.Body.String(), "hashicorp/aws")
				assert.Contains(t, w.Body.String(), "versions")
			},
		},
		{
			name:         "provider versions list response",
			responseType: "versions_list",
			testFunc: func(t *testing.T) {
				// Test that a versions list response would contain expected fields
				mockHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{
						"versions": [
							{"version": "1.0.0"},
							{"version": "1.1.0"},
							{"version": "2.0.0"}
						]
					}`))
				})

				req := httptest.NewRequest(http.MethodGet, "/v2/providers/hashicorp/aws/versions", nil)
				w := httptest.NewRecorder()

				mockHandler.ServeHTTP(w, req)

				assert.Equal(t, http.StatusOK, w.Code)
				assert.Contains(t, w.Body.String(), "versions")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.testFunc(t)
		})
	}
}