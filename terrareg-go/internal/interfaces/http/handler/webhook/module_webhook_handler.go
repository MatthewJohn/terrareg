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
		if len(h.uploadAPIKeys) > 0 {
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

		// Return success response
		if result.Success {
			terrareg.RespondJSON(w, http.StatusOK, map[string]interface{}{
				"status":  "success",
				"message": result.Message,
			})
		} else {
			terrareg.RespondJSON(w, http.StatusBadRequest, map[string]interface{}{
				"status":  "error",
				"message": result.Message,
			})
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
	Push struct {
		Changes []struct {
			Type string `json:"type"`
			Old  struct {
				Type string `json:"type"`
			} `json:"old"`
			New struct {
				Type string `json:"type"`
				Name string `json:"name"`
			} `json:"new"`
		} `json:"changes"`
	} `json:"push"`
	Repository struct {
		FullName string `json:"full_name"`
	} `json:"repository"`
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

	// Only process release events (matching Python behavior)
	if payload.Action != "published" && payload.Action != "created" {
		return &moduleService.WebhookResult{
			Success: true,
			Message: fmt.Sprintf("Ignoring non-release action: %s", payload.Action),
		}, nil
	}

	// Extract version from tag
	version := payload.Release.TagName
	if version == "" {
		return &moduleService.WebhookResult{
			Success: false,
			Message: "No tag found in release payload",
		}, nil
	}

	// Trigger module version creation for the tag
	fmt.Printf("GitHub webhook: Processing release %s for module %s/%s/%s\n", version, namespace, moduleName, provider)

	return h.triggerModuleVersionCreation(ctx, namespace, moduleName, provider, version)
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

	// Collect all versions to process
	var versionRequests []moduleService.VersionImportRequest
	for _, change := range payload.Push.Changes {
		// Only process TAG type changes with ADD or UPDATE
		if change.Type != "TAG" {
			continue
		}

		if change.Old.Type != "" && change.New.Type == "" {
			// Tag deletion - ignore for now
			continue
		}

		if change.New.Type == "TAG" && change.New.Name != "" {
			version := change.New.Name
			importRequest := module.ImportModuleVersionRequest{
				Namespace: namespace,
				Module:    moduleName,
				Provider:  provider,
				GitTag:    &version,
			}
			versionRequests = append(versionRequests, moduleService.VersionImportRequest{
				Version: version,
				Request: importRequest,
			})
		}
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
		return &moduleService.WebhookResult{
			Success: false,
			Message: multiResult.FailureSummary,
		}, nil
	} else {
		return &moduleService.WebhookResult{
			Success: true,
			Message: fmt.Sprintf("Successfully processed all %d versions", multiResult.SuccessCount),
		}, nil
	}
}

// triggerModuleVersionCreation triggers module version creation (matching Python workflow)
func (h *ModuleWebhookHandler) triggerModuleVersionCreation(ctx context.Context, namespace, moduleName, provider, version string) (*moduleService.WebhookResult, error) {
	// Integrate with the module service to:
	// 1. Get the module provider and validate git tag format regex
	// 2. Create ImportModuleVersionRequest
	// 3. Call webhookService.CreateModuleVersionFromTag()

	result, err := h.webhookService.CreateModuleVersionFromTag(ctx, namespace, moduleName, provider, version)
	if err != nil {
		return &moduleService.WebhookResult{
			Success: false,
			Message: fmt.Sprintf("Failed to create module version: %v", err),
		}, nil
	}

	return result, nil
}
