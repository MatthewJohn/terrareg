package model

import (
	"time"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
)

// AuditAction represents the type of audit action
type AuditAction string

const (
	// Namespace actions
	AuditActionNamespaceCreate AuditAction = "NAMESPACE_CREATE"
	AuditActionNamespaceUpdate AuditAction = "NAMESPACE_UPDATE"
	AuditActionNamespaceDelete AuditAction = "NAMESPACE_DELETE"

	// Module provider actions
	AuditActionModuleProviderCreate AuditAction = "MODULE_PROVIDER_CREATE"
	AuditActionModuleProviderUpdate AuditAction = "MODULE_PROVIDER_UPDATE"
	AuditActionModuleProviderDelete AuditAction = "MODULE_PROVIDER_DELETE"

	// Module version actions
	AuditActionModuleVersionCreate  AuditAction = "MODULE_VERSION_CREATE"
	AuditActionModuleVersionPublish AuditAction = "MODULE_VERSION_PUBLISH"
	AuditActionModuleVersionDelete  AuditAction = "MODULE_VERSION_DELETE"
	AuditActionModuleVersionIndex   AuditAction = "MODULE_VERSION_INDEX"

	// User group actions
	AuditActionUserGroupCreate AuditAction = "USER_GROUP_CREATE"
	AuditActionUserGroupUpdate AuditAction = "USER_GROUP_UPDATE"
	AuditActionUserGroupDelete AuditAction = "USER_GROUP_DELETE"

	// GPG key actions
	AuditActionGpgKeyCreate AuditAction = "GPG_KEY_CREATE"
	AuditActionGpgKeyUpdate AuditAction = "GPG_KEY_UPDATE"
	AuditActionGpgKeyDelete AuditAction = "GPG_KEY_DELETE"

	// Integration actions
	AuditActionIntegrationCreate AuditAction = "INTEGRATION_CREATE"
	AuditActionIntegrationUpdate AuditAction = "INTEGRATION_UPDATE"
	AuditActionIntegrationDelete AuditAction = "INTEGRATION_DELETE"
)

// AuditHistory represents an audit log entry
type AuditHistory struct {
	id         int
	timestamp  time.Time
	username   string
	action     AuditAction
	objectType string
	objectID   string
	oldValue   *string
	newValue   *string
}

// NewAuditHistory creates a new audit history entry
func NewAuditHistory(
	username string,
	action AuditAction,
	objectType string,
	objectID string,
	oldValue *string,
	newValue *string,
) *AuditHistory {
	return &AuditHistory{
		timestamp:  time.Now(),
		username:   username,
		action:     action,
		objectType: objectType,
		objectID:   objectID,
		oldValue:   oldValue,
		newValue:   newValue,
	}
}

// ID returns the audit history ID
func (a *AuditHistory) ID() int {
	return a.id
}

// Timestamp returns when the audit entry was created
func (a *AuditHistory) Timestamp() time.Time {
	return a.timestamp
}

// Username returns the user who performed the action
func (a *AuditHistory) Username() string {
	return a.username
}

// Action returns the audit action type
func (a *AuditHistory) Action() AuditAction {
	return a.action
}

// ObjectType returns the type of object that was acted upon
func (a *AuditHistory) ObjectType() string {
	return a.objectType
}

// ObjectID returns the ID of the object that was acted upon
func (a *AuditHistory) ObjectID() string {
	return a.objectID
}

// OldValue returns the previous value (if any)
func (a *AuditHistory) OldValue() *string {
	return a.oldValue
}

// NewValue returns the new value (if any)
func (a *AuditHistory) NewValue() *string {
	return a.newValue
}

// SetID sets the audit history ID
func (a *AuditHistory) SetID(id int) {
	a.id = id
}

// SetTimestamp sets the audit history timestamp
func (a *AuditHistory) SetTimestamp(timestamp time.Time) {
	a.timestamp = timestamp
}

// ToDBModel converts the domain model to the database model
func (a *AuditHistory) ToDBModel() *sqldb.AuditHistoryDB {
	action := sqldb.AuditAction(a.action)
	return &sqldb.AuditHistoryDB{
		ID:         a.id,
		Timestamp:  &a.timestamp,
		Username:   &a.username,
		Action:     action,
		ObjectType: &a.objectType,
		ObjectID:   &a.objectID,
		OldValue:   a.oldValue,
		NewValue:   a.newValue,
	}
}

// AuditHistorySearchResult represents the result of an audit history search
type AuditHistorySearchResult struct {
	Records       []*AuditHistory
	TotalCount    int
	FilteredCount int
	Draw          int
}

// AuditHistorySearchQuery represents search criteria for audit history
type AuditHistorySearchQuery struct {
	SearchValue string
	Length      int
	Start       int
	Draw        int
	OrderDir    string
	OrderColumn int
}
