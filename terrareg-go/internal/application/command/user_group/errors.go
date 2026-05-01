package user_group

import "errors"

var (
	// ErrUserGroupAlreadyExists is returned when a user group with the same name already exists
	ErrUserGroupAlreadyExists = errors.New("user group already exists")

	// ErrInvalidUserGroupName is returned when the user group name is invalid
	ErrInvalidUserGroupName = errors.New("invalid user group name")

	// ErrInvalidSiteAdminValue is returned when site_admin is not True or False
	ErrInvalidSiteAdminValue = errors.New("site_admin must be True or False")

	// ErrUserGroupNotFound is returned when a user group with the given name doesn't exist
	ErrUserGroupNotFound = errors.New("user group does not exist")

	// ErrNamespaceNotFound is returned when a namespace with the given name doesn't exist
	ErrNamespaceNotFound = errors.New("namespace does not exist")

	// ErrPermissionAlreadyExists is returned when the permission already exists for this user_group/namespace
	ErrPermissionAlreadyExists = errors.New("permission already exists for this user_group/namespace")

	// ErrPermissionNotFound is returned when the permission doesn't exist for this user_group/namespace
	ErrPermissionNotFound = errors.New("permission does not exist")

	// ErrInvalidPermissionType is returned when the permission type is invalid
	ErrInvalidPermissionType = errors.New("invalid namespace permission type")
)
