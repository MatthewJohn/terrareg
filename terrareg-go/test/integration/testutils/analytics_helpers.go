package testutils

import (
	"fmt"
	"testing"
	"time"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
)

// CreateAnalyticsData creates an analytics entry for testing
func CreateAnalyticsData(t *testing.T, db *sqldb.Database, moduleVersionID int, downloads int, timestamp time.Time) {
	t.Helper()

	for i := 0; i < downloads; i++ {
		analytics := sqldb.AnalyticsDB{
			ParentModuleVersion: moduleVersionID,
			Timestamp:           &timestamp,
			TerraformVersion:    stringPtr("1.5.0"),
			AnalyticsToken:      stringPtr("test-token"),
			AuthToken:           stringPtr("auth-token"),
			Environment:         stringPtr("production"),
			NamespaceName:       stringPtr("test-namespace"),
			ModuleName:          stringPtr("test-module"),
			ProviderName:        stringPtr("aws"),
		}
		err := db.DB.Create(&analytics).Error
		if err != nil {
			t.Fatalf("Failed to create analytics data: %v", err)
		}
	}
}

// CreateAnalyticsDataWithDetails creates an analytics entry with full details for testing
func CreateAnalyticsDataWithDetails(t *testing.T, db *sqldb.Database, moduleVersionID int, timestamp time.Time, terraformVersion, analyticsToken, authToken, environment, namespace, module, provider string) {
	t.Helper()

	analytics := sqldb.AnalyticsDB{
		ParentModuleVersion: moduleVersionID,
		Timestamp:           &timestamp,
		TerraformVersion:    stringPtr(terraformVersion),
		AnalyticsToken:      stringPtr(analyticsToken),
		AuthToken:           stringPtr(authToken),
		Environment:         stringPtr(environment),
		NamespaceName:       stringPtr(namespace),
		ModuleName:          stringPtr(module),
		ProviderName:        stringPtr(provider),
	}
	err := db.DB.Create(&analytics).Error
	if err != nil {
		t.Fatalf("Failed to create analytics data with details: %v", err)
	}
}

// CreateAuditLog creates an audit log entry for testing
func CreateAuditLog(t *testing.T, db *sqldb.Database, username, action, objectType string, objectID int) {
	t.Helper()

	objectIDStr := fmt.Sprintf("%d", objectID)
	auditLog := sqldb.AuditHistoryDB{
		Username:   &username,
		Action:     sqldb.AuditAction(action),
		ObjectType: &objectType,
		ObjectID:   &objectIDStr,
	}
	err := db.DB.Create(&auditLog).Error
	if err != nil {
		t.Fatalf("Failed to create audit log: %v", err)
	}
}

// CreateAuditLogWithDetails creates an audit log entry with full details for testing
func CreateAuditLogWithDetails(t *testing.T, db *sqldb.Database, timestamp time.Time, username, action, objectType, objectID, oldValue, newValue string) {
	t.Helper()

	auditLog := sqldb.AuditHistoryDB{
		Timestamp:  &timestamp,
		Username:   &username,
		Action:     sqldb.AuditAction(action),
		ObjectType: &objectType,
		ObjectID:   &objectID,
		OldValue:   &oldValue,
		NewValue:   &newValue,
	}
	err := db.DB.Create(&auditLog).Error
	if err != nil {
		t.Fatalf("Failed to create audit log with details: %v", err)
	}
}

// CreateMultipleAuditLogs creates multiple audit log entries for testing pagination
func CreateMultipleAuditLogs(t *testing.T, db *sqldb.Database, count int, username, action, objectType string) {
	t.Helper()

	timestamp := time.Now()
	for i := 0; i < count; i++ {
		objectID := fmt.Sprintf("%d", i+1)
		auditLog := sqldb.AuditHistoryDB{
			Timestamp:  &timestamp,
			Username:   &username,
			Action:     sqldb.AuditAction(action),
			ObjectType: &objectType,
			ObjectID:   &objectID,
		}
		err := db.DB.Create(&auditLog).Error
		if err != nil {
			t.Fatalf("Failed to create audit log %d: %v", i, err)
		}
		// Increment timestamp by 1 minute for each log to ensure consistent ordering
		timestamp = timestamp.Add(time.Minute)
	}
}
