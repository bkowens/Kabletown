package handlers

import (
	"net/http"

	"github.com/jellyfinhanced/shared/response"
)

// GetBrandingOptions handles GET /Branding/Configuration.
func (h *Handler) GetBrandingOptions(w http.ResponseWriter, r *http.Request) {
	response.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"LoginDisclaimer": "",
		"CustomCss":       "",
		"SplashscreenEnabled": false,
	})
}

// UpdateBrandingOptions handles POST /Branding/Configuration.
func (h *Handler) UpdateBrandingOptions(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

// GetBrandingCss handles GET /Branding/Css.
func (h *Handler) GetBrandingCss(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/css; charset=utf-8")
	w.WriteHeader(http.StatusOK)
}
