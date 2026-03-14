# Goose AI Prompt — JellyFinhanced → Go Micro-API Conversion Plan

> **Audience:** Goose AI agentic coding system
> **Objective:** Examine the JellyFinhanced Jellyfin media server codebase (C# / ASP.NET Core,
> located at `/home/bowens/Code/JellyFinhanced`) and produce a complete, immediately-executable
> development plan to rewrite it as a Go-based micro-API architecture — preserving the **existing
> MySQL schema verbatim** and maintaining **100% HTTP wire-format compatibility** with the
> `jellyfin-web` browser client.
> **Output destination:** `/home/bowens/Code/Kabletown/`

---

## Immovable Constraints

Read these before doing anything else. Every decision in every agent plan must satisfy all five.

### C1 — Schema Freeze
The MySQL database schema **must not change**. Every table, column name, column type, nullable
flag, primary key, foreign key, and index defined in the three migration files:
- `20260209015247_InitialCreate.cs`
- `20260308000000_AddVideoMetadataColumns.cs`
- `20260309000000_AddPerformanceIndexes.cs`

…must be used exactly as-is. No new columns, no renamed columns, no extra tables, no dropped
indexes. Go services read and write rows; they do not own schema migrations.

### C2 — Wire Format Fidelity
Every HTTP route path, method, query parameter name, JSON field name, response envelope shape,
status code, and response header that the `jellyfin-web` client currently sends or receives
must be reproduced byte-for-byte by the Go services. The web client is the integration test
oracle. If the browser app works without changes, the implementation is correct.

### C3 — Agent Isolation
Each service agent must be able to implement its service **without reading any other service's
code**. Agent task cards must be self-contained. Cross-service communication is limited to
database-shared state (MySQL) and clearly-specified internal HTTP calls.

### C4 — Go Idioms
All Go code must use:
- Go 1.23+ with modules
- `github.com/go-chi/chi/v5` for HTTP routing
- `github.com/jmoiron/sqlx` + `github.com/go-sql-driver/mysql` for database access
- `log/slog` with JSON handler for structured logging
- `context.Context` propagation through every call chain
- Graceful shutdown via `signal.NotifyContext` + `http.Server.Shutdown`

### C5 — Test Coverage
Every agent must produce:
- Unit tests (mock DB interface) covering happy path, auth failure, and 404 for every handler
- Integration tests using `testcontainers-go` against real MySQL

---

## Pre-Loaded Codebase Intelligence

The following facts were extracted from a full examination of the source tree. Goose must
**verify** these during Phase 0 but may use them immediately to avoid redundant reading.

### P1 — Project Layout

```
/home/bowens/Code/JellyFinhanced/
├── Jellyfin.Api/Controllers/          # 59 controller files (the API surface)
├── Jellyfin.Api/Auth/                 # Authorization attributes and context
├── Jellyfin.Data/                     # DTOs and request/response models
├── MediaBrowser.Model/Dto/            # Core DTO definitions (BaseItemDto, UserDto, etc.)
├── MediaBrowser.Model/Entities/       # Domain enums and value types
├── Jellyfin.Server.Implementations/   # Manager implementations
├── Emby.Server.Implementations/       # Library and user manager implementations
├── src/Jellyfin.Database/
│   ├── Jellyfin.Database.Implementations/
│   │   ├── JellyfinDbContext.cs        # EF Core DbContext — DbSet<T> inventory
│   │   └── DbRepository/              # Repository classes
│   └── Jellyfin.Database.Providers.MySql/
│       └── Migrations/                # 3 MySQL migration files — authoritative schema
└── jellyfin-web/                      # Web client (do not modify)
```

### P2 — The 59 Controllers and Their Domains

| File | Primary Route Prefix | Domain |
|---|---|---|
| ActivityLogController.cs | /ActivityLog | system |
| ApiKeyController.cs | /Auth/Keys | auth |
| ArtistsController.cs | /Artists | content |
| AudioController.cs | /Audio | stream |
| BackupController.cs | /System/Backups | system |
| BrandingController.cs | /Branding | system |
| ChannelsController.cs | /Channels | content |
| ClientLogController.cs | /Logs | system |
| CollectionController.cs | /Collections | content |
| ConfigurationController.cs | /System/Configuration | system |
| DashboardController.cs | /Dashboard | system |
| DevicesController.cs | /Devices | session |
| DisplayPreferencesController.cs | /DisplayPreferences | user |
| DynamicHlsController.cs | /Videos/{id}/master.m3u8 | stream |
| EnvironmentController.cs | /Environment | system |
| FilterController.cs | /Items/Filters | library |
| GenresController.cs | /Genres | library |
| HlsSegmentController.cs | /Videos/\*/Segments/\* | stream |
| ImageController.cs | /Items/{id}/Images | media |
| InstantMixController.cs | /Items/{id}/InstantMix | content |
| ItemLookupController.cs | /Items/RemoteSearch | metadata |
| ItemRefreshController.cs | /Items/{id}/Refresh | metadata |
| ItemsController.cs | /Items | library |
| ItemUpdateController.cs | /Items/{id} (POST) | metadata |
| LibraryController.cs | /Library | library |
| LibraryStructureController.cs | /LibraryStructure | library |
| LiveTvController.cs | /LiveTV | content |
| LocalizationController.cs | /Localization | system |
| LyricsController.cs | /Items/{id}/Lyrics | media |
| MediaInfoController.cs | /Items/{id}/MediaInfo | media |
| MediaSegmentsController.cs | /Items/{id}/MediaSegments | media |
| MoviesController.cs | /Movies | content |
| MusicGenresController.cs | /MusicGenres | library |
| PackageController.cs | /Packages | system |
| PersonsController.cs | /Persons | library |
| PlaylistsController.cs | /Playlists | content |
| PlaystateController.cs | /Users/{id}/PlayedItems | playstate |
| PluginsController.cs | /Plugins | system |
| QuickConnectController.cs | /QuickConnect | auth |
| RemoteImageController.cs | /RemoteImage | metadata |
| ScheduledTasksController.cs | /ScheduledTasks | system |
| SearchController.cs | /Search | library |
| SessionController.cs | /Sessions | session |
| StartupController.cs | /Startup | auth |
| StudiosController.cs | /Studios | library |
| SubtitleController.cs | /Items/{id}/Subtitles | media |
| SuggestionsController.cs | /Items/{id}/Similar | library |
| SyncPlayController.cs | /SyncPlay | session |
| SystemController.cs | /System | system |
| TimeSyncController.cs | /GetUtcTime | session |
| TrailersController.cs | /Trailers | library |
| TrickplayController.cs | /Videos/{id}/Trickplay | stream |
| TvShowsController.cs | /Shows | content |
| UniversalAudioController.cs | /Audio/{id}/Universal | stream |
| UserController.cs | /Users | user+auth |
| UserLibraryController.cs | /Users/{id}/Items | library |
| UserViewsController.cs | /Users/{id}/Views | library |
| VideoAttachmentsController.cs | /Videos/{id}/Attachments | media |
| VideosController.cs | /Videos | stream |
| YearsController.cs | /Years | library |

### P3 — MySQL Schema (Critical Tables)

The full schema is in `20260209015247_InitialCreate.cs`. Key tables with their column profiles:

**`BaseItems`** — the heart of the library (self-referencing via `ParentId`):
```
Id               char(36) PK               -- GUID, lowercase hyphenated
Type             varchar(128) NOT NULL      -- "Movie", "Series", "Episode", "Audio", "Folder", etc.
Data             longtext NULL              -- JSON blob with extended properties
Path             varchar(768) NULL          -- filesystem path
Name             longtext NULL              -- display name
SortName         varchar(128) NULL          -- sortable name (indexed)
CleanName        longtext NULL              -- searchable cleaned name
OriginalTitle    longtext NULL              -- original language title
Overview         longtext NULL
Tagline          longtext NULL
Genres           longtext NULL              -- PIPE-DELIMITED: "Action|Drama|Thriller"
Studios          longtext NULL              -- PIPE-DELIMITED
Tags             longtext NULL              -- PIPE-DELIMITED
Artists          longtext NULL              -- PIPE-DELIMITED (music)
AlbumArtists     longtext NULL              -- PIPE-DELIMITED (music)
ProductionLocations longtext NULL           -- PIPE-DELIMITED
ExtraIds         longtext NULL              -- PIPE-DELIMITED GUIDs
RunTimeTicks     bigint NULL                -- duration in 100-nanosecond ticks
DateCreated      datetime(6) NULL
DateModified     datetime(6) NULL
DateLastRefreshed datetime(6) NULL
DateLastSaved    datetime(6) NULL
PremiereDate     datetime(6) NULL
ProductionYear   int NULL
CommunityRating  float NULL
CriticRating     float NULL
OfficialRating   longtext NULL              -- "PG-13", "TV-MA", etc.
InheritedParentalRatingValue int NULL       -- numeric value for parental control
IsFolder         tinyint(1) NOT NULL
IsVirtualItem    tinyint(1) NOT NULL        -- TRUE = DB-only, no backing file
IsLocked         tinyint(1) NOT NULL
IsMovie          tinyint(1) NOT NULL
IsSeries         tinyint(1) NOT NULL
MediaType        varchar(64) NULL           -- "Video", "Audio", "Photo", "Book"
Width            int NULL
Height           int NULL
Size             bigint NULL
Audio            int NULL                   -- enum: Mono=0, Stereo=1, etc.
ExtraType        int NULL                   -- enum: extra category
TotalBitrate     int NULL
ParentId         char(36) NULL FK→BaseItems(Id)
TopParentId      char(36) NULL              -- root library folder GUID
SeriesId         char(36) NULL              -- episode→series GUID
SeasonId         char(36) NULL              -- episode→season GUID
SeriesName       longtext NULL
SeasonName       longtext NULL
EpisodeTitle     longtext NULL
IndexNumber      int NULL                   -- episode number
ParentIndexNumber int NULL                  -- season number
Album            longtext NULL              -- music album name
LUFS             float NULL
NormalizationGain float NULL
PresentationUniqueKey varchar(128) NULL
SeriesPresentationUniqueKey varchar(128) NULL
ChannelId        char(36) NULL
ExternalId       longtext NULL
ExternalSeriesId longtext NULL
ExternalServiceId longtext NULL
ShowId           longtext NULL
OwnerId          longtext NULL
PrimaryVersionId longtext NULL
```

**`Users`**:
```
Id                          char(36) PK
Username                    varchar(255) NOT NULL
Password                    longtext NULL           -- BCrypt hash
MustUpdatePassword          tinyint(1) NOT NULL
AuthenticationProviderId    varchar(255) NOT NULL
PasswordResetProviderId     varchar(255) NOT NULL
InvalidLoginAttemptCount    int NOT NULL
LastActivityDate            datetime(6) NULL
LastLoginDate               datetime(6) NULL
LoginAttemptsBeforeLockout  int NULL
MaxActiveSessions           int NOT NULL
SubtitleMode                int NOT NULL            -- enum
PlayDefaultAudioTrack       tinyint(1) NOT NULL
SubtitleLanguagePreference  varchar(255) NULL
AudioLanguagePreference     varchar(255) NULL
MaxParentalAgeRating        int NULL
RemoteClientBitrateLimit    int NOT NULL
InternalId                  bigint NOT NULL UNIQUE  -- auto-increment surrogate
EnableAutoLogin             tinyint(1) NOT NULL
RowVersion                  int unsigned NOT NULL   -- optimistic concurrency token
```

**`Devices`** (active sessions + access tokens):
```
Id                  int PK AUTO_INCREMENT
UserId              char(36) NOT NULL FK→Users(Id)
AccessToken         char(36) NOT NULL               -- the bearer token (GUID)
AppName             varchar(64) NOT NULL
AppVersion          varchar(32) NOT NULL
DeviceId            varchar(256) NOT NULL
DeviceName          varchar(64) NOT NULL
IsActive            tinyint(1) NOT NULL
DateCreated         datetime(6) NOT NULL
DateLastActivity    datetime(6) NOT NULL
```

**`ApiKeys`**:
```
Id                  int PK AUTO_INCREMENT
DateCreated         datetime(6) NOT NULL
DateLastActivity    datetime(6) NOT NULL
Name                varchar(64) NOT NULL
AccessToken         varchar(255) NOT NULL           -- opaque string (not a GUID)
```

**`UserData`** (per-user playback state per item):
```
ItemId                   char(36) PK (composite with UserId)
UserId                   char(36) PK
PlaybackPositionTicks    bigint NOT NULL
PlayCount                int NOT NULL
IsFavorite               tinyint(1) NOT NULL
Played                   tinyint(1) NOT NULL
Rating                   float NULL                 -- user star rating
LastPlayedDate           datetime(6) NULL
AudioStreamIndex         int NULL
SubtitleStreamIndex      int NULL
```

**`MediaStreamInfos`** (tracks per media item):
```
ItemId          char(36) FK→BaseItems(Id)
StreamIndex     int
StreamType      int NOT NULL    -- 0=Audio, 1=Video, 2=Subtitle, 3=EmbeddedImage, 4=Data, 5=Lyric
Codec           varchar(32) NULL
Language        varchar(32) NULL
ChannelLayout   varchar(32) NULL
BitRate         int NULL
BitDepth        int NULL
Width           int NULL
Height          int NULL
IsDefault       tinyint(1) NOT NULL
IsForced        tinyint(1) NOT NULL
IsExternal      tinyint(1) NOT NULL
IsHearingImpaired tinyint(1) NOT NULL
VideoRange      varchar(32) NULL
VideoRangeType  varchar(32) NULL
ColorSpace      varchar(32) NULL
ColorTransfer   varchar(32) NULL
ColorPrimaries  varchar(32) NULL
PixelFormat     varchar(32) NULL
AspectRatio     varchar(16) NULL
AverageFrameRate float NULL
RealFrameRate   float NULL
Profile         varchar(64) NULL
Level           float NULL
PacketLength    int NULL
TimestampOffset int NULL
IsAVC           tinyint(1) NULL
RefFrames       int NULL
CodecTimeBase   varchar(16) NULL
Comment         varchar(256) NULL
NalLengthSize   varchar(8) NULL
Title           varchar(256) NULL
TimeBase        varchar(16) NULL
CodecTag        varchar(16) NULL
DvVersionMajor  int NULL
DvVersionMinor  int NULL
DvProfile       int NULL
DvLevel         int NULL
RpuPresentFlag  int NULL
ElPresentFlag   int NULL
BlPresentFlag   int NULL
DvBlSignalCompatibilityId int NULL
Rotation        int NULL
```

**`ItemValues`** (deduplicated genre/artist/studio values):
```
ItemValueId     char(36) PK
Type            int NOT NULL    -- 0=Artist, 1=AlbumArtist, 2=Genre, 3=Studio, 4=Tag, 5=ProductionLocation, 6=MusicGenre
Value           varchar(255) NOT NULL
CleanValue      varchar(255) NOT NULL
```

**`ItemValuesMap`** (junction: items ↔ values):
```
ItemValueId     char(36) PK (composite with ItemId) FK→ItemValues
ItemId          char(36) PK FK→BaseItems
```

**`BaseItemImageInfos`** (images per item):
```
Id                  int PK AUTO_INCREMENT
ItemId              char(36) FK→BaseItems(Id)
Path                varchar(512) NOT NULL
ImageType           int NOT NULL    -- 0=Primary, 1=Art, 2=Backdrop, 3=Logo, 4=Thumb, 5=Disc, 6=Box, 7=Screenshot, 8=Menu, 9=Chapter, 10=BoxRear, 11=Profile, 12=Banner
DateModified        datetime(6) NULL
Width               int NULL
Height              int NULL
BlurHash            varchar(255) NULL
```

**`BaseItemProviders`** (external ID mappings):
```
ItemId          char(36) FK→BaseItems(Id)
ProviderId      varchar(255)    -- "tmdb", "imdb", "tvdb", "musicbrainztrack", etc.
ProviderValue   varchar(255)
PK: (ItemId, ProviderId)
```

**`Peoples`** and **`PeopleBaseItemMap`**:
```
-- Peoples
Id              char(36) PK
Name            varchar(255) NOT NULL
PersonType      longtext NULL       -- "Actor", "Director", "Writer", etc.

-- PeopleBaseItemMap
ItemId          char(36) PK (composite) FK→BaseItems
PeopleId        char(36) PK (composite) FK→Peoples
Role            varchar(1024) NULL  -- character name for actors
SortOrder       int NULL
ListOrder       int NULL
```

**`AncestorIds`** (flattened hierarchy for collection membership):
```
ItemId          char(36) PK (composite) FK→BaseItems
ParentItemId    char(36) PK FK→BaseItems
```

**`Chapters`** (chapter markers):
```
ItemId          char(36) FK→BaseItems(Id)
ChapterIndex    int PK (composite with ItemId)
StartPositionTicks  bigint NOT NULL
Name            longtext NULL
ImagePath       longtext NULL
ImageDateModified datetime(6) NOT NULL
```

**`MediaSegments`** (intro/outro detection):
```
Id              char(36) PK
ItemId          char(36) NOT NULL
Type            int NOT NULL    -- 0=Unknown, 1=Commercial, 2=Preview, 3=Recap, 4=Outro, 5=Intro, 6=Credits
StartTicks      bigint NOT NULL
EndTicks        bigint NOT NULL
SegmentProviderId longtext NOT NULL
```

**`TrickplayInfos`** (seek preview thumbnails):
```
ItemId          char(36) PK (composite with Width)
Width           int PK
Height          int NOT NULL
TileWidth       int NOT NULL
TileHeight      int NOT NULL
ThumbnailCount  int NOT NULL
Interval        int NOT NULL    -- milliseconds between thumbnails
Bandwidth       int NOT NULL
```

**`DisplayPreferences`**:
```
Id                      int PK AUTO_INCREMENT
UserId                  char(36) FK→Users(Id)
ItemId                  char(36) NOT NULL
Client                  varchar(32) NOT NULL
ViewType                varchar(32) NULL
SortBy                  varchar(64) NULL
SortOrder               int NOT NULL
RememberIndexing        tinyint(1) NOT NULL
ShowBackdrop            tinyint(1) NOT NULL
ShowSidebar             tinyint(1) NOT NULL
ScrollDirection         int NOT NULL
SkipForwardLength       int NOT NULL
SkipBackwardLength      int NOT NULL
RememberSorting         tinyint(1) NOT NULL
ChromecastVersion       int NOT NULL
EnableNextVideoInfoOverlay tinyint(1) NOT NULL
```

**`CustomItemDisplayPreferences`**:
```
Id          int PK AUTO_INCREMENT
UserId      char(36) NOT NULL
ItemId      char(36) NOT NULL
Client      varchar(32) NOT NULL
Key         varchar(255) NOT NULL
Value       longtext NULL
```

**`Permissions`** and **`Preferences`** (per-user key-value stores):
```
-- Permissions
Id          int PK AUTO_INCREMENT
UserId      char(36) FK→Users(Id)
Kind        int NOT NULL    -- enum: IsAdministrator=0, IsHidden=1, IsDisabled=2, EnableContentDeletion=5, EnableContentDownloading=6, EnableSyncTranscoding=7, EnableMediaConversion=8, EnableAllDevices=12, EnableAllChannels=13, EnableAllFolders=14, EnablePublicSharing=18, EnableRemoteControlOfOtherUsers=20, etc.
Value       tinyint(1) NOT NULL

-- Preferences
Id          int PK AUTO_INCREMENT
UserId      char(36) FK→Users(Id)
Kind        int NOT NULL    -- enum of user preference keys
Value       longtext NULL

-- AccessSchedules
Id          int PK AUTO_INCREMENT
UserId      char(36) FK→Users(Id)
DayOfWeek   int NOT NULL    -- enum: Sunday=0...Saturday=6
StartHour   float NOT NULL
EndHour     float NOT NULL
```

**`ActivityLogs`**:
```
Id              int PK AUTO_INCREMENT
Name            varchar(512) NOT NULL
Overview        varchar(512) NULL
ShortOverview   varchar(512) NULL
Type            varchar(256) NOT NULL
UserId          char(36) NOT NULL
ItemId          varchar(256) NULL
DateCreated     datetime(6) NOT NULL
LogSeverity     int NOT NULL    -- 0=Trace,1=Debug,2=Information,3=Warning,4=Error,5=Critical
RowVersion      int unsigned NOT NULL
```

**`DeviceOptions`**:
```
Id          int PK AUTO_INCREMENT
DeviceId    varchar(255) NOT NULL
CustomName  longtext NULL
```

**`KeyframeData`**:
```
ItemId      char(36) PK
KeyframeTicks longtext NOT NULL   -- JSON array of int64 tick values
```

**`AttachmentStreamInfos`**:
```
ItemId          char(36) FK→BaseItems(Id)
Index           int PK (composite with ItemId)
Codec           varchar(32) NULL
CodecTag        varchar(32) NULL
Comment         varchar(256) NULL
Filename        varchar(512) NULL
MimeType        varchar(64) NULL
```

**`BaseItemMetadataFields`**:
```
Id          int PK AUTO_INCREMENT
ItemId      char(36) FK→BaseItems(Id)
MetadataField int NOT NULL    -- enum: Cast=0, Genres=1, ProductionLocations=2, Studios=3, Tags=4, Name=5, Overview=6, Runtime=7, OfficialRating=8
```

**`BaseItemTrailerTypes`**:
```
Id          int PK AUTO_INCREMENT
ItemId      char(36) FK→BaseItems(Id)
TrailerType int NOT NULL    -- enum: ComingSoon=1, Archive=2, LocalTrailer=4, SeasonsTrailer=8
```

### P4 — Authentication Wire Format

**Header (sent by client on every authenticated request):**
```
X-Emby-Authorization: MediaBrowser Client="Jellyfin Web", Device="Chrome", DeviceId="abc123def456", Version="10.9.11", Token="3f2504e0-4f89-11d3-9a0c-0305e82c3301"
```

Alternative header names accepted (all identical parsing):
- `X-Emby-Authorization`
- `Authorization` (with `MediaBrowser ` prefix, not `Bearer `)

Alternative token delivery (query param, for stream URLs where headers aren't injectable):
- `?api_key=<token>` — both device tokens (GUID format) and API keys (string)
- `?X-Emby-Token=<token>` — some older clients

**Token resolution algorithm:**
```
1. Extract Token value from header or api_key query param
2. SELECT d.*, u.* FROM Devices d JOIN Users u ON d.UserId = u.Id
   WHERE d.AccessToken = ? AND d.IsActive = 1
   → if found: auth principal = that User, IsAdmin based on Permissions
3. If step 2 returns nothing:
   SELECT * FROM ApiKeys WHERE AccessToken = ?
   → if found: auth principal = server admin (no user; IsAdmin = true)
4. If neither: return 401 with body {"Message":"Access token is invalid or expired.","StatusCode":401}
5. On successful Devices match: UPDATE Devices SET DateLastActivity = NOW() WHERE Id = ?
```

**Response headers that must be present on ALL responses:**
```
X-Application-Version: 10.9.11          (from config or hardcoded version constant)
X-MediaBrowser-Server-Id: <server-uuid> (from system.xml ServerId field)
Content-Type: application/json; charset=utf-8
```

### P5 — JSON Wire Format Rules

These rules are absolute. Any violation breaks the web client.

**Rule W1 — Field names are PascalCase throughout:**
```json
{"Id": "...", "Name": "...", "RunTimeTicks": 72000000000, "UserData": {...}}
```
NOT camelCase. Go's `encoding/json` defaults to exact struct field names, so struct fields
must be PascalCase OR have explicit `json:"PascalCaseName"` tags.

**Rule W2 — GUIDs are lowercase hyphenated:**
```json
{"Id": "3f2504e0-4f89-11d3-9a0c-0305e82c3301"}
```
`uuid.UUID.String()` from `github.com/google/uuid` produces lowercase — use this library.
Never uppercase. Never strip hyphens.

**Rule W3 — Timestamps have exactly 7 decimal places in UTC:**
```json
{"DateCreated": "2024-01-15T22:30:00.0000000Z"}
```
Go's `time.RFC3339Nano` produces up to 9 decimal places. **Must use a custom marshaler:**
```go
type JellyfinTime struct{ time.Time }

func (t JellyfinTime) MarshalJSON() ([]byte, error) {
    if t.IsZero() {
        return []byte("null"), nil
    }
    formatted := t.UTC().Format("2006-01-02T15:04:05.0000000Z")
    return []byte(`"` + formatted + `"`), nil
}

func (t *JellyfinTime) UnmarshalJSON(data []byte) error {
    s := strings.Trim(string(data), `"`)
    parsed, err := time.Parse("2006-01-02T15:04:05.0000000Z", s)
    if err != nil {
        parsed, err = time.Parse(time.RFC3339Nano, s)
    }
    if err != nil { return err }
    t.Time = parsed
    return nil
}
```

**Rule W4 — Duration is in 100-nanosecond ticks (not milliseconds, not seconds):**
```json
{"RunTimeTicks": 72000000000}   // = 2 hours (2h * 3600s * 10_000_000 ticks/s)
```
Conversion: `ticks = durationNanoseconds / 100`
Conversion back: `duration = time.Duration(ticks * 100) * time.Nanosecond`

**Rule W5 — Paginated list responses always use this exact envelope:**
```json
{
  "Items": [],
  "TotalRecordCount": 0,
  "StartIndex": 0
}
```
Never omit `TotalRecordCount` or `StartIndex`. Never return a bare array. Even single-item
endpoints that logically return one thing use the envelope if their C# counterpart does.

**Rule W6 — Empty arrays are `[]`, never `null`:**
```go
// Always initialize slices before marshaling:
if result.Items == nil {
    result.Items = []ItemDTO{}
}
```

**Rule W7 — Pipe-delimited database strings must be split into arrays in responses:**
```go
// BaseItems.Genres is stored as "Action|Drama|Thriller"
// BaseItemDto.Genres must be []string{"Action", "Drama", "Thriller"}
func splitPipe(s *string) []string {
    if s == nil || *s == "" {
        return []string{}
    }
    return strings.Split(*s, "|")
}
// And when writing back:
func joinPipe(ss []string) string {
    return strings.Join(ss, "|")
}
```

**Rule W8 — HTTP status codes must match C# exactly:**
- `200 OK` for GET with body
- `204 No Content` for DELETE (no body)
- `400 Bad Request` for validation failures
- `401 Unauthorized` for missing/invalid token (with `WWW-Authenticate` header)
- `403 Forbidden` for insufficient permissions
- `404 Not Found` for missing resources (body: `{"Message":"Resource not found.","StatusCode":404}`)
- `409 Conflict` for duplicate names
- `500 Internal Server Error` (body: `{"Message":"...","StatusCode":500}`)

**Rule W9 — Image URLs must maintain exact format:**
```
/Items/{itemId}/Images/{imageType}?fillHeight=300&fillWidth=300&quality=90&tag={md5hash}
/Items/{itemId}/Images/{imageType}/{imageIndex}
```
The `tag` parameter is the MD5 hash of `DateModified.ToString("yyyyMMddHHmmss")`.

**Rule W10 — The `BaseItemDto` type discriminator field:**
```json
{"Type": "Movie", "MediaType": "Video"}
{"Type": "Series", "MediaType": "Video"}
{"Type": "Episode", "MediaType": "Video"}
{"Type": "Audio", "MediaType": "Audio"}
{"Type": "MusicAlbum", "MediaType": null}
{"Type": "Folder", "MediaType": null}
```
`Type` is the C# class name short-form. `MediaType` is the media category.

### P6 — ItemValues Type Enum (for Genres/Studios/Artists filtering)

```go
const (
    ItemValueTypeArtist           = 0
    ItemValueTypeAlbumArtist      = 1
    ItemValueTypeGenre            = 2
    ItemValueTypeStudio           = 3
    ItemValueTypeTag              = 4
    ItemValueTypeProductionLocation = 5
    ItemValueTypeMusicGenre       = 6
)
```

### P7 — Performance Index Strategy

Every SQL query in the Go services must be designed to use the indexes created by
`20260309000000_AddPerformanceIndexes.cs`. The most critical patterns:

```sql
-- Library browse (uses IX_BaseItems_Type_IsVirtualItem_SortName):
WHERE Type IN (?) AND IsVirtualItem = 0 ORDER BY SortName

-- Parent folder contents (uses IX_BaseItems_ParentId_IsVirtualItem_Type):
WHERE ParentId = ? AND IsVirtualItem = 0

-- Recently added (uses IX_BaseItems_Type_IsVirtualItem_DateCreated):
WHERE Type IN (?) AND IsVirtualItem = 0 ORDER BY DateCreated DESC

-- Series episodes (uses IX_BaseItems_SeriesId_IsVirtualItem):
WHERE SeriesId = ? AND IsVirtualItem = 0

-- User favorites (uses IX_UserData_UserId_IsFavorite):
WHERE UserId = ? AND IsFavorite = 1

-- User resume queue (uses IX_UserData_UserId_PlaybackPositionTicks):
WHERE UserId = ? AND PlaybackPositionTicks > 0

-- Full-text search (uses FT_BaseItems_Name_OriginalTitle):
WHERE MATCH(Name, OriginalTitle) AGAINST (? IN BOOLEAN MODE)
```

---

## Phase 0 — Required Examination Tasks

Before writing any plans, Goose must examine the following files and record the specified
information. Do not skip these reads even though much is pre-loaded above — the goal is to
catch anything that changed.

### E1 — Read Every Controller File
For each `.cs` file in `/home/bowens/Code/JellyFinhanced/Jellyfin.Api/Controllers/`:
- Extract all `[Http*]` route attributes with their exact route templates
- Note `[Authorize]`, `[AllowAnonymous]`, and `[Authorize(Policy = "...")]` decorations
- List all method parameters marked `[FromQuery]`, `[FromRoute]`, `[FromBody]`
- Note the return type (especially `ActionResult<QueryResult<T>>` vs `ActionResult<T>` vs `IActionResult`)

### E2 — Read the DbContext
File: `src/Jellyfin.Database/Jellyfin.Database.Implementations/JellyfinDbContext.cs`
- List every `DbSet<T>` property and its entity class name → this is the definitive table inventory

### E3 — Read the Auth Infrastructure
Files: `Jellyfin.Api/Auth/` (all files), `Jellyfin.Server/Middleware/` (if exists)
- Document exactly how `X-Emby-Authorization` is parsed (regex or string split?)
- Document which middleware runs before which authorization attribute

### E4 — Read the Key DTOs
Files: `MediaBrowser.Model/Dto/BaseItemDto.cs`, `MediaBrowser.Model/Dto/UserDto.cs`,
`Jellyfin.Data/Dto/` (all files)
- List every field with its C# type and JSON property name
- Note nullable vs. non-nullable

### E5 — Read the BaseItemRepository Query Builder
File: `src/Jellyfin.Database/Jellyfin.Database.Implementations/DbRepository/BaseItemRepository.cs`
(or closest equivalent)
- Document the `InternalItemsQuery`/`InternalQueryResult` filter parameters
- This is the most complex query — the Go library-service must replicate it

### E6 — Read HLS and Stream Routes
Files: `Jellyfin.Api/Controllers/DynamicHlsController.cs`, `Jellyfin.Api/Controllers/VideosController.cs`
- List every route template and every query parameter
- Note ffmpeg invocation approach (look for `_mediaEncoder` or `ProcessOptions`)

---

## Phase 1 — Architecture Blueprint

After completing Phase 0, produce the following architecture artifacts.

### Service Map

Define **13 Go services** with these exact names, ports, and responsibilities:

| # | Service | Port | Controllers Covered |
|---|---|---|---|
| 0 | `shared` | (library, no HTTP) | — |
| 1 | `gateway` | 8080 | Nginx/Caddy config |
| 2 | `auth-service` | 8001 | ApiKeyController, QuickConnectController, StartupController, UserController (auth endpoints only) |
| 3 | `user-service` | 8002 | UserController (CRUD), DisplayPreferencesController |
| 4 | `library-service` | 8003 | ItemsController, UserLibraryController, UserViewsController, LibraryController, LibraryStructureController, FilterController, GenresController, MusicGenresController, StudiosController, PersonsController, YearsController, TrailersController, SearchController, SuggestionsController |
| 5 | `playstate-service` | 8004 | PlaystateController |
| 6 | `media-service` | 8005 | ImageController, MediaInfoController, MediaSegmentsController, LyricsController, VideoAttachmentsController, SubtitleController |
| 7 | `stream-service` | 8006 | VideosController, AudioController, UniversalAudioController, DynamicHlsController, HlsSegmentController, TrickplayController |
| 8 | `session-service` | 8007 | SessionController, DevicesController, SyncPlayController, TimeSyncController |
| 9 | `metadata-service` | 8008 | ItemRefreshController, ItemUpdateController, ItemLookupController, RemoteImageController, ScheduledTasksController, PackageController, PluginsController, BackupController |
| 10 | `content-service` | 8009 | MoviesController, TvShowsController, ArtistsController, ChannelsController, LiveTvController, CollectionController, PlaylistsController, InstantMixController |
| 11 | `system-service` | 8010 | SystemController, ConfigurationController, DashboardController, BrandingController, EnvironmentController, LocalizationController, ClientLogController, ActivityLogController |

### Shared Package Specification (`shared/`)

The `shared` module (`github.com/jellyfinhanced/shared`) must contain these packages:

```
shared/
├── go.mod                          (module github.com/jellyfinhanced/shared)
├── auth/
│   ├── middleware.go               (chi middleware: parse header, query DB, set context)
│   ├── context.go                  (typed context keys + AuthInfo struct)
│   └── parser.go                   (X-Emby-Authorization header parser)
├── db/
│   ├── factory.go                  (NewDB(dsn string) (*sqlx.DB, error) with pool config)
│   └── tx.go                       (WithTx helper for transactions)
├── dto/
│   ├── base_item.go                (BaseItemDto, UserItemDataDto, ImageInfo, etc.)
│   ├── user.go                     (UserDto, AuthenticationResultDto)
│   ├── session.go                  (SessionInfoDto, DeviceInfoDto)
│   ├── time.go                     (JellyfinTime custom marshaler)
│   └── pagination.go               (PagedResult[T any] generic type)
├── config/
│   ├── system.go                   (read system.xml → SystemConfig struct)
│   └── network.go                  (read network.xml → NetworkConfig struct)
├── response/
│   ├── json.go                     (WriteJSON, WriteError, WriteProblem helpers)
│   └── headers.go                  (SetServerHeaders: X-Application-Version, X-MediaBrowser-Server-Id)
└── types/
    ├── guid.go                     (GUID parsing/formatting helpers)
    └── ticks.go                    (TicksToDuration, DurationToTicks)
```

**Critical implementations to specify in the plan:**

```go
// auth/context.go
type contextKey int
const (
    authInfoContextKey contextKey = iota
)

type AuthInfo struct {
    UserID      uuid.UUID
    DeviceID    string
    Token       string
    IsAdmin     bool
    IsApiKey    bool
    // Resolved from DB:
    Username    string
    MaxParentalRating *int
    Permissions map[int]bool   // kind → value from Permissions table
}

func GetAuth(ctx context.Context) (*AuthInfo, bool) {
    v, ok := ctx.Value(authInfoContextKey).(*AuthInfo)
    return v, ok
}

// auth/parser.go — parse "MediaBrowser Client="...", Token="..."" format
func ParseEmbyAuthHeader(header string) (token, deviceID string, err error) {
    // Header format: MediaBrowser key1="val1", key2="val2", ...
    // Use strings.Split on ", " then parse each key="value" pair
    // Return token and deviceId extracted values
}

// auth/middleware.go
func Middleware(db *sqlx.DB) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // 1. Try X-Emby-Authorization header
            // 2. Try Authorization header (same format)
            // 3. Try api_key query param
            // 4. If no token found → pass through (AllowAnonymous endpoints)
            // 5. Query Devices JOIN Users WHERE AccessToken = ?
            // 6. If not found → query ApiKeys WHERE AccessToken = ?
            // 7. UPDATE Devices DateLastActivity on success
            // 8. Set AuthInfo in context
        })
    }
}

// dto/pagination.go
type PagedResult[T any] struct {
    Items            []T `json:"Items"`
    TotalRecordCount int `json:"TotalRecordCount"`
    StartIndex       int `json:"StartIndex"`
}

func NewPagedResult[T any](items []T, total, start int) PagedResult[T] {
    if items == nil { items = []T{} }
    return PagedResult[T]{Items: items, TotalRecordCount: total, StartIndex: start}
}
```

### Gateway Configuration

Produce an Nginx configuration (`gateway/nginx.conf`) that:
- Listens on port 8080 (HTTP) and 8443 (HTTPS with self-signed cert for dev)
- Routes by path prefix to correct service port
- Passes `X-Emby-Authorization`, `Authorization`, `X-Real-IP`, `X-Forwarded-For` unchanged
- Serves `jellyfin-web/dist/` at `/web` with proper cache headers
- Has a health check endpoint `GET /ping` → `200 OK`
- Buffers streaming responses properly (for HLS, increase `proxy_buffering off` on stream routes)

```nginx
upstream auth_svc      { server 127.0.0.1:8001; }
upstream user_svc      { server 127.0.0.1:8002; }
upstream library_svc   { server 127.0.0.1:8003; }
upstream playstate_svc { server 127.0.0.1:8004; }
upstream media_svc     { server 127.0.0.1:8005; }
upstream stream_svc    { server 127.0.0.1:8006; }
upstream session_svc   { server 127.0.0.1:8007; }
upstream metadata_svc  { server 127.0.0.1:8008; }
upstream content_svc   { server 127.0.0.1:8009; }
upstream system_svc    { server 127.0.0.1:8010; }

# Auth routes
location ~ ^/(Auth|QuickConnect|Startup)/ { proxy_pass http://auth_svc; }
location ~ ^/Users/AuthenticateByName { proxy_pass http://auth_svc; }

# User routes
location ~ ^/Users/[^/]+/(Password|EasyPassword|Policy|Configuration)$ { proxy_pass http://user_svc; }
location /DisplayPreferences/ { proxy_pass http://user_svc; }

# Playstate routes
location ~ ^/Users/[^/]+/PlayedItems { proxy_pass http://playstate_svc; }
location ~ ^/Users/[^/]+/PlayingItems { proxy_pass http://playstate_svc; }

# Stream routes (buffering off for HLS)
location ~ ^/Videos/[^/]+/(master\.m3u8|hls|stream) { proxy_pass http://stream_svc; proxy_buffering off; }
location ~ ^/Audio/[^/]+/(stream|Universal) { proxy_pass http://stream_svc; proxy_buffering off; }
location ~ ^/Videos/[^/]+/Trickplay { proxy_pass http://stream_svc; }

# Library routes
location /Items/ { proxy_pass http://library_svc; }
location /Users/$1/Items { proxy_pass http://library_svc; }

# ... (complete all upstream mappings)
```

---

## Phase 2 — Agent Task Cards

For each of the 11 HTTP services, produce a self-contained task card using this exact template.
Each card must be usable by a fresh AI coding agent with zero prior context.

---

### AGENT CARD 0: shared

**Priority:** BLOCKING — all other agents depend on this module. Must be completed first.

**Goal:** Implement the shared Go module that provides authentication middleware, database
connection factory, DTO types, JSON helpers, and pagination utilities used by all 10 HTTP
services.

**Module path:** `github.com/jellyfinhanced/shared`
**Local path:** `shared/`

**Deliverables:**
- [ ] `go.mod` with correct module name and all dependencies pinned
- [ ] `auth/middleware.go` — Chi middleware implementing the full auth flow from P4
- [ ] `auth/context.go` — AuthInfo struct and typed context key functions
- [ ] `auth/parser.go` — X-Emby-Authorization header parser with unit tests
- [ ] `db/factory.go` — `NewDB(dsn string) (*sqlx.DB, error)` with pool tuning (MaxOpenConns=25, MaxIdleConns=10, ConnMaxLifetime=5m)
- [ ] `dto/base_item.go` — Complete BaseItemDto struct (100+ fields) with all json tags PascalCase
- [ ] `dto/user.go` — UserDto, AuthenticationResultDto, SessionInfoDto
- [ ] `dto/time.go` — JellyfinTime with custom 7-decimal-place marshaler (see W3)
- [ ] `dto/pagination.go` — Generic PagedResult[T any]
- [ ] `response/json.go` — WriteJSON, WriteError, WriteProblem helpers
- [ ] `response/headers.go` — SetServerHeaders middleware adding required response headers
- [ ] `types/guid.go` — Lowercase hyphenated GUID formatting
- [ ] `types/ticks.go` — Tick ↔ time.Duration conversions
- [ ] Unit tests for every package (≥ 90% coverage)

**Critical Implementation Details:**

The `BaseItemDto` struct must have every field present in `MediaBrowser.Model/Dto/BaseItemDto.cs`.
Key fields that agents frequently get wrong:
```go
type BaseItemDto struct {
    Id                       string              `json:"Id"`                       // GUID string
    Name                     *string             `json:"Name"`
    OriginalTitle            *string             `json:"OriginalTitle"`
    ServerId                 string              `json:"ServerId"`                 // X-MediaBrowser-Server-Id value
    Type                     string              `json:"Type"`                     // "Movie", "Series", etc.
    MediaType                *string             `json:"MediaType"`
    IsFolder                 bool                `json:"IsFolder"`
    ParentId                 *string             `json:"ParentId"`
    RunTimeTicks             *int64              `json:"RunTimeTicks"`             // 100-ns ticks
    ProductionYear           *int                `json:"ProductionYear"`
    PremiereDate             *JellyfinTime       `json:"PremiereDate"`
    DateCreated              *JellyfinTime       `json:"DateCreated"`
    Genres                   []string            `json:"Genres"`                   // split from pipe-delimited
    Studios                  []NameGuidPair      `json:"Studios"`
    People                   []BaseItemPerson    `json:"People"`
    Overview                 *string             `json:"Overview"`
    Tagline                  *string             `json:"Tagline"`
    CommunityRating          *float32            `json:"CommunityRating"`
    CriticRating             *float32            `json:"CriticRating"`
    OfficialRating           *string             `json:"OfficialRating"`
    MediaStreams             []MediaStream       `json:"MediaStreams"`
    ImageTags                map[string]string   `json:"ImageTags"`                // imageType → etag
    BackdropImageTags        []string            `json:"BackdropImageTags"`
    UserData                 *UserItemDataDto    `json:"UserData"`
    SeriesId                 *string             `json:"SeriesId"`
    SeriesName               *string             `json:"SeriesName"`
    SeasonId                 *string             `json:"SeasonId"`
    SeasonName               *string             `json:"SeasonName"`
    IndexNumber              *int                `json:"IndexNumber"`
    ParentIndexNumber        *int                `json:"ParentIndexNumber"`
    // ... (examine BaseItemDto.cs for complete field list)
}

type UserItemDataDto struct {
    Rating                  *float64        `json:"Rating"`
    PlayedPercentage        *float64        `json:"PlayedPercentage"`
    UnplayedItemCount       *int            `json:"UnplayedItemCount"`
    PlaybackPositionTicks   int64           `json:"PlaybackPositionTicks"`
    PlayCount               int             `json:"PlayCount"`
    IsFavorite              bool            `json:"IsFavorite"`
    Likes                   *bool           `json:"Likes"`
    LastPlayedDate          *JellyfinTime   `json:"LastPlayedDate"`
    Played                  bool            `json:"Played"`
    Key                     string          `json:"Key"`
    ItemId                  string          `json:"ItemId"`
}
```

---

### AGENT CARD 1: auth-service

**Priority:** HIGH — deploy before user/library/session services.

**Goal:** Handle all credential-based authentication, API key management, Quick Connect pairing,
and initial server setup wizard. This service issues no JWTs — it validates credentials and
creates/returns device access tokens stored in `Devices.AccessToken`.

**Module path:** `github.com/jellyfinhanced/auth-service`
**Local path:** `auth-service/`
**Port:** 8001

**Routes to implement** (verify exact templates from E1):

```
POST   /Users/AuthenticateByName           [AllowAnonymous]
POST   /Users/{userId}/Authenticate        [AllowAnonymous]

GET    /Auth/Keys                          [IsAdmin]
POST   /Auth/Keys                          [IsAdmin]   body: {Name: string}
DELETE /Auth/Keys/{id}                     [IsAdmin]

GET    /QuickConnect/Enabled               [AllowAnonymous]
POST   /QuickConnect/Initiate              [AllowAnonymous]
GET    /QuickConnect/Status                [AllowAnonymous]  ?secret=string
POST   /QuickConnect/Authorize             [Authorized]      body: {Code: string}
POST   /QuickConnect/Connect               [AllowAnonymous]  body: {Secret: string}

GET    /Startup/Configuration              [AllowAnonymous]
POST   /Startup/Configuration              [AllowAnonymous]
GET    /Startup/User                       [AllowAnonymous]
POST   /Startup/User                       [AllowAnonymous]
POST   /Startup/Complete                   [AllowAnonymous]
GET    /Startup/RemoteAccess               [AllowAnonymous]
POST   /Startup/RemoteAccess               [AllowAnonymous]
```

**Database Tables:**
- `Users` (READ: validate credentials; UPDATE: LastLoginDate, InvalidLoginAttemptCount)
- `Devices` (INSERT: new device on auth; READ: token lookup)
- `ApiKeys` (CRUD)
- `Permissions` (READ: IsAdministrator check)

**Key Queries:**
```sql
-- Authenticate by name
SELECT Id, Username, Password, InvalidLoginAttemptCount, LoginAttemptsBeforeLockout,
       MustUpdatePassword, AuthenticationProviderId
FROM Users
WHERE Username = ? COLLATE utf8mb4_general_ci;

-- Create device session on successful auth
INSERT INTO Devices (UserId, AccessToken, AppName, AppVersion, DeviceId, DeviceName,
                     IsActive, DateCreated, DateLastActivity)
VALUES (?, UUID(), ?, ?, ?, ?, 1, NOW(), NOW());

-- API key listing
SELECT Id, DateCreated, DateLastActivity, Name, AccessToken FROM ApiKeys ORDER BY Name;

-- Delete API key
DELETE FROM ApiKeys WHERE Id = ?;
```

**Authentication Business Logic:**
```go
// Password verification: BCrypt
import "golang.org/x/crypto/bcrypt"

func verifyPassword(storedHash, plaintext string) bool {
    // Jellyfin stores passwords as BCrypt hashes prefixed with "$2a$"
    err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(plaintext))
    return err == nil
}

// Lock account on too many failures
// InvalidLoginAttemptCount >= LoginAttemptsBeforeLockout → return 403, do not attempt auth
```

**Response Shapes:**
```go
// POST /Users/AuthenticateByName response
type AuthenticationResult struct {
    User          UserDto          `json:"User"`
    SessionInfo   SessionInfoDto   `json:"SessionInfo"`
    AccessToken   string           `json:"AccessToken"`
    ServerId      string           `json:"ServerId"`
}
```

**Wire Compatibility Checklist:**
- [ ] Returns 200 (not 201) on successful auth
- [ ] `AccessToken` field is the GUID value from `Devices.AccessToken`
- [ ] `SessionInfo.DeviceId` matches the `DeviceId` sent in the auth request header
- [ ] Lockout returns `401` with `{"Message": "Your account has been locked...", "StatusCode": 401}`

---

### AGENT CARD 2: user-service

**Priority:** HIGH.

**Goal:** Manage user accounts (CRUD), passwords, user policies, and display preferences.
The `GET /Users` and `GET /Users/{id}` routes are called on every page load; they must be fast.

**Port:** 8002

**Routes:**
```
GET    /Users                                        [Authorized or IsAdmin]
GET    /Users/Me                                     [Authorized]
GET    /Users/{userId}                               [Authorized]
POST   /Users/New                                    [IsAdmin]
DELETE /Users/{userId}                               [IsAdmin]
POST   /Users/{userId}                               [SelfOrAdmin]  (update user)
POST   /Users/{userId}/Password                      [SelfOrAdmin]
POST   /Users/{userId}/EasyPassword                  [SelfOrAdmin]
POST   /Users/{userId}/Policy                        [IsAdmin]
POST   /Users/{userId}/Configuration                 [SelfOrAdmin]
GET    /DisplayPreferences/{displayPreferencesId}    [Authorized] ?userId&client
POST   /DisplayPreferences/{displayPreferencesId}    [Authorized] ?userId&client
```

**Database Tables:**
- `Users` (CRUD)
- `Permissions` (R/W per-user permission flags)
- `Preferences` (R/W per-user preference values)
- `AccessSchedules` (R/W per-user time-based access restrictions)
- `DisplayPreferences` (R/W per-user/client display settings)
- `CustomItemDisplayPreferences` (R/W per-user/item/client custom prefs)

**Key Queries:**
```sql
-- Get user with all relations
SELECT u.*,
       p.Kind as PermKind, p.Value as PermValue,
       pr.Kind as PrefKind, pr.Value as PrefValue,
       a.Id as SchId, a.DayOfWeek, a.StartHour, a.EndHour
FROM Users u
LEFT JOIN Permissions p ON p.UserId = u.Id
LEFT JOIN Preferences pr ON pr.UserId = u.Id
LEFT JOIN AccessSchedules a ON a.UserId = u.Id
WHERE u.Id = ?;

-- Update password
UPDATE Users SET Password = ? WHERE Id = ?;

-- Upsert display preferences
INSERT INTO DisplayPreferences (UserId, ItemId, Client, SortBy, SortOrder, ViewType, ...)
VALUES (?, ?, ?, ?, ?, ?, ...)
ON DUPLICATE KEY UPDATE SortBy = VALUES(SortBy), SortOrder = VALUES(SortOrder), ...;
```

**UserDto Mapping:**
```go
type UserDto struct {
    Name                    string           `json:"Name"`
    ServerId                string           `json:"ServerId"`
    Id                      string           `json:"Id"`
    HasPassword             bool             `json:"HasPassword"`
    HasConfiguredPassword   bool             `json:"HasConfiguredPassword"`
    EnableAutoLogin         bool             `json:"EnableAutoLogin"`
    LastLoginDate           *JellyfinTime    `json:"LastLoginDate"`
    LastActivityDate        *JellyfinTime    `json:"LastActivityDate"`
    Configuration           UserConfiguration    `json:"Configuration"`
    Policy                  UserPolicy           `json:"Policy"`
    // NOTE: Password hash is NEVER included in response
}
```

---

### AGENT CARD 3: library-service

**Priority:** CRITICAL — serves the most traffic. This is the most complex service.

**Goal:** Implement the core library browsing API. This includes the `GET /Items` mega-endpoint
which supports 50+ filter parameters, plus all the taxonomy endpoints (genres, studios, persons,
years) and the search endpoint.

**Port:** 8003

**Routes (exhaustive):**
```
-- Core items
GET    /Items                                            [Authorized]
GET    /Items/{itemId}                                   [Authorized]
GET    /Items/{itemId}/Ancestors                         [Authorized]
GET    /Items/{itemId}/Similar                           [Authorized]
GET    /Items/Counts                                     [Authorized]
GET    /Items/Filters                                    [Authorized]
GET    /Items/Filters2                                   [Authorized]

-- User-scoped item browsing
GET    /Users/{userId}/Items                             [SelfOrAdmin]
GET    /Users/{userId}/Items/{itemId}                    [SelfOrAdmin]
GET    /Users/{userId}/Items/Latest                      [SelfOrAdmin]
GET    /Users/{userId}/Items/Resume                      [SelfOrAdmin]

-- User views (home screen library folders)
GET    /Users/{userId}/Views                             [SelfOrAdmin]

-- Library management
GET    /Items/{itemId}/File                              [Authorized]
GET    /Items/{itemId}/ThemeSongs                        [Authorized]
GET    /Items/{itemId}/ThemeVideos                       [Authorized]
GET    /Items/{itemId}/LocalTrailers                     [Authorized]
GET    /Items/{itemId}/SpecialFeatures                   [Authorized]
GET    /Items/{itemId}/ExternalIds                       [Authorized]
DELETE /Items/{itemId}                                   [IsAdmin]
POST   /Library/Media/Updated                            [Authorized]
POST   /Library/Series/Added                             [Authorized]
POST   /Library/Series/Updated                           [Authorized]
GET    /Library/PhysicalPaths                            [IsAdmin]
GET    /Library/SelectableMediaFolders                   [IsAdmin]
GET    /Library/AvailableOptions                         [IsAdmin]
GET    /LibraryStructure/VirtualFolders                  [IsAdmin]
POST   /LibraryStructure/VirtualFolders                  [IsAdmin]
DELETE /LibraryStructure/VirtualFolders                  [IsAdmin]
POST   /LibraryStructure/VirtualFolders/Paths            [IsAdmin]
DELETE /LibraryStructure/VirtualFolders/Paths            [IsAdmin]
POST   /LibraryStructure/VirtualFolders/LibraryOptions   [IsAdmin]

-- Taxonomy
GET    /Genres                                           [Authorized]
GET    /Genres/{genreName}                               [Authorized]
GET    /MusicGenres                                      [Authorized]
GET    /MusicGenres/{genreName}                          [Authorized]
GET    /Studios                                          [Authorized]
GET    /Studios/{studioName}                             [Authorized]
GET    /Persons                                          [Authorized]
GET    /Persons/{personName}                             [Authorized]
GET    /Years                                            [Authorized]
GET    /Years/{year}                                     [Authorized]
GET    /Trailers                                         [Authorized]

-- Search
GET    /Search/Hints                                     [Authorized]
GET    /Items/{itemId}/Similar                           [Authorized]
```

**The `GET /Items` Query Builder (most complex part of the entire codebase):**

The Go implementation must replicate the `BaseItemRepository.TranslateQuery` method. The
following query parameters must be supported with their exact semantics:

```go
// REQUIRED: implement all of these as optional SQL filter conditions
type ItemsQueryParams struct {
    // Pagination
    StartIndex          *int        // OFFSET
    Limit               *int        // LIMIT (default: no limit)

    // User context
    UserId              *uuid.UUID

    // Type filters
    IncludeItemTypes    []string    // WHERE Type IN (...)
    ExcludeItemTypes    []string    // WHERE Type NOT IN (...)
    MediaTypes          []string    // WHERE MediaType IN (...)
    IsFolder            *bool       // WHERE IsFolder = ?

    // Hierarchy
    ParentId            *uuid.UUID  // WHERE ParentId = ?
    AncestorIds         []uuid.UUID // JOIN AncestorIds WHERE ParentItemId IN (...)
    TopParentIds        []uuid.UUID // WHERE TopParentId IN (...)

    // Metadata filters
    HasTmdbId           *bool
    HasImdbId           *bool
    HasTvdbId           *bool
    HasOverview         *bool
    HasOfficialRating   *bool
    IsPlayed            *bool       // JOIN UserData WHERE UserId=? AND Played = ?
    IsFavorite          *bool       // JOIN UserData WHERE UserId=? AND IsFavorite = ?
    IsResumable         *bool       // JOIN UserData WHERE UserId=? AND PlaybackPositionTicks > 0
    IsVirtualItem       *bool       // WHERE IsVirtualItem = ? (default false)
    IsLocked            *bool
    IsUnaired           *bool       // WHERE PremiereDate > NOW()
    IsMissing           *bool       // WHERE IsVirtualItem = 1

    // Content filters
    Genres              []string    // JOIN ItemValuesMap+ItemValues WHERE Type=2 AND CleanValue IN (...)
    Tags                []string
    Studios             []string
    Artists             []string
    AlbumArtists        []string
    ContributingArtists []string
    PersonIds           []uuid.UUID // JOIN PeopleBaseItemMap WHERE PeopleId IN (...)
    PersonTypes         []string
    Years               []int       // WHERE ProductionYear IN (...)
    OfficialRatings     []string    // WHERE OfficialRating IN (?)
    VideoTypes          []string

    // Ratings
    MinCommunityRating  *float64
    MinCriticRating     *float64
    MaxOfficialRating   string      // parental control ceiling
    MinPremiereDate     *time.Time
    MaxPremiereDate     *time.Time
    MinDateLastSaved    *time.Time

    // Search
    SearchTerm          string      // MATCH(Name,OriginalTitle) AGAINST (? IN BOOLEAN MODE)
    NameStartsWith      string      // WHERE CleanName LIKE 'prefix%'
    NameStartsWithOrGreater string  // WHERE SortName >= ?
    NameLessThan        string      // WHERE SortName < ?

    // Series-specific
    SeriesId            *uuid.UUID  // WHERE SeriesId = ?
    SeasonId            *uuid.UUID  // WHERE SeasonId = ?
    SeriesPresentationUniqueKey string

    // Sorting
    SortBy              []string    // "SortName", "DateCreated", "CommunityRating", "ProductionYear", etc.
    SortOrder           []string    // "Ascending", "Descending"

    // Field selection (for DTO projection)
    Fields              []string    // which extra fields to populate in response
    EnableImages        *bool
    EnableUserData      *bool
    ImageTypeLimit      *int
    EnableTotalRecordCount *bool    // if false, skip COUNT(*) query for perf

    // Misc
    CollapseBoxSetItems *bool
    ExcludeItemIds      []uuid.UUID
    Ids                 []uuid.UUID // WHERE Id IN (...)
    MinIndexNumber      *int
    MaxIndexNumber      *int
    ParentIndexNumber   *int
}
```

**SQL Query Pattern for GET /Items:**
```sql
-- Count query (run first if EnableTotalRecordCount != false)
SELECT COUNT(*) FROM BaseItems bi
[LEFT JOIN UserData ud ON ud.ItemId = bi.Id AND ud.UserId = ?]
[JOIN ItemValuesMap ivm ON ivm.ItemId = bi.Id JOIN ItemValues iv ON iv.ItemValueId = ivm.ItemValueId AND iv.Type = 2]
WHERE bi.IsVirtualItem = 0
  [AND bi.Type IN (?,...)]
  [AND bi.ParentId = ?]
  [AND bi.SeriesId = ?]
  [... all filters ...]

-- Data query (same WHERE, add ORDER BY + LIMIT OFFSET)
SELECT bi.*,
       img.Path as ImgPath, img.ImageType as ImgType, img.BlurHash, img.Width as ImgW, img.Height as ImgH,
       prov.ProviderId, prov.ProviderValue,
       ud.PlaybackPositionTicks, ud.PlayCount, ud.IsFavorite, ud.Played, ud.LastPlayedDate,
       ud.Rating as UserRating, ud.AudioStreamIndex, ud.SubtitleStreamIndex
FROM BaseItems bi
LEFT JOIN BaseItemImageInfos img ON img.ItemId = bi.Id
LEFT JOIN BaseItemProviders prov ON prov.ItemId = bi.Id
[LEFT JOIN UserData ud ON ud.ItemId = bi.Id AND ud.UserId = ?]
WHERE bi.IsVirtualItem = 0
  [... same filters ...]
ORDER BY [SortBy columns]
LIMIT ? OFFSET ?;
```

**Parental Control Filtering:**
```go
// Applied when user has MaxParentalAgeRating set
// WHERE InheritedParentalRatingValue <= userMaxRating OR InheritedParentalRatingValue IS NULL
```

---

### AGENT CARD 4: playstate-service

**Priority:** HIGH — called every 30 seconds by the player during playback.

**Goal:** Track per-user playback state: mark items played/unplayed, update resume position,
manage favorites, and report play session start/stop/progress.

**Port:** 8004

**Routes:**
```
POST   /Users/{userId}/PlayedItems/{itemId}            [SelfOrAdmin]
DELETE /Users/{userId}/PlayedItems/{itemId}            [SelfOrAdmin]
POST   /Users/{userId}/Items/{itemId}/UserData         [SelfOrAdmin]
POST   /Users/{userId}/PlayingItems/{itemId}           [SelfOrAdmin]  (play session start)
DELETE /Users/{userId}/PlayingItems/{itemId}           [SelfOrAdmin]  (play session stop)
POST   /Users/{userId}/PlayingItems/{itemId}/Progress  [SelfOrAdmin]  (progress ping)
```

**Database Tables:**
- `UserData` (primary — upsert on every progress ping)
- `Devices` (update `DateLastActivity` on progress)
- `Sessions` (if maintaining in-memory session state → use a sync.Map)

**Critical Query — Upsert UserData:**
```sql
INSERT INTO UserData (ItemId, UserId, PlaybackPositionTicks, PlayCount, IsFavorite, Played,
                      Rating, LastPlayedDate, AudioStreamIndex, SubtitleStreamIndex)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
    PlaybackPositionTicks = VALUES(PlaybackPositionTicks),
    PlayCount = VALUES(PlayCount),
    Played = VALUES(Played),
    LastPlayedDate = VALUES(LastPlayedDate),
    AudioStreamIndex = VALUES(AudioStreamIndex),
    SubtitleStreamIndex = VALUES(SubtitleStreamIndex);
```

**Business Logic:**
- `POST /PlayedItems/{itemId}` → set `Played=1`, increment `PlayCount`, set `LastPlayedDate=NOW()`, set `PlaybackPositionTicks=0`
- `DELETE /PlayedItems/{itemId}` → set `Played=0`, set `PlayCount=MAX(PlayCount-1, 0)`
- Progress ping → update `PlaybackPositionTicks` only
- Stop → if position >= 90% of runtime → mark as played

---

### AGENT CARD 5: media-service

**Priority:** MEDIUM-HIGH — images are on every page.

**Goal:** Serve item images (with on-the-fly resizing), media stream info, media segments
(intro/outro data), lyrics, video attachments, and subtitle files.

**Port:** 8005

**Routes:**
```
GET    /Items/{itemId}/Images/{imageType}              [AllowAnonymous]
GET    /Items/{itemId}/Images/{imageType}/{imageIndex} [AllowAnonymous]
HEAD   /Items/{itemId}/Images/{imageType}              [AllowAnonymous]

GET    /Items/{itemId}/MediaInfo                       [Authorized]
GET    /Items/{itemId}/PlaybackInfo                    [Authorized]
POST   /Items/{itemId}/PlaybackInfo                    [Authorized]  (body has profile)

GET    /Items/{itemId}/MediaSegments                   [Authorized]

GET    /Items/{itemId}/Lyrics                          [Authorized]

GET    /Videos/{itemId}/Attachments/{index}            [Authorized]

GET    /Items/{itemId}/Subtitles/{index}/{startPositionTicks}/Stream.{format}  [AllowAnonymous]
GET    /Items/{itemId}/Subtitles/{index}/Stream.{format}                       [AllowAnonymous]
GET    /Items/{itemId}/RemoteSearch/Subtitles/{language}                        [IsAdmin]
POST   /Items/{itemId}/Subtitles/{index}               [IsAdmin]
DELETE /Items/{itemId}/Subtitles/{index}               [IsAdmin]
GET    /Providers/Subtitles/Subtitles                  [IsAdmin]
```

**Image Serving Implementation:**

```go
// Images are served from disk with ETags for caching
// Path on disk: {CacheDir}/images/{itemId}/{imageType}-{width}x{height}.jpg

func serveImage(w http.ResponseWriter, r *http.Request, imgInfo BaseItemImageInfo, params ImageParams) {
    // 1. Check If-None-Match ETag header → 304 if matches
    // 2. If fillWidth/fillHeight requested: resize using github.com/disintegration/imaging
    // 3. Set Cache-Control: max-age=31536000 (1 year) — images are content-addressed by tag
    // 4. Set ETag: params.Tag (MD5 of DateModified)
    // 5. Serve file
}

// ETag calculation — must match C# formula:
func imageETag(dateModified time.Time) string {
    h := md5.Sum([]byte(dateModified.UTC().Format("20060102150405")))
    return hex.EncodeToString(h[:])
}
```

**PlaybackInfo Response (complex, critical for player startup):**
```go
type PlaybackInfoResponse struct {
    MediaSources      []MediaSourceInfo   `json:"MediaSources"`
    PlaySessionId     string              `json:"PlaySessionId"`  // UUID for this play session
    ErrorCode         *string             `json:"ErrorCode"`
}

type MediaSourceInfo struct {
    Protocol              string              `json:"Protocol"`       // "File"
    Id                    string              `json:"Id"`
    Path                  string              `json:"Path"`
    Type                  string              `json:"Type"`           // "Default", "Grouping"
    Container             string              `json:"Container"`      // "mkv", "mp4", etc.
    Size                  *int64              `json:"Size"`
    Name                  string              `json:"Name"`
    IsRemote              bool                `json:"IsRemote"`
    ETag                  *string             `json:"ETag"`
    RunTimeTicks          *int64              `json:"RunTimeTicks"`
    SupportsTranscoding   bool                `json:"SupportsTranscoding"`
    SupportsDirectStream  bool                `json:"SupportsDirectStream"`
    SupportsDirectPlay    bool                `json:"SupportsDirectPlay"`
    IsInfiniteStream      bool                `json:"IsInfiniteStream"`
    Bitrate               *int                `json:"Bitrate"`
    DefaultAudioStreamIndex *int              `json:"DefaultAudioStreamIndex"`
    DefaultSubtitleStreamIndex *int           `json:"DefaultSubtitleStreamIndex"`
    MediaStreams          []MediaStream       `json:"MediaStreams"`
    TranscodingUrl        *string             `json:"TranscodingUrl"`
    TranscodingSubProtocol *string            `json:"TranscodingSubProtocol"`  // "hls"
    TranscodingContainer  *string             `json:"TranscodingContainer"`    // "ts"
    DirectStreamUrl       *string             `json:"DirectStreamUrl"`
}
```

---

### AGENT CARD 6: stream-service

**Priority:** HIGH — streaming is the core user activity.

**Goal:** Serve video/audio streams, HLS playlists and segments, and trickplay thumbnails.
This service shells out to `ffmpeg` for transcoding. Direct play (static file serving) must
also be supported for maximum performance.

**Port:** 8006

**Routes:**
```
-- Direct video streaming
GET  /Videos/{itemId}/stream                           [AllowAnonymous, token via api_key]
GET  /Videos/{itemId}/stream.{container}               [AllowAnonymous, token via api_key]
HEAD /Videos/{itemId}/stream                           [AllowAnonymous]

-- HLS adaptive streaming
GET  /Videos/{itemId}/master.m3u8                      [AllowAnonymous]
GET  /Videos/{itemId}/main.m3u8                        [AllowAnonymous]
GET  /Videos/{itemId}/hls1/{playSessionId}/{segmentId}.{container} [AllowAnonymous]

-- Audio streaming
GET  /Audio/{itemId}/stream                            [AllowAnonymous]
GET  /Audio/{itemId}/stream.{container}                [AllowAnonymous]
GET  /Audio/{itemId}/Universal                         [AllowAnonymous]
GET  /Audio/{itemId}/{filename}                        [AllowAnonymous]

-- Trickplay
GET  /Videos/{itemId}/Trickplay/{width}/{index}.jpg    [AllowAnonymous]

-- HLS segment management
DELETE /Videos/{itemId}/MasterHlsVideoPlaylist         [Authorized]
DELETE /Videos/{itemId}/HlsPlaylistSegment/{segmentId} [Authorized]
```

**Direct Play Implementation:**
```go
// For static=true requests: serve the file directly with http.ServeContent
// This handles Range requests automatically for seeking
func directPlay(w http.ResponseWriter, r *http.Request, filePath string, modTime time.Time) {
    f, err := os.Open(filePath)
    if err != nil { /* 404 */ return }
    defer f.Close()
    http.ServeContent(w, r, filepath.Base(filePath), modTime, f)
}
```

**HLS Transcoding Flow:**
```go
// 1. Parse request params: bitrate, videoCodec, audioCodec, audioStreamIndex, subtitleStreamIndex, etc.
// 2. Generate unique playSessionId if not provided
// 3. Create temp dir: {CacheDir}/transcodes/{playSessionId}/
// 4. Build ffmpeg command (replicate DynamicHlsController._mediaEncoder.GetInputArgument pattern)
// 5. Start ffmpeg as background process
// 6. Wait for first segment file to appear (with timeout)
// 7. Return master.m3u8 playlist
// 8. Serve subsequent segment requests from the temp dir

// Master playlist format:
const masterPlaylistTemplate = `#EXTM3U
#EXT-X-VERSION:3
#EXT-X-STREAM-INF:BANDWIDTH=%d,RESOLUTION=%dx%d,CODECS="%s"
%s`

// Segment playlist format:
const segmentPlaylistTemplate = `#EXTM3U
#EXT-X-VERSION:3
#EXT-X-TARGETDURATION:%d
#EXT-X-MEDIA-SEQUENCE:%d
%s
`
```

**Trickplay Serving:**
```go
// Trickplay images are pre-generated and stored as:
// {DataDir}/trickplay/{itemId}/{width}/{index}.jpg (individual) or
// {DataDir}/trickplay/{itemId}/{width}/0.jpg       (sprite sheet)
// TrickplayInfos table has Width, Height, TileWidth, TileHeight, ThumbnailCount, Interval

func serveTrickplay(w http.ResponseWriter, r *http.Request, itemId string, width, index int) {
    imgPath := filepath.Join(dataDir, "trickplay", itemId, strconv.Itoa(width),
                              strconv.Itoa(index)+".jpg")
    http.ServeFile(w, r, imgPath)
}
```

**Critical Gotchas for stream-service:**
- `api_key` query param is the ONLY auth mechanism for stream URLs (browser `<video>` tags can't set headers)
- Range request support is MANDATORY for video seeking — use `http.ServeContent`
- Content-Type must match container: `video/mp4`, `video/x-matroska`, `video/webm`, etc.
- `Accept-Ranges: bytes` header required in 200 responses
- HLS segments are temporary files; implement cleanup goroutine after play session ends (30 min TTL)
- `DELETE /Videos/{id}/MasterHlsVideoPlaylist` is called when client stops playback — kill ffmpeg process

---

### AGENT CARD 7: session-service

**Priority:** MEDIUM-HIGH — required for remote control and SyncPlay features.

**Goal:** Manage active client sessions, provide remote control capabilities, implement
SyncPlay (synchronized playback across multiple users), and time synchronization.

**Port:** 8007

**Routes:**
```
GET    /Sessions                                       [Authorized]
POST   /Sessions/{sessionId}/Message                   [Authorized]
POST   /Sessions/{sessionId}/System/{command}          [Authorized]
POST   /Sessions/{sessionId}/Playing                   [Authorized]
POST   /Sessions/{sessionId}/Playing/{command}         [Authorized]
POST   /Sessions/{sessionId}/Viewing                   [Authorized]
POST   /Sessions/{sessionId}/Users/{userId}            [IsAdmin]
DELETE /Sessions/{sessionId}/Users/{userId}            [IsAdmin]
POST   /Sessions/Logout                                [Authorized]
GET    /Sessions/Capabilities                          [Authorized]
POST   /Sessions/Capabilities                          [Authorized]
POST   /Sessions/Capabilities/Full                     [Authorized]

GET    /Devices                                        [IsAdmin]
DELETE /Devices                                        [IsAdmin]  ?id=deviceId
GET    /Devices/Options                                [IsAdmin]  ?id=deviceId
POST   /Devices/Options                                [IsAdmin]  ?id=deviceId

GET    /GetUtcTime                                     [AllowAnonymous]

-- SyncPlay (WebSocket-heavy, see note)
POST   /SyncPlay/New                                   [Authorized]
POST   /SyncPlay/Join                                  [Authorized]
POST   /SyncPlay/Leave                                 [Authorized]
POST   /SyncPlay/Ready                                 [Authorized]
POST   /SyncPlay/Ping                                  [Authorized]
GET    /SyncPlay/List                                  [Authorized]
POST   /SyncPlay/SetIgnoreWait                         [Authorized]
POST   /SyncPlay/Play                                  [Authorized]
POST   /SyncPlay/Stop                                  [Authorized]
POST   /SyncPlay/Pause                                 [Authorized]
POST   /SyncPlay/Unpause                               [Authorized]
POST   /SyncPlay/Seek                                  [Authorized]
POST   /SyncPlay/Queue                                 [Authorized]
POST   /SyncPlay/SetPlaylistItem                       [Authorized]
POST   /SyncPlay/RemoveFromPlaylist                    [Authorized]
POST   /SyncPlay/MovePlaylistItem                      [Authorized]
POST   /SyncPlay/NextItem                              [Authorized]
POST   /SyncPlay/PreviousItem                          [Authorized]
POST   /SyncPlay/SetRepeatMode                         [Authorized]
POST   /SyncPlay/SetShuffleMode                        [Authorized]
POST   /SyncPlay/BufferingDone                         [Authorized]
```

**Session State Management:**
```go
// Sessions are ephemeral (not persisted to DB in Jellyfin)
// Use sync.Map for in-memory session storage
// Key: sessionId (string), Value: *SessionState

type SessionState struct {
    mu              sync.RWMutex
    ID              string
    UserID          uuid.UUID
    DeviceID        string
    DeviceName      string
    AppName         string
    AppVersion      string
    Client          string
    RemoteEndPoint  string
    LastActivityDate time.Time
    NowPlayingItem  *BaseItemDto
    PlayState       *PlayerStateInfo
    Capabilities    *ClientCapabilities
}

var sessions sync.Map // sessionId → *SessionState
```

**WebSocket for SyncPlay:**
```go
// SyncPlay uses WebSocket for real-time group coordination
// Use github.com/gorilla/websocket
// SyncPlay group state: maintained in-memory with a mutex

type SyncPlayGroup struct {
    mu          sync.Mutex
    GroupId     uuid.UUID
    PlayingItem *QueueItem
    State       string      // "Idle", "Playing", "Paused", "Waiting"
    PositionTicks int64
    LastUpdated time.Time
    Members     map[string]*SyncPlayMember  // sessionId → member
}
```

**GET /Sessions Response:**
```go
type SessionInfoDto struct {
    PlayState           *PlayerStateInfo    `json:"PlayState"`
    AdditionalUsers     []SessionUserInfo   `json:"AdditionalUsers"`
    Capabilities        *ClientCapabilities `json:"Capabilities"`
    RemoteEndPoint      string              `json:"RemoteEndPoint"`
    Id                  string              `json:"Id"`
    UserId              string              `json:"UserId"`
    UserName            string              `json:"UserName"`
    Client              string              `json:"Client"`
    LastActivityDate    JellyfinTime        `json:"LastActivityDate"`
    LastPlaybackCheckIn JellyfinTime        `json:"LastPlaybackCheckIn"`
    DeviceName          string              `json:"DeviceName"`
    DeviceId            string              `json:"DeviceId"`
    NowPlayingItem      *BaseItemDto        `json:"NowPlayingItem"`
    ApplicationVersion  string              `json:"ApplicationVersion"`
    IsActive            bool                `json:"IsActive"`
    SupportsMediaControl bool               `json:"SupportsMediaControl"`
    SupportsRemoteControl bool              `json:"SupportsRemoteControl"`
    HasCustomDeviceName bool                `json:"HasCustomDeviceName"`
    ServerId            string              `json:"ServerId"`
}
```

---

### AGENT CARD 8: metadata-service

**Priority:** MEDIUM.

**Goal:** Handle metadata refresh (triggering re-fetch from external providers), item updates,
remote image search, plugin management stubs, scheduled task management, and backup operations.

**Port:** 8008

**Routes:**
```
POST   /Items/{itemId}/Refresh                         [Authorized]
POST   /Items/{itemId}                                 [Authorized]   (item field update)

GET    /Items/RemoteSearch/Movie                       [Authorized]
POST   /Items/RemoteSearch/Movie                       [Authorized]
POST   /Items/RemoteSearch/Tv                          [Authorized]
POST   /Items/RemoteSearch/Person                      [Authorized]
POST   /Items/RemoteSearch/Book                        [Authorized]
POST   /Items/RemoteSearch/Music                       [Authorized]
POST   /Items/RemoteSearch/MusicVideo                  [Authorized]
POST   /Items/RemoteSearch/ExternalIdLookup            [Authorized]
POST   /Items/RemoteSearch/Apply/{itemId}              [IsAdmin]

GET    /RemoteImage/Providers                          [Authorized]
GET    /RemoteImage/Images                             [Authorized]
POST   /Items/{itemId}/RemoteImages/Download           [IsAdmin]

GET    /ScheduledTasks                                 [IsAdmin]
GET    /ScheduledTasks/{taskId}                        [IsAdmin]
POST   /ScheduledTasks/Running/{taskId}                [IsAdmin]
DELETE /ScheduledTasks/Running/{taskId}                [IsAdmin]
POST   /ScheduledTasks/{taskId}/Triggers               [IsAdmin]

GET    /Packages                                       [AllowAnonymous]
GET    /Packages/{name}                                [AllowAnonymous]
POST   /Packages/Installed/{name}                      [IsAdmin]
DELETE /Packages/Installed/{name}/{version}            [IsAdmin]

GET    /Plugins                                        [IsAdmin]
GET    /Plugins/{pluginId}/Configuration               [IsAdmin]
POST   /Plugins/{pluginId}/Configuration               [IsAdmin]
DELETE /Plugins/{pluginId}                             [IsAdmin]

GET    /System/Backups                                 [IsAdmin]
POST   /System/Backups/Create                          [IsAdmin]
POST   /System/Backups/Restore                         [IsAdmin]
DELETE /System/Backups/{name}                          [IsAdmin]
```

**Item Update (POST /Items/{itemId}):**
```sql
UPDATE BaseItems SET
    Name = ?, Overview = ?, Tagline = ?, OfficialRating = ?,
    CommunityRating = ?, ProductionYear = ?, PremiereDate = ?,
    ForcedSortName = ?, Genres = ?, Studios = ?, Tags = ?,
    IsLocked = ?, DateModified = NOW()
WHERE Id = ?;

-- Also update BaseItemMetadataFields for locked fields
DELETE FROM BaseItemMetadataFields WHERE ItemId = ?;
INSERT INTO BaseItemMetadataFields (ItemId, MetadataField) VALUES (?, ?), ...;
```

**Scheduled Tasks — in-memory only:**
```go
// Task state is maintained in-memory (no DB table)
// Use a registry of known tasks:
type Task struct {
    Id           string
    Name         string
    Description  string
    Category     string
    State        string   // "Idle", "Running", "Cancelling"
    LastResult   *TaskResult
    Triggers     []TaskTrigger
    // Go implementation: wrap in goroutine with context cancellation
}
```

---

### AGENT CARD 9: content-service

**Priority:** MEDIUM.

**Goal:** Serve domain-specific content browsing for movies, TV shows, music artists, channels,
Live TV, collections, playlists, and instant mix generation. These are largely thin wrappers
over `BaseItems` queries with specific type filters and extra logic.

**Port:** 8009

**Routes:**
```
-- Movies
GET    /Movies/Recommendations                         [Authorized]

-- TV Shows
GET    /Shows/NextUp                                   [Authorized]
GET    /Shows/{seriesId}/Seasons                       [Authorized]
GET    /Shows/{seriesId}/Episodes                      [Authorized]
GET    /Shows/{seriesId}/Similar                       [Authorized]

-- Artists
GET    /Artists                                        [Authorized]
GET    /Artists/{name}                                 [Authorized]
GET    /Artists/AlbumArtists                           [Authorized]

-- Channels
GET    /Channels                                       [Authorized]
GET    /Channels/{channelId}                           [Authorized]
GET    /Channels/{channelId}/Items                     [Authorized]
GET    /Channels/{channelId}/LatestItems               [Authorized]
GET    /Channels/Features                              [Authorized]

-- Live TV
GET    /LiveTV/Channels                                [Authorized]
GET    /LiveTV/Programs                                [Authorized]
GET    /LiveTV/Programs/{programId}                    [Authorized]
POST   /LiveTV/Programs                                [Authorized]
GET    /LiveTV/Recordings                              [Authorized]
GET    /LiveTV/Recordings/{recordingId}                [Authorized]
DELETE /LiveTV/Recordings/{recordingId}                [IsAdmin]
GET    /LiveTV/Timers                                  [Authorized]
GET    /LiveTV/Timers/{timerId}                        [Authorized]
POST   /LiveTV/Timers                                  [IsAdmin]
DELETE /LiveTV/Timers/{timerId}                        [IsAdmin]
GET    /LiveTV/SeriesTimers                            [Authorized]
POST   /LiveTV/SeriesTimers                            [IsAdmin]
DELETE /LiveTV/SeriesTimers/{timerId}                  [IsAdmin]
GET    /LiveTV/GuideInfo                               [Authorized]
POST   /LiveTV/ChannelMappings                         [IsAdmin]
GET    /LiveTV/TunerHosts                              [IsAdmin]
POST   /LiveTV/TunerHosts                              [IsAdmin]
DELETE /LiveTV/TunerHosts                              [IsAdmin]

-- Collections
GET    /Collections                                    [Authorized]
POST   /Collections                                    [Authorized]
POST   /Collections/{collectionId}/Items               [Authorized]
DELETE /Collections/{collectionId}/Items               [Authorized]

-- Playlists
GET    /Playlists                                      [Authorized]
POST   /Playlists                                      [Authorized]
GET    /Playlists/{playlistId}/Items                   [Authorized]
POST   /Playlists/{playlistId}/Items                   [Authorized]
DELETE /Playlists/{playlistId}/Items                   [Authorized]
POST   /Playlists/{playlistId}/Items/{itemId}/Move/{newIndex} [Authorized]

-- Instant Mix
GET    /Items/{itemId}/InstantMix                      [Authorized]
GET    /Artists/{name}/InstantMix                      [Authorized]
GET    /Albums/{albumId}/InstantMix                    [Authorized]
GET    /Songs/{songId}/InstantMix                      [Authorized]
GET    /Playlists/{playlistId}/InstantMix              [Authorized]
GET    /MusicGenres/{name}/InstantMix                  [Authorized]
```

**NextUp Logic (GET /Shows/NextUp):**
```sql
-- For each series the user has started watching:
-- Find the first unwatched episode after their last watched episode
SELECT bi.*, ud.*
FROM BaseItems bi
JOIN UserData ud ON ud.ItemId = bi.Id AND ud.UserId = ?
WHERE bi.Type = 'Episode'
  AND bi.SeriesId IN (
      -- series user has at least one played episode
      SELECT DISTINCT bi2.SeriesId FROM BaseItems bi2
      JOIN UserData ud2 ON ud2.ItemId = bi2.Id AND ud2.UserId = ? AND ud2.Played = 1
      WHERE bi2.Type = 'Episode'
  )
  AND ud.Played = 0
  AND bi.IsVirtualItem = 0
ORDER BY bi.SeriesId, bi.ParentIndexNumber, bi.IndexNumber;
-- Then: GROUP BY SeriesId, take MIN per series
```

**Collections — stored as BaseItems with Type='BoxSet':**
```sql
-- Get collection members via AncestorIds
SELECT bi.* FROM BaseItems bi
JOIN AncestorIds ai ON ai.ItemId = bi.Id
WHERE ai.ParentItemId = ?  -- the collection Id
  AND bi.IsVirtualItem = 0;
```

---

### AGENT CARD 10: system-service

**Priority:** MEDIUM — required for admin UI and server info display.

**Goal:** Serve server information, system configuration, branding, environment details,
localization data, client log submission, and activity log browsing.

**Port:** 8010

**Routes:**
```
GET    /System/Info                                    [Authorized]
GET    /System/Info/Public                             [AllowAnonymous]
GET    /System/Ping                                    [AllowAnonymous]
POST   /System/Ping                                    [AllowAnonymous]
GET    /System/Logs                                    [IsAdmin]
GET    /System/Logs/Log                                [IsAdmin]  ?name=filename
POST   /System/Restart                                 [IsAdmin]
POST   /System/Shutdown                                [IsAdmin]
GET    /System/Endpoint                                [Authorized]

GET    /System/Configuration                           [Authorized]
PUT    /System/Configuration                           [IsAdmin]
GET    /System/Configuration/{key}                     [Authorized]
POST   /System/Configuration/{key}                     [IsAdmin]

GET    /Dashboard/ConfigurationPages                   [AllowAnonymous]
GET    /Dashboard/ConfigurationPage                    [AllowAnonymous]

GET    /Branding/Configuration                         [AllowAnonymous]
GET    /Branding/Css.css                               [AllowAnonymous]
GET    /Branding/Css                                   [AllowAnonymous]

GET    /Environment/DefaultDirectoryBrowser            [IsAdmin]
GET    /Environment/DirectoryContents                  [IsAdmin]
GET    /Environment/Drives                             [IsAdmin]
GET    /Environment/ParentPath                         [IsAdmin]
GET    /Environment/ValidatePath                       [IsAdmin]
POST   /Environment/ValidatePath                       [IsAdmin]

GET    /Localization/Countries                         [AllowAnonymous]
GET    /Localization/Cultures                          [AllowAnonymous]
GET    /Localization/Options                           [AllowAnonymous]
GET    /Localization/ParentalRatings                   [AllowAnonymous]

POST   /Logs/ClientLogging                             [Authorized]

GET    /ActivityLog/Entries                            [IsAdmin]
```

**System Info Response:**
```go
type SystemInfo struct {
    LocalAddress            *string     `json:"LocalAddress"`
    ServerName              string      `json:"ServerName"`
    Version                 string      `json:"Version"`           // "10.9.11"
    ProductName             string      `json:"ProductName"`       // "Jellyfin Server"
    OperatingSystem         string      `json:"OperatingSystem"`
    Id                      string      `json:"Id"`                // ServerId GUID from config
    StartupWizardCompleted  bool        `json:"StartupWizardCompleted"`
    OperatingSystemDisplayName string   `json:"OperatingSystemDisplayName"`
    CanSelfRestart          bool        `json:"CanSelfRestart"`
    CanLaunchWebBrowser     bool        `json:"CanLaunchWebBrowser"`
    ProgramDataPath         string      `json:"ProgramDataPath"`
    LogPath                 string      `json:"LogPath"`
    ItemsByNamePath         string      `json:"ItemsByNamePath"`
    CachePath               string      `json:"CachePath"`
    InternalMetadataPath    string      `json:"InternalMetadataPath"`
    TranscodingTempPath     string      `json:"TranscodingTempPath"`
    HasUpdateAvailable      bool        `json:"HasUpdateAvailable"`
    EncoderLocation         string      `json:"EncoderLocation"`  // "System"
    SystemArchitecture      string      `json:"SystemArchitecture"` // "X64"
}
```

**Configuration — read XML config files:**
```go
// Config files are in {ConfigDir}/:
// system.xml → ServerName, IsStartupWizardCompleted, EnableRemoteAccess, etc.
// encoding.xml → FFmpegPath, EnableHardwareEncoding, etc.
// network.xml → HttpServerPortNumber, LocalNetworkSubnets, etc.

// Parse system.xml:
type SystemXmlConfig struct {
    XMLName                  xml.Name `xml:"ServerConfiguration"`
    ServerName               string   `xml:"ServerName"`
    IsStartupWizardCompleted bool     `xml:"IsStartupWizardCompleted"`
    ServerId                 string   `xml:"ServerId"`
    EnableRemoteAccess       bool     `xml:"EnableRemoteAccess"`
    LogFileRetentionDays     int      `xml:"LogFileRetentionDays"`
}
```

---

## Phase 3 — Step-by-Step Implementation Plans

For each agent card above, expand into these ordered implementation steps.

### Universal Steps (apply to every service except shared):

```
Step 1 — Module scaffold
  go mod init github.com/jellyfinhanced/<service-name>
  go get github.com/go-chi/chi/v5@latest
  go get github.com/jmoiron/sqlx@latest
  go get github.com/go-sql-driver/mysql@latest
  go get github.com/google/uuid@latest
  go get golang.org/x/crypto@latest         (auth-service only)
  go get github.com/gorilla/websocket@latest (session-service only)
  go get github.com/disintegration/imaging@latest (media-service only)
  go get github.com/testcontainers/testcontainers-go@latest (test deps)
  go get github.com/stretchr/testify@latest
  # Add local replace directive:
  echo 'replace github.com/jellyfinhanced/shared => ../shared' >> go.mod

Step 2 — Directory structure
  mkdir -p cmd/server
  mkdir -p internal/{db,handlers,middleware,dto}
  touch cmd/server/main.go
  touch internal/db/{interface.go,mysql.go,mock.go}
  touch internal/handlers/{handler_file_per_controller.go}
  touch internal/middleware/{auth.go,logging.go,recovery.go}

Step 3 — Implement DB interface
  // interface.go: define interface with one method per query group
  type Store interface {
      // list all methods
  }
  // mysql.go: *sqlx.DB implementation
  // mock.go: mock using testify/mock for unit tests

Step 4 — Implement auth middleware (use shared package)
  // Simply wire shared/auth middleware into the Chi router

Step 5 — Implement handlers (one file per controller, one function per route)
  // Each handler function signature:
  func (h *Handler) MethodName(w http.ResponseWriter, r *http.Request) {
      auth, ok := sharedauth.GetAuth(r.Context())
      if !ok { response.WriteError(w, 401, "Unauthorized"); return }
      // ... parse params, call DB, map to DTO, call response.WriteJSON
  }

Step 6 — Wire router
  // cmd/server/main.go:
  func main() {
      ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
      defer stop()

      db := mustOpenDB()
      h := handlers.New(db)

      r := chi.NewRouter()
      r.Use(response.ServerHeaderMiddleware(serverID, version))
      r.Use(sharedauth.Middleware(db))
      r.Use(middleware.RequestLogger)
      r.Use(middleware.Recovery)

      // mount routes
      r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
          w.WriteHeader(200)
          w.Write([]byte(`{"status":"ok"}`))
      })
      // ... route registrations

      srv := &http.Server{Addr: ":"+port, Handler: r}
      go srv.ListenAndServe()
      <-ctx.Done()
      srv.Shutdown(context.Background())
  }

Step 7 — Write unit tests
  // For every handler: use httptest.NewRecorder, inject mock Store
  // Test: happy path, 401 on missing auth, 404 on unknown resource

Step 8 — Write integration tests
  // Use testcontainers-go to spin up MySQL
  // Load schema from migration SQL files
  // Insert seed data
  // Run actual HTTP requests against the service
  // Assert JSON response shape and status codes

Step 9 — Write Dockerfile
  FROM golang:1.23-alpine AS builder
  WORKDIR /app
  COPY go.mod go.sum ./
  COPY ../shared ../shared
  RUN go mod download
  COPY . .
  RUN CGO_ENABLED=0 GOOS=linux go build -o /service ./cmd/server

  FROM alpine:3.21
  COPY --from=builder /service /service
  EXPOSE <port>
  HEALTHCHECK CMD wget -qO- http://localhost:<port>/healthz || exit 1
  CMD ["/service"]

Step 10 — Manual verification
  # Compare key responses against live C# server:
  curl -s http://localhost:<port>/<route> | jq . > go_response.json
  curl -s http://csharp-server/<route> | jq . > cs_response.json
  diff go_response.json cs_response.json
```

---

## Phase 4 — Agent Execution Order & Dependency Graph

```
Phase 0 (BLOCKING, sequential):
  └── shared module (Agent 0)

Phase 1 (BLOCKING, sequential, depends on shared):
  └── auth-service (Agent 1)

Phase 2 (PARALLEL, depends on shared + auth-service):
  ├── user-service (Agent 2)
  ├── library-service (Agent 3)  ← longest, most complex
  ├── playstate-service (Agent 4)
  └── media-service (Agent 5)

Phase 3 (PARALLEL, depends on shared + auth-service):
  ├── stream-service (Agent 6)
  ├── session-service (Agent 7)
  └── metadata-service (Agent 8)

Phase 4 (PARALLEL, depends on Phase 2 + 3 designs):
  ├── content-service (Agent 9)   ← depends on library-service patterns
  └── system-service (Agent 10)

Phase 5 (SEQUENTIAL):
  └── gateway configuration (Nginx)

Phase 6 (SEQUENTIAL):
  └── Integration test suite
```

**Estimated parallel agent slots needed:** 4 (to handle Phase 2 in parallel)

---

## Phase 5 — Rollout Strategy

### Stage 1: Shadow Mode (Week 1-2)
Run both C# and all Go services simultaneously. All traffic goes to C#. Go services receive
mirrored traffic (via Nginx `mirror` module) but their responses are discarded. Monitor error
rates and compare DB state before/after.

```nginx
# In the C# proxy location:
mirror /mirror;
mirror_request_body on;

location /mirror {
    internal;
    proxy_pass http://go_gateway;
}
```

### Stage 2: Canary — Non-Streaming (Week 3)
Route 10% of requests for non-streaming endpoints to Go services. Observe:
- HTTP error rate delta < 0.1%
- p99 response time <= C# p99
- Zero "NaN" or null field complaints from web client

### Stage 3: Per-Service Cutover (Weeks 4-8)
Promote one service per week in order: system → auth → user → playstate → session → library →
media → metadata → content → stream

```nginx
# Cutover switch per service (change weight):
upstream library_backend {
    server 127.0.0.1:8003 weight=100;  # Go
    server 127.0.0.1:8096 weight=0;    # C# (keep at 0 until stable, remove after 72h)
}
```

### Stage 4: C# Shutdown (After 72h stability on all Go services)
```bash
systemctl stop jellyfin
# Keep C# process definition for 2 weeks (rollback option)
# After 2 weeks with no rollback → rm /usr/lib/jellyfin/
```

**Rollback procedure (any stage):**
```nginx
# Instant rollback: flip all upstream weights back to C# in nginx.conf, reload
nginx -t && nginx -s reload
```

**Smoke tests per service after promotion:**
```bash
# auth-service
curl -s -X POST http://localhost:8080/Users/AuthenticateByName \
  -H "Content-Type: application/json" \
  -H "X-Emby-Authorization: MediaBrowser Client=Test, Device=curl, DeviceId=test1, Version=1.0" \
  -d '{"Username":"admin","Pw":"password"}' | jq .AccessToken

# library-service
curl -s "http://localhost:8080/Items?IncludeItemTypes=Movie&Limit=5" \
  -H "X-Emby-Authorization: MediaBrowser Token=<token>" | jq '{count:.TotalRecordCount, first:.Items[0].Name}'

# playstate-service
curl -s -X POST "http://localhost:8080/Users/<uid>/PlayedItems/<itemId>" \
  -H "X-Emby-Authorization: MediaBrowser Token=<token>" -w "%{http_code}"
# Expected: 200

# stream-service
curl -sI "http://localhost:8080/Videos/<itemId>/stream?api_key=<token>&static=true" | head -5
# Expected: HTTP/1.1 200, Content-Type: video/*, Accept-Ranges: bytes
```

---

## Phase 6 — Risk Register & Open Decisions

Each item below requires a human decision before the relevant agent begins work.

### R1 — Password Storage (DECISION NEEDED: NONE)
**Finding:** Jellyfin uses BCrypt. `golang.org/x/crypto/bcrypt` is a direct equivalent.
No decision needed — use BCrypt with the same cost factor (10).

### R2 — Transcoding Strategy
**Decision needed:** Should stream-service shell out to `ffmpeg` binary (matching C# behavior)
or use CGO bindings to libav?
**Recommendation:** Shell out to `ffmpeg` binary. C# does the same. This avoids CGO complexity,
preserves identical command-line arguments, and makes debugging straightforward.
**Risk:** Process management overhead. Mitigate with a `sync.Map` of running processes and
`cmd.WaitDelay` to enforce cleanup.

### R3 — SyncPlay Horizontal Scaling
**Finding:** SyncPlay group state is in-memory in C#. If session-service runs as a single
instance, same approach works. If replicated, state must be in Redis.
**Decision needed:** Will session-service run as a single instance or be load-balanced?
**Recommendation:** Single instance during initial rollout. Add Redis-backed session state
in a follow-up if horizontal scaling becomes needed.

### R4 — Live TV / Tuner Integration
**Finding:** Live TV requires DVR tuner hardware drivers (HDHomeRun, IPTV M3U) and a guide
data fetcher. This is tightly coupled to C#-specific plugin interfaces.
**Recommendation:** Defer Live TV to a dedicated follow-up agent (`livetv-service`) in Phase 4
of rollout. During transition, proxy all `/LiveTV/*` routes to the C# server.

### R5 — Plugin System
**Finding:** Jellyfin's C# plugin system uses .NET Assembly loading. This cannot be replicated
in Go without a complete redesign.
**Decision needed:** Which currently-installed plugins are in active use?
**Recommendation:** Audit installed plugins (check `{DataDir}/plugins/`). For each active plugin,
either bake its functionality into the appropriate Go service or keep C# running in sidecar mode
for plugin traffic only.

### R6 — Image Resizing Library
**Decision needed:** Which Go image processing library to use?
**Options:**
- `github.com/disintegration/imaging` — pure Go, no CGO, good quality, slower
- `github.com/davidbyttow/govips` — CGO bindings to libvips, fast, requires libvips
- Standard `image/jpeg` + `golang.org/x/image` — no dependencies, slowest
**Recommendation:** Start with `disintegration/imaging` for zero-dependency simplicity. If
image serving becomes a CPU bottleneck (> 1000 resize ops/second), switch to govips.

### R7 — Metadata Provider HTTP Clients
**Finding:** ItemRefreshController triggers metadata re-fetch from TMDB, TVDb, MusicBrainz.
These are complex integrations (OAuth, rate limiting, multi-page responses).
**Recommendation:** For the initial rewrite, metadata-service's `/Items/{id}/Refresh` should
trigger the C# server's refresh endpoint (proxy the call) during the transition period.
After C# is retired, implement native Go TMDB/TVDb clients as a separate agent task.

### R8 — `Data` Column in BaseItems
**Finding:** `BaseItems.Data` is a `longtext` JSON blob that stores extended properties not
in dedicated columns (e.g., `ProductionLocations`, `ProviderIds`, `ExtraData`).
**Decision needed:** Should Go services deserialize this column?
**Recommendation:** Yes — define a `BaseItemData` struct and unmarshal this field. Reading it
is required to populate several `BaseItemDto` fields. Examine `MediaBrowser.Controller/Entities/BaseItem.cs`
and the SQLite migration's serialization format to understand the exact JSON schema.

### R9 — Quick Connect State
**Finding:** Quick Connect (QR code pairing) uses in-memory state (pending codes that expire).
**Recommendation:** Store Quick Connect state in a `sync.Map` with TTL in auth-service.
If auth-service restarts, pending Quick Connect attempts are lost (acceptable — user retries).

### R10 — Config File Writes During Transition
**Finding:** C# server writes XML config files when settings change.
**Recommendation:** During the shadow and canary phases, system-service should **read** XML
configs but route **write** operations (POST /System/Configuration) through to the C# server.
This preserves config consistency while C# is still running. After full cutover, system-service
writes config files directly.

---

## Output Requirements for Goose

Produce all output in `/home/bowens/Code/JellyFinhanced/Kabletown/` with this layout:

```
/
├── ARCHITECTURE.md          (Phase 1 architecture document)
├── EXECUTION-ORDER.md       (Phase 4 dependency graph with estimated parallelism)
├── RISKS.md                 (Phase 6 decisions with your recommendations)
├── shared/                  (Agent 0 full implementation)
├── gateway/
│   ├── nginx.conf
│   └── docker-compose.yml
├── auth-service/            (Agent 1 full implementation)
├── user-service/            (Agent 2 full implementation)
├── library-service/         (Agent 3 full implementation — stub queries if too large, mark TODO)
├── playstate-service/       (Agent 4 full implementation)
├── media-service/           (Agent 5 full implementation)
├── stream-service/          (Agent 6 full implementation)
├── session-service/         (Agent 7 full implementation)
├── metadata-service/        (Agent 8 full implementation)
├── content-service/         (Agent 9 full implementation)
├── system-service/          (Agent 10 full implementation)
└── integration-tests/
    ├── smoke_test.go        (curl-equivalent Go HTTP tests against full stack)
    └── docker-compose.test.yml
```

**If the full implementation of library-service is too large for a single response:**
1. Complete `shared/` and `auth-service/` fully.
2. For `library-service/`, produce complete scaffolding + the 10 most complex query implementations
   (GET /Items with full filter builder, GET /Users/{id}/Items, GET /Shows/NextUp).
3. Mark remaining handlers with `// TODO: Agent library-service continuation needed`.
4. For all other services, produce complete implementations.

**Verification step (run this at the end):**
```bash
cd go-rewrite
find . -name "*.go" | xargs gofmt -l    # must produce no output
find . -name "*.go" | xargs go vet ./... # must produce no errors
```

---

*End of Goose AI Prompt — JellyFinhanced Go Micro-API Conversion Plan*
*Generated: 2026-03-12 | Source: /home/bowens/Code/JellyFinhanced*
