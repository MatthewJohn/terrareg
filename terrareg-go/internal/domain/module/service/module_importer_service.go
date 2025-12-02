package service

import (
	"context"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/config"
	gitService "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/git/service" // Added import with alias
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/repository"
)

// ModuleImporterService handles the business logic for importing module versions.
type ModuleImporterService struct {
	moduleProviderRepo repository.ModuleProviderRepository
	gitClient          gitService.GitClient // Changed to use alias
	storageService     StorageService
	moduleParser       ModuleParser
	config             *config.Config
}

// NewModuleImporterService creates a new ModuleImporterService.
func NewModuleImporterService(
	moduleProviderRepo repository.ModuleProviderRepository,
	gitClient gitService.GitClient, // Changed to use alias
	storageService StorageService,
	moduleParser ModuleParser,
	config *config.Config,
) *ModuleImporterService {
	return &ModuleImporterService{
		moduleProviderRepo: moduleProviderRepo,
		gitClient:          gitClient,
		storageService:     storageService,
		moduleParser:       moduleParser,
		config:             config,
	}
}

// ImportModuleVersionRequest represents the data needed to import a module version.
type ImportModuleVersionRequest struct {
	Namespace string
	Module    string
	Provider  string
	Version   *string
	GitTag    *string
}

// ImportModuleVersion imports a module version from Git.
func (s *ModuleImporterService) ImportModuleVersion(ctx context.Context, req ImportModuleVersionRequest) error {
	// Validate that either version or git_tag is provided (not both, not neither)
	if (req.Version == nil && req.GitTag == nil) || (req.Version != nil && req.GitTag != nil) {
		return fmt.Errorf("either version or git_tag must be provided (but not both)")
	}

	// Find the module provider
	moduleProvider, err := s.moduleProviderRepo.FindByNamespaceModuleProvider(
		ctx, req.Namespace, req.Module, req.Provider,
	)
	if err != nil {
		return fmt.Errorf("module provider not found: %w", err)
	}

	// Validate that the module provider has Git configuration
	if moduleProvider.GitProviderID() == nil || moduleProvider.RepoCloneURLTemplate() == nil || *moduleProvider.RepoCloneURLTemplate() == "" {
		return fmt.Errorf("module provider is not a git based module")
	}

	// If git_tag is provided, derive version from it
	resolvedVersion := req.Version
	if req.GitTag != nil {
		if gitTagFormat := moduleProvider.GitTagFormat(); gitTagFormat != nil && *gitTagFormat != "" {
			re, err := regexp.Compile(*gitTagFormat)
			if err != nil {
				return fmt.Errorf("invalid git_tag_format regex: %w", err)
			}
			matches := re.FindStringSubmatch(*req.GitTag)
			if len(matches) > 1 {
				resolvedVersion = &matches[1]
			} else {
				return fmt.Errorf("git_tag '%s' does not match git_tag_format '%s'", *req.GitTag, *gitTagFormat)
			}
		} else {
			resolvedVersion = req.GitTag
		}
	}

	if resolvedVersion == nil || *resolvedVersion == "" {
		return fmt.Errorf("could not determine module version")
	}

	// Clone the Git repository
	var cloneURLTemplate string
	if tmpl := moduleProvider.RepoCloneURLTemplate(); tmpl != nil && *tmpl != "" {
		cloneURLTemplate = *tmpl
	} else if gp := moduleProvider.GitProvider(); gp != nil && gp.CloneURLTemplate != "" { // This assumes GitProvider is loaded with the moduleProvider. Needs review when GitProvider is implemented
		cloneURLTemplate = gp.CloneURLTemplate
	} else {
		return fmt.Errorf("no clone URL template configured for module provider")
	}

	replacer := strings.NewReplacer(
		"{protocol}", "https", // Assuming HTTPS for cloning
		"{namespace}", req.Namespace,
		"{name}", req.Module,
		"{provider}", req.Provider,
	)
	cloneURL := replacer.Replace(cloneURLTemplate)

	tmpDir, err := s.storageService.MkdirTemp("", "terrareg-git-")
	if err != nil {
		return fmt.Errorf("failed to create temp dir for git clone: %w", err)
	}
	defer func() {
		if rErr := s.storageService.RemoveAll(tmpDir); rErr != nil {
			fmt.Printf("Error removing temp dir %s: %v\n", tmpDir, rErr)
		}
	}()

	if err := s.gitClient.Clone(ctx, cloneURL, tmpDir); err != nil {
		return fmt.Errorf("failed to clone git repository: %w", err)
	}

	if req.GitTag != nil {
		if err := s.gitClient.Checkout(ctx, tmpDir, *req.GitTag); err != nil {
			return fmt.Errorf("failed to checkout git tag '%s': %w", *req.GitTag, err)
		}
	}

	// Determine source directory within the cloned repo
	srcDir := tmpDir
	if gitPath := moduleProvider.GitPath(); gitPath != nil && *gitPath != "" {
		srcDir = filepath.Join(tmpDir, *gitPath)
	}

	// Destination directory for module files
	destDir := filepath.Join(s.config.DataDirectory, "modules", req.Namespace, req.Module, req.Provider, *resolvedVersion)
	if err := s.storageService.MkdirAll(destDir, 0755); err != nil { // Ensure destination directory exists
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	if err := s.storageService.CopyDir(srcDir, destDir); err != nil {
		return fmt.Errorf("failed to copy module files: %w", err)
	}

	// Run terraform-docs to extract metadata
	parseResult, err := s.moduleParser.ParseModule(destDir)
	if err != nil {
		return fmt.Errorf("failed to parse module: %w", err)
	}

	// Create/update module version
	var details *model.ModuleDetails
	if parseResult != nil {
		details = model.NewModuleDetails([]byte(parseResult.ReadmeContent)) // Updated call
		if parseResult.RawTerraformDocs != nil {
			details = details.WithTerraformDocs(parseResult.RawTerraformDocs)
		}
		// TODO: Add other details like variables, outputs, resources, provider versions
	}

	// Get or create ModuleVersion
	// This part assumes that ModuleProvider has methods to manage its versions
	// and that the repository can update the ModuleProvider aggregate.
	moduleVersion, err := moduleProvider.GetVersion(*resolvedVersion)
	if err != nil {
		// Version not found, create new
		moduleVersion, err = moduleProvider.PublishVersion(*resolvedVersion, details, false) // Assuming not beta for initial import
		if err != nil {
			return fmt.Errorf("failed to publish new version: %w", err)
		}
	} else {
		// Version found, update its details
		// NOTE: In a real DDD scenario, updating an existing version might imply a new "version" of the ModuleVersion entity
		// or specific business rules for updating existing published versions. For now, we'll just set details.
		moduleVersion.SetDetails(details)
		if parseResult != nil {
			// Set owner and description directly on ModuleVersion
			moduleVersion.SetMetadata(&parseResult.Owner, &parseResult.Description)
		}
	}

	// Ensure the version is marked as published (if not already)
	if !moduleVersion.IsPublished() {
		if err := moduleVersion.Publish(); err != nil {
			return fmt.Errorf("failed to mark version as published: %w", err)
		}
	}

	// Save the updated module provider aggregate
	if err := s.moduleProviderRepo.Save(ctx, moduleProvider); err != nil {
		return fmt.Errorf("failed to save module provider: %w", err)
	}

	return nil
}
