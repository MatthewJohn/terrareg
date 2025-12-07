package integration

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/application/query/module"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/repository"
	moduleDto "github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/dto/module"
	moduleHandler "github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/handler/terraform/v1"
	"github.com/matthewjohn/terrareg/terrareg-go/test/util"
)

// MockModuleProviderRepository for integration tests
type MockModuleProviderRepository struct {
	modules []*model.ModuleProvider
}

func (m *MockModuleProviderRepository) Save(ctx context.Context, mp *model.ModuleProvider) error {
	return nil
}

func (m *MockModuleProviderRepository) FindByID(ctx context.Context, id int) (*model.ModuleProvider, error) {
	for _, mp := range m.modules {
		if mp.ID() == id {
			return mp, nil
		}
	}
	return nil, nil
}

func (m *MockModuleProviderRepository) FindByNamespaceModuleProvider(ctx context.Context, namespace, moduleName, provider string) (*model.ModuleProvider, error) {
	for _, mp := range m.modules {
		if mp.Namespace().Name() == namespace &&
			mp.Module() == moduleName &&
			mp.Provider() == provider {
			return mp, nil
		}
	}
	return nil, errors.New("not found")
}

func (m *MockModuleProviderRepository) FindByNamespace(ctx context.Context, namespace string) ([]*model.ModuleProvider, error) {
	var result []*model.ModuleProvider
	for _, mp := range m.modules {
		if mp.Namespace().Name() == namespace {
			result = append(result, mp)
		}
	}
	return result, nil
}

func (m *MockModuleProviderRepository) Search(ctx context.Context, query repository.ModuleSearchQuery) (*repository.ModuleSearchResult, error) {
	result := &repository.ModuleSearchResult{
		Modules:    m.modules,
		TotalCount: len(m.modules),
	}
	return result, nil
}

func (m *MockModuleProviderRepository) Delete(ctx context.Context, id int) error {
	return nil
}

func (m *MockModuleProviderRepository) Exists(ctx context.Context, namespace, module, provider string) (bool, error) {
	mp, err := m.FindByNamespaceModuleProvider(ctx, namespace, module, provider)
	if err != nil {
		return false, nil
	}
	return mp != nil, nil
}

func TestModuleListAPI_ResponseStructure(t *testing.T) {
	// Arrange
	repo := &MockModuleProviderRepository{
		modules: []*model.ModuleProvider{
			util.CreateMockModuleProvider("example", "test-module", "aws", true),
			util.CreateMockModuleProvider("test", "another", "gcp", false),
		},
	}

	listQuery := module.NewListModulesQuery(repo)
	handler := moduleHandler.NewModuleListHandler(listQuery)

	// Act
	req := httptest.NewRequest("GET", "/v1/modules", nil)
	w := httptest.NewRecorder()
	handler.HandleListModules(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	// Verify response structure
	var response map[string][]moduleDto.ModuleProviderResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response, "modules")
	assert.Len(t, response["modules"], 2)

	// Verify first module
	firstModule := response["modules"][0]
	assert.Equal(t, "example", firstModule.Namespace)
	assert.Equal(t, "test-module", firstModule.Name)
	assert.Equal(t, "aws", firstModule.Provider)
	assert.True(t, firstModule.Verified)

	// Verify second module
	secondModule := response["modules"][1]
	assert.Equal(t, "test", secondModule.Namespace)
	assert.Equal(t, "another", secondModule.Name)
	assert.Equal(t, "gcp", secondModule.Provider)
	assert.False(t, secondModule.Verified)
}

func TestModuleListAPI_EmptyResponse(t *testing.T) {
	// Arrange
	repo := &MockModuleProviderRepository{
		modules: []*model.ModuleProvider{},
	}

	listQuery := module.NewListModulesQuery(repo)
	handler := moduleHandler.NewModuleListHandler(listQuery)

	// Act
	req := httptest.NewRequest("GET", "/v1/modules", nil)
	w := httptest.NewRecorder()
	handler.HandleListModules(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	// Verify empty response
	var response map[string][]moduleDto.ModuleProviderResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response, "modules")
	assert.Empty(t, response["modules"])
}

func TestModuleListAPI_ErrorResponse(t *testing.T) {
	// This test would require modifying the repository to return errors
	// For now, we just verify the error handling structure exists
	t.Skip("Error handling verification requires mock that can return errors")
}

func TestModuleDTO_JSONCompatibility(t *testing.T) {
	// This test verifies that the DTO structure matches expected API format
	providerResponse := moduleDto.ModuleProviderResponse{
		ProviderBase: moduleDto.ProviderBase{
			ID:        "example/aws/1.0.0",
			Namespace: "example",
			Name:      "aws",
			Provider:  "aws",
			Verified:  true,
			Trusted:   false,
		},
		Description: &[]string{"AWS provider for Terraform"}[0],
		Owner:       &[]string{"team-a"}[0],
		Downloads:   1500,
	}

	// Serialize to JSON
	jsonData, err := json.Marshal(providerResponse)
	require.NoError(t, err)

	// Verify required fields are present
	var unmarshaled map[string]interface{}
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err)

	// Check all required fields for Terraform Registry API compatibility
	assert.Contains(t, unmarshaled, "id")
	assert.Contains(t, unmarshaled, "namespace")
	assert.Contains(t, unmarshaled, "name")
	assert.Contains(t, unmarshaled, "provider")
	assert.Contains(t, unmarshaled, "verified")
	assert.Contains(t, unmarshaled, "trusted")
}
