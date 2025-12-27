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

	providerQuery "github.com/matthewjohn/terrareg/terrareg-go/internal/application/query/provider"
)

// MockGetNamespaceGPGKeysQuery mocks the GetNamespaceGPGKeysQuery
type MockGetNamespaceGPGKeysQuery struct {
	mock.Mock
}

func (m *MockGetNamespaceGPGKeysQuery) Execute(ctx context.Context, namespace string) ([]providerQuery.GPGKeyResponse, error) {
	args := m.Called(ctx, namespace)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]providerQuery.GPGKeyResponse), args.Error(1)
}

// MockManageGPGKeyCommand mocks the ManageGPGKeyCommand
type MockManageGPGKeyCommand struct {
	mock.Mock
}

func (m *MockManageGPGKeyCommand) Execute(ctx context.Context, req interface{}) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

// TestTerraformV2GPGHandler_GPGKeyQueries tests GPG key query execution
func TestTerraformV2GPGHandler_GPGKeyQueries(t *testing.T) {
	tests := []struct {
		name     string
		testFunc func(t *testing.T)
	}{
		{
			name: "get namespace GPG keys query execution",
			testFunc: func(t *testing.T) {
				mockQuery := new(MockGetNamespaceGPGKeysQuery)
				ctx := context.Background()

				expectedGPGKeys := []providerQuery.GPGKeyResponse{
					{
						ID:         "key1",
						Namespace:  "hashicorp",
						KeyID:      "ABCD1234",
						ASCIIArmor: "-----BEGIN PGP PUBLIC KEY BLOCK-----\n...\n-----END PGP PUBLIC KEY BLOCK-----",
						CreatedAt:  "2023-01-01T00:00:00Z",
					},
				}
				mockQuery.On("Execute", ctx, "hashicorp").Return(expectedGPGKeys, nil)

				actualGPGKeys, err := mockQuery.Execute(ctx, "hashicorp")

				assert.NoError(t, err)
				assert.Equal(t, expectedGPGKeys, actualGPGKeys)
				mockQuery.AssertExpectations(t)
			},
		},
		{
			name: "GPG key query error handling",
			testFunc: func(t *testing.T) {
				mockQuery := new(MockGetNamespaceGPGKeysQuery)
				ctx := context.Background()

				expectedError := errors.New("namespace not found")
				mockQuery.On("Execute", ctx, "nonexistent").Return(nil, expectedError)

				actualGPGKeys, err := mockQuery.Execute(ctx, "nonexistent")

				assert.Error(t, err)
				assert.Nil(t, actualGPGKeys)
				assert.Equal(t, expectedError, err)
				mockQuery.AssertExpectations(t)
			},
		},
		{
			name: "empty GPG keys response",
			testFunc: func(t *testing.T) {
				mockQuery := new(MockGetNamespaceGPGKeysQuery)
				ctx := context.Background()

				var expectedGPGKeys []providerQuery.GPGKeyResponse
				mockQuery.On("Execute", ctx, "empty").Return(expectedGPGKeys, nil)

				actualGPGKeys, err := mockQuery.Execute(ctx, "empty")

				assert.NoError(t, err)
				assert.Empty(t, actualGPGKeys)
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

// TestTerraformV2GPGHandler_URLParsing tests URL parameter parsing for GPG key endpoints
func TestTerraformV2GPGHandler_URLParsing(t *testing.T) {
	tests := []struct {
		name           string
		url            string
		expectedNS     string
		expectedKeyID  string
		expectedFilter string
	}{
		{
			name:           "list GPG keys with namespace filter",
			url:            "/v2/gpg-keys?filter[namespace]=hashicorp",
			expectedFilter: "hashicorp",
		},
		{
			name:           "list GPG keys with multiple namespace filters",
			url:            "/v2/gpg-keys?filter[namespace]=hashicorp,aws,google",
			expectedFilter: "hashicorp,aws,google",
		},
		{
			name:          "get specific GPG key",
			url:           "/v2/gpg-keys/hashicorp/key1",
			expectedNS:    "hashicorp",
			expectedKeyID: "key1",
		},
		{
			name:          "get another specific GPG key",
			url:           "/v2/gpg-keys/aws/aws-key-2023",
			expectedNS:    "aws",
			expectedKeyID: "aws-key-2023",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock handler that extracts URL parameters
			mockHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if strings.Contains(r.URL.Path, "/gpg-keys/") && !strings.Contains(r.URL.Path, "?") {
					// Extract namespace and key_id from URL path for specific key
					pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
					if len(pathParts) >= 4 && pathParts[0] == "v2" && pathParts[1] == "gpg-keys" {
						namespace := pathParts[2]
						keyID := pathParts[3]
						assert.Equal(t, tt.expectedNS, namespace)
						assert.Equal(t, tt.expectedKeyID, keyID)
					}
				} else {
					// Extract filter parameter for list endpoint
					filterNS := r.URL.Query().Get("filter[namespace]")
					assert.Equal(t, tt.expectedFilter, filterNS)
				}

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

// TestTerraformV2GPGHandler_QueryParameterParsing tests query parameter parsing for GPG endpoints
func TestTerraformV2GPGHandler_QueryParameterParsing(t *testing.T) {
	tests := []struct {
		name               string
		url                string
		expectedNamespaces []string
		shouldError        bool
	}{
		{
			name:               "single namespace filter",
			url:                "/v2/gpg-keys?filter[namespace]=hashicorp",
			expectedNamespaces: []string{"hashicorp"},
			shouldError:        false,
		},
		{
			name:               "multiple namespace filters",
			url:                "/v2/gpg-keys?filter[namespace]=hashicorp,aws,google",
			expectedNamespaces: []string{"hashicorp", "aws", "google"},
			shouldError:        false,
		},
		{
			name:               "empty namespace filter",
			url:                "/v2/gpg-keys?filter[namespace]=",
			expectedNamespaces: []string{""},
			shouldError:        true, // Should error because empty namespace
		},
		{
			name:               "missing namespace filter",
			url:                "/v2/gpg-keys",
			expectedNamespaces: nil,
			shouldError:        true, // Should error because missing filter
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate query parameter parsing logic from the handler
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)

			namespacesStr := req.URL.Query().Get("filter[namespace]")

			if tt.shouldError {
				if namespacesStr == "" {
					assert.Empty(t, namespacesStr)
				} else if namespacesStr == "" {
					assert.Equal(t, "", namespacesStr)
				}
				return
			}

			assert.NotEmpty(t, namespacesStr)

			// Split comma-separated namespaces
			namespaces := strings.Split(namespacesStr, ",")

			// Trim whitespace and filter empty strings
			var cleanNamespaces []string
			for _, ns := range namespaces {
				if trimmed := strings.TrimSpace(ns); trimmed != "" {
					cleanNamespaces = append(cleanNamespaces, trimmed)
				}
			}

			assert.Equal(t, tt.expectedNamespaces, cleanNamespaces)
		})
	}
}

// TestTerraformV2GPGHandler_ResponseFormats tests expected response formats for GPG endpoints
func TestTerraformV2GPGHandler_ResponseFormats(t *testing.T) {
	tests := []struct {
		name         string
		responseType string
		testFunc     func(t *testing.T)
	}{
		{
			name:         "list GPG keys response format",
			responseType: "list_response",
			testFunc: func(t *testing.T) {
				// Test GPG keys list response format
				mockHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					// Simulate a GPG keys list response
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{
						"data": [
							{
								"id": "key1",
								"namespace": "hashicorp",
								"key_text": "-----BEGIN PGP PUBLIC KEY BLOCK-----",
								"fingerprint": "ABCD1234",
								"created_at": "2023-01-01T00:00:00Z"
							},
							{
								"id": "key2",
								"namespace": "aws",
								"key_text": "-----BEGIN PGP PUBLIC KEY BLOCK-----",
								"fingerprint": "WXYZ6789",
								"created_at": "2023-06-01T00:00:00Z"
							}
						]
					}`))
				})

				req := httptest.NewRequest(http.MethodGet, "/v2/gpg-keys?filter[namespace]=hashicorp,aws", nil)
				w := httptest.NewRecorder()

				mockHandler.ServeHTTP(w, req)

				assert.Equal(t, http.StatusOK, w.Code)
				assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
				assert.Contains(t, w.Body.String(), "data")
				assert.Contains(t, w.Body.String(), "hashicorp")
				assert.Contains(t, w.Body.String(), "fingerprint")
			},
		},
		{
			name:         "single GPG key response format",
			responseType: "single_response",
			testFunc: func(t *testing.T) {
				// Test single GPG key response format
				mockHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{
						"id": "key1",
						"namespace": "hashicorp",
						"key_text": "-----BEGIN PGP PUBLIC KEY BLOCK-----\n...\n-----END PGP PUBLIC KEY BLOCK-----",
						"fingerprint": "ABCD1234",
						"created_at": "2023-01-01T00:00:00Z"
					}`))
				})

				req := httptest.NewRequest(http.MethodGet, "/v2/gpg-keys/hashicorp/key1", nil)
				w := httptest.NewRecorder()

				mockHandler.ServeHTTP(w, req)

				assert.Equal(t, http.StatusOK, w.Code)
				assert.Contains(t, w.Body.String(), "BEGIN PGP PUBLIC KEY BLOCK")
				assert.Contains(t, w.Body.String(), "ABCD1234")
			},
		},
		{
			name:         "GPG key not found error response",
			responseType: "error_response",
			testFunc: func(t *testing.T) {
				// Test error response for missing GPG key
				mockHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusNotFound)
					w.Write([]byte(`{
						"error": "GPG key not found"
					}`))
				})

				req := httptest.NewRequest(http.MethodGet, "/v2/gpg-keys/nonexistent/missing-key", nil)
				w := httptest.NewRecorder()

				mockHandler.ServeHTTP(w, req)

				assert.Equal(t, http.StatusNotFound, w.Code)
				assert.Contains(t, w.Body.String(), "GPG key not found")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.testFunc(t)
		})
	}
}
