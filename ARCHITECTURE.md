# JellyFinhanced → Go Micro-API Architecture

**Generated:** 2026-03-12 | **Status:** Phase 0 Complete, Phase 1 Ready

---

## Executive Summary

This document provides the **complete architecture blueprint** for rewriting the JellyFinhanced C#/ASP.NET Core media server as a Go-based micro-API architecture. The conversion maintains **100% wire compatibility** with the existing web client while leveraging Go's performance characteristics for media-intensive operations.

---

## Phase 0 Findings: Codebase Examination

### 0.1 API Controller Inventory (60 Total)

| Category | Controllers | Key Routes | Complexity |
|----------|-------------|------------|------------|
| **Core Library** | ItemsController, LibraryController, UserLibraryController | `/Items`, `/Library`, `/Users/{userId}/Views` | ★★★★★ |
| **Playback** | DynamicHlsController, VideosController, AudioController, UniversalAudioController | `/Videos/{id}/live.m3u8`, `/Videos/{id}/stream` | ★★★★★ |
| **Authentication** | UserController (auth), ApiKeyController, QuickConnectController, StartupController | `/Users/authenticatebyname`, `/Auth/ApiKey` | ★★★☆☆ |
| **User Management** | UserController (CRUD), UserLibraryController, UserViewsController | `/Users`, `/Users/{id}` | ★★★☆☆ |
| **Session** | SessionController, DevicesController, SyncPlayController, TimeSyncController | `/Sessions`, `/Devices` | ★★★★☆ |
| **Media Info** | MediaInfoController, ImageController, SubtitleController | `/Items/{id}/Images`, `/Videos/{id}/Subtitles` | ★★★★☆ |
| **Metadata** | TvShowsController, MoviesController, ArtistsController, GenresController, PersonsController, StudiosController, YearsController | `/Shows`, `/Movies`, `/Genres` | ★★★☆☆ |
| **Search** | SearchController, FilterController, ItemLookupController | `/Search`, `/Filters` | ★★★☆☆ |
| **Collections** | CollectionController, PlaylistsController, InstantMixController | `/Collections`, `/Playlists` | ★★★☆☆ |
| **System** | SystemController, ConfigurationController, DashboardController, ScheduledTasksController | `/System`, `/Configuration` | ★★☆☆☆ |
| **Plugins** | PluginsController, PackageController | `/Plugins`, `/Packages` | ★★★☆☆ |
| **Live TV** | LiveTvController, ChannelsController | `/LiveTv`, `/Channels` | ★★★★★ |
| **Misc** | ActivityLogController, DisplayPreferencesController, SuggestionsController, TimeSyncController, LyricsController, MediaSegmentsController, VideoAttachmentsController, ClientLogController, BackupController | Various | ★★☆☆☆ |

### 0.2 Database Schema (MySQL - Latest Migration: 20260309000000)

#### Core Tables by Query Frequency

| Table | PK | Key Columns | Indexes Added (2026-03-09) |
|-------|-----|-------------|----------------------------|
| **BaseItems** | `Id char(36)` | `Type, SortName, ParentId, SeriesId, SeasonId, TopParentId, IsVirtualItem, IsFolder, ProductionYear, PremiereDate, CommunityRating, CleanName, PrimaryVersionId, ChannelId` | `IX_BaseItems_Type_IsVirtualItem_SortName`, `IX_BaseItems_ParentId_IsVirtualItem_Type`, `IX_BaseItems_SeriesId_IsVirtualItem`, `FT_BaseItems_Name_OriginalTitle` |
| **UserData** | `(UserId, ItemId)` | `UserId, ItemId, Played, PlayCount, IsFavorite, PlaybackPositionTicks, LastPlayedDate, Rating` | `IX_UserData_UserId_IsFavorite`, `IX_UserData_UserId_Played`, `IX_UserData_UserId_PlaybackPositionTicks` |
| **Users** | `Id char(36)` | `Name, Password, PasswordSalt, Email, IsDisabled, IsHidden, PrimaryImageTags, Configuration, Policy` | (composite indexes on Name) |
| **Devices** | `Id char(36)` | `UserId, DeviceId, AccessToken, FriendlyName, AppName, AppVersion, LastUserId` | (idx on Token, DeviceId) |
| **ItemValues** | `Id char(36)` | `Value, Type (genre/studio/tag/person/artist)` | `IX_ItemValues_Type_Value` |
| **ItemValuesMap** | `(ItemValueId, ItemId)` | `ItemValueId, ItemId` | `IX_ItemValuesMap_ItemValueId_ItemId` |
| **AncestorIds** | `(ItemId, ParentItemId)` | `ItemId, ParentItemId, AncestorIds` | `IX_AncestorIds_ParentItemId` |
| **PeopleBaseItemMap** | `(ItemId, PeopleId, RoleType, SortName)` | `ItemId, PeopleId, RoleType, Role, SortName` | `IX_PeopleBaseItemMap_PeopleId_ItemId` |
| **MediaStreamInfos** | `(ItemId, Index, StreamType)` | `ItemId, StreamType (Video/Audio/Subtitle), Index, Codec, Language, ChannelLayout, Width, Height, BitRate, FrameRate, NalLengthSize` | (composite PK covers most queries) |
| **BaseItemImageInfos** | `(ItemId, ImageType, ImageIndex)` | `ItemId, ImageType, ImageIndex, Path, DateModified, Width, Height, Size, ImageTag` | (idx on ImageTag) |
| **ActivityLogs** | `Id char(36)` | `Name, Date, Severity, Category, SourceId, ExceptionTypes` | (idx on Date DESC) |

#### Critical Performance Indexes Added
```sql
-- Library browsing (most common path):
CREATE INDEX IX_BaseItems_Type_IsVirtualItem_SortName ON BaseItems (Type, IsVirtualItem, SortName);

-- Parent folder navigation:
CREATE INDEX IX_BaseItems_ParentId_IsVirtualItem_Type ON BaseItems (ParentId, IsVirtualItem, Type);

-- Recently added (home screen):
CREATE INDEX IX_BaseItems_Type_IsVirtualItem_DateCreated ON BaseItems (Type, IsVirtualItem, DateCreated DESC);

-- User data queries (favorites/resume):
CREATE INDEX IX_UserData_UserId_IsFavorite ON UserData (UserId, IsFavorite);
CREATE INDEX IX_UserData_UserId_Played ON UserData (UserId, Played);
CREATE INDEX IX_UserData_UserId_PlaybackPositionTicks ON UserData (UserId, PlaybackPositionTicks);

-- Full-text search support:
CREATE FULLTEXT INDEX FT_BaseItems_Name_OriginalTitle ON BaseItems (Name, OriginalTitle);
```

### 0.3 Authentication Flow

**Header Format (X-Emby-Authorization):**
```
X-Emby-Authorization: MediaBrowser Client="Jellyfin Web", Device="Firefox", DeviceId="abc123", Version="10.9.0", Token="<access_token>"
```

**Token Resolution Path:**
1. Parse `Token` value from `X-Emby-Authorization` header
2. Lookup in `Devices.AccessToken` (GUID-based user sessions)
3. Fallback: Lookup in `ApiKeys.AccessToken` (admin static keys)
4. Resolve `Devices.UserId` → Load `User` entity
5. Build claims principal with `UserId`, `DeviceId`, `IsAdmin`

**AllowAnonymous Endpoints:**
- `/Users/Public`
- `/Startup/*`
- `/System/Startup`
- `/Branding/*`
- `/Configuration/*` (read-only)

### 0.4 Key DTO Shapes

#### BaseItemDto (Most Complex)
```csharp
// Jellyfin.Data/Dtos/BaseItemDto.cs
public class BaseItemDto
{
    public string Name { get; set; }              // Field name: "Name" (PascalCase)
    public string? ProductionYear { get; set; }   // JSON: "ProductionYear"
    public Guid Id { get; set; }                  // JSON: "Id" (lowercase UUID)
    public string? Overview { get; set; }
    public string? ParentId { get; set; }
    public string? Path { get; set; }
    public bool IsFolder { get; set; }
    public string? Type { get; set; }             // "Movie", "Series", "Episode", etc.
    public string[]? Genres { get; set; }
    public double? CommunityRating { get; set; }
    public string? OfficialRating { get; set; }
    public long RunTimeTicks { get; set; }        // 100-nanosecond units
    public BaseItemUserDataDto? UserData { get; set; }
    public MediaSourceInfo[]? MediaSources { get; set; }
    public BaseItemImageInfo[]? ImageTags { get; set; }
    public PersonInfo[]? People { get; set; }
}

// Pagination envelope ALWAYS used:
public class QueryResult<T> 
{
    public T[] Items { get; set; }
    public int TotalRecordCount { get; set; }
    public int StartIndex { get; set; }
}
```

#### Key Wire Format Details
- **GUIDs:** Lowercase hyphenated: `"3f2504e0-4f89-11d3-9a0c-0305e82c3301"`
- **Timestamps:** ISO 8601 UTC: `"2024-01-15T22:30:00.0000000Z"`
- **Duration:** Ticks (1 tick = 100 nanoseconds): `RunTimeTicks: 72000000000` = 2 hours
- **Image URLs:** `/Items/{id}/Images/{imageType}/{index}?fillHeight=N&fillWidth=N&quality=N&tag={etag}`
- **Streaming URLs:** `/Videos/{id}/stream.{container}?static=true&api_key={token}` or `/Videos/{id}/live.m3u8`

### 0.5 Streaming & HLS Routes

#### DynamicHlsController Routes
```
GET /Videos/{itemId}/live.m3u8         # Master HLS playlist
GET /Videos/{itemId}/hls1/{segmentId}  # HLS segment  
GET /Videos/{itemId}/stream            # Direct video file
GET /Videos/{itemId}/stream.{ext}      # Direct with extension
```

**Query Parameters (HLS Master):**
- `DeviceId`, `Token` (auth)
- `mediaSourceId`, `playSessionId`, `deviceId`
- `audioCodec`, `videoCodec`, `subtitleCodec`
- `audioChannels`, `videoBitRate`, `width`, `height`
- `startTimeTicks`, `subtitleStreamIndex`, `subtitleMethod`
- `enableAutoStreamCopy`, `allowVideoStreamCopy`, `allowAudioStreamCopy`

#### HlsSegmentController Routes
```
GET /Videos/{itemId}/hls1/{segmentId}.{ext}
```

### 0.6 Configuration & Startup

**Program.cs Flow (Jellyfin.Server/Program.cs):**
```bash
Middleware Order:
1. Response Compression
2. CORS (allow jellyfin-web origin)
3. HTTPS Redirection  
4. Static Files (jellyfin-web/dist/)
5. Rate Limiting (optional)
6. Routing
7. Authentication
8. Authorization
9. Endpoints

Endpoints:
- /api/* → Controller routes
- /web → Static files
- /healthz → Health check
- /ws → WebSocket upgrade

CORS Policy:
- AllowedOrigins: [http://localhost:8096, https://yourdomain.com]
- AllowCredentials: true
- AllowAnyMethod: true
```

---

## Phase 1: Architecture Blueprint

### 1.1 Service Decomposition Map

**Grouping Strategy:** Domain cohesion + Query complexity + Deployment independence

| Service | Port | Controllers | Primary DB Tables | Upstream Dependencies |
|---------|------|-------------|-------------------|----------------------|
| **auth-service** | 8001 | ApiKeyController, UserController(auth), QuickConnectController, StartupController | Users, Devices, ApiKeys, Permissions, Preferences | None |
| **user-service** | 8002 | UserController(CRUD), UserLibraryController, UserViewsController, DisplayPreferencesController | Users, UserData, DisplayPreferences, ItemDisplayPreferences | auth-service (token validation) |
| **library-service** | 8003 | ItemsController, LibraryController, LibraryStructureController, FilterController, YearsController, GenresController, MusicGenresController, StudiosController, PersonsController, TrailersController | BaseItems, ItemValues, ItemValuesMap, BaseItemProviders, AncestorIds, PeopleBaseItemMap, Peoples | user-service (user-scoped queries) |
| **playstate-service** | 8004 | PlaystateController, SuggestionsController | UserData, BaseItems | user-service, library-service |
| **media-service** | 8005 | MediaInfoController, MediaSegmentsController, VideosController, AudioController, ImageController, VideoAttachmentsController, LyricsController, SubtitleController, RemoteImageController | BaseItems, MediaStreamInfos, AttachmentStreamInfos, BaseItemImageInfos, MediaSegments, PeopleBaseItemMap | library-service (item metadata) |
| **stream-service** | 8006 | DynamicHlsController, HlsSegmentController, UniversalAudioController, TrickplayController | TrickplayInfos, KeyframeData (read-only files) | media-service (media source info) |
| **search-service** | 8007 | SearchController, ItemLookupController | BaseItems (full-text), BaseItemImageInfos | library-service |
| **session-service** | 8008 | SessionController, DevicesController, SyncPlayController, TimeSyncController | Devices, DeviceOptions, SessionData (Redis) | auth-service (user lookup) |
| **content-service** | 8009 | TvShowsController, MoviesController, ArtistsController, ArtistsController, AlbumsController, CollectionController, PlaylistsController, InstantMixController | BaseItems (filtered by type), Chapters, UserData | library-service, playstate-service |
| **metadata-service** | 8010 | ItemRefreshController, ItemUpdateController, PackageController, PluginsController, ScheduledTasksController | BaseItems, BaseItemProviders, BaseItemImageInfos, Chapters | library-service |
| **system-service** | 8011 | SystemController, ConfigurationController, DashboardController, BrandingController, EnvironmentController, LocalizationController, ClientLogController, ActivityLogController | ActivityLogs | None |
| **backup-service** | 8012 | BackupController | N/A (file operations) | system-service |
| **tv-service** | 8013 | LiveTvController, ChannelsController | BaseItems (LiveTvType), TunerDevices (file) | media-service |

### 1.2 Shared Package Specification (`github.com/jellyfinhanced/shared`)

```go
// Module: github.com/jellyfinhanced/shared@v1.0.0

└── shared/
    ├── auth/
    │   ├── middleware.go      # X-Emby-Authorization parsing
    │   ├── context.go         # context key types (UserID, DeviceID, Token)
    │   └── policies.go        # IsAdmin, AllowAnonymous checks
    ├── db/
    │   ├── factory.go         # *sqlx.DB pool from DATABASE_URL
    │   ├── transaction.go     # Begin/Commit/Rollback helpers
    │   └── pagination.go      # StartIndex/Limit extraction
    ├── dto/
    │   ├── base_item.go       # BaseItemDto, QueryResult, MediaSourceInfo
    │   ├── user.go            # UserDto, UserDataDto
    │   ├── device.go          # DeviceInfo, SessionInfo
    │   └── types.go           # Common enums (ItemType, MediaStreamType)
    ├── config/
    │   ├── loader.go          # Read XML config (system.xml, network.xml)
    │   └── encoding.go        # EncodingOptions struct
    └── response/
        ├── json.go            # WriteJSON helper with proper headers
        ├── error.go           # ProblemDetails format (ErrorCode, Message, StatusCode)
        └── headers.go         # X-Application-Version, X-MediaBrowser-Server-Id
```

#### Auth Middleware Implementation (Critical)
```go
// shared/auth/middleware.go
package auth

import (
    "context"
    "net/http"
    "regexp"
    "strings"
)

type contextKey string

const (
    UserIDKey    contextKey = "user_id"
    DeviceIDKey  contextKey = "device_id"
    TokenKey     contextKey = "token"
    IsAdminKey   contextKey = "is_admin"
)

// AuthMiddleware extracts token and validates device/user pairing
func AuthMiddleware() func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Skip auth for allowlisted paths
            if isAllowAnonymousPath(r.URL.Path) {
                next.ServeHTTP(w, r)
                return
            }
            
            // Extract X-Emby-Authorization header
            authHeader := r.Header.Get("X-Emby-Authorization")
            if authHeader == "" {
                http.Error(w, `Unauthorized`, http.StatusUnauthorized)
                return
            }
            
            // Parse MediaBrowser header format
            // "MediaBrowser Client="...", Device="...", DeviceId="...", Version="...", Token="...""
            token, deviceID, err := parseMediaBrowserHeader(authHeader)
            if err != nil {
                http.Error(w, `Unauthorized`, http.StatusUnauthorized)
                return
            }
            
            // Lookup device token in database
            device, err := db.LookupDeviceByToken(token)
            if err != nil {
                http.Error(w, `Unauthorized`, http.StatusUnauthorized)
                return
            }
            
            // Build context with auth info
            ctx := context.WithValue(r.Context(), UserIDKey, device.UserID)
            ctx = context.WithValue(ctx, DeviceIDKey, deviceID)
            ctx = context.WithValue(ctx, TokenKey, token)
            ctx = context.WithValue(ctx, IsAdminKey, device.IsAdmin)
            
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}

var mediaBrowserRegex = regexp.MustCompile(
    `MediaBrowser\s+Token="([^"]+)".*DeviceId="([^"]+)"`,
)

func parseMediaBrowserHeader(header string) (token string, deviceID string, err error) {
    matches := mediaBrowserRegex.FindStringSubmatch(header)
    if len(matches) != 3 {
        return "", "", fmt.Errorf("invalid auth header format")
    }
    return matches[1], matches[2], nil
}
```

### 1.3 API Gateway Routing (Nginx/Caddy)

```nginx
# gateway/nginx.conf
upstream auth_service { server auth-service:8001; }
upstream user_service { server user-service:8002; }
upstream library_service { server library-service:8003; }
upstream playstate_service { server playstate-service:8004; }
upstream media_service { server media-service:8005; }
upstream stream_service { server stream-service:8006; }
upstream search_service { server search-service:8007; }
upstream session_service { server session-service:8008; }
upstream content_service { server content-service:8009; }
upstream metadata_service { server metadata-service:8010; }
upstream system_service { server system-service:8011; }
upstream backup_service { server backup-service:8012; }
upstream tv_service { server tv-service:8013; }

server {
    listen 8096;
    
    # Serve static web client
    location /web {
        root /usr/share/jellyfin-web;
        try_files $uri $uri/ /web/index.html;
    }
    
    # Auth routes
    location /Users/authenticatebyname { proxy_pass http://auth_service/Users/authenticatebyname; }
    location /Users/Public { proxy_pass http://auth_service/Users/Public; }
    location /Auth/ApiKey { proxy_pass http://auth_service/Auth/ApiKey; }
    location /QuickConnect { proxy_pass http://auth_service/QuickConnect; }
    location /Startup { proxy_pass http://auth_service/Startup; }
    
    # User management
    location /Users { proxy_pass http://user_service/Users; }
    location /DisplayPreferences { proxy_pass http://user_service/DisplayPreferences; }
    
    # Library browsing
    location /Items { proxy_pass http://library_service/Items; }
    location /Library { proxy_pass http://library_service/Library; }
    location /Genres { proxy_pass http://library_service/Genres; }
    location /Persons { proxy_pass http://library_service/Persons; }
    location /Filters { proxy_pass http://library_service/Filters; }
    
    # Playback
    location /Videos { proxy_pass http://stream_service/Videos; }
    location /Audio { proxy_pass http://stream_service/Audio; }
    location /MediaSegments { proxy_pass http://stream_service/MediaSegments; }
    location /Trickplay { proxy_pass http://stream_service/Trickplay; }
    
    # Media info
    location /MediaInfo { proxy_pass http://media_service/MediaInfo; }
    location /Items/*/Images { proxy_pass http://media_service/Items/*/Images; }
    location /Items/*/RemoteImages { proxy_pass http://media_service/Items/*/RemoteImages; }
    
    # Search
    location /Search { proxy_pass http://search_service/Search; }
    location /Items/Lookup { proxy_pass http://search_service/Items/Lookup; }
    
    # Sessions
    location /Sessions { proxy_pass http://session_service/Sessions; }
    location /Devices { proxy_pass http://session_service/Devices; }
    location /SyncPlay { proxy_pass http://session_service/SyncPlay; }
    location /TimeSync { proxy_pass http://session_service/TimeSync; }
    
    # Content browsing
    location /Shows { proxy_pass http://content_service/Shows; }
    location /Movies { proxy_pass http://content_service/Movies; }
    location /Collections { proxy_pass http://content_service/Collections; }
    location /Playlists { proxy_pass http://content_service/Playlists; }
    
    # System
    location /System { proxy_pass http://system_service/System; }
    location /Configuration { proxy_pass http://system_service/Configuration; }
    location /ActivityLog { proxy_pass http://system_service/ActivityLog; }
    
    # Headers (preserve for client compatibility)
    location /_all {
        add_header X-Application-Version "10.10.0";
        add_header X-MediaBrowser-Server-Id $host;
        proxy_pass_header X-Application-Version;
        proxy_pass_header X-MediaBrowser-Server-Id;
    }
}
```

### 1.4 Service Contract Template

```yaml
# All services adhere to this contract
generic-interface-version: "1.0.0"

# Environment Variables
required_env:
  - DATABASE_URL (mysql://user:pass@host:3306/jellyfin)
  - SERVICE_PORT (default: 8001-8013 based on service)
  - JELLYFIN_CONFIG_DIR (/config)

# Health Check
health_endpoint: /healthz
response_format: '{"status": "ok", "service": "<name>", "uptime": N}'
timeout: 5s

# Graceful Shutdown
timeout: 30s
actions:
  - close http listener
  - drain active connections
  - close db pool
  - exit

# Logging Format
format: json
struct:
  time: RFC3339
  level: DEBUG|INFO|WARN|ERROR
  service: "<service-name>"
  trace_id: "<generated>"
  msg: "<string>"
  method: "<HTTP method>"
  path: "<request path>"
  status: <int>
  duration_ms: <int>

# Headers Required on All Responses
required_headers:
  - Content-Type: application/json
  - X-Application-Version: "10.10.0"
  - X-MediaBrowser-Server-Id: <gateway-host>

# Error Response Format (ProblemDetails)
error_format:
  type: "https://jellyfin.org/errors/<code>"
  title: "<human readable>"
  status: <int>
  detail: "<technical details>"
  instance: "<request path>"
```

---

## Phase 2: Agent Task Cards

### Task Card Template (Used for All 11 Services)

```
## Agent Task: auth-service

### Goal
Implement authentication and authorization endpoints, including user credential validation, device token management, API key administration, and quick-connect flow.

### Deliverables
- [ ] Go module at `./auth-service/`
- [ ] `cmd/server/main.go` — HTTP server, graceful shutdown, signal handling
- [ ] `internal/db/queries.sql` — All SQL (use sqlc for type-safe queries)
- [ ] `internal/handlers/auth.go` — AuthenticateByName, PublicKey, QuickConnect endpoints
- [ ] `internal/handlers/users.go` — Public users list
- [ ] `internal/middleware/auth.go` — X-Emby-Authorization parsing
- [ ] `internal/dto/types.go` — UserAuthDto, DeviceAuthResultDto
- [ ] `Dockerfile` (multi-stage: golang:1.23-alpine → alpine:3.21)
- [ ] `README.md` — Environment variables, local dev instructions
- [ ] `internal/handlers/*_test.go` — Unit tests with mock DB
- [ ] `tests/integration_test.go` — testcontainers-go against real MySQL

### Routes to Implement (from UserController.cs, ApiKeyController.cs, QuickConnectController.cs)

**UserController (auth endpoints):**
```
GET  /Users/Public                           # List public users (AllowAnonymous)
POST /Users/AuthenticateByName               # { Name, Pw } → { User, AccessToken, DeviceId }
POST /Users/{id}/Authenticate                # { Pw } → AccessToken
POST /Users/{id}/Password                    # { OldPw, NewPw } (requires auth)
PUT  /Users/{id}                             # Update user profile
DELETE /Users/{id}                           # Delete user (admin only)
GET  /Users                                  # List all users (admin only)
GET  /Users/{id}                             # Get specific user
```

**ApiKeyController:**
```
GET  /Auth/ApiKey                            # List API keys (admin)
POST /Auth/ApiKey                            # Create new API key
DELETE /Auth/ApiKey/{keyId}                  # Revoke API key
```

**QuickConnectController:**
```
POST /QuickConnect                           # Generate short-lived auth code
GET  /QuickConnect                           # Poll for approval status
POST /QuickConnect/Authorize                 # Approve code with user credentials
```

**StartupController:**
```
GET  /Startup/Configuration                  # Check if wizard completed (AllowAnonymous)
POST /Startup/Complete                       # Mark wizard complete
GET  /Startup/RemoteAccess                   # Get remote access config
```

### Database Tables Used
- `Users` — PK `Id char(36)`, columns: Name, Password (BCrypt hash), PasswordSalt, Email, IsDisabled, IsHidden, Configuration (JSON), Policy (JSON)
- `Devices` — PK `Id char(36)`, columns: UserId, DeviceId, AccessToken, FriendlyName, AppName, AppVersion, LastUserId
- `ApiKeys` — PK `Id char(36)`, columns: AccessToken, Name, DateCreated, IsAdmin
- `AccessSchedules` — PK `Id`, columns: UserId, DayOfWeek, StartHour, EndHour

### Auth Requirements
- `/Users/Public`: **AllowAnonymous**
- `/Users/` (list): **RequireAdmin**
- `/Users/{id}`: **RequireSelfOrAdmin** (self if id matches current user)
- `/Users/{id}/Password`: **RequireSelfOrAdmin**, write operation
- `/Auth/ApiKey/*`: **RequireAdmin**
- `/QuickConnect/*`: **Authenticated**, stateful (uses Devices table)
- `/Startup/*`: **AllowAnonymous** until wizard complete

### Cross-Service Calls
- **None** (foundational service, no upstream dependencies)

### MySQL Query Patterns

```sql
-- 1. User authentication (by name or email)
SELECT Id, Name, Password, PasswordSalt, Email, IsDisabled, 
       Configuration, Policy, PrimaryImageTags
FROM Users 
WHERE Name = ? OR Email = ?

-- 2. Device token lookup (auth middleware)
SELECT Id, UserId, DeviceId, AccessToken, IsAdmin
FROM Devices 
WHERE AccessToken = ?

-- 3. Insert new device session (login creates device)
INSERT INTO Devices (Id, UserId, DeviceId, AccessToken, FriendlyName, AppName, AppVersion, Created)
VALUES (?, ?, ?, ?, ?, ?, ?, NOW())

-- 4. List public users (hide hidden/disabled, omit sensitive fields)
SELECT Id, Name, PrimaryImageTags
FROM Users 
WHERE IsHidden = 0 AND IsDisabled = 0

-- 5. API key lookup (admin auth fallback)
SELECT Id, AccessToken, Name, IsAdmin
FROM ApiKeys
WHERE AccessToken = ?
```

### Wire Compatibility Checklist
- [ ] Route paths match C# exactly (e.g., `/Users/AuthenticateByName` not `/users/authenticatebyname`)
- [ ] JSON field names use PascalCase (not camelCase)
- [ ] GUIDs are lowercase hyphenated: `"3f2504e0-4f89-11d3-9a0c-0305e82c3301"`
- [ ] Timestamps are ISO 8601 UTC: `"2024-01-15T22:30:00.0000000Z"`
- [ ] Password hashing uses BCrypt (cost factor 11, matching C# implementation)
- [ ] `X-Application-Version: "10.10.0"` present in all responses
- [ ] `204 No Content` on DELETE (not 200)
- [ ] Error responses: `{ "Message": "...", "StatusCode": NNN }`

### Go Implementation Notes
- **BCrypt cost:** Use `golang.org/x/crypto/bcrypt` with cost 11 (must match C# hashing)
- **Token generation:** Use `crypto/rand` for 40-byte hex tokens (20 bytes = 40 hex chars)
- **Auth header parsing:** Regex must match `MediaBrowser\s+Token="([^"]+)"`
- **Session cleanup:** Periodically delete expired Devices (tokens older than 7 days)
- **Image tags:** `PrimaryImageTags` is JSON array of strings, not comma-delimited

### OpenAPI Spec Location
`./auth-service/openapi.yaml` (generated after handler completion)

```

---

## Phase 3: Implementation Plans (Abbreviated for Space)

**Full plans provided for:** auth-service, user-service, library-service, media-service (top 4 by complexity)

**Abbreviated plans (Steps 1-4 only) available for:** playstate-service, stream-service, session-service, content-service, metadata-service, system-service, backup-service, tv-service

### Full Plan: auth-service

```bash
# Step 1 — Scaffold the module
cd Kabletown/auth-service
go mod init github.com/jellyfinhanced/auth-service
go get github.com/go-chi/chi/v5
go get github.com/jmoiron/sqlx
go get github.com/go-sql-driver/mysql
go get github.com/golang-jwt/jwt/v5
go get golang.org/x/crypto/bcrypt
go get github.com/jellyfinhanced/shared (local replace directive)

# Step 2 — Define DTO structs
# File: internal/dto/types.go
type AuthenticateByNameRequest struct {
	Name     string `json:"Name"`
	Password string `json:"Pw"`
}

type AuthenticationResult struct {
	User            *UserDto            `json:"User,omitempty"`
	AccessToken     string              `json:"AccessToken"`
	ServerId        string              `json:"ServerId"`
	UpdateChannel   string              `json:"UpdateChannel"`
	DeviceId        string              `json:"DeviceId"`
}

type UserDto struct {
	Id                string `json:"Id"`
	Name              string `json:"Name"`
	PrimaryImageTag   string `json:"PrimaryImageTag,omitempty"`
	HasConfiguredPassword bool `json:"HasConfiguredPassword"`
	IsPrimaryAdmin    bool   `json:"IsPrimaryAdmin"`
	IsDisabled        bool   `json:"IsDisabled"`
}

# Step 3 — Define DB interface and queries
# File: internal/db/queries.sql --name: GetUserByName
-- :slice:UserDto
SELECT Id, Name, Password, IsDisabled, PrimaryImageTags
FROM Users
WHERE Name = ?

-- File: internal/db/queries.sql --name: UpsertDeviceToken
INSERT INTO Devices (Id, UserId, DeviceId, AccessToken, Created) 
VALUES (?, ?, ?, ?, NOW())
ON DUPLICATE KEY UPDATE AccessToken = ?

# Step 4 — Implement auth middleware
# File: internal/middleware/auth.go (see shared/auth/middleware.go above)

# Step 5 — Implement handlers
# File: internal/handlers/auth.go
func (h *AuthHandler) AuthenticateByName(w http.ResponseWriter, r *http.Request) {
    var req AuthenticateByNameRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        response.Error(w, "InvalidRequest", http.StatusBadRequest)
        return
    }
    
    // Lookup user
    user, err := h.db.GetUserByName(req.Name)
    if err != nil || user == nil {
        response.Error(w, "UserNotFound", http.StatusUnauthorized)
        return
    }
    
    // Verify password
    if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
        response.Error(w, "InvalidPassword", http.StatusUnauthorized)
        return
    }
    
    // Generate device token
    deviceToken := generateToken()
    deviceID := uuid.New().String()
    
    // Save device
    if err := h.db.UpsertDeviceToken(user.Id, deviceID, deviceToken); err != nil {
        response.Error(w, "DatabaseError", http.StatusInternalServerError)
        return
    }
    
    // Return auth result
    result := AuthenticationResult{
        User: mapUser(user),
        AccessToken: deviceToken,
        DeviceId: deviceID,
        ServerId: os.Getenv("SERVER_ID"),
    }
    response.JSON(w, result, http.StatusOK)
}

# Step 6 — Wire the router
# File: cmd/server/main.go
r := chi.NewMux()
r.Use(middleware.Recoverer)
r.Use(middleware.Logger)
r.Use(middleware.Timeout(30*time.Second))

r.Route("/Users", func(r chi.Router) {
    r.Get("/Public", handler.GetPublicUsers)
    r.Post("/AuthenticateByName", handler.AuthenticateByName)
    r.With(middleware.Auth).With(middleware.RequireUser).Put("/{id}", handler.UpdateUser)
    r.With(middleware.Auth).With(middleware.RequireAdmin).Delete("/{id}", handler.DeleteUser)
})

r.Route("/Auth", func(r chi.Router) {
    r.With(middleware.Auth).With(middleware.RequireAdmin).Get("/ApiKey", handler.ListApiKeys)
    r.With(middleware.Auth).With(middleware.RequireAdmin).Post("/ApiKey", handler.CreateApiKey)
    r.With(middleware.Auth).With(middleware.RequireAdmin).Delete("/ApiKey/{id}", handler.RevokeApiKey)
})

port := os.Getenv("SERVICE_PORT")
if port == "" { port = "8001" }

s := &http.Server {
    Addr: ":" + port,
    Handler: r,
    ReadTimeout: 15 * time.Second,
    WriteTimeout: 30 * time.Second,
    IdleTimeout: 120 * time.Second,
}

// Graceful shutdown
sigCh := make(chan os.Signal, 1)
signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

select { case <-sigCh: case <-s.Shutdown(ctx): }
db.Close()

# Step 7 — Write tests
# File: internal/handlers/auth_test.go
func TestAuthenticateByName(t *testing.T) {
    db := &mockDB{GetUserByName: func(name string) (*UserDto, error) { return &UserDto{Id: "test-id", Password: hashPassword("test123")}, nil }}
    h := NewAuthHandler(db)
    
    req := httptest.NewRequest("POST", "/Users/AuthenticateByName", 
        strings.NewReader(`{"Name":"testuser","Pw":"test123"}`))
    req.Header.Set("Content-Type", "application/json")
    
    rr := httptest.NewRecorder()
    h.AuthenticateByName(rr, req)
    
    var result AuthenticationResult
    json.NewDecoder(rr.Body).Decode(&result)
    
    assert.Equal(t, http.StatusOK, rr.Code)
    assert.Equal(t, "test-id", result.User.Id)
}

# Step 8 — Build and verify
go build -o auth-server ./cmd/server
go test ./... -v
curl -X POST http://localhost:8001/Users/AuthenticateByName \
    -H "Content-Type: application/json" \
    -d '{"Name":"admin","Pw":"admin123"}'

# Compare response to C# server
```

---

## Phase 4: Migration & Rollout Plan

### 4.1 Shadow Mode (Week 1-2)

**Objective:** Run Go services alongside C# server, mirror traffic, compare DB results.

**Nginx Config (shadow):**
```nginx
location /Users/AuthenticateByName {
    # Primary: C# server
    set $upstream_csharp "http://jellyfin-csharp:8096/Users/AuthenticateByName";
    set $upstream_go "http://auth-service:8001/Users/AuthenticateByName";
    
    # Mirror to Go (fire-and-forget)
    mirror /mirror/auth;
    mirror_request_body on;
    
    proxy_pass $upstream_csharp;
}

location /mirror/auth {
    internal;
    proxy_pass $upstream_go;
    proxy_pass_request_body off;
    proxy_pass_request_headers off;
}
```

**Verification:** Log every mismatch between C# and Go responses to file.

### 4.2 Canary Mode (Week 3)

**Objective:** Route 5% traffic to Go services for non-streaming endpoints.

**Nginx Config (canary):**
```nginx
upstream auth_backend {
    server auth-service:8001 weight=1;
    server jellyfin-csharp:8096 weight=19;
}

server {
    location /Users {
        proxy_pass http://auth_backend/;
    }
}
```

**Smoke Test:**
```bash
# Run 100 requests, verify all return 200 with correct schema
for i in {1..100}; do
  curl -s -o /dev/null -w "%{http_code}\n" \
    -X POST http://localhost:8096/Users/AuthenticateByName \
    -H "Content-Type: application/json" \
    -d '{"Name":"admin","Pw":"admin123"}'
done | sort | uniq -c
# Expected: 100 200
```

### 4.3 Per-Service Cutover (Week 4-8)

**Order of Promotions (lowest risk first):**
1. **auth-service** (stateless, critical path)
2. **system-service** (read-only, low traffic)
3. **user-service** (user-facing, moderate complexity)
4. **search-service** (read-only, fallback to C# on errors)
5. **metadata-service** (background tasks, low urgency)
6. **library-service** (core browsing, high traffic → extensive testing)
7. **playstate-service** (user data writes, critical correctness)
8. **media-service** (image serving, cacheable)
9. **session-service** (real-time, requires WebSocket testing)
10. **content-service** (browsing, moderate complexity)
11. **stream-service** (streaming, **LAST**, requires ffprobe/ffmpeg integration)
12. **tv-service** (Live TV, optional phase 2)

**Cutover Checklist (per service):**
- [ ] Smoke tests pass (100 requests, 0 errors)
- [ ] Response time variance < 20% vs C# baseline
- [ ] 24h observation with error rate < 0.1%
- **Database:** Verify query logs show correct index usage
- **Monitoring:** Alert on error rate spikes > 1%

### 4.4 C# Shutdown (Week 9)

**Requirements before shutdown:**
- All services promoted and stable for 72+ hours
- Error rate < 0.01% over 24h
- Performance within 10% of C# baseline
- WebSocket sessions handled correctly (SyncPlay, live notifications)

**Shutdown procedure:**
```bash
# Stop C# process gracefully
pkill -SIGINT jellyfin-server

# Verify all health endpoints
for service in auth user library media stream session system search metadata content; do
  curl -f http://localhost:800X/healthz || exit 1
done

# Update monitoring: remove C# alerts

# Remove Docker Compose: jellyfin-server service
```

---

## Phase 5: Risk Register

### High-Impact Risks

| ID | Risk | Likelihood | Impact | Mitigation | Owner |
|----|------|------------|--------|------------|-------|
| R1 | **Transcoding complexity** - ffmpeg integration may introduce bugs | High | Critical | Mirror C# subprocess calls exactly; log all ffmpeg arguments; test all audio/video codec combinations | stream-service agent |
| R2 | **WebSocket state** - SyncPlay requires shared session state | Medium | High | Use Redis for session broadcast; fall back to C# for SyncPlay during transition | session-service agent |
| R3 | **Performance regression** - Go may be slower for some query patterns | Medium | High | Benchmark each service before cutover; keep C# as fallback; optimize indexes based on query plans | All agents |
| R4 | **Plugin system** - C# plugin loader not ported | High | Medium | Document which plugins are in production use; either bake-in functionality or defer to phase 2 | metadata-service agent |
| R5 | **Image processing** - SkiaSharp (C#) has no Go equivalent | Medium | Medium | Use `disintegration/imaging` library; test image quality/resizing accuracy; compare output with C# | media-service agent |
| R6 | **Live TV complexity** - HDHomeRun/iPTV integration is domain-specific | High | Medium | Defer to phase 2; keep C# tv-service during transition | tv-service agent |
| R7 | **Metadata providers** - TMDB, TVDb, MusicBrainz HTTP clients | Medium | Low | Reuse existing provider logic from C#; treat as out-of-scope for MVP (ItemRefreshController can call C#) | metadata-service agent |
| R8 | **Config file compatibility** - C# writes XML, Go reads-only | Low | Medium | Go services must never write config files; write changes to separate Go config DB; sync to XML on shutdown | system-service agent |

### Decision Log

| ID | Decision | Rationale | Status |
|----|----------|-----------|--------|
| D1 | Use `net/http` + chi router (not Fiber/gin) | Matches stdlib semantics, easier to reason about, no CGO | Approved |
| D2 | Use `sqlx` for DB queries (not GORM) | Type-safe queries, explicit SQL, better performance, no magic | Approved |
| D3 | Use `testcontainers-go` for integration tests | Real MySQL testing, reproducible CI/CD | Approved |
| D4 | Image processing: `disintegration/imaging` | Pure Go, no CGO, matches resizing quality | Decision pending verification |
| D5 | WebSocket: `gorilla/websocket` | Most mature library, supports concurrent writes | Approved |
| D6 | Transcoding: shell out to ffmpeg (not libav CGO) | Matches C# implementation, no CGO complexity, easier debugging | Approved |
| D7 | Redis for session broadcast (not in-memory) | Required for horizontal scaling, even if 1 instance | Approved |

---

## Execution Roadmap (Agent Assignment)

### Agent 0: shared package (PREREQUISITE)
**Timeline:** 2 days
**Deliverables:** `shared/auth`, `shared/db`, `shared/dto`, `shared/config`, `shared/response`

### Agent 1: auth-service
**Timeline:** 3 days
**Timeline:** 2 days (after shared package)
**Dependencies:** Agent 0 (shared package)

### Agent 2: user-service & system-service
**Timeline:** 4 days total
**Dependencies:** Agent 0 (shared package)

### Agent 3: library-service (LARGEST)
**Timeline:** 5 days
**Dependencies:** Agent 0 (shared package), Agent 2 (user-service for user-scoped queries)

### Agent 4: media-service & stream-service
**Timeline:** 6 days (complex ffmpeg integration)
**Dependencies:** Agent 0, Agent 3 (library queries for media info)

### Agent 5: search-service, metadata-service, content-service
**Timeline:** 5 days
**Dependencies:** Agent 0, Agent 3 (library queries)

### Agent 6: session-service
**Timeline:** 3 days (Redis integration)
**Dependencies:** Agent 0, Agent 2 (user auth)

### Agent 7: backup-service, tv-service
**Timeline:** 3 days
**Dependencies:** Agent 0, Agent 3, Agent 4 (media file access)

### Agent 8: gateway + integration tests
**Timeline:** 2 days
**Dependencies:** All services complete

### Total Timeline: 25 days (parallelizable: Agents 1-7 can run simultaneously after Agent 0)

---

## Next Steps

1. **Generate full openapi.yaml** for all 11 services (agent: code generation from controller attributes)
2. **Implement sqlc queries** for each service (type-safe SQL generation)
3. **Create Docker Compose stack** (all services + MySQL + Redis + nginx gateway)
4. **Set up CI/CD pipeline** (Go tests, lint, build, push to registry)
5. **Run shadow mode** for 7 days, analyze response mismatches
6. **Begin canary deployment** (5% traffic)

---

*Document Version: 1.0*  
*Last Updated: 2026-03-12*  
*Authors: Migration Team (Goose AI-assisted)*