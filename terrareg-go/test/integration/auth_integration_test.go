package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	authQuery "github.com/matthewjohn/terrareg/terrareg-go/internal/application/query/auth"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth/service"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/config"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
	authHandler "github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/handler/terrareg"
	"github.com/matthewjohn/terrareg/terrareg-go/test/integration/testutils"
)

func TestAuthenticationIntegration(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer func() {
		require.NoError(t, db.Close())
	}()

	// Setup repositories
	sessionRepo := sqldb.NewSessionRepository(db.DB)
	userGroupRepo := sqldb.NewUserGroupRepository(db.DB)
	namespaceRepo := sqldb.NewNamespaceRepository(db.DB)
	userGroupNamespacePermissionRepo := sqldb.NewUserGroupNamespacePermissionRepository(db.DB)
	terraformIDPRepo := sqldb.NewTerraformIDPRepository(db.DB)

	// Create minimal services for testing
	sessionConfig := service.DefaultSessionDatabaseConfig()
	sessionService := service.NewSessionService(sessionRepo, sessionConfig)
	checkSessionQuery := authQuery.NewCheckSessionQuery(sessionRepo)

	// Create basic infrastructure config
	infraConfig := &config.InfrastructureConfig{
		ListenPort: 3000,
		PublicURL:  "http://localhost:3000",
		DomainName: "localhost",
		Debug:      true,
	}

	// Create auth service with minimal dependencies
	cookieService := service.NewCookieService(infraConfig)
	authService := service.NewAuthenticationService(sessionService, cookieService)

	handler := authHandler.NewAuthHandler(
		nil, // adminLoginCmd
		checkSessionQuery,
		nil, // isAuthenticatedQuery
		nil, // oidcLoginCmd
		nil, // oidcCallbackCmd
		nil, // samlLoginCmd
		nil, // samlMetadataCmd
		nil, // githubOAuthCmd
		authService,
		infraConfig,
	)

	t.Run("Create session", func(t *testing.T) {
		// Create session request
		sessionReq := map[string]interface{}{
			"auth_method": "anonymous",
		}
		reqBody, _ := json.Marshal(sessionReq)

		req := httptest.NewRequest("POST", "/v1/sessions", bytes.NewReader(reqBody))
		w := httptest.NewRecorder()

		// Add test cookies to simulate existing session
		req.AddCookie(&http.Cookie{Name: "test_session", Value: "test_value"})

		handler.HandleCreateSession(w, req)
		assert.Equal(t, http.StatusCreated, w.Code)

		// Check response has session token
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Contains(t, response, "token")
	})

	t.Run("Get session", func(t *testing.T) {
		// First create a session
		session := &model.Session{
			ID:         "test-session-id",
			AuthMethod: "anonymous",
			CreatedAt:  time.Now(),
			ExpiresAt:  time.Now().Add(24 * time.Hour),
		}

		ctx := context.Background()
		err := sessionRepo.Save(ctx, session)
		require.NoError(t, err)

		// Now get the session
		req := httptest.NewRequest("GET", "/v1/sessions/current", nil)
		req.AddCookie(&http.Cookie{Name: "terrareg_session", Value: session.ID})
		w := httptest.NewRecorder()

		handler.HandleGetSession(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, session.ID, response["id"])
		assert.Equal(t, session.AuthMethod, response["auth_method"])
	})

	t.Run("Delete session", func(t *testing.T) {
		// First create a session
		session := &model.Session{
			ID:         "test-session-id-2",
			AuthMethod: "anonymous",
			CreatedAt:  time.Now(),
			ExpiresAt:  time.Now().Add(24 * time.Hour),
		}

		ctx := context.Background()
		err := sessionRepo.Save(ctx, session)
		require.NoError(t, err)

		// Delete the session
		req := httptest.NewRequest("DELETE", "/v1/sessions/current", nil)
		req.AddCookie(&http.Cookie{Name: "terrareg_session", Value: session.ID})
		w := httptest.NewRecorder()

		handler.HandleDeleteSession(w, req)
		assert.Equal(t, http.StatusNoContent, w.Code)

		// Verify session is deleted
		_, err = sessionRepo.GetByID(ctx, session.ID)
		assert.Error(t, err)
	})

	t.Run("Handle expired session", func(t *testing.T) {
		// Create an expired session
		session := &model.Session{
			ID:         "expired-session-id",
			AuthMethod: "anonymous",
			CreatedAt:  time.Now().Add(-25 * time.Hour),
			ExpiresAt:  time.Now().Add(-1 * time.Hour), // Expired
		}

		ctx := context.Background()
		err := sessionRepo.Save(ctx, session)
		require.NoError(t, err)

		// Try to get the expired session
		req := httptest.NewRequest("GET", "/v1/sessions/current", nil)
		req.AddCookie(&http.Cookie{Name: "terrareg_session", Value: session.ID})
		w := httptest.NewRecorder()

		handler.HandleGetSession(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Contains(t, response["error"], "Session expired or not found")
	})

	t.Run("Handle missing session", func(t *testing.T) {
		// Try to get session without cookie
		req := httptest.NewRequest("GET", "/v1/sessions/current", nil)
		w := httptest.NewRecorder()

		handler.HandleGetSession(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Contains(t, response["error"], "No session found")
	})

	t.Run("Handle invalid session token", func(t *testing.T) {
		// Try to get session with invalid token
		req := httptest.NewRequest("GET", "/v1/sessions/current", nil)
		req.AddCookie(&http.Cookie{Name: "terrareg_session", Value: "invalid-token"})
		w := httptest.NewRecorder()

		handler.HandleGetSession(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Contains(t, response["error"], "Session expired or not found")
	})

	t.Run("Test session token validation", func(t *testing.T) {
		// Create a session with valid characters
		session := &model.Session{
			ID:         "valid-session-token-123",
			AuthMethod: "anonymous",
			CreatedAt:  time.Now(),
			ExpiresAt:  time.Now().Add(24 * time.Hour),
		}

		ctx := context.Background()
		err := sessionRepo.Save(ctx, session)
		require.NoError(t, err)

		// Get the session with valid token
		req := httptest.NewRequest("GET", "/v1/sessions/current", nil)
		req.AddCookie(&http.Cookie{Name: "terrareg_session", Value: session.ID})
		w := httptest.NewRecorder()

		handler.HandleGetSession(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, session.ID, response["id"])
	})

	t.Run("Test concurrent session operations", func(t *testing.T) {
		// Create multiple sessions concurrently
		sessions := make([]*model.Session, 5)
		for i := 0; i < 5; i++ {
			sessions[i] = &model.Session{
				ID:         fmt.Sprintf("concurrent-session-%d", i),
				AuthMethod: "anonymous",
				CreatedAt:  time.Now(),
				ExpiresAt:  time.Now().Add(24 * time.Hour),
			}

			ctx := context.Background()
			err := sessionRepo.Save(ctx, sessions[i])
			require.NoError(t, err)
		}

		// Verify all sessions can be retrieved
		for _, session := range sessions {
			req := httptest.NewRequest("GET", "/v1/sessions/current", nil)
			req.AddCookie(&http.Cookie{Name: "terrareg_session", Value: session.ID})
			w := httptest.NewRecorder()

			handler.HandleGetSession(w, req)
			assert.Equal(t, http.StatusOK, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)
			assert.Equal(t, session.ID, response["id"])
		}

		// Delete all sessions
		for _, session := range sessions {
			req := httptest.NewRequest("DELETE", "/v1/sessions/current", nil)
			req.AddCookie(&http.Cookie{Name: "terrareg_session", Value: session.ID})
			w := httptest.NewRecorder()

			handler.HandleDeleteSession(w, req)
			assert.Equal(t, http.StatusNoContent, w.Code)
		}
	})
}
