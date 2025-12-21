package auth

import (
	"context"
	"testing"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/config"
)

func TestNewAdminApiKeyAuthMethod(t *testing.T) {
	authMethod := NewAdminApiKeyAuthMethod(&config.InfrastructureConfig{})

	if authMethod == nil {
		t.Fatal("NewAdminApiKeyAuthMethod returned nil")
	}

	if authMethod.GetProviderType() != auth.AuthMethodAdminApiKey {
		t.Errorf("Expected provider type %v, got %v", auth.AuthMethodAdminApiKey, authMethod.GetProviderType())
	}

	// Should be disabled when no token is configured
	if authMethod.IsEnabled() {
		t.Error("Expected auth method to be disabled when no token configured")
	}
}

func TestNewAdminApiKeyAuthMethod_WithToken(t *testing.T) {
	config := &config.InfrastructureConfig{
		AdminAuthenticationToken: "test-admin-token",
	}
	authMethod := NewAdminApiKeyAuthMethod(config)

	if !authMethod.IsEnabled() {
		t.Error("Expected auth method to be enabled when token is configured")
	}
}

func TestAdminApiKeyAuthMethod_Authenticate_ValidKey(t *testing.T) {
	config := &config.InfrastructureConfig{
	AdminAuthenticationToken: "test-admin-token",
	}
	authMethod := NewAdminApiKeyAuthMethod(config)

	headers := map[string]string{
		"X-Terrareg-ApiKey": "test-admin-token",
	}

	authContext, err := authMethod.Authenticate(context.Background(), headers, map[string]string{}, map[string]string{})

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if authContext == nil {
		t.Fatal("Expected auth context to be returned")
	}

	// Test auth context methods (not auth method methods)
	if !authContext.IsAuthenticated() {
		t.Error("Expected auth context to be authenticated")
	}

	if !authContext.IsAdmin() {
		t.Error("Expected auth context to be admin")
	}

	if authContext.GetUsername() != "admin-api-key" {
		t.Errorf("Expected username 'admin-api-key', got '%s'", authContext.GetUsername())
	}

	if authContext.GetProviderType() != auth.AuthMethodAdminApiKey {
		t.Errorf("Expected provider type %v, got %v", auth.AuthMethodAdminApiKey, authContext.GetProviderType())
	}
}

func TestAdminApiKeyAuthMethod_Authenticate_InvalidKey(t *testing.T) {
	config := &config.InfrastructureConfig{
		AdminAuthenticationToken: "correct-token",
	}
	authMethod := NewAdminApiKeyAuthMethod(config)

	headers := map[string]string{
		"X-Terrareg-ApiKey": "wrong-token",
	}

	authContext, err := authMethod.Authenticate(context.Background(), headers, map[string]string{}, map[string]string{})

	if err != nil {
		t.Errorf("Unexpected error (should return nil context): %v", err)
	}

	if authContext != nil {
		t.Error("Expected auth context to be nil for invalid key")
	}
}

func TestAdminApiKeyAuthMethod_Authenticate_NoHeader(t *testing.T) {
	config := &config.InfrastructureConfig{
		AdminAuthenticationToken: "test-token",
	}
	authMethod := NewAdminApiKeyAuthMethod(config)

	headers := map[string]string{} // No X-Terrareg-ApiKey header

	authContext, err := authMethod.Authenticate(context.Background(), headers, map[string]string{}, map[string]string{})

	if err != nil {
		t.Errorf("Unexpected error (should return nil context): %v", err)
	}

	if authContext != nil {
		t.Error("Expected auth context to be nil when no header provided")
	}
}

func TestAdminApiKeyAuthMethod_Authenticate_Disabled(t *testing.T) {
	config := &config.InfrastructureConfig{
		AdminAuthenticationToken: "", // Empty token = disabled
	}
	authMethod := NewAdminApiKeyAuthMethod(config)

	headers := map[string]string{
		"X-Terrareg-ApiKey": "any-token",
	}

	authContext, err := authMethod.Authenticate(context.Background(), headers, map[string]string{}, map[string]string{})

	if err != nil {
		t.Errorf("Unexpected error (should return nil context): %v", err)
	}

	if authContext != nil {
		t.Error("Expected auth context to be nil when auth method is disabled")
	}
}

func TestAdminApiKeyAuthMethod_AuthContext_Permissions(t *testing.T) {
	config := &config.InfrastructureConfig{
		AdminAuthenticationToken: "test-admin-token",
	}
	authMethod := NewAdminApiKeyAuthMethod(config)

	headers := map[string]string{
		"X-Terrareg-ApiKey": "test-admin-token",
	}

	authContext, err := authMethod.Authenticate(context.Background(), headers, map[string]string{}, map[string]string{})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if authContext == nil {
		t.Fatal("Expected auth context to be returned")
	}

	// Test that admin context has all admin permissions
	if !authContext.IsAdmin() {
		t.Error("Expected admin context to have admin privileges")
	}

	if !authContext.IsBuiltInAdmin() {
		t.Error("Expected admin context to be built-in admin")
	}

	// Test API access permissions
	if !authContext.CanAccessReadAPI() {
		t.Error("Expected admin context to have read API access")
	}

	if !authContext.CanAccessTerraformAPI() {
		t.Error("Expected admin context to have Terraform API access")
	}

	// Test module permissions (admin should have full access)
	if !authContext.CanPublishModuleVersion("any-namespace") {
		t.Error("Expected admin context to be able to publish to any namespace")
	}

	if !authContext.CanUploadModuleVersion("any-namespace") {
		t.Error("Expected admin context to be able to upload to any namespace")
	}

	// Test namespace access
	if !authContext.CheckNamespaceAccess("FULL", "any-namespace") {
		t.Error("Expected admin context to have FULL access to any namespace")
	}

	// Test CSRF requirements
	if authContext.RequiresCSRF() {
		t.Error("Expected admin context to not require CSRF")
	}

	// Test auth state
	if !authContext.CheckAuthState() {
		t.Error("Expected admin context to have valid auth state")
	}
}

func TestAdminApiKeyAuthMethod_Interface_Compliance(t *testing.T) {
	// Test that the auth method implements the HeaderAuthMethod interface
	config := &config.InfrastructureConfig{
		AdminAuthenticationToken: "test-token",
	}
	authMethod := NewAdminApiKeyAuthMethod(config)

	// Verify it implements the base AuthMethod interface
	var _ auth.AuthMethod = authMethod

	// Verify it implements HeaderAuthMethod
	var headerAuthMethod auth.HeaderAuthMethod = authMethod

	// Test the interface methods work
	if headerAuthMethod.GetProviderType() != auth.AuthMethodAdminApiKey {
		t.Errorf("Expected provider type %v", auth.AuthMethodAdminApiKey)
	}

	if !headerAuthMethod.IsEnabled() {
		t.Error("Expected auth method to be enabled")
	}

	// Test Authenticate method returns AuthContext
	authContext, err := headerAuthMethod.Authenticate(
		context.Background(),
		map[string]string{"X-Terrareg-ApiKey": "test-token"},
		map[string]string{},
		map[string]string{},
	)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if authContext == nil {
		t.Error("Expected auth context to be returned")
	}

	if !authContext.IsAuthenticated() {
		t.Error("Expected auth context to be authenticated")
	}
}