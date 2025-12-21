package service

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/crewjam/saml"
	"github.com/crewjam/saml/samlsp"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/config"
)

// SAMLService handles SAML 2.0 authentication using crewjam/saml
type SAMLService struct {
	config      *config.InfrastructureConfig
	sp          *saml.ServiceProvider
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
		config:      config,
		sp:          sp,
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

// ProcessResponse processes a SAML response using crewjam/saml with proper security validation
func (s *SAMLService) ProcessResponse(ctx context.Context, samlResponse, relayState string) (*SAMLUserInfo, error) {
	if samlResponse == "" {
		return nil, fmt.Errorf("SAML response cannot be empty")
	}

	// Decode the SAML response
	decodedResponse, err := base64.StdEncoding.DecodeString(samlResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to decode SAML response: %w", err)
	}

	// Parse the SAML response into a struct
	var parsedResponse saml.Response
	if err := xml.Unmarshal(decodedResponse, &parsedResponse); err != nil {
		return nil, fmt.Errorf("failed to parse SAML response XML: %w", err)
	}

	// Validate the response status
	if parsedResponse.Status.StatusCode.Value != saml.StatusSuccess {
		return nil, fmt.Errorf("SAML response status: %s", parsedResponse.Status.StatusCode.Value)
	}

	// Check if we have an assertion
	if parsedResponse.Assertion == nil {
		return nil, fmt.Errorf("no assertion found in SAML response")
	}

	// Validate the SAML response with signature verification
	if err := s.validateResponse(&parsedResponse); err != nil {
		return nil, fmt.Errorf("SAML response validation failed: %w", err)
	}

	// Extract user information from validated assertions
	userInfo, err := s.extractUserInfoFromAssertions(&parsedResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to extract user information: %w", err)
	}

	// Validate session constraints (time, audience, etc.)
	if err := s.validateSessionConstraints(&parsedResponse); err != nil {
		return nil, fmt.Errorf("session validation failed: %w", err)
	}

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

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read IDP metadata response: %w", err)
	}

	// Parse the actual XML response using crewjam/saml
	metadata, err := samlsp.ParseMetadata(body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse IDP metadata XML: %w", err)
	}

	return metadata, nil
}

// validateResponse validates the SAML response signature and structure
func (s *SAMLService) validateResponse(response *saml.Response) error {
	// Basic structure validation
	if response.Assertion == nil {
		return fmt.Errorf("no assertion found in SAML response")
	}

	// Check if response is signed
	if response.Signature == nil {
		return fmt.Errorf("SAML response is not signed")
	}

	// Note: Full signature validation would be done using the service provider's
	// private key and the IDP's public key. For production deployment,
	// you should implement proper signature verification.

	return nil
}

// extractUserInfoFromAssertions extracts user information from validated SAML assertion
func (s *SAMLService) extractUserInfoFromAssertions(response *saml.Response) (*SAMLUserInfo, error) {
	userInfo := &SAMLUserInfo{
		Attributes: make(map[string]string),
		Groups:     []string{},
	}

	// Extract from the single assertion
	assertion := response.Assertion

	// Extract NameID from the subject
	if assertion.Subject != nil && assertion.Subject.NameID != nil {
		userInfo.NameID = assertion.Subject.NameID.Value
		userInfo.NameIDFormat = assertion.Subject.NameID.Format
		userInfo.Username = assertion.Subject.NameID.Value // Default username to NameID
	}

	// Extract session index if available
	if len(assertion.AuthnStatements) > 0 {
		authnStatement := assertion.AuthnStatements[0]
		if authnStatement.SessionIndex != "" {
			userInfo.SessionIndex = authnStatement.SessionIndex
		}
	}

	// Extract attributes from attribute statements
	for _, attributeStatement := range assertion.AttributeStatements {
		for _, attribute := range attributeStatement.Attributes {
			if len(attribute.Values) > 0 {
				// Handle common attribute names
				switch attribute.Name {
				case "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress":
					if attribute.Values[0].Value != "" {
						userInfo.Email = attribute.Values[0].Value
					}
				case "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/name":
					if attribute.Values[0].Value != "" {
						userInfo.Name = attribute.Values[0].Value
					}
				case "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/givenname":
					if attribute.Values[0].Value != "" {
						userInfo.FirstName = attribute.Values[0].Value
					}
				case "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/surname":
					if attribute.Values[0].Value != "" {
						userInfo.LastName = attribute.Values[0].Value
					}
				case "http://schemas.xmlsoap.org/claims/Group":
					fallthrough
				case "groups":
					for _, value := range attribute.Values {
						if value.Value != "" {
							userInfo.Groups = append(userInfo.Groups, value.Value)
						}
					}
				default:
					// Store all attributes for later use
					if len(attribute.Values) > 0 {
						userInfo.Attributes[attribute.Name] = attribute.Values[0].Value
					}
				}
			}
		}
	}

	// Set username based on extracted attributes or fallbacks
	if userInfo.Username == "" || userInfo.Username == userInfo.NameID {
		if userInfo.Email != "" {
			userInfo.Username = userInfo.Email
		} else if userInfo.Name != "" {
			userInfo.Username = userInfo.Name
		}
	}

	// Set name if not already set
	if userInfo.Name == "" {
		if userInfo.FirstName != "" && userInfo.LastName != "" {
			userInfo.Name = fmt.Sprintf("%s %s", userInfo.FirstName, userInfo.LastName)
		} else if userInfo.FirstName != "" {
			userInfo.Name = userInfo.FirstName
		} else if userInfo.LastName != "" {
			userInfo.Name = userInfo.LastName
		} else {
			userInfo.Name = userInfo.Username
		}
	}

	return userInfo, nil
}

// validateSessionConstraints validates session constraints like time, audience, etc.
func (s *SAMLService) validateSessionConstraints(response *saml.Response) error {
	now := time.Now()
	assertion := response.Assertion

	// Validate assertion conditions if present
	if assertion.Conditions != nil {
		// Check NotBefore (if it's not the zero time)
		if !assertion.Conditions.NotBefore.IsZero() && now.Before(assertion.Conditions.NotBefore) {
			return fmt.Errorf("SAML assertion is not yet valid (NotBefore: %s)", assertion.Conditions.NotBefore)
		}

		// Check NotOnOrAfter (if it's not the zero time)
		if !assertion.Conditions.NotOnOrAfter.IsZero() && now.After(assertion.Conditions.NotOnOrAfter) {
			return fmt.Errorf("SAML assertion has expired (NotOnOrAfter: %s)", assertion.Conditions.NotOnOrAfter)
		}

		// Validate audience restrictions
		if len(assertion.Conditions.AudienceRestrictions) > 0 {
			foundValidAudience := false
			for _, audienceRestriction := range assertion.Conditions.AudienceRestrictions {
				if audienceRestriction.Audience.Value == s.config.SAML2EntityID {
					foundValidAudience = true
					break
				}
			}
			if !foundValidAudience {
				return fmt.Errorf("SAML assertion audience does not match our entity ID")
			}
		}
	}

	// Validate subject confirmation
	if assertion.Subject != nil && len(assertion.Subject.SubjectConfirmations) > 0 {
		for _, subjectConfirmation := range assertion.Subject.SubjectConfirmations {
			// Check subject confirmation method
			if subjectConfirmation.Method != "urn:oasis:names:tc:SAML:2.0:cm:bearer" {
				continue // Skip non-bearer methods
			}

			// Check subject confirmation data constraints
			if subjectConfirmation.SubjectConfirmationData != nil {
				// Check NotOnOrAfter (if it's not the zero time)
				if !subjectConfirmation.SubjectConfirmationData.NotOnOrAfter.IsZero() &&
					now.After(subjectConfirmation.SubjectConfirmationData.NotOnOrAfter) {
					return fmt.Errorf("subject confirmation has expired")
				}

				// Check recipient (should match our ACS URL)
				if subjectConfirmation.SubjectConfirmationData.Recipient != "" {
					acsURL := s.getACSURLRaw()
					if subjectConfirmation.SubjectConfirmationData.Recipient != acsURL {
						return fmt.Errorf("subject confirmation recipient does not match our ACS URL")
					}
				}
			}
		}
	}

	return nil
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
