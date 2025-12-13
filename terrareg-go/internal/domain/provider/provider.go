package provider

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// Domain errors for Provider aggregate
var (
	ErrInvalidVersion               = errors.New("invalid version format")
	ErrVersionAlreadyExists         = errors.New("version already exists")
	ErrVersionNotFound              = errors.New("version not found")
	ErrCannotRemoveLatestVersion    = errors.New("cannot remove latest version")
	ErrGPGKeyNotFound               = errors.New("GPG key not found")
	ErrGPGKeyAlreadyExists          = errors.New("GPG key already exists")
	ErrInvalidBinaryPlatform        = errors.New("invalid binary platform")
	ErrBinaryAlreadyExists          = errors.New("binary already exists for platform")
)

// GPGKey represents a GPG key value object
type GPGKey struct {
	id             int
	keyText        string
	asciiArmor     string
	keyID          string
	trustSignature *string
	createdAt      time.Time
	updatedAt      time.Time
}

// NewGPGKey creates a new GPG key
func NewGPGKey(keyText, asciiArmor, keyID, trustSignature string) (*GPGKey, error) {
	// Validate key format
	if err := ValidateGPGKeyFormat(keyText); err != nil {
		return nil, err
	}

	// Validate keyID format
	if err := ValidateKeyID(keyID); err != nil {
		return nil, err
	}

	now := time.Now()
	return &GPGKey{
		keyText:        keyText,
		asciiArmor:     asciiArmor,
		keyID:          keyID,
		trustSignature: func() *string {
			if trustSignature != "" {
				return &trustSignature
			}
			return nil
		}(),
		createdAt: now,
		updatedAt: now,
	}, nil
}

// ReconstructGPGKey reconstructs a GPG key from persistence
func ReconstructGPGKey(
	id int,
	keyText string,
	asciiArmor string,
	keyID string,
	trustSignature *string,
	createdAt time.Time,
	updatedAt time.Time,
) *GPGKey {
	return &GPGKey{
		id:             id,
		keyText:        keyText,
		asciiArmor:     asciiArmor,
		keyID:          keyID,
		trustSignature: trustSignature,
		createdAt:      createdAt,
		updatedAt:      updatedAt,
	}
}

// Getters
func (g *GPGKey) ID() int                 { return g.id }
func (g *GPGKey) KeyText() string         { return g.keyText }
func (g *GPGKey) AsciiArmor() string      { return g.asciiArmor }
func (g *GPGKey) KeyID() string           { return g.keyID }
func (g *GPGKey) TrustSignature() *string { return g.trustSignature }
func (g *GPGKey) CreatedAt() time.Time    { return g.createdAt }
func (g *GPGKey) UpdatedAt() time.Time    { return g.updatedAt }

// Setters for repository operations
func (g *GPGKey) SetID(id int)                { g.id = id }
func (g *GPGKey) SetKeyText(keyText string)   { g.keyText = keyText }
func (g *GPGKey) SetAsciiArmor(armor string)  { g.asciiArmor = armor }
func (g *GPGKey) SetKeyID(keyID string)       { g.keyID = keyID }
func (g *GPGKey) SetTrustSignature(sig *string) {
	g.trustSignature = sig
	g.updatedAt = time.Now()
}

// ProviderBinary represents a provider binary for a specific platform
type ProviderBinary struct {
	id              int
	versionID       int
	operatingSystem string
	architecture    string
	fileName        string
	fileSize        int64
	fileHash        string
	downloadURL     string
}

// NewProviderBinary creates a new ProviderBinary instance
func NewProviderBinary(
	versionID int,
	operatingSystem string,
	architecture string,
	fileName string,
	fileHash string,
	downloadURL string,
	fileSize int64,
) *ProviderBinary {
	return &ProviderBinary{
		versionID:       versionID,
		operatingSystem: operatingSystem,
		architecture:    architecture,
		fileName:        fileName,
		fileSize:        fileSize,
		fileHash:        fileHash,
		downloadURL:     downloadURL,
	}
}

// ReconstructProviderBinary reconstructs a ProviderBinary from persistence
func ReconstructProviderBinary(
	id int,
	versionID int,
	operatingSystem string,
	architecture string,
	fileName string,
	fileSize int64,
	fileHash string,
	downloadURL string,
) *ProviderBinary {
	return &ProviderBinary{
		id:              id,
		versionID:       versionID,
		operatingSystem: operatingSystem,
		architecture:    architecture,
		fileName:        fileName,
		fileSize:        fileSize,
		fileHash:        fileHash,
		downloadURL:     downloadURL,
	}
}

// Getters
func (b *ProviderBinary) ID() int                 { return b.id }
func (b *ProviderBinary) VersionID() int          { return b.versionID }
func (b *ProviderBinary) OperatingSystem() string { return b.operatingSystem }
func (b *ProviderBinary) OS() string              { return b.operatingSystem }
func (b *ProviderBinary) Architecture() string    { return b.architecture }
func (b *ProviderBinary) Arch() string            { return b.architecture }
func (b *ProviderBinary) FileName() string        { return b.fileName }
func (b *ProviderBinary) Filename() string        { return b.fileName }
func (b *ProviderBinary) FileSize() int64         { return b.fileSize }
func (b *ProviderBinary) FileHash() string        { return b.fileHash }
func (b *ProviderBinary) DownloadURL() string     { return b.downloadURL }

// Setters for repository operations
func (b *ProviderBinary) SetID(id int)                    { b.id = id }
func (b *ProviderBinary) SetVersionID(versionID int)      { b.versionID = versionID }
func (b *ProviderBinary) SetOperatingSystem(os string)    { b.operatingSystem = os }
func (b *ProviderBinary) SetArchitecture(arch string)     { b.architecture = arch }
func (b *ProviderBinary) SetFileName(fileName string)     { b.fileName = fileName }
func (b *ProviderBinary) SetFileSize(fileSize int64)      { b.fileSize = fileSize }
func (b *ProviderBinary) SetFileHash(fileHash string)     { b.fileHash = fileHash }
func (b *ProviderBinary) SetDownloadURL(downloadURL string) { b.downloadURL = downloadURL }

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
	binaries         []*ProviderBinary
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
		binaries:         make([]*ProviderBinary, 0),
	}
}

// ReconstructProviderVersion reconstructs a ProviderVersion from persistence
func ReconstructProviderVersion(
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
		binaries:         make([]*ProviderBinary, 0),
	}
}

// Getters
func (pv *ProviderVersion) ID() int                     { return pv.id }
func (pv *ProviderVersion) ProviderID() int             { return pv.providerID }
func (pv *ProviderVersion) Version() string             { return pv.version }
func (pv *ProviderVersion) GitTag() *string             { return pv.gitTag }
func (pv *ProviderVersion) Beta() bool                  { return pv.beta }
func (pv *ProviderVersion) PublishedAt() *time.Time     { return pv.publishedAt }
func (pv *ProviderVersion) GPGKeyID() int               { return pv.gpgKeyID }
func (pv *ProviderVersion) ProtocolVersions() []string  { return pv.protocolVersions }
func (pv *ProviderVersion) Binaries() []*ProviderBinary { return pv.binaries }

// Setters for repository operations
func (pv *ProviderVersion) SetID(id int)                       { pv.id = id }
func (pv *ProviderVersion) SetProviderID(providerID int)       { pv.providerID = providerID }
func (pv *ProviderVersion) SetVersion(version string)          { pv.version = version }
func (pv *ProviderVersion) SetGitTag(gitTag *string)           { pv.gitTag = gitTag }
func (pv *ProviderVersion) SetBeta(beta bool)                  { pv.beta = beta }
func (pv *ProviderVersion) SetPublishedAt(publishedAt *time.Time) { pv.publishedAt = publishedAt }
func (pv *ProviderVersion) SetGPGKeyID(gpgKeyID int)           { pv.gpgKeyID = gpgKeyID }
func (pv *ProviderVersion) SetProtocolVersions(protocolVersions []string) { pv.protocolVersions = protocolVersions }

// Provider represents a Terraform provider aggregate root
type Provider struct {
	id                    int
	namespaceID           int
	name                  string
	description           *string
	tier                  string
	categoryID            *int
	repositoryID          *int
	latestVersionID       *int
	useProviderSourceAuth bool

	// Aggregate collections - loaded lazily or via repository
	versions []*ProviderVersion
	gpgKeys  []*GPGKey
}

// NewProvider creates a new Provider instance
func NewProvider(
	namespaceID int,
	name string,
	description *string,
	tier string,
	categoryID *int,
	repositoryID *int,
	useProviderSourceAuth bool,
) *Provider {
	return &Provider{
		namespaceID:           namespaceID,
		name:                  name,
		description:           description,
		tier:                  tier,
		categoryID:            categoryID,
		repositoryID:          repositoryID,
		useProviderSourceAuth: useProviderSourceAuth,
		versions:              make([]*ProviderVersion, 0),
		gpgKeys:               make([]*GPGKey, 0),
	}
}

// ReconstructProvider reconstructs a Provider from persistence
func ReconstructProvider(
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
		id:                    id,
		namespaceID:           namespaceID,
		name:                  name,
		description:           description,
		tier:                  tier,
		categoryID:            categoryID,
		repositoryID:          repositoryID,
		latestVersionID:       latestVersionID,
		useProviderSourceAuth: useProviderSourceAuth,
		versions:              make([]*ProviderVersion, 0),
		gpgKeys:               make([]*GPGKey, 0),
	}
}

// Getters
func (p *Provider) ID() int                     { return p.id }
func (p *Provider) NamespaceID() int            { return p.namespaceID }
func (p *Provider) Name() string                { return p.name }
func (p *Provider) Description() *string        { return p.description }
func (p *Provider) Tier() string                { return p.tier }
func (p *Provider) CategoryID() *int            { return p.categoryID }
func (p *Provider) RepositoryID() *int          { return p.repositoryID }
func (p *Provider) LatestVersionID() *int       { return p.latestVersionID }
func (p *Provider) UseProviderSourceAuth() bool { return p.useProviderSourceAuth }

// Setters for repository operations
func (p *Provider) SetID(id int)                            { p.id = id }
func (p *Provider) SetNamespaceID(namespaceID int)          { p.namespaceID = namespaceID }
func (p *Provider) SetName(name string)                     { p.name = name }
func (p *Provider) SetDescription(description *string)      { p.description = description }
func (p *Provider) SetTier(tier string)                     { p.tier = tier }
func (p *Provider) SetCategoryID(categoryID *int)           { p.categoryID = categoryID }
func (p *Provider) SetRepositoryID(repositoryID *int)       { p.repositoryID = repositoryID }
func (p *Provider) SetLatestVersionID(latestVersionID *int) { p.latestVersionID = latestVersionID }
func (p *Provider) SetUseProviderSourceAuth(use bool)       { p.useProviderSourceAuth = use }

// UpdateDetails updates provider details
func (p *Provider) UpdateDetails(description *string, tier string, source *string, alias *string) {
	if description != nil {
		p.description = description
	}
	if tier != "" {
		p.tier = tier
	}
	// Note: source and alias would need to be added to the struct if needed
}

// UpdateGitConfig updates git configuration
func (p *Provider) UpdateGitConfig(gitProviderID *int, repoCloneURL *string, gitTagFormat *string, gitPath *string) {
	// Note: These fields would need to be added to the struct if needed
	// For now, this is a placeholder implementation
}

// Aggregate Root Methods

// GetVersions returns all versions of this provider
func (p *Provider) GetVersions() []*ProviderVersion {
	return p.versions
}

// GetLatestVersion returns the latest version of this provider
func (p *Provider) GetLatestVersion() *ProviderVersion {
	if p.latestVersionID == nil {
		return nil
	}

	for _, version := range p.versions {
		if version.ID() == *p.latestVersionID {
			return version
		}
	}

	// If latestVersionID is set but version not in collection, return nil
	// Repository should handle loading the version
	return nil
}

// GetVersionByNumber finds a specific version by version number
func (p *Provider) GetVersionByNumber(versionNumber string) *ProviderVersion {
	for _, version := range p.versions {
		if version.Version() == versionNumber {
			return version
		}
	}
	return nil
}

// AddVersion adds a version to the provider's version collection
func (p *Provider) AddVersion(version *ProviderVersion) {
	// Ensure this version belongs to this provider
	if version.ProviderID() != p.id {
		return
	}

	// Check if version already exists
	for _, existingVersion := range p.versions {
		if existingVersion.ID() == version.ID() {
			return // Already exists
		}
	}

	p.versions = append(p.versions, version)
}

// PublishVersion publishes a new version of the provider
func (p *Provider) PublishVersion(version string, protocolVersions []string, isBeta bool) (*ProviderVersion, error) {
	// Validate version format (basic semantic versioning)
	if version == "" {
		return nil, ErrInvalidVersion
	}

	// Check if version already exists
	if existingVersion := p.GetVersionByNumber(version); existingVersion != nil {
		return nil, ErrVersionAlreadyExists
	}

	now := time.Now()
	providerVersion := NewProviderVersion(
		0, // ID will be set by repository
		p.id,
		version,
		nil, // gitTag - will be set by repository
		isBeta,
		&now,
		0, // gpgKeyID - will be set when GPG key is assigned
		protocolVersions,
	)

	p.versions = append(p.versions, providerVersion)
	return providerVersion, nil
}

// SetLatestVersion marks a version as the latest
func (p *Provider) SetLatestVersion(versionID int) error {
	// Find the version in our collection
	var foundVersion *ProviderVersion
	for _, version := range p.versions {
		if version.ID() == versionID {
			foundVersion = version
			break
		}
	}

	if foundVersion == nil {
		return ErrVersionNotFound
	}

	p.latestVersionID = &versionID
	return nil
}

// RemoveVersion removes a version from the provider
func (p *Provider) RemoveVersion(versionID int) error {
	// Cannot remove the latest version
	if p.latestVersionID != nil && *p.latestVersionID == versionID {
		return ErrCannotRemoveLatestVersion
	}

	for i, version := range p.versions {
		if version.ID() == versionID {
			p.versions = append(p.versions[:i], p.versions[i+1:]...)
			return nil
		}
	}

	return ErrVersionNotFound
}

// GetGPGKeys returns all GPG keys for this provider
func (p *Provider) GetGPGKeys() []*GPGKey {
	return p.gpgKeys
}

// FindGPGKeyByID finds a GPG key by its database ID
func (p *Provider) FindGPGKeyByID(keyID int) *GPGKey {
	for _, gpgKey := range p.gpgKeys {
		if gpgKey.ID() == keyID {
			return gpgKey
		}
	}
	return nil
}

// FindGPGKeyByKeyID finds a GPG key by its key identifier
func (p *Provider) FindGPGKeyByKeyID(keyIdentifier string) *GPGKey {
	for _, gpgKey := range p.gpgKeys {
		if gpgKey.KeyID() == keyIdentifier {
			return gpgKey
		}
	}
	return nil
}

// AddGPGKey adds a GPG key to the provider
func (p *Provider) AddGPGKey(gpgKey *GPGKey) error {
	// Check if key with same keyID already exists
	if existingKey := p.FindGPGKeyByKeyID(gpgKey.KeyID()); existingKey != nil {
		return ErrGPGKeyAlreadyExists
	}

	p.gpgKeys = append(p.gpgKeys, gpgKey)
	return nil
}

// RemoveGPGKey removes a GPG key from the provider by database ID
func (p *Provider) RemoveGPGKey(keyID int) error {
	for i, gpgKey := range p.gpgKeys {
		if gpgKey.ID() == keyID {
			p.gpgKeys = append(p.gpgKeys[:i], p.gpgKeys[i+1:]...)
			return nil
		}
	}
	return ErrGPGKeyNotFound
}

// RemoveGPGKeyByKeyID removes a GPG key by its key identifier
func (p *Provider) RemoveGPGKeyByKeyID(keyIdentifier string) error {
	for i, gpgKey := range p.gpgKeys {
		if gpgKey.KeyID() == keyIdentifier {
			p.gpgKeys = append(p.gpgKeys[:i], p.gpgKeys[i+1:]...)
			return nil
		}
	}
	return ErrGPGKeyNotFound
}

// Validation Functions

// ValidateGPGKeyFormat validates the format of a GPG key text
func ValidateGPGKeyFormat(keyText string) error {
	if keyText == "" {
		return ErrGPGKeyNotFound
	}

	// Basic validation - check if it looks like a GPG key
	if !strings.Contains(keyText, "-----BEGIN PGP") ||
	   !strings.Contains(keyText, "-----END PGP") {
		return fmt.Errorf("invalid GPG key format: missing PGP headers")
	}

	return nil
}

// ValidateKeyID validates the format of a GPG key ID
func ValidateKeyID(keyID string) error {
	if keyID == "" {
		return ErrGPGKeyNotFound
	}

	// GPG key IDs are typically 40-character hex strings
	// or shorter 8 or 16 character versions
	keyID = strings.ReplaceAll(keyID, " ", "")

	// Check if it's a valid hex string
	hexRegex := regexp.MustCompile(`^[0-9A-Fa-f]+$`)
	if !hexRegex.MatchString(keyID) {
		return fmt.Errorf("invalid key ID format: must be hexadecimal")
	}

	// Check reasonable length (8, 16, or 40 characters)
	if len(keyID) != 8 && len(keyID) != 16 && len(keyID) != 40 {
		return fmt.Errorf("invalid key ID length: expected 8, 16, or 40 characters")
	}

	return nil
}

// ExtractKeyIDFromText attempts to extract the key ID from key text
func ExtractKeyIDFromText(keyText string) (string, error) {
	lines := strings.Split(keyText, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "Key fingerprint =") {
			// Extract the fingerprint from the line
			parts := strings.Split(line, "Key fingerprint = ")
			if len(parts) == 2 {
				fingerprint := strings.ReplaceAll(strings.TrimSpace(parts[1]), " ", "")
				// Return the last 8 characters as the key ID
				if len(fingerprint) >= 8 {
					return fingerprint[len(fingerprint)-8:], nil
				}
			}
		}
	}
	return "", fmt.Errorf("could not extract key ID from key text")
}
