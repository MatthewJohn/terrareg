package model

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
)

// RepositoryKind represents the kind of repository (module or provider)
// Python reference: repository_kind.py
type RepositoryKind string

const (
	RepositoryKindNone     RepositoryKind = ""
	RepositoryKindModule   RepositoryKind = "module"
	RepositoryKindProvider RepositoryKind = "provider"
)

// tagRegex matches semantic version tags (v1.2.3 or v1.2.3-beta)
// Python reference: repository_release_metadata.py::_TAG_REGEX
var tagRegex = regexp.MustCompile(`^v([0-9]+\.[0-9]+\.[0-9]+(:?-[a-z0-9]+)?)$`)

// Repository represents a GitHub repository from a provider source
// Python reference: repository_model.py::Repository
type Repository struct {
	id                 int
	providerID         *string
	owner              string
	name               string
	description        *string
	cloneURL           *string
	logoURL            *string
	providerSourceName string
}

// NewRepository creates a new repository entity
// Python reference: repository_model.py::Repository.create()
func NewRepository(
	providerID *string,
	owner string,
	name string,
	description *string,
	cloneURL *string,
	logoURL *string,
	providerSourceName string,
) (*Repository, error) {
	if err := ValidateRepositoryOwner(owner); err != nil {
		return nil, err
	}
	if err := ValidateRepositoryName(name); err != nil {
		return nil, err
	}
	if providerSourceName == "" {
		return nil, fmt.Errorf("provider source name cannot be empty")
	}

	return &Repository{
		providerID:         providerID,
		owner:              owner,
		name:               name,
		description:        description,
		cloneURL:           cloneURL,
		logoURL:            logoURL,
		providerSourceName: providerSourceName,
	}, nil
}

// ReconstructRepository reconstructs a repository from persistence
// Used by repository implementation when loading from database
func ReconstructRepository(
	id int,
	providerID *string,
	owner string,
	name string,
	description *string,
	cloneURL *string,
	logoURL *string,
	providerSourceName string,
) *Repository {
	return &Repository{
		id:                 id,
		providerID:         providerID,
		owner:              owner,
		name:               name,
		description:        description,
		cloneURL:           cloneURL,
		logoURL:            logoURL,
		providerSourceName: providerSourceName,
	}
}

// GetID returns the string ID of the repository (owner/name format)
// Python reference: repository_model.py::Repository.id
func (r *Repository) GetID() string {
	return fmt.Sprintf("%s/%s", r.owner, r.name)
}

// ID returns the database primary key
// Python reference: repository_model.py::Repository.pk
func (r *Repository) ID() int {
	return r.id
}

// ProviderID returns the provider ID
// Python reference: repository_model.py::Repository.provider_id
func (r *Repository) ProviderID() *string {
	return r.providerID
}

// Owner returns the repository owner
// Python reference: repository_model.py::Repository.owner
func (r *Repository) Owner() string {
	return r.owner
}

// Name returns the repository name
// Python reference: repository_model.py::Repository.name
func (r *Repository) Name() string {
	return r.name
}

// Description returns the repository description
// Python reference: repository_model.py::Repository.description
func (r *Repository) Description() *string {
	return r.description
}

// CloneURL returns the clone URL
// Python reference: repository_model.py::Repository.clone_url
func (r *Repository) CloneURL() *string {
	return r.cloneURL
}

// LogoURL returns the logo URL
// Python reference: repository_model.py::Repository.logo_url
func (r *Repository) LogoURL() *string {
	return r.logoURL
}

// ProviderSourceName returns the provider source name
// Python reference: repository_model.py::Repository.provider_source
func (r *Repository) ProviderSourceName() string {
	return r.providerSourceName
}

// Kind returns the repository kind based on the repository name
// Python reference: repository_model.py::Repository.kind
func (r *Repository) Kind() RepositoryKind {
	if strings.HasPrefix(r.name, "terraform-provider-") {
		return RepositoryKindProvider
	}
	if strings.HasPrefix(r.name, "terraform-") {
		return RepositoryKindModule
	}
	return RepositoryKindNone
}

// SetID sets the database ID (used by repository after persistence)
func (r *Repository) SetID(id int) {
	r.id = id
}

// SetProviderID sets the provider ID
func (r *Repository) SetProviderID(providerID *string) {
	r.providerID = providerID
}

// SetOwner sets the repository owner
func (r *Repository) SetOwner(owner string) {
	r.owner = owner
}

// SetName sets the repository name
func (r *Repository) SetName(name string) {
	r.name = name
}

// SetDescription sets the repository description
func (r *Repository) SetDescription(description *string) {
	r.description = description
}

// SetCloneURL sets the clone URL
func (r *Repository) SetCloneURL(cloneURL *string) {
	r.cloneURL = cloneURL
}

// SetLogoURL sets the logo URL
func (r *Repository) SetLogoURL(logoURL *string) {
	r.logoURL = logoURL
}

// SetProviderSourceName sets the provider source name
func (r *Repository) SetProviderSourceName(providerSourceName string) {
	r.providerSourceName = providerSourceName
}

// ToDBModel converts the domain entity to a database model
func (r *Repository) ToDBModel() *sqldb.RepositoryDB {
	return &sqldb.RepositoryDB{
		ID:                 r.id,
		ProviderID:         r.providerID,
		Owner:              &r.owner,
		Name:               &r.name,
		Description:        encodeStringToBlob(r.description),
		CloneURL:           r.cloneURL,
		LogoURL:            r.logoURL,
		ProviderSourceName: r.providerSourceName,
	}
}

// encodeStringToBlob converts a string pointer to a byte slice for blob storage
func encodeStringToBlob(s *string) []byte {
	if s == nil {
		return []byte{}
	}
	return []byte(*s)
}

// ValidateRepositoryOwner validates a repository owner name
func ValidateRepositoryOwner(owner string) error {
	if owner == "" {
		return fmt.Errorf("repository owner cannot be empty")
	}
	return nil
}

// ValidateRepositoryName validates a repository name
func ValidateRepositoryName(name string) error {
	if name == "" {
		return fmt.Errorf("repository name cannot be empty")
	}
	return nil
}

// TagToVersion converts a git tag to a semantic version
// Python reference: repository_release_metadata.py::RepositoryReleaseMetadata.tag_to_version()
func TagToVersion(tag string) *string {
	if match := tagRegex.FindStringSubmatch(tag); len(match) > 1 {
		return &match[1]
	}
	return nil
}

// VersionToTag converts a semantic version to a git tag
// Python reference: repository_release_metadata.py::RepositoryReleaseMetadata.version_to_tag()
func VersionToTag(version string) string {
	return fmt.Sprintf("v%s", version)
}
