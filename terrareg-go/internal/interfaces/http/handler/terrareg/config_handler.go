package terrareg

import (
	"net/http"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/application/query/config"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/dto/terrareg"
)

// ConfigHandler handles config endpoint requests
type ConfigHandler struct {
	getConfigQuery *config.GetConfigQuery
}

// NewConfigHandler creates a new ConfigHandler
func NewConfigHandler(getConfigQuery *config.GetConfigQuery) *ConfigHandler {
	return &ConfigHandler{
		getConfigQuery: getConfigQuery,
	}
}

// HandleConfig handles the GET /v1/terrareg/config endpoint
func (h *ConfigHandler) HandleConfig(w http.ResponseWriter, r *http.Request) {
	// Only allow GET requests
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Execute query
	response, err := h.getConfigQuery.Execute(r.Context())
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Convert to DTO
	configResponse := terrareg.NewConfigResponse(response.Config)

	// Write JSON response
	RespondJSON(w, http.StatusOK, configResponse)
}
