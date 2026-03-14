package response

import (
	"crypto/rand"
	"fmt"
	"net/http"
	"sync"
)

// Server header constants
const (
	// XMediaBrowserServerID is the unique server identifier header
	XMediaBrowserServerID = "X-MediaBrowser-Server-Id"
	
	// XApplicationVersion is the application version header
	XApplicationVersion = "X-Application-Version"
	
	// XMediaBrowserToken is the token authentication header
	XMediaBrowserToken = "X-Emby-Authorization"
	
	// ContentTypeJSON is the JSON content type
	ContentTypeJSON = "application/json; charset=utf-8"
	
	// ContentTypeXML is the XML content type
	ContentTypeXML = "application/xml; charset=utf-8"
)

var (
	// ServerID is the unique server identifier (generated at startup)
	ServerID string
	
	// ApplicationVersion is the version of the application
	ApplicationVersion string = "1.0.0"
	
	// serverIDGenerator is used to generate unique server IDs
	serverIDOnce sync.Once
)

// SetServerHeaders sets common server response headers
// This should be called for all API responses
func SetServerHeaders(w http.ResponseWriter) {
	serverID := GetServerID()
	
	w.Header().Set(XMediaBrowserServerID, serverID)
	w.Header().Set(XApplicationVersion, ApplicationVersion)
}

// SetServerHeadersForRequest sets server headers on a request context
// Useful for middleware that needs to add headers before downstream processing
func SetServerHeadersForRequest(r *http.Request, w http.ResponseWriter) {
	SetServerHeaders(w)
}

// GetServerID returns the unique server identifier
// Generates a UUID-like ID on first call if not explicitly set
func GetServerID() string {
	if ServerID != "" {
		return ServerID
	}
	
	// Generate a simple UUID-like ID if not set
	serverIDOnce.Do(func() {
		if ServerID == "" {
			// Generate a UUID v4 format string
			ServerID = generateSimpleUUID()
		}
	})
	
	return ServerID
}

// SetServerID sets the server identifier explicitly
// Should be called during application startup
func SetServerID(id string) {
	ServerID = id
}

// SetApplicationVersion sets the application version
// Should be called during application startup
func SetApplicationVersion(version string) {
	ApplicationVersion = version
}

// AddCacheHeaders adds caching headers to prevent caching of streaming content
func AddCacheHeaders(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
}

// AddStreamingHeaders adds headers appropriate for HLS/streaming content
func AddStreamingHeaders(w http.ResponseWriter) {
	AddCacheHeaders(w)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
}

// AddJSONHeaders explicitly sets JSON content type headers
func AddJSONHeaders(w http.ResponseWriter) {
	w.Header().Set(ContentTypeHeader, ContentTypeJSON)
}

// ContentTypeHeader constant
const ContentTypeHeader = "Content-Type"

// addServerHeader is a helper to add server identification to any response
func addServerHeader(w http.ResponseWriter) {
	serverID := GetServerID()
	w.Header().Set(XMediaBrowserServerID, serverID)
	w.Header().Set(XApplicationVersion, ApplicationVersion)
}

// generateSimpleUUID creates a simple UUID v4 format string
// This is a simplified implementation for server identification
func generateSimpleUUID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		// Fallback to fixed value if crypto/rand fails
		return "kabletown-server-0000"
	}
	
	// Set version to 4 UUID
	b[6] = (b[6] & 0x0f) | 0x40
	
	// Set variant to RFC 4122
	b[8] = (b[8] & 0x3f) | 0x80
	
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

// CORSHeaders contains CORS-related header keys and values
type CORSHeaders struct {
	AllowOrigin      string
	AllowMethods     string
	AllowHeaders     string
	ExposeHeaders    string
	MaxAge           string
	AllowCredentials string
}

// DefaultCORSHeaders returns default CORS configuration
func DefaultCORSHeaders() CORSHeaders {
	return CORSHeaders{
		AllowOrigin:      "*",
		AllowMethods:     "GET, POST, PUT, DELETE, PATCH, OPTIONS",
		AllowHeaders:     "Content-Type, Authorization, X-Emby-Authorization",
		ExposeHeaders:    "X-Application-Version, X-MediaBrowser-Server-Id",
		MaxAge:           "86400", // 24 hours
		AllowCredentials: "false",
	}
}

// AddCORSHeaders adds CORS headers to the response
func AddCORSHeaders(w http.ResponseWriter, headers CORSHeaders) {
	w.Header().Set("Access-Control-Allow-Origin", headers.AllowOrigin)
	w.Header().Set("Access-Control-Allow-Methods", headers.AllowMethods)
	w.Header().Set("Access-Control-Allow-Headers", headers.AllowHeaders)
	w.Header().Set("Access-Control-Expose-Headers", headers.ExposeHeaders)
	w.Header().Set("Access-Control-Max-Age", headers.MaxAge)
	w.Header().Set("Access-Control-Allow-Credentials", headers.AllowCredentials)
}

// AddCORSHeadersWithOptions adds CORS headers with custom configuration
func AddCORSHeadersWithOptions(w http.ResponseWriter) {
	AddCORSHeaders(w, DefaultCORSHeaders())
}

// PreflightHandler handles CORS preflight OPTIONS requests
func PreflightHandler(w http.ResponseWriter, r *http.Request) {
	AddCORSHeadersWithOptions(w)
	w.WriteHeader(http.StatusNoContent)
}
