// Package handlers provides HTTP handlers for metadata service
package handlers

import (
	"github.com/jmoiron/sqlx"

	"net/http"

	"github.com/jellyfinhanced/shared/response"
	"github.com/jellyfinhanced/shared/types"
)

// ItemUpdateHandler handles metadata update operations
type ItemUpdateHandler struct {
	dbPool *sqlx.DB
}

// NewItemUpdateHandler creates a new ItemUpdateHandler
func NewItemUpdateHandler(pool *sqlx.DB) *ItemUpdateHandler {
	return &ItemUpdateHandler{dbPool: pool}
}

// UpdateItem handles POST /Items/{id}/UpdateMetadata
// @Summary Update item metadata
// @Description Update metadata for a specific item
// @Tags Metadata
// @Param id path string true "Item ID"
// @Success 204
// @Router /Items/{id}/UpdateMetadata [post]
func (h *ItemUpdateHandler) UpdateItem(w http.ResponseWriter, r *http.Request) {
	itemId := r.PathValue("id")
	if itemId == "" || !types.IsValidGUID(itemId) {
		response.WriteBadRequest(w, "Valid item ID required")
		return
	}

	// Update metadata from external sources
	// This would fetch new metadata from themoviedb, themoviedb, etc.
	response.WriteNoContent(w)
}