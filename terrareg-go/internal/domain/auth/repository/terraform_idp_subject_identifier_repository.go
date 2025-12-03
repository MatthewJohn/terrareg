package repository

import (
	"context"
	"time"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
)

// TerraformIdpSubjectIdentifierRepository defines the interface for OAuth subject identifier storage
type TerraformIdpSubjectIdentifierRepository interface {
	Create(ctx context.Context, key string, data []byte, expiry time.Time) error
	FindByKey(ctx context.Context, key string) (*sqldb.TerraformIDPSubjectIdentifierDB, error)
	DeleteByKey(ctx context.Context, key string) error
	DeleteExpired(ctx context.Context) (int64, error)
	WithTransaction(tx interface{}) TerraformIdpSubjectIdentifierRepository
}

// TerraformIdpSubjectIdentifierRepositoryImpl implements the repository using SQL database
type TerraformIdpSubjectIdentifierRepositoryImpl struct {
	db *sqldb.Database
}

// NewTerraformIdpSubjectIdentifierRepository creates a new subject identifier repository
func NewTerraformIdpSubjectIdentifierRepository(db *sqldb.Database) TerraformIdpSubjectIdentifierRepository {
	return &TerraformIdpSubjectIdentifierRepositoryImpl{
		db: db,
	}
}

// Create stores a new subject identifier in the database
func (r *TerraformIdpSubjectIdentifierRepositoryImpl) Create(ctx context.Context, key string, data []byte, expiry time.Time) error {
	subjectIdentifier := &sqldb.TerraformIDPSubjectIdentifierDB{
		Key:    key,
		Data:   data,
		Expiry: expiry,
	}

	return r.db.GetDB().WithContext(ctx).Create(subjectIdentifier).Error
}

// FindByKey retrieves a subject identifier by key if it hasn't expired
func (r *TerraformIdpSubjectIdentifierRepositoryImpl) FindByKey(ctx context.Context, key string) (*sqldb.TerraformIDPSubjectIdentifierDB, error) {
	var subjectIdentifier sqldb.TerraformIDPSubjectIdentifierDB
	err := r.db.GetDB().WithContext(ctx).
		Where("key = ? AND expiry > ?", key, time.Now()).
		First(&subjectIdentifier).
		Error

	if err != nil {
		return nil, err
	}

	return &subjectIdentifier, nil
}

// DeleteByKey removes a subject identifier by key
func (r *TerraformIdpSubjectIdentifierRepositoryImpl) DeleteByKey(ctx context.Context, key string) error {
	return r.db.GetDB().WithContext(ctx).
		Where("key = ?", key).
		Delete(&sqldb.TerraformIDPSubjectIdentifierDB{}).
		Error
}

// DeleteExpired removes all expired subject identifiers and returns the count of deleted records
func (r *TerraformIdpSubjectIdentifierRepositoryImpl) DeleteExpired(ctx context.Context) (int64, error) {
	result := r.db.GetDB().WithContext(ctx).
		Where("expiry <= ?", time.Now()).
		Delete(&sqldb.TerraformIDPSubjectIdentifierDB{})

	return result.RowsAffected, result.Error
}

// WithTransaction returns a new repository instance that works within the given transaction
func (r *TerraformIdpSubjectIdentifierRepositoryImpl) WithTransaction(tx interface{}) TerraformIdpSubjectIdentifierRepository {
	// Implementation would depend on the transaction interface used
	// This is a placeholder for transaction support
	return r
}