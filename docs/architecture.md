# Kabletown Architecture Blueprint

## System Overview

Kabletown is a **Go-based microservices replacement** for Jellyfin, designed for better performance, maintainability, and observability. It replicates Jellyfin's core functionality through loosely-coupled services communicating over HTTP/REST.

---

## Service Architecture

### 1. Auth Service (Port 8081)

**Responsibility:** Authentication, User Management, Device Registration, API Key Lifecycle

**Endpoints:**
- `POST /Devices` - Device registration (returns API key)
- `POST /Sessions` - User authentication (username/password → API key)
- `GET /Devices` - List user's registered devices (protected)
- `GET /Users` - Current user info (protected)

**Database Tables:**
- `users` - User accounts with SHA256 password hashes
- `api_keys` - Token-to-user mappings with device info
- `devices` - Device registration metadata
- `user_policies` - Permission flags (transcoding, favorites, etc.)

**Key Design Decisions:**
- Token format: 256-bit hex random (32 bytes = 64 hex chars)
- Token lifecycle: Stored until explicitly revoked by user
- Session management: No server-side expiration (JWT alternative)
- Password hashing: SHA256 (upgrade to bcrypt/argon2 if needed)

---

### 2. Item Service (Port 8004)

**Responsibility:** Media metadata queries, Library management, Item repository

**Endpoints:**
- `GET /Items` - Complex query (InternalItemsQuery 参数) - protected
- `GET /Items/{itemId}` - Single item lookup - public
- `GET /Items/RecentlyAdded` - Recently added feed - protected
- `GET /Items/NextEpisode` - Episode progression - protected

**Database Tables:**
- `base_items` - Core media metadata (P7 optimized)
- `item_values` - P6 filtering: Genre, Studio, Artist, CollectionType
- `user_data` - Per-user playstate: IsPlayed, IsFavorite, PlaybackPositionTicks
- `ancestors` - Hierarchical relationships (recursive CTE for tree traversal)

**Key Design Decisions:**
- **P7 Indexing**: TopParentId + Type + AncestorIds for folder queries
- **P6 Indexing**: ItemValues as separate lookup table with normalized values
- **User Context**: Every query optionally merges user_data (playstate)
- **Query Complexity**: Supports 80+ query parameters via InternalItemsQuery struct

---

### 3. Streaming Service (Port 8005)

**Responsibility:** Media streaming, Transcode job management, FFmpeg lifecycle

**Endpoints:**
- `GET /Videos/{itemId}/master.m3u8` - HLS master playlist - protected
- `GET /Videos/{itemId}/hls/{playlistId}/{segmentId}.ts` - HLS segments
- `GET /Videos/{itemId}/stream` - Progressive stream - protected
- `GET /Videos/{itemId}/stream.{container}` - Progressive by container

**Key Components:**
- `TranscodeManager` - FFmpeg process registry + kill timer (30s timeout)
- `StreamStateBuilder` - 48+ param decoding → codec/resolution/bitrate decision
- `PlaylistGenerator` - HLS manifest assembly (Master + Variant streams)
- `SegmentCleaner` - Removes old HLS segments to prevent disk explosion

**Key Design Decisions:**
- **Output Path**: `MD5(MediaPath-UserAgent-DeviceId-PlaySessionId)`
- **ActiveRequestCount**: Track concurrent segment requests (HLS only)
- **IsLiveOutput**: HLS playlists persist after disconnect (not cleaned)
- **Throttler**: Pause FFmpeg at keyframes when client paused (5min+ videos)
- **Log Files**: `FFmpeg.{Type}-{timestamp}_{MediaSourceId}_{guid8}.log`

---

## Data Flow Patterns

### Authentication Flow

```
Client Request
    ↓
X-Emby-Authorization header
    ↓
Auth Middleware: Parse Header (Token + DeviceId)
    ↓
Auth Service: Lookup Token → Get UserId + IsAdmin
    ↓
Populate Context (user_id, device_id, token, is_admin)
    ↓
Route Handler: Access Context Values
```

### Item Query Flow

```
GET /Items?TopParentId=xxx&IncludeItemTypes=Movie&OrderBy=DateCreated
    ↓
Middleware: Auth Check (optional for public items)
    ↓
Handler: Parse Query Params → InternalItemsQuery struct
    ↓
Repository: TranslateQuery() → SQL with P7/P6 indexes
    ↓
Execute Query → QueryResult[BaseItemDto]
    ↓
Serialize → Return JSON
```

### Streaming Flow (HLS)

```
GET /Videos/{itemId}/master.m3u8
    ↓
Middleware: Auth Check (REQUIRED)
    ↓
Handler: Build StreamState (48+ params)
    ↓
Decision: Transcode vs Direct Play?
    ↓
If Transcode:
    1. LockAsync(outputPath)
    2. StartFfMpeg(state, args)
    3. Register job with PlaySessionId
    4. WaitForMinimumSegments()
    5. Generate Master Playlist
If Direct Play:
    1. Return file reference
    ↓
Response: application/vnd.apple.mpegurl
```

### Streaming Flow (Progressive)

```
GET /Videos/{itemId}/stream?VideoCodec=h264
    ↓
Middleware: Auth Check (REQUIRED)
    ↓
Handler: Build StreamState
    ↓
Decision: Stream Copy vs Transcode?
    ↓
If Stream Copy:
    1. Return static file range request
If Transcode:
    1. LockAsync(outputPath)
    2. StartFfMpeg(state, args)
    3. Stream stdout directly to HTTP response
    ↓
Response: Media type based on container
```

---

## Inter-Service Communication

### Service Discovery

**Approach:** DNS-based (via Docker Compose networking)

```yaml
# docker-compose.yml
networks:
  kabletown:
    driver: bridge

# Service references via container name
auth-service:
  build: ./auth-service
  networks:
    - kabletown

item-service:
  build: ./item-service
  networks:
    - kabletown
  # Can reach auth-service at http://auth-service:8081
```

### Token Passing Model

**No internal service tokens** - each service validates X-Emby-Authorization independently

```
Client → Nginx → Auth Service
Client → Nginx → Item Service (with same token)
Client → Nginx → Streaming Service (with same token)
```

**Rationale:** Token is self-contained (just references api_keys table). Every service can validate it independently without inter-service RPC calls.

---

## Database Schema (Consolidated)

### MySQL 8.0 Shared Schema

```sql
-- Users (Auth Service)
CREATE TABLE users (
    Id CHAR(36) PRIMARY KEY,
    Username VARCHAR(100) UNIQUE NOT NULL,
    Email VARCHAR(255),
    PasswordHash CHAR(64) NOT NULL,
    IsAdmin BOOLEAN DEFAULT FALSE,
    EnableUser BOOLEAN DEFAULT TRUE,
    DateCreated TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    DateLastLogin TIMESTAMP NULL
);

-- API Keys (Auth Service)
CREATE TABLE api_keys (
    Id CHAR(36) PRIMARY KEY,
    UserId CHAR(36) NOT NULL,
    DeviceId CHAR(64) NOT NULL,
    Token CHAR(64) UNIQUE NOT NULL,
    Name VARCHAR(255),
    AppName VARCHAR(255),
    AppVersion VARCHAR(50),
    DateCreated TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    DateLastUsed TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    IsActive BOOLEAN DEFAULT TRUE,
    FOREIGN KEY (UserId) REFERENCES users(Id) ON DELETE CASCADE
);

-- Items (Item Service)
CREATE TABLE base_items (
    Id CHAR(36) PRIMARY KEY,
    Name VARCHAR(255) NOT NULL,
    Type VARCHAR(50) NOT NULL,
    TopParentId CHAR(36),
    Path VARCHAR(500),
    Container VARCHAR(100),
    DurationTicks BIGINT,
    Size BIGINT,
    DateCreated TIMESTAMP,
    DateModified TIMESTAMP,
    ExtraData JSON,  -- Provider IDs, custom fields
    INDEX idx_topparent_type (TopParentId, Type),
    INDEX idx_ancestors (AncestorIds)  -- Computed/trigger-maintained
);

-- ItemValues (P6) - Item Service
CREATE TABLE item_values (
    ItemId CHAR(36) NOT NULL,
    ValueType INT NOT NULL,  -- 2=Genre, 3=Studio, 4=Artist
    Value VARCHAR(255) NOT NULL,
    NormalizedValue VARCHAR(255) NOT NULL,  -- utf8mb4_unicode_ci
    PRIMARY KEY (ItemId, ValueType, NormalizedValue),
    INDEX idx_value_lookup (ValueType, NormalizedValue)
);

-- UserData (Item Service)
CREATE TABLE user_data (
    UserId CHAR(36) NOT NULL,
    ItemId CHAR(36) NOT NULL,
    IsPlayed BOOLEAN DEFAULT FALSE,
    IsFavorite BOOLEAN DEFAULT FALSE,
    PlaybackPositionTicks BIGINT DEFAULT 0,
    Rating INT,
    PRIMARY KEY (UserId, ItemId)
);

-- UserPreferences (Auth Service)
CREATE TABLE user_policies (
    UserId CHAR(36) PRIMARY KEY,
    EnableVideoTranscoding BOOLEAN DEFAULT TRUE,
    EnableAudioTranscoding BOOLEAN DEFAULT TRUE,
    MaxStreamingBitrate INT,
    FOREIGN KEY (UserId) REFERENCES users(Id) ON DELETE CASCADE
);
```

---

## Technology Stack

| Layer | Technology | Rationale |
| --- | --- | --- |
| **Language** | Go 1.22+ | Concurrency, performance, single binary deployment |
| **Web Framework** | Chi | Lightweight, middleware-friendly, no magic |
| **Database** | MySQL 8.0 | JSON support, window functions, recursive CTEs |
| **DB Driver** | go-sql-driver/mysql | Battle-tested, TLS support |
| **ORM** | sqlx | Lightweight query building, no automatic migrations |
| **Container** | Docker Compose | Simple orchestration for dev/test |
| **Reverse Proxy** | Nginx | TLS termination, load balancing, WebSocket support |
| **Media Processing** | FFmpeg | Industry standard transcoding |
| **Logging** | Log with JSON format | Structured logging for aggregation |

---

## API Design Principles

### 1. Consistent Error Format

```json
{
  "Message": "Invalid token",
  "StatusCode": 401
}
```

### 2. Query Parameters Follow Jellyfin Pattern

```json
/Items?
  TopParentId=abc123&
  IncludeItemTypes=Movie,Episode&
  OrderBy=DateCreated,Desc&
  StartIndex=0&
  Limit=20
```

### 3. Response Wrapping for Lists

```json
{
  "Items": [...],
  "TotalRecordCount": 150,
  "StartIndex": 0
}
```

### 4. Headers per Jellyfin Spec

- `X-MediaBrowser-Server-Id`: GUID identifying this server instance
- `X-MediaBrowser-Token`: API key (alternative to Authorization header)
- `Content-Type`: `application/json; charset=utf-8`

### 5. UUID Format

All identifiers use GUID format:
- Input: Dashed UUIDs (`550e8400-e29b-41d4-a716-446655440000`)
- Storage: CHAR(36) with dashes (preserve Jellyfin compatibility)
- Query: `WHERE Id = '550e8400-e29b-41d4-a716-446655440000'`

---

## Security Model

### Authentication

1. **Header Format:** `X-Emby-Authorization: MediaBrowser Token="abc123", DeviceId="def456" ...`
2. **Token Lifetime:** Indefinite (until revoked)
3. **Revocation:** User manually deletes device via UI

### Authorization

1. **Middleware Pattern:** Every protected route uses `AuthMiddleware`
2. **Context Values:** `user_id`, `device_id`, `is_admin`
3. **Role Checks:** Handler-level `if !isAdmin { return 403 }`

### Data Protection

1. **Password Hashing:** SHA256 (replace with argon2/bcrypt in production)
2. **TLS:** Handled by Nginx reverse proxy
3. **CORS:** Disabled (same-origin via Nginx)

---

## Deployment Architecture

```
┌─────────────────────────────────────────────────────┐
│                  Client Browser                     │
│              (iOS/Android/Chrome)                   │
└─────────────────────┬───────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────┐
│              Nginx Reverse Proxy                    │
│         (kabletown-nginx:8096)                      │
│  ┌─────────────────────────────────────────────┐  │
│  │ /auth/*   → auth-service:8081              │  │
│  │ /items/*  → item-service:8004              │  │
│  │ /videos/* → streaming-service:8005         │  │
│  └─────────────────────────────────────────────┘  │
└─────────────────────┬───────────────────────────────┘
                      │
┌─────────────────────┼───────────────────────────────┐
│                     │                               │
▼                     ▼                               ▼
┌─────────────┐ ┌─────────────┐ ┌─────────────┐      ┌─────────────┐
│ auth-service│ │ item-service│ │ streaming   │      │   MySQL     │
│  :8081      │ │  :8004      │ │  service    │      │  :3306      │
│             │ │             │ │  :8005      │      │             │
└─────────────┘ └─────────────┘ └─────────────┘      └─────────────┘
                      │                               │
                      ▼                               │
              ┌─────────────┐                         │
              │ FFmpeg      │                         │
              │ :N/A        │                         │
              │ (subprocess)│                         │
              └─────────────┘                         │
                      │                               │
                      ▼                               │
              ┌─────────────┐                         │
              │ /transcode  │                         │
              │ (volume)    │                         │
              └─────────────┘                         │
                                                     │
┌─────────────────────────────────────────────────────┤
│                    Docker Host                      │
│                  (Ubuntu 22.04)                     │
└─────────────────────────────────────────────────────┘
```

---

## Implementation Phases

### Phase 1: Architecture Blueprint (Current)
- ✅ Auth middleware pattern
- ✅ Service boundaries defined
- ✅ Database schema designed
- ⬜ FFmpeg process management documented
- ⬜ Docker Compose setup

### Phase 2: Auth Service MVP
- ⬜ Implement device registration
- ⬜ Implement user authentication
- ⬜ Implement token validation
- ⬜ Add password hashing (argon2)

### Phase 3: Item Service Core
- ⬜ Implement InternalItemsQuery parser
- ⬜ Add P7 indexes (TopParentId, Type, AncestorIds)
- ⬜ Implement BaseItemDto serialization
- ⬜ Add user_data joins

### Phase 4: Streaming Service Core
- ⬜ Implement TranscodeManager (FFmpeg process registry)
- ⬜ Implement StreamStateBuilder (48+ params)
- ⬜ Add HLS playlist generation
- ⬜ Add segment cleanup

### Phase 5: Integration & Testing
- ⬜ Integration tests for all services
- ⬜ Load testing for transcoding
- ⬜ Security audit

---

## Open Technical Questions

1. **Token Storage:** Should tokens be hashed or stored plaintext?
   - *Decision:* Store plaintext (reversible by nature, encrypted at rest via DB encryption)
   
2. **Inter-service Auth:** Should token validation hit auth-service or query shared DB?
   - *Decision:* Shared DB (avoids RPC chain, token is just lookup key)
   
3. **FFmpeg Process Management:** How many concurrent transcodes?
   - *Decision:* Start with 2 (configurable per-user policy)
   
4. **CDN / Edge Caching:** Should HLS segments be cacheable?
   - *Decision:* Add cache-control headers (4s TTL per segment)
