package integration

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	modulemodel "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/model"
	moduleRepo "github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb/module"
	"github.com/matthewjohn/terrareg/terrareg-go/test/integration/testutils"
)

const testUploadAPIKey = "test-upload-key" // Matches CreateTestInfraConfig

// TestGitHubWebhookIntegration tests GitHub webhook processing following Python pattern
func TestGitHubWebhookIntegration(t *testing.T) {
	// Change to terrareg-go directory for template loading
	oldWd, _ := os.Getwd()
	err := os.Chdir("/app/terrareg-go")
	require.NoError(t, err)
	defer func() {
		os.Chdir(oldWd)
	}()

	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Get the real server router from the container
	cont := testutils.CreateTestContainer(t, db)
	router := cont.Server.Router()

	// Setup repositories
	namespaceRepo := moduleRepo.NewNamespaceRepository(db.DB)
	moduleProviderRepo := moduleRepo.NewModuleProviderRepository(db.DB, namespaceRepo, nil)

	ctx := context.Background()

	// Create test data following Python pattern
	namespace, err := modulemodel.NewNamespace("test-namespace", nil, modulemodel.NamespaceTypeNone)
	require.NoError(t, err)
	err = namespaceRepo.Save(ctx, namespace)
	require.NoError(t, err)

	moduleProvider, err := modulemodel.NewModuleProvider(namespace, "test-module", "aws")
	require.NoError(t, err)

	// Configure git repository URL (required for webhooks)
	cloneURL := "https://github.com/testorg/test-module.git"
	moduleProvider.SetGitConfiguration(
		nil,       // gitProviderID
		nil,       // repoBaseURL
		&cloneURL, // repoCloneURL
		nil,       // repoBrowseURL
		nil,       // gitTagFormat
		nil,       // gitPath
		false,     // archiveGitPath
	)

	err = moduleProviderRepo.Save(ctx, moduleProvider)
	require.NoError(t, err)

	t.Run("test_github_webhook_with_published_release", func(t *testing.T) {
		// Create GitHub release webhook payload (following Python structure)
		payload := map[string]interface{}{
			"action": "published",
			"release": map[string]interface{}{
				"tag_name": "v1.0.0",
				"url":      "https://api.github.com/repos/testorg/test-module/releases/v1.0.0",
				"html_url": "https://github.com/testorg/test-module/releases/tag/v1.0.0",
			},
			"repository": map[string]interface{}{
				"id":       118,
				"name":     "test-module",
				"full_name": "testorg/test-module",
			},
			"sender": map[string]interface{}{
				"login": "testuser",
			},
		}

		payloadBytes, err := json.Marshal(payload)
		require.NoError(t, err)

		// Generate valid HMAC signature using the test config's upload API key
		signature := generateHMACSignature(testUploadAPIKey, payloadBytes)

		// Create request using the real server's route
		req := httptest.NewRequest(
			"POST",
			"/v1/terrareg/modules/test-namespace/test-module/aws/hooks/github",
			bytes.NewReader(payloadBytes),
		)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-GitHub-Event", "release")
		req.Header.Set("X-Hub-Signature-256", signature)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Webhook endpoint returns 400 if processing fails (matching Python behavior)
		// The webhook signature was validated (that passed), but the import will fail
		// because we don't have a real git repository in tests
		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// Response should have status field set to "error"
		assert.Contains(t, response, "status")
		assert.Equal(t, "error", response["status"])
		assert.Contains(t, response, "message")
	})

	t.Run("test_github_webhook_with_created_release", func(t *testing.T) {
		// Test with "created" action (also valid)
		payload := map[string]interface{}{
			"action": "created",
			"release": map[string]interface{}{
				"tag_name": "v1.1.0",
			},
			"repository": map[string]interface{}{
				"full_name": "testorg/test-module",
			},
		}

		payloadBytes, _ := json.Marshal(payload)
		signature := generateHMACSignature(testUploadAPIKey, payloadBytes)

		req := httptest.NewRequest(
			"POST",
			"/v1/terrareg/modules/test-namespace/test-module/aws/hooks/github",
			bytes.NewReader(payloadBytes),
		)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-GitHub-Event", "release")
		req.Header.Set("X-Hub-Signature-256", signature)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Webhook processing will fail (no real git repo) - expect 400
		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "error", response["status"])
	})

	t.Run("test_github_webhook_ignores_non_release_actions", func(t *testing.T) {
		// Test that non-release actions are ignored (following Python behavior)
		payload := map[string]interface{}{
			"action": "edited",
			"release": map[string]interface{}{
				"tag_name": "v1.2.0",
			},
			"repository": map[string]interface{}{
				"full_name": "testorg/test-module",
			},
		}

		payloadBytes, _ := json.Marshal(payload)
		signature := generateHMACSignature(testUploadAPIKey, payloadBytes)

		req := httptest.NewRequest(
			"POST",
			"/v1/terrareg/modules/test-namespace/test-module/aws/hooks/github",
			bytes.NewReader(payloadBytes),
		)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-GitHub-Event", "release")
		req.Header.Set("X-Hub-Signature-256", signature)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)

		// Should return success but indicate it was ignored
		assert.Equal(t, "success", response["status"])
	})

	t.Run("test_github_webhook_with_invalid_signature", func(t *testing.T) {
		payload := map[string]interface{}{
			"action": "published",
			"release": map[string]interface{}{
				"tag_name": "v1.3.0",
			},
		}

		payloadBytes, _ := json.Marshal(payload)

		req := httptest.NewRequest(
			"POST",
			"/v1/terrareg/modules/test-namespace/test-module/aws/hooks/github",
			bytes.NewReader(payloadBytes),
		)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Hub-Signature-256", "sha256=invalid_signature")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Should return unauthorized for invalid signature
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("test_github_webhook_with_missing_signature", func(t *testing.T) {
		payload := map[string]interface{}{
			"action": "published",
			"release": map[string]interface{}{
				"tag_name": "v1.4.0",
			},
		}

		payloadBytes, _ := json.Marshal(payload)

		req := httptest.NewRequest(
			"POST",
			"/v1/terrareg/modules/test-namespace/test-module/aws/hooks/github",
			bytes.NewReader(payloadBytes),
		)
		req.Header.Set("Content-Type", "application/json")
		// Don't set signature header

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Should return unauthorized for missing signature
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("test_github_webhook_with_missing_tag_name", func(t *testing.T) {
		// Test webhook with missing tag_name
		payload := map[string]interface{}{
			"action": "published",
			"release": map[string]interface{}{
				"url": "https://api.github.com/repos/testorg/test-module/releases/v1.0.0",
			},
		}

		payloadBytes, _ := json.Marshal(payload)
		signature := generateHMACSignature(testUploadAPIKey, payloadBytes)

		req := httptest.NewRequest(
			"POST",
			"/v1/terrareg/modules/test-namespace/test-module/aws/hooks/github",
			bytes.NewReader(payloadBytes),
		)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Hub-Signature-256", signature)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Python returns 400 when tag_name is missing
		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)

		// Should return error status
		assert.Equal(t, "error", response["status"])
		assert.Contains(t, response["message"], "No tag found")
	})
}

// TestBitbucketWebhookIntegration tests Bitbucket webhook processing following Python pattern
func TestBitbucketWebhookIntegration(t *testing.T) {
	// Change to terrareg-go directory for template loading
	oldWd, _ := os.Getwd()
	err := os.Chdir("/app/terrareg-go")
	require.NoError(t, err)
	defer func() {
		os.Chdir(oldWd)
	}()

	db := testutils.SetupTestDatabase(t)
	defer testutils.CleanupTestDatabase(t, db)

	// Get the real server router from the container
	cont := testutils.CreateTestContainer(t, db)
	router := cont.Server.Router()

	// Setup repositories
	namespaceRepo := moduleRepo.NewNamespaceRepository(db.DB)
	moduleProviderRepo := moduleRepo.NewModuleProviderRepository(db.DB, namespaceRepo, nil)

	ctx := context.Background()

	// Create test data
	namespace, err := modulemodel.NewNamespace("bitbucket-test", nil, modulemodel.NamespaceTypeNone)
	require.NoError(t, err)
	err = namespaceRepo.Save(ctx, namespace)
	require.NoError(t, err)

	moduleProvider, err := modulemodel.NewModuleProvider(namespace, "bb-module", "aws")
	require.NoError(t, err)

	cloneURL := "https://bitbucket.org/testorg/bb-module.git"
	moduleProvider.SetGitConfiguration(
		nil,       // gitProviderID
		nil,       // repoBaseURL
		&cloneURL, // repoCloneURL
		nil,       // repoBrowseURL
		nil,       // gitTagFormat
		nil,       // gitPath
		false,     // archiveGitPath
	)

	err = moduleProviderRepo.Save(ctx, moduleProvider)
	require.NoError(t, err)

	t.Run("test_bitbucket_webhook_with_single_tag", func(t *testing.T) {
		// Create Bitbucket push webhook payload (following Python structure)
		payload := map[string]interface{}{
			"eventKey": "repo:refs_changed",
			"date":     "2022-04-23T21:21:46+0000",
			"actor": map[string]interface{}{
				"name":         "admin",
				"emailAddress": "admin@localhost",
				"displayName":  "Administrator",
				"id":           1,
				"type":         "normal",
				"active":       true,
			},
			"repository": map[string]interface{}{
				"slug":  "bb-module",
				"id":    1,
				"name":  "bb-module",
				"scmId": "git",
			},
			"changes": []map[string]interface{}{
				{
					"ref": map[string]interface{}{
						"id":         "refs/tags/v5.1.2",
						"displayId":  "v5.1.2",
						"type":       "TAG",
					},
					"refId":    "refs/tags/v5.1.2",
					"fromHash": "0000000000000000000000000000000000000000",
					"toHash":   "1097d939669e3209ff33e6dfe982d84c204f6087",
					"type":     "ADD",
				},
			},
		}

		payloadBytes, _ := json.Marshal(payload)
		signature := generateHMACSignature(testUploadAPIKey, payloadBytes)

		req := httptest.NewRequest(
			"POST",
			"/v1/terrareg/modules/bitbucket-test/bb-module/aws/hooks/bitbucket",
			bytes.NewReader(payloadBytes),
		)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Event-Key", "repo:refs_changed")
		req.Header.Set("X-Hub-Signature", signature)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)

		assert.Contains(t, response, "status")
	})

	t.Run("test_bitbucket_webhook_with_multiple_tags", func(t *testing.T) {
		// Test multiple tag processing (following Python pattern)
		payload := map[string]interface{}{
			"eventKey": "repo:refs_changed",
			"repository": map[string]interface{}{
				"slug": "bb-module",
				"name": "bb-module",
			},
			"changes": []map[string]interface{}{
				{
					"ref": map[string]interface{}{
						"type": "TAG",
						"name": "v6.0.0",
					},
					"type": "ADD",
				},
				{
					"ref": map[string]interface{}{
						"type": "TAG",
						"name": "v6.1.0",
					},
					"type": "ADD",
				},
			},
		}

		payloadBytes, _ := json.Marshal(payload)
		signature := generateHMACSignature(testUploadAPIKey, payloadBytes)

		req := httptest.NewRequest(
			"POST",
			"/v1/terrareg/modules/bitbucket-test/bb-module/aws/hooks/bitbucket",
			bytes.NewReader(payloadBytes),
		)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Hub-Signature", signature)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)

		assert.Contains(t, response, "status")
	})

	t.Run("test_bitbucket_webhook_ignores_non_tag_changes", func(t *testing.T) {
		// Test that non-TAG changes are ignored
		payload := map[string]interface{}{
			"eventKey": "repo:refs_changed",
			"repository": map[string]interface{}{
				"slug": "bb-module",
			},
			"changes": []map[string]interface{}{
				{
					"ref": map[string]interface{}{
						"type": "BRANCH",
						"name": "main",
					},
					"type": "UPDATE",
				},
			},
		}

		payloadBytes, _ := json.Marshal(payload)
		signature := generateHMACSignature(testUploadAPIKey, payloadBytes)

		req := httptest.NewRequest(
			"POST",
			"/v1/terrareg/modules/bitbucket-test/bb-module/aws/hooks/bitbucket",
			bytes.NewReader(payloadBytes),
		)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Hub-Signature", signature)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)

		// Should indicate no tags to process
		assert.Equal(t, "success", response["status"])
	})

	t.Run("test_bitbucket_webhook_ignores_tag_deletion", func(t *testing.T) {
		// Test that tag deletions are ignored
		payload := map[string]interface{}{
			"eventKey": "repo:refs_changed",
			"repository": map[string]interface{}{
				"slug": "bb-module",
			},
			"changes": []map[string]interface{}{
				{
					"old": map[string]interface{}{
						"type": "TAG",
					},
					"new": map[string]interface{}{
						"type": "",
					},
					"type": "UPDATE",
				},
			},
		}

		payloadBytes, _ := json.Marshal(payload)
		signature := generateHMACSignature(testUploadAPIKey, payloadBytes)

		req := httptest.NewRequest(
			"POST",
			"/v1/terrareg/modules/bitbucket-test/bb-module/aws/hooks/bitbucket",
			bytes.NewReader(payloadBytes),
		)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Hub-Signature", signature)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)

		assert.Equal(t, "success", response["status"])
	})

	t.Run("test_bitbucket_webhook_with_invalid_signature", func(t *testing.T) {
		payload := map[string]interface{}{
			"eventKey": "repo:refs_changed",
			"repository": map[string]interface{}{
				"slug": "bb-module",
			},
			"changes": []map[string]interface{}{
				{
					"ref": map[string]interface{}{
						"type": "TAG",
						"name": "v5.0.0",
					},
					"type": "ADD",
				},
			},
		}

		payloadBytes, _ := json.Marshal(payload)

		req := httptest.NewRequest(
			"POST",
			"/v1/terrareg/modules/bitbucket-test/bb-module/aws/hooks/bitbucket",
			bytes.NewReader(payloadBytes),
		)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Hub-Signature", "invalid_signature")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

// Helper function to generate HMAC-SHA256 signature (following Python pattern)
func generateHMACSignature(apiKey string, payload []byte) string {
	hasher := hmac.New(sha256.New, []byte(apiKey))
	hasher.Write(payload)
	digest := hasher.Sum(nil)
	return fmt.Sprintf("sha256=%s", hex.EncodeToString(digest))
}
