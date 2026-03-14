package handlers

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-sql-driver/mysql"
	"github.com/jellyfinhanced/shared/auth"
	"github.com/jellyfinhanced/shared/db"
	"github.com/jellyfinhanced/shared/logger"
	"kabletown/session-service/internal/models"
)

var sessionLog = logger.NewLogger("session-handler")

type SessionHandler struct {
	db   *sql.DB
	ctx  context.Context
}

func NewSessionHandler(dbPool *sql.DB) *SessionHandler {
	return &SessionHandler{
		db:  dbPool,
		ctx: context.Background(),
	}
}

// GetSessions returns all active sessions
func (h *SessionHandler) GetSessions(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserFromContext(r.Context())
	query := `SELECT id, user_id, device_id, app_name, device_name, client, last_activity_date
		FROM sessions WHERE user_id = ? OR user_id IS NULL ORDER BY last_activity_date DESC`
	
	rows, err := h.db.QueryContext(r.Context(), query, userID)
	if err != nil {
		sessionLog.Error("Failed to query sessions", "error", err, "user_id", userID)
		http.Error(w, "Failed to retrieve sessions", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	sessions := []models.Session{}
	for rows.Next() {
		var s models.Session
		var userID sql.NullString
		if err := rows.Scan(&s.ID, &userID, &s.DeviceID, &s.AppName, &s.DeviceName, &s.Client, &s.LastActivityDate); err != nil {
			http.Error(w, "Failed to scan session", http.StatusInternalServerError)
			return
		}
		if userID.Valid {
			s.UserID = userID.String
		}
		sessions = append(sessions, s)
	}

	render.JSON(w, r, sessions)
}

// CreateSession creates a new session
func (h *SessionHandler) CreateSession(w http.ResponseWriter, r *http.Request) {
	var req models.CreateSessionRequest
	if err := render.DecodeJSON(r.Body, &req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	userID := auth.GetUserFromContext(r.Context())
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	sessionID := req.DeviceID // Use device ID as session ID or generate new one
	
	insertQuery := `INSERT INTO sessions (id, user_id, device_id, app_name, device_name, client, last_activity_date) 
		VALUES (?, ?, ?, ?, ?, ?, ?)`
	
	_, err := h.db.ExecContext(r.Context(), insertQuery,
		sessionID, userID, req.DeviceID, req.AppName, req.DeviceName, req.Client, time.Now())
	if err != nil {
		if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == 1062 {
			// Session already exists, update it
			updateQuery := `UPDATE sessions SET last_activity_date = ?, device_name = ?, client = ? WHERE id = ?`
			_, err = h.db.ExecContext(r.Context(), updateQuery, time.Now(), req.DeviceName, req.Client, sessionID)
			if err != nil {
				sessionLog.Error("Failed to update existing session", "error", err)
				http.Error(w, "Failed to update session", http.StatusInternalServerError)
				return
			}
		} else {
			sessionLog.Error("Failed to create session", "error", err)
			http.Error(w, "Failed to create session", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusCreated)
	render.JSON(w, r, map[string]string{"id": sessionID})
}

// GetSession returns a specific session
func (h *SessionHandler) GetSession(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "sessionId")
	user
	query := `SELECT id, user_id, device_id, app_name, device_name, client, last_activity_date
		FROM sessions WHERE id = ?`
	
	var s models.Session
	var userID sql.NullString
	err := h.db.QueryRowContext(r.Context(), query, sessionID).Scan(
		&s.ID, &userID, &s.DeviceID, &s.AppName, &s.DeviceName, &s.Client, &s.LastActivityDate)
	if err == sql.ErrNoRows {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}
	if err != nil {
		sessionLog.Error("Failed to query session", "error", err)
		http.Error(w, "Failed to retrieve session", http.StatusInternalServerError)
		return
	}
	if userID.Valid {
		s.UserID = userID.String
	}

	render.JSON(w, r, s)
}

// ReportSessionActivity updates session activity
func (h *SessionHandler) ReportSessionActivity(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "sessionId")
	
	updateQuery := `UPDATE sessions SET last_activity_date = ? WHERE id = ?`
	_, err := h.db.ExecContext(r.Context(), updateQuery, time.Now(), sessionID)
	if err != nil {
		sessionLog.Error("Failed to update session activity", "error", err)
		http.Error(w, "Failed to update activity", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// ReportViewingItem tracks what item a session is viewing
func (h *SessionHandler) ReportViewingItem(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "sessionId")
	itemID := chi.URLParam(r, "itemId")

	// Insert or update viewing item record
	query := `INSERT INTO session_viewing (session_id, item_id, viewed_at) 
		VALUES (?, ?, ?) ON DUPLICATE KEY UPDATE viewed_at = ?`
	
	_, err := h.db.ExecContext(r.Context(), query, sessionID, itemID, time.Now(), time.Now())
	if err != nil {
		sessionLog.Error("Failed to report viewing item", "error", err)
		http.Error(w, "Failed to report viewing", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// ReportPlaying reports playback state
func (h *SessionHandler) ReportPlaying(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "sessionId")
	
	var req models.PlaybackState
	if err := render.DecodeJSON(r.Body, &req); err != nil {
		// At minimum we have session ID, just update activity
		h.ReportSessionActivity(w, r)
		return
	}

	// Store playback state
	query := `INSERT INTO playback_state (session_id, item_id, play_position_ticks, is_playing, last_reported)
		VALUES (?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE play_position_ticks = ?, is_playing = ?, last_reported = ?`
	
	_, err := h.db.ExecContext(r.Context(), query,
		sessionID, req.ItemID, req.PlayPositionTicks, req.IsPlaying, time.Now(),
		req.PlayPositionTicks, req.IsPlaying, time.Now())
	if err != nil {
		sessionLog.Error("Failed to report playback state", "error", err)
		http.Error(w, "Failed to report playback", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// ReportStopped reports playback stopped
func (h *SessionHandler) ReportStopped(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "sessionId")
	
	var req models.PlaybackState
	if err := render.DecodeJSON(r.Body, &req); err != nil {
		h.ReportSessionActivity(w, r)
		return
	}

	// Update playback state to stopped
	query := `UPDATE playback_state SET is_playing = ?, last_reported = ? WHERE session_id = ?`
	_, err := h.db.ExecContext(r.Context(), query, false, time.Now(), sessionID)
	if err != nil {
		sessionLog.Error("Failed to report stopped state", "error", err)
		http.Error(w, "Failed to report stopped", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// UpdateSessionCapability updates session capability info
func (h *SessionHandler) UpdateSessionCapability(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "sessionId")
	
	var req models.SessionCapability
	if err := render.DecodeJSON(r.Body, &req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Store capability info (simplified)
	query := `UPDATE sessions SET capabilities = ? WHERE id = ?`
	_, err := h.db.ExecContext(r.Context(), query, req.toJSON(), sessionID)
	if err != nil {
		sessionLog.Error("Failed to update capability", "error", err)
		http.Error(w, "Failed to update capability", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// SendMessageToSession sends a message to a specific session
func (h *SessionHandler) SendMessageToSession(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "sessionId")
	
	var req models.MessageRequest
	if err := render.DecodeJSON(r.Body, &req); err != nil {
		http.Error(w, "Invalid message", http.StatusBadRequest)
		return
	}

	// Queue message for delivery (could use WebSocket/Redis PubSub in production)
	query := `INSERT INTO session_messages (session_id, message_type, header, text, timeout_ms, created_at)
		VALUES (?, ?, ?, ?, ?, ?)`
	
	_, err := h.db.ExecContext(r.Context(), query,
		sessionID, req.MessageType, req.Header, req.Text, req.TimeoutMS, time.Now())
	if err != nil {
		sessionLog.Error("Failed to queue message", "error", err)
		http.Error(w, "Failed to send message", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// CloseSession closes the current session
func (h *SessionHandler) CloseSession(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "sessionId")
	
	deleteQuery := `DELETE FROM sessions WHERE id = ?`
	result, err := h.db.ExecContext(r.Context(), deleteQuery, sessionID)
	if err != nil {
		sessionLog.Error("Failed to delete session", "error", err)
		http.Error(w, "Failed to delete session", http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// CloseSpecificSession closes a specific session
func (h *SessionHandler) CloseSpecificSession(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "sessionId")
	
	deleteQuery := `DELETE FROM sessions WHERE id = ?`
	result, err := h.db.ExecContext(r.Context(), deleteQuery, sessionID)
	if err != nil {
		sessionLog.Error("Failed to delete session", "error", err)
		http.Error(w, "Failed to delete session", http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// KeepAlive extends the session timeout
func (h *SessionHandler) KeepAlive(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "sessionId")
	
	query := `UPDATE sessions SET last_activity_date = ? WHERE id = ?`
	_, err := h.db.ExecContext(r.Context(), query, time.Now(), sessionID)
	if err != nil {
		sessionLog.Error("Failed to update session keepalive", "error", err)
		http.Error(w, "Failed to keepalive", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
