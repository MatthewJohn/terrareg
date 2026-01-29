package user_group

import (
	"context"
	"fmt"

	auditservice "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/audit/service"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth/repository"
)

// DeleteUserGroupCommand handles deleting a user group
// Matches Python: ApiTerraregAuthUserGroup._delete(user_group)
type DeleteUserGroupCommand struct {
	userGroupRepo         repository.UserGroupRepository
	userGroupAuditService *auditservice.UserGroupAuditService
}

// NewDeleteUserGroupCommand creates a new delete user group command
func NewDeleteUserGroupCommand(userGroupRepo repository.UserGroupRepository, userGroupAuditService *auditservice.UserGroupAuditService) *DeleteUserGroupCommand {
	return &DeleteUserGroupCommand{
		userGroupRepo:         userGroupRepo,
		userGroupAuditService: userGroupAuditService,
	}
}

// Execute deletes a user group by name
// Matches Python: UserGroup.delete() called via ApiTerraregAuthUserGroup._delete()
// Returns nil on success, error on failure
func (c *DeleteUserGroupCommand) Execute(ctx context.Context, userGroupName string) error {
	// Find user group by name
	// Python: user_group = UserGroup.get_by_group_name(user_group)
	userGroup, err := c.userGroupRepo.FindByName(ctx, userGroupName)
	if err != nil {
		return fmt.Errorf("failed to find user group: %w", err)
	}
	if userGroup == nil {
		return ErrUserGroupNotFound
	}

	// Delete user group (this should cascade delete namespace permissions)
	// Python: user_group.delete()
	if err := c.userGroupRepo.Delete(ctx, userGroup.ID); err != nil {
		return fmt.Errorf("failed to delete user group: %w", err)
	}

	// Log audit event (async, non-blocking)
	// Python reference: /app/terrareg/models.py:256 - AuditAction.USER_GROUP_DELETE
	go c.userGroupAuditService.LogUserGroupDelete(ctx, userGroupName)

	return nil
}
