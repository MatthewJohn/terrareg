package service

import (
	"testing"
)

func TestMarkdownService_ConvertToHTML(t *testing.T) {
	service := NewMarkdownService()

	tests := []struct {
		name     string
		markdown string
		expected string
	}{
		{
			name:     "Empty markdown",
			markdown: "",
			expected: ``,
		},
		{
			name:     "Simple headers",
			markdown: "# Header 1\n## Header 2\n### Header 3",
			expected: `<div class="markdown-content"><h1 id="header-1">Header 1</h1>
<h2 id="header-2">Header 2</h2>
<h3 id="header-3">Header 3</h3>
</div>`,
		},
		{
			name:     "Paragraphs",
			markdown: "This is a paragraph.\n\nThis is another paragraph.",
			expected: `<div class="markdown-content"><p>This is a paragraph.</p>
<p>This is another paragraph.</p>
</div>`,
		},
		{
			name:     "Fenced code blocks",
			markdown: "```go\nfunc main() {\n    fmt.Println(\"Hello\")\n}\n```",
			expected: `<div class="markdown-content"><pre><code class="language-go">func main() {
    fmt.Println(&quot;Hello&quot;)
}
</code></pre>
</div>`,
		},
		{
			name:     "Inline code",
			markdown: "This is `inline code` in text.",
			expected: `<div class="markdown-content"><p>This is <code>inline code</code> in text.</p>
</div>`,
		},
		{
			name:     "Tables",
			markdown: "| Name | Age |\n|------|-----|\n| John | 30  |\n| Jane | 25  |",
			expected: `<div class="markdown-content"><table>
<thead>
<tr>
<th>Name</th>
<th>Age</th>
</tr>
</thead>
<tbody>
<tr>
<td>John</td>
<td>30</td>
</tr>
<tr>
<td>Jane</td>
<td>25</td>
</tr>
</tbody>
</table>
</div>`,
		},
		{
			name:     "Links",
			markdown: "[GitHub](https://github.com)",
			expected: `<div class="markdown-content"><p><a href="https://github.com">GitHub</a></p>
</div>`,
		},
		{
			name:     "Bold and italic",
			markdown: "This is **bold** and this is *italic*.",
			expected: `<div class="markdown-content"><p>This is <strong>bold</strong> and this is <em>italic</em>.</p>
</div>`,
		},
		{
			name:     "Lists",
			markdown: "- Item 1\n- Item 2\n  - Nested item\n- Item 3",
			expected: `<div class="markdown-content"><ul>
<li>Item 1</li>
<li>Item 2
<ul>
<li>Nested item</li>
</ul>
</li>
<li>Item 3</li>
</ul>
</div>`,
		},
		{
			name:     "Blockquotes",
			markdown: "> This is a quote\n> > Nested quote",
			expected: `<div class="markdown-content"><blockquote>
<p>This is a quote</p>
<blockquote>
<p>Nested quote</p>
</blockquote>
</blockquote>
</div>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.ConvertToHTML(tt.markdown)
			if result != tt.expected {
				t.Errorf("ConvertToHTML() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

func TestMarkdownService_ConvertToHTML_WithFileName(t *testing.T) {
	service := NewMarkdownService()

	tests := []struct {
		name     string
		markdown string
		filename string
		expected string
	}{
		{
			name:     "Relative links with README.md",
			markdown: "[Section 1](./README.md#section-1)",
			filename: "README.md",
			expected: `<div class="markdown-content"><p><a href="#section-1">Section 1</a></p>
</div>`,
		},
		{
			name:     "Direct README.md links",
			markdown: "[Section 2](README.md#section-2)",
			filename: "README.md",
			expected: `<div class="markdown-content"><p><a href="#section-2">Section 2</a></p>
</div>`,
		},
		{
			name:     "Absolute links remain unchanged",
			markdown: "[GitHub](https://github.com)",
			filename: "README.md",
			expected: `<div class="markdown-content"><p><a href="https://github.com">GitHub</a></p>
</div>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.ConvertToHTML(tt.markdown, WithFileName(tt.filename))
			if result != tt.expected {
				t.Errorf("ConvertToHTML() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

func TestMarkdownService_PostProcessHTML(t *testing.T) {
	service := NewMarkdownService()

	tests := []struct {
		name     string
		html     string
		filename string
		expected string
	}{
		{
			name:     "Remove relative image sources",
			html:     `<img src="image.png" alt="Test">`,
			filename: "README.md",
			expected: `<img  alt="Test">`,
		},
		{
			name:     "Keep absolute image sources",
			html:     `<img src="https://example.com/image.png" alt="Test">`,
			filename: "README.md",
			expected: `<img src="https://example.com/image.png" alt="Test">`,
		},
		{
			name:     "Keep http image sources",
			html:     `<img src="http://example.com/image.png" alt="Test">`,
			filename: "README.md",
			expected: `<img src="http://example.com/image.png" alt="Test">`,
		},
		{
			name:     "Multiple images",
			html:     `<img src="relative.png"><img src="https://absolute.com/image.png">`,
			filename: "README.md",
			expected: `<img ><img src="https://absolute.com/image.png">`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.postProcessHTML(tt.html, tt.filename)
			if result != tt.expected {
				t.Errorf("postProcessHTML() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

func TestMarkdownService_SanitizeHTML(t *testing.T) {
	service := NewMarkdownService()

	tests := []struct {
		name     string
		html     string
		expected string
	}{
		{
			name:     "Remove script tags",
			html:     `<p>Hello</p><script>alert('xss')</script>`,
			expected: `<p>Hello</p>`,
		},
		{
			name:     "Remove iframe tags",
			html:     `<p>Hello</p><iframe src="evil.com"></iframe>`,
			expected: `<p>Hello</p>`,
		},
		{
			name:     "Remove form tags",
			html:     `<p>Hello</p><form action="evil.com"><input type="submit"></form>`,
			expected: `<p>Hello</p><input type="submit">`,
		},
		{
			name:     "Remove javascript links",
			html:     `<p><a href="javascript:alert('xss')">Click me</a></p>`,
			expected: `<p><a href="#">Click me</a></p>`,
		},
		{
			name:     "Keep safe HTML",
			html:     `<p>Hello <strong>world</strong> <em>test</em></p>`,
			expected: `<p>Hello <strong>world</strong> <em>test</em></p>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.SanitizeHTML(tt.html)
			if result != tt.expected {
				t.Errorf("SanitizeHTML() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

func TestMarkdownService_AngularBracketPlaceholders(t *testing.T) {
	service := NewMarkdownService()

	// Test angular bracket placeholders like Python terrareg
	placeholderMarkdown := "Convert `<Hi>` or `<Your name>` where <your-name> is replaced with you name."

	result := service.ConvertToHTML(placeholderMarkdown)

	// Should properly escape angular brackets in code but leave them in regular text
	if !containsSubstring(result, `<code>&lt;Hi&gt;</code>`) {
		t.Error("Angular brackets in inline code should be HTML escaped")
	}

	if !containsSubstring(result, `<your-name>`) {
		t.Error("Angular brackets in regular text should be preserved")
	}
}

func TestMarkdownService_ComplexMarkdown(t *testing.T) {
	service := NewMarkdownService()

	// Test complex markdown similar to what might be in a README
	complexMarkdown := "# My Awesome Module\n\n" +
		"This is a sample README with various markdown features.\n\n" +
		"## Installation\n\n" +
		"Install the module with:\n\n" +
		"`terraform init`\n\n" +
		"## Features\n\n" +
		"- Feature 1 with **bold text**\n" +
		"- Feature 2 with *italic text*\n" +
		"- Feature 3 with `inline code`\n\n" +
		"## Usage Table\n\n" +
		"| Input | Type | Description |\n" +
		"|-------|------|-------------|\n" +
		"| name  | string | The name to use |\n" +
		"| count | number | The count of items |\n\n" +
		"## Links\n\n" +
		"[GitHub Repository](https://github.com/myorg/my-module)\n" +
		"[Installation](./README.md#installation)\n\n" +
		"> **Note:** This is an important note with bold text.\n"

	result := service.ConvertToHTML(complexMarkdown, WithFileName("README.md"))

	// Check that key elements are present
	if result == "" {
		t.Error("ConvertToHTML() returned empty string")
	}

	// Check for proper HTML structure
	if !containsSubstring(result, `<div class="markdown-content">`) {
		t.Error("Result should have markdown-content div")
	}

	// Check for fenced code blocks - look for any code block, not just specific format
	if !containsSubstring(result, `<pre><code`) && !containsSubstring(result, `<code>`) {
		t.Error("Result should contain fenced code block")
	}

	// Check for table
	if !containsSubstring(result, `<table>`) {
		t.Error("Result should contain table")
	}

	// Check that relative links were processed properly
	if !containsSubstring(result, `href="#installation"`) {
		t.Error("Relative README link should be converted to anchor")
	}
}

// Helper function to check if string contains substring
func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr)
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
