// Package handlers provides HTTP handlers for metadata service
package handlers

import (
	"net/http"
)

// RegisterRoutes registers all metadata routes
func RegisterRoutes(mux *http.ServeMux,
	refresh *ItemRefreshHandler,
	update *ItemUpdateHandler,
	lookup *ItemLookupHandler,
	remoteImage *ItemRemoteImageHandler,
	tasks *ItemRefreshTaskHandler) {

	// Item refresh routes
	mux.HandleFunc("POST /Items/{id}/Refresh", refresh.RefreshItem)
	mux.HandleFunc("DELETE /Items/{id}/Refresh", refresh.CancelRefresh)
	mux.HandleFunc("POST /Library/Refresh", refresh.RefreshLibrary)

	// Item update routes
	mux.HandleFunc("POST /Items/{id}/UpdateMetadata", update.UpdateItem)

	// Item lookup routes
	mux.HandleFunc("GET /Items/{id}/Lookup", lookup.LookupItem)

	// Remote image routes
	mux.HandleFunc("GET /Items/{id}/RemoteImage", remoteImage.SearchImages)

	// Scheduled task routes
	mux.HandleFunc("GET /ScheduledTasks", tasks.ListTasks)
	mux.HandleFunc("POST /ScheduledTasks/{id}/Start", tasks.StartTask)
}
