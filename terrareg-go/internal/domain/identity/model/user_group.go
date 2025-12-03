package model

import (
	"time"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared"
)

// UserGroup represents a group of users for permissions management
type UserGroup struct {
	id          string
	name        string
	description string
	members     []*User
	permissions []GroupPermission
	createdAt   time.Time
	updatedAt   time.Time
}

// GroupPermission represents a permission granted to a user group
type GroupPermission struct {
	id           int
	resourceType ResourceType
	resourceID   string
	action       Action
	grantedAt    time.Time
	grantedBy    *User
}

// NewUserGroup creates a new user group
func NewUserGroup(name, description string) (*UserGroup, error) {
	if err := ValidateGroupName(name); err != nil {
		return nil, err
	}

	return &UserGroup{
		id:          shared.GenerateID(),
		name:        name,
		description: description,
		members:     make([]*User, 0),
		permissions: make([]GroupPermission, 0),
		createdAt:   time.Now(),
		updatedAt:   time.Now(),
	}, nil
}

// Business methods

// AddMember adds a user to the group
func (ug *UserGroup) AddMember(user *User) error {
	if user == nil {
		return ErrUserRequired
	}

	// Check if user is already a member
	for _, member := range ug.members {
		if member.id == user.ID() {
			return ErrUserAlreadyInGroup
		}
	}

	ug.members = append(ug.members, user)
	ug.updatedAt = time.Now()
	return nil
}

// RemoveMember removes a user from the group
func (ug *UserGroup) RemoveMember(userID string) error {
	if userID == "" {
		return ErrUserIDRequired
	}

	for i, member := range ug.members {
		if member.id == userID {
			ug.members = append(ug.members[:i], ug.members[i+1:]...)
			ug.updatedAt = time.Now()
			return nil
		}
	}

	return ErrUserNotInGroup
}

// HasMember checks if a user is a member of the group
func (ug *UserGroup) HasMember(userID string) bool {
	for _, member := range ug.members {
		if member.id == userID {
			return true
		}
	}
	return false
}

// AddPermission adds a permission to the group
func (ug *UserGroup) AddPermission(resourceType ResourceType, resourceID string, action Action, grantedBy *User) error {
	if grantedBy == nil {
		return ErrUserRequired
	}

	// Check if permission already exists
	for _, perm := range ug.permissions {
		if perm.resourceType == resourceType && perm.resourceID == resourceID && perm.action == action {
			return ErrPermissionAlreadyExists
		}
	}

	permission := GroupPermission{
		id:           shared.GenerateIntID(),
		resourceType: resourceType,
		resourceID:   resourceID,
		action:       action,
		grantedAt:    time.Now(),
		grantedBy:    grantedBy,
	}

	ug.permissions = append(ug.permissions, permission)
	ug.updatedAt = time.Now()
	return nil
}

// RemovePermission removes a permission from the group
func (ug *UserGroup) RemovePermission(permissionID int) error {
	for i, perm := range ug.permissions {
		if perm.id == permissionID {
			ug.permissions = append(ug.permissions[:i], ug.permissions[i+1:]...)
			ug.updatedAt = time.Now()
			return nil
		}
	}

	return ErrPermissionNotFound
}

// HasPermission checks if the group has a specific permission
func (ug *UserGroup) HasPermission(resourceType ResourceType, resourceID string, action Action) bool {
	for _, perm := range ug.permissions {
		if perm.resourceType == resourceType && perm.resourceID == resourceID {
			if perm.action == action || perm.action == ActionAdmin {
				return true
			}
		}
	}
	return false
}

// UpdateDetails updates the group's details
func (ug *UserGroup) UpdateDetails(name, description string) error {
	if name != "" {
		if err := ValidateGroupName(name); err != nil {
			return err
		}
		ug.name = name
	}
	if description != "" {
		ug.description = description
	}
	ug.updatedAt = time.Now()
	return nil
}

// GetActiveMembers returns only active members
func (ug *UserGroup) GetActiveMembers() []*User {
	var activeMembers []*User
	for _, member := range ug.members {
		if member.Active() {
			activeMembers = append(activeMembers, member)
		}
	}
	return activeMembers
}

// Getters
func (ug *UserGroup) ID() string                      { return ug.id }
func (ug *UserGroup) Name() string                    { return ug.name }
func (ug *UserGroup) Description() string              { return ug.description }
func (ug *UserGroup) Members() []*User                 { return ug.members }
func (ug *UserGroup) Permissions() []GroupPermission   { return ug.permissions }
func (ug *UserGroup) CreatedAt() time.Time             { return ug.createdAt }
func (ug *UserGroup) UpdatedAt() time.Time             { return ug.updatedAt }

// GroupPermission getters
func (gp GroupPermission) ID() int           { return gp.id }
func (gp GroupPermission) ResourceType() ResourceType { return gp.resourceType }
func (gp GroupPermission) ResourceID() string { return gp.resourceID }
func (gp GroupPermission) Action() Action     { return gp.action }
func (gp GroupPermission) GrantedAt() time.Time { return gp.grantedAt }
func (gp GroupPermission) GrantedBy() *User   { return gp.grantedBy }