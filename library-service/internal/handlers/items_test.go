package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/jellyfinhanced/shared/auth"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
)

const testItemUserID = "11111111-1111-1111-1111-111111111111"
const testItemID = "22222222-2222-2222-2222-222222222222"

// newTestItemsHandler creates an ItemsHandler backed by go-sqlmock.
func newTestItemsHandler(t *testing.T) (*ItemsHandler, sqlmock.Sqlmock) {
	t.Helper()
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
	h := NewItemsHandler(sqlxDB)
	return h, mock
}

// withItemsAuth injects auth context for items handler tests.
func withItemsAuth(r *http.Request, userID string) *http.Request {
	uid := uuid.MustParse(userID)
	info := &auth.AuthInfo{
		UserID:   uid,
		Username: "testuser",
		IsAdmin:  false,
	}
	return r.WithContext(auth.SetAuthInContext(r.Context(), info))
}

// TestGetItems_Unauthorized verifies 401 when no auth context.
// Jellyfin compat: /Items requires authentication.
func TestGetItems_Unauthorized(t *testing.T) {
	h, _ := newTestItemsHandler(t)

	req := httptest.NewRequest("GET", "/Items", nil)
	w := httptest.NewRecorder()

	h.GetItems(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
}

// TestGetItems_EmptyResult verifies empty item list returns proper structure.
// Jellyfin compat: Items query returns {Items: [], TotalRecordCount: 0, StartIndex: 0}.
func TestGetItems_EmptyResult(t *testing.T) {
	h, mock := newTestItemsHandler(t)

	// Main query returns no rows
	mock.ExpectQuery("SELECT").WillReturnRows(
		sqlmock.NewRows([]string{"Id", "Name", "Type"}))

	// Count query
	mock.ExpectQuery("SELECT COUNT").WillReturnRows(
		sqlmock.NewRows([]string{"count"}).AddRow(0))

	req := httptest.NewRequest("GET", "/Items?IncludeItemTypes=Movie", nil)
	req = withItemsAuth(req, testItemUserID)
	w := httptest.NewRecorder()

	h.GetItems(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", w.Code, w.Body.String())
	}

	var result map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &result)

	// Verify Jellyfin-compatible response shape
	if _, ok := result["Items"]; !ok {
		t.Error("response missing 'Items' field")
	}
	if _, ok := result["TotalRecordCount"]; !ok {
		t.Error("response missing 'TotalRecordCount' field")
	}
	if _, ok := result["StartIndex"]; !ok {
		t.Error("response missing 'StartIndex' field")
	}

	items, ok := result["Items"].([]interface{})
	if !ok {
		t.Fatal("Items should be an array")
	}
	if len(items) != 0 {
		t.Errorf("expected 0 items, got %d", len(items))
	}
}

// TestGetItems_WithResults verifies items are returned with correct structure.
// Jellyfin compat: each item must include Id, Name, Type fields.
func TestGetItems_WithResults(t *testing.T) {
	h, mock := newTestItemsHandler(t)

	rows := sqlmock.NewRows([]string{"Id", "Name", "Type"}).
		AddRow("item-1", "Test Movie", "Movie").
		AddRow("item-2", "Test Series", "Series")
	mock.ExpectQuery("SELECT").WillReturnRows(rows)

	mock.ExpectQuery("SELECT COUNT").WillReturnRows(
		sqlmock.NewRows([]string{"count"}).AddRow(2))

	req := httptest.NewRequest("GET", "/Items?IncludeItemTypes=Movie,Series", nil)
	req = withItemsAuth(req, testItemUserID)
	w := httptest.NewRecorder()

	h.GetItems(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	var result map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &result)

	items := result["Items"].([]interface{})
	if len(items) != 2 {
		t.Errorf("expected 2 items, got %d", len(items))
	}

	total := result["TotalRecordCount"].(float64)
	if total != 2 {
		t.Errorf("TotalRecordCount = %v, want 2", total)
	}
}

// TestGetItem_ValidID verifies GET /Items/{id} returns single item.
// Jellyfin compat: item response includes Id, Name, Type.
func TestGetItem_ValidID(t *testing.T) {
	h, mock := newTestItemsHandler(t)

	mock.ExpectQuery("SELECT .+ FROM base_items WHERE Id").WithArgs(testItemID).
		WillReturnRows(sqlmock.NewRows([]string{"Id", "Name", "Type"}).
			AddRow(testItemID, "Test Movie", "Movie"))

	// No UserData query expected since no auth
	req := httptest.NewRequest("GET", "/Items/"+testItemID, nil)
	// Use Go 1.22+ path value
	req.SetPathValue("id", testItemID)
	w := httptest.NewRecorder()

	h.GetItem(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", w.Code, w.Body.String())
	}

	var item map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &item)
	if item["Id"] != testItemID {
		t.Errorf("Id = %v, want %s", item["Id"], testItemID)
	}
	if item["Name"] != "Test Movie" {
		t.Errorf("Name = %v, want Test Movie", item["Name"])
	}
}

// TestGetItem_InvalidID verifies 400 for non-UUID ID.
// Jellyfin compat: invalid GUID format returns 400.
func TestGetItem_InvalidID(t *testing.T) {
	h, _ := newTestItemsHandler(t)

	req := httptest.NewRequest("GET", "/Items/not-a-uuid", nil)
	req.SetPathValue("id", "not-a-uuid")
	w := httptest.NewRecorder()

	h.GetItem(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

// TestGetItem_EmptyID verifies 400 when no ID is provided.
// Jellyfin compat: empty item ID returns 400.
func TestGetItem_EmptyID(t *testing.T) {
	h, _ := newTestItemsHandler(t)

	req := httptest.NewRequest("GET", "/Items/", nil)
	req.SetPathValue("id", "")
	w := httptest.NewRecorder()

	h.GetItem(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

// TestGetItem_NotFound verifies 404 when item does not exist.
// Jellyfin compat: missing item returns 404.
func TestGetItem_NotFound(t *testing.T) {
	h, mock := newTestItemsHandler(t)

	mock.ExpectQuery("SELECT .+ FROM base_items WHERE Id").WithArgs(testItemID).
		WillReturnRows(sqlmock.NewRows([]string{"Id", "Name", "Type"}))

	req := httptest.NewRequest("GET", "/Items/"+testItemID, nil)
	req.SetPathValue("id", testItemID)
	w := httptest.NewRecorder()

	h.GetItem(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", w.Code)
	}
}

// TestGetItem_WithUserData verifies UserData is included when auth is present.
// Jellyfin compat: authenticated requests include UserData in item responses.
func TestGetItem_WithUserData(t *testing.T) {
	h, mock := newTestItemsHandler(t)

	mock.ExpectQuery("SELECT .+ FROM base_items WHERE Id").WithArgs(testItemID).
		WillReturnRows(sqlmock.NewRows([]string{"Id", "Name", "Type"}).
			AddRow(testItemID, "Test Movie", "Movie"))

	mock.ExpectQuery("SELECT .+ FROM UserData WHERE ItemId").WithArgs(testItemID, testItemUserID).
		WillReturnRows(sqlmock.NewRows([]string{"Played", "PlaybackPositionTicks", "PlayCount", "IsFavorite"}).
			AddRow(true, int64(12345), 3, true))

	req := httptest.NewRequest("GET", "/Items/"+testItemID, nil)
	req.SetPathValue("id", testItemID)
	req = withItemsAuth(req, testItemUserID)
	w := httptest.NewRecorder()

	h.GetItem(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", w.Code, w.Body.String())
	}

	var item map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &item)

	ud, ok := item["UserData"].(map[string]interface{})
	if !ok {
		t.Fatal("expected UserData field in response")
	}
	if ud["Played"] != true {
		t.Error("UserData.Played should be true")
	}
	if ud["IsFavorite"] != true {
		t.Error("UserData.IsFavorite should be true")
	}
}

// TestGetNextUp_Unauthorized verifies 401 when no auth context.
// Jellyfin compat: NextUp requires authentication.
func TestGetNextUp_Unauthorized(t *testing.T) {
	h, _ := newTestItemsHandler(t)

	req := httptest.NewRequest("GET", "/Items/NextUp", nil)
	w := httptest.NewRecorder()

	h.GetNextUp(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
}

// TestGetNextUp_EmptyResult verifies empty NextUp returns empty array.
// Jellyfin compat: empty NextUp returns empty array, not null.
func TestGetNextUp_EmptyResult(t *testing.T) {
	h, mock := newTestItemsHandler(t)

	mock.ExpectQuery("SELECT .+ FROM base_items").WillReturnRows(
		sqlmock.NewRows([]string{"Id", "Name", "IndexNumber", "ParentId"}))

	req := httptest.NewRequest("GET", "/Items/NextUp", nil)
	req = withItemsAuth(req, testItemUserID)
	w := httptest.NewRecorder()

	h.GetNextUp(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}

	var items []interface{}
	json.Unmarshal(w.Body.Bytes(), &items)
	if items == nil {
		t.Error("expected empty array, not null")
	}
	if len(items) != 0 {
		t.Errorf("expected 0 items, got %d", len(items))
	}
}

// TestGetResume_Unauthorized verifies 401 when no auth context.
// Jellyfin compat: Resume requires authentication.
func TestGetResume_Unauthorized(t *testing.T) {
	h, _ := newTestItemsHandler(t)

	req := httptest.NewRequest("GET", "/Items/Resume", nil)
	w := httptest.NewRecorder()

	h.GetResume(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
}

// TestGetAncestors_InvalidID verifies 400 for non-UUID ancestor ID.
// Jellyfin compat: invalid GUID format returns 400.
func TestGetAncestors_InvalidID(t *testing.T) {
	h, _ := newTestItemsHandler(t)

	req := httptest.NewRequest("GET", "/Items/bad-id/Ancestors", nil)
	req.SetPathValue("id", "bad-id")
	w := httptest.NewRecorder()

	h.GetAncestors(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

// TestIsValidGUID verifies GUID validation.
// Jellyfin compat: both hyphenated and standard UUID formats must be accepted.
func TestIsValidGUID(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"11111111-1111-1111-1111-111111111111", true},
		{"00000000-0000-0000-0000-000000000000", true},
		{"not-a-guid", false},
		{"", false},
		{"11111111111111111111111111111111", true}, // uuid.Parse accepts non-hyphenated
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := isValidGUID(tt.input)
			if got != tt.want {
				t.Errorf("isValidGUID(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// TestRegisterRoutes_AllEndpoints verifies all item routes are registered.
// Jellyfin compat: all item API endpoints must be available.
func TestRegisterRoutes_AllEndpoints(t *testing.T) {
	h, mock := newTestItemsHandler(t)
	_ = mock

	mux := http.NewServeMux()
	RegisterRoutes(mux, h)

	// Test that routes are registered by making requests
	// (they may fail due to missing auth/params, but should not 404)
	// Verify routes are registered by checking mux handles key paths.
	// Go 1.22+ ServeMux uses "GET /Items" format patterns.
	routes := []struct {
		path string
		want int // expected status (not 404/405)
	}{
		{"/Items", http.StatusUnauthorized},    // no auth => 401
		{"/Items/NextUp", http.StatusUnauthorized}, // no auth => 401
		{"/Items/Resume", http.StatusUnauthorized}, // no auth => 401
	}

	for _, rt := range routes {
		t.Run(rt.path, func(t *testing.T) {
			req := httptest.NewRequest("GET", rt.path, nil)
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, req)
			// Just verify the mux matched a handler (not 404 from mux itself)
			if w.Code == http.StatusNotFound {
				t.Errorf("route %s not registered (got 404)", rt.path)
			}
		})
	}
}

// Ensure context import is used
var _ = context.Background
