// Package db provides shared database connection helpers for Kabletown services.
package db

import (
	"fmt"

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
}

// Connect opens and validates a MySQL connection using sqlx.
func Connect(cfg Config) (*sqlx.DB, error) {
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name,
	)
	db, err := sqlx.Connect("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("db.Connect: %w", err)
	}
	return db, nil
}
