package parser

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	configModel "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/config/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/service"
)

// ModuleParserImpl implements the service.ModuleParser interface.
type ModuleParserImpl struct {
	storageService service.StorageService
	config         *configModel.DomainConfig
}

// NewModuleParserImpl creates a new ModuleParserImpl.
func NewModuleParserImpl(storageService service.StorageService, config *configModel.DomainConfig) *ModuleParserImpl {
	return &ModuleParserImpl{
		storageService: storageService,
		config:         config,
	}
}

// ParseModule parses a module directory and extracts metadata.
func (p *ModuleParserImpl) ParseModule(modulePath string) (*service.ParseResult, error) {
	result := &service.ParseResult{}

	// Check if module directory exists
	_, err := p.storageService.Stat(modulePath)
	if err != nil {
		return nil, fmt.Errorf("module directory does not exist or is inaccessible: %s, error: %w", modulePath, err)
	}

	// Extract README content
	readmeContent, err := p.extractReadme(modulePath)
	if err == nil {
		result.ReadmeContent = readmeContent
		result.Description = p.extractDescriptionFromReadme(readmeContent)
	}

	// Run terraform-docs to extract variables, outputs, etc.
	cmd := exec.Command("terraform-docs", "json", modulePath)
	output, err := cmd.Output()
	if err == nil {
		result.RawTerraformDocs = output
		var tfdocs TerraformDocs
		if err := json.Unmarshal(output, &tfdocs); err == nil {
			for _, input := range tfdocs.Inputs {
				result.Variables = append(result.Variables, service.Variable{
					Name:        input.Name,
					Type:        input.Type,
					Description: input.Description,
					Default:     input.Default,
					Required:    input.Required,
				})
			}
			for _, out := range tfdocs.Outputs {
				result.Outputs = append(result.Outputs, service.Output{
					Name:        out.Name,
					Description: out.Description,
				})
			}
			for _, provider := range tfdocs.Providers {
				version := ""
				if provider.Version != nil {
					version = *provider.Version
				}
				result.ProviderVersions = append(result.ProviderVersions, service.ProviderVersion{
					Name:    provider.Name,
					Version: version,
				})
			}
			for _, resource := range tfdocs.Resources {
				result.Resources = append(result.Resources, service.Resource{
					Type: resource.Type,
					Name: resource.Name,
				})
			}
			// Extract module dependencies
			for _, module := range tfdocs.Modules {
				result.Dependencies = append(result.Dependencies, model.Dependency{
					Module:  module.Name,
					Source:  module.Source,
					Version: module.Version,
				})
			}
			// Extract terraform requirements
			for _, req := range tfdocs.Requirements {
				result.TerraformRequirements = append(result.TerraformRequirements, service.TerraformRequirement{
					Name:    req.Name,
					Version: req.Version,
				})
			}
			if tfdocs.Header != "" && result.Description == "" {
				result.Description = tfdocs.Header
			}
		}
	} else {
		// Log the error but don't fail, terraform-docs might not be installed or module might be invalid
		fmt.Printf("warning: terraform-docs failed for %s: %v\n", modulePath, err)
		if exitError, ok := err.(*exec.ExitError); ok {
			fmt.Printf("stderr: %s\n", exitError.Stderr)
		}
	}

	// TODO: Parse terrareg metadata files
	// TODO: Extract provider requirements
	// TODO: Detect submodules
	// TODO: Detect examples

	return result, nil
}

// TerraformDocs represents the structure of the terraform-docs JSON output
type TerraformDocs struct {
	Header string `json:"header"`
	Footer string `json:"footer"`
	Inputs []struct {
		Name        string      `json:"name"`
		Type        string      `json:"type"`
		Description string      `json:"description"`
		Default     interface{} `json:"default"`
		Required    bool        `json:"required"`
	} `json:"inputs"`
	Outputs []struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	} `json:"outputs"`
	Providers []struct {
		Name    string `json:"name"`
		Alias   *string `json:"alias"`
		Version *string `json:"version"`
	} `json:"providers"`
	Resources []struct {
		Type        string  `json:"type"`
		Name        string  `json:"name"`
		Provider    string  `json:"provider"`
		Source      string  `json:"source"`
		Mode        string  `json:"mode"`
		Version     string  `json:"version"`
		Description *string `json:"description"`
	} `json:"resources"`
	Modules []struct {
		Name    string `json:"name"`
		Source  string `json:"source"`
		Version string `json:"version"`
	} `json:"modules"`
	Requirements []struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	} `json:"requirements"`
}

// extractReadme reads the README.md file from the module directory
func (p *ModuleParserImpl) extractReadme(modulePath string) (string, error) {
	readmePath := filepath.Join(modulePath, "README.md")

	content, err := p.storageService.ReadFile(readmePath)
	if err != nil {
		return "", fmt.Errorf("failed to read README: %w", err)
	}

	return string(content), nil
}

// extractDescriptionFromReadme extracts a description from README content
func (p *ModuleParserImpl) extractDescriptionFromReadme(readmeContent string) string {
	if readmeContent == "" {
		return ""
	}

	lines := strings.Split(readmeContent, "\n")

	for _, line := range lines {
		// Skip empty lines and headers
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Skip lines with unwanted content
		if strings.Contains(line, "http://") ||
			strings.Contains(line, "https://") ||
			strings.Contains(line, "@") {
			continue
		}

		// Extract description from sentences, limiting length
		description := ""
		sentences := strings.Split(line, ". ")

		for _, sentence := range sentences {
			sentence = strings.TrimSpace(sentence)
			newDescription := description
			if description != "" {
				newDescription += ". "
			}
			newDescription += sentence

			// Limit description length (matching Python logic)
			if (newDescription != "" && len(newDescription) >= 80) ||
				(description == "" && len(newDescription) >= 130) {
				break
			}
			description = newDescription
		}

		if description != "" {
			return description
		}
	}

	return ""
}

// DetectSubmodules finds submodules in the module directory
// Matches Python implementation using recursive scanning for .tf files
func (p *ModuleParserImpl) DetectSubmodules(modulePath string) ([]string, error) {
	var submodules []string
	submoduleSet := make(map[string]bool)

	// Walk through the directory looking for .tf files recursively
	err := filepath.Walk(modulePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip if not a .tf file
		if !strings.HasSuffix(path, ".tf") || info.IsDir() {
			return nil
		}

		// Get the parent directory of the .tf file
		parentDir := filepath.Dir(path)

		// Skip if parent directory is the root module path (Python warns about this)
		if parentDir == modulePath {
			return nil
		}

		// Skip if parent directory is examples directory
		examplesDir := p.config.ExamplesDirectory
		if examplesDir == "" {
			examplesDir = "examples" // fallback to default
		}
		if filepath.Base(filepath.Dir(parentDir)) == examplesDir {
			return nil
		}

		// Get relative path from module root
		relativePath, err := filepath.Rel(modulePath, parentDir)
		if err != nil {
			return nil
		}

		// Skip hidden directories
		if strings.HasPrefix(relativePath, ".") {
			return nil
		}

		// Add to set to avoid duplicates
		if !submoduleSet[relativePath] {
			submoduleSet[relativePath] = true
			submodules = append(submodules, relativePath)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to scan module directory: %w", err)
	}

	return submodules, nil
}

// DetectExamples finds example directories in the module
// Matches Python implementation: scans examples/ directory non-recursively for subdirectories with .tf files
func (p *ModuleParserImpl) DetectExamples(modulePath string) ([]string, error) {
	var examples []string

	examplesDir := p.config.ExamplesDirectory
	if examplesDir == "" {
		examplesDir = "examples" // fallback to default
	}
	examplesPath := filepath.Join(modulePath, examplesDir)

	// Check if examples directory exists
	_, err := p.storageService.Stat(examplesPath)
	if err != nil {
		// No examples directory, return empty list
		return examples, nil
	}

	entries, err := p.storageService.ReadDir(examplesPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read examples directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Skip hidden directories
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		examplePath := filepath.Join(examplesPath, entry.Name())

		// Check if this directory contains .tf files
		hasTerraformFiles, err := p.hasTerraformFiles(examplePath)
		if err != nil {
			continue
		}

		if hasTerraformFiles {
			examples = append(examples, entry.Name())
		}
	}

	return examples, nil
}

// hasTerraformFiles checks if a directory contains any .tf files
func (p *ModuleParserImpl) hasTerraformFiles(dirPath string) (bool, error) {
	entries, err := p.storageService.ReadDir(dirPath)
	if err != nil {
		return false, err
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".tf") {
			return true, nil
		}
	}

	return false, nil
}
