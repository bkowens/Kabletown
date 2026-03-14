package handlers

import (
	"net/http"

	"github.com/jellyfinhanced/shared/response"
	"github.com/google/uuid"
)

// GetSubtitleStream serves a subtitle stream by format extension — stub returns 404.
func (h *Handler) GetSubtitleStream(w http.ResponseWriter, r *http.Request) {
	response.WriteError(w, http.StatusNotFound, "Subtitle not found")
}

// GetSubtitleStreamNoFormat serves a subtitle stream without a format — stub returns 404.
func (h *Handler) GetSubtitleStreamNoFormat(w http.ResponseWriter, r *http.Request) {
	response.WriteError(w, http.StatusNotFound, "Subtitle not found")
}

// playbackInfoResponse is the Jellyfin-compatible PlaybackInfo response.
type playbackInfoResponse struct {
	MediaSources  []struct{}  `json:"MediaSources"`
	PlaySessionId string      `json:"PlaySessionId"`
	ErrorCode     interface{} `json:"ErrorCode"`
}

// GetPlaybackInfo returns a stub PlaybackInfo with an empty MediaSources list.
func (h *Handler) GetPlaybackInfo(w http.ResponseWriter, r *http.Request) {
	response.WriteJSON(w, http.StatusOK, playbackInfoResponse{
		MediaSources:  []struct{}{},
		PlaySessionId: uuid.New().String(),
		ErrorCode:     nil,
	})
}
