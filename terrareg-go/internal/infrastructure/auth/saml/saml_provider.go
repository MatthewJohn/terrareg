package saml

import (
	"context"
	"encoding/base64"
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/crewjam/saml"
	"github.com/crewjam/saml/samlsp"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/identity/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/identity/service"
)

var (
	ErrSAMLNotConfigured    = errors.New("SAML provider not configured")
	ErrSAMLInvalidResponse  = errors.New("invalid SAML response")
	ErrSAMLAuthentication = errors.New("SAML authentication failed")
)

// SAMLProvider implements SAML2 authentication
type SAMLProvider struct {
	sp       *samlsp.ServiceProvider
	metadata *saml.EntityDescriptor
	config   SAMLConfig
}

// SAMLConfig holds SAML configuration
type SAMLConfig struct {
	EntityID          string
	KeyFile           string
	CertificateFile   string
	AllowIDPInitiated bool
	SSOURL           string
	SLOURL           string
	NameIDFormat      string
	AttributeMapping  map[string]string
	SessionTimeout    time.Duration
	ServiceProviderURL string
}

// SAMLAuthRequest represents a SAML authentication request
type SAMLAuthRequest struct {
	IDP     string
	Redirect string
}

// NewSAMLProvider creates a new SAML provider
func NewSAMLProvider(config SAMLConfig) (*SAMLProvider, error) {
	if config.KeyFile == "" || config.CertificateFile == "" {
		return nil, ErrSAMLNotConfigured
	}

	// Create SAML service provider
	opts := samlsp.Options{
		URL:               *mustParseURL(config.ServiceProviderURL),
		Key:                config.KeyFile,
		Certificate:         config.CertificateFile,
		AllowIDPInitiated:    config.AllowIDPInitiated,
		SignRequest:         true,
		SignResponse:        true,
		DefaultRedirectURI:   config.SSOURL,
	}

	sp, err := samlsp.New(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create SAML service provider: %w", err)
	}

	provider := &SAMLProvider{
		sp:     sp,
		config: config,
	}

	// Generate metadata if needed
	if provider.metadata == nil {
		provider.metadata = sp.Metadata()
	}

	return provider, nil
}

// GetAuthURL returns the SAML authentication URL
func (p *SAMLProvider) GetAuthURL(ctx context.Context, state string) (string, error) {
	// For SAML, we need to handle IDP-initiated SSO or return a URL that initiates SAML
	if p.config.SSOURL != "" {
		// For SP-initiated SSO, return the configured SSO URL
		return p.config.SSOURL, nil
	}

	// Generate SAML authentication request
	authRequest, err := p.sp.MakeAuthenticationRequest(samlsp.IdSSOBindingRedirect)
	if err != nil {
		return "", fmt.Errorf("failed to create SAML auth request: %w", err)
	}

	// Build redirect URL with SAML request
	redirectURL, err := authRequest.Redirect("")
	if err != nil {
		return "", fmt.Errorf("failed to build SAML redirect URL: %w", err)
	}

	// Add state parameter if provided
	if state != "" {
		redirectURL += "&RelayState=" + url.QueryEscape(state)
	}

	return redirectURL, nil
}

// Authenticate handles SAML authentication response
func (p *SAMLProvider) Authenticate(ctx context.Context, request *http.Request) (*service.AuthResult, error) {
	// Parse SAML response
	assertion, err := p.sp.ParseResponse(request, p.getOptionsFromRequest(request))
	if err != nil {
		return nil, fmt.Errorf("failed to parse SAML response: %w", err)
	}

	// Extract user information from SAML assertion
	userInfo, err := p.extractUserInfo(assertion)
	if err != nil {
		return nil, fmt.Errorf("failed to extract user info from SAML assertion: %w", err)
	}

	// Create auth result
	return &service.AuthResult{
		UserID:         userInfo.ID,
		Username:       userInfo.Username,
		Email:          userInfo.Email,
		DisplayName:    userInfo.DisplayName,
		ExternalID:     userInfo.ID,
		AuthProviderID: "saml",
		AccessToken:    "", // SAML doesn't use access tokens
		RefreshToken:   "", // SAML doesn't use refresh tokens
		ExpiresIn:      int64(p.config.SessionTimeout.Seconds()),
	}, nil
}

// GetMetadata returns the SAML metadata
func (p *SAMLProvider) GetMetadata(ctx context.Context) (string, error) {
	if p.metadata == nil {
		return "", errors.New("SAML metadata not available")
	}

	metadataBytes, err := xml.Marshal(p.metadata)
	if err != nil {
		return "", fmt.Errorf("failed to marshal SAML metadata: %w", err)
	}

	return string(metadataBytes), nil
}

// HandleLogout handles SAML logout request
func (p *SAMLProvider) HandleLogout(ctx context.Context, request *http.Request) (string, error) {
	// Parse SAML logout request
	logoutRequest, err := p.sp.ParseLogoutRequest(request, p.getOptionsFromRequest(request))
	if err != nil {
		return "", fmt.Errorf("failed to parse SAML logout request: %w", err)
	}

	// Get logout URL
	logoutURL, err := logoutRequest.Redirect("")
	if err != nil {
		return "", fmt.Errorf("failed to build SAML logout URL: %w", err)
	}

	return logoutURL, nil
}

// HandleLogoutResponse handles SAML logout response
func (p *SAMLProvider) HandleLogoutResponse(ctx context.Context, request *http.Request) error {
	_, err := p.sp.ParseLogoutResponse(request, p.getOptionsFromRequest(request))
	if err != nil {
		return fmt.Errorf("failed to parse SAML logout response: %w", err)
	}

	return nil
}

// extractUserInfo extracts user information from SAML assertion
func (p *SAMLProvider) extractUserInfo(assertion *saml.Assertion) (*service.UserInfo, error) {
	userInfo := &service.UserInfo{}

	// Extract NameID as the primary user identifier
	if assertion.Subject != nil && assertion.Subject.NameID != "" {
		userInfo.ID = assertion.Subject.NameID
		userInfo.Username = assertion.Subject.NameID
	}

	// Extract attributes from assertion
	for _, statement := range assertion.AttributeStatements {
		for _, attribute := range statement.Attributes {
			value := p.getAttributeValue(attribute)

			// Map attributes using configuration
			if mappedKey, exists := p.config.AttributeMapping[attribute.Name]; exists {
				switch mappedKey {
				case "email":
					userInfo.Email = value
				case "display_name", "name":
					userInfo.DisplayName = value
				case "username", "uid":
					userInfo.Username = value
				}
			} else {
				// Default mappings for common attribute names
				switch attribute.Name {
				case "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress":
					userInfo.Email = value
				case "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/name":
					userInfo.DisplayName = value
				case "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/upn":
					userInfo.Username = value
				}
			}
		}
	}

	// Validate required fields
	if userInfo.ID == "" {
		return nil, ErrSAMLInvalidResponse
	}

	return userInfo, nil
}

// getAttributeValue extracts value from SAML attribute
func (p *SAMLProvider) getAttributeValue(attribute *saml.Attribute) string {
	if len(attribute.Values) > 0 {
		return attribute.Values[0]
	}
	return ""
}

// getOptionsFromRequest extracts SAML options from HTTP request
func (p *SAMLProvider) getOptionsFromRequest(request *http.Request) samlsp.Options {
	// Extract RelayState from request if present
	relayState := request.FormValue("RelayState")

	// Extract SAMLRequest from request if present
	samlRequest := request.FormValue("SAMLRequest")

	// Extract SAMLResponse from request if present
	samlResponse := request.FormValue("SAMLResponse")

	options := samlsp.Options{}
	if relayState != "" {
		options.RelayState = relayState
	}
	if samlRequest != "" {
		options.Request = base64.StdEncoding.DecodeString(samlRequest)
	}
	if samlResponse != "" {
		options.Response = base64.StdEncoding.DecodeString(samlResponse)
	}

	return options
}

// mustParseURL parses a URL or panics if invalid
func mustParseURL(rawURL string) *url.URL {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		panic(fmt.Sprintf("invalid URL %q: %v", rawURL, err))
	}
	return parsed
}