package handlers

import (
	"net/http"

	"github.com/jellyfinhanced/shared/response"
)

// ListImages returns an empty list of images for the given item.
func (h *Handler) ListImages(w http.ResponseWriter, r *http.Request) {
	response.WriteJSON(w, http.StatusOK, []struct{}{})
}

// GetImage returns a 404 since no image data is stored in this stub.
func (h *Handler) GetImage(w http.ResponseWriter, r *http.Request) {
	response.WriteError(w, http.StatusNotFound, "Image not found")
}

// GetImageByIndex returns a 404 since no image data is stored in this stub.
func (h *Handler) GetImageByIndex(w http.ResponseWriter, r *http.Request) {
	response.WriteError(w, http.StatusNotFound, "Image not found")
}

// UploadImage accepts an image upload and returns 204 No Content.
func (h *Handler) UploadImage(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

// DeleteImage deletes an image and returns 204 No Content.
func (h *Handler) DeleteImage(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

// remoteImagesResult is the stub response for remote image search.
type remoteImagesResult struct {
	Images           []struct{} `json:"Images"`
	TotalRecordCount int        `json:"TotalRecordCount"`
}

// GetRemoteImages returns a stub empty remote images response.
func (h *Handler) GetRemoteImages(w http.ResponseWriter, r *http.Request) {
	response.WriteJSON(w, http.StatusOK, remoteImagesResult{
		Images:           []struct{}{},
		TotalRecordCount: 0,
	})
}

// DownloadRemoteImage queues a remote image download and returns 204.
func (h *Handler) DownloadRemoteImage(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}
