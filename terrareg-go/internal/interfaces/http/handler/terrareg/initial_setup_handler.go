package terrareg

import (
	"net/http"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/application/query/setup"
)

// InitialSetupHandler handles initial setup endpoint requests
type InitialSetupHandler struct {
	getInitialSetupQuery *setup.GetInitialSetupQuery
}

// NewInitialSetupHandler creates a new InitialSetupHandler
func NewInitialSetupHandler(getInitialSetupQuery *setup.GetInitialSetupQuery) *InitialSetupHandler {
	return &InitialSetupHandler{
		getInitialSetupQuery: getInitialSetupQuery,
	}
}

// HandleInitialSetup handles the GET /v1/terrareg/initial_setup endpoint
func (h *InitialSetupHandler) HandleInitialSetup(w http.ResponseWriter, r *http.Request) {
	// Only allow GET requests
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Execute query
	response, err := h.getInitialSetupQuery.Execute(r.Context())
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Write JSON response
	RespondJSON(w, http.StatusOK, response)
}
