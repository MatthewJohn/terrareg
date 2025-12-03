package identity

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	identityModel "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/identity/model"
	identityRepository "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/identity/repository"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
)

// UserRepositoryImpl implements UserRepository using GORM
type UserRepositoryImpl struct {
	db *gorm.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *gorm.DB) *UserRepositoryImpl {
	return &UserRepositoryImpl{db: db}
}

// Save creates or updates a user
func (r *UserRepositoryImpl) Save(ctx context.Context, user *identityModel.User) error {
	userDB := r.domainToDB(user)

	// Use upsert to handle both create and update
	result := r.db.WithContext(ctx).
		Where("id = ?", userDB.ID).
		Assign(userDB).
		FirstOrCreate(&userDB)

	return result.Error
}

// FindByID retrieves a user by ID
func (r *UserRepositoryImpl) FindByID(ctx context.Context, id string) (*identityModel.User, error) {
	var userDB sqldb.UserDB
	err := r.db.WithContext(ctx).
		Where("id = ? AND active = ?", id, true).
		First(&userDB).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	return r.dbToDomain(&userDB), nil
}

// FindByUsername retrieves a user by username
func (r *UserRepositoryImpl) FindByUsername(ctx context.Context, username string) (*identityModel.User, error) {
	var userDB sqldb.UserDB
	err := r.db.WithContext(ctx).
		Where("username = ? AND active = ?", username, true).
		First(&userDB).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	return r.dbToDomain(&userDB), nil
}

// FindByEmail retrieves a user by email
func (r *UserRepositoryImpl) FindByEmail(ctx context.Context, email string) (*identityModel.User, error) {
	var userDB sqldb.UserDB
	err := r.db.WithContext(ctx).
		Where("email = ? AND active = ?", email, true).
		First(&userDB).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	return r.dbToDomain(&userDB), nil
}

// Update updates a user
func (r *UserRepositoryImpl) Update(ctx context.Context, user *identityModel.User) error {
	userDB := r.domainToDB(user)
	return r.db.WithContext(ctx).
		Where("id = ?", userDB.ID).
		Updates(userDB).Error
}

// Delete soft-deletes a user (sets active = false)
func (r *UserRepositoryImpl) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).
		Model(&sqldb.UserDB{}).
		Where("id = ?", id).
		Update("active", false).Error
}

// FindByAuthProviderID finds user by auth method and provider ID
func (r *UserRepositoryImpl) FindByAuthProviderID(ctx context.Context, authMethod identityModel.AuthMethod, providerID string) (*identityModel.User, error) {
	var userDB sqldb.UserDB
	err := r.db.WithContext(ctx).
		Where("auth_method = ? AND auth_provider_id = ? AND active = ?",
			authMethod.String(), providerID, true).
		First(&userDB).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	return r.dbToDomain(&userDB), nil
}

// FindByExternalID finds user by auth method and external ID
func (r *UserRepositoryImpl) FindByExternalID(ctx context.Context, authMethod identityModel.AuthMethod, externalID string) (*identityModel.User, error) {
	var userDB sqldb.UserDB
	err := r.db.WithContext(ctx).
		Where("auth_method = ? AND external_id = ? AND active = ?",
			authMethod.String(), externalID, true).
		First(&userDB).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	return r.dbToDomain(&userDB), nil
}

// FindByAccessToken finds user by access token
func (r *UserRepositoryImpl) FindByAccessToken(ctx context.Context, accessToken string) (*identityModel.User, error) {
	var userDB sqldb.UserDB
	err := r.db.WithContext(ctx).
		Where("access_token = ? AND active = ? AND (token_expiry IS NULL OR token_expiry > ?)",
			accessToken, true, time.Now()).
		First(&userDB).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	return r.dbToDomain(&userDB), nil
}

// FindActive retrieves all active users
func (r *UserRepositoryImpl) FindActive(ctx context.Context) ([]*identityModel.User, error) {
	var userDBs []sqldb.UserDB
	err := r.db.WithContext(ctx).
		Where("active = ?", true).
		Find(&userDBs).Error

	if err != nil {
		return nil, err
	}

	return r.dbSliceToDomain(userDBs), nil
}

// FindInactive retrieves all inactive users
func (r *UserRepositoryImpl) FindInactive(ctx context.Context) ([]*identityModel.User, error) {
	var userDBs []sqldb.UserDB
	err := r.db.WithContext(ctx).
		Where("active = ?", false).
		Find(&userDBs).Error

	if err != nil {
		return nil, err
	}

	return r.dbSliceToDomain(userDBs), nil
}

// FindByUserGroupID finds users by user group membership
func (r *UserRepositoryImpl) FindByUserGroupID(ctx context.Context, userGroupID string) ([]*identityModel.User, error) {
	var userDBs []sqldb.UserDB
	err := r.db.WithContext(ctx).
		Table("user u").
		Select("u.*").
		Joins("INNER JOIN user_group_member ugm ON u.id = ugm.user_id").
		Where("ugm.user_group_id = ? AND u.active = ?", userGroupID, true).
		Find(&userDBs).Error

	if err != nil {
		return nil, err
	}

	return r.dbSliceToDomain(userDBs), nil
}

// List retrieves users with pagination
func (r *UserRepositoryImpl) List(ctx context.Context, offset, limit int) ([]*identityModel.User, error) {
	var userDBs []sqldb.UserDB
	err := r.db.WithContext(ctx).
		Where("active = ?", true).
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&userDBs).Error

	if err != nil {
		return nil, err
	}

	return r.dbSliceToDomain(userDBs), nil
}

// Count counts total active users
func (r *UserRepositoryImpl) Count(ctx context.Context) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&sqldb.UserDB{}).
		Where("active = ?", true).
		Count(&count).Error

	return int(count), err
}

// Search searches users by username or display name
func (r *UserRepositoryImpl) Search(ctx context.Context, query string, offset, limit int) ([]*identityModel.User, error) {
	var userDBs []sqldb.UserDB
	searchPattern := "%" + query + "%"

	err := r.db.WithContext(ctx).
		Where("active = ? AND (username ILIKE ? OR display_name ILIKE ?)",
			true, searchPattern, searchPattern).
		Offset(offset).
		Limit(limit).
		Order("username ASC").
		Find(&userDBs).Error

	if err != nil {
		return nil, err
	}

	return r.dbSliceToDomain(userDBs), nil
}

// WithTransaction returns repository with transaction support
func (r *UserRepositoryImpl) WithTransaction(tx interface{}) identityRepository.UserRepository {
	if txDB, ok := tx.(*gorm.DB); ok {
		return &UserRepositoryImpl{db: txDB}
	}
	return r
}

// Helper methods for domain-to-database mapping

func (r *UserRepositoryImpl) domainToDB(user *identityModel.User) sqldb.UserDB {
	return sqldb.UserDB{
		ID:            user.ID(),
		Username:      user.Username(),
		DisplayName:   user.DisplayName(),
		Email:         user.Email(),
		AuthMethod:     user.AuthMethod().String(),
		AuthProviderID: user.AuthProviderID(),
		ExternalID:     user.ExternalID(),
		AccessToken:    user.AccessToken(),
		RefreshToken:   user.RefreshToken(),
		TokenExpiry:    user.AccessTokenExpiry(),
		Active:         user.Active(),
		CreatedAt:      user.CreatedAt(),
		LastLoginAt:    user.LastLoginAt(),
	}
}

func (r *UserRepositoryImpl) dbToDomain(userDB *sqldb.UserDB) *identityModel.User {
	authMethod := identityModel.AuthMethodNone
	switch userDB.AuthMethod {
	case "SAML":
		authMethod = identityModel.AuthMethodSAML
	case "OIDC":
		authMethod = identityModel.AuthMethodOIDC
	case "GITHUB":
		authMethod = identityModel.AuthMethodGitHub
	case "API_KEY":
		authMethod = identityModel.AuthMethodAPIKey
	case "TERRAFORM":
		authMethod = identityModel.AuthMethodTerraform
	}

	user, err := identityModel.ReconstructUser(
		userDB.ID,
		userDB.Username,
		userDB.DisplayName,
		userDB.Email,
		authMethod,
		userDB.AuthProviderID,
		userDB.ExternalID,
		userDB.AccessToken,
		userDB.RefreshToken,
		userDB.TokenExpiry,
		userDB.LastLoginAt,
		userDB.CreatedAt,
		userDB.Active,
	)
	if err != nil {
		// In case of error, return nil to indicate reconstruction failed
		return nil
	}

	return user
}

func (r *UserRepositoryImpl) dbSliceToDomain(userDBs []sqldb.UserDB) []*identityModel.User {
	users := make([]*identityModel.User, len(userDBs))
	for i, userDB := range userDBs {
		users[i] = r.dbToDomain(&userDB)
	}
	return users
}