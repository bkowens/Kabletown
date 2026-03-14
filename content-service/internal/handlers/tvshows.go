package handlers

import (
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jellyfinhanced/shared/logger"
)

var tvShowsLog = logger.NewLogger("tvshows-handler")

type TvShowsHandler struct {
	db  *sql.DB
}

func NewTvShowsHandler(dbPool *sql.DB) *TvShowsHandler {
	return &TvShowsHandler{db: dbPool}
}

func (h *TvShowsHandler) GetTvShows(w http.ResponseWriter, r *http.Request) {
	tvShowsLog.Info("Get TV shows requested")
	w.WriteHeader(http.StatusOK)
	render.JSON(w, r, map[string]interface{}{"Items": []interface{}{}})
}

func (h *TvShowsHandler) GetTvShow(w http.ResponseWriter, r *http.Request) {
	_ = chi.URLParam(r, "itemId")
	w.WriteHeader(http.StatusOK)
	render.JSON(w, r, map[string]interface{}{})
}

func (h *TvShowsHandler) GetEpisodes(w http.ResponseWriter, r *http.Request) {
	_ = chi.URLParam(r, "showId")
	w.WriteHeader(http.StatusOK)
	render.JSON(w, r, map[string]interface{}{"Items": []interface{}{}})
}

func (h *TvShowsHandler) GetSeasons(w http.ResponseWriter, r *http.Request) {
	_ = chi.URLParam(r, "showId")
	w.WriteHeader(http.StatusOK)
	render.JSON(w, r, map[string]interface{}{"Items": []interface{}{}})
}

func (h *TvShowsHandler) GetSeasonEpisodes(w http.ResponseWriter, r *http.Request) {
	_ = chi.URLParam(r, "seasonId")
	w.WriteHeader(http.StatusOK)
	render.JSON(w, r, map[string]interface{}{"Items": []interface{}{}})
}
