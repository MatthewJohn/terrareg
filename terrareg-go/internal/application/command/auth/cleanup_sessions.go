package auth

import (
	"context"
	"time"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth/repository"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth/service"
	"github.com/rs/zerolog"
)

// CleanupSessionsCommand handles session cleanup
type CleanupSessionsCommand struct {
	sessionRepo repository.SessionRepository
	sessionService *service.CookieSessionService
	logger       zerolog.Logger
	interval     time.Duration
}

// NewCleanupSessionsCommand creates a new CleanupSessionsCommand
func NewCleanupSessionsCommand(
	sessionRepo repository.SessionRepository,
	sessionService *service.CookieSessionService,
	logger zerolog.Logger,
	interval time.Duration,
) *CleanupSessionsCommand {
	return &CleanupSessionsCommand{
		sessionRepo: sessionRepo,
		sessionService: sessionService,
		logger: logger,
		interval: interval,
	}
}

// Execute performs cleanup of expired sessions
func (c *CleanupSessionsCommand) Execute(ctx context.Context) error {
	c.logger.Info().Msg("Starting session cleanup")

	if err := c.sessionService.CleanupExpiredSessions(ctx); err != nil {
		c.logger.Error().Err(err).Msg("Failed to cleanup expired sessions")
		return err
	}

	c.logger.Info().Msg("Session cleanup completed")
	return nil
}

// Start starts a background cleanup goroutine
func (c *CleanupSessionsCommand) Start(ctx context.Context) {
	c.logger.Info().Dur("interval", c.interval).Msg("Starting session cleanup goroutine")

	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			c.logger.Info().Msg("Session cleanup goroutine stopping")
			return
		case <-ticker.C:
			if err := c.Execute(ctx); err != nil {
				// Log error but continue the cleanup loop
				c.logger.Error().Err(err).Msg("Session cleanup failed, will retry on next interval")
			}
		}
	}
}

// CleanupOnce performs a one-time cleanup without starting the background goroutine
func (c *CleanupSessionsCommand) CleanupOnce(ctx context.Context) error {
	return c.Execute(ctx)
}