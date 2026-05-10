package model

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider_source/service"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared/types"
)

// NamespaceType represents the type of namespace
type NamespaceType string

const (
	NamespaceTypeNone       NamespaceType = "NONE"
	NamespaceTypeGithubUser NamespaceType = "GITHUB_USER"
	NamespaceTypeGithubOrg  NamespaceType = "GITHUB_ORGANISATION"
)

// ProviderSourceFactory defines the interface for getting provider sources
// This is a minimal interface that the ProviderSourceFactory service must implement
type ProviderSourceFactory interface {
	GetProviderSourceByName(ctx context.Context, name string) (service.ProviderSourceInstance, error)
}

// Namespace represents a namespace for modules and providers
type Namespace struct {
	id                           int
	name                         types.NamespaceName
	displayName                  *string
	nsType                       NamespaceType
	defaultProviderSourceName    *string
	providerSourceFactory        ProviderSourceFactory
}

var namespaceNameRegex = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]*[a-zA-Z0-9]$`)

// NewNamespace creates a new namespace
func NewNamespace(name types.NamespaceName, displayName *string, nsType NamespaceType) (*Namespace, error) {
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
func ReconstructNamespace(id int, name types.NamespaceName, displayName *string, nsType NamespaceType, defaultProviderSourceName *string, providerSourceFactory ProviderSourceFactory) *Namespace {
	return &Namespace{
		id:                        id,
		name:                      name,
		displayName:               displayName,
		nsType:                    nsType,
		defaultProviderSourceName: defaultProviderSourceName,
		providerSourceFactory:     providerSourceFactory,
	}
}

// ValidateNamespaceName validates a namespace name
func ValidateNamespaceName(name types.NamespaceName) error {
	if name == "" {
		return fmt.Errorf("%w: namespace name cannot be empty", shared.ErrInvalidNamespace)
	}

	if len(name) < 2 {
		return fmt.Errorf("%w: namespace name must be at least 2 characters", shared.ErrInvalidNamespace)
	}

	if len(name) > 128 {
		return fmt.Errorf("%w: namespace name must not exceed 128 characters", shared.ErrInvalidNamespace)
	}

	stringName := string(name)

	// Check for double underscores (sequential underscores not allowed)
	if strings.Contains(stringName, "__") {
		return fmt.Errorf("%w: namespace name cannot contain sequential underscores", shared.ErrInvalidNamespace)
	}

	// Convert to lowercase for validation
	stringName = strings.ToLower(stringName)

	if !namespaceNameRegex.MatchString(stringName) {
		return fmt.Errorf("%w: namespace name must contain only alphanumeric characters, hyphens, and underscores", shared.ErrInvalidNamespace)
	}

	// Reserved names
	reserved := []string{"modules", "providers", "v1", "v2", "api", "admin", "login", "logout", "terrareg"}
	for _, r := range reserved {
		if strings.ToLower(stringName) == r {
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
func (n *Namespace) Name() types.NamespaceName {
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
	return string(n.name)
}

// DefaultProviderSourceName returns the default provider source name
func (n *Namespace) DefaultProviderSourceName() *string {
	return n.defaultProviderSourceName
}

// SetDefaultProviderSourceName sets the default provider source name
func (n *Namespace) SetDefaultProviderSourceName(name *string) {
	n.defaultProviderSourceName = name
}

// DefaultProviderSource returns the default provider source for this namespace
// Python reference: /app/terrareg/models.py lines 1028-1037
func (n *Namespace) DefaultProviderSource(ctx context.Context) (service.ProviderSourceInstance, error) {
	if n.defaultProviderSourceName == nil || n.providerSourceFactory == nil {
		return nil, nil
	}
	return n.providerSourceFactory.GetProviderSourceByName(ctx, *n.defaultProviderSourceName)
}

// UpdateDefaultProviderSource updates the default provider source for this namespace
// Python reference: /app/terrareg/models.py lines 1174-1213
func (n *Namespace) UpdateDefaultProviderSource(ctx context.Context, providerSourceName *string) error {
	// If nil, no change requested
	if providerSourceName == nil {
		return nil
	}

	// If empty string, unset (set to nil)
	var newValue *string
	if *providerSourceName == "" {
		newValue = nil
	} else {
		// Validate provider source exists
		if n.providerSourceFactory != nil {
			providerSource, err := n.providerSourceFactory.GetProviderSourceByName(ctx, *providerSourceName)
			if err != nil {
				return err
			}
			if providerSource == nil {
				return &InvalidProviderSourceNameError{Name: *providerSourceName}
			}
		}
		newValue = providerSourceName
	}

	// Update the field
	n.defaultProviderSourceName = newValue
	return nil
}

// SetProviderSourceFactory sets the provider source factory
func (n *Namespace) SetProviderSourceFactory(factory ProviderSourceFactory) {
	n.providerSourceFactory = factory
}

// InvalidProviderSourceNameError represents an error when an invalid provider source name is provided
// Python reference: /app/terrareg/errors.py
type InvalidProviderSourceNameError struct {
	Name string
}

func (e *InvalidProviderSourceNameError) Error() string {
	return fmt.Sprintf("Provider source '%s' does not exist", e.Name)
}
