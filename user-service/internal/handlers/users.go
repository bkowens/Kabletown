package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/bowens/kabletown/user-service/internal/db"
	"github.com/bowens/kabletown/user-service/internal/dto"
	"github.com/bowens/kabletown/shared/auth"
	"github.com/bowens/kabletown/shared/response"
)

// ListUsers handles GET /Users.
// Admins get all users; non-admins get only themselves.
func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	callerID := auth.GetUserID(r)
	isAdmin := auth.IsAdmin(r)

	if isAdmin {
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
		return
	}

	user, err := h.userRepo.GetUserByID(callerID)
	if err != nil || user == nil {
		response.Error(w, http.StatusNotFound, "user not found")
		return
	}
	response.JSON(w, http.StatusOK, []dto.UserDto{*userToDto(user, h.serverID)})
}

// GetUser handles GET /Users/{userId}.
func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	targetID := chi.URLParam(r, "userId")
	callerID := auth.GetUserID(r)

	if targetID == "me" {
		targetID = callerID
	}

	if targetID != callerID && !auth.IsAdmin(r) {
		response.Error(w, http.StatusForbidden, "Access denied")
		return
	}

	user, err := h.userRepo.GetUserByID(targetID)
	if err != nil || user == nil {
		response.Error(w, http.StatusNotFound, "User not found")
		return
	}

	d := userToDto(user, h.serverID)
	// Include policy and configuration
	if user.Policy != "" {
		var pol dto.UserPolicyDto
		if json.Unmarshal([]byte(user.Policy), &pol) == nil {
			d.Policy = &pol
		}
	}
	if d.Policy == nil {
		d.Policy = defaultPolicy(user.IsAdmin)
	}
	if user.Configuration != "" {
		var cfg dto.UserConfigDto
		if json.Unmarshal([]byte(user.Configuration), &cfg) == nil {
			d.Configuration = &cfg
		}
	}
	if d.Configuration == nil {
		d.Configuration = defaultConfig()
	}

	response.JSON(w, http.StatusOK, d)
}

// CreateUser handles POST /Users/New (admin only).
func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" {
		response.Error(w, http.StatusBadRequest, "Name is required")
		return
	}

	user, err := h.userRepo.CreateUser(req.Name, req.Password)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to create user")
		return
	}
	d := userToDto(user, h.serverID)
	d.Policy = defaultPolicy(false)
	d.Configuration = defaultConfig()
	response.JSON(w, http.StatusOK, d)
}

// UpdateUser handles PUT /Users/{userId}.
func (h *Handler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	targetID := chi.URLParam(r, "userId")
	callerID := auth.GetUserID(r)

	if targetID != callerID && !auth.IsAdmin(r) {
		response.Error(w, http.StatusForbidden, "Access denied")
		return
	}

	var req dto.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	cfgJSON := ""
	if req.Configuration != nil {
		b, _ := json.Marshal(req.Configuration)
		cfgJSON = string(b)
	}

	if err := h.userRepo.UpdateUser(targetID, req.Name, cfgJSON, ""); err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to update user")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// DeleteUser handles DELETE /Users/{userId} (admin only).
func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	targetID := chi.URLParam(r, "userId")
	if err := h.userRepo.DeleteUser(targetID); err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to delete user")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ChangePassword handles POST /Users/{userId}/Password.
func (h *Handler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	targetID := chi.URLParam(r, "userId")
	callerID := auth.GetUserID(r)

	if targetID != callerID && !auth.IsAdmin(r) {
		response.Error(w, http.StatusForbidden, "Access denied")
		return
	}

	var req dto.ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Non-admins must supply the current password.
	if !auth.IsAdmin(r) && !req.ResetPassword {
		user, err := h.userRepo.GetUserByID(targetID)
		if err != nil || user == nil {
			response.Error(w, http.StatusNotFound, "User not found")
			return
		}
		if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.CurrentPw)) != nil {
			response.Error(w, http.StatusUnauthorized, "Current password is incorrect")
			return
		}
	}

	if err := h.userRepo.UpdatePassword(targetID, req.NewPw); err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to update password")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// UpdatePolicy handles POST /Users/{userId}/Policy (admin only).
func (h *Handler) UpdatePolicy(w http.ResponseWriter, r *http.Request) {
	targetID := chi.URLParam(r, "userId")

	var pol dto.UserPolicyDto
	if err := json.NewDecoder(r.Body).Decode(&pol); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	b, _ := json.Marshal(pol)
	if err := h.userRepo.UpdateUser(targetID, "", "", string(b)); err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to update policy")
		return
	}

	// Sync IsAdministrator to the Permissions table.
	val := 0
	if pol.IsAdministrator {
		val = 1
	}
	h.db.Exec(
		`INSERT INTO Permissions (UserId, Kind, Value) VALUES (?, 0, ?)
		 ON DUPLICATE KEY UPDATE Value = ?`,
		targetID, val, val,
	) //nolint:errcheck

	w.WriteHeader(http.StatusNoContent)
}

// UpdateConfiguration handles POST /Users/{userId}/Configuration.
func (h *Handler) UpdateConfiguration(w http.ResponseWriter, r *http.Request) {
	targetID := chi.URLParam(r, "userId")
	callerID := auth.GetUserID(r)

	if targetID != callerID && !auth.IsAdmin(r) {
		response.Error(w, http.StatusForbidden, "Access denied")
		return
	}

	var cfg dto.UserConfigDto
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	b, _ := json.Marshal(cfg)
	if err := h.userRepo.UpdateUser(targetID, "", string(b), ""); err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to update configuration")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// userToDto converts an internal db.User to a public UserDto.
func userToDto(u *db.User, serverID string) *dto.UserDto {
	return &dto.UserDto{
		Id:                        u.Id,
		Name:                      u.Name,
		ServerId:                  serverID,
		PrimaryImageTag:           u.PrimaryImageTag,
		HasPassword:               u.Password != "",
		HasConfiguredPassword:     u.Password != "",
		HasConfiguredEasyPassword: false,
		EnableAutoLogin:           false,
		IsAdministrator:           u.IsAdmin,
		IsDisabled:                u.IsDisabled,
		IsHidden:                  u.IsHidden,
	}
}

func defaultPolicy(isAdmin bool) *dto.UserPolicyDto {
	return &dto.UserPolicyDto{
		IsAdministrator:            isAdmin,
		EnableUserPreferenceAccess: true,
		EnableRemoteAccess:         true,
		EnableLiveTvAccess:         true,
		EnableMediaPlayback:        true,
		EnableAudioPlaybackTranscoding: true,
		EnableVideoPlaybackTranscoding: true,
		EnablePlaybackRemuxing:     true,
		EnableContentDownloading:   true,
		EnableAllDevices:           true,
		EnableAllChannels:          true,
		EnableAllFolders:           true,
		EnablePublicSharing:        true,
		AuthenticationProviderId:   "Emby.Server.Implementations.Library.DefaultAuthenticationProvider",
		PasswordResetProviderId:    "Emby.Server.Implementations.Library.DefaultPasswordResetProvider",
		SyncPlayAccess:             "CreateAndJoinGroups",
	}
}

func defaultConfig() *dto.UserConfigDto {
	return &dto.UserConfigDto{
		PlayDefaultAudioTrack:      true,
		SubtitleMode:               "Default",
		HidePlayedInLatest:         true,
		RememberAudioSelections:    true,
		RememberSubtitleSelections: true,
		EnableNextEpisodeAutoPlay:  true,
	}
}
