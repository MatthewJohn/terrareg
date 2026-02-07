package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/model"
	types "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared/types"
	"github.com/stretchr/testify/assert"
)

func TestNewSecurityScanningService(t *testing.T) {
	_, err := NewSecurityScanningService(nil, nil, nil, nil, nil)
	assert.Error(t, err, "NewSecurityScanningService should return error when required dependencies are nil")
}

func TestSecurityScanningService_BasicStructure(t *testing.T) {
	// Basic test to verify the service structure exists
	_, err := NewSecurityScanningService(nil, nil, nil, nil, nil)
	assert.Error(t, err, "NewSecurityScanningService should return error when required dependencies are nil")
}

func TestSecurityScanningService_ExecuteSecurityScan_WithoutTfsec(t *testing.T) {
	// Test that the service handles missing dependencies gracefully
	_, err := NewSecurityScanningService(nil, nil, nil, nil, nil)
	assert.Error(t, err)
}

func TestSecurityScanSummary_Structure(t *testing.T) {
	// Test that SecurityScanSummary has the correct structure
	summary := &SecurityScanSummary{
		Total:    5,
		Critical: 1,
		High:     1,
		Medium:   2,
		Low:      1,
		Info:     0,
		Warnings: 0,
	}

	assert.Equal(t, 5, summary.Total)
	assert.Equal(t, 1, summary.Critical)
	assert.Equal(t, 1, summary.High)
	assert.Equal(t, 2, summary.Medium)
	assert.Equal(t, 1, summary.Low)
	assert.Equal(t, 0, summary.Info)
	assert.Equal(t, 0, summary.Warnings)
}

func TestSecurityScanResult_Structure(t *testing.T) {
	// Test that SecurityScanResult has the correct structure
	result := &SecurityScanResult{
		RuleID:      "AWS001",
		Severity:    "HIGH",
		Title:       "Insecure Security Group",
		Description: "Security group allows unrestricted ingress",
		Location: SecurityScanLocation{
			Filename:  "main.tf",
			StartLine: 10,
			EndLine:   15,
		},
		Links: []string{"https://tfsec.dev/docs/aws/AWS001/"},
	}

	assert.Equal(t, "AWS001", result.RuleID)
	assert.Equal(t, "HIGH", result.Severity)
	assert.Equal(t, "Insecure Security Group", result.Title)
	assert.Equal(t, "Security group allows unrestricted ingress", result.Description)
	assert.Equal(t, "main.tf", result.Location.Filename)
	assert.Equal(t, 10, result.Location.StartLine)
	assert.Equal(t, 15, result.Location.EndLine)
	assert.Len(t, result.Links, 1)
}

func TestSecurityScanResponse_Structure(t *testing.T) {
	// Test that SecurityScanResponse has the correct structure
	summary := &SecurityScanSummary{
		Total:    1,
		Critical: 0,
		High:     1,
		Medium:   0,
		Low:      0,
		Info:     0,
		Warnings: 0,
	}

	result := &SecurityScanResult{
		RuleID:      "AWS001",
		Severity:    "HIGH",
		Title:       "Insecure Security Group",
		Description: "Security group allows unrestricted ingress",
		Location: SecurityScanLocation{
			Filename:  "main.tf",
			StartLine: 10,
			EndLine:   15,
		},
	}

	response := &SecurityScanResponse{
		Results: []SecurityScanResult{*result},
		Summary: *summary,
	}

	assert.Len(t, response.Results, 1)
	assert.Equal(t, 1, response.Summary.Total)
	assert.Equal(t, 1, response.Summary.High)
}

// Mock repositories for testing updateModuleVersionDetails

type MockModuleDetailsRepository struct {
	saveAndReturnIDFunc func(ctx context.Context, details *model.ModuleDetails) (int, error)
}

func (m *MockModuleDetailsRepository) Save(ctx context.Context, details *model.ModuleDetails) (*model.ModuleDetails, error) {
	return nil, nil
}

func (m *MockModuleDetailsRepository) SaveAndReturnID(ctx context.Context, details *model.ModuleDetails) (int, error) {
	if m.saveAndReturnIDFunc != nil {
		return m.saveAndReturnIDFunc(ctx, details)
	}
	return 1, nil
}

func (m *MockModuleDetailsRepository) FindByID(ctx context.Context, id int) (*model.ModuleDetails, error) {
	return nil, nil
}

func (m *MockModuleDetailsRepository) Update(ctx context.Context, id int, details *model.ModuleDetails) (*model.ModuleDetails, error) {
	return nil, nil
}

func (m *MockModuleDetailsRepository) Delete(ctx context.Context, id int) error {
	return nil
}

func (m *MockModuleDetailsRepository) FindByModuleVersionID(ctx context.Context, moduleVersionID int) (*model.ModuleDetails, error) {
	return nil, nil
}

type MockModuleVersionRepositoryForUpdate struct {
	updateModuleDetailsIDFunc func(ctx context.Context, moduleVersionID int, moduleDetailsID int) error
	saveFunc                  func(ctx context.Context, moduleVersion *model.ModuleVersion) (*model.ModuleVersion, error)
}

func (m *MockModuleVersionRepositoryForUpdate) Save(ctx context.Context, moduleVersion *model.ModuleVersion) (*model.ModuleVersion, error) {
	if m.saveFunc != nil {
		return m.saveFunc(ctx, moduleVersion)
	}
	return moduleVersion, nil
}

func (m *MockModuleVersionRepositoryForUpdate) UpdateModuleDetailsID(ctx context.Context, moduleVersionID int, moduleDetailsID int) error {
	if m.updateModuleDetailsIDFunc != nil {
		return m.updateModuleDetailsIDFunc(ctx, moduleVersionID, moduleDetailsID)
	}
	return nil
}

// Minimal interface implementation to satisfy the compiler
func (m *MockModuleVersionRepositoryForUpdate) FindByID(ctx context.Context, id int) (*model.ModuleVersion, error) {
	return nil, nil
}

func (m *MockModuleVersionRepositoryForUpdate) FindByModuleProvider(ctx context.Context, moduleProviderID int, includeBeta, includeUnpublished bool) ([]*model.ModuleVersion, error) {
	return nil, nil
}

func (m *MockModuleVersionRepositoryForUpdate) FindByModuleProviderAndVersion(ctx context.Context, moduleProviderID int, version types.ModuleVersion) (*model.ModuleVersion, error) {
	return nil, nil
}

func (m *MockModuleVersionRepositoryForUpdate) Delete(ctx context.Context, id int) error {
	return nil
}

func (m *MockModuleVersionRepositoryForUpdate) Exists(ctx context.Context, moduleProviderID int, version types.ModuleVersion) (bool, error) {
	return false, nil
}

func TestSecurityScanningService_updateModuleVersionDetails_Success(t *testing.T) {
	// This test verifies that updateModuleVersionDetails:
	// 1. Saves the module details using moduleDetailsRepo.SaveAndReturnID()
	// 2. Updates the module version using moduleVersionRepo.UpdateModuleDetailsID()

	calledSave := false
	var savedDetails *model.ModuleDetails

	mockDetailsRepo := &MockModuleDetailsRepository{
		saveAndReturnIDFunc: func(ctx context.Context, details *model.ModuleDetails) (int, error) {
			calledSave = true
			savedDetails = details
			return 42, nil // Return mock details ID
		},
	}

	calledUpdate := false
	var capturedModuleDetailsID int

	mockVersionRepo := &MockModuleVersionRepositoryForUpdate{
		updateModuleDetailsIDFunc: func(ctx context.Context, moduleVersionID int, moduleDetailsID int) error {
			calledUpdate = true
			capturedModuleDetailsID = moduleDetailsID
			return nil
		},
	}

	service, err := NewSecurityScanningService(nil, mockVersionRepo, mockDetailsRepo, nil, nil)
	assert.NoError(t, err)
	assert.NotNil(t, service)

	// Create test module version with ID 123
	moduleVersion, err := model.NewModuleVersion("1.0.0", nil, false)
	assert.NoError(t, err)
	// We need to set the ID on the module version - for this test we'll check if it was called

	// Create test module details with tfsec data
	// Format matches Python tfsec output: {"results": [...]}
	tfsecJSON := []byte(`{"results": [{"rule_id": "AWS001", "severity": "HIGH"}]}`)
	details := model.NewModuleDetails([]byte{})
	updatedDetails := details.WithTfsec(tfsecJSON)

	// Call updateModuleVersionDetails
	ctx := context.Background()
	err = service.updateModuleVersionDetails(ctx, moduleVersion, updatedDetails)

	// This test will FAIL because updateModuleVersionDetails is currently a placeholder
	// After implementing the fix, these assertions should pass:
	assert.NoError(t, err, "updateModuleVersionDetails should not return an error")
	assert.True(t, calledSave, "SaveAndReturnID should have been called")
	assert.NotNil(t, savedDetails, "savedDetails should not be nil")
	assert.True(t, calledUpdate, "UpdateModuleDetailsID should have been called")
	// Note: We can't assert exact moduleVersionID without setting it on the moduleVersion object
	// For now, just verify the update was called with the correct details ID
	assert.Equal(t, 42, capturedModuleDetailsID, "UpdateModuleDetailsID should be called with details ID 42")
}

func TestSecurityScanningService_updateModuleVersionDetails_SaveFails(t *testing.T) {
	// This test verifies error handling when SaveAndReturnID fails

	mockDetailsRepo := &MockModuleDetailsRepository{
		saveAndReturnIDFunc: func(ctx context.Context, details *model.ModuleDetails) (int, error) {
			return 0, fmt.Errorf("database connection failed")
		},
	}

	mockVersionRepo := &MockModuleVersionRepositoryForUpdate{
		updateModuleDetailsIDFunc: func(ctx context.Context, moduleVersionID int, moduleDetailsID int) error {
			// This should NOT be called because SaveAndReturnID failed
			t.Error("UpdateModuleDetailsID should not be called when SaveAndReturnID fails")
			return nil
		},
	}

	service, err := NewSecurityScanningService(nil, mockVersionRepo, mockDetailsRepo, nil, nil)
	assert.NoError(t, err)
	assert.NotNil(t, service)

	moduleVersion, err := model.NewModuleVersion("1.0.0", nil, false)
	assert.NoError(t, err)
	tfsecJSON := []byte(`{"results": []}`)
	details := model.NewModuleDetails([]byte{})
	updatedDetails := details.WithTfsec(tfsecJSON)

	ctx := context.Background()
	err = service.updateModuleVersionDetails(ctx, moduleVersion, updatedDetails)

	// This test will FAIL because updateModuleVersionDetails is currently a placeholder
	// After implementing the fix, these assertions should pass:
	assert.Error(t, err, "updateModuleVersionDetails should return an error when SaveAndReturnID fails")
	if err != nil {
		assert.Contains(t, err.Error(), "failed to save module details", "error message should mention saving failed")
		assert.Contains(t, err.Error(), "database connection failed", "error message should include underlying error")
	}
}

func TestSecurityScanningService_updateModuleVersionDetails_UpdateFails(t *testing.T) {
	// This test verifies error handling when UpdateModuleDetailsID fails

	mockDetailsRepo := &MockModuleDetailsRepository{
		saveAndReturnIDFunc: func(ctx context.Context, details *model.ModuleDetails) (int, error) {
			return 42, nil // Save succeeds
		},
	}

	mockVersionRepo := &MockModuleVersionRepositoryForUpdate{
		updateModuleDetailsIDFunc: func(ctx context.Context, moduleVersionID int, moduleDetailsID int) error {
			return fmt.Errorf("constraint violation: foreign key mismatch")
		},
	}

	service, err := NewSecurityScanningService(nil, mockVersionRepo, mockDetailsRepo, nil, nil)
	assert.NoError(t, err)
	assert.NotNil(t, service)

	moduleVersion, err := model.NewModuleVersion("1.0.0", nil, false)
	assert.NoError(t, err)
	tfsecJSON := []byte(`{"results": []}`)
	details := model.NewModuleDetails([]byte{})
	updatedDetails := details.WithTfsec(tfsecJSON)

	ctx := context.Background()
	err = service.updateModuleVersionDetails(ctx, moduleVersion, updatedDetails)

	// This test will FAIL because updateModuleVersionDetails is currently a placeholder
	// After implementing the fix, these assertions should pass:
	assert.Error(t, err, "updateModuleVersionDetails should return an error when UpdateModuleDetailsID fails")
	if err != nil {
		assert.Contains(t, err.Error(), "failed to update module version with details ID", "error message should mention updating failed")
		assert.Contains(t, err.Error(), "constraint violation", "error message should include underlying error")
	}
}
