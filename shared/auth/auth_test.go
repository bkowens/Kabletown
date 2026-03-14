package auth

import (
	"context"
	"net/http"
	"testing"

	"github.com/google/uuid"
)

// TestParseMediaBrowserHeader_ValidComplete verifies full header parsing.
// Jellyfin compat: jellyfin-web sends X-Emby-Authorization with all these fields on every request.
func TestParseMediaBrowserHeader_ValidComplete(t *testing.T) {
	header := `MediaBrowser Token="tok123", DeviceId="dev456", Client="JellyfinWeb", Version="10.8.0", Device="Chrome", Platform="Linux"`
	fields, err := ParseMediaBrowserHeader(header)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fields.Token != "tok123" {
		t.Errorf("Token = %q, want %q", fields.Token, "tok123")
	}
	if fields.DeviceID != "dev456" {
		t.Errorf("DeviceID = %q, want %q", fields.DeviceID, "dev456")
	}
	if fields.Client != "JellyfinWeb" {
		t.Errorf("Client = %q, want %q", fields.Client, "JellyfinWeb")
	}
	if fields.Version != "10.8.0" {
		t.Errorf("Version = %q, want %q", fields.Version, "10.8.0")
	}
	if fields.Device != "Chrome" {
		t.Errorf("Device = %q, want %q", fields.Device, "Chrome")
	}
	if fields.Platform != "Linux" {
		t.Errorf("Platform = %q, want %q", fields.Platform, "Linux")
	}
}

// TestParseMediaBrowserHeader_MissingToken verifies that a header without Token returns an error.
// Jellyfin compat: the frontend always sends a Token; absence means the client is not authenticated.
func TestParseMediaBrowserHeader_MissingToken(t *testing.T) {
	header := `MediaBrowser Client="JellyfinWeb", DeviceId="abc"`
	_, err := ParseMediaBrowserHeader(header)
	if err == nil {
		t.Fatal("expected error for missing token, got nil")
	}
	pe, ok := err.(*ParseError)
	if !ok {
		t.Fatalf("expected *ParseError, got %T", err)
	}
	if pe.Message != "Token is required" {
		t.Errorf("error message = %q, want %q", pe.Message, "Token is required")
	}
}

// TestParseMediaBrowserHeader_EmptyHeader verifies that an empty string returns no error but empty fields.
// Jellyfin compat: some public endpoints may have no auth header at all.
func TestParseMediaBrowserHeader_EmptyHeader(t *testing.T) {
	fields, err := ParseMediaBrowserHeader("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fields.Token != "" {
		t.Errorf("expected empty Token, got %q", fields.Token)
	}
}

// TestParseMediaBrowserHeader_InvalidScheme verifies non-MediaBrowser schemes are rejected.
// Jellyfin compat: only MediaBrowser scheme is accepted by Jellyfin.
func TestParseMediaBrowserHeader_InvalidScheme(t *testing.T) {
	_, err := ParseMediaBrowserHeader(`Bearer Token="abc"`)
	if err == nil {
		t.Fatal("expected error for non-MediaBrowser scheme")
	}
}

// TestParseMediaBrowserHeader_CaseInsensitiveKeys verifies keys are matched case-insensitively.
// Jellyfin compat: different clients may send TOKEN vs Token vs token.
func TestParseMediaBrowserHeader_CaseInsensitiveKeys(t *testing.T) {
	header := `MediaBrowser TOKEN="upper", DEVICEID="dev", CLIENT="cli"`
	fields, err := ParseMediaBrowserHeader(header)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fields.Token != "upper" {
		t.Errorf("Token = %q, want %q", fields.Token, "upper")
	}
	if fields.DeviceID != "dev" {
		t.Errorf("DeviceID = %q, want %q", fields.DeviceID, "dev")
	}
	if fields.Client != "cli" {
		t.Errorf("Client = %q, want %q", fields.Client, "cli")
	}
}

// TestParseMediaBrowserHeader_UnquotedValues verifies values without quotes are accepted.
// Jellyfin compat: some mobile clients may omit quotes.
func TestParseMediaBrowserHeader_UnquotedValues(t *testing.T) {
	header := `MediaBrowser Token=unquoted`
	fields, err := ParseMediaBrowserHeader(header)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fields.Token != "unquoted" {
		t.Errorf("Token = %q, want %q", fields.Token, "unquoted")
	}
}

// TestGenerateToken_UniqueAndLength verifies tokens are 64 hex chars and unique.
// Jellyfin compat: access tokens must be unique across sessions.
func TestGenerateToken_UniqueAndLength(t *testing.T) {
	t1 := GenerateToken()
	t2 := GenerateToken()

	if len(t1) != 64 {
		t.Errorf("token length = %d, want 64", len(t1))
	}
	if t1 == t2 {
		t.Error("two consecutive tokens should not be identical")
	}
	if !ValidateTokenFormat(t1) {
		t.Error("generated token failed format validation")
	}
}

// TestHashToken_Deterministic verifies the same input produces the same hash.
// Jellyfin compat: token hashes stored in DB for lookup must be reproducible.
func TestHashToken_Deterministic(t *testing.T) {
	token := "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"
	h1 := HashToken(token)
	h2 := HashToken(token)
	if h1 != h2 {
		t.Errorf("HashToken not deterministic: %q != %q", h1, h2)
	}
	if len(h1) != 64 {
		t.Errorf("hash length = %d, want 64", len(h1))
	}
}

// TestValidateTokenFormat_Table verifies valid/invalid token formats.
// Jellyfin compat: API keys must be exactly 64 hex characters.
func TestValidateTokenFormat_Table(t *testing.T) {
	tests := []struct {
		name  string
		token string
		want  bool
	}{
		{"valid 64 hex", "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789", true},
		{"too short", "abcdef", false},
		{"too long", "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789aa", false},
		{"non-hex chars", "GGGGGG0123456789abcdef0123456789abcdef0123456789abcdef0123456789", false},
		{"empty", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidateTokenFormat(tt.token)
			if got != tt.want {
				t.Errorf("ValidateTokenFormat(%q) = %v, want %v", tt.token, got, tt.want)
			}
		})
	}
}

// TestSetAuthInContext_GetAuth verifies context round-trip.
// Jellyfin compat: auth info must survive through middleware chain.
func TestSetAuthInContext_GetAuth(t *testing.T) {
	info := &AuthInfo{
		UserID:   uuid.New(),
		Username: "testuser",
		DeviceID: uuid.New(),
		Token:    "testtoken",
		IsAdmin:  true,
		IsApiKey: false,
		Client:   "JellyfinWeb",
		Device:   "Chrome",
		Version:  "10.8.0",
	}
	ctx := SetAuthInContext(context.Background(), info)

	got, ok := GetAuth(ctx)
	if !ok {
		t.Fatal("GetAuth returned false")
	}
	if got.UserID != info.UserID {
		t.Errorf("UserID = %v, want %v", got.UserID, info.UserID)
	}
	if got.Username != info.Username {
		t.Errorf("Username = %q, want %q", got.Username, info.Username)
	}
	if got.IsAdmin != info.IsAdmin {
		t.Errorf("IsAdmin = %v, want %v", got.IsAdmin, info.IsAdmin)
	}
}

// TestSetAuthInContext_Nil verifies nil info returns original context.
// Jellyfin compat: middleware should handle missing auth gracefully.
func TestSetAuthInContext_Nil(t *testing.T) {
	ctx := SetAuthInContext(context.Background(), nil)
	_, ok := GetAuth(ctx)
	if ok {
		t.Error("GetAuth should return false when nil info is stored")
	}
}

// TestGetAuth_EmptyContext verifies empty context returns false.
// Jellyfin compat: unauthenticated requests have no auth context.
func TestGetAuth_EmptyContext(t *testing.T) {
	_, ok := GetAuth(context.Background())
	if ok {
		t.Error("GetAuth should return false for empty context")
	}
}

// TestGetUserIDAsGUID verifies GUID string formatting from context.
// Jellyfin compat: user IDs in JSON responses must be UUID strings.
func TestGetUserIDAsGUID(t *testing.T) {
	uid := uuid.New()
	ctx := SetAuthInContext(context.Background(), &AuthInfo{UserID: uid})
	got := GetUserIDAsGUID(ctx)
	if got != uid.String() {
		t.Errorf("GetUserIDAsGUID = %q, want %q", got, uid.String())
	}
}

// TestGetUserIDAsGUID_NilUser verifies Nil UUID returns empty string.
// Jellyfin compat: empty user IDs should not appear in responses.
func TestGetUserIDAsGUID_NilUser(t *testing.T) {
	ctx := SetAuthInContext(context.Background(), &AuthInfo{UserID: uuid.Nil})
	got := GetUserIDAsGUID(ctx)
	if got != "" {
		t.Errorf("GetUserIDAsGUID for Nil = %q, want empty", got)
	}
}

// TestGetUserIDAsGUID_NoAuth verifies no auth returns empty string.
// Jellyfin compat: unauthenticated requests should get empty user ID.
func TestGetUserIDAsGUID_NoAuth(t *testing.T) {
	got := GetUserIDAsGUID(context.Background())
	if got != "" {
		t.Errorf("GetUserIDAsGUID with no auth = %q, want empty", got)
	}
}

// TestGetDeviceIDAsGUID verifies device ID string formatting from context.
// Jellyfin compat: device IDs must be UUID strings in session data.
func TestGetDeviceIDAsGUID(t *testing.T) {
	did := uuid.New()
	ctx := SetAuthInContext(context.Background(), &AuthInfo{DeviceID: did})
	got := GetDeviceIDAsGUID(ctx)
	if got != did.String() {
		t.Errorf("GetDeviceIDAsGUID = %q, want %q", got, did.String())
	}
}

// TestIsAdminFromContext verifies admin flag extraction.
// Jellyfin compat: admin checks gate access to user management endpoints.
func TestIsAdminFromContext(t *testing.T) {
	t.Run("admin true", func(t *testing.T) {
		ctx := SetAuthInContext(context.Background(), &AuthInfo{IsAdmin: true})
		if !IsAdminFromContext(ctx) {
			t.Error("expected true")
		}
	})
	t.Run("admin false", func(t *testing.T) {
		ctx := SetAuthInContext(context.Background(), &AuthInfo{IsAdmin: false})
		if IsAdminFromContext(ctx) {
			t.Error("expected false")
		}
	})
	t.Run("no auth", func(t *testing.T) {
		if IsAdminFromContext(context.Background()) {
			t.Error("expected false with no auth")
		}
	})
}

// TestNewAuthMiddleware_CreatesMiddleware verifies the deprecated wrapper still works.
// Jellyfin compat: backward compatibility for services using the old API.
func TestNewAuthMiddleware_CreatesMiddleware(t *testing.T) {
	// Cannot test with real DB, but verify function returns a non-nil middleware.
	mw := NewAuthMiddleware(nil)
	if mw == nil {
		t.Error("NewAuthMiddleware returned nil")
	}
}

// TestRequireAdmin verifies combined auth + admin check.
// Jellyfin compat: many admin-only endpoints use this pattern.
func TestRequireAdmin(t *testing.T) {
	t.Run("admin", func(t *testing.T) {
		ctx := SetAuthInContext(context.Background(), &AuthInfo{IsAdmin: true, UserID: uuid.New()})
		info, ok := RequireAdmin(ctx)
		if !ok || info == nil {
			t.Error("expected admin auth")
		}
	})
	t.Run("non-admin", func(t *testing.T) {
		ctx := SetAuthInContext(context.Background(), &AuthInfo{IsAdmin: false})
		_, ok := RequireAdmin(ctx)
		if ok {
			t.Error("expected false for non-admin")
		}
	})
	t.Run("no auth", func(t *testing.T) {
		_, ok := RequireAdmin(context.Background())
		if ok {
			t.Error("expected false for no auth")
		}
	})
}

// TestContextHelpers_AllWrappers verifies all convenience context accessors.
// Jellyfin compat: these are used throughout all service handlers.
func TestContextHelpers_AllWrappers(t *testing.T) {
	uid := uuid.New()
	did := uuid.New()
	info := &AuthInfo{
		UserID:   uid,
		Username: "admin",
		DeviceID: did,
		Token:    "tok",
		IsAdmin:  true,
		IsApiKey: true,
		Client:   "web",
		Device:   "browser",
		Version:  "1.0",
	}
	ctx := SetAuthInContext(context.Background(), info)

	if GetUserFromContext(ctx) != uid {
		t.Error("GetUserFromContext mismatch")
	}
	if GetUsernameFromContext(ctx) != "admin" {
		t.Error("GetUsernameFromContext mismatch")
	}
	if GetDeviceIDFromContext(ctx) != did {
		t.Error("GetDeviceIDFromContext mismatch")
	}
	if GetTokenFromContext(ctx) != "tok" {
		t.Error("GetTokenFromContext mismatch")
	}
	if !IsApiKeyFromContext(ctx) {
		t.Error("IsApiKeyFromContext expected true")
	}
	if GetClientFromContext(ctx) != "web" {
		t.Error("GetClientFromContext mismatch")
	}
	if GetDeviceFromContext(ctx) != "browser" {
		t.Error("GetDeviceFromContext mismatch")
	}
	if GetVersionFromContext(ctx) != "1.0" {
		t.Error("GetVersionFromContext mismatch")
	}
}

// TestSetUserInContextFromGUID verifies GUID string parsing into context.
// Jellyfin compat: URL path params contain user IDs as GUID strings.
func TestSetUserInContextFromGUID(t *testing.T) {
	uid := uuid.New()
	ctx := SetUserInContextFromGUID(context.Background(), uid.String())
	got := GetUserIDAsGUID(ctx)
	if got != uid.String() {
		t.Errorf("got %q, want %q", got, uid.String())
	}
}

// TestSetUserInContextFromGUID_Empty verifies empty string is handled.
// Jellyfin compat: missing userId path params should not crash.
func TestSetUserInContextFromGUID_Empty(t *testing.T) {
	ctx := SetUserInContextFromGUID(context.Background(), "")
	got := GetUserIDAsGUID(ctx)
	if got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}

// TestSetAdminInContext verifies admin flag can be set on existing auth.
// Jellyfin compat: admin status is resolved from Permissions table after initial auth.
func TestSetAdminInContext(t *testing.T) {
	ctx := SetAuthInContext(context.Background(), &AuthInfo{UserID: uuid.New()})
	ctx = SetAdminInContext(ctx, true)
	if !IsAdminFromContext(ctx) {
		t.Error("expected admin after SetAdminInContext(true)")
	}
}

// TestHasAuth verifies the HasAuth convenience function.
// Jellyfin compat: used to check if a request is authenticated at all.
func TestHasAuth(t *testing.T) {
	if HasAuth(context.Background()) {
		t.Error("expected false for empty context")
	}
	ctx := SetAuthInContext(context.Background(), &AuthInfo{})
	if !HasAuth(ctx) {
		t.Error("expected true for context with auth")
	}
}

// TestRequireAuth verifies the RequireAuth convenience function.
// Jellyfin compat: shorthand for handlers that need auth but not the ok bool.
func TestRequireAuth(t *testing.T) {
	if RequireAuth(context.Background()) != nil {
		t.Error("expected nil for empty context")
	}
	ctx := SetAuthInContext(context.Background(), &AuthInfo{Username: "test"})
	info := RequireAuth(ctx)
	if info == nil || info.Username != "test" {
		t.Error("expected non-nil with username")
	}
}

// TestHasValidAuth_Request verifies full request-level auth checking.
// Jellyfin compat: verifies both X-Emby-Authorization and Authorization headers are checked.
func TestHasValidAuth_XEmbyHeader(t *testing.T) {
	r, _ := http.NewRequest("GET", "/test", nil)
	r.Header.Set("X-Emby-Authorization", `MediaBrowser Token="valid123"`)
	if !HasValidAuth(r) {
		t.Error("expected true for valid X-Emby-Authorization")
	}
}

// TestHasValidAuth_AuthorizationFallback verifies Authorization header fallback.
// Jellyfin compat: some clients use Authorization instead of X-Emby-Authorization.
func TestHasValidAuth_AuthorizationFallback(t *testing.T) {
	r, _ := http.NewRequest("GET", "/test", nil)
	r.Header.Set("Authorization", `MediaBrowser Token="fallback"`)
	if !HasValidAuth(r) {
		t.Error("expected true for valid Authorization header")
	}
}

// TestHasValidAuth_NoHeaders verifies missing auth returns false.
// Jellyfin compat: public endpoints should not crash on missing headers.
func TestHasValidAuth_NoHeaders(t *testing.T) {
	r, _ := http.NewRequest("GET", "/test", nil)
	if HasValidAuth(r) {
		t.Error("expected false for missing headers")
	}
}

// TestGetAuthHeader_PreferXEmby verifies X-Emby-Authorization takes precedence.
// Jellyfin compat: when both headers present, X-Emby-Authorization wins.
func TestGetAuthHeader_PreferXEmby(t *testing.T) {
	r, _ := http.NewRequest("GET", "/test", nil)
	r.Header.Set("X-Emby-Authorization", `MediaBrowser Token="emby"`)
	r.Header.Set("Authorization", `MediaBrowser Token="auth"`)
	h := GetAuthHeader(r)
	if h != `MediaBrowser Token="emby"` {
		t.Errorf("expected X-Emby-Authorization, got %q", h)
	}
}

// TestExtractFullHeader_Complete verifies map extraction from all fields.
// Jellyfin compat: used to serialize header data for session records.
func TestExtractFullHeader_Complete(t *testing.T) {
	h := &HeaderFields{
		Token:    "tok",
		DeviceID: "dev",
		Client:   "cli",
		Version:  "1.0",
		Device:   "browser",
		Platform: "linux",
	}
	m := ExtractFullHeader(h)
	expected := map[string]string{
		"token":    "tok",
		"deviceid": "dev",
		"client":   "cli",
		"version":  "1.0",
		"device":   "browser",
		"platform": "linux",
	}
	for k, v := range expected {
		if m[k] != v {
			t.Errorf("map[%q] = %q, want %q", k, m[k], v)
		}
	}
}

// TestExtractFullHeader_Nil verifies nil input returns empty map.
// Jellyfin compat: defensive coding for missing header data.
func TestExtractFullHeader_Nil(t *testing.T) {
	m := ExtractFullHeader(nil)
	if len(m) != 0 {
		t.Errorf("expected empty map, got %d entries", len(m))
	}
}
