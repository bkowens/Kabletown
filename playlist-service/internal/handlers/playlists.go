package handlers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/bowens/kabletown/playlist-service/internal/db"
	"github.com/bowens/kabletown/shared/response"
	"github.com/go-chi/chi/v5"
)

// getPlaylistRepo extracts repository from context
func getPlaylistRepo(r *http.Request) *db.PlaylistRepository {
	repo, ok := r.Context().Value("playlistRepo").(*db.PlaylistRepository)
	if !ok {
		log.Println("⚠️ playlistRepo not found in context")
		return nil
	}
	return repo
}

// GetPlaylists returns playlists for the current user (protected)
// GET /api/Playlists
func GetPlaylists(w http.ResponseWriter, r *http.Request) {
	repo := getPlaylistRepo(r)
	if repo == nil {
		response.InternalServerError(w, "Internal server error")
		return
	}

	// TODO: Get user ID from token/context
	userID := "00000000-0000-0000-0000-000000000000" // Placeholder

	playlists, err := repo.GetUserPlaylists(userID)
	if err != nil {
		response.InternalServerError(w, "Failed to fetch playlists")
		return
	}

	// W5: Use PaginatedResponse envelope (never bare array)
	// W6: ensure [] not null
	response.PaginatedResponse(w, playlists, len(playlists), 0, len(playlists))
}

// GetPlaylist returns a playlist by ID (public)
// GET /api/Playlists/{id}
func GetPlaylist(w http.ResponseWriter, r *http.Request) {
	playlistIDStr := chi.URLParam(r, "playlistId")
	playlistID, err := strconv.Atoi(playlistIDStr)
	if err != nil {
		response.BadRequest(w, "Invalid playlist ID")
		return
	}

	repo := getPlaylistRepo(r)
	if repo == nil {
		response.InternalServerError(w, "Internal server error")
		return
	}

	playlist, err := repo.GetPlaylistByID(playlistID)
	if err != nil {
		response.NotFound(w, "Playlist not found")
		return
	}

	response.OK(w, playlist)
}

// ListPlaylist returns a simple list endpoint (public)
// GET /api/Playlists (for discovery)
func ListPlaylist(w http.ResponseWriter, r *http.Request) {
	// W5: Use paginated envelope
	response.PaginatedResponse(w, []db.Playlist{}, 0, 0, 0)
}

// CreatePlaylist creates a new playlist (protected)
// POST /api/Playlists
func CreatePlaylist(w http.ResponseWriter, r *http.Request) {
	// TODO: Parse request body into Playlist
	// W1: Request body uses PascalCase: { "Name": "My Playlist", "UserId": "..." }
	// W2: UserId must be lowercase hyphenated GUID
	// W3: DateCreated will be automatically set with 7 decimal places

	repo := getPlaylistRepo(r)
	if repo == nil {
		response.InternalServerError(w, "Internal server error")
		return
	}

	// Placeholder - TODO: parse body
	playlist := &db.Playlist{
		Name:   "New Playlist",
		UserID: "00000000-0000-0000-0000-000000000000",
	}

	if err := repo.CreatePlaylist(playlist); err != nil {
		response.InternalServerError(w, "Failed to create playlist")
		return
	}

	response.Created(w, playlist)
}

// UpdatePlaylist updates an existing playlist (protected)
// PATCH /api/Playlists/{id}
func UpdatePlaylist(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement update logic
	response.OK(w, map[string]string{"Status": "Update not yet implemented"})
}

// DeletePlaylist deletes a playlist (protected)
// DELETE /api/Playlists/{id}
func DeletePlaylist(w http.ResponseWriter, r *http.Request) {
	playlistIDStr := chi.URLParam(r, "playlistId")
	playlistID, err := strconv.Atoi(playlistIDStr)
	if err != nil {
		response.BadRequest(w, "Invalid playlist ID")
		return
	}

	repo := getPlaylistRepo(r)
	if repo == nil {
		response.InternalServerError(w, "Internal server error")
		return
	}

	if err := repo.DeletePlaylist(playlistID); err != nil {
		response.NotFound(w, "Playlist not found")
		return
	}

	response.NoContent(w)
}

// AddToPlaylist adds an item to a playlist (protected)
// POST /api/Playlists/{id}/AddToPlaylist
func AddToPlaylist(w http.ResponseWriter, r *http.Request) {
	playlistIDStr := chi.URLParam(r, "playlistId")
	playlistID, err := strconv.Atoi(playlistIDStr)
	if err != nil {
		response.BadRequest(w, "Invalid playlist ID")
		return
	}

	repo := getPlaylistRepo(r)
	if repo == nil {
		response.InternalServerError(w, "Internal server error")
		return
	}

	// TODO: Parse libraryItemID from body
	// W2: libraryItemID must be lowercase hyphenated GUID
	libraryItemID := "00000000-0000-0000-0000-000000000000" // Placeholder

	item, err := repo.AddItemToPlaylist(playlistID, libraryItemID)
	if err != nil {
		response.InternalServerError(w, "Failed to add item to playlist")
		return
	}

	response.Created(w, item)
}

// GetPlaylistItems returns all items in a playlist (public)
// GET /api/Playlists/{playlistId}/Items
func GetPlaylistItems(w http.ResponseWriter, r *http.Request) {
	playlistIDStr := chi.URLParam(r, "playlistId")
	playlistID, err := strconv.Atoi(playlistIDStr)
	if err != nil {
		response.BadRequest(w, "Invalid playlist ID")
		return
	}

	repo := getPlaylistRepo(r)
	if repo == nil {
		response.InternalServerError(w, "Internal server error")
		return
	}

	items, err := repo.GetPlaylistItems(playlistID)
	if err != nil {
		response.NotFound(w, "Playlist not found")
		return
	}

	// W5: Use PaginatedResponse envelope (never bare array)
	// W6: ensure [] not null
	response.PaginatedResponse(w, items, len(items), 0, len(items))
}

// RemoveFromPlaylist removes an item from a playlist (protected)
// DELETE /api/Playlists/{playlistId}/RemoveFromPlaylist
func RemoveFromPlaylist(w http.ResponseWriter, r *http.Request) {
	// TODO: Parse item ID from body or query params
	response.OK(w, map[string]string{"Status": "Remove not yet implemented"})
}
