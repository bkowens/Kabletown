package handlers

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/jellyfinhanced/shared/response"
)

// TvShowsHandler handles TV show-related requests.
type TvShowsHandler struct {
	db *sql.DB
}

// NewTvShowsHandler creates a new TvShowsHandler.
func NewTvShowsHandler(dbPool *sql.DB) *TvShowsHandler {
	return &TvShowsHandler{db: dbPool}
}

// GetTvShows returns a list of TV shows.
func (h *TvShowsHandler) GetTvShows(w http.ResponseWriter, r *http.Request) {
	log.Printf("tvshows-handler: GetTvShows requested")
	response.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"Items":            []interface{}{},
		"TotalRecordCount": 0,
		"StartIndex":       0,
	})
}

// GetTvShow returns a specific TV show.
func (h *TvShowsHandler) GetTvShow(w http.ResponseWriter, r *http.Request) {
	_ = chi.URLParam(r, "itemId")
	response.WriteJSON(w, http.StatusOK, map[string]interface{}{})
}

// GetSeasons returns seasons for a series.
func (h *TvShowsHandler) GetSeasons(w http.ResponseWriter, r *http.Request) {
	_ = chi.URLParam(r, "seriesId")
	response.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"Items":            []interface{}{},
		"TotalRecordCount": 0,
		"StartIndex":       0,
	})
}

// GetEpisodes returns episodes for a series.
func (h *TvShowsHandler) GetEpisodes(w http.ResponseWriter, r *http.Request) {
	_ = chi.URLParam(r, "seriesId")
	response.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"Items":            []interface{}{},
		"TotalRecordCount": 0,
		"StartIndex":       0,
	})
}

// GetNextUp returns the next up episode for the user.
func (h *TvShowsHandler) GetNextUp(w http.ResponseWriter, r *http.Request) {
	response.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"Items":            []interface{}{},
		"TotalRecordCount": 0,
		"StartIndex":       0,
	})
}

// GetSeasonEpisodes returns episodes for a specific season.
func (h *TvShowsHandler) GetSeasonEpisodes(w http.ResponseWriter, r *http.Request) {
	_ = chi.URLParam(r, "seasonId")
	response.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"Items":            []interface{}{},
		"TotalRecordCount": 0,
		"StartIndex":       0,
	})
}
