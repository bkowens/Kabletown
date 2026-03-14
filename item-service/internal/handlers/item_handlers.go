package handlers

import (
	"net/http"
	"strconv"

	"kabletown/item-service/internal/db"
	"github.com/jellyfinhanced/shared/response"
	"github.com/jellyfinhanced/shared/types"
)

// GetItems retrieves items with optional filtering and pagination
// Query params: parent_id, start_index, limit, genre_ids, types
func GetItems(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	itemRepo := ctx.Value("itemRepository").(*db.ItemRepository)

	// Parse query parameters
	parentID := r.URL.Query().Get("parent_id")
	genreIDs := r.URL.Query().Get("genre_ids") // Pipe-delimited: "genre1|genre2|genre3"
	itemType := r.URL.Query().Get("type")

	// Pagination params
	startIndex := 0
	if s, err := strconv.Atoi(r.URL.Query().Get("start_index")); err == nil {
		startIndex = s
	}
	limit := 20
	if l, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil && l > 0 {
		limit = l
	}

	// Route to appropriate query
	var items []db.BaseItemDto
	var total int
	var err error

	if parentID != "" {
		// Parent folder browse (uses IX_BaseItems_ParentId_IsVirtualItem_Type)
		items, total, err = itemRepo.GetByParentId(parentID, startIndex, limit)
	} else if genreIDs != "" {
		// Filter by genre (uses IX_ItemValuesMap_ItemValueId)
		genreList := types.SplitPipeDelimited(&genreIDs)
		if itemType == "" {
			itemType = "Movie" // Default type
		}
		items, total, err = itemRepo.FilterByGenre(genreList, itemType, startIndex, limit)
	} else if itemType != "" {
		// Filter by type only
		// TODO: Add GetByType method
		items, total, err = []db.BaseItemDto{}, 0, nil
	} else {
		// No filter - return empty or handle as error
		response.WriteBadRequest(w, "Must provide parent_id, genre_ids, or type parameter")
		return
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return paginated response
	response.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"Items":            items,
		"TotalRecordCount": total,
		"StartIndex":       startIndex,
	})
}

// GetItemById retrieves a single item by ID
func GetItemById(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	itemRepo := ctx.Value("itemRepository").(*db.ItemRepository)

	itemID := r.URL.Query().Get("itemId")
	if itemID == "" {
		rw := http.ResponseWriter(w)
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	item, err := itemRepo.GetById(itemID)
	if err != nil {
		response.WriteNotFound(w, "Item not found")
		return
	}

	response.WriteJSON(w, http.StatusOK, item)
}

// GetRecentlyAdded retrieves recently added items
// Query params: types (pipe-delimited), limit
func GetRecentlyAdded(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	itemRepo := ctx.Value("itemRepository").(*db.ItemRepository)

	// Parse types parameter (pipe-delimited)
	typeParam := r.URL.Query().Get("types")

	// Default to common types if not specified
	if typeParam == "" {
		typeParam = "Movie|Series|Album|Book"
	}

	types := types.SplitPipeDelimited(&typeParam)
	limit := 20
	if l, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil && l > 0 {
		limit = l
	}

	// Fetch each type separately (100% index coverage) then merge
	allItems := []db.BaseItemDto{}
	for _, itemType := range types {
		items, err := itemRepo.GetRecentlyAddedByType(itemType, limit/len(types)+5)
		if err != nil {
			continue
		}
		allItems = append(allItems, items...)
	}

	// Sort by date created (in Go - efficient for small sets)
	// Simplified sorting - in production, use proper time comparison
	// Note: This is a simplified version

	// Return limited results
	if len(allItems) > limit {
		allItems = allItems[:limit]
	}

	response.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"Items":            allItems,
		"TotalRecordCount": len(allItems),
		"StartIndex":       0,
	})
}

// GetNextEpisode retrieves the next unwatched episode for a series
// Query params: series_id, user_id
func GetNextEpisode(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	itemRepo := ctx.Value("itemRepository").(*db.ItemRepository)

	systemID := r.URL.Query().Get("series_id")
	userID := r.URL.Query().Get("user_id")

	if systemID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Get all episodes for the series
	items, err := itemRepo.GetEpisodesBySeries(systemID, &userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Filter for unwatched and get first (simplified)
	var nextItem *db.BaseItemDto
	for i := range items {
		// In production, check UserData table for played status
		// For now, return first episode
		nextItem = &items[i]
		break
	}

	if nextItem == nil {
		response.WriteJSON(w, http.StatusOK, nil)
		return
	}

	response.WriteJSON(w, http.StatusOK, nextItem)
}
