package model

import (
	"time"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared"
)

// TerraformSession represents a Terraform OIDC session
type TerraformSession struct {
	id              string
	authorizationCode string
	clientID        string
	redirectURI     string
	scope           string
	state           string
	codeChallenge   *string
	codeChallengeMethod *string
	nonce           *string
	expiresAt       time.Time
	createdAt       time.Time
	exchangedAt     *time.Time
	accessTokenID   *string
	subjectIdentifier string
}

// TerraformAccessToken represents a Terraform OIDC access token
type TerraformAccessToken struct {
	id           string
	tokenValue   string
	tokenType    string
	expiresAt    time.Time
	createdAt    time.Time
	scopes       []string
	subjectIdentifier string
	authorizationCodeID *string
}

// TerraformSubjectIdentifier represents a Terraform OIDC subject identifier
type TerraformSubjectIdentifier struct {
	id               string
	subject          string
	issuer           string
	authMethod       string
	createdAt        time.Time
	lastSeenAt       time.Time
	metadata         map[string]interface{}
}

// NewTerraformSession creates a new Terraform OIDC session
func NewTerraformSession(authorizationCode, clientID, redirectURI, scope, state string, codeChallenge, codeChallengeMethod, nonce *string) (*TerraformSession, error) {
	if authorizationCode == "" {
		return nil, shared.ErrInvalidInput
	}
	if clientID == "" {
		return nil, shared.ErrInvalidInput
	}
	if redirectURI == "" {
		return nil, shared.ErrInvalidInput
	}

	expiresAt := time.Now().Add(10 * time.Minute) // Authorization codes expire in 10 minutes

	return &TerraformSession{
		id:                   shared.GenerateID(),
		authorizationCode:     authorizationCode,
		clientID:             clientID,
		redirectURI:          redirectURI,
		scope:                scope,
		state:                state,
		codeChallenge:        codeChallenge,
		codeChallengeMethod:  codeChallengeMethod,
		nonce:                nonce,
		expiresAt:            expiresAt,
		createdAt:            time.Now(),
	}, nil
}

// NewTerraformAccessToken creates a new Terraform access token
func NewTerraformAccessToken(tokenValue, tokenType string, expiresAt time.Time, scopes []string, subjectIdentifier string, authorizationCodeID *string) (*TerraformAccessToken, error) {
	if tokenValue == "" {
		return nil, shared.ErrInvalidInput
	}
	if tokenType == "" {
		return nil, shared.ErrInvalidInput
	}
	if subjectIdentifier == "" {
		return nil, shared.ErrInvalidInput
	}

	return &TerraformAccessToken{
		id:                 shared.GenerateID(),
		tokenValue:         tokenValue,
		tokenType:          tokenType,
		expiresAt:          expiresAt,
		createdAt:          time.Now(),
		scopes:             scopes,
		subjectIdentifier:  subjectIdentifier,
		authorizationCodeID: authorizationCodeID,
	}, nil
}

// NewTerraformSubjectIdentifier creates a new Terraform subject identifier
func NewTerraformSubjectIdentifier(subject, issuer, authMethod string) (*TerraformSubjectIdentifier, error) {
	if subject == "" {
		return nil, shared.ErrInvalidInput
	}
	if issuer == "" {
		return nil, shared.ErrInvalidInput
	}
	if authMethod == "" {
		return nil, shared.ErrInvalidInput
	}

	return &TerraformSubjectIdentifier{
		id:         shared.GenerateID(),
		subject:    subject,
		issuer:     issuer,
		authMethod: authMethod,
		createdAt:  time.Now(),
		lastSeenAt: time.Now(),
		metadata:   make(map[string]interface{}),
	}, nil
}

// TerraformSession getters
func (ts *TerraformSession) ID() string { return ts.id }
func (ts *TerraformSession) AuthorizationCode() string { return ts.authorizationCode }
func (ts *TerraformSession) ClientID() string { return ts.clientID }
func (ts *TerraformSession) RedirectURI() string { return ts.redirectURI }
func (ts *TerraformSession) Scope() string { return ts.scope }
func (ts *TerraformSession) State() string { return ts.state }
func (ts *TerraformSession) CodeChallenge() *string { return ts.codeChallenge }
func (ts *TerraformSession) CodeChallengeMethod() *string { return ts.codeChallengeMethod }
func (ts *TerraformSession) Nonce() *string { return ts.nonce }
func (ts *TerraformSession) ExpiresAt() time.Time { return ts.expiresAt }
func (ts *TerraformSession) CreatedAt() time.Time { return ts.createdAt }
func (ts *TerraformSession) ExchangedAt() *time.Time { return ts.exchangedAt }
func (ts *TerraformSession) AccessTokenID() *string { return ts.accessTokenID }
func (ts *TerraformSession) SubjectIdentifier() string { return ts.subjectIdentifier }

// TerraformSession domain methods
func (ts *TerraformSession) IsExpired() bool {
	return time.Now().After(ts.expiresAt)
}

func (ts *TerraformSession) IsExchanged() bool {
	return ts.exchangedAt != nil
}

func (ts *TerraformSession) Exchange(accessTokenID string) error {
	if ts.IsExchanged() {
		return ErrAuthorizationCodeAlreadyExchanged
	}
	if ts.IsExpired() {
		return ErrAuthorizationCodeExpired
	}

	now := time.Now()
	ts.exchangedAt = &now
	ts.accessTokenID = &accessTokenID

	return nil
}

func (ts *TerraformSession) SetSubjectIdentifier(subjectIdentifier string) {
	ts.subjectIdentifier = subjectIdentifier
}

// TerraformAccessToken getters
func (tat *TerraformAccessToken) ID() string { return tat.id }
func (tat *TerraformAccessToken) TokenValue() string { return tat.tokenValue }
func (tat *TerraformAccessToken) TokenType() string { return tat.tokenType }
func (tat *TerraformAccessToken) ExpiresAt() time.Time { return tat.expiresAt }
func (tat *TerraformAccessToken) CreatedAt() time.Time { return tat.createdAt }
func (tat *TerraformAccessToken) Scopes() []string { return tat.scopes }
func (tat *TerraformAccessToken) SubjectIdentifier() string { return tat.subjectIdentifier }
func (tat *TerraformAccessToken) AuthorizationCodeID() *string { return tat.authorizationCodeID }

// TerraformAccessToken domain methods
func (tat *TerraformAccessToken) IsExpired() bool {
	return time.Now().After(tat.expiresAt)
}

func (tat *TerraformAccessToken) IsValid() bool {
	return !tat.IsExpired()
}

func (tat *TerraformAccessToken) HasScope(scope string) bool {
	for _, s := range tat.scopes {
		if s == scope {
			return true
		}
	}
	return false
}

// TerraformSubjectIdentifier getters
func (tsi *TerraformSubjectIdentifier) ID() string { return tsi.id }
func (tsi *TerraformSubjectIdentifier) Subject() string { return tsi.subject }
func (tsi *TerraformSubjectIdentifier) Issuer() string { return tsi.issuer }
func (tsi *TerraformSubjectIdentifier) AuthMethod() string { return tsi.authMethod }
func (tsi *TerraformSubjectIdentifier) CreatedAt() time.Time { return tsi.createdAt }
func (tsi *TerraformSubjectIdentifier) LastSeenAt() time.Time { return tsi.lastSeenAt }
func (tsi *TerraformSubjectIdentifier) Metadata() map[string]interface{} { return tsi.metadata }

// TerraformSubjectIdentifier domain methods
func (tsi *TerraformSubjectIdentifier) UpdateLastSeen() {
	tsi.lastSeenAt = time.Now()
}

func (tsi *TerraformSubjectIdentifier) SetMetadata(key string, value interface{}) {
	tsi.metadata[key] = value
}

func (tsi *TerraformSubjectIdentifier) GetMetadata(key string) (interface{}, bool) {
	value, exists := tsi.metadata[key]
	return value, exists
}

// Reconstruction functions for repository use
func ReconstructTerraformSession(id, authorizationCode, clientID, redirectURI, scope, state string, codeChallenge, codeChallengeMethod, nonce *string, expiresAt, createdAt time.Time, exchangedAt *time.Time, accessTokenID *string, subjectIdentifier string) *TerraformSession {
	return &TerraformSession{
		id:                   id,
		authorizationCode:     authorizationCode,
		clientID:             clientID,
		redirectURI:          redirectURI,
		scope:                scope,
		state:                state,
		codeChallenge:        codeChallenge,
		codeChallengeMethod:  codeChallengeMethod,
		nonce:                nonce,
		expiresAt:            expiresAt,
		createdAt:            createdAt,
		exchangedAt:          exchangedAt,
		accessTokenID:        accessTokenID,
		subjectIdentifier:    subjectIdentifier,
	}
}

func ReconstructTerraformAccessToken(id, tokenValue, tokenType string, expiresAt, createdAt time.Time, scopes []string, subjectIdentifier string, authorizationCodeID *string) *TerraformAccessToken {
	return &TerraformAccessToken{
		id:                 id,
		tokenValue:         tokenValue,
		tokenType:          tokenType,
		expiresAt:          expiresAt,
		createdAt:          createdAt,
		scopes:             scopes,
		subjectIdentifier:  subjectIdentifier,
		authorizationCodeID: authorizationCodeID,
	}
}

func ReconstructTerraformSubjectIdentifier(id, subject, issuer, authMethod string, createdAt, lastSeenAt time.Time, metadata map[string]interface{}) *TerraformSubjectIdentifier {
	if metadata == nil {
		metadata = make(map[string]interface{})
	}
	return &TerraformSubjectIdentifier{
		id:         id,
		subject:    subject,
		issuer:     issuer,
		authMethod: authMethod,
		createdAt:  createdAt,
		lastSeenAt: lastSeenAt,
		metadata:   metadata,
	}
}