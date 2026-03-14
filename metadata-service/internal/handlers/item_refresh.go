// Package handlers provides HTTP handlers for metadata service
package handlers

import (
	"github.com/jmoiron/sqlx"

	"net/http"

	"github.com/jellyfinhanced/metadata-service/internal/dto"
	"github.com/jellyfinhanced/shared/response"
	"github.com/jellyfinhanced/shared/types"
)

// ItemRefreshHandler handles metadata refresh operations
type ItemRefreshHandler struct {
	dbPool *sqlx.DB
}

// NewItemRefreshHandler creates a new ItemRefreshHandler
func NewItemRefreshHandler(pool *sqlx.DB) *ItemRefreshHandler {
	return &ItemRefreshHandler{dbPool: pool}
}

// RefreshItem handles POST /Items/{id}/Refresh
// @Summary Refresh item metadata
// @Description Force metadata refresh for a specific item
// @Tags Metadata
// @Param id path string true "Item ID"
// @Param overwriteMetadata query bool false "Overwrite existing metadata" default(false)
// @Param recursive query bool false "Refresh child items" default(false)
// @Success 204
// @Router /Items/{id}/Refresh [post]
func (h *ItemRefreshHandler) RefreshItem(w http.ResponseWriter, r *http.Request) {
	itemId := r.PathValue("id")
	if itemId == "" || !types.IsValidGUID(itemId) {
		response.WriteBadRequest(w, "Valid item ID required")
		return
	}

	overwrite := r.URL.Query().Get("OverwriteMetadata") == "true"
	recursive := r.URL.Query().Get("Recursive") == "true"

	// Build refresh task record
	task := &dto.MetadataRefreshTask{
		ItemId:    itemId,
		UserId:    "",
		Overwrite: overwrite,
		Recursive: recursive,
		Status:    dto.Task{Id: "queued", Name: "refresh", State: "queued"},
	}

	// Queue refresh task
	err := h.queueRefreshTask(r.Context(), task)
	if err != nil {
		response.WriteInternalServerError(w, "Failed to queue refresh task")
		return
	}

	response.WriteNoContent(w)
}

// RefreshLibrary handles POST /Library/Refresh
// @Summary Refresh library
// @Description Trigger library scan and metadata refresh
// @Tags Metadata
// @Success 204
// @Router /Library/Refresh [post]
func (h *ItemRefreshHandler) RefreshLibrary(w http.ResponseWriter, r *http.Request) {
	// Trigger background library scan
	// This would be implemented via scheduled task system
	response.WriteJSON(w, http.StatusOK, dto.MessageDto{
		Message: "Library refresh queued",
	})
}

// CancelRefresh handles DELETE /Items/{id}/Refresh
// @Summary Cancel refresh
// @Description Cancel pending refresh task for an item
// @Tags Metadata
// @Param id path string true "Item ID"
// @Success 204
// @Router /Items/{id}/Refresh [delete]
func (h *ItemRefreshHandler) CancelRefresh(w http.ResponseWriter, r *http.Request) {
	itemId := r.PathValue("id")
	if itemId == "" {
		response.WriteBadRequest(w, "Item ID required")
		return
	}

	err := h.cancelRefreshTask(r.Context(), itemId)
	if err != nil {
		response.WriteInternalServerError(w, "Failed to cancel refresh")
		return
	}

	response.WriteNoContent(w)
}

// queueRefreshTask queues a metadata refresh task in the database
func (h *ItemRefreshHandler) queueRefreshTask(ctx interface{}, task *dto.MetadataRefreshTask) error {
	// Implementation: Insert task into scheduled_tasks table
	return nil // TODO: Implement
}

// cancelRefreshTask cancels a pending refresh task
func (h *ItemRefreshHandler) cancelRefreshTask(ctx interface{}, itemId string) error {
	// Implementation: Update task status to cancelled
	return nil // TODO: Implement
}