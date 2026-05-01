package terrareg_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/handler/terrareg"
	"github.com/matthewjohn/terrareg/terrareg-go/test/integration/testutils"
)

// TestGraphHandler_HandleGraphStatistics_Success tests successful graph statistics retrieval
func TestGraphHandler_HandleGraphStatistics_Success(t *testing.T) {
	// Create handler with nil dependencies - HandleGraphStatistics doesn't use them
	handler := terrareg.NewGraphHandler(nil, nil)

	req := httptest.NewRequest("GET", "/v1/terrareg/graph/statistics", nil)
	w := httptest.NewRecorder()

	handler.HandleGraphStatistics(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	response := testutils.GetJSONBody(t, w)

	// Check overview structure
	assert.Contains(t, response, "overview")
	overview := response["overview"].(map[string]interface{})
	assert.Contains(t, overview, "total_module_providers")
	assert.Contains(t, overview, "total_modules")
	assert.Contains(t, overview, "total_namespaces")
	assert.Contains(t, overview, "total_dependencies")

	// Check by_namespace structure
	assert.Contains(t, response, "by_namespace")
	byNamespace := response["by_namespace"].([]interface{})
	assert.GreaterOrEqual(t, len(byNamespace), 1)

	// Check top_dependencies structure
	assert.Contains(t, response, "top_dependencies")
	topDependencies := response["top_dependencies"].([]interface{})
	assert.GreaterOrEqual(t, len(topDependencies), 1)

	// Check metadata
	assert.Contains(t, response, "metadata")
	metadata := response["metadata"].(map[string]interface{})
	assert.Contains(t, metadata, "namespace")
	assert.Contains(t, metadata, "generated_at")
}

// TestGraphHandler_HandleGraphStatistics_WithNamespace tests with namespace filter
func TestGraphHandler_HandleGraphStatistics_WithNamespace(t *testing.T) {
	handler := terrareg.NewGraphHandler(nil, nil)

	req := httptest.NewRequest("GET", "/v1/terrareg/graph/statistics?namespace=hashicorp", nil)
	w := httptest.NewRecorder()

	handler.HandleGraphStatistics(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	response := testutils.GetJSONBody(t, w)

	// Check that namespace is included in metadata
	metadata := response["metadata"].(map[string]interface{})
	assert.Equal(t, "hashicorp", metadata["namespace"])
}

// TestGraphHandler_HandleGraphStatistics_ResponseStructure tests the complete response structure
func TestGraphHandler_HandleGraphStatistics_ResponseStructure(t *testing.T) {
	handler := terrareg.NewGraphHandler(nil, nil)

	req := httptest.NewRequest("GET", "/v1/terrareg/graph/statistics", nil)
	w := httptest.NewRecorder()

	handler.HandleGraphStatistics(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	response := testutils.GetJSONBody(t, w)

	// Verify overview values are present as numbers
	overview := response["overview"].(map[string]interface{})
	assert.IsType(t, float64(0), overview["total_module_providers"])
	assert.IsType(t, float64(0), overview["total_modules"])
	assert.IsType(t, float64(0), overview["total_namespaces"])
	assert.IsType(t, float64(0), overview["total_dependencies"])

	// Verify by_namespace has expected structure
	byNamespace := response["by_namespace"].([]interface{})
	for _, ns := range byNamespace {
		namespaceData := ns.(map[string]interface{})
		assert.Contains(t, namespaceData, "namespace")
		assert.Contains(t, namespaceData, "module_providers")
		assert.Contains(t, namespaceData, "modules")
		assert.Contains(t, namespaceData, "dependencies")
		assert.Contains(t, namespaceData, "avg_dependencies")
	}

	// Verify top_dependencies has expected structure
	topDependencies := response["top_dependencies"].([]interface{})
	for _, dep := range topDependencies {
		depData := dep.(map[string]interface{})
		assert.Contains(t, depData, "provider")
		assert.Contains(t, depData, "count")
		assert.Contains(t, depData, "percentage")
	}
}
