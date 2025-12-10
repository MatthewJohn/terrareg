package v2

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

	providerCmd "github.com/matthewjohn/terrareg/terrareg-go/internal/application/command/provider"
	providerQuery "github.com/matthewjohn/terrareg/terrareg-go/internal/application/query/provider"
)

// Mock implementations
type MockManageGPGKeyCommand struct {
	mock.Mock
}

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

func TestTerraformV2GPGHandler_HandleListGPGKeys(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    string
		mockSetup      func(*MockGetNamespaceGPGKeysQuery)
		expectedStatus int
		expectedKeys   int
	}{
		{
			name:        "successful listing with single namespace",
			queryParams: "?filter[namespace]=hashicorp",
			mockSetup: func(m *MockGetNamespaceGPGKeysQuery) {
				keys := []providerQuery.GPGKeyResponse{
					{
						ID:        "key123",
						Namespace: "hashicorp",
						ASCIIArmor: "-----BEGIN PGP PUBLIC KEY BLOCK-----\n...",
						TrustSignature: "trust",
						Source:     "github",
						CreatedAt:  "2023-01-01T00:00:00Z",
					},
				}
				m.On("Execute", mock.Anything, "hashicorp").Return(keys, nil)
			},
			expectedStatus: http.StatusOK,
			expectedKeys:   1,
		},
		{
			name:        "successful listing with multiple namespaces",
			queryParams: "?filter[namespace]=hashicorp,terraform",
			mockSetup: func(m *MockGetNamespaceGPGKeysQuery) {
				keys1 := []providerQuery.GPGKeyResponse{
					{ID: "key1", Namespace: "hashicorp"},
				}
				keys2 := []providerQuery.GPGKeyResponse{
					{ID: "key2", Namespace: "terraform"},
					{ID: "key3", Namespace: "terraform"},
				}
				m.On("Execute", mock.Anything, "hashicorp").Return(keys1, nil)
				m.On("Execute", mock.Anything, "terraform").Return(keys2, nil)
			},
			expectedStatus: http.StatusOK,
			expectedKeys:   3,
		},
		{
			name:        "empty namespace filter",
			queryParams: "",
			mockSetup:   func(m *MockGetNamespaceGPGKeysQuery) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "namespace with whitespace",
			queryParams: "?filter[namespace]=hashicorp, terraform",
			mockSetup: func(m *MockGetNamespaceGPGKeysQuery) {
				keys := []providerQuery.GPGKeyResponse{
					{ID: "key1", Namespace: "hashicorp"},
					{ID: "key2", Namespace: "terraform"},
				}
				m.On("Execute", mock.Anything, "hashicorp").Return(keys, nil)
				m.On("Execute", mock.Anything, "terraform").Return([]providerQuery.GPGKeyResponse{}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedKeys:   1,
		},
		{
			name:        "namespace not found",
			queryParams: "?filter[namespace]=nonexistent",
			mockSetup: func(m *MockGetNamespaceGPGKeysQuery) {
				m.On("Execute", mock.Anything, "nonexistent").Return(nil, assert.AnError)
			},
			expectedStatus: http.StatusOK, // Should continue and return empty list
			expectedKeys:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			manageGPGKeyCmd := &MockManageGPGKeyCommand{}
			getNamespaceGPGKeysQuery := &MockGetNamespaceGPGKeysQuery{}
			tt.mockSetup(getNamespaceGPGKeysQuery)

			handler := NewTerraformV2GPGHandler(manageGPGKeyCmd, getNamespaceGPGKeysQuery)

			// Create request
			req := httptest.NewRequest("GET", "/v2/gpg-keys"+tt.queryParams, nil)
			w := httptest.NewRecorder()

			// Execute handler
			handler.HandleListGPGKeys(w, req)

			// Verify response
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				// Verify response structure
				data, ok := response["data"].([]interface{})
				assert.True(t, ok)
				assert.Len(t, data, tt.expectedKeys)

				// If there are keys, verify their structure
				if len(data) > 0 {
					key, ok := data[0].(map[string]interface{})
					assert.True(t, ok)
					assert.NotEmpty(t, key["id"])
					assert.NotEmpty(t, key["namespace"])
				}
			} else {
				var errorResponse map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
				assert.NoError(t, err)
				assert.NotEmpty(t, errorResponse["error"])
			}

			getNamespaceGPGKeysQuery.AssertExpectations(t)
		})
	}
}

func TestTerraformV2GPGHandler_HandleGetGPGKey(t *testing.T) {
	tests := []struct {
		name           string
		namespace      string
		keyID          string
		mockSetup      func(*MockGetNamespaceGPGKeysQuery)
		expectedStatus int
		expectFound    bool
	}{
		{
			name:      "successful key retrieval",
			namespace: "hashicorp",
			keyID:     "key123",
			mockSetup: func(m *MockGetNamespaceGPGKeysQuery) {
				keys := []providerQuery.GPGKeyResponse{
					{
						ID:        "key123",
						Namespace: "hashicorp",
						ASCIIArmor: "-----BEGIN PGP PUBLIC KEY BLOCK-----\n...",
						TrustSignature: "trust",
						Source:     "github",
						CreatedAt:  "2023-01-01T00:00:00Z",
					},
					{
						ID:        "key456",
						Namespace: "hashicorp",
						ASCIIArmor: "-----BEGIN PGP PUBLIC KEY BLOCK-----\n...",
						TrustSignature: "trust",
						Source:     "github",
						CreatedAt:  "2023-01-01T00:00:00Z",
					},
				}
				m.On("Execute", mock.Anything, "hashicorp").Return(keys, nil)
			},
			expectedStatus: http.StatusOK,
			expectFound:    true,
		},
		{
			name:      "key not found",
			namespace: "hashicorp",
			keyID:     "nonexistent",
			mockSetup: func(m *MockGetNamespaceGPGKeysQuery) {
				keys := []providerQuery.GPGKeyResponse{
					{ID: "key123", Namespace: "hashicorp"},
				}
				m.On("Execute", mock.Anything, "hashicorp").Return(keys, nil)
			},
			expectedStatus: http.StatusNotFound,
			expectFound:    false,
		},
		{
			name:      "namespace not found",
			namespace: "nonexistent",
			keyID:     "key123",
			mockSetup: func(m *MockGetNamespaceGPGKeysQuery) {
				m.On("Execute", mock.Anything, "nonexistent").Return(nil, assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
			expectFound:    false,
		},
		{
			name:           "missing namespace",
			namespace:      "",
			keyID:          "key123",
			mockSetup:      func(m *MockGetNamespaceGPGKeysQuery) {},
			expectedStatus: http.StatusBadRequest,
			expectFound:    false,
		},
		{
			name:           "missing key ID",
			namespace:      "hashicorp",
			keyID:          "",
			mockSetup:      func(m *MockGetNamespaceGPGKeysQuery) {},
			expectedStatus: http.StatusBadRequest,
			expectFound:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			manageGPGKeyCmd := &MockManageGPGKeyCommand{}
			getNamespaceGPGKeysQuery := &MockGetNamespaceGPGKeysQuery{}
			tt.mockSetup(getNamespaceGPGKeysQuery)

			handler := NewTerraformV2GPGHandler(manageGPGKeyCmd, getNamespaceGPGKeysQuery)

			// Create request
			req := httptest.NewRequest("GET", "/v2/gpg-keys/"+tt.namespace+"/"+tt.keyID, nil)
			w := httptest.NewRecorder()

			// Setup chi context
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("namespace", tt.namespace)
			rctx.URLParams.Add("key_id", tt.keyID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			// Execute handler
			handler.HandleGetGPGKey(w, req)

			// Verify response
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK && tt.expectFound {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				// Verify response structure
				data, ok := response["data"].(map[string]interface{})
				assert.True(t, ok)
				assert.Equal(t, tt.keyID, data["id"])
				assert.Equal(t, tt.namespace, data["namespace"])
			} else {
				var errorResponse map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
				assert.NoError(t, err)
				assert.NotEmpty(t, errorResponse["error"])
			}

			if tt.namespace != "" && tt.keyID != "" {
				getNamespaceGPGKeysQuery.AssertExpectations(t)
			}
		})
	}
}

func TestTerraformV2GPGHandler_HandleCreateGPGKey(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name: "valid GPG key creation request",
			requestBody: map[string]interface{}{
				"data": map[string]interface{}{
					"type": "gpg-keys",
					"attributes": map[string]interface{}{
						"namespace":   "hashicorp",
						"ascii_armor": "-----BEGIN PGP PUBLIC KEY BLOCK-----\n...",
						"key_id":      "ABC123",
						"key_text":    "Some key text",
					},
				},
			},
			expectedStatus: http.StatusNotImplemented, // Returns 501 as TODO
			expectedError:  "Namespace-scoped GPG key management not yet implemented",
		},
		{
			name: "missing namespace",
			requestBody: map[string]interface{}{
				"data": map[string]interface{}{
					"type": "gpg-keys",
					"attributes": map[string]interface{}{
						"ascii_armor": "-----BEGIN PGP PUBLIC KEY BLOCK-----\n...",
						"key_id":      "ABC123",
					},
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Missing required fields: namespace, ascii_armor, key_id",
		},
		{
			name: "missing ascii_armor",
			requestBody: map[string]interface{}{
				"data": map[string]interface{}{
					"type": "gpg-keys",
					"attributes": map[string]interface{}{
						"namespace": "hashicorp",
						"key_id":    "ABC123",
					},
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Missing required fields: namespace, ascii_armor, key_id",
		},
		{
			name: "missing key_id",
			requestBody: map[string]interface{}{
				"data": map[string]interface{}{
					"type": "gpg-keys",
					"attributes": map[string]interface{}{
						"namespace":   "hashicorp",
						"ascii_armor": "-----BEGIN PGP PUBLIC KEY BLOCK-----\n...",
					},
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Missing required fields: namespace, ascii_armor, key_id",
		},
		{
			name:           "invalid JSON",
			requestBody:    "invalid json",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid JSON request body",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			manageGPGKeyCmd := &MockManageGPGKeyCommand{}
			getNamespaceGPGKeysQuery := &MockGetNamespaceGPGKeysQuery{}

			handler := NewTerraformV2GPGHandler(manageGPGKeyCmd, getNamespaceGPGKeysQuery)

			// Create request body
			var body []byte
			if strBody, ok := tt.requestBody.(string); ok {
				body = []byte(strBody)
			} else {
				var err error
				body, err = json.Marshal(tt.requestBody)
				assert.NoError(t, err)
			}

			req := httptest.NewRequest("POST", "/v2/gpg-keys", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			// Execute handler
			handler.HandleCreateGPGKey(w, req)

			// Verify response
			assert.Equal(t, tt.expectedStatus, w.Code)

			var errorResponse map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
			assert.NoError(t, err)
			assert.Contains(t, errorResponse["error"], tt.expectedError)
		})
	}
}

func TestTerraformV2GPGHandler_HandleDeleteGPGKey(t *testing.T) {
	tests := []struct {
		name           string
		namespace      string
		keyID          string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "valid delete request",
			namespace:      "hashicorp",
			keyID:          "ABC123",
			expectedStatus: http.StatusNotImplemented, // Returns 501 as TODO
			expectedError:  "GPG key deletion not yet implemented",
		},
		{
			name:           "missing namespace",
			namespace:      "",
			keyID:          "ABC123",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Missing required parameters: namespace and key_id",
		},
		{
			name:           "missing key ID",
			namespace:      "hashicorp",
			keyID:          "",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Missing required parameters: namespace and key_id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			manageGPGKeyCmd := &MockManageGPGKeyCommand{}
			getNamespaceGPGKeysQuery := &MockGetNamespaceGPGKeysQuery{}

			handler := NewTerraformV2GPGHandler(manageGPGKeyCmd, getNamespaceGPGKeysQuery)

			// Create request
			req := httptest.NewRequest("DELETE", "/v2/gpg-keys/"+tt.namespace+"/"+tt.keyID, nil)
			w := httptest.NewRecorder()

			// Setup chi context
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("namespace", tt.namespace)
			rctx.URLParams.Add("key_id", tt.keyID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			// Execute handler
			handler.HandleDeleteGPGKey(w, req)

			// Verify response
			assert.Equal(t, tt.expectedStatus, w.Code)

			var errorResponse map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
			assert.NoError(t, err)
			assert.Contains(t, errorResponse["error"], tt.expectedError)
		})
	}
}