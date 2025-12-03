package model

import (
	"errors"
	"regexp"
	"strings"
)

var (
	// User validation
	ErrUsernameRequired       = errors.New("username is required")
	ErrUsernameInvalid        = errors.New("username is invalid")
	ErrUsernameTooLong        = errors.New("username is too long")
	ErrEmailRequired          = errors.New("email is required")
	ErrEmailInvalid           = errors.New("email is invalid")
	ErrDisplayNameRequired     = errors.New("display name is required")
	ErrDisplayNameTooLong     = errors.New("display name is too long")
	ErrUserAlreadyDeactivated = errors.New("user is already deactivated")
	ErrAccessTokenRequired     = errors.New("access token is required")
	ErrPermissionAlreadyExists = errors.New("permission already exists")

	// User group validation
	ErrGroupNameRequired      = errors.New("group name is required")
	ErrGroupNameInvalid       = errors.New("group name is invalid")
	ErrGroupNameTooLong      = errors.New("group name is too long")
	ErrUserRequired          = errors.New("user is required")
	ErrUserIDRequired        = errors.New("user ID is required")
	ErrUserAlreadyInGroup    = errors.New("user is already in group")
	ErrUserNotInGroup        = errors.New("user is not in group")

	// Session validation
	ErrSessionInactive        = errors.New("session is inactive")

	// Permission validation
	ErrNamespaceIDRequired   = errors.New("namespace ID is required")
	ErrUserGroupIDRequired   = errors.New("user group ID is required")
	ErrActionsRequired       = errors.New("actions are required")
	ErrActionAlreadyExists   = errors.New("action already exists")
	ErrActionNotFound        = errors.New("action not found")
	ErrCannotRemoveAllActions = errors.New("cannot remove all actions")
	ErrPermissionNotFound    = errors.New("permission not found")

	// OAuth validation
	ErrInvalidAuthMethod         = errors.New("invalid auth method")
	ErrProviderNameRequired      = errors.New("provider name is required")
	ErrSubjectIDRequired         = errors.New("subject ID is required")
	ErrAuthorizationCodeRequired = errors.New("authorization code is required")
	ErrAuthorizationCodeExpired   = errors.New("authorization code is expired")
	ErrAuthorizationCodeAlreadyExchanged = errors.New("authorization code already exchanged")
	ErrClientIDRequired          = errors.New("client ID is required")
	ErrRedirectURIRequired       = errors.New("redirect URI is required")
	ErrScopesRequired            = errors.New("scopes are required")
	ErrTokenTypeRequired          = errors.New("token type is required")
)

// Validation patterns
var (
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	emailRegex    = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	groupNameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
)

// ValidateUsername validates a username
func ValidateUsername(username string) error {
	if username == "" {
		return ErrUsernameRequired
	}
	if len(username) > 50 {
		return ErrUsernameTooLong
	}
	if !usernameRegex.MatchString(username) {
		return ErrUsernameInvalid
	}
	return nil
}

// ValidateEmail validates an email address
func ValidateEmail(email string) error {
	if email == "" {
		return ErrEmailRequired
	}
	if !emailRegex.MatchString(email) {
		return ErrEmailInvalid
	}
	if len(email) > 255 {
		return ErrEmailInvalid // email too long
	}
	return nil
}

// ValidateDisplayName validates a display name
func ValidateDisplayName(displayName string) error {
	if displayName == "" {
		return ErrDisplayNameRequired
	}
	if len(displayName) > 100 {
		return ErrDisplayNameTooLong
	}
	return nil
}

// ValidateGroupName validates a group name
func ValidateGroupName(name string) error {
	if name == "" {
		return ErrGroupNameRequired
	}
	if len(name) > 50 {
		return ErrGroupNameTooLong
	}
	if !groupNameRegex.MatchString(name) {
		return ErrGroupNameInvalid
	}
	return nil
}

// ValidateAction validates an action
func ValidateAction(action Action) error {
	switch action {
	case ActionRead, ActionWrite, ActionAdmin:
		return nil
	default:
		return errors.New("invalid action")
	}
}

// ValidateResourceType validates a resource type
func ValidateResourceType(resourceType ResourceType) error {
	switch resourceType {
	case ResourceTypeNamespace, ResourceTypeModule, ResourceTypeProvider:
		return nil
	default:
		return errors.New("invalid resource type")
	}
}

// ValidateAuthMethod validates an auth method
func ValidateAuthMethod(authMethod AuthMethod) error {
	switch authMethod {
	case AuthMethodNone, AuthMethodSAML, AuthMethodOIDC, AuthMethodGitHub, AuthMethodAPIKey, AuthMethodTerraform:
		return nil
	default:
		return errors.New("invalid auth method")
	}
}

// SanitizeInput sanitizes input strings
func SanitizeInput(input string) string {
	return strings.TrimSpace(input)
}

// ValidateScopes validates OAuth scopes
func ValidateScopes(scopes []string) error {
	if len(scopes) == 0 {
		return ErrScopesRequired
	}

	for _, scope := range scopes {
		if scope == "" {
			return errors.New("empty scope not allowed")
		}
	}

	return nil
}