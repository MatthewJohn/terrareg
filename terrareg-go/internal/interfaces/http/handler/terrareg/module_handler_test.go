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

	moduleQuery "github.com/matthewjohn/terrareg/terrareg-go/internal/application/query/module"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/repository"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/dto"
)

// MockGetSubmodulesQuery is a mock for GetSubmodulesQuery
type MockGetSubmodulesQuery struct {
	mock.Mock
}

func (m *MockGetSubmodulesQuery) Execute(ctx context.Context, namespace, name, provider, version string) ([]moduleQuery.SubmoduleInfo, error) {
	args := m.Called(ctx, namespace, name, provider, version)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]moduleQuery.SubmoduleInfo), args.Error(1)
}

// MockGetExamplesQuery is a mock for GetExamplesQuery
type MockGetExamplesQuery struct {
	mock.Mock
}

func (m *MockGetExamplesQuery) Execute(ctx context.Context, namespace, name, provider, version string) ([]moduleQuery.ExampleInfo, error) {
	args := m.Called(ctx, namespace, name, provider, version)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]moduleQuery.ExampleInfo), args.Error(1)
}

func TestModuleHandler_HandleGetSubmodules_Success(t *testing.T) {
	// Arrange
	mockSubmodulesQuery := &MockGetSubmodulesQuery{}
	handler := &ModuleHandler{
		getSubmodulesQuery: mockSubmodulesQuery,
	}

	// Create test data with Href fields
	expectedSubmodules := []moduleQuery.SubmoduleInfo{
		{Path: "submodule1", Href: "/modules/testns/testmod/testprov/1.0.0/submodule/submodule1"},
		{Path: "submodule2", Href: "/modules/testns/testmod/testprov/1.0.0/submodule/submodule2"},
	}

	// Setup mock
	mockSubmodulesQuery.On("Execute", mock.Anything, "testns", "testmod", "testprov", "1.0.0").
		Return(expectedSubmodules, nil)

	// Create HTTP request
	req := httptest.NewRequest("GET", "/v1/terrareg/modules/testns/testmod/testprov/1.0.0/submodules", nil)

	// Setup chi router context
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("namespace", "testns")
	rctx.URLParams.Add("name", "testmod")
	rctx.URLParams.Add("provider", "testprov")
	rctx.URLParams.Add("version", "1.0.0")

	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	// Create response recorder
	w := httptest.NewRecorder()

	// Act
	handler.HandleGetSubmodules(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	// Response should be a direct array, not wrapped
	var response []interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response, 2)

	// Verify submodule data with both path and href
	submodule1 := response[0].(map[string]interface{})
	assert.Equal(t, "submodule1", submodule1["path"])
	assert.Equal(t, "/modules/testns/testmod/testprov/1.0.0/submodule/submodule1", submodule1["href"])

	submodule2 := response[1].(map[string]interface{})
	assert.Equal(t, "submodule2", submodule2["path"])
	assert.Equal(t, "/modules/testns/testmod/testprov/1.0.0/submodule/submodule2", submodule2["href"])

	mockSubmodulesQuery.AssertExpectations(t)
}

func TestModuleHandler_HandleGetSubmodules_EmptySubmodules(t *testing.T) {
	// Arrange
	mockSubmodulesQuery := &MockGetSubmodulesQuery{}
	handler := &ModuleHandler{
		getSubmodulesQuery: mockSubmodulesQuery,
	}

	// Setup mock to return empty slice
	mockSubmodulesQuery.On("Execute", mock.Anything, "testns", "testmod", "testprov", "1.0.0").
		Return([]moduleQuery.SubmoduleInfo{}, nil)

	// Create HTTP request
	req := httptest.NewRequest("GET", "/v1/terrareg/modules/testns/testmod/testprov/1.0.0/submodules", nil)

	// Setup chi router context
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("namespace", "testns")
	rctx.URLParams.Add("name", "testmod")
	rctx.URLParams.Add("provider", "testprov")
	rctx.URLParams.Add("version", "1.0.0")

	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	// Create response recorder
	w := httptest.NewRecorder()

	// Act
	handler.HandleGetSubmodules(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	// Response should be a direct empty array
	var response []interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response, 0) // Empty array, not null

	mockSubmodulesQuery.AssertExpectations(t)
}

func TestModuleHandler_HandleGetSubmodules_QueryError(t *testing.T) {
	// Arrange
	mockSubmodulesQuery := &MockGetSubmodulesQuery{}
	handler := &ModuleHandler{
		getSubmodulesQuery: mockSubmodulesQuery,
	}

	// Setup mock to return error
	mockSubmodulesQuery.On("Execute", mock.Anything, "testns", "testmod", "testprov", "1.0.0").
		Return(nil, assert.AnError)

	// Create HTTP request
	req := httptest.NewRequest("GET", "/v1/terrareg/modules/testns/testmod/testprov/1.0.0/submodules", nil)

	// Setup chi router context
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("namespace", "testns")
	rctx.URLParams.Add("name", "testmod")
	rctx.URLParams.Add("provider", "testprov")
	rctx.URLParams.Add("version", "1.0.0")

	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	// Create response recorder
	w := httptest.NewRecorder()

	// Act
	handler.HandleGetSubmodules(w, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response dto.Error
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.NotEmpty(t, response.Message)

	mockSubmodulesQuery.AssertExpectations(t)
}

func TestModuleHandler_HandleGetSubmodules_MissingParameters(t *testing.T) {
	// Arrange
	mockSubmodulesQuery := &MockGetSubmodulesQuery{}
	handler := &ModuleHandler{
		getSubmodulesQuery: mockSubmodulesQuery,
	}

	// Create HTTP request with missing version parameter
	req := httptest.NewRequest("GET", "/v1/terrareg/modules/testns/testmod/testprov//submodules", nil)

	// Setup chi router context with missing version
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("namespace", "testns")
	rctx.URLParams.Add("name", "testmod")
	rctx.URLParams.Add("provider", "testprov")
	// Version is intentionally missing

	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	// Create response recorder
	w := httptest.NewRecorder()

	// Act
	handler.HandleGetSubmodules(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response dto.Error
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response.Message, "Missing required path parameters")

	// Mock should not be called since validation happens first
	mockSubmodulesQuery.AssertNotCalled(t, "Execute")
}

func TestModuleHandler_HandleGetExamples_Success(t *testing.T) {
	// Arrange
	mockExamplesQuery := &MockGetExamplesQuery{}
	handler := &ModuleHandler{
		getExamplesQuery: mockExamplesQuery,
	}

	// Create test data
	expectedExamples := []moduleQuery.ExampleInfo{
		{Path: "example1"},
		{Path: "example2"},
	}

	// Setup mock
	mockExamplesQuery.On("Execute", mock.Anything, "testns", "testmod", "testprov", "1.0.0").
		Return(expectedExamples, nil)

	// Create HTTP request
	req := httptest.NewRequest("GET", "/v1/terrareg/modules/testns/testmod/testprov/1.0.0/examples", nil)

	// Setup chi router context
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("namespace", "testns")
	rctx.URLParams.Add("name", "testmod")
	rctx.URLParams.Add("provider", "testprov")
	rctx.URLParams.Add("version", "1.0.0")

	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	// Create response recorder
	w := httptest.NewRecorder()

	// Act
	handler.HandleGetExamples(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Contains(t, response, "examples")
	examples := response["examples"].([]interface{})
	assert.Len(t, examples, 2)

	// Verify example data
	example1 := examples[0].(map[string]interface{})
	assert.Equal(t, "example1", example1["path"])

	example2 := examples[1].(map[string]interface{})
	assert.Equal(t, "example2", example2["path"])

	mockExamplesQuery.AssertExpectations(t)
}

func TestModuleHandler_HandleGetExamples_EmptyExamples(t *testing.T) {
	// Arrange
	mockExamplesQuery := &MockGetExamplesQuery{}
	handler := &ModuleHandler{
		getExamplesQuery: mockExamplesQuery,
	}

	// Setup mock to return empty slice
	mockExamplesQuery.On("Execute", mock.Anything, "testns", "testmod", "testprov", "1.0.0").
		Return([]moduleQuery.ExampleInfo{}, nil)

	// Create HTTP request
	req := httptest.NewRequest("GET", "/v1/terrareg/modules/testns/testmod/testprov/1.0.0/examples", nil)

	// Setup chi router context
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("namespace", "testns")
	rctx.URLParams.Add("name", "testmod")
	rctx.URLParams.Add("provider", "testprov")
	rctx.URLParams.Add("version", "1.0.0")

	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	// Create response recorder
	w := httptest.NewRecorder()

	// Act
	handler.HandleGetExamples(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Contains(t, response, "examples")
	examples := response["examples"].([]interface{})
	assert.Len(t, examples, 0) // Empty array, not null

	mockExamplesQuery.AssertExpectations(t)
}

func TestModuleHandler_HandleGetExamples_QueryError(t *testing.T) {
	// Arrange
	mockExamplesQuery := &MockGetExamplesQuery{}
	handler := &ModuleHandler{
		getExamplesQuery: mockExamplesQuery,
	}

	// Setup mock to return error
	mockExamplesQuery.On("Execute", mock.Anything, "testns", "testmod", "testprov", "1.0.0").
		Return(nil, assert.AnError)

	// Create HTTP request
	req := httptest.NewRequest("GET", "/v1/terrareg/modules/testns/testmod/testprov/1.0.0/examples", nil)

	// Setup chi router context
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("namespace", "testns")
	rctx.URLParams.Add("name", "testmod")
	rctx.URLParams.Add("provider", "testprov")
	rctx.URLParams.Add("version", "1.0.0")

	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	// Create response recorder
	w := httptest.NewRecorder()

	// Act
	handler.HandleGetExamples(w, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response dto.Error
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.NotEmpty(t, response.Message)

	mockExamplesQuery.AssertExpectations(t)
}