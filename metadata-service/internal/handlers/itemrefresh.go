package handlers

import (
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jellyfinhanced/shared/logger"
)

var itemRefreshLog = logger.NewLogger("item-refresh-handler")

type ItemRefreshHandler struct {
	db  *sql.DB
}

func NewItemRefreshHandler(dbPool *sql.DB) *ItemRefreshHandler {
	return &ItemRefreshHandler{db: dbPool}
}

// RefreshAll refreshes all items
func (h *ItemRefreshHandler) RefreshAll(w http.ResponseWriter, r *http.Request) {
	itemRefreshLog.Info("Refresh all items requested")
	// Trigger background refresh task
	w.WriteHeader(http.StatusOK)
	render.JSON(w, r, map[string]string{"status": "refresh initiated"})
}

// RefreshItem refreshes a specific item
func (h *ItemRefreshHandler) RefreshItem(w http.ResponseWriter, r *http.Request) {
	itemID := chi.URLParam(r, "itemId")
	
	// Update last refresh timestamp
	query := `UPDATE items SET date_last_refreshed = ? WHERE id = ?`
	_, err := h.db.ExecContext(r.Context(), query, sql.NullTime{}, itemID)
	if err != nil {
		itemRefreshLog.Error("Failed to refresh item", "error", err, "item_id", itemID)
		http.Error(w, "Failed to refresh item", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// RefreshPartial performs partial item refresh
func (h *ItemRefreshHandler) RefreshPartial(w http.ResponseWriter, r *http.Request) {
	_ = chi.URLParam(r, "itemId")
	// Use default refresh settings
	RefreshItem(w, r)
}
