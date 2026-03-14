# Kabletown Service Architecture

## 15-Service Microservices Map

| # | Service Name | Port | Primary Routes | Controllers |
|---|---|---|---|---|
| 1 | `gateway` | 8080 | All external traffic | Nginx API Gateway |
| 2 | `auth-service` | 8001 | `/Auth`, `/Users/AuthenticateByName` | ApiKeyController, QuickConnectController, UserController (CRUD), DeviceController, SessionController (auth-related) |
| 3 | `user-service` | 8002 | `/Users`, `/DisplayPreferences` | UserController (preferences), DisplayPreferencesController |
| 4 | `library-service` | 8003 | `/Items`, `/UserLibrary`, `/UserViews`, `/Genres`, `/Studios`, `/Persons` | ItemsController (80+ params InternalItemsQuery), SearchController, ArtistsController (browse), AlbumsController, YearsController, SuggestionsController, FiltersController |
| 5 | `playstate-service` | 8004 | `/Playstate` | PlaystateController (playback start/stop/playing) |
| 6 | `media-service` | 8005 | `/Images`, `/MediaInfo`, `/Subtitles` | ImageController, MediaInfoController, SubtitleController, AttachmentController, LyricController |
| 7 | `stream-service` | 8006 | `/Videos`, `/Hls`, `/Audio`, `/Trickplay` | VideosController, AudioController, HlsController, TrickplayController |
| 8 | `session-service` | 8007 | `/Sessions`, `/Devices`, `/SyncPlay` | SessionController, DevicesController, SyncPlayController |
| 9 | `metadata-service` | 8008 | `/Metadata`, `/ScheduledTasks`, `/Items/Refresh`, `/Items/Lookup`, `/Items/{id}/Images/Remote` | ItemRefreshController, ItemUpdateController, ItemLookupController, RemoteImageController, ScheduledTasksController |
| 10 | `content-service` | 8009 | `/Movies`, `/TvShows`, `/Albums`, `/Artists`, `/Songs`, `/Collections`, `/Playlists`, `/Channels`, `/LiveTv` | MoviesController, TvShowsController, ArtistsController, ChannelsController, LiveTvController, CollectionController, PlaylistController, InstantMixController |
| 11 | `search-service` | 8010 | `/Search` | SearchController (global search across all content types) |
| 12 | `transcode-service` | 8011 | `/transcodes`, `/Transcoding` | TranscodeController (FFmpeg job management, segment delivery) |
| 13 | `collection-service` | 8011 | `/Collections`, `/BoxSets` | CollectionController, BoxSetController |
| 14 | `playlist-service` | 8012 | `/Playlists` | PlaylistController, PlaylistItemsController |
| 15 | `notification-service` | 8013 | `/MessageQueue`, `/Sse` | MessageController, SSEController, WebSocketController |
| 16 | `plugin-service` | 8014 | `/Plugins` | PluginController, PackageController, BackupController |

## Notes

### Port Assignment Rationale
- **8001-8006**: Core services (auth, user, library, playstate, media, stream)
- **8007**: Session management (user-facing device/session tracking)
- **8008**: Metadata operations (item refresh, external lookups, scheduled tasks)
- **8009**: Content browsing with rich relationships (movies, TV, genres, collections)
- **8010**: Search (full-text search across all content)
- **8011**: Transcode worker (internal FFmpeg orchestration - moved from 8007)
- **8011**: Collection service (same port as transcode worker, separate process)
- **8012**: Playlist service
- **8013**: Notification/SSE service
- **8014**: Plugin management

## Database Schema

All services share a single MySQL 8.0 database:
- **Users API:** `users`, `api_keys`, `devices`, `sessions`, `quick_connect_codes`
- **Items API:** `items`, `item_values`, `user_data`, `library_folders`, `collections`, `playlists`
- **Playstate API:** `playback_state`, `play_history`, `session_viewing`
- **SyncPlay API:** `syncplay_groups`, `syncplay_group_members`, `syncplay_commands`
- **ScheduledTasks API:** `scheduled_tasks`

### Key Indexing Strategy
- **P7 Index:** `items(top_parent_id, type)` - Optimizes recursive CTE hierarchy queries for folder structures
- **P6 Index:** `item_values(item_id, value_type, value)` - Normalized storage for multi-value fields (genres, studios, people)
- **P5 Index:** `user_data(user_id, item_id, is_favorite)` - Quick favorite/uns favorite checks

## Authentication Flow

1. Client requests `/Auth/Devices/...` or `/Users/AuthenticateByName` with username/password
2. Auth service validates credentials, generates 256-bit token, stores SHA256 hash
3. Token returned to client in response header: `X-Emby-Authorization: MediaBrowser Token=...`
4. Client includes this token in all subsequent requests
5. Each service middleware validates token by querying `api_keys` table directly
6. User ID injected into request context for authorization checks

## FFmpeg Transcoding Flow

1. Client requests `/Hls/{itemId}/master.m3u8?transcode=true`
2. Stream service detects transcoding needed, requests job from `/transcodes/jobs`
3. Transcode service spawns FFmpeg process with specific segment output
4. Transcode service responds with `202 Accepted` and job ID
5. Stream service polls `/transcodes/segments/{jobId}/{n}.ts` for segment availability
6. FFmpeg outputs 6-second TS segments to shared volume
7. Transcode service tracks `ActiveRequestCount`, resets 60s idle kill timer on each request
8. When idle for 60s, FFmpeg process is killed, temp files cleaned up

## Directory Structure

```
Kabletown/
├── auth-service/
│   ├── cmd/server/main.go
│   └── internal/handlers/(user, device, apikey, quickconnect, startup).go
├── user-service/
│   └── cmd/server/main.go
├── library-service/
│   ├── cmd/server/main.go
│   └── internal/query/parser.go  # InternalItemsQuery (80+ params)
├── playstate-service/
│   └── cmd/server/main.go
├── media-service/
│   └── cmd/server/main.go
├── stream-service/
│   └── cmd/server/main.go
├── session-service/
│   ├── cmd/server/main.go
│   └── internal/handlers/(session, device, syncplay).go
├── metadata-service/
│   ├── cmd/server/main.go
│   └── internal/handlers/(itemrefresh, itemupdate, remoteimage, scheduledtasks).go
├── content-service/
│   ├── cmd/server/main.go
│   └── internal/handlers/(movies, tvshows, artists, channels, liveTv, collections, playlists).go
├── search-service/
│   └── cmd/server/main.go
├── transcode-service/
│   ├── cmd/server/main.go
│   └── internal/transcode/transcodemanager.go
├── collection-service/
│   └── cmd/server/main.go
├── playlist-service/
│   └── cmd/server/main.go
├── notification-service/
│   └── cmd/server/main.go
├── plugin-service/
│   └── cmd/server/main.go
├── nginx/nginx.conf
├── docker-compose.yml
├── shared/
│   ├── auth/middleware.go
│   ├── db/mysql.go
│   └── logger/logger.go
├── migrations/
│   ├── 001_auth_schema.up.sql
│   ├── 001_auth_schema.down.sql
│   ├── 002_items_schema.up.sql
│   ├── 002_items_schema.down.sql
│   └── 003_playstate_schema.up.sql
└── docs/
    ├── internal-items-query.md  # 80+ param analysis
    └── hls-streaming-routes.md
```

## Service Dependencies

- **No inter-service RPC:** All services share single MySQL database
- **Nginx Gateway:** All external requests routed through nginx:8080
- **Shared Auth:** Token validation via direct DB query (no authservice RPC calls)
- **Transcode Worker:** Transcode management separate from stream delivery (transcode-service=8011, stream-service=8006)
- **File Sharing:** Transcode output volume mounted at `/tmp/transcoding`
