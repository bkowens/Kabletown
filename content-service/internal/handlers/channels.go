package handlers

import (
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jellyfinhanced/shared/logger"
)

var channelsLog = logger.NewLogger("channels-handler")

type ChannelsHandler struct {
	db  *sql.DB
}

func NewChannelsHandler(dbPool *sql.DB) *ChannelsHandler {
	return &ChannelsHandler{db: dbPool}
}

func (h *ChannelsHandler) GetChannels(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	render.JSON(w, r, map[string]interface{}{"Items": []interface{}{}})
}

func (h *ChannelsHandler) GetChannel(w http.ResponseWriter, r *http.Request) {
	_ = chi.URLParam(r, "itemId")
	w.WriteHeader(http.StatusOK)
	render.JSON(w, r, map[string]interface{}{})
}
