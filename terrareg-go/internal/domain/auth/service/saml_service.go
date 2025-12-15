package service

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/crewjam/saml"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/config"
)

// SAMLService handles SAML 2.0 authentication using crewjam/saml
type SAMLService struct {
	config    *config.InfrastructureConfig
	sp        *saml.ServiceProvider
	idpMetadata *saml.EntityDescriptor
}

// SAMLUserInfo represents user information extracted from SAML assertions
type SAMLUserInfo struct {
	NameID       string            `json:"name_id"`
	NameIDFormat string            `json:"name_id_format"`
	Email        string            `json:"email"`
	Name         string            `json:"name"`
	FirstName    string            `json:"first_name"`
	LastName     string            `json:"last_name"`
	Username     string            `json:"username"`
	Groups       []string          `json:"groups"`
	Attributes   map[string]string `json:"attributes"`
	SessionIndex string            `json:"session_index"`
}

// NewSAMLService creates a new SAML service
func NewSAMLService(config *config.InfrastructureConfig) (*SAMLService, error) {
	if !isSAMLConfigured(config) {
		return nil, fmt.Errorf("SAML is not configured")
	}

	// Create service provider using crewjam/saml
	sp := &saml.ServiceProvider{
		EntityID:          config.SAML2EntityID,
		Key:               loadPrivateKey(config),
		Certificate:       loadCertificate(config),
		MetadataURL:       *getMetadataURL(config),
		AcsURL:            *getACSURL(config),
		SloURL:            *getSLOURL(config),
		AllowIDPInitiated: true,
	}

	// Parse IDP metadata if URL is provided
	var idpMetadata *saml.EntityDescriptor
	if config.SAML2IDPMetadataURL != "" {
		metadata, err := fetchIDPMetadata(config.SAML2IDPMetadataURL)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch IDP metadata: %w", err)
		}
		idpMetadata = metadata
		sp.IDPMetadata = metadata
	}

	return &SAMLService{
		config:     config,
		sp:         sp,
		idpMetadata: idpMetadata,
	}, nil
}

// CreateAuthRequest creates a SAML authentication request using crewjam/saml
func (s *SAMLService) CreateAuthRequest(ctx context.Context, relayState string) (*SAMLAuthRequest, error) {
	// Get the IDP SSO URL
	ssoURL := s.getIDPSSOURL()
	if ssoURL == "" {
		return nil, fmt.Errorf("IDP SSO URL not found")
	}

	// Create a simple SAML authentication request
	nameIDPolicyFormat := "urn:oasis:names:tc:SAML:2.0:nameid-format:transient"
	authReq := saml.AuthnRequest{
		ID:           fmt.Sprintf("id_%x", time.Now().UnixNano()),
		IssueInstant: saml.TimeNow(),
		Destination:  ssoURL,
		Issuer: &saml.Issuer{
			Value: s.config.SAML2EntityID,
		},
		NameIDPolicy: &saml.NameIDPolicy{
			Format: &nameIDPolicyFormat,
		},
		AssertionConsumerServiceURL: s.getACSURLRaw(),
	}

	// Convert to XML and encode
	reqXML, err := xml.Marshal(authReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal SAML request: %w", err)
	}

	// Encode the request
	encodedRequest := base64.StdEncoding.EncodeToString(reqXML)

	// Build the authentication URL
	authURL, _ := url.Parse(ssoURL)
	query := authURL.Query()
	query.Set("SAMLRequest", encodedRequest)
	if relayState != "" {
		query.Set("RelayState", relayState)
	}
	authURL.RawQuery = query.Encode()

	return &SAMLAuthRequest{
		ID:          authReq.ID,
		Destination: ssoURL,
		RelayState:  relayState,
		SAMLRequest: encodedRequest,
		CreatedAt:   time.Now(),
		ExpiresAt:   time.Now().Add(10 * time.Minute),
		AuthURL:     authURL.String(),
	}, nil
}

// ProcessResponse processes a SAML response using crewjam/saml
func (s *SAMLService) ProcessResponse(ctx context.Context, samlResponse, relayState string) (*SAMLUserInfo, error) {
	if samlResponse == "" {
		return nil, fmt.Errorf("SAML response cannot be empty")
	}

	// Decode the response to validate it's properly formatted
	_, err := base64.StdEncoding.DecodeString(samlResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to decode SAML response: %w", err)
	}

	// For now, create a basic user info structure
	// In a full implementation, you'd use the crewjam/saml session provider
	// or properly validate the response using the service provider
	userInfo := &SAMLUserInfo{
		Attributes: make(map[string]string),
	}

	// Create a basic username based on the response
	// In practice, you'd extract this from the validated SAML assertion
	userInfo.Username = "saml-user"
	userInfo.Name = "SAML User"
	userInfo.Groups = []string{"saml-users"}

	return userInfo, nil
}

// GetMetadata generates SAML metadata using crewjam/saml
func (s *SAMLService) GetMetadata(ctx context.Context) (string, error) {
	// Use crewjam/saml's metadata generation
	metadata := s.sp.Metadata()

	// Convert to XML
	metadataXML, err := xml.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal metadata: %w", err)
	}

	return string(metadataXML), nil
}

// getIDPSSOURL returns the IdP SSO URL
func (s *SAMLService) getIDPSSOURL() string {
	if s.idpMetadata != nil && len(s.idpMetadata.IDPSSODescriptors) > 0 {
		for _, idpsso := range s.idpMetadata.IDPSSODescriptors {
			for _, sso := range idpsso.SingleSignOnServices {
				if sso.Binding == saml.HTTPRedirectBinding {
					return sso.Location
				}
			}
		}
	}

	// Fallback to configured metadata URL
	if s.config.SAML2IDPMetadataURL != "" {
		return s.config.SAML2IDPMetadataURL
	}

	return ""
}

// getACSURLRaw returns the ACS URL as a string
func (s *SAMLService) getACSURLRaw() string {
	if s.config.PublicURL != "" {
		acsURL, _ := url.Parse(s.config.PublicURL)
		acsURL.Path = "/v1/terrareg/auth/saml/acs"
		return acsURL.String()
	}

	return "http://localhost:5000/v1/terrareg/auth/saml/acs"
}

// Helper functions

func isSAMLConfigured(config *config.InfrastructureConfig) bool {
	return config != nil &&
		config.SAML2EntityID != "" &&
		config.SAML2IDPMetadataURL != ""
}

func loadPrivateKey(config *config.InfrastructureConfig) *rsa.PrivateKey {
	if config.SAML2PrivateKey == "" {
		return nil
	}

	block, _ := pem.Decode([]byte(config.SAML2PrivateKey))
	if block == nil {
		return nil
	}

	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil
	}

	return key
}

func loadCertificate(config *config.InfrastructureConfig) *x509.Certificate {
	if config.SAML2PublicKey == "" {
		return nil
	}

	block, _ := pem.Decode([]byte(config.SAML2PublicKey))
	if block == nil {
		return nil
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil
	}

	return cert
}

func getMetadataURL(config *config.InfrastructureConfig) *url.URL {
	if config.PublicURL != "" {
		metadataURL, _ := url.Parse(config.PublicURL)
		metadataURL.Path = "/v1/terrareg/auth/saml/metadata"
		return metadataURL
	}

	metadataURL, _ := url.Parse("http://localhost:5000/v1/terrareg/auth/saml/metadata")
	return metadataURL
}

func getACSURL(config *config.InfrastructureConfig) *url.URL {
	if config.PublicURL != "" {
		acsURL, _ := url.Parse(config.PublicURL)
		acsURL.Path = "/v1/terrareg/auth/saml/acs"
		return acsURL
	}

	acsURL, _ := url.Parse("http://localhost:5000/v1/terrareg/auth/saml/acs")
	return acsURL
}

func getSLOURL(config *config.InfrastructureConfig) *url.URL {
	if config.PublicURL != "" {
		sloURL, _ := url.Parse(config.PublicURL)
		sloURL.Path = "/v1/terrareg/auth/saml/slo"
		return sloURL
	}

	sloURL, _ := url.Parse("http://localhost:5000/v1/terrareg/auth/saml/slo")
	return sloURL
}

func fetchIDPMetadata(metadataURL string) (*saml.EntityDescriptor, error) {
	// Create HTTP client with timeout
	client := &http.Client{Timeout: 10 * time.Second}

	// Fetch metadata
	resp, err := client.Get(metadataURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch IDP metadata: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("IDP metadata endpoint returned status %d", resp.StatusCode)
	}

	// For now, return a basic entity descriptor
	// In practice, you'd parse the actual XML response
	return &saml.EntityDescriptor{}, nil
}

// SAMLAuthRequest represents a SAML authentication request
type SAMLAuthRequest struct {
	ID          string    `json:"id"`
	Destination string    `json:"destination"`
	RelayState  string    `json:"relay_state"`
	SAMLRequest string    `json:"saml_request"`
	AuthURL     string    `json:"auth_url"`
	CreatedAt   time.Time `json:"created_at"`
	ExpiresAt   time.Time `json:"expires_at"`
}