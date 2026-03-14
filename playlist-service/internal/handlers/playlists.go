package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/jellyfinhanced/playlist-service/internal/db"
	"github.com/jellyfinhanced/shared/response"
)

type contextKeyType string

const repoContextKey contextKeyType = "playlistRepo"

// RegisterRoutes wires all playlist routes onto the provided router.
func RegisterRoutes(r chi.Router, repo *db.PlaylistRepository) {
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			ctx := context.WithValue(req.Context(), repoContextKey, repo)
			next.ServeHTTP(w, req.WithContext(ctx))
		})
	})

	r.Post("/Playlists", CreatePlaylist)
	r.Get("/Playlists/{playlistId}", GetPlaylist)
	r.Delete("/Playlists/{playlistId}", DeletePlaylist)
	r.Get("/Playlists/{playlistId}/Items", GetPlaylistItems)
	r.Post("/Playlists/{playlistId}/Items", func(w http.ResponseWriter, r *http.Request) { response.WriteNotImplemented(w, "Add to playlist not implemented") })
	r.Delete("/Playlists/{playlistId}/Items", RemoveFromPlaylist)
	r.Post("/Playlists/{playlistId}/Items/Move/{itemId}", MovePlaylistItem)
}

// getRepo extracts the PlaylistRepository from the request context.
func getRepo(r *http.Request) *db.PlaylistRepository {
	repo, _ := r.Context().Value(repoContextKey).(*db.PlaylistRepository)
	return repo
}

// CreatePlaylistRequest is the POST /Playlists request body.
type CreatePlaylistRequest struct {
	Name    string   `json:"Name"`
	UserId  string   `json:"UserId"`
	ItemIds []string `json:"ItemIds,omitempty"`
}

// CreatePlaylist handles POST /Playlists.
func CreatePlaylist(w http.ResponseWriter, r *http.Request) {
	var req CreatePlaylistRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" {
		response.WriteError(w, http.StatusBadRequest, "Name is required")
		return
	}

	userID := req.UserId
	if userID == "" {
		userID = r.URL.Query().Get("userId")
	}
	if userID == "" {
		response.WriteError(w, http.StatusBadRequest, "UserId is required")
		return
	}

	repo := getRepo(r)
	if repo == nil {
		response.WriteError(w, http.StatusInternalServerError, "repository not initialized")
		return
	}

	playlist := &db.Playlist{
		Name:   req.Name,
		UserID: userID,
	}

	if err := repo.CreatePlaylist(playlist); err != nil {
		response.WriteError(w, http.StatusInternalServerError, "failed to create playlist")
		return
	}

	for _, itemID := range req.ItemIds {
		if _, err := repo.AddItemToPlaylist(playlist.ID, strings.TrimSpace(itemID)); err != nil {
			_ = repo.DeletePlaylist(playlist.ID)
			response.WriteError(w, http.StatusInternalServerError, "failed to add item to playlist")
			return
		}
	}

	response.WriteJSON(w, http.StatusCreated, playlist)
}

// GetPlaylist handles GET /Playlists/{playlistId}.
func GetPlaylist(w http.ResponseWriter, r *http.Request) {
	playlistID, err := parsePlaylistID(r)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid playlist ID")
		return
	}

	repo := getRepo(r)
	if repo == nil {
		response.WriteError(w, http.StatusInternalServerError, "repository not initialized")
		return
	}

	playlist, err := repo.GetPlaylistByID(playlistID)
	if err == db.ErrPlaylistNotFound {
		response.WriteError(w, http.StatusNotFound, "playlist not found")
		return
	}
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "failed to fetch playlist")
		return
	}

	response.WriteJSON(w, http.StatusOK, playlist)
}

// DeletePlaylist handles DELETE /Playlists/{playlistId}.
func DeletePlaylist(w http.ResponseWriter, r *http.Request) {
	playlistID, err := parsePlaylistID(r)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid playlist ID")
		return
	}

	repo := getRepo(r)
	if repo == nil {
		response.WriteError(w, http.StatusInternalServerError, "repository not initialized")
		return
	}

	if err := repo.DeletePlaylist(playlistID); err != nil {
		if err == db.ErrPlaylistNotFound {
			response.WriteError(w, http.StatusNotFound, "playlist not found")
			return
		}
		response.WriteError(w, http.StatusInternalServerError, "failed to delete playlist")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetPlaylistItems handles GET /Playlists/{playlistId}/Items.
func GetPlaylistItems(w http.ResponseWriter, r *http.Request) {
	playlistID, err := parsePlaylistID(r)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid playlist ID")
		return
	}

	repo := getRepo(r)
	if repo == nil {
		response.WriteError(w, http.StatusInternalServerError, "repository not initialized")
		return
	}

	items, err := repo.GetPlaylistItems(playlistID)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "failed to fetch playlist items")
		return
	}

	// WritePaginated(w, "playlists", items, totalCount)
	// if !result {
	//	return
	// }
	response.WriteJSON(w, http.StatusOK, items)
	return
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid playlist ID")
		return
	}

	var req struct {
		ItemIds []string `json:"ItemIds"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	repo = getRepo(r)
	if repo == nil {
		response.WriteError(w, http.StatusInternalServerError, "repository not initialized")
		return
	}

	var added []*db.PlaylistItem
	for _, itemID := range req.ItemIds {
		item, err := repo.AddItemToPlaylist(playlistID, strings.TrimSpace(itemID))
		if err != nil {
			response.WriteError(w, http.StatusInternalServerError, "failed to add item to playlist")
			return
		}
		added = append(added, item)
	}

	response.WriteJSON(w, http.StatusOK, map[string]string{"Message": "items added to playlist"})
}

// RemoveFromPlaylist handles DELETE /Playlists/{playlistId}/Items.
// Accepts a JSON body with ItemIds (library item IDs) or EntryIds (playlist item row IDs).
func RemoveFromPlaylist(w http.ResponseWriter, r *http.Request) {
	playlistID, err := parsePlaylistID(r)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid playlist ID")
		return
	}

	var req struct {
		ItemIds  []string `json:"ItemIds,omitempty"`
		EntryIds []string `json:"EntryIds,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	repo := getRepo(r)
	if repo == nil {
		response.WriteError(w, http.StatusInternalServerError, "repository not initialized")
		return
	}

	for _, libraryItemID := range req.ItemIds {
		if err := repo.RemoveItemFromPlaylistByLibraryID(playlistID, strings.TrimSpace(libraryItemID)); err != nil {
			if err == db.ErrPlaylistItemNotFound {
				response.WriteError(w, http.StatusNotFound, "playlist item not found")
				return
			}
			response.WriteError(w, http.StatusInternalServerError, "failed to remove item")
			return
		}
	}

	for _, entryIDStr := range req.EntryIds {
		entryID, err := strconv.Atoi(strings.TrimSpace(entryIDStr))
		if err != nil {
			response.WriteError(w, http.StatusBadRequest, "invalid entry ID")
			return
		}
		if err := repo.RemoveItemFromPlaylist(playlistID, entryID); err != nil {
			if err == db.ErrPlaylistItemNotFound {
				response.WriteError(w, http.StatusNotFound, "playlist item not found")
				return
			}
			response.WriteError(w, http.StatusInternalServerError, "failed to remove item")
			return
		}
	}

	response.WriteJSON(w, http.StatusOK, map[string]string{"Message": "items removed from playlist"})
}

// MovePlaylistItem handles POST /Playlists/{playlistId}/Items/Move/{itemId}.
func MovePlaylistItem(w http.ResponseWriter, r *http.Request) {
	playlistID, err := parsePlaylistID(r)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid playlist ID")
		return
	}

	itemIDStr := chi.URLParam(r, "itemId")
	itemID, err := strconv.Atoi(itemIDStr)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid item ID")
		return
	}

	newIndexStr := r.URL.Query().Get("newIndex")
	newIndex, err := strconv.Atoi(newIndexStr)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid newIndex")
		return
	}

	repo := getRepo(r)
	if repo == nil {
		response.WriteError(w, http.StatusInternalServerError, "repository not initialized")
		return
	}

	if err := repo.MoveItem(playlistID, itemID, newIndex); err != nil {
		response.WriteError(w, http.StatusInternalServerError, "failed to move item")
		return
	}

	response.WriteJSON(w, http.StatusOK, map[string]string{"Message": "item moved"})
}

// parsePlaylistID extracts and validates the playlistId URL parameter.
func parsePlaylistID(r *http.Request) (int, error) {
	return strconv.Atoi(chi.URLParam(r, "playlistId"))
}