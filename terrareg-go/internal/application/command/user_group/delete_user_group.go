package user_group

import (
	"context"
	"errors"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/identity/repository"
)

// DeleteUserGroupCommand deletes a user group
// Matches Python API: DELETE /v1/terrareg/auth/user-group/{group}
type DeleteUserGroupCommand struct {
	groupRepo repository.UserGroupRepository
}

// NewDeleteUserGroupCommand creates a new DeleteUserGroupCommand
func NewDeleteUserGroupCommand(groupRepo repository.UserGroupRepository) *DeleteUserGroupCommand {
	return &DeleteUserGroupCommand{
		groupRepo: groupRepo,
	}
}

// Execute deletes the user group
func (c *DeleteUserGroupCommand) Execute(ctx context.Context, groupName string) error {
	// Get group by name
	group, err := c.groupRepo.GetByName(ctx, groupName)
	if err != nil {
		return err
	}

	if group == nil {
		return errors.New("User group does not exist")
	}

	// Delete the group
	return c.groupRepo.Delete(ctx, group.ID())
}