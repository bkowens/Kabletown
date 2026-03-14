package handlers

import (
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/jellyfinhanced/shared/response"
)

// InstantMixHandler handles instant mix requests.
type InstantMixHandler struct {
	db *sql.DB
}

// NewInstantMixHandler creates a new InstantMixHandler.
func NewInstantMixHandler(dbPool *sql.DB) *InstantMixHandler {
	return &InstantMixHandler{db: dbPool}
}

// CreateInstantMix creates an instant mix.
func (h *InstantMixHandler) CreateInstantMix(w http.ResponseWriter, r *http.Request) {
	response.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"Items":            []interface{}{},
		"TotalRecordCount": 0,
		"StartIndex":       0,
	})
}

// CreateItemInstantMix creates an instant mix from a specific item.
func (h *InstantMixHandler) CreateItemInstantMix(w http.ResponseWriter, r *http.Request) {
	_ = chi.URLParam(r, "itemId")
	response.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"Items":            []interface{}{},
		"TotalRecordCount": 0,
		"StartIndex":       0,
	})
}
