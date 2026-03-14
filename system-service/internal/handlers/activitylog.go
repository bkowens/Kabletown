package handlers

import (
	"net/http"

	"github.com/jellyfinhanced/shared/response"
)

// GetActivityLog handles GET /System/ActivityLog/Entries.
func (h *Handler) GetActivityLog(w http.ResponseWriter, r *http.Request) {
	response.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"Items":            []interface{}{},
		"TotalRecordCount": 0,
		"StartIndex":       0,
	})
}
