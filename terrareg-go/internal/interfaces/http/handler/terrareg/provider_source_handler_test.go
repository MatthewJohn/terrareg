package terrareg

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	authservice "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth/service"
	provider_source_service "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider_source/service"
	provider_source_model "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider_source/model"
)

// MockProviderSourceFactory for testing
type MockProviderSourceFactory struct {
	providerSource provider_source_service.ProviderSourceInstance
	returnError    bool
}

func (m *MockProviderSourceFactory) GetProviderSourceByApiName(ctx context.Context, apiName string) (provider_source_service.ProviderSourceInstance, error) {
	if m.returnError || m.providerSource == nil {
		return nil, nil
	}
	return m.providerSource, nil
}

func (m *MockProviderSourceFactory) GetProviderSourceByName(ctx context.Context, name string) (*provider_source_model.ProviderSource, error) {
	if m.returnError || m.providerSource == nil {
		return nil, nil
	}
	// Return a mock provider source
	config := &provider_source_model.ProviderSourceConfig{
		BaseURL:      "https://github.com",
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
	}
	ps := provider_source_model.NewProviderSource("Test GitHub", name, provider_source_model.ProviderSourceTypeGithub, config)
	return ps, nil
}

// MockProviderSourceInstance for testing
type MockProviderSourceInstance struct {
	loginRedirectURL     string
	accessToken          string
	username             string
	organizations        []string
	shouldFailToken      bool
	shouldFailUsername   bool
}

func (m *MockProviderSourceInstance) Name() string {
	return "Test GitHub"
}

func (m *MockProviderSourceInstance) ApiName() string {
	return "test-github"
}

func (m *MockProviderSourceInstance) Type() provider_source_model.ProviderSourceType {
	return provider_source_model.ProviderSourceTypeGithub
}

func (m *MockProviderSourceInstance) GetLoginRedirectURL(ctx context.Context) (string, error) {
	if m.loginRedirectURL != "" {
		return m.loginRedirectURL, nil
	}
	return "https://github.com/login/oauth/authorize?client_id=test-client-id&state=12345&scope=read:org", nil
}

func (m *MockProviderSourceInstance) GetUserAccessToken(ctx context.Context, code string) (string, error) {
	if m.shouldFailToken {
		return "", nil
	}
	if m.accessToken != "" {
		return m.accessToken, nil
	}
	return "test-access-token", nil
}

func (m *MockProviderSourceInstance) GetUsername(ctx context.Context, accessToken string) (string, error) {
	if m.shouldFailUsername {
		return "", nil
	}
	if m.username != "" {
		return m.username, nil
	}
	return "test-user", nil
}

func (m *MockProviderSourceInstance) GetUserOrganizations(ctx context.Context, accessToken string) []string {
	if m.organizations != nil {
		return m.organizations
	}
	return []string{"test-org-1", "test-org-2"}
}

// MockAuthenticationService for testing
type MockAuthenticationService struct {
	createSessionFunc func(ctx context.Context, w http.ResponseWriter, authMethod string, providerData map[string]interface{}, ttl *time.Duration) error
	validateFunc      func(ctx context.Context, r *http.Request) (*authservice.AuthenticationContext, error)
}

func (m *MockAuthenticationService) CreateAuthenticatedSession(ctx context.Context, w http.ResponseWriter, authMethod string, providerData map[string]interface{}, ttl *time.Duration) error {
	if m.createSessionFunc != nil {
		return m.createSessionFunc(ctx, w, authMethod, providerData, ttl)
	}
	return nil
}

func (m *MockAuthenticationService) ValidateRequest(ctx context.Context, r *http.Request) (*authservice.AuthenticationContext, error) {
	if m.validateFunc != nil {
		return m.validateFunc(ctx, r)
	}
	// Return unauthenticated context by default
	return &authservice.AuthenticationContext{
		IsAuthenticated: false,
	}, nil
}

func (m *MockAuthenticationService) InvalidateSession(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (m *MockAuthenticationService) RefreshSession(ctx context.Context, w http.ResponseWriter, r *http.Request, ttl time.Duration) error {
	return nil
}

func (m *MockAuthenticationService) CreateSession(ctx context.Context, w http.ResponseWriter, sessionID string) error {
	return nil
}

func (m *MockAuthenticationService) ClearSession(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (m *MockAuthenticationService) CreateAdminSession(ctx context.Context, w http.ResponseWriter, sessionID string) error {
	return nil
}

// TestNewProviderSourceHandler tests the constructor
func TestNewProviderSourceHandler(t *testing.T) {
	factory := &MockProviderSourceFactory{}
	authService := &MockAuthenticationService{}

	handler := NewProviderSourceHandler(factory, authService)

	if handler == nil {
		t.Fatal("NewProviderSourceHandler returned nil")
	}

	if handler.providerSourceFactory == nil {
		t.Error("providerSourceFactory not set correctly")
	}

	if handler.authService == nil {
		t.Error("authService not set correctly")
	}
}

// TestProviderSourceHandler_HandleLogin tests the login handler
func TestProviderSourceHandler_HandleLogin(t *testing.T) {
	tests := []struct {
		name               string
		providerSource     string
		setupFactory       func(*MockProviderSourceFactory)
		expectedStatusCode int
		expectedRedirect   bool
	}{
		{
			name:           "successful login",
			providerSource: "test-github",
			setupFactory: func(f *MockProviderSourceFactory) {
				f.providerSource = &MockProviderSourceInstance{
					loginRedirectURL: "https://github.com/login/oauth/authorize?client_id=test-client-id",
				}
			},
			expectedStatusCode: http.StatusFound,
			expectedRedirect:   true,
		},
		{
			name:               "missing provider source",
			providerSource:     "",
			setupFactory:       func(f *MockProviderSourceFactory) {},
			expectedStatusCode: http.StatusBadRequest,
			expectedRedirect:   false,
		},
		{
			name:           "provider source not found",
			providerSource: "non-existent",
			setupFactory: func(f *MockProviderSourceFactory) {
				f.returnError = true
			},
			expectedStatusCode: http.StatusNotFound,
			expectedRedirect:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factory := &MockProviderSourceFactory{}
			tt.setupFactory(factory)

			handler := NewProviderSourceHandler(factory, &MockAuthenticationService{})

			req := httptest.NewRequest("GET", "/"+tt.providerSource+"/login", nil)
			w := httptest.NewRecorder()

			// Set chi context
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("provider_source", tt.providerSource)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			handler.HandleLogin(w, req)

			if w.Code != tt.expectedStatusCode {
				t.Errorf("HandleLogin() status = %d, want %d", w.Code, tt.expectedStatusCode)
			}

			isRedirect := w.Header().Get("Location") != ""
			if isRedirect != tt.expectedRedirect {
				t.Errorf("HandleLogin() redirect = %v, want %v", isRedirect, tt.expectedRedirect)
			}
		})
	}
}

// TestProviderSourceHandler_HandleLogin_RedirectURL tests the redirect URL
func TestProviderSourceHandler_HandleLogin_RedirectURL(t *testing.T) {
	expectedURL := "https://github.com/login/oauth/authorize?client_id=test-client-id&state=12345&scope=read:org"

	factory := &MockProviderSourceFactory{
		providerSource: &MockProviderSourceInstance{
			loginRedirectURL: expectedURL,
		},
	}

	handler := NewProviderSourceHandler(factory, &MockAuthenticationService{})

	req := httptest.NewRequest("GET", "/test-github/login", nil)
	w := httptest.NewRecorder()

	// Set chi context
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("provider_source", "test-github")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	handler.HandleLogin(w, req)

	location := w.Header().Get("Location")
	if location != expectedURL {
		t.Errorf("HandleLogin() redirect URL = %s, want %s", location, expectedURL)
	}
}

// TestProviderSourceHandler_HandleCallback_Success tests successful callback
func TestProviderSourceHandler_HandleCallback_Success(t *testing.T) {
	sessionCreated := false
	expectedUsername := "test-user"
	expectedOrgs := []string{"test-org-1", "test-org-2"}

	factory := &MockProviderSourceFactory{
		providerSource: &MockProviderSourceInstance{
			username:      expectedUsername,
			organizations: expectedOrgs,
		},
	}

	authService := &MockAuthenticationService{
		createSessionFunc: func(ctx context.Context, w http.ResponseWriter, authMethod string, providerData map[string]interface{}, ttl *time.Duration) error {
			sessionCreated = true

			// Verify provider data
			if providerData["provider_source"] != "test-github" {
				t.Errorf("provider_data[provider_source] = %v, want test-github", providerData["provider_source"])
			}

			if providerData["github_username"] != expectedUsername {
				t.Errorf("provider_data[github_username] = %v, want %s", providerData["github_username"], expectedUsername)
			}

			orgs, ok := providerData["organisations"].(map[string]string)
			if !ok {
				t.Error("provider_data[organisations] is not a map[string]string")
				return nil
			}

			// Check username is in organizations
			if orgs[expectedUsername] != "GITHUB_USER" {
				t.Errorf("organisations[%s] = %v, want GITHUB_USER", expectedUsername, orgs[expectedUsername])
			}

			// Check organizations are included
			for _, org := range expectedOrgs {
				if orgs[org] != "GITHUB_ORGANISATION" {
					t.Errorf("organisations[%s] = %v, want GITHUB_ORGANISATION", org, orgs[org])
				}
			}

			return nil
		},
	}

	handler := NewProviderSourceHandler(factory, authService)

	req := httptest.NewRequest("GET", "/test-github/callback?code=test-code", nil)
	w := httptest.NewRecorder()

	// Set chi context
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("provider_source", "test-github")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	handler.HandleCallback(w, req)

	if !sessionCreated {
		t.Error("HandleCallback() did not create session")
	}

	// Should redirect to home
	if w.Code != http.StatusFound {
		t.Errorf("HandleCallback() status = %d, want %d", w.Code, http.StatusFound)
	}

	location := w.Header().Get("Location")
	if location != "/" {
		t.Errorf("HandleCallback() redirect location = %s, want /", location)
	}
}

// TestProviderSourceHandler_HandleCallback_MissingCode tests callback without code
func TestProviderSourceHandler_HandleCallback_MissingCode(t *testing.T) {
	factory := &MockProviderSourceFactory{
		providerSource: &MockProviderSourceInstance{},
	}

	handler := NewProviderSourceHandler(factory, &MockAuthenticationService{})

	req := httptest.NewRequest("GET", "/test-github/callback", nil)
	w := httptest.NewRecorder()

	// Set chi context
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("provider_source", "test-github")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	handler.HandleCallback(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("HandleCallback() status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

// TestProviderSourceHandler_HandleCallback_MissingProviderSource tests callback without provider source
func TestProviderSourceHandler_HandleCallback_MissingProviderSource(t *testing.T) {
	handler := NewProviderSourceHandler(&MockProviderSourceFactory{}, &MockAuthenticationService{})

	req := httptest.NewRequest("GET", "//callback?code=test-code", nil)
	w := httptest.NewRecorder()

	// Set chi context
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("provider_source", "")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	handler.HandleCallback(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("HandleCallback() status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

// TestProviderSourceHandler_HandleCallback_ProviderSourceNotFound tests callback with non-existent provider source
func TestProviderSourceHandler_HandleCallback_ProviderSourceNotFound(t *testing.T) {
	factory := &MockProviderSourceFactory{
		returnError: true,
	}

	handler := NewProviderSourceHandler(factory, &MockAuthenticationService{})

	req := httptest.NewRequest("GET", "/non-existent/callback?code=test-code", nil)
	w := httptest.NewRecorder()

	// Set chi context
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("provider_source", "non-existent")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	handler.HandleCallback(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("HandleCallback() status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

// TestProviderSourceHandler_HandleCallback_FailedTokenExchange tests failed token exchange
func TestProviderSourceHandler_HandleCallback_FailedTokenExchange(t *testing.T) {
	factory := &MockProviderSourceFactory{
		providerSource: &MockProviderSourceInstance{
			shouldFailToken: true,
		},
	}

	handler := NewProviderSourceHandler(factory, &MockAuthenticationService{})

	req := httptest.NewRequest("GET", "/test-github/callback?code=test-code", nil)
	w := httptest.NewRecorder()

	// Set chi context
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("provider_source", "test-github")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	handler.HandleCallback(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("HandleCallback() status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

// TestProviderSourceHandler_HandleAuthStatus tests the auth status endpoint
func TestProviderSourceHandler_HandleAuthStatus(t *testing.T) {
	tests := []struct {
		name               string
		providerSource     string
		isAuthenticated    bool
		expectedStatusCode int
	}{
		{
			name:               "authenticated user",
			providerSource:     "test-github",
			isAuthenticated:    true,
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "unauthenticated user",
			providerSource:     "test-github",
			isAuthenticated:    false,
			expectedStatusCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factory := &MockProviderSourceFactory{
				providerSource: &MockProviderSourceInstance{},
			}

			authService := &MockAuthenticationService{
				validateFunc: func(ctx context.Context, r *http.Request) (*authservice.AuthenticationContext, error) {
					return &authservice.AuthenticationContext{
						IsAuthenticated: tt.isAuthenticated,
						Username:        "test-user",
						AuthMethod:      "GITHUB",
					}, nil
				},
			}

			handler := NewProviderSourceHandler(factory, authService)

			req := httptest.NewRequest("GET", "/"+tt.providerSource+"/auth/status", nil)
			w := httptest.NewRecorder()

			// Set chi context
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("provider_source", tt.providerSource)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			handler.HandleAuthStatus(w, req)

			if w.Code != tt.expectedStatusCode {
				t.Errorf("HandleAuthStatus() status = %d, want %d", w.Code, tt.expectedStatusCode)
			}

			// Verify response body
			var response map[string]interface{}
			if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
				t.Fatalf("Failed to decode response: %v", err)
			}

			authenticated, ok := response["authenticated"].(bool)
			if !ok {
				t.Error("response[authenticated] is not a bool")
			}

			if authenticated != tt.isAuthenticated {
				t.Errorf("response[authenticated] = %v, want %v", authenticated, tt.isAuthenticated)
			}

			providerType, ok := response["provider_type"].(string)
			if !ok {
				t.Error("response[provider_type] is not a string")
			}

			if providerType != string(provider_source_model.ProviderSourceTypeGithub) {
				t.Errorf("response[provider_type] = %s, want %s", providerType, provider_source_model.ProviderSourceTypeGithub)
			}

			if tt.isAuthenticated {
				username, ok := response["username"].(string)
				if !ok {
					t.Error("response[username] is not a string")
				}

				if username != "test-user" {
					t.Errorf("response[username] = %s, want test-user", username)
				}

				authMethod, ok := response["auth_method"].(string)
				if !ok {
					t.Error("response[auth_method] is not a string")
				}

				if authMethod != "GITHUB" {
					t.Errorf("response[auth_method] = %s, want GITHUB", authMethod)
				}
			}
		})
	}
}

// TestProviderSourceHandler_HandleAuthStatus_MissingProviderSource tests auth status without provider source
func TestProviderSourceHandler_HandleAuthStatus_MissingProviderSource(t *testing.T) {
	handler := NewProviderSourceHandler(&MockProviderSourceFactory{}, &MockAuthenticationService{})

	req := httptest.NewRequest("GET", "//auth/status", nil)
	w := httptest.NewRecorder()

	// Set chi context
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("provider_source", "")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	handler.HandleAuthStatus(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("HandleAuthStatus() status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

// TestProviderSourceHandler_HandleAuthStatus_ProviderSourceNotFound tests auth status with non-existent provider source
func TestProviderSourceHandler_HandleAuthStatus_ProviderSourceNotFound(t *testing.T) {
	factory := &MockProviderSourceFactory{
		returnError: true,
	}

	handler := NewProviderSourceHandler(factory, &MockAuthenticationService{})

	req := httptest.NewRequest("GET", "/non-existent/auth/status", nil)
	w := httptest.NewRecorder()

	// Set chi context
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("provider_source", "non-existent")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	handler.HandleAuthStatus(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("HandleAuthStatus() status = %d, want %d", w.Code, http.StatusNotFound)
	}
}
