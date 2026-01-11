package model

import "errors"

// Identity domain errors - these are specific to the identity domain
var (
	// User-related errors
	ErrUserNotFound       = errors.New("user not found")
	ErrUserInactive       = errors.New("user is inactive")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")

	// Authentication errors
	ErrAuthenticationFailed              = errors.New("authentication failed")
	ErrInvalidAuthMethod                 = errors.New("invalid authentication method")
	ErrTokenExpired                      = errors.New("token expired")
	ErrTokenInvalid                      = errors.New("token invalid")
	ErrAuthorizationCodeExpired          = errors.New("authorization code expired")
	ErrAuthorizationCodeAlreadyExchanged = errors.New("authorization code already exchanged")

	// API Key errors
	ErrAPIKeyNotFound = errors.New("API key not found")
	ErrAPIKeyExpired  = errors.New("API key has expired")
	ErrAPIKeyDisabled = errors.New("API key is disabled")
	ErrInvalidAPIKey  = errors.New("invalid API key")
	ErrTooManyAPIKeys = errors.New("too many API keys for user")

	// Session errors
	ErrSessionNotFound  = errors.New("session not found")
	ErrSessionExpired   = errors.New("session has expired")
	ErrSessionInvalid   = errors.New("session invalid")
	ErrUserNotInContext = errors.New("user not found in context")

	// OAuth/OIDC errors
	ErrOIDCConfiguration = errors.New("OIDC configuration required")
	ErrOIDCLoginFailed   = errors.New("OIDC login failed")
	ErrOIDCTokenInvalid  = errors.New("OIDC token invalid")

	// SAML errors
	ErrSAMLConfiguration   = errors.New("SAML configuration required")
	ErrSAMLLoginFailed     = errors.New("SAML login failed")
	ErrSAMLResponseInvalid = errors.New("SAML response invalid")

	// Required field errors
	ErrUserIDRequired            = errors.New("user ID required")
	ErrUsernameRequired          = errors.New("username required")
	ErrEmailRequired             = errors.New("email required")
	ErrPasswordRequired          = errors.New("password required")
	ErrClientIDRequired          = errors.New("client ID required")
	ErrClientSecretRequired      = errors.New("client secret required")
	ErrRedirectURIRequired       = errors.New("redirect URI required")
	ErrAuthorizationCodeRequired = errors.New("authorization code required")
	ErrAccessTokenRequired       = errors.New("access token required")
	ErrRefreshTokenRequired      = errors.New("refresh token required")
	ErrScopesRequired            = errors.New("scopes required")
	ErrProviderNameRequired      = errors.New("provider name required")
	ErrSubjectIDRequired         = errors.New("subject ID required")
	ErrTokenTypeRequired         = errors.New("token type required")

	// Validation errors
	ErrInvalidInput     = errors.New("invalid input")
	ErrInvalidName      = errors.New("invalid name")
	ErrInvalidEmail     = errors.New("invalid email")
	ErrInvalidPassword  = errors.New("invalid password")
	ErrInvalidTokenType = errors.New("invalid token type")
	ErrInvalidNamespace = errors.New("invalid namespace")
	ErrInvalidProvider  = errors.New("invalid provider")

	// Field-specific validation errors
	ErrUsernameInvalid         = errors.New("username is invalid")
	ErrUsernameTooLong         = errors.New("username is too long")
	ErrEmailInvalid            = errors.New("email is invalid")
	ErrDisplayNameRequired     = errors.New("display name is required")
	ErrDisplayNameTooLong      = errors.New("display name is too long")
	ErrUserAlreadyDeactivated  = errors.New("user is already deactivated")
	ErrGroupNameRequired       = errors.New("group name is required")
	ErrGroupNameInvalid        = errors.New("group name is invalid")
	ErrGroupNameTooLong        = errors.New("group name is too long")
	ErrUserRequired            = errors.New("user is required")
	ErrUserAlreadyInGroup      = errors.New("user is already in group")
	ErrUserNotInGroup          = errors.New("user is not in group")
	ErrSessionInactive         = errors.New("session is inactive")
	ErrNamespaceIDRequired     = errors.New("namespace ID is required")
	ErrUserGroupIDRequired     = errors.New("user group ID is required")
	ErrActionsRequired         = errors.New("actions are required")
	ErrActionAlreadyExists     = errors.New("action already exists")
	ErrActionNotFound          = errors.New("action not found")
	ErrCannotRemoveAllActions  = errors.New("cannot remove all actions")
	ErrPermissionNotFound      = errors.New("permission not found")
	ErrPermissionAlreadyExists = errors.New("permission already exists")

	// Authentication Token errors
	ErrTokenNotFound       = errors.New("authentication token not found")
	ErrTokenInactive       = errors.New("authentication token is inactive")
	ErrTokenAlreadyRevoked = errors.New("authentication token is already revoked")
	ErrTokenValueRequired  = errors.New("token value is required")
	ErrDescriptionRequired = errors.New("description is required")
	ErrDescriptionTooLong  = errors.New("description is too long")
	ErrCreatedByRequired   = errors.New("created by is required")
	ErrExpirationInPast    = errors.New("expiration date is in the past")
	ErrInvalidTokenValue   = errors.New("invalid token value format")
)
