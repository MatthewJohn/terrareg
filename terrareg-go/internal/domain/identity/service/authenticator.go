package service

import (
	"context"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/identity/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/identity/repository"
)

// Using identity model errors instead of redefining here

// Authenticator handles user authentication
type Authenticator struct {
	userRepo       repository.UserRepository
	sessionRepo    repository.SessionRepository
	sessionManager *SessionManager
	authProviders  map[model.AuthMethod]AuthProvider
}

// AuthProvider defines the interface for authentication providers
type AuthProvider interface {
	Authenticate(ctx context.Context, request AuthRequest) (*AuthResult, error)
	GetUserInfo(ctx context.Context, token string) (*UserInfo, error)
	RefreshToken(ctx context.Context, refreshToken string) (*AuthResult, error)
}

// AuthRequest represents an authentication request
type AuthRequest struct {
	Method       model.AuthMethod
	Token        string
	Code         string
	State        string
	RedirectURI  string
	ClientID     string
	ClientSecret string
	Scopes       []string
}

// AuthResult represents the result of an authentication attempt
type AuthResult struct {
	UserID         string
	Username       string
	Email          string
	DisplayName    string
	AccessToken    string
	RefreshToken   string
	ExpiresIn      int64
	ExternalID     string
	AuthProviderID string
}

// UserInfo represents user information from an auth provider
type UserInfo struct {
	ID          string
	Username    string
	Email       string
	DisplayName string
	AvatarURL   string
	Groups      []string
}

// NewAuthenticator creates a new authenticator
func NewAuthenticator(
	userRepo repository.UserRepository,
	sessionRepo repository.SessionRepository,
	sessionManager *SessionManager,
) *Authenticator {
	return &Authenticator{
		userRepo:       userRepo,
		sessionRepo:    sessionRepo,
		sessionManager: sessionManager,
		authProviders:  make(map[model.AuthMethod]AuthProvider),
	}
}

// RegisterProvider registers an authentication provider
func (a *Authenticator) RegisterProvider(method model.AuthMethod, provider AuthProvider) {
	a.authProviders[method] = provider
}

// Authenticate authenticates a user using the specified method
func (a *Authenticator) Authenticate(ctx context.Context, request AuthRequest) (*model.Session, *model.User, error) {
	// Get auth provider
	provider, exists := a.authProviders[request.Method]
	if !exists {
		return nil, nil, model.ErrInvalidAuthMethod
	}

	// Authenticate with provider
	authResult, err := provider.Authenticate(ctx, request)
	if err != nil {
		return nil, nil, err
	}

	// Find or create user
	user, err := a.findOrCreateUser(ctx, authResult, request.Method)
	if err != nil {
		return nil, nil, err
	}

	// Create session
	session, err := a.sessionManager.CreateSession(ctx, user.ID(), request.Method, SessionMetadata{
		IPAddress: "", // Set from context if available
		UserAgent: "", // Set from context if available
		Remember:  false,
	})
	if err != nil {
		return nil, nil, err
	}

	// Update user tokens
	err = user.Authenticate(authResult.AccessToken, authResult.RefreshToken, nil)
	if err != nil {
		return nil, nil, err
	}

	// Save updated user
	err = a.userRepo.Update(ctx, user)
	if err != nil {
		return nil, nil, err
	}

	return session, user, nil
}

// ValidateToken validates an access token and returns the user
func (a *Authenticator) ValidateToken(ctx context.Context, token string) (*model.Session, *model.User, error) {
	return a.sessionManager.ValidateSession(ctx, token)
}

// RefreshToken refreshes an access token
func (a *Authenticator) RefreshToken(ctx context.Context, refreshToken string, authMethod model.AuthMethod) (*model.Session, *model.User, error) {
	// Get auth provider
	provider, exists := a.authProviders[authMethod]
	if !exists {
		return nil, nil, model.ErrInvalidAuthMethod
	}

	// Refresh token with provider
	authResult, err := provider.RefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, nil, err
	}

	// Find user by external ID or create new one
	user, err := a.findOrCreateUser(ctx, authResult, authMethod)
	if err != nil {
		return nil, nil, err
	}

	// Update user tokens
	err = user.Authenticate(authResult.AccessToken, authResult.RefreshToken, nil)
	if err != nil {
		return nil, nil, err
	}

	// Save updated user
	err = a.userRepo.Update(ctx, user)
	if err != nil {
		return nil, nil, err
	}

	// Get or create session
	sessions, err := a.sessionManager.GetActiveUserSessions(ctx, user.ID())
	if err != nil || len(sessions) == 0 {
		session, err := a.sessionManager.CreateSession(ctx, user.ID(), authMethod, SessionMetadata{})
		if err != nil {
			return nil, nil, err
		}
		return session, user, nil
	}

	return sessions[0], user, nil
}

// Logout invalidates a session
func (a *Authenticator) Logout(ctx context.Context, sessionToken string) error {
	return a.sessionManager.InvalidateSession(ctx, sessionToken)
}

// LogoutAll invalidates all sessions for a user
func (a *Authenticator) LogoutAll(ctx context.Context, userID string) error {
	return a.sessionManager.InvalidateAllUserSessions(ctx, userID)
}

// findOrCreateUser finds an existing user or creates a new one
func (a *Authenticator) findOrCreateUser(ctx context.Context, authResult *AuthResult, authMethod model.AuthMethod) (*model.User, error) {
	// Try to find user by external ID
	user, err := a.userRepo.FindByExternalID(ctx, authMethod, authResult.ExternalID)
	if err == nil {
		// User found, update tokens
		user.Authenticate(authResult.AccessToken, authResult.RefreshToken, nil)
		return user, nil
	}

	// Try to find user by email
	if authResult.Email != "" {
		user, err = a.userRepo.FindByEmail(ctx, authResult.Email)
		if err == nil {
			// User found with same email but different auth method
			if user.AuthMethod() == authMethod {
				user.Authenticate(authResult.AccessToken, authResult.RefreshToken, nil)
				return user, nil
			}
			// User exists with different auth method - error out
			return nil, model.ErrAuthenticationFailed
		}
	}

	// Create new user
	user, err = model.NewUser(
		authResult.Username,
		authResult.DisplayName,
		authResult.Email,
		authMethod,
	)
	if err != nil {
		return nil, err
	}

	// Set external ID and provider info
	user.SetExternalID(authResult.ExternalID)
	user.SetAuthProviderID(authResult.AuthProviderID)

	// Authenticate user with tokens
	err = user.Authenticate(authResult.AccessToken, authResult.RefreshToken, nil)
	if err != nil {
		return nil, err
	}

	// Save user
	err = a.userRepo.Save(ctx, user)
	if err != nil {
		return nil, err
	}

	return user, nil
}
