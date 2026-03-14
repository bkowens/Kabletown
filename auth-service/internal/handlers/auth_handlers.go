// Package handlers implements the HTTP handlers for the auth-service.
package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/bowens/kabletown/auth-service/internal/db"
	"github.com/bowens/kabletown/auth-service/internal/dto"
	"github.com/bowens/kabletown/shared/auth"
	"github.com/bowens/kabletown/shared/response"
)

// Handler holds the database dependencies for all auth-service route handlers.
type Handler struct {
	db       *sqlx.DB
	resolver *db.TokenResolver
	userRepo *db.UserRepository
	keyRepo  *db.ApiKeyRepository
	serverID string
}

// New creates a Handler wired to the given database connection and server UUID.
func New(sqlxDB *sqlx.DB, serverID string) *Handler {
	return &Handler{
		db:       sqlxDB,
		resolver: db.NewTokenResolver(sqlxDB),
		userRepo: db.NewUserRepository(sqlxDB),
		keyRepo:  db.NewApiKeyRepository(sqlxDB),
		serverID: serverID,
	}
}

// RegisterRoutes wires all auth-service routes onto the given chi router.
// The lookup function is used by NewAuthMiddleware to validate bearer tokens.
func (h *Handler) RegisterRoutes(r chi.Router, lookup auth.DeviceLookupFunc) {
	anonymousPaths := []string{
		"/healthz",
		"/Users/Public",
		"/Users/AuthenticateByName",
		"/Users/AuthenticateWithQuickConnect",
		"/Users/ForgotPassword",
		"/Users/ForgotPasswordPin",
		"/Startup/Configuration",
		"/Startup/Complete",
		"/Startup/RemoteAccess",
		"/Startup/User",
		"/QuickConnect/Enabled",
		"/QuickConnect/Initiate",
		"/QuickConnect/Connect",
		"/System/Info/Public",
		"/Branding/Configuration",
	}

	authMiddleware := auth.NewAuthMiddleware(lookup, anonymousPaths)

	// Anonymous routes (no token required)
	r.Post("/Users/AuthenticateByName", h.AuthenticateByName)
	r.Post("/Users/{userId}/Authenticate", h.AuthenticateUser)
	r.Post("/Users/AuthenticateWithQuickConnect", h.AuthenticateWithQuickConnect)
	r.Get("/Users/Public", h.GetPublicUsers)

	// Startup wizard (anonymous)
	r.Get("/Startup/Configuration", h.StartupConfiguration)
	r.Post("/Startup/Complete", h.StartupComplete)
	r.Get("/Startup/RemoteAccess", h.GetRemoteAccess)
	r.Post("/Startup/RemoteAccess", h.SetRemoteAccess)
	r.Get("/Startup/User", h.GetStartupUser)
	r.Post("/Startup/User", h.SetStartupUser)

	// QuickConnect (anonymous initiation / polling)
	r.Get("/QuickConnect/Enabled", h.QuickConnectEnabled)
	r.Post("/QuickConnect/Initiate", h.QuickConnectInitiate)
	r.Post("/QuickConnect/Connect", h.QuickConnectConnect)

	// Forgot password (anonymous)
	r.Post("/Users/ForgotPassword", h.ForgotPassword)
	r.Post("/Users/ForgotPasswordPin", h.ForgotPasswordPin)

	// Authenticated routes
	r.Group(func(r chi.Router) {
		r.Use(authMiddleware)

		r.Get("/Users", auth.RequireAdmin(http.HandlerFunc(h.GetUsers)).ServeHTTP)
		r.Get("/Users/{userId}", auth.RequireAuth(http.HandlerFunc(h.GetUser)).ServeHTTP)

		r.Get("/Auth/Keys", auth.RequireAdmin(http.HandlerFunc(h.ListApiKeys)).ServeHTTP)
		r.Post("/Auth/Keys", auth.RequireAdmin(http.HandlerFunc(h.CreateApiKey)).ServeHTTP)
		r.Delete("/Auth/Keys/{key}", auth.RequireAdmin(http.HandlerFunc(h.RevokeApiKey)).ServeHTTP)

		// QuickConnect authorize (must be authenticated — caller approves a pending code)
		r.Post("/QuickConnect/Authorize", h.QuickConnectAuthorize)
	})
}

// AuthenticateByName handles POST /Users/AuthenticateByName.
// Request body: {"Name":"admin","Pw":"password123"}
func (h *Handler) AuthenticateByName(w http.ResponseWriter, r *http.Request) {
	var req dto.AuthenticateByNameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := h.userRepo.GetUserByName(req.Name)
	if err != nil || user == nil {
		response.Error(w, http.StatusUnauthorized, "Invalid username or password")
		return
	}
	if user.IsDisabled {
		response.Error(w, http.StatusUnauthorized, "User account is disabled")
		return
	}

	if err := db.CheckPassword(user.Password, req.Pw); err != nil {
		response.Error(w, http.StatusUnauthorized, "Invalid username or password")
		return
	}

	token, err := generateToken()
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	// Parse client metadata from the authorization header.
	var clientName, deviceName, deviceID, appVersion string
	if hdr := r.Header.Get("X-Emby-Authorization"); hdr != "" {
		_, deviceID, clientName, deviceName, appVersion, _ = auth.ParseMediaBrowserHeader(hdr)
	}
	if deviceName == "" {
		deviceName = "Unknown Device"
	}
	if clientName == "" {
		clientName = "Unknown Client"
	}

	if err := h.upsertDevice(user.Id, deviceID, deviceName, clientName, appVersion, token); err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to create session")
		return
	}

	result := dto.AuthenticationResult{
		User:        userToDto(user, h.serverID),
		AccessToken: token,
		ServerId:    h.serverID,
	}
	response.JSON(w, http.StatusOK, result)
}

// AuthenticateUser handles POST /Users/{userId}/Authenticate.
// The password is supplied as the query parameter pw=<password>.
func (h *Handler) AuthenticateUser(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userId")
	pw := r.URL.Query().Get("pw")

	user, err := h.userRepo.GetUserByID(userID)
	if err != nil || user == nil {
		response.Error(w, http.StatusUnauthorized, "Invalid user or password")
		return
	}
	if user.IsDisabled {
		response.Error(w, http.StatusUnauthorized, "User account is disabled")
		return
	}

	if err := db.CheckPassword(user.Password, pw); err != nil {
		response.Error(w, http.StatusUnauthorized, "Invalid user or password")
		return
	}

	token, err := generateToken()
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	var clientName, deviceName, deviceID, appVersion string
	if hdr := r.Header.Get("X-Emby-Authorization"); hdr != "" {
		_, deviceID, clientName, deviceName, appVersion, _ = auth.ParseMediaBrowserHeader(hdr)
	}
	if deviceName == "" {
		deviceName = "Unknown Device"
	}
	if clientName == "" {
		clientName = "Unknown Client"
	}

	if err := h.upsertDevice(user.Id, deviceID, deviceName, clientName, appVersion, token); err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to create session")
		return
	}

	result := dto.AuthenticationResult{
		User:        userToDto(user, h.serverID),
		AccessToken: token,
		ServerId:    h.serverID,
	}
	response.JSON(w, http.StatusOK, result)
}

// GetPublicUsers handles GET /Users/Public.
// Returns only non-hidden, non-disabled users with Id, Name, and PrimaryImageTag.
func (h *Handler) GetPublicUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.userRepo.ListPublicUsers()
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to list users")
		return
	}

	dtos := make([]dto.UserDto, 0, len(users))
	for i := range users {
		dtos = append(dtos, dto.UserDto{
			Id:              users[i].Id,
			Name:            users[i].Name,
			ServerId:        h.serverID,
			PrimaryImageTag: users[i].PrimaryImageTag,
		})
	}
	response.JSON(w, http.StatusOK, dtos)
}

// GetUsers handles GET /Users.
// Requires admin. Returns all users as []UserDto.
func (h *Handler) GetUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.userRepo.ListUsers()
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to list users")
		return
	}

	dtos := make([]dto.UserDto, 0, len(users))
	for i := range users {
		dtos = append(dtos, *userToDto(&users[i], h.serverID))
	}
	response.JSON(w, http.StatusOK, dtos)
}

// GetUser handles GET /Users/{userId}.
// A user may only fetch their own record unless they are an admin.
func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	targetID := chi.URLParam(r, "userId")
	callerID := auth.GetUserID(r)

	// Resolve "me" shorthand.
	if targetID == "me" {
		targetID = callerID
	}

	// Must be self or admin.
	if targetID != callerID && !auth.IsAdmin(r) {
		response.Error(w, http.StatusForbidden, "Access denied")
		return
	}

	user, err := h.userRepo.GetUserByID(targetID)
	if err != nil {
		response.Error(w, http.StatusNotFound, "User not found")
		return
	}

	response.JSON(w, http.StatusOK, userToDto(user, h.serverID))
}

// ListApiKeys handles GET /Auth/Keys.
// Requires admin. Returns a QueryResult of AuthenticationInfo.
func (h *Handler) ListApiKeys(w http.ResponseWriter, r *http.Request) {
	keys, err := h.keyRepo.ListApiKeys()
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to list API keys")
		return
	}

	infos := make([]dto.AuthenticationInfo, 0, len(keys))
	for _, k := range keys {
		infos = append(infos, dto.AuthenticationInfo{
			Id:          k.Id,
			AccessToken: k.AccessToken,
			AppName:     k.Name,
			DateCreated: k.DateCreated,
			IsActive:    true,
		})
	}

	result := response.PaginatedResponse(infos, len(infos), 0)
	response.JSON(w, http.StatusOK, result)
}

// CreateApiKey handles POST /Auth/Keys.
// Requires admin. The display name is supplied as the query param app=<name>.
// Returns 204 No Content on success.
func (h *Handler) CreateApiKey(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("app")
	if name == "" {
		name = "Unknown"
	}

	if _, err := h.keyRepo.CreateApiKey(name); err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to create API key")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// RevokeApiKey handles DELETE /Auth/Keys/{key}.
// Requires admin. Returns 204 No Content on success.
func (h *Handler) RevokeApiKey(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	if err := h.keyRepo.DeleteApiKey(key); err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to revoke API key")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- helpers ---

// generateToken creates a 40-character lowercase hex token from 20 random bytes.
func generateToken() (string, error) {
	b := make([]byte, 20)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generateToken: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// upsertDevice inserts or updates a Device row for the given session.
func (h *Handler) upsertDevice(userID, deviceID, friendlyName, appName, appVersion, token string) error {
	if deviceID == "" {
		deviceID = uuid.New().String()
	}
	id := uuid.New().String()
	now := time.Now().UTC().Format("2006-01-02 15:04:05")

	_, err := h.db.Exec(
		`INSERT INTO Devices
		    (Id, UserId, DeviceId, FriendlyName, AppName, AppVersion, AccessToken, Created, DateLastActivity)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		 ON DUPLICATE KEY UPDATE
		    AccessToken      = VALUES(AccessToken),
		    DateLastActivity = VALUES(DateLastActivity)`,
		id, userID, deviceID, friendlyName, appName, appVersion, token, now, now,
	)
	if err != nil {
		return fmt.Errorf("upsertDevice: %w", err)
	}
	return nil
}

// userToDto converts an internal User row to a public-facing UserDto.
func userToDto(u *db.User, serverID string) *dto.UserDto {
	return &dto.UserDto{
		Id:                       u.Id,
		Name:                     u.Name,
		ServerId:                 serverID,
		PrimaryImageTag:          u.PrimaryImageTag,
		HasPassword:              u.Password != "",
		HasConfiguredPassword:    u.Password != "",
		HasConfiguredEasyPassword: false,
		EnableAutoLogin:          false,
		IsAdministrator:          u.IsAdmin,
		IsDisabled:               u.IsDisabled,
		IsHidden:                 u.IsHidden,
	}
}
