package handlers

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/jellyfinhanced/shared/response"
)

// MoviesHandler handles movie-related requests.
type MoviesHandler struct {
	db *sql.DB
}

// NewMoviesHandler creates a new MoviesHandler.
func NewMoviesHandler(dbPool *sql.DB) *MoviesHandler {
	return &MoviesHandler{db: dbPool}
}

// GetMovies returns a list of movies.
func (h *MoviesHandler) GetMovies(w http.ResponseWriter, r *http.Request) {
	log.Printf("movies-handler: GetMovies requested")
	response.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"Items":            []interface{}{},
		"TotalRecordCount": 0,
		"StartIndex":       0,
	})
}

// GetMovie returns a specific movie.
func (h *MoviesHandler) GetMovie(w http.ResponseWriter, r *http.Request) {
	_ = chi.URLParam(r, "itemId")
	response.WriteJSON(w, http.StatusOK, map[string]interface{}{})
}

// PlayMovie initiates playback for a movie.
func (h *MoviesHandler) PlayMovie(w http.ResponseWriter, r *http.Request) {
	_ = chi.URLParam(r, "itemId")
	w.WriteHeader(http.StatusNoContent)
}
