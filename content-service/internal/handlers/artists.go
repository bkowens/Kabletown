package handlers

import (
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/jellyfinhanced/shared/response"
)

// ArtistsHandler handles artist-related requests.
type ArtistsHandler struct {
	db *sql.DB
}

// NewArtistsHandler creates a new ArtistsHandler.
func NewArtistsHandler(dbPool *sql.DB) *ArtistsHandler {
	return &ArtistsHandler{db: dbPool}
}

// GetArtists returns a list of artists.
func (h *ArtistsHandler) GetArtists(w http.ResponseWriter, r *http.Request) {
	response.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"Items":            []interface{}{},
		"TotalRecordCount": 0,
		"StartIndex":       0,
	})
}

// GetArtist returns a specific artist.
func (h *ArtistsHandler) GetArtist(w http.ResponseWriter, r *http.Request) {
	_ = chi.URLParam(r, "artistId")
	response.WriteJSON(w, http.StatusOK, map[string]interface{}{})
}

// GetArtistAlbums returns albums for an artist.
func (h *ArtistsHandler) GetArtistAlbums(w http.ResponseWriter, r *http.Request) {
	_ = chi.URLParam(r, "artistId")
	response.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"Items":            []interface{}{},
		"TotalRecordCount": 0,
		"StartIndex":       0,
	})
}

// GetArtistSongs returns songs for an artist.
func (h *ArtistsHandler) GetArtistSongs(w http.ResponseWriter, r *http.Request) {
	_ = chi.URLParam(r, "artistId")
	response.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"Items":            []interface{}{},
		"TotalRecordCount": 0,
		"StartIndex":       0,
	})
}
