package repository

import (
	"context"
	"errors"
	"regexp"
)

// UserGroupRepository defines operations for user group persistence
// Matches Python terrareg.models.UserGroup exactly
type UserGroupRepository interface {
	Create(ctx context.Context, name string, siteAdmin bool) (UserGroup, error)
	GetByName(ctx context.Context, name string) (UserGroup, error)
	GetByID(ctx context.Context, id int) (UserGroup, error)
	Delete(ctx context.Context, id int) error
	List(ctx context.Context) ([]UserGroup, error)
}

// UserGroup represents a user group matching Python schema
// Fields from Python: id, name, site_admin
type UserGroup interface {
	ID() int
	Name() string
	SiteAdmin() bool
}

// PermissionType matches Python UserGroupNamespacePermissionType
type PermissionType string

const (
	PermissionTypeFull   PermissionType = "FULL"
	PermissionTypeModify PermissionType = "MODIFY"
)

// UserGroupPermissionType represents the enum type
// From Python: class UsergroupnamespacepermissionType(enum.Enum)
type UserGroupPermissionType interface {
	Value() string
}

// userGroupPermissionType implements PermissionType
type userGroupPermissionType struct {
	value string
}

func (p userGroupPermissionType) Value() string { return p.value }

// NewUserGroup creates a new user group instance
func NewUserGroup(id int, name string, siteAdmin bool) UserGroup {
	return &userGroup{
		id:        id,
		name:      name,
		siteAdmin: siteAdmin,
	}
}

// userGroup implements UserGroup
type userGroup struct {
	id        int
	name      string
	siteAdmin bool
}

func (g *userGroup) ID() int         { return g.id }
func (g *userGroup) Name() string    { return g.name }
func (g *userGroup) SiteAdmin() bool { return g.siteAdmin }

// ValidateName matches Python _validate_name method
func ValidateName(name string) error {
	// Python regex: ^[\s0-9a-zA-Z-_]+$
	re := regexp.MustCompile(`^[\s0-9a-zA-Z-_]+$`)
	if !re.MatchString(name) {
		return errors.New("User group name is invalid")
	}
	return nil
}

// NewPermissionType creates a new permission type
func NewPermissionType(value string) UserGroupPermissionType {
	return userGroupPermissionType{value: value}
}

// Permission types matching Python enum
var (
	PermissionTypeFullEnum   = NewPermissionType("FULL")
	PermissionTypeModifyEnum = NewPermissionType("MODIFY")
)
