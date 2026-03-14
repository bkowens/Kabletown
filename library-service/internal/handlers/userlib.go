// Package handlers provides HTTP handlers for library service.
package handlers

import (
	"net/http"

	"github.com/jellyfinhanced/shared/response"
	"github.com/jmoiron/sqlx"

	"github.com/jellyfinhanced/library-service/internal/dto"
)

// UserLibHandler handles user library endpoints.
type UserLibHandler struct {
	db *sqlx.DB
}

// NewUserLibHandler creates a new UserLibHandler.
func NewUserLibHandler(db *sqlx.DB) *UserLibHandler {
	return &UserLibHandler{db: db}
}

// GetUserViews handles GET /Users/{userId}/Views.
func (h *UserLibHandler) GetUserViews(w http.ResponseWriter, r *http.Request) {
	userId := r.PathValue("userId")
	if userId == "" {
		response.WriteBadRequest(w, "User ID required")
		return
	}

	views := []dto.LibraryView{
		{Id: "movies", Name: "Movies", Type: "MovieLibrary"},
		{Id: "tvshows", Name: "TV Shows", Type: "TvSeriesLibrary"},
		{Id: "music", Name: "Music", Type: "MusicLibrary"},
	}

	response.WriteJSON(w, http.StatusOK, dto.QueryResult[dto.LibraryView]{
		Items:            views,
		TotalRecordCount: int64(len(views)),
	})
}

// RegisterUserLibRoutes registers user library routes.
func RegisterUserLibRoutes(mux *http.ServeMux, h *UserLibHandler) {
	mux.HandleFunc("GET /Users/{userId}/Views", h.GetUserViews)
}
