// Package dto contains shared data transfer objects used across Kabletown services.
// QueryResult lives in the response package — do not redefine it here.
package dto

// ItemTypeValue is the Jellyfin item type enum serialized as a string.
type ItemTypeValue int8

const (
	ItemTypeValueAggregateFolder    ItemTypeValue = 0
	ItemTypeValueAudio              ItemTypeValue = 1
	ItemTypeValueAudioBook          ItemTypeValue = 2
	ItemTypeValueBasePluginFolder   ItemTypeValue = 3
	ItemTypeValueBook               ItemTypeValue = 4
	ItemTypeValueBoxSet             ItemTypeValue = 5
	ItemTypeValueChannel            ItemTypeValue = 6
	ItemTypeValueChannelFolderItem  ItemTypeValue = 7
	ItemTypeValueCollectionFolder   ItemTypeValue = 8
	ItemTypeValueEpisode            ItemTypeValue = 9
	ItemTypeValueFolder             ItemTypeValue = 10
	ItemTypeValueGenre              ItemTypeValue = 11
	ItemTypeValueManualPlaylistsFolder ItemTypeValue = 12
	ItemTypeValueMovie              ItemTypeValue = 13
	ItemTypeValueLiveTvChannel      ItemTypeValue = 14
	ItemTypeValueLiveTvProgram      ItemTypeValue = 15
	ItemTypeValueMusicAlbum         ItemTypeValue = 16
	ItemTypeValueMusicArtist        ItemTypeValue = 17
	ItemTypeValueMusicGenre         ItemTypeValue = 18
	ItemTypeValueMusicVideo         ItemTypeValue = 19
	ItemTypeValuePerson             ItemTypeValue = 20
	ItemTypeValuePhoto              ItemTypeValue = 21
	ItemTypeValuePhotoAlbum         ItemTypeValue = 22
	ItemTypeValuePlaylist           ItemTypeValue = 23
	ItemTypeValuePlaylistsFolder    ItemTypeValue = 24
	ItemTypeValueProgram            ItemTypeValue = 25
	ItemTypeValueRecording          ItemTypeValue = 26
	ItemTypeValueSeason             ItemTypeValue = 27
	ItemTypeValueSeries             ItemTypeValue = 28
	ItemTypeValueStudio             ItemTypeValue = 29
	ItemTypeValueTrailer            ItemTypeValue = 30
	ItemTypeValueTvChannel          ItemTypeValue = 31
	ItemTypeValueTvProgram          ItemTypeValue = 32
	ItemTypeValueUserRootFolder     ItemTypeValue = 33
	ItemTypeValueUserView           ItemTypeValue = 34
	ItemTypeValueVideo              ItemTypeValue = 35
	ItemTypeValueYear               ItemTypeValue = 36
)

// ItemTypeValueString maps ItemTypeValue → its JSON string representation.
var ItemTypeValueString = map[ItemTypeValue]string{
	ItemTypeValueAggregateFolder:    "AggregateFolder",
	ItemTypeValueAudio:              "Audio",
	ItemTypeValueAudioBook:          "AudioBook",
	ItemTypeValueBasePluginFolder:   "BasePluginFolder",
	ItemTypeValueBook:               "Book",
	ItemTypeValueBoxSet:             "BoxSet",
	ItemTypeValueChannel:            "Channel",
	ItemTypeValueChannelFolderItem:  "ChannelFolderItem",
	ItemTypeValueCollectionFolder:   "CollectionFolder",
	ItemTypeValueEpisode:            "Episode",
	ItemTypeValueFolder:             "Folder",
	ItemTypeValueGenre:              "Genre",
	ItemTypeValueManualPlaylistsFolder: "ManualPlaylistsFolder",
	ItemTypeValueMovie:              "Movie",
	ItemTypeValueLiveTvChannel:      "LiveTvChannel",
	ItemTypeValueLiveTvProgram:      "LiveTvProgram",
	ItemTypeValueMusicAlbum:         "MusicAlbum",
	ItemTypeValueMusicArtist:        "MusicArtist",
	ItemTypeValueMusicGenre:         "MusicGenre",
	ItemTypeValueMusicVideo:         "MusicVideo",
	ItemTypeValuePerson:             "Person",
	ItemTypeValuePhoto:              "Photo",
	ItemTypeValuePhotoAlbum:         "PhotoAlbum",
	ItemTypeValuePlaylist:           "Playlist",
	ItemTypeValuePlaylistsFolder:    "PlaylistsFolder",
	ItemTypeValueProgram:            "Program",
	ItemTypeValueRecording:          "Recording",
	ItemTypeValueSeason:             "Season",
	ItemTypeValueSeries:             "Series",
	ItemTypeValueStudio:             "Studio",
	ItemTypeValueTrailer:            "Trailer",
	ItemTypeValueTvChannel:          "TvChannel",
	ItemTypeValueTvProgram:          "TvProgram",
	ItemTypeValueUserRootFolder:     "UserRootFolder",
	ItemTypeValueUserView:           "UserView",
	ItemTypeValueVideo:              "Video",
	ItemTypeValueYear:               "Year",
}

// ItemTypeValueByName maps string name → ItemTypeValue (for deserialization).
var ItemTypeValueByName = map[string]ItemTypeValue{
	"AggregateFolder":    ItemTypeValueAggregateFolder,
	"Audio":              ItemTypeValueAudio,
	"AudioBook":          ItemTypeValueAudioBook,
	"BasePluginFolder":   ItemTypeValueBasePluginFolder,
	"Book":               ItemTypeValueBook,
	"BoxSet":             ItemTypeValueBoxSet,
	"Channel":            ItemTypeValueChannel,
	"ChannelFolderItem":  ItemTypeValueChannelFolderItem,
	"CollectionFolder":   ItemTypeValueCollectionFolder,
	"Episode":            ItemTypeValueEpisode,
	"Folder":             ItemTypeValueFolder,
	"Genre":              ItemTypeValueGenre,
	"ManualPlaylistsFolder": ItemTypeValueManualPlaylistsFolder,
	"Movie":              ItemTypeValueMovie,
	"LiveTvChannel":      ItemTypeValueLiveTvChannel,
	"LiveTvProgram":      ItemTypeValueLiveTvProgram,
	"MusicAlbum":         ItemTypeValueMusicAlbum,
	"MusicArtist":        ItemTypeValueMusicArtist,
	"MusicGenre":         ItemTypeValueMusicGenre,
	"MusicVideo":         ItemTypeValueMusicVideo,
	"Person":             ItemTypeValuePerson,
	"Photo":              ItemTypeValuePhoto,
	"PhotoAlbum":         ItemTypeValuePhotoAlbum,
	"Playlist":           ItemTypeValuePlaylist,
	"PlaylistsFolder":    ItemTypeValuePlaylistsFolder,
	"Program":            ItemTypeValueProgram,
	"Recording":          ItemTypeValueRecording,
	"Season":             ItemTypeValueSeason,
	"Series":             ItemTypeValueSeries,
	"Studio":             ItemTypeValueStudio,
	"Trailer":            ItemTypeValueTrailer,
	"TvChannel":          ItemTypeValueTvChannel,
	"TvProgram":          ItemTypeValueTvProgram,
	"UserRootFolder":     ItemTypeValueUserRootFolder,
	"UserView":           ItemTypeValueUserView,
	"Video":              ItemTypeValueVideo,
	"Year":               ItemTypeValueYear,
}

// ImageTags maps image type names to their tag (etag) values.
type ImageTags map[string]string

// UserDataDto carries playback state for a user/item pair.
type UserDataDto struct {
	Rating                float64 `json:"Rating,omitempty"`
	PlayedPercentage      float64 `json:"PlayedPercentage,omitempty"`
	UnplayedItemCount     int     `json:"UnplayedItemCount,omitempty"`
	PlaybackPositionTicks int64   `json:"PlaybackPositionTicks"`
	PlayCount             int     `json:"PlayCount"`
	IsFavorite            bool    `json:"IsFavorite"`
	Likes                 *bool   `json:"Likes,omitempty"`
	LastPlayedDate        string  `json:"LastPlayedDate,omitempty"`
	Played                bool    `json:"Played"`
	Key                   string  `json:"Key"`
	ItemId                string  `json:"ItemId,omitempty"`
}

// MediaSourceInfo describes a playable source for a library item.
type MediaSourceInfo struct {
	Protocol                string            `json:"Protocol"`
	Id                      string            `json:"Id"`
	Path                    string            `json:"Path,omitempty"`
	EncoderPath             string            `json:"EncoderPath,omitempty"`
	EncoderProtocol         string            `json:"EncoderProtocol,omitempty"`
	Type                    string            `json:"Type"`
	Container               string            `json:"Container,omitempty"`
	Size                    int64             `json:"Size,omitempty"`
	Name                    string            `json:"Name,omitempty"`
	IsRemote                bool              `json:"IsRemote"`
	ETag                    string            `json:"ETag,omitempty"`
	RunTimeTicks            int64             `json:"RunTimeTicks,omitempty"`
	ReadAtNativeFramerate   bool              `json:"ReadAtNativeFramerate"`
	IgnoreDts               bool              `json:"IgnoreDts"`
	IgnoreIndex             bool              `json:"IgnoreIndex"`
	GenPtsInput             bool              `json:"GenPtsInput"`
	SupportsTranscoding     bool              `json:"SupportsTranscoding"`
	SupportsDirectStream    bool              `json:"SupportsDirectStream"`
	SupportsDirectPlay      bool              `json:"SupportsDirectPlay"`
	IsInfiniteStream        bool              `json:"IsInfiniteStream"`
	RequiresOpening         bool              `json:"RequiresOpening"`
	OpenToken               string            `json:"OpenToken,omitempty"`
	RequiresClosing         bool              `json:"RequiresClosing"`
	LiveStreamId            string            `json:"LiveStreamId,omitempty"`
	BufferMs                int               `json:"BufferMs,omitempty"`
	RequiresLooping         bool              `json:"RequiresLooping"`
	SupportsProbing         bool              `json:"SupportsProbing"`
	VideoType               string            `json:"VideoType,omitempty"`
	IsoType                 string            `json:"IsoType,omitempty"`
	Video3DFormat           string            `json:"Video3DFormat,omitempty"`
	Timestamp               string            `json:"Timestamp,omitempty"`
	MediaStreams             []MediaStreamInfo `json:"MediaStreams,omitempty"`
	MediaAttachments        []interface{}     `json:"MediaAttachments,omitempty"`
	Formats                 []string          `json:"Formats,omitempty"`
	Bitrate                 int               `json:"Bitrate,omitempty"`
	FallbackMaxStreamingBitrate int           `json:"FallbackMaxStreamingBitrate,omitempty"`
	TranscodingSubProtocol  string            `json:"TranscodingSubProtocol,omitempty"`
	TranscodingUrl          string            `json:"TranscodingUrl,omitempty"`
	TranscodingContainer    string            `json:"TranscodingContainer,omitempty"`
	AnalyzeDurationMs       int               `json:"AnalyzeDurationMs,omitempty"`
	DefaultAudioStreamIndex int               `json:"DefaultAudioStreamIndex,omitempty"`
	DefaultSubtitleStreamIndex int            `json:"DefaultSubtitleStreamIndex,omitempty"`
}

// MediaStreamInfo describes a single audio/video/subtitle stream.
type MediaStreamInfo struct {
	Codec                  string  `json:"Codec,omitempty"`
	CodecTag               string  `json:"CodecTag,omitempty"`
	Language               string  `json:"Language,omitempty"`
	ColorRange             string  `json:"ColorRange,omitempty"`
	ColorSpace             string  `json:"ColorSpace,omitempty"`
	ColorTransfer          string  `json:"ColorTransfer,omitempty"`
	ColorPrimaries         string  `json:"ColorPrimaries,omitempty"`
	DvVersionMajor         int     `json:"DvVersionMajor,omitempty"`
	DvVersionMinor         int     `json:"DvVersionMinor,omitempty"`
	DvProfile              int     `json:"DvProfile,omitempty"`
	DvLevel                int     `json:"DvLevel,omitempty"`
	RpuPresentFlag         int     `json:"RpuPresentFlag,omitempty"`
	ElPresentFlag          int     `json:"ElPresentFlag,omitempty"`
	BlPresentFlag          int     `json:"BlPresentFlag,omitempty"`
	DvBlSignalCompatibilityId int  `json:"DvBlSignalCompatibilityId,omitempty"`
	Comment                string  `json:"Comment,omitempty"`
	TimeBase               string  `json:"TimeBase,omitempty"`
	CodecTimeBase          string  `json:"CodecTimeBase,omitempty"`
	Title                  string  `json:"Title,omitempty"`
	VideoRange             string  `json:"VideoRange,omitempty"`
	VideoRangeType         string  `json:"VideoRangeType,omitempty"`
	VideoDoViTitle         string  `json:"VideoDoViTitle,omitempty"`
	LocalizedUndefined     string  `json:"LocalizedUndefined,omitempty"`
	LocalizedDefault       string  `json:"LocalizedDefault,omitempty"`
	LocalizedForced        string  `json:"LocalizedForced,omitempty"`
	LocalizedExternal      string  `json:"LocalizedExternal,omitempty"`
	LocalizedHearingImpaired string `json:"LocalizedHearingImpaired,omitempty"`
	DisplayTitle           string  `json:"DisplayTitle,omitempty"`
	NalLengthSize          string  `json:"NalLengthSize,omitempty"`
	IsInterlaced           bool    `json:"IsInterlaced"`
	IsAVC                  bool    `json:"IsAVC,omitempty"`
	ChannelLayout          string  `json:"ChannelLayout,omitempty"`
	BitRate                int     `json:"BitRate,omitempty"`
	BitDepth               int     `json:"BitDepth,omitempty"`
	RefFrames              int     `json:"RefFrames,omitempty"`
	PacketLength           int     `json:"PacketLength,omitempty"`
	Channels               int     `json:"Channels,omitempty"`
	SampleRate             int     `json:"SampleRate,omitempty"`
	IsDefault              bool    `json:"IsDefault"`
	IsForced               bool    `json:"IsForced"`
	IsHearingImpaired      bool    `json:"IsHearingImpaired"`
	Height                 int     `json:"Height,omitempty"`
	Width                  int     `json:"Width,omitempty"`
	AverageFrameRate       float64 `json:"AverageFrameRate,omitempty"`
	RealFrameRate          float64 `json:"RealFrameRate,omitempty"`
	Profile                string  `json:"Profile,omitempty"`
	Type                   string  `json:"Type"`
	AspectRatio            string  `json:"AspectRatio,omitempty"`
	Index                  int     `json:"Index"`
	Score                  int     `json:"Score,omitempty"`
	IsExternal             bool    `json:"IsExternal"`
	DeliveryMethod         string  `json:"DeliveryMethod,omitempty"`
	DeliveryUrl            string  `json:"DeliveryUrl,omitempty"`
	IsExternalUrl          bool    `json:"IsExternalUrl,omitempty"`
	IsTextSubtitleStream   bool    `json:"IsTextSubtitleStream"`
	SupportsExternalStream bool    `json:"SupportsExternalStream"`
	Path                   string  `json:"Path,omitempty"`
	PixelFormat            string  `json:"PixelFormat,omitempty"`
	Level                  float64 `json:"Level,omitempty"`
	IsAnamorphic           bool    `json:"IsAnamorphic,omitempty"`
}

// PersonInfo represents a cast/crew member attached to a library item.
type PersonInfo struct {
	Name            string `json:"Name"`
	Id              string `json:"Id,omitempty"`
	Role            string `json:"Role,omitempty"`
	Type            string `json:"Type,omitempty"`
	PrimaryImageTag string `json:"PrimaryImageTag,omitempty"`
	ImageBlurHashes map[string]map[string]string `json:"ImageBlurHashes,omitempty"`
}

// ItemValue is a lightweight name/id pair used in stub lists (genres, studios, etc.).
type ItemValue struct {
	Name string `json:"Name"`
	Id   string `json:"Id"`
}

// UserConfiguration holds per-user playback and display preferences.
type UserConfiguration struct {
	AudioLanguagePreference    string   `json:"AudioLanguagePreference,omitempty"`
	PlayDefaultAudioTrack      bool     `json:"PlayDefaultAudioTrack"`
	SubtitleLanguagePreference string   `json:"SubtitleLanguagePreference,omitempty"`
	DisplayMissingEpisodes     bool     `json:"DisplayMissingEpisodes"`
	GroupedFolders             []string `json:"GroupedFolders,omitempty"`
	SubtitleMode               string   `json:"SubtitleMode"`
	DisplayCollectionsView     bool     `json:"DisplayCollectionsView"`
	EnableLocalPassword        bool     `json:"EnableLocalPassword"`
	OrderedViews               []string `json:"OrderedViews,omitempty"`
	LatestItemsExcludes        []string `json:"LatestItemsExcludes,omitempty"`
	MyMediaExcludes            []string `json:"MyMediaExcludes,omitempty"`
	HidePlayedInLatest         bool     `json:"HidePlayedInLatest"`
	RememberAudioSelections    bool     `json:"RememberAudioSelections"`
	RememberSubtitleSelections bool     `json:"RememberSubtitleSelections"`
	EnableNextEpisodeAutoPlay  bool     `json:"EnableNextEpisodeAutoPlay"`
	CastReceiverId             string   `json:"CastReceiverId,omitempty"`
}

// UserPolicy holds permission/policy settings for a user.
type UserPolicy struct {
	IsAdministrator                  bool     `json:"IsAdministrator"`
	IsHidden                         bool     `json:"IsHidden"`
	IsDisabled                       bool     `json:"IsDisabled"`
	MaxParentalRating                int      `json:"MaxParentalRating,omitempty"`
	BlockedTags                      []string `json:"BlockedTags,omitempty"`
	AllowedTags                      []string `json:"AllowedTags,omitempty"`
	EnableUserPreferenceAccess       bool     `json:"EnableUserPreferenceAccess"`
	AccessSchedules                  []interface{} `json:"AccessSchedules,omitempty"`
	BlockUnratedItems                []string `json:"BlockUnratedItems,omitempty"`
	EnableRemoteControlOfOtherUsers  bool     `json:"EnableRemoteControlOfOtherUsers"`
	EnableSharedDeviceControl        bool     `json:"EnableSharedDeviceControl"`
	EnableRemoteAccess               bool     `json:"EnableRemoteAccess"`
	EnableLiveTvManagement           bool     `json:"EnableLiveTvManagement"`
	EnableLiveTvAccess               bool     `json:"EnableLiveTvAccess"`
	EnableMediaPlayback              bool     `json:"EnableMediaPlayback"`
	EnableAudioPlaybackTranscoding   bool     `json:"EnableAudioPlaybackTranscoding"`
	EnableVideoPlaybackTranscoding   bool     `json:"EnableVideoPlaybackTranscoding"`
	EnablePlaybackRemuxing           bool     `json:"EnablePlaybackRemuxing"`
	ForceRemoteSourceTranscoding     bool     `json:"ForceRemoteSourceTranscoding"`
	EnableContentDeletion            bool     `json:"EnableContentDeletion"`
	EnableContentDeletionFromFolders []string `json:"EnableContentDeletionFromFolders,omitempty"`
	EnableContentDownloading         bool     `json:"EnableContentDownloading"`
	EnableSyncTranscoding            bool     `json:"EnableSyncTranscoding"`
	EnableMediaConversion            bool     `json:"EnableMediaConversion"`
	EnabledDevices                   []string `json:"EnabledDevices,omitempty"`
	EnableAllDevices                 bool     `json:"EnableAllDevices"`
	EnabledChannels                  []string `json:"EnabledChannels,omitempty"`
	EnableAllChannels                bool     `json:"EnableAllChannels"`
	EnabledFolders                   []string `json:"EnabledFolders,omitempty"`
	EnableAllFolders                 bool     `json:"EnableAllFolders"`
	InvalidLoginAttemptCount         int      `json:"InvalidLoginAttemptCount"`
	LoginAttemptsBeforeLockout       int      `json:"LoginAttemptsBeforeLockout"`
	MaxActiveSessions                int      `json:"MaxActiveSessions"`
	EnablePublicSharing              bool     `json:"EnablePublicSharing"`
	BlockedMediaFolders              []string `json:"BlockedMediaFolders,omitempty"`
	BlockedChannels                  []string `json:"BlockedChannels,omitempty"`
	RemoteClientBitrateLimit         int      `json:"RemoteClientBitrateLimit"`
	AuthenticationProviderId         string   `json:"AuthenticationProviderId"`
	PasswordResetProviderId          string   `json:"PasswordResetProviderId"`
	SyncPlayAccess                   string   `json:"SyncPlayAccess"`
}

// DeviceInfo describes a client device that has connected to the server.
type DeviceInfo struct {
	Name             string `json:"Name"`
	CustomName       string `json:"CustomName,omitempty"`
	AccessToken      string `json:"AccessToken,omitempty"`
	Id               string `json:"Id"`
	LastUserName     string `json:"LastUserName,omitempty"`
	AppName          string `json:"AppName,omitempty"`
	AppVersion       string `json:"AppVersion,omitempty"`
	LastUserId       string `json:"LastUserId,omitempty"`
	DateLastActivity string `json:"DateLastActivity,omitempty"`
	IconUrl          string `json:"IconUrl,omitempty"`
}

// PlayerStateInfo holds current playback state for a session.
type PlayerStateInfo struct {
	PositionTicks    int64  `json:"PositionTicks,omitempty"`
	CanSeek          bool   `json:"CanSeek"`
	IsPaused         bool   `json:"IsPaused"`
	IsMuted          bool   `json:"IsMuted"`
	VolumeLevel      int    `json:"VolumeLevel,omitempty"`
	AudioStreamIndex int    `json:"AudioStreamIndex,omitempty"`
	SubtitleStreamIndex int `json:"SubtitleStreamIndex,omitempty"`
	MediaSourceId    string `json:"MediaSourceId,omitempty"`
	PlayMethod       string `json:"PlayMethod,omitempty"`
	RepeatMode       string `json:"RepeatMode,omitempty"`
	LiveStreamId     string `json:"LiveStreamId,omitempty"`
}

// SessionInfo describes an active playback session.
type SessionInfo struct {
	PlayState               *PlayerStateInfo `json:"PlayState,omitempty"`
	AdditionalUsers         []interface{}    `json:"AdditionalUsers,omitempty"`
	Capabilities            interface{}      `json:"Capabilities,omitempty"`
	RemoteEndPoint          string           `json:"RemoteEndPoint,omitempty"`
	PlayableMediaTypes      []string         `json:"PlayableMediaTypes,omitempty"`
	Id                      string           `json:"Id"`
	UserId                  string           `json:"UserId"`
	UserName                string           `json:"UserName,omitempty"`
	Client                  string           `json:"Client,omitempty"`
	LastActivityDate        string           `json:"LastActivityDate,omitempty"`
	LastPlaybackCheckIn     string           `json:"LastPlaybackCheckIn,omitempty"`
	DeviceName              string           `json:"DeviceName,omitempty"`
	DeviceType              string           `json:"DeviceType,omitempty"`
	NowPlayingItem          interface{}      `json:"NowPlayingItem,omitempty"`
	NowViewingItem          interface{}      `json:"NowViewingItem,omitempty"`
	DeviceId                string           `json:"DeviceId,omitempty"`
	ApplicationVersion      string           `json:"ApplicationVersion,omitempty"`
	TranscodingInfo         interface{}      `json:"TranscodingInfo,omitempty"`
	IsActive                bool             `json:"IsActive"`
	SupportsMediaControl    bool             `json:"SupportsMediaControl"`
	SupportsRemoteControl   bool             `json:"SupportsRemoteControl"`
	QueueableMediaTypes     []string         `json:"QueueableMediaTypes,omitempty"`
	HasCustomDeviceName     bool             `json:"HasCustomDeviceName"`
	PlaylistItemId          string           `json:"PlaylistItemId,omitempty"`
	ServerId                string           `json:"ServerId,omitempty"`
	UserPrimaryImageTag     string           `json:"UserPrimaryImageTag,omitempty"`
}

// BaseItemDto is the large DTO returned by library/browse endpoints.
type BaseItemDto struct {
	Name                         string            `json:"Name"`
	OriginalTitle                string            `json:"OriginalTitle,omitempty"`
	ServerId                     string            `json:"ServerId,omitempty"`
	Id                           string            `json:"Id"`
	Etag                         string            `json:"Etag,omitempty"`
	SourceType                   string            `json:"SourceType,omitempty"`
	PlaylistItemId               string            `json:"PlaylistItemId,omitempty"`
	DateCreated                  string            `json:"DateCreated,omitempty"`
	DateLastMediaAdded           string            `json:"DateLastMediaAdded,omitempty"`
	ExtraType                    string            `json:"ExtraType,omitempty"`
	AirsBeforeSeasonNumber       int               `json:"AirsBeforeSeasonNumber,omitempty"`
	AirsAfterSeasonNumber        int               `json:"AirsAfterSeasonNumber,omitempty"`
	AirsBeforeEpisodeNumber      int               `json:"AirsBeforeEpisodeNumber,omitempty"`
	CanDelete                    bool              `json:"CanDelete,omitempty"`
	CanDownload                  bool              `json:"CanDownload,omitempty"`
	HasLyrics                    bool              `json:"HasLyrics,omitempty"`
	HasSubtitles                 bool              `json:"HasSubtitles,omitempty"`
	PreferredMetadataLanguage    string            `json:"PreferredMetadataLanguage,omitempty"`
	PreferredMetadataCountryCode string            `json:"PreferredMetadataCountryCode,omitempty"`
	Container                    string            `json:"Container,omitempty"`
	SortName                     string            `json:"SortName,omitempty"`
	ForcedSortName               string            `json:"ForcedSortName,omitempty"`
	Video3DFormat                string            `json:"Video3DFormat,omitempty"`
	PremiereDate                 string            `json:"PremiereDate,omitempty"`
	ExternalUrls                 []interface{}     `json:"ExternalUrls,omitempty"`
	MediaSources                 []MediaSourceInfo `json:"MediaSources,omitempty"`
	CriticRating                 float64           `json:"CriticRating,omitempty"`
	ProductionLocations          []string          `json:"ProductionLocations,omitempty"`
	Path                         string            `json:"Path,omitempty"`
	EnableMediaSourceDisplay     bool              `json:"EnableMediaSourceDisplay,omitempty"`
	OfficialRating               string            `json:"OfficialRating,omitempty"`
	CustomRating                 string            `json:"CustomRating,omitempty"`
	ChannelId                    string            `json:"ChannelId,omitempty"`
	ChannelName                  string            `json:"ChannelName,omitempty"`
	Overview                     string            `json:"Overview,omitempty"`
	Taglines                     []string          `json:"Taglines,omitempty"`
	Genres                       []string          `json:"Genres,omitempty"`
	CommunityRating              float64           `json:"CommunityRating,omitempty"`
	CumulativeRunTimeTicks       int64             `json:"CumulativeRunTimeTicks,omitempty"`
	RunTimeTicks                 int64             `json:"RunTimeTicks,omitempty"`
	PlayAccess                   string            `json:"PlayAccess,omitempty"`
	AspectRatio                  string            `json:"AspectRatio,omitempty"`
	ProductionYear               int               `json:"ProductionYear,omitempty"`
	IsPlaceHolder                bool              `json:"IsPlaceHolder,omitempty"`
	Number                       string            `json:"Number,omitempty"`
	ChannelNumber                string            `json:"ChannelNumber,omitempty"`
	IndexNumber                  int               `json:"IndexNumber,omitempty"`
	IndexNumberEnd               int               `json:"IndexNumberEnd,omitempty"`
	ParentIndexNumber            int               `json:"ParentIndexNumber,omitempty"`
	RemoteTrailers               []interface{}     `json:"RemoteTrailers,omitempty"`
	ProviderIds                  map[string]string `json:"ProviderIds,omitempty"`
	IsHD                         bool              `json:"IsHD,omitempty"`
	IsFolder                     bool              `json:"IsFolder"`
	ParentId                     string            `json:"ParentId,omitempty"`
	Type                         ItemTypeValue     `json:"Type"`
	People                       []PersonInfo      `json:"People,omitempty"`
	Studios                      []ItemValue       `json:"Studios,omitempty"`
	GenreItems                   []ItemValue       `json:"GenreItems,omitempty"`
	ParentLogoItemId             string            `json:"ParentLogoItemId,omitempty"`
	ParentBackdropItemId         string            `json:"ParentBackdropItemId,omitempty"`
	ParentBackdropImageTags      []string          `json:"ParentBackdropImageTags,omitempty"`
	LocalTrailerCount            int               `json:"LocalTrailerCount,omitempty"`
	UserData                     *UserDataDto      `json:"UserData,omitempty"`
	RecursiveItemCount           int               `json:"RecursiveItemCount,omitempty"`
	ChildCount                   int               `json:"ChildCount,omitempty"`
	SeriesName                   string            `json:"SeriesName,omitempty"`
	SeriesId                     string            `json:"SeriesId,omitempty"`
	SeasonId                     string            `json:"SeasonId,omitempty"`
	SpecialFeatureCount          int               `json:"SpecialFeatureCount,omitempty"`
	DisplayPreferencesId         string            `json:"DisplayPreferencesId,omitempty"`
	Status                       string            `json:"Status,omitempty"`
	AirTime                      string            `json:"AirTime,omitempty"`
	AirDays                      []string          `json:"AirDays,omitempty"`
	Tags                         []string          `json:"Tags,omitempty"`
	PrimaryImageAspectRatio      float64           `json:"PrimaryImageAspectRatio,omitempty"`
	Artists                      []string          `json:"Artists,omitempty"`
	ArtistItems                  []ItemValue       `json:"ArtistItems,omitempty"`
	Album                        string            `json:"Album,omitempty"`
	CollectionType               string            `json:"CollectionType,omitempty"`
	DisplayOrder                 string            `json:"DisplayOrder,omitempty"`
	AlbumId                      string            `json:"AlbumId,omitempty"`
	AlbumPrimaryImageTag         string            `json:"AlbumPrimaryImageTag,omitempty"`
	SeriesPrimaryImageTag        string            `json:"SeriesPrimaryImageTag,omitempty"`
	AlbumArtist                  string            `json:"AlbumArtist,omitempty"`
	AlbumArtists                 []ItemValue       `json:"AlbumArtists,omitempty"`
	SeasonName                   string            `json:"SeasonName,omitempty"`
	MediaStreams                  []MediaStreamInfo `json:"MediaStreams,omitempty"`
	VideoType                    string            `json:"VideoType,omitempty"`
	PartCount                    int               `json:"PartCount,omitempty"`
	MediaSourceCount             int               `json:"MediaSourceCount,omitempty"`
	ImageTags                    ImageTags         `json:"ImageTags,omitempty"`
	BackdropImageTags            []string          `json:"BackdropImageTags,omitempty"`
	ScreenshotImageTags          []string          `json:"ScreenshotImageTags,omitempty"`
	ParentLogoImageTag           string            `json:"ParentLogoImageTag,omitempty"`
	ParentArtItemId              string            `json:"ParentArtItemId,omitempty"`
	ParentArtImageTag            string            `json:"ParentArtImageTag,omitempty"`
	SeriesThumbImageTag          string            `json:"SeriesThumbImageTag,omitempty"`
	ImageBlurHashes              map[string]map[string]string `json:"ImageBlurHashes,omitempty"`
	SeriesStudio                 string            `json:"SeriesStudio,omitempty"`
	ParentThumbItemId            string            `json:"ParentThumbItemId,omitempty"`
	ParentThumbImageTag          string            `json:"ParentThumbImageTag,omitempty"`
	ParentPrimaryImageItemId     string            `json:"ParentPrimaryImageItemId,omitempty"`
	ParentPrimaryImageTag        string            `json:"ParentPrimaryImageTag,omitempty"`
	Chapters                     []interface{}     `json:"Chapters,omitempty"`
	Trickplay                    interface{}       `json:"Trickplay,omitempty"`
	LocationType                 string            `json:"LocationType,omitempty"`
	IsoType                      string            `json:"IsoType,omitempty"`
	MediaType                    string            `json:"MediaType,omitempty"`
	EndDate                      string            `json:"EndDate,omitempty"`
	LockedFields                 []string          `json:"LockedFields,omitempty"`
	TrailerCount                 int               `json:"TrailerCount,omitempty"`
	MovieCount                   int               `json:"MovieCount,omitempty"`
	SeriesCount                  int               `json:"SeriesCount,omitempty"`
	ProgramCount                 int               `json:"ProgramCount,omitempty"`
	EpisodeCount                 int               `json:"EpisodeCount,omitempty"`
	SongCount                    int               `json:"SongCount,omitempty"`
	AlbumCount                   int               `json:"AlbumCount,omitempty"`
	ArtistCount                  int               `json:"ArtistCount,omitempty"`
	MusicVideoCount              int               `json:"MusicVideoCount,omitempty"`
	LockData                     bool              `json:"LockData,omitempty"`
	Width                        int               `json:"Width,omitempty"`
	Height                       int               `json:"Height,omitempty"`
	CameraMake                   string            `json:"CameraMake,omitempty"`
	CameraModel                  string            `json:"CameraModel,omitempty"`
	Software                     string            `json:"Software,omitempty"`
	ExposureTime                 float64           `json:"ExposureTime,omitempty"`
	FocalLength                  float64           `json:"FocalLength,omitempty"`
	ImageOrientation             string            `json:"ImageOrientation,omitempty"`
	Aperture                     float64           `json:"Aperture,omitempty"`
	ShutterSpeed                 float64           `json:"ShutterSpeed,omitempty"`
	Latitude                     float64           `json:"Latitude,omitempty"`
	Longitude                    float64           `json:"Longitude,omitempty"`
	Altitude                     float64           `json:"Altitude,omitempty"`
	IsoSpeedRating               int               `json:"IsoSpeedRating,omitempty"`
	SeriesTimerId                string            `json:"SeriesTimerId,omitempty"`
	ProgramId                    string            `json:"ProgramId,omitempty"`
	ChannelPrimaryImageTag       string            `json:"ChannelPrimaryImageTag,omitempty"`
	StartDate                    string            `json:"StartDate,omitempty"`
	CompletionPercentage         float64           `json:"CompletionPercentage,omitempty"`
	IsRepeat                     bool              `json:"IsRepeat,omitempty"`
	EpisodeTitle                 string            `json:"EpisodeTitle,omitempty"`
	ChannelType                  string            `json:"ChannelType,omitempty"`
	Audio                        string            `json:"Audio,omitempty"`
	IsMovie                      bool              `json:"IsMovie,omitempty"`
	IsSports                     bool              `json:"IsSports,omitempty"`
	IsSeries                     bool              `json:"IsSeries,omitempty"`
	IsLive                       bool              `json:"IsLive,omitempty"`
	IsNews                       bool              `json:"IsNews,omitempty"`
	IsKids                       bool              `json:"IsKids,omitempty"`
	IsPremiere                   bool              `json:"IsPremiere,omitempty"`
	TimerId                      string            `json:"TimerId,omitempty"`
	NormalizationGain            float64           `json:"NormalizationGain,omitempty"`
	CurrentProgram               interface{}       `json:"CurrentProgram,omitempty"`
}
