package handlers

import (
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/jellyfinhanced/shared/response"
)

// PlaylistsHandler handles playlist-related requests.
type PlaylistsHandler struct {
	db *sql.DB
}

// NewPlaylistsHandler creates a new PlaylistsHandler.
func NewPlaylistsHandler(dbPool *sql.DB) *PlaylistsHandler {
	return &PlaylistsHandler{db: dbPool}
}

// GetPlaylists returns a list of playlists.
func (h *PlaylistsHandler) GetPlaylists(w http.ResponseWriter, r *http.Request) {
	response.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"Items":            []interface{}{},
		"TotalRecordCount": 0,
		"StartIndex":       0,
	})
}

// CreatePlaylist creates a new playlist.
func (h *PlaylistsHandler) CreatePlaylist(w http.ResponseWriter, r *http.Request) {
	response.WriteJSON(w, http.StatusCreated, map[string]interface{}{"Id": "new-playlist-id"})
}

// GetPlaylist returns a specific playlist.
func (h *PlaylistsHandler) GetPlaylist(w http.ResponseWriter, r *http.Request) {
	_ = chi.URLParam(r, "itemId")
	response.WriteJSON(w, http.StatusOK, map[string]interface{}{})
}

// AddToPlaylist adds items to a playlist.
func (h *PlaylistsHandler) AddToPlaylist(w http.ResponseWriter, r *http.Request) {
	_ = chi.URLParam(r, "itemId")
	w.WriteHeader(http.StatusNoContent)
}

// RemoveFromPlaylist removes an item from a playlist.
func (h *PlaylistsHandler) RemoveFromPlaylist(w http.ResponseWriter, r *http.Request) {
	_ = chi.URLParam(r, "itemId")
	_ = chi.URLParam(r, "childId")
	w.WriteHeader(http.StatusNoContent)
}
