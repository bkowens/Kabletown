package handlers

import (
	"net/http"

	"github.com/jellyfinhanced/shared/response"
)

// GetConfigurationPage handles GET /web/ConfigurationPage (dashboard stub).
func (h *Handler) GetConfigurationPage(w http.ResponseWriter, r *http.Request) {
	response.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"Name":        "Dashboard",
		"EnableInMainMenu": false,
		"MenuSection": "server",
		"MenuIcon":    "dashboard",
		"DisplayName": "Dashboard",
	})
}
