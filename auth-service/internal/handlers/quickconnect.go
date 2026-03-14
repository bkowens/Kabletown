package handlers

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/bowens/kabletown/auth-service/internal/dto"
	"github.com/bowens/kabletown/shared/auth"
	"github.com/bowens/kabletown/shared/response"
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
// Always returns true — QuickConnect is always available.
func (h *Handler) QuickConnectEnabled(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, http.StatusOK, true)
}

// QuickConnectInitiate handles POST /QuickConnect/Initiate.
// Generates a Secret (UUID) and a 6-digit Code, stores the entry with a 15-minute TTL,
// and returns the entry details to the caller.
func (h *Handler) QuickConnectInitiate(w http.ResponseWriter, r *http.Request) {
	secret := uuid.New().String()
	code := fmt.Sprintf("%06d", rand.Intn(1_000_000)) //nolint:gosec

	// Capture optional DeviceId from the authorization header.
	deviceID := auth.GetDeviceID(r)
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

	response.JSON(w, http.StatusOK, map[string]string{
		"Secret":   secret,
		"Code":     code,
		"DeviceId": deviceID,
	})
}

// QuickConnectConnect handles POST /QuickConnect/Connect?secret=<secret>.
// Looks up the entry by secret and returns its current authentication status.
func (h *Handler) QuickConnectConnect(w http.ResponseWriter, r *http.Request) {
	secret := r.URL.Query().Get("secret")
	if secret == "" {
		response.Error(w, http.StatusBadRequest, "secret query parameter is required")
		return
	}

	raw, ok := quickConnectStore.Load(secret)
	if !ok {
		response.Error(w, http.StatusNotFound, "QuickConnect secret not found")
		return
	}

	entry := raw.(quickConnectEntry)
	if time.Now().After(entry.ExpiresAt) {
		quickConnectStore.Delete(secret)
		response.Error(w, http.StatusGone, "QuickConnect secret has expired")
		return
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"Secret":        entry.Secret,
		"Code":          entry.Code,
		"Authenticated": entry.Authenticated,
	})
}

// QuickConnectAuthorize handles POST /QuickConnect/Authorize?code=<code>.
// Requires authentication. Marks the QuickConnect entry identified by code as
// authenticated with the calling user's ID.
func (h *Handler) QuickConnectAuthorize(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		response.Error(w, http.StatusBadRequest, "code query parameter is required")
		return
	}

	callerID := auth.GetUserID(r)
	if callerID == "" {
		response.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var found bool
	quickConnectStore.Range(func(key, value interface{}) bool {
		entry := value.(quickConnectEntry)
		if entry.Code != code {
			return true // continue
		}
		if time.Now().After(entry.ExpiresAt) {
			quickConnectStore.Delete(key)
			return true
		}
		entry.Authenticated = true
		entry.AuthUserID = callerID
		quickConnectStore.Store(key, entry)
		found = true
		return false // stop
	})

	if !found {
		response.Error(w, http.StatusNotFound, "QuickConnect code not found or expired")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// AuthenticateWithQuickConnect handles POST /Users/AuthenticateWithQuickConnect.
// Body: {"Secret":"..."}
// If the secret is authenticated, creates a device token and returns an AuthenticationResult.
func (h *Handler) AuthenticateWithQuickConnect(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Secret string `json:"Secret"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Secret == "" {
		response.Error(w, http.StatusBadRequest, "Secret is required")
		return
	}

	raw, ok := quickConnectStore.Load(req.Secret)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "QuickConnect secret not found")
		return
	}

	entry := raw.(quickConnectEntry)
	if time.Now().After(entry.ExpiresAt) {
		quickConnectStore.Delete(req.Secret)
		response.Error(w, http.StatusUnauthorized, "QuickConnect secret has expired")
		return
	}
	if !entry.Authenticated {
		response.Error(w, http.StatusUnauthorized, "QuickConnect has not been authorized yet")
		return
	}

	// Fetch the user who authorized the QuickConnect session.
	user, err := h.userRepo.GetUserByID(entry.AuthUserID)
	if err != nil || user == nil {
		response.Error(w, http.StatusUnauthorized, "authorized user not found")
		return
	}
	if user.IsDisabled {
		response.Error(w, http.StatusUnauthorized, "User account is disabled")
		return
	}

	token, err := generateToken()
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	// Derive client metadata from the request header; fall back to QuickConnect defaults.
	var clientName, deviceName, deviceID, appVersion string
	if hdr := r.Header.Get("X-Emby-Authorization"); hdr != "" {
		_, deviceID, clientName, deviceName, appVersion, _ = auth.ParseMediaBrowserHeader(hdr)
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
		response.Error(w, http.StatusInternalServerError, "failed to create session")
		return
	}

	// Consume the secret — one-time use.
	quickConnectStore.Delete(req.Secret)

	result := dto.AuthenticationResult{
		User:        userToDto(user, h.serverID),
		AccessToken: token,
		ServerId:    h.serverID,
	}
	response.JSON(w, http.StatusOK, result)
}

// chi.URLParam is used in QuickConnectAuthorize when the code arrives as a URL segment
// rather than a query param. The route registers it as a query param per the spec, but
// the import is still needed for other handlers; keep it referenced.
var _ = chi.URLParam
