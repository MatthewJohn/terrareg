package user_group

import (
	"context"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/identity/repository"
)

// ListUserGroupsQuery retrieves all user groups
// Matches Python terrareg.models.UserGroup.get_all_user_groups()
type ListUserGroupsQuery struct {
	groupRepo      repository.UserGroupRepository
	permissionRepo repository.UserGroupNamespacePermissionRepository
}

// NewListUserGroupsQuery creates a new ListUserGroupsQuery
func NewListUserGroupsQuery(
	groupRepo repository.UserGroupRepository,
	permissionRepo repository.UserGroupNamespacePermissionRepository,
) *ListUserGroupsQuery {
	return &ListUserGroupsQuery{
		groupRepo:      groupRepo,
		permissionRepo: permissionRepo,
	}
}

// Execute retrieves all user groups
// Returns a slice matching Python API: [{name, site_admin, namespace_permissions}]
func (q *ListUserGroupsQuery) Execute(ctx context.Context) ([]map[string]interface{}, error) {
	groups, err := q.groupRepo.List(ctx)
	if err != nil {
		return nil, err
	}

	// Convert to response format matching Python exactly
	result := make([]map[string]interface{}, len(groups))
	for i, group := range groups {
		// Get namespace permissions for this group
		permissions, err := q.permissionRepo.GetByUserGroup(ctx, group.Name())
		if err != nil {
			return nil, err
		}

		// Convert permissions to format matching Python
		namespacePermissions := make([]map[string]string, len(permissions))
		for j, perm := range permissions {
			namespacePermissions[j] = map[string]string{
				"namespace":       perm.NamespaceName,
				"permission_type": string(perm.PermissionType),
			}
		}

		// Create group entry matching Python API
		result[i] = map[string]interface{}{
			"name":                  group.Name(),
			"site_admin":            group.SiteAdmin(),
			"namespace_permissions": namespacePermissions,
		}
	}

	return result, nil
}
