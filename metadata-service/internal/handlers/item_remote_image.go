// Package handlers provides HTTP handlers for metadata service
package handlers

import (
	"github.com/jmoiron/sqlx"

	"net/http"

	"github.com/jellyfinhanced/shared/response"
	"github.com/jellyfinhanced/shared/types"
)

// ItemRemoteImageHandler handles remote image operations
type ItemRemoteImageHandler struct {
	dbPool *sqlx.DB
}

// NewItemRemoteImageHandler creates a new ItemRemoteImageHandler
func NewItemRemoteImageHandler(pool *sqlx.DB) *ItemRemoteImageHandler {
	return &ItemRemoteImageHandler{dbPool: pool}
}

// SearchImages handles GET /Items/{id}/RemoteImage
// @Summary Search for remote images
// @Description Search for remote images for an item
// @Tags Metadata
// @Param id path string true "Item ID"
// @Success 200 {array} dto.ImageInfoDto
// @Router /Items/{id}/RemoteImage [get]
func (h *ItemRemoteImageHandler) SearchImages(w http.ResponseWriter, r *http.Request) {
	itemId := r.PathValue("id")
	if itemId == "" || !types.IsValidGUID(itemId) {
		response.WriteBadRequest(w, "Valid item ID required")
		return
	}

	// Search external image providers
	response.WriteJSON(w, http.StatusOK, []string{})
}