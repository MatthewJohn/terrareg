package user_group

import (
	"context"
	"fmt"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth/repository"
	moduleRepo "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/repository"
	types "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared/types"
)

// DeleteUserGroupNamespacePermissionCommand handles deleting a namespace permission for a user group
// Matches Python: ApiTerraregAuthUserGroupNamespacePermissions._delete(user_group, namespace)
type DeleteUserGroupNamespacePermissionCommand struct {
	userGroupRepo repository.UserGroupRepository
	namespaceRepo moduleRepo.NamespaceRepository
}

// NewDeleteUserGroupNamespacePermissionCommand creates a new delete namespace permission command
func NewDeleteUserGroupNamespacePermissionCommand(
	userGroupRepo repository.UserGroupRepository,
	namespaceRepo moduleRepo.NamespaceRepository,
) *DeleteUserGroupNamespacePermissionCommand {
	return &DeleteUserGroupNamespacePermissionCommand{
		userGroupRepo: userGroupRepo,
		namespaceRepo: namespaceRepo,
	}
}

// Execute deletes a namespace permission for a user group
// Matches Python: UserGroupNamespacePermission.delete() called via ApiTerraregAuthUserGroupNamespacePermissions._delete()
// Returns nil on success, error on failure
func (c *DeleteUserGroupNamespacePermissionCommand) Execute(
	ctx context.Context,
	userGroupName string,
	namespaceName types.NamespaceName,
) error {
	// Get namespace by name
	// Python: namespace_obj = Namespace.get(name=namespace)
	namespace, err := c.namespaceRepo.FindByName(ctx, types.NamespaceName(namespaceName))
	if err != nil {
		return fmt.Errorf("failed to find namespace: %w", err)
	}
	if namespace == nil {
		return ErrNamespaceNotFound
	}

	// Get user group by name
	// Python: user_group_obj = UserGroup.get_by_group_name(user_group)
	userGroup, err := c.userGroupRepo.FindByName(ctx, userGroupName)
	if err != nil {
		return fmt.Errorf("failed to find user group: %w", err)
	}
	if userGroup == nil {
		return ErrUserGroupNotFound
	}

	// Check if permission exists
	// Python: user_group_namespace_permission = UserGroupNamespacePermission.get_permissions_by_user_group_and_namespace(...)
	permissions, err := c.userGroupRepo.GetNamespacePermissions(ctx, userGroup.ID)
	if err != nil {
		return fmt.Errorf("failed to check existing permissions: %w", err)
	}

	permissionExists := false
	for _, perm := range permissions {
		if perm.NamespaceID == namespace.ID() {
			permissionExists = true
			break
		}
	}

	if !permissionExists {
		return ErrPermissionNotFound
	}

	// Delete permission
	// Python: user_group_namespace_permission.delete()
	if err := c.userGroupRepo.RemoveNamespacePermission(ctx, userGroup.ID, namespace.ID()); err != nil {
		return fmt.Errorf("failed to delete namespace permission: %w", err)
	}

	return nil
}
