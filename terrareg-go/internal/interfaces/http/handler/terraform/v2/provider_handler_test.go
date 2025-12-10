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

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider"
	providerCmd "github.com/matthewjohn/terrareg/terrareg-go/internal/application/command/provider"
	providerQuery "github.com/matthewjohn/terrareg/terrareg-go/internal/application/query/provider"
)

// Mock implementations for testing
type MockGetProviderQuery struct {
	mock.Mock
}

func (m *MockGetProviderQuery) Execute(ctx context.Context, namespace, providerName string) (*provider.Provider, error) {
	args := m.Called(ctx, namespace, providerName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*provider.Provider), args.Error(1)
}

type MockGetProviderVersionsQuery struct {
	mock.Mock
}

func (m *MockGetProviderVersionsQuery) Execute(ctx context.Context, providerID int) ([]*provider.ProviderVersion, error) {
	args := m.Called(ctx, providerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*provider.ProviderVersion), args.Error(1)
}

type MockGetProviderVersionQuery struct {
	mock.Mock
}

func (m *MockGetProviderVersionQuery) Execute(ctx context.Context, providerID int, version string) (*provider.ProviderVersion, error) {
	args := m.Called(ctx, providerID, version)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*provider.ProviderVersion), args.Error(1)
}

type MockListProvidersQuery struct {
	mock.Mock
}

func (m *MockListProvidersQuery) Execute(ctx context.Context) ([]*provider.Provider, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*provider.Provider), args.Error(1)
}

type MockGetProviderDownloadQuery struct {
	mock.Mock
}

func (m *MockGetProviderDownloadQuery) Execute(ctx context.Context, req *providerCmd.GetProviderDownloadRequest) (*providerCmd.ProviderDownloadResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*providerCmd.ProviderDownloadResponse), args.Error(1)
}

// Mock provider and related types
type MockProvider struct {
	id           int
	name         string
	description  string
	repositoryID string
	tier         string
}

func NewMockProvider(id int, name, description, repositoryID, tier string) *MockProvider {
	return &MockProvider{
		id:           id,
		name:         name,
		description:  description,
		repositoryID: repositoryID,
		tier:         tier,
	}
}

func (p *MockProvider) ID() int                     { return p.id }
func (p *MockProvider) Name() string               { return p.name }
func (p *MockProvider) Description() string        { return p.description }
func (p *MockProvider) RepositoryID() string       { return p.repositoryID }
func (p *MockProvider) Tier() string               { return p.tier }

type MockProviderVersion struct {
	id               int
	version          string
	beta             bool
	protocolVersions []string
	binaries         []provider.ProviderBinary
}

func NewMockProviderVersion(id int, version string, beta bool) *MockProviderVersion {
	return &MockProviderVersion{
		id:               id,
		version:          version,
		beta:             beta,
		protocolVersions: []string{"5.0"},
		binaries:         []provider.ProviderBinary{},
	}
}

func (pv *MockProviderVersion) ID() int                         { return pv.id }
func (pv *MockProviderVersion) Version() string                 { return pv.version }
func (pv *MockProviderVersion) Beta() bool                       { return pv.beta }
func (pv *MockProviderVersion) ProtocolVersions() []string       { return pv.protocolVersions }
func (pv *MockProviderVersion) Binaries() []provider.ProviderBinary { return pv.binaries }

type MockProviderBinary struct {
	os           string
	architecture string
	filename     string
	fileHash     string
	downloadURL  string
}

func NewMockProviderBinary(os, arch, filename, fileHash, downloadURL string) *MockProviderBinary {
	return &MockProviderBinary{
		os:           os,
		architecture: arch,
		filename:     filename,
		fileHash:     fileHash,
		downloadURL:  downloadURL,
	}
}

func (pb *MockProviderBinary) OS() string           { return pb.os }
func (pb *MockProviderBinary) Architecture() string { return pb.architecture }
func (pb *MockProviderBinary) Filename() string     { return pb.filename }
func (pb *MockProviderBinary) FileHash() string     { return pb.fileHash }
func (pb *MockProviderBinary) DownloadURL() string  { return pb.downloadURL }

func TestTerraformV2ProviderHandler_HandleProviderDetails(t *testing.T) {
	tests := []struct {
		name           string
		namespace      string
		provider       string
		mockSetup      func(*MockGetProviderQuery)
		expectedStatus int
		expectedError  string
	}{
		{
			name:      "successful provider details",
			namespace: "hashicorp",
			provider:  "aws",
			mockSetup: func(m *MockGetProviderQuery) {
				provider := NewMockProvider(1, "aws", "AWS provider", "repo123", "official")
				m.On("Execute", mock.Anything, "hashicorp", "aws").Return(provider, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:      "provider not found",
			namespace: "nonexistent",
			provider:  "provider",
			mockSetup: func(m *MockGetProviderQuery) {
				m.On("Execute", mock.Anything, "nonexistent", "provider").Return(nil, assert.AnError)
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "Provider nonexistent/provider not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			getProviderQuery := &MockGetProviderQuery{}
			tt.mockSetup(getProviderQuery)

			getProviderVersionsQuery := &MockGetProviderVersionsQuery{}
			getProviderVersionQuery := &MockGetProviderVersionQuery{}
			listProvidersQuery := &MockListProvidersQuery{}
			getProviderDownloadQuery := &MockGetProviderDownloadQuery{}

			handler := NewTerraformV2ProviderHandler(
				getProviderQuery,
				getProviderVersionsQuery,
				getProviderVersionQuery,
				listProvidersQuery,
				getProviderDownloadQuery,
			)

			// Create request
			req := httptest.NewRequest("GET", "/v2/providers/"+tt.namespace+"/"+tt.provider, nil)
			w := httptest.NewRecorder()

			// Setup chi context
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("namespace", tt.namespace)
			rctx.URLParams.Add("provider", tt.provider)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			// Execute handler
			handler.HandleProviderDetails(w, req)

			// Verify response
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				// Verify response structure
				data, ok := response["data"].(map[string]interface{})
				assert.True(t, ok)
				assert.Equal(t, "providers", data["type"])
				assert.Equal(t, float64(1), data["id"]) // JSON numbers are float64

				attributes, ok := data["attributes"].(map[string]interface{})
				assert.True(t, ok)
				assert.Equal(t, "aws", attributes["name"])
				assert.Equal(t, "hashicorp", attributes["namespace"])
			} else {
				var errorResponse map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
				assert.NoError(t, err)
				assert.Contains(t, errorResponse["error"], tt.expectedError)
			}

			getProviderQuery.AssertExpectations(t)
		})
	}
}

func TestTerraformV2ProviderHandler_HandleProviderVersions(t *testing.T) {
	tests := []struct {
		name           string
		namespace      string
		provider       string
		mockSetup      func(*MockGetProviderQuery, *MockGetProviderVersionsQuery)
		expectedStatus int
	}{
		{
			name:      "successful provider versions",
			namespace: "hashicorp",
			provider:  "aws",
			mockSetup: func(providerQuery *MockGetProviderQuery, versionsQuery *MockGetProviderVersionsQuery) {
				provider := NewMockProvider(1, "aws", "AWS provider", "repo123", "official")
				providerQuery.On("Execute", mock.Anything, "hashicorp", "aws").Return(provider, nil)

				version1 := NewMockProviderVersion(1, "1.0.0", false)
				version2 := NewMockProviderVersion(2, "1.1.0", false)
				binary1 := NewMockProviderBinary("linux", "amd64", "terraform-provider-aws_1.0.0_linux_amd64.zip", "abc123", "https://github.com/hashicorp/terraform-provider-aws/releases/download/v1.0.0/terraform-provider-aws_1.0.0_linux_amd64.zip")
				binary2 := NewMockProviderBinary("darwin", "amd64", "terraform-provider-aws_1.1.0_darwin_amd64.zip", "def456", "https://github.com/hashicorp/terraform-provider-aws/releases/download/v1.1.0/terraform-provider-aws_1.1.0_darwin_amd64.zip")

				version1.binaries = []provider.ProviderBinary{binary1}
				version2.binaries = []provider.ProviderBinary{binary2}

				versionsQuery.On("Execute", mock.Anything, 1).Return([]*provider.ProviderVersion{version1, version2}, nil)
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			getProviderQuery := &MockGetProviderQuery{}
			getProviderVersionsQuery := &MockGetProviderVersionsQuery{}
			getProviderVersionQuery := &MockGetProviderVersionQuery{}
			listProvidersQuery := &MockListProvidersQuery{}
			getProviderDownloadQuery := &MockGetProviderDownloadQuery{}

			tt.mockSetup(getProviderQuery, getProviderVersionsQuery)

			handler := NewTerraformV2ProviderHandler(
				getProviderQuery,
				getProviderVersionsQuery,
				getProviderVersionQuery,
				listProvidersQuery,
				getProviderDownloadQuery,
			)

			// Create request
			req := httptest.NewRequest("GET", "/v2/providers/"+tt.namespace+"/"+tt.provider+"/versions", nil)
			w := httptest.NewRecorder()

			// Setup chi context
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("namespace", tt.namespace)
			rctx.URLParams.Add("provider", tt.provider)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			// Execute handler
			handler.HandleProviderVersions(w, req)

			// Verify response
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				// Verify response structure
				assert.Equal(t, "hashicorp/aws", response["id"])
				versions, ok := response["versions"].([]interface{})
				assert.True(t, ok)
				assert.Len(t, versions, 2)

				permissions, ok := response["permissions"].(map[string]interface{})
				assert.True(t, ok)
				assert.Equal(t, false, permissions["can_delete"])
			}

			getProviderQuery.AssertExpectations(t)
			getProviderVersionsQuery.AssertExpectations(t)
		})
	}
}

func TestTerraformV2ProviderHandler_HandleProviderDownload(t *testing.T) {
	tests := []struct {
		name           string
		namespace      string
		provider       string
		version        string
		os             string
		arch           string
		headers        map[string]string
		mockSetup      func(*MockGetProviderDownloadQuery)
		expectedStatus int
		expectedError  string
	}{
		{
			name:      "successful provider download",
			namespace: "hashicorp",
			provider:  "aws",
			version:   "1.0.0",
			os:        "linux",
			arch:      "amd64",
			headers:   map[string]string{"User-Agent": "Terraform/1.0.0"},
			mockSetup: func(m *MockGetProviderDownloadQuery) {
				response := &providerCmd.ProviderDownloadResponse{
					Protocols:    []string{"5.0"},
					OS:           "linux",
					Arch:         "amd64",
					Filename:     "terraform-provider-aws_1.0.0_linux_amd64.zip",
					DownloadURL:  "https://github.com/hashicorp/terraform-provider-aws/releases/download/v1.0.0/terraform-provider-aws_1.0.0_linux_amd64.zip",
					Shasum:       "abc123def456",
					ShasumsURL:  "https://github.com/hashicorp/terraform-provider-aws/releases/download/v1.0.0/terraform-provider-aws_1.0.0_SHA256SUMS",
				}
				m.On("Execute", mock.Anything, mock.MatchedBy(func(req *providerCmd.GetProviderDownloadRequest) bool {
					return req.Namespace == "hashicorp" && req.Provider == "aws" && req.Version == "1.0.0" && req.OS == "linux" && req.Arch == "amd64"
				})).Return(response, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:      "provider not found",
			namespace: "nonexistent",
			provider:  "provider",
			version:   "1.0.0",
			os:        "linux",
			arch:      "amd64",
			mockSetup: func(m *MockGetProviderDownloadQuery) {
				m.On("Execute", mock.Anything, mock.Anything).Return(nil, assert.AnError)
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "Provider version nonexistent/provider/1.0.0 not found",
		},
		{
			name:      "unsupported OS",
			namespace: "hashicorp",
			provider:  "aws",
			version:   "1.0.0",
			os:        "unsupported",
			arch:      "amd64",
			mockSetup: func(m *MockGetProviderDownloadQuery) {
				m.On("Execute", mock.Anything, mock.Anything).Return(nil, assert.AnError)
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Unsupported OS: unsupported",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			getProviderQuery := &MockGetProviderQuery{}
			getProviderVersionsQuery := &MockGetProviderVersionsQuery{}
			getProviderVersionQuery := &MockGetProviderVersionQuery{}
			listProvidersQuery := &MockListProvidersQuery{}
			getProviderDownloadQuery := &MockGetProviderDownloadQuery{}

			tt.mockSetup(getProviderDownloadQuery)

			handler := NewTerraformV2ProviderHandler(
				getProviderQuery,
				getProviderVersionsQuery,
				getProviderVersionQuery,
				listProvidersQuery,
				getProviderDownloadQuery,
			)

			// Create request
			req := httptest.NewRequest("GET", "/v2/providers/"+tt.namespace+"/"+tt.provider+"/"+tt.version+"/download/"+tt.os+"/"+tt.arch, nil)

			// Add headers
			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}

			w := httptest.NewRecorder()

			// Setup chi context
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("namespace", tt.namespace)
			rctx.URLParams.Add("provider", tt.provider)
			rctx.URLParams.Add("version", tt.version)
			rctx.URLParams.Add("os", tt.os)
			rctx.URLParams.Add("arch", tt.arch)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			// Execute handler
			handler.HandleProviderDownload(w, req)

			// Verify response
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response providerCmd.ProviderDownloadResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				assert.Equal(t, "linux", response.OS)
				assert.Equal(t, "amd64", response.Arch)
				assert.NotEmpty(t, response.DownloadURL)
			} else {
				var errorResponse map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
				assert.NoError(t, err)
				assert.Contains(t, errorResponse["error"], tt.expectedError)
			}

			getProviderDownloadQuery.AssertExpectations(t)
		})
	}
}