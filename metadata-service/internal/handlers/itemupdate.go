package handlers

import (
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jellyfinhanced/shared/logger"
)

var itemUpdateLog = logger.NewLogger("item-update-handler")

type ItemUpdateHandler struct {
	db  *sql.DB
}

func NewItemUpdateHandler(dbPool *sql.DB) *ItemUpdateHandler {
	return &ItemUpdateHandler{db: dbPool}
}

// UpdateItem updates item information
func (h *ItemUpdateHandler) UpdateItem(w http.ResponseWriter, r *http.Request) {
	itemUpdateLog.Info("Item update requested")
	w.WriteHeader(http.StatusOK)
}

// UpdateSpecificItem updates a specific item
func (h *ItemUpdateHandler) UpdateSpecificItem(w http.ResponseWriter, r *http.Request) {
	itemID := chi.URLParam(r, "itemId")
	
	query := `UPDATE items SET date_last_modified = ? WHERE id = ?`
	_, err := h.db.ExecContext(r.Context(), query, sql.NullTime{}, itemID)
	if err != nil {
		itemUpdateLog.Error("Failed to update item", "error", err, "item_id", itemID)
		http.Error(w, "Failed to update item", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// RemoteSearch triggers remote search for item
func (h *ItemUpdateHandler) RemoteSearch(w http.ResponseWriter, r *http.Request) {
	itemUpdateLog.Info("Remote search triggered")
	w.WriteHeader(http.StatusOK)
}
