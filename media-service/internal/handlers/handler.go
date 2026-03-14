// Package handlers implements HTTP route handlers for media-service.
package handlers

import (
	"github.com/go-chi/chi/v5"
)

// Handler holds dependencies for media-service route handlers.
type Handler struct {
	serverID   string
	serverName string
}

// New creates a Handler with the given server identity.
func New(serverID, serverName string) *Handler {
	return &Handler{
		serverID:   serverID,
		serverName: serverName,
	}
}

// RegisterRoutes wires all media-service routes onto the given chi router.
func (h *Handler) RegisterRoutes(r chi.Router) {
	// Image routes
	r.Get("/Items/{itemId}/Images", h.ListImages)
	r.Get("/Items/{itemId}/Images/{imageType}", h.GetImage)
	r.Get("/Items/{itemId}/Images/{imageType}/{imageIndex}", h.GetImageByIndex)
	r.Post("/Items/{itemId}/Images/{imageType}", h.UploadImage)
	r.Delete("/Items/{itemId}/Images/{imageType}", h.DeleteImage)
	r.Get("/Items/{itemId}/RemoteImages", h.GetRemoteImages)
	r.Post("/Items/{itemId}/RemoteImages/Download", h.DownloadRemoteImage)

	// Subtitle routes
	r.Get("/Videos/{itemId}/Subtitles/{index}/Stream.{format}", h.GetSubtitleStream)
	r.Get("/Videos/{itemId}/Subtitles/{index}/Stream", h.GetSubtitleStreamNoFormat)

	// Playback
	r.Get("/Items/{itemId}/PlaybackInfo", h.GetPlaybackInfo)
	r.Post("/Items/{itemId}/PlaybackInfo", h.GetPlaybackInfo)
}
