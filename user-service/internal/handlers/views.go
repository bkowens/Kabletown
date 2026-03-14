package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/bowens/kabletown/shared/auth"
	"github.com/bowens/kabletown/shared/response"
	"github.com/bowens/kabletown/user-service/internal/dto"
)

// GetUserViews handles GET /Users/{userId}/Views.
// Returns a static set of top-level library views (populated by the media-service in production).
// For now we return an empty collection so clients don't error out.
func (h *Handler) GetUserViews(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userId")
	callerID := auth.GetUserID(r)

	if userID != callerID && !auth.IsAdmin(r) {
		response.Error(w, http.StatusForbidden, "Access denied")
		return
	}

	items := []dto.BaseItemDto{}
	response.JSON(w, http.StatusOK, map[string]interface{}{
		"Items":            items,
		"TotalRecordCount": 0,
		"StartIndex":       0,
	})
}

// GetLatestItems handles GET /Users/{userId}/Items/Latest.
// Returns an empty array; actual latest items come from the media-service.
func (h *Handler) GetLatestItems(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userId")
	callerID := auth.GetUserID(r)

	if userID != callerID && !auth.IsAdmin(r) {
		response.Error(w, http.StatusForbidden, "Access denied")
		return
	}

	response.JSON(w, http.StatusOK, []dto.BaseItemDto{})
}

// GetDisplayPreferences handles GET /DisplayPreferences/{displayPreferencesId}.
func (h *Handler) GetDisplayPreferences(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "displayPreferencesId")
	userID := auth.GetUserID(r)
	client := r.URL.Query().Get("client")
	if client == "" {
		client = "emby"
	}

	data, err := h.displayRepo.GetDisplayPreferences(id, userID, client)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to get display preferences")
		return
	}

	if data == "" {
		// Return default display preferences
		response.JSON(w, http.StatusOK, defaultDisplayPrefs(id, userID, client))
		return
	}

	var prefs map[string]interface{}
	if json.Unmarshal([]byte(data), &prefs) != nil {
		response.JSON(w, http.StatusOK, defaultDisplayPrefs(id, userID, client))
		return
	}
	response.JSON(w, http.StatusOK, prefs)
}

// SetDisplayPreferences handles POST /DisplayPreferences/{displayPreferencesId}.
func (h *Handler) SetDisplayPreferences(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "displayPreferencesId")
	userID := auth.GetUserID(r)
	client := r.URL.Query().Get("client")
	if client == "" {
		client = "emby"
	}

	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	b, _ := json.Marshal(body)
	if err := h.displayRepo.UpsertDisplayPreferences(id, userID, client, string(b)); err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to save display preferences")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func defaultDisplayPrefs(id, userID, client string) map[string]interface{} {
	return map[string]interface{}{
		"Id":                    id,
		"UserId":                userID,
		"Client":                client,
		"SortBy":                "SortName",
		"SortOrder":             "Ascending",
		"RememberIndexing":      false,
		"PrimaryImageHeight":    250,
		"PrimaryImageWidth":     0,
		"CustomPrefs":           map[string]interface{}{},
		"ScrollDirection":       "Horizontal",
		"ShowBackdrop":          true,
		"RememberSorting":       false,
		"ViewType":              "",
		"ShowSidebar":           false,
		"IndexBy":               nil,
	}
}
