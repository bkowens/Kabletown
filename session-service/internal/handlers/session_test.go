package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/jellyfinhanced/shared/auth"

	"github.com/DATA-DOG/go-sqlmock"
)

// newTestSessionHandler creates a SessionHandler backed by go-sqlmock.
func newTestSessionHandler(t *testing.T) (*SessionHandler, sqlmock.Sqlmock) {
	t.Helper()
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	h := NewSessionHandler(mockDB)
	return h, mock
}

const testUserUUID = "11111111-1111-1111-1111-111111111111"

// withSessionAuth injects auth info into the request context.
func withSessionAuth(r *http.Request, userID string) *http.Request {
	uid := uuid.MustParse(userID)
	info := &auth.AuthInfo{
		UserID:   uid,
		Username: "testuser",
		IsAdmin:  false,
	}
	return r.WithContext(auth.SetAuthInContext(r.Context(), info))
}

// withChiParam injects a chi URL parameter into the request context.
func withChiParam(r *http.Request, key, value string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, value)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

// TestGetSessions_Success verifies GET /Sessions returns array of sessions.
// Jellyfin compat: sessions endpoint returns array of Session objects.
func TestGetSessions_Success(t *testing.T) {
	h, mock := newTestSessionHandler(t)

	rows := sqlmock.NewRows([]string{"id", "user_id", "device_id", "app_name", "device_name", "client", "last_activity_date"}).
		AddRow("sess-1", testUserUUID, "dev-1", "Jellyfin Web", "Chrome", "web", time.Now())
	mock.ExpectQuery("SELECT .+ FROM sessions").WithArgs(testUserUUID).WillReturnRows(rows)

	req := httptest.NewRequest("GET", "/Sessions", nil)
	req = withSessionAuth(req, testUserUUID)
	w := httptest.NewRecorder()

	h.GetSessions(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", w.Code, w.Body.String())
	}

	var sessions []map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &sessions); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if len(sessions) != 1 {
		t.Errorf("got %d sessions, want 1", len(sessions))
	}
}

// TestGetSessions_EmptyResult verifies empty array when no sessions exist.
// Jellyfin compat: empty session list returns empty array, not null.
func TestGetSessions_EmptyResult(t *testing.T) {
	h, mock := newTestSessionHandler(t)

	rows := sqlmock.NewRows([]string{"id", "user_id", "device_id", "app_name", "device_name", "client", "last_activity_date"})
	mock.ExpectQuery("SELECT .+ FROM sessions").WithArgs(testUserUUID).WillReturnRows(rows)

	req := httptest.NewRequest("GET", "/Sessions", nil)
	req = withSessionAuth(req, testUserUUID)
	w := httptest.NewRecorder()

	h.GetSessions(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}

	var sessions []interface{}
	json.Unmarshal(w.Body.Bytes(), &sessions)
	if sessions == nil {
		t.Error("sessions should be empty array, not null")
	}
}

// TestCreateSession_Success verifies POST /Sessions creates a session.
// Jellyfin compat: session creation returns 201 with Id field.
func TestCreateSession_Success(t *testing.T) {
	h, mock := newTestSessionHandler(t)

	mock.ExpectExec("INSERT INTO sessions").WillReturnResult(sqlmock.NewResult(1, 1))

	body := `{"DeviceId":"dev-1","AppName":"Jellyfin Web","DeviceName":"Chrome","Client":"web"}`
	req := httptest.NewRequest("POST", "/Sessions", bytes.NewBufferString(body))
	req = withSessionAuth(req, testUserUUID)
	w := httptest.NewRecorder()

	h.CreateSession(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201; body = %s", w.Code, w.Body.String())
	}

	var result map[string]string
	json.Unmarshal(w.Body.Bytes(), &result)
	if result["Id"] == "" {
		t.Error("response should include Id field")
	}
}

// TestCreateSession_Unauthorized verifies 401 when no auth context.
// Jellyfin compat: session creation requires authentication.
func TestCreateSession_Unauthorized(t *testing.T) {
	h, _ := newTestSessionHandler(t)

	body := `{"DeviceId":"dev-1","AppName":"app","DeviceName":"dev","Client":"c"}`
	req := httptest.NewRequest("POST", "/Sessions", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.CreateSession(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
}

// TestCreateSession_InvalidBody verifies 400 for malformed JSON.
// Jellyfin compat: malformed body returns 400.
func TestCreateSession_InvalidBody(t *testing.T) {
	h, _ := newTestSessionHandler(t)

	req := httptest.NewRequest("POST", "/Sessions", bytes.NewBufferString("not json"))
	req = withSessionAuth(req, testUserUUID)
	w := httptest.NewRecorder()

	h.CreateSession(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

// TestGetSession_NotFound verifies 404 for nonexistent session.
// Jellyfin compat: unknown session returns 404.
func TestGetSession_NotFound(t *testing.T) {
	h, mock := newTestSessionHandler(t)

	mock.ExpectQuery("SELECT .+ FROM sessions WHERE id").WithArgs("no-such-session").
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "device_id", "app_name", "device_name", "client", "last_activity_date"}))

	req := httptest.NewRequest("GET", "/Sessions/no-such-session", nil)
	req = withChiParam(req, "sessionId", "no-such-session")
	w := httptest.NewRecorder()

	h.GetSession(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", w.Code)
	}
}

// TestReportSessionActivity_Success verifies session activity update returns 204.
// Jellyfin compat: activity report returns 204 No Content.
func TestReportSessionActivity_Success(t *testing.T) {
	h, mock := newTestSessionHandler(t)

	mock.ExpectExec("UPDATE sessions SET last_activity_date").WillReturnResult(sqlmock.NewResult(0, 1))

	req := httptest.NewRequest("POST", "/Sessions/sess-1/Activity", nil)
	req = withChiParam(req, "sessionId", "sess-1")
	w := httptest.NewRecorder()

	h.ReportSessionActivity(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("status = %d, want 204", w.Code)
	}
}

// TestReportPlaying_NoBody verifies playback report with empty body still returns 204.
// Jellyfin compat: playback reporting is best-effort, empty body is valid.
func TestReportPlaying_NoBody(t *testing.T) {
	h, mock := newTestSessionHandler(t)

	mock.ExpectExec("UPDATE sessions SET last_activity_date").WillReturnResult(sqlmock.NewResult(0, 1))

	req := httptest.NewRequest("POST", "/Sessions/sess-1/Playing", nil)
	req = withChiParam(req, "sessionId", "sess-1")
	w := httptest.NewRecorder()

	h.ReportPlaying(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("status = %d, want 204", w.Code)
	}
}

// TestReportPlaying_WithBody verifies playback report with state data.
// Jellyfin compat: playback state includes ItemId, PlayPositionTicks, IsPlaying.
func TestReportPlaying_WithBody(t *testing.T) {
	h, mock := newTestSessionHandler(t)

	mock.ExpectExec("INSERT INTO playback_state").WillReturnResult(sqlmock.NewResult(1, 1))

	body := `{"ItemId":"item-1","PlayPositionTicks":12345678,"IsPlaying":true}`
	req := httptest.NewRequest("POST", "/Sessions/sess-1/Playing", bytes.NewBufferString(body))
	req = withChiParam(req, "sessionId", "sess-1")
	w := httptest.NewRecorder()

	h.ReportPlaying(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("status = %d, want 204", w.Code)
	}
}

// TestReportStopped_Success verifies playback stopped report returns 204.
// Jellyfin compat: stopped report updates is_playing to false.
func TestReportStopped_Success(t *testing.T) {
	h, mock := newTestSessionHandler(t)

	mock.ExpectExec("UPDATE playback_state SET is_playing").WillReturnResult(sqlmock.NewResult(0, 1))

	body := `{"ItemId":"item-1","PlayPositionTicks":12345678,"IsPlaying":false}`
	req := httptest.NewRequest("POST", "/Sessions/sess-1/Stopped", bytes.NewBufferString(body))
	req = withChiParam(req, "sessionId", "sess-1")
	w := httptest.NewRecorder()

	h.ReportStopped(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("status = %d, want 204", w.Code)
	}
}

// TestCloseSession_Success verifies DELETE /Sessions/Logout returns 204.
// Jellyfin compat: logout deletes all sessions for the user.
func TestCloseSession_Success(t *testing.T) {
	h, mock := newTestSessionHandler(t)

	mock.ExpectExec("DELETE FROM sessions WHERE user_id").WillReturnResult(sqlmock.NewResult(0, 1))

	req := httptest.NewRequest("DELETE", "/Sessions/Logout", nil)
	req = withSessionAuth(req, testUserUUID)
	w := httptest.NewRecorder()

	h.CloseSession(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("status = %d, want 204", w.Code)
	}
}

// TestCloseSpecificSession_NotFound verifies 404 when session ID does not exist.
// Jellyfin compat: deleting unknown session returns 404.
func TestCloseSpecificSession_NotFound(t *testing.T) {
	h, mock := newTestSessionHandler(t)

	mock.ExpectExec("DELETE FROM sessions WHERE id").WillReturnResult(sqlmock.NewResult(0, 0))

	req := httptest.NewRequest("DELETE", "/Sessions/no-such-session", nil)
	req = withChiParam(req, "sessionId", "no-such-session")
	w := httptest.NewRecorder()

	h.CloseSpecificSession(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", w.Code)
	}
}

// TestCloseSpecificSession_Success verifies successful session deletion returns 204.
// Jellyfin compat: specific session deletion returns 204 No Content.
func TestCloseSpecificSession_Success(t *testing.T) {
	h, mock := newTestSessionHandler(t)

	mock.ExpectExec("DELETE FROM sessions WHERE id").WillReturnResult(sqlmock.NewResult(0, 1))

	req := httptest.NewRequest("DELETE", "/Sessions/sess-1", nil)
	req = withChiParam(req, "sessionId", "sess-1")
	w := httptest.NewRecorder()

	h.CloseSpecificSession(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("status = %d, want 204", w.Code)
	}
}

// TestKeepAlive_Success verifies session keepalive returns 204.
// Jellyfin compat: keepalive extends session timeout.
func TestKeepAlive_Success(t *testing.T) {
	h, mock := newTestSessionHandler(t)

	mock.ExpectExec("UPDATE sessions SET last_activity_date").WillReturnResult(sqlmock.NewResult(0, 1))

	req := httptest.NewRequest("POST", "/Sessions/sess-1/KeepAlive", nil)
	req = withChiParam(req, "sessionId", "sess-1")
	w := httptest.NewRecorder()

	h.KeepAlive(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("status = %d, want 204", w.Code)
	}
}

// TestSendMessageToSession_Success verifies sending a message returns 204.
// Jellyfin compat: session messaging returns 204 No Content.
func TestSendMessageToSession_Success(t *testing.T) {
	h, mock := newTestSessionHandler(t)

	mock.ExpectExec("INSERT INTO session_messages").WillReturnResult(sqlmock.NewResult(1, 1))

	body := `{"MessageType":"Notification","Header":"Test","Text":"Hello","TimeoutMs":5000}`
	req := httptest.NewRequest("POST", "/Sessions/sess-1/Message", bytes.NewBufferString(body))
	req = withChiParam(req, "sessionId", "sess-1")
	w := httptest.NewRecorder()

	h.SendMessageToSession(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("status = %d, want 204", w.Code)
	}
}

// TestSendMessageToSession_InvalidBody verifies 400 for malformed JSON.
// Jellyfin compat: malformed message body returns 400.
func TestSendMessageToSession_InvalidBody(t *testing.T) {
	h, _ := newTestSessionHandler(t)

	req := httptest.NewRequest("POST", "/Sessions/sess-1/Message", bytes.NewBufferString("not json"))
	req = withChiParam(req, "sessionId", "sess-1")
	w := httptest.NewRecorder()

	h.SendMessageToSession(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

// TestUpdateSessionCapability_InvalidBody verifies 400 for malformed JSON.
// Jellyfin compat: malformed capability body returns 400.
func TestUpdateSessionCapability_InvalidBody(t *testing.T) {
	h, _ := newTestSessionHandler(t)

	req := httptest.NewRequest("POST", "/Sessions/sess-1/Capabilities", bytes.NewBufferString("not json"))
	req = withChiParam(req, "sessionId", "sess-1")
	w := httptest.NewRecorder()

	h.UpdateSessionCapability(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}
