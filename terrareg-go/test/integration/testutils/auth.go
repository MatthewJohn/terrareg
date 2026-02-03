package testutils

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth/service"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/container"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
	"github.com/stretchr/testify/require"
)

// AuthHelper provides authentication setup for tests
type AuthHelper struct {
	db            *sqldb.Database
	t             *testing.T
	server        *TestServer
	container     *container.Container
	cookieService *service.CookieService
}

// TestServer represents a test server with router
type TestServer struct {
	// Router would be set by the actual container
	Router http.Handler
}

// NewAuthHelper creates a new auth helper
// Accepts either *TestServer or *testutils.TestContainer
// If server is a *TestContainer, it will extract the container and create a CookieService
func NewAuthHelper(t *testing.T, db *sqldb.Database, server interface{}) *AuthHelper {
	helper := &AuthHelper{db: db, t: t}

	// Try to extract container from TestContainer
	if tc, ok := server.(*TestContainer); ok && tc.Container != nil {
		helper.server = &TestServer{Router: tc.Router}
		helper.container = tc.Container
		// Create CookieService using the same config as the container
		// This uses the test secret key for encryption/decryption
		helper.cookieService = service.NewCookieService(tc.Container.InfraConfig)
	} else if ts, ok := server.(*TestServer); ok {
		helper.server = ts
	}

	return helper
}

// LoginAsBuiltInAdmin performs admin login and returns session cookie
// Uses SessionManagementService if available, otherwise attempts API login
func (h *AuthHelper) LoginAsBuiltInAdmin() string {
	// Prefer using SessionManagementService directly
	if h.container != nil && h.container.SessionManagementService != nil {
		ctx := context.Background()
		w := httptest.NewRecorder()

		// Use SessionManagementService to create session and cookie
		err := h.container.SessionManagementService.CreateSessionAndCookie(
			ctx,
			w,
			"ADMIN_SESSION",
			"admin",
			true, // site admin
			[]string{},
			nil, // permissions
			nil, // providerData
			nil, // ttl - use default
		)
		require.NoError(h.t, err, "Failed to create admin session")

		// Extract encrypted cookie from response
		cookies := w.Result().Cookies()
		for _, c := range cookies {
			if c.Name == "terrareg_session" {
				return fmt.Sprintf("terrareg_session=%s", c.Value)
			}
		}
		return ""
	}

	// Fallback: Try API login if router is available
	if h.server != nil && h.server.Router != nil {
		req := httptest.NewRequest("POST", "/v1/terrareg/auth/admin/login", nil)
		req.Header.Set("X-Terrareg-ApiKey", h.getAdminApiKey())

		w := httptest.NewRecorder()
		h.server.Router.ServeHTTP(w, req)

		// Extract session cookie from response
		cookies := w.Header().Get("Set-Cookie")
		if cookies != "" {
			// Parse the session cookie
			parts := strings.Split(cookies, ";")
			for _, part := range parts {
				part = strings.TrimSpace(part)
				if strings.HasPrefix(part, "terrareg_session=") {
					return part
				}
			}
		}
	}

	return ""
}

// getAdminApiKey returns the admin API key from environment or default
func (h *AuthHelper) getAdminApiKey() string {
	if key := os.Getenv("ADMIN_AUTH_TOKEN"); key != "" {
		return key
	}
	return "test-admin-api-key"
}

// getUploadApiKey returns the upload API key from environment or default
func (h *AuthHelper) getUploadApiKey() string {
	if key := os.Getenv("UPLOAD_AUTH_TOKEN"); key != "" {
		return key
	}
	return "test-upload-key"
}

// getPublishApiKey returns the publish API key from environment or default
func (h *AuthHelper) getPublishApiKey() string {
	if key := os.Getenv("PUBLISH_AUTH_TOKEN"); key != "" {
		return key
	}
	return "test-publish-key"
}

// CreateSessionForUser creates a session in the database for testing
// This bypasses the OAuth flow and directly creates a session with specified attributes
// Returns an encrypted session cookie that can be used in test requests
func (h *AuthHelper) CreateSessionForUser(username string, siteAdmin bool, userGroups []string, namespacePermissions map[string]string) string {
	// Build provider source auth JSON
	providerSourceAuth := map[string]interface{}{
		"user_id":     username,
		"username":    username,
		"email":       username + "@example.com",
		"user_groups": userGroups,
		"site_admin":  siteAdmin,
		"is_admin":    siteAdmin, // Support both field names
	}

	// Create session with proper encoding
	session := sqldb.SessionDB{
		ID:                 fmt.Sprintf("session-%s-%d", username, time.Now().UnixNano()),
		Expiry:             time.Now().Add(1 * time.Hour),
		ProviderSourceAuth: sqldb.EncodeBlob(providerSourceAuth),
	}

	// Use Save (upsert) to handle duplicate session IDs
	err := h.db.DB.Save(&session).Error
	require.NoError(h.t, err, "Failed to create session")

	// Add user to groups and set permissions
	for _, groupName := range userGroups {
		h.addUserToGroupByName(groupName, namespacePermissions)
	}

	// Create encrypted session cookie using CookieService
	if h.cookieService != nil {
		expiry := time.Now().Add(1 * time.Hour)
		sessionData := &auth.SessionData{
			SessionID:  session.ID,
			UserID:     username,
			Username:   username,
			AuthMethod: "session_password",
			SiteAdmin:  siteAdmin,
			IsAdmin:    siteAdmin,
			UserGroups: userGroups,
			Expiry:     &expiry,
		}
		encrypted, err := h.cookieService.EncryptSession(sessionData)
		require.NoError(h.t, err, "Failed to encrypt session cookie")
		return fmt.Sprintf("terrareg_session=%s", encrypted)
	}

	// Fallback: return unencrypted session ID (for tests without CookieService)
	// This maintains backward compatibility but won't work with encrypted cookie validation
	return fmt.Sprintf("terrareg_session=%s", session.ID)
}

// addUserToGroupByName adds a user to a group, creating the group if necessary
func (h *AuthHelper) addUserToGroupByName(groupName string, namespacePermissions map[string]string) {
	t := h.t

	// Find or create user group
	var group sqldb.UserGroupDB
	err := h.db.DB.Where("name = ?", groupName).First(&group).Error
	if err != nil {
		// Create the group
		group = sqldb.UserGroupDB{
			Name:      groupName,
			SiteAdmin: false,
		}
		err = h.db.DB.Create(&group).Error
		require.NoError(t, err, "Failed to create user group")
	}

	// Add namespace permissions if provided
	for namespaceName, permissionType := range namespacePermissions {
		// Find namespace
		var namespace sqldb.NamespaceDB
		err = h.db.DB.Where("namespace = ?", namespaceName).First(&namespace).Error
		if err != nil {
			// Create namespace if it doesn't exist
			namespace = CreateNamespace(t, h.db, namespaceName, nil)
		}

		// Check if permission already exists
		var existingPerm sqldb.UserGroupNamespacePermissionDB
		err = h.db.DB.Where("user_group_id = ? AND namespace_id = ?", group.ID, namespace.ID).First(&existingPerm).Error
		if err != nil {
			// Create permission - PermissionType is a string type
			perm := sqldb.UserGroupNamespacePermissionDB{
				UserGroupID:    group.ID,
				NamespaceID:    namespace.ID,
				PermissionType: sqldb.UserGroupNamespacePermissionType(permissionType),
			}
			err = h.db.DB.Create(&perm).Error
			require.NoError(t, err, "Failed to create namespace permission")
		}
	}
}

// SetupUserGroupWithPermissions creates a user group with namespace permissions
func (h *AuthHelper) SetupUserGroupWithPermissions(name string, siteAdmin bool, perms map[string]string) int {
	t := h.t

	// Determine if group already exists
	group := sqldb.UserGroupDB{}
	err := h.db.DB.Where("name = ?", name).First(&group).Error
	if err != nil {
		// Create the group
		group = sqldb.UserGroupDB{
			Name:      name,
			SiteAdmin: siteAdmin,
		}

		err = h.db.DB.Create(&group).Error
		require.NoError(t, err, "Failed to create user group")
	}

	// Add namespace permissions
	for namespaceName, permissionType := range perms {
		// Find or create namespace
		var namespace sqldb.NamespaceDB
		err = h.db.DB.Where("namespace = ?", namespaceName).First(&namespace).Error
		if err != nil {
			namespace = CreateNamespace(t, h.db, namespaceName, nil)
		}

		err := h.db.DB.Where("user_group_id = ? AND namespace_id = ?", group.ID, namespace.ID).Delete(sqldb.UserGroupNamespacePermissionDB{}).Error
		require.NoError(t, err, "Failed to delete existing permissions group")

		perm := sqldb.UserGroupNamespacePermissionDB{
			UserGroupID:    group.ID,
			NamespaceID:    namespace.ID,
			PermissionType: sqldb.UserGroupNamespacePermissionType(permissionType),
		}
		err = h.db.DB.Create(&perm).Error
		require.NoError(t, err, "Failed to create namespace permission")
	}

	return group.ID
}

// CreateTerraformIDPToken creates a Terraform IDP authorization and access token for testing
func (h *AuthHelper) CreateTerraformIDPToken(subject string, permissions map[string]string) string {
	t := h.t

	// Create authorization code - Note: fields are Key, Data, Expiry
	codeData := map[string]interface{}{
		"subject": subject,
	}
	codeDataJSON, _ := json.Marshal(codeData)

	// Add unique suffix to avoid UNIQUE constraint violations
	authCodeKey := fmt.Sprintf("test-auth-code-%s-%d", subject, time.Now().UnixNano())
	authCode := sqldb.TerraformIDPAuthorizationCodeDB{
		Key:    authCodeKey,
		Data:   codeDataJSON,
		Expiry: time.Now().Add(5 * time.Minute),
	}
	err := h.db.DB.Where("key = ?", authCodeKey).Delete(&sqldb.TerraformIDPAuthorizationCodeDB{}).Error
	require.NoError(t, err, "Failed to cleanup authorization code")

	err = h.db.DB.Create(&authCode).Error
	require.NoError(t, err, "Failed to create authorization code")

	// Create access token - Note: fields are Key, Data, Expiry
	accessTokenData := map[string]interface{}{
		"subject": subject,
	}
	// FIX: Store namespace_permissions as raw JSON object, not string
	if len(permissions) > 0 {
		permissionsJSON, _ := json.Marshal(permissions)
		var permissionsMap map[string]interface{}
		json.Unmarshal(permissionsJSON, &permissionsMap)
		accessTokenData["namespace_permissions"] = permissionsMap
	}
	accessTokenDataJSON, _ := json.Marshal(accessTokenData)

	// Add unique suffix to avoid UNIQUE constraint violations
	accessTokenKey := fmt.Sprintf("test-access-token-%s-%d", subject, time.Now().UnixNano())
	accessToken := sqldb.TerraformIDPAccessTokenDB{
		Key:    accessTokenKey,
		Data:   accessTokenDataJSON,
		Expiry: time.Now().Add(1 * time.Hour),
	}
	err = h.db.DB.Create(&accessToken).Error
	require.NoError(t, err, "Failed to create access token")

	// Create subject identifier
	subjectData := map[string]interface{}{
		"alias": subject,
	}
	subjectDataJSON, _ := json.Marshal(subjectData)

	// Use Save (upsert) for subject identifier to handle duplicates
	// First try to delete any existing with same subject
	err = h.db.DB.Where("key = ?", subject).Delete(&sqldb.TerraformIDPSubjectIdentifierDB{}).Error
	require.NoError(t, err, "Failed to cleanup subject identifier")

	subjectIdentifier := sqldb.TerraformIDPSubjectIdentifierDB{
		Key:    subject,
		Data:   subjectDataJSON,
		Expiry: time.Now().Add(1 * time.Hour),
	}
	err = h.db.DB.Create(&subjectIdentifier).Error
	require.NoError(t, err, "Failed to create subject identifier")

	return accessToken.Key
}

// MakeAuthenticatedRequest creates an HTTP request with authentication
func (h *AuthHelper) MakeAuthenticatedRequest(method, path string, authMethod string, headers map[string]string) *http.Request {
	req := httptest.NewRequest(method, path, nil)

	// Apply authentication based on method
	switch authMethod {
	case "unauthenticated":
		// No authentication
	case "admin_api_key":
		req.Header.Set("X-Terrareg-ApiKey", h.getAdminApiKey())
	case "upload_api_key":
		req.Header.Set("X-Terrareg-UploadKey", h.getUploadApiKey())
	case "publish_api_key":
		req.Header.Set("X-Terrareg-PublishKey", h.getPublishApiKey())
	case "terraform_idp":
		token := h.CreateTerraformIDPToken("test-subject", nil)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	case "terraform_static":
		// Terraform static token - if configured
		if token := os.Getenv("TERRAFORM_AUTH_TOKEN"); token != "" {
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
		}
	case "session", "admin_session":
		cookie := h.LoginAsBuiltInAdmin()
		if cookie != "" {
			req.Header.Set("Cookie", cookie)
		}
	case "user_session":
		// Create a regular user session
		cookie := h.CreateSessionForUser("testuser", false, []string{}, nil)
		req.Header.Set("Cookie", cookie)
	case "user_with_read_permission":
		cookie := h.CreateSessionForUser("readuser", false, []string{"read-group"}, map[string]string{"test-ns": "READ"})
		req.Header.Set("Cookie", cookie)
	case "user_with_modify_permission":
		cookie := h.CreateSessionForUser("modifyuser", false, []string{"modify-group"}, map[string]string{"test-ns": "MODIFY"})
		req.Header.Set("Cookie", cookie)
	case "user_with_full_permission":
		cookie := h.CreateSessionForUser("fulluser", false, []string{"full-group"}, map[string]string{"test-ns": "FULL"})
		req.Header.Set("Cookie", cookie)
	case "site_admin_user":
		cookie := h.CreateSessionForUser("siteadmin", true, []string{"admin-group"}, nil)
		req.Header.Set("Cookie", cookie)
	}

	// Add custom headers
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	return req
}

// GetUserGroupsForSubject gets user groups for a Terraform IDP subject
func (h *AuthHelper) GetUserGroupsForSubject(subject string, groups []string, siteAdmin bool) {
	t := h.t

	// Create user group if it doesn't exist
	for _, groupName := range groups {
		var group sqldb.UserGroupDB
		err := h.db.DB.Where("name = ?", groupName).First(&group).Error
		if err != nil {
			group = sqldb.UserGroupDB{
				Name:      groupName,
				SiteAdmin: siteAdmin,
			}
			err = h.db.DB.Create(&group).Error
			require.NoError(t, err, "Failed to create user group")
		}

		// Update subject identifier Data field to include group membership
		var subjectID sqldb.TerraformIDPSubjectIdentifierDB
		err = h.db.DB.Where("key = ?", subject).First(&subjectID).Error
		if err == nil {
			// Update Data to include group
			var dataMap map[string]interface{}
			if len(subjectID.Data) > 0 {
				json.Unmarshal(subjectID.Data, &dataMap)
			}
			if dataMap == nil {
				dataMap = make(map[string]interface{})
			}
			if dataMap["user_groups"] == nil {
				dataMap["user_groups"] = []string{}
			}
			userGroups := dataMap["user_groups"].([]string)
			userGroups = append(userGroups, groupName)
			dataMap["user_groups"] = userGroups

			newData, _ := json.Marshal(dataMap)
			subjectID.Data = newData
			h.db.DB.Save(&subjectID)
		}
	}
}
