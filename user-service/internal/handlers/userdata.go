package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/jellyfinhanced/shared/auth"
	"github.com/jellyfinhanced/shared/response"
	"github.com/jellyfinhanced/user-service/internal/db"
	"github.com/jellyfinhanced/user-service/internal/dto"
)

// MarkFavorite handles POST /Users/{userId}/FavoriteItems/{itemId}.
func (h *Handler) MarkFavorite(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userId")
	itemID := chi.URLParam(r, "itemId")
	callerID := auth.GetUserIDAsGUID(r.Context())

	if userID != callerID && !auth.IsAdminFromContext(r.Context()) {
		response.WriteError(w, http.StatusForbidden, "Access denied")
		return
	}

	if err := h.userDataRepo.SetFavorite(userID, itemID, true); err != nil {
		response.WriteError(w, http.StatusInternalServerError, "failed to mark favorite")
		return
	}

	ud, _ := h.userDataRepo.GetUserData(userID, itemID)
	response.WriteJSON(w, http.StatusOK, udToDto(ud, itemID))
}

// UnmarkFavorite handles DELETE /Users/{userId}/FavoriteItems/{itemId}.
func (h *Handler) UnmarkFavorite(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userId")
	itemID := chi.URLParam(r, "itemId")
	callerID := auth.GetUserIDAsGUID(r.Context())

	if userID != callerID && !auth.IsAdminFromContext(r.Context()) {
		response.WriteError(w, http.StatusForbidden, "Access denied")
		return
	}

	if err := h.userDataRepo.SetFavorite(userID, itemID, false); err != nil {
		response.WriteError(w, http.StatusInternalServerError, "failed to unmark favorite")
		return
	}

	ud, _ := h.userDataRepo.GetUserData(userID, itemID)
	response.WriteJSON(w, http.StatusOK, udToDto(ud, itemID))
}

// MarkPlayed handles POST /Users/{userId}/PlayedItems/{itemId}.
func (h *Handler) MarkPlayed(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userId")
	itemID := chi.URLParam(r, "itemId")
	callerID := auth.GetUserIDAsGUID(r.Context())

	if userID != callerID && !auth.IsAdminFromContext(r.Context()) {
		response.WriteError(w, http.StatusForbidden, "Access denied")
		return
	}

	if err := h.userDataRepo.MarkPlayed(userID, itemID, nil); err != nil {
		response.WriteError(w, http.StatusInternalServerError, "failed to mark played")
		return
	}

	ud, _ := h.userDataRepo.GetUserData(userID, itemID)
	response.WriteJSON(w, http.StatusOK, udToDto(ud, itemID))
}

// MarkUnplayed handles DELETE /Users/{userId}/PlayedItems/{itemId}.
func (h *Handler) MarkUnplayed(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userId")
	itemID := chi.URLParam(r, "itemId")
	callerID := auth.GetUserIDAsGUID(r.Context())

	if userID != callerID && !auth.IsAdminFromContext(r.Context()) {
		response.WriteError(w, http.StatusForbidden, "Access denied")
		return
	}

	if err := h.userDataRepo.MarkUnplayed(userID, itemID); err != nil {
		response.WriteError(w, http.StatusInternalServerError, "failed to mark unplayed")
		return
	}

	ud, _ := h.userDataRepo.GetUserData(userID, itemID)
	response.WriteJSON(w, http.StatusOK, udToDto(ud, itemID))
}

// MarkPlaying handles POST /Users/{userId}/PlayingItems (now-playing stub).
func (h *Handler) MarkPlaying(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func udToDto(ud *db.UserData, itemID string) dto.UserItemDataDto {
	out := dto.UserItemDataDto{
		ItemId: itemID,
		Key:    itemID,
	}
	if ud == nil {
		return out
	}
	out.IsFavorite = ud.IsFavorite
	out.Played = ud.Played
	out.PlayCount = ud.PlayCount
	out.PlaybackPositionTicks = ud.PlaybackPositionTicks
	return out
}
