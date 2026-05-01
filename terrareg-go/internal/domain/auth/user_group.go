package auth

import (
	"fmt"
)

// UserGroup represents a group of users for permissions management
// Matches Python's user_group table structure
type UserGroup struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	SiteAdmin bool   `json:"site_admin"`
}

// NamespacePermission represents a permission granted to a user group for a namespace
// Matches Python's user_group_namespace_permission table structure
type NamespacePermission struct {
	UserGroupID    int            `json:"user_group_id"`
	NamespaceID    int            `json:"namespace_id"`
	PermissionType PermissionType `json:"permission_type"`
}

// PermissionType represents permission types for namespace access
type PermissionType string

const (
	PermissionRead   PermissionType = "READ"
	PermissionModify PermissionType = "MODIFY"
	PermissionFull   PermissionType = "FULL"
)

// NewUserGroup creates a new user group
func NewUserGroup(name string, siteAdmin bool) (*UserGroup, error) {
	if err := ValidateGroupName(name); err != nil {
		return nil, err
	}

	return &UserGroup{
		Name:      name,
		SiteAdmin: siteAdmin,
	}, nil
}

// NewNamespacePermission creates a new namespace permission
func NewNamespacePermission(userGroupID, namespaceID int, permissionType PermissionType) (*NamespacePermission, error) {
	if userGroupID <= 0 {
		return nil, fmt.Errorf("user group ID required")
	}
	if namespaceID <= 0 {
		return nil, fmt.Errorf("namespace ID required")
	}

	return &NamespacePermission{
		UserGroupID:    userGroupID,
		NamespaceID:    namespaceID,
		PermissionType: permissionType,
	}, nil
}

// Business methods

// UpdateName updates the group name
func (ug *UserGroup) UpdateName(name string) error {
	if err := ValidateGroupName(name); err != nil {
		return err
	}

	ug.Name = name
	return nil
}

// SetSiteAdmin updates the site admin flag
func (ug *UserGroup) SetSiteAdmin(siteAdmin bool) {
	ug.SiteAdmin = siteAdmin
}

// HasPermission checks if this permission grants the specified access
func (np *NamespacePermission) HasPermission(requestedType PermissionType) bool {
	switch requestedType {
	case PermissionRead:
		return np.PermissionType == PermissionRead ||
			np.PermissionType == PermissionModify ||
			np.PermissionType == PermissionFull
	case PermissionModify:
		return np.PermissionType == PermissionModify ||
			np.PermissionType == PermissionFull
	case PermissionFull:
		return np.PermissionType == PermissionFull
	default:
		return false
	}
}

// PermissionLevel represents the hierarchical level of a permission
func (np *NamespacePermission) PermissionLevel() int {
	switch np.PermissionType {
	case PermissionRead:
		return 1
	case PermissionModify:
		return 2
	case PermissionFull:
		return 3
	default:
		return 0
	}
}

// IsHigherThan checks if this permission is higher level than another
func (np *NamespacePermission) IsHigherThan(other PermissionType) bool {
	otherPerm, _ := NewNamespacePermission(0, 0, other)
	return np.PermissionLevel() > otherPerm.PermissionLevel()
}

// Getters for UserGroup
func (ug *UserGroup) GetID() int        { return ug.ID }
func (ug *UserGroup) GetName() string   { return ug.Name }
func (ug *UserGroup) IsSiteAdmin() bool { return ug.SiteAdmin }

// Getters for NamespacePermission
func (np *NamespacePermission) GetUserGroupID() int               { return np.UserGroupID }
func (np *NamespacePermission) GetNamespaceID() int               { return np.NamespaceID }
func (np *NamespacePermission) GetPermissionType() PermissionType { return np.PermissionType }

// Permission checking helper functions

// CheckNamespacePermission checks if a set of permissions includes access to a namespace
func CheckNamespacePermission(permissions []NamespacePermission, namespaceID int, requestedType PermissionType) bool {
	for _, perm := range permissions {
		if perm.NamespaceID == namespaceID && perm.HasPermission(requestedType) {
			return true
		}
	}
	return false
}

// GetHighestNamespacePermission returns the highest level permission for a namespace
func GetHighestNamespacePermission(permissions []NamespacePermission, namespaceID int) PermissionType {
	highest := PermissionType("")

	for _, perm := range permissions {
		if perm.NamespaceID == namespaceID {
			if highest == "" || perm.IsHigherThan(highest) {
				highest = perm.PermissionType
			}
		}
	}

	return highest
}

// GroupPermissionsByNamespace groups permissions by namespace
func GroupPermissionsByNamespace(permissions []NamespacePermission) map[int][]NamespacePermission {
	grouped := make(map[int][]NamespacePermission)

	for _, perm := range permissions {
		grouped[perm.NamespaceID] = append(grouped[perm.NamespaceID], perm)
	}

	return grouped
}

// Validation function

// ValidateGroupName validates a user group name
func ValidateGroupName(name string) error {
	if name == "" {
		return fmt.Errorf("group name cannot be empty")
	}
	if len(name) > 128 {
		return fmt.Errorf("group name cannot exceed 128 characters")
	}
	// Add additional validation as needed
	return nil
}
