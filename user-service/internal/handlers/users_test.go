package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/jellyfinhanced/shared/auth"
	"github.com/jellyfinhanced/user-service/internal/db"
	"github.com/jellyfinhanced/user-service/internal/dto"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

// suppress unused import warning for db
var _ = db.NewUserRepository

// newTestHandler creates a Handler backed by go-sqlmock.
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

// withAuth injects auth info into the request context.
func withAuth(r *http.Request, userID string, isAdmin bool) *http.Request {
	uid := uuid.MustParse(userID)
	info := &auth.AuthInfo{
		UserID:   uid,
		Username: "testuser",
		IsAdmin:  isAdmin,
	}
	return r.WithContext(auth.SetAuthInContext(r.Context(), info))
}

// withChi injects chi URL params into the request context.
func withChi(r *http.Request, params map[string]string) *http.Request {
	rctx := chi.NewRouteContext()
	for k, v := range params {
		rctx.URLParams.Add(k, v)
	}
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

const testUserID = "11111111-1111-1111-1111-111111111111"
const testUserID2 = "22222222-2222-2222-2222-222222222222"

// TestListUsers_Admin verifies admin gets all users.
// Jellyfin compat: admin GET /Users returns array of all UserDto objects.
func TestListUsers_Admin(t *testing.T) {
	h, mock := newTestHandler(t)

	rows := sqlmock.NewRows([]string{"Id", "Name", "Password", "IsDisabled", "IsHidden", "PrimaryImageTag", "Configuration", "Policy"}).
		AddRow(testUserID, "admin", "hash", false, false, "", "", "").
		AddRow(testUserID2, "user2", "", false, false, "", "", "")
	mock.ExpectQuery("SELECT .+ FROM Users").WillReturnRows(rows)

	// isAdmin calls for each user
	mock.ExpectQuery("SELECT Value FROM Permissions").WithArgs(testUserID).
		WillReturnRows(sqlmock.NewRows([]string{"Value"}).AddRow(1))
	mock.ExpectQuery("SELECT Value FROM Permissions").WithArgs(testUserID2).
		WillReturnRows(sqlmock.NewRows([]string{"Value"}).AddRow(0))

	req := httptest.NewRequest("GET", "/Users", nil)
	req = withAuth(req, testUserID, true)
	w := httptest.NewRecorder()

	h.ListUsers(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", w.Code, w.Body.String())
	}

	var users []dto.UserDto
	if err := json.Unmarshal(w.Body.Bytes(), &users); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if len(users) != 2 {
		t.Errorf("got %d users, want 2", len(users))
	}
}

// TestListUsers_NonAdmin_GetsSelfOnly verifies non-admin gets only their own user.
// Jellyfin compat: non-admin GET /Users returns single-element array with caller only.
func TestListUsers_NonAdmin_GetsSelfOnly(t *testing.T) {
	h, mock := newTestHandler(t)

	rows := sqlmock.NewRows([]string{"Id", "Name", "Password", "IsDisabled", "IsHidden", "PrimaryImageTag", "Configuration", "Policy"}).
		AddRow(testUserID, "regularuser", "", false, false, "", "", "")
	mock.ExpectQuery("SELECT .+ FROM Users WHERE Id").WithArgs(testUserID).WillReturnRows(rows)

	mock.ExpectQuery("SELECT Value FROM Permissions").WithArgs(testUserID).
		WillReturnRows(sqlmock.NewRows([]string{"Value"}).AddRow(0))

	req := httptest.NewRequest("GET", "/Users", nil)
	req = withAuth(req, testUserID, false)
	w := httptest.NewRecorder()

	h.ListUsers(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	var users []dto.UserDto
	json.Unmarshal(w.Body.Bytes(), &users)
	if len(users) != 1 {
		t.Errorf("got %d users, want 1", len(users))
	}
	if len(users) > 0 && users[0].Id != testUserID {
		t.Errorf("got Id %q, want %q", users[0].Id, testUserID)
	}
}

// TestGetUser_Self verifies a user can get their own profile.
// Jellyfin compat: GET /Users/{userId} with own ID includes Policy and Configuration.
func TestGetUser_Self(t *testing.T) {
	h, mock := newTestHandler(t)

	rows := sqlmock.NewRows([]string{"Id", "Name", "Password", "IsDisabled", "IsHidden", "PrimaryImageTag", "Configuration", "Policy"}).
		AddRow(testUserID, "me", "hash", false, false, "", "", "")
	mock.ExpectQuery("SELECT .+ FROM Users WHERE Id").WithArgs(testUserID).WillReturnRows(rows)

	mock.ExpectQuery("SELECT Value FROM Permissions").WithArgs(testUserID).
		WillReturnRows(sqlmock.NewRows([]string{"Value"}).AddRow(0))

	req := httptest.NewRequest("GET", "/Users/"+testUserID, nil)
	req = withAuth(req, testUserID, false)
	req = withChi(req, map[string]string{"userId": testUserID})
	w := httptest.NewRecorder()

	h.GetUser(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", w.Code, w.Body.String())
	}

	var user dto.UserDto
	json.Unmarshal(w.Body.Bytes(), &user)
	if user.Id != testUserID {
		t.Errorf("Id = %q, want %q", user.Id, testUserID)
	}
	// Should include default policy and config when user has none stored
	if user.Policy == nil {
		t.Error("Policy should not be nil (should have default)")
	}
	if user.Configuration == nil {
		t.Error("Configuration should not be nil (should have default)")
	}
}

// TestGetUser_MeAlias verifies "me" alias resolves to caller's user ID.
// Jellyfin compat: GET /Users/me must resolve to the authenticated user.
func TestGetUser_MeAlias(t *testing.T) {
	h, mock := newTestHandler(t)

	rows := sqlmock.NewRows([]string{"Id", "Name", "Password", "IsDisabled", "IsHidden", "PrimaryImageTag", "Configuration", "Policy"}).
		AddRow(testUserID, "me", "hash", false, false, "", "", "")
	mock.ExpectQuery("SELECT .+ FROM Users WHERE Id").WithArgs(testUserID).WillReturnRows(rows)

	mock.ExpectQuery("SELECT Value FROM Permissions").WithArgs(testUserID).
		WillReturnRows(sqlmock.NewRows([]string{"Value"}).AddRow(0))

	req := httptest.NewRequest("GET", "/Users/me", nil)
	req = withAuth(req, testUserID, false)
	req = withChi(req, map[string]string{"userId": "me"})
	w := httptest.NewRecorder()

	h.GetUser(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	var user dto.UserDto
	json.Unmarshal(w.Body.Bytes(), &user)
	if user.Id != testUserID {
		t.Errorf("Id = %q, want %q (me should resolve to caller)", user.Id, testUserID)
	}
}

// TestGetUser_OtherUser_NonAdmin_Forbidden verifies non-admin cannot access other users.
// Jellyfin compat: non-admin accessing another user returns 403.
func TestGetUser_OtherUser_NonAdmin_Forbidden(t *testing.T) {
	h, _ := newTestHandler(t)

	req := httptest.NewRequest("GET", "/Users/"+testUserID2, nil)
	req = withAuth(req, testUserID, false) // not admin, accessing testUserID2
	req = withChi(req, map[string]string{"userId": testUserID2})
	w := httptest.NewRecorder()

	h.GetUser(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("status = %d, want 403", w.Code)
	}
}

// TestCreateUser_Success verifies POST /Users/New creates a user.
// Jellyfin compat: returns the created UserDto with default Policy and Configuration.
func TestCreateUser_Success(t *testing.T) {
	h, mock := newTestHandler(t)

	// CreateUser does INSERT, then INSERT permissions, then GetUserByID
	mock.ExpectExec("INSERT INTO Users").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO Permissions").WillReturnResult(sqlmock.NewResult(1, 1))

	// GetUserByID after create
	rows := sqlmock.NewRows([]string{"Id", "Name", "Password", "IsDisabled", "IsHidden", "PrimaryImageTag", "Configuration", "Policy"}).
		AddRow("new-user-id", "newuser", "bcrypthash", false, false, "", "", "")
	mock.ExpectQuery("SELECT .+ FROM Users WHERE Id").WillReturnRows(rows)
	mock.ExpectQuery("SELECT Value FROM Permissions").WillReturnRows(sqlmock.NewRows([]string{"Value"}).AddRow(0))

	body := `{"Name":"newuser","Password":"pass123"}`
	req := httptest.NewRequest("POST", "/Users/New", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.CreateUser(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", w.Code, w.Body.String())
	}

	var user dto.UserDto
	json.Unmarshal(w.Body.Bytes(), &user)
	if user.Name != "newuser" {
		t.Errorf("Name = %q, want newuser", user.Name)
	}
	if user.Policy == nil {
		t.Error("created user should have default Policy")
	}
	if user.Configuration == nil {
		t.Error("created user should have default Configuration")
	}
}

// TestCreateUser_MissingName verifies 400 when Name is empty.
// Jellyfin compat: Name is required for user creation.
func TestCreateUser_MissingName(t *testing.T) {
	h, _ := newTestHandler(t)

	body := `{"Name":"","Password":"pass"}`
	req := httptest.NewRequest("POST", "/Users/New", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.CreateUser(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

// TestCreateUser_InvalidBody verifies 400 for malformed JSON.
// Jellyfin compat: malformed body returns 400.
func TestCreateUser_InvalidBody(t *testing.T) {
	h, _ := newTestHandler(t)

	req := httptest.NewRequest("POST", "/Users/New", bytes.NewBufferString("not json"))
	w := httptest.NewRecorder()

	h.CreateUser(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

// TestDeleteUser_Success verifies DELETE /Users/{userId} returns 204.
// Jellyfin compat: user deletion returns 204 No Content.
func TestDeleteUser_Success(t *testing.T) {
	h, mock := newTestHandler(t)

	// DeleteUser runs 4 DELETE statements
	mock.ExpectExec("DELETE FROM UserData").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("DELETE FROM DisplayPreferences").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("DELETE FROM Permissions").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("DELETE FROM Users").WillReturnResult(sqlmock.NewResult(0, 1))

	req := httptest.NewRequest("DELETE", "/Users/"+testUserID, nil)
	req = withChi(req, map[string]string{"userId": testUserID})
	w := httptest.NewRecorder()

	h.DeleteUser(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("status = %d, want 204", w.Code)
	}
}

// TestChangePassword_NonAdmin_WrongCurrentPassword verifies 401 for wrong current password.
// Jellyfin compat: non-admin must provide correct current password.
func TestChangePassword_NonAdmin_WrongCurrentPassword(t *testing.T) {
	h, mock := newTestHandler(t)

	// bcrypt hash of "correctpassword"
	hash, _ := bcryptHash("correctpassword")
	rows := sqlmock.NewRows([]string{"Id", "Name", "Password", "IsDisabled", "IsHidden", "PrimaryImageTag", "Configuration", "Policy"}).
		AddRow(testUserID, "user", hash, false, false, "", "", "")
	mock.ExpectQuery("SELECT .+ FROM Users WHERE Id").WithArgs(testUserID).WillReturnRows(rows)
	mock.ExpectQuery("SELECT Value FROM Permissions").WithArgs(testUserID).
		WillReturnRows(sqlmock.NewRows([]string{"Value"}).AddRow(0))

	body := `{"CurrentPw":"wrongpassword","NewPw":"newpass"}`
	req := httptest.NewRequest("POST", "/Users/"+testUserID+"/Password", bytes.NewBufferString(body))
	req = withAuth(req, testUserID, false)
	req = withChi(req, map[string]string{"userId": testUserID})
	w := httptest.NewRecorder()

	h.ChangePassword(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
}

// TestChangePassword_NonAdmin_Forbidden verifies non-admin cannot change other user's password.
// Jellyfin compat: non-admin changing another user's password returns 403.
func TestChangePassword_NonAdmin_Forbidden(t *testing.T) {
	h, _ := newTestHandler(t)

	body := `{"CurrentPw":"pass","NewPw":"newpass"}`
	req := httptest.NewRequest("POST", "/Users/"+testUserID2+"/Password", bytes.NewBufferString(body))
	req = withAuth(req, testUserID, false) // not admin, trying to change testUserID2's password
	req = withChi(req, map[string]string{"userId": testUserID2})
	w := httptest.NewRecorder()

	h.ChangePassword(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("status = %d, want 403", w.Code)
	}
}

// TestUpdatePolicy_Success verifies POST /Users/{userId}/Policy returns 204.
// Jellyfin compat: policy update returns 204 No Content.
func TestUpdatePolicy_Success(t *testing.T) {
	h, mock := newTestHandler(t)

	mock.ExpectExec("UPDATE Users SET").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO Permissions").WillReturnResult(sqlmock.NewResult(0, 1))

	body := `{"IsAdministrator":true,"EnableRemoteAccess":true}`
	req := httptest.NewRequest("POST", "/Users/"+testUserID+"/Policy", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req = withChi(req, map[string]string{"userId": testUserID})
	w := httptest.NewRecorder()

	h.UpdatePolicy(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("status = %d, want 204; body = %s", w.Code, w.Body.String())
	}
}

// TestUpdatePolicy_InvalidBody verifies 400 for malformed JSON.
// Jellyfin compat: malformed body returns 400.
func TestUpdatePolicy_InvalidBody(t *testing.T) {
	h, _ := newTestHandler(t)

	req := httptest.NewRequest("POST", "/Users/"+testUserID+"/Policy", bytes.NewBufferString("not json"))
	req = withChi(req, map[string]string{"userId": testUserID})
	w := httptest.NewRecorder()

	h.UpdatePolicy(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

// TestUpdateConfiguration_NonAdmin_Forbidden verifies non-admin cannot update other user's config.
// Jellyfin compat: non-admin modifying another user returns 403.
func TestUpdateConfiguration_NonAdmin_Forbidden(t *testing.T) {
	h, _ := newTestHandler(t)

	body := `{"PlayDefaultAudioTrack":true}`
	req := httptest.NewRequest("POST", "/Users/"+testUserID2+"/Configuration", bytes.NewBufferString(body))
	req = withAuth(req, testUserID, false) // not admin
	req = withChi(req, map[string]string{"userId": testUserID2})
	w := httptest.NewRecorder()

	h.UpdateConfiguration(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("status = %d, want 403", w.Code)
	}
}

// TestUserToDto_Fields verifies the dto conversion.
// Jellyfin compat: UserDto must include all required fields for the web client.
func TestUserToDto_Fields(t *testing.T) {
	tests := []struct {
		name      string
		user      *db.User
		wantPW    bool
		wantAdmin bool
	}{
		{
			"admin with password",
			&db.User{Id: "id-1", Name: "admin", Password: "hash", IsAdmin: true},
			true, true,
		},
		{
			"user without password",
			&db.User{Id: "id-2", Name: "user", Password: "", IsAdmin: false},
			false, false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := userToDto(tt.user, "srv-1")
			if d.HasPassword != tt.wantPW {
				t.Errorf("HasPassword = %v, want %v", d.HasPassword, tt.wantPW)
			}
			if d.IsAdministrator != tt.wantAdmin {
				t.Errorf("IsAdministrator = %v, want %v", d.IsAdministrator, tt.wantAdmin)
			}
			if d.ServerId != "srv-1" {
				t.Errorf("ServerId = %q, want srv-1", d.ServerId)
			}
		})
	}
}

// TestDefaultPolicy_AdminFlag verifies default policy sets IsAdministrator correctly.
// Jellyfin compat: default policy must set correct admin flag and enable standard permissions.
func TestDefaultPolicy_AdminFlag(t *testing.T) {
	tests := []struct {
		name    string
		isAdmin bool
	}{
		{"admin policy", true},
		{"non-admin policy", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pol := defaultPolicy(tt.isAdmin)
			if pol.IsAdministrator != tt.isAdmin {
				t.Errorf("IsAdministrator = %v, want %v", pol.IsAdministrator, tt.isAdmin)
			}
			if !pol.EnableMediaPlayback {
				t.Error("EnableMediaPlayback should be true by default")
			}
			if !pol.EnableAllFolders {
				t.Error("EnableAllFolders should be true by default")
			}
			if pol.AuthenticationProviderId == "" {
				t.Error("AuthenticationProviderId should not be empty")
			}
		})
	}
}

// TestDefaultConfig_Values verifies default user configuration values.
// Jellyfin compat: default config must enable standard playback options.
func TestDefaultConfig_Values(t *testing.T) {
	cfg := defaultConfig()
	if !cfg.PlayDefaultAudioTrack {
		t.Error("PlayDefaultAudioTrack should be true")
	}
	if cfg.SubtitleMode != "Default" {
		t.Errorf("SubtitleMode = %q, want Default", cfg.SubtitleMode)
	}
	if !cfg.EnableNextEpisodeAutoPlay {
		t.Error("EnableNextEpisodeAutoPlay should be true")
	}
}

// bcryptHash is a test helper to generate a bcrypt hash (cost 11 to match Jellyfin).
func bcryptHash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 11)
	return string(hash), err
}
