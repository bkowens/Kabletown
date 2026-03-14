// Package handlers implements HTTP route handlers for search-service.
package handlers

import (
	"github.com/go-chi/chi/v5"

	internalDB "github.com/jellyfinhanced/search-service/internal/db"
)

// Handler holds dependencies for search-service route handlers.
type Handler struct {
	serverID   string
	serverName string
	repo       *internalDB.SearchRepository
}

// New creates a Handler with the given server identity and search repository.
func New(serverID, serverName string, repo *internalDB.SearchRepository) *Handler {
	return &Handler{
		serverID:   serverID,
		serverName: serverName,
		repo:       repo,
	}
}

// RegisterRoutes wires all search-service routes onto the given chi router.
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get("/Search/Hints", h.SearchHints)
	r.Get("/Items", h.ListItems)
}
