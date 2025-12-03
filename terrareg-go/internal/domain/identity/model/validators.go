package model

import (
	"errors"
	"regexp"
	"strings"
)

// Using errors from errors.go instead of redefining here

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