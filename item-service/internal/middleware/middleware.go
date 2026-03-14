package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var (
	serverID     string
	serverIDOnce sync.Once
)

// InitializeServerID returns a unique server ID for authentication
func InitializeServerID(dataDir string) string {
	serverIDOnce.Do(func() {
		// Try to read existing ID first
		serverIDFile := filepath.Join(dataDir, "server-id.txt")

		if content, err := os.ReadFile(serverIDFile); err == nil && len(content) < 100 {
			serverID = string(content)
		} else {
			serverID = generateServerID()
			os.WriteFile(serverIDFile, []byte(serverID), 0644)
		}
	})

	return serverID
}

// ResponseHeadersMiddleware adds required response headers per W1-W10
func ResponseHeadersMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// W1: Add content type headers
			w.Header().Set("Content-Type", "application/json; charset=utf-8")

			// Add server ID header
			w.Header().Set("X-MediaBrowser-Server-Id", GetServerID())

			// Continue to next handler
			next.ServeHTTP(w, r)
		})
	}
}

// AuthMiddleware adds authentication support by reading user info from context
// set by the shared auth middleware
func AuthMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// User info is already set by shared/auth middleware
			// Just ensure context values default to empty strings if not present
			if r.Context().Value("user_id") == nil {
				ctx := context.WithValue(r.Context(), "user_id", "")
				r = r.WithContext(ctx)
			}
			if r.Context().Value("device_id") == nil {
				ctx := context.WithValue(r.Context(), "device_id", "")
				r = r.WithContext(ctx)
			}

			next.ServeHTTP(w, r)
		})
	}
}

// GetServerID returns the current server ID
func GetServerID() string {
	return serverID
}

// generateServerID creates a unique server identifier
func generateServerID() string {
	// Generate UUID-like ID using crypto/rand
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return generateFallbackID()
	}
	return formatAsUUID(bytes)
}

// formatAsUUID formats bytes as a UUID string
func formatAsUUID(bytes []byte) string {
	hexStr := hex.EncodeToString(bytes)
	return hexStr[0:8] + "-" + hexStr[8:12] + "-" + hexStr[12:16] + "-" + hexStr[16:20] + "-" + hexStr[20:32]
}

// generateFallbackID creates ID from current time (for crypto failure)
func generateFallbackID() string {
	return time.Now().Format("20060102150405") + hex.EncodeToString([]byte("kabletown"))
}
