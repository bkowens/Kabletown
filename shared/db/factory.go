// Package db provides shared database connection helpers for Kabletown services.
package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

// Config holds the parameters needed to open a MySQL connection.
type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	// Pool settings (optional, uses defaults if not set)
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

// NewDB creates a new SQLx database connection with pool configuration.
// Main factory function - uses default pool settings if not specified.
func NewDB(dsn string) (*sqlx.DB, error) {
	return NewDBWithConfig(dsn, Config{})
}

// NewDBWithConfig creates a new SQLx database connection with custom pool config.
func NewDBWithConfig(dsn string, cfg Config) (*sqlx.DB, error) {
	db, err := sqlx.Connect("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("db.NewDB: %w", err)
	}

	// Apply pool settings or use defaults
	applyPoolSettings(db.DB, cfg)

	// Verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("db.NewDB: failed to ping: %w", err)
	}

	return db, nil
}

// Connect opens and validates a MySQL connection using sqlx (legacy helper).
// Deprecated: Use NewDB() or NewDBWithConfig() instead.
func Connect(cfg Config) (*sqlx.DB, error) {
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name,
	)
	return NewDBWithConfig(dsn, cfg)
}

// DefaultConfig returns common pool configuration
func DefaultConfig() Config {
	return Config{
		MaxOpenConns:    25,
		MaxIdleConns:    10,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 1 * time.Minute,
	}
}

// applyPoolSettings applies pool settings to the database connection.
func applyPoolSettings(sqldb *sql.DB, cfg Config) {
	defaults := DefaultConfig()

	// Use cfg values or defaults
	maxOpen := cfg.MaxOpenConns
	if maxOpen <= 0 {
		maxOpen = defaults.MaxOpenConns
	}

	maxIdle := cfg.MaxIdleConns
	if maxIdle <= 0 {
		maxIdle = defaults.MaxIdleConns
	}

	lifetime := cfg.ConnMaxLifetime
	if lifetime <= 0 {
		lifetime = defaults.ConnMaxLifetime
	}

	idleTime := cfg.ConnMaxIdleTime
	if idleTime <= 0 {
		idleTime = defaults.ConnMaxIdleTime
	}

	sqldb.SetMaxOpenConns(maxOpen)
	sqldb.SetMaxIdleConns(maxIdle)
	sqldb.SetConnMaxLifetime(lifetime)
	sqldb.SetConnMaxIdleTime(idleTime)
}

// WithTx executes a function within a database transaction.
// If the function returns an error, the transaction is rolled back.
// Otherwise, the transaction is committed.
// This is the preferred way to perform multi-table operations.
//
// Example usage:
//
//	err := db.WithTx(func(tx *sqlx.Tx) error {
//	    if _, err := tx.Exec("INSERT INTO ..."); err != nil {
//	        return err
//	    }
//	    if _, err := tx.Exec("UPDATE ..."); err != nil {
//	        return err
//	    }
//	    return nil
//
//	})
func WithTx(db *sqlx.DB, fn func(tx *sqlx.Tx) error) error {
	tx, err := db.BeginTxx(context.Background(), nil)
	if err != nil {
		return fmt.Errorf("db.WithTx: failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p) // re-throw panic after rollback
		} else if err != nil {
			tx.Rollback()
		}
	}()

	err = fn(tx)
	if err != nil {
		return fmt.Errorf("db.WithTx: transaction failed: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("db.WithTx: failed to commit: %w", err)
	}

	return nil
}

// WithTxContext executes a function within a database transaction with context.
// Similar to WithTx but accepts a context for timeout/cancellation support.
func WithTxContext(ctx context.Context, db *sqlx.DB, fn func(tx *sqlx.Tx) error) error {
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("db.WithTxContext: failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		}
	}()

	err = fn(tx)
	if err != nil {
		return fmt.Errorf("db.WithTxContext: transaction failed: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("db.WithTxContext: failed to commit: %w", err)
	}

	return nil
}

// NewMySQLPool creates a standard sql.DB connection pool.
// This is provided for services that don't need sqlx features.
func NewMySQLPool(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("db.NewMySQLPool: %w", err)
	}

	// Apply default pool settings
	cfg := DefaultConfig()
	applyPoolSettings(db, cfg)

	// Verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("db.NewMySQLPool: failed to ping: %w", err)
	}

	return db, nil
}
