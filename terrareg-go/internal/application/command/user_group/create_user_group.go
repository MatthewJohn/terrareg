package user_group

import (
	"context"
	"errors"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/identity/repository"
)

// CreateUserGroupCommand creates a new user group
// Matches Python UserGroup.create()
type CreateUserGroupCommand struct {
	groupRepo repository.UserGroupRepository
}

// NewCreateUserGroupCommand creates a new CreateUserGroupCommand
func NewCreateUserGroupCommand(groupRepo repository.UserGroupRepository) *CreateUserGroupCommand {
	return &CreateUserGroupCommand{
		groupRepo: groupRepo,
	}
}

// CreateUserGroupRequest represents the request to create a user group
// Matches Python API: {name: string, site_admin: boolean}
type CreateUserGroupRequest struct {
	Name      string `json:"name"`
	SiteAdmin bool   `json:"site_admin"`
}

// CreateUserGroupResponse represents the response after creating a user group
// Matches Python API response
type CreateUserGroupResponse struct {
	Name      string `json:"name"`
	SiteAdmin bool   `json:"site_admin"`
	Success   bool   `json:"success"`
}

// Execute creates a new user group
func (c *CreateUserGroupCommand) Execute(ctx context.Context, req *CreateUserGroupRequest) (*CreateUserGroupResponse, error) {
	// Validate group name
	if err := repository.ValidateName(req.Name); err != nil {
		return nil, err
	}

	// Check if group already exists
	existing, err := c.groupRepo.GetByName(ctx, req.Name)
	if err != nil {
		return nil, err
	}

	if existing != nil {
		return nil, nil // Python returns None if group exists
	}

	// Create new group
	group, err := c.groupRepo.Create(ctx, req.Name, req.SiteAdmin)
	if err != nil {
		return nil, err
	}

	if group == nil {
		return nil, errors.New("failed to create user group")
	}

	return &CreateUserGroupResponse{
		Name:      group.Name(),
		SiteAdmin: group.SiteAdmin(),
		Success:   true,
	}, nil
}
