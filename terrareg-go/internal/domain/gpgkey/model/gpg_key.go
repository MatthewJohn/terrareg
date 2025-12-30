package gpgkey

import (
	"time"
)

// GPGKey represents a GPG key entity in the domain layer
type GPGKey struct {
	id             int
	namespaceID    int
	namespace      *Namespace
	asciiArmor     string
	keyID          string
	fingerprint    string
	source         string
	sourceURL      *string
	trustSignature *string
	createdAt      time.Time
	updatedAt      time.Time
}

// Namespace represents a namespace entity for GPG key scoping
type Namespace struct {
	id   int
	name string
}

// NewGPGKey creates a new GPG key entity
func NewGPGKey(
	namespaceID int,
	asciiArmor string,
	keyID string,
	fingerprint string,
) (*GPGKey, error) {
	if asciiArmor == "" {
		return nil, ErrInvalidASCIIArmor
	}
	if keyID == "" {
		return nil, ErrInvalidKeyID
	}
	if fingerprint == "" {
		return nil, ErrInvalidFingerprint
	}

	return &GPGKey{
		namespaceID:    namespaceID,
		asciiArmor:     asciiArmor,
		keyID:          keyID,
		fingerprint:    fingerprint,
		source:         "", // Default empty source
		trustSignature: nil,
		createdAt:      time.Now(),
		updatedAt:      time.Now(),
	}, nil
}

// ID returns the GPG key ID
func (k *GPGKey) ID() int {
	return k.id
}

// NamespaceID returns the namespace ID
func (k *GPGKey) NamespaceID() int {
	return k.namespaceID
}

// Namespace returns the namespace entity
func (k *GPGKey) Namespace() *Namespace {
	return k.namespace
}

// SetNamespace sets the namespace entity
func (k *GPGKey) SetNamespace(namespace *Namespace) {
	k.namespace = namespace
}

// ASCIILArmor returns the ASCII armor representation of the GPG key
func (k *GPGKey) ASCIIArmor() string {
	return k.asciiArmor
}

// KeyID returns the short key ID (last 16 characters of fingerprint)
func (k *GPGKey) KeyID() string {
	return k.keyID
}

// Fingerprint returns the full GPG fingerprint
func (k *GPGKey) Fingerprint() string {
	return k.fingerprint
}

// Source returns the source of the GPG key
func (k *GPGKey) Source() string {
	return k.source
}

// SetSource sets the source of the GPG key
func (k *GPGKey) SetSource(source string) {
	k.source = source
	k.updatedAt = time.Now()
}

// SourceURL returns the source URL of the GPG key
func (k *GPGKey) SourceURL() *string {
	return k.sourceURL
}

// SetSourceURL sets the source URL of the GPG key
func (k *GPGKey) SetSourceURL(sourceURL *string) {
	k.sourceURL = sourceURL
	k.updatedAt = time.Now()
}

// TrustSignature returns the trust signature
func (k *GPGKey) TrustSignature() *string {
	return k.trustSignature
}

// SetTrustSignature sets the trust signature
func (k *GPGKey) SetTrustSignature(trustSignature *string) {
	k.trustSignature = trustSignature
	k.updatedAt = time.Now()
}

// CreatedAt returns the creation timestamp
func (k *GPGKey) CreatedAt() time.Time {
	return k.createdAt
}

// UpdatedAt returns the last update timestamp
func (k *GPGKey) UpdatedAt() time.Time {
	return k.updatedAt
}

// SetID sets the GPG key ID (used by repository during hydration)
func (k *GPGKey) SetID(id int) {
	k.id = id
}

// Namespace entity methods

// NewNamespace creates a new namespace entity
func NewNamespace(id int, name string) *Namespace {
	return &Namespace{
		id:   id,
		name: name,
	}
}

// ID returns the namespace ID
func (n *Namespace) ID() int {
	return n.id
}

// Name returns the namespace name
func (n *Namespace) Name() string {
	return n.name
}

// Error definitions

var (
	// ErrInvalidASCIIArmor is returned when the ASCII armor is invalid
	ErrInvalidASCIIArmor = NewError("invalid ASCII armor")

	// ErrInvalidKeyID is returned when the key ID is invalid
	ErrInvalidKeyID = NewError("invalid key ID")

	// ErrInvalidFingerprint is returned when the fingerprint is invalid
	ErrInvalidFingerprint = NewError("invalid fingerprint")

	// ErrDuplicateFingerprint is returned when a GPG key with the same fingerprint already exists
	ErrDuplicateFingerprint = NewError("GPG key with this fingerprint already exists")

	// ErrGPGKeyNotFound is returned when a GPG key is not found
	ErrGPGKeyNotFound = NewError("GPG key not found")

	// ErrGPGKeyInUse is returned when attempting to delete a GPG key that is in use
	ErrGPGKeyInUse = NewError("cannot delete GPG key that is in use")
)

// GPGKeyError represents a domain error for GPG key operations
type GPGKeyError struct {
	message string
}

// NewError creates a new GPG key error
func NewError(message string) *GPGKeyError {
	return &GPGKeyError{message: message}
}

// Error implements the error interface
func (e *GPGKeyError) Error() string {
	return e.message
}
