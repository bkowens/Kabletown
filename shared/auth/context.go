// Package auth provides authentication utilities for Jellyfin/Emby API clients
package auth

import (
	"context"

	"github.com/google/uuid"
)

// authInfoContextKey is an unexported type for context keys to avoid collisions
type authInfoContextKey struct{}

// AuthInfo contains all authentication/authorization data for a request
type AuthInfo struct {
	UserID   uuid.UUID
	Username string
	DeviceID uuid.UUID
	Token    string
	IsAdmin  bool
	IsApiKey bool
	Client   string
	Device   string
	Version  string
}

// SetAuthInContext creates a new context with complete auth info stored atomically
func SetAuthInContext(ctx context.Context, info *AuthInfo) context.Context {
	if info == nil {
		return ctx
	}
	return context.WithValue(ctx, authInfoContextKey{}, info)
}

// GetAuth retrieves complete AuthInfo from context
// Returns (*AuthInfo, true) if authenticated, (nil, false) otherwise
func GetAuth(ctx context.Context) (*AuthInfo, bool) {
	v, ok := ctx.Value(authInfoContextKey{}).(*AuthInfo)
	if !ok {
		return nil, false
	}
	return v, true
}

// RequireAuth is the same as GetAuth but doesn't return the ok bool
// Use GetAuth(ctx) instead for safer error handling
func RequireAuth(ctx context.Context) *AuthInfo {
	if auth, ok := GetAuth(ctx); ok {
		return auth
	}
	return nil
}

// HasAuth checks if user is authenticated
func HasAuth(ctx context.Context) bool {
	_, ok := GetAuth(ctx)
	return ok
}

// GetUserFromContext retrieves user ID from context (convenience wrapper)
func GetUserFromContext(ctx context.Context) uuid.UUID {
	if auth, ok := GetAuth(ctx); ok {
		return auth.UserID
	}
	return uuid.Nil
}

// GetUsernameFromContext retrieves username from context (convenience wrapper)
func GetUsernameFromContext(ctx context.Context) string {
	if auth, ok := GetAuth(ctx); ok {
		return auth.Username
	}
	return ""
}

// GetDeviceIDFromContext retrieves device ID from context (convenience wrapper)
func GetDeviceIDFromContext(ctx context.Context) uuid.UUID {
	if auth, ok := GetAuth(ctx); ok {
		return auth.DeviceID
	}
	return uuid.Nil
}

// GetTokenFromContext retrieves token from context (convenience wrapper)
func GetTokenFromContext(ctx context.Context) string {
	if auth, ok := GetAuth(ctx); ok {
		return auth.Token
	}
	return ""
}

// IsAdminFromContext retrieves admin flag from context (convenience wrapper)
func IsAdminFromContext(ctx context.Context) bool {
	if auth, ok := GetAuth(ctx); ok {
		return auth.IsAdmin
	}
	return false
}

// IsApiKeyFromContext retrieves IsApiKey flag from context (convenience wrapper)
func IsApiKeyFromContext(ctx context.Context) bool {
	if auth, ok := GetAuth(ctx); ok {
		return auth.IsApiKey
	}
	return false
}

// GetClientFromContext retrieves client string from context (convenience wrapper)
func GetClientFromContext(ctx context.Context) string {
	if auth, ok := GetAuth(ctx); ok {
		return auth.Client
	}
	return ""
}

// GetDeviceFromContext retrieves device name from context (convenience wrapper)
func GetDeviceFromContext(ctx context.Context) string {
	if auth, ok := GetAuth(ctx); ok {
		return auth.Device
	}
	return ""
}

// GetVersionFromContext retrieves version string from context (convenience wrapper)
func GetVersionFromContext(ctx context.Context) string {
	if auth, ok := GetAuth(ctx); ok {
		return auth.Version
	}
	return ""
}

// RequireAdmin is a convenience function that checks both auth and admin status
func RequireAdmin(ctx context.Context) (*AuthInfo, bool) {
	if auth, ok := GetAuth(ctx); ok {
		if auth.IsAdmin {
			return auth, true
		}
	}
	return nil, false
}

// RequireApiKey is a convenience function that checks API key auth
func RequireApiKey(ctx context.Context) (*AuthInfo, bool) {
	if auth, ok := GetAuth(ctx); ok {
		if auth.IsApiKey {
			return auth, true
		}
	}
	return nil, false
}
