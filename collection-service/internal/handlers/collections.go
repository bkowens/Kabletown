package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/jellyfinhanced/collection-service/internal/db"
	"github.com/jellyfinhanced/shared/response"
)

type contextKeyType string

const repoContextKey contextKeyType = "collectionRepo"

// RegisterRoutes wires all collection routes onto the provided router.
func RegisterRoutes(r chi.Router, repo *db.CollectionRepository) {
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			ctx := context.WithValue(req.Context(), repoContextKey, repo)
			next.ServeHTTP(w, req.WithContext(ctx))
		})
	})

	r.Post("/Collections", CreateCollection)
	r.Get("/Collections/{collectionId}", GetCollection)
	r.Delete("/Collections/{collectionId}", DeleteCollection)
	r.Post("/Collections/{collectionId}/Items", AddToCollection)
	r.Delete("/Collections/{collectionId}/Items", RemoveFromCollection)
}

// getRepo extracts the CollectionRepository from the request context.
func getRepo(r *http.Request) *db.CollectionRepository {
	repo, _ := r.Context().Value(repoContextKey).(*db.CollectionRepository)
	return repo
}

// CreateCollectionRequest represents the POST /Collections request body.
type CreateCollectionRequest struct {
	Name    string   `json:"Name"`
	ItemIds []string `json:"ItemIds,omitempty"`
}

// GetCollection handles GET /Collections/{collectionId}.
func GetCollection(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "collectionId")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid collectionId")
		return
	}

	repo := getRepo(r)
	if repo == nil {
		response.WriteError(w, http.StatusInternalServerError, "repository not initialized")
		return
	}

	collection, err := repo.GetCollectionDetails(id)
	if err == db.ErrCollectionNotFound {
		response.WriteError(w, http.StatusNotFound, "collection not found")
		return
	}
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "failed to fetch collection")
		return
	}

	items, _ := repo.GetCollectionItems(id)
	if items == nil {
		items = []string{}
	}

	result := struct {
		db.CollectionDetails
		ItemIds []string `json:"ItemIds"`
	}{
		CollectionDetails: *collection,
		ItemIds:           items,
	}

	response.WriteJSON(w, http.StatusOK, result)
}

// CreateCollection handles POST /Collections.
func CreateCollection(w http.ResponseWriter, r *http.Request) {
	var req CreateCollectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" {
		response.WriteError(w, http.StatusBadRequest, "Name is required")
		return
	}

	userID := r.URL.Query().Get("userId")
	if userID == "" {
		response.WriteError(w, http.StatusBadRequest, "userId required")
		return
	}

	repo := getRepo(r)
	if repo == nil {
		response.WriteError(w, http.StatusInternalServerError, "repository not initialized")
		return
	}

	collection, err := repo.CreateCollection(req.Name, userID, "")
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "failed to create collection")
		return
	}

	for _, itemID := range req.ItemIds {
		if _, err := repo.AddItemToCollection(collection.ID, strings.TrimSpace(itemID), nil); err != nil {
			_ = repo.DeleteCollection(collection.ID)
			response.WriteError(w, http.StatusInternalServerError, "failed to add item to collection")
			return
		}
	}

	response.WriteJSON(w, http.StatusCreated, collection)
}

// DeleteCollection handles DELETE /Collections/{collectionId}.
func DeleteCollection(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "collectionId")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid collectionId")
		return
	}

	repo := getRepo(r)
	if repo == nil {
		response.WriteError(w, http.StatusInternalServerError, "repository not initialized")
		return
	}

	if err := repo.DeleteCollection(id); err != nil {
		if err == db.ErrCollectionNotFound {
			response.WriteError(w, http.StatusNotFound, "collection not found")
			return
		}
		response.WriteError(w, http.StatusInternalServerError, "failed to delete collection")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// AddToCollection handles POST /Collections/{collectionId}/Items.
func AddToCollection(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "collectionId")
	collectionID, err := strconv.Atoi(idStr)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid collectionId")
		return
	}

	var req struct {
		ItemIds []string `json:"ItemIds"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	repo := getRepo(r)
	if repo == nil {
		response.WriteError(w, http.StatusInternalServerError, "repository not initialized")
		return
	}

	for _, itemID := range req.ItemIds {
		if _, err := repo.AddItemToCollection(collectionID, strings.TrimSpace(itemID), nil); err != nil {
			response.WriteError(w, http.StatusInternalServerError, "failed to add item")
			return
		}
	}

	response.WriteJSON(w, http.StatusOK, map[string]string{"Message": "items added to collection"})
}

// RemoveFromCollection handles DELETE /Collections/{collectionId}/Items.
func RemoveFromCollection(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "collectionId")
	collectionID, err := strconv.Atoi(idStr)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid collectionId")
		return
	}

	var req struct {
		ItemIds []string `json:"ItemIds"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	repo := getRepo(r)
	if repo == nil {
		response.WriteError(w, http.StatusInternalServerError, "repository not initialized")
		return
	}

	for _, itemID := range req.ItemIds {
		if err := repo.RemoveItemFromCollection(collectionID, strings.TrimSpace(itemID)); err != nil {
			if err == db.ErrCollectionItemNotFound {
				response.WriteError(w, http.StatusNotFound, "collection item not found")
				return
			}
			response.WriteError(w, http.StatusInternalServerError, "failed to remove item")
			return
		}
	}

	response.WriteJSON(w, http.StatusOK, map[string]string{"Message": "items removed from collection"})
}
