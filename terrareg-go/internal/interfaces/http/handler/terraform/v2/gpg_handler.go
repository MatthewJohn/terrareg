package v2

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	providerCmd "github.com/matthewjohn/terrareg/terrareg-go/internal/application/command/provider"
	providerQuery "github.com/matthewjohn/terrareg/terrareg-go/internal/application/query/provider"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/handler/terrareg"
)

// TerraformV2GPGHandler groups all /v2/gpg-keys handlers
type TerraformV2GPGHandler struct {
	manageGPGKeyCmd        *providerCmd.ManageGPGKeyCommand
	getNamespaceGPGKeysQuery *providerQuery.GetNamespaceGPGKeysQuery
}

// NewTerraformV2GPGHandler creates a new TerraformV2GPGHandler
func NewTerraformV2GPGHandler(
	manageGPGKeyCmd *providerCmd.ManageGPGKeyCommand,
	getNamespaceGPGKeysQuery *providerQuery.GetNamespaceGPGKeysQuery,
) *TerraformV2GPGHandler {
	return &TerraformV2GPGHandler{
		manageGPGKeyCmd:        manageGPGKeyCmd,
		getNamespaceGPGKeysQuery: getNamespaceGPGKeysQuery,
	}
}

// HandleListGPGKeys handles GET /v2/gpg-keys
func (h *TerraformV2GPGHandler) HandleListGPGKeys(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse namespace filter parameter
	namespacesStr := r.URL.Query().Get("filter[namespace]")
	if namespacesStr == "" {
		terrareg.RespondError(w, http.StatusBadRequest, "Missing required parameter: filter[namespace]")
		return
	}

	// Split comma-separated namespaces
	namespaces := strings.Split(namespacesStr, ",")

	var allGPGKeys []providerQuery.GPGKeyResponse

	// Get GPG keys for each namespace
	for _, namespace := range namespaces {
		if namespace = strings.TrimSpace(namespace); namespace != "" {
			gpgKeys, err := h.getNamespaceGPGKeysQuery.Execute(ctx, namespace)
			if err != nil {
				// Continue with other namespaces if one fails
				continue
			}
			allGPGKeys = append(allGPGKeys, gpgKeys...)
		}
	}

	// Build response following Python terrareg format
	response := map[string]interface{}{
		"data": allGPGKeys,
	}

	terrareg.RespondJSON(w, http.StatusOK, response)
}

// HandleGetGPGKey handles GET /v2/gpg-keys/{namespace}/{key_id}
func (h *TerraformV2GPGHandler) HandleGetGPGKey(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	namespace := chi.URLParam(r, "namespace")
	keyID := chi.URLParam(r, "key_id")

	if namespace == "" || keyID == "" {
		terrareg.RespondError(w, http.StatusBadRequest, "Missing required parameters: namespace and key_id")
		return
	}

	// Get GPG keys for the namespace
	gpgKeys, err := h.getNamespaceGPGKeysQuery.Execute(ctx, namespace)
	if err != nil {
		terrareg.RespondError(w, http.StatusInternalServerError, "Failed to retrieve GPG keys")
		return
	}

	// Find the specific GPG key
	var foundKey *providerQuery.GPGKeyResponse
	for _, key := range gpgKeys {
		if key.ID == keyID {
			foundKey = &key
			break
		}
	}

	if foundKey == nil {
		terrareg.RespondError(w, http.StatusNotFound, "GPG key not found")
		return
	}

	// Build response following Python terrareg format
	response := map[string]interface{}{
		"data": foundKey,
	}

	terrareg.RespondJSON(w, http.StatusOK, response)
}

// HandleCreateGPGKey handles POST /v2/gpg-keys
func (h *TerraformV2GPGHandler) HandleCreateGPGKey(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var createRequest struct {
		Data struct {
			Type       string `json:"type"`
			Attributes struct {
				Namespace string `json:"namespace"`
				KeyText    string `json:"key_text"`
				ASCIIArmor string `json:"ascii_armor"`
				KeyID      string `json:"key_id"`
			} `json:"attributes"`
		} `json:"data"`
	}

	if err := json.NewDecoder(r.Body).Decode(&createRequest); err != nil {
		terrareg.RespondError(w, http.StatusBadRequest, "Invalid JSON request body")
		return
	}

	// Validate required fields
	if createRequest.Data.Attributes.Namespace == "" ||
	   createRequest.Data.Attributes.ASCIIArmor == "" ||
	   createRequest.Data.Attributes.KeyID == "" {
		terrareg.RespondError(w, http.StatusBadRequest, "Missing required fields: namespace, ascii_armor, key_id")
		return
	}

	// Execute command (TODO: Create namespace-scoped GPG key management)
	// For now, return 501 as this needs proper domain model support
	terrareg.RespondError(w, http.StatusNotImplemented, "Namespace-scoped GPG key management not yet implemented")
}

// HandleDeleteGPGKey handles DELETE /v2/gpg-keys/{namespace}/{key_id}
func (h *TerraformV2GPGHandler) HandleDeleteGPGKey(w http.ResponseWriter, r *http.Request) {
	namespace := chi.URLParam(r, "namespace")
	keyID := chi.URLParam(r, "key_id")

	if namespace == "" || keyID == "" {
		terrareg.RespondError(w, http.StatusBadRequest, "Missing required parameters: namespace and key_id")
		return
	}

	// TODO: Implement GPG key deletion when domain model supports it
	// This would involve:
	// 1. Check if GPG key is in use by any provider versions
	// 2. Delete GPG key if not in use
	// 3. Return appropriate error if in use
	terrareg.RespondError(w, http.StatusNotImplemented, "GPG key deletion not yet implemented")
}
