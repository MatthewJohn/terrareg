package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared/service"
	URLService "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/url/service"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/logging"
)

// InfracostServiceImpl handles cost analysis of module examples using Infracost CLI
type InfracostServiceImpl struct {
	config         *InfracostConfig
	logger         logging.Logger
	commandService service.SystemCommandService
	urlService     URLService.URLService
}

// InfracostConfig holds configuration for infracost operations
type InfracostConfig struct {
	// InfracostAPIKey is the API key for infracost (required for infracost to run)
	InfracostAPIKey string
	// InternalExtractionAnalyticsToken is used for Terraform Cloud authentication
	// when the example module references Terraform Cloud modules
	InternalExtractionAnalyticsToken string
	// PublicURL is used to derive the Terraform Cloud host
	PublicURL string
}

// NewInfracostService creates a new InfracostService implementation
func NewInfracostService(
	config *InfracostConfig,
	logger logging.Logger,
	commandService service.SystemCommandService,
	urlService *URLService.URLService,
) (*InfracostServiceImpl, error) {
	if urlService == nil {
		return nil, fmt.Errorf("urlService must not be nil in NewInfracostService")
	}
	if config == nil {
		return nil, fmt.Errorf("config must not be nil in NewInfracostService")
	}
	return &InfracostServiceImpl{
		config:         config,
		logger:         logger,
		commandService: commandService,
		urlService:     *urlService,
	}, nil
}

// AnalyzeExample runs infracost on a module example and returns the JSON results
// Returns (nil, nil) if infracost is not configured (not an error)
// Returns (results, nil) on success
// Returns (nil, error) on execution failure
func (s *InfracostServiceImpl) AnalyzeExample(ctx context.Context, examplePath string) ([]byte, error) {
	// 1. Check if API key is configured
	if !s.IsAvailable() {
		s.logger.Debug().Msg("infracost API key not configured, skipping cost analysis")
		return nil, nil // Not an error - just skip
	}

	// 2. Validate example path
	if examplePath == "" {
		return nil, fmt.Errorf("example path cannot be empty")
	}

	// Check if path exists
	if _, err := os.Stat(examplePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("example path does not exist: %s", examplePath)
	}

	// 3. Create temporary file for output
	tmpFile, err := os.CreateTemp("", "infracost-*.json")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file for infracost: %w", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	// 4. Build command arguments
	args := []string{
		"breakdown",
		"--path", examplePath,
		"--format", "json",
		"--out-file", tmpPath,
	}

	// 5. Build environment variables
	env := s.buildEnvironment()

	// 6. Execute command
	cmd := &service.Command{
		Name: "infracost",
		Args: args,
		Env:  env,
	}

	s.logger.Debug().Str("path", examplePath).Msg("running infracost")
	cmdResult, err := s.commandService.Execute(ctx, cmd)
	if err != nil {
		// Check if it's a "not found" error
		if strings.Contains(err.Error(), "executable file not found") || strings.Contains(err.Error(), "file not found") {
			s.logger.Warn().Msg("infracost executable not found, skipping cost analysis")
			return nil, nil // Not an error - skip gracefully
		}
		return nil, fmt.Errorf("infracost failed: %w: %s", err, cmdResult.Stderr)
	}

	// 7. Read results from temp file
	result, err := os.ReadFile(tmpPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read infracost output from temp file: %w", err)
	}

	// 8. Validate JSON output
	var rawJSON map[string]interface{}
	if err := json.Unmarshal(result, &rawJSON); err != nil {
		s.logger.Warn().Err(err).Str("output", string(result)).Msg("infracost output is not valid JSON")
		return nil, fmt.Errorf("infracost output is not valid JSON: %w", err)
	}

	s.logger.Debug().
		Int("bytes", len(result)).
		Str("example", filepath.Base(examplePath)).
		Msg("infracost analysis completed successfully")

	return result, nil
}

// IsAvailable returns true if infracost API key is configured
func (s *InfracostServiceImpl) IsAvailable() bool {
	return s.config != nil && s.config.InfracostAPIKey != ""
}

// buildEnvironment builds the environment variables for infracost
func (s *InfracostServiceImpl) buildEnvironment() []string {
	env := os.Environ()

	// Parse PUBLIC_URL to get domain name
	urlDetails := s.urlService.GetPublicURLDetails(nil)

	// Add Terraform Cloud environment variables if configured
	// This matches Python's behavior: domain from PUBLIC_URL, token from INTERNAL_EXTRACTION_ANALYTICS_TOKEN
	// This is slightly modified, as we use the parsed PULBIC_URL and check scheme
	if s.config.InternalExtractionAnalyticsToken != "" && urlDetails.Domain != "" && urlDetails.Protocol == URLService.ProtocolHttps {
		// Set INFRACOST_TERRAFORM_CLOUD_TOKEN from INTERNAL_EXTRACTION_ANALYTICS_TOKEN
		env = append(env, fmt.Sprintf("INFRACOST_TERRAFORM_CLOUD_TOKEN=%s", s.config.InternalExtractionAnalyticsToken))
		// Set INFRACOST_TERRAFORM_CLOUD_HOST from PublicURL domain
		env = append(env, fmt.Sprintf("INFRACOST_TERRAFORM_CLOUD_HOST=%s", urlDetails.Domain))
		s.logger.Debug().
			Str("host", urlDetails.Domain).
			Msg("configured Terraform Cloud for infracost")
	} else {
		s.logger.Debug().
			Str("public_url protocol", string(urlDetails.Protocol)).
			Str("public_url domain", urlDetails.Domain).
			Str("internal_extraction_analytics_token", s.config.InternalExtractionAnalyticsToken).
			Msg("INTERNAL_EXTRACTION_ANALYTICS_TOKEN or PUBLIC_URL is not set or is http, so not used for Terraform Cloud host")
	}

	return env
}
