package sqldb

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Database wraps the GORM database connection
type Database struct {
	DB *gorm.DB
}

// NewDatabase creates a new database connection
func NewDatabase(databaseURL string, debug bool) (*Database, error) {
	var dialector gorm.Dialector

	// Parse database URL and determine driver
	if strings.HasPrefix(databaseURL, "sqlite://") {
		// SQLite
		dbPath := strings.TrimPrefix(databaseURL, "sqlite://")
		dialector = sqlite.Open(dbPath)
	} else if strings.HasPrefix(databaseURL, "mysql://") {
		// MySQL
		dsn := convertMySQLURL(databaseURL)
		dialector = mysql.Open(dsn)
	} else if strings.HasPrefix(databaseURL, "postgresql://") || strings.HasPrefix(databaseURL, "postgres://") {
		// PostgreSQL
		dialector = postgres.Open(databaseURL)
	} else {
		return nil, fmt.Errorf("unsupported database URL format: %s", databaseURL)
	}

	// Configure logger
	logLevel := logger.Silent
	if debug {
		logLevel = logger.Info
	}

	// Open database connection
	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get underlying SQL DB for connection pool settings
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying SQL DB: %w", err)
	}

	// Set connection pool settings
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	// Test connection
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Database{DB: db}, nil
}

// Close closes the database connection
func (d *Database) Close() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// WithTransaction executes a function within a database transaction
func (d *Database) WithTransaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return d.DB.WithContext(ctx).Transaction(fn)
}

// GetDB returns the underlying GORM DB instance
func (d *Database) GetDB() *gorm.DB {
	return d.DB
}

// Helper function to convert MySQL URL format
// mysql://user:pass@host:port/dbname -> user:pass@tcp(host:port)/dbname
func convertMySQLURL(url string) string {
	url = strings.TrimPrefix(url, "mysql://")

	// Split into user:pass@host:port/dbname
	parts := strings.SplitN(url, "@", 2)
	if len(parts) != 2 {
		return url
	}

	userPass := parts[0]
	remaining := parts[1]

	// Split host:port/dbname
	parts = strings.SplitN(remaining, "/", 2)
	if len(parts) != 2 {
		return url
	}

	hostPort := parts[0]
	dbName := parts[1]

	// Reconstruct as MySQL DSN
	return fmt.Sprintf("%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", userPass, hostPort, dbName)
}

// TransactionManager provides transaction management
type TransactionManager struct {
	db *gorm.DB
}

// NewTransactionManager creates a new transaction manager
func NewTransactionManager(db *gorm.DB) *TransactionManager {
	return &TransactionManager{db: db}
}

// WithTransaction executes the given function within a transaction
func (tm *TransactionManager) WithTransaction(ctx context.Context, fn func(context.Context) error) error {
	return tm.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Store transaction in context
		txCtx := context.WithValue(ctx, txContextKey, tx)
		return fn(txCtx)
	})
}

// GetDB returns the DB from context if in transaction, otherwise returns the main DB
func (tm *TransactionManager) GetDB(ctx context.Context) *gorm.DB {
	if tx, ok := ctx.Value(txContextKey).(*gorm.DB); ok {
		return tx
	}
	return tm.db.WithContext(ctx)
}

// txContextKey is the key for storing transaction in context
type contextKey string

const txContextKey contextKey = "tx"

// GetTxFromContext retrieves the transaction from context
func GetTxFromContext(ctx context.Context) (*gorm.DB, bool) {
	tx, ok := ctx.Value(txContextKey).(*gorm.DB)
	return tx, ok
}

// BeginTx starts a new transaction and stores it in context
func BeginTx(ctx context.Context, db *gorm.DB) (context.Context, *gorm.DB, error) {
	tx := db.Begin(&sql.TxOptions{})
	if tx.Error != nil {
		return ctx, nil, tx.Error
	}

	txCtx := context.WithValue(ctx, txContextKey, tx)
	return txCtx, tx, nil
}
