package handlers

import (
	"net/http"

	"github.com/jellyfinhanced/shared/response"
)

// notificationListResult is the response for listing notifications.
type notificationListResult struct {
	Notifications    []struct{} `json:"Notifications"`
	TotalRecordCount int        `json:"TotalRecordCount"`
}

// unreadCountResult is the response for the unread count endpoint.
type unreadCountResult struct {
	UnreadCount int `json:"UnreadCount"`
}

// ListNotifications returns an empty notifications list for the given user.
func (h *Handler) ListNotifications(w http.ResponseWriter, r *http.Request) {
	response.WriteJSON(w, http.StatusOK, notificationListResult{
		Notifications:    []struct{}{},
		TotalRecordCount: 0,
	})
}

// GetUnreadCount returns zero unread notifications for the given user.
func (h *Handler) GetUnreadCount(w http.ResponseWriter, r *http.Request) {
	response.WriteJSON(w, http.StatusOK, unreadCountResult{UnreadCount: 0})
}

// MarkRead marks all notifications as read and returns 204.
func (h *Handler) MarkRead(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

// GetNotificationTypes returns the list of supported notification types — stub returns empty array.
func (h *Handler) GetNotificationTypes(w http.ResponseWriter, r *http.Request) {
	response.WriteJSON(w, http.StatusOK, []struct{}{})
}

// GetNotificationServices returns the list of notification services — stub returns empty array.
func (h *Handler) GetNotificationServices(w http.ResponseWriter, r *http.Request) {
	response.WriteJSON(w, http.StatusOK, []struct{}{})
}

// WebSocket returns 400 because WebSocket is not supported in this stub.
func (h *Handler) WebSocket(w http.ResponseWriter, r *http.Request) {
	response.WriteError(w, http.StatusBadRequest, "WebSocket not supported")
}

// StartSessionEvents writes SSE headers and immediately closes — stub implementation.
func (h *Handler) StartSessionEvents(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
}
