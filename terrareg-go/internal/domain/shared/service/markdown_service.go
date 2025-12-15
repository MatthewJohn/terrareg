package service

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

// MarkdownService provides markdown to HTML conversion functionality
type MarkdownService struct {
	converter goldmark.Markdown
}

// NewMarkdownService creates a new markdown service
func NewMarkdownService() *MarkdownService {
	// Create a goldmark instance with extensions similar to Python terrareg
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM, // GitHub Flavored Markdown - includes tables, fenced code, etc.
			extension.Footnote,
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
			parser.WithAttribute(),
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithXHTML(),
			html.WithUnsafe(), // Allow raw HTML (like Python's markdown)
		),
	)

	return &MarkdownService{
		converter: md,
	}
}

// ConvertToHTML converts markdown content to HTML with optional file-specific processing
func (s *MarkdownService) ConvertToHTML(markdown string, options ...MarkdownOption) string {
	if markdown == "" {
		return ""
	}

	opts := &MarkdownOptions{}
	for _, opt := range options {
		opt(opts)
	}

	var buf bytes.Buffer
	source := []byte(markdown)

	// Parse the markdown to AST for custom processing
	if opts.FileName != "" {
		processedSource := s.processMarkdownWithFileContext(source, opts.FileName)
		source = processedSource
	}

	if err := s.converter.Convert(source, &buf); err != nil {
		// Fallback to basic escaping if markdown conversion fails
		return fmt.Sprintf(`<div class="markdown-content"><pre>%s</pre></div>`, markdown)
	}

	html := buf.String()

	// Post-process HTML if needed
	if opts.FileName != "" {
		html = s.postProcessHTML(html, opts.FileName)
	}

	return fmt.Sprintf(`<div class="markdown-content">%s</div>`, html)
}

// MarkdownOptions provides options for markdown conversion
type MarkdownOptions struct {
	FileName string // For link processing and anchor generation
}

// MarkdownOption represents a functional option for markdown conversion
type MarkdownOption func(*MarkdownOptions)

// WithFileName sets the filename for markdown processing (for link/anchor handling)
func WithFileName(filename string) MarkdownOption {
	return func(opts *MarkdownOptions) {
		opts.FileName = filename
	}
}

// processMarkdownWithFileContext processes markdown content with file-specific context
// Similar to Python terrareg's markdown_link_modifier functionality
func (s *MarkdownService) processMarkdownWithFileContext(source []byte, filename string) []byte {
	// This is a simplified version - in a full implementation,
	// we'd use AST visitors to modify links and anchors like Python does
	// For now, we'll do basic preprocessing
	content := string(source)

	// Convert relative links to anchor links within the same file
	// Example: ./README.md#section -> #section
	re := regexp.MustCompile(`\.\/` + regexp.QuoteMeta(filename) + `#(.+)`)
	content = re.ReplaceAllString(content, `#$1`)

	// Convert just the filename with anchor to proper anchor
	re = regexp.MustCompile(regexp.QuoteMeta(filename) + `#(.+)`)
	content = re.ReplaceAllString(content, `#$1`)

	return []byte(content)
}

// postProcessHTML performs post-processing on generated HTML
func (s *MarkdownService) postProcessHTML(html, filename string) string {
	// Remove relative image sources to avoid broken images
	// Similar to Python's ImageSourceCheck
	imgSrcRegex := regexp.MustCompile(`<img[^>]+src=["']([^"']+)["'][^>]*>`)
	html = imgSrcRegex.ReplaceAllStringFunc(html, func(match string) string {
		// Extract src attribute
		srcRegex := regexp.MustCompile(`src=["']([^"']+)["']`)
		if matches := srcRegex.FindStringSubmatch(match); len(matches) > 1 {
			src := matches[1]
			// Remove src if it's not an absolute URL
			if !strings.HasPrefix(src, "http://") && !strings.HasPrefix(src, "https://") {
				// Remove the src attribute
				return strings.Replace(match, `src="`+src+`"`, "", 1)
			}
		}
		return match
	})

	return html
}

// SanitizeHTML performs basic HTML sanitization for security
// In a production environment, consider using a proper HTML sanitizer like bluemonday
func (s *MarkdownService) SanitizeHTML(html string) string {
	// Basic sanitization - remove potentially dangerous elements
	dangerousTags := []string{"script", "iframe", "object", "embed"}

	for _, tag := range dangerousTags {
		// Remove opening tags and everything between them
		openTagRegex := regexp.MustCompile(`(?s)<` + tag + `[^>]*>.*?</` + tag + `>`)
		html = openTagRegex.ReplaceAllString(html, "")

		// Remove standalone opening tags
		openTagStandaloneRegex := regexp.MustCompile(`<` + tag + `[^>]*>`)
		html = openTagStandaloneRegex.ReplaceAllString(html, "")

		// Remove closing tags
		closeTagRegex := regexp.MustCompile(`</` + tag + `>`)
		html = closeTagRegex.ReplaceAllString(html, "")
	}

	// For form tags, remove just the form tags but preserve the content
	formTagRegex := regexp.MustCompile(`</?form[^>]*>`)
	html = formTagRegex.ReplaceAllString(html, "")

	// Remove dangerous attributes
	dangerousAttrs := []string{"onload", "onerror", "onclick", "onmouseover"}
	for _, attr := range dangerousAttrs {
		attrRegex := regexp.MustCompile(`\s` + attr + `=["'][^"']*["']`)
		html = attrRegex.ReplaceAllString(html, "")
	}

	// Remove javascript: links using a more precise pattern
	jsLinkRegex := regexp.MustCompile(`href\s*=\s*"(?:[^""]*javascript[^""]*)"`)
	html = jsLinkRegex.ReplaceAllString(html, `href="#"`)

	// Also handle single quotes
	jsLinkRegexSingle := regexp.MustCompile(`href\s*=\s*'(?:[^'']*javascript[^'']*)'`)
	html = jsLinkRegexSingle.ReplaceAllString(html, `href='#'`)

	return html
}
