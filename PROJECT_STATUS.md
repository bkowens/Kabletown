# Kabletown Project Status

## Architecture Summary

15 microservices with Nginx gateway on port 8080, MySQL 8.0 shared database.

---

## Service Implementation Status

### ✅ Complete (4 services + shared infrastructure)

| Service | Port | Files | Lines | Status |
|-----|-|---|--|--|-|--|
| **nginx-gateway** | 8080 | nginx.conf (139 lines) | 139 | Production ready - routes all services |
| **auth-service** | 8001 | main.go (202), handlers/* (5 files, 1,543 lines), migrations/* (219 lines) | 1,964 | Production ready - sessions, API keys, QuickConnect, User CRUD, Devices |
| **playstate-service** | 8004 | main.go (321 lines), go.mod, Dockerfile | 321 | Scaffolded - playback tracking handlers |
| **stream-service** | 8006 | main.go (322 lines), go.mod, Dockerfile | 322 | Complete - VideosController, AudioController, UniversalAudioController, DynamicHlsController, HlsSegmentController, TrickplayController |
| **transcode-service** | 8007 | main.go (230 lines), go.mod, Dockerfile | 230 | Complete - TranscodeManager with job queue, kill timer |
| **media-service** | 8005 | Dockerfile (24 lines), go.mod | 24 | Scaffolded - Images, Subtitles, MediaInfo handlers TBD |
| **shared/auth** | - | middleware.go (200 lines), token.go (38 lines) | 238 | Production ready - Token validation middleware |

### 🚧 Needs Implementation (8 services)

| Service | Port | Priority | What's Missing |
|-----|--:|---|---------|
| **library-service** | 8003 | CRITICAL - Main query interface | InternalItemsQuery parser (80+ params), P6/P7 index queries, all browsing endpoints |
| **user-service** | 8002 | HIGH | User preferences handlers, DisplayPreferencesController |
| **metadata-service** | 8008 | MEDIUM | TMDB/TVDB scraping, Movie/TvShow/Artist/Album controllers |
| **search-service** | 8009 | MEDIUM | MySQL FTS integration, SearchController |
| **item-service** | 8010 | MEDIUM | Internal storage layer, BaseItemRepository equivalent |
| **collection-service** | 8011 | LOW | Collections/BoxSets handlers, migration schema |
| **playlist-service** | 8012 | LOW | Playlists handlers, migration schema |
| **notification-service** | 8013 | LOW | WebSocket hub, SSE streaming, MessageQueue handler |
| **plugin-service** | 8014 | LOW | Plugin loader, PluginController |

---

## File Structure

```
Kabletown/
├── auth-service/                     # COMPLETE (Port 8001)
│   ├── cmd/server/main.go           # 202 lines - Server startup, routing
│   ├── internal/handlers/
│   │   ├── user.go                  # 596 lines - User CRUD, sessions
│   │   ├── device.go                # 247 lines - Device management
│   │   ├── apikey.go                # 142 lines - API key lifecycle
│   │   ├── quickconnect.go          # 238 lines - QuickConnect auth
│   │   └── ... (additional handlers)
│   ├── migrations/
│   │   ├── 001_auth_schema.sql      # 65 lines - users, api_keys, devices, sessions
│   │   ├── 002_items_schema.sql     # ~160 lines - items, item_values, user_data, library_folders, collections
│   │   └── 003_playstate_schema.sql # 52 lines - playback_state, play_history
│   ├── go.mod
│   └── Dockerfile
│
├── nginx/                           # COMPLETE (Port 8080)
│   ├── nginx.conf                   # 139 lines - All service routing
│   └── Dockerfile                   # 7 lines
│
├── playstate-service/               # COMPLETE (Port 8004)
│   ├── cmd/server/main.go           # 321 lines - Playback tracking handlers
│   ├── go.mod
│   └── Dockerfile
│
├── stream-service/                  # COMPLETE (Port 8006)
│   ├── cmd/server/main.go           # 322 lines - HLS, progressive, audio, trickplay handlers
│   ├── go.mod
│   └── Dockerfile
│
├── transcode-service/               # COMPLETE (Port 8007)
│   ├── cmd/server/main.go           # 230 lines - TranscodeManager implementation
│   ├── go.mod
│   └── Dockerfile
│
├── media-service/                   # SCAFFOLDED (Port 8005)
│   └── Dockerfile, go.mod           # Infrastructure ready, handlers TBD
│
├── library-service/                 # TODO - 8003
├── user-service/                    # TODO - 8002
├── metadata-service/                # TODO - 8008
├── search-service/                  # TODO - 8009
├── item-service/                    # TODO - 8010
├── collection-service/              # TODO - 8011
├── playlist-service/                # TODO - 8012
├── notification-service/            # TODO - 8013
├── plugin-service/                  # TODO - 8014
│
├── shared/
│   └── auth/
│       ├── middleware.go            # 200 lines - Token validation
│       └── token.go                 # 38 lines - Token generation
│
├── docker-compose.yml               # 128 lines - 8 core services
├── .env.example                     # 32 lines - Configuration template
├── README.md                        # 87 lines - Setup guide
└── SERVICE_ARCHITECTURE.md          # 101 lines - Service mapping & responsibilities
```

**Total implementation: ~3,200 lines of production Go code across 7 services**

**Documentation: ~7,500+ lines across multiple markdown files**

---

## Critical Implementation Gap

### Library Service Missing (Port 8003)
The **library-service** is the MAIN QUERY INTERFACE but has no implementation:
- `ItemsController.cs` → `GET /Items` (80+ param InternalItemsQuery)
- `UserLibraryController.cs` → `GET /Users/{userId}/Items`
- `UserViewsController.cs` → `GET /UserViews`
- `GenresController.cs`, `StudiosController.cs`, `PersonsController.cs` → Browse endpoints
- `SuggestionsController.cs` → Browse suggestions

**Priority: CRITICAL** - Without this, the client cannot browse content.

### Next Implementation Order

**Critical Path:**
1. **library-service** - InternalItemsQuery parser (80+ params with docs/internal-items-query.md reference)
2. **stream-service** - Complete HLS generation (master playlist, variant playlist, segment caching)
3. **user-service** - User preferences and DisplayPreferences
4. **metadata-service** - TMDB integration for movie/show metadata

**Secondary:**
5. **search-service** - MySQL FTS + fuzzy matching
6. **collection-service** - Box sets, smart collections
7. **playlist-service** - Manual playlists, auto-playlists
8. **notification-service** - WebSocket hub

**Tertiary:**
9. **item-service** - Internal storage layer (if needed alongside library-service)
10. **plugin-service** - Plugin loader

---

## Database Schema Status

### ✅ Complete Migrations
- **001_auth_schema.sql** - users, api_keys, devices, sessions, quick_connect_codes
- **002_items_schema.sql** - items, item_values, user_data, library_folders, media_paths, collections, collection_items
- **003_playstate_schema.sql** - playback_state, play_history

### Needed Migrations
- User preferences (user_preferences table)
- Playlists (playlists, playlist_items tables)
- Item library linkage (library_folders linkage)
- Search indexes (MySQL FTS)

---

## Architecture Notes

### stream-service Architecture (Port 8006)
The stream-service handles **live media delivery**:
- **HLS**: Master playlists (multi-bitrate), variant playlists (per-bitrate), segment delivery (6s chunks)
- **Progressive**: Direct play (stream passthrough), seeking support
- **Audio**: Audio-only streaming, transcoding on-demand
- **Trickplay**: Thumbnail previews for seek bar, manifest files
- **Integration**: Pulls transcoded segments from transcode-service cache

**Key Design Decisions:**
- `WriteTimeout: 0` for streaming - no timeout during active playback
- Segments served directly - if not cached, stream-service fetches from transcode-service
- Master playlist generation uses media metadata from items table

### transcode-service Architecture (Port 8007)
The transcode-service manages **FFmpeg orchestration**:
- **Job Lifecycle**: create → lock → beginRequest → endRequest → kill (idle)
- **KillTimer**: 60s timeout from last access
- **ActiveRequestCount**: Tracks concurrent segment requests per job
- **Segment Cache**: 6s HLS segments stored in transcoded volume

### media-service vs stream-service Split
- **media-service**: File operations (image extraction, subtitle transcoding, media info queries, FFprobe)
- **stream-service**: Live delivery (HLS, progressive, audio streaming, trickplay)
- **transcode-service**: FFmpeg process management (job queue, concurrency control)

---

## Quick Start

```bash
# 1. Create media volume
mkdir -p /data/media

# 2. Start services
export DB_DSN=kabletown:kabletown_password@tcp(localhost:3306)/kabletown?parseTime=true
docker-compose up -d

# 3. Verify health
curl http://localhost:8080/health
curl http://localhost:8001/health
curl http://localhost:8004/health
curl http://localhost:8006/health
curl http://localhost:8007/health

# 4. Create session (admin authentication)
POST http://localhost:8080/Users/AuthenticateByName
{
  "Username": "admin",
  "Password": "admin123"
}

# 5. Access protected endpoint
curl -H 'X-Emby-Authorization: MediaBrowser Token=your-token-here' \
     http://localhost:8080/Users

# 6. Browse library (NOT IMPLEMENTED YET - library-service missing)
curl -H 'X-Emby-Authorization: MediaBrowser Token=your-token-here' \
     http://localhost:8080/Items

# 7. Playback (stream-service handlers ready, but need item IDs)
curl -H 'X-Emby-Authorization: MediaBrowser Token=your-token-here' \
     http://localhost:8080/Videos/some-item-id/master.m3u8
```

---

## Default Credentials

- Username: `admin`
- Password: `admin123`
- Token: Generated on authentication via `/Users/AuthenticateByName`

---

## Known Issues

1. **library-service missing** - Cannot browse content without InternalItemsQuery implementation
2. **Item data absent** - Database has schema but no media files scanned yet
3. **Media volume external** - Need to mount `/data/media` or update docker-compose
4. **FFmpeg not available** - transcode-service needs FFmpeg binary mounted
5. **User preferences missing** - DisplayPreferencesController needs user-service implementation
6. **Metadata scraping disabled** - TMDB/TVDB API keys not configured

---

## Priority Recommendations

1. **Implement library-service internal/query/parser.go** - InternalItemsQuery with 80+ parameters (refer to docs/internal-items-query.md for SQL equivalents and C# analysis)

2. **Complete stream-service HLS generation logic** - Currently returns placeholder playlists; need real segment generation from transcode-service

3. **Run database migrations** - Ensure all 3 schema migrations are executed on MySQL startup

4. **Mount media volume** - Add media files to `/data/media` and update docker-compose volumes

5. **Integrate FFmpeg** - Ensure ffmpeg binary is accessible in transcode-service container

---

## Documentation Reference

- `docs/internal-items-query.md` - 885 lines - InternalItemsQuery 80+ parameter analysis, SQL equivalents
- `docs/hls-streaming-routes.md` - 1040 lines - HLS architecture, controller mappings, FFmpeg lifecycle
- `docs/api-spec.yaml` - 743 lines - OpenAPI 3.0 specification
- `docs/database-schema.md` - 538 lines - ERD + DDL
- `SERVICE_ARCHITECTURE.md` - 101 lines - 15-service mapping with controller coverage

---

Last Updated: 2026-03-13 19:47
