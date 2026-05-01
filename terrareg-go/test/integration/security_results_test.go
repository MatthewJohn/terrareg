package integration

import (
	"encoding/json"
	"testing"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSecurityResultsParsingFromDBFormat tests that security results
// are correctly parsed from the actual database format and filtered
// to only include failures (status == 0), matching Python behavior.
// Reference: terrareg/models.py get_tfsec_failures()
func TestSecurityResultsParsingFromDBFormat(t *testing.T) {
	// This is the actual format stored in the database by the Python module_extractor
	// and by the Golang SecurityScanningService
	tfsecJSON := []byte(`{
		"results": [
			{
				"rule_id": "AVD-AWS-0132",
				"long_id": "aws-s3-encryption-customer-key",
				"rule_description": "S3 encryption should use Customer Managed Keys",
				"rule_provider": "aws",
				"rule_service": "s3",
				"impact": "Using AWS managed keys does not allow for fine grained control",
				"resolution": "Enable encryption using customer managed keys",
				"links": [
					"https://aquasecurity.github.io/tfsec/v1.28.14/checks/aws/s3/encryption-customer-key/"
				],
				"description": "Bucket does not encrypt data with a customer managed key.",
				"severity": "HIGH",
				"warning": false,
				"status": 0,
				"resource": "aws_s3_bucket.mybucket",
				"location": {
					"filename": "main.tf",
					"start_line": 2,
					"end_line": 8
				}
			},
			{
				"rule_id": "AVD-AWS-0090",
				"long_id": "aws-s3-enable-versioning",
				"rule_description": "S3 Data should be versioned",
				"rule_provider": "aws",
				"rule_service": "s3",
				"impact": "Deleted or modified data would not be recoverable",
				"resolution": "Enable versioning to protect against accidental/malicious removal or modification",
				"links": [
					"https://aquasecurity.github.io/tfsec/v1.28.14/checks/aws/s3/enable-versioning/"
				],
				"description": "",
				"severity": "MEDIUM",
				"warning": false,
				"status": 1,
				"resource": "aws_s3_bucket.mybucket",
				"location": {
					"filename": "main.tf",
					"start_line": 2,
					"end_line": 8
				}
			},
			{
				"rule_id": "AVD-AWS-0091",
				"long_id": "aws-s3-ignore-public-acls",
				"rule_description": "S3 Access Block should Ignore Public Acl",
				"rule_provider": "aws",
				"rule_service": "s3",
				"impact": "PUT calls with public ACLs specified can make objects public",
				"resolution": "Enable ignoring the application of public ACLs in PUT calls",
				"links": ["https://aquasecurity.github.io/tfsec/v1.28.14/checks/aws/s3/ignore-public-acls/"],
				"description": "",
				"severity": "HIGH",
				"warning": false,
				"status": 1,
				"resource": "aws_s3_bucket.mybucket",
				"location": {
					"filename": "main.tf",
					"start_line": 2,
					"end_line": 8
				}
			}
		]
	}`)

	details := model.NewModuleDetails([]byte{}).WithTfsec(tfsecJSON)
	mv, err := model.NewModuleVersion("1.0.0", details, false)
	require.NoError(t, err)

	// Get security results - should only include failures (status == 0)
	results := mv.GetSecurityResults()

	// Only 1 result should be returned (the one with status == 0)
	assert.Len(t, results, 1, "Only failures (status == 0) should be returned")

	// Verify the result is the failing one
	assert.Equal(t, "AVD-AWS-0132", results[0].RuleID)
	assert.Equal(t, "HIGH", results[0].Severity)
	assert.Equal(t, "Bucket does not encrypt data with a customer managed key.", results[0].Description)
	assert.Equal(t, "main.tf", results[0].Location.Filename)
	assert.Equal(t, 2, results[0].Location.StartLine)
	assert.Equal(t, 8, results[0].Location.EndLine)

	// Verify GetSecurityFailures returns the correct count
	assert.Equal(t, 1, mv.GetSecurityFailures())
}

// TestSecurityResultsWithAllFailures tests when all results are failures
func TestSecurityResultsWithAllFailures(t *testing.T) {
	tfsecJSON := []byte(`{
		"results": [
			{
				"rule_id": "AWS001",
				"severity": "HIGH",
				"description": "First failure",
				"status": 0,
				"location": {"filename": "main.tf", "start_line": 1, "end_line": 5}
			},
			{
				"rule_id": "AWS002",
				"severity": "LOW",
				"description": "Second failure",
				"status": 0,
				"location": {"filename": "vars.tf", "start_line": 10, "end_line": 15}
			}
		]
	}`)

	details := model.NewModuleDetails([]byte{}).WithTfsec(tfsecJSON)
	mv, err := model.NewModuleVersion("1.0.0", details, false)
	require.NoError(t, err)

	results := mv.GetSecurityResults()
	assert.Len(t, results, 2, "All failures should be returned")
	assert.Equal(t, 2, mv.GetSecurityFailures())
}

// TestSecurityResultsWithNoFailures tests when all results are passes
func TestSecurityResultsWithNoFailures(t *testing.T) {
	tfsecJSON := []byte(`{
		"results": [
			{
				"rule_id": "AWS001",
				"severity": "HIGH",
				"description": "Passed check",
				"status": 2,
				"location": {"filename": "main.tf", "start_line": 1, "end_line": 5}
			},
			{
				"rule_id": "AWS002",
				"severity": "LOW",
				"description": "Another passed check",
				"status": 1,
				"location": {"filename": "vars.tf", "start_line": 10, "end_line": 15}
			}
		]
	}`)

	details := model.NewModuleDetails([]byte{}).WithTfsec(tfsecJSON)
	mv, err := model.NewModuleVersion("1.0.0", details, false)
	require.NoError(t, err)

	results := mv.GetSecurityResults()
	assert.Empty(t, results, "No failures should be returned when all status != 0")
	assert.Equal(t, 0, mv.GetSecurityFailures())
}

// TestSecurityResultsToAPIResponse tests that security results
// are correctly converted to the API response format
func TestSecurityResultsToAPIResponse(t *testing.T) {
	tfsecJSON := []byte(`{
		"results": [
			{
				"rule_id": "AVD-AWS-0132",
				"severity": "HIGH",
				"description": "Bucket does not encrypt data with a customer managed key.",
				"status": 0,
				"location": {
					"filename": "modules/test/main.tf",
					"start_line": 2,
					"end_line": 8
				}
			}
		]
	}`)

	details := model.NewModuleDetails([]byte{}).WithTfsec(tfsecJSON)
	mv, err := model.NewModuleVersion("1.0.0", details, false)
	require.NoError(t, err)

	results := mv.GetSecurityResults()
	require.Len(t, results, 1)

	// Convert to JSON to verify API format
	jsonOutput, err := json.Marshal(results[0])
	require.NoError(t, err)

	var apiResult map[string]interface{}
	err = json.Unmarshal(jsonOutput, &apiResult)
	require.NoError(t, err)

	// Verify the expected API fields
	assert.Equal(t, "AVD-AWS-0132", apiResult["rule_id"])
	assert.Equal(t, "HIGH", apiResult["severity"])
	assert.Equal(t, "Bucket does not encrypt data with a customer managed key.", apiResult["description"])

	// Verify location is properly nested
	location, ok := apiResult["location"].(map[string]interface{})
	require.True(t, ok, "Location should be a nested object")
	assert.Equal(t, "modules/test/main.tf", location["filename"])
	assert.Equal(t, float64(2), location["start_line"])
	assert.Equal(t, float64(8), location["end_line"])
}

// TestSecurityResultsEmptyDatabase tests handling of empty/invalid JSON
func TestSecurityResultsEmptyDatabase(t *testing.T) {
	tests := []struct {
		name        string
		tfsecJSON   []byte
		expectEmpty bool
	}{
		{
			name:        "nil tfsec",
			tfsecJSON:   nil,
			expectEmpty: true,
		},
		{
			name:        "empty results array",
			tfsecJSON:   []byte(`{"results": []}`),
			expectEmpty: true,
		},
		{
			name:        "null results",
			tfsecJSON:   []byte(`{"results": null}`),
			expectEmpty: true,
		},
		{
			name:        "invalid JSON",
			tfsecJSON:   []byte(`not json`),
			expectEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			details := model.NewModuleDetails([]byte{}).WithTfsec(tt.tfsecJSON)
			mv, err := model.NewModuleVersion("1.0.0", details, false)
			require.NoError(t, err)

			results := mv.GetSecurityResults()
			if tt.expectEmpty {
				assert.Empty(t, results)
			}
			assert.Equal(t, 0, mv.GetSecurityFailures())
		})
	}
}
