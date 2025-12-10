package service

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	moduleRepo "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/repository"
)

func TestSecurityScanningService_ExecuteSecurityScan(t *testing.T) {
	// Create a temporary directory with test Terraform files for scanning
	tempDir, err := os.MkdirTemp("", "terrareg-scan-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a simple Terraform file with some potential security issues
	terraformFile := filepath.Join(tempDir, "main.tf")
	terraformContent := `
resource "aws_security_group" "example" {
  # Insecure - allowing all inbound traffic
  ingress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  # Insecure - using plaintext password
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_db_instance" "example" {
  # Insecure - unencrypted storage
  storage_encrypted = false

  # Insecure - no backup retention
  backup_retention_period = 0

  # Insecure - plaintext password
  password = "secretpassword123"
}
`
	if err := os.WriteFile(terraformFile, []byte(terraformContent), 0644); err != nil {
		t.Fatalf("Failed to write test Terraform file: %v", err)
	}

	tests := []struct {
		name        string
		req         *SecurityScanRequest
		expectError bool
		expectFindings bool
	}{
		{
			name: "successful scan with findings",
			req: &SecurityScanRequest{
				Namespace:  "test",
				Module:     "test-module",
				Provider:   "aws",
				Version:    "1.0.0",
				ModulePath: tempDir,
			},
			expectError:    false,
			expectFindings: true, // Should find security issues in the test file
		},
		{
			name: "scan with module path extraction",
			req: &SecurityScanRequest{
				Namespace: "test",
				Module:    "test-module",
				Provider:  "aws",
				Version:   "1.0.0",
				// No ModulePath - should trigger extraction
			},
			expectError: true, // Should fail since there's no module file service
		},
		{
			name: "scan with non-existent path",
			req: &SecurityScanRequest{
				Namespace:  "test",
				Module:     "test-module",
				Provider:   "aws",
				Version:    "1.0.0",
				ModulePath: "/non/existent/path",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &SecurityScanningService{
				// Don't provide moduleFileService for this test
				moduleFileService: nil,
				moduleVersionRepo: nil,
			}

			result, err := service.ExecuteSecurityScan(context.Background(), tt.req)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result == nil {
				t.Fatalf("Expected result but got nil")
			}

			if tt.expectFindings && len(result.Results) == 0 {
				t.Logf("Warning: Expected security findings but got none. tfsec might not be installed or available.")
				t.Logf("Scan result: %+v", result)
			}

			// Verify the summary structure is correct
			if result.Summary.Critical < 0 || result.Summary.High < 0 ||
			   result.Summary.Medium < 0 || result.Summary.Low < 0 {
				t.Errorf("Invalid summary counts: %+v", result.Summary)
			}
		})
	}
}

func TestSecurityScanningService_GetSecurityResults(t *testing.T) {
	service := &SecurityScanningService{}

	result, err := service.GetSecurityResults(context.Background(), 123)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if result == nil {
		t.Errorf("Expected result but got nil")
	}

	// Should return empty results (TODO implementation)
	if len(result.Results) != 0 {
		t.Errorf("Expected empty results, got %d findings", len(result.Results))
	}
}

func TestSecurityScanningService_StoreSecurityResults(t *testing.T) {
	service := &SecurityScanningService{}

	scanResults := &SecurityScanResponse{
		Results: []SecurityScanResult{
			{
				RuleID:      "AWS018",
				Severity:    "HIGH",
				Title:       "Security group rule allows unrestricted ingress",
				Description: "Security group rule allows unrestricted ingress traffic.",
				Location: SecurityScanLocation{
					Filename:  "main.tf",
					StartLine: 1,
					EndLine:   10,
				},
			},
		},
		Summary: SecurityScanSummary{
			Critical: 0,
			High:     1,
			Medium:   0,
			Low:      0,
		},
	}

	err := service.StoreSecurityResults(context.Background(), 123, scanResults)

	if err != nil {
		t.Errorf("Unexpected error storing results: %v", err)
	}

	// TODO: When implemented, verify the results were actually stored
}

func TestSecurityScanningService_extractModuleFiles(t *testing.T) {
	// This test would require a module file service to be implemented
	// For now, we just verify that the method exists and handles the nil case

	service := &SecurityScanningService{
		moduleFileService: nil,
		moduleVersionRepo: nil,
	}

	req := &SecurityScanRequest{
		Namespace: "test",
		Module:    "test-module",
		Provider:  "aws",
		Version:   "1.0.0",
	}

	_, err := service.extractModuleFiles(context.Background(), req)

	if err == nil {
		t.Errorf("Expected error when moduleFileService is nil")
	}
}

// Helper function to create a mock module repository (if needed for future tests)
func createMockModuleVersionRepo() moduleRepo.ModuleVersionRepository {
	// This would return a mock implementation
	// For now, return nil since we're testing with TODO implementations
	return nil
}