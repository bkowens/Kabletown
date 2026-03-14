package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"kabletown/shared/auth"
)

// ApiKeyHandler handles API key management
type ApiKeyHandler struct {
	db *sql.DB
}

func NewApiKeyHandler(db *sql.DB) *ApiKeyHandler {
	return &ApiKeyHandler{db: db}
}

// GetApiKeys handles GET /Plugins/{userId}/ApiKey
func (h *ApiKeyHandler) GetApiKeys(w http.ResponseWriter, r *http.Request) {
	authUser := auth.TokenFromContext(r)
	if authUser == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	rows, err := h.db.Query(
		`SELECT id, device_id, device_name, created_at, last_use_date, is_active 
         FROM api_keys 
         WHERE user_id = ?`,
		authUser.UserID,
	)
	if err != nil {
		http.Error(w, "Failed to query API keys", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	keys := []map[string]interface{}{}
	for rows.Next() {
		var id, deviceId, deviceName string
		var createdAt, lastUseDate time.Time
		var isActive bool

		if err := rows.Scan(&id, &deviceId, &deviceName, &createdAt, &lastUseDate, &isActive); err != nil {
			continue
		}

		keys = append(keys, map[string]interface{}{
			"Id":            id,
			"DeviceId":      deviceId,
			"DeviceName":    deviceName,
			"CreatedAt":     createdAt.UTC().Format(time.RFC3339Nano),
			"LastUseDate":   lastUseDate.UTC().Format(time.RFC3339Nano),
			"IsActive":      isActive,
			"Token":         generateMaskedToken(id),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(keys)
}

// CreateApiKey handles POST /Plugins/{userId}/ApiKey
func (h *ApiKeyHandler) CreateApiKey(w http.ResponseWriter, r *http.Request) {
	authUser := auth.TokenFromContext(r)
	if authUser == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		App      string `json:"AppName"`
		Version  string `json:"AppVersion"`
		DeviceId string `json:"DeviceId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	token := auth.GenerateToken()
	now := time.Now().UTC()

	_, err := h.db.Exec(
		`INSERT INTO api_keys (id, user_id, api_key_hash, device_id, device_name, created_at, is_active)
         VALUES (?, ?, ?, ?, ?, ?, 1)`,
		generateUUID(), authUser.UserID, auth.HashToken(token), req.DeviceId, req.App, now,
	)

	if err != nil {
		log.Printf("Error creating API key: %v", err)
		http.Error(w, "Failed to create API key", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"Id":       token,
		"Token":    token,
		"IsAll":    true,
		"CanDelete": true,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// DeleteApiKey handles DELETE /Plugins/{userId}/ApiKey/{keyId}
func (h *ApiKeyHandler) DeleteApiKey(w http.ResponseWriter, r *http.Request) {
	authUser := auth.TokenFromContext(r)
	if authUser == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	keyId := chi.URLParam(r, "keyId")

	_, err := h.db.Exec(
		"UPDATE api_keys SET is_active = 0 WHERE id = ? AND user_id = ?",
		keyId, authUser.UserID,
	)

	if err != nil {
		http.Error(w, "Failed to revoke API key", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func generateMaskedToken(token string) string {
	if len(token) < 8 {
		return "***"
	}
	return token[:4] + "..." + token[len(token)-4:]
}
