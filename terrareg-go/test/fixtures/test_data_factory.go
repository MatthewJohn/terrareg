package fixtures

import (
	"time"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
)

// TestDataFactory provides factory methods for creating test data
type TestDataFactory struct {
	namespaceCounter int
	moduleCounter    int
	versionCounter   int
	userCounter      int
}

// NewTestDataFactory creates a new test data factory
func NewTestDataFactory() *TestDataFactory {
	return &TestDataFactory{}
}

// CreateNamespace creates a test namespace with default or custom values
func (f *TestDataFactory) CreateNamespace(overrides ...NamespaceOverride) sqldb.NamespaceDB {
	f.namespaceCounter++

	displayName := f.generateDisplayName("Namespace")
	// Default values
	namespace := sqldb.NamespaceDB{
		Namespace:     f.generateNamespaceName(),
		DisplayName:   &displayName,
		NamespaceType: sqldb.NamespaceTypeNone,
	}

	// Apply overrides
	for _, override := range overrides {
		override(&namespace)
	}

	return namespace
}

// NamespaceOverride is a function to override namespace defaults
type NamespaceOverride func(*sqldb.NamespaceDB)

// WithNamespaceDisplayName sets a custom display name
func WithNamespaceDisplayName(displayName string) NamespaceOverride {
	return func(n *sqldb.NamespaceDB) {
		n.DisplayName = &displayName
	}
}

// WithNamespaceType sets a custom namespace type
func WithNamespaceType(namespaceType sqldb.NamespaceType) NamespaceOverride {
	return func(n *sqldb.NamespaceDB) {
		n.NamespaceType = namespaceType
	}
}

// WithPrivateNamespace makes the namespace private
func WithPrivateNamespace() NamespaceOverride {
	return func(n *sqldb.NamespaceDB) {
		// Private namespaces could use a specific namespace type
		// For now, this is a placeholder as the actual implementation may vary
	}
}

// CreateModuleProvider creates a test module provider with default or custom values
func (f *TestDataFactory) CreateModuleProvider(namespaceID int, overrides ...ModuleProviderOverride) sqldb.ModuleProviderDB {
	f.moduleCounter++

	module := sqldb.ModuleProviderDB{
		NamespaceID:           namespaceID,
		Module:                f.generateModuleName(),
		Provider:              f.generateProviderName(),
		Verified:              nil, // false by default
		GitProviderID:         nil,
		RepoBaseURLTemplate:   nil,
		RepoCloneURLTemplate:  nil,
		RepoBrowseURLTemplate: nil,
		GitTagFormat:          nil,
		GitPath:               nil,
		ArchiveGitPath:        false,
		LatestVersionID:       nil,
	}

	// Apply overrides
	for _, override := range overrides {
		override(&module)
	}

	return module
}

// ModuleProviderOverride is a function to override module provider defaults
type ModuleProviderOverride func(*sqldb.ModuleProviderDB)

// WithModuleName sets a custom module name
func WithModuleName(name string) ModuleProviderOverride {
	return func(mp *sqldb.ModuleProviderDB) {
		mp.Module = name
	}
}

// WithProviderName sets a custom provider name
func WithProviderName(name string) ModuleProviderOverride {
	return func(mp *sqldb.ModuleProviderDB) {
		mp.Provider = name
	}
}

// WithVerified marks the module provider as verified
func WithVerified() ModuleProviderOverride {
	return func(mp *sqldb.ModuleProviderDB) {
		mp.Verified = &[]bool{true}[0]
	}
}

// WithGitConfig sets Git configuration for the module provider
func WithGitConfig(gitProviderID int, tagFormat, gitPath string) ModuleProviderOverride {
	return func(mp *sqldb.ModuleProviderDB) {
		mp.GitProviderID = &gitProviderID
		mp.GitTagFormat = &tagFormat
		mp.GitPath = &gitPath
	}
}

// CreateModuleVersion creates a test module version with default or custom values
func (f *TestDataFactory) CreateModuleVersion(moduleProviderID int, overrides ...ModuleVersionOverride) sqldb.ModuleVersionDB {
	f.versionCounter++

	version := sqldb.ModuleVersionDB{
		ModuleProviderID:      moduleProviderID,
		Version:               f.generateVersion(),
		Beta:                  false,
		Internal:              false,
		Published:             &[]bool{false}[0], // false by default
		PublishedAt:           nil,
		GitSHA:                nil,
		GitPath:               nil,
		ArchiveGitPath:        false,
		RepoBaseURLTemplate:   nil,
		RepoCloneURLTemplate:  nil,
		RepoBrowseURLTemplate: nil,
		Owner:                 nil,
		Description:           nil,
		VariableTemplate:      nil,
		ExtractionVersion:     nil,
		ModuleDetailsID:       nil,
	}

	// Apply overrides
	for _, override := range overrides {
		override(&version)
	}

	return version
}

// ModuleVersionOverride is a function to override module version defaults
type ModuleVersionOverride func(*sqldb.ModuleVersionDB)

// WithVersion sets a custom version
func WithVersion(version string) ModuleVersionOverride {
	return func(mv *sqldb.ModuleVersionDB) {
		mv.Version = version
	}
}

// WithPublished marks the version as published
func WithPublished() ModuleVersionOverride {
	return func(mv *sqldb.ModuleVersionDB) {
		mv.PublishedAt = &[]time.Time{time.Now()}[0]
		mv.Published = &[]bool{true}[0]
	}
}

// WithBeta marks the version as beta
func WithBeta() ModuleVersionOverride {
	return func(mv *sqldb.ModuleVersionDB) {
		mv.Beta = true
	}
}

// WithInternal marks the version as internal
func WithInternal() ModuleVersionOverride {
	return func(mv *sqldb.ModuleVersionDB) {
		mv.Internal = true
	}
}

// WithDescription sets a description for the version
func WithDescription(description string) ModuleVersionOverride {
	return func(mv *sqldb.ModuleVersionDB) {
		mv.Description = &description
	}
}

// WithOwner sets the owner of the version
func WithOwner(owner string) ModuleVersionOverride {
	return func(mv *sqldb.ModuleVersionDB) {
		mv.Owner = &owner
	}
}

// CreateModuleDetails creates test module details
func (f *TestDataFactory) CreateModuleDetails(overrides ...ModuleDetailsOverride) sqldb.ModuleDetailsDB {
	details := sqldb.ModuleDetailsDB{
		ReadmeContent:    []byte("# Test Module\n\nThis is a test module for testing purposes."),
		TerraformDocs:    []byte(`{"variables": [], "outputs": [], "resources": []}`),
		Tfsec:            []byte(`{"results": [], "summary": {"total": 0}}`),
		Infracost:        []byte(`{"total_monthly_cost": 0, "projects": []}`),
		TerraformGraph:   []byte(`{"nodes": [], "edges": []}`),
		TerraformModules: []byte(`{}`),
		TerraformVersion: []byte(">= 1.0.0"),
	}

	// Apply overrides
	for _, override := range overrides {
		override(&details)
	}

	return details
}

// ModuleDetailsOverride is a function to override module details defaults
type ModuleDetailsOverride func(*sqldb.ModuleDetailsDB)

// WithReadmeContent sets custom README content
func WithReadmeContent(content string) ModuleDetailsOverride {
	return func(md *sqldb.ModuleDetailsDB) {
		md.ReadmeContent = []byte(content)
	}
}

// WithTerraformDocs sets custom Terraform docs
func WithTerraformDocs(docs string) ModuleDetailsOverride {
	return func(md *sqldb.ModuleDetailsDB) {
		md.TerraformDocs = []byte(docs)
	}
}

// WithSecurityIssues sets custom security scan results
func WithSecurityIssues(issues string) ModuleDetailsOverride {
	return func(md *sqldb.ModuleDetailsDB) {
		md.Tfsec = []byte(issues)
	}
}

// CreateUserGroup creates a test user group with default or custom values
func (f *TestDataFactory) CreateUserGroup(overrides ...UserGroupOverride) sqldb.UserGroupDB {
	f.userCounter++

	userGroup := sqldb.UserGroupDB{
		Name:      f.generateGroupName(),
		SiteAdmin: false,
	}

	// Apply overrides
	for _, override := range overrides {
		override(&userGroup)
	}

	return userGroup
}

// UserGroupOverride is a function to override user group defaults
type UserGroupOverride func(*sqldb.UserGroupDB)

// WithGroupName sets a custom group name
func WithGroupName(name string) UserGroupOverride {
	return func(ug *sqldb.UserGroupDB) {
		ug.Name = name
	}
}

// WithSiteAdmin makes the group a site admin group
func WithSiteAdmin() UserGroupOverride {
	return func(ug *sqldb.UserGroupDB) {
		ug.SiteAdmin = true
	}
}

// WithMembers sets group members (Note: members are tracked separately, not in UserGroupDB)
func WithMembers(members ...string) UserGroupOverride {
	return func(ug *sqldb.UserGroupDB) {
		// Members are tracked in UserGroupMemberDB, not UserGroupDB
		// This is a placeholder for future implementation
	}
}

// CreateAnalytics creates test analytics data
func (f *TestDataFactory) CreateAnalytics(parentModuleVersionID int, overrides ...AnalyticsOverride) sqldb.AnalyticsDB {
	token := f.generateAnalyticsToken()
	timestamp := time.Now()
	terraformVersion := "1.0.0"
	analytics := sqldb.AnalyticsDB{
		ParentModuleVersion: parentModuleVersionID,
		Timestamp:           &timestamp,
		AnalyticsToken:      &token,
		TerraformVersion:    &terraformVersion,
	}

	// Apply overrides
	for _, override := range overrides {
		override(&analytics)
	}

	return analytics
}

// AnalyticsOverride is a function to override analytics defaults
type AnalyticsOverride func(*sqldb.AnalyticsDB)

// WithAnalyticsToken sets a custom analytics token
func WithAnalyticsToken(token string) AnalyticsOverride {
	return func(a *sqldb.AnalyticsDB) {
		a.AnalyticsToken = &token
	}
}

// WithTimestamp sets a custom timestamp
func WithTimestamp(timestamp time.Time) AnalyticsOverride {
	return func(a *sqldb.AnalyticsDB) {
		a.Timestamp = &timestamp
	}
}

// CreateSubmodule creates a test submodule
func (f *TestDataFactory) CreateSubmodule(parentModuleVersionID, moduleDetailsID int, overrides ...SubmoduleOverride) sqldb.SubmoduleDB {
	submodule := sqldb.SubmoduleDB{
		ParentModuleVersion: parentModuleVersionID,
		Path:                "submodule",
		Name:                &[]string{"Test Submodule"}[0],
		Type:                &[]string{"module"}[0],
		ModuleDetailsID:     &moduleDetailsID,
	}

	// Apply overrides
	for _, override := range overrides {
		override(&submodule)
	}

	return submodule
}

// SubmoduleOverride is a function to override submodule defaults
type SubmoduleOverride func(*sqldb.SubmoduleDB)

// WithSubmodulePath sets a custom submodule path
func WithSubmodulePath(path string) SubmoduleOverride {
	return func(s *sqldb.SubmoduleDB) {
		s.Path = path
	}
}

// WithSubmoduleName sets a custom submodule name
func WithSubmoduleName(name string) SubmoduleOverride {
	return func(s *sqldb.SubmoduleDB) {
		s.Name = &name
	}
}

// WithExampleType marks the submodule as an example
func WithExampleType() SubmoduleOverride {
	return func(s *sqldb.SubmoduleDB) {
		s.Type = &[]string{"example"}[0]
	}
}

// CreateExampleFile creates a test example file
func (f *TestDataFactory) CreateExampleFile(submoduleID int, overrides ...ExampleFileOverride) sqldb.ExampleFileDB {
	exampleFile := sqldb.ExampleFileDB{
		SubmoduleID: submoduleID,
		Path:        "main.tf",
		Content:     []byte(`resource "null_resource" "example" {}`),
	}

	// Apply overrides
	for _, override := range overrides {
		override(&exampleFile)
	}

	return exampleFile
}

// ExampleFileOverride is a function to override example file defaults
type ExampleFileOverride func(*sqldb.ExampleFileDB)

// WithFilePath sets a custom file path
func WithFilePath(path string) ExampleFileOverride {
	return func(ef *sqldb.ExampleFileDB) {
		ef.Path = path
	}
}

// WithFileContent sets custom file content
func WithFileContent(content string) ExampleFileOverride {
	return func(ef *sqldb.ExampleFileDB) {
		ef.Content = []byte(content)
	}
}

// Helper methods for generating unique test data

func (f *TestDataFactory) generateNamespaceName() string {
	f.namespaceCounter++
	return "test-namespace-" + string(rune(f.namespaceCounter+'0'))
}

func (f *TestDataFactory) generateDisplayName(base string) string {
	return base + " " + string(rune(f.namespaceCounter+'0'))
}

func (f *TestDataFactory) generateModuleName() string {
	f.moduleCounter++
	return "test-module-" + string(rune(f.moduleCounter+'0'))
}

func (f *TestDataFactory) generateProviderName() string {
	return "test-provider-" + string(rune(f.moduleCounter+'0'))
}

func (f *TestDataFactory) generateVersion() string {
	f.versionCounter++
	return "1." + string(rune(f.versionCounter+'0')) + ".0"
}

func (f *TestDataFactory) generateGroupName() string {
	f.userCounter++
	return "test-group-" + string(rune(f.userCounter+'0'))
}

func (f *TestDataFactory) generateAnalyticsToken() string {
	f.userCounter++
	return "test-token-" + string(rune(f.userCounter+'0'))
}

// Predefined test data sets for common scenarios

// CreateCompleteModuleSet creates a complete set of related test data (namespace, module provider, versions, etc.)
func (f *TestDataFactory) CreateCompleteModuleSet() (sqldb.NamespaceDB, sqldb.ModuleProviderDB, []sqldb.ModuleVersionDB, []sqldb.ModuleDetailsDB) {
	namespace := f.CreateNamespace()
	moduleProvider := f.CreateModuleProvider(namespace.ID)

	versions := []sqldb.ModuleVersionDB{
		f.CreateModuleVersion(moduleProvider.ID, WithVersion("1.0.0"), WithPublished()),
		f.CreateModuleVersion(moduleProvider.ID, WithVersion("1.1.0"), WithPublished()),
		f.CreateModuleVersion(moduleProvider.ID, WithVersion("1.2.0-beta.1"), WithBeta()),
	}

	details := []sqldb.ModuleDetailsDB{
		f.CreateModuleDetails(WithReadmeContent("# Version 1.0.0\n\nInitial release")),
		f.CreateModuleDetails(WithReadmeContent("# Version 1.1.0\n\nBug fixes and improvements")),
		f.CreateModuleDetails(WithReadmeContent("# Version 1.2.0-beta.1\n\nBeta release with new features")),
	}

	// Link versions to details
	for i := range versions {
		versions[i].ModuleDetailsID = &details[i].ID
	}

	return namespace, moduleProvider, versions, details
}

// CreateNamespaceWithPermissions creates a namespace with user group permissions
func (f *TestDataFactory) CreateNamespaceWithPermissions() (sqldb.NamespaceDB, sqldb.UserGroupDB, sqldb.UserGroupNamespacePermissionDB) {
	namespace := f.CreateNamespace()
	userGroup := f.CreateUserGroup(WithGroupName("test-group"))
	// Note: This permission cannot be properly created without database IDs
	// This is a placeholder for testing - in actual tests, save namespace and user group first
	permission := sqldb.UserGroupNamespacePermissionDB{
		UserGroupID:     1, // Placeholder - needs actual DB ID
		NamespaceID:     1, // Placeholder - needs actual DB ID
		PermissionType:  sqldb.PermissionTypeFull,
	}

	return namespace, userGroup, permission
}
