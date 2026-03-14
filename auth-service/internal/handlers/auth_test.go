package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/jellyfinhanced/auth-service/internal/db"
	"github.com/jellyfinhanced/auth-service/internal/dto"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
)

// Ensure context import is used.
var _ = context.Background

// newTestHandler creates a Handler backed by go-sqlmock for unit testing.
func newTestHandler(t *testing.T) (*Handler, sqlmock.Sqlmock) {
	t.Helper()
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
	h := New(sqlxDB, "test-server-id")
	return h, mock
}

// buildChiContext sets chi URL params on a request.
func buildChiContext(r *http.Request, params map[string]string) *http.Request {
	rctx := chi.NewRouteContext()
	for k, v := range params {
		rctx.URLParams.Add(k, v)
	}
	return r.WithContext(
		r.Context(),
	)
}

// TestAuthenticateByName_ValidCredentials verifies successful login returns AccessToken and User.
// Jellyfin compat: POST /Users/AuthenticateByName must return AuthenticationResult with User, AccessToken, ServerId.
func TestAuthenticateByName_ValidCredentials(t *testing.T) {
	h, mock := newTestHandler(t)

	// hashPassword for "testpassword"
	pw := h.hashPassword("testpassword")

	// GetUserByName query
	userRows := sqlmock.NewRows([]string{"Id", "Name", "Password", "IsDisabled", "IsHidden", "PrimaryImageTag"}).
		AddRow("user-id-1", "admin", pw, false, false, "")
	mock.ExpectQuery("SELECT .+ FROM\\s+Users").WithArgs("admin").WillReturnRows(userRows)

	// isAdmin check
	permRows := sqlmock.NewRows([]string{"Value"}).AddRow(1)
	mock.ExpectQuery("SELECT Value FROM Permissions").WithArgs("user-id-1").WillReturnRows(permRows)

	// upsertDevice
	mock.ExpectQuery("INSERT INTO Devices").WillReturnRows(sqlmock.NewRows(nil))

	body := `{"Username":"admin","Password":"testpassword"}`
	req := httptest.NewRequest("POST", "/Users/AuthenticateByName", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.AuthenticateByName(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", w.Code, w.Body.String())
	}

	var result dto.AuthenticationResult
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if result.AccessToken == "" {
		t.Error("AccessToken should not be empty")
	}
	if result.ServerId != "test-server-id" {
		t.Errorf("ServerId = %q, want %q", result.ServerId, "test-server-id")
	}
	if result.User == nil {
		t.Fatal("User should not be nil")
	}
	if result.User.Name != "admin" {
		t.Errorf("User.Name = %q, want %q", result.User.Name, "admin")
	}
	if !result.User.IsAdministrator {
		t.Error("User.IsAdministrator should be true")
	}
}

// TestAuthenticateByName_WrongPassword verifies 401 for incorrect password.
// Jellyfin compat: wrong password must return 401 with "Invalid username or password".
func TestAuthenticateByName_WrongPassword(t *testing.T) {
	h, mock := newTestHandler(t)

	correctHash := h.hashPassword("correctpassword")
	userRows := sqlmock.NewRows([]string{"Id", "Name", "Password", "IsDisabled", "IsHidden", "PrimaryImageTag"}).
		AddRow("user-id-1", "admin", correctHash, false, false, "")
	mock.ExpectQuery("SELECT .+ FROM\\s+Users").WithArgs("admin").WillReturnRows(userRows)

	permRows := sqlmock.NewRows([]string{"Value"}).AddRow(1)
	mock.ExpectQuery("SELECT Value FROM Permissions").WithArgs("user-id-1").WillReturnRows(permRows)

	body := `{"Username":"admin","Password":"wrongpassword"}`
	req := httptest.NewRequest("POST", "/Users/AuthenticateByName", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.AuthenticateByName(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}

	var errBody map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &errBody)
	if msg, _ := errBody["message"].(string); msg != "Invalid username or password" {
		t.Errorf("message = %q, want %q", msg, "Invalid username or password")
	}
}

// TestAuthenticateByName_MissingUsername verifies 400 when Username is empty.
// Jellyfin compat: empty Username returns 400.
func TestAuthenticateByName_MissingUsername(t *testing.T) {
	h, _ := newTestHandler(t)

	body := `{"Username":"","Password":"password"}`
	req := httptest.NewRequest("POST", "/Users/AuthenticateByName", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.AuthenticateByName(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

// TestAuthenticateByName_MissingPassword verifies 400 when both Password and PasswordMd5 are empty.
// Jellyfin compat: must provide either Password or PasswordMd5.
func TestAuthenticateByName_MissingPassword(t *testing.T) {
	h, mock := newTestHandler(t)

	userRows := sqlmock.NewRows([]string{"Id", "Name", "Password", "IsDisabled", "IsHidden", "PrimaryImageTag"}).
		AddRow("user-id-1", "admin", "somehash", false, false, "")
	mock.ExpectQuery("SELECT .+ FROM\\s+Users").WithArgs("admin").WillReturnRows(userRows)

	permRows := sqlmock.NewRows([]string{"Value"}).AddRow(0)
	mock.ExpectQuery("SELECT Value FROM Permissions").WithArgs("user-id-1").WillReturnRows(permRows)

	body := `{"Username":"admin"}`
	req := httptest.NewRequest("POST", "/Users/AuthenticateByName", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.AuthenticateByName(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

// TestAuthenticateByName_DisabledUser verifies 401 for disabled user.
// Jellyfin compat: disabled users must get 401 "User is disabled".
func TestAuthenticateByName_DisabledUser(t *testing.T) {
	h, mock := newTestHandler(t)

	pw := h.hashPassword("pass")
	userRows := sqlmock.NewRows([]string{"Id", "Name", "Password", "IsDisabled", "IsHidden", "PrimaryImageTag"}).
		AddRow("user-id-1", "disabled", pw, true, false, "")
	mock.ExpectQuery("SELECT .+ FROM\\s+Users").WithArgs("disabled").WillReturnRows(userRows)

	permRows := sqlmock.NewRows([]string{"Value"}).AddRow(0)
	mock.ExpectQuery("SELECT Value FROM Permissions").WithArgs("user-id-1").WillReturnRows(permRows)

	body := `{"Username":"disabled","Password":"pass"}`
	req := httptest.NewRequest("POST", "/Users/AuthenticateByName", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.AuthenticateByName(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
}

// TestAuthenticateByName_InvalidBody verifies 400 for malformed JSON.
// Jellyfin compat: malformed body returns 400.
func TestAuthenticateByName_InvalidBody(t *testing.T) {
	h, _ := newTestHandler(t)

	req := httptest.NewRequest("POST", "/Users/AuthenticateByName", bytes.NewBufferString("not json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.AuthenticateByName(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

// TestAuthenticateByName_UserNotFound verifies 401 when user does not exist.
// Jellyfin compat: unknown user must return same error as wrong password (401).
func TestAuthenticateByName_UserNotFound(t *testing.T) {
	h, mock := newTestHandler(t)

	mock.ExpectQuery("SELECT .+ FROM\\s+Users").WithArgs("nobody").
		WillReturnRows(sqlmock.NewRows([]string{"Id", "Name", "Password", "IsDisabled", "IsHidden", "PrimaryImageTag"}))

	body := `{"Username":"nobody","Password":"pass"}`
	req := httptest.NewRequest("POST", "/Users/AuthenticateByName", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.AuthenticateByName(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
}

// TestAuthenticateByName_ResponseShape verifies all required JSON fields.
// Jellyfin compat: response must contain User, AccessToken, ServerId at top level.
func TestAuthenticateByName_ResponseShape(t *testing.T) {
	h, mock := newTestHandler(t)

	pw := h.hashPassword("pass")
	userRows := sqlmock.NewRows([]string{"Id", "Name", "Password", "IsDisabled", "IsHidden", "PrimaryImageTag"}).
		AddRow("user-id-1", "admin", pw, false, false, "")
	mock.ExpectQuery("SELECT .+ FROM\\s+Users").WithArgs("admin").WillReturnRows(userRows)

	permRows := sqlmock.NewRows([]string{"Value"}).AddRow(1)
	mock.ExpectQuery("SELECT Value FROM Permissions").WithArgs("user-id-1").WillReturnRows(permRows)

	mock.ExpectQuery("INSERT INTO Devices").WillReturnRows(sqlmock.NewRows(nil))

	body := `{"Username":"admin","Password":"pass"}`
	req := httptest.NewRequest("POST", "/Users/AuthenticateByName", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.AuthenticateByName(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", w.Code, w.Body.String())
	}

	var raw map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &raw)

	requiredFields := []string{"User", "AccessToken", "ServerId"}
	for _, f := range requiredFields {
		if _, ok := raw[f]; !ok {
			t.Errorf("response missing required field %q", f)
		}
	}

	// Verify User object shape
	user, ok := raw["User"].(map[string]interface{})
	if !ok {
		t.Fatal("User field is not an object")
	}
	userFields := []string{"Id", "Name", "HasPassword", "IsAdministrator", "IsDisabled", "IsHidden"}
	for _, f := range userFields {
		if _, ok := user[f]; !ok {
			t.Errorf("User missing required field %q", f)
		}
	}
}

// TestValidateUser_Found verifies GET /Users/{userId}/Authorize returns user info.
// Jellyfin compat: returns Id, Name, isAdmin, hasPassword.
func TestValidateUser_Found(t *testing.T) {
	h, mock := newTestHandler(t)

	userRows := sqlmock.NewRows([]string{"Id", "Name", "Password", "IsDisabled", "IsHidden", "PrimaryImageTag"}).
		AddRow("user-id-1", "admin", "hash", false, false, "")
	mock.ExpectQuery("SELECT .+ FROM\\s+Users").WithArgs("user-id-1").WillReturnRows(userRows)

	permRows := sqlmock.NewRows([]string{"Value"}).AddRow(1)
	mock.ExpectQuery("SELECT Value FROM Permissions").WithArgs("user-id-1").WillReturnRows(permRows)

	req := httptest.NewRequest("GET", "/Users/user-id-1/Authorize", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("userId", "user-id-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	h.ValidateUser(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}

	var body map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &body)
	if body["Id"] != "user-id-1" {
		t.Errorf("Id = %v, want user-id-1", body["Id"])
	}
}

// TestValidateUser_NotFound verifies 404 for unknown userId.
// Jellyfin compat: missing user returns 404.
func TestValidateUser_NotFound(t *testing.T) {
	h, mock := newTestHandler(t)

	mock.ExpectQuery("SELECT .+ FROM\\s+Users").WithArgs("no-such-user").
		WillReturnRows(sqlmock.NewRows([]string{"Id", "Name", "Password", "IsDisabled", "IsHidden", "PrimaryImageTag"}))

	req := httptest.NewRequest("GET", "/Users/no-such-user/Authorize", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("userId", "no-such-user")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	h.ValidateUser(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", w.Code)
	}
}

// TestValidateUser_EmptyUserId verifies 400 when userId param is empty.
// Jellyfin compat: empty userId returns 400.
func TestValidateUser_EmptyUserId(t *testing.T) {
	h, _ := newTestHandler(t)

	req := httptest.NewRequest("GET", "/Users//Authorize", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("userId", "")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	h.ValidateUser(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

// TestHashPassword_Deterministic verifies same input produces same hash.
// Jellyfin compat: SHA256 hash must be deterministic for token comparison.
func TestHashPassword_Deterministic(t *testing.T) {
	h, _ := newTestHandler(t)
	h1 := h.hashPassword("hello")
	h2 := h.hashPassword("hello")
	if h1 != h2 {
		t.Error("hashPassword not deterministic")
	}
	if len(h1) != 64 {
		t.Errorf("hash length = %d, want 64", len(h1))
	}
}

// TestHashPassword_DifferentInputs verifies different inputs produce different hashes.
func TestHashPassword_DifferentInputs(t *testing.T) {
	h, _ := newTestHandler(t)
	h1 := h.hashPassword("password1")
	h2 := h.hashPassword("password2")
	if h1 == h2 {
		t.Error("different passwords produced same hash")
	}
}

// TestUserToDto_Fields verifies all dto fields are correctly populated.
// Jellyfin compat: UserDto must include all required fields for web client.
func TestUserToDto_Fields(t *testing.T) {
	tests := []struct {
		name     string
		user     *db.User
		wantPW   bool
		wantAdmin bool
	}{
		{
			"admin with password",
			&db.User{Id: "id-1", Name: "admin", Password: "hash", IsAdmin: true, IsDisabled: false, IsHidden: false},
			true, true,
		},
		{
			"regular user no password",
			&db.User{Id: "id-2", Name: "user", Password: "", IsAdmin: false, IsDisabled: false, IsHidden: false},
			false, false,
		},
		{
			"disabled user",
			&db.User{Id: "id-3", Name: "disabled", Password: "hash", IsAdmin: false, IsDisabled: true, IsHidden: false},
			true, false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := userToDto(tt.user, "server-1")
			if d.Id != tt.user.Id {
				t.Errorf("Id = %q, want %q", d.Id, tt.user.Id)
			}
			if d.Name != tt.user.Name {
				t.Errorf("Name = %q, want %q", d.Name, tt.user.Name)
			}
			if d.ServerId != "server-1" {
				t.Errorf("ServerId = %q, want %q", d.ServerId, "server-1")
			}
			if d.HasPassword != tt.wantPW {
				t.Errorf("HasPassword = %v, want %v", d.HasPassword, tt.wantPW)
			}
			if d.HasConfiguredPassword != tt.wantPW {
				t.Errorf("HasConfiguredPassword = %v, want %v", d.HasConfiguredPassword, tt.wantPW)
			}
			if d.IsAdministrator != tt.wantAdmin {
				t.Errorf("IsAdministrator = %v, want %v", d.IsAdministrator, tt.wantAdmin)
			}
			if d.IsDisabled != tt.user.IsDisabled {
				t.Errorf("IsDisabled = %v, want %v", d.IsDisabled, tt.user.IsDisabled)
			}
		})
	}
}

// TestStartupConfiguration verifies GET /Startup/Configuration response shape.
// Jellyfin compat: must return UICulture, MetadataCountryCode, PreferredMetadataLanguage.
func TestStartupConfiguration(t *testing.T) {
	h, _ := newTestHandler(t)

	req := httptest.NewRequest("GET", "/Startup/Configuration", nil)
	w := httptest.NewRecorder()

	h.StartupConfiguration(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}

	var body map[string]string
	json.Unmarshal(w.Body.Bytes(), &body)

	required := []string{"UICulture", "MetadataCountryCode", "PreferredMetadataLanguage"}
	for _, f := range required {
		if _, ok := body[f]; !ok {
			t.Errorf("missing field %q", f)
		}
	}
}

// TestStartupComplete verifies POST /Startup/Complete returns 204.
// Jellyfin compat: startup wizard completion returns 204 No Content.
func TestStartupComplete(t *testing.T) {
	h, _ := newTestHandler(t)

	req := httptest.NewRequest("POST", "/Startup/Complete", nil)
	w := httptest.NewRecorder()

	h.StartupComplete(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("status = %d, want 204", w.Code)
	}
}

// TestGenerateToken_Format verifies generated tokens meet format requirements.
// Jellyfin compat: tokens must be 64 hex characters (32 bytes).
func TestGenerateToken_Format(t *testing.T) {
	token, err := generateToken()
	if err != nil {
		t.Fatalf("generateToken() error: %v", err)
	}
	if len(token) != 64 {
		t.Errorf("token length = %d, want 64", len(token))
	}
	for _, c := range token {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			t.Errorf("token contains non-hex char %q", string(c))
		}
	}
}

// TestGenerateToken_Unique verifies no collisions.
func TestGenerateToken_Unique(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 100; i++ {
		token, _ := generateToken()
		if seen[token] {
			t.Fatalf("duplicate token generated at iteration %d", i)
		}
		seen[token] = true
	}
}
