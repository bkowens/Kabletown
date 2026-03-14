package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/jellyfinhanced/auth-service/internal/db"
	"github.com/jellyfinhanced/shared/response"
)

// StartupConfiguration handles GET /Startup/Configuration.
// Returns the default startup configuration values expected by Jellyfin clients.
func (h *Handler) StartupConfiguration(w http.ResponseWriter, r *http.Request) {
	response.WriteJSON(w, http.StatusOK, map[string]string{
		"UICulture":                    "en-US",
		"MetadataCountryCode":          "US",
		"PreferredMetadataLanguage":    "en",
	})
}

// StartupComplete handles POST /Startup/Complete.
// Returns 204 No Content — the wizard is considered done.
func (h *Handler) StartupComplete(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

// GetRemoteAccess handles GET /Startup/RemoteAccess.
func (h *Handler) GetRemoteAccess(w http.ResponseWriter, r *http.Request) {
	response.WriteJSON(w, http.StatusOK, map[string]bool{
		"EnableRemoteAccess": true,
		"EnableUPnP":         false,
	})
}

// SetRemoteAccess handles POST /Startup/RemoteAccess.
// Accepts a JSON body and returns 204 No Content.
func (h *Handler) SetRemoteAccess(w http.ResponseWriter, r *http.Request) {
	// Drain the body to be well-behaved; ignore the content.
	var body interface{}
	json.NewDecoder(r.Body).Decode(&body) //nolint:errcheck
	w.WriteHeader(http.StatusNoContent)
}

// GetStartupUser handles GET /Startup/User.
// Queries the first admin user and returns their Name with an empty Password field.
func (h *Handler) GetStartupUser(w http.ResponseWriter, r *http.Request) {
	var id, name string
	err := h.db.QueryRow(
		`SELECT Id, Name FROM Users
		 WHERE Id IN (SELECT UserId FROM Permissions WHERE Kind = 0 AND Value = 1)
		 LIMIT 1`,
	).Scan(&id, &name)

	if err != nil {
		// No admin user exists yet — return empty object so the wizard can proceed.
		response.WriteJSON(w, http.StatusOK, map[string]string{
			"Name":     "",
			"Password": "",
		})
		return
	}

	response.WriteJSON(w, http.StatusOK, map[string]string{
		"Name":     name,
		"Password": "",
	})
}

// SetStartupUser handles POST /Startup/User.
// Body: {"Name":"admin","Password":"password123"}
// Upserts the user with a bcrypt-11 hash and grants admin permission. Returns 204.
func (h *Handler) SetStartupUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name     string `json:"Name"`
		Password string `json:"Password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" {
		response.WriteError(w, http.StatusBadRequest, "Name is required")
		return
	}

	hash, err := db.HashPassword(req.Password)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "failed to hash password")
		return
	}

	// Check whether the user already exists.
	var existingID string
	_ = h.db.QueryRow(`SELECT Id FROM Users WHERE Name = ?`, req.Name).Scan(&existingID)

	if existingID != "" {
		// Update existing user's password.
		if _, err := h.db.Exec(
			`UPDATE Users SET Password = ? WHERE Id = ?`,
			hash, existingID,
		); err != nil {
			response.WriteError(w, http.StatusInternalServerError, "failed to update user")
			return
		}
		// Ensure the admin permission row exists.
		h.db.Exec( //nolint:errcheck
			`INSERT INTO Permissions (UserId, Kind, Value)
			 VALUES (?, 0, 1)
			 ON DUPLICATE KEY UPDATE Value = 1`,
			existingID,
		)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Create a new admin user.
	newID, genErr := generateToken()
	if genErr != nil {
		response.WriteError(w, http.StatusInternalServerError, "failed to generate user id")
		return
	}

	tx, err := h.db.Begin()
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "failed to start transaction")
		return
	}

	if _, err := tx.Exec(
		`INSERT INTO Users (Id, Name, Password, IsDisabled, IsHidden, PrimaryImageTag)
		 VALUES (?, ?, ?, 0, 0, '')`,
		newID, req.Name, hash,
	); err != nil {
		tx.Rollback() //nolint:errcheck
		response.WriteError(w, http.StatusInternalServerError, "failed to create user")
		return
	}

	if _, err := tx.Exec(
		`INSERT INTO Permissions (UserId, Kind, Value) VALUES (?, 0, 1)`,
		newID,
	); err != nil {
		tx.Rollback() //nolint:errcheck
		response.WriteError(w, http.StatusInternalServerError, "failed to set admin permission")
		return
	}

	if err := tx.Commit(); err != nil {
		response.WriteError(w, http.StatusInternalServerError, "failed to commit transaction")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
