package auth

import (
	"context"
	"testing"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth"
)

func TestNewAdminApiKeyAuthMethod(t *testing.T) {
	authMethod := NewAdminApiKeyAuthMethod()

	if authMethod == nil {
		t.Fatal("NewAdminApiKeyAuthMethod returned nil")
	}

	if authMethod.GetProviderType() != auth.AuthMethodAdminApiKey {
		t.Errorf("Expected provider type %v, got %v", auth.AuthMethodAdminApiKey, authMethod.GetProviderType())
	}

	if authMethod.CheckAuthState() {
		t.Error("Expected initial auth state to be false")
	}

	if authMethod.IsAdmin() {
		t.Error("Expected initial admin status to be false")
	}

	if authMethod.IsAuthenticated() {
		t.Error("Expected initial authenticated status to be false")
	}
}

func TestAdminApiKeyAuthMethod_Authenticate_ValidKey(t *testing.T) {
	authMethod := NewAdminApiKeyAuthMethod()

	headers := map[string]string{
		"Authorization": "Bearer admin-api-key-12345",
	}

	err := authMethod.Authenticate(context.Background(), headers, map[string]string{})

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if !authMethod.CheckAuthState() {
		t.Error("Expected auth state to be true after valid authentication")
	}

	if !authMethod.IsAdmin() {
		t.Error("Expected admin status to be true after valid authentication")
	}

	if !authMethod.IsAuthenticated() {
		t.Error("Expected authenticated status to be true after valid authentication")
	}

	expectedUsername := "Admin API Key User"
	if authMethod.GetUsername() != expectedUsername {
		t.Errorf("Expected username %s, got %s", expectedUsername, authMethod.GetUsername())
	}
}

func TestAdminApiKeyAuthMethod_Authenticate_InvalidKey(t *testing.T) {
	authMethod := NewAdminApiKeyAuthMethod()

	headers := map[string]string{
		"Authorization": "Bearer invalid-key",
	}

	err := authMethod.Authenticate(context.Background(), headers, map[string]string{})

	if err == nil {
		t.Error("Expected error for invalid key")
	}

	if authMethod.CheckAuthState() {
		t.Error("Expected auth state to remain false for invalid key")
	}

	if authMethod.IsAdmin() {
		t.Error("Expected admin status to remain false for invalid key")
	}

	if authMethod.IsAuthenticated() {
		t.Error("Expected authenticated status to remain false for invalid key")
	}
}

func TestAdminApiKeyAuthMethod_Authenticate_MissingHeader(t *testing.T) {
	authMethod := NewAdminApiKeyAuthMethod()

	headers := map[string]string{}

	err := authMethod.Authenticate(context.Background(), headers, map[string]string{})

	if err == nil {
		t.Error("Expected error for missing authorization header")
	}

	if authMethod.CheckAuthState() {
		t.Error("Expected auth state to remain false for missing header")
	}
}

func TestAdminApiKeyAuthMethod_Authenticate_InvalidFormat(t *testing.T) {
	authMethod := NewAdminApiKeyAuthMethod()

	tests := []struct {
		name   string
		header string
	}{
		{"No Bearer prefix", "admin-api-key-12345"},
		{"Empty Bearer", "Bearer "},
		{"Wrong prefix", "Token admin-api-key-12345"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			headers := map[string]string{
				"Authorization": test.header,
			}

			err := authMethod.Authenticate(context.Background(), headers, map[string]string{})

			if err == nil {
				t.Errorf("Expected error for invalid format: %s", test.header)
			}
		})
	}
}

func TestAdminApiKeyAuthMethod_CanPublishModuleVersion(t *testing.T) {
	authMethod := NewAdminApiKeyAuthMethod()

	// Before authentication
	if authMethod.CanPublishModuleVersion("any-namespace") {
		t.Error("Expected CanPublishModuleVersion to return false before authentication")
	}

	// After successful authentication
	headers := map[string]string{
		"Authorization": "Bearer admin-api-key-12345",
	}
	_ = authMethod.Authenticate(context.Background(), headers, map[string]string{})

	if !authMethod.CanPublishModuleVersion("any-namespace") {
		t.Error("Expected CanPublishModuleVersion to return true for admin after authentication")
	}
}

func TestAdminApiKeyAuthMethod_CanUploadModuleVersion(t *testing.T) {
	authMethod := NewAdminApiKeyAuthMethod()

	// Before authentication
	if authMethod.CanUploadModuleVersion("any-namespace") {
		t.Error("Expected CanUploadModuleVersion to return false before authentication")
	}

	// After successful authentication
	headers := map[string]string{
		"Authorization": "Bearer admin-api-key-12345",
	}
	_ = authMethod.Authenticate(context.Background(), headers, map[string]string{})

	if !authMethod.CanUploadModuleVersion("any-namespace") {
		t.Error("Expected CanUploadModuleVersion to return true for admin after authentication")
	}
}

func TestAdminApiKeyAuthMethod_CheckNamespaceAccess(t *testing.T) {
	authMethod := NewAdminApiKeyAuthMethod()

	// Before authentication
	if authMethod.CheckNamespaceAccess("READ", "any-namespace") {
		t.Error("Expected CheckNamespaceAccess to return false before authentication")
	}

	// After successful authentication
	headers := map[string]string{
		"Authorization": "Bearer admin-api-key-12345",
	}
	_ = authMethod.Authenticate(context.Background(), headers, map[string]string{})

	tests := []struct {
		permission string
		namespace  string
		expected   bool
	}{
		{"READ", "any-namespace", true},
		{"MODIFY", "any-namespace", true},
		{"FULL", "any-namespace", true},
		{"READ", "different-namespace", true},
		{"MODIFY", "different-namespace", true},
		{"FULL", "different-namespace", true},
	}

	for _, test := range tests {
		result := authMethod.CheckNamespaceAccess(test.permission, test.namespace)
		if result != test.expected {
			t.Errorf("Expected CheckNamespaceAccess(%s, %s) to be %v, got %v",
				test.permission, test.namespace, test.expected, result)
		}
	}
}

func TestAdminApiKeyAuthMethod_GetAllNamespacePermissions(t *testing.T) {
	authMethod := NewAdminApiKeyAuthMethod()

	// Before authentication
	permissions := authMethod.GetAllNamespacePermissions()
	if len(permissions) != 0 {
		t.Error("Expected no permissions before authentication")
	}

	// After successful authentication
	headers := map[string]string{
		"Authorization": "Bearer admin-api-key-12345",
	}
	_ = authMethod.Authenticate(context.Background(), headers, map[string]string{})

	permissions = authMethod.GetAllNamespacePermissions()
	if len(permissions) != 0 {
		t.Error("Expected empty permissions map for admin (signifies full access)")
	}
}

func TestAdminApiKeyAuthMethod_GetUserGroupNames(t *testing.T) {
	authMethod := NewAdminApiKeyAuthMethod()

	groups := authMethod.GetUserGroupNames()
	expected := []string{"admin"}

	if len(groups) != 1 || groups[0] != expected[0] {
		t.Errorf("Expected groups %v, got %v", expected, groups)
	}
}

func TestAdminApiKeyAuthMethod_CanAccessReadAPI(t *testing.T) {
	authMethod := NewAdminApiKeyAuthMethod()

	// Before authentication
	if authMethod.CanAccessReadAPI() {
		t.Error("Expected CanAccessReadAPI to return false before authentication")
	}

	// After successful authentication
	headers := map[string]string{
		"Authorization": "Bearer admin-api-key-12345",
	}
	_ = authMethod.Authenticate(context.Background(), headers, map[string]string{})

	if !authMethod.CanAccessReadAPI() {
		t.Error("Expected CanAccessReadAPI to return true after authentication")
	}
}

func TestAdminApiKeyAuthMethod_CanAccessTerraformAPI(t *testing.T) {
	authMethod := NewAdminApiKeyAuthMethod()

	// Before authentication
	if authMethod.CanAccessTerraformAPI() {
		t.Error("Expected CanAccessTerraformAPI to return false before authentication")
	}

	// After successful authentication
	headers := map[string]string{
		"Authorization": "Bearer admin-api-key-12345",
	}
	_ = authMethod.Authenticate(context.Background(), headers, map[string]string{})

	if !authMethod.CanAccessTerraformAPI() {
		t.Error("Expected CanAccessTerraformAPI to return true for admin after authentication")
	}
}

func TestAdminApiKeyAuthMethod_GetProviderData(t *testing.T) {
	authMethod := NewAdminApiKeyAuthMethod()

	// Before authentication
	data := authMethod.GetProviderData()
	if data == nil {
		t.Error("Expected provider data to be non-nil")
	}

	if _, exists := data["api_key"]; exists {
		t.Error("Expected api_key to not exist in provider data before authentication")
	}

	// After successful authentication
	headers := map[string]string{
		"Authorization": "Bearer admin-api-key-12345",
	}
	_ = authMethod.Authenticate(context.Background(), headers, map[string]string{})

	data = authMethod.GetProviderData()
	if data == nil {
		t.Error("Expected provider data to be non-nil")
	}

	if apiKey, exists := data["api_key"]; !exists || apiKey != "admin-api-key-12345" {
		t.Errorf("Expected api_key to be 'admin-api-key-12345', got %v", apiKey)
	}

	if isAdmin, exists := data["is_admin"]; !exists || !isAdmin.(bool) {
		t.Error("Expected is_admin to be true")
	}
}

func TestAdminApiKeyAuthMethod_IsEnabled(t *testing.T) {
	authMethod := NewAdminApiKeyAuthMethod()

	if !authMethod.IsEnabled() {
		t.Error("Expected IsEnabled to return true")
	}
}

func TestAdminApiKeyAuthMethod_RequiresCSRF(t *testing.T) {
	authMethod := NewAdminApiKeyAuthMethod()

	if authMethod.RequiresCSRF() {
		t.Error("Expected RequiresCSRF to return false for API key authentication")
	}
}

func TestAdminApiKeyAuthMethod_IsBuiltInAdmin(t *testing.T) {
	authMethod := NewAdminApiKeyAuthMethod()

	if !authMethod.IsBuiltInAdmin() {
		t.Error("Expected IsBuiltInAdmin to return true for admin API key method")
	}
}

func TestAdminApiKeyAuthMethod_GetTerraformAuthToken(t *testing.T) {
	authMethod := NewAdminApiKeyAuthMethod()

	// Before authentication
	token := authMethod.GetTerraformAuthToken()
	if token != "" {
		t.Error("Expected empty token before authentication")
	}

	// After successful authentication
	headers := map[string]string{
		"Authorization": "Bearer admin-api-key-12345",
	}
	_ = authMethod.Authenticate(context.Background(), headers, map[string]string{})

	token = authMethod.GetTerraformAuthToken()
	if token != "admin-api-key-12345" {
		t.Errorf("Expected token to be 'admin-api-key-12345', got %s", token)
	}
}

func TestAdminApiKeyError(t *testing.T) {
	err := adminApiKeyErr("test error")
	if err == nil {
		t.Error("Expected error to be created")
	}

	adminErr, ok := err.(*AdminApiKeyError)
	if !ok {
		t.Error("Expected error to be of type AdminApiKeyError")
	}

	if adminErr.Message != "test error" {
		t.Errorf("Expected error message 'test error', got %s", adminErr.Message)
	}

	expectedErrorString := "Admin API Key authentication failed: test error"
	if adminErr.Error() != expectedErrorString {
		t.Errorf("Expected error string '%s', got %s", expectedErrorString, adminErr.Error())
	}
}
