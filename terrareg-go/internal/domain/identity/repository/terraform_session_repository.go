package repository

import (
	"context"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth/model"
)

// TerraformSessionRepository handles persistence of Terraform OIDC sessions
type TerraformSessionRepository interface {
	// Session operations
	SaveSession(ctx context.Context, session *model.TerraformSession) error
	FindByAuthorizationCode(ctx context.Context, authorizationCode string) (*model.TerraformSession, error)
	FindByID(ctx context.Context, id string) (*model.TerraformSession, error)
	DeleteExpiredSessions(ctx context.Context) error
	DeleteSession(ctx context.Context, id string) error

	// Access token operations
	SaveAccessToken(ctx context.Context, token *model.TerraformAccessToken) error
	FindByAccessToken(ctx context.Context, tokenValue string) (*model.TerraformAccessToken, error)
	FindTokensBySubject(ctx context.Context, subjectIdentifier string) ([]*model.TerraformAccessToken, error)
	RevokeToken(ctx context.Context, tokenID string) error
	DeleteExpiredTokens(ctx context.Context) error

	// Subject identifier operations
	SaveSubjectIdentifier(ctx context.Context, subject *model.TerraformSubjectIdentifier) error
	FindBySubject(ctx context.Context, subject, issuer string) (*model.TerraformSubjectIdentifier, error)
	FindSubjectByID(ctx context.Context, id string) (*model.TerraformSubjectIdentifier, error)
	UpdateLastSeen(ctx context.Context, id string) error
}

// TerraformSessionRepositoryTx extends TerraformSessionRepository with transaction support
type TerraformSessionRepositoryTx interface {
	TerraformSessionRepository
	WithTransaction(tx interface{}) TerraformSessionRepository
}