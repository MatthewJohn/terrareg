package integration

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestRouter_BasicChiSetup(t *testing.T) {
	// Test basic chi router functionality without terrareg dependencies
	r := chi.NewRouter()

	// Add a simple test route
	r.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response"))
	})

	// Test that the route works
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "test response", w.Body.String())
}

func TestRouter_NotFound(t *testing.T) {
	// Test that chi router returns 404 for non-existent routes
	r := chi.NewRouter()

	// No routes added
	req := httptest.NewRequest("GET", "/nonexistent", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestRouter_MethodNotAllowed(t *testing.T) {
	// Test that chi router handles method not allowed
	r := chi.NewRouter()

	// Add GET route only
	r.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Test POST to GET route
	req := httptest.NewRequest("POST", "/test", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestRouter_PathParameters(t *testing.T) {
	// Test that chi router handles path parameters correctly
	r := chi.NewRouter()

	r.Get("/modules/{namespace}/{name}/{provider}", func(w http.ResponseWriter, r *http.Request) {
		namespace := chi.URLParam(r, "namespace")
		name := chi.URLParam(r, "name")
		provider := chi.URLParam(r, "provider")

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(namespace + "/" + name + "/" + provider))
	})

	// Test with valid parameters
	req := httptest.NewRequest("GET", "/modules/testns/testmod/testprov", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "testns/testmod/testprov", w.Body.String())
}
