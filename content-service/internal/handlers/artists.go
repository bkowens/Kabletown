package handlers

import (
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jellyfinhanced/shared/logger"
)

var artistsLog = logger.NewLogger("artists-handler")

type ArtistsHandler struct {
	db  *sql.DB
}

func NewArtistsHandler(dbPool *sql.DB) *ArtistsHandler {
	return &ArtistsHandler{db: dbPool}
}

func (h *ArtistsHandler) GetArtists(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	render.JSON(w, r, map[string]interface{}{"Items": []interface{}{}})
}

func (h *ArtistsHandler) GetArtist(w http.ResponseWriter, r *http.Request) {
	_ = chi.URLParam(r, "artistId")
	w.WriteHeader(http.StatusOK)
	render.JSON(w, r, map[string]interface{}{})
}

func (h *ArtistsHandler) GetArtistAlbums(w http.ResponseWriter, r *http.Request) {
	_ = chi.URLParam(r, "artistId")
	w.WriteHeader(http.StatusOK)
	render.JSON(w, r, map[string]interface{}{"Items": []interface{}{}})
}

func (h *ArtistsHandler) GetArtistSongs(w http.ResponseWriter, r *http.Request) {
	_ = chi.URLParam(r, "artistId")
	w.WriteHeader(http.StatusOK)
	render.JSON(w, r, map[string]interface{}{"Items": []interface{}{}})
}
