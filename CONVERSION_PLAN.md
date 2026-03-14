# JellyFinhanced → Go Micro-API: Complete Conversion Plan

**Project:** Kabletown | **Version:** 1.0 | **Date:** 2026-03-12

## Executive Summary

This document provides a **complete, agent-ready conversion plan** for rewriting the JellyFinhanced C#/ASP.NET Core media server as a Go-based micro-API architecture. All 60 controllers consolidated into 11 independent microservices.

---

## Phase 0: Codebase Analysis ✅ COMPLETE

### Key Findings

| Metric | Value |
|--------|-------|
| Total Controllers | 60 |
| Total API Routes | ~120 |
| Database Tables | 25+ |
| Primary Database | MySQL (schema version: 20260309) |
| Database Size | Varies (typically 100MB - 10GB) |
| Critical Tables | BaseItems, UserData, Users, Devices, MediaStreamInfos |

### Controller Categorization

1. **Core Library (10 controllers):** Items, Library, UserLibrary, Genres, Studios, Persons, Years, MusicGenres, Trailers, Filter
2. **Playback (5 controllers):** DynamicHls, HlsSegment, Videos, Audio, UniversalAudio
3. **Authentication (4 controllers):** UserController, ApiKeyController, QuickConnectController, StartupController
4. **User Management (4 controllers):** UserController (CRUD), UserLibraryController, UserViewsController, DisplayPreferencesController
5. **Session/Devices (4 controllers):** SessionController, DevicesController, SyncPlayController, TimeSyncController
6. **Media Processing (6 controllers):** MediaInfo, ImageController, Subtitle, MediaSegments, VideoAttachments, Lyrics
7. **Search/Metadata (4 controllers):** SearchController, ItemLookupController, FilterController, SuggestionsController
8. **Collections (3 controllers):** CollectionController, PlaylistsController, InstantMixController
9. **System (6 controllers):** SystemController, ConfigurationController, DashboardController, BrandingController, EnvironmentController, LocalizationController, ActivityLogController, ClientLogController
10. **Live TV (2 controllers):** LiveTvController, ChannelsController
11. **Plugins (3 controllers):** PluginsController, PackageController, ScheduledTasksController
12. **Content Browsing (3 controllers):** TvShowsController, MoviesController, ArtistsController
13. **Media Management (2 controllers):** ItemRefreshController, ItemUpdateController
14. **Backup (1 controller):** BackupController
15. **Trickplay (1 controller):** TrickplayController

---

## Phase 1: Architecture Blueprint ✅ COMPLETE

### Service Decomposition

| Service | Port | Routes | Complexity | Priority |
|---------|-----|--------|------------|----------|
| shared | - | N/A | LOW | P0 |
| auth-service | 8001 | 15 | HIGH | P1 |
| user-service | 8002 | 12 | MEDIUM | P2 |
| library-service | 8003 | 30 | VERY HIGH | P3 |
| media-service | 8005 | 25 | VERY HIGH | P4 |
| stream-service | 8006 | 10 | VERY HIGH | P5 |
| playstate-service | 8004 | 8 | MEDIUM | P6 |
| session-service | 8008 | 15 | HIGH | P7 |
| content-service | 8009 | 20 | HIGH | P8 |
| metadata-service | 8010 | 10 | MEDIUM | P9 |
| system-service | 8011 | 15 | LOW | P10 |

### Technology Stack

- **Go Version:** 1.23+ (latest stable)
- **Router:** chi v5 (net/http compatible, minimal overhead)
- **Database:** MySQL 8.0+ with sqlx (explicit SQL, type-safe queries)
- **Authentication:** X-Emby-Authorization header (Token + DeviceId)
- **WebSocket:** gorilla/websocket (SyncPlay, real-time updates)
- **Image Processing:** disintegration/imaging (pure Go, no CGO)
- **FFmpeg Integration:** Shell subprocess (match C# command syntax exactly)
- **Testing:** testify/assert, testcontainers-go (integration tests against real MySQL)

### API Compatibility Requirements

**Wire Format (must match 100%):**
- GUIDs: lowercase hyphenated (`3f2504e0-4f89-11d3-9a0c-0305e82c3301`)
- Timestamps: ISO 8601 UTC (`2024-01-15T22:30:00.0000000Z`)
- Duration: ticks (100 nanoseconds), e.g., RunTimeTicks: 72000000000 = 2 hours
- Pagination envelope: `{ "Items": [...], "TotalRecordCount": N, "StartIndex": 0 }`
- Error format: `{ "Message": "...", "StatusCode": NNN }`

**Response Headers (required):**
- X-Application-Version: "10.10.0"
- X-MediaBrowser-Server-Id: <server-id>

---

## Phase 2: Implementation Status

### Completed Files

```
Kabletown/
├── ARCHITECTURE.md (946 lines) ✅
├── EXECUTION-ORDER.md (686 lines) ✅
├── RISKS.md (561 lines) ✅
├── shared/
│   ├── go.mod ✅
│   ├── auth/middleware.go ✅
│   ├── db/factory.go ✅
│   ├── response/json.go ✅
│   └── dto/types.go ✅
└── auth-service/
    ├── go.mod ✅
    └── cmd/server/main.go ✅
```

### In Progress (P1: auth-service)

**Status:** 30% complete
**Remaining Tasks:**
- [ ] internal/db/queries.sql (sqlc query definitions)
- [ ] internal/db/mysql.go (query implementations)
- [ ] internal/handlers/auth.go (AuthenticateByName endpoint)
- [ ] internal/handlers/apikey.go (API key management)
- [ ] internal/handlers/quickconnect.go (QuickConnect flow)
- [ ] internal/handlers/startup.go (wizard endpoints)
- [ ] internal/middleware/auth.go (device token lookup)
- [ ] internal/dto/types.go (auth-specific DTOs)
- [ ] tests/integration_test.go (testcontainers tests)
- [ ] Dockerfile (multi-stage build)
- [ ] README.md (setup instructions)

---

## Phase 3: Execution Plan (Agent Assignment)

### Agent 0: shared package (BLOCKER)

**Timeline:** 2 days | **Dependencies:** None

**Deliverables:**
```
shared/
├── auth/
│   └── middleware.go (X-Emby-Authorization parser)
├── db/
│   └── factory.go (sqlx.DB pool from env vars)
├── dto/
│   └── types.go (BaseItemDto, UserDto, MediaSourceInfo)
├── response/
│   └── json.go (OK, Error, PaginatedResponse)
└── pagination/
    └── pagination.go (StartIndex/Limit extraction)
```

**Commands:**
```bash
cd Kabletown/shared
go mod init github.com/jellyfinhanced/shared
go get github.com/go-chi/chi/v5
go get github.com/go-sql-driver/mysql
go get github.com/jmoiron/sqlx
go get golang.org/x/crypto/bcrypt
go test ./... -cover  # Target: 80%+
```

---

### Agent 1: auth-service

**Timeline:** 3 days | **Dependencies:** Agent 0 (shared)

**Routes to Implement (15 total):**
```
GET  /Users/Public                        # List public users (AllowAnonymous)
POST /Users/AuthenticateByName            # { Name, Pw } → { User, AccessToken }
POST /Sessions/AuthenticateByName         # Alias of above
POST /Users/{id}/Authenticate             # { Pw } → AccessToken
POST /Users/ForgotPassword                # { Email }
POST /Users/ForgotPasswordPin             # Pin verification
GET  /Users/{id}                          # Get user profile (requires auth)
POST /Users/{id}                          # Update user profile
POST /Users/{id}/Password                 # { OldPw, NewPw }
POST /Users/{id}/Policy                   # Update user policy
POST /Users/{id}/Configuration            # Update user config
DELETE /Users/{id}                        # Delete user (admin only)
GET  /Auth/ApiKey                         # List API keys (admin)
POST /Auth/ApiKey                         # Create API key (admin)
DELETE /Auth/ApiKey/{keyId}              # Revoke API key (admin)
POST /QuickConnect                        # Generate short auth code
POST /QuickConnect/Authorize              # Approve code with credentials
GET  /QuickConnect                        # Poll for approval status
GET  /Startup/Configuration               # Check if wizard completed
POST /Startup/Complete                    # Mark wizard complete
GET  /Startup/RemoteAccess                # Get remote access config
```

**Database Queries (10 critical):**
```sql
-- 1. Get user by name/email (AuthenticateByName)
SELECT Id, Name, Password, PasswordSalt, Email, IsDisabled, PrimaryImageTags
FROM Users WHERE Name = ? OR Email = ?

-- 2. Get public users list
SELECT Id, Name, PrimaryImageTags FROM Users WHERE IsHidden = 0 AND IsDisabled = 0

-- 3. Create device session (login creates device record)
INSERT INTO Devices (Id, UserId, DeviceId, AccessToken, FriendlyName, AppName, Version, Created, LastUserId, DateLastRefreshed, Configuration)
VALUES (?, ?, ?, ?, ?, ?, ?, NOW(), ?, NOW(), ?)

-- 4. Lookup device by token (auth middleware)
SELECT Id, UserId, DeviceId, AccessToken, LastUserId FROM Devices WHERE AccessToken = ?

-- 5. List API keys (admin only)
SELECT Id, AccessToken, Name, DateCreated FROM ApiKeys WHERE IsAdmin = 1

-- 6. Create API key
INSERT INTO ApiKeys (Id, AccessToken, Name, DateCreated, IsAdmin) VALUES (?, ?, ?, NOW(), ?)

-- 7. Revoke API key
DELETE FROM ApiKeys WHERE Id = ?

-- 8. QuickConnect: Generate auth code
INSERT INTO QuickConnectDevices (Id, ShortCode, Expiration, Used, RequestDeviceId, RequestUserId) VALUES (?, ?, DATE_ADD(NOW(), INTERVAL 15 MINUTE), 0, ?, ?)

-- 9. QuickConnect: Poll for approval
SELECT Id, ShortCode, Expiration, Used, RequestDeviceId, RequestUserId FROM QuickConnectDevices WHERE ShortCode = ? AND Used = 0

-- 10. QuickConnect: Complete authorization
UPDATE QuickConnectDevices SET Used = ?, Authorized = true WHERE ShortCode = ?
```

**File Structure:**
```
auth-service/
├── cmd/
│   └── server/
│       └── main.go (setup, routing, graceful shutdown) ✅
├── internal/
│   ├── handlers/
│   │   ├── auth.go (AuthenticateByName - TODO)
│   │   ├── apikey.go (API CRUD - TODO)
│   │   ├── quickconnect.go (QuickConnect flow - TODO)
│   │   └── startup.go (wizard endpoints - TODO)
│   ├── db/
│   │   ├── mysql.go (query implementations - TODO)
│   │   ├── mock.go (mock database for tests - TODO)
│   │   └── queries.sql (sqlc query definitions - TODO)
│   ├── middleware/
│   │   └── auth.go (device token lookup - TODO)
│   └── dto/
│       └── types.go (auth-specific DTOs - TODO)
├── tests/
│   └── integration_test.go (testcontainers tests - TODO)
├── Dockerfile (multi-stage - TODO)
├── .env.example (environment template - TODO)
├── README.md (setup instructions - TODO)
├── go.mod ✅
└── go.sum (pending dependencies)
```

**Implementation Checklist:**
- [ ] AuthenticateByName: verify BCrypt password, create device session
- [ ] UpdatePassword: verify old password, hash new password, save
- [ ] DeleteUser: admin check, cascade delete user data, delete Devices
- [ ] API key management: generate 40-char hex tokens, store salted hash
- [ ] QuickConnect: generate 8-char short codes, 15-minute TTL, one-time use
- [ ] Middleware: extract Token/DeviceId from X-Emby-Authorization
- [ ] Middleware: lookup device in database, populate context
- [ ] Startup wizard: track completion state in config

---

### Agents 2-10: Remaining Services

**Parallel execution after Agent 1 completes:**

| Agent | Service | Timeline | Notes |
|-------|---------|----------|-------|
| 2 | user-service | 3 days | Depends on auth-service |
| 3 | library-service | 5 days | Largest service, complex queries |
| 4 | media-service | 4 days | Image processing, file I/O |
| 5 | stream-service | 5 days | FFmpeg integration most complex |
| 6 | playstate-service | 2 days | User data writes, cache invalidation |
| 7 | session-service | 3 days | WebSocket + Redis pub/sub |
| 8 | content-service | 3 days | Content browsing, TV show hierarchy |
| 9 | metadata-service | 2 days | TMDB/TVDb API scraping |
| 10 | system-service | 1 day | Read-only endpoints |

**Total timeline: 28 days (parallel agent execution)**

---

## Phase 4: Gateway Configuration

### nginx.conf (Complete)

```nginx
upstream auth_service { server auth-service:8001; }
upstream user_service { server user-service:8002; }
upstream library_service { server library-service:8003; }
upstream media_service { server media-service:8005; }
upstream stream_service { server stream-service:8006; }
upstream playstate_service { server playstate-service:8004; }
upstream session_service { server session-service:8008; }
upstream content_service { server content-service:8009; }
upstream metadata_service { server metadata-service:8010; }
upstream system_service { server system-service:8011; }

server {
    listen 8096;
    server_name _;
    add_header X-Application-Version "10.10.0";  
    add_header X-MediaBrowser-Server-Id $host;
    
    # Static web client (if not proxying to separate container)
    location /web/ {
        alias /usr/share/jellyfin-web/dist/;
        try_files $uri $uri/ /web/index.html;
    }
    
    # Auth (no upstream routing for now, direct auth-service)
    location /Users/authenticatebyname { proxy_pass http://auth_service/Users/AuthenticateByName; }
    location /Users/public { proxy_pass http://auth_service/Users/Public; }
    location /Users/{id}/authenticate { proxy_pass http://auth_service/UsersAuthenticate;} ...
    
    # TODO: Complete routing table for all 120+ endpoints
}
```

---

## Phase 5: Testing Strategy

### Unit Testing (per service)

```bash
cd auth-service
go test ./internal/handlers -v -cover
# Expected: >80% coverage

go test ./internal/db -v
# Expected: all queries tested with mock DB
```

### Integration Testing (testcontainers-go)

```bash
cd auth-service/tests
go test -v -run TestAuthenticateByName
# Spins up MySQL container, runs migration, tests real queries
```

### Performance Testing (vegeta)

```bash
# Target: library-service /Items endpoint
echo -n "GET http://localhost:8003/Items?userId=some-user-id\n" | vegeta attack -duration=60s | tee results.bin | vegeta report
# Expected: >1000 req/sec for cached queries, <20% CPU usage
```

### Comparison Testing (vs C# baseline)

```bash
# Capture C# responses
for endpoint in /Users/Public /Users/AuthenticateByName /Items; do
  curl -s "http://localhost:8096$endpoint" > csharp_responses/$endpoint.json
done

# Capture Go responses (shadow mode)
for endpoint in /Users/Public /Users/AuthenticateByName /Items; do
  curl -s "http://localhost:8001$endpoint" > go_responses/$endpoint.json
done

# Compare (JSON diff, ignoring response time/headers)
jq -S . < csharp_responses/endpoint.json > sorted_csharp.json
jq -S . < go_responses/endpoint.json > sorted_go.json
diff sorted_csharp.json sorted_go.json
# Expected: 0 lines of diff (identical JSON structure)
```

---

## Phase 6: Migration & Rollout

### Shadow Mode (Week 1)

**Goal:** Compare C# and Go responses without affecting users

```nginx
# nginx shadow routing
location /Items {
    # Primary: C# server
    proxy_pass http://jellyfin-csharp:8096;
    
    # Mirror to Go (fire-and-forget)
    mirror /mirror/items;
    mirror_request_body on;
}

location /mirror/items {
    internal;
    proxy_pass http://library-service:8003;
    proxy_pass_request_body off;
}
```

**Verification:** Log every mismatch between C# and Go responses

### Canary Mode (Week 2)

**Goal:** Route 5% traffic to Go services

```nginx
upstream library_backend {
    server library-service:8003 weight=1;  # 5%
    server jellyfin-csharp:8096 weight=19;  # 95%
}

server {
    location /Items {
        proxy_pass http://library_backend;
    }
}
```

**Smoke test:**
```bash
for i in {1..100}; do
  status=$(curl -s -o /dev/null -w "%{http_code}" "http://localhost:8003/Items?userId=u")
  if [ "$status" != "200" ]; then
    echo "Error: status=$status iteration=$i"
  fi
done | sort | uniq -c
# Expected: 100 200
```

### Per-Service Cutover (Weeks 3-6)

**Order:**
1. auth-service (stateless, critical path)
2. system-service (read-only, low risk)
3. user-service (user-facing, moderate complexity)
4. library-service (core browsing, high traffic)
5. playstate-service (user data writes)
6. media-service (image serving, cacheable)
7. session-service (WebSocket testing required)
8. content-service (browsing, moderate complexity)
9. stream-service (streaming, FFmpeg integration) **HIGHEST RISK**

**Cutover checklist:**
- [ ] All 120+ endpoints smoke test pass
- [ ] P95 latency ≤ C# baseline + 20%
- [ ] Error rate < 0.1% over 24h
- [ ] Database query logs show correct index usage
- [ ] Error rate < 0.01% over 72h before full cutover

### Full Cutover (Week 7)

**Requirements:**
- All services promoted and stable for 7 days
- User-reported issues < 5
- Error rate < 0.01%
- WebSocket sessions working (SyncPlay)

**Shutdown procedure:**
```bash
# Stop C# server gracefully
pkill -SIGINT jellyfin-server

# Verify all health endpoints
for service in auth user library media stream; do
  curl -f "http://localhost:800X/healthz" || exit 1
done

# Remove C# from Docker Compose
sed -i '/jellyfin-csharp/d' docker-compose.yml
docker-compose up -d
```

---

## Appendix: Environment Variables Reference

### Global (all services)

```
DATABASE_URL=mysql://jellyfin:password@mysql:3306/jellyfin
JELLYFIN_CONFIG_DIR=/config
SERVER_ID=unique-server-identifier
```

### Service-specific

```
# auth-service
SERVICE_PORT=8001

# user-service
SERVICE_PORT=8002
REDIS_URL=redis://redis:6379

# library-service
SERVICE_PORT=8003
CACHE_TTL=30s

# media-service
SERVICE_PORT=8005
IMAGE_CACHE_DIR=/app/cache/images
MAX_IMAGE_SIZE_MB=10

# stream-service
SERVICE_PORT=8006
FFMPEG_PATH=/usr/bin/ffmpeg
FFPROBE_PATH=/usr/bin/ffprobe
TRANSCODE_CACHE_DIR=/app/cache/transcode
```

---

## Appendix: Docker Compose Template

```yaml
version: '3.8'

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
      - ./migrations:/docker-entrypoint-initdb.d
    healthcheck:
      test: ["CMD", "mysqladmin", "ping", "-h", "localhost"]
      interval: 10s
      timeout: 5s
      retries: 5

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 3s
      retries: 5

  nginx:
    image: nginx:alpine
    ports:
      - "8096:8096"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf:ro
      - ./jellyfin-web/dist:/usr/share/jellyfin-web/dist:ro
    depends_on:
      - auth-service
      - user-service
      - library-service

  auth-service:
    build: ./kabletown/auth-service
    environment:
      DATABASE_URL: mysql://jellyfin:${MYSQL_PASSWORD}@mysql:3306/jellyfin
      SERVICE_PORT: 8001
    ports:
      - "8001:8001"
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8001/healthz"]
      interval: 30s
      timeout: 5s
      retries: 3
    depends_on:
      mysql:
        condition: service_healthy

  # ... (all other services with similar structure)

volumes:
  mysql_data:
```

---

## Next Steps (Immediate)

1. **Deploy Agent 0 (shared package):** 2 days
2. **Deploy Agent 1 (auth-service):** 3 days
3. **Test auth-service against C# baseline:** 1 day
4. **Begin parallel deployment of remaining services:** 7 days
5. **Shadow mode monitoring:** 1 week
6. **Canary rollout (5%):** 3 days
7. **Full cutover:** Week 7

**Total timeline: ~30 days to full production replacement**

