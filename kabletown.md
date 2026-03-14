# Kabletown — Go Microservice Rewrite: Complete Agent Task List

**Project:** Kabletown
**Goal:** 100% wire-compatible Go replacement for JellyFinhanced (.NET 10 / ASP.NET Core)
**Frontend:** React 18 / jellyfin-web (unchanged — must receive identical JSON)
**Date:** 2026-03-13

---

## 1. LOCKED TECHNOLOGY STACK

| Concern | Choice | Reason |
|---------|--------|--------|
| Go version | 1.23+ | Required for range-over-func |
| HTTP router | `github.com/go-chi/chi/v5` | stdlib-compatible, composable middleware |
| Database driver | `github.com/jmoiron/sqlx` + `github.com/go-sql-driver/mysql` | Explicit SQL, type-safe scans, no magic |
| Password hashing | `golang.org/x/crypto/bcrypt` cost=11 | Must match C# BCrypt cost exactly |
| WebSocket | `github.com/gorilla/websocket` | Most mature, concurrent-write safe |
| Session broadcast | Redis pub/sub (`github.com/redis/go-redis/v9`) | Required for horizontal scaling |
| Image processing | `github.com/disintegration/imaging` | Pure Go, no CGO |
| Logging | `go.uber.org/zap` | Structured JSON, matches Serilog output shape |
| Config | `github.com/spf13/viper` | Env + file config |
| UUID | `github.com/google/uuid` | Standard UUID v4 generation |
| Testing | `github.com/stretchr/testify` + `github.com/testcontainers/testcontainers-go` | Real MySQL integration tests |
| FFmpeg | shell subprocess (`os/exec`) | Matches C# implementation exactly |
| OpenAPI docs | `github.com/swaggo/swag` | Auto-generate from handler annotations |
| Gateway | nginx | Reverse proxy, static file serving |

---

## 2. WIRE COMPATIBILITY REQUIREMENTS (NON-NEGOTIABLE)

The React frontend uses `@jellyfin/sdk` which validates responses strictly.

### 2.1 JSON Serialization
- **Field names:** PascalCase always. `{"Id":"...","Name":"..."}` NOT `{"id":"...","name":"..."}`
- **GUIDs:** lowercase hyphenated: `"3f2504e0-4f89-11d3-9a0c-0305e82c3301"`
- **Timestamps:** ISO 8601 UTC with 7 decimal places: `"2024-01-15T22:30:00.0000000Z"`
- **Duration:** Ticks (1 tick = 100 nanoseconds). 2 hours = `72000000000`
- **Empty arrays:** `[]` never `null`
- **Absent optional fields:** omit from JSON (use `omitempty`)
- **Pagination envelope:** always `{"Items":[...],"TotalRecordCount":N,"StartIndex":N}`
- **Error format:** `{"Message":"...","StatusCode":NNN}`

### 2.2 Required Response Headers (all endpoints)
```
Content-Type: application/json; charset=utf-8
X-Application-Version: 10.10.0
X-MediaBrowser-Server-Id: <server-uuid from config>
```

### 2.3 Authentication Header Format
```
X-Emby-Authorization: MediaBrowser Client="Jellyfin Web", Device="Firefox", DeviceId="abc123", Version="10.9.0", Token="<access_token>"
```
Also accepted: `Authorization: MediaBrowser Token="<token>"` and `?api_key=<token>` query param.

### 2.4 HTTP Status Codes
| Scenario | Code |
|----------|------|
| Success with body | 200 |
| Success no body (DELETE, some POST) | 204 |
| Invalid input | 400 |
| Missing/invalid token | 401 |
| Insufficient permissions | 403 |
| Resource not found | 404 |
| Method not allowed | 405 |
| Payload too large | 413 |
| Server error | 500 |

### 2.5 Image URL Format
```
/Items/{itemId}/Images/{imageType}?fillHeight=400&fillWidth=400&quality=90&tag={etag}
/Items/{itemId}/Images/{imageType}/{imageIndex}
```

### 2.6 Streaming URL Formats
```
/Videos/{itemId}/stream                          # direct file stream
/Videos/{itemId}/stream.{container}              # with extension
/Videos/{itemId}/live.m3u8                       # HLS master playlist
/Videos/{itemId}/hls1/{segmentId}/stream.ts      # HLS segment
/Audio/{itemId}/stream
/Audio/{itemId}/stream.{container}
```

---

## 3. PROJECT DIRECTORY STRUCTURE

```
Kabletown/
├── shared/                    ✅ STARTED — foundation package
├── auth-service/              ✅ STARTED — authentication, API keys, QuickConnect
├── collection-service/        ✅ STARTED — collections
├── playlist-service/          ✅ STARTED — playlists
├── user-service/              🔴 TODO — user CRUD, user views, display prefs
├── library-service/           🔴 TODO — items, library, filters, years, user library
├── browse-service/            🔴 TODO — artists, genres, movies, shows, studios, persons, trailers, channels
├── playback-service/          🔴 TODO — HLS, audio, sessions, playstate, syncplay
├── media-service/             🔴 TODO — images, subtitles, trickplay, lyrics, attachments, media info
├── metadata-service/          🔴 TODO — item refresh, item update, remote search, packages, plugins, tasks
├── admin-service/             🔴 TODO — system, config, activity, branding, localization, backup, dashboard
├── livetv-service/            🔴 TODO (Phase 2) — live TV, channels
├── search-service/            🔴 TODO — search, suggestions, instant mix
├── nginx.conf                 🟡 PARTIAL — needs complete routing table
├── docker-compose.yml         🟡 PARTIAL — needs all services
└── kabletown.md               ← this file
```

Each service follows this internal layout:
```
{service}/
├── cmd/server/main.go         # Entry point, router setup, graceful shutdown
├── internal/
│   ├── handlers/              # HTTP handler funcs (one file per controller group)
│   ├── service/               # Business logic
│   ├── db/                    # sqlx queries, repository pattern
│   │   └── mock.go            # Mock DB for unit tests
│   ├── middleware/            # Service-specific middleware
│   └── dto/                   # Service-specific request/response types
├── tests/
│   └── integration_test.go    # testcontainers-go tests
├── Dockerfile                 # Multi-stage: golang:1.23-alpine → alpine:3.21
├── .env.example
├── go.mod
└── README.md
```

---

## 4. DATABASE SCHEMA (MySQL 8.0)

### Core Tables (used across services)

#### `Users`
```sql
Id               CHAR(36) PRIMARY KEY,  -- lowercase UUID
Name             VARCHAR(255) NOT NULL,
Password         VARCHAR(255),          -- BCrypt hash, cost=11
PasswordSalt     VARCHAR(255),
Email            VARCHAR(255),
IsDisabled       TINYINT(1) DEFAULT 0,
IsHidden         TINYINT(1) DEFAULT 0,
PrimaryImageTag  VARCHAR(255),
Configuration    LONGTEXT,              -- JSON blob
Policy           LONGTEXT               -- JSON blob
```

#### `Devices`
```sql
Id                    CHAR(36) PRIMARY KEY,
UserId                CHAR(36) NOT NULL REFERENCES Users(Id),
DeviceId              VARCHAR(255) NOT NULL,
AccessToken           VARCHAR(255) NOT NULL UNIQUE,
FriendlyName          VARCHAR(255),
AppName               VARCHAR(255),
AppVersion            VARCHAR(50),
Created               DATETIME NOT NULL DEFAULT NOW(),
DateLastActivity      DATETIME,
LastUserId            CHAR(36),
IsActive              TINYINT(1) DEFAULT 1,
Configuration         LONGTEXT           -- JSON blob
INDEX idx_devices_token (AccessToken),
INDEX idx_devices_device_id (DeviceId)
```

#### `ApiKeys`
```sql
Id            CHAR(36) PRIMARY KEY,
AccessToken   VARCHAR(255) NOT NULL UNIQUE,
Name          VARCHAR(255) NOT NULL,
DateCreated   DATETIME NOT NULL DEFAULT NOW(),
IsAdmin       TINYINT(1) DEFAULT 1
INDEX idx_apikeys_token (AccessToken)
```

#### `BaseItems`
```sql
Id                      CHAR(36) PRIMARY KEY,
Type                    VARCHAR(100) NOT NULL,     -- Movie, Series, Episode, etc.
Name                    VARCHAR(500),
CleanName               VARCHAR(500),
OriginalTitle           VARCHAR(500),
SortName                VARCHAR(500),
ParentId                CHAR(36),
SeriesId                CHAR(36),
SeasonId                CHAR(36),
TopParentId             CHAR(36),
PresentationUniqueKey   VARCHAR(255),
Path                    LONGTEXT,
IsVirtualItem           TINYINT(1) DEFAULT 0,
IsFolder                TINYINT(1) DEFAULT 0,
IsMovie                 TINYINT(1) DEFAULT 0,
IsSeries                TINYINT(1) DEFAULT 0,
IsNews                  TINYINT(1) DEFAULT 0,
IsKids                  TINYINT(1) DEFAULT 0,
IsSports                TINYINT(1) DEFAULT 0,
Overview                LONGTEXT,
Tagline                 VARCHAR(1000),
OfficialRating          VARCHAR(50),
CustomRating            VARCHAR(50),
CommunityRating         FLOAT,
CriticRating            FLOAT,
ProductionYear          INT,
PremiereDate            DATETIME,
DateCreated             DATETIME,
DateLastSaved           DATETIME,
DateLastMediaAdded      DATETIME,
RunTimeTicks            BIGINT DEFAULT 0,
ExtraType               VARCHAR(100),
IndexNumber             INT,
ParentIndexNumber       INT,
ChannelId               VARCHAR(255),
ExternalId              VARCHAR(255),
ExternalSeriesId        VARCHAR(255),
ProviderIds             LONGTEXT,       -- JSON
Audio                   VARCHAR(50),
MediaType               VARCHAR(50),
ForcedSortName          VARCHAR(500),
LockedFields            VARCHAR(1000),
Studios                 LONGTEXT,       -- JSON array
Genres                  LONGTEXT,       -- JSON array
Tags                    LONGTEXT,       -- JSON array
ImageInfosJson          LONGTEXT,       -- JSON (BaseItemImageInfos)
SeriesPresentationUniqueKey VARCHAR(255),
AlbumId                 CHAR(36),
AlbumArtists            LONGTEXT,       -- JSON array
ExtraIds                LONGTEXT,
TotalBitrate            INT,
Size                    BIGINT,
Width                   INT,
Height                  INT,

-- Performance indexes (from 2026-03-09 migration)
INDEX IX_BaseItems_Type_IsVirtualItem_SortName (Type, IsVirtualItem, SortName),
INDEX IX_BaseItems_ParentId_IsVirtualItem_Type (ParentId, IsVirtualItem, Type),
INDEX IX_BaseItems_SeriesId_IsVirtualItem (SeriesId, IsVirtualItem),
INDEX IX_BaseItems_Type_IsVirtualItem_DateCreated (Type, IsVirtualItem, DateCreated DESC),
INDEX IX_BaseItems_TopParentId (TopParentId),
FULLTEXT INDEX FT_BaseItems_Name_OriginalTitle (Name, OriginalTitle)
```

#### `UserData`
```sql
UserId                  CHAR(36) NOT NULL REFERENCES Users(Id),
ItemId                  CHAR(36) NOT NULL REFERENCES BaseItems(Id),
Played                  TINYINT(1) DEFAULT 0,
PlayCount               INT DEFAULT 0,
IsFavorite              TINYINT(1) DEFAULT 0,
PlaybackPositionTicks   BIGINT DEFAULT 0,
LastPlayedDate          DATETIME,
Rating                  FLOAT,
PRIMARY KEY (UserId, ItemId),
INDEX IX_UserData_UserId_IsFavorite (UserId, IsFavorite),
INDEX IX_UserData_UserId_Played (UserId, Played),
INDEX IX_UserData_UserId_PlaybackPositionTicks (UserId, PlaybackPositionTicks)
```

#### `ActivityLogs`
```sql
Id              CHAR(36) PRIMARY KEY,
Name            VARCHAR(512) NOT NULL,
Overview        VARCHAR(512),
ShortOverview   VARCHAR(512),
Type            VARCHAR(256) NOT NULL,
ItemId          VARCHAR(256),
Date            DATETIME NOT NULL,
UserId          CHAR(36),
Username        VARCHAR(256),
Severity        VARCHAR(50) NOT NULL,
INDEX idx_activity_date (Date DESC),
INDEX idx_activity_type (Type),
INDEX idx_activity_userid (UserId)
```

#### `MediaStreamInfos`
```sql
ItemId          CHAR(36) NOT NULL,
StreamIndex     INT NOT NULL,
StreamType      VARCHAR(50),           -- Video, Audio, Subtitle, Data
Codec           VARCHAR(50),
Language        VARCHAR(10),
ChannelLayout   VARCHAR(50),
Profile         VARCHAR(50),
AspectRatio     VARCHAR(20),
Path            LONGTEXT,
IsInterlaced    TINYINT(1),
BitRate         INT,
Channels        INT,
SampleRate      INT,
IsDefault       TINYINT(1),
IsForced        TINYINT(1),
IsExternal      TINYINT(1),
Height          INT,
Width           INT,
AverageFrameRate FLOAT,
RealFrameRate   FLOAT,
Level           FLOAT,
PixelFormat     VARCHAR(50),
BitDepth        INT,
IsAnamorphic    TINYINT(1),
RefFrames       INT,
CodecTag        VARCHAR(50),
Comment         VARCHAR(500),
NalLengthSize   VARCHAR(10),
Title           VARCHAR(500),
TimeBase        VARCHAR(50),
CodecTimeBase   VARCHAR(50),
ColorPrimaries  VARCHAR(50),
ColorSpace      VARCHAR(50),
ColorTransfer   VARCHAR(50),
ColorRange      VARCHAR(50),
DisplayTitle    VARCHAR(500),
IsHearingImpaired TINYINT(1),
PRIMARY KEY (ItemId, StreamIndex)
```

#### `BaseItemImageInfos`
```sql
ItemId      CHAR(36) NOT NULL,
ImageType   INT NOT NULL,              -- 0=Primary, 1=Art, 2=Backdrop, ...
ImageIndex  INT NOT NULL DEFAULT 0,
Path        LONGTEXT,
DateModified DATETIME,
Width       INT,
Height      INT,
Size        BIGINT,
ImageTag    VARCHAR(255),
BlurHash    VARCHAR(255),
PRIMARY KEY (ItemId, ImageType, ImageIndex)
```

#### `Permissions` / `Preferences`
```sql
-- Permissions
Id          INT AUTO_INCREMENT PRIMARY KEY,
UserId      CHAR(36) NOT NULL REFERENCES Users(Id),
Kind        INT NOT NULL,              -- PermissionKind enum value
Value       TINYINT(1) NOT NULL DEFAULT 0

-- Preferences
Id          INT AUTO_INCREMENT PRIMARY KEY,
UserId      CHAR(36) NOT NULL REFERENCES Users(Id),
Kind        INT NOT NULL,
Value       LONGTEXT
```

#### `ItemValues` / `ItemValuesMap` (genres, studios, tags, artists)
```sql
-- ItemValues
Id          CHAR(36) PRIMARY KEY,
Type        INT NOT NULL,   -- 0=Artist,1=AlbumArtist,2=MusicGenre,3=Genre,4=Studio,5=Tag,6=Person
Value       VARCHAR(500) NOT NULL COLLATE utf8mb4_unicode_ci,
CleanValue  VARCHAR(500),
INDEX IX_ItemValues_Type_Value (Type, CleanValue)

-- ItemValuesMap
ItemValueId CHAR(36) NOT NULL REFERENCES ItemValues(Id),
ItemId      CHAR(36) NOT NULL REFERENCES BaseItems(Id),
PRIMARY KEY (ItemValueId, ItemId),
INDEX IX_ItemValuesMap_ItemId (ItemId)
```

#### `AncestorIds`
```sql
ItemId          CHAR(36) NOT NULL,
ParentItemId    CHAR(36) NOT NULL,
AncestorIdText  VARCHAR(100),
PRIMARY KEY (ItemId, ParentItemId),
INDEX IX_AncestorIds_ParentItemId (ParentItemId)
```

#### `Chapters`
```sql
ItemId          CHAR(36) NOT NULL,
ChapterIndex    INT NOT NULL,
StartPositionTicks BIGINT,
Name            VARCHAR(500),
ImagePath       LONGTEXT,
ImageDateModified DATETIME,
PRIMARY KEY (ItemId, ChapterIndex)
```

#### `MediaSegments`
```sql
Id          CHAR(36) PRIMARY KEY,
ItemId      CHAR(36) NOT NULL,
Type        INT NOT NULL,     -- 0=Unknown,1=Commercial,2=Preview,3=Recap,4=Outro,5=Intro,6=Credits
TypeIndex   INT,
StartTicks  BIGINT NOT NULL,
EndTicks    BIGINT NOT NULL,
INDEX idx_mediasegments_itemid (ItemId)
```

#### `TrickplayInfos`
```sql
ItemId          CHAR(36) NOT NULL,
Width           INT NOT NULL,
Interval        INT NOT NULL,
TileWidth       INT NOT NULL,
TileHeight      INT NOT NULL,
ThumbnailCount  INT NOT NULL,
Height          INT NOT NULL,
Bandwidth       INT,
PRIMARY KEY (ItemId, Width)
```

---

## 5. AUTHENTICATION SYSTEM

### 5.1 Token Resolution (middleware shared across all services)

```
Request arrives with X-Emby-Authorization header
  → Parse: Client, Device, DeviceId, Version, Token
  → SELECT * FROM Devices WHERE AccessToken = ? AND IsActive = 1
      → Found: load userId, isAdmin from device row
  → Fallback: SELECT * FROM ApiKeys WHERE AccessToken = ?
      → Found: isApiKey=true, isAdmin=true
  → Neither found: 401 Unauthorized
  → Store in request context: userId, deviceId, isAdmin, isApiKey, token
```

### 5.2 Authorization Policies

Every handler must enforce one of these policies:

| Policy | Rule |
|--------|------|
| `Anonymous` | No token required |
| `Authenticated` | Valid token required |
| `RequireAdmin` | IsAdmin = true |
| `SelfOrAdmin` | userId == pathParam{userId} OR isAdmin |
| `LocalOrAdmin` | Request from LAN subnet OR isAdmin |
| `CollectionManagement` | Permission bit `EnableCollectionManagement` |
| `Download` | Permission bit `EnableContentDownloading` |
| `SyncPlay` | Permission bit `EnableSyncPlay` |

### 5.3 AllowAnonymous Paths
```
GET  /System/Info/Public
GET  /Branding/Configuration
GET  /Branding/Css
GET  /Branding/Css.css
GET  /Users/Public
GET  /Startup/Configuration
POST /Startup/Complete
GET  /Startup/RemoteAccess
POST /Startup/User
GET  /System/Ping
POST /System/Ping
GET  /healthz
```

---

## 6. WEBSOCKET PROTOCOL

### 6.1 Endpoint
```
GET /socket?api_key=<token>&deviceId=<id>
```
Nginx upgrades this to WebSocket and proxies to `session-service:8008`.

### 6.2 Message Format (both directions)
```json
{
  "MessageType": "Sessions",
  "Data": "...",
  "MessageId": "uuid"
}
```

### 6.3 Server → Client Message Types
| MessageType | Data Type | When Sent |
|-------------|-----------|-----------|
| `Sessions` | `SessionInfoDto[]` | Playback start/stop, session state change |
| `ScheduledTasksInfo` | `TaskInfo[]` | Task progress updates |
| `ActivityLogEntry` | `ActivityLogEntry` | New activity logged |
| `UserDataChanged` | `UserDataChangeInfo` | Playback position saved |
| `LibraryChanged` | `LibraryUpdateInfo` | Items added/removed |
| `KeepAlive` | `null` | Every 30s heartbeat |
| `ForceKeepAlive` | `int` (seconds) | Tell client to refresh |

### 6.4 Client → Server Message Types
| MessageType | Action |
|-------------|--------|
| `SessionsStart` | Subscribe to session updates (Data: "0,1500") |
| `SessionsStop` | Unsubscribe |
| `ScheduledTasksInfoStart` | Subscribe to task updates |
| `ScheduledTasksInfoStop` | Unsubscribe |
| `ActivityLogEntryStart` | Subscribe to activity log |
| `ActivityLogEntryStop` | Unsubscribe |

---

## 7. SERVICE TASK CARDS

---

### SERVICE 0: `shared` ✅ STARTED

**Status:** Foundation types exist. Verify completeness against this checklist.

**Files required:**
```
shared/
├── auth/
│   ├── middleware.go    ✅ ParseMediaBrowserHeader, context keys
│   ├── context.go       🔴 GetUserID(ctx), GetDeviceID(ctx), IsAdmin(ctx), IsApiKey(ctx)
│   └── policies.go      🔴 RequireAdmin(next), RequireAuth(next), RequireSelfOrAdmin(next)
├── db/
│   ├── factory.go       ✅ *sqlx.DB from DATABASE_URL env
│   ├── pagination.go    🔴 type PaginationParams{StartIndex,Limit}; ExtractPagination(r)
│   └── transaction.go   🔴 WithTransaction(db, func(tx) error) error
├── dto/
│   ├── types.go         ✅ BaseItemDto, QueryResult, UserDataDto, MediaSourceInfo
│   ├── user.go          🔴 UserDto, AuthenticationResult, UserPolicy, UserConfiguration
│   ├── session.go       🔴 SessionInfoDto, ClientCapabilitiesDto, PlaystateDto
│   ├── system.go        🔴 SystemInfo, PublicSystemInfo, TaskInfo, LogFile
│   └── enums.go         🔴 ItemType consts, ImageType consts, SortBy consts
├── config/
│   └── loader.go        ✅ viper config loading
├── response/
│   ├── json.go          ✅ OK, Created, NoContent, BadRequest, Unauthorized, etc.
│   └── headers.go       🔴 AddRequiredHeaders(w, serverID) — adds X-Application-Version etc.
└── middleware/
    ├── logger.go        🔴 Zap structured request logging (method, path, status, duration_ms)
    ├── recovery.go      🔴 Panic recovery → 500
    ├── response_time.go 🔴 X-Response-Time-ms header
    └── cors.go          🔴 CORS: allow all origins, credentials, all methods
```

**`dto/user.go` must define:**
```go
type UserDto struct {
    Id                       string            `json:"Id"`
    Name                     string            `json:"Name"`
    ServerId                  string            `json:"ServerId,omitempty"`
    PrimaryImageTag           string            `json:"PrimaryImageTag,omitempty"`
    HasPassword               bool              `json:"HasPassword"`
    HasConfiguredPassword     bool              `json:"HasConfiguredPassword"`
    HasConfiguredEasyPassword bool              `json:"HasConfiguredEasyPassword"`
    EnableAutoLogin           bool              `json:"EnableAutoLogin"`
    LastLoginDate             *time.Time        `json:"LastLoginDate,omitempty"`
    LastActivityDate          *time.Time        `json:"LastActivityDate,omitempty"`
    Configuration             *UserConfiguration `json:"Configuration,omitempty"`
    Policy                    *UserPolicy        `json:"Policy,omitempty"`
    IsAdministrator           bool              `json:"IsAdministrator"`
    IsDisabled                bool              `json:"IsDisabled"`
    IsHidden                  bool              `json:"IsHidden"`
}

type AuthenticationResult struct {
    User          *UserDto `json:"User"`
    SessionInfo   *SessionInfoDto `json:"SessionInfo,omitempty"`
    AccessToken   string   `json:"AccessToken"`
    ServerId      string   `json:"ServerId"`
}
```

---

### SERVICE 1: `auth-service` ✅ STARTED (port 8001)

**Status:** Scaffold exists. Handlers not yet implemented.

**Controllers:** `UserController` (auth methods), `ApiKeyController`, `QuickConnectController`, `StartupController`

#### All Routes to Implement

```
# Anonymous
GET  /Users/Public                        → list non-hidden non-disabled users (Id, Name, PrimaryImageTag only)
POST /Users/AuthenticateByName            → body: {Name, Pw} → AuthenticationResult
POST /Users/{userId}/Authenticate         → query: pw → AuthenticationResult (legacy)
POST /Users/AuthenticateWithQuickConnect  → body: {Secret} → AuthenticationResult

# Authenticated
POST /Users/ForgotPassword                → body: {EnteredUsername} → ForgotPasswordResult
POST /Users/ForgotPasswordPin             → body: {EnteredPin} → PinRedeemResult

# Admin only
GET  /Auth/Keys                           → QueryResult<AuthenticationInfo>
POST /Auth/Keys                           → query: app=<name> → 204
DELETE /Auth/Keys/{key}                   → 204

# QuickConnect (anonymous to start, authenticated to authorize)
GET  /QuickConnect/Enabled                → bool (is QuickConnect enabled)
POST /QuickConnect/Initiate               → QuickConnectResult {Secret, Code, DeviceId}
POST /QuickConnect/Connect                → query: secret=<secret> → QuickConnectResult
POST /QuickConnect/Authorize              → query: code=<code> (authenticated) → 204

# Startup wizard (anonymous)
GET  /Startup/Configuration               → {HasCustomCertificate, SupportedAlgorithms}
POST /Startup/Complete                    → 204
GET  /Startup/RemoteAccess                → {EnableRemoteAccess, EnableUPnP}
POST /Startup/RemoteAccess                → body: {EnableRemoteAccess, EnableUPnP} → 204
GET  /Startup/User                        → first admin user
POST /Startup/User                        → body: {Name, Password} → 204
```

#### Critical Implementation Details

**AuthenticateByName handler:**
```go
// 1. Decode body {Name, Pw}
// 2. SELECT Id, Name, Password, IsDisabled, Policy FROM Users WHERE LOWER(Name)=LOWER(?)
// 3. If user.IsDisabled → 403 Forbidden "User is disabled"
// 4. bcrypt.CompareHashAndPassword(user.Password, []byte(req.Pw)) → 401 if mismatch
// 5. Generate token: hex(crypto/rand 20 bytes) = 40-char token
// 6. Parse X-Emby-Authorization for Client, Device, DeviceId, Version
// 7. INSERT INTO Devices (Id=uuid, UserId, DeviceId, AccessToken=token, FriendlyName=Device, AppName=Client, AppVersion=Version, Created=NOW())
//    ON DUPLICATE KEY UPDATE AccessToken=token, DateLastActivity=NOW()
// 8. Return AuthenticationResult{User: mapUserDto(user), AccessToken: token, ServerId: serverID}
```

**API Key creation:**
```go
// token = hex(crypto/rand 20 bytes)
// INSERT INTO ApiKeys (Id=uuid, AccessToken=token, Name=appParam, DateCreated=NOW(), IsAdmin=1)
// Return 204 (no body) — C# returns NoContent
```

**QuickConnect flow:**
```go
// Initiate: generate Secret=uuid, Code=6-digit random, store in memory/Redis with 15min TTL
// Connect: client polls with secret → return {Authenticated: true/false}
// Authorize: authenticated user approves code → mark as authorized
// FinalAuth: POST /Users/AuthenticateWithQuickConnect {Secret} → validate secret is authorized → create device token
```

#### DB Queries
```sql
-- GetUserByName
SELECT Id, Name, Password, IsDisabled, IsHidden, PrimaryImageTag, Policy
FROM Users WHERE LOWER(Name) = LOWER(?) LIMIT 1

-- GetPublicUsers
SELECT Id, Name, PrimaryImageTag FROM Users
WHERE IsHidden = 0 AND IsDisabled = 0 ORDER BY Name

-- CreateDevice
INSERT INTO Devices (Id, UserId, DeviceId, AccessToken, FriendlyName, AppName, AppVersion, Created, DateLastActivity)
VALUES (?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
ON DUPLICATE KEY UPDATE AccessToken = ?, DateLastActivity = NOW()

-- ListApiKeys
SELECT Id, AccessToken, Name, DateCreated FROM ApiKeys ORDER BY DateCreated DESC

-- DeleteApiKey
DELETE FROM ApiKeys WHERE Id = ? OR AccessToken = ?
```

---

### SERVICE 2: `user-service` 🔴 NEW (port 8002)

**Controllers:** `UserController` (CRUD methods), `UserLibraryController`, `UserViewsController`, `DisplayPreferencesController`

#### All Routes to Implement

```
# User CRUD
GET    /Users                              → [requires auth] list all users (admin: all; user: just self)
GET    /Users/{userId}                     → UserDto (self or admin)
POST   /Users/New                          → body: {Name, Password} → UserDto (admin only)
PUT    /Users/{userId}                     → body: UserDto updates (self or admin) → 204
DELETE /Users/{userId}                     → 204 (admin only)
POST   /Users/{userId}/Password            → body: {CurrentPw, NewPw, ResetPassword} → 204
POST   /Users/{userId}/EasyPassword        → body: {NewPassword} → 204
POST   /Users/{userId}/Policy              → body: UserPolicy → 204 (admin only)
POST   /Users/{userId}/Configuration       → body: UserConfiguration → 204

# User Library (items scoped to user's libraries)
GET    /Users/{userId}/Items               → QueryResult<BaseItemDto> (same params as /Items)
GET    /Users/{userId}/Items/{itemId}      → BaseItemDto
GET    /Users/{userId}/Items/Latest        → BaseItemDto[] (recently added)
GET    /Users/{userId}/Items/{itemId}/LocalTrailers    → BaseItemDto[]
GET    /Users/{userId}/Items/{itemId}/SpecialFeatures  → BaseItemDto[]
GET    /Users/{userId}/Items/{itemId}/AdditionalParts  → BaseItemDto[]
GET    /Users/{userId}/Items/{itemId}/ThemeVideos      → BaseItemDto[]
GET    /Users/{userId}/Items/{itemId}/ThemeSongs       → BaseItemDto[]
GET    /Users/{userId}/Items/{itemId}/Intros           → BaseItemDto[]

# User Views
GET    /Users/{userId}/Views               → QueryResult<BaseItemDto> (user's library views)
GET    /Users/{userId}/GroupingOptions     → VirtualFolderInfo[]

# Display Preferences
GET    /DisplayPreferences/{displayPreferencesId}      → query: userId, client → DisplayPreferences
POST   /DisplayPreferences/{displayPreferencesId}      → body: DisplayPreferences → 204

# Favorites / Played state (delegated to playback-service but routed here)
POST   /Users/{userId}/FavoriteItems/{itemId}          → UserItemDataDto
DELETE /Users/{userId}/FavoriteItems/{itemId}          → UserItemDataDto
POST   /Users/{userId}/PlayedItems/{itemId}            → query: datePlayed → UserItemDataDto
DELETE /Users/{userId}/PlayedItems/{itemId}            → UserItemDataDto
POST   /Users/{userId}/PlayingItems/{itemId}           → query: mediaSourceId, audioStreamIndex, subtitleStreamIndex → 204
DELETE /Users/{userId}/PlayingItems/{itemId}           → query: mediaSourceId, positionTicks → 204
POST   /Users/{userId}/PlayingItems/{itemId}/Progress  → query: positionTicks, ... → 204
```

#### Key Implementation Notes

- `/Users/{userId}/Items` is a passthrough to library-service with userId scoping
- `/Users/{userId}/Items/Latest` → `SELECT * FROM BaseItems bi LEFT JOIN UserData ud ON ud.ItemId=bi.Id AND ud.UserId=? WHERE bi.TopParentId IN (user's library root IDs) ORDER BY bi.DateCreated DESC LIMIT 20`
- `/Users/{userId}/Views` returns virtual library roots (CollectionType folders)
- `DisplayPreferences` stored in DB table `DisplayPreferences` (Id, UserId, Client, ItemDisplayPreferences JSON)

#### DB Queries
```sql
-- GetAllUsers (admin)
SELECT Id, Name, IsDisabled, IsHidden, PrimaryImageTag, Policy, Configuration
FROM Users ORDER BY Name

-- CreateUser
INSERT INTO Users (Id, Name, Password, IsDisabled, IsHidden, Configuration, Policy)
VALUES (?, ?, ?, 0, 0, ?, ?)

-- UpdateUserPassword
UPDATE Users SET Password = ? WHERE Id = ?

-- GetLatestItems (per-user library)
SELECT bi.*, ud.Played, ud.IsFavorite, ud.PlaybackPositionTicks
FROM BaseItems bi
LEFT JOIN UserData ud ON ud.ItemId = bi.Id AND ud.UserId = ?
WHERE bi.TopParentId IN (SELECT Id FROM BaseItems WHERE Type = 'UserRootFolder' AND ParentId IS NULL)
  AND bi.IsVirtualItem = 0
  AND bi.Type NOT IN ('UserRootFolder', 'CollectionFolder')
ORDER BY bi.DateCreated DESC
LIMIT ?

-- GetDisplayPreferences
SELECT * FROM DisplayPreferences WHERE Id = ? AND UserId = ? AND Client = ?

-- UpsertDisplayPreferences
INSERT INTO DisplayPreferences (Id, UserId, Client, Data) VALUES (?, ?, ?, ?)
ON DUPLICATE KEY UPDATE Data = ?
```

---

### SERVICE 3: `library-service` 🔴 NEW (port 8003)

**Controllers:** `ItemsController`, `LibraryController`, `LibraryStructureController`, `FilterController`, `YearsController`

This is the largest and most performance-critical service.

#### All Routes to Implement

```
# Core item querying (most complex endpoint in the entire system)
GET  /Items                                → QueryResult<BaseItemDto>
GET  /Items/{itemId}                       → BaseItemDto
DELETE /Items/{itemId}                     → 204 (admin only)
POST /Items/Delete                         → body: {Ids:[]} → 204 (admin only)
GET  /Items/{itemId}/Ancestors             → BaseItemDto[]
GET  /Items/{itemId}/PlaybackInfo          → PlaybackInfoResponse (delegate to media-service)
GET  /Items/{itemId}/Similar               → QueryResult<BaseItemDto>
GET  /Items/{itemId}/Reviews               → QueryResult<ItemReview>
GET  /Items/{itemId}/ThemeMedia            → AllThemeMediaResult
GET  /Items/{itemId}/ExternalIds           → ExternalIdInfo[]
GET  /Items/Counts                         → LibraryCounts (movie/show/song counts)

# Library management
GET  /Library/SelectableMediaFolders       → VirtualFolderInfo[] (admin)
GET  /Library/VirtualFolders               → VirtualFolderInfo[] (admin)
POST /Library/VirtualFolders               → body: AddVirtualFolderDto → 204 (admin)
DELETE /Library/VirtualFolders             → query: name → 204 (admin)
POST /Library/VirtualFolders/Paths         → body: MediaPathDto → 204 (admin)
DELETE /Library/VirtualFolders/Paths       → query: name, path → 204 (admin)
POST /Library/Refresh                      → 204 — trigger library scan
POST /Library/Media/Updated                → body: MediaUpdateInfoDto[] → 204
GET  /Library/FileSystem                   → body: FileSystemEntryInfo[]
GET  /Library/MediaFolders                 → QueryResult<BaseItemDto>
GET  /Library/PhysicalPaths               → string[]
GET  /Library/AvailableOptions             → LibraryOptionInfoDto

# Filters
GET  /Filters                              → query: userId, parentId, includeItemTypes[] → QueryFilters
GET  /Filters2                             → query: userId, parentId, includeItemTypes[] → QueryFiltersLegacy

# Years
GET  /Years                                → QueryResult<BaseItemDto>
GET  /Years/{year}                         → BaseItemDto
```

#### `/Items` Query Parameters (all optional)

```
userId, maxOfficialRating, hasThemeSong, hasThemeVideo, hasSubtitles,
hasSpecialFeature, hasTrailer, adjacentTo, indexNumber, parentIndexNumber,
hasParentalRating, isHd, is4K, locationTypes[], excludeLocationTypes[],
isMissing, isUnaired, minCommunityRating, minCriticRating,
minPremiereDate, minDateLastSaved, minDateLastSavedForUser,
maxPremiereDate, hasOverview, hasImdbId, hasTmdbId, hasTvdbId,
isMovie, isSeries, isNews, isKids, isSports,
excludeItemIds[], startIndex, limit, recursive, searchTerm,
sortOrder[], parentId, fields[], excludeItemTypes[], includeItemTypes[],
filters[], isFavorite, mediaTypes[], imageTypes[], sortBy[],
isPlayed, genres[], officialRatings[], tags[], years[],
enableUserData, imageTypeLimit, enableImageTypes[], person, personIds[],
personTypes[], studios[], artists[], excludeArtistIds[], artistIds[],
albumArtistIds[], contributingArtistIds[], albums[], albumIds[],
ids[], videoTypes[], minOfficialRating, isLocked, isPlaceHolder,
hasOfficialRating, collapseBoxSetItems, minWidth, minHeight,
maxWidth, maxHeight, is3D, seriesStatus[], nameStartsWithOrGreater,
nameStartsWith, nameLessThan, studioIds[], genreIds[],
enableTotalRecordCount=true, enableImages=true
```

#### Dynamic Query Builder (Critical)

Implement a SQL query builder in `internal/db/querybuilder.go`:

```go
type ItemQuery struct {
    UserId           string
    ParentId         string
    IncludeItemTypes []string    // filter to these BaseItem types
    ExcludeItemTypes []string
    Filters          []string    // IsPlayed, IsUnplayed, IsFavorite, etc.
    SortBy           []string    // SortName, DateCreated, CommunityRating, etc.
    SortOrder        []string    // Ascending, Descending
    StartIndex       int
    Limit            int
    Recursive        bool
    SearchTerm       string
    Genres           []string
    GenreIds         []string
    StudioIds        []string
    PersonIds        []string
    Years            []int
    IsPlayed         *bool
    IsFavorite       *bool
    IsFolder         *bool
    // ... all other filter fields
}

func BuildItemQuery(q ItemQuery) (sql string, args []interface{})
// Returns SELECT with proper JOINs, WHERE, ORDER BY, LIMIT/OFFSET
// Must LEFT JOIN UserData when userId provided
// Must JOIN ItemValues for genre/studio filtering
// Must use AncestorIds for recursive=true
```

#### DB Query Patterns
```sql
-- Base query structure
SELECT bi.Id, bi.Name, bi.Type, bi.SortName, bi.ParentId, bi.SeriesId,
       bi.ProductionYear, bi.PremiereDate, bi.CommunityRating, bi.OfficialRating,
       bi.RunTimeTicks, bi.IsFolder, bi.IsVirtualItem, bi.Overview, bi.ImageInfosJson,
       bi.DateCreated, bi.Genres, bi.Studios, bi.ProviderIds,
       ud.Played, ud.IsFavorite, ud.PlaybackPositionTicks, ud.PlayCount, ud.LastPlayedDate
FROM BaseItems bi
LEFT JOIN UserData ud ON ud.ItemId = bi.Id AND ud.UserId = ?
WHERE bi.IsVirtualItem = 0
  AND bi.Type IN (?)           -- dynamic based on includeItemTypes
  AND bi.ParentId = ?          -- if not recursive
ORDER BY bi.SortName ASC       -- dynamic
LIMIT ? OFFSET ?

-- Recursive (uses AncestorIds)
JOIN AncestorIds ai ON ai.ItemId = bi.Id AND ai.ParentItemId = ?

-- Genre filter
JOIN ItemValuesMap ivm ON ivm.ItemId = bi.Id
JOIN ItemValues iv ON iv.Id = ivm.ItemValueId AND iv.Type = 3 AND LOWER(iv.CleanValue) IN (?)

-- Full-text search
WHERE MATCH(bi.Name, bi.OriginalTitle) AGAINST (? IN BOOLEAN MODE)

-- Count query (same WHERE, no LIMIT/ORDER)
SELECT COUNT(*) FROM BaseItems bi [same JOINs] WHERE [same conditions]
```

---

### SERVICE 4: `browse-service` 🔴 NEW (port 8004)

**Controllers:** `ArtistsController`, `GenresController`, `MusicGenresController`, `StudiosController`, `PersonsController`, `MoviesController`, `TvShowsController`, `TrailersController`, `ChannelsController`, `InstantMixController`

#### All Routes to Implement

```
# Artists
GET  /Artists                              → QueryResult<BaseItemDto>
GET  /Artists/AlbumArtists                 → QueryResult<BaseItemDto>
GET  /Artists/{name}                       → BaseItemDto (query: userId)

# Genres
GET  /Genres                               → QueryResult<BaseItemDto>
GET  /Genres/{genreName}                   → BaseItemDto (query: userId)
GET  /MusicGenres                          → QueryResult<BaseItemDto>
GET  /MusicGenres/{genreName}              → BaseItemDto (query: userId)

# Studios
GET  /Studios                              → QueryResult<BaseItemDto>
GET  /Studios/{name}                       → BaseItemDto (query: userId)

# Persons (cast/crew)
GET  /Persons                              → QueryResult<BaseItemDto>
GET  /Persons/{name}                       → BaseItemDto (query: userId)

# Movies
GET  /Movies/Recommendations               → RecommendationDto[] (query: userId, categoryLimit, itemLimit)

# TV Shows
GET  /Shows/NextUp                         → QueryResult<BaseItemDto> (query: userId, limit, fields[], ...)
GET  /Shows/Upcoming                       → QueryResult<BaseItemDto>
GET  /Shows/{seriesId}/Seasons             → QueryResult<BaseItemDto>
GET  /Shows/{seriesId}/Episodes            → QueryResult<BaseItemDto> (query: seasonId, ...)
GET  /Shows/{seriesId}/Similar             → QueryResult<BaseItemDto>

# Trailers
GET  /Trailers                             → QueryResult<BaseItemDto>

# Channels
GET  /Channels                             → QueryResult<BaseItemDto>
GET  /Channels/Features                    → ChannelFeatures[]
GET  /Channels/{channelId}/Features        → ChannelFeatures
GET  /Channels/{channelId}/Items           → QueryResult<BaseItemDto>
GET  /Channels/Items/Latest                → QueryResult<BaseItemDto>

# Instant Mix
GET  /Songs/{id}/InstantMix                → QueryResult<BaseItemDto>
GET  /Albums/{id}/InstantMix               → QueryResult<BaseItemDto>
GET  /Artists/{id}/InstantMix             → QueryResult<BaseItemDto>
GET  /MusicGenres/{name}/InstantMix        → QueryResult<BaseItemDto>
GET  /Playlists/{id}/InstantMix            → QueryResult<BaseItemDto>
GET  /Items/{id}/InstantMix                → QueryResult<BaseItemDto>
```

#### Key Implementation Notes

- Artists/Genres/Studios/Persons all query `ItemValues` table by type
- `GET /Shows/NextUp` → last watched episode per series + next episode: complex query
- Instant Mix → random items from same genre/artist pool

```sql
-- NextUp: find next episode after last watched
SELECT ep.*, ud.Played, ud.LastPlayedDate
FROM BaseItems ep
WHERE ep.Type = 'Episode'
  AND ep.SeriesId IN (
    SELECT DISTINCT bi.SeriesId FROM BaseItems bi
    JOIN UserData ud ON ud.ItemId = bi.Id AND ud.UserId = ?
    WHERE bi.Type = 'Episode' AND ud.Played = 1
  )
  AND ep.Id NOT IN (SELECT ItemId FROM UserData WHERE UserId = ? AND Played = 1)
ORDER BY ep.SeriesId, ep.ParentIndexNumber, ep.IndexNumber
-- Then group by SeriesId and take first unwatched per series
```

---

### SERVICE 5: `playback-service` 🔴 NEW (port 8005)

**Controllers:** `DynamicHlsController`, `HlsSegmentController`, `AudioController`, `UniversalAudioController`, `VideosController`, `PlaystateController`, `MediaInfoController`, `SyncPlayController`, `TimeSyncController`, `SessionController`

#### All Routes to Implement

```
# Playstate
POST   /Users/{userId}/PlayedItems/{itemId}            → query: datePlayed → UserItemDataDto
DELETE /Users/{userId}/PlayedItems/{itemId}            → UserItemDataDto
POST   /Users/{userId}/FavoriteItems/{itemId}          → UserItemDataDto
DELETE /Users/{userId}/FavoriteItems/{itemId}          → UserItemDataDto
POST   /Users/{userId}/PlayingItems/{itemId}           → 204
DELETE /Users/{userId}/PlayingItems/{itemId}           → 204
POST   /Users/{userId}/PlayingItems/{itemId}/Progress  → 204
POST   /PlaystateCommands/{command}                    → 204 (send playstate command to client)

# HLS Video Streaming
GET  /Videos/{itemId}/live.m3u8                        → text/plain (HLS master playlist)
GET  /Videos/{itemId}/master.m3u8                      → text/plain
GET  /Videos/{itemId}/main.m3u8                        → text/plain
GET  /Videos/{itemId}/hls1/{segmentId}/stream.ts       → video/mp2t (HLS segment)
GET  /Videos/{itemId}/hls1/{segmentId}/stream.aac      → audio/aac
GET  /Videos/{itemId}/hls1/{segmentId}/stream.mp3      → audio/mpeg
DELETE /Videos/ActiveEncodings                         → query: deviceId, playSessionId → 204

# Direct Video Streaming
GET  /Videos/{itemId}/stream                           → video/* stream
GET  /Videos/{itemId}/stream.{container}               → video/* stream
HEAD /Videos/{itemId}/stream                           → headers only
HEAD /Videos/{itemId}/stream.{container}               → headers only

# Audio Streaming
GET  /Audio/{itemId}/stream                            → audio/* stream
GET  /Audio/{itemId}/stream.{container}                → audio/* stream
HEAD /Audio/{itemId}/stream                            → headers only
GET  /Audio/{itemId}/universal                         → best format selection
HEAD /Audio/{itemId}/universal                         → headers only

# Media Info
GET  /Items/{itemId}/PlaybackInfo                      → PlaybackInfoResponse
POST /Items/{itemId}/PlaybackInfo                      → body: PlaybackInfoRequestDto → PlaybackInfoResponse
GET  /Videos/{itemId}/AdditionalParts                  → BaseItemDto[]
GET  /Items/{itemId}/DownloadToken                     → {Token}
POST /MediaInfo/LiveStreamOpen                         → LiveStreamResponse
POST /MediaInfo/LiveStreamClose                        → query: liveStreamId → 204

# Sessions
GET    /Sessions                           → SessionInfoDto[]
POST   /Sessions/Logout                    → 204
POST   /Sessions/{sessionId}/Playing       → body: PlayRequest → 204
POST   /Sessions/{sessionId}/Playing/Stop  → 204
POST   /Sessions/{sessionId}/Playing/Pause → 204
POST   /Sessions/{sessionId}/Playing/Unpause → 204
POST   /Sessions/{sessionId}/Playing/NextTrack → 204
POST   /Sessions/{sessionId}/Playing/PreviousTrack → 204
POST   /Sessions/{sessionId}/Playing/Seek  → query: seekPositionTicks → 204
POST   /Sessions/{sessionId}/Viewing       → body: {ItemType, ItemId, ItemName} → 204
POST   /Sessions/{sessionId}/Message       → body: {Header, Text, TimeoutMs} → 204
POST   /Sessions/{sessionId}/Browse        → body: {ItemType, ItemId, ItemName, Context} → 204
POST   /Sessions/{sessionId}/Users/{userId} → 204 (add user to session)
DELETE /Sessions/{sessionId}/Users/{userId} → 204 (remove user from session)
POST   /Sessions/{sessionId}/Capabilities  → body: ClientCapabilitiesDto → 204
POST   /Sessions/Capabilities              → body: ClientCapabilitiesDto → 204
POST   /Sessions/Capabilities/Full         → body: ClientCapabilitiesDto → 204

# SyncPlay
POST /SyncPlay/New                         → body: NewGroupRequestDto → 204
POST /SyncPlay/Join                        → body: JoinGroupRequestDto → 204
POST /SyncPlay/Leave                       → 204
GET  /SyncPlay/List                        → GroupInfoDto[]
POST /SyncPlay/MovePlaylistItem            → body: MovePlaylistItemRequestDto → 204
POST /SyncPlay/NextItem                    → body: NextItemRequestDto → 204
POST /SyncPlay/Pause                       → body: PauseRequestDto → 204
POST /SyncPlay/Ping                        → body: PingRequestDto → 204
POST /SyncPlay/PreviousItem                → body: PreviousItemRequestDto → 204
POST /SyncPlay/Queue                       → body: QueueRequestDto → 204
POST /SyncPlay/Ready                       → body: ReadyRequestDto → 204
POST /SyncPlay/RemoveFromPlaylist          → body: RemoveFromPlaylistRequestDto → 204
POST /SyncPlay/Seek                        → body: SeekRequestDto → 204
POST /SyncPlay/SetIgnoreWait              → body: IgnoreWaitRequestDto → 204
POST /SyncPlay/SetNewQueue                 → body: PlayRequestDto → 204
POST /SyncPlay/SetPlaylistItem             → body: SetPlaylistItemRequestDto → 204
POST /SyncPlay/SetRepeatMode              → body: SetRepeatModeRequestDto → 204
POST /SyncPlay/SetShuffleMode             → body: SetShuffleModeRequestDto → 204
POST /SyncPlay/Stop                        → body: StopRequestDto → 204
POST /SyncPlay/Unpause                     → body: UnpauseRequestDto → 204

# Time Sync
GET  /GetUtcTime                           → UtcTimeResponse {RequestReceptionTime, ResponseTransmissionTime}
```

#### HLS Implementation (Critical)

```go
// internal/ffmpeg/hls.go

type TranscodeJob struct {
    ItemId        string
    PlaySessionId string
    MediaSourceId string
    AudioCodec    string
    VideoCodec    string
    VideoBitRate  int
    AudioChannels int
    Width         int
    Height        int
    StartTimeTicks int64
    SegmentLength int  // default 3
    OutputDir     string
}

// BuildFFmpegArgs generates exact CLI args matching C# output
func BuildFFmpegArgs(job TranscodeJob, filePath string) []string {
    args := []string{
        "-i", filePath,
        "-map_metadata", "-1",
        "-map_chapters", "-1",
        "-threads", "0",
        "-codec:v:0", job.VideoCodec,
        "-b:v", fmt.Sprintf("%d", job.VideoBitRate),
        "-maxrate", fmt.Sprintf("%d", int(float64(job.VideoBitRate)*1.5)),
        "-bufsize", fmt.Sprintf("%d", job.VideoBitRate*2),
    }
    // ... add audio args, subtitle args, segment args
    // -f hls -hls_time 3 -hls_list_size 0 -hls_segment_filename "%s%s_%d.ts"
    return args
}

// GenerateMasterPlaylist creates the .m3u8 master playlist
func GenerateMasterPlaylist(item *BaseItem, mediaSources []MediaSource, params HlsParams) string
```

#### Sessions (in-memory + Redis)

Sessions are NOT stored in MySQL. Store in Redis:
```
Key: session:{sessionId}
Value: JSON SessionInfoDto
TTL: 1 hour (refreshed on activity)
```

```sql
-- UserData: mark played
INSERT INTO UserData (UserId, ItemId, Played, PlayCount, LastPlayedDate)
VALUES (?, ?, 1, COALESCE((SELECT PlayCount FROM UserData WHERE UserId=? AND ItemId=?), 0)+1, ?)
ON DUPLICATE KEY UPDATE
  Played = 1,
  PlayCount = PlayCount + 1,
  LastPlayedDate = ?

-- UserData: save position
INSERT INTO UserData (UserId, ItemId, PlaybackPositionTicks)
VALUES (?, ?, ?)
ON DUPLICATE KEY UPDATE PlaybackPositionTicks = ?

-- UserData: toggle favorite
INSERT INTO UserData (UserId, ItemId, IsFavorite) VALUES (?, ?, ?)
ON DUPLICATE KEY UPDATE IsFavorite = ?
```

---

### SERVICE 6: `media-service` 🔴 NEW (port 8006)

**Controllers:** `ImageController`, `RemoteImageController`, `SubtitleController`, `LyricsController`, `MediaSegmentsController`, `VideoAttachmentsController`, `TrickplayController`

#### All Routes to Implement

```
# Images
GET  /Items/{itemId}/Images                            → ImageInfo[]
GET  /Items/{itemId}/Images/{imageType}                → image/* binary (primary)
GET  /Items/{itemId}/Images/{imageType}/{imageIndex}   → image/* binary
HEAD /Items/{itemId}/Images/{imageType}                → headers only (ETag, Last-Modified)
HEAD /Items/{itemId}/Images/{imageType}/{imageIndex}   → headers only
DELETE /Items/{itemId}/Images/{imageType}              → 204
DELETE /Items/{itemId}/Images/{imageType}/{imageIndex} → 204
POST /Items/{itemId}/Images/{imageType}/{imageIndex}   → upload image → 204
POST /Items/{itemId}/Images/{imageType}                → upload image → 204
POST /Items/{itemId}/Images/{imageType}/{imageIndex}/Index → body: {NewIndex} → 204

# User images
GET  /Users/{userId}/Images/{imageType}                → image/* binary
HEAD /Users/{userId}/Images/{imageType}                → headers only
DELETE /Users/{userId}/Images/{imageType}              → 204
POST /Users/{userId}/Images/{imageType}                → upload → 204

# Remote images
GET  /Items/{itemId}/RemoteImages                      → query: type, includeAllLanguages → RemoteImageResult
GET  /Items/{itemId}/RemoteImages/Providers            → ImageProviderInfo[]
POST /Items/{itemId}/RemoteImages/Download             → query: type, imageUrl → 204

# Subtitles
GET  /Videos/{itemId}/Subtitles/{index}/Stream.{format}  → subtitle file (SRT, VTT, etc.)
GET  /Videos/{itemId}/Subtitles/{index}/{partialId}/Stream.{format} → subtitle file
DELETE /Videos/{itemId}/Subtitles/{index}              → 204
POST /Videos/{itemId}/Subtitles/{index}                → search/download → SubtitleInfo[]
GET  /Providers/Subtitles/Subtitles                    → subtitle provider search
POST /Videos/{itemId}/RemoteSubtitles/{subtitleId}     → download remote subtitle → 204
GET  /Videos/{itemId}/Subtitles/{index}/Search         → RemoteSubtitleInfo[]
POST /Videos/SubtitleSegments                          → upload subtitle → 204

# Lyrics
GET  /Audio/{itemId}/Lyrics                            → LyricDto
POST /Audio/{itemId}/Lyrics                            → upload → LyricDto
DELETE /Audio/{itemId}/Lyrics                          → 204
GET  /Providers/Lyrics                                 → LyricProviderInfo[]
GET  /Audio/{itemId}/RemoteLyrics/{lyricId}            → LyricDto (preview)
POST /Audio/{itemId}/RemoteLyrics/{lyricId}            → download → 204

# Media Segments
GET  /MediaSegments/{itemId}                           → query: includeSegmentTypes[] → MediaSegmentDto[]
POST /MediaSegments/{itemId}                           → body: MediaSegmentDto[] → 204 (admin)
DELETE /MediaSegments/{itemId}/{segmentId}             → 204 (admin)

# Trickplay
GET  /Videos/{itemId}/Trickplay/{width}/GetBIF         → application/x-bif binary
GET  /Videos/{itemId}/Trickplay/{width}/{index}/GetBIF → application/x-bif binary

# Video Attachments
GET  /Videos/{videoId}/Attachments/{mediaSourceId}/{index}  → attachment binary
```

#### Image Serving Implementation

```go
// internal/handlers/image.go

func ServeImage(w http.ResponseWriter, r *http.Request) {
    itemId := chi.URLParam(r, "itemId")
    imageType := chi.URLParam(r, "imageType")

    // 1. Look up image info from BaseItems.ImageInfosJson
    imageInfo, err := db.GetImageInfo(itemId, imageType, imageIndex)

    // 2. Check ETag (If-None-Match header) → 304 Not Modified if matches
    if r.Header.Get("If-None-Match") == imageInfo.ImageTag {
        w.WriteHeader(http.StatusNotModified)
        return
    }

    // 3. Parse resize params: fillHeight, fillWidth, quality, tag
    fillHeight := queryInt(r, "fillHeight", 0)
    fillWidth  := queryInt(r, "fillWidth", 0)
    quality    := queryInt(r, "quality", 90)

    // 4. Check disk cache: /cache/images/{itemId}_{type}_{fillH}x{fillW}_q{quality}.jpg
    cachePath := buildCachePath(itemId, imageType, fillHeight, fillWidth, quality)
    if _, err := os.Stat(cachePath); err == nil {
        serveFile(w, r, cachePath)
        return
    }

    // 5. Load original from disk, resize with imaging.Fit()
    img := imaging.Open(imageInfo.Path)
    if fillHeight > 0 || fillWidth > 0 {
        img = imaging.Fit(img, fillWidth, fillHeight, imaging.Lanczos)
    }

    // 6. Encode to JPEG at quality, write to cache
    imaging.Save(img, cachePath, imaging.JPEGQuality(quality))

    // 7. Set headers: Content-Type, ETag, Cache-Control: max-age=31536000
    w.Header().Set("ETag", imageInfo.ImageTag)
    w.Header().Set("Cache-Control", "max-age=31536000")
    serveFile(w, r, cachePath)
}
```

---

### SERVICE 7: `metadata-service` 🔴 NEW (port 8007)

**Controllers:** `ItemRefreshController`, `ItemUpdateController`, `ItemLookupController`, `PackageController`, `PluginsController`, `ScheduledTasksController`

#### All Routes to Implement

```
# Item Refresh
POST /Items/{itemId}/Refresh               → query: metadataRefreshMode, imageRefreshMode, replaceAllMetadata, replaceAllImages → 204

# Item Update
GET  /Items/{itemId}/MetadataEditor        → MetadataEditorInfo
POST /Items/{itemId}                       → body: BaseItemDto updates → 204

# Item Lookup (remote search)
POST /Items/RemoteSearch/Movie             → body: MovieInfoRemoteSearchQuery → RemoteSearchResult[]
POST /Items/RemoteSearch/Trailer           → body: TrailerInfoRemoteSearchQuery → RemoteSearchResult[]
POST /Items/RemoteSearch/Series            → body: SeriesInfoRemoteSearchQuery → RemoteSearchResult[]
POST /Items/RemoteSearch/Music             → body: AlbumInfoRemoteSearchQuery → RemoteSearchResult[]
POST /Items/RemoteSearch/MusicArtist       → body: ArtistInfoRemoteSearchQuery → RemoteSearchResult[]
POST /Items/RemoteSearch/MusicAlbum        → body: AlbumInfoRemoteSearchQuery → RemoteSearchResult[]
POST /Items/RemoteSearch/Person            → body: PersonLookupInfoRemoteSearchQuery → RemoteSearchResult[]
POST /Items/RemoteSearch/Book              → body: BookInfoRemoteSearchQuery → RemoteSearchResult[]
POST /Items/RemoteSearch/Apply/{itemId}    → body: RemoteSearchResult → 204

# Packages / Plugins
GET  /Packages                             → PackageInfo[]
GET  /Packages/{name}                      → PackageInfo
POST /Packages/Installed/{name}            → query: assemblyGuid, version → 204 (admin)
DELETE /Packages/Installed/{packageName}   → 204 (admin)
GET  /Plugins                              → PluginInfo[]
GET  /Plugins/{pluginId}/Configuration     → plugin config object
POST /Plugins/{pluginId}/Configuration     → body: config → 204
DELETE /Plugins/{pluginId}                 → 204 (admin)
GET  /Plugins/{pluginId}/Manifest          → PluginManifest

# Scheduled Tasks
GET  /ScheduledTasks                       → query: isHidden?, isEnabled? → TaskInfo[]
GET  /ScheduledTasks/{taskId}              → TaskInfo
POST /ScheduledTasks/Running/{taskId}      → 204 (start task)
DELETE /ScheduledTasks/Running/{taskId}    → 204 (stop task)
POST /ScheduledTasks/{taskId}/Triggers     → body: TaskTriggerInfo[] → 204 (admin)
```

#### Scheduled Tasks Implementation

```go
// internal/tasks/manager.go

type Task struct {
    Id          string
    Name        string
    Description string
    Category    string
    IsHidden    bool
    State       string   // "Idle", "Running", "Cancelling"
    CurrentProgress *TaskProgress
    LastExecutionResult *TaskResult
    Triggers    []TaskTriggerInfo
}

// Built-in tasks to register:
// - "RefreshGuide"         — Live TV guide refresh
// - "CleanupMedia"         — Remove missing media from library
// - "RefreshPeople"        — Re-scan people/cast images
// - "CleanupImages"        — Remove orphaned image cache
// - "OptimizeDatabase"     — OPTIMIZE TABLE / ANALYZE
// - "RefreshLibrary"       — Scan all library folders for new media

// Task progress broadcast via WebSocket (ScheduledTasksInfoStart subscription)
```

---

### SERVICE 8: `collection-service` ✅ STARTED (port 8008)

**Controllers:** `CollectionController`

**Verify implementation includes:**
```
POST   /Collections                          → query: name, ids[], parentId, isLocked → CollectionCreationResult{Id}
POST   /Collections/{collectionId}/Items     → query: ids[] (comma-separated) → 204
DELETE /Collections/{collectionId}/Items     → query: ids[] → 204
```

Wire format:
- `CollectionCreationResult: {"Id": "uuid"}`
- POST requires policy `CollectionManagement`

---

### SERVICE 9: `playlist-service` ✅ STARTED (port 8009)

**Controllers:** `PlaylistsController`

**Verify implementation includes:**
```
POST   /Playlists                            → body: CreatePlaylistDto → PlaylistCreationResult
GET    /Playlists/{playlistId}/Items         → query: userId, startIndex, limit, fields[], enableImages, enableUserData → QueryResult<BaseItemDto>
POST   /Playlists/{playlistId}/Items         → query: ids[], userId → 204
DELETE /Playlists/{playlistId}/Items         → query: entryIds[] → 204
POST   /Playlists/{playlistId}/Items/{itemId}/Move/{newIndex} → 204
GET    /Playlists/{playlistId}/Instantmix    → QueryResult<BaseItemDto>
GET    /Playlists/{playlistId}               → PlaylistDto (BaseItemDto variant)
DELETE /Playlists/{playlistId}               → 204
```

---

### SERVICE 10: `admin-service` 🔴 NEW (port 8010)

**Controllers:** `SystemController`, `ConfigurationController`, `DashboardController`, `BrandingController`, `EnvironmentController`, `LocalizationController`, `ActivityLogController`, `ClientLogController`, `BackupController`, `DevicesController`

#### All Routes to Implement

```
# System
GET  /System/Info                          → SystemInfo (authenticated)
GET  /System/Info/Public                   → PublicSystemInfo (anonymous)
GET  /System/Info/Storage                  → SystemStorageDto (admin)
GET  /System/Ping                          → "Jellyfin Server"
POST /System/Ping                          → "Jellyfin Server"
POST /System/Restart                       → 204 (local or admin)
POST /System/Shutdown                      → 204 (admin only)
GET  /System/Logs                          → LogFile[] (admin)
GET  /System/Logs/Log                      → query: name → text/plain log content (admin)
GET  /System/Endpoint                      → EndPointInfo
GET  /System/WakeOnLanInfo                 → WakeOnLanInfo[]

# Configuration
GET  /System/Configuration                 → ServerConfiguration JSON object
POST /System/Configuration                 → body: ServerConfiguration → 204 (admin)
GET  /System/Configuration/{key}           → named config section JSON
POST /System/Configuration/{key}           → body: JSON → 204 (admin)
GET  /System/Configuration/MetadataOptions/Default → MetadataOptions (admin)
POST /System/Configuration/Branding        → body: BrandingOptionsDto → 204 (admin)

# Dashboard
GET  /Dashboard/ConfigurationPages         → ConfigurationPageInfo[]
GET  /Dashboard/ConfigurationPage          → query: name → config page content

# Branding (anonymous)
GET  /Branding/Configuration               → BrandingOptionsDto
GET  /Branding/Css                         → text/css
GET  /Branding/Css.css                     → text/css

# Environment
GET  /Environment/DefaultDirectoryBrowser  → DefaultDirectoryBrowserInfo (admin)
GET  /Environment/Drives                   → FileSystemEntryInfo[] (admin)
GET  /Environment/NetworkDevices           → FileSystemEntryInfo[] (admin)
GET  /Environment/NetworkShares            → FileSystemEntryInfo[] (admin)
GET  /Environment/ParentPath               → query: path → string (admin)
GET  /Environment/DirectoryContents        → query: path, includeFiles, includeDirectories → FileSystemEntryInfo[] (admin)

# Localization
GET  /Localization/Cultures                → CultureDto[]
GET  /Localization/Countries               → CountryInfo[]
GET  /Localization/Options                 → LocalizationOption[]
GET  /Localization/ParentalRatings         → ParentalRating[]

# Activity Log
GET  /System/ActivityLog/Entries           → QueryResult<ActivityLogEntry>
  query: startIndex, limit, minDate, hasUserId

# Client Log
POST /Log/ClientEntries                    → body: ClientLogDocumentDto → ClientLogDocumentResponseDto
  (text/plain body, max 1MB, requires auth)

# Backup
GET  /System/Backups                       → BackupManifestDto[] (admin)
GET  /System/Backups/Manifest              → query: path → BackupManifestDto (admin)
POST /System/Backups/Create               → body: BackupOptionsDto → BackupManifestDto (admin)
POST /System/Backups/Restore              → body: BackupRestoreRequestDto → 204 (admin)

# Devices
GET  /Devices                              → query: userId? → DeviceInfo[] (admin)
GET  /Devices/Info                         → query: id → DeviceInfo (admin)
GET  /Devices/Options                      → query: id → DeviceOptions (admin)
POST /Devices/Options                      → query: id, body: DeviceOptions → 204 (admin)
DELETE /Devices                            → query: id → 204 (admin)
```

#### Key DTOs

```go
type PublicSystemInfo struct {
    LocalAddress         string `json:"LocalAddress"`
    ServerName           string `json:"ServerName"`
    Version              string `json:"Version"`     // "10.10.0"
    ProductName          string `json:"ProductName"` // "Jellyfin Server"
    OperatingSystem      string `json:"OperatingSystem"`
    Id                   string `json:"Id"`
    StartupWizardCompleted bool  `json:"StartupWizardCompleted"`
}

type SystemInfo struct {
    PublicSystemInfo               // embed all public fields
    LogPath                string  `json:"LogPath"`
    ItemsByNamePath         string  `json:"ItemsByNamePath"`
    CachePath               string  `json:"CachePath"`
    InternalMetadataPath    string  `json:"InternalMetadataPath"`
    TranscodingTempPath     string  `json:"TranscodingTempPath"`
    HardwareAccelerationRequiresRestart bool `json:"HardwareAccelerationRequiresRestart"`
    EncoderLocationType     string  `json:"EncoderLocationType"`
    SystemArchitecture      string  `json:"SystemArchitecture"`
    PackageName             string  `json:"PackageName"`
    CanSelfRestart          bool    `json:"CanSelfRestart"`
    CanLaunchWebBrowser     bool    `json:"CanLaunchWebBrowser"`
    HasUpdateAvailable      bool    `json:"HasUpdateAvailable"`
    SupportsLibraryMonitor  bool    `json:"SupportsLibraryMonitor"`
    CompletedInstallations  []InstallationInfo `json:"CompletedInstallations"`
}

type BrandingOptionsDto struct {
    LoginDisclaimer      string `json:"LoginDisclaimer"`
    CustomCss            string `json:"CustomCss"`
    SplashscreenEnabled  bool   `json:"SplashscreenEnabled"`
}
```

---

### SERVICE 11: `search-service` 🔴 NEW (port 8011)

**Controllers:** `SearchController`, `SuggestionsController`

#### All Routes to Implement

```
# Search
GET /Search/Hints                          → SearchHintResult
  query: startIndex, limit, userId, searchTerm (required),
         includeItemTypes[], excludeItemTypes[], mediaTypes[],
         parentId, isMovie, isSeries, isNews, isKids, isSports,
         includePeople, includeMedia, includeGenres, includeStudios, includeArtists

# Suggestions
GET /Items/{userId}/Suggestions            → QueryResult<BaseItemDto>
  query: mediaType[], type[], startIndex, limit
```

#### SearchHintResult Structure

```go
type SearchHintResult struct {
    SearchHints      []SearchHint `json:"SearchHints"`
    TotalRecordCount int          `json:"TotalRecordCount"`
}

type SearchHint struct {
    ItemId           string `json:"ItemId"`
    Id               string `json:"Id"`
    Name             string `json:"Name"`
    MatchedTerm      string `json:"MatchedTerm,omitempty"`
    IndexNumber      *int   `json:"IndexNumber,omitempty"`
    ProductionYear   *int   `json:"ProductionYear,omitempty"`
    ParentIndexNumber *int  `json:"ParentIndexNumber,omitempty"`
    PrimaryImageTag  string `json:"PrimaryImageTag,omitempty"`
    ThumbImageTag    string `json:"ThumbImageTag,omitempty"`
    ThumbImageItemId string `json:"ThumbImageItemId,omitempty"`
    BackdropImageTag string  `json:"BackdropImageTag,omitempty"`
    BackdropImageItemId string `json:"BackdropImageItemId,omitempty"`
    Type             string `json:"Type"`
    MediaType        string `json:"MediaType,omitempty"`
    Series           string `json:"Series,omitempty"`
    Status           string `json:"Status,omitempty"`
    Album            string `json:"Album,omitempty"`
    AlbumId          string `json:"AlbumId,omitempty"`
    AlbumArtist      string `json:"AlbumArtist,omitempty"`
    Artists          []string `json:"Artists,omitempty"`
    SongCount        *int   `json:"SongCount,omitempty"`
    EpisodeCount     *int   `json:"EpisodeCount,omitempty"`
    IsFolder         bool   `json:"IsFolder"`
    RunTimeTicks     int64  `json:"RunTimeTicks,omitempty"`
    ChannelId        string `json:"ChannelId,omitempty"`
    ChannelName      string `json:"ChannelName,omitempty"`
}
```

#### Search Implementation

```sql
-- Full-text search (MySQL FULLTEXT)
SELECT Id, Name, Type, ProductionYear, SortName, ImageInfosJson,
       ParentId, SeriesId, IndexNumber, ParentIndexNumber
FROM BaseItems
WHERE IsVirtualItem = 0
  AND (MATCH(Name, OriginalTitle) AGAINST (? IN BOOLEAN MODE)
       OR Name LIKE CONCAT('%', ?, '%'))
  AND Type IN (?)  -- filtered by includeItemTypes
LIMIT ?

-- Suggestions: recently added + resume-worthy
SELECT bi.*, ud.PlaybackPositionTicks
FROM BaseItems bi
JOIN UserData ud ON ud.ItemId = bi.Id AND ud.UserId = ?
WHERE ud.PlaybackPositionTicks > 0 AND ud.Played = 0
  AND bi.Type IN ('Movie', 'Episode')
ORDER BY ud.LastPlayedDate DESC
LIMIT 10
```

---

### SERVICE 12: `livetv-service` 🔴 TODO — Phase 2 (port 8012)

**Controllers:** `LiveTvController`, `ChannelsController`

**Defer to Phase 2.** Complexity is extremely high (hardware tuner integration, EPG, HDHomeRun protocol). During Phase 1, keep C# live TV service running alongside Go services. The nginx gateway should route `/LiveTV/*` and `/Channels/*` to C# during Phase 1.

---

## 8. NGINX GATEWAY (Complete Routing Table)

File: `Kabletown/nginx.conf` — replace existing partial with this complete version:

```nginx
upstream auth_service    { server auth-service:8001;     keepalive 32; }
upstream user_service    { server user-service:8002;     keepalive 32; }
upstream library_service { server library-service:8003;  keepalive 64; }
upstream browse_service  { server browse-service:8004;   keepalive 32; }
upstream playback_service { server playback-service:8005; keepalive 64; }
upstream media_service   { server media-service:8006;    keepalive 64; }
upstream metadata_service { server metadata-service:8007; keepalive 16; }
upstream collection_service { server collection-service:8008; keepalive 16; }
upstream playlist_service { server playlist-service:8009; keepalive 16; }
upstream admin_service   { server admin-service:8010;    keepalive 16; }
upstream search_service  { server search-service:8011;   keepalive 16; }

map $http_upgrade $connection_upgrade {
    default upgrade;
    '' close;
}

server {
    listen 8096;
    server_name _;

    # Required headers on all responses
    add_header X-Application-Version "10.10.0" always;

    # Static web client
    location /web {
        root /usr/share/jellyfin-web;
        try_files $uri $uri/ /web/index.html;
        add_header Cache-Control "no-cache";
    }

    # WebSocket (session-service)
    location /socket {
        proxy_pass http://playback_service;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection $connection_upgrade;
        proxy_set_header Host $host;
        proxy_read_timeout 3600;
    }

    # === AUTH SERVICE ===
    location ~ ^/Users/Public$                      { proxy_pass http://auth_service; }
    location ~ ^/Users/AuthenticateByName$          { proxy_pass http://auth_service; }
    location ~ ^/Users/AuthenticateWithQuickConnect$ { proxy_pass http://auth_service; }
    location ~ ^/Users/ForgotPassword               { proxy_pass http://auth_service; }
    location ~ ^/Users/[^/]+/Authenticate$          { proxy_pass http://auth_service; }
    location ~ ^/Auth/Keys                          { proxy_pass http://auth_service; }
    location ~ ^/QuickConnect                       { proxy_pass http://auth_service; }
    location ~ ^/Startup                            { proxy_pass http://auth_service; }

    # === USER SERVICE ===
    location ~ ^/Users/New$                         { proxy_pass http://user_service; }
    location ~ ^/Users/[^/]+/Password$              { proxy_pass http://user_service; }
    location ~ ^/Users/[^/]+/EasyPassword$          { proxy_pass http://user_service; }
    location ~ ^/Users/[^/]+/Policy$                { proxy_pass http://user_service; }
    location ~ ^/Users/[^/]+/Configuration$         { proxy_pass http://user_service; }
    location ~ ^/Users/[^/]+/Views$                 { proxy_pass http://user_service; }
    location ~ ^/Users/[^/]+/GroupingOptions$       { proxy_pass http://user_service; }
    location ~ ^/Users/[^/]+/Items/Latest$          { proxy_pass http://user_service; }
    location ~ ^/Users/[^/]+/Items/[^/]+/(LocalTrailers|SpecialFeatures|AdditionalParts|ThemeVideos|ThemeSongs|Intros)$ { proxy_pass http://user_service; }
    location ~ ^/Users/[^/]+/Items                  { proxy_pass http://user_service; }
    location ~ ^/Users                              { proxy_pass http://user_service; }
    location ~ ^/DisplayPreferences                 { proxy_pass http://user_service; }

    # === PLAYBACK SERVICE (playstate, sessions, HLS, syncplay) ===
    location ~ ^/Users/[^/]+/(Played|Playing|Favorite)Items { proxy_pass http://playback_service; }
    location ~ ^/Sessions                           { proxy_pass http://playback_service; }
    location ~ ^/SyncPlay                           { proxy_pass http://playback_service; }
    location ~ ^/GetUtcTime                         { proxy_pass http://playback_service; }
    location ~ ^/Videos/[^/]+/live\.m3u8            { proxy_pass http://playback_service; }
    location ~ ^/Videos/[^/]+/(master|main)\.m3u8   { proxy_pass http://playback_service; }
    location ~ ^/Videos/[^/]+/hls1/                 { proxy_pass http://playback_service; }
    location ~ ^/Videos/[^/]+/(stream|stream\.)      { proxy_pass http://playback_service; }
    location ~ ^/Videos/ActiveEncodings             { proxy_pass http://playback_service; }
    location ~ ^/Audio/[^/]+(stream|universal)      { proxy_pass http://playback_service; }
    location ~ ^/MediaInfo                          { proxy_pass http://playback_service; }
    location ~ ^/Items/[^/]+/PlaybackInfo           { proxy_pass http://playback_service; }

    # === MEDIA SERVICE (images, subtitles, trickplay) ===
    location ~ ^/Items/[^/]+/Images                 { proxy_pass http://media_service; }
    location ~ ^/Users/[^/]+/Images                 { proxy_pass http://media_service; }
    location ~ ^/Items/[^/]+/RemoteImages           { proxy_pass http://media_service; }
    location ~ ^/Videos/[^/]+/Subtitles             { proxy_pass http://media_service; }
    location ~ ^/Providers/Subtitles                { proxy_pass http://media_service; }
    location ~ ^/Audio/[^/]+/Lyrics                 { proxy_pass http://media_service; }
    location ~ ^/Providers/Lyrics                   { proxy_pass http://media_service; }
    location ~ ^/MediaSegments                      { proxy_pass http://media_service; }
    location ~ ^/Videos/[^/]+/Trickplay             { proxy_pass http://media_service; }
    location ~ ^/Videos/[^/]+/Attachments           { proxy_pass http://media_service; }
    location ~ ^/Videos/[^/]+/SubtitleSegments      { proxy_pass http://media_service; }

    # === LIBRARY SERVICE ===
    location ~ ^/Items/RemoteSearch                 { proxy_pass http://metadata_service; }
    location ~ ^/Items/[^/]+/Refresh$              { proxy_pass http://metadata_service; }
    location ~ ^/Items/[^/]+/MetadataEditor$       { proxy_pass http://metadata_service; }
    location ~ ^/Items/Counts$                     { proxy_pass http://library_service; }
    location ~ ^/Items/[^/]+/Ancestors$            { proxy_pass http://library_service; }
    location ~ ^/Items/[^/]+/Similar$              { proxy_pass http://library_service; }
    location ~ ^/Items/[^/]+/ExternalIds$          { proxy_pass http://library_service; }
    location ~ ^/Items/[^/]+/ThemeMedia$           { proxy_pass http://library_service; }
    location ~ ^/Items                             { proxy_pass http://library_service; }
    location ~ ^/Library                           { proxy_pass http://library_service; }
    location ~ ^/Filters                           { proxy_pass http://library_service; }
    location ~ ^/Years                             { proxy_pass http://library_service; }

    # === BROWSE SERVICE ===
    location ~ ^/Artists                           { proxy_pass http://browse_service; }
    location ~ ^/Genres                            { proxy_pass http://browse_service; }
    location ~ ^/MusicGenres                       { proxy_pass http://browse_service; }
    location ~ ^/Studios                           { proxy_pass http://browse_service; }
    location ~ ^/Persons                           { proxy_pass http://browse_service; }
    location ~ ^/Shows                             { proxy_pass http://browse_service; }
    location ~ ^/Movies                            { proxy_pass http://browse_service; }
    location ~ ^/Trailers                          { proxy_pass http://browse_service; }
    location ~ ^/Channels                          { proxy_pass http://browse_service; }
    location ~ ^/(Songs|Albums|Artists|MusicGenres|Playlists)/[^/]+/InstantMix { proxy_pass http://browse_service; }
    location ~ ^/Items/[^/]+/InstantMix            { proxy_pass http://browse_service; }

    # === METADATA SERVICE ===
    location ~ ^/ScheduledTasks                    { proxy_pass http://metadata_service; }
    location ~ ^/Packages                          { proxy_pass http://metadata_service; }
    location ~ ^/Plugins                           { proxy_pass http://metadata_service; }

    # === COLLECTION SERVICE ===
    location ~ ^/Collections                       { proxy_pass http://collection_service; }

    # === PLAYLIST SERVICE ===
    location ~ ^/Playlists                         { proxy_pass http://playlist_service; }

    # === SEARCH SERVICE ===
    location ~ ^/Search                            { proxy_pass http://search_service; }
    location ~ ^/Items/[^/]+/Suggestions           { proxy_pass http://search_service; }

    # === ADMIN SERVICE ===
    location ~ ^/System                            { proxy_pass http://admin_service; }
    location ~ ^/System/Configuration             { proxy_pass http://admin_service; }
    location ~ ^/Branding                         { proxy_pass http://admin_service; }
    location ~ ^/Environment                      { proxy_pass http://admin_service; }
    location ~ ^/Localization                     { proxy_pass http://admin_service; }
    location ~ ^/Log                              { proxy_pass http://admin_service; }
    location ~ ^/Devices                          { proxy_pass http://admin_service; }

    # Common proxy settings
    proxy_http_version 1.1;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_connect_timeout 5s;
    proxy_read_timeout 300s;   # Long for streaming
    proxy_send_timeout 300s;
    proxy_buffering off;       # Critical for streaming responses

    # Health check
    location /healthz {
        return 200 '{"status":"ok"}';
        add_header Content-Type application/json;
    }
}
```

---

## 9. DOCKER COMPOSE

```yaml
version: '3.9'

services:
  mysql:
    image: mysql:8.0
    environment:
      MYSQL_ROOT_PASSWORD: ${MYSQL_ROOT_PASSWORD}
      MYSQL_DATABASE: jellyfin
      MYSQL_USER: jellyfin
      MYSQL_PASSWORD: ${MYSQL_PASSWORD}
    volumes:
      - mysql_data:/var/lib/mysql
    healthcheck:
      test: ["CMD", "mysqladmin", "ping", "-h", "localhost", "-u", "jellyfin", "-p${MYSQL_PASSWORD}"]
      interval: 10s
      timeout: 5s
      retries: 5

  redis:
    image: redis:7-alpine
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s

  nginx:
    image: nginx:alpine
    ports:
      - "8096:8096"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf:ro
      - ${JELLYFIN_WEB_PATH}:/usr/share/jellyfin-web:ro
    depends_on: [auth-service, user-service, library-service]

  auth-service:
    build: ./auth-service
    environment:
      DATABASE_URL: mysql://jellyfin:${MYSQL_PASSWORD}@mysql:3306/jellyfin?parseTime=true
      SERVICE_PORT: "8001"
      SERVER_ID: ${SERVER_ID}
    depends_on:
      mysql: { condition: service_healthy }
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8001/healthz"]
      interval: 30s

  user-service:
    build: ./user-service
    environment:
      DATABASE_URL: mysql://jellyfin:${MYSQL_PASSWORD}@mysql:3306/jellyfin?parseTime=true
      SERVICE_PORT: "8002"
      SERVER_ID: ${SERVER_ID}
      REDIS_URL: redis://redis:6379
    depends_on:
      mysql: { condition: service_healthy }
      redis: { condition: service_healthy }

  library-service:
    build: ./library-service
    environment:
      DATABASE_URL: mysql://jellyfin:${MYSQL_PASSWORD}@mysql:3306/jellyfin?parseTime=true
      SERVICE_PORT: "8003"
      SERVER_ID: ${SERVER_ID}
    depends_on:
      mysql: { condition: service_healthy }

  browse-service:
    build: ./browse-service
    environment:
      DATABASE_URL: mysql://jellyfin:${MYSQL_PASSWORD}@mysql:3306/jellyfin?parseTime=true
      SERVICE_PORT: "8004"
      SERVER_ID: ${SERVER_ID}
    depends_on:
      mysql: { condition: service_healthy }

  playback-service:
    build: ./playback-service
    environment:
      DATABASE_URL: mysql://jellyfin:${MYSQL_PASSWORD}@mysql:3306/jellyfin?parseTime=true
      SERVICE_PORT: "8005"
      SERVER_ID: ${SERVER_ID}
      REDIS_URL: redis://redis:6379
      FFMPEG_PATH: /usr/bin/ffmpeg
      FFPROBE_PATH: /usr/bin/ffprobe
      TRANSCODE_CACHE_DIR: /cache/transcode
      MEDIA_PATH: ${MEDIA_PATH}
    volumes:
      - transcode_cache:/cache/transcode
      - ${MEDIA_PATH}:${MEDIA_PATH}:ro
    depends_on:
      mysql: { condition: service_healthy }
      redis: { condition: service_healthy }

  media-service:
    build: ./media-service
    environment:
      DATABASE_URL: mysql://jellyfin:${MYSQL_PASSWORD}@mysql:3306/jellyfin?parseTime=true
      SERVICE_PORT: "8006"
      SERVER_ID: ${SERVER_ID}
      IMAGE_CACHE_DIR: /cache/images
      METADATA_PATH: ${METADATA_PATH}
    volumes:
      - image_cache:/cache/images
      - ${METADATA_PATH}:${METADATA_PATH}:ro
    depends_on:
      mysql: { condition: service_healthy }

  metadata-service:
    build: ./metadata-service
    environment:
      DATABASE_URL: mysql://jellyfin:${MYSQL_PASSWORD}@mysql:3306/jellyfin?parseTime=true
      SERVICE_PORT: "8007"
      SERVER_ID: ${SERVER_ID}
    depends_on:
      mysql: { condition: service_healthy }

  collection-service:
    build: ./collection-service
    environment:
      DATABASE_URL: mysql://jellyfin:${MYSQL_PASSWORD}@mysql:3306/jellyfin?parseTime=true
      SERVICE_PORT: "8008"
      SERVER_ID: ${SERVER_ID}
    depends_on:
      mysql: { condition: service_healthy }

  playlist-service:
    build: ./playlist-service
    environment:
      DATABASE_URL: mysql://jellyfin:${MYSQL_PASSWORD}@mysql:3306/jellyfin?parseTime=true
      SERVICE_PORT: "8009"
      SERVER_ID: ${SERVER_ID}
    depends_on:
      mysql: { condition: service_healthy }

  admin-service:
    build: ./admin-service
    environment:
      DATABASE_URL: mysql://jellyfin:${MYSQL_PASSWORD}@mysql:3306/jellyfin?parseTime=true
      SERVICE_PORT: "8010"
      SERVER_ID: ${SERVER_ID}
      JELLYFIN_LOG_DIR: /logs
      JELLYFIN_CONFIG_DIR: /config
    volumes:
      - ${JELLYFIN_LOG_DIR}:/logs:ro
      - ${JELLYFIN_CONFIG_DIR}:/config
    depends_on:
      mysql: { condition: service_healthy }

  search-service:
    build: ./search-service
    environment:
      DATABASE_URL: mysql://jellyfin:${MYSQL_PASSWORD}@mysql:3306/jellyfin?parseTime=true
      SERVICE_PORT: "8011"
      SERVER_ID: ${SERVER_ID}
    depends_on:
      mysql: { condition: service_healthy }

volumes:
  mysql_data:
  image_cache:
  transcode_cache:
```

---

## 10. MULTI-STAGE DOCKERFILE TEMPLATE

All services use this pattern:

```dockerfile
FROM golang:1.23-alpine AS builder
RUN apk add --no-cache git ca-certificates
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /server ./cmd/server

FROM alpine:3.21
RUN apk add --no-cache ca-certificates curl
WORKDIR /app
COPY --from=builder /server /app/server
EXPOSE 8001
HEALTHCHECK --interval=30s --timeout=5s CMD curl -f http://localhost:8001/healthz || exit 1
ENTRYPOINT ["/app/server"]
```

For playback-service, add ffmpeg:
```dockerfile
RUN apk add --no-cache ffmpeg
```

---

## 11. STANDARD `cmd/server/main.go` TEMPLATE

Every service must follow this pattern:

```go
package main

import (
    "context"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/go-chi/chi/v5"
    chimiddleware "github.com/go-chi/chi/v5/middleware"
    "github.com/jellyfinhanced/shared/auth"
    "github.com/jellyfinhanced/shared/db"
    "github.com/jellyfinhanced/shared/auth"
    "github.com/jellyfinhanced/shared/response"

    "github.com/jellyfinhanced/{service}/internal/handlers"
)

func main() {
    // 1. Connect to MySQL
    database, err := db.Connect(os.Getenv("DATABASE_URL"))
    if err != nil {
        log.Fatalf("db connect: %v", err)
    }
    defer database.Close()

    // 2. Build handler dependencies
    h := handlers.New(database)

    // 3. Setup router
    r := chi.NewMux()
    r.Use(chimiddleware.RealIP)
    r.Use(middleware.Logger)
    r.Use(middleware.Recovery)
    r.Use(middleware.ResponseTime)
    r.Use(middleware.CORS)
    r.Use(response.AddRequiredHeaders(os.Getenv("SERVER_ID")))

    // 4. Health check (anonymous)
    r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        w.Write([]byte(`{"status":"ok"}`))
    })

    // 5. Register routes (service-specific)
    h.RegisterRoutes(r, auth.NewMiddleware(database))

    // 6. Start server with graceful shutdown
    port := os.Getenv("SERVICE_PORT")
    if port == "" {
        port = "8001"
    }
    srv := &http.Server{
        Addr:         ":" + port,
        Handler:      r,
        ReadTimeout:  15 * time.Second,
        WriteTimeout: 300 * time.Second,
        IdleTimeout:  120 * time.Second,
    }

    go func() {
        log.Printf("Starting on :%s", port)
        if err := srv.ListenAndServe(); err != http.ErrServerClosed {
            log.Fatalf("server: %v", err)
        }
    }()

    // 7. Wait for signal
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
    <-sigCh

    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    srv.Shutdown(ctx)
}
```

---

## 12. TESTING REQUIREMENTS

### Per-Service Unit Tests
Every handler must have a corresponding `_test.go` using mock DB:
```go
// Minimum coverage target: 80%
go test ./... -cover -covermode=atomic
```

### Integration Tests (testcontainers-go)
```go
// tests/integration_test.go pattern for all services
func TestIntegration(t *testing.T) {
    ctx := context.Background()

    mysql, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
        ContainerRequest: testcontainers.ContainerRequest{
            Image: "mysql:8.0",
            Env: map[string]string{
                "MYSQL_ROOT_PASSWORD": "test",
                "MYSQL_DATABASE":      "jellyfin_test",
                "MYSQL_USER":          "jellyfin",
                "MYSQL_PASSWORD":      "test",
            },
            WaitingFor: wait.ForLog("ready for connections"),
        },
        Started: true,
    })

    // Run migrations
    // Start service
    // Run HTTP assertions against real MySQL
}
```

### Wire Compatibility Tests
For each endpoint, compare Go response to recorded C# response:
```go
func TestWireCompatibility(t *testing.T) {
    // Load golden file: testdata/responses/{endpoint}.json
    // Make request to Go service
    // Assert JSON fields match exactly (ignore ordering)
    // Assert status code matches
    // Assert Content-Type header matches
}
```

### Load Tests (vegeta)
```bash
# library-service /Items
echo "GET http://localhost:8003/Items?userId=test-id&limit=50&recursive=true" \
  | vegeta attack -duration=30s -rate=100 \
  | vegeta report
# Target: p95 < 200ms at 100 req/sec
```

---

## 13. AGENT ASSIGNMENT MATRIX

| Agent | Service | Port | Blocker | Priority | Status |
|-------|---------|------|---------|----------|--------|
| 0 | shared | — | None | P0 | ✅ Started |
| 1 | auth-service | 8001 | shared | P1 | ✅ Started (30%) |
| 2 | user-service | 8002 | shared | P2 | 🔴 New |
| 3 | library-service | 8003 | shared | P2 | 🔴 New |
| 4 | browse-service | 8004 | shared | P3 | 🔴 New |
| 5 | playback-service | 8005 | shared, redis | P2 | 🔴 New |
| 6 | media-service | 8006 | shared | P3 | 🔴 New |
| 7 | metadata-service | 8007 | library-service | P4 | 🔴 New |
| 8 | collection-service | 8008 | shared | P3 | ✅ Started |
| 9 | playlist-service | 8009 | shared | P3 | ✅ Started |
| 10 | admin-service | 8010 | shared | P3 | 🔴 New |
| 11 | search-service | 8011 | library-service | P4 | 🔴 New |
| 12 | livetv-service | 8012 | media-service | P5 (Phase 2) | 🔴 Deferred |
| 13 | nginx gateway | — | all services | P1 | 🟡 Partial |

**Parallelization strategy:**
- Week 1: Complete `shared` package (Agent 0)
- Week 2: `auth-service`, `user-service`, `library-service`, `playback-service` in parallel (Agents 1-5)
- Week 3: `browse-service`, `media-service`, `collection-service`, `playlist-service`, `admin-service` in parallel
- Week 4: `metadata-service`, `search-service`, nginx gateway finalization
- Week 5: Integration testing, shadow mode, wire compatibility validation
- Week 6-7: Canary → full cutover → C# shutdown

---

## 14. CRITICAL CORRECTNESS CHECKLIST

Before any service goes to production, validate every item:

- [ ] JSON uses PascalCase field names (not camelCase)
- [ ] GUIDs are lowercase hyphenated (scan with: `grep -r "strings.ToUpper.*UUID\|uuid.Upper" .`)
- [ ] Timestamps formatted as `2006-01-02T15:04:05.0000000Z` (7 decimal places)
- [ ] `RunTimeTicks` is int64 ticks (×10,000,000 for seconds)
- [ ] Empty slice fields serialize as `[]` not `null`
- [ ] `QueryResult.Items` is never null (use `make([]T, 0)`)
- [ ] DELETE endpoints return 204 with no body
- [ ] `X-Application-Version: 10.10.0` on every response
- [ ] `X-MediaBrowser-Server-Id: <uuid>` on every response
- [ ] Password hashing uses bcrypt cost=11
- [ ] Token generation uses `crypto/rand` (not `math/rand`)
- [ ] SQL parameters use `?` placeholders (MySQL style, NOT `$1`)
- [ ] `parseTime=true` in MySQL DSN (required for time.Time scanning)
- [ ] Auth header also accepted as query param `?api_key=<token>`
- [ ] Image ETag/Cache-Control headers set for browser caching
- [ ] HLS responses use `Content-Type: application/x-mpegURL`
- [ ] Binary responses (images, video) use `proxy_buffering off` in nginx

---

*Document version: 1.0 — 2026-03-13*
*Source: JellyFinhanced .NET 10 codebase analysis + Kabletown planning documents*
