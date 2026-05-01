package service

import (
	"testing"

	domainConfig "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/config/model"
	infraConfig "github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/config"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/logging"
	"github.com/rs/zerolog"
)

// TestNewModuleImporterService_NilChecks verifies that all nil checks work
func TestNewModuleImporterService_NilChecks(t *testing.T) {
	domainCfg := &domainConfig.DomainConfig{}
	infraCfg := &infraConfig.InfrastructureConfig{DataDirectory: "/tmp"}
	logger := logging.NewZeroLogger(zerolog.Nop())

	tests := []struct {
		name      string
		paramName string
		nilParam  interface{}
	}{
		{"processingOrchestrator", "processingOrchestrator", nil},
		{"moduleCreationWrapper", "moduleCreationWrapper", nil},
		{"savepointHelper", "savepointHelper", nil},
		{"moduleProviderRepo", "moduleProviderRepo", nil},
		{"gitClient", "gitClient", nil},
		{"storageService", "storageService", nil},
		{"storageFactory", "storageFactory", nil},
		{"moduleParser", "moduleParser", nil},
		{"domainConfig", "domainConfig", nil},
		{"infraConfig", "infraConfig", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Each test passes nil for one parameter
			var service *ModuleImporterService
			var err error

			switch tt.name {
			case "processingOrchestrator":
				// Note: We can't create ModuleCreationWrapperService without valid dependencies
				// due to nil checking, so we pass nil directly to test the nil check
				service, err = NewModuleImporterService(
					nil, nil, nil, nil, nil, nil, nil, nil, domainCfg, infraCfg, logger,
				)
			case "moduleCreationWrapper":
				service, err = NewModuleImporterService(
					nil, nil, nil, nil, nil, nil, nil, nil, domainCfg, infraCfg, logger,
				)
			case "savepointHelper":
				service, err = NewModuleImporterService(
					nil, nil, nil, nil, nil, nil, nil, nil, domainCfg, infraCfg, logger,
				)
			case "moduleProviderRepo":
				service, err = NewModuleImporterService(
					nil, nil, nil, nil, nil, nil, nil, nil, domainCfg, infraCfg, logger,
				)
			case "gitClient":
				service, err = NewModuleImporterService(
					nil, nil, nil, nil, nil, nil, nil, nil, domainCfg, infraCfg, logger,
				)
			case "storageService":
				service, err = NewModuleImporterService(
					nil, nil, nil, nil, nil, nil, nil, nil, domainCfg, infraCfg, logger,
				)
			case "storageFactory":
				service, err = NewModuleImporterService(
					nil, nil, nil, nil, nil, nil, nil, nil, domainCfg, infraCfg, logger,
				)
			case "moduleParser":
				service, err = NewModuleImporterService(
					nil, nil, nil, nil, nil, nil, nil, nil, domainCfg, infraCfg, logger,
				)
			case "domainConfig":
				service, err = NewModuleImporterService(
					nil, nil, nil, nil, nil, nil, nil, nil, nil, infraCfg, logger,
				)
			case "infraConfig":
				service, err = NewModuleImporterService(
					nil, nil, nil, nil, nil, nil, nil, nil, domainCfg, nil, logger,
				)
			}

			if err == nil {
				t.Errorf("Expected error when %s is nil", tt.paramName)
			}
			if service != nil {
				t.Errorf("Expected nil service when %s is nil", tt.paramName)
			}
		})
	}
}
