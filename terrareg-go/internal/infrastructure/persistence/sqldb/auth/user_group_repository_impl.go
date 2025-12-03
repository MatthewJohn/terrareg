package auth

import (
	"context"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth/repository"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
	"gorm.io/gorm"
)

// UserGroupRepositoryImpl implements the UserGroupRepository interface using SQL database
type UserGroupRepositoryImpl struct {
	db *gorm.DB
}

// NewUserGroupRepository creates a new UserGroupRepository implementation
func NewUserGroupRepository(db *gorm.DB) repository.UserGroupRepository {
	return &UserGroupRepositoryImpl{
		db: db,
	}
}

// Save saves a user group (creates or updates)
func (r *UserGroupRepositoryImpl) Save(ctx context.Context, userGroup *auth.UserGroup) error {
	if userGroup.ID == 0 {
		// Create new user group
		dbUserGroup := &sqldb.UserGroupDB{
			Name:      userGroup.Name,
			SiteAdmin: userGroup.SiteAdmin,
		}

		result := r.db.WithContext(ctx).Create(dbUserGroup)
		if result.Error != nil {
			return result.Error
		}

		userGroup.ID = dbUserGroup.ID
		return nil
	} else {
		// Update existing user group
		return r.Update(ctx, userGroup)
	}
}

// Create creates a new user group (alias for Save when ID is 0)
func (r *UserGroupRepositoryImpl) Create(ctx context.Context, userGroup *auth.UserGroup) error {
	return r.Save(ctx, userGroup)
}

// FindByID finds a user group by ID
func (r *UserGroupRepositoryImpl) FindByID(ctx context.Context, id int) (*auth.UserGroup, error) {
	var dbUserGroup sqldb.UserGroupDB
	err := r.db.WithContext(ctx).First(&dbUserGroup, id).Error
	if err != nil {
		return nil, err
	}

	return &auth.UserGroup{
		ID:        dbUserGroup.ID,
		Name:      dbUserGroup.Name,
		SiteAdmin: dbUserGroup.SiteAdmin,
	}, nil
}

// FindByName finds a user group by name
func (r *UserGroupRepositoryImpl) FindByName(ctx context.Context, name string) (*auth.UserGroup, error) {
	var dbUserGroup sqldb.UserGroupDB
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&dbUserGroup).Error
	if err != nil {
		return nil, err
	}

	return &auth.UserGroup{
		ID:        dbUserGroup.ID,
		Name:      dbUserGroup.Name,
		SiteAdmin: dbUserGroup.SiteAdmin,
	}, nil
}

// Update updates a user group
func (r *UserGroupRepositoryImpl) Update(ctx context.Context, userGroup *auth.UserGroup) error {
	dbUserGroup := &sqldb.UserGroupDB{
		ID:        userGroup.ID,
		Name:      userGroup.Name,
		SiteAdmin: userGroup.SiteAdmin,
	}

	return r.db.WithContext(ctx).Save(dbUserGroup).Error
}

// Delete deletes a user group by ID
func (r *UserGroupRepositoryImpl) Delete(ctx context.Context, id int) error {
	return r.db.WithContext(ctx).Delete(&sqldb.UserGroupDB{}, id).Error
}

// List returns user groups with offset and limit
func (r *UserGroupRepositoryImpl) List(ctx context.Context, offset, limit int) ([]*auth.UserGroup, error) {
	var dbUserGroups []sqldb.UserGroupDB
	err := r.db.WithContext(ctx).Offset(offset).Limit(limit).Find(&dbUserGroups).Error
	if err != nil {
		return nil, err
	}

	userGroups := make([]*auth.UserGroup, len(dbUserGroups))
	for i, dbUserGroup := range dbUserGroups {
		userGroups[i] = &auth.UserGroup{
			ID:        dbUserGroup.ID,
			Name:      dbUserGroup.Name,
			SiteAdmin: dbUserGroup.SiteAdmin,
		}
	}

	return userGroups, nil
}

// FindSiteAdminGroups returns all site admin user groups
func (r *UserGroupRepositoryImpl) FindSiteAdminGroups(ctx context.Context) ([]*auth.UserGroup, error) {
	var dbUserGroups []sqldb.UserGroupDB
	err := r.db.WithContext(ctx).Where("site_admin = ?", true).Find(&dbUserGroups).Error
	if err != nil {
		return nil, err
	}

	userGroups := make([]*auth.UserGroup, len(dbUserGroups))
	for i, dbUserGroup := range dbUserGroups {
		userGroups[i] = &auth.UserGroup{
			ID:        dbUserGroup.ID,
			Name:      dbUserGroup.Name,
			SiteAdmin: dbUserGroup.SiteAdmin,
		}
	}

	return userGroups, nil
}

// GetNamespacePermissions returns all namespace permissions for a user group
func (r *UserGroupRepositoryImpl) GetNamespacePermissions(ctx context.Context, userGroupID int) ([]auth.NamespacePermission, error) {
	var dbPermissions []sqldb.UserGroupNamespacePermissionDB
	err := r.db.WithContext(ctx).Where("user_group_id = ?", userGroupID).Find(&dbPermissions).Error
	if err != nil {
		return nil, err
	}

	permissions := make([]auth.NamespacePermission, len(dbPermissions))
	for i, dbPerm := range dbPermissions {
		permissions[i] = auth.NamespacePermission{
			UserGroupID:    dbPerm.UserGroupID,
			NamespaceID:    dbPerm.NamespaceID,
			PermissionType: auth.PermissionType(dbPerm.PermissionType),
		}
	}

	return permissions, nil
}

// GetGroupsForUser returns groups for a user (placeholder implementation)
func (r *UserGroupRepositoryImpl) GetGroupsForUser(ctx context.Context, userID string) ([]*auth.UserGroup, error) {
	// This is a placeholder implementation
	// In a real system, you would have a user_groups table mapping users to groups
	// For now, return empty list
	return []*auth.UserGroup{}, nil
}

// AddNamespacePermission adds a namespace permission for a user group (implements interface)
func (r *UserGroupRepositoryImpl) AddNamespacePermission(ctx context.Context, userGroupID, namespaceID int, permissionType auth.PermissionType) error {
	dbPermission := &sqldb.UserGroupNamespacePermissionDB{
		UserGroupID:    userGroupID,
		NamespaceID:    namespaceID,
		PermissionType: sqldb.UserGroupNamespacePermissionType(permissionType),
	}

	return r.db.WithContext(ctx).Create(dbPermission).Error
}

// CreateNamespacePermission creates a new namespace permission for a user group (alias for AddNamespacePermission)
func (r *UserGroupRepositoryImpl) CreateNamespacePermission(ctx context.Context, userGroupID, namespaceID int, permissionType auth.PermissionType) error {
	return r.AddNamespacePermission(ctx, userGroupID, namespaceID, permissionType)
}

// RemoveNamespacePermission removes a namespace permission for a user group (implements interface)
func (r *UserGroupRepositoryImpl) RemoveNamespacePermission(ctx context.Context, userGroupID, namespaceID int) error {
	return r.db.WithContext(ctx).
		Where("user_group_id = ? AND namespace_id = ?", userGroupID, namespaceID).
		Delete(&sqldb.UserGroupNamespacePermissionDB{}).Error
}

// DeleteNamespacePermission deletes a namespace permission for a user group (alias for RemoveNamespacePermission)
func (r *UserGroupRepositoryImpl) DeleteNamespacePermission(ctx context.Context, userGroupID, namespaceID int) error {
	return r.RemoveNamespacePermission(ctx, userGroupID, namespaceID)
}

// Count returns the total number of user groups
func (r *UserGroupRepositoryImpl) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&sqldb.UserGroupDB{}).Count(&count).Error
	return count, err
}

// HasNamespacePermission checks if a user group has a specific namespace permission
func (r *UserGroupRepositoryImpl) HasNamespacePermission(ctx context.Context, userGroupID, namespaceID int, permissionType auth.PermissionType) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&sqldb.UserGroupNamespacePermissionDB{}).
		Where("user_group_id = ? AND namespace_id = ? AND permission_type = ?", userGroupID, namespaceID, string(permissionType)).
		Count(&count).Error
	return count > 0, err
}

// GetHighestNamespacePermission gets the highest permission level for a user group in a namespace
func (r *UserGroupRepositoryImpl) GetHighestNamespacePermission(ctx context.Context, userGroupID, namespaceID int) (auth.PermissionType, error) {
	var dbPerm sqldb.UserGroupNamespacePermissionDB
	err := r.db.WithContext(ctx).
		Where("user_group_id = ? AND namespace_id = ?", userGroupID, namespaceID).
		Order("CASE permission_type WHEN 'FULL' THEN 1 WHEN 'MODIFY' THEN 2 WHEN 'READ' THEN 3 END").
		First(&dbPerm).Error

	if err != nil {
		return "", err
	}

	return auth.PermissionType(dbPerm.PermissionType), nil
}

// FindGroupsByNamespace finds all groups that have permissions for a specific namespace
func (r *UserGroupRepositoryImpl) FindGroupsByNamespace(ctx context.Context, namespaceID int) ([]*auth.UserGroup, error) {
	var dbUserGroups []sqldb.UserGroupDB
	err := r.db.WithContext(ctx).
		Joins("INNER JOIN user_group_namespace_permissions ON user_groups.id = user_group_namespace_permissions.user_group_id").
		Where("user_group_namespace_permissions.namespace_id = ?", namespaceID).
		Distinct().
		Find(&dbUserGroups).Error

	if err != nil {
		return nil, err
	}

	userGroups := make([]*auth.UserGroup, len(dbUserGroups))
	for i, dbUserGroup := range dbUserGroups {
		userGroups[i] = &auth.UserGroup{
			ID:        dbUserGroup.ID,
			Name:      dbUserGroup.Name,
			SiteAdmin: dbUserGroup.SiteAdmin,
		}
	}

	return userGroups, nil
}

// SearchByName searches user groups by name pattern
func (r *UserGroupRepositoryImpl) SearchByName(ctx context.Context, query string, offset, limit int) ([]*auth.UserGroup, error) {
	var dbUserGroups []sqldb.UserGroupDB
	err := r.db.WithContext(ctx).
		Where("name ILIKE ?", "%"+query+"%").
		Offset(offset).
		Limit(limit).
		Find(&dbUserGroups).Error

	if err != nil {
		return nil, err
	}

	userGroups := make([]*auth.UserGroup, len(dbUserGroups))
	for i, dbUserGroup := range dbUserGroups {
		userGroups[i] = &auth.UserGroup{
			ID:        dbUserGroup.ID,
			Name:      dbUserGroup.Name,
			SiteAdmin: dbUserGroup.SiteAdmin,
		}
	}

	return userGroups, nil
}

// WithTransaction returns a new repository instance that works within the given transaction
func (r *UserGroupRepositoryImpl) WithTransaction(tx interface{}) repository.UserGroupRepository {
	// Implementation would depend on the transaction interface used
	// This is a placeholder for transaction support
	return r
}

// UserGroupNamespacePermission implements the NamespacePermission interface
type UserGroupNamespacePermission struct {
	UserGroupID    int
	NamespaceID    int
	PermissionType auth.PermissionType
}

// GetUserGroupID returns the user group ID
func (p *UserGroupNamespacePermission) GetUserGroupID() int {
	return p.UserGroupID
}

// GetNamespaceID returns the namespace ID
func (p *UserGroupNamespacePermission) GetNamespaceID() int {
	return p.NamespaceID
}

// GetPermissionType returns the permission type
func (p *UserGroupNamespacePermission) GetPermissionType() auth.PermissionType {
	return p.PermissionType
}
