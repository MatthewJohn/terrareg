package model

import (
	"testing"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/model"
	"github.com/stretchr/testify/assert"
)

// TestModuleVersionFile_NewModuleVersionFile tests creating module version files
func TestModuleVersionFile_NewModuleVersionFile(t *testing.T) {
	t.Run("terraform file", func(t *testing.T) {
		file := model.NewModuleVersionFile(1, nil, "main.tf", "resource \"test\" \"test\" {}")
		assert.Equal(t, "main.tf", file.FileName())
		assert.Equal(t, "main.tf", file.Path())
		assert.Equal(t, "text/plain", file.ContentType())
		assert.Equal(t, "resource \"test\" \"test\" {}", file.Content())
		assert.False(t, file.IsMarkdown())
	})

	t.Run("markdown file", func(t *testing.T) {
		file := model.NewModuleVersionFile(1, nil, "README.md", "# README")
		assert.Equal(t, "README.md", file.FileName())
		assert.Equal(t, "README.md", file.Path())
		assert.Equal(t, "text/markdown", file.ContentType())
		assert.Equal(t, "# README", file.Content())
		assert.True(t, file.IsMarkdown())
	})

	t.Run("json file", func(t *testing.T) {
		file := model.NewModuleVersionFile(1, nil, "data.json", `{"key": "value"}`)
		assert.Equal(t, "data.json", file.FileName())
		assert.Equal(t, "application/json", file.ContentType())
	})

	t.Run("yaml file", func(t *testing.T) {
		file := model.NewModuleVersionFile(1, nil, "config.yaml", "key: value")
		assert.Equal(t, "config.yaml", file.FileName())
		assert.Equal(t, "application/x-yaml", file.ContentType())
	})

	t.Run("yml file", func(t *testing.T) {
		file := model.NewModuleVersionFile(1, nil, "config.yml", "key: value")
		assert.Equal(t, "config.yml", file.FileName())
		assert.Equal(t, "application/x-yaml", file.ContentType())
	})

	t.Run("file with path", func(t *testing.T) {
		file := model.NewModuleVersionFile(1, nil, "path/to/file.tf", "content")
		assert.Equal(t, "file.tf", file.FileName())
		assert.Equal(t, "path/to/file.tf", file.Path())
	})
}

// TestModuleVersionFile_ValidatePath tests path validation
func TestModuleVersionFile_ValidatePath(t *testing.T) {
	testCases := []struct {
		name      string
		path      string
		wantError bool
	}{
		{"valid path", "main.tf", false},
		{"nested path", "path/to/main.tf", false},
		{"deeply nested", "a/b/c/d/e/f/file.tf", false},
		{"empty path", "", true},
		{"path traversal", "../main.tf", true},
		{"traversal in middle", "path/../main.tf", true},
		{"absolute path", "/absolute/path.tf", true},
		{"traversal at end", "path/..", true},
		{"complex traversal", "path/../../etc/passwd", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			file := model.NewModuleVersionFile(1, nil, tc.path, "content")
			err := file.ValidatePath()
			if tc.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestModuleVersionFile_ID tests ID getter
func TestModuleVersionFile_ID(t *testing.T) {
	file := model.NewModuleVersionFile(42, nil, "test.tf", "content")
	assert.Equal(t, 42, file.ID())
}

// TestModuleVersionFile_FileName tests file name extraction
func TestModuleVersionFile_FileName(t *testing.T) {
	testCases := []struct {
		path         string
		expectedName string
	}{
		{"main.tf", "main.tf"},
		{"path/to/main.tf", "main.tf"},
		{"deeply/nested/path/file.tf", "file.tf"},
		{"README.md", "README.md"},
		{".hidden", ".hidden"},
		{"file", "file"},
	}

	for _, tc := range testCases {
		t.Run(tc.path, func(t *testing.T) {
			file := model.NewModuleVersionFile(1, nil, tc.path, "content")
			assert.Equal(t, tc.expectedName, file.FileName())
		})
	}
}

// TestModuleVersionFile_IsMarkdown tests markdown detection
func TestModuleVersionFile_IsMarkdown(t *testing.T) {
	testCases := []struct {
		path     string
		expected bool
	}{
		{"README.md", true},
		{"readme.md", true},
		{"path/to/file.md", true},
		{"path/to/file.MD", true},
		{"path/to/file.Md", true},
		{"main.tf", false},
		{"data.json", false},
		{"config.yaml", false},
		{"file.txt", false},
		{"file", false},
	}

	for _, tc := range testCases {
		t.Run(tc.path, func(t *testing.T) {
			file := model.NewModuleVersionFile(1, nil, tc.path, "content")
			assert.Equal(t, tc.expected, file.IsMarkdown())
		})
	}
}

// TestModuleVersionFile_ContentType tests content type detection
func TestModuleVersionFile_ContentType(t *testing.T) {
	testCases := []struct {
		path            string
		expectedType    string
	}{
		{"README.md", "text/markdown"},
		{"file.md", "text/markdown"},
		{"file.MD", "text/markdown"},
		{"data.json", "application/json"},
		{"file.json", "application/json"},
		{"config.yaml", "application/x-yaml"},
		{"config.yml", "application/x-yaml"},
		{"main.tf", "text/plain"},
		{"variables.tf", "text/plain"},
		{"file.txt", "text/plain"},
		{"LICENSE", "text/plain"},
		{"file", "text/plain"},
	}

	for _, tc := range testCases {
		t.Run(tc.path, func(t *testing.T) {
			file := model.NewModuleVersionFile(1, nil, tc.path, "content")
			assert.Equal(t, tc.expectedType, file.ContentType())
		})
	}
}

// TestModuleVersionFile_Content tests content getter
func TestModuleVersionFile_Content(t *testing.T) {
	content := "resource \"test\" \"test\" {}"
	file := model.NewModuleVersionFile(1, nil, "main.tf", content)
	assert.Equal(t, content, file.Content())
}

// TestModuleVersionFile_Path tests path getter
func TestModuleVersionFile_Path(t *testing.T) {
	path := "path/to/main.tf"
	file := model.NewModuleVersionFile(1, nil, path, "content")
	assert.Equal(t, path, file.Path())
}
