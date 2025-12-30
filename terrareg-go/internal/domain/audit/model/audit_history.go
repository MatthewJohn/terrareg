package model

import (
	"time"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
)

// AuditAction represents the type of audit action
type AuditAction string

const (
	// Namespace actions (matching Python audit_action.py snake_case format)
	AuditActionNamespaceCreate            AuditAction = "namespace_create"
	AuditActionNamespaceModifyName        AuditAction = "namespace_modify_name"
	AuditActionNamespaceModifyDisplayName AuditAction = "namespace_modify_display_name"
	AuditActionNamespaceDelete            AuditAction = "namespace_delete"

	// Module provider actions
	AuditActionModuleProviderCreate                   AuditAction = "module_provider_create"
	AuditActionModuleProviderDelete                   AuditAction = "module_provider_delete"
	AuditActionModuleProviderUpdateGitTagFormat       AuditAction = "module_provider_update_git_tag_format"
	AuditActionModuleProviderUpdateGitProvider        AuditAction = "module_provider_update_git_provider"
	AuditActionModuleProviderUpdateGitPath            AuditAction = "module_provider_update_git_path"
	AuditActionModuleProviderUpdateArchiveGitPath     AuditAction = "module_provider_update_archive_git_path"
	AuditActionModuleProviderUpdateGitCustomBaseURL   AuditAction = "module_provider_update_git_custom_base_url"
	AuditActionModuleProviderUpdateGitCustomCloneURL  AuditAction = "module_provider_update_git_custom_clone_url"
	AuditActionModuleProviderUpdateGitCustomBrowseURL AuditAction = "module_provider_update_git_custom_browse_url"
	AuditActionModuleProviderUpdateVerified           AuditAction = "module_provider_update_verified"
	AuditActionModuleProviderUpdateNamespace          AuditAction = "module_provider_update_namespace"
	AuditActionModuleProviderUpdateModuleName         AuditAction = "module_provider_update_module_name"
	AuditActionModuleProviderUpdateProviderName       AuditAction = "module_provider_update_provider_name"
	AuditActionModuleProviderRedirectDelete           AuditAction = "module_provider_redirect_delete"

	// Module version actions
	AuditActionModuleVersionIndex   AuditAction = "module_version_index"
	AuditActionModuleVersionPublish AuditAction = "module_version_publish"
	AuditActionModuleVersionDelete  AuditAction = "module_version_delete"

	// User group actions
	AuditActionUserGroupCreate                    AuditAction = "user_group_create"
	AuditActionUserGroupDelete                    AuditAction = "user_group_delete"
	AuditActionUserGroupNamespacePermissionAdd    AuditAction = "user_group_namespace_permission_add"
	AuditActionUserGroupNamespacePermissionModify AuditAction = "user_group_namespace_permission_modify"
	AuditActionUserGroupNamespacePermissionDelete AuditAction = "user_group_namespace_permission_delete"

	// User authentication
	AuditActionUserLogin AuditAction = "user_login"

	// GPG key actions
	AuditActionGpgKeyCreate AuditAction = "gpg_key_create"
	AuditActionGpgKeyDelete AuditAction = "gpg_key_delete"

	// Provider actions
	AuditActionProviderCreate        AuditAction = "provider_create"
	AuditActionProviderDelete        AuditAction = "provider_delete"
	AuditActionProviderVersionIndex  AuditAction = "provider_version_index"
	AuditActionProviderVersionDelete AuditAction = "provider_version_delete"

	// Repository actions
	AuditActionRepositoryCreate AuditAction = "repository_create"
	AuditActionRepositoryUpdate AuditAction = "repository_update"
	AuditActionRepositoryDelete AuditAction = "repository_delete"
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
