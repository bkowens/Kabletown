package handlers

import (
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/jellyfinhanced/shared/response"
)

// LiveTvHandler handles LiveTV-related requests.
type LiveTvHandler struct {
	db *sql.DB
}

// NewLiveTvHandler creates a new LiveTvHandler.
func NewLiveTvHandler(dbPool *sql.DB) *LiveTvHandler {
	return &LiveTvHandler{db: dbPool}
}

// GetLiveTvInfo returns LiveTV service info (stub).
func (h *LiveTvHandler) GetLiveTvInfo(w http.ResponseWriter, r *http.Request) {
	response.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"IsEnabled": false,
	})
}

// GetPrograms returns a list of programs.
func (h *LiveTvHandler) GetPrograms(w http.ResponseWriter, r *http.Request) {
	response.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"Items":            []interface{}{},
		"TotalRecordCount": 0,
		"StartIndex":       0,
	})
}

// GetLiveChannels returns a list of live TV channels.
func (h *LiveTvHandler) GetLiveChannels(w http.ResponseWriter, r *http.Request) {
	response.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"Items":            []interface{}{},
		"TotalRecordCount": 0,
		"StartIndex":       0,
	})
}

// GetChannelPrograms returns programs for a specific channel.
func (h *LiveTvHandler) GetChannelPrograms(w http.ResponseWriter, r *http.Request) {
	_ = chi.URLParam(r, "channelId")
	response.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"Items":            []interface{}{},
		"TotalRecordCount": 0,
		"StartIndex":       0,
	})
}

// GetProgram returns a specific program.
func (h *LiveTvHandler) GetProgram(w http.ResponseWriter, r *http.Request) {
	_ = chi.URLParam(r, "programId")
	response.WriteJSON(w, http.StatusOK, map[string]interface{}{})
}
