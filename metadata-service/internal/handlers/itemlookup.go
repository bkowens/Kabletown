package handlers

import (
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jellyfinhanced/shared/logger"
)

var itemLookupLog = logger.NewLogger("item-lookup-handler")

type ItemLookupHandler struct {
	db  *sql.DB
}

func NewItemLookupHandler(dbPool *sql.DB) *ItemLookupHandler {
	return &ItemLookupHandler{db: dbPool}
}

// GetProviderInfo returns provider information for an item
func (h *ItemLookupHandler) GetProviderInfo(w http.ResponseWriter, r *http.Request) {
	_ = chi.URLParam(r, "itemId")
	
	w.WriteHeader(http.StatusOK)
	render.JSON(w, r, map[string]interface{}{
		"providers": []string{"tmdb", "tvdb", "omdb"},
	})
}

// SearchMetadata searches for metadata
func (h *ItemLookupHandler) SearchMetadata(w http.ResponseWriter, r *http.Request) {
	_ = chi.URLParam(r, "itemId")
	
	w.WriteHeader(http.StatusOK)
	render.JSON(w, r, map[string]interface{}{
		"results": []interface{}{},
	})
}
