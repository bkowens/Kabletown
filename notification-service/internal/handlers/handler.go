// Package handlers implements HTTP route handlers for notification-service.
package handlers

import (
	"github.com/go-chi/chi/v5"
)

// Handler holds dependencies for notification-service route handlers.
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

// RegisterRoutes wires all notification-service routes onto the given chi router.
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get("/Notifications/{userId}", h.ListNotifications)
	r.Get("/Notifications/{userId}/Unread", h.GetUnreadCount)
	r.Post("/Notifications/{userId}/Read", h.MarkRead)
	r.Get("/Notifications/Types", h.GetNotificationTypes)
	r.Get("/Notifications/Services", h.GetNotificationServices)

	r.Get("/socket", h.WebSocket)
	r.Get("/Events/Sessions/Start", h.StartSessionEvents)
}
