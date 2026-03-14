package handlers

import (
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jellyfinhanced/shared/logger"
)

var remoteImageLog = logger.NewLogger("remote-image-handler")

type RemoteImageHandler struct {
	db  *sql.DB
}

func NewRemoteImageHandler(dbPool *sql.DB) *RemoteImageHandler {
	return &RemoteImageHandler{db: dbPool}
}

// GetRemoteImages returns available remote images for an item
func (h *RemoteImageHandler) GetRemoteImages(w http.ResponseWriter, r *http.Request) {
	_ = chi.URLParam(r, "itemId")
	
	w.WriteHeader(http.StatusOK)
	render.JSON(w, r, map[string]interface{}{
		"images": []interface{}{},
	})
}

// GetImageProviders returns available image providers
func (h *RemoteImageHandler) GetImageProviders(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	render.JSON(w, r, map[string]interface{}{
		"providers": []string{"tmdb", "tvdb", "fanart"},
	})
}

// DownloadRemoteImage downloads an image from a remote source
func (h *RemoteImageHandler) DownloadRemoteImage(w http.ResponseWriter, r *http.Request) {
	_ = chi.URLParam(r, "itemId")
	
	w.WriteHeader(http.StatusOK)
}

// UploadRemoteImage uploads an image
func (h *RemoteImageHandler) UploadRemoteImage(w http.ResponseWriter, r *http.Request) {
	_ = chi.URLParam(r, "imageType")
	
	w.WriteHeader(http.StatusOK)
}

// DeleteRemoteImage deletes an image
func (h *RemoteImageHandler) DeleteRemoteImage(w http.ResponseWriter, r *http.Request) {
	_ = chi.URLParam(r, "imageType")
	
	w.WriteHeader(http.StatusOK)
}
