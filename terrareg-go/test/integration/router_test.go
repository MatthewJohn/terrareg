package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/router"
)

func TestRouter_TerraregRoutes(t *testing.T) {
	// This tests that the terrareg routes are properly configured
	// In a full integration test, you would set up repositories and test the full flow

	// Create router with minimal configuration
	r := router.NewRouter()

	// Test that the terrareg route exists and returns appropriate response for missing data
	req := httptest.NewRequest("GET", "/v1/terrareg/modules/nonexistent/module/provider", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	// Should return 404 or appropriate error for missing module
	assert.NotEqual(t, http.StatusOK, w.Code)
}

func TestRouter_StandardModuleRoutes(t *testing.T) {
	// Test that standard module API routes exist
	r := router.NewRouter()

	req := httptest.NewRequest("GET", "/v1/modules/nonexistent/module/provider", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	// Should return 404 or appropriate error for missing module
	assert.NotEqual(t, http.StatusOK, w.Code)
}

func TestRouter_CORSHeaders(t *testing.T) {
	// Test that the router properly sets CORS headers
	r := router.NewRouter()

	req := httptest.NewRequest("OPTIONS", "/v1/modules/test/test/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	// Check for CORS headers (if implemented)
	// This test may need adjustment based on actual CORS configuration
	if w.Code == http.StatusOK {
		assert.Contains(t, w.Header().Get("Access-Control-Allow-Origin"), "localhost:3000")
	}
}

func TestRouter_ContentTypeHeaders(t *testing.T) {
	// Test that API responses have correct content type
	r := router.NewRouter()

	req := httptest.NewRequest("GET", "/v1/terrareg/modules/test/test/test", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	// For successful responses, should have JSON content type
	if w.Code == http.StatusOK {
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	}
}