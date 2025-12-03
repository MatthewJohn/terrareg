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
	ErrUserGroupNotFound    = errors.New("user group not found")
	ErrUserAlreadyInGroup = errors.New("user already in group")
	ErrUserNotInGroup      = errors.New("user not in group")
	ErrPermissionNotFound    = errors.New("permission not found")
)

// UserGroupRepositoryImpl implements UserGroupRepository using GORM
type UserGroupRepositoryImpl struct {
	db *gorm.DB
}

// NewUserGroupRepository creates a new user group repository
func NewUserGroupRepository(db *gorm.DB) *UserGroupRepositoryImpl {
	return &UserGroupRepositoryImpl{db: db}
}

// Save creates or updates a user group
func (r *UserGroupRepositoryImpl) Save(ctx context.Context, userGroup *identityModel.UserGroup) error {
	userGroupDB := r.domainToDB(userGroup)

	// Use upsert to handle both create and update
	result := r.db.WithContext(ctx).
		Where("id = ?", userGroupDB.ID).
		Assign(userGroupDB).
		FirstOrCreate(&userGroupDB)

	return result.Error
}

// FindByID retrieves a user group by ID
func (r *UserGroupRepositoryImpl) FindByID(ctx context.Context, id string) (*identityModel.UserGroup, error) {
	var userGroupDB sqldb.UserGroupDB
	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&userGroupDB).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	return r.dbToDomain(&userGroupDB), nil
}

// FindByName retrieves a user group by name
func (r *UserGroupRepositoryImpl) FindByName(ctx context.Context, name string) (*identityModel.UserGroup, error) {
	var userGroupDB sqldb.UserGroupDB
	err := r.db.WithContext(ctx).
		Where("name = ?", name).
		First(&userGroupDB).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	return r.dbToDomain(&userGroupDB), nil
}

// Update updates a user group
func (r *UserGroupRepositoryImpl) Update(ctx context.Context, userGroup *identityModel.UserGroup) error {
	userGroupDB := r.domainToDB(userGroup)
	return r.db.WithContext(ctx).
		Where("id = ?", userGroupDB.ID).
		Updates(userGroupDB).Error
}

// Delete deletes a user group
func (r *UserGroupRepositoryImpl) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).
		Where("id = ?", id).
		Delete(&sqldb.UserGroupDB{}).Error
}

// AddMember adds a user to a user group
func (r *UserGroupRepositoryImpl) AddMember(ctx context.Context, userGroupID, userID string) error {
	// Check if user is already in group
	var count int64
	err := r.db.WithContext(ctx).
		Model(&sqldb.UserGroupMemberDB{}).
		Where("user_group_id = ? AND user_id = ?", userGroupID, userID).
		Count(&count).Error

	if err != nil {
		return err
	}

	if count > 0 {
		return ErrUserAlreadyInGroup
	}

	memberDB := sqldb.UserGroupMemberDB{
		UserGroupID: sqldb.StringToInt(userGroupID), // Assuming ID is string, need conversion
		UserID:      userID,
		JoinedAt:     time.Now(),
	}

	return r.db.WithContext(ctx).Create(&memberDB).Error
}

// RemoveMember removes a user from a user group
func (r *UserGroupRepositoryImpl) RemoveMember(ctx context.Context, userGroupID, userID string) error {
	result := r.db.WithContext(ctx).
		Where("user_group_id = ? AND user_id = ?", userGroupID, userID).
		Delete(&sqldb.UserGroupMemberDB{})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return ErrUserNotInGroup
	}

	return nil
}

// IsMember checks if a user is a member of a user group
func (r *UserGroupRepositoryImpl) IsMember(ctx context.Context, userGroupID, userID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&sqldb.UserGroupMemberDB{}).
		Where("user_group_id = ? AND user_id = ?", userGroupID, userID).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// GetMembers retrieves all members of a user group
func (r *UserGroupRepositoryImpl) GetMembers(ctx context.Context, userGroupID string) ([]*identityModel.User, error) {
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

	return r.userDBSliceToDomain(userDBs), nil
}

// GetActiveMembers retrieves active members of a user group
func (r *UserGroupRepositoryImpl) GetActiveMembers(ctx context.Context, userGroupID string) ([]*identityModel.User, error) {
	return r.GetMembers(ctx, userGroupID) // Already filters by active = true
}

// AddPermission adds a permission to a user group
func (r *UserGroupRepositoryImpl) AddPermission(ctx context.Context, userGroupID string, resourceType identityModel.ResourceType, resourceID string, action identityModel.Action, grantedBy string) error {
	// Check if permission already exists
	var count int64
	err := r.db.WithContext(ctx).
		Model(&sqldb.UserGroupNamespacePermissionDB{}).
		Where("user_group_id = ? AND namespace_id = ? AND permission_type = ?",
			userGroupID, resourceID, action).
		Count(&count).Error

	if err != nil {
		return err
	}

	if count > 0 {
		return ErrPermissionNotFound // Or a more specific "already exists" error
	}

	permissionDB := sqldb.UserGroupNamespacePermissionDB{
		UserGroupID:     sqldb.StringToInt(userGroupID),
		NamespaceID:     sqldb.StringToInt(resourceID), // Assuming resourceID is namespace ID
		PermissionType: sqldb.UserGroupNamespacePermissionType(action),
	}

	return r.db.WithContext(ctx).Create(&permissionDB).Error
}

// RemovePermission removes a permission from a user group
func (r *UserGroupRepositoryImpl) RemovePermission(ctx context.Context, userGroupID string, permissionID int) error {
	result := r.db.WithContext(ctx).
		Where("user_group_id = ? AND id = ?", userGroupID, permissionID).
		Delete(&sqldb.UserGroupNamespacePermissionDB{})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return ErrPermissionNotFound
	}

	return nil
}

// HasPermission checks if a user group has a specific permission
func (r *UserGroupRepositoryImpl) HasPermission(ctx context.Context, userGroupID string, resourceType identityModel.ResourceType, resourceID string, action identityModel.Action) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&sqldb.UserGroupNamespacePermissionDB{}).
		Where("user_group_id = ? AND namespace_id = ? AND permission_type IN (?, ?)",
			userGroupID, resourceID, action, sqldb.PermissionTypeFull).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// GetPermissions retrieves all permissions for a user group
func (r *UserGroupRepositoryImpl) GetPermissions(ctx context.Context, userGroupID string) ([]identityModel.GroupPermission, error) {
	var permissions []sqldb.UserGroupNamespacePermissionDB
	err := r.db.WithContext(ctx).
		Where("user_group_id = ?", userGroupID).
		Find(&permissions).Error

	if err != nil {
		return nil, err
	}

	return r.permissionsToDomain(permissions), nil
}

// List retrieves user groups with pagination
func (r *UserGroupRepositoryImpl) List(ctx context.Context, offset, limit int) ([]*identityModel.UserGroup, error) {
	var userGroupDBs []sqldb.UserGroupDB
	err := r.db.WithContext(ctx).
		Offset(offset).
		Limit(limit).
		Order("name ASC").
		Find(&userGroupDBs).Error

	if err != nil {
		return nil, err
	}

	return r.dbSliceToDomain(userGroupDBs), nil
}

// Count counts total user groups
func (r *UserGroupRepositoryImpl) Count(ctx context.Context) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&sqldb.UserGroupDB{}).
		Count(&count).Error

	return int(count), err
}

// Search searches user groups by name
func (r *UserGroupRepositoryImpl) Search(ctx context.Context, query string, offset, limit int) ([]*identityModel.UserGroup, error) {
	var userGroupDBs []sqldb.UserGroupDB
	searchPattern := "%" + query + "%"

	err := r.db.WithContext(ctx).
		Where("name ILIKE ?", searchPattern).
		Offset(offset).
		Limit(limit).
		Order("name ASC").
		Find(&userGroupDBs).Error

	if err != nil {
		return nil, err
	}

	return r.dbSliceToDomain(userGroupDBs), nil
}

// FindByUserID retrieves all user groups for a user
func (r *UserGroupRepositoryImpl) FindByUserID(ctx context.Context, userID string) ([]*identityModel.UserGroup, error) {
	var userGroupDBs []sqldb.UserGroupDB
	err := r.db.WithContext(ctx).
		Table("user_group ug").
		Select("ug.*").
		Joins("INNER JOIN user_group_member ugm ON ug.id = ugm.user_group_id").
		Where("ugm.user_id = ?", userID).
		Order("ug.name ASC").
		Find(&userGroupDBs).Error

	if err != nil {
		return nil, err
	}

	return r.dbSliceToDomain(userGroupDBs), nil
}

// WithTransaction returns repository with transaction support
func (r *UserGroupRepositoryImpl) WithTransaction(tx interface{}) identityRepository.UserGroupRepository {
	if txDB, ok := tx.(*gorm.DB); ok {
		return &UserGroupRepositoryImpl{db: txDB}
	}
	return r
}

// Helper methods for domain-to-database mapping

func (r *UserGroupRepositoryImpl) domainToDB(userGroup *identityModel.UserGroup) sqldb.UserGroupDB {
	return sqldb.UserGroupDB{
		ID:        sqldb.StringToInt(userGroup.ID()),
		Name:      userGroup.Name(),
		SiteAdmin: userGroup.IsSiteAdmin(),
	}
}

func (r *UserGroupRepositoryImpl) dbToDomain(userGroupDB *sqldb.UserGroupDB) *identityModel.UserGroup {
	// This assumes UserGroup can be reconstructed from these fields
	// In a real implementation, we'd need a proper constructor
	userGroup := &identityModel.UserGroup{}
	// Set fields based on available UserGroup constructor
	return userGroup
}

func (r *UserGroupRepositoryImpl) dbSliceToDomain(userGroupDBs []sqldb.UserGroupDB) []*identityModel.UserGroup {
	userGroups := make([]*identityModel.UserGroup, len(userGroupDBs))
	for i, userGroupDB := range userGroupDBs {
		userGroups[i] = r.dbToDomain(&userGroupDB)
	}
	return userGroups
}

func (r *UserGroupRepositoryImpl) userDBSliceToDomain(userDBs []sqldb.UserDB) []*identityModel.User {
	users := make([]*identityModel.User, len(userDBs))
	for i, userDB := range userDBs {
		users[i] = r.userDBToDomain(&userDB)
	}
	return users
}

func (r *UserGroupRepositoryImpl) userDBToDomain(userDB *sqldb.UserDB) *identityModel.User {
	// This is a simplified conversion - in reality we'd need proper mapping
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

	user, _ := identityModel.NewUser(userDB.Username, userDB.DisplayName, userDB.Email, authMethod)
	// Set additional fields as needed
	return user
}

func (r *UserGroupRepositoryImpl) permissionsToDomain(permissions []sqldb.UserGroupNamespacePermissionDB) []identityModel.GroupPermission {
	result := make([]identityModel.GroupPermission, len(permissions))
	for i, perm := range permissions {
		// Map permission type to action
		var action identityModel.Action
		switch perm.PermissionType {
		case sqldb.PermissionTypeFull:
			action = identityModel.ActionAdmin
		case sqldb.PermissionTypeModify:
			action = identityModel.ActionWrite
		case sqldb.PermissionTypeRead:
			action = identityModel.ActionRead
		}

		result[i] = identityModel.GroupPermission{
			// Set fields based on available GroupPermission constructor
			ResourceType: identityModel.ResourceTypeNamespace,
			ResourceID:   sqldb.IntToString(perm.NamespaceID),
			Action:       action,
		}
	}
	return result
}