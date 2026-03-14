// Package handlers provides HTTP handlers for library service.
package handlers

import (
	"github.com/jmoiron/sqlx"
)

// Handler is the base handler with a shared database connection.
type Handler struct {
	db *sqlx.DB
}

// NewHandler creates a new base handler.
func NewHandler(db *sqlx.DB) *Handler {
	return &Handler{db: db}
}
