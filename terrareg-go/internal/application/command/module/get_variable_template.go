package module

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/model"
	moduleRepo "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/repository"
	moduleService "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/service"
)

// GetVariableTemplateQuery handles retrieving variable templates for module versions
type GetVariableTemplateQuery struct {
	moduleProviderRepo moduleRepo.ModuleProviderRepository
	moduleFileService  *moduleService.ModuleFileService
}

// NewGetVariableTemplateQuery creates a new get variable template query
func NewGetVariableTemplateQuery(
	moduleProviderRepo moduleRepo.ModuleProviderRepository,
	moduleFileService *moduleService.ModuleFileService,
) *GetVariableTemplateQuery {
	return &GetVariableTemplateQuery{
		moduleProviderRepo: moduleProviderRepo,
		moduleFileService:  moduleFileService,
	}
}

// GetVariableTemplateRequest represents a request to get variable template
type GetVariableTemplateRequest struct {
	Namespace string
	Module    string
	Provider  string
	Version   string
	Output    string // "md" or "html" (default: "md")
}

// VariableTemplate represents a variable definition from the template
type VariableTemplate struct {
	Name           string      `json:"name"`
	Type           string      `json:"type"`           // text, boolean, number, list, select
	AdditionalHelp string      `json:"additional_help"`
	QuoteValue     bool        `json:"quote_value"`
	Required       bool        `json:"required"`
	DefaultValue   interface{} `json:"default_value"`
}

// GetVariableTemplateResponse represents the response with variable template data
type GetVariableTemplateResponse struct {
	Variables []VariableTemplate `json:"variables"`
}

// Execute retrieves the variable template for the module version
func (q *GetVariableTemplateQuery) Execute(ctx context.Context, req *GetVariableTemplateRequest) (*GetVariableTemplateResponse, error) {
	// Validate request
	if req.Namespace == "" || req.Module == "" || req.Provider == "" || req.Version == "" {
		return nil, fmt.Errorf("missing required parameters")
	}

	// Set default output format
	outputFormat := req.Output
	if outputFormat == "" {
		outputFormat = "md"
	}

	// Get module provider
	moduleProvider, err := q.moduleProviderRepo.FindByNamespaceModuleProvider(
		ctx,
		req.Namespace,
		req.Module,
		req.Provider,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get module provider: %w", err)
	}

	// Get module version
	moduleVersion, err := moduleProvider.GetVersion(req.Version)
	if err != nil || moduleVersion == nil {
		return nil, fmt.Errorf("module version not found: %w", err)
	}

	// Get variable template from module version
	variableTemplateJSON, err := q.getVariableTemplateFromVersion(ctx, moduleVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to get variable template: %w", err)
	}

	// Parse and process variables
	variables, err := q.parseVariableTemplate(variableTemplateJSON, outputFormat)
	if err != nil {
		return nil, fmt.Errorf("failed to parse variable template: %w", err)
	}

	return &GetVariableTemplateResponse{
		Variables: variables,
	}, nil
}

// getVariableTemplateFromVersion gets the variable template JSON from module version
func (q *GetVariableTemplateQuery) getVariableTemplateFromVersion(ctx context.Context, moduleVersion *model.ModuleVersion) ([]byte, error) {
	// Get variable template file (typically named variables.json or similar)
	files, err := q.moduleFileService.ListModuleFiles(
		ctx,
		moduleVersion.ModuleProvider().Namespace().Name(),
		moduleVersion.ModuleProvider().Module(),
		moduleVersion.ModuleProvider().Provider(),
		moduleVersion.Version().String(),
	)
	if err != nil {
		return nil, err
	}

	// Look for variable template file
	var templateContent string
	for _, file := range files {
		// Check for common variable template filenames
		if strings.EqualFold(file.Path(), "variables.json") ||
		   strings.EqualFold(file.Path(), "variable_template.json") ||
		   strings.EqualFold(file.Path(), "variables.tf") {
			templateContent = file.Content()
			break
		}
	}

	// If no dedicated template file found, check if variable_template column exists
	if templateContent == "" {
		// TODO: Access the variable_template column from module version
		// For now, return empty template
		return []byte("[]"), nil
	}

	return []byte(templateContent), nil
}

// parseVariableTemplate parses and processes the variable template JSON
func (q *GetVariableTemplateQuery) parseVariableTemplate(templateJSON []byte, outputFormat string) ([]VariableTemplate, error) {
	var variables []VariableTemplate

	if len(templateJSON) == 0 || string(templateJSON) == "[]" {
		return variables, nil
	}

	var rawVariables []map[string]interface{}
	if err := json.Unmarshal(templateJSON, &rawVariables); err != nil {
		return nil, fmt.Errorf("failed to parse JSON template: %w", err)
	}

	for _, rawVar := range rawVariables {
		variable := VariableTemplate{
			Name:           getStringValue(rawVar, "name"),
			Type:           getStringValue(rawVar, "type"),
			AdditionalHelp: getStringValue(rawVar, "additional_help"),
			QuoteValue:     getBoolValue(rawVar, "quote_value", true),
			Required:       getBoolValue(rawVar, "required", true),
			DefaultValue:   rawVar["default_value"],
		}

		// Apply defaults
		if variable.Type == "" {
			variable.Type = "text"
		}

		// Process help text based on output format
		if outputFormat == "html" {
			variable.AdditionalHelp = processMarkdownToHTML(variable.AdditionalHelp)
		}

		variables = append(variables, variable)
	}

	return variables, nil
}

// Helper functions for parsing
func getStringValue(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func getBoolValue(m map[string]interface{}, key string, defaultValue bool) bool {
	if val, ok := m[key]; ok {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return defaultValue
}

// processMarkdownToHTML converts markdown to HTML (basic implementation)
func processMarkdownToHTML(markdown string) string {
	// Basic markdown to HTML conversion
	// In a full implementation, this would use a proper markdown parser
	html := markdown

	// Convert line breaks to <br>
	html = strings.ReplaceAll(html, "\n", "<br>")

	// Convert bold text **text** to <strong>text</strong>
	html = strings.ReplaceAll(html, "**", "<strong>")
	html = strings.ReplaceAll(html, "**", "</strong>")

	return html
}