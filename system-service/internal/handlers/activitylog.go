package handlers

import (
	"net/http"

	"github.com/bowens/kabletown/shared/response"
)

// GetActivityLog handles GET /System/ActivityLog/Entries.
func (h *Handler) GetActivityLog(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, http.StatusOK, map[string]interface{}{
		"Items":            []interface{}{},
		"TotalRecordCount": 0,
		"StartIndex":       0,
	})
}
