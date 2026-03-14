// Package handlers provides HTTP handlers for metadata service
package handlers

import (
	"github.com/jmoiron/sqlx"

	"net/http"

	"github.com/jellyfinhanced/shared/response"
)

// ItemLookupHandler handles metadata lookup operations
type ItemLookupHandler struct {
	dbPool *sqlx.DB
}

// NewItemLookupHandler creates a new ItemLookupHandler
func NewItemLookupHandler(pool *sqlx.DB) *ItemLookupHandler {
	return &ItemLookupHandler{dbPool: pool}
}

// LookupItem handles GET /Items/{id}/Lookup
// @Summary Lookup item metadata
// @Description Lookup external metadata for an item
// @Tags Metadata
// @Param id path string true "Item ID"
// @Success 200 {object} dto.ItemMetadataDto
// @Router /Items/{id}/Lookup [get]
func (h *ItemLookupHandler) LookupItem(w http.ResponseWriter, r *http.Request) {
	itemId := r.PathValue("id")
	if itemId == "" {
		response.WriteBadRequest(w, "Item ID required")
		return
	}

	// Lookup metadata from external providers
	response.WriteNotImplemented(w, "Lookup not implemented")
}