package terrareg

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/git/service"
)

// GitProvidersHandler handles requests for listing git providers
// Python reference: server/api/terrareg_git_providers.py::ApiTerraregGitProviders
type GitProvidersHandler struct {
	gitProviderFactory *service.GitProviderFactory
}

// NewGitProvidersHandler creates a new git providers handler
func NewGitProvidersHandler(
	gitProviderFactory *service.GitProviderFactory,
) *GitProvidersHandler {
	return &GitProvidersHandler{
		gitProviderFactory: gitProviderFactory,
	}
}

// GitProviderResponse represents a git provider in the API response
type GitProviderResponse struct {
	ID               int    `json:"id"`
	Name             string `json:"name"`
	GitPathTemplate  string `json:"git_path_template"`
}

// ServeHTTP handles the HTTP request
// Python reference: server/api/terrareg_git_providers.py::_get()
// Note: Python uses auth_wrapper('can_access_read_api') but in Go this is handled at the router level with OptionalAuth middleware
func (h *GitProvidersHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// Get all git providers
	gitProviders, err := h.gitProviderFactory.GetAll(ctx)
	if err != nil {
		http.Error(w, "Failed to retrieve git providers", http.StatusInternalServerError)
		return
	}

	// Build response
	response := make([]GitProviderResponse, len(gitProviders))
	for i, gp := range gitProviders {
		response[i] = GitProviderResponse{
			ID:               gp.ID,
			Name:             gp.Name,
			GitPathTemplate:  gp.GitPathTemplate,
		}
	}

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}
