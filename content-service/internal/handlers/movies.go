package handlers

import (
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jellyfinhanced/shared/logger"
)

var moviesLog = logger.NewLogger("movies-handler")

type MoviesHandler struct {
	db  *sql.DB
}

func NewMoviesHandler(dbPool *sql.DB) *MoviesHandler {
	return &MoviesHandler{db: dbPool}
}

func (h *MoviesHandler) GetMovies(w http.ResponseWriter, r *http.Request) {
	moviesLog.Info("Get movies requested")
	w.WriteHeader(http.StatusOK)
	render.JSON(w, r, map[string]interface{}{"Items": []interface{}{}})
}

func (h *MoviesHandler) GetMovie(w http.ResponseWriter, r *http.Request) {
	_ = chi.URLParam(r, "itemId")
	w.WriteHeader(http.StatusOK)
	render.JSON(w, r, map[string]interface{}{})
}

func (h *MoviesHandler) PlayMovie(w http.ResponseWriter, r *http.Request) {
	_ = chi.URLParam(r, "itemId")
	w.WriteHeader(http.StatusOK)
}
