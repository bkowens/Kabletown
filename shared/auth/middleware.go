package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"
)

// contextKey is an unexported type for context keys to avoid collisions.
type contextKey string

// Context key constants for storing auth values in request context.
const (
	UserIDKey   contextKey = "userID"
	DeviceIDKey contextKey = "deviceID"
	TokenKey    contextKey = "token"
	IsAdminKey  contextKey = "isAdmin"
)

// DeviceLookupFunc resolves a bearer token to a user identity.
type DeviceLookupFunc func(token string) (userID string, isAdmin bool, err error)

// ParseMediaBrowserHeader parses a Jellyfin/Emby MediaBrowser authorization header value.
// The value is a comma-separated list of key="value" pairs, optionally prefixed with "MediaBrowser ".
// Returns an error if the token field is absent.
func ParseMediaBrowserHeader(header string) (token, deviceId, client, device, version string, err error) {
	var tok, did, cli, dev, ver string

	parts := strings.Split(header, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		// Strip the leading "MediaBrowser " scheme prefix if present on the first segment.
		if strings.HasPrefix(part, "MediaBrowser ") {
			part = strings.TrimSpace(strings.TrimPrefix(part, "MediaBrowser "))
		}
		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			continue
		}
		key := strings.TrimSpace(kv[0])
		val := strings.TrimSpace(kv[1])
		// Strip surrounding double-quotes from the value.
		if len(val) >= 2 && val[0] == '"' && val[len(val)-1] == '"' {
			val = val[1 : len(val)-1]
		}
		switch key {
		case "Token":
			tok = val
		case "DeviceId":
			did = val
		case "Client":
			cli = val
		case "Device":
			dev = val
		case "Version":
			ver = val
		}
	}

	if tok == "" {
		return "", "", "", "", "", errors.New("auth: token not found in MediaBrowser header")
	}
	return tok, did, cli, dev, ver, nil
}

// ExtractToken extracts a bearer token from the request using Jellyfin auth conventions.
// Priority order: X-Emby-Authorization header → Authorization header → api_key query param.
func ExtractToken(r *http.Request) (token string, ok bool) {
	// 1. X-Emby-Authorization header
	if h := r.Header.Get("X-Emby-Authorization"); h != "" {
		if tok, _, _, _, _, parseErr := ParseMediaBrowserHeader(h); parseErr == nil && tok != "" {
			return tok, true
		}
	}

	// 2. Authorization header (MediaBrowser scheme)
	if h := r.Header.Get("Authorization"); h != "" && strings.HasPrefix(h, "MediaBrowser ") {
		if tok, _, _, _, _, parseErr := ParseMediaBrowserHeader(h); parseErr == nil && tok != "" {
			return tok, true
		}
		// Bare "MediaBrowser Token=<value>" without surrounding quotes
		rest := strings.TrimSpace(strings.TrimPrefix(h, "MediaBrowser "))
		if strings.HasPrefix(rest, "Token=") {
			tok := strings.Trim(strings.TrimPrefix(rest, "Token="), "\"")
			if tok != "" {
				return tok, true
			}
		}
	}

	// 3. api_key query parameter
	if key := r.URL.Query().Get("api_key"); key != "" {
		return key, true
	}

	return "", false
}

// NewAuthMiddleware returns HTTP middleware that authenticates every request.
// Requests whose path exactly matches an entry in anonymousPaths bypass auth.
func NewAuthMiddleware(lookup DeviceLookupFunc, anonymousPaths []string) func(http.Handler) http.Handler {
	anonSet := make(map[string]struct{}, len(anonymousPaths))
	for _, p := range anonymousPaths {
		anonSet[p] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Anonymous paths bypass authentication.
			if _, isAnon := anonSet[r.URL.Path]; isAnon {
				next.ServeHTTP(w, r)
				return
			}

			// Extract bearer token from the request.
			token, ok := ExtractToken(r)
			if !ok {
				w.Header().Set("Content-Type", "application/json; charset=utf-8")
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"Message":"Unauthorized","StatusCode":401}`)) //nolint:errcheck
				return
			}

			// Resolve token → userID + isAdmin.
			userID, isAdmin, lookupErr := lookup(token)
			if lookupErr != nil || userID == "" {
				w.Header().Set("Content-Type", "application/json; charset=utf-8")
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"Message":"Unauthorized","StatusCode":401}`)) //nolint:errcheck
				return
			}

			// Propagate identity through context.
			ctx := r.Context()
			ctx = context.WithValue(ctx, UserIDKey, userID)
			ctx = context.WithValue(ctx, IsAdminKey, isAdmin)
			ctx = context.WithValue(ctx, TokenKey, token)

			// Extract DeviceId from the authorization header and store in context.
			if h := r.Header.Get("X-Emby-Authorization"); h != "" {
				if _, deviceID, _, _, _, parseErr := ParseMediaBrowserHeader(h); parseErr == nil && deviceID != "" {
					ctx = context.WithValue(ctx, DeviceIDKey, deviceID)
				}
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserID returns the authenticated user ID stored in the request context.
func GetUserID(r *http.Request) string {
	v, _ := r.Context().Value(UserIDKey).(string)
	return v
}

// GetDeviceID returns the device ID stored in the request context.
func GetDeviceID(r *http.Request) string {
	v, _ := r.Context().Value(DeviceIDKey).(string)
	return v
}

// GetToken returns the bearer token stored in the request context.
func GetToken(r *http.Request) string {
	v, _ := r.Context().Value(TokenKey).(string)
	return v
}

// IsAdmin returns true if the authenticated user has administrator privileges.
func IsAdmin(r *http.Request) bool {
	v, _ := r.Context().Value(IsAdminKey).(bool)
	return v
}

// RequireAdmin is an HTTP middleware that returns 403 Forbidden when the caller is not an admin.
func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !IsAdmin(r) {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte(`{"Message":"Forbidden","StatusCode":403}`)) //nolint:errcheck
			return
		}
		next.ServeHTTP(w, r)
	})
}

// RequireAuth is an HTTP middleware that returns 401 Unauthorized when no user ID is in context.
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if GetUserID(r) == "" {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"Message":"Unauthorized","StatusCode":401}`)) //nolint:errcheck
			return
		}
		next.ServeHTTP(w, r)
	})
}
