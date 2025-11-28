package main

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/terrareg/terrareg/internal/config"
	"github.com/terrareg/terrareg/internal/infrastructure/persistence/sqldb"
	"github.com/terrareg/terrareg/internal/interfaces/http"
)

func main() {
	// Setup logger
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	logger := log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	logger.Info().Msg("Starting Terrareg Go Server")

	// Load configuration
	cfg, err := config.New()
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		logger.Fatal().Err(err).Msg("Invalid configuration")
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
		logger.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer db.Close()

	logger.Info().Msg("Database connected successfully")

	// Auto-migrate database schema (for development)
	if cfg.Debug {
		logger.Info().Msg("Running database auto-migration")
		if err := autoMigrate(db); err != nil {
			logger.Fatal().Err(err).Msg("Failed to auto-migrate database")
		}
	}

	// TODO: Initialize repositories, services, and application layer
	// This will be implemented in Phase 2+

	// Create HTTP server
	server := http.NewServer(cfg, logger)

	// Start server
	logger.Info().Int("port", cfg.ListenPort).Msg("Starting HTTP server")
	if err := server.Start(); err != nil {
		logger.Fatal().Err(err).Msg("HTTP server failed")
	}
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
