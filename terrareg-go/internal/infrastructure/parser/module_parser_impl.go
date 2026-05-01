package parser

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	configModel "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/config/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/service"
	sharedService "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared/service"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/logging"
)

// ModuleParserImpl implements the service.ModuleParser interface.
type ModuleParserImpl struct {
	storageService service.StorageService
	config         *configModel.DomainConfig
	commandService sharedService.SystemCommandService
	logger         logging.Logger
}

// NewModuleParserImpl creates a new ModuleParserImpl.
// Python: No direct equivalent (constructor pattern)
func NewModuleParserImpl(
	storageService service.StorageService,
	config *configModel.DomainConfig,
	commandService sharedService.SystemCommandService,
	logger logging.Logger,
) *ModuleParserImpl {
	return &ModuleParserImpl{
		storageService: storageService,
		config:         config,
		commandService: commandService,
		logger:         logger,
	}
}

// ParseModule parses a module directory and extracts metadata.
// Python: ModuleExtractor._run_terraform_docs() + _get_readme_content() + _extract_description() + _get_terrareg_metadata()
func (p *ModuleParserImpl) ParseModule(modulePath string) (*service.ParseResult, error) {
	result := &service.ParseResult{}

	// Check if module directory exists
	_, err := p.storageService.Stat(modulePath)
	if err != nil {
		return nil, fmt.Errorf("module directory does not exist or is inaccessible: %s, error: %w", modulePath, err)
	}

	// Parse terrareg metadata files first (terrareg.json/.terrareg.json)
	// Python: ModuleExtractor._get_terrareg_metadata()
	// This must be done before README extraction as metadata description takes precedence
	metadata, err := p.parseTerraregMetadata(modulePath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse terrareg metadata: %w", err)
	}
	if metadata != nil {
		// Populate result with metadata fields
		// Python: Uses metadata.owner, metadata.description, etc.
		if metadata.Owner != nil {
			result.Owner = *metadata.Owner
		}
		// Note: Other metadata fields like RepoCloneURL are handled by higher-level services
	}

	// Extract README content
	// Python: ModuleExtractor._get_readme_content()
	readmeContent, err := p.extractReadme(modulePath)
	if err == nil {
		result.ReadmeContent = readmeContent
		// Use description from metadata if available, otherwise extract from README
		// Python: description = terrareg_metadata.get('description', None); if not description: description = _extract_description(readme_content)
		if metadata != nil && metadata.Description != nil && *metadata.Description != "" {
			result.Description = *metadata.Description
		} else {
			result.Description = p.extractDescriptionFromReadme(readmeContent)
		}
	}

	// Remove terraform-docs config files before running to prevent module's own config from being used
	// Python: ModuleExtractor._run_terraform_docs() lines 111-114
	for _, configFile := range []string{".terraform-docs.yml", ".terraform-docs.yaml"} {
		configPath := filepath.Join(modulePath, configFile)
		if _, err := p.storageService.Stat(configPath); err == nil {
			// Remove the file so terraform-docs doesn't use module's config
			if err := p.storageService.RemoveAll(configPath); err != nil {
				p.logger.Warn().Err(err).Str("path", configPath).Msg("Failed to remove terraform-docs config")
			}
		}
	}

	// Run terraform-docs to extract variables, outputs, etc.
	// Python: ModuleExtractor._run_terraform_docs()
	cmd := &sharedService.Command{
		Name: "terraform-docs",
		Args: []string{"json", modulePath},
	}
	cmdResult, err := p.commandService.Execute(context.Background(), cmd)
	if err == nil && cmdResult.ExitCode == 0 {
		output := []byte(cmdResult.Stdout)
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
			// Python: Gets from terraform-docs output (same as Go) - Already implemented
			for _, req := range tfdocs.Requirements {
				result.TerraformRequirements = append(result.TerraformRequirements, service.TerraformRequirement{
					Name:    req.Name,
					Version: req.Version,
				})
			}
			// Use terraform-docs header as fallback description if no description set yet
			// Python: if tfdocs.Header != "" and result.Description == "" { result.Description = tfdocs.Header }
			if tfdocs.Header != "" && result.Description == "" {
				result.Description = tfdocs.Header
			}
		}
	} else {
		// Log the error but don't fail, terraform-docs might not be installed or module might be invalid
		// Python: Raises UnableToProcessTerraformError in production, logs in debug mode
		p.logger.Warn().
			Err(err).
			Str("module_path", modulePath).
			Int("exit_code", cmdResult.ExitCode).
			Str("stderr", cmdResult.Stderr).
			Msg("terraform-docs failed, continuing without terraform-docs output")
	}

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
		Name    string  `json:"name"`
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

// parseTerraregMetadata reads and validates terrareg.json/.terrareg.json files.
// Returns the parsed metadata or nil if no metadata file exists (not an error).
// Python: ModuleExtractor._get_terrareg_metadata()
func (p *ModuleParserImpl) parseTerraregMetadata(modulePath string) (*service.TerraregMetadata, error) {
	// Try terrareg.json first, then .terrareg.json
	// Python: self.TERRAREG_METADATA_FILES = ['terrareg.json', '.terrareg.json']
	metadataFiles := []string{"terrareg.json", ".terrareg.json"}

	var metadataPath string
	for _, filename := range metadataFiles {
		path := filepath.Join(modulePath, filename)
		if _, err := p.storageService.Stat(path); err == nil {
			metadataPath = path
			break
		}
	}

	// No metadata file found - this is not an error
	if metadataPath == "" {
		return nil, nil
	}

	// Read and parse metadata file
	content, err := p.storageService.ReadFile(metadataPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata file %s: %w", metadataPath, err)
	}

	var metadata service.TerraregMetadata
	if err := json.Unmarshal(content, &metadata); err != nil {
		return nil, fmt.Errorf("failed to parse metadata JSON from %s: %w", metadataPath, err)
	}

	// Delete the metadata file after reading to prevent it from being included in archives
	// Python: os.unlink(path) after reading
	if err := p.storageService.RemoveAll(metadataPath); err != nil {
		// Log the error but don't fail - the metadata was successfully read
		p.logger.Warn().Err(err).Str("path", metadataPath).Msg("Failed to delete metadata file after reading")
	}

	return &metadata, nil
}

// ParseSubmodule parses a single submodule directory and extracts metadata.
// Returns a ParseResult with the submodule's terraform-docs output and README.
// Python: ModuleExtractor._process_submodule() (partial - terraform-docs and README only)
func (p *ModuleParserImpl) ParseSubmodule(submodulePath string) (*service.ParseResult, error) {
	// Reuse ParseModule functionality for submodules
	// Python: _process_submodule calls _run_terraform_docs and _get_readme_content
	return p.ParseModule(submodulePath)
}

// ParseExample parses a single example directory and extracts metadata.
// Returns an ExampleParseResult with the example's parse results and optional infracost data.
// Note: Infracost analysis should be performed separately using the InfracostService.
// Python: ModuleExtractor._process_submodule() for Example (including _run_infracost)
func (p *ModuleParserImpl) ParseExample(examplePath string, infracostJSON []byte) (*service.ExampleParseResult, error) {
	// Parse the example directory
	parseResult, err := p.ParseModule(examplePath)
	if err != nil {
		return nil, err
	}

	// Create example parse result with optional infracost data
	return &service.ExampleParseResult{
		ParseResult:   parseResult,
		InfracostJSON: infracostJSON,
	}, nil
}

// ExtractExampleFiles extracts all terraform-related files from an example directory.
// Returns a list of ExampleFile pointers with file paths and contents.
// Uses the ExampleFileExtensions from config to determine which files to extract.
// Python: ModuleExtractor._extract_example_files()
func (p *ModuleParserImpl) ExtractExampleFiles(examplePath string) ([]*model.ExampleFile, error) {
	// Get file extensions from config
	// Python: Config().EXAMPLE_FILE_EXTENSIONS defaults to ["tf", "tfvars", "sh", "json"]
	extensions := p.config.ExampleFileExtensions
	if len(extensions) == 0 {
		// Use default if not configured
		extensions = []string{".tf", ".tfvars", ".sh", ".json"}
	}

	var exampleFiles []*model.ExampleFile

	// Read all entries in the example directory
	entries, err := p.storageService.ReadDir(examplePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read example directory: %w", err)
	}

	for _, entry := range entries {
		// Skip directories
		if entry.IsDir() {
			continue
		}

		filename := entry.Name()

		// Check if file has one of the desired extensions
		// Config extensions don't include the dot, so we need to add it
		hasMatchingExtension := false
		for _, ext := range extensions {
			if !strings.HasPrefix(ext, ".") {
				ext = "." + ext
			}
			if strings.HasSuffix(filename, ext) {
				hasMatchingExtension = true
				break
			}
		}

		if !hasMatchingExtension {
			continue
		}

		// Read file contents
		filePath := filepath.Join(examplePath, filename)
		content, err := p.storageService.ReadFile(filePath)
		if err != nil {
			// Log warning but continue with other files
			p.logger.Warn().Err(err).Str("path", filePath).Msg("Failed to read example file")
			continue
		}

		// Create example file using the constructor
		// Python: ExampleFile.create(example=example, path=tf_file)
		exampleFile := model.NewExampleFile(filename, content)
		exampleFiles = append(exampleFiles, exampleFile)
	}

	return exampleFiles, nil
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

// extractDescriptionFromReadme extracts a description from README content.
// Python: ModuleExtractor._extract_description()
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

		// Check minimum character count (20 alphabetic characters)
		// Python: len(re.sub(r'[^a-zA-Z]', '', line)) < 20
		alphaCount := 0
		for _, r := range line {
			if unicode.IsLetter(r) {
				alphaCount++
			}
		}
		if alphaCount < 20 {
			continue
		}

		// Check minimum word count (6 words)
		// Python: word_match is None or len(word_match) < 6
		wordCount := 0
		inWord := false
		for _, r := range line {
			if unicode.IsLetter(r) || unicode.IsDigit(r) {
				if !inWord {
					wordCount++
					inWord = true
				}
			} else {
				inWord = false
			}
		}
		if wordCount < 6 {
			continue
		}

		// Extract description from sentences, limiting length
		description := ""
		sentences := strings.Split(line, ". ")

		for _, sentence := range sentences {
			sentence = strings.TrimSpace(sentence)
			if sentence == "" {
				continue
			}
			newDescription := description
			if description != "" {
				newDescription += ". "
			}
			newDescription += sentence

			// Check length limits
			// Python: 80 chars for combined, 130 for first sentence
			// If description is empty (first sentence) and it would be too long, return it
			// If description is not empty and adding this sentence would make it too long, return the previous description
			if description == "" {
				// First sentence - allow up to 130 chars
				if len(newDescription) >= 130 {
					return newDescription
				}
			} else {
				// Not first sentence - allow up to 80 chars combined
				if len(newDescription) >= 80 {
					// Return previous description (without this sentence)
					return description
				}
			}
			description = newDescription
		}

		if description != "" {
			return description
		}
	}

	return ""
}

// DetectSubmodules finds submodules in the module directory.
// Returns a list of submodule paths relative to the module root.
// Python: ModuleExtractor._scan_submodules() for MODULES_DIRECTORY (scanning only, not processing)
func (p *ModuleParserImpl) DetectSubmodules(modulePath string) ([]string, error) {
	var submodules []string
	submoduleSet := make(map[string]bool)

	// Get the configured modules directory
	modulesDir := p.config.ModulesDirectory
	if modulesDir == "" {
		modulesDir = "modules" // fallback to default
	}

	// Determine the scan path based on MODULES_DIRECTORY configuration
	var scanPath string
	if modulesDir == "" {
		// If MODULES_DIRECTORY is empty, scan the root module directory
		scanPath = modulePath
	} else {
		// Otherwise, only scan within the configured modules directory
		scanPath = filepath.Join(modulePath, modulesDir)

		// Check if the configured directory exists
		_, err := p.storageService.Stat(scanPath)
		if err != nil {
			// Modules directory doesn't exist, return empty list (Python behavior)
			return submodules, nil
		}
	}

	// Get examples directory name for exclusion
	examplesDir := p.config.ExamplesDirectory
	if examplesDir == "" {
		examplesDir = "examples" // fallback to default
	}

	// Walk through the scan path looking for .tf files recursively
	err := filepath.Walk(scanPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip if not a .tf file
		if !strings.HasSuffix(path, ".tf") || info.IsDir() {
			return nil
		}

		// Get the parent directory of the .tf file
		parentDir := filepath.Dir(path)

		// Skip if parent directory is the scan path (i.e., root of scan)
		// Python: "WARNING: submodule is in root of submodules directory."
		if parentDir == scanPath {
			return nil
		}

		// Get relative path from module root (not scan path)
		relativePath, err := filepath.Rel(modulePath, parentDir)
		if err != nil {
			return nil
		}

		// Skip hidden directories
		if strings.HasPrefix(relativePath, ".") {
			return nil
		}

		// Skip if parent directory is examples directory
		// Check if the first path component is the examples directory
		firstComponent := strings.Split(relativePath, string(filepath.Separator))[0]
		if firstComponent == examplesDir {
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

// DetectExamples finds example directories in the module.
// Returns a list of example directory names (not paths) relative to the examples directory.
// Python: ModuleExtractor._scan_submodules() for EXAMPLES_DIRECTORY (scanning only, not processing)
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
