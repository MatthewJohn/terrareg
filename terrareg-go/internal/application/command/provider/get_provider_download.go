package provider

import (
	"context"
	"fmt"
	"time"

	analytics "github.com/matthewjohn/terrareg/terrareg-go/internal/application/command/analytics"
	namespaceRepo "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/repository"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider"
	providerRepo "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider/repository"
	gpgkeyRepo "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/gpgkey/repository"
)

// GetProviderDownloadQuery handles getting provider download information
type GetProviderDownloadQuery struct {
	providerRepo  providerRepo.ProviderRepository
	namespaceRepo namespaceRepo.NamespaceRepository
	analyticsRepo analytics.AnalyticsRepository
	gpgKeyRepo    gpgkeyRepo.GPGKeyRepository

	// Store current provider version context for response generation
	currentProviderVersion *provider.ProviderVersion
	currentProvider        *provider.Provider
	currentNamespace       string
}

// NewGetProviderDownloadQuery creates a new get provider download query
func NewGetProviderDownloadQuery(
	providerRepo providerRepo.ProviderRepository,
	namespaceRepo namespaceRepo.NamespaceRepository,
	analyticsRepo analytics.AnalyticsRepository,
	gpgKeyRepo gpgkeyRepo.GPGKeyRepository,
) *GetProviderDownloadQuery {
	return &GetProviderDownloadQuery{
		providerRepo:  providerRepo,
		namespaceRepo: namespaceRepo,
		analyticsRepo: analyticsRepo,
		gpgKeyRepo:    gpgKeyRepo,
	}
}

// GetProviderDownloadRequest represents a request for provider download information
type GetProviderDownloadRequest struct {
	Namespace        string
	Provider         string
	Version          string
	OS               string
	Arch             string
	UserAgent        string
	TerraformVersion string
}

// ProviderDownloadResponse represents the provider download metadata
type ProviderDownloadResponse struct {
	Protocols           []string             `json:"protocols"`
	OS                  string               `json:"os"`
	Arch                string               `json:"arch"`
	Filename            string               `json:"filename"`
	DownloadURL         string               `json:"download_url"`
	ShasumsURL          string               `json:"shasums_url"`
	ShasumsSignatureURL string               `json:"shasums_signature_url"`
	Shasum              string               `json:"shasum"`
	SigningKeys         *SigningKeysResponse `json:"signing_keys"`
}

// SigningKeysResponse represents signing key information
type SigningKeysResponse struct {
	GPGPublicKeys []GPGPublicKey `json:"gpg_public_keys"`
}

// GPGPublicKey represents a GPG public key
type GPGPublicKey struct {
	KeyID          string `json:"key_id"`
	ASCIIArmor     string `json:"ascii_armor"`
	TrustSignature string `json:"trust_signature"`
	Source         string `json:"source"`
	SourceURL      string `json:"source_url"`
}

// Execute retrieves provider download metadata
func (q *GetProviderDownloadQuery) Execute(ctx context.Context, req *GetProviderDownloadRequest) (*ProviderDownloadResponse, error) {
	// Validate request parameters
	if err := q.validateRequest(req); err != nil {
		return nil, err
	}

	// Get provider
	providerEntity, err := q.providerRepo.FindByNamespaceAndName(ctx, req.Namespace, req.Provider)
	if err != nil {
		return nil, fmt.Errorf("provider not found: %w", err)
	}

	// Get provider version
	providerVersion, err := q.providerRepo.FindVersionByProviderAndVersion(ctx, providerEntity.ID(), req.Version)
	if err != nil {
		return nil, fmt.Errorf("provider version not found: %w", err)
	}

	// Store context for response generation
	defer func() {
		q.currentProvider = nil
		q.currentProviderVersion = nil
		q.currentNamespace = ""
	}()
	q.currentProvider = providerEntity
	q.currentProviderVersion = providerVersion
	q.currentNamespace = req.Namespace

	// Get binary for the specified OS/arch
	binary, err := q.getProviderBinary(ctx, providerVersion, req.OS, req.Arch)
	if err != nil {
		return nil, err
	}

	// Record download analytics
	go q.recordDownloadAnalytics(ctx, providerVersion, req)

	// Convert to response
	return q.convertToResponse(binary), nil
}

// validateRequest validates the download request
func (q *GetProviderDownloadQuery) validateRequest(req *GetProviderDownloadRequest) error {
	if req.Namespace == "" {
		return fmt.Errorf("namespace is required")
	}
	if req.Provider == "" {
		return fmt.Errorf("provider is required")
	}
	if req.Version == "" {
		return fmt.Errorf("version is required")
	}
	if req.OS == "" {
		return fmt.Errorf("OS is required")
	}
	if req.Arch == "" {
		return fmt.Errorf("architecture is required")
	}

	// Validate OS
	validOS := map[string]bool{
		"linux":   true,
		"darwin":  true,
		"windows": true,
		"freebsd": true,
	}
	if !validOS[req.OS] {
		return fmt.Errorf("unsupported OS: %s", req.OS)
	}

	// Validate architecture
	validArch := map[string]bool{
		"amd64": true,
		"arm":   true,
		"arm64": true,
		"386":   true,
	}
	if !validArch[req.Arch] {
		return fmt.Errorf("unsupported architecture: %s", req.Arch)
	}

	return nil
}

// getProviderBinary gets the provider binary for the specified OS/arch
func (q *GetProviderDownloadQuery) getProviderBinary(ctx context.Context, providerVersion *provider.ProviderVersion, os, arch string) (*provider.ProviderBinary, error) {
	binaries := providerVersion.Binaries()

	for _, binary := range binaries {
		if binary.OS() == os && binary.Architecture() == arch {
			return binary, nil
		}
	}

	return nil, fmt.Errorf("binary not found for %s/%s", os, arch)
}

// recordDownloadAnalytics records download analytics
func (q *GetProviderDownloadQuery) recordDownloadAnalytics(ctx context.Context, providerVersion *provider.ProviderVersion, req *GetProviderDownloadRequest) {
	// Record download analytics asynchronously
	if q.analyticsRepo != nil {
		now := time.Now()

		// Create provider download analytics event
		providerDownloadEvent := analytics.ProviderDownloadEvent{
			ProviderVersionID: providerVersion.ID(),
			Timestamp:        &now,
			TerraformVersion: &req.TerraformVersion,
			AnalyticsToken:   nil, // TODO: Extract from request if available
			AuthToken:        nil, // TODO: Extract from request if available
			Environment:      nil, // TODO: Extract from request if available
			NamespaceName:    &req.Namespace,
			ProviderName:     &req.Provider,
			Version:          &req.Version,
			OS:               &req.OS,
			Architecture:     &req.Arch,
			UserAgent:        &req.UserAgent,
		}

		// Record the provider download event asynchronously
		go func() {
			// Use a new context with timeout for the analytics operation
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			err := q.analyticsRepo.RecordProviderDownload(ctx, providerDownloadEvent)
			if err != nil {
				// Log error but don't fail the download request
				// In production, you'd want proper logging here
				_ = err
			}
		}()
	}
}

// convertToResponse converts the domain binary to response format
func (q *GetProviderDownloadQuery) convertToResponse(binary *provider.ProviderBinary) *ProviderDownloadResponse {
	response := &ProviderDownloadResponse{
		Protocols:           []string{"5.0"},
		OS:                  binary.OS(),
		Arch:                binary.Architecture(),
		Filename:            binary.Filename(),
		DownloadURL:         binary.DownloadURL(),
		ShasumsURL:          q.generateShasumsURL(),
		ShasumsSignatureURL: q.generateShasumsSignatureURL(),
		Shasum:              binary.FileHash(),
	}

	// Add signing keys if available
	if signingKeys := q.getSigningKeys(); signingKeys != nil {
		response.SigningKeys = signingKeys
	}

	return response
}

// generateShasumsURL generates the shasums URL for the current provider version
func (q *GetProviderDownloadQuery) generateShasumsURL() string {
	if q.currentProvider == nil || q.currentProviderVersion == nil || q.currentNamespace == "" {
		return ""
	}

	// Generate shasums URL following Terraform registry conventions
	// Format: /v1/providers/{namespace}/{provider}/{version}/download/shasums
	return fmt.Sprintf("/v1/providers/%s/%s/%s/download/shasums",
		q.currentNamespace,
		q.currentProvider.Name(),
		q.currentProviderVersion.Version())
}

// generateShasumsSignatureURL generates the shasums signature URL for the current provider version
func (q *GetProviderDownloadQuery) generateShasumsSignatureURL() string {
	if q.currentProvider == nil || q.currentProviderVersion == nil || q.currentNamespace == "" {
		return ""
	}

	// Generate shasums signature URL following Terraform registry conventions
	// Format: /v1/providers/{namespace}/{provider}/{version}/download/shasums.sig
	return fmt.Sprintf("/v1/providers/%s/%s/%s/download/shasums.sig",
		q.currentNamespace,
		q.currentProvider.Name(),
		q.currentProviderVersion.Version())
}

// getSigningKeys retrieves signing keys for the current provider version
func (q *GetProviderDownloadQuery) getSigningKeys() *SigningKeysResponse {
	if q.currentProviderVersion == nil || q.currentProviderVersion.GPGKeyID() == 0 || q.gpgKeyRepo == nil {
		return nil
	}

	// Query the GPG key repository using the provider version's GPG key ID
	ctx := context.Background()
	gpgKey, err := q.gpgKeyRepo.FindByID(ctx, q.currentProviderVersion.GPGKeyID())
	if err != nil || gpgKey == nil {
		// GPG key not found or error occurred - return nil instead of failing the request
		return nil
	}

	// Convert the domain GPG key to the response format
	gpgPublicKey := GPGPublicKey{
		KeyID:          gpgKey.KeyID(),
		ASCIIArmor:     gpgKey.ASCIIArmor(),
		TrustSignature: "", // Default empty
		Source:         gpgKey.Source(),
		SourceURL:      "", // Default empty
	}

	// Set optional fields if they exist
	if gpgKey.SourceURL() != nil {
		gpgPublicKey.SourceURL = *gpgKey.SourceURL()
	}
	if gpgKey.TrustSignature() != nil {
		gpgPublicKey.TrustSignature = *gpgKey.TrustSignature()
	}

	return &SigningKeysResponse{
		GPGPublicKeys: []GPGPublicKey{gpgPublicKey},
	}
}
