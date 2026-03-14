// Package auth provides HTTP authentication middleware
package auth

import (
	"context"
	"database/sql"
	"net/http"
	"net/url"
	"strings"

	"github.com/google/uuid"
	"github.com/jellyfinhanced/shared/response"
	"github.com/jellyfinhanced/shared/types"
	"github.com/jmoiron/sqlx"
)

// AuthMiddleware wraps an http.Handler with authentication
// Supports: X-Emby-Authorization header, api_key query param
func AuthMiddleware(pool *sqlx.DB, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if endpoint is public/allow anonymous
		if isPublicEndpoint(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		// Extract and validate auth from multiple sources
		ctx, err := AuthenticateRequest(r, pool)
		if err != nil {
			// If no auth provided (not invalid), allow through for AllowAnonymous endpoints
			if _, ok := err.(*NoAuthProvidedError); ok {
				next.ServeHTTP(w, r)
				return
			}
			response.WriteUnauthorized(w, "Unauthorized: "+err.Error())
			return
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// NoAuthProvidedError indicates no authentication was provided in the request
type NoAuthProvidedError struct {
	Message string
}

func (e *NoAuthProvidedError) Error() string {
	return e.Message
}

// AuthenticateRequest validates authentication from multiple sources
// Returns NoAuthProvidedError if no auth headers/params found
func AuthenticateRequest(r *http.Request, pool *sqlx.DB) (context.Context, error) {
	ctx := r.Context()

	// 1. Try X-Emby-Authorization header
	authHeader := r.Header.Get("X-Emby-Authorization")
	if authHeader != "" {
		return AuthenticateFromHeader(ctx, authHeader, pool)
	}

	// 2. Try api_key query parameter
	apiKey := r.URL.Query().Get("api_key")
	if apiKey != "" {
		return AuthenticateFromAPIKeyQuery(ctx, apiKey, pool)
	}

	// 3. No authentication provided - caller decides if allowed
	return nil, &NoAuthProvidedError{"No authentication provided"}
}

// AuthenticateFromHeader validates X-Emby-Authorization header
func AuthenticateFromHeader(ctx context.Context, authHeader string, pool *sqlx.DB) (context.Context, error) {
	// Parse the MediaBrowser header
	header, err := ParseMediaBrowserHeader(authHeader)
	if err != nil {
		return nil, &AuthError{"Invalid authorization header: " + err.Error()}
	}

	// Extract token
	token := header.Token
	if token == "" {
		return nil, &AuthError{"Token missing in authorization header"}
	}

	// Validate token against database
	info, err := ValidateToken(ctx, token, pool)
	if err != nil {
		return nil, err
	}

	// Populate context with auth info
	ctx = SetAuthInContext(ctx, info)

	return ctx, nil
}

// AuthenticateFromAPIKeyQuery validates api_key from query parameters
func AuthenticateFromAPIKeyQuery(ctx context.Context, apiKey string, pool *sqlx.DB) (context.Context, error) {
	// Validate API key format
	if !types.ValidateTokenFormat(apiKey) {
		return nil, &AuthError{"Invalid API key format"}
	}

	// Validate token against database
	info, err := ValidateToken(ctx, apiKey, pool)
	if err != nil {
		return nil, err
	}

	// Populate context with auth info
	ctx = SetAuthInContext(ctx, info)

	return ctx, nil
}

// isPublicEndpoint checks if the path should skip authentication
func isPublicEndpoint(path string) bool {
	// Decode URL path to handle encoded characters
	decodedPath, err := url.PathUnescape(path)
	if err != nil {
		decodedPath = path
	}

	// Public endpoints that don't require authentication
	publicPaths := []string{
		"/health",
		"/healthcheck",
		"/System/PublicStartupInfo",
		"/Sessions",
		"/Hls",
		"/Stream",
		"/Audio",
		"/Items/\u003cid>/Primary",
		"/Items/\u003cid>/Backdrop",
		"/Items/\u003cid>/Logo",
		"/Items/\u003cid>/Art",
		"/Items/\u003cid>/Thumb",
		"/Videos/\u003cid>/Stream",
		"/Streams/Audio",
		"/Streams/Video",
		"/Items/\u003cid>/SpecialFeatures",
		"/Items/\u003cid>/Latest",
		"/Users/Public",
	}

	for _, pp := range publicPaths {
		// Handle wildcard paths like /Items/{id}/Primary
		if strings.Contains(pp, "\u003cid\u003e") {
			// Check if path matches pattern (e.g., /Items/123/Primary)
			parts := strings.Split(pp, "/")
			requestParts := strings.Split(decodedPath, "/")
			
			if len(parts) == len(requestParts) {
				match := true
				for i := 0; i < len(parts); i++ {
					if parts[i] != "\u003cid\u003e" && parts[i] != requestParts[i] {
						match = false
						break
					}
				}
				if match {
					return true
				}
			}
		} else {
			if decodedPath == pp || (len(decodedPath) > len(pp) && decodedPath[:len(pp)+1] == pp+"/") {
				return true
			}
		}
	}

	return false
}

// AuthError represents an authentication error
type AuthError struct {
	Message string
}

func (e *AuthError) Error() string {
	return e.Message
}

// ValidateToken validates an API token and returns associated user info
func ValidateToken(ctx context.Context, token string, pool *sqlx.DB) (*AuthInfo, error) {
	// Hash the token for DB lookup
	tokenHash := types.HashToken(token)

	// Query for the API key
	sqlQuery := `SELECT id, user_id, username, device_id, client, device, version, is_admin
		FROM api_keys
		WHERE token = ? AND (expires_at IS NULL OR expires_at > NOW())
		LIMIT 1`

	var (
		id       uuid.UUID
		userID   uuid.UUID
		username string
		deviceID uuid.UUID
		client   string
		device   string
		version  string
		isAdmin  bool
	)

	err := pool.QueryRowxContext(ctx, sqlQuery, tokenHash).Scan(
		&id, &userID, &username, &deviceID, &client, &device, &version, &isAdmin,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &AuthError{"Token validation failed"}
		}
		return nil, &AuthError{"Token validation error"}
	}

	return &AuthInfo{
		UserID:   userID,
		Username: username,
		DeviceID: deviceID,
		Token:    token,
		IsAdmin:  isAdmin,
		IsApiKey: true,
		Client:   client,
		Device:   device,
		Version:  version,
	}, nil
}

// ValidateSession validates a session token (alternative to API keys)
func ValidateSession(ctx context.Context, sessionId string, pool *sqlx.DB) (*AuthInfo, error) {
	// Query for the session
	sqlQuery := `SELECT id, user_id, username, device_id, client, device, version, is_admin
		FROM sessions
		WHERE id = ? AND expiry > NOW()
		LIMIT 1`

	var (
		userID   uuid.UUID
		username string
		deviceID uuid.UUID
		client   string
		device   string
		version  string
		isAdmin  bool
	)

	err := pool.QueryRowxContext(ctx, sqlQuery, sessionId).Scan(
		&userID, &username, &deviceID, &client, &device, &version, &isAdmin,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &AuthError{"Session validation failed"}
		}
		return nil, &AuthError{"Session validation error"}
	}

	return &AuthInfo{
		UserID:   userID,
		Username: username,
		DeviceID: deviceID,
		Token:    sessionId,
		IsAdmin:  isAdmin,
		IsApiKey: false,
		Client:   client,
		Device:   device,
		Version:  version,
	}, nil
}
