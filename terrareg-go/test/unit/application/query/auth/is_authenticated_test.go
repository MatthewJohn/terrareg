package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/application/query/auth"
	authService "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth/service"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/dto"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/middleware"
)

// Use a mock session key for testing
type testSessionKey string

const testSessionDataKey testSessionKey = "sessionData"

// Helper function to add session data to context for testing
func withTestSessionData(ctx context.Context, sessionData *authService.SessionData) context.Context {
	return context.WithValue(ctx, testSessionDataKey, sessionData)
}

// Helper function to get session data from context for testing
func getTestSessionData(ctx context.Context) *authService.SessionData {
	if sessionData, ok := ctx.Value(testSessionDataKey).(*authService.SessionData); ok {
		return sessionData
	}
	return nil
}

func TestIsAuthenticatedQuery_Execute(t *testing.T) {
	// Test 1: No session data (unauthenticated)
	t.Run("Unauthenticated", func(t *testing.T) {
		ctx := context.Background()
		query := auth.NewIsAuthenticatedQuery()

		response, err := query.Execute(ctx)
		if err != nil {
			t.Fatalf("Failed to execute query: %v", err)
		}

		if response.Authenticated != false {
			t.Errorf("Expected Authenticated=false, got %v", response.Authenticated)
		}
		if response.ReadAccess != false {
			t.Errorf("Expected ReadAccess=false, got %v", response.ReadAccess)
		}
		if response.SiteAdmin != false {
			t.Errorf("Expected SiteAdmin=false, got %v", response.SiteAdmin)
		}
		if len(response.NamespacePermissions) != 0 {
			t.Errorf("Expected empty permissions, got %v", response.NamespacePermissions)
		}

		fmt.Printf("✓ Unauthenticated test passed: %+v\n", response)
	})

	// Test 2: Admin session (simulated)
	t.Run("Query Structure", func(t *testing.T) {
		// Since we can't use the real sessionDataKey, we'll test that the query structure is correct
		ctx := context.Background()
		query := auth.NewIsAuthenticatedQuery()

		// We can't directly test with mocked context since sessionDataKey is not exported
		// But we can verify the query doesn't panic and handles nil context properly
		response, err := query.Execute(ctx)
		if err != nil {
			t.Fatalf("Failed to execute query: %v", err)
		}

		// Should return unauthenticated response when no session data
		if response.Authenticated != false {
			t.Errorf("Expected Authenticated=false for empty context, got %v", response.Authenticated)
		}

		fmt.Printf("✓ Query handles empty context correctly: %+v\n", response)
	})

	// Test 3: Test JSON serialization matches expected format
	t.Run("Response Format", func(t *testing.T) {
		response := dto.IsAuthenticatedResponse{
			Authenticated: true,
			ReadAccess: true,
			SiteAdmin: false,
			NamespacePermissions: map[string]string{
				"ns1": "FULL",
				"ns2": "READ",
			},
		}

		jsonData, err := json.Marshal(response)
		if err != nil {
			t.Fatalf("Failed to marshal response: %v", err)
		}

		var unmarshaled dto.IsAuthenticatedResponse
		err = json.Unmarshal(jsonData, &unmarshaled)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if unmarshaled.Authenticated != response.Authenticated ||
		   unmarshaled.ReadAccess != response.ReadAccess ||
		   unmarshaled.SiteAdmin != response.SiteAdmin ||
		   len(unmarshaled.NamespacePermissions) != len(response.NamespacePermissions) {
			t.Error("JSON serialization/deserialization did not preserve response")
		}

		fmt.Printf("✓ Response format test passed: %s\n", string(jsonData))
	})
}

func TestGetSessionData(t *testing.T) {
	// Test the GetSessionData function from middleware
	t.Run("GetSessionData from empty context", func(t *testing.T) {
		ctx := context.Background()
		sessionData := middleware.GetSessionData(ctx)

		if sessionData != nil {
			t.Errorf("Expected nil session data from empty context, got %v", sessionData)
		}

		fmt.Printf("✓ GetSessionData correctly returns nil for empty context\n")
	})
}