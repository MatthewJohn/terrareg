package auth

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestNewUserGroupRepository(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	require.NoError(t, err)

	repo := NewUserGroupRepository(gormDB)
	assert.NotNil(t, repo)
}

func TestUserGroupRepositoryImpl_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	require.NoError(t, err)

	repo := NewUserGroupRepository(gormDB)

	userGroup := &auth.UserGroup{
		Name:        "test-group",
		SiteAdmin:   false,
		Description: "Test group description",
	}

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "user_groups"`).
		WithArgs(userGroup.Name, userGroup.SiteAdmin, userGroup.Description).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	err = repo.Create(context.Background(), userGroup)

	assert.NoError(t, err)
	assert.Equal(t, 1, userGroup.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserGroupRepositoryImpl_FindByID_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	require.NoError(t, err)

	repo := NewUserGroupRepository(gormDB)

	expectedUserGroup := &auth.UserGroup{
		ID:          1,
		Name:        "test-group",
		SiteAdmin:   false,
		Description: "Test group description",
	}

	rows := sqlmock.NewRows([]string{"id", "name", "site_admin", "description"}).
		AddRow(1, "test-group", false, "Test group description")

	mock.ExpectQuery(`SELECT \* FROM "user_groups"`).
		WithArgs(1).
		WillReturnRows(rows)

	result, err := repo.FindByID(context.Background(), 1)

	assert.NoError(t, err)
	assert.Equal(t, expectedUserGroup, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserGroupRepositoryImpl_FindByID_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	require.NoError(t, err)

	repo := NewUserGroupRepository(gormDB)

	mock.ExpectQuery(`SELECT \* FROM "user_groups"`).
		WithArgs(999).
		WillReturnError(gorm.ErrRecordNotFound)

	result, err := repo.FindByID(context.Background(), 999)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, gorm.ErrRecordNotFound, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserGroupRepositoryImpl_FindByName_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	require.NoError(t, err)

	repo := NewUserGroupRepository(gormDB)

	expectedUserGroup := &auth.UserGroup{
		ID:          1,
		Name:        "test-group",
		SiteAdmin:   true,
		Description: "Test admin group",
	}

	rows := sqlmock.NewRows([]string{"id", "name", "site_admin", "description"}).
		AddRow(1, "test-group", true, "Test admin group")

	mock.ExpectQuery(`SELECT \* FROM "user_groups"`).
		WithArgs("test-group").
		WillReturnRows(rows)

	result, err := repo.FindByName(context.Background(), "test-group")

	assert.NoError(t, err)
	assert.Equal(t, expectedUserGroup, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserGroupRepositoryImpl_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	require.NoError(t, err)

	repo := NewUserGroupRepository(gormDB)

	userGroup := &auth.UserGroup{
		ID:          1,
		Name:        "updated-group",
		SiteAdmin:   true,
		Description: "Updated description",
	}

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "user_groups"`).
		WithArgs(userGroup.Name, userGroup.SiteAdmin, userGroup.Description, userGroup.ID).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err = repo.Update(context.Background(), userGroup)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserGroupRepositoryImpl_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	require.NoError(t, err)

	repo := NewUserGroupRepository(gormDB)

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "user_groups"`).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err = repo.Delete(context.Background(), 1)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserGroupRepositoryImpl_List(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	require.NoError(t, err)

	repo := NewUserGroupRepository(gormDB)

	rows := sqlmock.NewRows([]string{"id", "name", "site_admin", "description"}).
		AddRow(1, "admin-group", true, "Admin group").
		AddRow(2, "user-group", false, "User group")

	mock.ExpectQuery(`SELECT \* FROM "user_groups"`).
		WillReturnRows(rows)

	result, err := repo.List(context.Background())

	assert.NoError(t, err)
	assert.Len(t, result, 2)

	assert.Equal(t, 1, result[0].ID)
	assert.Equal(t, "admin-group", result[0].Name)
	assert.True(t, result[0].SiteAdmin)
	assert.Equal(t, "Admin group", result[0].Description)

	assert.Equal(t, 2, result[1].ID)
	assert.Equal(t, "user-group", result[1].Name)
	assert.False(t, result[1].SiteAdmin)
	assert.Equal(t, "User group", result[1].Description)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserGroupRepositoryImpl_FindSiteAdminGroups(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	require.NoError(t, err)

	repo := NewUserGroupRepository(gormDB)

	rows := sqlmock.NewRows([]string{"id", "name", "site_admin", "description"}).
		AddRow(1, "admin-group", true, "Admin group").
		AddRow(2, "super-admin", true, "Super admin group")

	mock.ExpectQuery(`SELECT \* FROM "user_groups"`).
		WithArgs(true).
		WillReturnRows(rows)

	result, err := repo.FindSiteAdminGroups(context.Background())

	assert.NoError(t, err)
	assert.Len(t, result, 2)

	for _, group := range result {
		assert.True(t, group.SiteAdmin)
	}

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserGroupRepositoryImpl_CreateNamespacePermission(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	require.NoError(t, err)

	repo := NewUserGroupRepository(gormDB)

	userGroupID := 1
	namespaceID := 1
	permissionType := auth.PermissionFull

	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO "user_group_namespace_permissions"`).
		WithArgs(userGroupID, namespaceID, string(permissionType)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err = repo.CreateNamespacePermission(context.Background(), userGroupID, namespaceID, permissionType)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserGroupRepositoryImpl_DeleteNamespacePermission(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	require.NoError(t, err)

	repo := NewUserGroupRepository(gormDB)

	userGroupID := 1
	namespaceID := 1

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "user_group_namespace_permissions"`).
		WithArgs(userGroupID, namespaceID).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err = repo.DeleteNamespacePermission(context.Background(), userGroupID, namespaceID)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserGroupRepositoryImpl_GetNamespacePermissions(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	require.NoError(t, err)

	repo := NewUserGroupRepository(gormDB)

	userGroupID := 1

	rows := sqlmock.NewRows([]string{"id", "user_group_id", "namespace_id", "permission_type"}).
		AddRow(1, 1, 1, "READ").
		AddRow(2, 1, 2, "FULL")

	mock.ExpectQuery(`SELECT \* FROM "user_group_namespace_permissions"`).
		WithArgs(userGroupID).
		WillReturnRows(rows)

	result, err := repo.GetNamespacePermissions(context.Background(), userGroupID)

	assert.NoError(t, err)
	assert.Len(t, result, 2)

	perm1 := result[0].(*UserGroupNamespacePermission)
	assert.Equal(t, 1, perm1.ID)
	assert.Equal(t, 1, perm1.UserGroupID)
	assert.Equal(t, 1, perm1.NamespaceID)
	assert.Equal(t, auth.PermissionRead, perm1.PermissionType)

	perm2 := result[1].(*UserGroupNamespacePermission)
	assert.Equal(t, 2, perm2.ID)
	assert.Equal(t, 1, perm2.UserGroupID)
	assert.Equal(t, 2, perm2.NamespaceID)
	assert.Equal(t, auth.PermissionFull, perm2.PermissionType)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserGroupNamespacePermission_Getters(t *testing.T) {
	perm := &UserGroupNamespacePermission{
		ID:             1,
		UserGroupID:    2,
		NamespaceID:    3,
		PermissionType: auth.PermissionFull,
	}

	assert.Equal(t, 1, perm.GetID())
	assert.Equal(t, 2, perm.GetUserGroupID())
	assert.Equal(t, 3, perm.GetNamespaceID())
	assert.Equal(t, auth.PermissionFull, perm.GetPermissionType())
}