package handlers

import (
	"database/sql"
	"net/http"

	"github.com/jellyfinhanced/shared/logger"
)

var instantMixLog = logger.NewLogger("instantmix-handler")

type InstantMixHandler struct {
	db  *sql.DB
}

func NewInstantMixHandler(dbPool *sql.DB) *InstantMixHandler {
	return &InstantMixHandler{db: dbPool}
}

func (h *InstantMixHandler) CreateInstantMix(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	render.JSON(w, r, map[string]interface{}{"Items": []interface{}{}})
}

func (h *InstantMixHandler) CreateItemInstantMix(w http.ResponseWriter, r *http.Request) {
	_ = chi.URLParam(r, "itemId")
	w.WriteHeader(http.StatusOK)
	render.JSON(w, r, map[string]interface{}{"Items": []interface{}{}})
}
