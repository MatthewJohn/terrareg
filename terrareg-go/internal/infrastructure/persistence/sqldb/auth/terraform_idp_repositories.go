package auth

import (
	"context"
	"time"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth/repository"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
	"gorm.io/gorm"
)

// TerraformIdpAuthorizationCodeRepositoryImpl implements the Terraform IDP authorization code repository
type TerraformIdpAuthorizationCodeRepositoryImpl struct {
	db *gorm.DB
}

// NewTerraformIdpAuthorizationCodeRepository creates a new Terraform IDP authorization code repository
func NewTerraformIdpAuthorizationCodeRepository(db *gorm.DB) repository.TerraformIdpAuthorizationCodeRepository {
	return &TerraformIdpAuthorizationCodeRepositoryImpl{
		db: db,
	}
}

// Create creates a new authorization code
func (r *TerraformIdpAuthorizationCodeRepositoryImpl) Create(ctx context.Context, key string, data []byte, expiry time.Time) error {
	authCode := &sqldb.TerraformIDPAuthorizationCodeDB{
		Key:    key,
		Data:   data,
		Expiry: expiry,
	}

	return r.db.WithContext(ctx).Create(authCode).Error
}

// FindByKey finds an authorization code by key if it hasn't expired
func (r *TerraformIdpAuthorizationCodeRepositoryImpl) FindByKey(ctx context.Context, key string) (*sqldb.TerraformIDPAuthorizationCodeDB, error) {
	var authCode sqldb.TerraformIDPAuthorizationCodeDB
	err := r.db.WithContext(ctx).
		Where("key = ? AND expiry > ?", key, time.Now()).
		First(&authCode).
		Error

	if err != nil {
		return nil, err
	}

	return &authCode, nil
}

// DeleteByKey deletes an authorization code by key
func (r *TerraformIdpAuthorizationCodeRepositoryImpl) DeleteByKey(ctx context.Context, key string) error {
	return r.db.WithContext(ctx).
		Where("key = ?", key).
		Delete(&sqldb.TerraformIDPAuthorizationCodeDB{}).
		Error
}

// DeleteExpired deletes all expired authorization codes and returns the count
func (r *TerraformIdpAuthorizationCodeRepositoryImpl) DeleteExpired(ctx context.Context) (int64, error) {
	result := r.db.WithContext(ctx).
		Where("expiry <= ?", time.Now()).
		Delete(&sqldb.TerraformIDPAuthorizationCodeDB{})

	return result.RowsAffected, result.Error
}

// WithTransaction returns a new repository instance that works within the given transaction
func (r *TerraformIdpAuthorizationCodeRepositoryImpl) WithTransaction(tx interface{}) repository.TerraformIdpAuthorizationCodeRepository {
	// Implementation would depend on the transaction interface used
	// This is a placeholder for transaction support
	return r
}

// TerraformIdpAccessTokenRepositoryImpl implements the Terraform IDP access token repository
type TerraformIdpAccessTokenRepositoryImpl struct {
	db *gorm.DB
}

// NewTerraformIdpAccessTokenRepository creates a new Terraform IDP access token repository
func NewTerraformIdpAccessTokenRepository(db *gorm.DB) repository.TerraformIdpAccessTokenRepository {
	return &TerraformIdpAccessTokenRepositoryImpl{
		db: db,
	}
}

// Create creates a new access token
func (r *TerraformIdpAccessTokenRepositoryImpl) Create(ctx context.Context, key string, data []byte, expiry time.Time) error {
	accessToken := &sqldb.TerraformIDPAccessTokenDB{
		Key:    key,
		Data:   data,
		Expiry: expiry,
	}

	return r.db.WithContext(ctx).Create(accessToken).Error
}

// FindByKey finds an access token by key if it hasn't expired
func (r *TerraformIdpAccessTokenRepositoryImpl) FindByKey(ctx context.Context, key string) (*sqldb.TerraformIDPAccessTokenDB, error) {
	var accessToken sqldb.TerraformIDPAccessTokenDB
	err := r.db.WithContext(ctx).
		Where("key = ? AND expiry > ?", key, time.Now()).
		First(&accessToken).
		Error

	if err != nil {
		return nil, err
	}

	return &accessToken, nil
}

// DeleteByKey deletes an access token by key
func (r *TerraformIdpAccessTokenRepositoryImpl) DeleteByKey(ctx context.Context, key string) error {
	return r.db.WithContext(ctx).
		Where("key = ?", key).
		Delete(&sqldb.TerraformIDPAccessTokenDB{}).
		Error
}

// DeleteExpired deletes all expired access tokens and returns the count
func (r *TerraformIdpAccessTokenRepositoryImpl) DeleteExpired(ctx context.Context) (int64, error) {
	result := r.db.WithContext(ctx).
		Where("expiry <= ?", time.Now()).
		Delete(&sqldb.TerraformIDPAccessTokenDB{})

	return result.RowsAffected, result.Error
}

// WithTransaction returns a new repository instance that works within the given transaction
func (r *TerraformIdpAccessTokenRepositoryImpl) WithTransaction(tx interface{}) repository.TerraformIdpAccessTokenRepository {
	// Implementation would depend on the transaction interface used
	// This is a placeholder for transaction support
	return r
}

// TerraformIdpSubjectIdentifierRepositoryImpl implements the Terraform IDP subject identifier repository
type TerraformIdpSubjectIdentifierRepositoryImpl struct {
	db *gorm.DB
}

// NewTerraformIdpSubjectIdentifierRepository creates a new Terraform IDP subject identifier repository
func NewTerraformIdpSubjectIdentifierRepository(db *gorm.DB) repository.TerraformIdpSubjectIdentifierRepository {
	return &TerraformIdpSubjectIdentifierRepositoryImpl{
		db: db,
	}
}

// Create creates a new subject identifier
func (r *TerraformIdpSubjectIdentifierRepositoryImpl) Create(ctx context.Context, key string, data []byte, expiry time.Time) error {
	subjectIdentifier := &sqldb.TerraformIDPSubjectIdentifierDB{
		Key:    key,
		Data:   data,
		Expiry: expiry,
	}

	return r.db.WithContext(ctx).Create(subjectIdentifier).Error
}

// FindByKey finds a subject identifier by key if it hasn't expired
func (r *TerraformIdpSubjectIdentifierRepositoryImpl) FindByKey(ctx context.Context, key string) (*sqldb.TerraformIDPSubjectIdentifierDB, error) {
	var subjectIdentifier sqldb.TerraformIDPSubjectIdentifierDB
	err := r.db.WithContext(ctx).
		Where("key = ? AND expiry > ?", key, time.Now()).
		First(&subjectIdentifier).
		Error

	if err != nil {
		return nil, err
	}

	return &subjectIdentifier, nil
}

// DeleteByKey deletes a subject identifier by key
func (r *TerraformIdpSubjectIdentifierRepositoryImpl) DeleteByKey(ctx context.Context, key string) error {
	return r.db.WithContext(ctx).
		Where("key = ?", key).
		Delete(&sqldb.TerraformIDPSubjectIdentifierDB{}).
		Error
}

// DeleteExpired deletes all expired subject identifiers and returns the count
func (r *TerraformIdpSubjectIdentifierRepositoryImpl) DeleteExpired(ctx context.Context) (int64, error) {
	result := r.db.WithContext(ctx).
		Where("expiry <= ?", time.Now()).
		Delete(&sqldb.TerraformIDPSubjectIdentifierDB{})

	return result.RowsAffected, result.Error
}

// WithTransaction returns a new repository instance that works within the given transaction
func (r *TerraformIdpSubjectIdentifierRepositoryImpl) WithTransaction(tx interface{}) repository.TerraformIdpSubjectIdentifierRepository {
	// Implementation would depend on the transaction interface used
	// This is a placeholder for transaction support
	return r
}