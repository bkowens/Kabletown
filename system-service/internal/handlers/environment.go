package handlers

import (
	"net/http"

	"github.com/jellyfinhanced/shared/response"
)

// GetDefaultDirectoryBrowser handles GET /Environment/DefaultDirectoryBrowser.
func (h *Handler) GetDefaultDirectoryBrowser(w http.ResponseWriter, r *http.Request) {
	response.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"Path": "/",
	})
}

// GetDrives handles GET /Environment/Drives.
func (h *Handler) GetDrives(w http.ResponseWriter, r *http.Request) {
	response.WriteJSON(w, http.StatusOK, []map[string]interface{}{
		{"Name": "/", "Path": "/", "Type": "Fixed"},
	})
}

// GetParentPath handles GET /Environment/ParentPath.
func (h *Handler) GetParentPath(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	if path == "" || path == "/" {
		response.WriteJSON(w, http.StatusOK, map[string]interface{}{"Path": "/"})
		return
	}
	// Simple parent path: strip last segment
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' && i > 0 {
			response.WriteJSON(w, http.StatusOK, map[string]interface{}{"Path": path[:i]})
			return
		}
	}
	response.WriteJSON(w, http.StatusOK, map[string]interface{}{"Path": "/"})
}

// GetDirectoryContents handles POST /Environment/DirectoryContents.
func (h *Handler) GetDirectoryContents(w http.ResponseWriter, r *http.Request) {
	response.WriteJSON(w, http.StatusOK, []interface{}{})
}

// GetNetworkShares handles GET /Environment/NetworkShares.
func (h *Handler) GetNetworkShares(w http.ResponseWriter, r *http.Request) {
	response.WriteJSON(w, http.StatusOK, []interface{}{})
}
