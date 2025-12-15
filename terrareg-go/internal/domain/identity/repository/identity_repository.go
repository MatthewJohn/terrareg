package repository

import (
	"context"
	"errors"
)

// IdentityRepository defines operations for identity persistence
type IdentityRepository interface {
	Save(ctx context.Context, identity interface{}) error
	GetByToken(ctx context.Context, token string) (interface{}, error)
	GetByProviderID(ctx context.Context, provider string, providerID string) (interface{}, error)
	Delete(ctx context.Context, id int) error
}

// ErrNotFound is returned when a record is not found
var ErrNotFound = errors.New("record not found")
