package service

import (
	"context"
	"time"

	"github.com/rs/zerolog"
)

// SessionCleanupService periodically cleans up expired sessions
type SessionCleanupService struct {
	sessionService  *SessionService
	logger          zerolog.Logger
	cleanupInterval time.Duration
	stopChan        chan struct{}
}

// NewSessionCleanupService creates a new session cleanup service
func NewSessionCleanupService(sessionService *SessionService, logger zerolog.Logger, cleanupInterval time.Duration) *SessionCleanupService {
	if cleanupInterval == 0 {
		cleanupInterval = 1 * time.Hour // Default cleanup interval
	}

	return &SessionCleanupService{
		sessionService:  sessionService,
		logger:          logger,
		cleanupInterval: cleanupInterval,
		stopChan:        make(chan struct{}),
	}
}

// Start begins the periodic cleanup process
func (scs *SessionCleanupService) Start(ctx context.Context) {
	scs.logger.Info().Dur("interval", scs.cleanupInterval).Msg("Starting session cleanup service")

	ticker := time.NewTicker(scs.cleanupInterval)
	defer ticker.Stop()

	// Run cleanup immediately on start
	scs.runCleanup(ctx)

	for {
		select {
		case <-ctx.Done():
			scs.logger.Info().Msg("Session cleanup service stopped due to context cancellation")
			return
		case <-scs.stopChan:
			scs.logger.Info().Msg("Session cleanup service stopped")
			return
		case <-ticker.C:
			scs.runCleanup(ctx)
		}
	}
}

// Stop stops the cleanup service
func (scs *SessionCleanupService) Stop() {
	close(scs.stopChan)
}

// runCleanup performs the actual cleanup of expired sessions
func (scs *SessionCleanupService) runCleanup(ctx context.Context) {
	scs.logger.Debug().Msg("Running session cleanup")

	if err := scs.sessionService.CleanupExpiredSessions(ctx); err != nil {
		scs.logger.Error().Err(err).Msg("Failed to cleanup expired sessions")
	} else {
		scs.logger.Debug().Msg("Session cleanup completed successfully")
	}
}
