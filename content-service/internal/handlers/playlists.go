package handlers

import (
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jellyfinhanced/shared/logger"
)

var playlistsLog = logger.NewLogger("playlists-handler")

type PlaylistsHandler struct {
	db  *sql.DB
}

func NewPlaylistsHandler(dbPool *sql.DB) *PlaylistsHandler {
	return &PlaylistsHandler{db: dbPool}
}

func (h *PlaylistsHandler) GetPlaylists(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	render.JSON(w, r, map[string]interface{}{"Items": []interface{}{}})
}

func (h *PlaylistsHandler) CreatePlaylist(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusCreated)
	render.JSON(w, r, map[string]interface{}{"id": "new-playlist-id"})
}

func (h *PlaylistsHandler) GetPlaylist(w http.ResponseWriter, r *http.Request) {
	_ = chi.URLParam(r, "itemId")
	w.WriteHeader(http.StatusOK)
	render.JSON(w, r, map[string]interface{}{})
}

func (h *PlaylistsHandler) AddToPlaylist(w http.ResponseWriter, r *http.Request) {
	_ = chi.URLParam(r, "itemId")
	w.WriteHeader(http.StatusOK)
}

func (h *PlaylistsHandler) RemoveFromPlaylist(w http.ResponseWriter, r *http.Request) {
	_ = chi.URLParam(r, "itemId")
	_ = chi.URLParam(r, "childId")
	w.WriteHeader(http.StatusOK)
}
