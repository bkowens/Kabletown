# JellyFinhanced → Go Migration: Execution Order

**Purpose:** Dependency graph for parallel agent execution with estimated timelines

---

## Phase 0: Foundation (Week 1)

### Week 1-2: Parallel Examination & Planning
**Status:** ✅ COMPLETE

| Task | Agent | Duration | Dependencies |
|------|-------|---------|-------------|
| 0.1 Controller analysis (60 files) | All agents | 4h | None |
| 0.2 Database schema analysis | Database agent | 2h | None |
| 0.3 Auth middleware analysis | Auth agent | 2h | None |
| 0.4 DTO shapes analysis | All agents | 2h | None |
| 0.5 Streaming/HLS analysis | Stream agent | 2h | None |
| 0.6 Configuration analysis | System agent | 2h | None |
| Create ARCHITECTURE.md | Architect agent | 8h | 0.1-0.6 |
| Create EXECUTION-ORDER.md | Architect agent | 2h | ARCHITECTURE.md |
| Create RISKS.md | Risk analyst | 2h | ARCHITECTURE.md |

**Deliverables:** All planning documents complete

---

## Phase 1: Core Infrastructure (Week 2)

### Week 2: Shared Package & Gateway
**Status:** 🟡 PLANNED

```
┌─────────────────────────────────┐
│   Agent 0: shared package       │
│   Duration: 2 days              │
│   Deliverables:                 │
│   - shared/auth/middleware.go   │
│   - shared/db/factory.go        │
│   - shared/dto/*.go             │
│   - shared/config/loader.go     │
│   - shared/response/errors.go   │
└──────────────┬──────────────────┘
               │
               ▼
    ┌──────────────────────┐
    │   All other agents   │
    │   (can start in parallel)  │
    └──────────────────────┘
```

**Agent 0 Tasks:**
```bash
# Directory structure
cd Kabletown
git clone shared

cd shared
go mod init github.com/jellyfinhanced/shared

# Create package structure
mkdir -p auth db dto config response pagination

# Week 2 schedule:
# Day 1: auth/, db/, config/
# Day 2: dto/, response/, pagination/, README, tests

# Dependencies:
# - github.com/go-chi/chi/v5
# - github.com/jmoiron/sqlx
# - github.com/go-sql-driver/mysql
# - golang.org/x/crypto/bcrypt
# - github.com/golang-jwt/jwt/v5 (for JWT auth if needed)

# Testing:
# - Unit tests: go test ./... (mock DB)
# - Integration tests: go test ./tests (testcontainers-go)
```

**Acceptance Criteria:**
- [ ] All unit tests pass (go test ./... -cover: 80%+)
- [ ] Docker Compose file for shared package (local dev MySQL)
- [ ] README.md with setup instructions
- [ ] API documentation draft (OpenAPI 3.0)

---

## Phase 2: Core Services (Weeks 3-5)

### Week 3: Authentication & User Management
**Dependencies:** Agent 0 (shared package)

```
         ┌────────────────┐
         │  Agent 0       │
         │  (COMPLETE)    │
         └────┬───────────┘
              │
    ┌─────────┴─────────┐
    │                   │
    ▼                   ▼
┌──────────┐      ┌──────────┐
│ Agent 1  │      │ Agent 2  │
│ auth     │      │ user     │
│ service  │      │ service  │
└──────────┘      └──────────┘
                   (also system-service)
```

**Agent 1: auth-service (2 days)**
```
Week 3, Day 1-2
Routes: 15 endpoints
Complexity: HIGH (crypto, session management)
Dependencies: None (foundational service)

Files to create:
├── cmd/server/main.go
├── internal/
│   ├── middleware/
│   │   └── auth.go          # X-Emby-Authorization parsing
│   ├── handlers/
│   │   ├── auth.go          # AuthenticateByName, QuickConnect
│   │   ├── apikey.go        # API key CRUD
│   │   └── startup.go       # Startup wizard endpoints
│   ├── db/
│   │   ├── mysql.go         # sqlx implementation
│   │   ├── mock.go          # mock for tests
│   │   └── queries.sql      # sqlc queries
│   └── dto/
│       └── types.go         # All auth DTOs
├── Dockerfile
├── docker-compose.yml
├── openapi.yaml
├── go.mod
├── go.sum
└── README.md

Test Coverage:
- Unit tests: 80%+
- Integration test: testcontainers-go
```

**Agent 2: user-service + system-service (3 days)**
```
Week 3, Days 3-5

USER SERVICE: 12 endpoints
  - /Users (list, create)
  - /Users/{id} (get, update, delete)
  - /Users/{id}/Password
  - /Users/{id}/Authenticate
  - /Users/{id}/Policy
  - /Users/{id}/Configuration
  - /Users/{id}/Views
  - /DisplayPreferences

SYSTEM SERVICE: 15 endpoints
  - /System (info, logs, tasks)
  - /Configuration (read/write)
  - /Dashboard (metrics)
  - /Branding (read-only)
  - /Environment
  - /Localization
  - /ClientLog
  - /ActivityLog

Dependencies:
  - shared (required)
  - auth (for token validation)

Risk: Medium (user data writes)
```

**Cutover Order for Week 3:**
1. Deploy auth-service to shadow mode
2. Deploy user-service to shadow mode
3. Deploy system-service to shadow mode
4. Run validation tests
5. Switch to 5% canary

---

## Phase 3: Library & Content (Weeks 4-6)

### Week 4-6: Browsing Core
**Dependencies:** Agent 0, Agent 1, Agent 2

```
         ┌─────────────┐
         │ Agent 0     │
         │ auth+user   │
         └──────┬──────┘
                │
    ┌───────────┴───────────┐
    │                       │
    ▼                       ▼
┌────────────┐        ┌────────────┐
│ Agent 3    │        │ Agent 5    │
│ library    │        │ content    │
│ service    │        │ service    │
└─────┬──────┘        └─────┬──────┘
      │                     │
      └───────────┬─────────┘
                  ▼
         ┌──────────────┐
         │ Agent 4      │
         │ search       │
         │ metadata     │
         └──────────────┘
```

**Agent 3: library-service (5 days - LARGEST)**
```
Week 4, Days 1-5
Routes: 30+ endpoints (most complex service)
Complexity: VERY HIGH (multi-filter queries, joins, performance critical)
Dependencies:
  - shared (required)
  - user-service (optional user-scoped filtering)

Critical routes:
  - GET /Items (50+ filter parameters)
  - GET /Items/{id} (complex join logic)
  - GET /Library/VirtualFolders
  - GET /Genres, /Studios, /Persons
  - GET /Years, /MusicGenres
  - GET /ItemLookup, ItemUpdate
  - GET /Filters

Database complexity:
  - BaseItems (main table, complex WHERE/ORDER BY)
  - ItemValues (genre/studio joins)
  - ItemValuesMap (junction table)
  - AncestorIds (recursive queries)
  - PeopleBaseItemMap (cast/crew)
  - MediaStreamInfos (video/audio track listing)
  - BaseItemImageInfos (image serving)

Must implement SQL query builder:
  - Dynamic WHERE clause (50+ optional filters)
  - Pagination (StartIndex, Limit, TotalRecordCount)
  - Sorting (15+ columns, ASC/DESC)
  - Item type filtering
  - User data joins (for play state)

Perf requirements:
  - Query time < 50ms (cached results)
  - Query time < 200ms (uncached, complex filters)
  - Must use composite indexes from 2026-03-09 migration

Files to create:
  ├── internal/db/
  │   ├── queries_baseitems.sql      # 30+ queries
  │   ├── queries_itemvalues.sql     # 10+ queries
  │   ├── queries_people.sql         # 15+ queries
  │   ├── queries_images.sql         # 5+ queries
  │   └── repositories/
  │       ├── baseitem.go            # BaseItemRepository
  │       ├── itemvalue.go           # ItemValueRepository
  │       └── people.go              # PeopleRepository
  ├── internal/handlers/
  │   ├── items.go                   # ItemsController
  │   ├── library.go                 # LibraryController
  │   ├── metadata.go                # Genres/Studios/Persons
  │   └── filters.go                 # FilterController
  └── ...

Testing:
  - Unit tests for query builder
  - Integration tests with real MySQL
  - Benchmark tests for performance
  - Comparison tests vs C# responses
```

**Agent 4: media-service + stream-service (6 days)**
```
Week 5, Days 1-6
Routes: 25 endpoints
Complexity: VERY HIGH (ffmpeg integration, transcoding, file I/O)
Dependencies:
  - shared (required)
  - library-service (media info lookup)

MEDIA SERVICE:
  - /MediaInfo (codec detection, media probes)
  - /MediaSegments (chapter info)
  - /Items/{id}/Images (resizing, serving)
  - /Items/{id}/RemoteImages (TMDB/TVDb API calls)
  - /Subtitles (subtitle serving, scraping)
  - /VideoAttachments
  - /Lyrics

STREAM SERVICE:
  - /Videos/{id}/live.m3u8 (HLS master playlist)
  - /Videos/{id}/hls1/{segmentId} (HLS segments)
  - /Videos/{id}/stream (direct video)
  - /Audio/{id}/stream (audio playback)
  - /Trickplay (preview thumbnails)

Transcoding complexity:
  - Shell out to ffmpeg (exact C# command matching)
  - HLS segment generation (ffmpeg HLS muxer)
  - Video codec detection (h264, h265, vp9, av1)
  - Audio codec detection (aac, opus, mp3, flac)
  - Subtitle burn-in (ass/srt embedding)
  - Segment cache management
  - Transcoding job cleanup

FFmpeg requirements:
  - Version detection (auto-detect min version)
  - Command logging (for debugging)
  - Error handling (FFmpeg exit codes)
  - Process management (timeout, kill)

Perf requirements:
  - Image resize time < 500ms (primary)
  - Image resize time < 100ms (thumbnails)
  - HLS playlist generation < 100ms
  - HLS segment serving < 10ms (cached)
  - Video transcoding: match C# bitrates

Files to create:
  ├── internal/ffmpeg/
  │   ├── encoder.go            # FFmpeg wrapper
  │   ├── hls_playlist.go       # HLS master/segment generation
  │   ├── image_processor.go    # Image resizer
  │   ├── media_probe.go        # ffprobe wrapper
  │   └── commands/
  │       ├── hls.go            # ffmpeg HLS commands
  │       ├── image.go          # ffmpeg image commands
  │       └── subtitle.go       # ffmpeg subtitle commands
  └── ...

Testing:
  - Mock FFmpeg tests (unit)
  - FFmpeg integration tests (real ffmpeg)
  - Image quality tests (compare output files)
  - Performance tests (concurrent image serving)
```

**Agent 5: search-service + metadata-service (3 days)**
```
Week 5, parallel with Agent 4

SEARCH SERVICE:
  - /Search (full-text search, fuzzy matching)
  - /ItemLookup (metadata enrichment)
  - /Filters (saved filters)
  - /Suggestions (autocomplete hints)

METADATA SERVICE:
  - /Items/Refresh (scrape metadata providers)
  - /Items/Update (manual metadata edit)
  - /Packages (plugin marketplace)
  - /Plugins (plugin lifecycle)
  - /ScheduledTasks (CRON jobs)
  - /Plugin/{id}/Update

Dependencies:
  - library-service (item lookup)

Testing:
  - Search result accuracy vs C#
  - Metadata provider API mocking
```

**Agent 6: content-service (3 days)**
```
Week 6, Days 1-3
Routes: 20 endpoints
Complexity: HIGH (complex filtering, aggregation)
Dependencies:
  - shared (required)
  - library-service (item queries)
  - playstate-service (user data)

Routes:
  - /Shows (TV shows with seasons/episodes)
  - /Shows/{id}/Seasons
  - /Shows/{id}/Episodes
  - /Shows/{id}/NextUp (watchlist generation)
  - /Movies
  - /Artists
  - /Artists/{id}/Albums
  - /Albums/{id}/Songs
  - /Collections (smart collections)
  - /Playlists (create/list/edit)
  - /InstantMix (recommendation generation)
  - /Trailers

Testing:
  - TV show hierarchy queries (series → season → episode)
  - NextUp algorithm (match C# output)
  - Album/Song grouping
  - Collection smart rules (AND/OR filtering)
```

**Agent 7: playstate-service (2 days)**
```
Week 6, Days 1-2
Routes: 8 endpoints
Complexity: MEDIUM (user data writes, cache invalidation)
Dependencies:
  - shared (required)
  - library-service (item lookup)

Routes:
  - /PlayedItems/markPlayed
  - /PlayedItems/markUnplayed
  - /PlayedItems/markPlayed/{batch}
  - /ResumePoints/{id}/start
  - /ResumePoints/{id}/end
  - /ResumePoints/{id}/delete
  - /UserData
  - /UserData/{id}/favorite

Notes:
  - User data changes broadcast via WebSocket
  - Invalidation cache for recent activity

Testing:
  - Concurrent user data writes
  - Race condition handling
  - Cache invalidation
```

---

## Phase 4: Session & Advanced (Weeks 5-7)

### Week 7: Real-Time Features
**Dependencies:** Agent 0, Agent 1, Agent 2

```
┌──────────────────────────────────┐
│ Agent 6: session-service         │
│ Duration: 3 days                 │
│ Dependencies: shared, auth       │
└──────────────────────────────────┘
```

**Agent 8: session-service (3 days)**
```
Week 7, Days 1-3
Routes: 15 endpoints
Complexity: HIGH (WebSocket state management, Redis)
Dependencies:
  - shared (required)
  - auth-service (user validation)

Routes:
  - /Sessions (list, kill)
  - /Devices (list, remove, block)
  - /SyncPlay/* (WebSocket hub)
  - /TimeSync (clock sync)

WebSocket requirements:
  - gorilla/websocket for connections
  - Redis pub/sub for broadcast across instances
  - Client heartbeats (5s timeout)
  - SyncPlay commands:
    - CreateGroup, JoinGroup, LeaveGroup
    - PlaybackStart, Pause, Seek
    - BufferingStart, BufferingStop

Testing:
  - 100 concurrent WebSocket connections
  - SyncPlay state sync across instances
  - Graceful disconnection handling
```

---

## Phase 5: Final Services (Weeks 7-8)

### Week 8: Remaining Services
**Dependencies:** All previous agents

```
         ┌───────────┐  ┌─────────────┐  ┌──────────┐
         │ Agent 9   │  │ Agent 10    │  │ Agent 11 │
         │ tv        │  │ backup      │  │ gateway  │
         │ service   │  │ service     │  │ + tests  │
         └───────────┘  └─────────────┘  └──────────┘
```

**Agent 9: tv-service (2 days) - OPTIONAL / PHASE 2**
```
Week 8, Days 1-2
Routes: 10 endpoints
Complexity: HIGH (HDHomeRun integration, tuner management)
Dependencies:
  - library-service (LiveTv item queries)
  - media-service (tuner stream proxy)

Routes:
  - /LiveTv (tuner status, channel guide)
  - /LiveTv/Tuners (list/add/remove tuners)
  - /LiveTv/Guide (EPG fetching)
  - /LiveTv/Recordings (list/create/delete)
  - /LiveTv/Series (series linking)
  - /LiveTv/Preferences

Notes:
  - Defer to phase 2 (complex hardware integration)
  - Keep C# service during transition
```

**Agent 10: backup-service (1 day)**
```
Week 8, Day 3
Routes: 4 endpoints
Complexity: LOW (file operations, archive creation)
Dependencies: None

Routes:
  - /System/Backup/Create (archive config + DB dump)
  - /System/Backup/Restore (upload + apply)
  - /System/Backup/Export (export user data)
  - /System/Backup/Import (import JSON config)

Implementation:
  - Use standard library archives (tar/zip)
  - File path traversal protection
  - Atomic rollback on failure
```

**Agent 11: gateway + integration tests (2 days)**
```
Week 8, Days 4-5
Deliverables:
  - nginx.conf (complete routing table)
  - docker-compose.yml (all 11 services + MySQL + Redis)
  - integration/smoke_test.go (Go HTTP tests)
  - integration/docker-compose.test.yml
  - Load test scenarios (vegeta/k6 scripts)

Smoke tests:
  - All 120+ endpoints (happy path)
  - Auth failure scenarios (401/403)
  - Missing resource scenarios (404)
  - Request validation (400)
  - Response schema validation (JSON schema)

Load tests:
  - 100 req/sec baseline for auth/user endpoints
  - 500 req/sec for library browsing
  - 2000 req/sec for static images
  - 500 concurrent WebSocket connections
```

---

## Total Timeline Summary

| Phase | Agent(s) | Duration | Dependencies |
|-------|----------|---------|-------------|
| 0. Planning | All | 2 weeks | None |
| 1. Shared | Agent 0 | 2 days | None |
| 2. Auth+User | Agents 1-2 | 3 days | Agent 0 |
| 3. Library+Content | Agents 3-7 | 9 days | Agent 0,1,2 |
| 4. Session | Agent 8 | 3 days | Agent 0,1 |
| 5. Final | Agents 9-11 | 3 days | All previous |
| 6. Testing+migration | All | 2 weeks | All services |
| **Total** | **Parallel agents** | **~30 days** | **Key path: 11 days (sequential)** |

**Parallelization Strategy:**
- Days 1-2: Agent 0 (shared package) - BLOCKER
- Days 3-5: Agents 1-2 (auth+user+system) - parallel
- Days 6-10: Agents 3-8 (library/media/session/content) - all parallel
- Days 11-13: Agents 9-11 (tv+backup+gateway) - all parallel
- Days 14-28: Integration testing + shadow mode
- Days 29-32: Canary → full cutover → C# shutdown

---

## Risk Mitigation Timeline

| Week | Risk | Mitigation Action |
|------|--|--|
| 1 | Unclear DTO shapes | Review ARCHITECTURE.md Phase 0 findings |
| 2 | DB migration blockers | Test queries against existing MySQL schema |
| 3 | Performance regression | Run benchmarks on library-service queries |
| 4 | FFmpeg integration issues | Log exact commands, compare with C# trace |
| 5 | WebSocket scaling problems | Redis pub/sub integration tests |
| 6 | Image quality mismatch | Compare Go images with C# side-by-side |
| 7 | Plugin migration gaps | Audit production plugin usage |
| 8 | Live TV hardware issues | Defer to phase 2 if blockers |
| 9-12| User-reported bugs | Shadow mode logs, quick rollback plan |

---

## Decision Points (Required Before Cutover)

1. **Week 2:** Choose image processing library (`disintegration/imaging` vs `vips` CGO)
2. **Week 3:** Finalize JWT or session token approach for Auth
3. **Week 4:** Confirm FFmpeg binary path/availability in production
4. **Week 5:** Redis instance sizing (memory for session cache)
5. **Week 6:** CDN strategy for image serving (if external)
6. **Week 7:** Live TV tuner hardware compatibility list
7. **Week 8:** Plugin system decision (bake-in vs. defer)

---

## Rollback Triggers (Per-Service Cutover)

**Automated alerts (rollback immediately):**
- Error rate > 5% for 5 minutes
- P95 latency > 1000ms (stream-service), > 500ms (others)
- DB connection pool exhaustion
- Memory usage > 80% for 10 minutes

**Manual rollback (human judgment):**
- Missing/incorrect response fields (JSON schema mismatch)
- User-reported auth failures
- Incorrect library data (missing episodes/genres)
- Playback failures (stream-service)

**Rollback procedure:**
```bash
# 1. Update nginx config (5 min)
sed -i 's/auth-service:8001/jellyfin-csharp:8096/' /etc/nginx/conf.d/services.conf

# 2. Reload nginx
nginx -s reload

# 3. Capture error logs
cp /var/log/nginx/error.log /tmp/rollback_$(date +%Y%m%d_%H%M%S).log

# 4. Report to Slack (#jellyfin-migration) with error samples
```

---

## Success Criteria (Go Cutover Complete)

- [ ] All 120+ endpoints returning 200/400 in integration tests
- [ ] Error rate < 0.01% over 7 days
- [ ] P95 latency ≤ C# baseline (measured per service)
- [ ] Memory usage ≤ current C# usage
- [ ] CPU usage ≤ 80% of C# baseline (headroom for scaling)
- [ ] User-reported issues < 5 per week
- [ ] Plugin system functional (or documented defer plan)
- [ ] WebSocket sessions (SyncPlay) operational
- [ ] Image quality ≥ C# visual quality (manual audit)
- [ ] FFmpeg transcoding matches C# output (bitrate, codec, duration)
- [ ] Production stability: 72 hours with 0 incidents

---

## Communication Cadence

| Audience | Frequency | Format |
|----------|----------|--------|
| Engineers | Daily | 15-min standup (Slack huddle) |
| Stakeholders | Weekly | Progress report (Slack #migration-updates) |
| Users | Bi-weekly | Changelog (GitHub Releases) |
| Incident review | Post-mortem | Slack thread + written summary |

---

## Appendix: OpenAPI Generation

```bash
# After each agent completes handlers:

# 1. Generate OpenAPI from code (go-swagger)
go get -u github.com/swaggo/swag/cmd/swag
swag init -g cmd/server/main.go -o ./docs

# 2. Validate spec
swagger validate ./docs/swagger.yaml

# 3. Generate client SDK (optional)
swag-codegen generate client -l go --input ./docs/swagger.yaml

# 4. Publish spec
gh release upload --clobber v0.1.0 ./docs/swagger.yaml
```

---

*Document Version: 1.0*
*Last Updated: 2026-03-12*
*Source: Kabletown migration planning*
