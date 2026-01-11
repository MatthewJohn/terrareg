package main

import (
	"context"
	"flag"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	domainConfigService "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/config/service"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/container"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/version"
)

func main() {
	// Parse command-line flags
	sslCertPrivateKey := flag.String("ssl-cert-private-key", "", "Path to SSL private key certificate file")
	sslCertPublicKey := flag.String("ssl-cert-public-key", "", "Path to SSL public key certificate file")
	flag.Parse()

	// Setup logger
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	logger := log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	logger.Info().Msg("Starting Terrareg Go Server")

	// Use new configuration service architecture
	c, err := bootstrap(logger, *sslCertPrivateKey, *sslCertPublicKey)
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
func bootstrap(logger zerolog.Logger, sslCertPrivateKey, sslCertPublicKey string) (*container.Container, error) {
	// Load configuration using the new configuration service
	versionReader := version.NewVersionReader()
	configService := domainConfigService.NewConfigurationService(
		domainConfigService.ConfigurationServiceOptions{
			SSLCertPrivateKey: sslCertPrivateKey,
			SSLCertPublicKey:  sslCertPublicKey,
		},
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

	// Initialize dependency injection container with new configuration architecture
	logger.Info().Msg("Initializing application container")
	c, err := container.NewContainer(
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
		&sqldb.AuthenticationTokenDB{},
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
