package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/bowens/kabletown/collection-service/internal/db"
	"github.com/bowens/kabletown/shared/response"
	"github.com/go-chi/chi/v5"
)

// CreateCollectionRequest represents the API request body
type CreateCollectionRequest struct {
	Name    string   `json:"Name"`
	ItemIds []string `json:"ItemIds,omitempty"`
}

// GetCollections handles GET /Collections
func GetCollections(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("userId")
	if userID == "" {
		response.BadRequest(w, "userId required")
		return
	}

	repo, ok := r.Context().Value("collectionRepo").(*db.CollectionRepository)
	if !ok {
		response.InternalServerError(w, "repository not initialized")
		return
	}

	collections, err := repo.GetCollectionsByUserID(userID)
	if err != nil {
		response.InternalServerError(w, "failed to fetch collections")
		return
	}

	// Map to CollectionDetails with ItemCount
	result := make([]db.CollectionDetails, len(collections))
	for i, c := range collections {
		details, _ := repo.GetCollectionDetails(c.ID)
		if details != nil {
			result[i] = *details
		}
	}

	// W5: Use PaginatedResponse envelope (never bare array)
	// W6: make([]T, 0) ensures [] not null
	response.PaginatedResponse(w, result, len(result), 0, len(result))
}

// GetCollection handles GET /Collections/{collectionId}
func GetCollection(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "collectionId")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.BadRequest(w, "invalid collectionId")
		return
	}

	repo, ok := r.Context().Value("collectionRepo").(*db.CollectionRepository)
	if !ok {
		response.InternalServerError(w, "repository not initialized")
		return
	}

	collection, err := repo.GetCollectionDetails(id)
	if err == db.ErrCollectionNotFound {
		response.NotFound(w, "collection")
		return
	}
	if err != nil {
		response.InternalServerError(w, "failed to fetch collection")
		return
	}

	// Add items
	items, _ := repo.GetCollectionItems(id)
	collectionDetails := struct {
		db.CollectionDetails
		ItemIds []string `json:"ItemIds"`
	}{
		CollectionDetails: *collection,
		ItemIds:           items,
	}

	response.OK(w, collectionDetails)
}

// CreateCollection handles POST /Collections
func CreateCollection(w http.ResponseWriter, r *http.Request) {
	var req CreateCollectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	if req.Name == "" {
		response.BadRequest(w, "Name is required")
		return
	}

	userID := r.URL.Query().Get("userId")
	if userID == "" {
		response.BadRequest(w, "userId required")
		return
	}

	repo, ok := r.Context().Value("collectionRepo").(*db.CollectionRepository)
	if !ok {
		response.InternalServerError(w, "repository not initialized")
		return
	}

	// Create collection
	collection, err := repo.CreateCollection(req.Name, userID, "")
	if err != nil {
		response.InternalServerError(w, "failed to create collection")
		return
	}

	// Add items if provided
	for _, itemID := range req.ItemIds {
		_, err := repo.AddItemToCollection(collection.ID, strings.TrimSpace(itemID), nil)
		if err != nil {
			// Clean up - delete the collection
			_ = repo.DeleteCollection(collection.ID)
			response.InternalServerError(w, "failed to add item to collection")
			return
		}
	}

	response.Created(w, collection)
}

// UpdateCollection handles PATCH /Collections/{collectionId}
func UpdateCollection(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "collectionId")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.BadRequest(w, "invalid collectionId")
		return
	}

	var req struct {
		Name string `json:"Name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	repo, ok := r.Context().Value("collectionRepo").(*db.CollectionRepository)
	if !ok {
		response.InternalServerError(w, "repository not initialized")
		return
	}

	if err := repo.UpdateCollection(id, req.Name); err != nil {
		if err == db.ErrCollectionNotFound {
			response.NotFound(w, "collection")
			return
		}
		response.InternalServerError(w, "failed to update collection")
		return
	}

	response.OK(w, map[string]string{"message": "collection updated"})
}

// DeleteCollection handles DELETE /Collections/{collectionId}
func DeleteCollection(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "collectionId")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.BadRequest(w, "invalid collectionId")
		return
	}

	repo, ok := r.Context().Value("collectionRepo").(*db.CollectionRepository)
	if !ok {
		response.InternalServerError(w, "repository not initialized")
		return
	}

	if err := repo.DeleteCollection(id); err != nil {
		if err == db.ErrCollectionNotFound {
			response.NotFound(w, "collection")
			return
		}
		response.InternalServerError(w, "failed to delete collection")
		return
	}

	response.NoContent(w)
}

// AddToCollection handles POST /Collections/{collectionId}/AddToCollection
func AddToCollection(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "collectionId")
	collectionID, err := strconv.Atoi(idStr)
	if err != nil {
		response.BadRequest(w, "invalid collectionId")
		return
	}

	var req struct {
		ItemIds []string `json:"ItemIds"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	repo, ok := r.Context().Value("collectionRepo").(*db.CollectionRepository)
	if !ok {
		response.InternalServerError(w, "repository not initialized")
		return
	}

	for _, itemID := range req.ItemIds {
		_, err := repo.AddItemToCollection(collectionID, strings.TrimSpace(itemID), nil)
		if err != nil {
			response.InternalServerError(w, "failed to add item")
			return
		}
	}

	response.OK(w, map[string]string{"message": "items added to collection"})
}

// RemoveFromCollection handles DELETE /Collections/{collectionId}/RemoveFromCollection
func RemoveFromCollection(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "collectionId")
	collectionID, err := strconv.Atoi(idStr)
	if err != nil {
		response.BadRequest(w, "invalid collectionId")
		return
	}

	var req struct {
		ItemIds []string `json:"ItemIds"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	repo, ok := r.Context().Value("collectionRepo").(*db.CollectionRepository)
	if !ok {
		response.InternalServerError(w, "repository not initialized")
		return
	}

	for _, itemID := range req.ItemIds {
		err := repo.RemoveItemFromCollection(collectionID, strings.TrimSpace(itemID))
		if err != nil {
			if err == db.ErrCollectionItemNotFound {
				response.NotFound(w, "collection item")
				return
			}
			response.InternalServerError(w, "failed to remove item")
			return
		}
	}

	response.OK(w, map[string]string{"message": "items removed from collection"})
}
