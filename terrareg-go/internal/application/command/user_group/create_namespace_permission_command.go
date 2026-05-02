package user_group

import (
	"context"
	"fmt"

	auditservice "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/audit/service"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth/repository"
	moduleRepo "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/repository"
	types "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared/types"
)

// CreateUserGroupNamespacePermissionCommand handles creating a namespace permission for a user group
// Matches Python: ApiTerraregAuthUserGroupNamespacePermissions._post(user_group, namespace)
type CreateUserGroupNamespacePermissionCommand struct {
	userGroupRepo         repository.UserGroupRepository
	namespaceRepo         moduleRepo.NamespaceRepository
	userGroupAuditService auditservice.UserGroupAuditServiceInterface
}

// NewCreateUserGroupNamespacePermissionCommand creates a new create namespace permission command
func NewCreateUserGroupNamespacePermissionCommand(
	userGroupRepo repository.UserGroupRepository,
	namespaceRepo moduleRepo.NamespaceRepository,
	userGroupAuditService auditservice.UserGroupAuditServiceInterface,
) *CreateUserGroupNamespacePermissionCommand {
	return &CreateUserGroupNamespacePermissionCommand{
		userGroupRepo:         userGroupRepo,
		namespaceRepo:         namespaceRepo,
		userGroupAuditService: userGroupAuditService,
	}
}

// CreateNamespacePermissionRequest represents the request to create a namespace permission
// Matches Python JSON input: {permission_type}
type CreateNamespacePermissionRequest struct {
	PermissionType string `json:"permission_type"`
}

// CreateNamespacePermissionResponse represents the response after creating a namespace permission
// Matches Python JSON response: {user_group, namespace, permission_type}
type CreateNamespacePermissionResponse struct {
	UserGroup      string              `json:"user_group"`
	Namespace      types.NamespaceName `json:"namespace"`
	PermissionType string              `json:"permission_type"`
}

// ValidPermissionTypes matches Python UserGroupNamespacePermissionType enum
// Python: FULL, MODIFY (note: READ is not in the Python enum)
var validPermissionTypes = map[string]bool{
	"FULL":   true,
	"MODIFY": true,
}

// Execute creates a namespace permission for a user group
// Matches Python: UserGroupNamespacePermission.create(user_group, namespace, permission_type)
// Returns CreateNamespacePermissionResponse on success, error on failure
func (c *CreateUserGroupNamespacePermissionCommand) Execute(
	ctx context.Context,
	userGroupName string,
	namespaceName types.NamespaceName,
	req CreateNamespacePermissionRequest,
) (*CreateNamespacePermissionResponse, error) {
	// Validate permission_type is valid enum value
	// Python: try: permission_type_enum = UserGroupNamespacePermissionType(permission_type)
	if !validPermissionTypes[req.PermissionType] {
		return nil, ErrInvalidPermissionType
	}

	// Get namespace by name
	// Python: namespace_obj = Namespace.get(name=namespace)
	namespace, err := c.namespaceRepo.FindByName(ctx, namespaceName)
	if err != nil {
		return nil, fmt.Errorf("failed to find namespace: %w", err)
	}
	if namespace == nil {
		return nil, ErrNamespaceNotFound
	}

	// Get user group by name
	// Python: user_group_obj = UserGroup.get_by_group_name(user_group)
	userGroup, err := c.userGroupRepo.FindByName(ctx, userGroupName)
	if err != nil {
		return nil, fmt.Errorf("failed to find user group: %w", err)
	}
	if userGroup == nil {
		return nil, ErrUserGroupNotFound
	}

	// Check if permission already exists
	// Python: if cls.get_permissions_by_user_group_and_namespace(user_group=user_group, namespace=namespace):
	permissions, err := c.userGroupRepo.GetNamespacePermissions(ctx, userGroup.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing permissions: %w", err)
	}

	for _, perm := range permissions {
		if perm.NamespaceID == namespace.ID() {
			return nil, ErrPermissionAlreadyExists
		}
	}

	// Create permission
	// Python: UserGroupNamespacePermission.create(user_group=user_group_obj, namespace=namespace_obj, permission_type=permission_type_enum)
	permissionType := auth.PermissionType(req.PermissionType)
	if err := c.userGroupRepo.AddNamespacePermission(ctx, userGroup.ID, namespace.ID(), permissionType); err != nil {
		return nil, fmt.Errorf("failed to create namespace permission: %w", err)
	}

	// Log audit event (synchronous)
	// Python reference: /app/terrareg/models.py:385 - AuditAction.USER_GROUP_NAMESPACE_PERMISSION_ADD
	if c.userGroupAuditService != nil {
		c.userGroupAuditService.LogUserGroupNamespacePermissionAdd(ctx, userGroupName, namespaceName, req.PermissionType)
	}

	// Return response matching Python format
	return &CreateNamespacePermissionResponse{
		UserGroup:      userGroup.Name,
		Namespace:      namespace.Name(),
		PermissionType: req.PermissionType,
	}, nil
}
