package v2

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/handler/terrareg"
)

// TerraformV2GPGHandler groups all /v2/gpg-keys handlers
type TerraformV2GPGHandler struct {
	// Add GPG service here when available
}

// NewTerraformV2GPGHandler creates a new TerraformV2GPGHandler
func NewTerraformV2GPGHandler() *TerraformV2GPGHandler {
	return &TerraformV2GPGHandler{}
}

// HandleListGPGKeys handles GET /v2/gpg-keys
func (h *TerraformV2GPGHandler) HandleListGPGKeys(w http.ResponseWriter, r *http.Request) {
	// For now, return empty list - can be enhanced with actual GPG key service later
	response := map[string]interface{}{
		"gpg_keys": []interface{}{},
	}
	terrareg.RespondJSON(w, http.StatusOK, response)
}

// HandleGetGPGKey handles GET /v2/gpg-keys/{namespace}/{key_id}
func (h *TerraformV2GPGHandler) HandleGetGPGKey(w http.ResponseWriter, r *http.Request) {
	_ = chi.URLParam(r, "namespace") // Parameter for future implementation
	_ = chi.URLParam(r, "key_id")   // Parameter for future implementation

	// For now, return 404 - can be enhanced with actual GPG key service later
	terrareg.RespondError(w, http.StatusNotFound, "GPG key not found")
}

// HandleCreateGPGKey handles POST /v2/gpg-keys
func (h *TerraformV2GPGHandler) HandleCreateGPGKey(w http.ResponseWriter, r *http.Request) {
	// For now, return 501 - can be enhanced with actual GPG key service later
	terrareg.RespondError(w, http.StatusNotImplemented, "GPG key creation not yet implemented")
}