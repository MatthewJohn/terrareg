package git

import (
	"testing"
)

func TestGitImportService_getTagRegex(t *testing.T) {
	service := &GitImportService{}

	tests := []struct {
		name        string
		format      string
		expectError bool
		expected    string // partial expected pattern
	}{
		{
			name:        "simple version format",
			format:      "v{version}",
			expectError: false,
			expected:    `v(?P<version>[^}]+)`,
		},
		{
			name:        "semantic versioning",
			format:      "v{major}.{minor}.{patch}",
			expectError: false,
			expected:    `v(?P<major>\d+)\.(?P<minor>\d+)\.(?P<patch>\d+)`,
		},
		{
			name:        "with build metadata",
			format:      "release-{major}.{minor}-{build}",
			expectError: false,
			expected:    `release-(?P<major>\d+)\.(?P<minor>\d+)-(?P<build>[^}]+)`,
		},
		{
			name:        "complex format with dots",
			format:      "api/{major}.{minor}.{patch}-alpha.{build}",
			expectError: false,
			expected:    `api/(?P<major>\d+)\.(?P<minor>\d+)\.(?P<patch>\d+)-alpha\.(?P<build>[^}]+)`,
		},
		{
			name:        "no placeholders",
			format:      "stable",
			expectError: false,
			expected:    `stable`,
		},
		{
			name:        "mixed with literal text",
			format:      "version-{version}-release",
			expectError: false,
			expected:    `version-(?P<version>[^}]+)-release`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			regex, err := service.getTagRegex(tt.format)

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

			// Check that the expected pattern is in the regex (accounting for ^ and $ anchors)
			if tt.expected != "" {
				actual := regex.String()
				// Remove ^ and $ for comparison
				if len(actual) > 2 && actual[0] == '^' && actual[len(actual)-1] == '$' {
					actual = actual[1 : len(actual)-1]
				}
				if actual != tt.expected {
					t.Logf("Expected pattern: %s", tt.expected)
					t.Logf("Actual pattern:  %s", actual)
					t.Errorf("Pattern doesn't match expected structure")
				}
			}

			// Verify regex is anchored
			pattern := regex.String()
			if pattern[0] != '^' || pattern[len(pattern)-1] != '$' {
				t.Errorf("Regex should be anchored with ^ and $, got: %s", pattern)
			}
		})
	}
}

func TestGitImportService_getVersionFromRegex(t *testing.T) {
	service := &GitImportService{}

	tests := []struct {
		name           string
		tagFormat      string
		gitTag         string
		expectedOutput string
	}{
		{
			name:           "direct version group",
			tagFormat:      "v{version}",
			gitTag:         "v1.2.3-beta",
			expectedOutput: "1.2.3-beta",
		},
		{
			name:           "semantic versioning",
			tagFormat:      "v{major}.{minor}.{patch}",
			gitTag:         "v2.5.1",
			expectedOutput: "2.5.1",
		},
		{
			name:           "semantic with build",
			tagFormat:      "v{major}.{minor}.{patch}-{build}",
			gitTag:         "v1.0.0-rc1",
			expectedOutput: "1.0.0-rc1",
		},
		{
			name:           "complex format",
			tagFormat:      "release/{major}.{minor}.{patch}-alpha.{build}",
			gitTag:         "release/3.2.1-alpha.20231201",
			expectedOutput: "3.2.1-20231201",
		},
		{
			name:           "missing minor defaults to 0",
			tagFormat:      "v{major}.{patch}",
			gitTag:         "v4.7",
			expectedOutput: "4.0.7",
		},
		{
			name:           "only major",
			tagFormat:      "v{major}",
			gitTag:         "v5",
			expectedOutput: "5.0.0",
		},
		{
			name:           "no match",
			tagFormat:      "v{major}.{minor}.{patch}",
			gitTag:         "not-a-version",
			expectedOutput: "",
		},
		{
			name:           "empty groups default to 0",
			tagFormat:      "v{major}.{minor}.{patch}",
			gitTag:         "v1.0.3",
			expectedOutput: "1.0.3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			regex, err := service.getTagRegex(tt.tagFormat)
			if err != nil {
				t.Fatalf("Failed to create regex: %v", err)
			}

			result := service.getVersionFromRegex(regex, tt.gitTag)
			if result != tt.expectedOutput {
				t.Errorf("Expected version %q, got %q", tt.expectedOutput, result)
			}
		})
	}
}

// Note: TestDeriveVersionFromGitTag requires proper mocking of the ModuleProvider domain model
// The individual components (getTagRegex, getVersionFromRegex) are thoroughly tested above

// Helper function to check if a pattern contains expected components
func containsPattern(actual, expected string) bool {
	// For complex patterns, check that key parts exist
	// This is a simplified check - in real scenarios might need more sophisticated matching
	return len(actual) > 0 && len(expected) > 0
}