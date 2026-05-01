package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider"
	providermodel "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider/model"
)

// ProviderDocumentationService handles extracting provider documentation using tfplugindocs
// Python reference: provider_extractor.py::ProviderExtractor.extract_documentation
type ProviderDocumentationService struct {
	// Dependencies would be injected here
}

// NewProviderDocumentationService creates a new documentation service
func NewProviderDocumentationService() *ProviderDocumentationService {
	return &ProviderDocumentationService{}
}

// DocumentationResult represents extracted documentation with metadata
type DocumentationResult struct {
	Type        providermodel.DocumentationType
	Name        string
	Content     string
	Title       *string
	Subcategory *string
	Description *string
}

// ExtractDocumentation extracts documentation from provider source using tfplugindocs
// Python reference: provider_extractor.py::ProviderExtractor.extract_documentation
func (s *ProviderDocumentationService) ExtractDocumentation(
	ctx context.Context,
	sourceDir string,
	providerEntity *provider.Provider,
) ([]*DocumentationResult, error) {
	docs := make([]*DocumentationResult, 0)

	// Check if tfplugindocs is available
	tfplugindocsPath, err := s.findTfplugindocs()
	if err != nil {
		// tfplugindocs is optional - return empty docs if not found
		return docs, nil
	}

	// Check for pre-existing documentation
	preexistingDocs, err := s.extractPreexistingDocumentation(sourceDir)
	if err == nil && len(preexistingDocs) > 0 {
		return preexistingDocs, nil
	}

	// Run tfplugindocs generate
	generatedDocs, err := s.runTfplugindocsGenerate(ctx, sourceDir, tfplugindocsPath, providerEntity.Name())
	if err != nil {
		// tfplugindocs failures are not fatal - just log and continue
		fmt.Printf("Warning: tfplugindocs generation failed: %v\n", err)
		return docs, nil
	}

	docs = append(docs, generatedDocs...)

	return docs, nil
}

// findTfplugindocs finds the tfplugindocs executable
func (s *ProviderDocumentationService) findTfplugindocs() (string, error) {
	// Check common locations
	paths := []string{
		"tfplugindocs",
		"/usr/local/bin/tfplugindocs",
		"/usr/bin/tfplugindocs",
	}

	for _, path := range paths {
		if _, err := exec.LookPath(path); err == nil {
			return path, nil
		}
	}

	return "", fmt.Errorf("tfplugindocs not found in PATH")
}

// extractPreexistingDocumentation extracts pre-existing documentation from the source directory
// Python reference: test_provider_extractor_documentation.py::TestProviderExtractorDocumentation::test_preexisting_docs
func (s *ProviderDocumentationService) extractPreexistingDocumentation(sourceDir string) ([]*DocumentationResult, error) {
	docs := make([]*DocumentationResult, 0)

	// Check for docs directory
	docsDir := filepath.Join(sourceDir, "docs")
	if _, err := os.Stat(docsDir); os.IsNotExist(err) {
		return docs, nil
	}

	// Walk through docs directory
	err := filepath.Walk(docsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// Check if file is markdown
		if strings.HasSuffix(strings.ToLower(info.Name()), ".md") {
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			// Extract metadata from frontmatter
			title, subcategory, description := s.parseMarkdownFrontmatter(content)

			// Determine documentation type from path
			docType := s.determineDocumentationTypeFromPath(path, docsDir)

			// Get relative name
			relPath, err := filepath.Rel(docsDir, path)
			if err != nil {
				relPath = info.Name()
			}

			// Remove .md extension
			name := strings.TrimSuffix(relPath, ".md")

			docs = append(docs, &DocumentationResult{
				Type:        docType,
				Name:        name,
				Content:     string(content),
				Title:       title,
				Subcategory: subcategory,
				Description: description,
			})
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return docs, nil
}

// parseMarkdownFrontmatter parses YAML frontmatter from markdown content
// Python reference: provider_extractor.py (frontmatter extraction)
func (s *ProviderDocumentationService) parseMarkdownFrontmatter(content []byte) (*string, *string, *string) {
	contentStr := string(content)

	// Check for frontmatter delimiter
	if !strings.HasPrefix(contentStr, "---") {
		return nil, nil, nil
	}

	// Find end delimiter
	endIndex := strings.Index(contentStr[4:], "---")
	if endIndex == -1 {
		return nil, nil, nil
	}

	frontmatter := contentStr[4 : endIndex+4]

	var title, subcategory, description *string

	// Parse key-value pairs
	lines := strings.Split(frontmatter, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(strings.ToLower(parts[0]))
		value := strings.TrimSpace(parts[1])

		// Remove quotes if present
		value = strings.Trim(value, `"'`)

		switch key {
		case "page_title", "title":
			title = &value
		case "subcategory":
			subcategory = &value
		case "description":
			description = &value
		}
	}

	return title, subcategory, description
}

// determineDocumentationTypeFromPath determines the documentation type from the file path
// Python reference: provider_extractor.py::ProviderExtractor._get_doc_type
func (s *ProviderDocumentationService) determineDocumentationTypeFromPath(filePath, docsDir string) providermodel.DocumentationType {
	relPath, err := filepath.Rel(docsDir, filePath)
	if err != nil {
		return providermodel.DocumentationTypeGuide
	}

	lowerPath := strings.ToLower(relPath)

	// Check for overview
	if strings.Contains(lowerPath, "index") || strings.Contains(lowerPath, "overview") {
		return providermodel.DocumentationTypeOverview
	}

	// Check for resources
	if strings.Contains(lowerPath, "resources") || strings.Contains(lowerPath, "r/") {
		return providermodel.DocumentationTypeResource
	}

	// Check for data sources
	if strings.Contains(lowerPath, "data-sources") || strings.Contains(lowerPath, "d/") {
		return providermodel.DocumentationTypeDataSource
	}

	// Default to guide
	return providermodel.DocumentationTypeGuide
}

// runTfplugindocsGenerate runs tfplugindocs generate command
// Python reference: provider_extractor.py::ProviderExtractor.extract_documentation
func (s *ProviderDocumentationService) runTfplugindocsGenerate(
	ctx context.Context,
	sourceDir string,
	tfplugindocsPath string,
	providerName string,
) ([]*DocumentationResult, error) {
	docs := make([]*DocumentationResult, 0)

	// Create output directory
	outputDir := filepath.Join(sourceDir, "docs")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create docs directory: %w", err)
	}

	// Run tfplugindocs generate
	cmd := exec.CommandContext(ctx, tfplugindocsPath, "generate")
	cmd.Dir = sourceDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("tfplugindocs generate failed: %s\noutput: %s", err, string(output))
	}

	// Parse generated documentation
	generatedDocs, err := s.extractPreexistingDocumentation(sourceDir)
	if err != nil {
		return nil, fmt.Errorf("failed to extract generated documentation: %w", err)
	}

	docs = append(docs, generatedDocs...)

	return docs, nil
}

// TfplugindocsMetadata represents the tfplugindocs metadata file
type TfplugindocsMetadata struct {
	ProviderName string `json:"provider_name"`
	Version      string `json:"version"`
	Renderer     struct {
		// MarkDownOptions allows customization of markdown rendering
		ToC bool `json:"toc"`
	} `json:"render"`
}

// ParseTfplugindocsMetadata parses the tfplugindocs metadata file
func (s *ProviderDocumentationService) ParseTfplugindocsMetadata(sourceDir string) (*TfplugindocsMetadata, error) {
	metadataPath := filepath.Join(sourceDir, "tfplugindocs-metadata.json")

	content, err := os.ReadFile(metadataPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata file: %w", err)
	}

	var metadata TfplugindocsMetadata
	if err := json.Unmarshal(content, &metadata); err != nil {
		return nil, fmt.Errorf("failed to parse metadata file: %w", err)
	}

	return &metadata, nil
}

// ValidateDocumentationStructure validates that the documentation structure is correct
func (s *ProviderDocumentationService) ValidateDocumentationStructure(sourceDir string) error {
	// Check for at least one markdown file in docs directory
	docsDir := filepath.Join(sourceDir, "docs")
	if _, err := os.Stat(docsDir); os.IsNotExist(err) {
		// No docs directory is OK - tfplugindocs will create it
		return nil
	}

	foundDocs := false
	err := filepath.Walk(docsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(strings.ToLower(info.Name()), ".md") {
			foundDocs = true
		}

		return nil
	})

	if err != nil {
		return err
	}

	if !foundDocs {
		return fmt.Errorf("docs directory exists but contains no markdown files")
	}

	return nil
}
