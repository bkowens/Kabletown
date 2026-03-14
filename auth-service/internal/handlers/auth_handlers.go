package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jellyfinhanced/auth-service/internal/db"
	"github.com/jellyfinhanced/auth-service/internal/dto"
	"github.com/jellyfinhanced/shared/auth"
	"github.com/jellyfinhanced/shared/response"
)

// AuthenticateByName handles POST /Users/AuthenticateByName.
// Body: {"Username":"...", "Password":"...", "PasswordMd5":"..."}
func (h *Handler) AuthenticateByName(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username    string `json:"Username"`
		Password    string `json:"Password"`
		PasswordMd5 string `json:"PasswordMd5"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Username == "" {
		response.WriteError(w, http.StatusBadRequest, "Username is required")
		return
	}

	// Extract auth header details
	authHeader := r.Header.Get("X-Emby-Authorization")
	hf, _ := auth.ParseMediaBrowserHeader(authHeader)
	var deviceID, deviceName, clientName, appVersion string
	if hf != nil {
		deviceID = hf.DeviceID
		deviceName = hf.Device
		clientName = hf.Client
		appVersion = hf.Version
	}
	if deviceID == "" {
		deviceID = uuid.New().String()
	}

	// Get user by name
	user, err := h.userRepo.GetUserByName(req.Username)
	if err != nil || user == nil {
		response.WriteError(w, http.StatusUnauthorized, "Invalid username or password")
		return
	}

	// Verify password
	var hashToCheck string
	if req.Password != "" {
		hashToCheck = h.hashPassword(req.Password)
	} else if req.PasswordMd5 != "" {
		hashToCheck = strings.ToLower(req.PasswordMd5)
	} else {
		response.WriteError(w, http.StatusBadRequest, "Password or PasswordMd5 is required")
		return
	}

	if hashToCheck != user.Password {
		response.WriteError(w, http.StatusUnauthorized, "Invalid username or password")
		return
	}

	if user.IsDisabled {
		response.WriteError(w, http.StatusUnauthorized, "User is disabled")
		return
	}

	// Generate device token
	token, err := generateToken()
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	// Upsert device
	if err := h.upsertDevice(user.Id, deviceID, deviceName, clientName, appVersion, token); err != nil {
		response.WriteError(w, http.StatusInternalServerError, "failed to create session")
		return
	}

	result := dto.AuthenticationResult{
		User:        userToDto(user, h.serverID),
		AccessToken: token,
		ServerId:    h.serverID,
	}
	response.WriteJSON(w, http.StatusOK, result)
}

// ValidateUser handles GET /Users/{userId}/Authorize.
func (h *Handler) ValidateUser(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userId")
	if userID == "" {
		response.WriteError(w, http.StatusBadRequest, "userId is required")
		return
	}

	user, err := h.userRepo.GetUserByID(userID)
	if err != nil || user == nil {
		response.WriteError(w, http.StatusNotFound, "User not found")
		return
	}

	response.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"Id":          user.Id,
		"Name":        user.Name,
		"isAdmin":     true,
		"hasPassword": true,
	})
}

// upsertDevice inserts or updates a device/session record for the authenticated user.
func (h *Handler) upsertDevice(userID, deviceID, deviceName, clientName, appVersion, token string) error {
	return h.db.QueryRowx(
		`INSERT INTO Devices (UserId, DeviceId, DeviceName, AppName, AppVersion, AccessToken, DateCreated, DateLastActivity)
		 VALUES (?, ?, ?, ?, ?, ?, NOW(), NOW())
		 ON DUPLICATE KEY UPDATE
		   DeviceName = VALUES(DeviceName),
		   AppVersion = VALUES(AppVersion),
		   AccessToken = VALUES(AccessToken),
		   DateLastActivity = NOW()`,
		userID, deviceID, deviceName, clientName, appVersion, token,
	).Err()
}

// userToDto converts an internal db.User to a public UserDto.
func userToDto(u *db.User, serverID string) *dto.UserDto {
	return &dto.UserDto{
		Id:                        u.Id,
		Name:                      u.Name,
		ServerId:                  serverID,
		HasPassword:               u.Password != "",
		HasConfiguredPassword:     u.Password != "",
		HasConfiguredEasyPassword: false,
		EnableAutoLogin:           false,
		IsAdministrator:           u.IsAdmin,
		IsDisabled:                u.IsDisabled,
		IsHidden:                  u.IsHidden,
	}
}

// hashPassword returns the SHA256 hash of the password (lowercase hex).
func (h *Handler) hashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}
