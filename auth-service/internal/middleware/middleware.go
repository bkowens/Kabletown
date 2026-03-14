package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"os"
	"path/filepath"

	chi_middleware "github.com/go-chi/chi/v5/middleware"
)

// serverID is initialized once at startup
var serverID string

// InitializeServerID generates or loads a unique server ID
func InitializeServerID(dataDir string) string {
	idFile := filepath.Join(dataDir, "server.id")
	
	// Try to load existing ID
	if data, err := os.ReadFile(idFile); err == nil {
		serverID = string(data)
		return serverID
	}
	
	// Generate new ID
	bytes := make([]byte, 16)
	rand.Read(bytes)
	serverID = hex.EncodeToString(bytes)
	
	// Save to file
	os.MkdirAll(dataDir, 0755)
	os.WriteFile(idFile, []byte(serverID), 0644)
	
	return serverID
}

// GetServerID returns the current server ID
func GetServerID() string {
	return serverID
}

// ResponseHeadersMiddleware sets standard response headers (W5, W8 compliance)
func ResponseHeadersMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Application-Version", "10.9.11")
			w.Header().Set("X-MediaBrowser-Server-Id", serverID)
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			next.ServeHTTP(w, r)
		})
	}
}

// RequestID - alias for chi middleware
var RequestID = chi_middleware.RequestID

// RealIP - alias for chi middleware
var RealIP = chi_middleware.RealIP

// Logger - alias for chi middleware  
var Logger = chi_middleware.Logger

// Recoverer - alias for chi middleware
var Recoverer = chi_middleware.Recoverer
