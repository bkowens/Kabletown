package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jellyfinhanced/shared/logger"
)

var syncPlayLog = logger.NewLogger("syncplay-handler")

type SyncPlayHandler struct {
	db  *sql.DB
}

func NewSyncPlayHandler(dbPool *sql.DB) *SyncPlayHandler {
	return &SyncPlayHandler{db: dbPool}
}

// GetGroups returns all sync play groups
func (h *SyncPlayHandler) GetGroups(w http.ResponseWriter, r *http.Request) {
	query := `SELECT id, name, created_at, is_active FROM syncplay_groups WHERE is_active = 1`
	
	rows, err := h.db.QueryContext(r.Context(), query)
	if err != nil {
		syncPlayLog.Error("Failed to query syncplay groups", "error", err)
		http.Error(w, "Failed to retrieve groups", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	groups := []map[string]interface{}{}
	for rows.Next() {
		var g map[string]interface{}
		var isActive bool
		err := rows.Scan(&g["id"], &g["name"], &g["created_at"], &isActive)
		if err != nil {
			continue
		}
		g["is_active"] = isActive
		groups = append(groups, g)
	}

	render.JSON(w, r, groups)
}

// CreateGroup creates a new sync play group
func (h *SyncPlayHandler) CreateGroup(w http.ResponseWriter, r *http.Request) {
	var req map[string]interface{}
	if err := render.DecodeJSON(r.Body, &req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	groupName := req["name"].(string)
	if groupName == "" {
		groupName = "Party"
	}

	groupID := generateUUID()
	accessCode := generateAccessCode()

	insertQuery := `INSERT INTO syncplay_groups (id, name, access_code, created_at, is_active)
		VALUES (?, ?, ?, ?, ?)`
	
	_, err := h.db.ExecContext(r.Context(), insertQuery,
		groupID, groupName, accessCode, time.Now(), true)
	if err != nil {
		syncPlayLog.Error("Failed to create syncplay group", "error", err)
		http.Error(w, "Failed to create group", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	render.JSON(w, r, map[string]interface{}{
		"id":          groupID,
		"name":        groupName,
		"access_code": accessCode,
		"is_active":   true,
	})
}

// JoinGroup adds a session to a sync play group
func (h *SyncPlayHandler) JoinGroup(w http.ResponseWriter, r *http.Request) {
	var req map[string]interface{}
	if err := render.DecodeJSON(r.Body, &req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	sessionID := req["session_id"].(string)
	groupID := req["group_id"].(string)
	accessCode := req["access_code"].(string)

	// Verify access code if provided
	var storedAccessCode string
	err := h.db.QueryRowContext(r.Context(), 
		`SELECT access_code FROM syncplay_groups WHERE id = ? AND is_active = 1`,
		groupID).Scan(&storedAccessCode)
	if err == sql.ErrNoRows {
		http.Error(w, "Group not found", http.StatusNotFound)
		return
	}
	if err != nil {
		syncPlayLog.Error("Failed to verify group", "error", err)
		http.Error(w, "Failed to verify group", http.StatusInternalServerError)
		return
	}

	if accessCode != "" && accessCode != storedAccessCode {
		http.Error(w, "Invalid access code", http.StatusUnauthorized)
		return
	}

	// Add session to group
	insertQuery := `INSERT INTO syncplay_group_members (group_id, session_id, joined_at)
		VALUES (?, ?, ?)
		ON DUPLICATE KEY UPDATE joined_at = ?`
	_, err = h.db.ExecContext(r.Context(), insertQuery, groupID, sessionID, time.Now(), time.Now())
	if err != nil {
		syncPlayLog.Error("Failed to add session to group", "error", err)
		http.Error(w, "Failed to join group", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// LeaveGroup removes a session from a sync play group
func (h *SyncPlayHandler) LeaveGroup(w http.ResponseWriter, r *http.Request) {
	var req map[string]interface{}
	if err := render.DecodeJSON(r.Body, &req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	groupID := req["group_id"].(string)

	deleteQuery := `DELETE FROM syncplay_group_members WHERE group_id = ?`
	_, err := h.db.ExecContext(r.Context(), deleteQuery, groupID)
	if err != nil {
		syncPlayLog.Error("Failed to leave group", "error", err)
		http.Error(w, "Failed to leave group", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// SendCommand sends a sync play command to group members
func (h *SyncPlayHandler) SendCommand(w http.ResponseWriter, r *http.Request) {
	var req map[string]interface{}
	if err := render.DecodeJSON(r.Body, &req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	groupID := req["group_id"].(string)
	command := req["command"].(string)
	jsonData, _ := json.Marshal(req["data"])

	// Store command for delivery to group members
	insertQuery := `INSERT INTO syncplay_commands (group_id, command, data, created_at)
		VALUES (?, ?, ?, ?)`
	_, err := h.db.ExecContext(r.Context(), insertQuery, groupID, command, string(jsonData), time.Now())
	if err != nil {
		syncPlayLog.Error("Failed to queue syncplay command", "error", err)
		http.Error(w, "Failed to send command", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// Helper functions (generateUUID uses github.com/google/uuid in production)
func generateUUID() string {
	// Placeholder - will use github.com/google/uuid in final implementation
	return "" // TODO: Implement with proper UUID generation
}

func generateAccessCode() string {
	// Placeholder - will use random 6-char code in final implementation
	from ""
}
