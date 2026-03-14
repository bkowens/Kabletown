package testutil

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
)

// AssertStatus verifies the response has the expected HTTP status code.
func AssertStatus(t *testing.T, w *httptest.ResponseRecorder, want int) {
	t.Helper()
	if w.Code != want {
		t.Errorf("status = %d, want %d; body = %s", w.Code, want, w.Body.String())
	}
}

// AssertJSONContentType verifies the response has application/json Content-Type.
func AssertJSONContentType(t *testing.T, w *httptest.ResponseRecorder) {
	t.Helper()
	ct := w.Header().Get("Content-Type")
	if ct != "application/json; charset=utf-8" {
		t.Errorf("Content-Type = %q, want application/json; charset=utf-8", ct)
	}
}

// ParseJSONBody unmarshals the response body into a map.
func ParseJSONBody(t *testing.T, w *httptest.ResponseRecorder) map[string]interface{} {
	t.Helper()
	var body map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to unmarshal JSON body: %v; raw = %s", err, w.Body.String())
	}
	return body
}

// ParseJSONArray unmarshals the response body into a slice.
func ParseJSONArray(t *testing.T, w *httptest.ResponseRecorder) []interface{} {
	t.Helper()
	var body []interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to unmarshal JSON array: %v; raw = %s", err, w.Body.String())
	}
	return body
}

// AssertJSONField verifies a top-level JSON field exists in the response body.
func AssertJSONField(t *testing.T, body map[string]interface{}, field string) {
	t.Helper()
	if _, ok := body[field]; !ok {
		t.Errorf("JSON response missing field %q", field)
	}
}

// AssertJSONFields verifies multiple top-level JSON fields exist.
func AssertJSONFields(t *testing.T, body map[string]interface{}, fields ...string) {
	t.Helper()
	for _, f := range fields {
		if _, ok := body[f]; !ok {
			t.Errorf("JSON response missing field %q", f)
		}
	}
}

// SampleUserFixture returns a map representing a minimal Jellyfin user JSON object.
// Jellyfin compat: these fields are the minimum set expected by jellyfin-web.
func SampleUserFixture() map[string]interface{} {
	return map[string]interface{}{
		"Id":                   DefaultTestUserID,
		"Name":                 "TestUser",
		"HasPassword":          true,
		"HasConfiguredPassword": true,
		"IsAdministrator":      false,
		"IsDisabled":           false,
		"IsHidden":             false,
	}
}

// SampleItemFixture returns a map representing a minimal Jellyfin item JSON object.
// Jellyfin compat: these are the minimum fields expected by jellyfin-web for list display.
func SampleItemFixture() map[string]interface{} {
	return map[string]interface{}{
		"Id":       "22222222-2222-2222-2222-222222222222",
		"Name":     "Test Movie",
		"Type":     "Movie",
		"IsFolder": false,
	}
}

// SampleAuthResultFixture returns a map representing a Jellyfin AuthenticationResult.
// Jellyfin compat: login response shape expected by all Jellyfin clients.
func SampleAuthResultFixture() map[string]interface{} {
	return map[string]interface{}{
		"User": map[string]interface{}{
			"Id":              DefaultTestUserID,
			"Name":            "admin",
			"IsAdministrator": true,
			"HasPassword":     true,
		},
		"AccessToken": "64charshextoken0123456789abcdef0123456789abcdef0123456789abcdef",
		"ServerId":    DefaultServerID,
	}
}
