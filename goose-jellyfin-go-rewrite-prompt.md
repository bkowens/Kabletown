# Goose AI Prompt: Jellyfin → Go Microservices Rewrite

## Mission

You are a senior Go architect. Your task is to analyze the JellyFinhanced Jellyfin codebase (a C# / ASP.NET Core media server) located at `/home/bowens/Code/JellyFinhanced` and produce a **complete, agent-ready development plan** for rewriting it as a Go-based micro-API architecture.

### Non-Negotiable Constraints
1. **MySQL schema compatibility** — Every microservice must use the _existing_ MySQL schema as-is. No DDL changes; all tables, columns, indexes, and relationships remain identical.
2. **Web client compatibility** — All HTTP routes, request/response shapes, status codes, headers, and authentication mechanisms must be wire-compatible with the existing Jellyfin web client (`jellyfin-web/`). The browser app must work against the new backend with zero changes.
3. **One agent per API domain** — The plan must be sliced so each major API group can be independently implemented by a separate AI coding agent, with clean boundaries and minimal cross-agent dependencies.
4. **Go idioms** — Use idiomatic Go: `net/http` or Chi/Fiber, `database/sql` with `sqlx` or `sqlc`, structured logging (`slog`), context propagation, and graceful shutdown.

---

## Phase 0 — Codebase Examination (Do This First)

Before producing any plan, examine the following key locations in the source tree and extract the information listed under each path.

### 0.1 API Controllers
**Path:** `Jellyfin.Api/Controllers/`

For **every** `.cs` file in this directory, record:
- Controller class name
- All HTTP methods and route templates (look for `[HttpGet]`, `[HttpPost]`, `[HttpDelete]`, `[HttpPut]`, `[Route]`, `[ApiController]`)
- Authorization requirements (`[Authorize]`, `[AllowAnonymous]`, policy names)
- Input parameters (query string, route params, request body DTOs)
- Return types (response DTOs, status codes)
- Which service interfaces are injected (constructor parameters)

### 0.2 Database Schema
**Path:** `src/Jellyfin.Database/Jellyfin.Database.Implementations/JellyfinDbContext.cs`
**Path:** `src/Jellyfin.Database/Jellyfin.Database.Providers.MySql/Migrations/`

Record every `DbSet<T>`, table name, column name/type, primary keys, foreign keys, unique constraints, and indexes from the latest migration file (`20260309000000_AddPerformanceIndexes`). Pay particular attention to the `BaseItems`, `Users`, `UserData`, `Devices`, `MediaStreamInfos`, and `ItemValues` tables as they carry the highest query volume.

### 0.3 Authentication Middleware
**Path:** `Jellyfin.Api/Auth/` (and any `Middleware/` subdirectory under `Jellyfin.Server/`)

Record:
- How the `Authorization` header is parsed (Bearer token vs. API key)
- How `X-Emby-Authorization` / `MediaBrowser Token` header is parsed
- The device-ID + access-token pairing flow
- How the `UserId` is extracted from claims and passed to controllers
- Which endpoints allow anonymous access

### 0.4 DTO Shapes
**Path:** `Jellyfin.Data/` and `MediaBrowser.Model/Dto/`

For the most-used response DTOs record every field name, type, and whether it is nullable:
- `BaseItemDto`
- `UserDto`
- `SessionInfoDto`
- `DeviceInfoDto`
- `AuthenticationResultDto`
- `PlaybackInfoResponse`

### 0.5 Streaming & HLS
**Path:** `Jellyfin.Api/Controllers/DynamicHlsController.cs` and `VideosController.cs`

Record the exact route templates for:
- Master playlist (`master.m3u8`)
- HLS segment requests (`hls1/{segmentId}/{segmentContainer}`)
- Direct video stream
- Video file download

Note all query parameters that control transcoding (bitrate, codec, container, audioStreamIndex, etc.).

### 0.6 Configuration & Startup
**Path:** `Jellyfin.Server/Program.cs` and `Jellyfin.Server/Startup.cs`

Record: middleware registration order, CORS policy, static file serving for the web client, health check endpoint, and any custom response headers set globally.

---

## Phase 1 — Architecture Blueprint

After completing Phase 0, produce an architecture document covering:

### 1.1 Service Decomposition

Group the 60 controllers into **logical microservices** based on domain cohesion and data ownership. Use the groupings below as a starting point but adjust based on what you find in the code:

| Service Name | Controllers It Covers | Primary DB Tables |
|---|---|---|
| **auth-service** | ApiKeyController, UserController (auth endpoints), QuickConnectController, StartupController | Users, Devices, ApiKeys, Permissions, Preferences, AccessSchedules |
| **library-service** | ItemsController, LibraryController, LibraryStructureController, FilterController, YearsController, GenresController, MusicGenresController, StudiosController, PersonsController, TrailersController | BaseItems, ItemValues, ItemValuesMap, BaseItemProviders, AncestorIds, PeopleBaseItemMap, Peoples |
| **user-service** | UserController (CRUD), UserLibraryController, UserViewsController, DisplayPreferencesController, PlaystateController, SuggestionsController | Users, UserData, DisplayPreferences, ItemDisplayPreferences |
| **media-service** | MediaInfoController, MediaSegmentsController, VideosController, AudioController, ImageController, VideoAttachmentsController, LyricsController | BaseItems, MediaStreamInfos, AttachmentStreamInfos, BaseItemImageInfos, MediaSegments |
| **stream-service** | DynamicHlsController, HlsSegmentController, UniversalAudioController, TrickplayController | TrickplayInfos, KeyframeData (read-only from disk segments) |
| **search-service** | SearchController, ItemLookupController, RemoteImageController | BaseItems (full-text), BaseItemImageInfos |
| **session-service** | SessionController, DevicesController, SyncPlayController, TimeSyncController | Devices, DeviceOptions |
| **metadata-service** | ItemRefreshController, ItemUpdateController, BackupController, PackageController, PluginsController, ScheduledTasksController | BaseItems, BaseItemProviders, BaseItemImageInfos, Chapters |
| **tv-service** | TvShowsController, MoviesController, ArtistsController, ChannelsController, LiveTvController | BaseItems (filtered by type), Chapters |
| **collection-service** | CollectionController, PlaylistsController, InstantMixController | BaseItems, UserData |
| **system-service** | SystemController, ConfigurationController, DashboardController, BrandingController, EnvironmentController, LocalizationController, ClientLogController, ActivityLogController | ActivityLogs |

For each service, specify:
- Go module name (e.g., `github.com/jellyfinhanced/auth-service`)
- Port number
- Which other services it calls (upstream dependencies)
- Shared packages it consumes from `github.com/jellyfinhanced/shared`

### 1.2 Shared Package (`/shared`)

Define the shared Go module that all services import. It must contain:
- **MySQL client factory** — `*sqlx.DB` pool builder from environment variable `DATABASE_URL`
- **Auth middleware** — Parses `X-Emby-Authorization`, Bearer tokens, and API keys; populates `context.Context` with `UserID`, `DeviceID`, `Token`, `IsAdmin`
- **Response helpers** — JSON envelope, error response (matching Jellyfin's `ProblemDetails` format), pagination (`StartIndex`/`TotalRecordCount`/`Items` envelope)
- **DTO types** — All shared response structs generated from the DTO shapes found in Phase 0
- **Config loader** — Reads `JELLYFIN_CONFIG_DIR` for XML config files and `DATABASE_URL` for DB connection

### 1.3 API Gateway

Define a lightweight reverse proxy (Nginx or Caddy configuration, or a minimal Go gateway) that:
- Routes `/Users*` → user-service
- Routes `/Items*`, `/Library*`, `/Genres*`, `/Persons*` → library-service
- Routes `/Videos*`, `/Audio*`, `/Images*` → media-service or stream-service
- Routes `/Sessions*`, `/Devices*` → session-service
- Routes `/System*`, `/Configuration*` → system-service
- Serves the static web client (`jellyfin-web/dist/`) at `/web`
- Forwards `X-Emby-Authorization` and `Authorization` headers unchanged

### 1.4 Infrastructure Contracts

For each service specify:
- `Dockerfile` (multi-stage: `golang:1.23-alpine` build → `alpine:3.21` runtime)
- Environment variables (with defaults)
- Health check endpoint (`GET /healthz`)
- Graceful shutdown timeout (suggest 30 s)
- Log format (JSON via `slog`, fields: `time`, `level`, `service`, `trace_id`, `msg`)

---

## Phase 2 — Per-Agent Implementation Plans

For each service defined in Phase 1, produce a **self-contained agent task card**. Each card must be usable by a fresh AI coding agent with no prior context.

### Task Card Template

```
## Agent Task: <ServiceName>

### Goal
<One paragraph describing what this service does and why it exists>

### Deliverables
- [ ] Go module at `./<service-name>/`
- [ ] `cmd/server/main.go` — HTTP server setup, graceful shutdown, signal handling
- [ ] `internal/db/queries.go` — All SQL queries (use sqlx named queries or sqlc)
- [ ] `internal/handlers/` — One file per controller group, matching Jellyfin routes exactly
- [ ] `internal/middleware/` — Auth extraction, logging, recovery
- [ ] `internal/dto/` — Request/response types matching Jellyfin's JSON field names exactly
- [ ] `Dockerfile`
- [ ] `README.md` — Environment variables, local dev instructions
- [ ] Unit tests for handler logic (mock DB interface)
- [ ] Integration test against real MySQL (use testcontainers-go)

### Routes to Implement
<Exhaustive list from Phase 0 for this service's controllers>

### Database Tables Used
<List from Phase 1>

### Auth Requirements
<Which endpoints require auth, which policies apply>

### Cross-Service Calls
<Which other services this service must call and what endpoints>

### MySQL Query Patterns
<The 5-10 most important queries this service must execute, with expected indexes>

### Wire Compatibility Checklist
- [ ] Route paths are byte-for-byte identical to C# routes
- [ ] JSON field names use exact same casing (PascalCase per Jellyfin convention)
- [ ] `TotalRecordCount` + `Items` + `StartIndex` envelope present on all list responses
- [ ] `X-Application-Version`, `X-Server-Id`, `X-MediaBrowser-Server-Id` response headers present
- [ ] `204 No Content` returned on DELETE (not 200)
- [ ] Error responses match `{ "Message": "...", "StatusCode": NNN }` shape

### Go Implementation Notes
<Specific gotchas for this service: e.g., pipe-delimited genres in BaseItems, ticks vs. milliseconds for RunTimeTicks, Base64 image data in ImageController, HLS segment file path construction>
```

---

## Phase 3 — Detailed Implementation Plans per Agent

After producing all task cards, expand **each** into a step-by-step implementation plan the agent can execute sequentially:

### Step-by-Step Format (per service)

```
### Step 1 — Scaffold the module
go mod init github.com/jellyfinhanced/<service-name>
go get github.com/go-chi/chi/v5
go get github.com/jmoiron/sqlx
go get github.com/go-sql-driver/mysql
go get golang.org/x/crypto
go get github.com/jellyfinhanced/shared (local replace directive)

### Step 2 — Define DTO structs
File: internal/dto/types.go
<List exact Go struct definitions with json tags matching Jellyfin's PascalCase>

### Step 3 — Define DB interface and queries
File: internal/db/interface.go   — interface with one method per query
File: internal/db/mysql.go       — sqlx implementation
File: internal/db/mock.go        — mock for unit tests
<List each method signature and the SQL it runs>

### Step 4 — Implement middleware
File: internal/middleware/auth.go
<Exact parsing logic for X-Emby-Authorization header and Bearer token>

### Step 5 — Implement handlers (one subsection per controller)
File: internal/handlers/<controller>.go
<For each route: function signature, auth check, DB call, DTO mapping, response>

### Step 6 — Wire the router
File: cmd/server/main.go
<Chi router setup, middleware chain, port binding, graceful shutdown>

### Step 7 — Write tests
File: internal/handlers/<controller>_test.go
<httptest.NewRecorder tests for happy path, auth failure, 404>

### Step 8 — Build and verify
go build ./...
go test ./...
curl -H "X-Emby-Authorization: MediaBrowser Token=..." http://localhost:<port>/<route>
Compare response to live C# server response (diff the JSON shapes)
```

---

## Phase 4 — Migration & Rollout Plan

Produce a phased rollout plan that allows running the Go services alongside the existing C# server:

1. **Shadow mode** — Go services receive mirrored traffic but responses are discarded; compare DB query results for accuracy.
2. **Canary** — Route 5% of traffic for non-streaming endpoints to Go services; monitor error rates.
3. **Per-service cutover** — Promote services one at a time: auth → library → user → media → stream → all others.
4. **C# shutdown** — Once all services are promoted and stable for 72 h, terminate the C# process.

For each phase specify:
- Nginx/Caddy configuration snippet
- Smoke test curl commands
- Rollback procedure

---

## Phase 5 — Open Questions & Risk Register

Identify and document any areas where the plan requires human judgment before an agent can proceed:

1. **Transcoding** — The `stream-service` must invoke `ffmpeg` for transcoding. Document whether the plan is to shell out to `ffmpeg` directly (same as C#) or integrate `libav` via CGO. Default recommendation: shell out, matching the C# approach.
2. **Plugin system** — The C# server has a dynamic plugin loader. The Go rewrite will likely drop this. Document which plugins are in active use and whether their functionality needs to be baked in.
3. **SyncPlay** — Requires WebSocket support. Document which Go WebSocket library to use (`gorilla/websocket` or `nhooyr.io/websocket`) and how session state is shared across horizontally scaled instances (Redis recommended).
4. **Live TV** — HDHomeRun and IPTV tuner integration is complex. Flag this as optional scope for a later phase.
5. **Metadata providers** — TMDB, TVDb, MusicBrainz, etc. are currently called from the C# process. Document whether the `metadata-service` agent should implement these HTTP clients or treat them as out-of-scope (metadata can still be triggered via the existing `ItemRefreshController` shelling out to the C# process during transition).
6. **Image processing** — `ImageController` resizes images on-the-fly using SkiaSharp. Document Go alternatives (`imaging`, `disintegration/imaging`, or CGO to `libvips`).
7. **Config file compatibility** — The C# server writes XML config files. The Go services should read but not write these files during the transition to preserve C# operability.

---

## Execution Instructions for Goose

When you execute this prompt, produce your output in the following order:

1. **Phase 0 Findings** — Structured tables/lists from code examination (no prose, maximum density)
2. **Phase 1 Architecture Document** — Service map, shared package spec, gateway routing table
3. **Phase 2 Task Cards** — One card per service (11 cards total)
4. **Phase 3 Implementation Plans** — Expanded step-by-step per service
5. **Phase 4 Rollout Plan**
6. **Phase 5 Risk Register**

For Phase 3, if the total output would be extremely long, produce the three highest-priority services in full (`auth-service`, `library-service`, `user-service`) and provide abbreviated plans (Steps 1–4 only) for the remaining 8 services, noting that each requires a follow-up agent invocation.

**Output format:** GitHub-flavored Markdown. Use fenced code blocks for all Go, SQL, shell, and Nginx/Caddy snippets. Use tables wherever structured comparison is useful.

---

## Context Summary (Pre-Loaded for You)

To save examination time, here is a summary of what the codebase contains. Verify these facts during Phase 0 but use them as your starting map:

### Controller Count
60 controllers in `Jellyfin.Api/Controllers/`. Key high-traffic ones:
- `ItemsController` — Core library browse, complex multi-filter queries against `BaseItems`
- `UserLibraryController` — User-scoped item queries joined with `UserData`
- `DynamicHlsController` — HLS master playlist + segment serving, wraps ffmpeg
- `SessionController` — Active session listing and remote control
- `ImageController` — Resized image serving from disk cache

### Database Tables (MySQL, 40+ tables)
Core tables by query frequency:
1. `BaseItems` — ~30 columns, all media items, full-text index on `Name`/`OriginalTitle`
2. `UserData` — Composite PK (`UserId`, `ItemId`), tracks play state, favorites, ratings
3. `Users` — Auth credentials + preferences
4. `Devices` — Client sessions, access tokens
5. `MediaStreamInfos` — Video/audio/subtitle track metadata per item
6. `ItemValues` — Deduplicated genre/artist/studio values
7. `ItemValuesMap` — Junction table linking items to values
8. `BaseItemImageInfos` — Image paths + metadata per item
9. `PeopleBaseItemMap` — Cast/crew junction table
10. `ActivityLogs` — Audit trail

### Authentication Flow
1. Client sends `X-Emby-Authorization: MediaBrowser Client="...", Device="...", DeviceId="...", Version="...", Token="<access_token>"` header on every request.
2. Server looks up `Token` in `Devices.AccessToken` (GUID).
3. Resolves `Devices.UserId` → loads `User` → builds auth principal.
4. For API keys: `Token` found in `ApiKeys.AccessToken` (string) → treated as admin.
5. Endpoints annotated `[AllowAnonymous]` skip this check.

### Key Wire-Format Details
- All GUIDs are lowercase hyphenated: `"3f2504e0-4f89-11d3-9a0c-0305e82c3301"`
- Timestamps are ISO 8601 UTC: `"2024-01-15T22:30:00.0000000Z"`
- Duration is in **ticks** (1 tick = 100 nanoseconds): `RunTimeTicks: 72000000000` = 2 hours
- List responses always have envelope: `{ "Items": [...], "TotalRecordCount": N, "StartIndex": 0 }`
- Image URLs: `/Items/{id}/Images/{imageType}/{index}?fillHeight=N&fillWidth=N&quality=N&tag={etag}`
- Streaming URL: `/Videos/{id}/stream.{container}?static=true&api_key={token}` or via HLS master
- PascalCase JSON field names throughout (not camelCase)

### Shared Package Contents Required
```
github.com/jellyfinhanced/shared/
├── auth/         — Middleware + context key types
├── db/           — *sqlx.DB factory, transaction helpers
├── dto/          — All shared response struct definitions
├── config/       — XML config file reader (system.xml, network.xml)
├── response/     — JSON write helpers, error response builder
└── pagination/   — StartIndex/Limit extraction from query params
```

### Go Version & Dependencies
- Target: **Go 1.23+**
- Router: **go-chi/chi v5** (lightweight, stdlib-compatible, route param extraction matches C# patterns)
- Database: **jmoiron/sqlx** + **go-sql-driver/mysql** (or **sqlc** for generated type-safe queries)
- Logging: **log/slog** (stdlib, structured JSON)
- Testing: **testify/assert** + **testcontainers-go** for integration tests
- Config: **encoding/xml** stdlib (no external XML library needed)
- Crypto: **golang.org/x/crypto/bcrypt** (password hashing matches C# BCrypt usage)
- WebSocket (session/sync): **gorilla/websocket**

---

*End of Goose AI Prompt*
