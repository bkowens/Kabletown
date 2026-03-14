package testutil

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/jellyfinhanced/shared/auth"
)

// DefaultTestUserID is a stable UUID used across tests.
const DefaultTestUserID = "11111111-1111-1111-1111-111111111111"

// DefaultTestAdminID is a stable admin UUID used across tests.
const DefaultTestAdminID = "00000000-0000-0000-0000-000000000001"

// DefaultServerID is a stable server UUID used across tests.
const DefaultServerID = "test-server-00000000-0000-0000-0000-000000000000"

// InjectAuth adds authentication context to an HTTP request.
// This simulates having passed through auth middleware with a valid token.
func InjectAuth(r *http.Request, userID string, isAdmin bool) *http.Request {
	uid := uuid.MustParse(userID)
	info := &auth.AuthInfo{
		UserID:   uid,
		Username: "testuser",
		IsAdmin:  isAdmin,
		Token:    "test-token-64chars-aabbccddee00112233445566778899aabbccddee0011223344",
		Client:   "TestClient",
		Device:   "TestDevice",
		Version:  "1.0.0",
	}
	return r.WithContext(auth.SetAuthInContext(r.Context(), info))
}

// InjectAdminAuth adds admin authentication context to an HTTP request.
func InjectAdminAuth(r *http.Request) *http.Request {
	return InjectAuth(r, DefaultTestAdminID, true)
}

// InjectUserAuth adds non-admin authentication context to an HTTP request.
func InjectUserAuth(r *http.Request) *http.Request {
	return InjectAuth(r, DefaultTestUserID, false)
}

// AuthContext creates a context with authentication info, useful for testing
// functions that operate on context directly rather than http.Request.
func AuthContext(userID string, isAdmin bool) context.Context {
	uid := uuid.MustParse(userID)
	info := &auth.AuthInfo{
		UserID:   uid,
		Username: "testuser",
		IsAdmin:  isAdmin,
	}
	return auth.SetAuthInContext(context.Background(), info)
}

// BuildMediaBrowserHeader constructs a valid X-Emby-Authorization header string.
// Jellyfin compat: the header format is MediaBrowser Client="...", Device="...", etc.
func BuildMediaBrowserHeader(token, client, device, deviceID, version string) string {
	return `MediaBrowser Client="` + client + `", Device="` + device + `", DeviceId="` + deviceID + `", Version="` + version + `", Token="` + token + `"`
}
