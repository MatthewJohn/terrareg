package sqldb

import (
	"time"
)

// Common constants for column sizes (matching Python version)
const (
	GeneralColumnSize  = 128
	LargeColumnSize    = 1024
	URLColumnSize      = 1024
	MediumBlobSize     = (1 << 24) - 1 // 16MB - 1
	IdentityColumnSize = 255
)

// Enum types

type NamespaceType string

const (
	NamespaceTypeNone       NamespaceType = "NONE"
	NamespaceTypeGithubUser NamespaceType = "GITHUB_USER"
	NamespaceTypeGithubOrg  NamespaceType = "GITHUB_ORGANISATION"
)

type UserGroupNamespacePermissionType string

const (
	PermissionTypeFull   UserGroupNamespacePermissionType = "FULL"
	PermissionTypeModify UserGroupNamespacePermissionType = "MODIFY"
	PermissionTypeRead   UserGroupNamespacePermissionType = "READ"
)

type ProviderTier string

const (
	ProviderTierOfficial  ProviderTier = "official"
	ProviderTierPartner   ProviderTier = "partner"
	ProviderTierCommunity ProviderTier = "community"
)

type ProviderSourceType string

const (
	ProviderSourceTypeGithub    ProviderSourceType = "github"
	ProviderSourceTypeGitlab    ProviderSourceType = "gitlab"
	ProviderSourceTypeBitbucket ProviderSourceType = "bitbucket"
)

type ProviderDocumentationType string

const (
	ProviderDocTypeOverview   ProviderDocumentationType = "overview"
	ProviderDocTypeResource   ProviderDocumentationType = "resource"
	ProviderDocTypeDataSource ProviderDocumentationType = "data-source"
	ProviderDocTypeGuide      ProviderDocumentationType = "guide"
	ProviderDocTypeFunction   ProviderDocumentationType = "function"
)

type ProviderBinaryOperatingSystemType string

const (
	OSLinux   ProviderBinaryOperatingSystemType = "linux"
	OSDarwin  ProviderBinaryOperatingSystemType = "darwin"
	OSWindows ProviderBinaryOperatingSystemType = "windows"
	OSFreeBSD ProviderBinaryOperatingSystemType = "freebsd"
)

type ProviderBinaryArchitectureType string

const (
	ArchAMD64 ProviderBinaryArchitectureType = "amd64"
	ArchARM64 ProviderBinaryArchitectureType = "arm64"
	ArchARM   ProviderBinaryArchitectureType = "arm"
	Arch386   ProviderBinaryArchitectureType = "386"
)

type AuditAction string

const (
	// Audit action constants matching Python audit_action.py (snake_case format)
	AuditActionNamespaceCreate                        AuditAction = "namespace_create"
	AuditActionNamespaceModifyName                    AuditAction = "namespace_modify_name"
	AuditActionNamespaceModifyDisplayName             AuditAction = "namespace_modify_display_name"
	AuditActionNamespaceDelete                        AuditAction = "namespace_delete"
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
	AuditActionModuleVersionIndex                     AuditAction = "module_version_index"
	AuditActionModuleVersionPublish                   AuditAction = "module_version_publish"
	AuditActionModuleVersionDelete                    AuditAction = "module_version_delete"
	AuditActionUserGroupCreate                        AuditAction = "user_group_create"
	AuditActionUserGroupDelete                        AuditAction = "user_group_delete"
	AuditActionUserGroupNamespacePermissionAdd        AuditAction = "user_group_namespace_permission_add"
	AuditActionUserGroupNamespacePermissionModify     AuditAction = "user_group_namespace_permission_modify"
	AuditActionUserGroupNamespacePermissionDelete     AuditAction = "user_group_namespace_permission_delete"
	AuditActionUserLogin                              AuditAction = "user_login"
	AuditActionGpgKeyCreate                           AuditAction = "gpg_key_create"
	AuditActionGpgKeyDelete                           AuditAction = "gpg_key_delete"
	AuditActionProviderCreate                         AuditAction = "provider_create"
	AuditActionProviderDelete                         AuditAction = "provider_delete"
	AuditActionProviderVersionIndex                   AuditAction = "provider_version_index"
	AuditActionProviderVersionDelete                  AuditAction = "provider_version_delete"
	AuditActionRepositoryCreate                       AuditAction = "repository_create"
	AuditActionRepositoryUpdate                       AuditAction = "repository_update"
	AuditActionRepositoryDelete                       AuditAction = "repository_delete"
)

// Database Models (matching Python SQLAlchemy schema exactly)

// SessionDB represents the session table
type SessionDB struct {
	ID                 string    `gorm:"type:varchar(128);primaryKey"`
	Expiry             time.Time `gorm:"not null"`
	ProviderSourceAuth []byte    `gorm:"type:mediumblob"`
}

func (SessionDB) TableName() string {
	return "session"
}

// TerraformIDPAuthorizationCodeDB represents OAuth authorization codes
type TerraformIDPAuthorizationCodeDB struct {
	ID     int       `gorm:"primaryKey;autoIncrement"`
	Key    string    `gorm:"type:varchar(128);not null;uniqueIndex"`
	Data   []byte    `gorm:"type:mediumblob"`
	Expiry time.Time `gorm:"not null"`
}

func (TerraformIDPAuthorizationCodeDB) TableName() string {
	return "terraform_idp_authorization_code"
}

// TerraformIDPAccessTokenDB represents OAuth access tokens
type TerraformIDPAccessTokenDB struct {
	ID     int       `gorm:"primaryKey;autoIncrement"`
	Key    string    `gorm:"type:varchar(128);not null;uniqueIndex"`
	Data   []byte    `gorm:"type:mediumblob"`
	Expiry time.Time `gorm:"not null"`
}

func (TerraformIDPAccessTokenDB) TableName() string {
	return "terraform_idp_access_token"
}

// TerraformIDPSubjectIdentifierDB represents OAuth subject identifiers
type TerraformIDPSubjectIdentifierDB struct {
	ID     int       `gorm:"primaryKey;autoIncrement"`
	Key    string    `gorm:"type:varchar(128);not null;uniqueIndex"`
	Data   []byte    `gorm:"type:mediumblob"`
	Expiry time.Time `gorm:"not null"`
}

func (TerraformIDPSubjectIdentifierDB) TableName() string {
	return "terraform_idp_subject_identifier"
}

// UserGroupDB represents user groups
type UserGroupDB struct {
	ID        int    `gorm:"primaryKey;autoIncrement"`
	Name      string `gorm:"type:varchar(128);not null;uniqueIndex"`
	SiteAdmin bool   `gorm:"default:false;not null"`
}

func (UserGroupDB) TableName() string {
	return "user_group"
}

// UserGroupNamespacePermissionDB represents namespace permissions
type UserGroupNamespacePermissionDB struct {
	UserGroupID    int                              `gorm:"primaryKey;not null"`
	NamespaceID    int                              `gorm:"primaryKey;not null"`
	PermissionType UserGroupNamespacePermissionType `gorm:"type:varchar(50)"`

	UserGroup UserGroupDB `gorm:"foreignKey:UserGroupID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Namespace NamespaceDB `gorm:"foreignKey:NamespaceID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (UserGroupNamespacePermissionDB) TableName() string {
	return "user_group_namespace_permission"
}

// GitProviderDB represents git providers
type GitProviderDB struct {
	ID                int    `gorm:"primaryKey"`
	Name              string `gorm:"type:varchar(128);uniqueIndex"`
	BaseURLTemplate   string `gorm:"type:varchar(1024)"`
	CloneURLTemplate  string `gorm:"type:varchar(1024)"`
	BrowseURLTemplate string `gorm:"type:varchar(1024)"`
	GitPathTemplate   string `gorm:"type:varchar(1024)"`
}

func (GitProviderDB) TableName() string {
	return "git_provider"
}

// NamespaceDB represents namespaces
type NamespaceDB struct {
	ID            int           `gorm:"primaryKey;autoIncrement"`
	Namespace     string        `gorm:"type:varchar(128);not null"`
	DisplayName   *string       `gorm:"type:varchar(128)"`
	NamespaceType NamespaceType `gorm:"type:varchar(50);not null;default:'NONE'"`
}

func (NamespaceDB) TableName() string {
	return "namespace"
}

// NamespaceRedirectDB represents namespace redirects
type NamespaceRedirectDB struct {
	ID          int    `gorm:"primaryKey;autoIncrement"`
	Name        string `gorm:"type:varchar(128);not null"`
	NamespaceID int    `gorm:"not null"`

	Namespace NamespaceDB `gorm:"foreignKey:NamespaceID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (NamespaceRedirectDB) TableName() string {
	return "namespace_redirect"
}

// ModuleProviderDB represents module providers
type ModuleProviderDB struct {
	ID                    int     `gorm:"primaryKey;autoIncrement"`
	NamespaceID           int     `gorm:"not null"`
	Module                string  `gorm:"type:varchar(128)"`
	Provider              string  `gorm:"type:varchar(128)"`
	RepoBaseURLTemplate   *string `gorm:"type:varchar(1024)"`
	RepoCloneURLTemplate  *string `gorm:"type:varchar(1024)"`
	RepoBrowseURLTemplate *string `gorm:"type:varchar(1024)"`
	GitTagFormat          *string `gorm:"type:varchar(128)"`
	GitPath               *string `gorm:"type:varchar(1024)"`
	ArchiveGitPath        bool    `gorm:"default:false"`
	Verified              *bool   `gorm:"default:null"`
	GitProviderID         *int    `gorm:"default:null"`
	LatestVersionID       *int    `gorm:"default:null"`

	Namespace     NamespaceDB      `gorm:"foreignKey:NamespaceID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	GitProvider   *GitProviderDB   `gorm:"foreignKey:GitProviderID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	LatestVersion *ModuleVersionDB `gorm:"foreignKey:LatestVersionID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
}

func (ModuleProviderDB) TableName() string {
	return "module_provider"
}

// ModuleProviderSearchResult represents a module provider with relevance score for search
type ModuleProviderSearchResult struct {
	// Module Provider fields
	ID                    int     `gorm:"column:module_provider_id"`
	NamespaceID           int     `gorm:"column:module_provider_namespace_id"`
	Module                string  `gorm:"column:module_provider_module"`
	Provider              string  `gorm:"column:module_provider_provider"`
	RepoBaseURLTemplate   *string `gorm:"column:module_provider_repo_base_url_template"`
	RepoCloneURLTemplate  *string `gorm:"column:module_provider_repo_clone_url_template"`
	RepoBrowseURLTemplate *string `gorm:"column:module_provider_repo_browse_url_template"`
	GitTagFormat          *string `gorm:"column:module_provider_git_tag_format"`
	GitPath               *string `gorm:"column:module_provider_git_path"`
	ArchiveGitPath        bool    `gorm:"column:module_provider_archive_git_path"`
	Verified              *bool   `gorm:"column:module_provider_verified"`
	GitProviderID         *int    `gorm:"column:module_provider_git_provider_id"`
	LatestVersionID       *int    `gorm:"column:module_provider_latest_version_id"`

	// Namespace fields
	NamespaceName        string `gorm:"column:namespace_namespace"`
	NamespaceDisplayName string `gorm:"column:namespace_display_name"`
	NamespaceType        string `gorm:"column:namespace_type"`

	// Search relevance score
	RelevanceScore *int `gorm:"column:relevance_score"`
}

// ModuleProviderRedirectDB represents module provider redirects
type ModuleProviderRedirectDB struct {
	ID               int    `gorm:"primaryKey;autoIncrement"`
	Module           string `gorm:"type:varchar(128);not null"`
	Provider         string `gorm:"type:varchar(128);not null"`
	NamespaceID      int    `gorm:"not null"`
	ModuleProviderID int    `gorm:"not null"`

	Namespace      NamespaceDB      `gorm:"foreignKey:NamespaceID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	ModuleProvider ModuleProviderDB `gorm:"foreignKey:ModuleProviderID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (ModuleProviderRedirectDB) TableName() string {
	return "module_provider_redirect"
}

// ModuleDetailsDB represents module metadata
type ModuleDetailsDB struct {
	ID               int    `gorm:"primaryKey;autoIncrement"`
	ReadmeContent    []byte `gorm:"type:mediumblob"`
	TerraformDocs    []byte `gorm:"type:mediumblob"`
	Tfsec            []byte `gorm:"type:mediumblob"`
	Infracost        []byte `gorm:"type:mediumblob"`
	TerraformGraph   []byte `gorm:"type:mediumblob"`
	TerraformModules []byte `gorm:"type:mediumblob"`
	TerraformVersion []byte `gorm:"type:mediumblob"`
}

func (ModuleDetailsDB) TableName() string {
	return "module_details"
}

// ModuleVersionDB represents module versions
type ModuleVersionDB struct {
	ID                    int     `gorm:"primaryKey;autoIncrement"`
	ModuleProviderID      int     `gorm:"not null"`
	Version               string  `gorm:"type:varchar(128)"`
	GitSHA                *string `gorm:"type:varchar(128)"`
	GitPath               *string `gorm:"type:varchar(1024)"`
	ArchiveGitPath        bool    `gorm:"default:false"`
	ModuleDetailsID       *int    `gorm:"default:null"`
	Beta                  bool    `gorm:"not null"`
	Owner                 *string `gorm:"type:varchar(128)"`
	Description           *string `gorm:"type:varchar(1024)"`
	RepoBaseURLTemplate   *string `gorm:"type:varchar(1024)"`
	RepoCloneURLTemplate  *string `gorm:"type:varchar(1024)"`
	RepoBrowseURLTemplate *string `gorm:"type:varchar(1024)"`
	PublishedAt           *time.Time
	VariableTemplate      []byte `gorm:"type:mediumblob"`
	Internal              bool   `gorm:"not null"`
	Published             *bool  `gorm:"default:null"`
	ExtractionVersion     *int   `gorm:"default:null"`

	ModuleProvider ModuleProviderDB `gorm:"foreignKey:ModuleProviderID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	ModuleDetails  *ModuleDetailsDB `gorm:"foreignKey:ModuleDetailsID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (ModuleVersionDB) TableName() string {
	return "module_version"
}

// SubmoduleDB represents submodules
type SubmoduleDB struct {
	ID                  int     `gorm:"primaryKey;autoIncrement"`
	ParentModuleVersion int     `gorm:"not null"`
	ModuleDetailsID     *int    `gorm:"default:null"`
	Type                *string `gorm:"type:varchar(128)"`
	Path                string  `gorm:"type:varchar(1024)"`
	Name                *string `gorm:"type:varchar(128)"`

	ParentVersion *ModuleVersionDB `gorm:"foreignKey:ParentModuleVersion;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	ModuleDetails *ModuleDetailsDB `gorm:"foreignKey:ModuleDetailsID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (SubmoduleDB) TableName() string {
	return "submodule"
}

// AnalyticsDB represents module analytics
type AnalyticsDB struct {
	ID                  int `gorm:"primaryKey;autoIncrement"`
	ParentModuleVersion int `gorm:"index;not null"`
	Timestamp           *time.Time
	TerraformVersion    *string `gorm:"type:varchar(128)"`
	AnalyticsToken      *string `gorm:"type:varchar(128)"`
	AuthToken           *string `gorm:"type:varchar(128)"`
	Environment         *string `gorm:"type:varchar(128)"`
	NamespaceName       *string `gorm:"type:varchar(128)"`
	ModuleName          *string `gorm:"type:varchar(128)"`
	ProviderName        *string `gorm:"type:varchar(128)"`
}

func (AnalyticsDB) TableName() string {
	return "analytics"
}

// ProviderAnalyticsDB represents provider analytics
type ProviderAnalyticsDB struct {
	ID                int `gorm:"primaryKey;autoIncrement"`
	ProviderVersionID int `gorm:"index;not null"`
	Timestamp         *time.Time
	TerraformVersion  *string `gorm:"type:varchar(128)"`
	NamespaceName     *string `gorm:"type:varchar(128)"`
	ProviderName      *string `gorm:"type:varchar(128)"`
}

func (ProviderAnalyticsDB) TableName() string {
	return "provider_analytics"
}

// ExampleFileDB represents example files
type ExampleFileDB struct {
	ID          int    `gorm:"primaryKey;autoIncrement"`
	SubmoduleID int    `gorm:"not null"`
	Path        string `gorm:"type:varchar(128);not null"`
	Content     []byte `gorm:"type:mediumblob"`

	Submodule SubmoduleDB `gorm:"foreignKey:SubmoduleID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (ExampleFileDB) TableName() string {
	return "example_file"
}

// ModuleVersionFileDB represents additional module files
type ModuleVersionFileDB struct {
	ID              int    `gorm:"primaryKey;autoIncrement"`
	ModuleVersionID int    `gorm:"not null"`
	Path            string `gorm:"type:varchar(128);not null"`
	Content         []byte `gorm:"type:mediumblob"`

	ModuleVersion ModuleVersionDB `gorm:"foreignKey:ModuleVersionID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (ModuleVersionFileDB) TableName() string {
	return "module_version_file"
}

// GPGKeyDB represents GPG keys
type GPGKeyDB struct {
	ID          int     `gorm:"primaryKey;autoIncrement"`
	NamespaceID int     `gorm:"not null"`
	ASCIIArmor  []byte  `gorm:"type:mediumblob"`
	KeyID       *string `gorm:"type:varchar(1024)"`
	Fingerprint *string `gorm:"type:varchar(1024)"`
	Source      *string `gorm:"type:varchar(1024)"`
	SourceURL   *string `gorm:"type:varchar(1024)"`
	CreatedAt   *time.Time
	UpdatedAt   *time.Time

	Namespace NamespaceDB `gorm:"foreignKey:NamespaceID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (GPGKeyDB) TableName() string {
	return "gpg_key"
}

// ProviderSourceDB represents provider sources (GitHub, GitLab, etc.)
type ProviderSourceDB struct {
	Name               string             `gorm:"type:varchar(128);primaryKey"`
	APIName            *string            `gorm:"type:varchar(128)"`
	ProviderSourceType ProviderSourceType `gorm:"type:varchar(50)"`
	Config             []byte             `gorm:"type:mediumblob"`
}

func (ProviderSourceDB) TableName() string {
	return "provider_source"
}

// ProviderCategoryDB represents provider categories
type ProviderCategoryDB struct {
	ID             int     `gorm:"primaryKey;autoIncrement"`
	Name           *string `gorm:"type:varchar(128)"`
	Slug           string  `gorm:"type:varchar(128);uniqueIndex"`
	UserSelectable bool    `gorm:"default:true"`
}

func (ProviderCategoryDB) TableName() string {
	return "provider_category"
}

// RepositoryDB represents repositories
type RepositoryDB struct {
	ID                 int     `gorm:"primaryKey;autoIncrement"`
	ProviderID         *string `gorm:"type:varchar(128)"`
	Owner              *string `gorm:"type:varchar(128)"`
	Name               *string `gorm:"type:varchar(128)"`
	Description        []byte  `gorm:"type:mediumblob"`
	CloneURL           *string `gorm:"type:varchar(1024)"`
	LogoURL            *string `gorm:"type:varchar(1024)"`
	ProviderSourceName string  `gorm:"type:varchar(128);not null"`

	ProviderSource ProviderSourceDB `gorm:"foreignKey:ProviderSourceName;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (RepositoryDB) TableName() string {
	return "repository"
}

// ProviderDB represents providers
type ProviderDB struct {
	ID                        int          `gorm:"primaryKey;autoIncrement"`
	NamespaceID               int          `gorm:"not null"`
	Name                      string       `gorm:"type:varchar(128)"`
	Description               *string      `gorm:"type:varchar(1024)"`
	Tier                      ProviderTier `gorm:"type:varchar(50)"`
	DefaultProviderSourceAuth bool         `gorm:"default:false"`
	ProviderCategoryID        *int         `gorm:"default:null"`
	RepositoryID              *int         `gorm:"default:null"`
	LatestVersionID           *int         `gorm:"default:null"`

	Namespace        NamespaceDB         `gorm:"foreignKey:NamespaceID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	ProviderCategory *ProviderCategoryDB `gorm:"foreignKey:ProviderCategoryID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	Repository       *RepositoryDB       `gorm:"foreignKey:RepositoryID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	LatestVersion    *ProviderVersionDB  `gorm:"foreignKey:LatestVersionID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
}

func (ProviderDB) TableName() string {
	return "provider"
}

// ProviderVersionDB represents provider versions
type ProviderVersionDB struct {
	ID                int     `gorm:"primaryKey;autoIncrement"`
	ProviderID        int     `gorm:"not null"`
	GPGKeyID          int     `gorm:"not null"`
	Version           string  `gorm:"type:varchar(128)"`
	GitTag            *string `gorm:"type:varchar(128)"`
	Beta              bool    `gorm:"not null"`
	PublishedAt       *time.Time
	ExtractionVersion *int   `gorm:"default:null"`
	ProtocolVersions  []byte `gorm:"type:mediumblob"`

	Provider ProviderDB `gorm:"foreignKey:ProviderID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	GPGKey   GPGKeyDB   `gorm:"foreignKey:GPGKeyID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (ProviderVersionDB) TableName() string {
	return "provider_version"
}

// ProviderVersionDocumentationDB represents provider documentation
type ProviderVersionDocumentationDB struct {
	ID                int                       `gorm:"primaryKey;autoIncrement"`
	ProviderVersionID int                       `gorm:"not null"`
	Name              string                    `gorm:"type:varchar(128);not null"`
	Slug              string                    `gorm:"type:varchar(128);not null"`
	Title             *string                   `gorm:"type:varchar(128)"`
	Description       []byte                    `gorm:"type:mediumblob"`
	Language          string                    `gorm:"type:varchar(128);not null"`
	Subcategory       *string                   `gorm:"type:varchar(128)"`
	Filename          string                    `gorm:"type:varchar(128);not null"`
	DocumentationType ProviderDocumentationType `gorm:"type:varchar(50);not null"`
	Content           []byte                    `gorm:"type:mediumblob"`

	ProviderVersion ProviderVersionDB `gorm:"foreignKey:ProviderVersionID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (ProviderVersionDocumentationDB) TableName() string {
	return "provider_version_documentation"
}

// ProviderVersionBinaryDB represents provider binaries
type ProviderVersionBinaryDB struct {
	ID                int                               `gorm:"primaryKey;autoIncrement"`
	ProviderVersionID int                               `gorm:"not null"`
	Name              string                            `gorm:"type:varchar(128);not null"`
	OperatingSystem   ProviderBinaryOperatingSystemType `gorm:"type:varchar(50);not null"`
	Architecture      ProviderBinaryArchitectureType    `gorm:"type:varchar(50);not null"`
	Checksum          string                            `gorm:"type:varchar(128);not null"`

	ProviderVersion ProviderVersionDB `gorm:"foreignKey:ProviderVersionID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (ProviderVersionBinaryDB) TableName() string {
	return "provider_version_binary"
}

// UserDB represents users (derived from auth data and session storage)
type UserDB struct {
	ID             string `gorm:"type:varchar(128);primaryKey"`
	Username       string `gorm:"type:varchar(128);not null"`
	DisplayName    string `gorm:"type:varchar(128)"`
	Email          string `gorm:"type:varchar(128)"`
	AuthMethod     string `gorm:"type:varchar(50);not null"`
	AuthProviderID string `gorm:"type:varchar(128)"`
	ExternalID     string `gorm:"type:varchar(128)"`
	AccessToken    string `gorm:"type:varchar(1024)"`
	RefreshToken   string `gorm:"type:varchar(1024)"`
	TokenExpiry    *time.Time
	Active         bool      `gorm:"default:true;not null"`
	CreatedAt      time.Time `gorm:"not null"`
	LastLoginAt    *time.Time
}

func (UserDB) TableName() string {
	return "user"
}

// UserGroupMemberDB represents user-group membership
type UserGroupMemberDB struct {
	UserGroupID int       `gorm:"primaryKey;not null"`
	UserID      string    `gorm:"type:varchar(128);primaryKey;not null"`
	JoinedAt    time.Time `gorm:"not null"`

	UserGroup UserGroupDB `gorm:"foreignKey:UserGroupID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (UserGroupMemberDB) TableName() string {
	return "user_group_member"
}

// UserPermissionDB represents direct user permissions (inherited from group membership)
type UserPermissionDB struct {
	ID           int       `gorm:"primaryKey;autoIncrement"`
	UserID       string    `gorm:"type:varchar(128);not null"`
	ResourceType string    `gorm:"type:varchar(50);not null"`
	ResourceID   string    `gorm:"type:varchar(128);not null"`
	Action       string    `gorm:"type:varchar(50);not null"`
	GrantedBy    string    `gorm:"type:varchar(128)"`
	GrantedAt    time.Time `gorm:"not null"`
}

func (UserPermissionDB) TableName() string {
	return "user_permission"
}

// AuditHistoryDB represents audit trail
type AuditHistoryDB struct {
	ID         int `gorm:"primaryKey;autoIncrement"`
	Timestamp  *time.Time
	Username   *string     `gorm:"type:varchar(128)"`
	Action     AuditAction `gorm:"type:varchar(50)"`
	ObjectType *string     `gorm:"type:varchar(128)"`
	ObjectID   *string     `gorm:"type:varchar(128)"`
	OldValue   *string     `gorm:"type:varchar(128)"`
	NewValue   *string     `gorm:"type:varchar(128)"`
}

func (AuditHistoryDB) TableName() string {
	return "audit_history"
}

// AuthenticationTokenDB represents authentication tokens for API access
type AuthenticationTokenDB struct {
	ID          int        `gorm:"primaryKey;autoIncrement"`
	TokenType   string     `gorm:"type:enum('admin','upload','publish');not null"`
	TokenValue  string     `gorm:"type:varchar(255);not null;uniqueIndex"`
	NamespaceID *int       `gorm:"default:null"` // Only for publish tokens
	Description string     `gorm:"type:text;not null"`
	CreatedAt   time.Time  `gorm:"not null"`
	ExpiresAt   *time.Time `gorm:"default:null"`
	IsActive    bool       `gorm:"default:true;not null"`
	CreatedBy   string     `gorm:"type:varchar(128);not null"`

	// Relationships
	Namespace *NamespaceDB `gorm:"foreignKey:NamespaceID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
}

func (AuthenticationTokenDB) TableName() string {
	return "authentication_tokens"
}
