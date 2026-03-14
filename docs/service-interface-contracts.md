# Kabletown Service Architecture

## Overview

Kabletown consists of three core microservices:
- **Auth Service** (Port 8081): Authentication, user management, API keys
- **Item Service** (Port 8004): Media metadata queries, library management
- **Streaming Service** (Port 8005): HLS/streaming, FFmpeg transcoding

---

## Service Boundaries

### What Each Service Owns

| Service | Data Ownership | External Dependencies |
|---------|----------------|----------------------|
| Auth Service | users, api_keys, devices, user_policies | None (standalone) |
| Item Service | base_items, item_values, user_data | MySQL only |
| Streaming Service | /transcode (ephemeral) | MySQL, FFmpeg, media files |

---

## Service Contracts

### Auth Service Interface

#### Device Registration

**Endpoint:** `POST /Devices`

**Request:**
```json
{
  "name": "My iPhone",
  "appName": "Kabletown iOS",
  "appVersion": "1.2.0"
}
```

**Response (200):**
```json
{
  "Id": "123e4567-e89b-12d3-a456-426614174000",
  "Name": "My iPhone",
  "AccessToken": "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
  "AppId": "987fed65-e43b-21d4-b567-537725285111"
}
```

**Behavior:**
- Creates new device record
- Generates 256-bit random API token
- Returns token for client to use in Authorization header
- No authentication required (public endpoint)

---

#### User Authentication

**Endpoint:** `POST /Sessions`

**Request:**
```json
{
  "Username": "admin",
  "Pw": "changeme"
}
```

**Response (200):**
```json
{
  "Id": "123e4567-e89b-12d3-a456-426614174000",
  "Name": "admin",
  "AccessToken": "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"
}
```

**Behavior:**
- Validates username + password (SHA256 hash comparison)
- Creates new session token in api_keys table
- Updates DateLastLogin timestamp
- Returns token for client to use in Authorization header

---

#### Token Validation (Internal)

**Method:** `ValidateToken(token string) (userID string, isAdmin bool, err error)`

**Input:** `AccessToken` string from X-Emby-Authorization header

**Output:**
- `userID`: UUID string of authenticated user
- `isAdmin`: Boolean admin flag
- `err`: sql.ErrNoRows if token invalid/revoked

**Query:**
```sql
SELECT ak.UserId, u.IsAdmin
FROM api_keys ak
JOIN users u ON ak.UserId = u.Id
WHERE ak.Token = ? AND ak.IsActive = 1
```

**Usage:** Called by auth middleware to validate requests before passing to handlers.

---

### Item Service Interface

#### Query Items (InternalItemsQuery)

**Endpoint:** `GET /Items`

**Query Parameters (80+ possible):**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| TopParentId | UUID | No | Filter by library/root folder |
| IncludeItemTypes | string | No | Comma-separated: Movie,Episode,Season |
| ExcludeItemTypes | string | No | Comma-separated types to exclude |
| OrderBy | string | No | Field,Direction pairs (DateCreated,Descending) |
| StartIndex | int | No | Pagination start (default 0) |
| Limit | int | No | Page size (default 20) |
| UserId | UUID | No | Attach user_data for playstate |
| Recursive | bool | No | Include child folders |
| IsFavorite | bool | No | Filter by favorite status |
| IsPlayed | bool | No | Filter by play status |
| GenreIds | string | No | Filter by Genre (ValueType=2) |
| StudioIds | string | No | Filter by Studio (ValueType=3) |

**Request:**
```
GET /Items?TopParentId=abc123&IncludeItemTypes=Movie&OrderBy=DateCreated,Descending&StartIndex=0&Limit=20&UserId=def456
```

**Response:**
```json
{
  "Items": [
    {
      "Id": "123e4567-e89b-12d3-a456-426614174000",
      "Name": "The Matrix",
      "Type": "Movie",
      "RunTimeTicks": 8280000000000,
      "UserData": {
        "IsPlayed": false,
        "PlaybackPositionTicks": 0,
        "IsFavorite": true
      }
    }
  ],
  "TotalRecordCount": 150,
  "StartIndex": 0
}
```

**Query Complexity:**
- Joins: `base_items` LEFT JOIN `user_data`
- Subqueries: Nested ItemValue lookups (Genre, Studio, etc.)
- CTE: Recursive hierarchy traversal (when Recursive=true)
- P6 Index: `WHERE ValueType=2 AND NormalizedValue='action'`
- P7 Index: `WHERE TopParentId=xxx AND Type='Movie'`

---

#### Get Single Item

**Endpoint:** `GET /Items/{itemId}`

**Path:** `itemId` = UUID string

**Query Param:** `UserId` (optional, for user_data)

**Response (200):**
```json
{
  "Id": "123e4567-e89b-12d3-a456-426614174000",
  "Name": "The Matrix",
  "Type": "Movie",
  "Path": "/media/movies/The Matrix.mkv",
  "Container": "matroska",
  "RunTimeTicks": 8280000000000,
  "Width": 1920,
  "Height": 1080,
  "ProductionYear": 1999,
  "UserData": {
    "IsPlayed": false,
    "PlaybackPositionTicks": 0,
    "IsFavorite": true,
    "Rating": 9
  }
}
```

**Response (404):**
```json
{"Message": "Item not found", "StatusCode": 404}
```

---

#### Get Recently Added

**Endpoint:** `GET /Items/RecentlyAdded`

**Query Parameters:**
- `UserId` (required) - User to attach playstate for
- `Limit` (optional) - Number of items (default 20)
- `TopParentId` (optional) - Filter to specific library

**Implementation:**
```sql
SELECT b.*, ud.IsPlayed, ud.PlaybackPositionTicks
FROM base_items b
LEFT JOIN user_data ud ON b.Id = ud.ItemId AND ud.UserId = ?
WHERE b.Type != 'Folder'
ORDER BY b.DateCreated DESC
LIMIT ?
```

---

### Streaming Service Interface

#### HLS Master Playlist

**Endpoint:** `GET /Videos/{itemId}/master.m3u8`

**Query Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| MediaSourceId | UUID | Yes | Select media version |
| StartTimeTicks | int64 | No | Start position (0=beginning) |
| DeviceId | string | Yes | Used to stop encoding on disconnect |
| PlaySessionId | string | Yes | Unique session tracking |
| VideoCodec | string | No | Target: h264, hevc, copy |
| AudioCodec | string | No | Target: aac, ac3, mp3, copy |
| VideoBitRate | int | No | Target bitrate (bps) |
| MaxWidth/MaxHeight | int | No | Resolution caps |
| TranscodeReasons | string | No | Client hints (VideoCodecNotSupported, etc.) |

**Response:**
```
#EXTM3U
#EXT-X-VERSION:3
#EXT-X-INDEPENDENT-SEGMENTS
#EXT-X-STREAM-INF:BANDWIDTH=8000000,RESOLUTION=1920x1080,CODECS="avc1.640028,mp4a.40.2"
1234567890123456789012345678901234567890-variant1.m3u8
#EXT-X-STREAM-INF:BANDWIDTH=4000000,RESOLUTION=1280x720,CODECS="avc1.4d401f,mp4a.40.2"
1234567890123456789012345678901234567890-variant2.m3u8
```

**Behavior:**
1. Look up item from Item Service (or local base_items cache)
2. Choose MediaSource based on MediaSourceId
3. Calculate available bitrate/quality options
4. For transcoding: Start FFmpeg job, register in TranscodeManager
5. Generate master playlist referencing variant playlists
6. Track job in TranscodeManager: ActiveRequestCount++
7. Return application/vnd.apple.mpegurl

---

#### HLS Segment

**Endpoint:** `GET /Videos/{itemId}/hls/{playlistId}/{segmentId}.{segmentContainer}`

**Path Params:**
- `playlistId` - Variant playlist identifier
- `segmentId` - Segment number (0, 1, 2, ...)
- `segmentContainer` - ts, mp4, or m4s

**Response:**
- Status: 200 OK with video segment binary
- Content-Type: video/mp2t (ts) or video/mp4

**Behavior:**
1. Resolve path from segmentId
2. LockAsync(outputPath) - prevent concurrent FFmpeg
3. Check if segment exists in disk
4. Start FFmpeg if job not running (via TranscodeManager)
5. Serve segment file
6. Track access: LastPingTime update

---

#### Progressive Stream

**Endpoint:** `GET /Videos/{itemId}/stream`

**Query Parameters:**
- `Static` (bool) - Stream copy vs transcode
- `StartTimeTicks` (int64) - Start position
- `VideoCodec` (string) - Target codec if transcode
- `AudioCodec` (string) - Target codec if transcode

**Request (Direct Play):**
```
GET /Videos/123e4567-e89b-12d3-a456-426614174000/stream?Static=true&MediaSourceId=abc123
```

**Response:**
- Content-Type: video/matroska (or appropriate)
- Body: Direct file stream with Range header support

**Request (Transcode):**
```
GET /Videos/123e4567-e89b-12d3-a456-426614174000/stream?VideoCodec=h264&AudioCodec=aac&VideoBitRate=5000000
```

**Response:**
- Content-Type: video/mp4
- Body: FFmpeg stdout (transcoded stream)

**Behavior:**
1. Build StreamState (48+ params)
2. Decision: Transcode vs Direct Play? (see TranscodeDecisionFlow)
3. If Direct Play: Stream file with Range header support
4. If Transcode:
   - LockAsync(outputPath)
   - StartFfMpeg(state, ffmpegArgs)
   - Stream stdout directly to HTTP response

---

## Authentication Flow

### Header Format

```
X-Emby-Authorization: MediaBrowser Token="abcdef1234...", DeviceId="xyz789...", Client="Kabletown", Version="1.0.0"
```

### Middleware Chain

```
Incoming Request
    ↓
Auth Middleware
    ↓
ParseMediaBrowserHeader(header)
    ↓
Extract Token + DeviceId
    ↓
ValidateToken(Token) → (userId, isAdmin, err)
    ↓
If err != nil: return 401 Unauthorized
    ↓
Populate Context:
  context.Set("user_id", userId)
  context.Set("device_id", DeviceId)
  context.Set("is_admin", isAdmin)
    ↓
Call Next Handler
```

### Helper Functions (Context)

```go
// Extract from context (returns empty string if not set)
func GetUserID(r *http.Request) string {
    if v := r.Context().Value("user_id"); v != nil {
        return v.(string)
    }
    return ""
}

func GetDeviceID(r *http.Request) string {
    if v := r.Context().Value("device_id"); v != nil {
        return v.(string)
    }
    return ""
}

func IsAdmin(r *http.Request) bool {
    if v := r.Context().Value("is_admin"); v != nil {
        return v.(bool)
    }
    return false
}
```

---

## Error Handling Patterns

### Standard Error Response

```json
{
  "Message": "Invalid token",
  "StatusCode": 401
}
```

### HTTP Status Codes by Service

| Service | 400 | 401 | 403 | 404 | 500 |
|---------|-----|-----|-----|-----|-----|
| Auth | Invalid JSON | Wrong password/token | N/A | User/Device not found | DB error |
| Item | Bad query params | Auth required | N/A | Item not found | Query error |
| Streaming | Bad stream params | Auth required | N/A | Item/segment not found | FFmpeg error |

---

## FFmpeg Process Management

### TranscodeManager API

```go
type TranscodeManager struct {
    Jobs map[string]*TranscodingJob  // keyed by PlaySessionId
    MutexMap *AsyncKeyedLocker[string] // keyed by output path
}

// Lifecycle methods
func (m *TranscodeManager) StartFFmpeg(state *StreamState, outputPath, args string) (*TranscodingJob, error)
func (m *TranscodeManager) OnTranscodeBeginRequest(PlaySessionId string) *TranscodingJob
func (m *TranscodeManager) OnTranscodeEndRequest(job *TranscodingJob) 
func (m *TranscodeManager) KillTranscodingJobs(deviceId, playSessionId string)
```

### Job State Machine

```
┌─────────────┐
│   Created   │ ← StartFFmpeg
└──────┬──────┘
       │
       │ FFmpeg process launched
       ▼
┌─────────────┐
│   Running   │ ─────┐
└──────┬──────┘      │
       │             │ (HLS)
       │             │ OnTranscodeBeginRequest
       │             │ ActiveRequestCount++
       │             │
       │             │ OnTranscodeEndRequest
       │             │ ActiveRequestCount--
       │             │ If 0: StartKillTimer(10s/60s)
       │             │
       │             │ Client disconnects
       │             │
       │             └─────────────┐
       │                           │
       │                   KillTimer fires
       │                           │
       │                           ▼
       │                   ┌─────────────┐
       │                   │   Terminated│
       │                   │ Delete files│
       │                   └─────────────┘
       │
       │ (Progressive)
       │ FFmpeg finishes normally
       ▼
┌─────────────┐
│   Finished  │
└─────────────┘
```

### Key Decision Points

**Transcode vs Direct Play (Progressive):**

```go
func needsTranscode(state *StreamState) bool {
    // Stream copy allowed if:
    // 1. Static=true
    // 2. Client supports container
    // 3. Client supports video/audio codecs
    // 4. Container compatibility OK
    // 5. Bitrate within limits
    
    if state.Request.Static {
        return false  // Force stream copy
    }
    
    if !clientSupportsContainer(state.ClientCapabilities) {
        return true  // Must transcode container
    }
    
    if !clientSupportsVideoCodec(state.VideoStream, state.Request.VideoCodec) {
        return true  // Must transcode video
    }
    
    if !clientSupportsAudioCodec(state.AudioStream, state.Request.AudioCodec) {
        return true  // Must transcode audio
    }
    
    if state.VideoStream.BitRate > state.MaxStreamingBitrate {
        return true  // Exceeds user policy
    }
    
    return false  // Direct play OK
}
```

**HLS Variant Bitrate Selection:**

```go
func selectHlsVariants(state *StreamState) []string {
    // Return list of variant bitrates/resolutions
    
    variants := []string{
        "1920x1080@8000000",  // 8 Mbps
        "1280x720@4000000",   // 4 Mbps
        "854x480@1500000",    // 1.5 Mbps
        "640x360@800000",     // 0.8 Mbps
    }
    
    // Filter by client caps (MaxWidth, MaxHeight, MaxVideoBitRate)
    // Return filtered list as variant playlist URLs
}
```

---

## Inter-Service Communication

### No Direct Service-to-Service Calls

Each service is self-contained and validates requests independently:

```
Client → Nginx → Auth Service (validate token)
Client → Nginx → Item Service (validate same token)
Client → Nginx → Streaming Service (validate same token)
```

### Shared Database Pattern

- **Auth Service**: `users`, `api_keys`, `devices`, `user_policies`
- **Item Service**: `base_items`, `item_values`, `user_data`, `user_policies`
- **Streaming Service**: Reads `base_items`, `user_policies` (for bitrate caps)
- **All services**: Shared MySQL connection

### Why This Works

1. **Token is a lookup key**: `SELECT userId FROM api_keys WHERE token = ?`
2. **Each service can validate**: No need to call Auth Service
3. **Consistent user context**: All services query same `user_id`
4. **Reduced coupling**: Auth Service can be taken down without breaking Item/Streaming

---

## Observability

### Logging Format (JSON)

```json
{"level":"info","timestamp":"2026-03-13T16:50:00Z","service":"streaming","PlaySessionId":"abc123","DeviceId":"xyz789","Action":"FFmpegStarted","CmdArgs":"-i /media/movie.mkv -c:v libx264 ..."}
```

### Metrics

| Name | Type | Service | Description |
|------|------|---------|-------------|
| `transcoding_jobs_active` | Gauge | streaming | Currently running FFmpeg jobs |
| `transcoding_bytes_sent_total` | Counter | streaming | Total transcoded bytes |
| `auth_tokens_created_total` | Counter | auth | Total tokens created |
| `item_queries_total` | Counter | item | Total item queries |
| `item_query_duration_seconds` | Histogram | item | Query execution time |

### Health Check Endpoint

```bash
GET /health
```

**Response:**
```json
{"status":"ok"}
```

**All Services:**
- Auth: Database connection test
- Item: Database connection test
- Streaming: Database connection test + FFmpeg availability

---

## Configuration

### Environment Variables

| Variable | Auth Service | Item Service | Streaming Service |
|----------|--------------|--------------|-------------------|
| `PORT` | 8081 | 8004 | 8005 |
| `DB_HOST` | - | mysql | mysql |
| `DB_PORT` | - | 3306 | 3306 |
| `DB_USER` | - | jellyfin | jellyfin |
| `DB_PASSWORD` | - | jellyfin_password | jellyfin_password |
| `DB_NAME` | - | jellyfin | jellyfin |
| `TRANSCODE_DIR` | - | - | /transcode |
| `MEDIA_DIR` | - | - | /media |
| `FFMPEG_PATH` | - | - | /usr/bin/ffmpeg |

---

## Security Considerations

### Path Traversal Prevention

All file paths validated before file operations:

```go
func validatePath(outputPath string) bool {
    // Ensure path is within allowed directory
    transcodeDir := "/transcode"
    absPath, _ := filepath.Abs(outputPath)
    return strings.HasPrefix(absPath, transcodeDir)
}
```

### SQL Injection Prevention

All queries use parameterized statements:

```go
// GOOD
row := db.QueryRow("SELECT Id FROM users WHERE Username = ?", username)

// BAD - DO NOT DO THIS
query := "SELECT Id FROM users WHERE Username = '" + username + "'"
```

### Container Validation Regex

Prevent shell injection via query params:

```go
const ContainerValidationRegex = `^[\w. ,\-]*$`

// Used in route definition:
[FromQuery][RegularExpression(ContainerValidationRegex)]
```

---

## Performance Guidelines

### Connection Pooling

```go
// Item Service - MySQL
config := &sqlconfig.Config{
    MaxOpenConns: 20,  // Per service
    MaxIdleConns: 5,
}
```

### Memory Limits

- **HLS segment buffer**: 4MB per job
- **FFmpeg temp files**: Clean every 30 minutes
- **TranscodeManager cache**: Max 10 concurrent jobs

### Caching Strategies

- **Item metadata**: Redis cache (15 min TTL)
- **User data**: In-memory (invalidate on update)
- **Transcode jobs**: TranscodeManager registry

---

## Migration Path (Jellyfin → Kabletown)

1. **Phase 1: Coexistence**
   - Run both Jellyfin and Kabletown on separate ports
   - Users can migrate accounts (export/import DB)

2. **Phase 2: Auth Sync**
   - Shared MySQL database
   - Create users in both systems

3. **Phase 3: Full Migration**
   - Point clients to Kabletown
   - Decommission Jellyfin

### Data Transfer

```bash
# Backup Jellyfin MySQL
mysqldump -h jellyfin_db -u root -p jellyfin > jellyfin_backup.sql

# Transfer to Kabletown
mysql -h kabletown_mysql -u jellyfin -p jellyfin < jellyfin_backup.sql

# Run migration scripts
mysql -h kabletown_mysql -u jellyfin -p jellyfin < Kabletown/migrations/sync_from_jellyfin.sql
```
