package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

// TestGetSystemInfo_ResponseShape verifies all required fields in GET /System/Info.
// Jellyfin compat: jellyfin-web relies on Id, ServerName, Version, OperatingSystem, ProductName.
func TestGetSystemInfo_ResponseShape(t *testing.T) {
	h := New("test-server-id", "TestServer")

	req := httptest.NewRequest("GET", "/System/Info", nil)
	w := httptest.NewRecorder()

	h.GetSystemInfo(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}

	ct := w.Header().Get("Content-Type")
	if ct != "application/json; charset=utf-8" {
		t.Errorf("Content-Type = %q, want application/json; charset=utf-8", ct)
	}

	var body map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	requiredFields := []string{
		"Id", "ServerName", "Version", "ProductName", "OperatingSystem",
		"StartupWizardCompleted", "LocalAddress", "CanSelfRestart",
		"CanLaunchWebBrowser", "HasPendingRestart", "IsShuttingDown",
		"SupportsLibraryMonitor", "WebSocketPortNumber", "SystemArchitecture",
	}
	for _, f := range requiredFields {
		if _, ok := body[f]; !ok {
			t.Errorf("response missing required field %q", f)
		}
	}

	if body["Id"] != "test-server-id" {
		t.Errorf("Id = %v, want test-server-id", body["Id"])
	}
	if body["ServerName"] != "TestServer" {
		t.Errorf("ServerName = %v, want TestServer", body["ServerName"])
	}
	if body["Version"] != "10.10.0" {
		t.Errorf("Version = %v, want 10.10.0", body["Version"])
	}
}

// TestGetPublicSystemInfo_SubsetOfFull verifies GET /System/Info/Public returns fewer fields.
// Jellyfin compat: public info is a strict subset, no paths or internal details.
func TestGetPublicSystemInfo_SubsetOfFull(t *testing.T) {
	h := New("test-server-id", "TestServer")

	req := httptest.NewRequest("GET", "/System/Info/Public", nil)
	w := httptest.NewRecorder()

	h.GetPublicSystemInfo(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}

	var body map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &body)

	requiredFields := []string{"Id", "ServerName", "Version", "ProductName", "OperatingSystem", "StartupWizardCompleted", "LocalAddress"}
	for _, f := range requiredFields {
		if _, ok := body[f]; !ok {
			t.Errorf("response missing required field %q", f)
		}
	}

	// Public info should NOT contain internal paths
	forbiddenFields := []string{"ProgramDataPath", "CachePath", "LogPath", "TranscodingTempPath"}
	for _, f := range forbiddenFields {
		if _, ok := body[f]; ok {
			t.Errorf("public info should not contain %q", f)
		}
	}
}

// TestPing_Response verifies GET /System/Ping returns "Jellyfin" string.
// Jellyfin compat: ping endpoint returns the string "Jellyfin" (or "Jellyfin Server").
func TestPing_Response(t *testing.T) {
	h := New("test-server-id", "TestServer")

	tests := []struct {
		method string
	}{
		{"GET"},
		{"POST"},
	}

	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/System/Ping", nil)
			w := httptest.NewRecorder()

			h.Ping(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("status = %d, want 200", w.Code)
			}

			var result string
			if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
				t.Fatalf("failed to unmarshal: %v", err)
			}
			if result != "Jellyfin" {
				t.Errorf("body = %q, want %q", result, "Jellyfin")
			}
		})
	}
}

// TestRestart_NoContent verifies POST /System/Restart returns 204.
// Jellyfin compat: restart returns 204 No Content.
func TestRestart_NoContent(t *testing.T) {
	h := New("test-server-id", "TestServer")

	req := httptest.NewRequest("POST", "/System/Restart", nil)
	w := httptest.NewRecorder()

	h.Restart(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("status = %d, want 204", w.Code)
	}
}

// TestShutdown_NoContent verifies POST /System/Shutdown returns 204.
// Jellyfin compat: shutdown returns 204 No Content.
func TestShutdown_NoContent(t *testing.T) {
	h := New("test-server-id", "TestServer")

	req := httptest.NewRequest("POST", "/System/Shutdown", nil)
	w := httptest.NewRecorder()

	h.Shutdown(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("status = %d, want 204", w.Code)
	}
}

// TestGetLogFiles_EmptyArray verifies GET /System/Logs returns empty array.
// Jellyfin compat: log files endpoint returns array (possibly empty).
func TestGetLogFiles_EmptyArray(t *testing.T) {
	h := New("test-server-id", "TestServer")

	req := httptest.NewRequest("GET", "/System/Logs", nil)
	w := httptest.NewRecorder()

	h.GetLogFiles(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}

	var body []interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to unmarshal as array: %v", err)
	}
	if len(body) != 0 {
		t.Errorf("expected empty array, got %d elements", len(body))
	}
}

// TestGetEndpointInfo_ResponseShape verifies GET /System/Endpoint response.
// Jellyfin compat: endpoint info must include IsLocal, IsInNetwork.
func TestGetEndpointInfo_ResponseShape(t *testing.T) {
	h := New("test-server-id", "TestServer")

	req := httptest.NewRequest("GET", "/System/Endpoint", nil)
	w := httptest.NewRecorder()

	h.GetEndpointInfo(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}

	var body map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &body)

	if _, ok := body["IsLocal"]; !ok {
		t.Error("missing IsLocal field")
	}
	if _, ok := body["IsInNetwork"]; !ok {
		t.Error("missing IsInNetwork field")
	}
}

// TestGetBrandingOptions_ResponseShape verifies GET /Branding/Configuration response.
// Jellyfin compat: branding must include LoginDisclaimer, CustomCss, SplashscreenEnabled.
func TestGetBrandingOptions_ResponseShape(t *testing.T) {
	h := New("test-server-id", "TestServer")

	req := httptest.NewRequest("GET", "/Branding/Configuration", nil)
	w := httptest.NewRecorder()

	h.GetBrandingOptions(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}

	var body map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &body)

	required := []string{"LoginDisclaimer", "CustomCss", "SplashscreenEnabled"}
	for _, f := range required {
		if _, ok := body[f]; !ok {
			t.Errorf("missing field %q", f)
		}
	}
}

// TestUpdateBrandingOptions_NoContent verifies POST /Branding/Configuration returns 204.
// Jellyfin compat: branding update returns 204 No Content.
func TestUpdateBrandingOptions_NoContent(t *testing.T) {
	h := New("test-server-id", "TestServer")

	req := httptest.NewRequest("POST", "/Branding/Configuration", nil)
	w := httptest.NewRecorder()

	h.UpdateBrandingOptions(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("status = %d, want 204", w.Code)
	}
}

// TestGetBrandingCss_ContentType verifies GET /Branding/Css returns text/css.
// Jellyfin compat: CSS endpoint must return text/css content type.
func TestGetBrandingCss_ContentType(t *testing.T) {
	h := New("test-server-id", "TestServer")

	req := httptest.NewRequest("GET", "/Branding/Css", nil)
	w := httptest.NewRecorder()

	h.GetBrandingCss(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}

	ct := w.Header().Get("Content-Type")
	if ct != "text/css; charset=utf-8" {
		t.Errorf("Content-Type = %q, want text/css; charset=utf-8", ct)
	}
}

// TestRegisterRoutes_RoutesExist verifies all system routes are registered.
// Jellyfin compat: all required system API endpoints must be available.
func TestRegisterRoutes_RoutesExist(t *testing.T) {
	h := New("test-server-id", "TestServer")

	// Test via a chi router with all routes registered
	r := chi.NewRouter()
	h.RegisterRoutes(r)

	routes := []struct {
		method string
		path   string
		want   int
	}{
		{"GET", "/System/Info", http.StatusOK},
		{"GET", "/System/Info/Public", http.StatusOK},
		{"GET", "/System/Ping", http.StatusOK},
		{"POST", "/System/Ping", http.StatusOK},
		{"GET", "/System/Logs", http.StatusOK},
		{"GET", "/System/Endpoint", http.StatusOK},
		{"GET", "/Branding/Configuration", http.StatusOK},
		{"GET", "/Branding/Css", http.StatusOK},
	}

	for _, rt := range routes {
		t.Run(rt.method+" "+rt.path, func(t *testing.T) {
			req := httptest.NewRequest(rt.method, rt.path, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != rt.want {
				t.Errorf("status = %d, want %d", w.Code, rt.want)
			}
		})
	}
}

// TestGetSystemInfo_ServerIDPropagation verifies the server ID from constructor appears in response.
// Jellyfin compat: every server must have a unique, stable server ID.
func TestGetSystemInfo_ServerIDPropagation(t *testing.T) {
	customID := "custom-server-uuid-12345"
	h := New(customID, "MyServer")

	req := httptest.NewRequest("GET", "/System/Info", nil)
	w := httptest.NewRecorder()

	h.GetSystemInfo(w, req)

	var body map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &body)

	if body["Id"] != customID {
		t.Errorf("Id = %v, want %v", body["Id"], customID)
	}
	if body["ServerName"] != "MyServer" {
		t.Errorf("ServerName = %v, want MyServer", body["ServerName"])
	}
}
