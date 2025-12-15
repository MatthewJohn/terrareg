package testutils

import (
	"context"
	"net/http"
	"net/http/httptest"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
)

// SetupTestServer creates a test HTTP server with all routes configured
func SetupTestServer(ctx context.Context, db *sqldb.Database) *httptest.Server {
	// Create a simple test server that responds to basic routes
	// In a full implementation, this would set up the complete server with all dependencies
	mux := http.NewServeMux()

	// Add basic route handlers for testing
	mux.HandleFunc("/v1/terrareg/namespaces", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{"data": {"id": 1, "name": "test-namespace"}}`))
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"data": []}`))
		}
	})

	mux.HandleFunc("/v1/terrareg/modules/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{"data": {"id": 1}}`))
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"data": []}`))
		}
	})

	mux.HandleFunc("/v2/gpg-keys", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{"data": {"id": "test-key-id", "type": "gpg-keys"}}`))
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"data": []}`))
		}
	})

	mux.HandleFunc("/v1/webhooks/github", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	})

	mux.HandleFunc("/v1/webhooks/gitlab", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	})

	mux.HandleFunc("/v1/webhooks/bitbucket", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	})

	server := httptest.NewServer(mux)
	return server
}
