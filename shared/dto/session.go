// Package dto contains shared data transfer objects used across Kabletown services.
package dto

import (
	"time"
)

// SessionInfoDto represents detailed session information for API responses
type SessionInfoDto struct {
	// Core identification
	ID              string   `json:"Id"`
	UserID          string   `json:"UserId"`
	UserName        string   `json:"UserName,omitempty"`
	DeviceID        string   `json:"DeviceId"`
	DeviceName      string   `json:"DeviceName"`
	Client          string   `json:"Client"`
	AppVersion      string   `json:"AppVersion,omitempty"`

	// Session lifecycle
	LastActivityDate     time.Time   `json:"LastActivityDate"`
	SessionStartDate     time.Time   `json:"SessionStartDate,omitempty"`
	LastPlaybackCheck    time.Time   `json:"LastPlaybackCheck,omitempty"`

	// Now playing state
	NowPlayingItem       *BaseItemDto `json:"NowPlayingItem,omitempty"`
	PlayableMediaTypes   []string    `json:"PlayableMediaTypes,omitempty"`
	AdditionalInfo       string      `json:"AdditionalInfo,omitempty"`

	// Client connection
	RemoteEndPoint       string      `json:"RemoteEndPoint,omitempty"`
	TranscodingSessions []TranscodeSessionInfo `json:"TranscodingSessions,omitempty"`

	// Capabilities
	Capabilities         ClientCapabilitiesDto `json:"Capabilities"`
	PlayState            PlayerStateInfoDto    `json:"PlayState"`

	// Session control
	SupportsMediaControl     bool `json:"SupportsMediaControl"`
	SupportsMediaControlV2   bool `json:"SupportsMediaControlV2,omitempty"`
	IsUserSession            bool `json:"IsUserSession"`

	// Application details
	ApplicationVersionInfo ApplicationVersionInfoDto `json:"ApplicationVersionInfo,omitempty"`
	DeviceInfo             DeviceInfoDto             `json:"DeviceInfo,omitempty"`
}

// TranscodeSessionInfo represents an active transcoding session
type TranscodeSessionInfo struct {
	// Identification
	ID                 string   `json:"Id"`
	SessionID         string   `json:"SessionId"`
	MediaSourceID     string   `json:"MediaSourceId,omitempty"`
	ItemId            string   `json:"ItemId,omitempty"`

	// Transport details
	ClientIPAddress   string   `json:"ClientIpAddress,omitempty"`
	Container         string   `json:"Container,omitempty"`

	// Playback parameters
	VideoCodec        string   `json:"VideoCodec,omitempty"`
	AudioCodec        string   `json:"AudioCodec,omitempty"`
	Protocol          string   `json:"Protocol,omitempty"`
	PlayMethod        string   `json:"PlayMethod,omitempty"`
	PlaySessionID     string   `json:"PlaySessionId,omitempty"`
	TimeOffsetTicks   int64    `json:"TimeOffsetTicks"`

	// Transcoding info
	IsVideoDevice      bool     `json:"IsVideoDevice"`
	TranscodingThrottleFactor int `json:"TranscodingThrottleFactor"`
	EstimatedTargetBitRate int `json:"EstimatedTargetBitRate"`
	TranscodingOffset  int64    `json:"TranscodingOffset"`
}

// ClientCapabilitiesDto represents what a client device can do
type ClientCapabilitiesDto struct {
	// Supported features
	PlayableMediaTypes []string `json:"PlayableMediaTypes,omitempty"`
	ProfileCapabilities ProfileCapabilitiesDto `json:"ProfileCapabilities,omitempty"`

	// Command support
	SupportedCommands []string `json:"SupportedCommands,omitempty"`
	SupportsMediaControl bool `json:"SupportsMediaControl"`
	SupportsPersistentId bool `json:"SupportsPersistentId"`
	SupportsSyncPlay   bool   `json:"SupportsSyncPlay"`
}

// ProfileCapabilitiesDto represents client profile information
type ProfileCapabilitiesDto struct {
	ContainerProfile     string   `json:"ContainerProfile,omitempty"`
	VideoCodecProfile    string   `json:"VideoCodecProfile,omitempty"`
	AudioCodecProfile    string   `json:"AudioCodecProfile,omitempty"`
}

// PlayerStateInfoDto represents current player state for a session
type PlayerStateInfoDto struct {
	// State flags
	IsPaused   bool   `json:"IsPaused"`
	IsMuted    bool   `json:"IsMuted"`
	IsVolumeMuted bool `json:"IsVolumeMuted"`

	// Position
	PositionTicks         int64  `json:"PositionTicks"`
	SubtitleOffset        int64  `json:"SubtitleOffset"`
	SubtitleTrackIndex    int    `json:"SubtitleTrackIndex"`
	SubtitleTrackId       string `json:"SubtitleTrackId,omitempty"`

	// Media source
	MediaSourceID         string `json:"MediaSourceId,omitempty"`
	PlayMethod            string `json:"PlayMethod,omitempty"`
	PlaySessionID         string `json:"PlaySessionId,omitempty"`

	// Repeat/shuffle
	RepeatMode         string `json:"RepeatMode,omitempty"`
	IsShuffle          bool   `json:"IsShuffle"`
}

// ApplicationVersionInfoDto represents client application version info
type ApplicationVersionInfoDto struct {
	Version           string `json:"Version,omitempty"`
	Client            string `json:"Client,omitempty"`
	OperatingSystem   string `json:"OperatingSystem,omitempty"`
	BuildNumber       string `json:"BuildNumber,omitempty"`
}

// DeviceInfoDto represents device information for a session
type DeviceInfoDto struct {
	// Identification
	ID           string   `json:"Id"`
	Name         string   `json:"Name"`
	DeviceID     string   `json:"DeviceId"`
	LastUserId   string   `json:"LastUserId,omitempty"`

	// Hardware
	Manufacturer string   `json:"Manufacturer,omitempty"`
	Model        string   `json:"Model,omitempty"`
	DeviceType   string   `json:"DeviceType,omitempty"`

	// Connection
	LastActivityDate     time.Time   `json:"LastActivityDate"`
	DateLastActivityed   time.Time   `json:"DateLastActivityed,omitempty"`

	// App info
	AppName         string   `json:"AppName,omitempty"`
	AppVersion      string   `json:"AppVersion,omitempty"`
}

// SessionCapabilityInfo represents capability information during session creation
type SessionCapabilityInfo struct {
	PlayableMediaTypes []string `json:"PlayableMediaTypes"`
	Profile            string   `json:"Profile,omitempty"`
	DeviceOptions        DeviceOptions `json:"DeviceOptions,omitempty"`
	DisplayPreferences []DisplayPreference `json:"DisplayPreferences,omitempty"`
	SupportedCommands  []string `json:"SupportedCommands"`
	SupportsMediaControl bool `json:"SupportsMediaControl"`
	SupportsPersistentId bool `json:"SupportsPersistentId"`
}

// DeviceOptions contains device-specific options
type DeviceOptions struct {
	EnableRealtimeTranscoding bool `json:"EnableRealtimeTranscoding"`
	EnableDirectPlay          bool `json:"EnableDirectPlay"`
	EnableDirectStream        bool `json:"EnableDirectStream"`
}

// DisplayPreference represents a display preference setting
type DisplayPreference struct {
	ItemID             string `json:"ItemId"`
	Type              string `json:"Type"`
	IsFolder          bool   `json:"IsFolder"`
	SortOrder         string `json:"SortOrder,omitempty"`
	ViewType          string `json:"ViewType,omitempty"`
}
