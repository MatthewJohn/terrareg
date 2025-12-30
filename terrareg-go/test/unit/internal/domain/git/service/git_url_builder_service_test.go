package service_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	gitmodel "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/git/model"
	gitservice "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/git/service"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared"
)

func TestGitURLBuilderService_BuildCloneURL(t *testing.T) {
	gitService := gitservice.NewGitURLBuilderService()

	tests := []struct {
		name        string
		request     *gitservice.URLBuilderRequest
		expectedURL string
		expectError bool
	}{
		{
			name: "Basic template substitution",
			request: &gitservice.URLBuilderRequest{
				Template:  "https://github.com/{namespace}/{module}.git",
				Namespace: "test-org",
				Module:    "test-module",
				Provider:  "aws",
			},
			expectedURL: "https://github.com/test-org/test-module.git",
			expectError: false,
		},
		{
			name: "Template with all placeholders",
			request: &gitservice.URLBuilderRequest{
				Template:  "https://gitlab.com/{namespace}/{module}-{provider}.git",
				Namespace: "mycompany",
				Module:    "terraform",
				Provider:  "aws",
				GitTag:    stringPtr("v1.0.0"),
				Version:   shared.NewVersion(1, 0, 0),
			},
			expectedURL: "https://gitlab.com/mycompany/terraform-aws.git",
			expectError: false,
		},
		{
			name: "Template with git tag substitution",
			request: &gitservice.URLBuilderRequest{
				Template:  "https://github.com/{namespace}/{module}.git?ref={git_tag}",
				Namespace: "hashicorp",
				Module:    "terraform",
				Provider:  "aws",
				GitTag:    stringPtr("v2.0.0"),
			},
			expectedURL: "https://github.com/hashicorp/terraform.git?ref=v2.0.0",
			expectError: false,
		},
		{
			name: "Template with version substitution",
			request: &gitservice.URLBuilderRequest{
				Template:  "https://github.com/{namespace}/{module}-provider/archive/v{version}.tar.gz",
				Namespace: "terraform-aws-modules",
				Module:    "vpc",
				Provider:  "aws",
				Version:   shared.NewVersion(5, 1, 0),
			},
			expectedURL: "https://github.com/terraform-aws-modules/vpc-provider/archive/v5.1.0.tar.gz",
			expectError: false,
		},
		{
			name: "SSH template",
			request: &gitservice.URLBuilderRequest{
				Template:  "git@github.com:{namespace}/{module}.git",
				Namespace: "org",
				Module:    "repo",
				Provider:  "provider",
			},
			expectedURL: "git@github.com:org/repo.git",
			expectError: false,
		},
		{
			name: "Invalid template",
			request: &gitservice.URLBuilderRequest{
				Template:  "https://github.com/{namespace}/{invalid}.git",
				Namespace: "test",
				Module:    "module",
				Provider:  "provider",
			},
			expectedURL: "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url, err := gitService.BuildCloneURL(tt.request)

			if tt.expectError {
				assert.Error(t, err)
				assert.Empty(t, url)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedURL, url)
			}
		})
	}
}

func TestGitURLBuilderService_BuildCloneURL_WithCredentials(t *testing.T) {
	gitService := gitservice.NewGitURLBuilderService()

	tests := []struct {
		name        string
		request     *gitservice.URLBuilderRequest
		expectedURL string
	}{
		{
			name: "HTTPS with username only",
			request: &gitservice.URLBuilderRequest{
				Template:  "https://github.com/{namespace}/{module}.git",
				Namespace: "org",
				Module:    "repo",
				Provider:  "provider",
				Credentials: &gitmodel.GitCredentials{
					Username: "user",
				},
			},
			expectedURL: "https://user@github.com/org/repo.git",
		},
		{
			name: "HTTPS with username and password",
			request: &gitservice.URLBuilderRequest{
				Template:  "https://github.com/{namespace}/{module}.git",
				Namespace: "org",
				Module:    "repo",
				Provider:  "provider",
				Credentials: &gitmodel.GitCredentials{
					Username: "user",
					Password: "pass",
				},
			},
			expectedURL: "https://user:pass@github.com/org/repo.git",
		},
		{
			name: "HTTPS with credentials containing special characters",
			request: &gitservice.URLBuilderRequest{
				Template:  "https://github.com/{namespace}/{module}.git",
				Namespace: "org",
				Module:    "repo",
				Provider:  "provider",
				Credentials: &gitmodel.GitCredentials{
					Username: "user@email.com",
					Password: "pass word",
				},
			},
			expectedURL: "https://user%40email.com:pass%20word@github.com/org/repo.git",
		},
		{
			name: "SSH with credentials (should not inject)",
			request: &gitservice.URLBuilderRequest{
				Template:  "git@github.com:{namespace}/{module}.git",
				Namespace: "org",
				Module:    "repo",
				Provider:  "provider",
				Credentials: &gitmodel.GitCredentials{
					Username: "user",
					Password: "pass",
				},
			},
			expectedURL: "git@github.com:org/repo.git",
		},
		{
			name: "No credentials",
			request: &gitservice.URLBuilderRequest{
				Template:    "https://github.com/{namespace}/{module}.git",
				Namespace:   "org",
				Module:      "repo",
				Provider:    "provider",
				Credentials: &gitmodel.GitCredentials{},
			},
			expectedURL: "https://github.com/org/repo.git",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url, err := gitService.BuildCloneURL(tt.request)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedURL, url)
		})
	}
}

func TestGitURLBuilderService_BuildBrowseURL(t *testing.T) {
	gitService := gitservice.NewGitURLBuilderService()

	tests := []struct {
		name           string
		template       string
		namespace      string
		module         string
		provider       string
		path           string
		expectedResult string
		expectError    bool
	}{
		{
			name:           "Basic browse URL",
			template:       "https://github.com/{namespace}/{module}",
			namespace:      "org",
			module:         "repo",
			provider:       "provider",
			path:           "",
			expectedResult: "https://github.com/org/repo",
			expectError:    false,
		},
		{
			name:           "Browse URL with path",
			template:       "https://github.com/{namespace}/{module}/tree/main",
			namespace:      "org",
			module:         "repo",
			provider:       "provider",
			path:           "/src/main.tf",
			expectedResult: "https://github.com/org/repo/tree/main/src/main.tf",
			expectError:    false,
		},
		{
			name:           "Browse URL with path starting without slash",
			template:       "https://github.com/{namespace}/{module}/blob/main",
			namespace:      "org",
			module:         "repo",
			provider:       "provider",
			path:           "README.md",
			expectedResult: "https://github.com/org/repo/blob/main/README.md",
			expectError:    false,
		},
		{
			name:           "Invalid template",
			template:       "https://github.com/{namespace}/{invalid}",
			namespace:      "org",
			module:         "repo",
			provider:       "provider",
			path:           "",
			expectedResult: "",
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := gitService.BuildBrowseURL(tt.template, tt.namespace, tt.module, tt.provider, tt.path)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, *result)
			}
		})
	}
}

func TestGitURLBuilderService_BuildArchiveURL(t *testing.T) {
	gitService := gitservice.NewGitURLBuilderService()

	tests := []struct {
		name           string
		template       string
		namespace      string
		module         string
		provider       string
		version        *shared.Version
		expectedResult string
		expectError    bool
	}{
		{
			name:           "GitHub archive URL",
			template:       "https://github.com/{namespace}/{module}/archive/v{version}.tar.gz",
			namespace:      "hashicorp",
			module:         "terraform",
			provider:       "aws",
			version:        shared.NewVersion(1, 5, 7),
			expectedResult: "https://github.com/hashicorp/terraform/archive/v1.5.7.tar.gz",
			expectError:    false,
		},
		{
			name:           "GitLab archive URL",
			template:       "https://gitlab.com/{namespace}/{module}-provider/-/archive/v{version}.tar.gz",
			namespace:      "terraform-aws-modules",
			module:         "vpc",
			provider:       "aws",
			version:        shared.NewVersion(3, 0, 0),
			expectedResult: "https://gitlab.com/terraform-aws-modules/vpc-provider/-/archive/v3.0.0.tar.gz",
			expectError:    false,
		},
		{
			name:           "Invalid template",
			template:       "https://github.com/{namespace}/{invalid}/archive/v{version}.tar.gz",
			namespace:      "org",
			module:         "repo",
			provider:       "provider",
			version:        shared.NewVersion(1, 0, 0),
			expectedResult: "",
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := gitService.BuildArchiveURL(tt.template, tt.namespace, tt.module, tt.provider, tt.version)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, *result)
			}
		})
	}
}

func TestGitURLBuilderService_ValidateTemplate(t *testing.T) {
	gitService := gitservice.NewGitURLBuilderService()

	tests := []struct {
		name        string
		template    string
		expectError bool
	}{
		{
			name:        "Valid template",
			template:    "https://github.com/{namespace}/{module}.git",
			expectError: false,
		},
		{
			name:        "Valid template with all placeholders",
			template:    "https://github.com/{namespace}/{module}-{provider}/archive/v{version}.tar.gz",
			expectError: false,
		},
		{
			name:        "Invalid placeholder",
			template:    "https://github.com/{namespace}/{invalid}.git",
			expectError: true,
		},
		{
			name:        "Empty template",
			template:    "",
			expectError: true,
		},
		{
			name:        "Template with malformed placeholder",
			template:    "https://github.com/{namespace/module.git",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := gitService.ValidateTemplate(tt.template)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGitURLBuilderService_ParseTemplateVariables(t *testing.T) {
	gitService := gitservice.NewGitURLBuilderService()

	tests := []struct {
		name              string
		template          string
		expectedVariables []string
		expectError       bool
	}{
		{
			name:              "Template with single variable",
			template:          "https://github.com/{namespace}/repo.git",
			expectedVariables: []string{"{namespace}"},
			expectError:       false,
		},
		{
			name:              "Template with multiple variables",
			template:          "https://github.com/{namespace}/{module}-{provider}.git",
			expectedVariables: []string{"{namespace}", "{module}", "{provider}"},
			expectError:       false,
		},
		{
			name:              "Template with all variables",
			template:          "https://github.com/{namespace}/{module}/{provider}/archive/v{version}.tar.gz?ref={git_tag}",
			expectedVariables: []string{"{namespace}", "{module}", "{provider}", "{version}", "{git_tag}"},
			expectError:       false,
		},
		{
			name:              "Template with no variables",
			template:          "https://github.com/fixed/repo.git",
			expectedVariables: []string{},
			expectError:       false,
		},
		{
			name:              "Invalid template",
			template:          "https://github.com/{namespace}/{invalid}.git",
			expectedVariables: nil,
			expectError:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			variables, err := gitService.ParseTemplateVariables(tt.template)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, variables)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedVariables, variables)
			}
		})
	}
}

func TestGitURLBuilderService_IsSSHTemplate(t *testing.T) {
	gitService := gitservice.NewGitURLBuilderService()

	tests := []struct {
		name     string
		template string
		expected bool
	}{
		{
			name:     "SSH URL format",
			template: "git@github.com:org/repo.git",
			expected: true,
		},
		{
			name:     "SSH protocol",
			template: "ssh://git@github.com/org/repo.git",
			expected: true,
		},
		{
			name:     "HTTPS URL",
			template: "https://github.com/org/repo.git",
			expected: false,
		},
		{
			name:     "HTTP URL",
			template: "http://github.com/org/repo.git",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, gitService.IsSSHTemplate(tt.template))
		})
	}
}

func TestGitURLBuilderService_RequiresCredentials(t *testing.T) {
	gitService := gitservice.NewGitURLBuilderService()

	tests := []struct {
		name     string
		template string
		expected bool
	}{
		{
			name:     "HTTPS URL",
			template: "https://github.com/org/repo.git",
			expected: true,
		},
		{
			name:     "HTTP URL",
			template: "http://github.com/org/repo.git",
			expected: true,
		},
		{
			name:     "SSH URL",
			template: "git@github.com:org/repo.git",
			expected: false,
		},
		{
			name:     "SSH protocol",
			template: "ssh://git@github.com/org/repo.git",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, gitService.RequiresCredentials(tt.template))
		})
	}
}

func TestGitURLBuilderService_IntegrationTest(t *testing.T) {
	gitService := gitservice.NewGitURLBuilderService()

	// Test building URL with all options set
	req := &gitservice.URLBuilderRequest{
		Template:  "https://github.com/{namespace}/{module}-{provider}.git",
		Namespace: "terraform-aws-modules",
		Module:    "vpc",
		Provider:  "aws",
		GitTag:    stringPtr("v3.0.0"),
		Version:   shared.NewVersion(3, 0, 0),
		Credentials: &gitmodel.GitCredentials{
			Username: "git-user",
			Password: "secure-pass-123",
		},
	}

	url, err := gitService.BuildCloneURL(req)
	require.NoError(t, err)

	expectedURL := "https://git-user:secure-pass-123@github.com/terraform-aws-modules/vpc-aws.git"
	assert.Equal(t, expectedURL, url)

	// Validate template
	err = gitService.ValidateTemplate(req.Template)
	assert.NoError(t, err)

	// Parse variables
	variables, err := gitService.ParseTemplateVariables(req.Template)
	require.NoError(t, err)
	assert.Equal(t, []string{"{namespace}", "{module}", "{provider}"}, variables)

	// Check template properties
	assert.False(t, gitService.IsSSHTemplate(req.Template))
	assert.True(t, gitService.RequiresCredentials(req.Template))
}

func TestURLTemplate_Model(t *testing.T) {
	tests := []struct {
		name        string
		templateStr string
		expected    *gitmodel.URLTemplate
	}{
		{
			name:        "Valid template",
			templateStr: "https://github.com/{namespace}/{module}.git",
			expected: &gitmodel.URLTemplate{
				Raw:     "https://github.com/{namespace}/{module}.git",
				IsValid: true,
			},
		},
		{
			name:        "Invalid template",
			templateStr: "https://github.com/{namespace}/{invalid}.git",
			expected: &gitmodel.URLTemplate{
				Raw:     "https://github.com/{namespace}/{invalid}.git",
				IsValid: false,
			},
		},
		{
			name:        "Empty template",
			templateStr: "",
			expected: &gitmodel.URLTemplate{
				Raw:     "",
				IsValid: false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			template := gitmodel.NewURLTemplate(tt.templateStr)
			assert.Equal(t, tt.expected.Raw, template.Raw)
			assert.Equal(t, tt.expected.IsValid, template.IsValid)
		})
	}
}

func TestURLTemplate_GetPlaceholders(t *testing.T) {
	tests := []struct {
		name                 string
		template             string
		expectedPlaceholders []string
	}{
		{
			name:                 "Single placeholder",
			template:             "https://github.com/{namespace}/repo.git",
			expectedPlaceholders: []string{"{namespace}"},
		},
		{
			name:                 "Multiple placeholders",
			template:             "https://github.com/{namespace}/{module}-{provider}/archive/v{version}.tar.gz",
			expectedPlaceholders: []string{"{namespace}", "{module}", "{provider}", "{version}"},
		},
		{
			name:                 "Duplicate placeholders",
			template:             "https://github.com/{namespace}/{module}/{namespace}/tree/{version}",
			expectedPlaceholders: []string{"{namespace}", "{module}", "{version}"},
		},
		{
			name:                 "No placeholders",
			template:             "https://github.com/fixed/repo.git",
			expectedPlaceholders: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			template := gitmodel.NewURLTemplate(tt.template)
			assert.Equal(t, tt.expectedPlaceholders, template.GetPlaceholders())
		})
	}
}

// Helper function
func stringPtr(s string) *string {
	return &s
}
