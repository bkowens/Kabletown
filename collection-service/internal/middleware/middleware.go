package middleware

import (
	"net/http"
	"time"
)

// JELLYFIN_VERSION is the compatibility version matching Jellyfin 10.9.11
const JELLYFIN_VERSION = "10.9.11"

// ServerID is the unique server identifier (set at startup)
var ServerID string

// SetServerID sets the server ID
func SetServerID(id string) {
	ServerID = id
}

// InitializeServerID initializes the server ID from system.xml or generates one
func InitializeServerID(dataDir string) string {
	// Simplified - in production, read actual system.xml
	if ServerID == "" {
		ServerID = "go-collection-server-00000000-0000-0000-0000-000000000000"
	}
	return ServerID
}

// ResponseHeadersMiddleware sets standard Jellyfin-compatible response headers
func ResponseHeadersMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Application-Version", JELLYFIN_VERSION)
			w.Header().Set("Date", time.Now().UTC().Format(http.TimeFormat))
			w.Header().Set("Content-Type", "application/json; charset=utf-8")

			if ServerID != "" {
				w.Header().Set("X-MediaBrowser-Server-Id", ServerID)
			}

			if r.Method == http.MethodGet {
				w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
				w.Header().Set("Pragma", "no-cache")
				w.Header().Set("Expires", "0")
			}

			next.ServeHTTP(w, r.WithContext(r.Context()))
		})
	}
}
