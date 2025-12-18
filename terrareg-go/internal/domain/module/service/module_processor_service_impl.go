package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/repository"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
)

// ModuleDetailsWithID wraps ModuleDetails with a database ID
type ModuleDetailsWithID struct {
	*model.ModuleDetails
	ID int
}

// ProcessedSubmoduleInfo represents a processed submodule with detailed information
type ProcessedSubmoduleInfo struct {
	Path          string         `json:"path"`
	Source        string         `json:"source"`
	Version       string         `json:"version"`
	Description   string         `json:"description"`
	ReadmeContent string         `json:"readme_content,omitempty"`
	Variables     []VariableInfo `json:"variables,omitempty"`
	Outputs       []OutputInfo   `json:"outputs,omitempty"`
}

// ProcessedExampleFile represents a file within an example directory
type ProcessedExampleFile struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

// ModuleProcessorServiceImpl implements the ModuleProcessorService interface
type ModuleProcessorServiceImpl struct {
	moduleParser       ModuleParser
	moduleDetailsRepo  repository.ModuleDetailsRepository
	moduleVersionRepo  repository.ModuleVersionRepository
	submoduleRepo      repository.SubmoduleRepository
	exampleFileRepo    repository.ExampleFileRepository
	logger            zerolog.Logger
}

// NewModuleProcessorServiceImpl creates a new ModuleProcessorService implementation
func NewModuleProcessorServiceImpl(
	moduleParser ModuleParser,
	moduleDetailsRepo repository.ModuleDetailsRepository,
	moduleVersionRepo repository.ModuleVersionRepository,
	submoduleRepo repository.SubmoduleRepository,
	exampleFileRepo repository.ExampleFileRepository,
	logger zerolog.Logger,
) ModuleProcessorService {
	return &ModuleProcessorServiceImpl{
		moduleParser:       moduleParser,
		moduleDetailsRepo:  moduleDetailsRepo,
		moduleVersionRepo:  moduleVersionRepo,
		submoduleRepo:      submoduleRepo,
		exampleFileRepo:    exampleFileRepo,
		logger:            logger,
	}
}

// ProcessModule processes a module directory and extracts metadata
func (s *ModuleProcessorServiceImpl) ProcessModule(
	ctx context.Context,
	moduleDir string,
	metadata *ModuleProcessingMetadata,
) (*ModuleProcessingResult, error) {
	s.logger.Debug().
		Str("module_dir", moduleDir).
		Int("module_version_id", metadata.ModuleVersionID).
		Msg("Starting module processing")

	// 1. Parse module using existing ModuleParserImpl
	parseResult, err := s.moduleParser.ParseModule(moduleDir)
	if err != nil {
		s.logger.Error().Err(err).Str("module_dir", moduleDir).Msg("Failed to parse module")
		// Continue with partial processing even if parsing fails
		parseResult = &ParseResult{}
	}

	// 2. Extract terraform version
	terraformVersion, err := s.extractTerraformVersion(moduleDir)
	if err != nil {
		s.logger.Warn().Err(err).Msg("Failed to extract terraform version")
		terraformVersion = ""
	}

	// 3. Create ModuleDetails entity with parsed content
	details := model.NewCompleteModuleDetails(
		[]byte(parseResult.ReadmeContent),
		parseResult.RawTerraformDocs,
		nil, // tfsec - to be implemented later
		nil, // infracost - to be implemented later
		s.extractTerraformGraph(moduleDir), // terraform graph
		s.extractTerraformModules(parseResult.TerraformRequirements),
		terraformVersion,
	)

	// 3. Save ModuleDetails to database and get ID
	moduleDetailsID, err := s.moduleDetailsRepo.SaveAndReturnID(ctx, details)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to save module details")
		return nil, fmt.Errorf("failed to save module details: %w", err)
	}

	// 4. Update ModuleVersion with ModuleDetails ID
	s.logger.Info().
		Int("module_version_id", metadata.ModuleVersionID).
		Int("module_details_id", moduleDetailsID).
		Msg("About to call UpdateModuleDetailsID")

	err = s.moduleVersionRepo.UpdateModuleDetailsID(
		ctx,
		metadata.ModuleVersionID,
		moduleDetailsID,
	)
	if err != nil {
		s.logger.Error().Err(err).Int("module_version_id", metadata.ModuleVersionID).Msg("Failed to update module version with module details ID")
		return nil, fmt.Errorf("failed to update module version with module details ID: %w", err)
	}

	s.logger.Info().
		Int("module_details_id", moduleDetailsID).
		Int("module_version_id", metadata.ModuleVersionID).
		Msg("Module details saved and linked to module version")

	// 5. Process submodules
	submodules, err := s.processSubmodules(ctx, moduleDir, metadata.ModuleVersionID)
	if err != nil {
		s.logger.Warn().Err(err).Msg("Failed to process submodules")
		submodules = []SubmoduleInfo{} // Continue with empty submodules
	}

	// 6. Process examples
	examples, err := s.processExamples(ctx, moduleDir, metadata.ModuleVersionID)
	if err != nil {
		s.logger.Warn().Err(err).Msg("Failed to process examples")
		examples = []ExampleInfo{} // Continue with empty examples
	}

	// 7. Convert parse result to module metadata
	moduleMetadata := s.convertToModuleMetadata(parseResult, metadata)

	// 8. Generate variable template (JSON)
	variableTemplate := s.generateVariableTemplate(parseResult.Variables)

	result := &ModuleProcessingResult{
		ModuleMetadata:   moduleMetadata,
		Submodules:       submodules,
		Examples:         examples,
		ReadmeContent:    parseResult.ReadmeContent,
		VariableTemplate: variableTemplate,
		ProcessedFiles:   []string{"README.md", "*.tf"}, // Simplified for now
	}

	s.logger.Info().
		Int("variables_count", len(parseResult.Variables)).
		Int("outputs_count", len(parseResult.Outputs)).
		Int("readme_length", len(parseResult.ReadmeContent)).
		Msg("Module processing completed successfully")

	return result, nil
}

// ValidateModuleStructure validates that a module directory has the required structure
func (s *ModuleProcessorServiceImpl) ValidateModuleStructure(ctx context.Context, moduleDir string) error {
	// Check if module directory exists
	if _, err := os.Stat(moduleDir); os.IsNotExist(err) {
		return fmt.Errorf("module directory does not exist: %s", moduleDir)
	}

	// Look for at least one .tf file
	hasTerraformFiles := false
	err := filepath.Walk(moduleDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".tf") {
			hasTerraformFiles = true
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("error scanning module directory: %w", err)
	}

	if !hasTerraformFiles {
		return fmt.Errorf("no Terraform files (.tf) found in module directory: %s", moduleDir)
	}

	return nil
}

// ExtractMetadata extracts metadata from a module directory
func (s *ModuleProcessorServiceImpl) ExtractMetadata(ctx context.Context, moduleDir string) (*ModuleMetadata, error) {
	parseResult, err := s.moduleParser.ParseModule(moduleDir)
	if err != nil {
		return nil, fmt.Errorf("failed to parse module: %w", err)
	}

	// Create basic metadata without Git information
	metadata := &ModuleMetadata{
		Name:        filepath.Base(moduleDir),
		Description: parseResult.Description,
		Version:     "", // Will be filled by caller
		Providers:   s.convertProviderVersions(parseResult.ProviderVersions),
		Variables:   s.convertVariables(parseResult.Variables),
		Outputs:     s.convertOutputs(parseResult.Outputs),
		Resources:   s.convertResources(parseResult.Resources),
		Dependencies: s.convertDependencies(parseResult.Dependencies),
	}

	return metadata, nil
}

// convertToModuleMetadata converts ParseResult to ModuleMetadata
func (s *ModuleProcessorServiceImpl) convertToModuleMetadata(parseResult *ParseResult, metadata *ModuleProcessingMetadata) *ModuleMetadata {
	return &ModuleMetadata{
		Name:        filepath.Base(metadata.GitPath),
		Description: parseResult.Description,
		Version:     metadata.GitTag,
		Providers:   s.convertProviderVersions(parseResult.ProviderVersions),
		Variables:   s.convertVariables(parseResult.Variables),
		Outputs:     s.convertOutputs(parseResult.Outputs),
		Resources:   s.convertResources(parseResult.Resources),
		Dependencies: s.convertDependencies(parseResult.Dependencies),
	}
}

// convertVariables converts service.Variable to VariableInfo
func (s *ModuleProcessorServiceImpl) convertVariables(variables []Variable) []VariableInfo {
	var result []VariableInfo
	for _, v := range variables {
		result = append(result, VariableInfo{
			Name:        v.Name,
			Type:        v.Type,
			Description: v.Description,
			Default:     v.Default,
			Required:    v.Required,
		})
	}
	return result
}

// convertOutputs converts service.Output to OutputInfo
func (s *ModuleProcessorServiceImpl) convertOutputs(outputs []Output) []OutputInfo {
	var result []OutputInfo
	for _, o := range outputs {
		result = append(result, OutputInfo{
			Name:        o.Name,
			Description: o.Description,
			Value:       nil, // Outputs don't have values during extraction
			Sensitive:   false, // service.Output doesn't have Sensitive field, default to false
		})
	}
	return result
}

// convertProviderVersions converts service.ProviderVersion to ProviderInfo
func (s *ModuleProcessorServiceImpl) convertProviderVersions(providerVersions []ProviderVersion) []ProviderInfo {
	var result []ProviderInfo
	for _, pv := range providerVersions {
		result = append(result, ProviderInfo{
			Name:    pv.Name,
			Version: pv.Version,
			Source:  "", // Will be extracted from terraform docs in future enhancement
		})
	}
	return result
}

// convertResources converts service.Resource to ResourceInfo
func (s *ModuleProcessorServiceImpl) convertResources(resources []Resource) []ResourceInfo {
	var result []ResourceInfo
	for _, r := range resources {
		result = append(result, ResourceInfo{
			Type: r.Type,
			Name: r.Name,
		})
	}
	return result
}

// convertDependencies converts model.Dependency to DependencyInfo
func (s *ModuleProcessorServiceImpl) convertDependencies(dependencies []model.Dependency) []DependencyInfo {
	var result []DependencyInfo
	for _, d := range dependencies {
		result = append(result, DependencyInfo{
			Source:  d.Source,
			Version: d.Version,
		})
	}
	return result
}

// generateVariableTemplate creates a JSON template for variables
func (s *ModuleProcessorServiceImpl) generateVariableTemplate(variables []Variable) string {
	if len(variables) == 0 {
		return "{}"
	}

	template := make(map[string]interface{})
	for _, v := range variables {
		variableData := map[string]interface{}{
			"description": v.Description,
			"type":        v.Type,
			"required":    v.Required,
		}
		if v.Default != nil {
			variableData["default"] = v.Default
		}
		template[v.Name] = variableData
	}

	jsonBytes, err := json.MarshalIndent(template, "", "  ")
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to generate variable template JSON")
		return "{}"
	}

	return string(jsonBytes)
}
// extractTerraformVersion extracts the terraform version for a module directory
func (s *ModuleProcessorServiceImpl) extractTerraformVersion(moduleDir string) (string, error) {
	// Run terraform version command
	cmd := exec.Command("terraform", "version")
	cmd.Dir = moduleDir
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get terraform version: %w", err)
	}

	// Parse terraform version output
	// Expected format: "Terraform v1.14.3 on linux_arm64"
	outputStr := string(output)
	lines := strings.Split(outputStr, "\n")
	if len(lines) > 0 {
		versionLine := strings.TrimSpace(lines[0])
		if strings.HasPrefix(versionLine, "Terraform v") {
			version := strings.TrimPrefix(versionLine, "Terraform v")
			// Remove platform information if present
			if spaceIdx := strings.Index(version, " "); spaceIdx > 0 {
				version = version[:spaceIdx]
			}
			return version, nil
		}
	}

	return "", fmt.Errorf("unable to parse terraform version from output: %s", outputStr)
}

// extractTerraformGraph generates terraform graph for the module
func (s *ModuleProcessorServiceImpl) extractTerraformGraph(moduleDir string) []byte {
	cmd := exec.Command("terraform", "graph")
	cmd.Dir = moduleDir
	output, err := cmd.Output()
	if err != nil {
		s.logger.Warn().Err(err).Str("module_dir", moduleDir).Msg("Failed to generate terraform graph")
		return nil
	}
	return output
}

// extractTerraformModules creates a JSON representation of terraform requirements
func (s *ModuleProcessorServiceImpl) extractTerraformModules(requirements []TerraformRequirement) []byte {
	if len(requirements) == 0 {
		return nil
	}

	jsonBytes, err := json.Marshal(requirements)
	if err != nil {
		s.logger.Warn().Err(err).Msg("Failed to marshal terraform requirements")
		return nil
	}
	return jsonBytes
}

// processSubmodules detects and processes submodules in the module directory
func (s *ModuleProcessorServiceImpl) processSubmodules(ctx context.Context, moduleDir string, moduleVersionID int) ([]SubmoduleInfo, error) {
	// Detect submodules
	submoduleNames, err := s.moduleParser.DetectSubmodules(moduleDir)
	if err != nil {
		return nil, fmt.Errorf("failed to detect submodules: %w", err)
	}

	s.logger.Info().
		Str("module_dir", moduleDir).
		Int("detected_count", len(submoduleNames)).
		Msg("Detected submodules")

	var submodules []SubmoduleInfo

	for _, submoduleName := range submoduleNames {
		submodulePath := filepath.Join(moduleDir, submoduleName)

		// Parse submodule content
		submoduleParseResult, err := s.moduleParser.ParseModule(submodulePath)
		if err != nil {
			s.logger.Warn().Err(err).
				Str("submodule", submoduleName).
				Msg("Failed to parse submodule, skipping")
			continue
		}

		// Create submodule ModuleDetails
		submoduleDetails := model.NewCompleteModuleDetails(
			[]byte(submoduleParseResult.ReadmeContent),
			submoduleParseResult.RawTerraformDocs,
			nil, // tfsec - to be implemented later
			nil, // infracost - to be implemented later
			nil, // terraform graph - optional for submodules
			nil, // terraform modules - optional for submodules
			"",  // terraform version - optional for submodules
		)

		// Save submodule details
		submoduleDetailsID, err := s.moduleDetailsRepo.SaveAndReturnID(ctx, submoduleDetails)
		if err != nil {
			s.logger.Warn().Err(err).
				Str("submodule", submoduleName).
				Msg("Failed to save submodule details, skipping")
			continue
		}

		// Save submodule record to database
		submoduleDB := &sqldb.SubmoduleDB{
			ModuleDetailsID: &submoduleDetailsID,
			Path:            submoduleName,
			Name:            &submoduleName,
		}

		_, err = s.submoduleRepo.SaveWithDetails(ctx, moduleVersionID, submoduleDB, submoduleDetailsID)
		if err != nil {
			s.logger.Warn().Err(err).
				Str("submodule", submoduleName).
				Msg("Failed to save submodule record, skipping")
			continue
		}

		// Create SubmoduleInfo for response
		submodule := SubmoduleInfo{
			Path:    submoduleName,
			Source:  "", // TODO: Add source logic if needed
			Version: "", // TODO: Add version logic if needed
		}

		submodules = append(submodules, submodule)

		s.logger.Debug().
			Str("submodule", submoduleName).
			Int("details_id", submoduleDetailsID).
			Msg("Processed submodule")
	}

	return submodules, nil
}

// processExamples detects and processes examples in the module directory
func (s *ModuleProcessorServiceImpl) processExamples(ctx context.Context, moduleDir string, moduleVersionID int) ([]ExampleInfo, error) {
	// Detect examples
	exampleNames, err := s.moduleParser.DetectExamples(moduleDir)
	if err != nil {
		return nil, fmt.Errorf("failed to detect examples: %w", err)
	}

	s.logger.Info().
		Str("module_dir", moduleDir).
		Int("detected_count", len(exampleNames)).
		Msg("Detected examples")

	var examples []ExampleInfo

	for _, exampleName := range exampleNames {
		examplePath := filepath.Join(moduleDir, "examples", exampleName)

		// Extract all files from example directory
		exampleFiles, err := s.extractExampleFiles(examplePath)
		if err != nil {
			s.logger.Warn().Err(err).
				Str("example", exampleName).
				Msg("Failed to extract example files, skipping")
			continue
		}

		// Create example record in database
		exampleType := "example"
		exampleSubmodule := &sqldb.SubmoduleDB{
			Type: &exampleType,
			Path: filepath.Join("examples", exampleName),
			Name: &exampleName,
		}

		// Save example submodule record
		savedSubmodule, err := s.submoduleRepo.Save(ctx, moduleVersionID, exampleSubmodule)
		if err != nil {
			s.logger.Warn().Err(err).
				Str("example", exampleName).
				Msg("Failed to save example record, skipping")
			continue
		}

		// Save example files
		var exampleFileDBs []*sqldb.ExampleFileDB
		for _, exampleFile := range exampleFiles {
			exampleFileDBs = append(exampleFileDBs, &sqldb.ExampleFileDB{
				SubmoduleID: savedSubmodule.ID,
				Path:        exampleFile.Path,
				Content:     []byte(exampleFile.Content), // Convert string to []byte
			})
		}

		if len(exampleFileDBs) > 0 {
			_, err = s.exampleFileRepo.SaveBatch(ctx, exampleFileDBs)
			if err != nil {
				s.logger.Warn().Err(err).
					Str("example", exampleName).
					Msg("Failed to save example files, skipping")
				continue
			}
		}

		// Create example info for response
		example := ExampleInfo{
			Name:        exampleName,
			Description: fmt.Sprintf("Example: %s", exampleName),
			Files:       make([]string, len(exampleFiles)),
		}

		for i, file := range exampleFiles {
			example.Files[i] = file.Path
		}

		examples = append(examples, example)

		s.logger.Debug().
			Str("example", exampleName).
			Int("file_count", len(exampleFiles)).
			Msg("Processed example")
	}

	return examples, nil
}

// extractExampleFiles extracts all files from an example directory
func (s *ModuleProcessorServiceImpl) extractExampleFiles(exampleDir string) ([]ProcessedExampleFile, error) {
	var files []ProcessedExampleFile

	err := filepath.Walk(exampleDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Skip hidden files and temporary files
		if strings.HasPrefix(info.Name(), ".") || strings.HasSuffix(info.Name(), "~") {
			return nil
		}

		// Read file content
		contentBytes, err := os.ReadFile(path)
		if err != nil {
			s.logger.Warn().Err(err).
				Str("file_path", path).
				Msg("Failed to read example file, skipping")
			return nil
		}

		// Get relative path from example directory
		relPath, err := filepath.Rel(exampleDir, path)
		if err != nil {
			relPath = filepath.Base(path)
		}

		files = append(files, ProcessedExampleFile{
			Path:    relPath,
			Content: string(contentBytes), // Convert []byte to string
		})

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk example directory %s: %w", exampleDir, err)
	}

	return files, nil
}
