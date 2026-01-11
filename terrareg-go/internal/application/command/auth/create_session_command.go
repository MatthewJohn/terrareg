package auth

import (
	"context"
	"encoding/json"
	"time"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth/service"
)

// CreateSessionCommand handles session creation requests
// This wraps the CookieSessionService for use in HTTP handlers
type CreateSessionCommand struct {
	cookieSessionService *service.CookieSessionService
}

// NewCreateSessionCommand creates a new CreateSessionCommand
func NewCreateSessionCommand(
	cookieSessionService *service.CookieSessionService,
) *CreateSessionCommand {
	return &CreateSessionCommand{
		cookieSessionService: cookieSessionService,
	}
}

// Execute executes the session creation command
func (c *CreateSessionCommand) Execute(ctx context.Context, req *CreateSessionRequest) (*CreateSessionResponse, error) {
	// Convert request to service request
	serviceReq := &service.CreateSessionRequest{
		AuthMethod:   req.AuthMethod,
		Username:     req.Username,
		IsAdmin:      req.IsAdmin,
		SiteAdmin:    req.SiteAdmin,
		UserGroups:   req.UserGroups,
		ProviderData: req.ProviderData,
		TTL:          req.TTL,
	}

	// Create session using service
	serviceResp, err := c.cookieSessionService.CreateSession(ctx, serviceReq)
	if err != nil {
		return nil, err
	}

	// Convert service response to command response
	return &CreateSessionResponse{
		SessionID:     serviceResp.SessionID,
		Expiry:        serviceResp.Expiry,
		Authenticated: serviceResp.Authenticated,
		Username:      serviceResp.Username,
		AuthMethod:    req.AuthMethod, // Use from request since service response doesn't have it
		IsAdmin:       serviceResp.IsAdmin,
		SiteAdmin:     serviceResp.SiteAdmin,
		UserGroups:    serviceResp.UserGroups,
	}, nil
}

// CreateSessionRequest represents a session creation request
type CreateSessionRequest struct {
	AuthMethod   string          `json:"auth_method"`
	Username     string          `json:"username,omitempty"`
	IsAdmin      bool            `json:"is_admin,omitempty"`
	SiteAdmin    bool            `json:"site_admin,omitempty"`
	UserGroups   []string        `json:"user_groups,omitempty"`
	ProviderData json.RawMessage `json:"provider_data,omitempty"`
	TTL          *time.Duration  `json:"ttl,omitempty"`
}

// CreateSessionResponse represents a session creation response
type CreateSessionResponse struct {
	SessionID     string    `json:"session_id"`
	Expiry        time.Time `json:"expiry"`
	Authenticated bool      `json:"authenticated"`
	Username      string    `json:"username,omitempty"`
	AuthMethod    string    `json:"auth_method,omitempty"`
	IsAdmin       bool      `json:"is_admin,omitempty"`
	SiteAdmin     bool      `json:"site_admin,omitempty"`
	UserGroups    []string  `json:"user_groups,omitempty"`
}
