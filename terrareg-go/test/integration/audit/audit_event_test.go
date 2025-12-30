package audit

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/audit/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/audit/service"
	auditrepo "github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb/audit"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
	testutils "github.com/matthewjohn/terrareg/terrareg-go/test/integration/testutils"
)

// TestAuditEvent_CreateAuditEvent_AllAuditActions tests creating audit events with all 33 audit action types
func TestAuditEvent_CreateAuditEvent_AllAuditActions(t *testing.T) {
	auditActions := []model.AuditAction{
		model.AuditActionNamespaceCreate,
		model.AuditActionNamespaceModifyName,
		model.AuditActionNamespaceModifyDisplayName,
		model.AuditActionNamespaceDelete,
		model.AuditActionModuleProviderCreate,
		model.AuditActionModuleProviderDelete,
		model.AuditActionModuleProviderUpdateGitTagFormat,
		model.AuditActionModuleProviderUpdateGitProvider,
		model.AuditActionModuleProviderUpdateGitPath,
		model.AuditActionModuleProviderUpdateArchiveGitPath,
		model.AuditActionModuleProviderUpdateGitCustomBaseURL,
		model.AuditActionModuleProviderUpdateGitCustomCloneURL,
		model.AuditActionModuleProviderUpdateGitCustomBrowseURL,
		model.AuditActionModuleProviderUpdateVerified,
		model.AuditActionModuleProviderUpdateNamespace,
		model.AuditActionModuleProviderUpdateModuleName,
		model.AuditActionModuleProviderUpdateProviderName,
		model.AuditActionModuleProviderRedirectDelete,
		model.AuditActionModuleVersionIndex,
		model.AuditActionModuleVersionPublish,
		model.AuditActionModuleVersionDelete,
		model.AuditActionUserGroupCreate,
		model.AuditActionUserGroupDelete,
		model.AuditActionUserGroupNamespacePermissionAdd,
		model.AuditActionUserGroupNamespacePermissionModify,
		model.AuditActionUserGroupNamespacePermissionDelete,
		model.AuditActionUserLogin,
		model.AuditActionGpgKeyCreate,
		model.AuditActionGpgKeyDelete,
		model.AuditActionProviderCreate,
		model.AuditActionProviderDelete,
		model.AuditActionProviderVersionIndex,
		model.AuditActionProviderVersionDelete,
		model.AuditActionRepositoryCreate,
		model.AuditActionRepositoryUpdate,
		model.AuditActionRepositoryDelete,
	}

	for _, auditAction := range auditActions {
		t.Run(string(auditAction), func(t *testing.T) {
			db := testutils.SetupTestDatabase(t)
			defer testutils.CleanupTestDatabase(t, db)

			// Clean audit table
			db.DB.Exec("DELETE FROM audit_history")

			username := "test-user"
			oldValue := "old"
			newValue := "new"

			// Create audit service
			auditRepo := auditrepo.NewAuditHistoryRepository(db.DB)
			auditService := service.NewAuditService(auditRepo)

			// Create audit event
			audit := model.NewAuditHistory(
				username,
				auditAction,
				"unittest-object-type",
				"unittest/object/id",
				&oldValue,
				&newValue,
			)

			err := auditService.LogEvent(context.Background(), audit)
			require.NoError(t, err)

			// Verify audit event from database
			var auditHistoryDB sqldb.AuditHistoryDB
			err = db.DB.First(&auditHistoryDB).Error
			require.NoError(t, err)

			assert.Equal(t, username, *auditHistoryDB.Username)
			assert.Equal(t, sqldb.AuditAction(auditAction), auditHistoryDB.Action)
			assert.Equal(t, "unittest-object-type", *auditHistoryDB.ObjectType)
			assert.Equal(t, "unittest/object/id", *auditHistoryDB.ObjectID)
			assert.Equal(t, oldValue, *auditHistoryDB.OldValue)
			assert.Equal(t, newValue, *auditHistoryDB.NewValue)
			assert.NotNil(t, auditHistoryDB.Timestamp)
			assert.WithinDuration(t, time.Now(), *auditHistoryDB.Timestamp, time.Minute)
		})
	}
}

// TestAuditEvent_OldValue tests creating audit events with various old_value types
func TestAuditEvent_OldValue(t *testing.T) {
	testCases := []struct {
		name     string
		oldValue interface{}
	}{
		{"nil", nil},
		{"empty string", ""},
		{"string value", "testvalue"},
		{"zero int", 0},
		{"int 1234", 1234},
		{"string 1234", "1234"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db := testutils.SetupTestDatabase(t)
			defer testutils.CleanupTestDatabase(t, db)

			// Clean audit table
			db.DB.Exec("DELETE FROM audit_history")

			var oldValuePtr *string
			if tc.oldValue != nil {
				if str, ok := tc.oldValue.(string); ok {
					oldValuePtr = &str
				} else {
					// Convert to string for non-string types
					str := string(rune(tc.oldValue.(int)))
					oldValuePtr = &str
				}
			}

			newValue := "new"
			newValuePtr := &newValue

			// Create audit service
			auditRepo := auditrepo.NewAuditHistoryRepository(db.DB)
			auditService := service.NewAuditService(auditRepo)

			// Create audit event
			audit := model.NewAuditHistory(
				"test-user",
				model.AuditActionModuleProviderCreate,
				"unittest-object-type",
				"unittest/object/id",
				oldValuePtr,
				newValuePtr,
			)

			err := auditService.LogEvent(context.Background(), audit)
			require.NoError(t, err)

			// Verify audit event from database
			var auditHistoryDB sqldb.AuditHistoryDB
			err = db.DB.First(&auditHistoryDB).Error
			require.NoError(t, err)

			if tc.oldValue == nil {
				assert.Nil(t, auditHistoryDB.OldValue)
			} else {
				assert.NotNil(t, auditHistoryDB.OldValue)
			}
		})
	}
}

// TestAuditEvent_NewValue tests creating audit events with various new_value types
func TestAuditEvent_NewValue(t *testing.T) {
	testCases := []struct {
		name     string
		newValue interface{}
	}{
		{"nil", nil},
		{"empty string", ""},
		{"string value", "testvalue"},
		{"zero int", 0},
		{"int 1234", 1234},
		{"string 1234", "1234"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db := testutils.SetupTestDatabase(t)
			defer testutils.CleanupTestDatabase(t, db)

			// Clean audit table
			db.DB.Exec("DELETE FROM audit_history")

			oldValue := "old"
			oldValuePtr := &oldValue

			var newValuePtr *string
			if tc.newValue != nil {
				if str, ok := tc.newValue.(string); ok {
					newValuePtr = &str
				} else {
					// Convert to string for non-string types
					str := string(rune(tc.newValue.(int)))
					newValuePtr = &str
				}
			}

			// Create audit service
			auditRepo := auditrepo.NewAuditHistoryRepository(db.DB)
			auditService := service.NewAuditService(auditRepo)

			// Create audit event
			audit := model.NewAuditHistory(
				"test-user",
				model.AuditActionModuleProviderCreate,
				"unittest-object-type",
				"unittest/object/id",
				oldValuePtr,
				newValuePtr,
			)

			err := auditService.LogEvent(context.Background(), audit)
			require.NoError(t, err)

			// Verify audit event from database
			var auditHistoryDB sqldb.AuditHistoryDB
			err = db.DB.First(&auditHistoryDB).Error
			require.NoError(t, err)

			if tc.newValue == nil {
				assert.Nil(t, auditHistoryDB.NewValue)
			} else {
				assert.NotNil(t, auditHistoryDB.NewValue)
			}
		})
	}
}

// TestAuditEvent_Username tests creating audit events with various username values
func TestAuditEvent_Username(t *testing.T) {
	testCases := []struct {
		name     string
		username string
	}{
		{"normal username", "testusername"},
		{"built-in admin", "Built-in admin"},
		{"empty string", ""},
		{"empty with nil handling", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db := testutils.SetupTestDatabase(t)
			defer testutils.CleanupTestDatabase(t, db)

			// Clean audit table
			db.DB.Exec("DELETE FROM audit_history")

			oldValue := "old"
			newValue := "new"

			// Create audit service
			auditRepo := auditrepo.NewAuditHistoryRepository(db.DB)
			auditService := service.NewAuditService(auditRepo)

			// Create audit event
			audit := model.NewAuditHistory(
				tc.username,
				model.AuditActionModuleProviderCreate,
				"unittest-object-type",
				"unittest/object/id",
				&oldValue,
				&newValue,
			)

			err := auditService.LogEvent(context.Background(), audit)
			require.NoError(t, err)

			// Verify audit event from database
			var auditHistoryDB sqldb.AuditHistoryDB
			err = db.DB.First(&auditHistoryDB).Error
			require.NoError(t, err)

			assert.Equal(t, tc.username, *auditHistoryDB.Username)
		})
	}
}
