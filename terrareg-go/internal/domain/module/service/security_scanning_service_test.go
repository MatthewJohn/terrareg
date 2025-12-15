package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSecurityScanningService(t *testing.T) {
	service := NewSecurityScanningService(nil, nil)
	assert.NotNil(t, service)
}

func TestSecurityScanningService_BasicStructure(t *testing.T) {
	// Basic test to verify the service structure exists
	service := NewSecurityScanningService(nil, nil)
	assert.NotNil(t, service)
	// These are nil because we passed nil to constructor, which is expected for this test
	assert.Nil(t, service.moduleFileService)
	assert.Nil(t, service.moduleVersionRepo)
}

func TestSecurityScanningService_ExecuteSecurityScan_WithoutTfsec(t *testing.T) {
	// Test that the service handles missing dependencies gracefully
	service := NewSecurityScanningService(nil, nil)

	req := &SecurityScanRequest{
		Namespace:  "test",
		Module:     "test-module",
		Provider:   "aws",
		Version:    "1.0.0",
		ModulePath: "", // Empty to trigger module file extraction
	}

	// Since moduleFileService is nil, this should fail gracefully with module files error
	result, err := service.ExecuteSecurityScan(context.Background(), req)

	// We expect an error because the service can't function without proper dependencies
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "module file service not initialized")
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
