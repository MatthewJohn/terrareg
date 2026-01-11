package v1

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/application/query/module"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/repository"
)

// MockModuleProviderRepository is a mock for ModuleProviderRepository
type MockModuleProviderRepository struct {
	mock.Mock
}

func (m *MockModuleProviderRepository) Save(ctx context.Context, mp *model.ModuleProvider) error {
	args := m.Called(ctx, mp)
	return args.Error(0)
}

func (m *MockModuleProviderRepository) FindByID(ctx context.Context, id int) (*model.ModuleProvider, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.ModuleProvider), args.Error(1)
}

func (m *MockModuleProviderRepository) FindByNamespaceModuleProvider(ctx context.Context, namespace, moduleName, provider string) (*model.ModuleProvider, error) {
	args := m.Called(ctx, namespace, moduleName, provider)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.ModuleProvider), args.Error(1)
}

func (m *MockModuleProviderRepository) FindByNamespace(ctx context.Context, namespace string) ([]*model.ModuleProvider, error) {
	args := m.Called(ctx, namespace)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.ModuleProvider), args.Error(1)
}

func (m *MockModuleProviderRepository) Search(ctx context.Context, query repository.ModuleSearchQuery) (*repository.ModuleSearchResult, error) {
	args := m.Called(ctx, query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.ModuleSearchResult), args.Error(1)
}

func (m *MockModuleProviderRepository) Delete(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockModuleProviderRepository) Exists(ctx context.Context, namespace, module, provider string) (bool, error) {
	args := m.Called(ctx, namespace, module, provider)
	return args.Bool(0), args.Error(1)
}

func createMockNamespace(name string) *model.Namespace {
	namespace, err := model.NewNamespace(name, nil, "NONE")
	if err != nil {
		panic(err) // In tests, panic is acceptable for mock creation failures
	}
	return namespace
}

func createMockModuleProvider(namespace, moduleName, provider string, verified bool) *model.ModuleProvider {
	ns := createMockNamespace(namespace)
	mp, err := model.NewModuleProvider(ns, moduleName, provider)
	if err != nil {
		panic(err) // In tests, panic is acceptable for mock creation failures
	}

	if verified {
		mp.Verify()
	}

	return mp
}

func createMockModuleVersion(version string, owner, description *string, publishedAt *time.Time) *model.ModuleVersion {
	details := model.NewModuleDetails(nil)

	mv, _ := model.ReconstructModuleVersion(
		1, // id
		version,
		details,
		false,              // beta
		false,              // internal
		publishedAt != nil, // published
		publishedAt,
		nil,   // gitSHA
		nil,   // gitPath
		false, // archiveGitPath
		nil,   // repoBaseURLTemplate
		nil,   // repoCloneURLTemplate
		nil,   // repoBrowseURLTemplate
		owner,
		description,
		nil,        // variableTemplate
		nil,        // extractionVersion
		time.Now(), // createdAt
		time.Now(), // updatedAt
	)

	return mv
}

func TestModuleListHandler_HandleListModules_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockModuleProviderRepository)
	listQuery := module.NewListModulesQuery(mockRepo)
	handler := NewModuleListHandler(listQuery)

	// Create mock module providers
	publishedAt := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	owner := "team-a"
	description := "Test module"
	mv := createMockModuleVersion("1.0.0", &owner, &description, &publishedAt)

	mp1 := createMockModuleProvider("example", "aws", "aws", true)
	mp1.AddVersion(mv)

	mp2 := createMockModuleProvider("test", "gcp", "gcp", false)

	modules := []*model.ModuleProvider{mp1, mp2}
	searchResult := &repository.ModuleSearchResult{
		Modules:    modules,
		TotalCount: 2,
	}
	mockRepo.On("Search", mock.Anything, repository.ModuleSearchQuery{}).Return(searchResult, nil)

	// Create HTTP request
	req := httptest.NewRequest("GET", "/v1/modules", nil)
	w := httptest.NewRecorder()

	// Act
	handler.HandleListModules(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	// Verify the response structure
	expectedBody := `{"modules":[`
	assert.Contains(t, w.Body.String(), expectedBody)
	assert.Contains(t, w.Body.String(), `"namespace":"example"`)
	assert.Contains(t, w.Body.String(), `"name":"aws"`)
	assert.Contains(t, w.Body.String(), `"provider":"aws"`)
	assert.Contains(t, w.Body.String(), `"verified":true`)

	// Verify the mock was called
	mockRepo.AssertExpectations(t)
}

func TestModuleListHandler_HandleListModules_Empty(t *testing.T) {
	// Arrange
	mockRepo := new(MockModuleProviderRepository)
	listQuery := module.NewListModulesQuery(mockRepo)
	handler := NewModuleListHandler(listQuery)

	// Return empty list
	modules := []*model.ModuleProvider{}
	searchResult := &repository.ModuleSearchResult{
		Modules:    modules,
		TotalCount: 0,
	}
	mockRepo.On("Search", mock.Anything, repository.ModuleSearchQuery{}).Return(searchResult, nil)

	// Create HTTP request
	req := httptest.NewRequest("GET", "/v1/modules", nil)
	w := httptest.NewRecorder()

	// Act
	handler.HandleListModules(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `{"modules":[]}`)
}

func TestModuleListHandler_HandleListModules_Error(t *testing.T) {
	// Arrange
	mockRepo := new(MockModuleProviderRepository)
	listQuery := module.NewListModulesQuery(mockRepo)
	handler := NewModuleListHandler(listQuery)

	// Return error
	mockRepo.On("Search", mock.Anything, repository.ModuleSearchQuery{}).Return(nil, errors.New("database connection failed"))

	// Create HTTP request
	req := httptest.NewRequest("GET", "/v1/modules", nil)
	w := httptest.NewRecorder()

	// Act
	handler.HandleListModules(w, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	// The error response is JSON with "error" field
	assert.Contains(t, w.Body.String(), "Internal Server Error")
}

func TestModuleListHandler_HandleListModules_WithUnverified(t *testing.T) {
	// Arrange
	mockRepo := new(MockModuleProviderRepository)
	listQuery := module.NewListModulesQuery(mockRepo)
	handler := NewModuleListHandler(listQuery)

	// Create mock module providers (one verified, one not)
	mp1 := createMockModuleProvider("verified", "module1", "aws", true)
	mp2 := createMockModuleProvider("unverified", "module2", "gcp", false)

	modules := []*model.ModuleProvider{mp1, mp2}
	searchResult := &repository.ModuleSearchResult{
		Modules:    modules,
		TotalCount: 2,
	}
	mockRepo.On("Search", mock.Anything, repository.ModuleSearchQuery{}).Return(searchResult, nil)

	// Create HTTP request
	req := httptest.NewRequest("GET", "/v1/modules", nil)
	w := httptest.NewRecorder()

	// Act
	handler.HandleListModules(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	// Verify both modules are included
	body := w.Body.String()
	assert.Contains(t, body, `"namespace":"verified"`)
	assert.Contains(t, body, `"verified":true`)
	assert.Contains(t, body, `"namespace":"unverified"`)
	assert.Contains(t, body, `"verified":false`)
}

func TestConvertModuleProviderToListResponse(t *testing.T) {
	// Arrange
	publishedAt := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	owner := "test-owner"
	description := "Test description"
	mv := createMockModuleVersion("1.0.0", &owner, &description, &publishedAt)

	mp := createMockModuleProvider("testns", "testmodule", "testprovider", true)
	mp.AddVersion(mv)

	// Act
	result := convertModuleProviderToListResponse(mp)

	// Assert
	assert.NotNil(t, result)
	assert.Equal(t, "0", result.ID) // New ModuleProvider has ID 0
	assert.Equal(t, "testns", result.Namespace)
	assert.Equal(t, "testmodule", result.Name)
	assert.Equal(t, "testprovider", result.Provider)
	assert.True(t, result.Verified)
	assert.False(t, result.Trusted)          // TODO: Get from namespace service
	assert.Equal(t, "testns", *result.Owner) // Owner uses namespace name
	assert.Equal(t, "Test description", *result.Description)
	assert.Equal(t, "2023-01-01T12:00:00Z", *result.PublishedAt)
	assert.Equal(t, 0, result.Downloads) // Placeholder
}
