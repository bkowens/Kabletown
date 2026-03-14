package handlers

import (
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jellyfinhanced/shared/logger"
)

var liveTvLog = logger.NewLogger("livetv-handler")

type LiveTvHandler struct {
	db  *sql.DB
}

func NewLiveTvHandler(dbPool *sql.DB) *LiveTvHandler {
	return &LiveTvHandler{db: dbPool}
}

func (h *LiveTvHandler) GetPrograms(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	render.JSON(w, r, map[string]interface{}{"Items": []interface{}{}})
}

func (h *LiveTvHandler) GetLiveChannels(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	render.JSON(w, r, map[string]interface{}{"Items": []interface{}{}})
}

func (h *LiveTvHandler) GetChannelPrograms(w http.ResponseWriter, r *http.Request) {
	_ = chi.URLParam(r, "channelId")
	w.WriteHeader(http.StatusOK)
	render.JSON(w, r, map[string]interface{}{"Items": []interface{}{}})
}

func (h *LiveTvHandler) GetProgram(w http.ResponseWriter, r *http.Request) {
	_ = chi.URLParam(r, "programId")
	w.WriteHeader(http.StatusOK)
	render.JSON(w, r, map[string]interface{}{})
}
