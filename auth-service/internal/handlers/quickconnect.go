package handlers

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/jellyfinhanced/auth-service/internal/dto"
	"github.com/jellyfinhanced/shared/auth"
	"github.com/jellyfinhanced/shared/response"
)

// quickConnectEntry holds an in-memory QuickConnect session.
type quickConnectEntry struct {
	Secret        string
	Code          string
	DeviceID      string
	Authenticated bool
	AuthUserID    string
	ExpiresAt     time.Time
}

// quickConnectStore is the package-level in-memory store for QuickConnect sessions.
var quickConnectStore sync.Map

// QuickConnectEnabled handles GET /QuickConnect/Enabled.
func (h *Handler) QuickConnectEnabled(w http.ResponseWriter, r *http.Request) {
	response.WriteJSON(w, http.StatusOK, true)
}

// QuickConnectInitiate handles POST /QuickConnect/Initiate.
func (h *Handler) QuickConnectInitiate(w http.ResponseWriter, r *http.Request) {
	secret := uuid.New().String()
	code := fmt.Sprintf("%06d", rand.Intn(1_000_000)) //nolint:gosec

	deviceID := auth.GetDeviceIDAsGUID(r.Context())
	if deviceID == "" {
		deviceID = uuid.New().String()
	}

	entry := quickConnectEntry{
		Secret:    secret,
		Code:      code,
		DeviceID:  deviceID,
		ExpiresAt: time.Now().Add(15 * time.Minute),
	}
	quickConnectStore.Store(secret, entry)

	response.WriteJSON(w, http.StatusOK, map[string]string{
		"Secret":   secret,
		"Code":     code,
		"DeviceId": deviceID,
	})
}

// QuickConnectConnect handles POST /QuickConnect/Connect?secret=<secret>.
func (h *Handler) QuickConnectConnect(w http.ResponseWriter, r *http.Request) {
	secret := r.URL.Query().Get("secret")
	if secret == "" {
		response.WriteError(w, http.StatusBadRequest, "secret query parameter is required")
		return
	}

	raw, ok := quickConnectStore.Load(secret)
	if !ok {
		response.WriteError(w, http.StatusNotFound, "QuickConnect secret not found")
		return
	}

	entry := raw.(quickConnectEntry)
	if time.Now().After(entry.ExpiresAt) {
		quickConnectStore.Delete(secret)
		response.WriteError(w, http.StatusGone, "QuickConnect secret has expired")
		return
	}

	response.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"Secret":        entry.Secret,
		"Code":          entry.Code,
		"Authenticated": entry.Authenticated,
	})
}

// QuickConnectAuthorize handles POST /QuickConnect/Authorize?code=<code>.
func (h *Handler) QuickConnectAuthorize(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		response.WriteError(w, http.StatusBadRequest, "code query parameter is required")
		return
	}

	callerID := auth.GetUserIDAsGUID(r.Context())
	if callerID == "" {
		response.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var found bool
	quickConnectStore.Range(func(key, value interface{}) bool {
		entry := value.(quickConnectEntry)
		if entry.Code != code {
			return true
		}
		if time.Now().After(entry.ExpiresAt) {
			quickConnectStore.Delete(key)
			return true
		}
		entry.Authenticated = true
		entry.AuthUserID = callerID
		quickConnectStore.Store(key, entry)
		found = true
		return false
	})

	if !found {
		response.WriteError(w, http.StatusNotFound, "QuickConnect code not found or expired")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// AuthenticateWithQuickConnect handles GET /Users/AuthenticateWithQuickConnect.
// Body: {"Secret":"..."}
func (h *Handler) AuthenticateWithQuickConnect(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Secret string `json:"Secret"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Secret == "" {
		response.WriteError(w, http.StatusBadRequest, "Secret is required")
		return
	}

	raw, ok := quickConnectStore.Load(req.Secret)
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "QuickConnect secret not found")
		return
	}

	entry := raw.(quickConnectEntry)
	if time.Now().After(entry.ExpiresAt) {
		quickConnectStore.Delete(req.Secret)
		response.WriteError(w, http.StatusUnauthorized, "QuickConnect secret has expired")
		return
	}
	if !entry.Authenticated {
		response.WriteError(w, http.StatusUnauthorized, "QuickConnect has not been authorized yet")
		return
	}

	user, err := h.userRepo.GetUserByID(entry.AuthUserID)
	if err != nil || user == nil {
		response.WriteError(w, http.StatusUnauthorized, "authorized user not found")
		return
	}
	if user.IsDisabled {
		response.WriteError(w, http.StatusUnauthorized, "User account is disabled")
		return
	}

	token, err := generateToken()
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	var clientName, deviceName, deviceID, appVersion string
	if hdr := r.Header.Get("X-Emby-Authorization"); hdr != "" {
		if hf, parseErr := auth.ParseMediaBrowserHeader(hdr); parseErr == nil && hf != nil {
			deviceID = hf.DeviceID
			deviceName = hf.Device
			clientName = hf.Client
			appVersion = hf.Version
		}
	}
	if deviceID == "" {
		deviceID = entry.DeviceID
	}
	if deviceName == "" {
		deviceName = "QuickConnect"
	}
	if clientName == "" {
		clientName = "QuickConnect"
	}

	if err := h.upsertDevice(user.Id, deviceID, deviceName, clientName, appVersion, token); err != nil {
		response.WriteError(w, http.StatusInternalServerError, "failed to create session")
		return
	}

	quickConnectStore.Delete(req.Secret)

	result := dto.AuthenticationResult{
		User:        userToDto(user, h.serverID),
		AccessToken: token,
		ServerId:    h.serverID,
	}
	response.WriteJSON(w, http.StatusOK, result)
}
