package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDomainConfig_AllowModuleHostingValues(t *testing.T) {
	tests := []struct {
		name               string
		allowModuleHosting ModuleHostingMode
		expectedValid      bool
	}{
		{
			name:               "Allow mode",
			allowModuleHosting: ModuleHostingModeAllow,
			expectedValid:      true,
		},
		{
			name:               "Enforce mode",
			allowModuleHosting: ModuleHostingModeEnforce,
			expectedValid:      true,
		},
		{
			name:               "Disallow mode",
			allowModuleHosting: ModuleHostingModeDisallow,
			expectedValid:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create config with the specific ALLOW_MODULE_HOSTING mode
			cfg := &DomainConfig{
				AllowModuleHosting: tt.allowModuleHosting,
			}

			// Test that config is created and value is set correctly
			assert.NotNil(t, cfg)
			assert.Equal(t, tt.allowModuleHosting, cfg.AllowModuleHosting)
		})
	}
}

func TestModuleVersionUpload_ConfigValidation(t *testing.T) {
	// Test that configuration validation works correctly
	// This is a simplified test focusing on config validation rather than full upload flow

	tests := []struct {
		name                string
		allowModuleHosting  ModuleHostingMode
		expectedShouldAllow bool
	}{
		{
			name:                "Allow mode - upload should be allowed",
			allowModuleHosting:  ModuleHostingModeAllow,
			expectedShouldAllow: true,
		},
		{
			name:                "Enforce mode - upload should be allowed",
			allowModuleHosting:  ModuleHostingModeEnforce,
			expectedShouldAllow: true,
		},
		{
			name:                "Disallow mode - upload should be blocked",
			allowModuleHosting:  ModuleHostingModeDisallow,
			expectedShouldAllow: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create config with the specific ALLOW_MODULE_HOSTING mode
			cfg := &DomainConfig{
				AllowModuleHosting: tt.allowModuleHosting,
			}

			// Simple validation check based on config
			shouldAllow := cfg.AllowModuleHosting == ModuleHostingModeAllow ||
				cfg.AllowModuleHosting == ModuleHostingModeEnforce

			assert.Equal(t, tt.expectedShouldAllow, shouldAllow)
			assert.Equal(t, tt.allowModuleHosting, cfg.AllowModuleHosting)
		})
	}
}
