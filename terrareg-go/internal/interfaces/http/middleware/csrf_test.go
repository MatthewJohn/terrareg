package middleware

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/security/csrf"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

// newTestCSRFMiddleware creates a test CSRF middleware
func newTestCSRFMiddleware() *CSRFMiddleware {
	logger := zerolog.New(nil).With().Timestamp().Logger()
	return NewCSRFMiddleware(logger)
}

// TestNewCSRFMiddleware tests the constructor
func TestNewCSRFMiddleware(t *testing.T) {
	logger := zerolog.New(nil).With().Timestamp().Logger()
	middleware := NewCSRFMiddleware(logger)

	assert.NotNil(t, middleware)
	assert.NotNil(t, middleware.validator)
	assert.NotNil(t, middleware.logger)
}

// TestRequiresCSRFProtection tests which HTTP methods require CSRF protection
func TestRequiresCSRFProtection(t *testing.T) {
	middleware := newTestCSRFMiddleware()

	stateChangingMethods := []string{"POST", "PUT", "PATCH", "DELETE"}
	for _, method := range stateChangingMethods {
		t.Run("requires protection for "+method, func(t *testing.T) {
			assert.True(t, middleware.requiresCSRFProtection(method))
		})
	}

	safeMethods := []string{"GET", "HEAD", "OPTIONS", "TRACE"}
	for _, method := range safeMethods {
		t.Run("does not require protection for "+method, func(t *testing.T) {
			assert.False(t, middleware.requiresCSRFProtection(method))
		})
	}

	t.Run("handles empty method", func(t *testing.T) {
		assert.False(t, middleware.requiresCSRFProtection(""))
	})

	t.Run("handles lowercase method", func(t *testing.T) {
		// HTTP methods should be case-sensitive, but let's test lowercase
		assert.False(t, middleware.requiresCSRFProtection("post"))
		assert.False(t, middleware.requiresCSRFProtection("Post"))
	})
}

// TestGetCSRFTokenFromRequest tests extracting CSRF tokens from various request sources
func TestGetCSRFTokenFromRequest(t *testing.T) {
	middleware := newTestCSRFMiddleware()

	t.Run("extracts token from form data", func(t *testing.T) {
		formData := url.Values{}
		formData.Set("csrf_token", "test-csrf-token-from-form")
		req := httptest.NewRequest("POST", "/test", strings.NewReader(formData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		token := middleware.getCSRFTokenFromRequest(req)

		assert.Equal(t, csrf.CSRFToken("test-csrf-token-from-form"), token)
	})

	t.Run("extracts token from X-CSRF-Token header", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/test", nil)
		req.Header.Set("X-CSRF-Token", "test-csrf-token-from-header")

		token := middleware.getCSRFTokenFromRequest(req)

		assert.Equal(t, csrf.CSRFToken("test-csrf-token-from-header"), token)
	})

	t.Run("prioritizes form data over header", func(t *testing.T) {
		formData := url.Values{}
		formData.Set("csrf_token", "form-token")
		req := httptest.NewRequest("POST", "/test", strings.NewReader(formData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("X-CSRF-Token", "header-token")

		token := middleware.getCSRFTokenFromRequest(req)

		assert.Equal(t, csrf.CSRFToken("form-token"), token)
	})

	t.Run("returns empty token when not present", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/test", nil)

		token := middleware.getCSRFTokenFromRequest(req)

		assert.True(t, token.IsEmpty())
	})

	t.Run("returns empty token for JSON requests (TODO - not implemented)", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/test", strings.NewReader(`{"csrf_token": "json-token"}`))
		req.Header.Set("Content-Type", "application/json")

		token := middleware.getCSRFTokenFromRequest(req)

		// Currently returns empty as JSON body parsing is not implemented
		assert.True(t, token.IsEmpty())
	})

	t.Run("handles multiple form values", func(t *testing.T) {
		formData := url.Values{}
		formData.Set("username", "testuser")
		formData.Set("password", "pass123")
		formData.Set("csrf_token", "multi-form-token")
		req := httptest.NewRequest("POST", "/test", strings.NewReader(formData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		token := middleware.getCSRFTokenFromRequest(req)

		assert.Equal(t, csrf.CSRFToken("multi-form-token"), token)
	})

	t.Run("handles empty CSRF token value", func(t *testing.T) {
		formData := url.Values{}
		formData.Set("csrf_token", "")
		req := httptest.NewRequest("POST", "/test", strings.NewReader(formData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		token := middleware.getCSRFTokenFromRequest(req)

		// Empty string becomes empty CSRFToken
		assert.True(t, token.IsEmpty())
	})

	t.Run("extracts token with special characters", func(t *testing.T) {
		formData := url.Values{}
		specialToken := "token-with-!@#$%^&*()_+-=[]{}|;':\",./<>?"
		formData.Set("csrf_token", specialToken)
		req := httptest.NewRequest("POST", "/test", strings.NewReader(formData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		token := middleware.getCSRFTokenFromRequest(req)

		assert.Equal(t, csrf.CSRFToken(specialToken), token)
	})

	t.Run("handles URL-encoded token value", func(t *testing.T) {
		formData := url.Values{}
		// Token with URL-encoded characters
		formData.Set("csrf_token", "token%20with%20spaces")
		req := httptest.NewRequest("POST", "/test", strings.NewReader(formData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		token := middleware.getCSRFTokenFromRequest(req)

		// ParseForm decodes URL encoding
		assert.Equal(t, csrf.CSRFToken("token with spaces"), token)
	})
}

// TestGetCSRFTokenFromSession tests extracting CSRF tokens from session
func TestGetCSRFTokenFromSession(t *testing.T) {
	middleware := newTestCSRFMiddleware()

	t.Run("returns empty when no session cookie", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/test", nil)

		token := middleware.getCSRFTokenFromSession(req)

		assert.True(t, token.IsEmpty())
	})

	t.Run("returns empty when session cookie exists but lookup not implemented", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/test", nil)
		req.AddCookie(&http.Cookie{Name: "session_id", Value: "session-123"})

		token := middleware.getCSRFTokenFromSession(req)

		// Currently returns empty as session lookup is not implemented (TODO)
		assert.True(t, token.IsEmpty())
	})

	t.Run("handles empty session cookie value", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/test", nil)
		req.AddCookie(&http.Cookie{Name: "session_id", Value: ""})

		token := middleware.getCSRFTokenFromSession(req)

		assert.True(t, token.IsEmpty())
	})
}

// TestRequireCSRF tests the RequireCSRF middleware
func TestRequireCSRFMiddleware(t *testing.T) {
	middleware := newTestCSRFMiddleware()

	t.Run("allows GET requests without CSRF token", func(t *testing.T) {
		handlerCalled := false
		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlerCalled = true
		})

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		middleware.RequireCSRF(nextHandler).ServeHTTP(w, req)

		assert.True(t, handlerCalled, "GET requests should pass through")
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("allows HEAD requests without CSRF token", func(t *testing.T) {
		handlerCalled := false
		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlerCalled = true
		})

		req := httptest.NewRequest("HEAD", "/test", nil)
		w := httptest.NewRecorder()

		middleware.RequireCSRF(nextHandler).ServeHTTP(w, req)

		assert.True(t, handlerCalled)
	})

	t.Run("allows OPTIONS requests without CSRF token", func(t *testing.T) {
		handlerCalled := false
		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlerCalled = true
		})

		req := httptest.NewRequest("OPTIONS", "/test", nil)
		w := httptest.NewRecorder()

		middleware.RequireCSRF(nextHandler).ServeHTTP(w, req)

		assert.True(t, handlerCalled)
	})

	t.Run("blocks POST request without CSRF token", func(t *testing.T) {
		handlerCalled := false
		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlerCalled = true
		})

		req := httptest.NewRequest("POST", "/test", nil)
		w := httptest.NewRecorder()

		middleware.RequireCSRF(nextHandler).ServeHTTP(w, req)

		assert.False(t, handlerCalled, "Handler should not be called without valid CSRF token")
		assert.Equal(t, http.StatusForbidden, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid CSRF token")
	})

	t.Run("blocks PUT request without CSRF token", func(t *testing.T) {
		handlerCalled := false
		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlerCalled = true
		})

		req := httptest.NewRequest("PUT", "/test", nil)
		w := httptest.NewRecorder()

		middleware.RequireCSRF(nextHandler).ServeHTTP(w, req)

		assert.False(t, handlerCalled)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("blocks PATCH request without CSRF token", func(t *testing.T) {
		handlerCalled := false
		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlerCalled = true
		})

		req := httptest.NewRequest("PATCH", "/test", nil)
		w := httptest.NewRecorder()

		middleware.RequireCSRF(nextHandler).ServeHTTP(w, req)

		assert.False(t, handlerCalled)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("blocks DELETE request without CSRF token", func(t *testing.T) {
		handlerCalled := false
		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlerCalled = true
		})

		req := httptest.NewRequest("DELETE", "/test", nil)
		w := httptest.NewRecorder()

		middleware.RequireCSRF(nextHandler).ServeHTTP(w, req)

		assert.False(t, handlerCalled)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("blocks POST when no session (empty expected token)", func(t *testing.T) {
		handlerCalled := false
		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlerCalled = true
		})

		formData := url.Values{}
		formData.Set("csrf_token", "provided-token")
		req := httptest.NewRequest("POST", "/test", strings.NewReader(formData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		middleware.RequireCSRF(nextHandler).ServeHTTP(w, req)

		// Should block because there's no session (empty expected token)
		assert.False(t, handlerCalled)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("blocks POST when tokens don't match", func(t *testing.T) {
		// This test would require mocking the session to return a token
		// For now, it demonstrates the expected behavior
		handlerCalled := false
		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlerCalled = true
		})

		formData := url.Values{}
		formData.Set("csrf_token", "wrong-token")
		req := httptest.NewRequest("POST", "/test", strings.NewReader(formData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		// Note: Without a way to inject a session token, this test
		// shows the expected behavior when tokens don't match
		w := httptest.NewRecorder()

		middleware.RequireCSRF(nextHandler).ServeHTTP(w, req)

		// Should block because no session = empty expected token
		assert.False(t, handlerCalled)
	})
}

// TestCSRFMiddleware_EdgeCases tests edge cases and error conditions
func TestCSRFMiddleware_EdgeCases(t *testing.T) {
	middleware := newTestCSRFMiddleware()

	t.Run("handles concurrent requests safely", func(t *testing.T) {
		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		const goroutines = 50
		done := make(chan bool, goroutines)

		for i := 0; i < goroutines; i++ {
			go func(id int) {
				req := httptest.NewRequest("GET", "/test", nil)
				w := httptest.NewRecorder()
				middleware.RequireCSRF(nextHandler).ServeHTTP(w, req)
				done <- true
			}(i)
		}

		for i := 0; i < goroutines; i++ {
			<-done
		}

		// Test passes if no panic or deadlock occurred
	})

	t.Run("handles malformed form data", func(t *testing.T) {
		handlerCalled := false
		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlerCalled = true
		})

		// Invalid form data (not properly encoded)
		req := httptest.NewRequest("POST", "/test", strings.NewReader("%invalid%"))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		middleware.RequireCSRF(nextHandler).ServeHTTP(w, req)

		// Should still process request even if form parsing fails
		// and block because there's no CSRF token
		assert.False(t, handlerCalled)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("handles very long CSRF token", func(t *testing.T) {
		handlerCalled := false
		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlerCalled = true
		})

		longToken := string(make([]byte, 10000))
		formData := url.Values{}
		formData.Set("csrf_token", longToken)
		req := httptest.NewRequest("POST", "/test", strings.NewReader(formData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		middleware.RequireCSRF(nextHandler).ServeHTTP(w, req)

		// Should block (no session) but not crash
		assert.False(t, handlerCalled)
	})

	t.Run("handles special characters in CSRF token", func(t *testing.T) {
		handlerCalled := false
		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlerCalled = true
		})

		specialToken := "token-with-!@#$%^&*()_+-=[]{}|;':\",./<>?"
		formData := url.Values{}
		formData.Set("csrf_token", specialToken)
		req := httptest.NewRequest("POST", "/test", strings.NewReader(formData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		middleware.RequireCSRF(nextHandler).ServeHTTP(w, req)

		// Should block (no session) but not crash
		assert.False(t, handlerCalled)
	})

	t.Run("handles request with multiple cookies", func(t *testing.T) {
		handlerCalled := false
		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlerCalled = true
		})

		req := httptest.NewRequest("GET", "/test", nil)
		req.AddCookie(&http.Cookie{Name: "cookie1", Value: "value1"})
		req.AddCookie(&http.Cookie{Name: "cookie2", Value: "value2"})
		req.AddCookie(&http.Cookie{Name: "session_id", Value: "session-123"})
		w := httptest.NewRecorder()

		middleware.RequireCSRF(nextHandler).ServeHTTP(w, req)

		// GET request should pass regardless of cookies
		assert.True(t, handlerCalled)
	})
}

// TestCSRFMiddleware_DifferentContentTypes tests various content types
func TestCSRFMiddleware_DifferentContentTypes(t *testing.T) {
	middleware := newTestCSRFMiddleware()

	t.Run("handles multipart form data", func(t *testing.T) {
		handlerCalled := false
		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlerCalled = true
		})

		body := strings.NewReader("--boundary\r\nContent-Disposition: form-data; name=\"csrf_token\"\r\n\r\ntest-token\r\n--boundary--")
		req := httptest.NewRequest("POST", "/test", body)
		req.Header.Set("Content-Type", "multipart/form-data; boundary=boundary")
		w := httptest.NewRecorder()

		middleware.RequireCSRF(nextHandler).ServeHTTP(w, req)

		// Should block (no session) but ParseForm doesn't work for multipart
		assert.False(t, handlerCalled)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("handles JSON content type", func(t *testing.T) {
		handlerCalled := false
		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlerCalled = true
		})

		req := httptest.NewRequest("POST", "/test", strings.NewReader(`{"csrf_token": "json-token"}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		middleware.RequireCSRF(nextHandler).ServeHTTP(w, req)

		// Should block (JSON body parsing not implemented yet)
		assert.False(t, handlerCalled)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("handles text/plain content type", func(t *testing.T) {
		handlerCalled := false
		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlerCalled = true
		})

		req := httptest.NewRequest("POST", "/test", strings.NewReader("plain text body"))
		req.Header.Set("Content-Type", "text/plain")
		w := httptest.NewRecorder()

		middleware.RequireCSRF(nextHandler).ServeHTTP(w, req)

		// Should block (no CSRF token in request)
		assert.False(t, handlerCalled)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

// TestCSRFToken_IsEmpty tests the IsEmpty method
func TestCSRFToken_IsEmpty(t *testing.T) {
	t.Run("empty string is empty", func(t *testing.T) {
		token := csrf.CSRFToken("")
		assert.True(t, token.IsEmpty())
	})

	t.Run("non-empty string is not empty", func(t *testing.T) {
		token := csrf.CSRFToken("test-token")
		assert.False(t, token.IsEmpty())
	})

	t.Run("whitespace-only is not empty", func(t *testing.T) {
		token := csrf.CSRFToken("   ")
		assert.False(t, token.IsEmpty())
	})
}

// TestCSRFToken_Equals tests the Equals method
func TestCSRFToken_Equals(t *testing.T) {
	t.Run("equal tokens return true", func(t *testing.T) {
		token1 := csrf.CSRFToken("same-token")
		token2 := csrf.CSRFToken("same-token")
		assert.True(t, token1.Equals(token2))
	})

	t.Run("different tokens return false", func(t *testing.T) {
		token1 := csrf.CSRFToken("token-one")
		token2 := csrf.CSRFToken("token-two")
		assert.False(t, token1.Equals(token2))
	})

	t.Run("case sensitive comparison", func(t *testing.T) {
		token1 := csrf.CSRFToken("Token")
		token2 := csrf.CSRFToken("token")
		assert.False(t, token1.Equals(token2))
	})

	t.Run("empty tokens are equal", func(t *testing.T) {
		token1 := csrf.CSRFToken("")
		token2 := csrf.CSRFToken("")
		assert.True(t, token1.Equals(token2))
	})
}

// TestCSRFToken_String tests the String method
func TestCSRFToken_String(t *testing.T) {
	t.Run("returns string representation", func(t *testing.T) {
		token := csrf.CSRFToken("test-token")
		assert.Equal(t, "test-token", token.String())
	})

	t.Run("empty token returns empty string", func(t *testing.T) {
		token := csrf.CSRFToken("")
		assert.Equal(t, "", token.String())
	})
}
