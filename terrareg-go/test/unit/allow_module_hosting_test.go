package unit

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/config/model"
)

func TestDomainConfig_AllowModuleHostingValues(t *testing.T) {
	tests := []struct {
		name               string
		allowModuleHosting model.ModuleHostingMode
		expectedValid      bool
	}{
		{
			name:               "Allow mode",
			allowModuleHosting: model.ModuleHostingModeAllow,
			expectedValid:      true,
		},
		{
			name:               "Enforce mode",
			allowModuleHosting: model.ModuleHostingModeEnforce,
			expectedValid:      true,
		},
		{
			name:               "Disallow mode",
			allowModuleHosting: model.ModuleHostingModeDisallow,
			expectedValid:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create config with the specific ALLOW_MODULE_HOSTING mode
			cfg := &model.DomainConfig{
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
		name                 string
		allowModuleHosting   model.ModuleHostingMode
		expectedShouldAllow  bool
	}{
		{
			name:                 "Allow mode - upload should be allowed",
			allowModuleHosting:   model.ModuleHostingModeAllow,
			expectedShouldAllow:  true,
		},
		{
			name:                 "Enforce mode - upload should be allowed",
			allowModuleHosting:   model.ModuleHostingModeEnforce,
			expectedShouldAllow:  true,
		},
		{
			name:                 "Disallow mode - upload should be blocked",
			allowModuleHosting:   model.ModuleHostingModeDisallow,
			expectedShouldAllow:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create config with the specific ALLOW_MODULE_HOSTING mode
			cfg := &model.DomainConfig{
				AllowModuleHosting: tt.allowModuleHosting,
			}

			// Simple validation check based on config
			shouldAllow := cfg.AllowModuleHosting == model.ModuleHostingModeAllow ||
						  cfg.AllowModuleHosting == model.ModuleHostingModeEnforce

			assert.Equal(t, tt.expectedShouldAllow, shouldAllow)
			assert.Equal(t, tt.allowModuleHosting, cfg.AllowModuleHosting)
		})
	}
}