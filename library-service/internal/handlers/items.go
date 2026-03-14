// Package handlers provides HTTP handlers for library service.
package handlers

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/jellyfinhanced/shared/auth"
	"github.com/jellyfinhanced/shared/response"
	"github.com/jmoiron/sqlx"

	"github.com/jellyfinhanced/library-service/internal/dto"
	libraryQuery "github.com/jellyfinhanced/library-service/internal/query"
)

// ItemsHandler handles item-related endpoints.
type ItemsHandler struct {
	db *sqlx.DB
}

// NewItemsHandler creates a new ItemsHandler.
func NewItemsHandler(db *sqlx.DB) *ItemsHandler {
	return &ItemsHandler{db: db}
}

// GetItems handles GET /Items.
func (h *ItemsHandler) GetItems(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserFromContext(r.Context())
	if userID == uuid.Nil {
		response.WriteUnauthorized(w, "Authorization required")
		return
	}

	query, err := libraryQuery.Parse(r.URL.Query(), userID.String())
	if err != nil {
		response.WriteBadRequest(w, err.Error())
		return
	}

	if err := query.Validate(); err != nil {
		response.WriteBadRequest(w, err.Error())
		return
	}

	builder := libraryQuery.NewSQLBuilder(query)
	sqlStr, params := builder.Build()

	rows, err := h.db.QueryContext(r.Context(), sqlStr, params...)
	if err != nil {
		response.WriteInternalServerError(w, "Failed to query items")
		return
	}
	defer rows.Close()

	var items []dto.BaseItemDto
	for rows.Next() {
		var item dto.BaseItemDto
		if err := rows.Scan(&item.Id, &item.Name, &item.Type); err != nil {
			continue
		}
		items = append(items, item)
	}
	if items == nil {
		items = []dto.BaseItemDto{}
	}

	countSQL, countParams := builder.GetCountQuery()
	var totalCount int64
	if err := h.db.QueryRowContext(r.Context(), countSQL, countParams...).Scan(&totalCount); err != nil {
		totalCount = int64(len(items))
	}

	response.WriteJSON(w, http.StatusOK, dto.QueryResult[dto.BaseItemDto]{
		Items:            items,
		TotalRecordCount: totalCount,
		StartIndex:       query.StartIndex,
	})
}

// GetItem handles GET /Items/{id}.
func (h *ItemsHandler) GetItem(w http.ResponseWriter, r *http.Request) {
	itemId := r.PathValue("id")
	if itemId == "" || !isValidGUID(itemId) {
		response.WriteBadRequest(w, "Valid item ID required")
		return
	}

	var item dto.BaseItemDto
	err := h.db.QueryRowContext(r.Context(),
		`SELECT Id, Name, Type FROM base_items WHERE Id = ? LIMIT 1`, itemId,
	).Scan(&item.Id, &item.Name, &item.Type)
	if err != nil {
		response.WriteNotFound(w, "Item not found")
		return
	}

	userID := auth.GetUserFromContext(r.Context())
	if userID != uuid.Nil {
		userData, _ := h.getUserData(r.Context(), userID.String(), itemId)
		item.UserData = userData
	}

	response.WriteJSON(w, http.StatusOK, item)
}

// GetNextUp handles GET /Items/NextUp.
func (h *ItemsHandler) GetNextUp(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserFromContext(r.Context())
	if userID == uuid.Nil {
		response.WriteUnauthorized(w, "Authorization required")
		return
	}

	rows, err := h.db.QueryContext(r.Context(), `
		SELECT e.Id, e.Name, e.IndexNumber, e.ParentId
		FROM base_items e
		INNER JOIN base_items s ON e.ParentId = s.Id AND s.Type = 'Series'
		LEFT JOIN UserData ud ON e.Id = ud.ItemId AND ud.UserId = ?
		WHERE e.Type = 'Episode'
		AND (ud.Played = 0 OR ud.Played IS NULL)
		ORDER BY e.IndexNumber, e.PremiereDate DESC
		LIMIT 50`, userID.String())
	if err != nil {
		response.WriteInternalServerError(w, "Failed to get next up episodes")
		return
	}
	defer rows.Close()

	var items []dto.BaseItemDto
	for rows.Next() {
		var item dto.BaseItemDto
		var indexNum *int
		if err := rows.Scan(&item.Id, &item.Name, &indexNum, &item.ParentId); err != nil {
			continue
		}
		item.IndexNumber = indexNum
		items = append(items, item)
	}
	if items == nil {
		items = []dto.BaseItemDto{}
	}
	response.WriteJSON(w, http.StatusOK, items)
}

// GetResume handles GET /Items/Resume.
func (h *ItemsHandler) GetResume(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserFromContext(r.Context())
	if userID == uuid.Nil {
		response.WriteUnauthorized(w, "Authorization required")
		return
	}

	rows, err := h.db.QueryContext(r.Context(), `
		SELECT i.Id, i.Name, i.Type, ud.PlaybackPositionTicks
		FROM base_items i
		INNER JOIN UserData ud ON i.Id = ud.ItemId AND ud.UserId = ?
		WHERE ud.PlaybackPositionTicks > 0 AND (ud.Played = 0 OR ud.Played IS NULL)
		ORDER BY ud.ItemId DESC
		LIMIT 50`, userID.String())
	if err != nil {
		response.WriteInternalServerError(w, "Failed to get resume items")
		return
	}
	defer rows.Close()

	var items []dto.BaseItemDto
	for rows.Next() {
		var item dto.BaseItemDto
		var resumeTicks int64
		if err := rows.Scan(&item.Id, &item.Name, &item.Type, &resumeTicks); err != nil {
			continue
		}
		item.UserData = &dto.UserDataDto{PlaybackPositionTicks: resumeTicks}
		items = append(items, item)
	}
	if items == nil {
		items = []dto.BaseItemDto{}
	}
	response.WriteJSON(w, http.StatusOK, items)
}

// GetAncestors handles GET /Items/{id}/Ancestors.
func (h *ItemsHandler) GetAncestors(w http.ResponseWriter, r *http.Request) {
	itemId := r.PathValue("id")
	if itemId == "" || !isValidGUID(itemId) {
		response.WriteBadRequest(w, "Valid item ID required")
		return
	}

	rows, err := h.db.QueryContext(r.Context(), `
		WITH RECURSIVE ancestors AS (
			SELECT Id, Name, Type FROM base_items WHERE Id = ?
			UNION ALL
			SELECT i.Id, i.Name, i.Type FROM base_items i
			INNER JOIN ancestors a ON i.Id = a.ParentId
		)
		SELECT Id, Name, Type FROM ancestors WHERE Id != ? ORDER BY Id`,
		itemId, itemId)
	if err != nil {
		response.WriteInternalServerError(w, "Failed to get ancestors")
		return
	}
	defer rows.Close()

	var items []dto.BaseItemDto
	for rows.Next() {
		var item dto.BaseItemDto
		if err := rows.Scan(&item.Id, &item.Name, &item.Type); err != nil {
			continue
		}
		items = append(items, item)
	}
	if items == nil {
		items = []dto.BaseItemDto{}
	}
	response.WriteJSON(w, http.StatusOK, items)
}

func isValidGUID(id string) bool {
	_, err := uuid.Parse(id)
	return err == nil
}

func (h *ItemsHandler) getUserData(ctx context.Context, userID, itemID string) (*dto.UserDataDto, error) {
	var ud dto.UserDataDto
	err := h.db.QueryRowContext(ctx, `
		SELECT Played, PlaybackPositionTicks, PlayCount, IsFavorite
		FROM UserData WHERE ItemId = ? AND UserId = ?`, itemID, userID,
	).Scan(&ud.Played, &ud.PlaybackPositionTicks, &ud.PlayCount, &ud.IsFavorite)
	if err != nil {
		return nil, err
	}
	return &ud, nil
}

// RegisterRoutes registers all item routes.
func RegisterRoutes(mux *http.ServeMux, h *ItemsHandler) {
	mux.HandleFunc("GET /Items", h.GetItems)
	mux.HandleFunc("GET /Items/{id}", h.GetItem)
	mux.HandleFunc("GET /Items/{id}/Ancestors", h.GetAncestors)
	mux.HandleFunc("GET /Items/NextUp", h.GetNextUp)
	mux.HandleFunc("GET /Items/Resume", h.GetResume)
}
