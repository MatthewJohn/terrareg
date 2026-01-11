package repository

import (
	"context"
	"time"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
)

// TerraformIdpAccessTokenRepository defines the interface for OAuth access token storage
type TerraformIdpAccessTokenRepository interface {
	Create(ctx context.Context, key string, data []byte, expiry time.Time) error
	FindByKey(ctx context.Context, key string) (*sqldb.TerraformIDPAccessTokenDB, error)
	DeleteByKey(ctx context.Context, key string) error
	DeleteExpired(ctx context.Context) (int64, error)
	WithTransaction(tx interface{}) TerraformIdpAccessTokenRepository
}

// TerraformIdpAccessTokenRepositoryImpl implements the repository using SQL database
type TerraformIdpAccessTokenRepositoryImpl struct {
	db *sqldb.Database
}

// NewTerraformIdpAccessTokenRepository creates a new access token repository
func NewTerraformIdpAccessTokenRepository(db *sqldb.Database) TerraformIdpAccessTokenRepository {
	return &TerraformIdpAccessTokenRepositoryImpl{
		db: db,
	}
}

// Create stores a new access token in the database
func (r *TerraformIdpAccessTokenRepositoryImpl) Create(ctx context.Context, key string, data []byte, expiry time.Time) error {
	accessToken := &sqldb.TerraformIDPAccessTokenDB{
		Key:    key,
		Data:   data,
		Expiry: expiry,
	}

	return r.db.GetDB().WithContext(ctx).Create(accessToken).Error
}

// FindByKey retrieves an access token by key if it hasn't expired
func (r *TerraformIdpAccessTokenRepositoryImpl) FindByKey(ctx context.Context, key string) (*sqldb.TerraformIDPAccessTokenDB, error) {
	var accessToken sqldb.TerraformIDPAccessTokenDB
	err := r.db.GetDB().WithContext(ctx).
		Where("key = ? AND expiry > ?", key, time.Now()).
		First(&accessToken).
		Error

	if err != nil {
		return nil, err
	}

	return &accessToken, nil
}

// DeleteByKey removes an access token by key
func (r *TerraformIdpAccessTokenRepositoryImpl) DeleteByKey(ctx context.Context, key string) error {
	return r.db.GetDB().WithContext(ctx).
		Where("key = ?", key).
		Delete(&sqldb.TerraformIDPAccessTokenDB{}).
		Error
}

// DeleteExpired removes all expired access tokens and returns the count of deleted records
func (r *TerraformIdpAccessTokenRepositoryImpl) DeleteExpired(ctx context.Context) (int64, error) {
	result := r.db.GetDB().WithContext(ctx).
		Where("expiry <= ?", time.Now()).
		Delete(&sqldb.TerraformIDPAccessTokenDB{})

	return result.RowsAffected, result.Error
}

// WithTransaction returns a new repository instance that works within the given transaction
func (r *TerraformIdpAccessTokenRepositoryImpl) WithTransaction(tx interface{}) TerraformIdpAccessTokenRepository {
	// Implementation would depend on the transaction interface used
	// This is a placeholder for transaction support
	return r
}
