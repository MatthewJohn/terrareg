package model

import (
	"time"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared"
)

// NamespacePermission represents a permission granted to a user group for a namespace
type NamespacePermission struct {
	id           int
	namespaceID  int
	userGroupID  int
	grantedBy    *User
	grantedAt    time.Time
	permissions  []Action
}

// NewNamespacePermission creates a new namespace permission
func NewNamespacePermission(namespaceID, userGroupID int, actions []Action, grantedBy *User) (*NamespacePermission, error) {
	if namespaceID <= 0 {
		return nil, ErrNamespaceIDRequired
	}
	if userGroupID <= 0 {
		return nil, ErrUserGroupIDRequired
	}
	if grantedBy == nil {
		return nil, ErrUserRequired
	}
	if len(actions) == 0 {
		return nil, ErrActionsRequired
	}

	return &NamespacePermission{
		id:          shared.GenerateIntID(),
		namespaceID: namespaceID,
		userGroupID: userGroupID,
		grantedBy:   grantedBy,
		grantedAt:   time.Now(),
		permissions: actions,
	}, nil
}

// HasAction checks if the permission includes a specific action
func (np *NamespacePermission) HasAction(action Action) bool {
	for _, permAction := range np.permissions {
		if permAction == action || permAction == ActionAdmin {
			return true
		}
	}
	return false
}

// AddAction adds an action to the permission
func (np *NamespacePermission) AddAction(action Action) error {
	// Check if action already exists
	if np.HasAction(action) {
		return ErrActionAlreadyExists
	}

	np.permissions = append(np.permissions, action)
	return nil
}

// RemoveAction removes an action from the permission
func (np *NamespacePermission) RemoveAction(action Action) error {
	for i, permAction := range np.permissions {
		if permAction == action {
			np.permissions = append(np.permissions[:i], np.permissions[i+1:]...)
			if len(np.permissions) == 0 {
				return ErrCannotRemoveAllActions
			}
			return nil
		}
	}
	return ErrActionNotFound
}

// Getters
func (np *NamespacePermission) ID() int           { return np.id }
func (np *NamespacePermission) NamespaceID() int  { return np.namespaceID }
func (np *NamespacePermission) UserGroupID() int  { return np.userGroupID }
func (np *NamespacePermission) GrantedBy() *User   { return np.grantedBy }
func (np *NamespacePermission) GrantedAt() time.Time { return np.grantedAt }
func (np *NamespacePermission) Permissions() []Action { return np.permissions }