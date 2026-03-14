package config

import (
	"encoding/xml"
	"os"
	"time"
)

// SystemConfig holds system-wide configuration parsed from system.xml
type SystemConfig struct {
	// Authentication settings
	AuthenticationProtocol            string `xml:"AuthenticationProtocol" json:"authentication_protocol"`
	EnableBasicAuth                   bool   `xml:"EnableBasicAuth" json:"enable_basic_auth"`
	EnableClientCertificateAuth       bool   `xml:"EnableClientCertificateAuth" json:"enable_client_certificate_auth"`
	EnableSimpleBearerAuth            bool   `xml:"EnableSimpleBearerAuth" json:"enable_simple_bearer_auth"`
	EnforceAuth                       bool   `xml:"EnforceAuth" json:"enforce_auth"`
	LocalAuthPolicy                   string `xml:"LocalAuthPolicy" json:"local_auth_policy"`

	// Display preferences
	AppName                           string `xml:"AppName" json:"app_name"`
	AppVersion                        string `xml:"AppVersion" json:"app_version"`
	EnableAutoDiscovery               bool   `xml:"EnableAutoDiscovery" json:"enable_auto_discovery"`

	// Encoding settings
	EncodingOptions                   *EncodingOptionsType `xml:"EncodingOptions" json:"encoding_options"`

	// Image extraction
	ImageExtractionIntervalSec        int    `xml:"ImageExtractionIntervalSec" json:"image_extraction_interval_sec"`
	ImageExtractionMaxCount           int    `xml:"ImageExtractionMaxCount" json:"image_extraction_max_count"`

	// Library settings
	EnableLibraryMonitor              bool   `xml:"EnableLibraryMonitor" json:"enable_library_monitor"`
	EnableLibraryMonitorDuringScan    bool   `xml:"EnableLibraryMonitorDuringScan" json:"enable_library_monitor_during_scan"`

	// Metadata settings
	EnableAutomaticRefresh            bool   `xml:"EnableAutomaticRefresh" json:"enable_automatic_refresh"`

	// Notification settings
	EnableNotifications               bool   `xml:"EnableNotifications" json:"enable_notifications"`

	// Playback settings
	EnableExternalMediaPlayback       bool   `xml:"EnableExternalMediaPlayback" json:"enable_external_media_playback"`

	// Scheduled tasks
	EnableScheduledTask               bool   `xml:"EnableScheduledTask" json:"enable_scheduled_task"`

	// Transcoding settings
	EnableTranscoding                 bool   `xml:"EnableTranscoding" json:"enable_transcoding"`
	TranscodingTempPath               string `xml:"TranscodingTempPath" json:"transcoding_temp_path"`
	TempDirKeepIdleTimeMin            int    `xml:"TempDirKeepIdleTimeMin" json:"temp_dir_keep_idle_time_min"`

	// UI/UX settings
	EnableQuickConnect                bool   `xml:"EnableQuickConnect" json:"enable_quick_connect"`
	EnableUserPasswordReset           bool   `xml:"EnableUserPasswordReset" json:"enable_user_password_reset"`

	// Default values
	DefaultAuthenticationProtocol     string `json:"default_authentication_protocol"`
	DefaultAppName                    string `json:"default_app_name"`
	DefaultAppVersion                 string `json:"default_app_version"`
}

// EncodingOptionsType contains transcoding encoding options
type EncodingOptionsType struct {
	// Audio codec options
	EnableAudioCache              bool   `xml:"EnableAudioCache" json:"enable_audio_cache"`
	EnableAudioTranscoding        bool   `xml:"EnableAudioTranscoding" json:"enable_audio_transcoding"`

	// Video codec options
	EnableVideoCodecCache         bool   `xml:"EnableVideoCodecCache" json:"enable_video_codec_cache"`
	EnableVideoTranscoding        bool   `xml:"EnableVideoTranscoding" json:"enable_video_transcoding"`

	// Hardware acceleration
	EnableHardwareDecoding        bool   `xml:"EnableHardwareDecoding" json:"enable_hardware_decoding"`
	EnableHardwareEncoding        bool   `xml:"EnableHardwareEncoding" json:"enable_hardware_encoding"`

	// Transcoding quality
	MaxStreamingBitrate           int    `xml:"MaxStreamingBitrate" json:"max_streaming_bitrate"`
	MaxStaticBitrate              int    `xml:"MaxStaticBitrate" json:"max_static_bitrate"`

	// Throttling
	EnableThrottling              bool   `xml:"EnableThrottling" json:"enable_throttling"`
	ThrottlingSec                 int    `xml:"ThrottlingSec" json:"throttling_sec"`

	// FFmpeg paths
	FFmpegPath                    string `xml:"FFmpegPath" json:"ffmpeg_path"`
	FFprobePath                   string `xml:"FFprobePath" json:"ffprobe_path"`
}

// SystemXML is the XML structure matching system.xml format
type SystemXML struct {
	XMLName xml.Name `xml:"ServerConfiguration"`

	// Authentication
	AuthenticationProtocol            string `xml:"AuthenticationProtocol"`
	EnableBasicAuth                   bool   `xml:"EnableBasicAuth"`
	EnableClientCertificateAuth       bool   `xml:"EnableClientCertificateAuth"`
	EnableSimpleBearerAuth            bool   `xml:"EnableSimpleBearerAuth"`
	EnforceAuth                       bool   `xml:"EnforceAuth"`
	LocalAuthPolicy                   string `xml:"LocalAuthPolicy"`

	// UI/Display
	AppName                           string `xml:"AppName"`
	AppVersion                        string `xml:"AppVersion"`
	EnableAutoDiscovery               bool   `xml:"EnableAutoDiscovery"`

	// Image extraction
	ImageExtractionIntervalSec        int    `xml:"ImageExtractionIntervalSec"`
	ImageExtractionMaxCount           int    `xml:"ImageExtractionMaxCount"`

	// Library
	EnableLibraryMonitor              bool   `xml:"EnableLibraryMonitor"`
	EnableLibraryMonitorDuringScan    bool   `xml:"EnableLibraryMonitorDuringScan"`

	// Metadata
	EnableAutomaticRefresh            bool   `xml:"EnableAutomaticRefresh"`

	// Notifications
	EnableNotifications               bool   `xml:"EnableNotifications"`

	// Playback
	EnableExternalMediaPlayback       bool   `xml:"EnableExternalMediaPlayback"`

	// Scheduled tasks
	EnableScheduledTask               bool   `xml:"EnableScheduledTask"`

	// Transcoding
	EnableTranscoding                 bool   `xml:"EnableTranscoding"`
	TranscodingTempPath               string `xml:"TranscodingTempPath"`
	TempDirKeepIdleTimeMin            int    `xml:"TempDirKeepIdleTimeMin"`

	// QuickConnect
	EnableQuickConnect                bool   `xml:"EnableQuickConnect"`
	EnableUserPasswordReset           bool   `xml:"EnableUserPasswordReset"`

	// Encoding options
	EncodingOptions                   *EncodingOptionsType `xml:"EncodingOptions"`
}

// DefaultSystemConfig returns a SystemConfig with sensible defaults
func DefaultSystemConfig() *SystemConfig {
	return &SystemConfig{
		DefaultAuthenticationProtocol: "MediaBrowser",
		DefaultAppName:                "Kabletown",
		DefaultAppVersion:             "1.0.0",
		AuthenticationProtocol:        "MediaBrowser",
		AppName:                       "Kabletown",
		AppVersion:                    "1.0.0",
		EnableAutoDiscovery:           true,
		EnableLibraryMonitor:          true,
		EnableAutomaticRefresh:        true,
		EnableNotifications:           true,
		EnableQuickConnect:            true,
		EnableUserPasswordReset:       true,
		EnableTranscoding:             true,
		ImageExtractionIntervalSec:    300, // 5 minutes
		ImageExtractionMaxCount:       10,
		TempDirKeepIdleTimeMin:        20,
		EnableExternalMediaPlayback:   true,
		EnableScheduledTask:            true,
	}
}

// LoadSystemConfig reads and parses system.xml from the given path
func LoadSystemConfig(systemXMLPath string) (*SystemConfig, error) {
	// Check if file exists
	if _, err := os.Stat(systemXMLPath); os.IsNotExist(err) {
		// Return defaults if file doesn't exist
		return DefaultSystemConfig(), nil
	}

	// Read file contents
	data, err := os.ReadFile(systemXMLPath)
	if err != nil {
		return DefaultSystemConfig(), err
	}

	// Parse XML
	var systemXML SystemXML
	if err := xml.Unmarshal(data, &systemXML); err != nil {
		return DefaultSystemConfig(), err
	}

	// Convert to SystemConfig
	config := &SystemConfig{
		AuthenticationProtocol:            systemXML.AuthenticationProtocol,
		EnableBasicAuth:                   systemXML.EnableBasicAuth,
		EnableClientCertificateAuth:       systemXML.EnableClientCertificateAuth,
		EnableSimpleBearerAuth:            systemXML.EnableSimpleBearerAuth,
		EnforceAuth:                       systemXML.EnforceAuth,
		LocalAuthPolicy:                   systemXML.LocalAuthPolicy,
		AppName:                           systemXML.AppName,
		AppVersion:                        systemXML.AppVersion,
		EnableAutoDiscovery:               systemXML.EnableAutoDiscovery,
		ImageExtractionIntervalSec:        systemXML.ImageExtractionIntervalSec,
		ImageExtractionMaxCount:           systemXML.ImageExtractionMaxCount,
		EnableLibraryMonitor:              systemXML.EnableLibraryMonitor,
		EnableLibraryMonitorDuringScan:    systemXML.EnableLibraryMonitorDuringScan,
		EnableAutomaticRefresh:            systemXML.EnableAutomaticRefresh,
		EnableNotifications:               systemXML.EnableNotifications,
		EnableExternalMediaPlayback:       systemXML.EnableExternalMediaPlayback,
		EnableScheduledTask:               systemXML.EnableScheduledTask,
		EnableTranscoding:                 systemXML.EnableTranscoding,
		TranscodingTempPath:               systemXML.TranscodingTempPath,
		TempDirKeepIdleTimeMin:            systemXML.TempDirKeepIdleTimeMin,
		EnableQuickConnect:                systemXML.EnableQuickConnect,
		EnableUserPasswordReset:           systemXML.EnableUserPasswordReset,
		EncodingOptions:                   systemXML.EncodingOptions,
		DefaultAuthenticationProtocol:     "MediaBrowser",
		DefaultAppName:                    "Kabletown",
		DefaultAppVersion:                 "1.0.0",
	}

	// Set defaults for empty values
	if config.AppName == "" {
		config.AppName = config.DefaultAppName
	}
	if config.AppVersion == "" {
		config.AppVersion = config.DefaultAppVersion
	}
	if config.AuthenticationProtocol == "" {
		config.AuthenticationProtocol = config.DefaultAuthenticationProtocol
	}
	if config.ImageExtractionIntervalSec == 0 {
		config.ImageExtractionIntervalSec = 300
	}
	if config.ImageExtractionMaxCount == 0 {
		config.ImageExtractionMaxCount = 10
	}
	if config.TempDirKeepIdleTimeMin == 0 {
		config.TempDirKeepIdleTimeMin = 20
	}
	if !config.EnableQuickConnect {
		config.EnableQuickConnect = true
	}
	if !config.EnableUserPasswordReset {
		config.EnableUserPasswordReset = true
	}

	return config, nil
}

// AuthProtocol is the authentication protocol used by Jellyfin-style services
const AuthProtocol = "MediaBrowser"

// GetConfigWithDefaults returns either the loaded config or defaults
func GetConfigWithDefaults(config *SystemConfig) *SystemConfig {
	if config == nil {
		return DefaultSystemConfig()
	}
	return config
}

// WaitBetweenTasks returns the interval between scheduled tasks (in seconds)
func (sc *SystemConfig) WaitBetweenTasks() time.Duration {
	// Default is 5 seconds
	return 5 * time.Second
}