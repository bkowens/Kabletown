// Package dto contains shared data transfer objects used across Kabletown services.
package dto

import (
	"time"

	"github.com/google/uuid"
)

// UserDto represents a user in the system
type UserDto struct {
	// Core identification
	ID                uuid.UUID            `json:"Id"`
	Name              string               `json:"Name"`
	HasPassword       bool                 `json:"HasPassword"`
	PrimaryImageTag   string               `json:"PrimaryImageTag,omitempty"`
	PrimaryImageId    string               `json:"PrimaryImageId,omitempty"`
	LastLoginDate     *time.Time           `json:"LastLoginDate,omitempty"`

	// User preferences
	ShowMissingEpisodes bool               `json:"ShowMissingEpisodes"`
	HidePlayedInLatest  bool               `json:"HidePlayedInLatest"`
	RememberAudioSelections bool             `json:"RememberAudioSelections"`
	RememberSubtitleSelections bool         `json:"RememberSubtitleSelections"`
	EnableNextEpisodeAutoPlay bool           `json:"EnableNextEpisodeAutoPlay"`
	CollectionsSyncToFolders bool           `json:"CollectionsSyncToFolders"`

	// Policy/permissions
	Policy PolicyDto `json:"Policy"`

	// Configuration
	Configuration UserConfigurationDto `json:"Configuration"`

	// Sync settings
	SyncLibraries        []string     `json:"SyncLibraries,omitempty"`
	DisplayPreferencesId string       `json:"DisplayPreferencesId,omitempty"`
}

// PolicyDto represents user access control policy
type PolicyDto struct {
	// Authentication
	IsAdministrator              bool     `json:"IsAdministrator"`
	IsHidden                     bool     `json:"IsHidden"`
	IsDisabled                   bool     `json:"IsDisabled"`
	IsProtected                  bool     `json:"IsProtected"`

	// Access control
	EnableAudioPlayback         bool     `json:"EnableAudioPlayback"`
	EnableAudioDownload         bool     `json:"EnableAudioDownload"`
	EnableCommunityRatingEditing bool    `json:"EnableCommunityRatingEditing"`
	EnableContentDeletion       bool     `json:"EnableContentDeletion"`
	EnableContentDeletionFromFolders bool `json:"EnableContentDeletionFromFolders,omitempty"`
	EnableContentPlaying        bool     `json:"EnableContentPlaying"`
	EnableContentRecording      bool     `json:"EnableContentRecording"`
	EnableContentSharing        bool     `json:"EnableContentSharing"`
	EnableContentUploading      bool     `json:"EnableContentUploading"`
	EnableFavoriteOperations    bool     `json:"EnableFavoriteOperations"`
	EnableLabelEditing          bool     `json:"EnableLabelEditing"`
	EnableLiveTvManagement      bool     `json:"EnableLiveTvManagement"`
	EnableLiveTvAccess          bool     `json:"EnableLiveTvAccess"`
	EnableMediaConversion       bool     `json:"EnableMediaConversion"`
	EnableMediaDeletion         bool     `json:"EnableMediaDeletion"`
	EnableMediaDeletionFromFolders   bool `json:"EnableMediaDeletionFromFolders,omitempty"`
	EnableMediaPlayback         bool     `json:"EnableMediaPlayback"`
	EnableMediaReading          bool     `json:"EnableMediaReading"`
	EnableMediaSharing          bool     `json:"EnableMediaSharing"`
	EnableProfileEditing        bool     `json:"EnableProfileEditing"`
	EnableRemoteControlOtherUsers bool   `json:"EnableRemoteControlOtherUsers"`
	EnableRemoteClient          bool     `json:"EnableRemoteClient"`
	EnableSubtitleEditing       bool     `json:"EnableSubtitleEditing"`
	EnableSubtitleDownloading   bool     `json:"EnableSubtitleDownloading"`
	EnableSubtitleUpload        bool     `json:"EnableSubtitleUpload"`
	EnableSynchPlay             bool     `json:"EnableSyncPlay"`
	EnableCollectionCreation    bool     `json:"EnableCollectionCreation"`
	EnablePlaylistCreation      bool     `json:"EnablePlaylistCreation"`
	EnablePublicServerSharing   bool     `json:"EnablePublicServerSharing"`

	// Access restrictions
	BlockedTags []string `json:"BlockedTags,omitempty"`
	AllowedTags []string `json:"AllowedTags,omitempty"`
	EnableAllChannels     bool     `json:"EnableAllChannels"`
	EnableAllDevices      bool     `json:"EnableAllDevices"`
	EnableAllFolders      bool     `json:"EnableAllFolders"`
	EnabledChannels       []string `json:"EnabledChannels,omitempty"`
	EnabledDevices        []string `json:"EnabledDevices,omitempty"`
	EnabledFolders        []string `json:"EnabledFolders,omitempty"`

	// Audio/subtitle limits
	MaxActiveStreams    int      `json:"MaxActiveStreams"`
	EnablePublicSharing bool     `json:"EnablePublicSharing"`
	SubtitleEncodingOption string `json:"SubtitleEncodingOption,omitempty"`
	SubtitlePlaybackMode   SubtitlePlaybackMode `json:"SubtitlePlaybackMode"`

	// Content limits
	EnableAudioPlaybackRestriction bool       `json:"EnableAudioPlaybackRestriction"`
	EnableAudioPlaybackLimit       bool       `json:"EnableAudioPlaybackLimit"`
	MaxAudioPlaybackBitrate       int64      `json:"MaxAudioPlaybackBitrate"`
	MaxAudioPlaybackChannels      int        `json:"MaxAudioPlaybackChannels"`
	EnableVideoPlaybackRestriction bool       `json:"EnableVideoPlaybackRestriction"`
	EnableVideoPlaybackLimit       bool       `json:"EnableVideoPlaybackLimit"`
	MaxVideoPlaybackBitrate       int64      `json:"MaxVideoPlaybackBitrate"`
	MaxVideoPlaybackResolution    string     `json:"MaxVideoPlaybackResolution,omitempty"`

	// Library access
	EnableSyncPlayLibraryChange   bool     `json:"EnableSyncPlayLibraryChange"`
}

// UserConfigurationDto represents user-specific configuration
type UserConfigurationDto struct {
	// Media player settings
	AudioLanguagePreference    string         `json:"AudioLanguagePreference,omitempty"`
	PlayDefaultAudioTrack      bool           `json:"PlayDefaultAudioTrack"`
	SubtitleLanguagePreference string         `json:"SubtitleLanguagePreference,omitempty"`
	DisplayLanguagePreference  string         `json:"DisplayLanguagePreference,omitempty"`
	DisplaySubtitlePreference  SubtitleDisplayPreference `json:"DisplaySubtitlePreference"`
	PlayDefaultSubtitleTrack   bool           `json:"PlayDefaultSubtitleTrack"`

	// Playback settings
	AlwaysAudioSubtitleMode   SubtitlePlaybackMode `json:"AlwaysAudioSubtitleMode"`
	AlwaysVideoSubtitleMode   SubtitlePlaybackMode `json:"AlwaysVideoSubtitleMode"`

	// App settings
	HomeSectionTypes          []string           `json:"HomeSectionTypes,omitempty"`
	ScreenSaveMode            string             `json:"ScreenSaveMode,omitempty"`
	SkipPromptForSeriesResume bool               `json:"SkipPromptForSeriesResume"`

	// Collections
	CollectionFolderId        string             `json:"CollectionFolderId,omitempty"`
	EnableGroupCollections    bool               `json:"EnableGroupCollections"`

	// Media source preferences
	MediaSourceName           string             `json:"MediaSourceName,omitempty"`
	MediaSourceProfileId      string             `json:"MediaSourceProfileId,omitempty"`
}

// SubtitlePlaybackMode represents subtitle playback behavior
type SubtitlePlaybackMode string

const (
	SubtitlePlaybackModeDefault    SubtitlePlaybackMode = "Default"
	SubtitlePlaybackModeAlways     SubtitlePlaybackMode = "Always"
	SubtitlePlaybackModeAutomatic  SubtitlePlaybackMode = "Automatic"
	SubtitlePlaybackModeNever      SubtitlePlaybackMode = "Never"
	SubtitlePlaybackModeBurnIn     SubtitlePlaybackMode = "BurnIn"
)

// SubtitleDisplayPreference represents subtitle display preferences
type SubtitleDisplayPreference string

const (
	SubtitleDisplayPreferenceDefault  SubtitleDisplayPreference = "Default"
	SubtitleDisplayPreferenceAlways   SubtitleDisplayPreference = "Always"
	SubtitleDisplayPreferenceNever    SubtitleDisplayPreference = "Never"
)

// AuthenticationResultDto represents the result of an authentication attempt
type AuthenticationResultDto struct {
	// Auth token
	AccessToken string `json:"AccessToken"`

	// User info
	User UserDto `json:"User"`

	// Session info
	SessionInfo SessionDto `json:"SessionInfo"`

	// User policy (simplified)
	Policy PolicyDto `json:"Policy"`

	// Config
	Config UserConfigurationDto `json:"Config"`
}

// SessionDto represents an active user session
type SessionDto struct {
	// Identification
	ID               string              `json:"Id"`
	UserID           string              `json:"UserId"`
	UserName         string              `json:"UserName,omitempty"`
	DeviceID         string              `json:"DeviceId"`
	DeviceName       string              `json:"DeviceName"`
	AppName          string              `json:"AppName"`
	Client           string              `json:"Client"`
	Version          string              `json:"Version,omitempty"`

	// Timestamp
	LastActivityDate   time.Time           `json:"LastActivityDate"`
	SessionStartDate   time.Time           `json:"SessionStartDate,omitempty"`

	// Current activity
	NowPlayingItem     BaseItemDto         `json:"NowPlayingItem,omitempty"`
	PlayableMediaTypes []string            `json:"PlayableMediaTypes,omitempty"`

	// Client capabilities
	Capabilities       ClientCapabilities  `json:"Capabilities"`
	PlayState          PlayerStateDto      `json:"PlayState"`
	AdditionalInfo     string              `json:"AdditionalInfo,omitempty"`

	// Session control
	SupportsMediaControl bool             `json:"SupportsMediaControl"`
	SupportsSupportedCommands []string    `json:"SupportsSupportedCommands"`
}

// ClientCapabilities represents what a client device can do
type ClientCapabilities struct {
	PlayableMediaTypes []string `json:"PlayableMediaTypes,omitempty"`
	SupportedCommands  []string `json:"SupportedCommands,omitempty"`
	SupportsMediaControl bool `json:"SupportsMediaControl"`
	SupportsPersistentId bool `json:"SupportsPersistentId"`
}

// PlayerStateDto represents current player state for a session
type PlayerStateDto struct {
	IsPaused           bool    `json:"IsPaused"`
	IsMuted            bool    `json:"IsMuted"`
	PositionTicks      int64   `json:"PositionTicks"`
	MediaSourceID      string  `json:"MediaSourceId,omitempty"`
	PlayMethod         string  `json:"PlayMethod,omitempty"`
	RepeatMode         string  `json:"RepeatMode,omitempty"`
}
