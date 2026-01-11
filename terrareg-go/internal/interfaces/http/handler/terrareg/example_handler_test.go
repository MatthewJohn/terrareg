package terrareg_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/application/query/module"
)

// MockGetExampleDetailsQuery is a mock for GetExampleDetailsQuery
type MockGetExampleDetailsQuery struct {
	mock.Mock
}

func (m *MockGetExampleDetailsQuery) Execute(ctx context.Context, namespace, moduleName, provider, version, examplePath string) (*module.ExampleDetails, error) {
	args := m.Called(ctx, namespace, moduleName, provider, version, examplePath)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*module.ExampleDetails), args.Error(1)
}

// MockGetExampleReadmeHTMLQuery is a mock for GetExampleReadmeHTMLQuery
type MockGetExampleReadmeHTMLQuery struct {
	mock.Mock
}

func (m *MockGetExampleReadmeHTMLQuery) Execute(ctx context.Context, namespace, moduleName, provider, version, examplePath string) (*module.ExampleReadmeHTMLResponse, error) {
	args := m.Called(ctx, namespace, moduleName, provider, version, examplePath)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*module.ExampleReadmeHTMLResponse), args.Error(1)
}

// MockGetExampleFileListQuery is a mock for GetExampleFileListQuery
type MockGetExampleFileListQuery struct {
	mock.Mock
}

func (m *MockGetExampleFileListQuery) Execute(ctx context.Context, namespace, moduleName, provider, version, examplePath string) ([]module.ExampleFileInfo, error) {
	args := m.Called(ctx, namespace, moduleName, provider, version, examplePath)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]module.ExampleFileInfo), args.Error(1)
}

// MockGetExampleFileQuery is a mock for GetExampleFileQuery
type MockGetExampleFileQuery struct {
	mock.Mock
}

func (m *MockGetExampleFileQuery) Execute(ctx context.Context, namespace, moduleName, provider, version, examplePath, filePath string) (*module.ExampleFileResponse, error) {
	args := m.Called(ctx, namespace, moduleName, provider, version, examplePath, filePath)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*module.ExampleFileResponse), args.Error(1)
}

func TestExampleHandler_HandleExampleDetails(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		url            string
		expectedStatus int
		setupMocks     func(*MockGetExampleDetailsQuery, *MockGetExampleReadmeHTMLQuery, *MockGetExampleFileListQuery, *MockGetExampleFileQuery)
		expectedBody   string
	}{
		{
			name:           "invalid method",
			method:         "POST",
			url:            "/modules/test/mod/provider/1.0.0/examples/details/example",
			expectedStatus: http.StatusMethodNotAllowed,
			setupMocks: func(*MockGetExampleDetailsQuery, *MockGetExampleReadmeHTMLQuery, *MockGetExampleFileListQuery, *MockGetExampleFileQuery) {
			},
		},
		{
			name:           "missing namespace parameter",
			method:         "GET",
			url:            "/modules//mod/provider/1.0.0/examples/details/example",
			expectedStatus: http.StatusBadRequest,
			setupMocks: func(*MockGetExampleDetailsQuery, *MockGetExampleReadmeHTMLQuery, *MockGetExampleFileListQuery, *MockGetExampleFileQuery) {
			},
			expectedBody: "Missing required path parameters",
		},
		{
			name:           "missing example parameter",
			method:         "GET",
			url:            "/modules/test/mod/provider/1.0.0/examples/details/",
			expectedStatus: http.StatusBadRequest,
			setupMocks: func(*MockGetExampleDetailsQuery, *MockGetExampleReadmeHTMLQuery, *MockGetExampleFileListQuery, *MockGetExampleFileQuery) {
			},
			expectedBody: "Missing required path parameters",
		},
		{
			name:           "example not found",
			method:         "GET",
			url:            "/modules/test/mod/provider/1.0.0/examples/details/example",
			expectedStatus: http.StatusNotFound,
			setupMocks: func(mockDetails *MockGetExampleDetailsQuery, mockReadme *MockGetExampleReadmeHTMLQuery, mockFileList *MockGetExampleFileListQuery, mockFile *MockGetExampleFileQuery) {
				mockDetails.On("Execute", mock.Anything, "test", "mod", "provider", "1.0.0", "example").
					Return(nil, assert.AnError).
					Once()
			},
			expectedBody: "example not found",
		},
		{
			name:           "successful response",
			method:         "GET",
			url:            "/modules/test/mod/provider/1.0.0/examples/details/example",
			expectedStatus: http.StatusOK,
			setupMocks: func(mockDetails *MockGetExampleDetailsQuery, mockReadme *MockGetExampleReadmeHTMLQuery, mockFileList *MockGetExampleFileListQuery, mockFile *MockGetExampleFileQuery) {
				exampleDetails := &module.ExampleDetails{
					Path:  "example",
					Name:  stringPtr("Test Example"),
					Empty: false,
					Inputs: []module.Input{
						{
							Name:        "var1",
							Type:        "string",
							Description: "A test variable",
							Required:    true,
						},
					},
					Resources: []module.Resource{
						{
							Type: "aws_instance",
							Name: "example",
						},
					},
				}
				mockDetails.On("Execute", mock.Anything, "test", "mod", "provider", "1.0.0", "example").
					Return(exampleDetails, nil).
					Once()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockDetails := &MockGetExampleDetailsQuery{}
			mockReadme := &MockGetExampleReadmeHTMLQuery{}
			mockFileList := &MockGetExampleFileListQuery{}
			mockFile := &MockGetExampleFileQuery{}
			tt.setupMocks(mockDetails, mockReadme, mockFileList, mockFile)

			handler := NewExampleHandler(mockDetails, mockReadme, mockFileList, mockFile)

			// Create request
			req := httptest.NewRequest(tt.method, tt.url, nil)

			// Set up chi context
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("namespace", "test")
			rctx.URLParams.Add("name", "mod")
			rctx.URLParams.Add("provider", "provider")
			rctx.URLParams.Add("version", "1.0.0")
			rctx.URLParams.Add("example", "example")
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			// Create response recorder
			w := httptest.NewRecorder()

			// Act
			handler.HandleExampleDetails(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedBody != "" {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				if err == nil {
					// If it's JSON, check the error message
					if errMsg, ok := response["error"].(string); ok {
						assert.Contains(t, errMsg, tt.expectedBody)
					}
				} else {
					// If not JSON, check the raw body
					assert.Contains(t, w.Body.String(), tt.expectedBody)
				}
			}

			mockDetails.AssertExpectations(t)
		})
	}
}

func TestExampleHandler_HandleExampleFileList(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		url            string
		expectedStatus int
		setupMocks     func(*MockGetExampleDetailsQuery, *MockGetExampleReadmeHTMLQuery, *MockGetExampleFileListQuery, *MockGetExampleFileQuery)
		expectedFiles  []string
	}{
		{
			name:           "successful response",
			method:         "GET",
			url:            "/modules/test/mod/provider/1.0.0/examples/filelist/example",
			expectedStatus: http.StatusOK,
			setupMocks: func(mockDetails *MockGetExampleDetailsQuery, mockReadme *MockGetExampleReadmeHTMLQuery, mockFileList *MockGetExampleFileListQuery, mockFile *MockGetExampleFileQuery) {
				fileList := []module.ExampleFileInfo{
					{Path: "main.tf"},
					{Path: "variables.tf"},
					{Path: "outputs.tf"},
				}
				mockFileList.On("Execute", mock.Anything, "test", "mod", "provider", "1.0.0", "example").
					Return(fileList, nil).
					Once()
			},
			expectedFiles: []string{"main.tf", "variables.tf", "outputs.tf"},
		},
		{
			name:           "empty file list",
			method:         "GET",
			url:            "/modules/test/mod/provider/1.0.0/examples/filelist/empty",
			expectedStatus: http.StatusOK,
			setupMocks: func(mockDetails *MockGetExampleDetailsQuery, mockReadme *MockGetExampleReadmeHTMLQuery, mockFileList *MockGetExampleFileListQuery, mockFile *MockGetExampleFileQuery) {
				fileList := []module.ExampleFileInfo{}
				mockFileList.On("Execute", mock.Anything, "test", "mod", "provider", "1.0.0", "empty").
					Return(fileList, nil).
					Once()
			},
			expectedFiles: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockDetails := &MockGetExampleDetailsQuery{}
			mockReadme := &MockGetExampleReadmeHTMLQuery{}
			mockFileList := &MockGetExampleFileListQuery{}
			mockFile := &MockGetExampleFileQuery{}
			tt.setupMocks(mockDetails, mockReadme, mockFileList, mockFile)

			handler := NewExampleHandler(mockDetails, mockReadme, mockFileList, mockFile)

			// Create request
			req := httptest.NewRequest(tt.method, tt.url, nil)

			// Set up chi context - extract example from URL
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("namespace", "test")
			rctx.URLParams.Add("name", "mod")
			rctx.URLParams.Add("provider", "provider")
			rctx.URLParams.Add("version", "1.0.0")

			// Extract example name from URL
			if tt.url[len("/modules/test/mod/provider/1.0.0/examples/filelist/"):] == "example" {
				rctx.URLParams.Add("example", "example")
			} else {
				rctx.URLParams.Add("example", "empty")
			}

			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			// Create response recorder
			w := httptest.NewRecorder()

			// Act
			handler.HandleExampleFileList(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

			var response []module.ExampleFileInfo
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, len(tt.expectedFiles), len(response))

			for i, expectedFile := range tt.expectedFiles {
				assert.Equal(t, expectedFile, response[i].Path)
			}

			mockFileList.AssertExpectations(t)
		})
	}
}

func TestExampleHandler_HandleExampleFile(t *testing.T) {
	tests := []struct {
		name            string
		method          string
		url             string
		expectedStatus  int
		setupMocks      func(*MockGetExampleDetailsQuery, *MockGetExampleReadmeHTMLQuery, *MockGetExampleFileListQuery, *MockGetExampleFileQuery)
		expectedContent string
		expectedType    string
	}{
		{
			name:            "successful response - Terraform file",
			method:          "GET",
			url:             "/modules/test/mod/provider/1.0.0/examples/file/example/main.tf",
			expectedStatus:  http.StatusOK,
			expectedContent: "resource \"aws_instance\" \"example\" {}",
			expectedType:    "text/x-hcl",
			setupMocks: func(mockDetails *MockGetExampleDetailsQuery, mockReadme *MockGetExampleReadmeHTMLQuery, mockFileList *MockGetExampleFileListQuery, mockFile *MockGetExampleFileQuery) {
				fileResponse := &module.ExampleFileResponse{
					Path:        "main.tf",
					Content:     "resource \"aws_instance\" \"example\" {}",
					ContentType: "text/x-hcl",
				}
				mockFile.On("Execute", mock.Anything, "test", "mod", "provider", "1.0.0", "example", "main.tf").
					Return(fileResponse, nil).
					Once()
			},
		},
		{
			name:            "successful response - JSON file",
			method:          "GET",
			url:             "/modules/test/mod/provider/1.0.0/examples/file/example/config.json",
			expectedStatus:  http.StatusOK,
			expectedContent: "{\"key\": \"value\"}",
			expectedType:    "application/json",
			setupMocks: func(mockDetails *MockGetExampleDetailsQuery, mockReadme *MockGetExampleReadmeHTMLQuery, mockFileList *MockGetExampleFileListQuery, mockFile *MockGetExampleFileQuery) {
				fileResponse := &module.ExampleFileResponse{
					Path:        "config.json",
					Content:     "{\"key\": \"value\"}",
					ContentType: "application/json",
				}
				mockFile.On("Execute", mock.Anything, "test", "mod", "provider", "1.0.0", "example", "config.json").
					Return(fileResponse, nil).
					Once()
			},
		},
		{
			name:           "file not found",
			method:         "GET",
			url:            "/modules/test/mod/provider/1.0.0/examples/file/example/missing.tf",
			expectedStatus: http.StatusNotFound,
			setupMocks: func(mockDetails *MockGetExampleDetailsQuery, mockReadme *MockGetExampleReadmeHTMLQuery, mockFileList *MockGetExampleFileListQuery, mockFile *MockGetExampleFileQuery) {
				mockFile.On("Execute", mock.Anything, "test", "mod", "provider", "1.0.0", "example", "missing.tf").
					Return(nil, assert.AnError).
					Once()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockDetails := &MockGetExampleDetailsQuery{}
			mockReadme := &MockGetExampleReadmeHTMLQuery{}
			mockFileList := &MockGetExampleFileListQuery{}
			mockFile := &MockGetExampleFileQuery{}
			tt.setupMocks(mockDetails, mockReadme, mockFileList, mockFile)

			handler := NewExampleHandler(mockDetails, mockReadme, mockFileList, mockFile)

			// Create request
			req := httptest.NewRequest(tt.method, tt.url, nil)

			// Set up chi context
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("namespace", "test")
			rctx.URLParams.Add("name", "mod")
			rctx.URLParams.Add("provider", "provider")
			rctx.URLParams.Add("version", "1.0.0")
			rctx.URLParams.Add("example", "example")
			rctx.URLParams.Add("file", "main.tf") // For the first test
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			// Create response recorder
			w := httptest.NewRecorder()

			// Act
			handler.HandleExampleFile(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				assert.Equal(t, tt.expectedType, w.Header().Get("Content-Type"))
				assert.Equal(t, tt.expectedContent, w.Body.String())
			}

			mockFile.AssertExpectations(t)
		})
	}
}

func TestExampleHandler_HandleExampleReadmeHTML(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		url            string
		expectedStatus int
		setupMocks     func(*MockGetExampleDetailsQuery, *MockGetExampleReadmeHTMLQuery, *MockGetExampleFileListQuery, *MockGetExampleFileQuery)
		checkHeaders   bool
		expectedHTML   string
	}{
		{
			name:           "no readme content",
			method:         "GET",
			url:            "/modules/test/mod/provider/1.0.0/examples/readme_html/example",
			expectedStatus: http.StatusOK,
			setupMocks: func(mockDetails *MockGetExampleDetailsQuery, mockReadme *MockGetExampleReadmeHTMLQuery, mockFileList *MockGetExampleFileListQuery, mockFile *MockGetExampleFileQuery) {
				mockReadme.On("Execute", mock.Anything, "test", "mod", "provider", "1.0.0", "example").
					Return(nil, assert.AnError).
					Once()
			},
			checkHeaders: true,
			expectedHTML: `<div class="alert alert-warning">No README found for this example</div>`,
		},
		{
			name:           "successful response",
			method:         "GET",
			url:            "/modules/test/mod/provider/1.0.0/examples/readme_html/example",
			expectedStatus: http.StatusOK,
			setupMocks: func(mockDetails *MockGetExampleDetailsQuery, mockReadme *MockGetExampleReadmeHTMLQuery, mockFileList *MockGetExampleFileListQuery, mockFile *MockGetExampleFileQuery) {
				readmeResponse := &module.ExampleReadmeHTMLResponse{
					HTML: "<h1>Test Example</h1><p>This is a test example</p>",
				}
				mockReadme.On("Execute", mock.Anything, "test", "mod", "provider", "1.0.0", "example").
					Return(readmeResponse, nil).
					Once()
			},
			checkHeaders: true,
			expectedHTML: "<h1>Test Example</h1><p>This is a test example</p>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockDetails := &MockGetExampleDetailsQuery{}
			mockReadme := &MockGetExampleReadmeHTMLQuery{}
			mockFileList := &MockGetExampleFileListQuery{}
			mockFile := &MockGetExampleFileQuery{}
			tt.setupMocks(mockDetails, mockReadme, mockFileList, mockFile)

			handler := NewExampleHandler(mockDetails, mockReadme, mockFileList, mockFile)

			// Create request
			req := httptest.NewRequest(tt.method, tt.url, nil)

			// Set up chi context
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("namespace", "test")
			rctx.URLParams.Add("name", "mod")
			rctx.URLParams.Add("provider", "provider")
			rctx.URLParams.Add("version", "1.0.0")
			rctx.URLParams.Add("example", "example")
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			// Create response recorder
			w := httptest.NewRecorder()

			// Act
			handler.HandleExampleReadmeHTML(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.checkHeaders {
				assert.Equal(t, "text/html; charset=utf-8", w.Header().Get("Content-Type"))
			}

			if tt.expectedHTML != "" {
				assert.Equal(t, tt.expectedHTML, w.Body.String())
			}

			mockReadme.AssertExpectations(t)
		})
	}
}

func TestExampleHelperFunctions(t *testing.T) {
	t.Run("splitExampleFilePath", func(t *testing.T) {
		tests := []struct {
			input    string
			expected []string
		}{
			{
				input:    "example/main.tf",
				expected: []string{"example", "main.tf"},
			},
			{
				input:    "example/subdir/main.tf",
				expected: []string{"example", "subdir/main.tf"},
			},
			{
				input:    "example",
				expected: []string{"example"},
			},
		}

		for _, tt := range tests {
			result := splitExampleFilePath(tt.input)
			assert.Equal(t, tt.expected, result)
		}
	})

	t.Run("joinPathParts", func(t *testing.T) {
		tests := []struct {
			input    []string
			expected string
		}{
			{
				input:    []string{"example", "main.tf"},
				expected: "example/main.tf",
			},
			{
				input:    []string{"example", "subdir", "main.tf"},
				expected: "example/subdir/main.tf",
			},
			{
				input:    []string{"single"},
				expected: "single",
			},
			{
				input:    []string{},
				expected: "",
			},
		}

		for _, tt := range tests {
			result := joinPathParts(tt.input...)
			assert.Equal(t, tt.expected, result)
		}
	})
}
