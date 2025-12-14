package v2

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	gpgkeyCmd "github.com/matthewjohn/terrareg/terrareg-go/internal/application/command/gpgkey"
	gpgkeyQuery "github.com/matthewjohn/terrareg/terrareg-go/internal/application/query/gpgkey"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/handler/terrareg"
)

// TerraformV2GPGHandler groups all /v2/gpg-keys handlers
type TerraformV2GPGHandler struct {
	manageGPGKeyCmd                 *gpgkeyCmd.ManageGPGKeyCommand
	getNamespaceGPGKeysQuery        *gpgkeyQuery.GetNamespaceGPGKeysQuery
	getMultipleNamespaceGPGKeysQuery *gpgkeyQuery.GetMultipleNamespaceGPGKeysQuery
	getGPGKeyQuery                  *gpgkeyQuery.GetGPGKeyQuery
}

// NewTerraformV2GPGHandler creates a new TerraformV2GPGHandler
func NewTerraformV2GPGHandler(
	manageGPGKeyCmd *gpgkeyCmd.ManageGPGKeyCommand,
	getNamespaceGPGKeysQuery *gpgkeyQuery.GetNamespaceGPGKeysQuery,
	getMultipleNamespaceGPGKeysQuery *gpgkeyQuery.GetMultipleNamespaceGPGKeysQuery,
	getGPGKeyQuery *gpgkeyQuery.GetGPGKeyQuery,
) *TerraformV2GPGHandler {
	return &TerraformV2GPGHandler{
		manageGPGKeyCmd:                 manageGPGKeyCmd,
		getNamespaceGPGKeysQuery:        getNamespaceGPGKeysQuery,
		getMultipleNamespaceGPGKeysQuery: getMultipleNamespaceGPGKeysQuery,
		getGPGKeyQuery:                  getGPGKeyQuery,
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

	// Split comma-separated namespaces and trim whitespace
	namespaces := strings.Split(namespacesStr, ",")
	cleanNamespaces := make([]string, 0, len(namespaces))
	for _, ns := range namespaces {
		if ns = strings.TrimSpace(ns); ns != "" {
			cleanNamespaces = append(cleanNamespaces, ns)
		}
	}

	if len(cleanNamespaces) == 0 {
		terrareg.RespondError(w, http.StatusBadRequest, "No valid namespaces provided")
		return
	}

	var allGPGKeys []gpgkeyQuery.GPGKeyResponse
	var err error

	if len(cleanNamespaces) == 1 {
		// Use single namespace query for efficiency
		allGPGKeys, err = h.getNamespaceGPGKeysQuery.Execute(ctx, cleanNamespaces[0])
	} else {
		// Use multiple namespace query
		allGPGKeys, err = h.getMultipleNamespaceGPGKeysQuery.Execute(ctx, cleanNamespaces)
	}

	if err != nil {
		terrareg.RespondError(w, http.StatusInternalServerError, "Failed to retrieve GPG keys")
		return
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

	// Get the specific GPG key
	gpgKey, err := h.getGPGKeyQuery.Execute(ctx, namespace, keyID)
	if err != nil {
		terrareg.RespondError(w, http.StatusInternalServerError, "Failed to retrieve GPG key")
		return
	}

	if gpgKey == nil {
		terrareg.RespondError(w, http.StatusNotFound, "GPG key not found")
		return
	}

	// Build response following Python terrareg format
	response := map[string]interface{}{
		"data": gpgKey,
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
				Namespace      string  `json:"namespace"`
				ASCIILArmor    string  `json:"ascii-armor"`
				TrustSignature *string `json:"trust-signature,omitempty"`
				Source         *string `json:"source,omitempty"`
				SourceURL      *string `json:"source-url,omitempty"`
			} `json:"attributes"`
		} `json:"data"`
		CSRFToken string `json:"csrf_token"`
	}

	if err := json.NewDecoder(r.Body).Decode(&createRequest); err != nil {
		terrareg.RespondError(w, http.StatusBadRequest, "Invalid JSON request body")
		return
	}

	// Validate required fields
	if createRequest.Data.Attributes.Namespace == "" {
		terrareg.RespondError(w, http.StatusBadRequest, "Missing required field: namespace")
		return
	}
	if createRequest.Data.Attributes.ASCIILArmor == "" {
		terrareg.RespondError(w, http.StatusBadRequest, "Missing required field: ascii-armor")
		return
	}

	// Convert to command request
	cmdRequest := gpgkeyCmd.CreateGPGKeyRequest{
		Namespace:     createRequest.Data.Attributes.Namespace,
		ASCIILArmor:   createRequest.Data.Attributes.ASCIILArmor,
		TrustSignature: createRequest.Data.Attributes.TrustSignature,
		Source:        createRequest.Data.Attributes.Source,
		SourceURL:     createRequest.Data.Attributes.SourceURL,
	}

	// Execute command
	response, err := h.manageGPGKeyCmd.CreateGPGKey(r.Context(), cmdRequest)
	if err != nil {
		if strings.Contains(err.Error(), "does not exist") {
			terrareg.RespondError(w, http.StatusNotFound, err.Error())
			return
		}
		if strings.Contains(err.Error(), "duplicate") {
			terrareg.RespondError(w, http.StatusConflict, err.Error())
			return
		}
		if strings.Contains(err.Error(), "invalid") {
			terrareg.RespondError(w, http.StatusBadRequest, err.Error())
			return
		}
		terrareg.RespondError(w, http.StatusInternalServerError, "Failed to create GPG key")
		return
	}

	// Build response following Python terrareg format
	apiResponse := map[string]interface{}{
		"data": response,
	}

	terrareg.RespondJSON(w, http.StatusCreated, apiResponse)
}

// HandleDeleteGPGKey handles DELETE /v2/gpg-keys/{namespace}/{key_id}
func (h *TerraformV2GPGHandler) HandleDeleteGPGKey(w http.ResponseWriter, r *http.Request) {
	namespace := chi.URLParam(r, "namespace")
	keyID := chi.URLParam(r, "key_id")

	if namespace == "" || keyID == "" {
		terrareg.RespondError(w, http.StatusBadRequest, "Missing required parameters: namespace and key_id")
		return
	}

	// Convert to command request
	cmdRequest := gpgkeyCmd.DeleteGPGKeyRequest{
		Namespace: namespace,
		KeyID:     keyID,
	}

	// Execute command
	err := h.manageGPGKeyCmd.DeleteGPGKey(r.Context(), cmdRequest)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			terrareg.RespondError(w, http.StatusNotFound, err.Error())
			return
		}
		if strings.Contains(err.Error(), "in use") {
			terrareg.RespondError(w, http.StatusConflict, err.Error())
			return
		}
		terrareg.RespondError(w, http.StatusInternalServerError, "Failed to delete GPG key")
		return
	}

	// Return empty response with 204 No Content
	w.WriteHeader(http.StatusNoContent)
}
