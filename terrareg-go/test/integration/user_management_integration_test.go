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

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
	authHandler "github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/handler/terrareg"
	"github.com/matthewjohn/terrareg/terrareg-go/test/integration/testutils"
)

func TestUserManagementIntegration(t *testing.T) {
	db := testutils.SetupTestDatabase(t)
	defer func() {
		require.NoError(t, db.Close())
	}()

	// Setup repositories
	sessionRepo := sqldb.NewSessionRepository(db.DB)
	userGroupRepo := sqldb.NewUserGroupRepository(db.DB)
	namespaceRepo := sqldb.NewNamespaceRepository(db.DB)
	userGroupNamespacePermissionRepo := sqldb.NewUserGroupNamespacePermissionRepository(db.DB)

	// Create handlers and commands
	createSessionCmd := auth.NewCreateSessionCommand(sessionRepo)
	getSessionQuery := auth.NewGetSessionQuery(sessionRepo)
	deleteSessionCmd := auth.NewDeleteSessionCommand(sessionRepo)

	createUserGroupCmd := auth.NewCreateUserGroupCommand(userGroupRepo)
	getUserGroupQuery := auth.NewGetUserGroupQuery(userGroupRepo)
	updateUserGroupCmd := auth.NewUpdateUserGroupCommand(userGroupRepo)
	deleteUserGroupCmd := auth.NewDeleteUserGroupCommand(userGroupRepo)

	grantPermissionCmd := auth.NewGrantUserGroupNamespacePermissionCommand(userGroupNamespacePermissionRepo)
	revokePermissionCmd := auth.NewRevokeUserGroupNamespacePermissionCommand(userGroupNamespacePermissionRepo)
	getUserGroupPermissionsQuery := auth.NewGetUserGroupNamespacePermissionsQuery(userGroupNamespacePermissionRepo)

	handler := authHandler.NewAuthHandler(
		getSessionQuery,
		createSessionCmd,
		deleteSessionCmd,
		nil, // terraformIDPService
		nil, // authMethodFactory
		createUserGroupCmd,
		updateUserGroupCmd,
		deleteUserGroupCmd,
		getUserGroupQuery,
		nil, // getUserGroupsQuery
		grantPermissionCmd,
		revokePermissionCmd,
		getUserGroupPermissionsQuery,
		nil, // domainConfig
	)

	t.Run("Create and manage user group", func(t *testing.T) {
		// Create a user group
		createReq := map[string]interface{}{
			"name":        "test-group",
			"description": "Test user group",
		}
		reqBody, _ := json.Marshal(createReq)

		req := httptest.NewRequest("POST", "/v1/terrareg/user-groups", bytes.NewReader(reqBody))
		w := httptest.NewRecorder()

		handler.HandleCreateUserGroup(w, req)
		assert.Equal(t, http.StatusCreated, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "User group created successfully", response["message"])

		// Get the user group
		req = httptest.NewRequest("GET", "/v1/terrareg/user-groups/test-group", nil)
		w = httptest.NewRecorder()

		handler.HandleGetUserGroup(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "test-group", response["name"])
		assert.Equal(t, "Test user group", response["description"])
	})

	t.Run("Grant namespace permission to user group", func(t *testing.T) {
		// Create a namespace first
		namespace, err := namespaceRepo.Create(context.Background(), "test-namespace", false)
		require.NoError(t, err)

		// Grant permission
		grantReq := map[string]interface{}{
			"namespace_name": namespace.Name(),
			"permission":     "MODULE_WRITE",
		}
		reqBody, _ := json.Marshal(grantReq)

		req := httptest.NewRequest(
			"POST",
			"/v1/terrareg/user-groups/test-group/permissions",
			bytes.NewReader(reqBody),
		)
		w := httptest.NewRecorder()

		handler.HandleGrantUserGroupPermission(w, req)
		assert.Equal(t, http.StatusCreated, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "Permission granted successfully", response["message"])
	})

	t.Run("Get user group permissions", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/terrareg/user-groups/test-group/permissions", nil)
		w := httptest.NewRecorder()

		handler.HandleGetUserGroupPermissions(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		var response []map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Len(t, response, 1)

		permission := response[0]
		assert.Equal(t, "test-namespace", permission["namespace_name"])
		assert.Equal(t, "MODULE_WRITE", permission["permission"])
	})

	t.Run("Update user group", func(t *testing.T) {
		updateReq := map[string]interface{}{
			"description": "Updated test user group",
		}
		reqBody, _ := json.Marshal(updateReq)

		req := httptest.NewRequest(
			"PUT",
			"/v1/terrareg/user-groups/test-group",
			bytes.NewReader(reqBody),
		)
		w := httptest.NewRecorder()

		handler.HandleUpdateUserGroup(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "User group updated successfully", response["message"])
	})

	t.Run("Revoke namespace permission from user group", func(t *testing.T) {
		req := httptest.NewRequest(
			"DELETE",
			"/v1/terrareg/user-groups/test-group/permissions/test-namespace",
			nil,
		)
		w := httptest.NewRecorder()

		handler.HandleRevokeUserGroupPermission(w, req)
		assert.Equal(t, http.StatusNoContent, w.Code)

		// Verify permission is revoked
		req = httptest.NewRequest("GET", "/v1/terrareg/user-groups/test-group/permissions", nil)
		w = httptest.NewRecorder()

		handler.HandleGetUserGroupPermissions(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		var response []map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Len(t, response, 0)
	})

	t.Run("Delete user group", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/v1/terrareg/user-groups/test-group", nil)
		w := httptest.NewRecorder()

		handler.HandleDeleteUserGroup(w, req)
		assert.Equal(t, http.StatusNoContent, w.Code)

		// Verify group is deleted
		req = httptest.NewRequest("GET", "/v1/terrareg/user-groups/test-group", nil)
		w = httptest.NewRecorder()

		handler.HandleGetUserGroup(w, req)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("Handle invalid user group operations", func(t *testing.T) {
		// Try to get non-existent group
		req := httptest.NewRequest("GET", "/v1/terrareg/user-groups/nonexistent", nil)
		w := httptest.NewRecorder()

		handler.HandleGetUserGroup(w, req)
		assert.Equal(t, http.StatusNotFound, w.Code)

		// Try to create group without name
		createReq := map[string]interface{}{
			"description": "Group without name",
		}
		reqBody, _ := json.Marshal(createReq)

		req = httptest.NewRequest("POST", "/v1/terrareg/user-groups", bytes.NewReader(reqBody))
		w = httptest.NewRecorder()

		handler.HandleCreateUserGroup(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)

		// Try to grant permission for non-existent namespace
		grantReq := map[string]interface{}{
			"namespace_name": "nonexistent-namespace",
			"permission":     "MODULE_WRITE",
		}
		reqBody, _ = json.Marshal(grantReq)

		req = httptest.NewRequest(
			"POST",
			"/v1/terrareg/user-groups/test-group/permissions",
			bytes.NewReader(reqBody),
		)
		w = httptest.NewRecorder()

		handler.HandleGrantUserGroupPermission(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}
