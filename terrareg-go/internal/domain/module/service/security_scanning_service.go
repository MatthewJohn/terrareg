package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	moduleRepo "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/repository"
)

// SecurityScanningService handles tfsec security scanning of module versions
type SecurityScanningService struct {
	moduleFileService *ModuleFileService
	moduleVersionRepo moduleRepo.ModuleVersionRepository
}

// NewSecurityScanningService creates a new security scanning service
func NewSecurityScanningService(
	moduleFileService *ModuleFileService,
	moduleVersionRepo moduleRepo.ModuleVersionRepository,
) *SecurityScanningService {
	return &SecurityScanningService{
		moduleFileService: moduleFileService,
		moduleVersionRepo: moduleVersionRepo,
	}
}

// SecurityScanRequest represents a request to scan a module version
type SecurityScanRequest struct {
	Namespace  string
	Module     string
	Provider   string
	Version    string
	ModulePath string
}

// SecurityScanResult represents a tfsec scan result
type SecurityScanResult struct {
	RuleID      string                  `json:"rule_id"`
	Severity    string                  `json:"severity"`
	Title       string                  `json:"title"`
	Description string                  `json:"description"`
	Location    SecurityScanLocation    `json:"location"`
	Links       []string                `json:"links,omitempty"`
}

// SecurityScanLocation represents the location of a security issue
type SecurityScanLocation struct {
	Filename  string `json:"filename"`
	StartLine int    `json:"start_line"`
	EndLine   int    `json:"end_line"`
}

// SecurityScanResponse represents the response from a security scan
type SecurityScanResponse struct {
	Results []SecurityScanResult `json:"results"`
	Summary SecurityScanSummary  `json:"summary"`
}

// SecurityScanSummary provides a summary of security scan results
type SecurityScanSummary struct {
	Total    int `json:"total"`
	High     int `json:"high"`
	Medium   int `json:"medium"`
	Low      int `json:"low"`
	Info     int `json:"info"`
	Warnings int `json:"warnings"`
}

// ExecuteSecurityScan runs a tfsec security scan on a module version
func (s *SecurityScanningService) ExecuteSecurityScan(ctx context.Context, req *SecurityScanRequest) (*SecurityScanResponse, error) {
	// TODO: Add security scanning configuration check when available
	// For now, always allow security scanning

	// If module path is not provided, extract files temporarily
	scanPath := req.ModulePath
	if scanPath == "" {
		tempDir, err := s.extractModuleFiles(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("failed to extract module files: %w", err)
		}
		defer os.RemoveAll(tempDir)
		scanPath = tempDir
	}

	// Run tfsec command
	results, err := s.runTfsecScan(scanPath)
	if err != nil {
		return nil, fmt.Errorf("failed to run tfsec scan: %w", err)
	}

	// Process and clean results
	processedResults := s.processResults(results, scanPath)

	// Generate summary
	summary := s.generateSummary(processedResults)

	return &SecurityScanResponse{
		Results: processedResults,
		Summary: summary,
	}, nil
}

// GetSecurityResults retrieves stored security scan results for a module version
func (s *SecurityScanningService) GetSecurityResults(ctx context.Context, moduleVersionID int) (*SecurityScanResponse, error) {
	// TODO: Implement when security scan data is added to module version domain model
	// For now, return empty results
	return &SecurityScanResponse{
		Results: []SecurityScanResult{},
		Summary: SecurityScanSummary{},
	}, nil
}

// StoreSecurityResults stores security scan results for a module version
func (s *SecurityScanningService) StoreSecurityResults(ctx context.Context, moduleVersionID int, results *SecurityScanResponse) error {
	// TODO: Implement when security scan data is added to module version domain model
	// For now, just log that we would store results
	fmt.Printf("TODO: Store security scan results for module version %d\n", moduleVersionID)
	return nil
}

// extractModuleFiles extracts module files to a temporary directory for scanning
func (s *SecurityScanningService) extractModuleFiles(ctx context.Context, req *SecurityScanRequest) (string, error) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", fmt.Sprintf("terrareg-scan-%s-%s-%s-%s", req.Namespace, req.Module, req.Provider, req.Version))
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Get module files
	files, err := s.moduleFileService.ListModuleFiles(ctx, req.Namespace, req.Module, req.Provider, req.Version)
	if err != nil {
		os.RemoveAll(tempDir)
		return "", fmt.Errorf("failed to get module files: %w", err)
	}

	// Write files to temporary directory
	for _, file := range files {
		filePath := filepath.Join(tempDir, file.Path())
		dirPath := filepath.Dir(filePath)

		// Create directory if it doesn't exist
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			os.RemoveAll(tempDir)
			return "", fmt.Errorf("failed to create directory: %w", err)
		}

		// Write file content
		if err := os.WriteFile(filePath, []byte(file.Content()), 0644); err != nil {
			os.RemoveAll(tempDir)
			return "", fmt.Errorf("failed to write file: %w", err)
		}
	}

	return tempDir, nil
}

// runTfsecScan executes the tfsec command and returns raw results
func (s *SecurityScanningService) runTfsecScan(modulePath string) (map[string]interface{}, error) {
	// Build tfsec command with options matching Python implementation
	args := []string{
		"--ignore-hcl-errors",
		"--format", "json",
		"--no-module-downloads",
		"--soft-fail",
		"--no-colour",
		"--include-ignored",
		"--include-passed",
		"--disable-grouping",
		modulePath,
	}

	// Execute tfsec command
	cmd := exec.Command("tfsec", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// tfsec returns non-zero exit code when security issues are found
		// This is expected behavior, so we continue with the output
		// Only fail if the command itself couldn't execute
		if !cmd.ProcessState.Success() && len(output) == 0 {
			return nil, fmt.Errorf("tfsec execution failed: %w", err)
		}
	}

	// Parse JSON output
	var results map[string]interface{}
	if err := json.Unmarshal(output, &results); err != nil {
		return nil, fmt.Errorf("failed to parse tfsec JSON output: %w", err)
	}

	return results, nil
}

// processResults cleans and processes tfsec results
func (s *SecurityScanningService) processResults(results map[string]interface{}, modulePath string) []SecurityScanResult {
	var processedResults []SecurityScanResult

	// Get results array from tfsec output
	if resultsArray, ok := results["results"].([]interface{}); ok {
		for _, result := range resultsArray {
			if resultMap, ok := result.(map[string]interface{}); ok {
				processedResult := s.processSingleResult(resultMap, modulePath)
				if processedResult != nil {
					processedResults = append(processedResults, *processedResult)
				}
			}
		}
	}

	return processedResults
}

// processSingleResult processes a single tfsec result
func (s *SecurityScanningService) processSingleResult(result map[string]interface{}, modulePath string) *SecurityScanResult {
	// Extract rule ID
	ruleID, _ := result["rule_id"].(string)
	if ruleID == "" {
		return nil
	}

	// Extract severity
	severity, _ := result["severity"].(string)

	// Extract title
	title, _ := result["title"].(string)

	// Extract description
	description, _ := result["description"].(string)

	// Extract location
	var location SecurityScanLocation
	if locationMap, ok := result["location"].(map[string]interface{}); ok {
		if filename, ok := locationMap["filename"].(string); ok {
			// Clean up path - remove temporary directory prefix
			location.Filename = strings.TrimPrefix(filename, modulePath)
			location.Filename = strings.TrimPrefix(location.Filename, "/")
		}
		if startLine, ok := locationMap["start_line"].(float64); ok {
			location.StartLine = int(startLine)
		}
		if endLine, ok := locationMap["end_line"].(float64); ok {
			location.EndLine = int(endLine)
		}
	}

	// Extract links
	var links []string
	if linksArray, ok := result["links"].([]interface{}); ok {
		for _, link := range linksArray {
			if linkStr, ok := link.(string); ok {
				links = append(links, linkStr)
			}
		}
	}

	return &SecurityScanResult{
		RuleID:      ruleID,
		Severity:    severity,
		Title:       title,
		Description: description,
		Location:    location,
		Links:       links,
	}
}

// generateSummary creates a summary of security scan results
func (s *SecurityScanningService) generateSummary(results []SecurityScanResult) SecurityScanSummary {
	summary := SecurityScanSummary{
		Total: len(results),
	}

	for _, result := range results {
		switch strings.ToLower(result.Severity) {
		case "high", "critical":
			summary.High++
		case "medium":
			summary.Medium++
		case "low":
			summary.Low++
		case "info":
			summary.Info++
		case "warning":
			summary.Warnings++
		}
	}

	return summary
}

// GetSecurityFailures returns only the failure results (excluding info and warnings)
func (s *SecurityScanningService) GetSecurityFailures(ctx context.Context, moduleVersionID int) ([]SecurityScanResult, error) {
	results, err := s.GetSecurityResults(ctx, moduleVersionID)
	if err != nil {
		return nil, err
	}

	var failures []SecurityScanResult
	for _, result := range results.Results {
		// Only include failures (exclude info and warnings)
		severity := strings.ToLower(result.Severity)
		if severity != "info" && severity != "warning" {
			failures = append(failures, result)
		}
	}

	return failures, nil
}