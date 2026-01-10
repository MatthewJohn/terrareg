package testutils

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// This is a simpler version than AssertJSONResponse in helpers.go which also validates response body
func AssertJSONContentTypeAndCode(t *testing.T, w *httptest.ResponseRecorder, expectedCode int) {
	assert.Equal(t, expectedCode, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	// Verify body is valid JSON
	var body interface{}
	err := json.Unmarshal(w.Body.Bytes(), &body)
	require.NoError(t, err, "Response body should be valid JSON")
}

// AssertJSONResponseWithBody asserts that the response has the expected status code and body
func AssertJSONResponseWithBody(t *testing.T, w *httptest.ResponseRecorder, expectedCode int, expectedBody interface{}) {
	assert.Equal(t, expectedCode, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	expectedJSON, _ := json.Marshal(expectedBody)
	actualJSON, _ := json.Marshal(response)

	assert.JSONEq(t, string(expectedJSON), string(actualJSON))
}

// AssertErrorResponse asserts that the response is an error response
func AssertErrorResponse(t *testing.T, w *httptest.ResponseRecorder, expectedCode int, expectedMessage string) {
	assert.Equal(t, expectedCode, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	if expectedMessage != "" {
		if msg, ok := response["error"]; ok {
			assert.Contains(t, msg.(string), expectedMessage)
		} else if msg, ok := response["message"]; ok {
			assert.Contains(t, msg.(string), expectedMessage)
		}
	}
}

// AssertPaginatedResponseV2 asserts that the response contains pagination metadata
// This is an alternative to AssertPaginatedResponse with different field names
func AssertPaginatedResponseV2(t *testing.T, w *httptest.ResponseRecorder, expectedLimit, expectedOffset, expectedTotal int) {
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Check for meta.pagination structure (JSON-API style)
	if meta, ok := response["meta"].(map[string]interface{}); ok {
		if pagination, ok := meta["pagination"].(map[string]interface{}); ok {
			if limit, ok := pagination["limit"]; ok {
				assert.Equal(t, float64(expectedLimit), limit)
			}
			if offset, ok := pagination["offset"]; ok {
				assert.Equal(t, float64(expectedOffset), offset)
			}
		}
	}

	// Check for direct pagination fields
	if limit, ok := response["limit"]; ok {
		assert.Equal(t, float64(expectedLimit), limit)
	}
	if offset, ok := response["offset"]; ok {
		assert.Equal(t, float64(expectedOffset), offset)
	}
	if total, ok := response["total"]; ok {
		assert.Equal(t, float64(expectedTotal), total)
	}
}

// AssertDataTablesResponse asserts that the response matches DataTables format
func AssertDataTablesResponse(t *testing.T, w *httptest.ResponseRecorder, expectedDraw, expectedRecordsTotal, expectedRecordsFiltered int) {
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response, "draw")
	assert.Contains(t, response, "recordsTotal")
	assert.Contains(t, response, "recordsFiltered")
	assert.Contains(t, response, "data")

	draw := int(response["draw"].(float64))
	recordsTotal := int(response["recordsTotal"].(float64))
	recordsFiltered := int(response["recordsFiltered"].(float64))

	assert.Equal(t, expectedDraw, draw)
	assert.Equal(t, expectedRecordsTotal, recordsTotal)
	assert.Equal(t, expectedRecordsFiltered, recordsFiltered)
}

// CreateRequestWithChiParams creates an HTTP request with Chi route context parameters
func CreateRequestWithChiParams(t *testing.T, method, url string, params map[string]string) *http.Request {
	req := httptest.NewRequest(method, url, nil)

	// Create Chi route context
	rctx := chi.NewRouteContext()
	for key, value := range params {
		rctx.URLParams.Add(key, value)
	}

	// Add context to request
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	return req
}

// AddChiContext adds Chi route context to an existing request
func AddChiContext(t *testing.T, req *http.Request, params map[string]string) *http.Request {
	rctx := chi.NewRouteContext()
	for key, value := range params {
		rctx.URLParams.Add(key, value)
	}

	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
}

// ParseJSONAPIResponse parses a JSON-API response
func ParseJSONAPIResponse(t *testing.T, body []byte) map[string]interface{} {
	var response map[string]interface{}
	err := json.Unmarshal(body, &response)
	require.NoError(t, err)
	return response
}

// GetJSONBodyMap gets the JSON body from a response recorder (alias for GetJSONBody to avoid conflict)
func GetJSONBodyMap(t *testing.T, w *httptest.ResponseRecorder) map[string]interface{} {
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	return response
}

// BuildQueryString builds a URL query string from a map
func BuildQueryString(params map[string]string) string {
	values := url.Values{}
	for key, value := range params {
		values.Add(key, value)
	}
	return values.Encode()
}

// BuildModuleSearchRequest builds query parameters for module search
func BuildModuleSearchRequest(query string, namespaces []string, providers []string, limit, offset int) string {
	params := url.Values{}
	if query != "" {
		params.Add("q", query)
	}
	for _, ns := range namespaces {
		params.Add("namespace", ns)
	}
	for _, p := range providers {
		params.Add("provider", p)
	}
	if limit > 0 {
		params.Add("limit", string(rune(limit)))
	}
	if offset > 0 {
		params.Add("offset", string(rune(offset)))
	}
	return params.Encode()
}

// BuildDataTablesRequest builds query parameters for DataTables requests
func BuildDataTablesRequest(search string, draw, start, length int, orderCol, orderDir string) string {
	params := url.Values{}
	params.Add("draw", string(rune(draw)))
	params.Add("start", string(rune(start)))
	params.Add("length", string(rune(length)))
	if search != "" {
		params.Add("search[value]", search)
	}
	if orderCol != "" {
		params.Add("order[0][column]", orderCol)
	}
	if orderDir != "" {
		params.Add("order[0][dir]", orderDir)
	}
	return params.Encode()
}

// AssertContentType asserts the response has the expected content type
func AssertContentType(t *testing.T, w *httptest.ResponseRecorder, contentType string) {
	actual := w.Header().Get("Content-Type")
	if contentType == "application/json" {
		// Allow charset suffix
		assert.True(t,
			strings.HasPrefix(actual, "application/json"),
			"Expected content type to start with application/json, got %s", actual)
	} else {
		assert.Equal(t, contentType, actual)
	}
}

// AssertHeader asserts a specific header value
func AssertHeader(t *testing.T, w *httptest.ResponseRecorder, key, expectedValue string) {
	assert.Equal(t, expectedValue, w.Header().Get(key))
}

// AssertCookie asserts a specific cookie is set
func AssertCookie(t *testing.T, w *httptest.ResponseRecorder, cookieName string) *http.Cookie {
	cookies := w.Result().Cookies()
	for _, c := range cookies {
		if c.Name == cookieName {
			return c
		}
	}
	t.Fatalf("Cookie %s not found in response", cookieName)
	return nil
}

// AssertCookieValue asserts a specific cookie has the expected value
func AssertCookieValue(t *testing.T, w *httptest.ResponseRecorder, cookieName, expectedValue string) {
	cookie := AssertCookie(t, w, cookieName)
	assert.Equal(t, expectedValue, cookie.Value)
}
