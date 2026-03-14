// Package auth provides additional context helper functions
package auth

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/jellyfinhanced/shared/types"
	"github.com/jmoiron/sqlx"
)

// SetUserInContextFromGUID creates a new context with user ID from string GUID
func SetUserInContextFromGUID(ctx context.Context, userID string) context.Context {
	if userID == "" {
		return ctx
	}
	id, err := types.ParseGUID(userID)
	if err != nil {
		return ctx
	}

	// Get existing auth or create new
	if auth, ok := GetAuth(ctx); ok {
		newAuth := *auth
		newAuth.UserID = id
		return SetAuthInContext(ctx, &newAuth)
	}

	return SetAuthInContext(ctx, &AuthInfo{
		UserID: id,
	})
}

// SetDeviceInContextFromGUID creates a new context with device ID from string GUID
func SetDeviceInContextFromGUID(ctx context.Context, deviceID string) context.Context {
	if deviceID == "" {
		return ctx
	}
	id, err := types.ParseGUID(deviceID)
	if err != nil {
		return ctx
	}

	// Get existing auth or create new
	if auth, ok := GetAuth(ctx); ok {
		newAuth := *auth
		newAuth.DeviceID = id
		return SetAuthInContext(ctx, &newAuth)
	}

	return SetAuthInContext(ctx, &AuthInfo{
		DeviceID: id,
	})
}

// SetUsernameInContext creates a new context with username
func SetUsernameInContext(ctx context.Context, username string) context.Context {
	if auth, ok := GetAuth(ctx); ok {
		newAuth := *auth
		newAuth.Username = username
		return SetAuthInContext(ctx, &newAuth)
	}

	return SetAuthInContext(ctx, &AuthInfo{
		Username: username,
	})
}

// SetAdminInContext creates a new context with admin flag
func SetAdminInContext(ctx context.Context, isAdmin bool) context.Context {
	if auth, ok := GetAuth(ctx); ok {
		newAuth := *auth
		newAuth.IsAdmin = isAdmin
		return SetAuthInContext(ctx, &newAuth)
	}

	return SetAuthInContext(ctx, &AuthInfo{
		IsAdmin: isAdmin,
	})
}

// SetIsApiKeyInContext creates a new context with IsApiKey flag
func SetIsApiKeyInContext(ctx context.Context, isApiKey bool) context.Context {
	if auth, ok := GetAuth(ctx); ok {
		newAuth := *auth
		newAuth.IsApiKey = isApiKey
		return SetAuthInContext(ctx, &newAuth)
	}

	return SetAuthInContext(ctx, &AuthInfo{
		IsApiKey: isApiKey,
	})
}

// GetUserIDAsGUID returns user ID as string (for JSON serialization)
func GetUserIDAsGUID(ctx context.Context) string {
	if auth, ok := GetAuth(ctx); ok {
		if auth.UserID == uuid.Nil {
			return ""
		}
		return auth.UserID.String()
	}
	return ""
}

// GetDeviceIDAsGUID returns device ID as string (for JSON serialization)
func GetDeviceIDAsGUID(ctx context.Context) string {
	if auth, ok := GetAuth(ctx); ok {
		if auth.DeviceID == uuid.Nil {
			return ""
		}
		return auth.DeviceID.String()
	}
	return ""
}

// GetUserFromString parses a string GUID to uuid.UUID
func GetUserFromString(userIdStr string) uuid.UUID {
	if userIdStr == "" {
		return uuid.Nil
	}
	id, _ := types.ParseGUID(userIdStr)
	return id
}

// GetDeviceFromString parses a string GUID to uuid.UUID
func GetDeviceFromString(deviceIdStr string) uuid.UUID {
	if deviceIdStr == "" {
		return uuid.Nil
	}
	id, _ := types.ParseGUID(deviceIdStr)
	return id
}

// ===== Compatibility wrappers for deprecated auth APIs ===
// These functions are kept for backward compatibility

// NewAuthMiddleware creates auth middleware - DEPRECATED
// Use AuthMiddleware(pool *sqlx.DB, next http.Handler) instead
func NewAuthMiddleware(pool *sqlx.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return AuthMiddleware(pool, next)
	}
}

// DeviceLookupFunc returns a function that looks up devices by user ID and device ID
// DEPRECATED - Use auth.GetAuth(context) and check DeviceID field
func DeviceLookupFunc(userID string) func(*http.Request) string {
	return func(r *http.Request) string {
		if info, ok := GetAuth(r.Context()); ok && info != nil {
			return info.DeviceID.String()
		}
		return ""
	}
}
