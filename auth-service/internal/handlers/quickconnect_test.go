package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/jellyfinhanced/shared/auth"
)

// TestQuickConnectEnabled verifies GET /QuickConnect/Enabled returns true.
// Jellyfin compat: QuickConnect must report as enabled.
func TestQuickConnectEnabled(t *testing.T) {
	h, _ := newTestHandler(t)

	req := httptest.NewRequest("GET", "/QuickConnect/Enabled", nil)
	w := httptest.NewRecorder()

	h.QuickConnectEnabled(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}

	var result bool
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if !result {
		t.Error("QuickConnect should be enabled (true)")
	}
}

// TestQuickConnectInitiate_ResponseShape verifies POST /QuickConnect/Initiate returns Secret, Code, DeviceId.
// Jellyfin compat: initiate response must include Secret, Code, DeviceId fields.
func TestQuickConnectInitiate_ResponseShape(t *testing.T) {
	h, _ := newTestHandler(t)

	req := httptest.NewRequest("POST", "/QuickConnect/Initiate", nil)
	w := httptest.NewRecorder()

	h.QuickConnectInitiate(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}

	var body map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	required := []string{"Secret", "Code", "DeviceId"}
	for _, f := range required {
		if body[f] == "" {
			t.Errorf("missing or empty field %q", f)
		}
	}

	// Code should be 6 digits
	code := body["Code"]
	if len(code) != 6 {
		t.Errorf("Code length = %d, want 6", len(code))
	}
}

// TestQuickConnectConnect_NotFound verifies 404 for unknown secret.
// Jellyfin compat: unknown secret returns 404.
func TestQuickConnectConnect_NotFound(t *testing.T) {
	h, _ := newTestHandler(t)

	req := httptest.NewRequest("POST", "/QuickConnect/Connect?secret=unknown-secret", nil)
	w := httptest.NewRecorder()

	h.QuickConnectConnect(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", w.Code)
	}
}

// TestQuickConnectConnect_MissingSecret verifies 400 when secret param is absent.
// Jellyfin compat: missing secret returns 400.
func TestQuickConnectConnect_MissingSecret(t *testing.T) {
	h, _ := newTestHandler(t)

	req := httptest.NewRequest("POST", "/QuickConnect/Connect", nil)
	w := httptest.NewRecorder()

	h.QuickConnectConnect(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

// TestQuickConnectFlow_InitiateConnectAuthorize tests the full QuickConnect lifecycle.
// Jellyfin compat: initiate -> check status (not authenticated) -> authorize -> check status (authenticated).
func TestQuickConnectFlow_InitiateConnectAuthorize(t *testing.T) {
	h, _ := newTestHandler(t)

	// Step 1: Initiate
	initReq := httptest.NewRequest("POST", "/QuickConnect/Initiate", nil)
	initW := httptest.NewRecorder()
	h.QuickConnectInitiate(initW, initReq)

	if initW.Code != http.StatusOK {
		t.Fatalf("Initiate status = %d, want 200", initW.Code)
	}

	var initBody map[string]string
	json.Unmarshal(initW.Body.Bytes(), &initBody)
	secret := initBody["Secret"]
	code := initBody["Code"]

	if secret == "" || code == "" {
		t.Fatal("Initiate must return non-empty Secret and Code")
	}

	// Step 2: Connect (should not be authenticated yet)
	connectReq := httptest.NewRequest("POST", "/QuickConnect/Connect?secret="+secret, nil)
	connectW := httptest.NewRecorder()
	h.QuickConnectConnect(connectW, connectReq)

	if connectW.Code != http.StatusOK {
		t.Fatalf("Connect status = %d, want 200", connectW.Code)
	}

	var connectBody map[string]interface{}
	json.Unmarshal(connectW.Body.Bytes(), &connectBody)
	if connectBody["Authenticated"] != false {
		t.Error("should not be authenticated before authorization")
	}

	// Step 3: Authorize (requires auth context with user ID)
	authReq := httptest.NewRequest("POST", "/QuickConnect/Authorize?code="+code, nil)
	userUUID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	authInfo := &auth.AuthInfo{
		UserID:   userUUID,
		Username: "testuser",
		IsAdmin:  true,
	}
	ctx := auth.SetAuthInContext(authReq.Context(), authInfo)
	authReq = authReq.WithContext(ctx)
	authW := httptest.NewRecorder()
	h.QuickConnectAuthorize(authW, authReq)

	if authW.Code != http.StatusNoContent {
		t.Errorf("Authorize status = %d, want 204; body = %s", authW.Code, authW.Body.String())
	}

	// Step 4: Connect again (should be authenticated now)
	connect2Req := httptest.NewRequest("POST", "/QuickConnect/Connect?secret="+secret, nil)
	connect2W := httptest.NewRecorder()
	h.QuickConnectConnect(connect2W, connect2Req)

	if connect2W.Code != http.StatusOK {
		t.Fatalf("Connect2 status = %d, want 200", connect2W.Code)
	}

	var connect2Body map[string]interface{}
	json.Unmarshal(connect2W.Body.Bytes(), &connect2Body)
	if connect2Body["Authenticated"] != true {
		t.Error("should be authenticated after authorization")
	}
}

// TestQuickConnectAuthorize_MissingCode verifies 400 when code param is absent.
// Jellyfin compat: missing code returns 400.
func TestQuickConnectAuthorize_MissingCode(t *testing.T) {
	h, _ := newTestHandler(t)

	req := httptest.NewRequest("POST", "/QuickConnect/Authorize", nil)
	userUUID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	authInfo := &auth.AuthInfo{UserID: userUUID, Username: "testuser", IsAdmin: true}
	ctx := auth.SetAuthInContext(req.Context(), authInfo)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.QuickConnectAuthorize(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

// TestQuickConnectAuthorize_Unauthorized verifies 401 when no auth context.
// Jellyfin compat: authorize requires authenticated user.
func TestQuickConnectAuthorize_Unauthorized(t *testing.T) {
	h, _ := newTestHandler(t)

	req := httptest.NewRequest("POST", "/QuickConnect/Authorize?code=123456", nil)
	w := httptest.NewRecorder()

	h.QuickConnectAuthorize(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
}

// TestQuickConnectAuthorize_InvalidCode verifies 404 for wrong code.
// Jellyfin compat: unknown code returns 404.
func TestQuickConnectAuthorize_InvalidCode(t *testing.T) {
	h, _ := newTestHandler(t)

	req := httptest.NewRequest("POST", "/QuickConnect/Authorize?code=999999", nil)
	userUUID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	authInfo := &auth.AuthInfo{UserID: userUUID, Username: "testuser", IsAdmin: true}
	ctx := auth.SetAuthInContext(req.Context(), authInfo)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.QuickConnectAuthorize(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", w.Code)
	}
}

// TestAuthenticateWithQuickConnect_MissingSecret verifies 400 for empty Secret.
// Jellyfin compat: Secret field is required.
func TestAuthenticateWithQuickConnect_MissingSecret(t *testing.T) {
	h, _ := newTestHandler(t)

	body := `{"Secret":""}`
	req := httptest.NewRequest("GET", "/Users/AuthenticateWithQuickConnect", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.AuthenticateWithQuickConnect(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

// TestAuthenticateWithQuickConnect_UnknownSecret verifies 401 for unknown secret.
// Jellyfin compat: unknown secret returns 401.
func TestAuthenticateWithQuickConnect_UnknownSecret(t *testing.T) {
	h, _ := newTestHandler(t)

	body := `{"Secret":"unknown-secret-id"}`
	req := httptest.NewRequest("GET", "/Users/AuthenticateWithQuickConnect", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.AuthenticateWithQuickConnect(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
}

// TestAuthenticateWithQuickConnect_NotYetAuthorized verifies 401 when QuickConnect is not yet authorized.
// Jellyfin compat: attempting to authenticate before authorization returns 401.
func TestAuthenticateWithQuickConnect_NotYetAuthorized(t *testing.T) {
	h, _ := newTestHandler(t)

	// First, initiate a QuickConnect session
	initReq := httptest.NewRequest("POST", "/QuickConnect/Initiate", nil)
	initW := httptest.NewRecorder()
	h.QuickConnectInitiate(initW, initReq)

	var initBody map[string]string
	json.Unmarshal(initW.Body.Bytes(), &initBody)
	secret := initBody["Secret"]

	// Try to authenticate without authorizing first
	body := `{"Secret":"` + secret + `"}`
	req := httptest.NewRequest("GET", "/Users/AuthenticateWithQuickConnect", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.AuthenticateWithQuickConnect(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
}

// TestAuthenticateWithQuickConnect_InvalidBody verifies 400 for malformed JSON.
// Jellyfin compat: malformed request body returns 400.
func TestAuthenticateWithQuickConnect_InvalidBody(t *testing.T) {
	h, _ := newTestHandler(t)

	req := httptest.NewRequest("GET", "/Users/AuthenticateWithQuickConnect", bytes.NewBufferString("not json"))
	w := httptest.NewRecorder()

	h.AuthenticateWithQuickConnect(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

// setAuthCtx is a test helper that injects auth info into a request's context.
func setAuthCtx(r *http.Request, userID uuid.UUID, username string, isAdmin bool) *http.Request {
	info := &auth.AuthInfo{
		UserID:   userID,
		Username: username,
		IsAdmin:  isAdmin,
	}
	return r.WithContext(auth.SetAuthInContext(r.Context(), info))
}

// context import needed by chi test helpers
var _ = context.Background
