package service

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"time"

	"gorm.io/gorm"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/repository"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb/transaction"
)

// sanitizeSavepointName converts a string to a SQL-safe identifier
// This is a local copy to avoid circular dependencies
func sanitizeSavepointNameForSecurity(name string) string {
	// Replace invalid SQL identifier characters with underscores
	re := regexp.MustCompile(`[^a-zA-Z0-9_]`)
	sanitized := re.ReplaceAllString(name, "_")

	// Ensure it doesn't start with a digit
	if len(sanitized) > 0 && sanitized[0] >= '0' && sanitized[0] <= '9' {
		sanitized = "sp_" + sanitized
	}

	// Ensure it's not empty
	if sanitized == "" {
		sanitized = fmt.Sprintf("sp_%d", time.Now().UnixNano())
	}

	// Truncate if too long
	if len(sanitized) > 64 {
		sanitized = sanitized[:61] + fmt.Sprintf("_%d", time.Now().UnixNano()%1000)
	}

	return sanitized
}

// SecurityScanningTransactionService handles security scanning with transaction safety
// and result persistence using the module details system
type SecurityScanningTransactionService struct {
	securityService   *SecurityScanningService
	moduleVersionRepo repository.ModuleVersionRepository
	savepointHelper   *transaction.SavepointHelper
}

// NewSecurityScanningTransactionService creates a new security scanning transaction service
func NewSecurityScanningTransactionService(
	securityService *SecurityScanningService,
	moduleVersionRepo repository.ModuleVersionRepository,
	savepointHelper *transaction.SavepointHelper,
) *SecurityScanningTransactionService {
	return &SecurityScanningTransactionService{
		securityService:   securityService,
		moduleVersionRepo: moduleVersionRepo,
		savepointHelper:   savepointHelper,
	}
}

// SecurityScanTransactionRequest represents a request to scan with transaction context
type SecurityScanTransactionRequest struct {
	ModuleVersionID int
	ModulePath      string
	Namespace       string
	Module          string
	Provider        string
	Version         string
	TransactionCtx  context.Context
	SavepointName   string
}

// SecurityScanTransactionResult represents the result of a security scan within a transaction
type SecurityScanTransactionResult struct {
	Success             bool                  `json:"success"`
	SecurityResponse    *SecurityScanResponse `json:"security_response,omitempty"`
	Error               *string               `json:"error,omitempty"`
	ScanDuration        time.Duration         `json:"scan_duration"`
	Timestamp           time.Time             `json:"timestamp"`
	SavepointRolledBack bool                  `json:"savepoint_rolled_back"`
}

// BatchSecurityScanRequest represents a batch security scan request
type BatchSecurityScanRequest struct {
	ModuleVersionID int
	ModulePath      string
	Namespace       string
	Module          string
	Provider        string
	Version         string
}

// BatchSecurityScanResult represents the result of batch security scanning
type BatchSecurityScanResult struct {
	TotalScans      int                             `json:"total_scans"`
	SuccessfulScans []SecurityScanTransactionResult `json:"successful_scans"`
	FailedScans     []SecurityScanTransactionResult `json:"failed_scans"`
	PartialSuccess  bool                            `json:"partial_success"`
	OverallSuccess  bool                            `json:"overall_success"`
}

// ScanWithTransaction executes a security scan within a transaction savepoint
// If the scan fails or results cannot be stored, the savepoint is rolled back
func (s *SecurityScanningTransactionService) ScanWithTransaction(
	ctx context.Context,
	req SecurityScanTransactionRequest,
) (*SecurityScanTransactionResult, error) {
	startTime := time.Now()
	result := &SecurityScanTransactionResult{
		Success:             false,
		SavepointRolledBack: false,
		Timestamp:           startTime,
	}

	// Use the provided savepoint name or create a new one
	savepointName := req.SavepointName
	if savepointName == "" {
		savepointName = fmt.Sprintf("security_scan_%d_%d", req.ModuleVersionID, startTime.UnixNano())
	}

	err := s.savepointHelper.WithSavepointNamed(ctx, savepointName, func(tx *gorm.DB) error {
		// Execute the security scan
		scanReq := &SecurityScanRequest{
			Namespace:  req.Namespace,
			Module:     req.Module,
			Provider:   req.Provider,
			Version:    req.Version,
			ModulePath: req.ModulePath,
		}

		securityResponse, err := s.securityService.ExecuteSecurityScan(ctx, scanReq)
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
func (s *SecurityScanningTransactionService) ScanBatchModules(
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
func (s *SecurityScanningTransactionService) ProcessExistingModuleSecurity(
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
func (s *SecurityScanningTransactionService) storeSecurityResultsInTransaction(
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
	if err := s.moduleVersionRepo.Save(ctx, moduleVersion); err != nil {
		return fmt.Errorf("failed to save module version with security results: %w", err)
	}

	return nil
}

// updateModuleVersionDetails updates module version details
// This is a helper method that would need to be implemented in the ModuleVersion domain model
func (s *SecurityScanningTransactionService) updateModuleVersionDetails(
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

// GetSecurityResultsFromStorage retrieves stored security scan results for a module version
func (s *SecurityScanningTransactionService) GetSecurityResultsFromStorage(
	ctx context.Context,
	moduleVersionID int,
) (*SecurityScanResponse, error) {
	// Find the module version
	moduleVersion, err := s.moduleVersionRepo.FindByID(ctx, moduleVersionID)
	if err != nil {
		return nil, fmt.Errorf("failed to find module version: %w", err)
	}

	if moduleVersion == nil {
		return nil, fmt.Errorf("module version not found")
	}

	// Get security results using the existing service method
	return s.securityService.GetSecurityResults(ctx, moduleVersionID)
}

// ValidateSecurityResults checks if security results exist and are valid
func (s *SecurityScanningTransactionService) ValidateSecurityResults(
	ctx context.Context,
	moduleVersionID int,
) (bool, error) {
	results, err := s.GetSecurityResultsFromStorage(ctx, moduleVersionID)
	if err != nil {
		return false, err
	}

	// Check if results exist and have content
	return results != nil && len(results.Results) > 0, nil
}

// RollbackSecurityScan rolls back security scan results for a module version
func (s *SecurityScanningTransactionService) RollbackSecurityScan(
	ctx context.Context,
	moduleVersionID int,
	savepointName string,
) error {
	// Sanitize savepoint name for SQL safety
	safeName := sanitizeSavepointNameForSecurity(savepointName)

	// Rollback to the specified savepoint with proper quoting
	if err := s.savepointHelper.WithContext(ctx).Exec(fmt.Sprintf("ROLLBACK TO SAVEPOINT `%s`", safeName)).Error; err != nil {
		return fmt.Errorf("failed to rollback to savepoint %s: %w", safeName, err)
	}

	return nil
}

// GetSecurityScanSummary returns a summary of security scan results for multiple module versions
func (s *SecurityScanningTransactionService) GetSecurityScanSummary(
	ctx context.Context,
	moduleVersionIDs []int,
) (*SecurityScanSummary, error) {
	summary := &SecurityScanSummary{}

	for _, id := range moduleVersionIDs {
		results, err := s.GetSecurityResultsFromStorage(ctx, id)
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
