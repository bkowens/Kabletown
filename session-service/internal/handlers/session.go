package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-sql-driver/mysql"

	"github.com/jellyfinhanced/shared/auth"
	"github.com/jellyfinhanced/shared/response"
	"github.com/jellyfinhanced/session-service/internal/models"
)

// SessionHandler handles session-related requests.
type SessionHandler struct {
	db *sql.DB
}

// NewSessionHandler creates a new SessionHandler.
func NewSessionHandler(dbPool *sql.DB) *SessionHandler {
	return &SessionHandler{db: dbPool}
}

// GetSessions returns all active sessions.
func (h *SessionHandler) GetSessions(w http.ResponseWriter, r *http.Request) {
	userID := func() string { info, ok := auth.GetAuth(r.Context()); if !ok || info == nil { return "" }; return info.UserID.String() }()
	query := `SELECT id, user_id, device_id, app_name, device_name, client, last_activity_date
		FROM sessions WHERE user_id = ? OR user_id IS NULL ORDER BY last_activity_date DESC`

	rows, err := h.db.QueryContext(r.Context(), query, userID)
	if err != nil {
		log.Printf("session-handler: GetSessions: query failed: %v", err)
		response.WriteError(w, http.StatusInternalServerError, "Failed to retrieve sessions")
		return
	}
	defer rows.Close()

	sessions := []models.Session{}
	for rows.Next() {
		var s models.Session
		var uid sql.NullString
		if err := rows.Scan(&s.ID, &uid, &s.DeviceID, &s.AppName, &s.DeviceName, &s.Client, &s.LastActivityDate); err != nil {
			response.WriteError(w, http.StatusInternalServerError, "Failed to scan session")
			return
		}
		if uid.Valid {
			s.UserID = uid.String
		}
		sessions = append(sessions, s)
	}

	response.WriteJSON(w, http.StatusOK, sessions)
}

// CreateSession creates a new session.
func (h *SessionHandler) CreateSession(w http.ResponseWriter, r *http.Request) {
	var req models.CreateSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	userID := func() string { info, ok := auth.GetAuth(r.Context()); if !ok || info == nil { return "" }; return info.UserID.String() }()
	if userID == "" {
		response.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	sessionID := req.DeviceID

	insertQuery := `INSERT INTO sessions (id, user_id, device_id, app_name, device_name, client, last_activity_date)
		VALUES (?, ?, ?, ?, ?, ?, ?)`

	_, err := h.db.ExecContext(r.Context(), insertQuery,
		sessionID, userID, req.DeviceID, req.AppName, req.DeviceName, req.Client, time.Now())
	if err != nil {
		if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == 1062 {
			updateQuery := `UPDATE sessions SET last_activity_date = ?, device_name = ?, client = ? WHERE id = ?`
			_, err = h.db.ExecContext(r.Context(), updateQuery, time.Now(), req.DeviceName, req.Client, sessionID)
			if err != nil {
				log.Printf("session-handler: CreateSession: update failed: %v", err)
				response.WriteError(w, http.StatusInternalServerError, "Failed to update session")
				return
			}
		} else {
			log.Printf("session-handler: CreateSession: insert failed: %v", err)
			response.WriteError(w, http.StatusInternalServerError, "Failed to create session")
			return
		}
	}

	response.WriteJSON(w, http.StatusCreated, map[string]string{"Id": sessionID})
}

// GetSession returns a specific session.
func (h *SessionHandler) GetSession(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "sessionId")

	query := `SELECT id, user_id, device_id, app_name, device_name, client, last_activity_date
		FROM sessions WHERE id = ?`

	var s models.Session
	var uid sql.NullString
	err := h.db.QueryRowContext(r.Context(), query, sessionID).Scan(
		&s.ID, &uid, &s.DeviceID, &s.AppName, &s.DeviceName, &s.Client, &s.LastActivityDate)
	if err == sql.ErrNoRows {
		response.WriteError(w, http.StatusNotFound, "Session not found")
		return
	}
	if err != nil {
		log.Printf("session-handler: GetSession: query failed: %v", err)
		response.WriteError(w, http.StatusInternalServerError, "Failed to retrieve session")
		return
	}
	if uid.Valid {
		s.UserID = uid.String
	}

	response.WriteJSON(w, http.StatusOK, s)
}

// ReportSessionActivity updates session activity.
func (h *SessionHandler) ReportSessionActivity(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "sessionId")

	updateQuery := `UPDATE sessions SET last_activity_date = ? WHERE id = ?`
	_, err := h.db.ExecContext(r.Context(), updateQuery, time.Now(), sessionID)
	if err != nil {
		log.Printf("session-handler: ReportSessionActivity: update failed: %v", err)
		response.WriteError(w, http.StatusInternalServerError, "Failed to update activity")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ReportPlaying reports playback state (started or progress).
func (h *SessionHandler) ReportPlaying(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "sessionId")

	var req models.PlaybackState
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Minimal update — just touch activity date.
		updateQuery := `UPDATE sessions SET last_activity_date = ? WHERE id = ?`
		h.db.ExecContext(r.Context(), updateQuery, time.Now(), sessionID) //nolint:errcheck
		w.WriteHeader(http.StatusNoContent)
		return
	}

	query := `INSERT INTO playback_state (session_id, item_id, play_position_ticks, is_playing, last_reported)
		VALUES (?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE play_position_ticks = ?, is_playing = ?, last_reported = ?`

	_, err := h.db.ExecContext(r.Context(), query,
		sessionID, req.ItemID, req.PlayPositionTicks, req.IsPlaying, time.Now(),
		req.PlayPositionTicks, req.IsPlaying, time.Now())
	if err != nil {
		log.Printf("session-handler: ReportPlaying: exec failed: %v", err)
		response.WriteError(w, http.StatusInternalServerError, "Failed to report playback")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ReportStopped reports playback stopped.
func (h *SessionHandler) ReportStopped(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "sessionId")

	var req models.PlaybackState
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	query := `UPDATE playback_state SET is_playing = ?, last_reported = ? WHERE session_id = ?`
	_, err := h.db.ExecContext(r.Context(), query, false, time.Now(), sessionID)
	if err != nil {
		log.Printf("session-handler: ReportStopped: update failed: %v", err)
		response.WriteError(w, http.StatusInternalServerError, "Failed to report stopped")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// UpdateSessionCapability updates session capability info.
func (h *SessionHandler) UpdateSessionCapability(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "sessionId")

	var req models.SessionCapability
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	query := `UPDATE sessions SET capabilities = ? WHERE id = ?`
	_, err := h.db.ExecContext(r.Context(), query, req.ToJSON(), sessionID)
	if err != nil {
		log.Printf("session-handler: UpdateSessionCapability: update failed: %v", err)
		response.WriteError(w, http.StatusInternalServerError, "Failed to update capability")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// SendMessageToSession sends a message to a specific session.
func (h *SessionHandler) SendMessageToSession(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "sessionId")

	var req models.MessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid message")
		return
	}

	query := `INSERT INTO session_messages (session_id, message_type, header, text, timeout_ms, created_at)
		VALUES (?, ?, ?, ?, ?, ?)`

	_, err := h.db.ExecContext(r.Context(), query,
		sessionID, req.MessageType, req.Header, req.Text, req.TimeoutMS, time.Now())
	if err != nil {
		log.Printf("session-handler: SendMessageToSession: insert failed: %v", err)
		response.WriteError(w, http.StatusInternalServerError, "Failed to send message")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// CloseSession closes the current session (DELETE /Sessions/Logout).
func (h *SessionHandler) CloseSession(w http.ResponseWriter, r *http.Request) {
	userID := func() string { info, ok := auth.GetAuth(r.Context()); if !ok || info == nil { return "" }; return info.UserID.String() }()

	deleteQuery := `DELETE FROM sessions WHERE user_id = ?`
	_, err := h.db.ExecContext(r.Context(), deleteQuery, userID)
	if err != nil {
		log.Printf("session-handler: CloseSession: delete failed: %v", err)
		response.WriteError(w, http.StatusInternalServerError, "Failed to delete session")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// CloseSpecificSession closes a specific session.
func (h *SessionHandler) CloseSpecificSession(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "sessionId")

	deleteQuery := `DELETE FROM sessions WHERE id = ?`
	result, err := h.db.ExecContext(r.Context(), deleteQuery, sessionID)
	if err != nil {
		log.Printf("session-handler: CloseSpecificSession: delete failed: %v", err)
		response.WriteError(w, http.StatusInternalServerError, "Failed to delete session")
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		response.WriteError(w, http.StatusNotFound, "Session not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// KeepAlive extends the session timeout.
func (h *SessionHandler) KeepAlive(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "sessionId")

	query := `UPDATE sessions SET last_activity_date = ? WHERE id = ?`
	_, err := h.db.ExecContext(r.Context(), query, time.Now(), sessionID)
	if err != nil {
		log.Printf("session-handler: KeepAlive: update failed: %v", err)
		response.WriteError(w, http.StatusInternalServerError, "Failed to keepalive")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}