package model

import (
	"time"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared"
)

// OAuthTokenType represents the type of OAuth token
type OAuthTokenType int

const (
	OAuthTokenTypeAuthorizationCode OAuthTokenType = iota
	OAuthTokenTypeAccessToken
	OAuthTokenTypeRefreshToken
)

// IDPSubjectIdentifier represents a subject identifier from an Identity Provider
type IDPSubjectIdentifier struct {
	id           int
	authMethod   AuthMethod
	providerName string
	subjectID    string
	userID       string
	createdAt    time.Time
}

// IDPAuthorizationCode represents an OAuth authorization code
type IDPAuthorizationCode struct {
	id           int
	authMethod   AuthMethod
	providerName string
	code         string
	clientID     string
	redirectURI  string
	scopes       []string
	state        string
	userID       string
	createdAt    time.Time
	expiresAt    time.Time
	exchanged    bool
}

// IDPAccessToken represents an OAuth access token
type IDPAccessToken struct {
	id           int
	authMethod   AuthMethod
	providerName string
	token        string
	tokenType    string
	clientID     string
	userID       string
	scopes       []string
	createdAt    time.Time
	expiresAt    time.Time
}

// NewIDPSubjectIdentifier creates a new IDP subject identifier
func NewIDPSubjectIdentifier(authMethod AuthMethod, providerName, subjectID, userID string) (*IDPSubjectIdentifier, error) {
	if authMethod == AuthMethodNone || authMethod == AuthMethodAPIKey {
		return nil, ErrInvalidAuthMethod
	}
	if providerName == "" {
		return nil, ErrProviderNameRequired
	}
	if subjectID == "" {
		return nil, ErrSubjectIDRequired
	}
	if userID == "" {
		return nil, ErrUserIDRequired
	}

	return &IDPSubjectIdentifier{
		id:           shared.GenerateIntID(),
		authMethod:   authMethod,
		providerName: providerName,
		subjectID:    subjectID,
		userID:       userID,
		createdAt:    time.Now(),
	}, nil
}

// NewIDPAuthorizationCode creates a new authorization code
func NewIDPAuthorizationCode(authMethod AuthMethod, providerName, code, clientID, redirectURI string, scopes []string, state, userID string, ttl time.Duration) (*IDPAuthorizationCode, error) {
	if authMethod == AuthMethodNone || authMethod == AuthMethodAPIKey {
		return nil, ErrInvalidAuthMethod
	}
	if providerName == "" {
		return nil, ErrProviderNameRequired
	}
	if code == "" {
		return nil, ErrAuthorizationCodeRequired
	}
	if clientID == "" {
		return nil, ErrClientIDRequired
	}
	if redirectURI == "" {
		return nil, ErrRedirectURIRequired
	}
	if len(scopes) == 0 {
		return nil, ErrScopesRequired
	}
	if userID == "" {
		return nil, ErrUserIDRequired
	}

	now := time.Now()
	return &IDPAuthorizationCode{
		id:          shared.GenerateIntID(),
		authMethod:  authMethod,
		providerName: providerName,
		code:        code,
		clientID:    clientID,
		redirectURI: redirectURI,
		scopes:      scopes,
		state:       state,
		userID:      userID,
		createdAt:   now,
		expiresAt:   now.Add(ttl),
		exchanged:   false,
	}, nil
}

// NewIDPAccessToken creates a new access token
func NewIDPAccessToken(authMethod AuthMethod, providerName, token, tokenType, clientID string, scopes []string, userID string, ttl time.Duration) (*IDPAccessToken, error) {
	if authMethod == AuthMethodNone || authMethod == AuthMethodAPIKey {
		return nil, ErrInvalidAuthMethod
	}
	if providerName == "" {
		return nil, ErrProviderNameRequired
	}
	if token == "" {
		return nil, ErrAccessTokenRequired
	}
	if tokenType == "" {
		return nil, ErrTokenTypeRequired
	}
	if clientID == "" {
		return nil, ErrClientIDRequired
	}
	if len(scopes) == 0 {
		return nil, ErrScopesRequired
	}
	if userID == "" {
		return nil, ErrUserIDRequired
	}

	now := time.Now()
	return &IDPAccessToken{
		id:          shared.GenerateIntID(),
		authMethod:  authMethod,
		providerName: providerName,
		token:       token,
		tokenType:   tokenType,
		clientID:    clientID,
		userID:      userID,
		scopes:      scopes,
		createdAt:   now,
		expiresAt:   now.Add(ttl),
	}, nil
}

// Authorization Code methods

// IsExpired checks if the authorization code is expired
func (ac *IDPAuthorizationCode) IsExpired() bool {
	return time.Now().After(ac.expiresAt)
}

// IsValid checks if the authorization code is valid
func (ac *IDPAuthorizationCode) IsValid() bool {
	return !ac.exchanged && !ac.IsExpired()
}

// Exchange marks the authorization code as exchanged
func (ac *IDPAuthorizationCode) Exchange() error {
	if ac.exchanged {
		return ErrAuthorizationCodeAlreadyExchanged
	}
	if ac.IsExpired() {
		return ErrAuthorizationCodeExpired
	}

	ac.exchanged = true
	return nil
}

// Access Token methods

// IsExpired checks if the access token is expired
func (at *IDPAccessToken) IsExpired() bool {
	return time.Now().After(at.expiresAt)
}

// IsValid checks if the access token is valid
func (at *IDPAccessToken) IsValid() bool {
	return !at.IsExpired()
}

// HasScope checks if the token has a specific scope
func (at *IDPAccessToken) HasScope(scope string) bool {
	for _, s := range at.scopes {
		if s == scope {
			return true
		}
	}
	return false
}

// Getters for IDPSubjectIdentifier
func (idp *IDPSubjectIdentifier) ID() int           { return idp.id }
func (idp *IDPSubjectIdentifier) AuthMethod() AuthMethod { return idp.authMethod }
func (idp *IDPSubjectIdentifier) ProviderName() string { return idp.providerName }
func (idp *IDPSubjectIdentifier) SubjectID() string    { return idp.subjectID }
func (idp *IDPSubjectIdentifier) UserID() string       { return idp.userID }
func (idp *IDPSubjectIdentifier) CreatedAt() time.Time { return idp.createdAt }

// Getters for IDPAuthorizationCode
func (ac *IDPAuthorizationCode) ID() int           { return ac.id }
func (ac *IDPAuthorizationCode) AuthMethod() AuthMethod { return ac.authMethod }
func (ac *IDPAuthorizationCode) ProviderName() string { return ac.providerName }
func (ac *IDPAuthorizationCode) Code() string        { return ac.code }
func (ac *IDPAuthorizationCode) ClientID() string    { return ac.clientID }
func (ac *IDPAuthorizationCode) RedirectURI() string { return ac.redirectURI }
func (ac *IDPAuthorizationCode) Scopes() []string   { return ac.scopes }
func (ac *IDPAuthorizationCode) State() string       { return ac.state }
func (ac *IDPAuthorizationCode) UserID() string       { return ac.userID }
func (ac *IDPAuthorizationCode) CreatedAt() time.Time { return ac.createdAt }
func (ac *IDPAuthorizationCode) ExpiresAt() time.Time { return ac.expiresAt }
func (ac *IDPAuthorizationCode) Exchanged() bool     { return ac.exchanged }

// Getters for IDPAccessToken
func (at *IDPAccessToken) ID() int           { return at.id }
func (at *IDPAccessToken) AuthMethod() AuthMethod { return at.authMethod }
func (at *IDPAccessToken) ProviderName() string { return at.providerName }
func (at *IDPAccessToken) Token() string       { return at.token }
func (at *IDPAccessToken) TokenType() string  { return at.tokenType }
func (at *IDPAccessToken) ClientID() string  { return at.clientID }
func (at *IDPAccessToken) UserID() string     { return at.userID }
func (at *IDPAccessToken) Scopes() []string   { return at.scopes }
func (at *IDPAccessToken) CreatedAt() time.Time { return at.createdAt }
func (at *IDPAccessToken) ExpiresAt() time.Time { return at.expiresAt }