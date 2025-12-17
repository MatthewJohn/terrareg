package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gorm.io/gorm"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb/transaction"
)

// TerraregMetadata represents terrareg.json metadata
type TerraregMetadata struct {
	Owner            *string                `json:"owner"`
	Description      *string                `json:"description"`
	RepoCloneURL     *string                `json:"repo_clone_url"`
	RepoBrowseURL    *string                `json:"repo_browse_url"`
	IssuesURL        *string                `json:"issues_url"`
	License          *string                `json:"license"`
	Provider         map[string]*string     `json:"provider"`
	VariableTemplate map[string]interface{} `json:"variable_template"`
}

// MetadataProcessingRequest represents a request to process metadata
type MetadataProcessingRequest struct {
	ModuleVersionID    int
	MetadataPath       string
	ModulePath         string
	TransactionCtx     context.Context
	RequiredAttributes []string
}

// MetadataProcessingResult represents the result of metadata processing
type MetadataProcessingResult struct {
	Success           bool              `json:"success"`
	Metadata          *TerraregMetadata `json:"metadata,omitempty"`
	MetadataFound     bool              `json:"metadata_found"`
	Validated         bool              `json:"validated"`
	Error             *string           `json:"error,omitempty"`
	MissingAttributes []string          `json:"missing_attributes,omitempty"`
	Duration          time.Duration     `json:"duration"`
}

// PathspecFilter represents pathspec filtering rules
type PathspecFilter struct {
	Rules []string `json:"rules"`
}

// MetadataProcessingService handles metadata processing with validation rollback
type MetadataProcessingService struct {
	savepointHelper *transaction.SavepointHelper
}

// NewMetadataProcessingService creates a new metadata processing service
func NewMetadataProcessingService(savepointHelper *transaction.SavepointHelper) *MetadataProcessingService {
	return &MetadataProcessingService{
		savepointHelper: savepointHelper,
	}
}

// ProcessMetadataWithTransaction processes metadata with transaction safety
func (s *MetadataProcessingService) ProcessMetadataWithTransaction(
	ctx context.Context,
	req MetadataProcessingRequest,
) (*MetadataProcessingResult, error) {
	startTime := time.Now()
	result := &MetadataProcessingResult{
		Success:           false,
		MetadataFound:     false,
		Validated:         false,
		MissingAttributes: req.RequiredAttributes,
		Duration:          0,
	}

	savepointName := fmt.Sprintf("metadata_processing_%d", startTime.UnixNano())

	err := s.savepointHelper.WithSmartSavepointOrTransaction(ctx, savepointName, func(tx *gorm.DB) error {
		// Check if metadata file exists
		metadataPath := s.findMetadataFile(req.MetadataPath)
		if metadataPath == "" {
			// No metadata file found - this is not an error
			result.Success = true
			result.MetadataFound = false
			return nil
		}

		result.MetadataFound = true

		// Read and parse metadata file
		metadata, err := s.readMetadataFile(metadataPath)
		if err != nil {
			return fmt.Errorf("failed to read metadata file: %w", err)
		}

		// Validate required attributes
		missingAttrs := s.validateRequiredAttributes(metadata, req.RequiredAttributes)
		if len(missingAttrs) > 0 {
			result.MissingAttributes = missingAttrs
			return fmt.Errorf("missing required attributes: %v", missingAttrs)
		}

		// Store parsed metadata
		result.Metadata = metadata
		result.Validated = true
		result.Success = true

		return nil
	})

	result.Duration = time.Since(startTime)

	if err != nil {
		errorMsg := err.Error()
		result.Error = &errorMsg
		return result, nil
	}

	return result, nil
}

// ProcessTerraregMetadata processes terrareg.json metadata with rollback on validation failure
func (s *MetadataProcessingService) ProcessTerraregMetadata(
	ctx context.Context,
	metadataPath string,
	moduleVersionID int,
) (*MetadataProcessingResult, error) {
	// This is a convenience method for processing terrareg.json specifically
	req := MetadataProcessingRequest{
		ModuleVersionID:    moduleVersionID,
		MetadataPath:       metadataPath,
		RequiredAttributes: []string{}, // Would pass configured required attributes
		TransactionCtx:     ctx,
	}

	return s.ProcessMetadataWithTransaction(ctx, req)
}

// findMetadataFile searches for terrareg.json or .terrareg.json in the given directory
func (s *MetadataProcessingService) findMetadataFile(basePath string) string {
	// Priority order: terrareg.json, then .terrareg.json
	metadataFiles := []string{"terrareg.json", ".terrareg.json"}

	for _, filename := range metadataFiles {
		path := filepath.Join(basePath, filename)
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return ""
}

// readMetadataFile reads and parses a metadata file
func (s *MetadataProcessingService) readMetadataFile(metadataPath string) (*TerraregMetadata, error) {
	data, err := os.ReadFile(metadataPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata file: %w", err)
	}

	var metadata TerraregMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("failed to parse metadata JSON: %w", err)
	}

	return &metadata, nil
}

// validateRequiredAttributes checks if all required attributes are present
func (s *MetadataProcessingService) validateRequiredAttributes(
	metadata *TerraregMetadata,
	requiredAttributes []string,
) []string {
	var missing []string

	for _, attr := range requiredAttributes {
		switch attr {
		case "owner":
			if metadata.Owner == nil || *metadata.Owner == "" {
				missing = append(missing, "owner")
			}
		case "description":
			if metadata.Description == nil || *metadata.Description == "" {
				missing = append(missing, "description")
			}
		case "repo_clone_url":
			if metadata.RepoCloneURL == nil || *metadata.RepoCloneURL == "" {
				missing = append(missing, "repo_clone_url")
			}
		case "repo_browse_url":
			if metadata.RepoBrowseURL == nil || *metadata.RepoBrowseURL == "" {
				missing = append(missing, "repo_browse_url")
			}
		case "license":
			if metadata.License == nil || *metadata.License == "" {
				missing = append(missing, "license")
			}
		default:
			// For custom attributes, check in provider map
			if metadata.Provider != nil {
				if _, exists := metadata.Provider[attr]; !exists {
					missing = append(missing, attr)
				}
			}
		}
	}

	return missing
}

// GetPathspecFilter reads and parses .terraformignore file if it exists
func (s *MetadataProcessingService) GetPathspecFilter(ctx context.Context, modulePath string) (*PathspecFilter, error) {
	ignorePath := filepath.Join(modulePath, ".terraformignore")
	data, err := os.ReadFile(ignorePath)
	if err != nil {
		// .terraformignore not found is not an error
		if os.IsNotExist(err) {
			return &PathspecFilter{Rules: []string{}}, nil
		}
		return nil, fmt.Errorf("failed to read .terraformignore file: %w", err)
	}

	// Parse ignore patterns (simplified - would use pathspec library in full implementation)
	lines := filepath.SplitList(string(data))
	var rules []string
	for _, line := range lines {
		line = filepath.Clean(line)
		if line != "" && !startsWith(line, "#") {
			rules = append(rules, line)
		}
	}

	return &PathspecFilter{Rules: rules}, nil
}

// startsWith checks if a string starts with a prefix (case-sensitive)
func startsWith(s, prefix string) bool {
	return len(s) >= len(prefix) && s[0:len(prefix)] == prefix
}
