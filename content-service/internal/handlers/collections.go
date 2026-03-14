package handlers

import (
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/jellyfinhanced/shared/response"
)

// CollectionHandler handles collection-related requests.
type CollectionHandler struct {
	db *sql.DB
}

// NewCollectionHandler creates a new CollectionHandler.
func NewCollectionHandler(dbPool *sql.DB) *CollectionHandler {
	return &CollectionHandler{db: dbPool}
}

// GetCollections returns a list of collections.
func (h *CollectionHandler) GetCollections(w http.ResponseWriter, r *http.Request) {
	response.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"Items":            []interface{}{},
		"TotalRecordCount": 0,
		"StartIndex":       0,
	})
}

// CreateCollection creates a new collection.
func (h *CollectionHandler) CreateCollection(w http.ResponseWriter, r *http.Request) {
	response.WriteJSON(w, http.StatusCreated, map[string]interface{}{"Id": "new-collection-id"})
}

// GetCollection returns a specific collection.
func (h *CollectionHandler) GetCollection(w http.ResponseWriter, r *http.Request) {
	_ = chi.URLParam(r, "itemId")
	response.WriteJSON(w, http.StatusOK, map[string]interface{}{})
}

// AddToCollection adds items to a collection.
func (h *CollectionHandler) AddToCollection(w http.ResponseWriter, r *http.Request) {
	_ = chi.URLParam(r, "itemId")
	w.WriteHeader(http.StatusNoContent)
}

// RemoveFromCollection removes an item from a collection.
func (h *CollectionHandler) RemoveFromCollection(w http.ResponseWriter, r *http.Request) {
	_ = chi.URLParam(r, "itemId")
	_ = chi.URLParam(r, "childId")
	w.WriteHeader(http.StatusNoContent)
}
