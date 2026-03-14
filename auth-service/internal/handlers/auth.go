package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"kabletown/shared/auth"
)

type AuthHandler struct {
	db *sql.DB
}

func NewAuthHandler(db *sql.DB) *AuthHandler {
	return &AuthHandler{db: db}
}

// AuthenticateUser handles POST /Users/AuthenticateByName
func (h *AuthHandler) AuthenticateUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		SubmitPassword bool   `json:"SubmitPassword"`
		Username       string `json:"Username"`
		Password       string `json:"Password"`
		DeviceId       string `json:"DeviceId"`
		App            string `json:"App"`
		AppVersion     string `json:"AppVersion"`
		DeviceName     string `json:"DeviceName"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Hash password and validate
	passwordHash := auth.HashToken(req.Password)

	var userId, username string
	var isAdmin bool
	err := h.db.QueryRow(
		`SELECT Id, Username, IsAdmin FROM users WHERE Username = ? AND password_hash = ?`,
		req.Username, passwordHash,
	).Scan(&userId, &username, &isAdmin)

	if err != nil {
		log.Printf("Authentication failed for user %s: %v", req.Username, err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Generate session token
	token := auth.GenerateToken()

	// Create/record device
	err = h.recordDevice(userId, req.DeviceId, req.DeviceName, req.App)
	if err != nil {
		log.Printf("Error recording device: %v", err)
	}

	// Create API key
	err = h.storeAPIKey(userId, token, req.DeviceId, req.DeviceName)
	if err != nil {
		log.Printf("Error creating API key: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Update last login
	now := time.Now().UTC()
	_, err = h.db.Exec(
		"UPDATE users SET last_login_date = ? WHERE Id = ?",
		now, userId,
	)
	if err != nil {
		log.Printf("Error updating last login: %v", err)
	}

	// Return auth response
	response := map[string]interface{}{
		"Id":       userId,
		"Username": username,
		"Policy": map[string]interface{}{
			"IsAdministrator": isAdmin,
			"IsDisabled":      false,
			"IsHidden":        false,
		},
		"SessionInfo": map[string]interface{}{
			"Id":                  token,
			"DeviceId":            req.DeviceId,
			"ApplicationName":     req.App,
			"DeviceName":          req.DeviceName,
			"LastActivityDate":    now.Format(time.RFC3339Nano),
			"LastUserName":        username,
			"SupportsMediaDeletion": true,
			"SupportsPersistentId": true,
		},
		"AccessToken": token,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetPublicUsers handles GET /Users/Public
func (h *AuthHandler) GetPublicUsers(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query(
		`SELECT Id, Username, PrimaryImageTag, HasConfiguredPassword, HasConfiguredEmail 
         FROM users 
         WHERE IsHidden = 0 AND IsDisabled = 0`,
	)
	if err != nil {
		http.Error(w, "Failed to query users", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	users := []map[string]interface{}{}
	for rows.Next() {
		var id, username, tag string
		var hasPassword, hasEmail bool
		if err := rows.Scan(&id, &username, &tag, &hasPassword, &hasEmail); err != nil {
			continue
		}
		users = append(users, map[string]interface{}{
			"Id":                       id,
			"Username":                 username,
			"PrimaryImageTag":          tag,
			"HasConfiguredPassword":    hasPassword,
			"HasConfiguredEmail":       hasEmail,
			"HasPrimaryImage":          tag != "",
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

// ReportClientInfo handles POST /Clients
func (h *AuthHandler) ReportClientInfo(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Application string `json:"Application"`
		Version     string `json:"Version"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Log client info for compatibility
	log.Printf("Client reported: %s v%s", req.Application, req.Version)

	w.WriteHeader(http.StatusNoContent)
}

// ValidateUser handles PUT /Users/ValidatePassword
func (h *AuthHandler) ValidateUser(w http.ResponseWriter, r *http.Request) {
	authRequired := r.Header.Get("X-Emby-Authorization") != ""
	if !authRequired {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	user := auth.TokenFromContext(r)
	if user == nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	var req struct {
		Password string `json:"pw"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Validate password
	passwordHash := auth.HashToken(req.Password)
	var count int
	err := h.db.QueryRow(
		"SELECT COUNT(*) FROM users WHERE Id = ? AND password_hash = ?",
		user.UserID, passwordHash,
	).Scan(&count)

	if err != nil || count == 0 {
		http.Error(w, "Invalid password", http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// GetSessionInfo handles GET /Sessions (with optional user filtering)
func (h *AuthHandler) GetSessionInfo(w http.ResponseWriter, r *http.Request) {
	user := auth.TokenFromContext(r)
	
	query := `SELECT s.Id, u.Username, s.DeviceId, s.ClientType, s.IPAddress, 
                     s.LastActivityDate, s.UserId, u.IsAdmin
              FROM sessions s
              JOIN users u ON s.UserId = u.Id
              WHERE s.IsActive = 1`

	var rows *sql.Rows
	var err error
	
	if user != nil {
		rows, err = h.db.Query(query+" AND s.UserId = ?", user.UserID)
	} else {
		rows, err = h.db.Query(query)
	}

	if err != nil {
		http.Error(w, "Failed to query sessions", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	sessions := []map[string]interface{}{}
	for rows.Next() {
		var id, username, deviceId, clientType, ipAddress string
		var lastActivity time.Time
		var userId string
		var isAdmin bool
		
		if err := rows.Scan(&id, &username, &deviceId, &clientType, &ipAddress,
			&lastActivity, &userId, &isAdmin); err != nil {
			continue
		}
		
		sessions = append(sessions, map[string]interface{}{
			"Id":               id,
			"UserId":           userId,
			"ApplicationName":  username,
			"DeviceId":         deviceId,
			"DeviceName":       clientType,
			"Client":           clientType,
			"LastActivityDate": lastActivity.Format(time.RFC3339Nano),
			"ServerId":         "kabletown-main",
			"SupportsMediaControl": true,
			"SupportsRemoteControl": true,
			"PlayState": map[string]interface{}{
				"IsPaused":         false,
				"IsMuted":          false,
				"VolumeLevel":      100,
				"Can_seek":         false,
				"IsPlaying":        false,
			},
			"CanBeHidden": false,
			"HasCustomDeviceName": false,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sessions)
}

// KillSession handles DELETE /Sessions/{id}
func (h *AuthHandler) KillSession(w http.ResponseWriter, r *http.Request) {
	user := auth.TokenFromContext(r)
	
	// Check if user is admin or killing own session
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	sessionId := chi.URLParam(r, "id")
	
	if !user.IsAdmin && sessionId != user.Token {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	_, err := h.db.Exec(
		"UPDATE sessions SET IsActive = 0 WHERE Id = ?",
		sessionId,
	)
	if err != nil {
		http.Error(w, "Failed to kill session", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// GetSessionInfoByID handles GET /Sessions/{id}
func (h *AuthHandler) GetSessionInfoByID(w http.ResponseWriter, r *http.Request) {
	sessionId := chi.URLParam(r, "id")
	
	var id, username, deviceId, clientType, ipAddress string
	var lastActivity time.Time
	var userId string
	var isAdmin bool

	err := h.db.QueryRow(
		`SELECT s.Id, u.Username, s.DeviceId, s.ClientType, s.IPAddress, 
                s.LastActivityDate, s.UserId, u.IsAdmin
         FROM sessions s
         JOIN users u ON s.UserId = u.Id
         WHERE s.Id = ?`,
		sessionId,
	).Scan(&id, &username, &deviceId, &clientType, &ipAddress,
	     &lastActivity, &userId, &isAdmin)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Session not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to query session", http.StatusInternalServerError)
		return
	}

	session := map[string]interface{}{
		"Id":                      id,
		"UserId":                  userId,
		"ApplicationName":         username,
		"DeviceId":                deviceId,
		"DeviceName":              clientType,
		"Client":                  clientType,
		"LastActivityDate":        lastActivity.Format(time.RFC3339Nano),
		"ServerId":                "kabletown-main",
		"SupportsMediaControl":    true,
		"SupportsRemoteControl":   true,
		"CanBeHidden":             false,
		"HasCustomDeviceName":     false,
		"IsVisible":               true,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(session)
}

// Helper methods
func (h *AuthHandler) recordDevice(userId, deviceId, deviceName, app string) error {
	var count int
	err := h.db.QueryRow(
		"SELECT COUNT(*) FROM devices WHERE device_id = ?",
		deviceId,
	).Scan(&count)

	if err != nil {
		return err
	}

	if count == 0 {
		_, err = h.db.Exec(
			`INSERT INTO devices (id, user_id, device_id, device_name, device_type, created_at) 
             VALUES (?, ?, ?, ?, ?, ?)`,
			generateUUID(), userId, deviceId, deviceName, app, time.Now().UTC(),
		)
		return err
	}

	_, err = h.db.Exec(
		"UPDATE devices SET last_activity_date = ? WHERE device_id = ?",
		time.Now().UTC(), deviceId,
	)
	return err
}

func (h *AuthHandler) storeAPIKey(userId, token, deviceId, deviceName string) error {
	now := time.Now().UTC()
	return h.withinTransaction(func(tx *sql.Tx) error {
		// Revoke existing keys for this device
		tx.Exec(
			"UPDATE api_keys SET is_active = 0 WHERE user_id = ? AND device_id = ?",
			userId, deviceId,
		)

		// Insert new key
		_, err := tx.Exec(
			`INSERT INTO api_keys (id, user_id, api_key_hash, device_id, device_name, created_at, is_active)
             VALUES (?, ?, ?, ?, ?, ?, 1)`,
			generateUUID(), userId, auth.HashToken(token), deviceId, deviceName, now,
		)
		return err
	})
}

func (h *AuthHandler) withinTransaction(fn func(*sql.Tx) error) error {
	tx, err := h.db.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}

	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func generateUUID() string {
	return "00000000-0000-0000-0000-000000000000" // TODO: Generate real UUID
}
