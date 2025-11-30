package auth

import (
	"context"

	"github.com/terrareg/terrareg/internal/domain/auth"
	authRepo "github.com/terrareg/terrareg/internal/domain/auth/repository"
)

// CheckSessionQuery handles checking if a session is valid
type CheckSessionQuery struct {
	sessionRepo authRepo.SessionRepository
}

// NewCheckSessionQuery creates a new query
func NewCheckSessionQuery(sessionRepo authRepo.SessionRepository) *CheckSessionQuery {
	return &CheckSessionQuery{
		sessionRepo: sessionRepo,
	}
}

// Execute checks if a session exists and is valid
func (q *CheckSessionQuery) Execute(ctx context.Context, sessionID string) (*auth.Session, error) {
	return q.sessionRepo.FindByID(ctx, sessionID)
}
