package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/jellyfinhanced/shared/response"
)

// searchHintItem is the JSON representation of a single search hint.
type searchHintItem struct {
	ItemId          string  `json:"ItemId"`
	Id              string  `json:"Id"`
	Name            string  `json:"Name"`
	Type            string  `json:"Type"`
	RunTimeTicks    *int64  `json:"RunTimeTicks"`
	ProductionYear  *int    `json:"ProductionYear"`
	PrimaryImageTag string  `json:"PrimaryImageTag"`
}

// searchHintsResult is the top-level search response envelope.
type searchHintsResult struct {
	SearchHints      []searchHintItem `json:"SearchHints"`
	TotalRecordCount int              `json:"TotalRecordCount"`
}

// itemsStubResult is the stub response for the /Items endpoint.
type itemsStubResult struct {
	Items            []struct{} `json:"Items"`
	TotalRecordCount int        `json:"TotalRecordCount"`
	StartIndex       int        `json:"StartIndex"`
}

// SearchHints handles GET /Search/Hints.
func (h *Handler) SearchHints(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	searchTerm := q.Get("SearchTerm")
	if searchTerm == "" {
		response.WriteError(w, http.StatusBadRequest, "SearchTerm is required")
		return
	}

	startIndex := 0
	if s := q.Get("StartIndex"); s != "" {
		if v, err := strconv.Atoi(s); err == nil && v >= 0 {
			startIndex = v
		}
	}

	limit := 20
	if l := q.Get("Limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 {
			limit = v
		}
	}

	var includeTypes []string
	if raw := q.Get("IncludeItemTypes"); raw != "" {
		for _, t := range strings.Split(raw, ",") {
			t = strings.TrimSpace(t)
			if t != "" {
				includeTypes = append(includeTypes, t)
			}
		}
	}

	rows, total, err := h.repo.Search(searchTerm, includeTypes, limit, startIndex)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "search failed")
		return
	}

	hints := make([]searchHintItem, 0, len(rows))
	for _, row := range rows {
		hints = append(hints, searchHintItem{
			ItemId:          row.Id,
			Id:              row.Id,
			Name:            row.Name,
			Type:            row.Type,
			RunTimeTicks:    row.DurationTicks,
			ProductionYear:  row.ProductionYear,
			PrimaryImageTag: "",
		})
	}

	response.WriteJSON(w, http.StatusOK, searchHintsResult{
		SearchHints:      hints,
		TotalRecordCount: total,
	})
}

// ListItems handles GET /Items — stub returning an empty paginated result.
func (h *Handler) ListItems(w http.ResponseWriter, r *http.Request) {
	response.WriteJSON(w, http.StatusOK, itemsStubResult{
		Items:            []struct{}{},
		TotalRecordCount: 0,
		StartIndex:       0,
	})
}
