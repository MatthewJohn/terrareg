package terrareg_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/application/query/config"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/config/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/handler/terrareg"
	"github.com/matthewjohn/terrareg/terrareg-go/test/integration/testutils"
)

// FakeConfigRepository is a fake implementation of ConfigRepository for testing
type FakeConfigRepository struct {
	version string
}

func (f *FakeConfigRepository) GetConfig(ctx context.Context) (*model.UIConfig, error) {
	return nil, nil
}

func (f *FakeConfigRepository) GetVersion(ctx context.Context) (string, error) {
	return f.version, nil
}

func (f *FakeConfigRepository) IsOpenIDConnectEnabled(ctx context.Context) bool {
	return false
}

func (f *FakeConfigRepository) IsSAMLEnabled(ctx context.Context) bool {
	return false
}

func (f *FakeConfigRepository) IsAdminLoginEnabled(ctx context.Context) bool {
	return false
}

func (f *FakeConfigRepository) GetProviderSources(ctx context.Context) ([]model.ProviderSource, error) {
	return nil, nil
}

// TestVersionHandler_HandleVersion_Success tests successful version retrieval
func TestVersionHandler_HandleVersion_Success(t *testing.T) {
	// Create fake config repository
	fakeRepo := &FakeConfigRepository{version: "1.0.0-test"}
	getVersionQuery := config.NewGetVersionQuery(fakeRepo)
	handler := terrareg.NewVersionHandler(getVersionQuery)

	req := httptest.NewRequest("GET", "/v1/terrareg/version", nil)
	w := httptest.NewRecorder()

	handler.HandleVersion(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	response := testutils.GetJSONBody(t, w)
	assert.Contains(t, response, "version")
	assert.Equal(t, "1.0.0-test", response["version"])
}

// TestVersionHandler_HandleVersion_CustomVersion tests with custom version string
func TestVersionHandler_HandleVersion_CustomVersion(t *testing.T) {
	fakeRepo := &FakeConfigRepository{version: "2.5.0-rc1"}
	getVersionQuery := config.NewGetVersionQuery(fakeRepo)
	handler := terrareg.NewVersionHandler(getVersionQuery)

	req := httptest.NewRequest("GET", "/v1/terrareg/version", nil)
	w := httptest.NewRecorder()

	handler.HandleVersion(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	response := testutils.GetJSONBody(t, w)
	assert.Equal(t, "2.5.0-rc1", response["version"])
}

// TestVersionHandler_HandleVersion_EmptyVersion tests with empty version string
func TestVersionHandler_HandleVersion_EmptyVersion(t *testing.T) {
	fakeRepo := &FakeConfigRepository{version: ""}
	getVersionQuery := config.NewGetVersionQuery(fakeRepo)
	handler := terrareg.NewVersionHandler(getVersionQuery)

	req := httptest.NewRequest("GET", "/v1/terrareg/version", nil)
	w := httptest.NewRecorder()

	handler.HandleVersion(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	response := testutils.GetJSONBody(t, w)
	assert.Contains(t, response, "version")
	assert.Equal(t, "", response["version"])
}

// TestVersionHandler_HandleVersion_MethodNotAllowed tests that POST is rejected
func TestVersionHandler_HandleVersion_MethodNotAllowed(t *testing.T) {
	fakeRepo := &FakeConfigRepository{version: "1.0.0"}
	getVersionQuery := config.NewGetVersionQuery(fakeRepo)
	handler := terrareg.NewVersionHandler(getVersionQuery)

	req := httptest.NewRequest("POST", "/v1/terrareg/version", nil)
	w := httptest.NewRecorder()

	handler.HandleVersion(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

// TestVersionHandler_HandleVersion_PutMethodNotAllowed tests that PUT is rejected
func TestVersionHandler_HandleVersion_PutMethodNotAllowed(t *testing.T) {
	fakeRepo := &FakeConfigRepository{version: "1.0.0"}
	getVersionQuery := config.NewGetVersionQuery(fakeRepo)
	handler := terrareg.NewVersionHandler(getVersionQuery)

	req := httptest.NewRequest("PUT", "/v1/terrareg/version", nil)
	w := httptest.NewRecorder()

	handler.HandleVersion(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

// TestVersionHandler_HandleVersion_DeleteMethodNotAllowed tests that DELETE is rejected
func TestVersionHandler_HandleVersion_DeleteMethodNotAllowed(t *testing.T) {
	fakeRepo := &FakeConfigRepository{version: "1.0.0"}
	getVersionQuery := config.NewGetVersionQuery(fakeRepo)
	handler := terrareg.NewVersionHandler(getVersionQuery)

	req := httptest.NewRequest("DELETE", "/v1/terrareg/version", nil)
	w := httptest.NewRecorder()

	handler.HandleVersion(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}
