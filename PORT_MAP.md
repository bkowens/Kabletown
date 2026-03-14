# Kabletown Architecture Port Map (Final)

| # | Service | Port | Controllers/Routes | Status |
|---|---|---|---|---|
| 1 | `nginx` | 8080 | API Gateway | ✅ Created |
| 2 | `auth-service` | 8001 | Sessions, APIKeys, Devices, Users, QuickConnect | ✅ Implemented |
| 3 | `user-service` | 8002 | UserController, DisplayPreferences | 🟡 Scaffolded |
| 4 | `library-service` | 8003 | ItemsController (80+ params), SearchController | 🟡 Scaffolded |
| 5 | `playstate-service` | 8004 | PlaystateController | ✅ Implemented |
| 6 | `media-service` | 8005 | ImageController, SubtitleController | 🟡 Scaffolded |
| 7 | `stream-service` | 8006 | VideosController, HlsController, AudioController | ✅ Implemented |
| 8 | `session-service` | 8007 | SessionController, DevicesController, SyncPlayController | ✅ Created |
| 9 | `metadata-service` | 8008 | ItemRefresh, ItemLookup, RemoteImage, ScheduledTasks | ✅ Created |
| 10 | `content-service` | 8009 | Movies, TvShows, Artists, Channels, LiveTv, Collections, Playlists | ✅ Created |
| 11 | `search-service` | 8010 | SearchController | ✅ Scaffolded |
| 12 | `transcode-service` | 8011 | TranscodeController (FFmpeg) | ✅ Created |
| 13 | `collection-service` | 8012 | CollectionController, BoxSetController | ✅ Scaffolded |
| 14 | `playlist-service` | 8013 | PlaylistController | ✅ Scaffolded |
| 15 | `notification-service` | 8014 | MessageController, SSEController | ✅ Scaffolded |
| 16 | `plugin-service` | 8015 | PluginController, PackageController | ✅ Scaffolded |

## Port Conflict Resolution

**Previous state:** transcode-service at 8007
**Resolution:** Moved transcode-service to 8011, assigned session-service to 8007

## Routing Map (nginx.conf)

```
/Sessions, /Devices → session-service:8007
/Metadata, /ScheduledTasks → metadata-service:8008
/Movies, /TvShows, /Artists, /Channels, /LiveTv, /Collections, /Playlists → content-service:8009
/Search → search-service:8010
/transcodes, /Transcoding → transcode-service:8011
```

## Implementation Status

### ✅ Complete (7 services)
- auth-service: Full CRUD for users, devices, API keys, sessions
- playstate-service: Playback tracking (start/stop/playing)
- stream-service: HLS streaming, audio, trickplay
- session-service: Session management, device registration, SyncPlay
- metadata-service: Item refresh, lookup, remote images, scheduled tasks
- content-service: Movies, TV shows, artists, collections, playlists
- transcode-service: FFmpeg job management at port 8011

### 🟡 Scaffolded (9 services)
- user-service: Dockerfile, go.mod, main.go only
- library-service: Dockerfile, go.mod, main.go only (CRITICAL - needs InternalItemsQuery parser)
- media-service: Dockerfile, go.mod, main.go only
- search-service: Dockerfile, go.mod, main.go only
- collection-service: Dockerfile, go.mod, main.go only
- playlist-service: Dockerfile, go.mod, main.go only
- notification-service: Dockerfile, go.mod, main.go only
- plugin-service: Dockerfile, go.mod, main.go only

### 📄 Documentation
- SERVICE_ARCHITECTURE.md: Complete 15-service map
- PROJECT_STATUS.md: Implementation tracker
- docs/internal-items-query.md: 80+ param query analysis
- docs/hls-streaming-routes.md: Streaming architecture

## Next Critical Tasks

1. **library-service (8003)** - Implement InternalItemsQuery parser for `/Items` endpoint
   - Use docs/internal-items-query.md as reference
   - Parse 80+ URL query parameters
   - Support recursive CTE for folder hierarchy traversal
   - Apply P7 index optimization `(top_parent_id, type)`

2. **Fill in handlers** for remaining scaffolded services:
   - user-service: DisplayPreferencesHandler, UserPreferencesHandler
   - media-service: ImageHandler, SubtitleHandler, MediaInfoHandler
   - search-service: GlobalSearchHandler (query across all content types)

3. **Database migrations** - Ensure all tables exist:
   - Add syncplay tables (syncplay_groups, syncplay_group_members, syncplay_commands)
   - Add scheduled_tasks table
   - Verify collection/playlists tables

4. **Test deployment** - Run docker-compose up and verify all services connect to MySQL
