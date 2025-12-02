package auth

import (
	"context"
	"time"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/config"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth"
	authRepo "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth/repository"
	authPersistence "github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb/auth"
)

// CreateAdminSessionCommand handles creating admin sessions
type CreateAdminSessionCommand struct {
	sessionRepo authRepo.SessionRepository
	config      *config.Config
}

// NewCreateAdminSessionCommand creates a new command
func NewCreateAdminSessionCommand(sessionRepo authRepo.SessionRepository, cfg *config.Config) *CreateAdminSessionCommand {
	return &CreateAdminSessionCommand{
		sessionRepo: sessionRepo,
		config:      cfg,
	}
}

// Execute creates a new admin session
func (c *CreateAdminSessionCommand) Execute(ctx context.Context) (*auth.Session, error) {
	// Generate session ID
	sessionID, err := authPersistence.GenerateSessionID()
	if err != nil {
		return nil, err
	}

	// Calculate expiry time
	expiry := time.Now().Add(time.Duration(c.config.AdminSessionExpiryMins) * time.Minute)

	// Create session
	session := auth.NewSession(sessionID, expiry)

	// Save to database
	if err := c.sessionRepo.Create(ctx, session); err != nil {
		return nil, err
	}

	// Cleanup old sessions opportunistically
	_ = c.sessionRepo.CleanupExpired(ctx)

	return session, nil
}
