package provider

import (
	"time"
)

// Provider represents a Terraform provider aggregate root
type Provider struct {
	id                  int
	namespaceID         int
	name                string
	description         *string
	tier                string
	categoryID          *int
	repositoryID        *int
	latestVersionID     *int
	useProviderSourceAuth bool
}

// NewProvider creates a new Provider instance
func NewProvider(
	id int,
	namespaceID int,
	name string,
	description *string,
	tier string,
	categoryID *int,
	repositoryID *int,
	latestVersionID *int,
	useProviderSourceAuth bool,
) *Provider {
	return &Provider{
		id:                  id,
		namespaceID:         namespaceID,
		name:                name,
		description:         description,
		tier:                tier,
		categoryID:          categoryID,
		repositoryID:        repositoryID,
		latestVersionID:     latestVersionID,
		useProviderSourceAuth: useProviderSourceAuth,
	}
}

// Getters
func (p *Provider) ID() int                          { return p.id }
func (p *Provider) NamespaceID() int                 { return p.namespaceID }
func (p *Provider) Name() string                     { return p.name }
func (p *Provider) Description() *string             { return p.description }
func (p *Provider) Tier() string                     { return p.tier }
func (p *Provider) CategoryID() *int                 { return p.categoryID }
func (p *Provider) RepositoryID() *int               { return p.repositoryID }
func (p *Provider) LatestVersionID() *int            { return p.latestVersionID }
func (p *Provider) UseProviderSourceAuth() bool      { return p.useProviderSourceAuth }

// ProviderVersion represents a provider version
type ProviderVersion struct {
	id               int
	providerID       int
	version          string
	gitTag           *string
	beta             bool
	publishedAt      *time.Time
	gpgKeyID         int
	protocolVersions []string
}

// NewProviderVersion creates a new ProviderVersion instance
func NewProviderVersion(
	id int,
	providerID int,
	version string,
	gitTag *string,
	beta bool,
	publishedAt *time.Time,
	gpgKeyID int,
	protocolVersions []string,
) *ProviderVersion {
	return &ProviderVersion{
		id:               id,
		providerID:       providerID,
		version:          version,
		gitTag:           gitTag,
		beta:             beta,
		publishedAt:      publishedAt,
		gpgKeyID:         gpgKeyID,
		protocolVersions: protocolVersions,
	}
}

// Getters
func (pv *ProviderVersion) ID() int                   { return pv.id }
func (pv *ProviderVersion) ProviderID() int           { return pv.providerID }
func (pv *ProviderVersion) Version() string           { return pv.version }
func (pv *ProviderVersion) GitTag() *string           { return pv.gitTag }
func (pv *ProviderVersion) Beta() bool                { return pv.beta }
func (pv *ProviderVersion) PublishedAt() *time.Time   { return pv.publishedAt }
func (pv *ProviderVersion) GPGKeyID() int             { return pv.gpgKeyID }
func (pv *ProviderVersion) ProtocolVersions() []string { return pv.protocolVersions }
