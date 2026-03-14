# Data Model Dictionary

This document defines all data structures used across Kabletown microservices, including field types, constraints, and usage patterns.

---

## Authentication Models

### User

**Purpose:** Core user account

| Field | Type | Constraints | Description |
|-------|------|-------------|---|-|
| `Id` | string (UUID) | PRIMARY KEY | User identifier (dashed format) |
| `Username` | string | UNIQUE(100), NOT NULL | Login username, case-insensitive |
| `Email` | string | UNIQUE(255), NULLABLE | User email address |
| `PasswordHash` | string | CHAR(64), NOT NULL | SHA256 hash (hex-encoded) |
| `IsAdmin` | bool | DEFAULT FALSE | Admin permission flag |
| `EnableUser` | bool | DEFAULT TRUE | Account active flag |
| `HasPassword` | bool | DEFAULT TRUE | Password set flag |
| `DateCreated` | timestamp | DEFAULT CURRENT_TIMESTAMP | Account creation time |
| `DateLastLogin` | timestamp | NULLABLE | Last successful login time |

**Usage in Code:**
```go
type User struct {
    Id            string     `json:"Id"`
    Username      string     `json:"Username"`
    Email         string     `json:"Email"`
    PasswordHash  string     `json:"-"` // Never serialize
    IsAdmin       bool       `json:"IsAdmin"`
    EnableUser    bool       `json:"EnableUser"`
    HasPassword   bool       `json:"HasPassword"`
    DateCreated   time.Time  `json:"DateCreated"`
    DateLastLogin *time.Time `json:"DateLastLogin"`
}
```

**Index Patterns:**
```sql
-- By username (login lookup)
CREATE UNIQUE INDEX idx_users_username ON users(Username);

-- By enable status (list active users)
CREATE INDEX idx_users_enable ON users(EnableUser) WHERE EnableUser = TRUE;
```

---

### API Key

**Purpose:** Token-to-user mapping for API authentication

| Field | Type | Constraints | Description |
|-------|------|-------------|---|-|
| `Id` | string (UUID) | PRIMARY KEY | API key record ID |
| `UserId` | string (UUID) | FOREIGN KEY → users(Id), NOT NULL | Owner of this token |
| `DeviceId` | string | CHAR(64), NOT NULL | Client device identifier |
| `Token` | string | CHAR(64), UNIQUE NOT NULL | 256-bit random hex token |
| `Name` | string | VARCHAR(255), NULLABLE | User-assigned device name |
| `AppName` | string | VARCHAR(255), NULLABLE | Client application name |
| `AppVersion` | string | VARCHAR(50), NULLABLE | Client application version |
| `DateCreated` | timestamp | DEFAULT CURRENT_TIMESTAMP | Token creation time |
| `DateLastUsed` | timestamp | DEFAULT CURRENT_TIMESTAMP | Last request timestamp |
| `IsActive` | bool | DEFAULT TRUE | Revocation flag |

**Usage in Code:**
```go
type APIKey struct {
    Id           string     `json:"Id"`
    UserId       string     `json:"UserId"`
    DeviceId     string     `json:"DeviceId"`
    Token        string     `json:"-"` // Never serialize (security)
    Name         string     `json:"Name"`
    AppName      string     `json:"AppName"`
    AppVersion   string     `json:"AppVersion"`
    DateCreated  time.Time  `json:"DateCreated"`
    DateLastUsed time.Time  `json:"DateLastUsed"`
    IsActive     bool       `json:"IsActive"`
}
```

**Index Patterns:**
```sql
-- Token validation (auth middleware)
CREATE UNIQUE INDEX idx_api_keys_token ON api_keys(Token);

-- List user devices
CREATE INDEX idx_api_keys_userid ON api_keys(UserId);

-- Device identification
CREATE INDEX idx_api_keys_deviceid ON api_keys(DeviceId);

-- Active key lookup
CREATE INDEX idx_api_keys_isactive ON api_keys(IsActive);
```

**API Token Format:**
```
Length: 64 hex characters (256 bits)
Generation: crypto/rand + hex encoding
Example: abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890
```

---

### Device

**Purpose:** Client device registration

| Field | Type | Constraints | Description |
|-------|------|-------------|---|-|
| `Id` | string (UUID) | PRIMARY KEY | Device record ID |
| `UserId` | string (UUID) | FOREIGN KEY → users(Id), NULLABLE | Owner (null for anonymous) |
| `Name` | string | VARCHAR(255), NULLABLE | User-assigned name |
| `DeviceId` | string | CHAR(64), UNIQUE | Client device identifier |
| `AppName` | string | VARCHAR(255), NULLABLE | Client app name |
| `AppVersion` | string | VARCHAR(50), NULLABLE | Client app version |
| `DateRegistered` | timestamp | DEFAULT CURRENT_TIMESTAMP | Registration time |
| `LastActivity` | timestamp | DEFAULT CURRENT_TIMESTAMP | Last activity time |

**Usage in Code:**
```go
type Device struct {
    Id              string     `json:"Id"`
    UserId          string     `json:"UserId"`
    Name            string     `json:"Name"`
    DeviceId        string     `json:"DeviceId"`
    AppName         string     `json:"AppName"`
    AppVersion      string     `json:"AppVersion"`
    DateRegistered  time.Time  `json:"DateRegistered"`
    LastActivity    time.Time  `json:"LastActivity"`
}
```

---

### UserPolicy

**Purpose:** User permission flags and limits

| Field | Type | Constraints | Description |
|-------|------|-------------|---|-|
| `UserId` | string (UUID) | PRIMARY KEY, FOREIGN KEY → users(Id) | User identifier |
| `EnableVideoTranscoding` | bool | DEFAULT TRUE | Allow video transcoding |
| `EnableAudioTranscoding` | bool | DEFAULT TRUE | Allow audio transcoding |
| `MaxStreamingBitrate` | int | NULLABLE | Max bitrate (bps) for streaming |
| `EnablePlaybackRemuxing` | bool | DEFAULT TRUE | Allow stream copy (remux) |

**Usage in Code:**
```go
type UserPolicy struct {
    UserId                     string `json:"UserId"`
    EnableVideoTranscoding     bool   `json:"EnableVideoTranscoding"`
    EnableAudioTranscoding     bool   `json:"EnableAudioTranscoding"`
    MaxStreamingBitrate        int    `json:"MaxStreamingBitrate"`
    EnablePlaybackRemuxing     bool   `json:"EnablePlaybackRemuxing"`
}
```

**Index Patterns:**
```sql
-- Primary key (one row per user)
-- No additional indexes needed (always accessed by UserId)
```

---

## Item Metadata Models (Jellyfin-Compatible)

### BaseItem

**Purpose:** Core media entity (all media types)

| Field | Type | Constraints | Description |
|-------|------|-------------|---|-|
| `Id` | string (UUID) | PRIMARY KEY | Item identifier |
| `Name` | string | VARCHAR(255), NOT NULL | Display name |
| `Type` | string | VARCHAR(50), NOT NULL | BaseItemKind enum |
| `IsFolder` | bool | DEFAULT FALSE | Folder vs file |
| `ParentId` | string (UUID) | NULLABLE | Immediate parent folder |
| `TopParentId` | string (UUID) | NULLABLE | Root library folder |
| `Path` | string | VARCHAR(500), NULLABLE | File system path |
| `Container` | string | VARCHAR(100), NULLABLE | Media container (mkv, mp4) |
| `DurationTicks` | int64 | NULLABLE | Duration (100-ns units) |
| `Size` | int64 | NULLABLE | File size in bytes |
| `Width` | int | NULLABLE | Video width (pixels) |
| `Height` | int | NULLABLE | Video height (pixels) |
| `BitRate` | int | NULLABLE | Video bitrate (bps) |
| `ProductionYear` | int | NULLABLE | Release year |
| `PremiereDate` | timestamp | NULLABLE | Release date |
| `DateCreated` | timestamp | DEFAULT CURRENT_TIMESTAMP | Metadata creation |
| `DateModified` | timestamp | DEFAULT CURRENT_TIMESTAMP | Metadata update |
| `ExtraData` | json | NULLABLE | Provider IDs, custom fields |
| `AncestorIds` | text | NULLABLE | Recursive CTE result |
| `Overview` | string | TEXT, NULLABLE | Description/synopsis |
| `OfficialRating` | string | VARCHAR(50), NULLABLE | Content rating (PG-13, R) |

**Usage in Code:**
```go
type BaseItem struct {
    Id              string     `json:"Id"`
    Name            string     `json:"Name"`
    Type            string     `json:"Type"`
    IsFolder        bool       `json:"IsFolder"`
    ParentId        string     `json:"ParentId"`
    TopParentId     string     `json:"TopParentId"`
    Path            string     `json:"Path"`
    Container       string     `json:"Container"`
    DurationTicks   int64      `json:"RunTimeTicks"`
    Size            int64      `json:"Size"`
    Width           int        `json:"Width"`
    Height          int        `json:"Height"`
    BitRate         int        `json:"BitRate"`
    ProductionYear  int        `json:"ProductionYear"`
    PremiereDate    *time.Time `json:"PremiereDate"`
    DateCreated     time.Time  `json:"DateCreated"`
    DateModified    time.Time  `json:"DateModified"`
    ExtraData       json.RawMessage `json:"ExtraData"`
    AncestorIds     string     `json:"AncestorIds"`
    Overview        string     `json:"Overview"`
    OfficialRating  string     `json:"OfficialRating"`
}

// BaseItemKind enum values
const (
    BaseItemKindMovie       = "Movie"
    BaseItemKindEpisode     = "Episode"
    BaseItemKindSeries      = "Series"
    BaseItemKindSeason      = "Season"
    BaseItemKindFolder      = "Folder"
    BaseItemKindPhoto       = "Photo"
    BaseItemKindAudio       = "Audio"
)
```

**Index Patterns (P7 Optimized):**
```sql
-- Core query: Library + Type filtering (most common)
CREATE INDEX idx_base_items_topparent_type ON base_items(TopParentId, Type);

-- Parent folder queries
CREATE INDEX idx_base_items_parentid ON base_items(ParentId);

-- Recently added (DateCreated DESC)
CREATE INDEX idx_base_items_datecreated ON base_items(DateCreated DESC);

-- Production year filtering
CREATE INDEX idx_base_items_year ON base_items(ProductionYear);

-- Full-text search (FTS5 or MySQL full-text)
CREATE FULLTEXT INDEX idx_base_items_search ON base_items(Name, Overview);
```

---

### MediaStream

**Purpose:** Video/audio/subtitle stream metadata

| Field | Type | Constraints | Description |
|-------|------|-------------|---|-|
| `Index` | int | NOT NULL | Stream index in container |
| `Codec` | string | VARCHAR(50), NOT NULL | Codec name (h264, aac, etc.) |
| `CodecTag` | string | VARCHAR(50), NULLABLE | Codec tag from container |
| `Language` | string | VARCHAR(10), NULLABLE | Language code (eng, spa) |
| `IsExternal` | bool | DEFAULT FALSE | External stream flag |
| `IsDefault` | bool | DEFAULT FALSE | Default stream flag |
| `IsForced` | bool | DEFAULT FALSE | Forced subtitle flag |
| `Type` | string | VARCHAR(20), NOT NULL | Video/Audio/Subtitle |
| `Width` | int | NULLABLE | Video width (pixels) |
| `Height` | int | NULLABLE | Video height (pixels) |
| `BitRate` | int | NULLABLE | Stream bitrate (bps) |
| `BitDepth` | int | NULLABLE | Video bit depth (8, 10) |
| `PixelFormat` | string | VARCHAR(50), NULLABLE | Pixel format (yuv420p) |
| `FrameRate` | float | NULLABLE | Frame rate (23.976, 29.97) |
| `ChannelCount` | int | NULLABLE | Audio channels (2, 6) |
| `SampleRate` | int | NULLABLE | Sample rate (44100, 48000) |
| `Profile` | string | VARCHAR(50), NULLABLE | Codec profile (main, high) |
| `Level` | float | NULLABLE | Codec level (3.1, 4.0) |

**Usage in Code:**
```go
type MediaStream struct {
    Index        int     `json:"Index"`
    Codec        string  `json:"Codec"`
    CodecTag     string  `json:"CodecTag"`
    Language     string  `json:"Language"`
    IsExternal   bool    `json:"IsExternal"`
    IsDefault    bool    `json:"IsDefault"`
    IsForced     bool    `json:"IsForced"`
    Type         string  `json:"Type"`
    Width        int     `json:"Width"`
    Height       int     `json:"Height"`
    BitRate      int     `json:"BitRate"`
    BitDepth     int     `json:"BitDepth"`
    PixelFormat  string  `json:"PixelFormat"`
    FrameRate    float64 `json:"FrameRate"`
    ChannelCount int     `json:"ChannelCount"`
    SampleRate   int     `json:"SampleRate"`
    Profile      string  `json:"Profile"`
    Level        float64 `json:"Level"`
}
```

---

### MediaSource

**Purpose:** Physical media source file

| Field | Type | Constraints | Description |
|-------|------|-------------|---|-|
| `Id` | string (UUID) | PRIMARY KEY | Source identifier |
| `ItemId` | string (UUID) | FOREIGN KEY → base_items(Id), NOT NULL | Parent item |
| `Path` | string | VARCHAR(500), NOT NULL | File system path |
| `Container` | string | VARCHAR(100), NOT NULL | Container format |
| `Size` | int64 | NOT NULL | File size (bytes) |
| `DurationTicks` | int64 | NULLABLE | Duration (100-ns units) |
| `BitRate` | int | NULLABLE | Total bitrate (bps) |
| `VideoStream` | json | NULLABLE | Video stream metadata |
| `AudioStreams` | json | NULLABLE | Array of audio streams |
| `SubtitleStreams` | json | NULLABLE | Array of subtitle streams |

**Usage in Code:**
```go
type MediaSource struct {
    Id           string          `json:"Id"`
    ItemId       string          `json:"ItemId"`
    Path         string          `json:"Path"`
    Container    string          `json:"Container"`
    Size         int64           `json:"Size"`
    DurationTicks int64          `json:"RunTimeTicks"`
    BitRate      int             `json:"BitRate"`
    VideoStream  *MediaStream    `json:"VideoStream"`
    AudioStreams []MediaStream   `json:"MediaStreams"` // Note: plural in API
    SubtitleStreams []MediaStream`json:"SubtitleStreams"`\n}
```

---

## P6 Filtering Models (Many-to-Many)

### ItemValue

**Purpose:** Normalized genre/studio/artist/collection/tag values (P6)

| Field | Type | Constraints | Description |
|-------|------|-------------|---|-|
| `ItemId` | string (UUID) | FOREIGN KEY → base_items(Id), PART OF PK |
| `ValueType` | int | NOT NULL, PART OF PK | 2=Genre, 3=Studio, 4=Artist, 5=Collection, 6=Tag |
| `Value` | string | VARCHAR(255), NOT NULL | Original display value |
| `NormalizedValue` | string | VARCHAR(255), NOT NULL, PART OF PK | Lowercase, trimmed, no diacritics |

**Usage in Code:**
```go
type ItemValue struct {
    ItemId           string `json:"ItemId"`
    ValueType        int    `json:"ValueType"`  // 2/3/4/5/6
    Value            string `json:"Value"`
    NormalizedValue  string `json:"NormalizedValue"`
}

// ItemValue types
const (
    ItemValueTypeGenre       = 2
    ItemValueTypeStudio      = 3
    ItemValueTypeArtist      = 4
    ItemValueTypeCollection  = 5
    ItemValueTypeTag         = 6
    ItemValueTypePerson      = 1
)

// Normalize function
func NormalizeValue(value string) string {
    value = strings.ToLower(strings.TrimSpace(value))
    value = removeDiacritics(value)  // café → cafe
    value = collapseSpaces(value)    // multiple space → single space
    return value
}
```

**Query Patterns:**
```sql
-- Get all movies in "Action" genre
SELECT b.*
FROM base_items b
INNER JOIN item_values iv ON b.Id = iv.ItemId
WHERE b.Type = 'Movie'
AND iv.ValueType = 2
AND iv.NormalizedValue = 'action';

-- Get all items with "Disney" as studio
SELECT b.*
FROM base_items b
INNER JOIN item_values iv ON b.Id = iv.ItemId
WHERE iv.ValueType = 3
AND iv.NormalizedValue = 'disney';

-- Get genre list for single item
SELECT v.Value
FROM item_values iv
WHERE iv.ItemId = 'uuid123'
AND iv.ValueType = 2;
```

**Index Patterns (P6 Optimization):**
```sql
-- Primary key (compound)
CREATE PRIMARY KEY idx_itemvalues_item_valuetype_normalized
  ON item_values(ItemId, ValueType, NormalizedValue);

-- Fast value lookup for filtering
CREATE INDEX idx_itemvalues_lookup ON item_values(ValueType, NormalizedValue);

-- All items with specific value (e.g., all action movies)
CREATE INDEX idx_itemvalues_valuetype ON item_values(ValueType);
```

---

## User Data Models (Playstate)

### UserData

**Purpose:** Per-user playstate, favorites, ratings

| Field | Type | Constraints | Description |
|-------|------|-------------|---|-|
| `UserId` | string (UUID) | FOREIGN KEY → users(Id), PART OF PK |
| `ItemId` | string (UUID) | FOREIGN KEY → base_items(Id), PART OF PK |
| `IsPlayed` | bool | DEFAULT FALSE | Watch completion flag |
| `IsFavorite` | bool | DEFAULT FALSE | User favorite flag |
| `PlaybackPositionTicks` | int64 | DEFAULT 0 | Resume position (100-ns units) |
| `Rating` | int | NULLABLE | User rating (1-10 scale) |
| `PlayCount` | int | DEFAULT 0 | Number of times played |
| `LastPlayedDate` | timestamp | NULLABLE | Last playback timestamp |

**Usage in Code:**
```go
type UserData struct {
    UserId               string  `json:"UserId"`
    ItemId               string  `json:"ItemId"`
    IsPlayed             bool    `json:"IsPlayed"`
    IsFavorite           bool    `json:"IsFavorite"`
    PlaybackPositionTicks int64  `json:"PlaybackPositionTicks"`
    Rating               *int    `json:"Rating"`  // Nullable
    PlayCount            int     `json:"PlayCount"`
    LastPlayedDate       *time.Time`json:"LastPlayedDate"`
}

// PlayState (DTO for API response)
type PlayState struct {
    IsPlaying        bool  `json:"IsPlaying"`
    IsPaused         bool  `json:"IsPaused"`
    IsActive         bool  `json:"IsActive"`
    CanSeek          bool  `json:"CanSeek"`
    PositionTicks    int64 `json:"PositionTicks"`
    PlayMethod       string`json:"PlayMethod"`  // DirectPlay, Transcode, DirectStream
}
```

**Query Patterns:**
```sql
-- User playstate for specific item
SELECT * FROM user_data
WHERE UserId = 'user123' AND ItemId = 'item456';

-- Recently played items (last 30 days)
SELECT b.*, ud.PlaybackPositionTicks, ud.IsPlayed
FROM base_items b
INNER JOIN user_data ud ON b.Id = ud.ItemId
WHERE ud.UserId = 'user123'
AND ud.PlaybackPositionTicks > 0
ORDER BY ud.LastPlayedDate DESC
LIMIT 20;

-- Unwatched items for Continue Watching row
SELECT b.*
FROM base_items b
INNER JOIN user_data ud ON b.Id = ud.ItemId
WHERE ud.UserId = 'user123'
AND ud.IsPlayed = FALSE
AND ud.PlaybackPositionTicks > 0
ORDER BY ud.LastPlayedDate DESC;

-- Favorite items
SELECT b.*
FROM base_items b
INNER JOIN user_data ud ON b.Id = ud.ItemId
WHERE ud.UserId = 'user123'
AND ud.IsFavorite = TRUE;
```

**Index Patterns:**
```sql
-- Primary key (compound - always query by both)
CREATE PRIMARY KEY idx_userdata_useritem
  ON user_data(UserId, ItemId);

-- Recently played (sorted by playback time)
CREATE INDEX idx_userdata_position
  ON user_data(UserId, PlaybackPositionTicks DESC);

-- Favorite items
CREATE INDEX idx_userdata_favorite
  ON user_data(UserId, IsFavorite) WHERE IsFavorite = TRUE;

-- Unwatched items
CREATE INDEX idx_userdata_unwatched
  ON user_data(UserId, IsPlayed) WHERE IsPlayed = FALSE;

-- Recent activity (LastPlayedDate sort)
CREATE INDEX idx_userdata_lastplayed
  ON user_data(UserId, LastPlayedDate DESC);
```

---

## Streaming Models (Transcode State)

### StreamState

**Purpose:** Transcode decision state (not persisted, in-memory only)

| Field | Type | Description |
|-------|------|-------------|
| `Request` | `VideoRequestDto` | Client request params |
| `MediaPath` | string | Source file path |
| `MediaSource` | `MediaSourceInfo` | Full media metadata |
| `VideoStream` | `MediaStream` | Selected video stream |
| `AudioStream` | `MediaStream` | Selected audio stream |
| `SubtitleStream` | `MediaStream` | Selected subtitle stream |
| `OutputVideoCodec` | string | Target video codec |
| `OutputAudioCodec` | string | Target audio codec |
| `OutputContainer` | string | Target container |
| `OutputFilePath` | string | Output file path (/transcode/)
| `OutputWidth` | int | Target resolution width |
| `OutputHeight` | int | Target resolution height |
| `OutputBitRate` | int | Target bitrate (bps) |
| `DirectStreamProvider` | boolean | Can stream directly? |
| `InputProtocol` | string | http/file/cdrom |

**Usage in Code:**
```go
type StreamState struct {
    Request          VideoRequestDto
    MediaPath        string
    MediaSource      *MediaSourceInfo
    VideoStream      *MediaStream
    AudioStream      *MediaStream
    SubtitleStream   *MediaStream
    OutputVideoCodec string
    OutputAudioCodec string
    OutputContainer  string
    OutputFilePath   string
    OutputWidth      int
    OutputHeight     int
    OutputBitRate    int
    DirectStreamProvider bool
    InputProtocol    string
    // ... 100+ more fields
}
```

---

### VideoRequestDto (Client Request)

**Purpose:** 48+ query parameters for streaming

| Field | Type | Required | Description |
|-------|------|:--------:|-------------|
| `Id` | UUID | Yes | Item ID |
| `Static` | bool | No | Stream copy only (default: false) |
| `MediaSourceId` | UUID | No | Select media version |
| `StartTimeTicks` | int64 | No | Start position (100-ns) |
| `VideoCodec` | string | No | Target video codec |
| `AudioCodec` | string | No | Target audio codec |
| `VideoBitRate` | int | No | Target bitrate (bps) |
| `MaxWidth` | int | No | Max resolution width |
| `MaxHeight` | int | No | Max resolution height |
| `SubtitleStreamIndex` | int | No | Subtitle stream index |
| `SubtitleMethod` | string | No | Embed/External/Hardsub |
| `DeviceId` | string | Yes | Device identifier (job tracking) |
| `PlaySessionId` | string | Yes | Session identifier (HLS tracking) |
| `MaxStreamingBitrate` | int | No | Max bitrate client can handle |
| `TranscodeReasons` | string | No | Client hints (comma-separated) |

**Usage in Code:**
```go
type VideoRequestDto struct {
    Id                    string `json:"Id"`
    Static                bool   `json:"Static"`
    MediaSourceId         string `json:"MediaSourceId"`
    StartTimeTicks        *int64 `json:"StartTimeTicks"`
    VideoCodec            string `json:"VideoCodec"`
    AudioCodec            string `json:"AudioCodec"`
    VideoBitRate          *int   `json:"VideoBitRate"`
    MaxWidth              *int   `json:"MaxWidth"`
    MaxHeight             *int   `json:"MaxHeight"`
    SubtitleStreamIndex   *int   `json:"SubtitleStreamIndex"`
    SubtitleMethod        string `json:"SubtitleMethod"`
    DeviceId              string `json:"DeviceId"`
    PlaySessionId         string `json:"PlaySessionId"`
    MaxStreamingBitrate   *int   `json:"MaxStreamingBitrate"`
    TranscodeReasons      string `json:"TranscodeReasons"`
    // ... 30+ more fields
}
```

---

### TranscodingJob

**Purpose:** FFmpeg process registry (in-memory only, not persisted)

| Field | Type | Description |
|-------|------|-------------|
| `Id` | string (UUID) | Job identifier |
| `Path` | string | Output file path |
| `Process` | `*os.Process` | FFmpeg subprocess handle |
| `JobType` | enum | HLS or Progressive |
| `IsLiveOutput` | bool | HLS flag (don't delete on disconnect) |
| `ActiveRequestCount` | int | Concurrent segment requests (HLS) |
| `PlaySessionId` | string | Session identifier |
| `DeviceId` | string | Client device identifier |
| `UserPaused` | bool | Client pause state |
| `LastPingTime` | time.Time | Last client checkin |
| `KillTimer` | `*time.Timer` | Auto-kill after timeout |
| `CancelFunc` | `context.CancelFunc` | Context cancellation |

**Usage in Code:**
```go
type TranscodingJob struct {
    Id                string
    Path              string
    Process           *os.Process
    JobType           TranscodingJobType // HLS | Progressive
    IsLiveOutput      bool
    ActiveRequestCount int
    PlaySessionId     string
    DeviceId          string
    UserPaused        bool
    LastPingTime      time.Time
    KillTimer         *time.Timer
    CancelFunc        context.CancelFunc
}

type TranscodingJobType int

const (
    TranscodingJobTypeHLS TranscodingJobType = iota
    TranscodingJobTypeProgressive
)
```

---

## DTO Models (API Responses)

### BaseItemDto

**Purpose:** API response format for item queries

```go
type BaseItemDto struct {
    Id                    string    `json:"Id"`
    Name                  string    `json:"Name"`
    OriginalTitle         string    `json:"OriginalTitle,omitempty"`
    Type                  string    `json:"Type"`
    IsFolder              bool      `json:"IsFolder"`
    ParentId              string    `json:"ParentId,omitempty"`
    TopParentId           string    `json:"TopParentId,omitempty"`
    Path                  string    `json:"Path,omitempty"`
    Container             string    `json:"Container,omitempty"`
    ChannelId             string    `json:"ChannelId,omitempty"`
    ChannelName           string    `json:"ChannelName,omitempty"`
    RunTimeTicks          int64     `json:"RunTimeTicks"`
    ProductionYear        int       `json:"ProductionYear,omitempty"`
    PremiereDate          *time.Time`json:"PremiereDate,omitempty"`
    OfficialRating        string    `json:"OfficialRating,omitempty"`
    Overview              string    `json:"Overview,omitempty"`
    PrimaryImageAspectRatio float32  `json:"PrimaryImageAspectRatio,omitempty"`
    CommunityRating       float32    `json:"CommunityRating,omitempty"`
    IndexNumber           int       `json:"IndexNumber,omitempty"`      // Episode number
    ParentIndexNumber     int       `json:"ParentIndexNumber,omitempty"` // Season number
    PlayState             *PlayState`json:"PlayState,omitempty"`
    UserData              *UserData `json:"UserData,omitempty"`
    MediaSources          []MediaSource`json:"MediaSources,omitempty"`
}
```

### QueryResult

**Purpose:** Paginated list response

```go
type QueryResult struct {
    Items             []BaseItemDto `json:"Items"`
    TotalRecordCount  int           `json:"TotalRecordCount"`
    StartIndex        int           `json:"StartIndex"`
}
```

---

## Type Conversions

### Ticks ↔ Duration

Ticks are 100-nanosecond units (Jellyfin standard):

```go
// Constants
const TicksPerSecond = 10_000_000
const TicksPerMillisecond = 10_000
const TicksPerMicrosecond = 100

// Duration → Ticks
func DurationToTicks(d time.Duration) int64 {
    return int64(d) / 100  // nanoseconds → 100-ns units
}

// Ticks → Duration
func TicksToDuration(ticks int64) time.Duration {
    return time.Duration(ticks * 100)  // 100-ns → nanoseconds
}

// Example: 23 minutes 48 seconds
// 23:48 = 1428 seconds = 14280000000 ticks
```

### UUID Format

All UUIDs are stored and transmitted with dashes:

```go
// Parse UUID
id, err := uuid.Parse("550e8400-e29b-41d4-a716-446655440000")

// Generate UUID
id := uuid.New()

// Database storage (CHAR(36))
idString := id.String()  // "550e8400-e29b-41d4-a716-446655440000"
```

---

## Summary Table

| Model | Service | Persistence | Primary Key |
|-------|---------|-------------|------|--------|
| User | Auth | MySQL | Id (UUID) |
| APIKey | Auth | MySQL | Id (UUID) |
| Device | Auth | MySQL | Id (UUID) |
| UserPolicy | Auth | MySQL | UserId (UUID) |
| base_items | Item | MySQL | Id (UUID) |
| item_values | Item | MySQL | (ItemId, ValueType, NormalizedValue) |
| user_data | Item | MySQL | (UserId, ItemId) |
| StreamState | Streaming | Memory | N/A (ephemeral) |
| TranscodingJob | Streaming | Memory | Id (UUID) |
