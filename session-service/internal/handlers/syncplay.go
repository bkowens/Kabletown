package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/jellyfinhanced/shared/response"
)

// SyncPlayHandler handles SyncPlay group requests.
type SyncPlayHandler struct {
	db *sql.DB
}

// NewSyncPlayHandler creates a new SyncPlayHandler.
func NewSyncPlayHandler(dbPool *sql.DB) *SyncPlayHandler {
	return &SyncPlayHandler{db: dbPool}
}

// GetGroups returns all active sync play groups.
func (h *SyncPlayHandler) GetGroups(w http.ResponseWriter, r *http.Request) {
	query := `SELECT id, name, created_at, is_active FROM syncplay_groups WHERE is_active = 1`

	rows, err := h.db.QueryContext(r.Context(), query)
	if err != nil {
		log.Printf("syncplay-handler: GetGroups: query failed: %v", err)
		response.WriteError(w, http.StatusInternalServerError, "Failed to retrieve groups")
		return
	}
	defer rows.Close()

	groups := []map[string]interface{}{}
	for rows.Next() {
		var id, name string
		var createdAt time.Time
		var isActive bool
		if err := rows.Scan(&id, &name, &createdAt, &isActive); err != nil {
			continue
		}
		groups = append(groups, map[string]interface{}{
			"Id":        id,
			"Name":      name,
			"CreatedAt": createdAt,
			"IsActive":  isActive,
		})
	}

	response.WriteJSON(w, http.StatusOK, groups)
}

// CreateGroup creates a new sync play group.
func (h *SyncPlayHandler) CreateGroup(w http.ResponseWriter, r *http.Request) {
	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	groupName := "Party"
	if name, ok := req["Name"].(string); ok && name != "" {
		groupName = name
	}

	groupID := generateGroupID()
	accessCode := generateAccessCode()

	insertQuery := `INSERT INTO syncplay_groups (id, name, access_code, created_at, is_active)
		VALUES (?, ?, ?, ?, ?)`

	_, err := h.db.ExecContext(r.Context(), insertQuery,
		groupID, groupName, accessCode, time.Now(), true)
	if err != nil {
		log.Printf("syncplay-handler: CreateGroup: insert failed: %v", err)
		response.WriteError(w, http.StatusInternalServerError, "Failed to create group")
		return
	}

	response.WriteJSON(w, http.StatusCreated, map[string]interface{}{
		"Id":         groupID,
		"Name":       groupName,
		"AccessCode": accessCode,
		"IsActive":   true,
	})
}

// JoinGroup adds a session to a sync play group.
func (h *SyncPlayHandler) JoinGroup(w http.ResponseWriter, r *http.Request) {
	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	sessionID, _ := req["SessionId"].(string)
	groupID, _ := req["GroupId"].(string)
	accessCode, _ := req["AccessCode"].(string)

	var storedAccessCode string
	err := h.db.QueryRowContext(r.Context(),
		`SELECT access_code FROM syncplay_groups WHERE id = ? AND is_active = 1`,
		groupID).Scan(&storedAccessCode)
	if err == sql.ErrNoRows {
		response.WriteError(w, http.StatusNotFound, "Group not found")
		return
	}
	if err != nil {
		log.Printf("syncplay-handler: JoinGroup: query failed: %v", err)
		response.WriteError(w, http.StatusInternalServerError, "Failed to verify group")
		return
	}

	if accessCode != "" && accessCode != storedAccessCode {
		response.WriteError(w, http.StatusUnauthorized, "Invalid access code")
		return
	}

	insertQuery := `INSERT INTO syncplay_group_members (group_id, session_id, joined_at)
		VALUES (?, ?, ?)
		ON DUPLICATE KEY UPDATE joined_at = ?`
	_, err = h.db.ExecContext(r.Context(), insertQuery, groupID, sessionID, time.Now(), time.Now())
	if err != nil {
		log.Printf("syncplay-handler: JoinGroup: insert failed: %v", err)
		response.WriteError(w, http.StatusInternalServerError, "Failed to join group")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// LeaveGroup removes a session from a sync play group.
func (h *SyncPlayHandler) LeaveGroup(w http.ResponseWriter, r *http.Request) {
	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	groupID, _ := req["GroupId"].(string)

	deleteQuery := `DELETE FROM syncplay_group_members WHERE group_id = ?`
	_, err := h.db.ExecContext(r.Context(), deleteQuery, groupID)
	if err != nil {
		log.Printf("syncplay-handler: LeaveGroup: delete failed: %v", err)
		response.WriteError(w, http.StatusInternalServerError, "Failed to leave group")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// SendCommand sends a sync play command to group members.
func (h *SyncPlayHandler) SendCommand(w http.ResponseWriter, r *http.Request) {
	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	groupID, _ := req["GroupId"].(string)
	command, _ := req["Command"].(string)
	jsonData, _ := json.Marshal(req["Data"])

	insertQuery := `INSERT INTO syncplay_commands (group_id, command, data, created_at)
		VALUES (?, ?, ?, ?)`
	_, err := h.db.ExecContext(r.Context(), insertQuery, groupID, command, string(jsonData), time.Now())
	if err != nil {
		log.Printf("syncplay-handler: SendCommand: insert failed: %v", err)
		response.WriteError(w, http.StatusInternalServerError, "Failed to send command")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func generateGroupID() string {
	return fmt.Sprintf("%016x", rand.Int63())
}

func generateAccessCode() string {
	const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 6)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}
	return string(b)
}