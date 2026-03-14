package handlers

import (
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jellyfinhanced/shared/logger"
)

var collectionLog = logger.NewLogger("collection-handler")

type CollectionHandler struct {
	db  *sql.DB
}

func NewCollectionHandler(dbPool *sql.DB) *CollectionHandler {
	return &CollectionHandler{db: dbPool}
}

func (h *CollectionHandler) GetCollections(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	render.JSON(w, r, map[string]interface{}{"Items": []interface{}{}})
}

func (h *CollectionHandler) CreateCollection(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusCreated)
	render.JSON(w, r, map[string]interface{}{"id": "new-collection-id"})
}

func (h *CollectionHandler) GetCollection(w http.ResponseWriter, r *http.Request) {
	_ = chi.URLParam(r, "itemId")
	w.WriteHeader(http.StatusOK)
	render.JSON(w, r, map[string]interface{}{})
}

func (h *CollectionHandler) AddToCollection(w http.ResponseWriter, r *http.Request) {
	_ = chi.URLParam(r, "itemId")
	w.WriteHeader(http.StatusOK)
}

func (h *CollectionHandler) RemoveFromCollection(w http.ResponseWriter, r *http.Request) {
	_ = chi.URLParam(r, "itemId")
	_ = chi.URLParam(r, "childId")
	w.WriteHeader(http.StatusOK)
}
