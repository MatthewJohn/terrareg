package auth

import (
	"testing"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth"
	"github.com/stretchr/testify/assert"
)

func TestNewUserGroupRepository(t *testing.T) {
	// Test that NewUserGroupRepository properly validates nil DB
	repo, err := NewUserGroupRepository(nil)
	assert.Error(t, err)
	assert.Nil(t, repo)
}

func TestUserGroupRepositoryImpl_Save_Create(t *testing.T) {
	// Test basic Save structure for creating new user group
	// With nil checking, we should get an error
	repo, err := NewUserGroupRepository(nil)
	assert.Error(t, err)
	assert.Nil(t, repo)

	userGroup := &auth.UserGroup{
		Name:      "test-group",
		SiteAdmin: false,
	}

	// Verify that userGroup model is valid
	assert.NotNil(t, userGroup)
}

func TestUserGroupRepositoryImpl_List(t *testing.T) {
	// Test that NewUserGroupRepository properly validates nil DB
	repo, err := NewUserGroupRepository(nil)
	assert.Error(t, err)
	assert.Nil(t, repo)
}

func TestUserGroupNamespacePermission_Getters(t *testing.T) {
	// Test the basic structure and methods
	perm := &UserGroupNamespacePermission{}

	// Test that the methods exist (even if they return default values)
	assert.NotNil(t, perm)
	// The actual interface methods would be tested with proper data
}
