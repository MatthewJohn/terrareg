package terrareg

import (
	"encoding/json"
	"net/http"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/application/query/config"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/dto/terrareg"
)

// VersionHandler handles version endpoint requests
type VersionHandler struct {
	getVersionQuery *config.GetVersionQuery
}

// NewVersionHandler creates a new VersionHandler
func NewVersionHandler(getVersionQuery *config.GetVersionQuery) *VersionHandler {
	return &VersionHandler{
		getVersionQuery: getVersionQuery,
	}
}

// HandleVersion handles the GET /v1/terrareg/version endpoint
func (h *VersionHandler) HandleVersion(w http.ResponseWriter, r *http.Request) {
	// Only allow GET requests
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Execute query
	response, err := h.getVersionQuery.Execute(r.Context())
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Convert to DTO
	versionResponse := terrareg.NewVersionResponse(response.Version)

	// Write JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(versionResponse); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
