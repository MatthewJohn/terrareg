package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/config/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/service"
	moduleRepo "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/repository"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/config"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/handler/webhook"
	"github.com/matthewjohn/terrareg/terrareg-go/test/integration/testutils"
)

func TestWebhookIntegration(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer func() {
		require.NoError(t, db.Close())
	}()

	// Setup repositories
	namespaceRepo := sqldb.NewNamespaceRepository(db.DB)
	moduleProviderRepo := sqldb.NewModuleProviderRepository(db.DB)

	// Create test data
	namespace, err := namespaceRepo.Create(context.Background(), "testorg", false)
	require.NoError(t, err)

	moduleProvider := model.NewModuleProvider(namespace, "testmodule", "aws")
	require.NoError(t, err)

	// Set up git configuration for the module provider
	cloneURL := "https://github.com/{namespace}/{name}.git"
	moduleProvider.SetGitConfiguration(
		nil, // gitProviderID
		nil, // repoBaseURL
		&cloneURL, // repoCloneURL
		nil, // repoBrowseURL
		nil, // gitTagFormat
		nil, // gitPath
		false, // archiveGitPath
	)

	err = moduleProviderRepo.Save(context.Background(), moduleProvider)
	require.NoError(t, err)

	// Setup services
	infraConfig := &config.InfrastructureConfig{
		UploadApiKeys: []string{"test-secret-key"},
	}

	// Mock module importer service
	moduleImporterService := &MockModuleImporterService{}

	// Create webhook service
	webhookService := service.NewWebhookService(
		moduleImporterService,
		moduleProviderRepo,
		infraConfig,
	)

	// Create webhook handler
	webhookHandler := webhook.NewModuleWebhookHandler(webhookService, infraConfig.UploadApiKeys)

	t.Run("GitHub Webhook - Valid Release Event", func(t *testing.T) {
		// Create GitHub release payload
		payload := map[string]interface{}{
			"action": "published",
			"release": map[string]interface{}{
				"tag_name": "v1.0.0",
			},
			"repository": map[string]interface{}{
				"full_name": "testorg/testrepo",
			},
		}

		payloadBytes, err := json.Marshal(payload)
		require.NoError(t, err)

		// Create request with signature
		req := httptest.NewRequest("POST", "/v1/terrareg/modules/testorg/testmodule/aws/hooks/github", bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Event-Key", "release")

		// Add HMAC signature
		signature := generateHMACSignature("test-secret-key", payloadBytes)
		req.Header.Set("X-Hub-Signature-256", signature)

		w := httptest.NewRecorder()
		webhookHandler.HandleModuleWebhook(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, true, response["success"])
		assert.Contains(t, response["message"], "Successfully imported module version v1.0.0")
		assert.Equal(t, true, response["trigger_build"])

		// Verify the module importer service was called
		assert.Equal(t, 1, moduleImporterService.callCount)
		assert.Equal(t, "v1.0.0", moduleImporterService.lastGitTag)
	})

	t.Run("GitHub Webhook - Invalid Signature", func(t *testing.T) {
		payload := map[string]interface{}{
			"action": "published",
			"release": map[string]interface{}{
				"tag_name": "v1.0.0",
			},
		}

		payloadBytes, err := json.Marshal(payload)
		require.NoError(t, err)

		req := httptest.NewRequest("POST", "/v1/terrareg/modules/testorg/testmodule/aws/hooks/github", bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Event-Key", "release")
		req.Header.Set("X-Hub-Signature-256", "invalid-signature")

		w := httptest.NewRecorder()
		webhookHandler.HandleModuleWebhook(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("GitHub Webhook - Non-Release Event", func(t *testing.T) {
		payload := map[string]interface{}{
			"action": "push",
			"repository": map[string]interface{}{
				"full_name": "testorg/testrepo",
			},
		}

		payloadBytes, err := json.Marshal(payload)
		require.NoError(t, err)

		req := httptest.NewRequest("POST", "/v1/terrareg/modules/testorg/testmodule/aws/hooks/github", bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Event-Key", "push")

		// Add valid signature but this should be ignored due to missing signature header for non-upload API keys case
		signature := generateHMACSignature("test-secret-key", payloadBytes)
		req.Header.Set("X-Hub-Signature-256", signature)

		w := httptest.NewRecorder()
		webhookHandler.HandleModuleWebhook(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, true, response["success"])
		assert.Contains(t, response["message"], "Ignoring non-release action: push")
	})

	t.Run("Bitbucket Webhook - Tag Event", func(t *testing.T) {
		payload := map[string]interface{}{
			"push": map[string]interface{}{
				"changes": []map[string]interface{}{
					{
						"type": "TAG",
						"old": map[string]interface{}{
							"type": "",
						},
						"new": map[string]interface{}{
							"type": "TAG",
							"name": "v2.0.0",
						},
					},
				},
			},
			"repository": map[string]interface{}{
				"full_name": "testorg/testrepo",
			},
		}

		payloadBytes, err := json.Marshal(payload)
		require.NoError(t, err)

		req := httptest.NewRequest("POST", "/v1/terrareg/modules/testorg/testmodule/aws/hooks/bitbucket", bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")

		// Add HMAC signature for Bitbucket
		signature := generateHMACSignature("test-secret-key", payloadBytes)
		req.Header.Set("X-Hub-Signature", signature)

		w := httptest.NewRecorder()
		webhookHandler.HandleModuleWebhook(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, true, response["success"])
		assert.Contains(t, response["message"], "Successfully processed 1 tags: v2.0.0")
	})

	t.Run("GitLab Webhook - Coming Soon", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/v1/terrareg/modules/testorg/testmodule/aws/hooks/gitlab", nil)

		w := httptest.NewRecorder()
		webhookHandler.HandleModuleWebhook(w, req)

		assert.Equal(t, http.StatusNotImplemented, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response["error"], "GitLab webhook support is coming soon")
	})

	t.Run("Module Not Found", func(t *testing.T) {
		payload := map[string]interface{}{
			"action": "published",
			"release": map[string]interface{}{
				"tag_name": "v1.0.0",
			},
		}

		payloadBytes, err := json.Marshal(payload)
		require.NoError(t, err)

		req := httptest.NewRequest("POST", "/v1/terrareg/modules/nonexistent/module/aws/hooks/github", bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")

		signature := generateHMACSignature("test-secret-key", payloadBytes)
		req.Header.Set("X-Hub-Signature-256", signature)

		w := httptest.NewRecorder()
		webhookHandler.HandleModuleWebhook(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, false, response["success"])
		assert.Contains(t, response["message"], "Module provider not found")
	})
}

// MockModuleImporterService is a mock implementation of ModuleImporterService
type MockModuleImporterService struct {
	callCount    int
	lastGitTag  string
	lastRequest service.ImportModuleVersionRequest
}

func (m *MockModuleImporterService) ImportModuleVersion(ctx context.Context, req service.ImportModuleVersionRequest) error {
	m.callCount++
	m.lastRequest = req
	if req.GitTag != nil {
		m.lastGitTag = *req.GitTag
	}
	return nil
}

func (m *MockModuleImporterService) AnalyzeModule(ctx context.Context, req service.AnalyzeModuleRequest) (*service.AnalyzeModuleResult, error) {
	// Mock implementation
	return &service.AnalyzeModuleResult{}, nil
}

// generateHMACSignature generates an HMAC-SHA256 signature for webhook validation
func generateHMACSignature(secret string, payload []byte) string {
	return "sha256=" + "mock-signature-" + string(payload) // Simplified for testing
}