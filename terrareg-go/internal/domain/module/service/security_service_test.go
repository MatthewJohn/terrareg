package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSecurityService_ValidateFilePath_ValidSimplePath(t *testing.T) {
	service := NewSecurityService()
	path := "main.tf"
	err := service.ValidateFilePath(path)

	assert.NoError(t, err,
		"Path '%s' should be valid type", path)
}

func TestSecurityService_ValidateFilePath_PathTraversal_DoubleDot(t *testing.T) {
	service := NewSecurityService()
	path := "../etc/passwd"
	err := service.ValidateFilePath(path)

	assert.Error(t, err,
		"Path '%s' should be invalid type - contains path traversal", path)
	assert.ErrorIs(t, err, ErrInvalidFilePath)
}

func TestSecurityService_ValidateFilePath_AbsolutePath(t *testing.T) {
	service := NewSecurityService()
	path := "/etc/passwd"
	err := service.ValidateFilePath(path)

	assert.Error(t, err,
		"Path '%s' should be invalid type - absolute path", path)
	assert.ErrorIs(t, err, ErrInvalidFilePath)
}

func TestSecurityService_ValidateFilePath_ForbiddenChar(t *testing.T) {
	service := NewSecurityService()
	path := "main<.tf"
	err := service.ValidateFilePath(path)

	assert.Error(t, err,
		"Path '%s' contains forbidden character", path)
	assert.ErrorIs(t, err, ErrInvalidFilePath)
}

func TestSecurityService_ValidateFilePath_NullByte(t *testing.T) {
	service := NewSecurityService()
	path := ""
	err := service.ValidateFilePath(path)

	assert.Error(t, err,
		"Path '%s' is null - byte injection detected", path)
	assert.ErrorIs(t, err, ErrInvalidFilePath)
}

func TestSecurityService_ValidateFileType_ValidExtension(t *testing.T) {
	service := NewSecurityService()
	fileName := "module.tf"
	err := service.ValidateFileType(fileName)

	assert.NoError(t, err)
}

func TestSecurityService_ValidateFileType_InvalidExtension(t *testing.T) {
	service := NewSecurityService()
	fileName := "malware.exe"
	err := service.ValidateFileType(fileName)

	assert.Error(t, err,
		"File '%s' should be invalid type", fileName)
	assert.ErrorIs(t, err, ErrInvalidFileType)
}

func TestSecurityService_ValidateFileType_NoExtension(t *testing.T) {
	service := NewSecurityService()
	fileName := "readme"
	err := service.ValidateFileType(fileName)

	assert.Error(t, err,
		"File '%s' has no extension", fileName)
	assert.ErrorIs(t, err, ErrInvalidFileType)
}

func TestSecurityService_ValidateFileType_DoubleExtension(t *testing.T) {
	service := NewSecurityService()
	fileName := "file.tf.exe"
	err := service.ValidateFileType(fileName)

	assert.Error(t, err,
		"File '%s' has double extension", fileName)
	assert.ErrorIs(t, err, ErrInvalidFileType)
}

func TestSecurityService_ValidateFileType_TrailingDot(t *testing.T) {
	service := NewSecurityService()
	fileName := "config."
	err := service.ValidateFileType(fileName)

	assert.Error(t, err,
		"File '%s' has trailing dot", fileName)
	assert.ErrorIs(t, err, ErrInvalidFileType)
}

func TestSecurityService_ValidateFileType_LeadingDot(t *testing.T) {
	service := NewSecurityService()
	fileName := ".hidden"
	err := service.ValidateFileType(fileName)

	assert.Error(t, err,
		"File '%s' has leading dot", fileName)
	assert.ErrorIs(t, err, ErrInvalidFileType)
}

func TestSecurityService_ValidateFileType_CaseInsensitiveValid(t *testing.T) {
	service := NewSecurityService()
	fileName := "MODULE.TF"
	err := service.ValidateFileType(fileName)

	assert.NoError(t, err,
		"File '%s' should be valid type - case insensitive", fileName)
}

func TestSecurityService_ValidateFileType_NullFilename(t *testing.T) {
	service := NewSecurityService()
	fileName := ""
	err := service.ValidateFileType(fileName)

	assert.Error(t, err,
		"Path '%s' is null - byte injection detected", fileName)
	assert.ErrorIs(t, err, ErrInvalidFileType)
}

func TestSecurityService_SanitizeContent_RemoveScriptTag(t *testing.T) {
	service := NewSecurityService()
	content := "<script>alert('xss')</script>hello"
	err := service.SanitizeContent(&content)

	assert.NoError(t, err)
	assert.NotContains(t, content, "<script")
	assert.NotContains(t, content, "</script")
	assert.NotContains(t, content, "alert")
	assert.Equal(t, "hello", content)
}

func TestSecurityService_SanitizeContent_RemoveUppercaseScript(t *testing.T) {
	service := NewSecurityService()
	content := "<SCRIPT>alert('xss')</SCRIPT>hello"
	err := service.SanitizeContent(&content)

	assert.NoError(t, err)
	assert.NotContains(t, content, "<script")
	assert.NotContains(t, content, "</script")
	assert.NotContains(t, content, "alert")
	assert.Equal(t, "hello", content)
}

func TestSecurityService_SanitizeContent_RemoveJavaScriptProtocol(t *testing.T) {
	service := NewSecurityService()
	content := "<a href=\"javascript:alert('xss')\">click</a>"
	err := service.SanitizeContent(&content)

	assert.NoError(t, err)
	assert.NotContains(t, content, "javascript:")
}

func TestSecurityService_SanitizeContent_RemoveDataProtocol(t *testing.T) {
	service := NewSecurityService()
	content := "<img src=data:text/html,<script>alert(1)</script>>"
	err := service.SanitizeContent(&content)

	assert.NoError(t, err)
	assert.NotContains(t, content, "data:")
	assert.NotContains(t, content, "alert(1)")
}

func TestSecurityService_SanitizeContent_RemoveVbscript(t *testing.T) {
	service := NewSecurityService()
	content := "vbscript:msgbox('xss')"
	err := service.SanitizeContent(&content)

	assert.NoError(t, err)
	assert.NotContains(t, content, "vbscript:")
	assert.Equal(t, "msgbox('xss')", content)
}

func TestSecurityService_SanitizeContent_RemoveFile(t *testing.T) {
	service := NewSecurityService()
	content := "file:///etc/passwd"
	err := service.SanitizeContent(&content)

	assert.NoError(t, err)
	assert.NotContains(t, content, "file:")
	assert.Equal(t, "///etc/passwd", content)
}

func TestSecurityService_SanitizeContent_RemoveIframe(t *testing.T) {
	service := NewSecurityService()
	content := "<iframe src=evil.com></iframe>content"
	err := service.SanitizeContent(&content)

	assert.NoError(t, err)
	assert.NotContains(t, content, "<iframe")
	assert.NotContains(t, content, "</iframe")
	assert.NotContains(t, content, "src=evil.com")
	assert.Equal(t, "content", content)
}

func TestSecurityService_SanitizeContent_RemoveObject(t *testing.T) {
	service := NewSecurityService()
	content := "<object data=evil.swf></object>content"
	err := service.SanitizeContent(&content)

	assert.NoError(t, err)
	assert.NotContains(t, content, "<object")
	assert.NotContains(t, content, "</object")
	assert.NotContains(t, content, "data=evil.swf")
	assert.Equal(t, "content", content)
}

func TestSecurityService_SanitizeContent_RemoveEmbed(t *testing.T) {
	service := NewSecurityService()
	content := "<embed src=evil.swf>content"
	err := service.SanitizeContent(&content)

	assert.NoError(t, err)
	assert.NotContains(t, content, "<embed")
	assert.NotContains(t, content, "src=evil.swf")
	assert.Equal(t, "content", content)
}

func TestSecurityService_SanitizeContent_RemoveFormTag(t *testing.T) {
	service := NewSecurityService()
	content := "<form action=evil.com></form>content"
	err := service.SanitizeContent(&content)

	assert.NoError(t, err)
	assert.NotContains(t, content, "<form")
	assert.NotContains(t, content, "</form")
	assert.NotContains(t, content, "action=evil.com")
	assert.Equal(t, "content", content)
}

func TestSecurityService_SanitizeContent_RemoveInputTag(t *testing.T) {
	service := NewSecurityService()
	content := "<input type=hidden>content"
	err := service.SanitizeContent(&content)

	assert.NoError(t, err)
	assert.NotContains(t, content, "<input")
	assert.NotContains(t, content, "type=hidden")
	assert.Equal(t, "content", content)
}

func TestSecurityService_SanitizeContent_RemoveOnload(t *testing.T) {
	service := NewSecurityService()
	content := "<img src=x onload=alert('xss')\"\">"
	err := service.SanitizeContent(&content)

	assert.NoError(t, err)
	assert.NotContains(t, content, "onload")
	assert.NotContains(t, content, "alert")
}

func TestSecurityService_SanitizeContent_RemoveOnblur(t *testing.T) {
	service := NewSecurityService()
	content := "<input onblur='alert(\"xss\")'\">"
	err := service.SanitizeContent(&content)

	assert.NoError(t, err)
	assert.NotContains(t, content, "onblur")
	assert.NotContains(t, content, "alert")
}

func TestSecurityService_SanitizeContent_RemoveOnfocus(t *testing.T) {
	service := NewSecurityService()
	content := "<input onfocus=\"alert('xss')\">"
	err := service.SanitizeContent(&content)

	assert.NoError(t, err)
	assert.NotContains(t, content, "onfocus")
	assert.NotContains(t, content, "alert")
}

func TestSecurityService_SanitizeContent_RemoveOnchange(t *testing.T) {
	service := NewSecurityService()
	content := "<form onchange='alert(\"xss\")'></form>"
	err := service.SanitizeContent(&content)

	assert.NoError(t, err)
	assert.NotContains(t, content, "onchange")
	assert.NotContains(t, content, "alert")
}

func TestSecurityService_SanitizeContent_RemoveOnreset(t *testing.T) {
	service := NewSecurityService()
	content := "<form onreset='alert(\"xss\")'></form>"
	err := service.SanitizeContent(&content)

	assert.NoError(t, err)
	assert.NotContains(t, content, "onreset")
	assert.NotContains(t, content, "alert")
}

func TestSecurityService_SanitizeContent_RemoveMultipleScriptTags(t *testing.T) {
	service := NewSecurityService()
	content := "<script>alert(1)</script><script>alert(2)</script>"
	err := service.SanitizeContent(&content)

	assert.NoError(t, err)
	assert.NotContains(t, content, "<script")
	assert.NotContains(t, content, "</script")
	assert.NotContains(t, content, "alert")
	assert.Equal(t, "", content)
}

func TestSecurityService_SanitizeContent_RemoveMixedInjections(t *testing.T) {
	service := NewSecurityService()
	content := "<script>alert(1)</script>javascript:alert(2)<iframe src=evil>"
	err := service.SanitizeContent(&content)

	assert.NoError(t, err)
	assert.NotContains(t, content, "<script")
	assert.NotContains(t, content, "javascript:")
	assert.NotContains(t, content, "<iframe")
	assert.NotContains(t, content, "alert(1)")
	assert.NotContains(t, content, "alert(2)")
}

func TestSecurityService_SanitizeContent_NestedDangers(t *testing.T) {
	service := NewSecurityService()
	content := "<script><iframe>evil</iframe></script>"
	err := service.SanitizeContent(&content)

	assert.NoError(t, err)
	assert.NotContains(t, content, "<script")
	assert.NotContains(t, content, "</script")
	assert.NotContains(t, content, "<iframe")
	assert.NotContains(t, content, "</iframe")
	assert.Equal(t, "", content)
}

func TestSecurityService_SanitizeContent_PlainText(t *testing.T) {
	service := NewSecurityService()
	content := "Hello, World!"
	err := service.SanitizeContent(&content)

	assert.NoError(t, err)
	assert.Equal(t, "Hello, World!", content)
}

func TestSecurityService_SanitizeContent_EmptyString(t *testing.T) {
	service := NewSecurityService()
	content := ""
	err := service.SanitizeContent(&content)

	assert.NoError(t, err)
	assert.Empty(t, content)
}

func TestSecurityService_SanitizeContent_SafeHTML(t *testing.T) {
	service := NewSecurityService()
	content := "<p>hello</p><div>world</div>"
	err := service.SanitizeContent(&content)

	assert.NoError(t, err)
	assert.Equal(t, "<p>hello</p><div>world</div>", content)
}

func TestSecurityService_SanitizeContent_NilContent(t *testing.T) {
	service := NewSecurityService()
	var content *string = nil
	err := service.SanitizeContent(content)

	assert.NoError(t, err)
	assert.Nil(t, content)
}

func TestSecurityService_NewService(t *testing.T) {
	service := NewSecurityService()

	assert.NotNil(t, service)
	assert.NotNil(t, service.allowedFileTypes)
	assert.NotNil(t, service.pathValidator)
}

func TestSecurityService_PathValidatorComplex(t *testing.T) {
	service := NewSecurityService()

	testCases := []struct {
		path   string
		valid  bool
		reason string
	}{
		{"modules/network/main.tf", true, "valid nested path"},
		{"../escape.tf", false, "path traversal"},
		{"/absolute/path.tf", false, "absolute path"},
		{"file<name.tf", false, "forbidden char <"},
		{"file>name.tf", false, "forbidden char >"},
		{"file:name.tf", false, "forbidden char :"},
		{"normal_file.tf", true, "underscore allowed"},
		{"file with spaces.tf", false, "spaces not allowed"},
		{"special@chars.tf", false, "forbidden char @"},
		{"test-123.tf", true, "hyphen and numbers allowed"},
		{"./current.tf", true, "current directory reference"},
		{"dir/./file.tf", true, "current dir in middle"},
		{"dir/../file.tf", false, "path traversal in middle"},
		{"", false, "empty path"},
	}

	for _, tc := range testCases {
		t.Run(tc.path, func(t *testing.T) {
			err := service.ValidateFilePath(tc.path)
			if tc.valid {
				assert.NoError(t, err, tc.reason)
			} else {
				assert.Error(t, err, tc.reason)
				assert.ErrorIs(t, err, ErrInvalidFilePath)
			}
		})
	}
}
