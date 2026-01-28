package user_group

import (
	"context"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth/repository"
	moduleRepo "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/repository"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared/types"
)

// ListUserGroupsQuery retrieves all user groups with their namespace permissions
// Matches Python: ApiTerraregAuthUserGroups._get()
type ListUserGroupsQuery struct {
	userGroupRepo repository.UserGroupRepository
	namespaceRepo moduleRepo.NamespaceRepository
}

// NewListUserGroupsQuery creates a new list user groups query
func NewListUserGroupsQuery(
	userGroupRepo repository.UserGroupRepository,
	namespaceRepo moduleRepo.NamespaceRepository,
) *ListUserGroupsQuery {
	return &ListUserGroupsQuery{
		userGroupRepo: userGroupRepo,
		namespaceRepo: namespaceRepo,
	}
}

// UserGroupResponse represents a user group in the API response
// Matches Python format: {name, site_admin, namespace_permissions}
type UserGroupResponse struct {
	Name                 string                        `json:"name"`
	SiteAdmin            bool                          `json:"site_admin"`
	NamespacePermissions []NamespacePermissionResponse `json:"namespace_permissions"`
}

// NamespacePermissionResponse represents a namespace permission in the API response
// Matches Python format: {namespace, permission_type}
type NamespacePermissionResponse struct {
	Namespace      types.NamespaceName `json:"namespace"`
	PermissionType string              `json:"permission_type"`
}

// Execute retrieves all user groups with their namespace permissions
// Matches Python: ApiTerraregAuthUserGroups._get()
// Returns list of {name, site_admin, namespace_permissions: [{namespace, permission_type}]}
func (q *ListUserGroupsQuery) Execute(ctx context.Context) ([]UserGroupResponse, error) {
	// Get all user groups (offset=0, limit=-1 for all)
	// Python calls: UserGroup.get_all_user_groups()
	userGroups, err := q.userGroupRepo.List(ctx, 0, -1)
	if err != nil {
		return nil, err
	}

	responses := make([]UserGroupResponse, len(userGroups))

	for i, ug := range userGroups {
		responses[i] = UserGroupResponse{
			Name:      ug.Name,
			SiteAdmin: ug.SiteAdmin,
		}

		// Get namespace permissions for this user group
		// Python calls: UserGroupNamespacePermission.get_permissions_by_user_group(user_group)
		permissions, err := q.userGroupRepo.GetNamespacePermissions(ctx, ug.ID)
		if err != nil {
			// If getting permissions fails, continue with empty list
			responses[i].NamespacePermissions = []NamespacePermissionResponse{}
			continue
		}

		// Convert permissions to response format
		namespacePerms := []NamespacePermissionResponse{} // Initialize as empty slice, not nil
		for _, perm := range permissions {
			// Get namespace name from namespace ID
			namespace, err := q.namespaceRepo.FindByID(ctx, perm.NamespaceID)
			if err != nil || namespace == nil {
				// Skip this permission if namespace not found
				continue
			}

			namespacePerms = append(namespacePerms, NamespacePermissionResponse{
				Namespace:      namespace.Name(),
				PermissionType: string(perm.PermissionType),
			})
		}
		responses[i].NamespacePermissions = namespacePerms
	}

	return responses, nil
}
