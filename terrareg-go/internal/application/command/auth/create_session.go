package auth

import (
	"context"
	"fmt"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth/repository"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth/service"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared"
)

// CreateSessionCommand handles session creation
type CreateSessionCommand struct {
	sessionRepo repository.SessionRepository
	sessionService *service.CookieSessionService
}

// NewCreateSessionCommand creates a new CreateSessionCommand
func NewCreateSessionCommand(
	sessionRepo repository.SessionRepository,
	sessionService *service.CookieSessionService,
) *CreateSessionCommand {
	return &CreateSessionCommand{
		sessionRepo: sessionRepo,
		sessionService: sessionService,
	}
}

// Execute creates a new session for the user
func (c *CreateSessionCommand) Execute(ctx context.Context, req *CreateSessionRequest) (*CreateSessionResponse, error) {
	// Validate input
	if req == nil {
		return nil, shared.ErrInvalidInput
	}

	if req.AuthMethod == "" {
		return nil, fmt.Errorf("%w: auth method is required", shared.ErrInvalidInput)
	}

	// Create session
	sessionData, err := c.sessionService.CreateSession(
		ctx,
		req.UserID,
		req.Username,
		req.AuthMethod,
		req.IsAdmin,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Set additional session data if provided
	if req.Permissions != nil {
		sessionData.Permissions = req.Permissions
	}
	if len(req.UserGroups) > 0 {
		sessionData.UserGroups = req.UserGroups
	}

	// Store provider source auth data if provided
	if len(req.ProviderSourceAuth) > 0 {
		if err := c.sessionRepo.UpdateProviderSourceAuth(
			ctx,
			sessionData.SessionID,
			req.ProviderSourceAuth,
		); err != nil {
			// Clean up session if auth data storage fails
			c.sessionService.DeleteSession(ctx, sessionData.SessionID)
			return nil, fmt.Errorf("failed to store provider source auth data: %w", err)
		}
		sessionData.ProviderSourceAuth = req.ProviderSourceAuth
	}

	return &CreateSessionResponse{
		SessionData: sessionData,
	}, nil
}

// CreateSessionRequest represents a session creation request
type CreateSessionRequest struct {
	UserID            string            `json:"user_id,omitempty"`
	Username          string            `json:"username,omitempty"`
	AuthMethod        string            `json:"auth_method"`
	IsAdmin           bool              `json:"is_admin,omitempty"`
	Permissions       map[string]string `json:"permissions,omitempty"`
	UserGroups        []string          `json:"user_groups,omitempty"`
	ProviderSourceAuth []byte          `json:"provider_source_auth,omitempty"`
}

// CreateSessionResponse represents a session creation response
type CreateSessionResponse struct {
	SessionData *service.SessionData `json:"session_data"`
}