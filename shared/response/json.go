package response

import (
	"encoding/json"
	"net/http"
)

// QueryResult is the Jellyfin-compatible paginated response envelope.
// It contains exactly the three fields required by the Jellyfin API — no extras.
type QueryResult[T any] struct {
	Items            []T `json:"Items"`
	TotalRecordCount int `json:"TotalRecordCount"`
	StartIndex       int `json:"StartIndex"`
}

// JSON writes v as a JSON response with the given status code.
// It sets Content-Type to application/json; charset=utf-8.
func JSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		// Nothing useful we can do at this point — headers already sent.
		return
	}
}

// Error writes a Jellyfin-compatible error body.
func Error(w http.ResponseWriter, status int, message string) {
	JSON(w, status, map[string]any{
		"Message":    message,
		"StatusCode": status,
	})
}

// PaginatedResponse builds a QueryResult from a slice, total count, and start index.
func PaginatedResponse[T any](items []T, total, startIndex int) QueryResult[T] {
	if items == nil {
		items = []T{}
	}
	return QueryResult[T]{
		Items:            items,
		TotalRecordCount: total,
		StartIndex:       startIndex,
	}
}
