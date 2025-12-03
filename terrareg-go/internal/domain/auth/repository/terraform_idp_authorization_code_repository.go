package repository

import (
	"context"
	"time"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
)

// TerraformIdpAuthorizationCodeRepository defines the interface for OAuth authorization code storage
type TerraformIdpAuthorizationCodeRepository interface {
	Create(ctx context.Context, key string, data []byte, expiry time.Time) error
	FindByKey(ctx context.Context, key string) (*sqldb.TerraformIDPAuthorizationCodeDB, error)
	DeleteByKey(ctx context.Context, key string) error
	DeleteExpired(ctx context.Context) (int64, error)
	WithTransaction(tx interface{}) TerraformIdpAuthorizationCodeRepository
}

// TerraformIdpAuthorizationCodeRepositoryImpl implements the repository using SQL database
type TerraformIdpAuthorizationCodeRepositoryImpl struct {
	db *sqldb.Database
}

// NewTerraformIdpAuthorizationCodeRepository creates a new authorization code repository
func NewTerraformIdpAuthorizationCodeRepository(db *sqldb.Database) TerraformIdpAuthorizationCodeRepository {
	return &TerraformIdpAuthorizationCodeRepositoryImpl{
		db: db,
	}
}

// Create stores a new authorization code in the database
func (r *TerraformIdpAuthorizationCodeRepositoryImpl) Create(ctx context.Context, key string, data []byte, expiry time.Time) error {
	authCode := &sqldb.TerraformIDPAuthorizationCodeDB{
		Key:    key,
		Data:   data,
		Expiry: expiry,
	}

	return r.db.GetDB().WithContext(ctx).Create(authCode).Error
}

// FindByKey retrieves an authorization code by key if it hasn't expired
func (r *TerraformIdpAuthorizationCodeRepositoryImpl) FindByKey(ctx context.Context, key string) (*sqldb.TerraformIDPAuthorizationCodeDB, error) {
	var authCode sqldb.TerraformIDPAuthorizationCodeDB
	err := r.db.GetDB().WithContext(ctx).
		Where("key = ? AND expiry > ?", key, time.Now()).
		First(&authCode).
		Error

	if err != nil {
		return nil, err
	}

	return &authCode, nil
}

// DeleteByKey removes an authorization code by key
func (r *TerraformIdpAuthorizationCodeRepositoryImpl) DeleteByKey(ctx context.Context, key string) error {
	return r.db.GetDB().WithContext(ctx).
		Where("key = ?", key).
		Delete(&sqldb.TerraformIDPAuthorizationCodeDB{}).
		Error
}

// DeleteExpired removes all expired authorization codes and returns the count of deleted records
func (r *TerraformIdpAuthorizationCodeRepositoryImpl) DeleteExpired(ctx context.Context) (int64, error) {
	result := r.db.GetDB().WithContext(ctx).
		Where("expiry <= ?", time.Now()).
		Delete(&sqldb.TerraformIDPAuthorizationCodeDB{})

	return result.RowsAffected, result.Error
}

// WithTransaction returns a new repository instance that works within the given transaction
func (r *TerraformIdpAuthorizationCodeRepositoryImpl) WithTransaction(tx interface{}) TerraformIdpAuthorizationCodeRepository {
	// Implementation would depend on the transaction interface used
	// This is a placeholder for transaction support
	return r
}