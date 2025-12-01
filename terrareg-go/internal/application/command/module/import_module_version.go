package module

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/terrareg/terrareg/internal/config"
	"github.com/terrareg/terrareg/internal/domain/module/repository"
	"github.com/terrareg/terrareg/internal/domain/module/service"
)

// ImportModuleVersionCommand handles importing module versions from Git
type ImportModuleVersionCommand struct {
	moduleProviderRepo repository.ModuleProviderRepository
	config             *config.Config
}

// NewImportModuleVersionCommand creates a new command
func NewImportModuleVersionCommand(
	moduleProviderRepo repository.ModuleProviderRepository,
	config *config.Config,
) *ImportModuleVersionCommand {
	return &ImportModuleVersionCommand{
		moduleProviderRepo: moduleProviderRepo,
		config:             config,
	}
}

// ImportModuleVersionRequest represents the import request
type ImportModuleVersionRequest struct {
	Namespace string
	Module    string
	Provider  string
	Version   *string // Optional - derived from git tag if not provided
	GitTag    *string // Optional - conflicts with Version
}

// Execute imports a module version from Git
func (c *ImportModuleVersionCommand) Execute(ctx context.Context, req ImportModuleVersionRequest) error {
	// Validate that either version or git_tag is provided (not both, not neither)
	if (req.Version == nil && req.GitTag == nil) || (req.Version != nil && req.GitTag != nil) {
		return fmt.Errorf("either version or git_tag must be provided (but not both)")
	}

	// Find the module provider
	moduleProvider, err := c.moduleProviderRepo.FindByNamespaceModuleProvider(
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
	if req.GitTag != nil {
		if gitTagFormat := moduleProvider.GitTagFormat(); gitTagFormat != nil && *gitTagFormat != "" {
			re, err := regexp.Compile(*gitTagFormat)
			if err != nil {
				return fmt.Errorf("invalid git_tag_format regex: %w", err)
			}
			matches := re.FindStringSubmatch(*req.GitTag)
			if len(matches) > 1 {
				req.Version = &matches[1]
			} else {
				return fmt.Errorf("git_tag '%s' does not match git_tag_format '%s'", *req.GitTag, *gitTagFormat)
			}
		} else {
			req.Version = req.GitTag
		}
	}

	// Clone the Git repository
	var cloneURLTemplate string
	if tmpl := moduleProvider.RepoCloneURLTemplate(); tmpl != nil && *tmpl != "" {
		cloneURLTemplate = *tmpl
	} else if gp := moduleProvider.GitProvider(); gp != nil && gp.CloneURLTemplate != "" {
		cloneURLTemplate = gp.CloneURLTemplate
	} else {
		return fmt.Errorf("no clone URL template configured for module provider")
	}

	replacer := strings.NewReplacer(
		"{protocol}", "https",
		"{namespace}", req.Namespace,
		"{name}", req.Module,
		"{provider}", req.Provider,
	)
	cloneURL := replacer.Replace(cloneURLTemplate)

	tmpDir, err := os.MkdirTemp("", "terrareg-git-")
	if err != nil {
		return fmt.Errorf("failed to create temp dir for git clone: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	cmd := exec.Command("git", "clone", cloneURL, tmpDir)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to clone git repository: %w", err)
	}

	if req.GitTag != nil {
		cmd = exec.Command("git", "-C", tmpDir, "checkout", *req.GitTag)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to checkout git tag '%s': %w", *req.GitTag, err)
		}
	}

	// Extract module files
	srcDir := tmpDir
	if gitPath := moduleProvider.GitPath(); gitPath != nil && *gitPath != "" {
		srcDir = filepath.Join(tmpDir, *gitPath)
	}

	destDir := filepath.Join(c.config.DataDirectory, "modules", req.Namespace, req.Module, req.Provider, *req.Version)
	if err := copyDir(srcDir, destDir); err != nil {
		return fmt.Errorf("failed to copy module files: %w", err)
	}

	// Run terraform-docs to extract metadata
	parser := service.NewModuleParser()
	parseResult, _ := parser.ParseModule(destDir)

	// Create/update module version
	var details *model.ModuleDetails
	if parseResult != nil {
		details = model.NewModuleDetails([]byte(parseResult.ReadmeContent))
		if parseResult.RawTerraformDocs != nil {
			details = details.WithTerraformDocs(parseResult.RawTerraformDocs)
		}
	}

	version, err := moduleProvider.GetVersion(*req.Version)
	if err != nil {
		// Not found, create new
		version, err = moduleProvider.PublishVersion(*req.Version, details, false) // assuming not beta
		if err != nil {
			return err
		}
	} else {
		// Found, update
		version.SetDetails(details)
	}

	if parseResult != nil && parseResult.Description != "" {
		version.SetMetadata(nil, &parseResult.Description)
	}

	// Publish the version
	if err := version.Publish(); err != nil {
		return fmt.Errorf("failed to publish version: %w", err)
	}

	if err := c.moduleProviderRepo.Save(ctx, moduleProvider); err != nil {
		return fmt.Errorf("failed to save module provider: %w", err)
	}

	return nil
}

func copyDir(src, dest string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dest, info.Mode()); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		destPath := filepath.Join(dest, entry.Name())

		if entry.IsDir() {
			if err := copyDir(srcPath, destPath); err != nil {
				return err
			}
		} else {
			// Prevent symlink traversal
			if entry.Type()&os.ModeSymlink != 0 {
				continue
			}
			if err := copyFile(srcPath, destPath); err != nil {
				return err
			}
		}
	}
	return nil
}

func copyFile(src, dest string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Close()
}
