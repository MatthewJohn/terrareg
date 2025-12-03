package sqldb

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/identity/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/identity/repository"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared"
)

// IdentityRepositoryImpl implements the IdentityRepository interface using existing Terraform IDP models
type IdentityRepositoryImpl struct {
	db *gorm.DB
}

// NewIdentityRepositoryImpl creates a new GORM-based identity repository
func NewIdentityRepositoryImpl(db *gorm.DB) repository.IdentityRepository {
	return &IdentityRepositoryImpl{db: db}
}

// Save saves an identity using the existing Terraform IDP models
func (r *IdentityRepositoryImpl) Save(ctx context.Context, identity interface{}) error {
	terraformIdentity, ok := identity.(*model.TerraformIdentity)
	if !ok {
		return fmt.Errorf("expected *model.TerraformIdentity, got %T", identity)
	}

	// For OIDC identities, use the existing TerraformIDPSubjectIdentifierDB and TerraformIDPAccessTokenDB
	if terraformIdentity.IdentityType() == model.TerraformIdentityTypeOIDC {
		return r.saveOIDCIdentity(ctx, terraformIdentity)
	}

	// For static token identities, store in TerraformIDPAccessTokenDB with metadata
	return r.saveStaticTokenIdentity(ctx, terraformIdentity)
}

// FindByID finds an identity by ID
func (r *IdentityRepositoryImpl) FindByID(ctx context.Context, id string) (interface{}, error) {
	// First try to find in access tokens
	var accessTokenDB TerraformIDPAccessTokenDB
	if err := r.db.WithContext(ctx).Where("key = ?", id).First(&accessTokenDB).Error; err == nil {
		return r.mapAccessTokenToDomain(&accessTokenDB)
	}

	// Then try to find in subject identifiers
	var subjectIdentifierDB TerraformIDPSubjectIdentifierDB
	if err := r.db.WithContext(ctx).Where("key = ?", id).First(&subjectIdentifierDB).Error; err == nil {
		return r.mapSubjectIdentifierToDomain(&subjectIdentifierDB)
	}

	return nil, fmt.Errorf("identity not found: %w", repository.ErrNotFound)
}

// FindBySubject finds an identity by subject
func (r *IdentityRepositoryImpl) FindBySubject(ctx context.Context, subject string) (interface{}, error) {
	// Search in subject identifiers
	var subjectIdentifierDB TerraformIDPSubjectIdentifierDB
	if err := r.db.WithContext(ctx).Where("key = ?", subject).First(&subjectIdentifierDB).Error; err != nil {
		return nil, fmt.Errorf("identity not found: %w", repository.ErrNotFound)
	}

	return r.mapSubjectIdentifierToDomain(&subjectIdentifierDB)
}

// FindByAccessToken finds an identity by access token
func (r *IdentityRepositoryImpl) FindByAccessToken(ctx context.Context, accessToken string) (interface{}, error) {
	// Search in access tokens
	var accessTokenDB TerraformIDPAccessTokenDB
	if err := r.db.WithContext(ctx).Where("key = ?", accessToken).First(&accessTokenDB).Error; err != nil {
		return nil, fmt.Errorf("identity not found: %w", repository.ErrNotFound)
	}

	return r.mapAccessTokenToDomain(&accessTokenDB)
}

// Delete deletes an identity by ID
func (r *IdentityRepositoryImpl) Delete(ctx context.Context, id string) error {
	// Delete from both tables
	r.db.WithContext(ctx).Where("key = ?", id).Delete(&TerraformIDPAccessTokenDB{})
	return r.db.WithContext(ctx).Where("key = ?", id).Delete(&TerraformIDPSubjectIdentifierDB{}).Error
}

// Exists checks if an identity exists by ID
func (r *IdentityRepositoryImpl) Exists(ctx context.Context, id string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&TerraformIDPAccessTokenDB{}).Where("key = ?", id).Count(&count).Error
	if err == nil && count > 0 {
		return true, nil
	}

	err = r.db.WithContext(ctx).Model(&TerraformIDPSubjectIdentifierDB{}).Where("key = ?", id).Count(&count).Error
	return count > 0, err
}

// List returns a list of identities with pagination
func (r *IdentityRepositoryImpl) List(ctx context.Context, offset, limit int) ([]interface{}, error) {
	// Get from access tokens first
	var accessTokens []TerraformIDPAccessTokenDB
	if err := r.db.WithContext(ctx).Offset(offset).Limit(limit).Order("created_at DESC").Find(&accessTokens).Error; err != nil {
		return nil, fmt.Errorf("failed to list access tokens: %w", err)
	}

	result := make([]interface{}, 0, len(accessTokens))
	for _, token := range accessTokens {
		identity, err := r.mapAccessTokenToDomain(&token)
		if err != nil {
			continue // Skip invalid entries
		}
		result = append(result, identity)
	}

	return result, nil
}

// Count returns the total number of identities
func (r *IdentityRepositoryImpl) Count(ctx context.Context) (int64, error) {
	var accessTokenCount int64
	r.db.WithContext(ctx).Model(&TerraformIDPAccessTokenDB{}).Count(&accessTokenCount)

	var subjectIdentifierCount int64
	r.db.WithContext(ctx).Model(&TerraformIDPSubjectIdentifierDB{}).Count(&subjectIdentifierCount)

	return accessTokenCount + subjectIdentifierCount, nil
}

// FindByType finds identities by type with pagination
func (r *IdentityRepositoryImpl) FindByType(ctx context.Context, identityType string, offset, limit int) ([]interface{}, error) {
	// Since type is stored in metadata, we need to search differently
	var accessTokens []TerraformIDPAccessTokenDB
	if err := r.db.WithContext(ctx).Offset(offset).Limit(limit).Order("created_at DESC").Find(&accessTokens).Error; err != nil {
		return nil, fmt.Errorf("failed to find identities by type: %w", err)
	}

	result := make([]interface{}, 0)
	for _, token := range accessTokens {
		identity, err := r.mapAccessTokenToDomain(&token)
		if err != nil {
			continue
		}
		if identity.IdentityType().String() == identityType {
			result = append(result, identity)
		}
	}

	return result, nil
}

// CleanupExpired removes expired identities
func (r *IdentityRepositoryImpl) CleanupExpired(ctx context.Context) error {
	now := time.Now()

	// Clean up expired access tokens
	if err := r.db.WithContext(ctx).Where("expiry < ?", now).Delete(&TerraformIDPAccessTokenDB{}).Error; err != nil {
		return fmt.Errorf("failed to cleanup expired access tokens: %w", err)
	}

	// Clean up expired subject identifiers
	if err := r.db.WithContext(ctx).Where("expiry < ?", now).Delete(&TerraformIDPSubjectIdentifierDB{}).Error; err != nil {
		return fmt.Errorf("failed to cleanup expired subject identifiers: %w", err)
	}

	return nil
}

// saveOIDCIdentity saves an OIDC identity using the existing Terraform IDP tables
func (r *IdentityRepositoryImpl) saveOIDCIdentity(ctx context.Context, identity *model.TerraformIdentity) error {
	// Create subject identifier entry
	subjectData := map[string]interface{}{
		"subject":     identity.Subject(),
		"type":        identity.IdentityType().String(),
		"permissions": identity.Permissions(),
		"metadata":    identity.Metadata(),
	}

	subjectDataJSON, err := json.Marshal(subjectData)
	if err != nil {
		return fmt.Errorf("failed to marshal subject data: %w", err)
	}

	subjectIdentifierDB := TerraformIDPSubjectIdentifierDB{
		Key:    identity.Subject(),
		Data:   subjectDataJSON,
		Expiry: time.Now().Add(24 * time.Hour), // Default expiry for OIDC
	}

	if err := r.db.WithContext(ctx).Save(&subjectIdentifierDB).Error; err != nil {
		return fmt.Errorf("failed to save subject identifier: %w", err)
	}

	// Create access token entry if access token is present
	if identity.AccessToken() != "" {
		tokenData := map[string]interface{}{
			"subject":     identity.Subject(),
			"type":        identity.IdentityType().String(),
			"permissions": identity.Permissions(),
			"metadata":    identity.Metadata(),
		}

		tokenDataJSON, err := json.Marshal(tokenData)
		if err != nil {
			return fmt.Errorf("failed to marshal token data: %w", err)
		}

		accessTokenDB := TerraformIDPAccessTokenDB{
			Key:    identity.AccessToken(),
			Data:   tokenDataJSON,
			Expiry: identity.ExpiresAt() || time.Now().Add(time.Hour),
		}

		if err := r.db.WithContext(ctx).Save(&accessTokenDB).Error; err != nil {
			return fmt.Errorf("failed to save access token: %w", err)
		}
	}

	return nil
}

// saveStaticTokenIdentity saves a static token identity
func (r *IdentityRepositoryImpl) saveStaticTokenIdentity(ctx context.Context, identity *model.TerraformIdentity) error {
	tokenData := map[string]interface{}{
		"subject":     identity.Subject(),
		"type":        identity.IdentityType().String(),
		"permissions": identity.Permissions(),
		"metadata":    identity.Metadata(),
	}

	tokenDataJSON, err := json.Marshal(tokenData)
	if err != nil {
		return fmt.Errorf("failed to marshal token data: %w", err)
	}

	accessTokenDB := TerraformIDPAccessTokenDB{
		Key:    identity.AccessToken(),
		Data:   tokenDataJSON,
		Expiry: identity.ExpiresAt() || time.Now().Add(24 * time.Hour),
	}

	return r.db.WithContext(ctx).Save(&accessTokenDB).Error
}

// mapAccessTokenToDomain converts AccessTokenDB to domain model
func (r *IdentityRepositoryImpl) mapAccessTokenToDomain(tokenDB *TerraformIDPAccessTokenDB) (*model.TerraformIdentity, error) {
	var tokenData map[string]interface{}
	if err := json.Unmarshal(tokenDB.Data, &tokenData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal token data: %w", err)
	}

	subject, _ := tokenData["subject"].(string)
	identityTypeStr, _ := tokenData["type"].(string)
	permissionsInterface, _ := tokenData["permissions"].([]interface{})
	metadataInterface, _ := tokenData["metadata"].(map[string]interface{})

	identityType := model.TerraformIdentityTypeFromString(identityTypeStr)

	permissions := make([]string, len(permissionsInterface))
	for i, p := range permissionsInterface {
		if permStr, ok := p.(string); ok {
			permissions[i] = permStr
		}
	}

	metadata := make(map[string]string)
	for k, v := range metadataInterface {
		if vStr, ok := v.(string); ok {
			metadata[k] = vStr
		}
	}

	terraformIdentity := model.ReconstructTerraformIdentity(
		tokenDB.Key,
		subject,
		identityType,
		tokenDB.Key,
		&tokenDB.Expiry,
		permissions,
		metadata,
	)

	return terraformIdentity, nil
}

// mapSubjectIdentifierToDomain converts SubjectIdentifierDB to domain model
func (r *IdentityRepositoryImpl) mapSubjectIdentifierToDomain(subjectDB *TerraformIDPSubjectIdentifierDB) (*model.TerraformIdentity, error) {
	var subjectData map[string]interface{}
	if err := json.Unmarshal(subjectDB.Data, &subjectData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal subject data: %w", err)
	}

	subject, _ := subjectData["subject"].(string)
	identityTypeStr, _ := subjectData["type"].(string)
	permissionsInterface, _ := subjectData["permissions"].([]interface{})
	metadataInterface, _ := subjectData["metadata"].(map[string]interface{})

	identityType := model.TerraformIdentityTypeFromString(identityTypeStr)

	permissions := make([]string, len(permissionsInterface))
	for i, p := range permissionsInterface {
		if permStr, ok := p.(string); ok {
			permissions[i] = permStr
		}
	}

	metadata := make(map[string]string)
	for k, v := range metadataInterface {
		if vStr, ok := v.(string); ok {
			metadata[k] = vStr
		}
	}

	terraformIdentity := model.ReconstructTerraformIdentity(
		subjectDB.Key,
		subject,
		identityType,
		"", // No access token for subject identifiers
		&subjectDB.Expiry,
		permissions,
		metadata,
	)

	return terraformIdentity, nil
}