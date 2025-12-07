package util

import (
	"time"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/model"
)

// CreateMockNamespace creates a mock namespace for testing
func CreateMockNamespace(name string) *model.Namespace {
	namespace, _ := model.NewNamespace(name, nil, "NONE")
	return namespace
}

// CreateMockNamespaceWithDetails creates a mock namespace with display name and type
func CreateMockNamespaceWithDetails(name string, displayName *string, nsType string) *model.Namespace {
	namespace, _ := model.NewNamespace(name, displayName, model.NamespaceType(nsType))
	return namespace
}

// CreateMockModuleProvider creates a mock module provider for testing
func CreateMockModuleProvider(namespace, moduleName, provider string, verified bool) *model.ModuleProvider {
	ns := CreateMockNamespace(namespace)
	mp, _ := model.NewModuleProvider(ns, moduleName, provider)

	if verified {
		mp.Verify()
	}

	return mp
}

// CreateMockModuleVersion creates a mock module version for testing
func CreateMockModuleVersion(version string, owner, description *string, publishedAt *time.Time) *model.ModuleVersion {
	details := model.NewModuleDetails(nil)

	// Use ReconstructModuleVersion to create with owner and description
	mv, _ := model.ReconstructModuleVersion(
		1, // id
		version,
		details,
		false, // beta
		false, // internal
		publishedAt != nil, // published
		publishedAt,
		nil, // gitSHA
		nil, // gitPath
		false, // archiveGitPath
		nil, // repoBaseURLTemplate
		nil, // repoCloneURLTemplate
		nil, // repoBrowseURLTemplate
		owner,
		description,
		nil, // variableTemplate
		nil, // extractionVersion
		time.Now(), // createdAt
		time.Now(), // updatedAt
	)

	return mv
}

// CreateMockModuleVersionWithBeta creates a mock beta module version
func CreateMockModuleVersionWithBeta(version string, beta bool, internal bool) *model.ModuleVersion {
	details := model.NewModuleDetails(nil)

	mv, _ := model.ReconstructModuleVersion(
		1, // id
		version,
		details,
		beta,
		internal,
		false, // published
		nil,   // publishedAt
		nil,   // gitSHA
		nil,   // gitPath
		false, // archiveGitPath
		nil,   // repoBaseURLTemplate
		nil,   // repoCloneURLTemplate
		nil,   // repoBrowseURLTemplate
		nil,   // owner
		nil,   // description
		nil,   // variableTemplate
		nil,   // extractionVersion
		time.Now(), // createdAt
		time.Now(), // updatedAt
	)

	return mv
}

// BuildTestPaginationMeta creates a pagination meta for testing
func BuildTestPaginationMeta(limit, offset, total int) map[string]interface{} {
	return map[string]interface{}{
		"limit":      limit,
		"offset":     offset,
		"total_count": total,
	}
}