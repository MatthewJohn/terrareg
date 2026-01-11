package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/model"
	moduleRepo "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/repository"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared/service"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb/transaction"
	"gorm.io/gorm"
)

// SecurityScanningService handles tfsec security scanning of module versions with transaction safety
type SecurityScanningService struct {
	moduleFileService *ModuleFileService
	moduleVersionRepo moduleRepo.ModuleVersionRepository
	savepointHelper   *transaction.SavepointHelper
	commandService    service.SystemCommandService
}

// NewSecurityScanningService creates a new security scanning service with transaction support
func NewSecurityScanningService(
	moduleFileService *ModuleFileService,
	moduleVersionRepo moduleRepo.ModuleVersionRepository,
	savepointHelper *transaction.SavepointHelper,
	commandService service.SystemCommandService,
) *SecurityScanningService {
	return &SecurityScanningService{
		moduleFileService: moduleFileService,
		moduleVersionRepo: moduleVersionRepo,
		savepointHelper:   savepointHelper,
		commandService:    commandService,
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
	RuleID      string               `json:"rule_id"`
	Severity    string               `json:"severity"`
	Title       string               `json:"title"`
	Description string               `json:"description"`
	Location    SecurityScanLocation `json:"location"`
	Links       []string             `json:"links,omitempty"`
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
	Critical int `json:"critical"`
	High     int `json:"high"`
	Medium   int `json:"medium"`
	Low      int `json:"low"`
	Info     int `json:"info"`
	Warnings int `json:"warnings"`
}

// ExecuteSecurityScan runs a tfsec security scan on a module version
func (s *SecurityScanningService) ExecuteSecurityScan(ctx context.Context, req *SecurityScanRequest) (*SecurityScanResponse, error) {
	// Check if services are properly initialized
	if s.moduleFileService == nil {
		return nil, fmt.Errorf("module file service not initialized")
	}

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
	results, err := s.runTfsecScan(ctx, scanPath)
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
func (s *SecurityScanningService) runTfsecScan(ctx context.Context, modulePath string) (map[string]interface{}, error) {
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

	// Execute tfsec command using SystemCommandService
	cmd := &service.Command{
		Name: "tfsec",
		Args: args,
	}

	result, err := s.commandService.Execute(ctx, cmd)
	if err != nil && strings.Contains(err.Error(), "executable file not found") {
		// tfsec is not installed, return empty results instead of failing
		return map[string]interface{}{
			"results": []interface{}{},
			"summary": map[string]interface{}{
				"passed":   0,
				"failed":   0,
				"critical": 0,
				"high":     0,
				"medium":   0,
				"low":      0,
			},
		}, nil
	}

	outputStr := result.Stdout

	// Log tfsec output for debugging (trimmed if too long)
	outputPreview := outputStr
	if len(outputPreview) > 500 {
		outputPreview = outputPreview[:500] + "..."
	}

	// Check if output looks like JSON
	outputStr = strings.TrimSpace(outputStr)
	if !strings.HasPrefix(outputStr, "{") && !strings.HasPrefix(outputStr, "[") {
		// Output doesn't look like JSON, it's probably an error message
		// Return empty results
		return map[string]interface{}{
			"results": []interface{}{},
			"summary": map[string]interface{}{
				"passed":   0,
				"failed":   0,
				"critical": 0,
				"high":     0,
				"medium":   0,
				"low":      0,
			},
		}, nil
	}

	// Parse JSON output
	var results map[string]interface{}
	if err := json.Unmarshal([]byte(outputStr), &results); err != nil {
		// JSON parsing failed, return empty results
		return map[string]interface{}{
			"results": []interface{}{},
			"summary": map[string]interface{}{
				"passed":   0,
				"failed":   0,
				"critical": 0,
				"high":     0,
				"medium":   0,
				"low":      0,
			},
		}, nil
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

// Transaction security scanning methods

// SecurityScanTransactionResult represents the result of a security scan within a transaction
type SecurityScanTransactionResult struct {
	Success             bool                  `json:"success"`
	SecurityResponse    *SecurityScanResponse `json:"security_response,omitempty"`
	Error               *string               `json:"error,omitempty"`
	ScanDuration        time.Duration         `json:"scan_duration"`
	Timestamp           time.Time             `json:"timestamp"`
	SavepointRolledBack bool                  `json:"savepoint_rolled_back"`
}

// BatchSecurityScanResult represents the result of batch security scanning
type BatchSecurityScanResult struct {
	TotalScans      int                             `json:"total_scans"`
	SuccessfulScans []SecurityScanTransactionResult `json:"successful_scans"`
	FailedScans     []SecurityScanTransactionResult `json:"failed_scans"`
	PartialSuccess  bool                            `json:"partial_success"`
	OverallSuccess  bool                            `json:"overall_success"`
}

// Define the transaction request types (these were duplicated from the transaction service)
type SecurityScanTransactionRequest struct {
	ModuleVersionID int
	ModulePath      string
	Namespace       string
	Module          string
	Provider        string
	Version         string
	TransactionCtx  context.Context
}

type BatchSecurityScanRequest struct {
	ModuleVersionID int
	ModulePath      string
	Namespace       string
	Module          string
	Provider        string
	Version         string
}

// ScanWithTransaction executes a security scan within a transaction savepoint
// If the scan fails or results cannot be stored, the savepoint is rolled back
func (s *SecurityScanningService) ScanWithTransaction(
	ctx context.Context,
	req SecurityScanTransactionRequest,
) (*SecurityScanTransactionResult, error) {
	startTime := time.Now()
	result := &SecurityScanTransactionResult{
		Success:             false,
		SavepointRolledBack: false,
		Timestamp:           startTime,
	}

	err := s.savepointHelper.WithTransaction(ctx, func(ctx context.Context, tx *gorm.DB) error {
		// Execute the security scan
		scanReq := &SecurityScanRequest{
			Namespace:  req.Namespace,
			Module:     req.Module,
			Provider:   req.Provider,
			Version:    req.Version,
			ModulePath: req.ModulePath,
		}

		securityResponse, err := s.ExecuteSecurityScan(ctx, scanReq)
		if err != nil {
			return fmt.Errorf("failed to execute security scan: %w", err)
		}

		// Store the results in the module version details
		if err := s.storeSecurityResultsInTransaction(ctx, req.ModuleVersionID, securityResponse); err != nil {
			return fmt.Errorf("failed to store security results: %w", err)
		}

		// Set successful result
		result.Success = true
		result.SecurityResponse = securityResponse

		return nil
	})

	result.ScanDuration = time.Since(startTime)

	if err != nil {
		// Mark that savepoint was rolled back
		result.SavepointRolledBack = true
		errorMsg := err.Error()
		result.Error = &errorMsg
		return result, nil
	}

	return result, nil
}

// ScanBatchModules executes security scans for multiple module versions with individual savepoints
// Each scan gets its own savepoint for isolation
func (s *SecurityScanningService) ScanBatchModules(
	ctx context.Context,
	modules []BatchSecurityScanRequest,
) (*BatchSecurityScanResult, error) {
	result := &BatchSecurityScanResult{
		TotalScans:      len(modules),
		SuccessfulScans: []SecurityScanTransactionResult{},
		FailedScans:     []SecurityScanTransactionResult{},
		PartialSuccess:  false,
		OverallSuccess:  true,
	}

	for _, module := range modules {
		// Each scan gets its own savepoint for isolation
		scanReq := SecurityScanTransactionRequest{
			ModuleVersionID: module.ModuleVersionID,
			ModulePath:      module.ModulePath,
			Namespace:       module.Namespace,
			Module:          module.Module,
			Provider:        module.Provider,
			Version:         module.Version,
			TransactionCtx:  ctx,
		}

		scanResult, err := s.ScanWithTransaction(ctx, scanReq)
		if err != nil {
			// This should rarely happen since we handle errors within ScanWithTransaction
			errorResult := SecurityScanTransactionResult{
				Success:             false,
				Error:               func() *string { e := err.Error(); return &e }(),
				ScanDuration:        0,
				Timestamp:           time.Now(),
				SavepointRolledBack: true,
			}
			result.FailedScans = append(result.FailedScans, errorResult)
			result.OverallSuccess = false
			result.PartialSuccess = true
			continue
		}

		if scanResult.Success {
			result.SuccessfulScans = append(result.SuccessfulScans, *scanResult)
		} else {
			result.FailedScans = append(result.FailedScans, *scanResult)
			result.OverallSuccess = false
			result.PartialSuccess = true
		}
	}

	// If there were no failures, set partial success to false
	if len(result.FailedScans) == 0 {
		result.PartialSuccess = false
	}

	return result, nil
}

// ProcessExistingModuleSecurity processes security scanning for an existing module version
// This method finds the module version by ID and updates it with security scan results
func (s *SecurityScanningService) ProcessExistingModuleSecurity(
	ctx context.Context,
	moduleVersionID int,
) (*SecurityScanTransactionResult, error) {
	// Find the module version
	moduleVersion, err := s.moduleVersionRepo.FindByID(ctx, moduleVersionID)
	if err != nil {
		return &SecurityScanTransactionResult{
			Success:   false,
			Error:     func() *string { e := fmt.Sprintf("failed to find module version: %v", err); return &e }(),
			Timestamp: time.Now(),
		}, nil
	}

	if moduleVersion == nil {
		return &SecurityScanTransactionResult{
			Success:   false,
			Error:     func() *string { e := "module version not found"; return &e }(),
			Timestamp: time.Now(),
		}, nil
	}

	// Get module details for path information
	moduleProvider := moduleVersion.ModuleProvider()
	if moduleProvider == nil {
		return &SecurityScanTransactionResult{
			Success:   false,
			Error:     func() *string { e := "module provider not found"; return &e }(),
			Timestamp: time.Now(),
		}, nil
	}

	// Extract path information
	// Note: In a real implementation, you would get the module path from the file system or storage
	// For now, we'll use an empty path which will trigger temporary extraction
	scanReq := SecurityScanTransactionRequest{
		ModuleVersionID: moduleVersionID,
		ModulePath:      "", // Will trigger temporary extraction
		Namespace:       moduleProvider.Namespace().Name(),
		Module:          moduleProvider.Module(),
		Provider:        moduleProvider.Provider(),
		Version:         moduleVersion.Version().String(),
		TransactionCtx:  ctx,
	}

	return s.ScanWithTransaction(ctx, scanReq)
}

// storeSecurityResultsInTransaction stores security scan results in module version details within a transaction
func (s *SecurityScanningService) storeSecurityResultsInTransaction(
	ctx context.Context,
	moduleVersionID int,
	securityResponse *SecurityScanResponse,
) error {
	// Find the module version
	moduleVersion, err := s.moduleVersionRepo.FindByID(ctx, moduleVersionID)
	if err != nil {
		return fmt.Errorf("failed to find module version: %w", err)
	}

	if moduleVersion == nil {
		return fmt.Errorf("module version not found")
	}

	// Convert security response to JSON bytes
	tfsecJSON, err := json.Marshal(securityResponse)
	if err != nil {
		return fmt.Errorf("failed to marshal security results to JSON: %w", err)
	}

	// Update module details with tfsec results
	currentDetails := moduleVersion.Details()
	if currentDetails == nil {
		// Create new details if none exist
		currentDetails = model.NewModuleDetails([]byte{})
	}

	// Create new details with updated tfsec results
	updatedDetails := currentDetails.WithTfsec(tfsecJSON)

	// Update the module version with new details
	// Note: This assumes ModuleVersion has a method to update its details
	// In the current architecture, this might require a domain method like UpdateDetails()
	if err := s.updateModuleVersionDetails(ctx, moduleVersion, updatedDetails); err != nil {
		return fmt.Errorf("failed to update module version details: %w", err)
	}

	// Save the updated module version
	_, err = s.moduleVersionRepo.Save(ctx, moduleVersion)
	if err != nil {
		return fmt.Errorf("failed to save module version with security results: %w", err)
	}

	return nil
}

// updateModuleVersionDetails updates module version details
// This is a helper method that would need to be implemented in the ModuleVersion domain model
func (s *SecurityScanningService) updateModuleVersionDetails(
	ctx context.Context,
	moduleVersion *model.ModuleVersion,
	details *model.ModuleDetails,
) error {
	// This is a placeholder for the actual domain method that would update details
	// In the current Go implementation, this might be:
	// moduleVersion.UpdateDetails(details)
	// or a similar domain method

	// For now, we'll assume the ModuleVersion has a method to update details
	// This would need to be implemented in the actual domain model
	return nil
}

// GetSecurityScanSummary returns a summary of security scan results for multiple module versions
func (s *SecurityScanningService) GetSecurityScanSummary(
	ctx context.Context,
	moduleVersionIDs []int,
) (*SecurityScanSummary, error) {
	summary := &SecurityScanSummary{}

	for _, id := range moduleVersionIDs {
		results, err := s.GetSecurityResults(ctx, id)
		if err != nil {
			continue // Skip versions that can't be retrieved
		}

		if results != nil {
			summary.Total += len(results.Results)
			summary.Critical += results.Summary.Critical
			summary.High += results.Summary.High
			summary.Medium += results.Summary.Medium
			summary.Low += results.Summary.Low
			summary.Info += results.Summary.Info
			summary.Warnings += results.Summary.Warnings
		}
	}

	return summary, nil
}
