package handlers

import (
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/jellyfinhanced/shared/response"
)

// ChannelsHandler handles channel-related requests.
type ChannelsHandler struct {
	db *sql.DB
}

// NewChannelsHandler creates a new ChannelsHandler.
func NewChannelsHandler(dbPool *sql.DB) *ChannelsHandler {
	return &ChannelsHandler{db: dbPool}
}

// GetChannels returns a list of channels (stub).
func (h *ChannelsHandler) GetChannels(w http.ResponseWriter, r *http.Request) {
	response.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"Items":            []interface{}{},
		"TotalRecordCount": 0,
		"StartIndex":       0,
	})
}

// GetChannel returns a specific channel.
func (h *ChannelsHandler) GetChannel(w http.ResponseWriter, r *http.Request) {
	_ = chi.URLParam(r, "itemId")
	response.WriteJSON(w, http.StatusOK, map[string]interface{}{})
}
