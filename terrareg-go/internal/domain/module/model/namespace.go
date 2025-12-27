package model

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared"
)

// NamespaceType represents the type of namespace
type NamespaceType string

const (
	NamespaceTypeNone       NamespaceType = "NONE"
	NamespaceTypeGithubUser NamespaceType = "GITHUB_USER"
	NamespaceTypeGithubOrg  NamespaceType = "GITHUB_ORGANISATION"
)

// Namespace represents a namespace for modules and providers
type Namespace struct {
	id          int
	name        string
	displayName *string
	nsType      NamespaceType
}

var namespaceNameRegex = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]*[a-zA-Z0-9]$`)

// NewNamespace creates a new namespace
func NewNamespace(name string, displayName *string, nsType NamespaceType) (*Namespace, error) {
	if err := ValidateNamespaceName(name); err != nil {
		return nil, err
	}

	return &Namespace{
		name:        name,
		displayName: displayName,
		nsType:      nsType,
	}, nil
}

// ReconstructNamespace reconstructs a namespace from persistence (used by repository)
func ReconstructNamespace(id int, name string, displayName *string, nsType NamespaceType) *Namespace {
	return &Namespace{
		id:          id,
		name:        name,
		displayName: displayName,
		nsType:      nsType,
	}
}

// ValidateNamespaceName validates a namespace name
func ValidateNamespaceName(name string) error {
	if name == "" {
		return fmt.Errorf("%w: namespace name cannot be empty", shared.ErrInvalidNamespace)
	}

	if len(name) < 2 {
		return fmt.Errorf("%w: namespace name must be at least 2 characters", shared.ErrInvalidNamespace)
	}

	if len(name) > 128 {
		return fmt.Errorf("%w: namespace name must not exceed 128 characters", shared.ErrInvalidNamespace)
	}

	// Convert to lowercase for validation
	name = strings.ToLower(name)

	if !namespaceNameRegex.MatchString(name) {
		return fmt.Errorf("%w: namespace name must contain only alphanumeric characters, hyphens, and underscores", shared.ErrInvalidNamespace)
	}

	// Reserved names
	reserved := []string{"modules", "providers", "v1", "v2", "api", "admin", "login", "logout", "terrareg"}
	for _, r := range reserved {
		if strings.ToLower(name) == r {
			return fmt.Errorf("%w: '%s' is a reserved namespace name", shared.ErrInvalidNamespace, name)
		}
	}

	return nil
}

// ID returns the namespace ID
func (n *Namespace) ID() int {
	return n.id
}

// Name returns the namespace name
func (n *Namespace) Name() string {
	return n.name
}

// DisplayName returns the display name
func (n *Namespace) DisplayName() *string {
	return n.displayName
}

// Type returns the namespace type
func (n *Namespace) Type() NamespaceType {
	return n.nsType
}

// SetDisplayName sets the display name
func (n *Namespace) SetDisplayName(displayName *string) {
	n.displayName = displayName
}

// SetType sets the namespace type
func (n *Namespace) SetType(nsType NamespaceType) {
	n.nsType = nsType
}

// IsGithub returns true if this is a GitHub namespace
func (n *Namespace) IsGithub() bool {
	return n.nsType == NamespaceTypeGithubUser || n.nsType == NamespaceTypeGithubOrg
}

// String returns the string representation
func (n *Namespace) String() string {
	return n.name
}
