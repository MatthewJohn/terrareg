package mocks

import (
	"encoding/json"
	"fmt"
)

// TfsecMock provides a mock implementation of tfsec security scanner output
// This is used for testing module security scanning without requiring the actual tfsec binary
type TfsecMock struct {
	ExpectedResults []SecurityIssue
	ShouldError     bool
	ErrorMessage    string
}

// SecurityIssue represents a security issue found by tfsec
type SecurityIssue struct {
	RuleID      string        `json:"rule_id"`
	Severity    string        `json:"severity"`
	Title       string        `json:"title"`
	Description string        `json:"description"`
	Impact      string        `json:"impact"`
	Resolution  string        `json:"resolution"`
	Links       []string      `json:"links,omitempty"`
	Location    SecurityLocation `json:"location"`
}

// SecurityLocation represents the location of a security issue
type SecurityLocation struct {
	Filename  string `json:"filename"`
	StartLine int    `json:"start_line"`
	EndLine   int    `json:"end_line"`
}

// TfsecOutput represents the full tfsec output format
type TfsecOutput struct {
	Results []SecurityIssue `json:"results"`
	Metadata TfsecMetadata  `json:"metadata"`
}

// TfsecMetadata represents metadata from tfsec scan
type TfsecMetadata struct {
	TFSecVersion    string `json:"tfsec_version"`
	TerraformVersion string `json:"terraform_version,omitempty"`
}

// Scan returns the tfsec output as JSON or an error
func (m *TfsecMock) Scan() ([]byte, error) {
	if m.ShouldError {
		if m.ErrorMessage != "" {
			return nil, fmt.Errorf("tfsec mock error: %s", m.ErrorMessage)
		}
		return nil, fmt.Errorf("tfsec mock error")
	}

	output := TfsecOutput{
		Results: m.ExpectedResults,
		Metadata: TfsecMetadata{
			TFSecVersion: "v1.28.0",
		},
	}

	jsonOutput, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal tfsec output: %w", err)
	}

	return jsonOutput, nil
}

// SetResults sets the expected security issues for the mock
func (m *TfsecMock) SetResults(results []SecurityIssue) {
	m.ExpectedResults = results
}

// SetError configures the mock to return an error
func (m *TfsecMock) SetError(message string) {
	m.ShouldError = true
	m.ErrorMessage = message
}

// ClearError clears the error state
func (m *TfsecMock) ClearError() {
	m.ShouldError = false
	m.ErrorMessage = ""
}

// NewTfsecMock creates a new mock with default test data
func NewTfsecMock() *TfsecMock {
	return &TfsecMock{
		ExpectedResults: []SecurityIssue{
			{
				RuleID:      "AWS018",
				Severity:    "MEDIUM",
				Title:       "S3 Bucket should have versioning enabled",
				Description: "S3 Bucket versioning keeps a variant of the S3 object when it is modified",
				Impact:     "Objects can be overwritten or deleted without a way to recover them",
				Resolution: "Enable versioning on the S3 bucket",
				Links:      []string{"https://tfsec.dev/docs/aws/s3/bucket-versioning/"},
				Location: SecurityLocation{
					Filename:  "main.tf",
					StartLine: 15,
					EndLine:   18,
				},
			},
		},
		ShouldError: false,
	}
}

// NewCleanTfsecMock creates a mock with no security issues
func NewCleanTfsecMock() *TfsecMock {
	return &TfsecMock{
		ExpectedResults: []SecurityIssue{},
		ShouldError:     false,
	}
}

// NewErrorTfsecMock creates a mock that always returns an error
func NewErrorTfsecMock(message string) *TfsecMock {
	return &TfsecMock{
		ShouldError:  true,
		ErrorMessage: message,
	}
}

// Helper functions to create common security issue outputs

// CreateCriticalSecurityIssues creates security issues with critical severity
func CreateCriticalSecurityIssues() []SecurityIssue {
	return []SecurityIssue{
		{
			RuleID:      "AWS046",
			Severity:    "CRITICAL",
			Title:       "S3 Bucket should not have ACLs that allow public access",
			Description: "S3 Bucket ACLs should not grant public read access",
			Impact:     "Sensitive data could be exposed publicly",
			Resolution: "Remove public access from the S3 bucket ACL",
			Links:      []string{"https://tfsec.dev/docs/aws/s3/bucket-acl-grants/"},
			Location: SecurityLocation{
				Filename:  "storage.tf",
				StartLine: 10,
				EndLine:   25,
			},
		},
		{
			RuleID:      "AWS093",
			Severity:    "CRITICAL",
			Title:       "S3 Bucket should encrypt data at rest",
			Description: "Unencrypted S3 buckets expose sensitive data",
			Impact:     "Data stored in the bucket is not encrypted",
			Resolution: "Enable default encryption on the S3 bucket",
			Links:      []string{"https://tfsec.dev/docs/aws/s3/bucket-encryption/"},
			Location: SecurityLocation{
				Filename:  "storage.tf",
				StartLine: 10,
				EndLine:   25,
			},
		},
	}
}

// CreateHighSecurityIssues creates security issues with high severity
func CreateHighSecurityIssues() []SecurityIssue {
	return []SecurityIssue{
		{
			RuleID:      "AWS017",
			Severity:    "HIGH",
			Title:       "S3 Bucket should have a logging configuration",
			Description: "S3 access logging provides visibility into who is accessing your data",
			Impact:     "Without logging, it is difficult to detect unauthorized access",
			Resolution: "Enable access logging on the S3 bucket",
			Links:      []string{"https://tfsec.dev/docs/aws/s3/bucket-logging/"},
			Location: SecurityLocation{
				Filename:  "storage.tf",
				StartLine: 10,
				EndLine:   25,
			},
		},
		{
			RuleID:      "AWS057",
			Severity:    "HIGH",
			Title:       "IAM user should not have inline policies",
			Description: "Inline policies are difficult to manage and audit",
			Impact:     "Policy management becomes complex and error-prone",
			Resolution: "Use managed policies instead of inline policies",
			Links:      []string{"https://tfsec.dev/docs/aws/iam/user-inline-policy/"},
			Location: SecurityLocation{
				Filename:  "iam.tf",
				StartLine: 5,
				EndLine:   20,
			},
		},
	}
}

// CreateMediumSecurityIssues creates security issues with medium severity
func CreateMediumSecurityIssues() []SecurityIssue {
	return []SecurityIssue{
		{
			RuleID:      "AWS018",
			Severity:    "MEDIUM",
			Title:       "S3 Bucket should have versioning enabled",
			Description: "S3 Bucket versioning keeps a variant of the S3 object when it is modified",
			Impact:     "Objects can be overwritten or deleted without a way to recover them",
			Resolution: "Enable versioning on the S3 bucket",
			Links:      []string{"https://tfsec.dev/docs/aws/s3/bucket-versioning/"},
			Location: SecurityLocation{
				Filename:  "storage.tf",
				StartLine: 10,
				EndLine:   25,
			},
		},
		{
			RuleID:      "AWS019",
			Severity:    "MEDIUM",
			Title:       "S3 Bucket should have default server-side encryption",
			Description: "S3 buckets should have server-side encryption enabled by default",
			Impact:     "Data stored in the bucket is not encrypted by default",
			Resolution: "Enable default server-side encryption on the S3 bucket",
			Links:      []string{"https://tfsec.dev/docs/aws/s3/bucket-default-encryption/"},
			Location: SecurityLocation{
				Filename:  "storage.tf",
				StartLine: 10,
				EndLine:   25,
			},
		},
		{
			RuleID:      "AWS085",
			Severity:    "MEDIUM",
			Title:       "S3 Bucket Access Control should not set 'AllUsers' or 'AuthorizedUsers' to READ",
			Description: "S3 bucket access control should not grant read access to all users",
			Impact:     "Bucket contents could be publicly readable",
			Resolution: "Remove public read access from the bucket policy",
			Links:      []string{"https://tfsec.dev/docs/aws/s3/bucket-acl-grants/"},
			Location: SecurityLocation{
				Filename:  "storage.tf",
				StartLine: 30,
				EndLine:   45,
			},
		},
	}
}

// CreateLowSecurityIssues creates security issues with low severity
func CreateLowSecurityIssues() []SecurityIssue {
	return []SecurityIssue{
		{
			RuleID:      "AWS025",
			Severity:    "LOW",
			Title:       "S3 Bucket should not have a public access policy",
			Description: "S3 buckets should not have a policy that allows public access",
			Impact:     "Bucket contents could be publicly accessible",
			Resolution: "Remove public access from the bucket policy",
			Links:      []string{"https://tfsec.dev/docs/aws/s3/bucket-public-read-access/"},
			Location: SecurityLocation{
				Filename:  "storage.tf",
				StartLine: 50,
				EndLine:   65,
			},
		},
		{
			RuleID:      "GEN001",
			Severity:    "LOW",
			Title:       "Missing required provider version",
			Description: "Provider versions should be pinned",
			Impact:     "Inconsistent provider versions can cause unexpected behavior",
			Resolution: "Pin provider versions in the required_providers block",
			Links:      []string{"https://tfsec.dev/docs/general/required-provider-version/"},
			Location: SecurityLocation{
				Filename:  "main.tf",
				StartLine: 1,
				EndLine:   10,
			},
		},
	}
}

// CreateMixedSecurityIssues creates security issues with mixed severity levels
func CreateMixedSecurityIssues() []SecurityIssue {
	issues := []SecurityIssue{}
	issues = append(issues, CreateCriticalSecurityIssues()...)
	issues = append(issues, CreateHighSecurityIssues()...)
	issues = append(issues, CreateMediumSecurityIssues()[0:1]...)
	issues = append(issues, CreateLowSecurityIssues()[0:1]...)
	return issues
}

// CreateSecurityIssue creates a single security issue with given parameters
func CreateSecurityIssue(ruleID, severity, title, description, filename string, startLine, endLine int) SecurityIssue {
	return SecurityIssue{
		RuleID:      ruleID,
		Severity:    severity,
		Title:       title,
		Description: description,
		Impact:     "See description for impact",
		Resolution: "See description for resolution",
		Links:      []string{fmt.Sprintf("https://tfsec.dev/docs/aws/%s/", ruleID)},
		Location: SecurityLocation{
			Filename:  filename,
			StartLine: startLine,
			EndLine:   endLine,
		},
	}
}
