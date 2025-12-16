package service

import (
	"path/filepath"
	"strings"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/storage/model"
)

// PathBuilderService implements PathBuilder interface
// This replicates the Python safe_join_paths and path construction logic exactly
type PathBuilderService struct {
	config *StoragePathConfig
}

// NewPathBuilderService creates a new path builder service
func NewPathBuilderService(config *StoragePathConfig) *PathBuilderService {
	return &PathBuilderService{
		config: config,
	}
}

// BuildNamespacePath builds the path for a namespace
// Python equivalent: safe_join_paths('/modules', namespace_name)
func (p *PathBuilderService) BuildNamespacePath(namespace string) string {
	return p.SafeJoinPaths(p.config.ModulesPath, namespace)
}

// BuildModulePath builds the path for a module
// Python equivalent: safe_join_paths(namespace.base_directory, module_name)
func (p *PathBuilderService) BuildModulePath(namespace string, module string) string {
	namespacePath := p.BuildNamespacePath(namespace)
	return p.SafeJoinPaths(namespacePath, module)
}

// BuildProviderPath builds the path for a provider
// Python equivalent: safe_join_paths(module.base_directory, provider_name)
func (p *PathBuilderService) BuildProviderPath(namespace string, module string, provider string) string {
	modulePath := p.BuildModulePath(namespace, module)
	return p.SafeJoinPaths(modulePath, provider)
}

// BuildVersionPath builds the path for a module version
// Python equivalent: safe_join_paths(provider.base_directory, version)
func (p *PathBuilderService) BuildVersionPath(namespace string, module string, provider string, version string) string {
	providerPath := p.BuildProviderPath(namespace, module, provider)
	return p.SafeJoinPaths(providerPath, version)
}

// BuildArchivePath builds the path for an archive file
// Python equivalent: safe_join_paths(version.base_directory, archive_name)
func (p *PathBuilderService) BuildArchivePath(namespace string, module string, provider string, version string, archiveName string) string {
	versionPath := p.BuildVersionPath(namespace, module, provider, version)
	return p.SafeJoinPaths(versionPath, archiveName)
}

// BuildUploadPath builds the path for uploaded files
// Python equivalent: paths in upload directory
func (p *PathBuilderService) BuildUploadPath(filename string) string {
	return p.SafeJoinPaths(p.config.UploadPath, filename)
}

// SafeJoinPaths safely joins paths to prevent directory traversal
// This replicates the Python safe_join_paths function from utils.py exactly
func (p *PathBuilderService) SafeJoinPaths(basePath string, subPaths ...string) string {
	// Ensure basePath doesn't end with separator (Python behavior)
	basePath = strings.TrimSuffix(basePath, string(filepath.Separator))

	// Build the full path
	result := basePath
	for _, subPath := range subPaths {
		// Ensure sub-path doesn't start with separator (prevents traversal)
		subPath = strings.TrimPrefix(subPath, string(filepath.Separator))
		if subPath == "" {
			continue
		}

		// Add separator if needed
		if !strings.HasSuffix(result, string(filepath.Separator)) {
			result += string(filepath.Separator)
		}
		result += subPath
	}

	return result
}

// ValidatePath validates a path to ensure it's safe
// This replicates the Python path validation logic
func (p *PathBuilderService) ValidatePath(path string) error {
	// Check for path traversal attempts
	if strings.Contains(path, "..") {
		return model.ErrPathTraversal
	}

	// Check for absolute paths in sub-paths
	if p.config != nil {
		if strings.HasPrefix(path, string(filepath.Separator)) && path != p.config.BasePath && !strings.HasPrefix(path, p.config.BasePath) {
			return model.ErrInvalidPath
		}
	}

	// Additional validation can be added here
	return nil
}

// GeneratePath generates a path from components
func (p *PathBuilderService) GeneratePath(pathComponents ...string) string {
	if len(pathComponents) == 0 {
		return ""
	}

	basePath := pathComponents[0]
	subPaths := pathComponents[1:]
	return p.SafeJoinPaths(basePath, subPaths...)
}

// ExtractPathComponents extracts storage path components from a full path
func (p *PathBuilderService) ExtractPathComponents(fullPath string) *model.StoragePath {
	// Remove base path
	relativePath := strings.TrimPrefix(fullPath, p.config.ModulesPath)
	relativePath = strings.TrimPrefix(relativePath, string(filepath.Separator))

	// Split into components
	components := strings.Split(relativePath, string(filepath.Separator))

	storagePath := &model.StoragePath{}

	if len(components) >= 1 {
		storagePath.Namespace = components[0]
	}
	if len(components) >= 2 {
		storagePath.Module = components[1]
	}
	if len(components) >= 3 {
		storagePath.Provider = components[2]
	}
	if len(components) >= 4 {
		storagePath.Version = components[3]
	}

	return storagePath
}

// IsArchivePath checks if a path represents an archive file
func (p *PathBuilderService) IsArchivePath(path string) bool {
	return strings.HasSuffix(path, ".tar.gz") || strings.HasSuffix(path, ".zip")
}

// GetArchiveName extracts archive name from path
func (p *PathBuilderService) GetArchiveName(path string) string {
	if !p.IsArchivePath(path) {
		return ""
	}
	return filepath.Base(path)
}

// GetDefaultPathConfig returns the default path configuration matching Python
func GetDefaultPathConfig(dataDirectory string) *StoragePathConfig {
	// Create a temporary path builder to use SafeJoinPaths
	tempPathBuilder := &PathBuilderService{}

	// Ensure data directory ends with separator
	if !strings.HasSuffix(dataDirectory, string(filepath.Separator)) {
		dataDirectory += string(filepath.Separator)
	}

	return &StoragePathConfig{
		BasePath:      dataDirectory,
		ModulesPath:   tempPathBuilder.SafeJoinPaths(dataDirectory, "modules"),
		ProvidersPath: tempPathBuilder.SafeJoinPaths(dataDirectory, "providers"),
		UploadPath:    tempPathBuilder.SafeJoinPaths(dataDirectory, "upload"),
		ArchivePrefix: "source",
		SourcePrefix:  "source",
	}
}

// BuildModuleArchivePaths builds all possible archive paths for a module version
func (p *PathBuilderService) BuildModuleArchivePaths(namespace string, module string, provider string, version string) []string {
	basePath := p.BuildVersionPath(namespace, module, provider, version)

	return []string{
		p.SafeJoinPaths(basePath, "source.tar.gz"),
		p.SafeJoinPaths(basePath, "source.zip"),
	}
}

// BuildModulePath builds the complete module path structure
func (p *PathBuilderService) BuildModulePathStructure(storagePath *model.StoragePath) map[string]string {
	structure := make(map[string]string)

	if storagePath.Namespace != "" {
		structure["namespace"] = p.BuildNamespacePath(storagePath.Namespace)
	}

	if storagePath.Module != "" {
		structure["module"] = p.BuildModulePath(storagePath.Namespace, storagePath.Module)
	}

	if storagePath.Provider != "" {
		structure["provider"] = p.BuildProviderPath(storagePath.Namespace, storagePath.Module, storagePath.Provider)
	}

	if storagePath.Version != "" {
		structure["version"] = p.BuildVersionPath(storagePath.Namespace, storagePath.Module, storagePath.Provider, storagePath.Version)

		// Add archive paths
		archivePaths := p.BuildModuleArchivePaths(storagePath.Namespace, storagePath.Module, storagePath.Provider, storagePath.Version)
		for i, archivePath := range archivePaths {
			if i == 0 {
				structure["archive_tar_gz"] = archivePath
			} else {
				structure["archive_zip"] = archivePath
			}
		}
	}

	return structure
}