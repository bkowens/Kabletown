package dto

// UserDto is the public-facing representation of a user.
type UserDto struct {
	Id                          string         `json:"Id"`
	Name                        string         `json:"Name"`
	ServerId                    string         `json:"ServerId,omitempty"`
	PrimaryImageTag             string         `json:"PrimaryImageTag,omitempty"`
	HasPassword                 bool           `json:"HasPassword"`
	HasConfiguredPassword       bool           `json:"HasConfiguredPassword"`
	HasConfiguredEasyPassword   bool           `json:"HasConfiguredEasyPassword"`
	EnableAutoLogin             bool           `json:"EnableAutoLogin"`
	LastLoginDate               *string        `json:"LastLoginDate,omitempty"`
	LastActivityDate            *string        `json:"LastActivityDate,omitempty"`
	IsAdministrator             bool           `json:"IsAdministrator"`
	IsDisabled                  bool           `json:"IsDisabled"`
	IsHidden                    bool           `json:"IsHidden"`
	Policy                      *UserPolicyDto `json:"Policy,omitempty"`
	Configuration               *UserConfigDto `json:"Configuration,omitempty"`
}

// UserPolicyDto carries permissions/policy settings for a user.
type UserPolicyDto struct {
	IsAdministrator                  bool     `json:"IsAdministrator"`
	IsHidden                         bool     `json:"IsHidden"`
	IsDisabled                       bool     `json:"IsDisabled"`
	EnableUserPreferenceAccess       bool     `json:"EnableUserPreferenceAccess"`
	EnableRemoteAccess               bool     `json:"EnableRemoteAccess"`
	EnableLiveTvManagement           bool     `json:"EnableLiveTvManagement"`
	EnableLiveTvAccess               bool     `json:"EnableLiveTvAccess"`
	EnableMediaPlayback              bool     `json:"EnableMediaPlayback"`
	EnableAudioPlaybackTranscoding   bool     `json:"EnableAudioPlaybackTranscoding"`
	EnableVideoPlaybackTranscoding   bool     `json:"EnableVideoPlaybackTranscoding"`
	EnablePlaybackRemuxing           bool     `json:"EnablePlaybackRemuxing"`
	EnableContentDeletion            bool     `json:"EnableContentDeletion"`
	EnableContentDownloading         bool     `json:"EnableContentDownloading"`
	EnableAllDevices                 bool     `json:"EnableAllDevices"`
	EnableAllChannels                bool     `json:"EnableAllChannels"`
	EnableAllFolders                 bool     `json:"EnableAllFolders"`
	InvalidLoginAttemptCount         int      `json:"InvalidLoginAttemptCount"`
	LoginAttemptsBeforeLockout       int      `json:"LoginAttemptsBeforeLockout"`
	MaxActiveSessions                int      `json:"MaxActiveSessions"`
	EnablePublicSharing              bool     `json:"EnablePublicSharing"`
	AuthenticationProviderId         string   `json:"AuthenticationProviderId"`
	PasswordResetProviderId          string   `json:"PasswordResetProviderId"`
	SyncPlayAccess                   string   `json:"SyncPlayAccess"`
}

// UserConfigDto carries per-user playback and display configuration.
type UserConfigDto struct {
	PlayDefaultAudioTrack      bool   `json:"PlayDefaultAudioTrack"`
	SubtitleMode               string `json:"SubtitleMode"`
	HidePlayedInLatest         bool   `json:"HidePlayedInLatest"`
	RememberAudioSelections    bool   `json:"RememberAudioSelections"`
	RememberSubtitleSelections bool   `json:"RememberSubtitleSelections"`
	EnableNextEpisodeAutoPlay  bool   `json:"EnableNextEpisodeAutoPlay"`
}

// CreateUserRequest is the body for POST /Users/New.
type CreateUserRequest struct {
	Name     string `json:"Name"`
	Password string `json:"Password"`
}

// UpdateUserRequest is the body for PUT /Users/{userId}.
type UpdateUserRequest struct {
	Name          string         `json:"Name"`
	Configuration *UserConfigDto `json:"Configuration,omitempty"`
}

// ChangePasswordRequest is the body for POST /Users/{userId}/Password.
type ChangePasswordRequest struct {
	CurrentPw     string `json:"CurrentPw"`
	NewPw         string `json:"NewPw"`
	ResetPassword bool   `json:"ResetPassword"`
}

// UserItemDataDto is returned by favorite/played state endpoints.
type UserItemDataDto struct {
	Rating                float64 `json:"Rating,omitempty"`
	PlayedPercentage      float64 `json:"PlayedPercentage,omitempty"`
	UnplayedItemCount     int     `json:"UnplayedItemCount,omitempty"`
	PlaybackPositionTicks int64   `json:"PlaybackPositionTicks"`
	PlayCount             int     `json:"PlayCount"`
	IsFavorite            bool    `json:"IsFavorite"`
	Played                bool    `json:"Played"`
	Key                   string  `json:"Key"`
	ItemId                string  `json:"ItemId,omitempty"`
}

// BaseItemDto is a lightweight item DTO for views/latest endpoints.
type BaseItemDto struct {
	Id              string `json:"Id"`
	Name            string `json:"Name"`
	ServerId        string `json:"ServerId,omitempty"`
	Type            string `json:"Type"`
	IsFolder        bool   `json:"IsFolder"`
	CollectionType  string `json:"CollectionType,omitempty"`
	PrimaryImageTag string `json:"PrimaryImageTag,omitempty"`
	ParentId        string `json:"ParentId,omitempty"`
	RunTimeTicks    int64  `json:"RunTimeTicks,omitempty"`
}
