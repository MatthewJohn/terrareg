package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog"

	configmodel "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/config/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/model"
)

// SubmoduleScanningService handles scanning for submodules and examples
type SubmoduleScanningService struct {
	domainConfig *configmodel.DomainConfig
	logger       zerolog.Logger
}

// NewSubmoduleScanningService creates a new submodule scanning service
func NewSubmoduleScanningService(
	domainConfig *configmodel.DomainConfig,
	logger zerolog.Logger,
) *SubmoduleScanningService {
	return &SubmoduleScanningService{
		domainConfig: domainConfig,
		logger:       logger,
	}
}

// ScanSubmodules scans for submodules in the configured modules directory
func (s *SubmoduleScanningService) ScanSubmodules(
	ctx context.Context,
	modulePath string,
	moduleVersion *model.ModuleVersion,
) error {
	submodulesDir := s.domainConfig.ModulesDirectory
	if submodulesDir == "" {
		s.logger.Debug().Msg("MODULES_DIRECTORY not configured, skipping submodule scan")
		return nil
	}

	return s.scanSubdirectory(ctx, modulePath, submodulesDir, moduleVersion, true)
}

// ScanExamples scans for examples in the configured examples directory
func (s *SubmoduleScanningService) ScanExamples(
	ctx context.Context,
	modulePath string,
	moduleVersion *model.ModuleVersion,
) error {
	examplesDir := s.domainConfig.ExamplesDirectory
	if examplesDir == "" {
		s.logger.Debug().Msg("EXAMPLES_DIRECTORY not configured, skipping example scan")
		return nil
	}

	return s.scanSubdirectory(ctx, modulePath, examplesDir, moduleVersion, false)
}

// scanSubdirectory scans a subdirectory for .tf files and creates submodules or examples
func (s *SubmoduleScanningService) scanSubdirectory(
	ctx context.Context,
	modulePath string,
	subdirectory string,
	moduleVersion *model.ModuleVersion,
	isSubmodule bool,
) error {
	// Build the full path to the subdirectory
	scanPath := filepath.Join(modulePath, subdirectory)

	// Check if the directory exists
	if _, err := os.Stat(scanPath); os.IsNotExist(err) {
		s.logger.Debug().Str("path", scanPath).Msg("Subdirectory does not exist, skipping scan")
		return nil
	}

	// Collect unique parent directories containing .tf files
	parentDirs := make(map[string]bool)
	err := filepath.Walk(scanPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-.tf files
		if info.IsDir() || !strings.HasSuffix(path, ".tf") {
			return nil
		}

		// Get the parent directory
		parentDir := filepath.Dir(path)

		// Make the path relative to the module base path
		relPath, err := filepath.Rel(modulePath, parentDir)
		if err != nil {
			s.logger.Debug().Err(err).Str("path", parentDir).Msg("Failed to get relative path")
			return nil
		}

		// Skip if the parent is the root of the subdirectory
		if relPath == subdirectory {
			s.logger.Warn().
				Str("path", path).
				Str("subdirectory", subdirectory).
				Msg("Submodule/example is in root of subdirectory, skipping")
			return nil
		}

		// Normalize the path to use forward slashes (like Python)
		relPath = filepath.ToSlash(relPath)
		parentDirs[relPath] = true

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to scan directory %s: %w", scanPath, err)
	}

	// Create submodules/examples for each unique parent directory
	for path := range parentDirs {
		if isSubmodule {
			submodule := model.NewSubmodule(path, nil, nil, nil)
			moduleVersion.AddSubmodule(submodule)
			s.logger.Debug().Str("path", path).Msg("Added submodule")
		} else {
			example := model.NewExample(path, nil, nil)
			moduleVersion.AddExample(example)
			s.logger.Debug().Str("path", path).Msg("Added example")
		}
	}

	s.logger.Info().
		Str("subdirectory", subdirectory).
		Int("count", len(parentDirs)).
		Msg("Scanned subdirectory for submodules/examples")

	return nil
}

// ScanAll scans both submodules and examples
func (s *SubmoduleScanningService) ScanAll(
	ctx context.Context,
	modulePath string,
	moduleVersion *model.ModuleVersion,
) error {
	if err := s.ScanSubmodules(ctx, modulePath, moduleVersion); err != nil {
		return fmt.Errorf("failed to scan submodules: %w", err)
	}

	if err := s.ScanExamples(ctx, modulePath, moduleVersion); err != nil {
		return fmt.Errorf("failed to scan examples: %w", err)
	}

	return nil
}
