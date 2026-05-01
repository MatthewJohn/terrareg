package service_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	configmodule "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/config/model"
	configmodel "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/service"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared/types"
)

func TestSecurityService_ValidateFilePath(t *testing.T) {
	securityService := service.NewSecurityService()

	testCases := []struct {
		name        string
		path        string
		expectError bool
	}{
		{
			name:        "Valid relative path",
			path:        "main.tf",
			expectError: false,
		},
		{
			name:        "Valid nested path",
			path:        "modules/example/main.tf",
			expectError: false,
		},
		{
			name:        "Empty path",
			path:        "",
			expectError: true,
		},
		{
			name:        "Path traversal attempt",
			path:        "../etc/passwd",
			expectError: true,
		},
		{
			name:        "Absolute path",
			path:        "/etc/passwd",
			expectError: true,
		},
		{
			name:        "Path with dangerous characters",
			path:        "file<name>.tf",
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := securityService.ValidateFilePath(tc.path)

			if tc.expectError {
				assert.Error(t, err)
				assert.Equal(t, service.ErrInvalidFilePath, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSecurityService_SanitizeContent(t *testing.T) {
	securityService := service.NewSecurityService()

	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Safe content",
			input:    "This is safe content",
			expected: "This is safe content", // Safe content passes through unchanged
		},
		{
			name:     "Script tag",
			input:    "<script>alert('xss')</script>",
			expected: "", // Script tags and their content are completely removed
		},
		{
			name:     "Mixed case dangerous tags",
			input:    "<SCRIPT>alert('xss')</SCRIPT><iframe src='evil'></iframe>",
			expected: "", // All dangerous tags and their content are removed
		},
		{
			name:     "Event handlers",
			input:    "<div onclick=\"alert('xss')\">content</div>",
			expected: "<div>content</div>", // Event handlers are removed
		},
		{
			name:     "JavaScript protocol",
			input:    "<a href=\"javascript:alert('xss')\">link</a>",
			expected: "<a href=\"\">link</a>", // Dangerous protocol is removed from href
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			content := tc.input
			err := securityService.SanitizeContent(&content)

			require.NoError(t, err)
			assert.Equal(t, tc.expected, content)
		})
	}
}

func TestSecurityService_ValidateFileType(t *testing.T) {
	securityService := service.NewSecurityService()

	testCases := []struct {
		name        string
		filename    string
		expectError bool
	}{
		{
			name:        "Valid Terraform file",
			filename:    "main.tf",
			expectError: false,
		},
		{
			name:        "Valid JSON file",
			filename:    "variables.tf.json",
			expectError: false,
		},
		{
			name:        "Valid Markdown file",
			filename:    "README.md",
			expectError: false,
		},
		{
			name:        "Valid YAML file",
			filename:    "values.yml",
			expectError: false,
		},
		{
			name:        "Executable file",
			filename:    "malware.exe",
			expectError: true,
		},
		{
			name:        "File without extension",
			filename:    "noextension",
			expectError: true,
		},
		{
			name:        "Hidden dangerous file",
			filename:    ".bashrc",
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := securityService.ValidateFileType(tc.filename)

			if tc.expectError {
				assert.Error(t, err)
				assert.Equal(t, service.ErrInvalidFileType, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNamespaceService_IsTrusted(t *testing.T) {
	// Create domain config with trusted namespaces
	config := &configmodule.DomainConfig{
		TrustedNamespaces: []string{"hashicorp", "aws-ia"},
	}

	namespaceService := service.NewNamespaceService(config)

	testCases := []struct {
		name        string
		namespace   types.NamespaceName
		expected    bool
		expectError bool
	}{
		{
			name:        "Trusted namespace",
			namespace:   "hashicorp",
			expected:    true,
			expectError: false,
		},
		{
			name:        "Untrusted namespace",
			namespace:   "random-namespace",
			expected:    false,
			expectError: false,
		},
		{
			name:        "Nil namespace",
			namespace:   "",
			expected:    false,
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var namespace *configmodel.Namespace
			var err error

			if tc.namespace != "" {
				namespace, err = configmodel.NewNamespace(tc.namespace, nil, configmodel.NamespaceTypeNone)
				if tc.expectError {
					require.Error(t, err)
					return
				}
				require.NoError(t, err)
			}

			result := namespaceService.IsTrusted(namespace)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestNamespaceService_IsAutoVerified(t *testing.T) {
	// Create domain config with verified namespaces
	config := &configmodule.DomainConfig{
		VerifiedModuleNamespaces: []string{"hashicorp", "terraform-aws-modules"},
	}

	namespaceService := service.NewNamespaceService(config)

	testCases := []struct {
		name        string
		namespace   types.NamespaceName
		expected    bool
		expectError bool
	}{
		{
			name:        "Auto-verified namespace",
			namespace:   "hashicorp",
			expected:    true,
			expectError: false,
		},
		{
			name:        "Regular namespace",
			namespace:   "my-namespace",
			expected:    false,
			expectError: false,
		},
		{
			name:        "Nil namespace",
			namespace:   "",
			expected:    false,
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var namespace *configmodel.Namespace
			var err error

			if tc.namespace != "" {
				namespace, err = configmodel.NewNamespace(tc.namespace, nil, configmodel.NamespaceTypeNone)
				if tc.expectError {
					require.Error(t, err)
					return
				}
				require.NoError(t, err)
			}

			result := namespaceService.IsAutoVerified(namespace)
			assert.Equal(t, tc.expected, result)
		})
	}
}
