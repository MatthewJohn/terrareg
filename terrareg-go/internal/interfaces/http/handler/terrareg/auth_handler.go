package terrareg

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"

	authCmd "github.com/matthewjohn/terrareg/terrareg-go/internal/application/command/auth"
	userGroupCmd "github.com/matthewjohn/terrareg/terrareg-go/internal/application/command/user_group"
	authQuery "github.com/matthewjohn/terrareg/terrareg-go/internal/application/query/auth"
	userGroupQuery "github.com/matthewjohn/terrareg/terrareg-go/internal/application/query/user_group"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth/service"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared/types"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/config"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/dto"
)

// AuthHandler handles authentication-related requests
type AuthHandler struct {
	adminLoginCmd            *authCmd.AdminLoginCommand
	checkSessionQuery        *authQuery.CheckSessionQuery
	isAuthenticatedQuery     *authQuery.IsAuthenticatedQuery
	oidcLoginCmd             *authCmd.OidcLoginCommand
	oidcCallbackCmd          *authCmd.OidcCallbackCommand
	samlLoginCmd             *authCmd.SamlLoginCommand
	samlMetadataCmd          *authCmd.SamlMetadataCommand
	githubOAuthCmd           *authCmd.GithubOAuthCommand
	sessionManagementService *service.SessionManagementService
	stateStorageService      *service.StateStorageService
	infraConfig              *config.InfrastructureConfig
	listUserGroupsQuery      *userGroupQuery.ListUserGroupsQuery
	createUserGroupCmd       *userGroupCmd.CreateUserGroupCommand
	deleteUserGroupCmd       *userGroupCmd.DeleteUserGroupCommand
	createNsPermCmd          *userGroupCmd.CreateUserGroupNamespacePermissionCommand
	deleteNsPermCmd          *userGroupCmd.DeleteUserGroupNamespacePermissionCommand
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(
	adminLoginCmd *authCmd.AdminLoginCommand,
	checkSessionQuery *authQuery.CheckSessionQuery,
	isAuthenticatedQuery *authQuery.IsAuthenticatedQuery,
	oidcLoginCmd *authCmd.OidcLoginCommand,
	oidcCallbackCmd *authCmd.OidcCallbackCommand,
	samlLoginCmd *authCmd.SamlLoginCommand,
	samlMetadataCmd *authCmd.SamlMetadataCommand,
	githubOAuthCmd *authCmd.GithubOAuthCommand,
	sessionManagementService *service.SessionManagementService,
	stateStorageService *service.StateStorageService,
	infraConfig *config.InfrastructureConfig,
	listUserGroupsQuery *userGroupQuery.ListUserGroupsQuery,
	createUserGroupCmd *userGroupCmd.CreateUserGroupCommand,
	deleteUserGroupCmd *userGroupCmd.DeleteUserGroupCommand,
	createNsPermCmd *userGroupCmd.CreateUserGroupNamespacePermissionCommand,
	deleteNsPermCmd *userGroupCmd.DeleteUserGroupNamespacePermissionCommand,
) *AuthHandler {
	return &AuthHandler{
		adminLoginCmd:            adminLoginCmd,
		checkSessionQuery:        checkSessionQuery,
		isAuthenticatedQuery:     isAuthenticatedQuery,
		oidcLoginCmd:             oidcLoginCmd,
		oidcCallbackCmd:          oidcCallbackCmd,
		samlLoginCmd:             samlLoginCmd,
		samlMetadataCmd:          samlMetadataCmd,
		githubOAuthCmd:           githubOAuthCmd,
		sessionManagementService: sessionManagementService,
		stateStorageService:      stateStorageService,
		infraConfig:              infraConfig,
		listUserGroupsQuery:      listUserGroupsQuery,
		createUserGroupCmd:       createUserGroupCmd,
		deleteUserGroupCmd:       deleteUserGroupCmd,
		createNsPermCmd:          createNsPermCmd,
		deleteNsPermCmd:          deleteNsPermCmd,
	}
}

// HandleAdminLogin handles POST /v1/terrareg/auth/admin/login
// Now follows DDD principles and matches Python API contract
func (h *AuthHandler) HandleAdminLogin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Only allow POST requests
	if r.Method != http.MethodPost {
		RespondJSON(w, http.StatusMethodNotAllowed, map[string]interface{}{
			"message": "Method not allowed",
		})
		return
	}

	// Execute admin login command using the request (extracts X-Terrareg-ApiKey header)
	response, err := h.adminLoginCmd.ExecuteWithRequest(ctx, r)
	if err != nil {
		// Handle different error scenarios appropriately
		if !h.adminLoginCmd.IsConfigured() {
			RespondJSON(w, http.StatusForbidden, map[string]interface{}{
				"message": "Admin authentication is not enabled",
			})
		} else {
			RespondJSON(w, http.StatusUnauthorized, map[string]interface{}{
				"message": "Invalid API key",
			})
		}
		return
	}

	// If authentication failed, return 401
	if !response.Authenticated {
		RespondJSON(w, http.StatusUnauthorized, map[string]interface{}{
			"message": "Authentication failed",
		})
		return
	}

	// Set session cookie for the existing session created by the admin login command
	log.Info().
		Str("session_id", response.SessionID).
		Msg("Setting admin session cookie")

	if err := h.sessionManagementService.SetCookieForExistingSession(ctx, w, response.SessionID, "Built-in admin", "ADMIN_API_KEY"); err != nil {
		// If cookie setting fails, return error
		log.Error().
			Err(err).
			Str("session_id", response.SessionID).
			Msg("Failed to set admin session cookie")

		RespondJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"message": "Failed to set session cookie",
		})
		return
	}

	log.Info().
		Str("session_id", response.SessionID).
		Msg("Admin session cookie set successfully")

	// Respond with Python-compatible format
	RespondJSON(w, http.StatusOK, map[string]interface{}{
		"authenticated": response.Authenticated,
	})
}

// HandleIsAuthenticated handles GET /v1/terrareg/auth/admin/is_authenticated
func (h *AuthHandler) HandleIsAuthenticated(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get authentication status using the query
	response, err := h.isAuthenticatedQuery.Execute(ctx)
	if err != nil {
		// If there's an error, return unauthenticated status
		response = &dto.IsAuthenticatedResponse{
			Authenticated:        false,
			ReadAccess:           false,
			SiteAdmin:            false,
			NamespacePermissions: make(map[string]string),
		}
	}

	RespondJSON(w, http.StatusOK, response)
}

// HandleOIDCLogin handles GET /v1/terrareg/auth/oidc/login
func (h *AuthHandler) HandleOIDCLogin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get redirect URL from query parameter
	redirectURL := r.URL.Query().Get("redirect_url")
	if redirectURL == "" {
		redirectURL = "/"
	}

	// Generate state parameter for CSRF protection
	state := generateRandomState()

	// Execute OIDC login command
	response, err := h.oidcLoginCmd.Execute(ctx, &authCmd.OidcLoginRequest{
		RedirectURL: redirectURL,
		State:       state,
	})
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Store state using secure state storage
	if h.stateStorageService != nil {
		// Generate and store secure state for CSRF protection
		redirectURL := getRedirectURL(r)
		stateParam, err := h.stateStorageService.GenerateAndStoreState(ctx, redirectURL, "oidc")
		if err != nil {
			RespondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to generate secure state: %v", err))
			return
		}

		// Use the generated state instead of the one from response
		response.State = stateParam
	}

	// Redirect to OIDC provider
	http.Redirect(w, r, response.AuthURL, http.StatusFound)
}

// HandleOIDCCallback handles GET /v1/terrareg/auth/oidc/callback
func (h *AuthHandler) HandleOIDCCallback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get authorization code and state from query parameters
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	errorParam := r.URL.Query().Get("error")
	errorDescription := r.URL.Query().Get("error_description")

	// Execute OIDC callback command
	response, err := h.oidcCallbackCmd.Execute(ctx, &authCmd.OidcCallbackRequest{
		Code:             code,
		State:            state,
		Error:            errorParam,
		ErrorDescription: errorDescription,
	})
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Check if authentication failed
	if !response.Authenticated {
		RespondError(w, http.StatusUnauthorized, response.ErrorMessage)
		return
	}

	// Set session cookie for the existing session
	// Note: Passing empty username/authMethod to skip audit logging here
	// The audit logging should be done in the command layer when full user context is available
	if err := h.sessionManagementService.SetCookieForExistingSession(ctx, w, response.SessionID, "", ""); err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to set session cookie")
		return
	}

	// Redirect to application
	http.Redirect(w, r, "/", http.StatusFound)
}

// HandleSAMLLogin handles GET /v1/terrareg/auth/saml/login
func (h *AuthHandler) HandleSAMLLogin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get redirect URL from query parameter
	redirectURL := r.URL.Query().Get("redirect_url")
	if redirectURL == "" {
		redirectURL = "/"
	}

	// Execute SAML login command
	response, err := h.samlLoginCmd.Execute(ctx, &authCmd.SamlLoginRequest{
		RelayState: redirectURL,
	})
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Redirect to SAML IDP
	http.Redirect(w, r, response.AuthURL, http.StatusFound)
}

// HandleSAMLACS handles POST /v1/terrareg/auth/saml/acs
func (h *AuthHandler) HandleSAMLACS(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.Method != http.MethodPost {
		RespondJSON(w, http.StatusMethodNotAllowed, map[string]interface{}{
			"message": "Method not allowed",
		})
		return
	}

	// Extract SAMLResponse from form data
	if err := r.ParseForm(); err != nil {
		log.Error().Err(err).Msg("Failed to parse form data")
		RespondJSON(w, http.StatusBadRequest, map[string]interface{}{
			"message": "Failed to parse form data",
		})
		return
	}

	samlResponse := r.FormValue("SAMLResponse")
	relayState := r.FormValue("RelayState")

	if samlResponse == "" {
		log.Warn().Msg("Missing SAMLResponse parameter")
		RespondJSON(w, http.StatusBadRequest, map[string]interface{}{
			"message": "Missing SAMLResponse",
		})
		return
	}

	// Log authentication attempt
	log.Info().
		Str("provider", "saml").
		Str("ip_address", getClientIP(r)).
		Msg("Processing SAML authentication")

	// TODO: This would ideally use a SAML callback command to properly validate
	// the SAML response and extract user attributes.
	// For now, create a minimal SamlAuthContext with placeholder data.
	// Production implementation should:
	// 1. Validate SAML response signature
	// 2. Decrypt encrypted assertions
	// 3. Extract user attributes (nameID, email, username, groups, etc.)

	// Create minimal SamlAuthContext - proper SAML parsing is TODO
	samlAuthCtx := auth.NewSamlAuthContext(ctx, relayState, make(map[string][]string))
	samlAuthCtx.ExtractUserDetails() // Extract what we can from attributes

	// Extract data from AuthContext to create session
	providerData := samlAuthCtx.GetProviderData()
	ttl := 24 * time.Hour

	// Create session and set cookie using SessionManagementService
	err := h.sessionManagementService.CreateSessionAndCookie(
		ctx,
		w,
		samlAuthCtx.GetProviderType(),
		samlAuthCtx.GetUsername(),
		samlAuthCtx.IsAdmin(),
		samlAuthCtx.GetUserGroupNames(),
		samlAuthCtx.GetAllNamespacePermissions(),
		providerData,
		&ttl,
	)
	if err != nil {
		log.Error().
			Err(err).
			Str("provider", "saml").
			Msg("Failed to create SAML session")
		RespondJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"message": "Failed to create session",
		})
		return
	}

	log.Info().
		Str("provider", "saml").
		Msg("SAML authentication completed")

	// Redirect to application
	redirectURL := relayState
	if redirectURL == "" {
		redirectURL = "/"
	}

	http.Redirect(w, r, redirectURL, http.StatusFound)
}

// HandleSAMLMetadata handles GET /v1/terrareg/auth/saml/metadata
func (h *AuthHandler) HandleSAMLMetadata(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Execute SAML metadata command
	response, err := h.samlMetadataCmd.Execute(ctx, &authCmd.SamlMetadataRequest{})
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Return SAML metadata XML
	w.Header().Set("Content-Type", h.samlMetadataCmd.GetMetadataContentType())
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(response.Metadata))
}

// HandleGitHubOAuth handles GET /v1/terrareg/auth/github/oauth
func (h *AuthHandler) HandleGitHubOAuth(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get redirect URL from query parameter
	redirectURL := r.URL.Query().Get("redirect_url")
	if redirectURL == "" {
		redirectURL = "/"
	}

	// Generate state parameter for CSRF protection
	state := generateRandomState()

	// Execute GitHub OAuth login command
	response, err := h.githubOAuthCmd.ExecuteLogin(ctx, &authCmd.GithubOAuthLoginRequest{
		RedirectURL: redirectURL,
		State:       state,
	})
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Store state using secure state storage
	if h.stateStorageService != nil {
		// Generate and store secure state for CSRF protection
		redirectURL := getRedirectURL(r)
		stateParam, err := h.stateStorageService.GenerateAndStoreState(ctx, redirectURL, "github")
		if err != nil {
			RespondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to generate secure state: %v", err))
			return
		}

		// Use the generated state instead of the one from response
		response.State = stateParam
	}

	// Redirect to GitHub OAuth provider
	http.Redirect(w, r, response.AuthURL, http.StatusFound)
}

// HandleLogout handles POST /v1/terrareg/auth/logout
func (h *AuthHandler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Clear session and cookie
	if err := h.sessionManagementService.ClearSessionAndCookie(ctx, w, r); err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Return success response
	RespondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Successfully logged out",
	})
}

// HandleUserGroupList handles GET /v1/terrareg/user-groups
// Matches Python: ApiTerraregAuthUserGroups._get()
func (h *AuthHandler) HandleUserGroupList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Only allow GET requests
	if r.Method != http.MethodGet {
		RespondJSON(w, http.StatusMethodNotAllowed, map[string]interface{}{
			"message": "Method not allowed",
		})
		return
	}

	// Execute list user groups query
	userGroups, err := h.listUserGroupsQuery.Execute(ctx)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Return response matching Python format
	RespondJSON(w, http.StatusOK, userGroups)
}

// HandleUserGroupCreate handles POST /v1/terrareg/user-groups
// Matches Python: ApiTerraregAuthUserGroups._post()
func (h *AuthHandler) HandleUserGroupCreate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Only allow POST requests
	if r.Method != http.MethodPost {
		RespondJSON(w, http.StatusMethodNotAllowed, map[string]interface{}{
			"message": "Method not allowed",
		})
		return
	}

	// Parse request body
	var req userGroupCmd.CreateUserGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondJSON(w, http.StatusBadRequest, map[string]interface{}{
			"message": "Invalid request body",
		})
		return
	}

	// Execute create user group command
	response, err := h.createUserGroupCmd.Execute(ctx, req)
	if err != nil {
		// Handle specific errors with appropriate HTTP status codes
		switch {
		case errors.Is(err, userGroupCmd.ErrInvalidUserGroupName):
			RespondJSON(w, http.StatusBadRequest, map[string]interface{}{
				"message": "Invalid user group name",
			})
		case errors.Is(err, userGroupCmd.ErrInvalidSiteAdminValue):
			RespondJSON(w, http.StatusBadRequest, map[string]interface{}{
				"message": "site_admin must be True or False",
			})
		case errors.Is(err, userGroupCmd.ErrUserGroupAlreadyExists):
			RespondJSON(w, http.StatusBadRequest, map[string]interface{}{
				"message": "User group already exists",
			})
		default:
			RespondError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	// Return response with 201 Created (matches Python)
	RespondJSON(w, http.StatusCreated, response)
}

// HandleUserGroupDelete handles DELETE /v1/terrareg/user-groups/{group}
// Matches Python: ApiTerraregAuthUserGroup._delete(user_group)
func (h *AuthHandler) HandleUserGroupDelete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Only allow DELETE requests
	if r.Method != http.MethodDelete {
		RespondJSON(w, http.StatusMethodNotAllowed, map[string]interface{}{
			"message": "Method not allowed",
		})
		return
	}

	// Extract user group name from URL parameter using chi
	// URL pattern: /v1/terrareg/user-groups/{group}
	userGroupName := chi.URLParam(r, "group")
	if userGroupName == "" {
		RespondJSON(w, http.StatusBadRequest, map[string]interface{}{
			"message": "User group name is required",
		})
		return
	}

	// Execute delete user group command
	err := h.deleteUserGroupCmd.Execute(ctx, userGroupName)
	if err != nil {
		if errors.Is(err, userGroupCmd.ErrUserGroupNotFound) {
			RespondJSON(w, http.StatusBadRequest, map[string]interface{}{
				"message": "User group does not exist",
			})
		} else {
			RespondError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	// Return 200 OK with empty body (matches Python)
	RespondJSON(w, http.StatusOK, map[string]interface{}{})
}

// HandleUserGroupNamespacePermissionsCreate handles POST /v1/terrareg/user-groups/{group}/permissions/{namespace}
// Matches Python: ApiTerraregAuthUserGroupNamespacePermissions._post(user_group, namespace)
func (h *AuthHandler) HandleUserGroupNamespacePermissionsCreate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Only allow POST requests
	if r.Method != http.MethodPost {
		RespondJSON(w, http.StatusMethodNotAllowed, map[string]interface{}{
			"message": "Method not allowed",
		})
		return
	}

	// Parse request body
	var req userGroupCmd.CreateNamespacePermissionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error().Err(err).Str("path", r.URL.Path).Msg("Failed to decode create namespace permission request")
		RespondJSON(w, http.StatusBadRequest, map[string]interface{}{
			"message": "Invalid request body",
		})
		return
	}

	log.Info().Str("permission_type", req.PermissionType).Msg("Decoded permission type")

	// Extract user group name and namespace from URL parameters using chi
	// URL pattern: /v1/terrareg/user-groups/{group}/permissions/{namespace}
	userGroupName := chi.URLParam(r, "group")
	namespaceName := chi.URLParam(r, "namespace")
	log.Info().Str("user_group", userGroupName).Str("namespace", namespaceName).Msg("URL params")
	if userGroupName == "" || namespaceName == "" {
		log.Warn().Msg("Missing user group or namespace in URL params")
		RespondJSON(w, http.StatusBadRequest, map[string]interface{}{
			"message": "User group name and namespace are required",
		})
		return
	}

	// Execute create namespace permission command
	response, err := h.createNsPermCmd.Execute(ctx, userGroupName, types.NamespaceName(namespaceName), req)
	if err != nil {
		// Handle specific errors with appropriate HTTP status codes
		log.Error().Err(err).Str("user_group", userGroupName).Str("namespace", namespaceName).Msg("Failed to create namespace permission")
		switch {
		case errors.Is(err, userGroupCmd.ErrInvalidPermissionType):
			RespondJSON(w, http.StatusBadRequest, map[string]interface{}{
				"message": "Invalid namespace permission type",
			})
		case errors.Is(err, userGroupCmd.ErrUserGroupNotFound):
			RespondJSON(w, http.StatusBadRequest, map[string]interface{}{
				"message": "User group does not exist",
			})
		case errors.Is(err, userGroupCmd.ErrNamespaceNotFound):
			RespondJSON(w, http.StatusBadRequest, map[string]interface{}{
				"message": "Namespace does not exist",
			})
		case errors.Is(err, userGroupCmd.ErrPermissionAlreadyExists):
			RespondJSON(w, http.StatusBadRequest, map[string]interface{}{
				"message": "Permission already exists for this user_group/namespace",
			})
		default:
			RespondError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	// Return response with 201 Created (matches Python)
	RespondJSON(w, http.StatusCreated, response)
}

// HandleUserGroupNamespacePermissionsDelete handles DELETE /v1/terrareg/user-groups/{group}/permissions/{namespace}
// Matches Python: ApiTerraregAuthUserGroupNamespacePermissions._delete(user_group, namespace)
func (h *AuthHandler) HandleUserGroupNamespacePermissionsDelete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Only allow DELETE requests
	if r.Method != http.MethodDelete {
		RespondJSON(w, http.StatusMethodNotAllowed, map[string]interface{}{
			"message": "Method not allowed",
		})
		return
	}

	// Extract user group name and namespace from URL parameters using chi
	// URL pattern: /v1/terrareg/user-groups/{group}/permissions/{namespace}
	userGroupName := chi.URLParam(r, "group")
	namespaceName := chi.URLParam(r, "namespace")
	if userGroupName == "" || namespaceName == "" {
		RespondJSON(w, http.StatusBadRequest, map[string]interface{}{
			"message": "User group name and namespace are required",
		})
		return
	}

	// Execute delete namespace permission command
	err := h.deleteNsPermCmd.Execute(ctx, userGroupName, types.NamespaceName(namespaceName))
	if err != nil {
		// Handle specific errors with appropriate HTTP status codes
		switch {
		case errors.Is(err, userGroupCmd.ErrUserGroupNotFound):
			RespondJSON(w, http.StatusBadRequest, map[string]interface{}{
				"message": "User group does not exist",
			})
		case errors.Is(err, userGroupCmd.ErrNamespaceNotFound):
			RespondJSON(w, http.StatusBadRequest, map[string]interface{}{
				"message": "Namespace does not exist",
			})
		case errors.Is(err, userGroupCmd.ErrPermissionNotFound):
			RespondJSON(w, http.StatusBadRequest, map[string]interface{}{
				"message": "Permission does not exist",
			})
		default:
			RespondError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	// Return 200 OK with empty body (matches Python)
	RespondJSON(w, http.StatusOK, map[string]interface{}{})
}

// generateRandomState generates a random state parameter for CSRF protection
func generateRandomState() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}

// getRedirectURL extracts the redirect URL from the request
func getRedirectURL(r *http.Request) string {
	redirectURL := r.URL.Query().Get("redirect_url")
	if redirectURL == "" {
		redirectURL = "/"
	}
	return redirectURL
}

// getClientIP extracts the client IP address from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first (for load balancers)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	return r.RemoteAddr
}
