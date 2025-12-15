package repository

import (
	"context"
)

// UserGroupNamespacePermissionRepository defines operations for user group namespace permissions
// Matches Python UserGroupNamespacePermission model
type UserGroupNamespacePermissionRepository interface {
	GetByUserGroup(ctx context.Context, userGroupName string) ([]UserGroupNamespacePermission, error)
}

// UserGroupNamespacePermission represents namespace permissions for user groups
// Matches Python UserGroupNamespacePermission model
type UserGroupNamespacePermission struct {
	UserGroupID    int
	NamespaceID    int
	NamespaceName  string // Added for convenience (joined from namespace table)
	PermissionType PermissionType
}
