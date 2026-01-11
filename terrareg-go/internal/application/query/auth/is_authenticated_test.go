package auth_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/application/query/auth"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/dto"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/middleware"
)

func TestIsAuthenticatedQuery_Execute(t *testing.T) {
	tests := []struct {
		name                string
		setupContext        func() context.Context
		expectAuthenticated bool
		expectReadAccess    bool
		expectSiteAdmin     bool
		expectPermissions   int
	}{
		{
			name: "Unauthenticated - empty context",
			setupContext: func() context.Context {
				return context.Background()
			},
			expectAuthenticated: false,
			expectReadAccess:    false,
			expectSiteAdmin:     false,
			expectPermissions:   0,
		},
		{
			name: "Unauthenticated - context without session data",
			setupContext: func() context.Context {
				return context.WithValue(context.Background(), "other_key", "value")
			},
			expectAuthenticated: false,
			expectReadAccess:    false,
			expectSiteAdmin:     false,
			expectPermissions:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setupContext()
			query := auth.NewIsAuthenticatedQuery()

			response, err := query.Execute(ctx)
			require.NoError(t, err, "Execute should not return an error")

			assert.Equal(t, tt.expectAuthenticated, response.Authenticated, "Authenticated mismatch")
			assert.Equal(t, tt.expectReadAccess, response.ReadAccess, "ReadAccess mismatch")
			assert.Equal(t, tt.expectSiteAdmin, response.SiteAdmin, "SiteAdmin mismatch")
			assert.Len(t, response.NamespacePermissions, tt.expectPermissions, "NamespacePermissions length mismatch")
		})
	}
}

func TestIsAuthenticatedResponse_JSONSerialization(t *testing.T) {
	tests := []struct {
		name     string
		response dto.IsAuthenticatedResponse
	}{
		{
			name: "Fully authenticated response",
			response: dto.IsAuthenticatedResponse{
				Authenticated: true,
				ReadAccess:    true,
				SiteAdmin:     true,
				NamespacePermissions: map[string]string{
					"ns1": "FULL",
					"ns2": "READ",
					"ns3": "MODIFY",
				},
			},
		},
		{
			name: "Partially authenticated response",
			response: dto.IsAuthenticatedResponse{
				Authenticated: true,
				ReadAccess:    true,
				SiteAdmin:     false,
				NamespacePermissions: map[string]string{
					"ns1": "READ",
				},
			},
		},
		{
			name: "Unauthenticated response",
			response: dto.IsAuthenticatedResponse{
				Authenticated: false,
				ReadAccess:    false,
				SiteAdmin:     false,
				NamespacePermissions: map[string]string{},
			},
		},
		{
			name: "Authenticated without permissions",
			response: dto.IsAuthenticatedResponse{
				Authenticated: true,
				ReadAccess:    false,
				SiteAdmin:     false,
				NamespacePermissions: nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test JSON marshaling
			jsonData, err := json.Marshal(tt.response)
			require.NoError(t, err, "Marshal should not return an error")

			// Test JSON unmarshaling
			var unmarshaled dto.IsAuthenticatedResponse
			err = json.Unmarshal(jsonData, &unmarshaled)
			require.NoError(t, err, "Unmarshal should not return an error")

			// Verify data preservation
			assert.Equal(t, tt.response.Authenticated, unmarshaled.Authenticated, "Authenticated not preserved")
			assert.Equal(t, tt.response.ReadAccess, unmarshaled.ReadAccess, "ReadAccess not preserved")
			assert.Equal(t, tt.response.SiteAdmin, unmarshaled.SiteAdmin, "SiteAdmin not preserved")
			assert.Len(t, unmarshaled.NamespacePermissions, len(tt.response.NamespacePermissions), "NamespacePermissions length mismatch")

			// Verify each permission is preserved
			for key, value := range tt.response.NamespacePermissions {
				assert.Equal(t, value, unmarshaled.NamespacePermissions[key], "Permission %s not preserved", key)
			}
		})
	}
}

func TestGetSessionData(t *testing.T) {
	tests := []struct {
		name             string
		setupContext     func() context.Context
		expectNilSession bool
	}{
		{
			name: "Empty context",
			setupContext: func() context.Context {
				return context.Background()
			},
			expectNilSession: true,
		},
		{
			name: "Context with unrelated value",
			setupContext: func() context.Context {
				return context.WithValue(context.Background(), "unrelated_key", "unrelated_value")
			},
			expectNilSession: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setupContext()
			sessionData := middleware.GetSessionData(ctx)

			if tt.expectNilSession {
				assert.Nil(t, sessionData, "Expected nil session data")
			} else {
				assert.NotNil(t, sessionData, "Expected non-nil session data")
			}
		})
	}
}
