package main

import (
	"context"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/config"
	appConfig "github.com/matthewjohn/terrareg/terrareg-go/internal/config"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/container"
	domainConfigService "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/config/service"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/config"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/version"
)

func main() {
	// Setup logger
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	logger := log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	logger.Info().Msg("Starting Terrareg Go Server")

	// Use new configuration service architecture
	c, err := bootstrapWithConfigService(logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to bootstrap application")
	}

	// Start session cleanup service
	go c.SessionCleanupService.Start(context.Background())

	// Start server
	logger.Info().Int("port", c.InfraConfig.ListenPort).Msg("Starting HTTP server")
	if err := c.Server.Start(); err != nil {
		logger.Fatal().Err(err).Msg("HTTP server failed")
	}
}

// bootstrapWithConfigService bootstraps the application using the new configuration service
func bootstrapWithConfigService(logger zerolog.Logger) (*container.Container, error) {
	// Load configuration using the new configuration service
	versionReader := version.NewVersionReader()
	configService := domainConfigService.NewConfigurationService(
		domainConfigService.ConfigurationServiceOptions{},
		versionReader,
	)

	domainConfig, infraConfig, err := configService.LoadConfiguration()
	if err != nil {
		return nil, err
	}

	logger.Info().
		Int("port", infraConfig.ListenPort).
		Str("public_url", infraConfig.PublicURL).
		Str("database_url", maskDatabaseURL(infraConfig.DatabaseURL)).
		Msg("Configuration loaded")

	// Initialize database
	logger.Info().Msg("Connecting to database")
	db, err := sqldb.NewDatabase(infraConfig.DatabaseURL, infraConfig.Debug)
	if err != nil {
		return nil, err
	}

	logger.Info().Msg("Database connected successfully")

	// Auto-migrate database schema (for development)
	if infraConfig.Debug {
		logger.Info().Msg("Running database auto-migration")
		// if err := autoMigrate(db); err != nil {
		// 	logger.Fatal().Err(err).Msg("Failed to auto-migrate database")
		// }
	}

	// Load legacy config for backward compatibility during migration
	legacyConfig, err := appConfig.New()
	if err != nil {
		logger.Warn().Err(err).Msg("Failed to load legacy config, some features may not work")
		legacyConfig = nil // Continue without legacy config
	}

	// Initialize dependency injection container with new configuration architecture
	logger.Info().Msg("Initializing application container")
	c, err := container.NewContainerWithConfigService(
		legacyConfig,
		domainConfig,
		infraConfig,
		configService,
		logger,
		db,
	)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// bootstrapLegacy bootstraps the application using the legacy configuration (for fallback)
func bootstrapLegacy(logger zerolog.Logger) (*container.Container, error) {
	// Load legacy configuration
	cfg, err := config.New()
	if err != nil {
		return nil, err
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	logger.Info().
		Int("port", cfg.ListenPort).
		Str("public_url", cfg.PublicURL).
		Str("database_url", maskDatabaseURL(cfg.DatabaseURL)).
		Msg("Configuration loaded")

	// Initialize database
	logger.Info().Msg("Connecting to database")
	db, err := sqldb.NewDatabase(cfg.DatabaseURL, cfg.Debug)
	if err != nil {
		return nil, err
	}

	logger.Info().Msg("Database connected successfully")

	// Auto-migrate database schema (for development)
	if cfg.Debug {
		logger.Info().Msg("Running database auto-migration")
		// if err := autoMigrate(db); err != nil {
		// 	logger.Fatal().Err(err).Msg("Failed to auto-migrate database")
		// }
	}

	// Initialize dependency injection container with legacy configuration
	logger.Info().Msg("Initializing application container")
	c, err := container.NewContainer(cfg, logger, db)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// autoMigrate runs GORM auto-migration for all models
func autoMigrate(db *sqldb.Database) error {
	// Import all models and run AutoMigrate
	return db.DB.AutoMigrate(
		&sqldb.SessionDB{},
		&sqldb.TerraformIDPAuthorizationCodeDB{},
		&sqldb.TerraformIDPAccessTokenDB{},
		&sqldb.TerraformIDPSubjectIdentifierDB{},
		&sqldb.UserGroupDB{},
		&sqldb.NamespaceDB{},
		&sqldb.UserGroupNamespacePermissionDB{},
		&sqldb.GitProviderDB{},
		&sqldb.NamespaceRedirectDB{},
		&sqldb.ModuleDetailsDB{},
		&sqldb.ModuleProviderDB{},
		&sqldb.ModuleVersionDB{},
		&sqldb.ModuleProviderRedirectDB{},
		&sqldb.SubmoduleDB{},
		&sqldb.AnalyticsDB{},
		&sqldb.ProviderAnalyticsDB{},
		&sqldb.ExampleFileDB{},
		&sqldb.ModuleVersionFileDB{},
		&sqldb.GPGKeyDB{},
		&sqldb.ProviderSourceDB{},
		&sqldb.ProviderCategoryDB{},
		&sqldb.RepositoryDB{},
		&sqldb.ProviderDB{},
		&sqldb.ProviderVersionDB{},
		&sqldb.ProviderVersionDocumentationDB{},
		&sqldb.ProviderVersionBinaryDB{},
		&sqldb.AuditHistoryDB{},
	)
}

// maskDatabaseURL masks sensitive information in database URL
func maskDatabaseURL(url string) string {
	// Simple masking - just show the database type
	if len(url) > 20 {
		return url[:10] + "****" + url[len(url)-6:]
	}
	return "****"
}
