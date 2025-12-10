package provider

import (
	"context"
	"fmt"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider"
	providerRepo "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider/repository"
	namespaceRepo "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/repository"
	analytics "github.com/matthewjohn/terrareg/terrareg-go/internal/application/command/analytics"
)

// GetProviderDownloadQuery handles getting provider download information
type GetProviderDownloadQuery struct {
	providerRepo        providerRepo.ProviderRepository
	namespaceRepo       namespaceRepo.NamespaceRepository
	analyticsRepo       analytics.AnalyticsRepository
}

// NewGetProviderDownloadQuery creates a new get provider download query
func NewGetProviderDownloadQuery(
	providerRepo providerRepo.ProviderRepository,
	namespaceRepo namespaceRepo.NamespaceRepository,
	analyticsRepo analytics.AnalyticsRepository,
) *GetProviderDownloadQuery {
	return &GetProviderDownloadQuery{
		providerRepo:  providerRepo,
		namespaceRepo: namespaceRepo,
		analyticsRepo: analyticsRepo,
	}
}

// GetProviderDownloadRequest represents a request for provider download information
type GetProviderDownloadRequest struct {
	Namespace string
	Provider  string
	Version   string
	OS        string
	Arch      string
	UserAgent string
	TerraformVersion string
}

// ProviderDownloadResponse represents the provider download metadata
type ProviderDownloadResponse struct {
	Protocols               []string               `json:"protocols"`
	OS                      string                  `json:"os"`
	Arch                    string                  `json:"arch"`
	Filename                string                  `json:"filename"`
	DownloadURL             string                  `json:"download_url"`
	ShasumsURL             string                  `json:"shasums_url"`
	ShasumsSignatureURL    string                  `json:"shasums_signature_url"`
	Shasum                  string                  `json:"shasum"`
	SigningKeys             *SigningKeysResponse    `json:"signing_keys"`
}

// SigningKeysResponse represents signing key information
type SigningKeysResponse struct {
	GPGPublicKeys []GPGPublicKey `json:"gpg_public_keys"`
}

// GPGPublicKey represents a GPG public key
type GPGPublicKey struct {
	KeyID         string `json:"key_id"`
	ASCIIArmor    string `json:"ascii_armor"`
	TrustSignature string `json:"trust_signature"`
	Source        string `json:"source"`
	SourceURL     string `json:"source_url"`
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
	// TODO: Implement analytics recording
	// This would call analyticsRepo.RecordProviderDownload()
}

// convertToResponse converts the domain binary to response format
func (q *GetProviderDownloadQuery) convertToResponse(binary *provider.ProviderBinary) *ProviderDownloadResponse {
	response := &ProviderDownloadResponse{
		Protocols:               []string{"5.0"},
		OS:                      binary.OS(),
		Arch:                    binary.Architecture(),
		Filename:                binary.Filename(),
		DownloadURL:             binary.DownloadURL(),
		ShasumsURL:             "", // TODO: Generate from provider version
		ShasumsSignatureURL:    "", // TODO: Generate from provider version
		Shasum:                  binary.FileHash(),
	}

	// TODO: Add signing keys if available when GPG key support is implemented
	// if signingKeys := binary.SigningKeys(); signingKeys != nil {
	//     response.SigningKeys = &SigningKeysResponse{
	//         GPGPublicKeys: []GPGPublicKey{
	//             {
	//                 KeyID:         signingKeys.KeyID(),
	//                 ASCIIArmor:    signingKeys.ASCCIArmor(),
	//                 TrustSignature: signingKeys.TrustSignature(),
	//                 Source:        signingKeys.Source(),
	//                 SourceURL:     signingKeys.SourceURL(),
	//             },
	//         },
	//     }
	// }

	return response
}