package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestWriteJSON_ContentTypeAndStatus verifies correct Content-Type header and status code.
// Jellyfin compat: all JSON responses must have application/json charset=utf-8.
func TestWriteJSON_ContentTypeAndStatus(t *testing.T) {
	tests := []struct {
		name   string
		status int
		data   interface{}
	}{
		{"200 with data", http.StatusOK, map[string]string{"key": "val"}},
		{"201 with data", http.StatusCreated, map[string]int{"id": 1}},
		{"200 with nil", http.StatusOK, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			ok := WriteJSON(w, tt.status, tt.data)
			if !ok {
				t.Error("WriteJSON returned false")
			}
			if w.Code != tt.status {
				t.Errorf("status = %d, want %d", w.Code, tt.status)
			}
			ct := w.Header().Get("Content-Type")
			if ct != "application/json; charset=utf-8" {
				t.Errorf("Content-Type = %q, want %q", ct, "application/json; charset=utf-8")
			}
		})
	}
}

// TestWriteJSON_Body verifies JSON encoding of the response body.
// Jellyfin compat: response body must be valid JSON.
func TestWriteJSON_Body(t *testing.T) {
	w := httptest.NewRecorder()
	WriteJSON(w, http.StatusOK, map[string]string{"Name": "test"})
	var result map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal body: %v", err)
	}
	if result["Name"] != "test" {
		t.Errorf("Name = %q, want %q", result["Name"], "test")
	}
}

// TestWriteError_Format verifies error response shape matches expected format.
// Jellyfin compat: error responses must contain "code" and "message" fields.
func TestWriteError_Format(t *testing.T) {
	tests := []struct {
		name       string
		status     int
		message    string
		wantStatus int
	}{
		{"400", http.StatusBadRequest, "bad input", http.StatusBadRequest},
		{"401", http.StatusUnauthorized, "not authorized", http.StatusUnauthorized},
		{"403", http.StatusForbidden, "denied", http.StatusForbidden},
		{"404", http.StatusNotFound, "missing", http.StatusNotFound},
		{"409", http.StatusConflict, "conflict", http.StatusConflict},
		{"500", http.StatusInternalServerError, "server error", http.StatusInternalServerError},
		{"418", 418, "teapot", 418}, // non-standard status
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			ok := WriteError(w, tt.status, tt.message)
			if !ok {
				t.Error("WriteError returned false")
			}
			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}
			var body map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
				t.Fatalf("failed to unmarshal: %v", err)
			}
			if _, ok := body["message"]; !ok {
				t.Error("response missing 'message' field")
			}
			if _, ok := body["code"]; !ok {
				t.Error("response missing 'code' field")
			}
			if msg, _ := body["message"].(string); msg != tt.message {
				t.Errorf("message = %q, want %q", msg, tt.message)
			}
		})
	}
}

// TestWriteNotFound_DefaultMessage verifies default message when empty string is passed.
// Jellyfin compat: 404 responses should always include a message.
func TestWriteNotFound_DefaultMessage(t *testing.T) {
	w := httptest.NewRecorder()
	WriteNotFound(w, "")
	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", w.Code)
	}
	var body map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &body)
	if msg, _ := body["message"].(string); msg == "" {
		t.Error("expected default message for empty input")
	}
}

// TestWriteUnauthorized_DefaultMessage verifies default message when empty string is passed.
// Jellyfin compat: 401 responses must include a message for the web frontend to display.
func TestWriteUnauthorized_DefaultMessage(t *testing.T) {
	w := httptest.NewRecorder()
	WriteUnauthorized(w, "")
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
	var body map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &body)
	if msg, _ := body["message"].(string); msg == "" {
		t.Error("expected default message for empty input")
	}
}

// TestWriteBadRequest_DefaultMessage verifies default message when empty string is passed.
// Jellyfin compat: 400 responses must include a message.
func TestWriteBadRequest_DefaultMessage(t *testing.T) {
	w := httptest.NewRecorder()
	WriteBadRequest(w, "")
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
	var body map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &body)
	if msg, _ := body["message"].(string); msg == "" {
		t.Error("expected default message for empty input")
	}
}

// TestWriteInternalServerError_DefaultMessage verifies default message.
// Jellyfin compat: 500 responses should not expose internal details.
func TestWriteInternalServerError_DefaultMessage(t *testing.T) {
	w := httptest.NewRecorder()
	WriteInternalServerError(w, "")
	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", w.Code)
	}
	var body map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &body)
	if msg, _ := body["message"].(string); msg == "" {
		t.Error("expected default message for empty input")
	}
}

// TestWritePaginated_ResponseShape verifies the paginated response structure.
// Jellyfin compat: all list endpoints must return Items, TotalRecordCount, StartIndex.
func TestWritePaginated_ResponseShape(t *testing.T) {
	w := httptest.NewRecorder()
	items := []map[string]string{{"Id": "1"}, {"Id": "2"}}
	WritePaginated(w, items, 100, 0, 20)
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
	// The WritePaginated wraps in WriteSuccess which wraps in APIResponse
	var body map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	// The data field contains the paginated response
	data, ok := body["data"].(map[string]interface{})
	if !ok {
		t.Fatal("response missing 'data' field or wrong type")
	}
	if _, ok := data["Items"]; !ok {
		t.Error("paginated response missing 'Items' field")
	}
	if _, ok := data["TotalRecordCount"]; !ok {
		t.Error("paginated response missing 'TotalRecordCount' field")
	}
	if _, ok := data["StartIndex"]; !ok {
		t.Error("paginated response missing 'StartIndex' field")
	}
}

// TestWriteNoContent_Status verifies 204 No Content has empty body.
// Jellyfin compat: DELETE and update endpoints return 204 with no body.
func TestWriteNoContent_Status(t *testing.T) {
	w := httptest.NewRecorder()
	ok := WriteNoContent(w)
	if !ok {
		t.Error("WriteNoContent returned false")
	}
	if w.Code != http.StatusNoContent {
		t.Errorf("status = %d, want 204", w.Code)
	}
	if w.Body.Len() != 0 {
		t.Errorf("expected empty body, got %d bytes", w.Body.Len())
	}
}

// TestWriteValidationError_ProblemDetailsFormat verifies RFC 7807 response shape.
// Jellyfin compat: validation errors follow problem+json format.
func TestWriteValidationError_ProblemDetailsFormat(t *testing.T) {
	w := httptest.NewRecorder()
	errors := map[string][]string{
		"Name": {"Name is required"},
	}
	WriteValidationError(w, errors)
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
	ct := w.Header().Get("Content-Type")
	if !strings.Contains(ct, "application/problem+json") {
		t.Errorf("Content-Type = %q, want problem+json", ct)
	}
	var body map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &body)
	if _, ok := body["errors"]; !ok {
		t.Error("response missing 'errors' field")
	}
	if _, ok := body["status"]; !ok {
		t.Error("response missing 'status' field")
	}
}

// TestSetServerHeaders_SetsRequiredHeaders verifies server identification headers.
// Jellyfin compat: X-MediaBrowser-Server-Id must be present on responses.
func TestSetServerHeaders_SetsRequiredHeaders(t *testing.T) {
	SetServerID("test-server-id-123")
	w := httptest.NewRecorder()
	SetServerHeaders(w)
	sid := w.Header().Get(XMediaBrowserServerID)
	if sid != "test-server-id-123" {
		t.Errorf("X-MediaBrowser-Server-Id = %q, want %q", sid, "test-server-id-123")
	}
	ver := w.Header().Get(XApplicationVersion)
	if ver == "" {
		t.Error("X-Application-Version header is empty")
	}
}

// TestGetServerID_Generated verifies server ID is auto-generated when not set.
// Jellyfin compat: every server instance needs a unique identifier.
func TestGetServerID_Generated(t *testing.T) {
	// Reset for test
	ServerID = ""
	id := GetServerID()
	if id == "" {
		t.Error("GetServerID returned empty string")
	}
	// Should be stable once generated
	id2 := GetServerID()
	if id != id2 {
		t.Error("GetServerID not stable across calls")
	}
}

// TestSetServerID_Explicit verifies explicit server ID setting.
// Jellyfin compat: server ID is configured at startup.
func TestSetServerID_Explicit(t *testing.T) {
	SetServerID("explicit-id")
	if GetServerID() != "explicit-id" {
		t.Errorf("got %q, want %q", GetServerID(), "explicit-id")
	}
}

// TestDefaultCORSHeaders verifies CORS configuration.
// Jellyfin compat: CORS must allow X-Emby-Authorization header.
func TestDefaultCORSHeaders_AllowsEmbyAuth(t *testing.T) {
	h := DefaultCORSHeaders()
	if !strings.Contains(h.AllowHeaders, "X-Emby-Authorization") {
		t.Error("CORS AllowHeaders should include X-Emby-Authorization")
	}
	if h.AllowOrigin != "*" {
		t.Errorf("AllowOrigin = %q, want %q", h.AllowOrigin, "*")
	}
}

// TestAddCORSHeaders_SetsAll verifies all CORS headers are applied.
// Jellyfin compat: CORS headers must be present for web frontend.
func TestAddCORSHeaders_SetsAll(t *testing.T) {
	w := httptest.NewRecorder()
	AddCORSHeaders(w, DefaultCORSHeaders())
	headers := []string{
		"Access-Control-Allow-Origin",
		"Access-Control-Allow-Methods",
		"Access-Control-Allow-Headers",
		"Access-Control-Expose-Headers",
		"Access-Control-Max-Age",
	}
	for _, h := range headers {
		if w.Header().Get(h) == "" {
			t.Errorf("missing CORS header: %s", h)
		}
	}
}

// TestWriteConflict verifies 409 response.
// Jellyfin compat: duplicate resource creation returns 409.
func TestWriteConflict(t *testing.T) {
	w := httptest.NewRecorder()
	WriteConflict(w, "already exists")
	if w.Code != http.StatusConflict {
		t.Errorf("status = %d, want 409", w.Code)
	}
}

// TestWriteForbidden verifies 403 response.
// Jellyfin compat: non-admin access to admin endpoints returns 403.
func TestWriteForbidden(t *testing.T) {
	w := httptest.NewRecorder()
	WriteForbidden(w, "access denied")
	if w.Code != http.StatusForbidden {
		t.Errorf("status = %d, want 403", w.Code)
	}
}

// TestWriteNotImplemented verifies 501 response.
// Jellyfin compat: stub endpoints return 501 for unimplemented features.
func TestWriteNotImplemented(t *testing.T) {
	w := httptest.NewRecorder()
	WriteNotImplemented(w, "")
	if w.Code != http.StatusNotImplemented {
		t.Errorf("status = %d, want 501", w.Code)
	}
}

// TestPreflightHandler verifies OPTIONS preflight response.
// Jellyfin compat: CORS preflight must return 204 with all CORS headers.
func TestPreflightHandler(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("OPTIONS", "/test", nil)
	PreflightHandler(w, r)
	if w.Code != http.StatusNoContent {
		t.Errorf("status = %d, want 204", w.Code)
	}
	if w.Header().Get("Access-Control-Allow-Origin") == "" {
		t.Error("missing CORS header on preflight response")
	}
}

// TestPaginatedResponse_JSONFields verifies the JSON field names match Jellyfin's format.
// Jellyfin compat: jellyfin-web expects Items, TotalRecordCount, StartIndex.
func TestPaginatedResponse_JSONFields(t *testing.T) {
	resp := PaginatedResponse{
		Items:      []string{"a", "b"},
		TotalCount: 10,
		StartIndex: 0,
		Limit:      20,
	}
	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	var m map[string]interface{}
	json.Unmarshal(data, &m)

	requiredFields := []string{"Items", "TotalRecordCount", "StartIndex", "Limit"}
	for _, f := range requiredFields {
		if _, ok := m[f]; !ok {
			t.Errorf("missing JSON field %q in PaginatedResponse", f)
		}
	}
}
