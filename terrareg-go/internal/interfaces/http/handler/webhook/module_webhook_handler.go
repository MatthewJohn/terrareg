package webhook

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module"
	moduleService "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/service"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared/types"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/handler/terrareg"
)

// ModuleWebhookHandler handles module webhook events for different Git providers
type ModuleWebhookHandler struct {
	webhookService *moduleService.WebhookService
	uploadAPIKeys  []string // For webhook signature validation
}

// NewModuleWebhookHandler creates a new module webhook handler
func NewModuleWebhookHandler(
	webhookService *moduleService.WebhookService,
	uploadAPIKeys []string,
) *ModuleWebhookHandler {
	return &ModuleWebhookHandler{
		webhookService: webhookService,
		uploadAPIKeys:  uploadAPIKeys,
	}
}

// HandleModuleWebhook handles module webhook events for all Git providers
func (h *ModuleWebhookHandler) HandleModuleWebhook(gitProvider string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Extract path parameters
		namespace := chi.URLParam(r, "namespace")
		moduleName := chi.URLParam(r, "name")
		provider := chi.URLParam(r, "provider")

		// Read the request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			terrareg.RespondError(w, http.StatusBadRequest, "Failed to read request body")
			return
		}

		// Validate webhook signature if upload API keys are configured
		// GitLab uses X-Gitlab-Token instead of HMAC signatures, so skip validation for GitLab
		if len(h.uploadAPIKeys) > 0 && gitProvider != "gitlab" {
			signature := r.Header.Get("X-Hub-Signature-256")
			if signature == "" {
				signature = r.Header.Get("X-Hub-Signature") // Fallback for Bitbucket
			}

			if signature == "" {
				terrareg.RespondError(w, http.StatusUnauthorized, "Missing signature header")
				return
			}

			if !h.validateSignature(body, signature) {
				terrareg.RespondError(w, http.StatusUnauthorized, "Invalid webhook signature")
				return
			}
		}

		// Route to appropriate Git provider handler
		var result *moduleService.WebhookResult
		switch gitProvider {
		case "github":
			result, err = h.processGitHubWebhook(ctx, namespace, moduleName, provider, body)
		case "bitbucket":
			result, err = h.processBitbucketWebhook(ctx, namespace, moduleName, provider, body)
		case "gitlab":
			// GitLab support marked as "coming soon" in Python version
			terrareg.RespondError(w, http.StatusNotImplemented, "GitLab webhook support is coming soon")
			return
		default:
			terrareg.RespondError(w, http.StatusBadRequest, fmt.Sprintf("Unsupported Git provider: %s", gitProvider))
			return
		}

		if err != nil {
			fmt.Printf("Webhook processing error: %v\n", err)
			terrareg.RespondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to process webhook: %s", err.Error()))
			return
		}

		// Return success response - match Python API format exactly
		// Python: {"status": "Success", "message": "...", "tag": "..."}
		if result.Success {
			response := map[string]interface{}{
				"status":  "Success", // Capitalized to match Python
				"message": result.Message,
			}
			// Add tag field if present (for GitHub webhooks)
			if result.Tag != "" {
				response["tag"] = result.Tag
			}
			// Add tags field for Bitbucket multi-version processing
			if result.Versions != nil {
				response["tags"] = result.Versions
			}
			terrareg.RespondJSON(w, http.StatusOK, response)
		} else {
			response := map[string]interface{}{
				"status":  "Error", // Capitalized to match Python
				"message": result.Message,
			}
			// Add tag field if present (for GitHub webhooks) - even in error responses
			if result.Tag != "" {
				response["tag"] = result.Tag
			}
			terrareg.RespondJSON(w, http.StatusBadRequest, response)
		}
	}
}

// validateSignature validates the webhook signature using HMAC-SHA256
func (h *ModuleWebhookHandler) validateSignature(body []byte, signature string) bool {
	// Extract hash from signature header
	parts := strings.SplitN(signature, "=", 2)
	if len(parts) != 2 || parts[0] != "sha256" {
		return false
	}

	receivedHash, err := hex.DecodeString(parts[1])
	if err != nil {
		return false
	}

	// Try each configured upload API key
	for _, apiKey := range h.uploadAPIKeys {
		// Calculate expected hash
		hasher := hmac.New(sha256.New, []byte(apiKey))
		hasher.Write(body)
		expectedHash := hasher.Sum(nil)

		// Compare hashes
		if hmac.Equal(receivedHash, expectedHash) {
			return true
		}
	}

	return false
}

// GitHub webhook payload structures (matching Python implementation)
type GitHubReleaseWebhookPayload struct {
	Action  string `json:"action"`
	Release struct {
		TagName string `json:"tag_name"`
	} `json:"release"`
	Repository struct {
		FullName string `json:"full_name"`
	} `json:"repository"`
}

// Bitbucket webhook payload structures (matching Python implementation)
type BitbucketPushWebhookPayload struct {
	Changes []struct {
		Ref struct {
			ID        string `json:"id"`        // refs/tags/v4.0.6
			DisplayID string `json:"displayId"` // v4.0.6
			Type      string `json:"type"`      // TAG
		} `json:"ref"`
		Type string `json:"type"` // ADD or UPDATE
	} `json:"changes"`
}

// processGitHubWebhook processes GitHub webhook events (release events only, matching Python)
func (h *ModuleWebhookHandler) processGitHubWebhook(ctx context.Context, namespace, moduleName, provider string, body []byte) (*moduleService.WebhookResult, error) {
	var payload GitHubReleaseWebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return &moduleService.WebhookResult{
			Success: false,
			Message: "Failed to parse GitHub payload",
		}, nil
	}

	// Extract version from tag
	tag := payload.Release.TagName
	if tag == "" {
		return &moduleService.WebhookResult{
			Success: false,
			Message: "No tag found in release payload",
		}, nil
	}

	// Handle delete/unpublished actions (matching Python behavior)
	if payload.Action == "deleted" || payload.Action == "unpublished" {
		result, err := h.deleteModuleVersion(ctx, namespace, moduleName, provider, tag)
		if err != nil {
			return result, err
		}
		// Python returns just {'status': 'Success'} for delete/unpublished (no tag field)
		return result, nil
	}

	// Only process published/created events (matching Python behavior)
	if payload.Action != "published" && payload.Action != "created" {
		return &moduleService.WebhookResult{
			Success: true,
			Message: fmt.Sprintf("Ignoring non-release action: %s", payload.Action),
			Tag:     tag, // Include tag even for ignored actions
		}, nil
	}

	// Trigger module version creation for the tag
	fmt.Printf("GitHub webhook: Processing release %s for module %s/%s/%s\n", tag, namespace, moduleName, provider)

	return h.triggerModuleVersionCreation(ctx, namespace, moduleName, provider, tag)
}

// processBitbucketWebhook processes Bitbucket webhook events with savepoint isolation (matching Python pattern)
func (h *ModuleWebhookHandler) processBitbucketWebhook(ctx context.Context, namespace, moduleName, provider string, body []byte) (*moduleService.WebhookResult, error) {
	var payload BitbucketPushWebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return &moduleService.WebhookResult{
			Success: false,
			Message: "Failed to parse Bitbucket payload",
		}, nil
	}

	// Convert to typed values
	namespaceTyped := types.NamespaceName(namespace)
	moduleNameTyped := types.ModuleName(moduleName)
	providerTyped := types.ModuleProviderName(provider)

	// Collect all versions to process
	var versionRequests []moduleService.VersionImportRequest
	for _, change := range payload.Changes {
		// Only process TAG type refs
		if change.Ref.Type != "TAG" {
			continue
		}

		// Only process ADD or UPDATE types
		if change.Type != "ADD" && change.Type != "UPDATE" {
			continue
		}

		// Extract version from ref.id (remove "refs/tags/" prefix if present)
		version := strings.TrimPrefix(change.Ref.ID, "refs/tags/")
		if version == "" {
			continue
		}

		importRequest := module.ImportModuleVersionRequest{
			Namespace: namespaceTyped,
			Module:    moduleNameTyped,
			Provider:  providerTyped,
			GitTag:    types.GitTag(version),
		}
		versionRequests = append(versionRequests, moduleService.VersionImportRequest{
			Version: version,
			Request: importRequest,
		})
	}

	if len(versionRequests) == 0 {
		return &moduleService.WebhookResult{
			Success: true,
			Message: "No tag changes found to process",
		}, nil
	}

	// Process versions with savepoint isolation using the enhanced webhook service
	multiResult, err := h.webhookService.ProcessMultipleVersionsWithSavepoints(
		ctx, namespace, moduleName, provider, versionRequests)
	if err != nil {
		return &moduleService.WebhookResult{
			Success: false,
			Message: fmt.Sprintf("Failed to process versions: %v", err),
		}, nil
	}

	// Convert MultiVersionResult to WebhookResult for backward compatibility
	if multiResult.HasFailures {
		// Convert versions map to format expected by Python API
		versions := make(map[string]interface{})
		for version, result := range multiResult.VersionsProcessed {
			versions[version] = map[string]interface{}{
				"status": result.Status,
			}
		}
		return &moduleService.WebhookResult{
			Success:  false,
			Message:  multiResult.FailureSummary,
			Versions: versions,
		}, nil
	} else {
		// Convert versions map to format expected by Python API
		versions := make(map[string]interface{})
		for version, result := range multiResult.VersionsProcessed {
			versions[version] = map[string]interface{}{
				"status": result.Status,
			}
		}
		return &moduleService.WebhookResult{
			Success:  true,
			Message:  "Imported all provided tags", // Python: "Imported all provided tags"
			Versions: versions,
		}, nil
	}
}

// triggerModuleVersionCreation triggers module version creation (matching Python workflow)
func (h *ModuleWebhookHandler) triggerModuleVersionCreation(ctx context.Context, namespace, moduleName, provider, tag string) (*moduleService.WebhookResult, error) {
	// Integrate with the module service to:
	// 1. Get the module provider and validate git tag format regex
	// 2. Create ImportModuleVersionRequest
	// 3. Call webhookService.CreateModuleVersionFromTag()

	// Convert strings to typed values
	result, err := h.webhookService.CreateModuleVersionFromTag(ctx,
		types.NamespaceName(namespace),
		types.ModuleName(moduleName),
		types.ModuleProviderName(provider),
		types.ModuleVersion(tag))
	if err != nil {
		return &moduleService.WebhookResult{
			Success: false,
			Message: fmt.Sprintf("Failed to create module version: %v", err),
			Tag:     tag,
		}, nil
	}

	// Add tag to result for Python API parity
	result.Tag = tag
	return result, nil
}

// deleteModuleVersion deletes a module version (matching Python behavior for deleted/unpublished actions)
func (h *ModuleWebhookHandler) deleteModuleVersion(ctx context.Context, namespace, moduleName, provider, version string) (*moduleService.WebhookResult, error) {
	// Call webhook service to delete module version
	result, err := h.webhookService.DeleteModuleVersion(ctx, namespace, moduleName, provider, version)
	if err != nil {
		return &moduleService.WebhookResult{
			Success: false,
			Message: fmt.Sprintf("Failed to delete module version: %v", err),
			Tag:     version,
		}, nil
	}

	// Add tag to result for Python API parity
	result.Tag = version
	return result, nil
}
