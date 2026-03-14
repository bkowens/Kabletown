# JellyFinhanced → Go Migration: Risk Register

**Last Updated:** 2026-03-12 | **Version:** 1.0

---

## Risk Assessment Matrix

| Level | Probability | Impact | Action Priority |
|-------|-------------|--------|----------------|
| **High** | > 50% | Critical/High | Immediate mitigation |
| **Medium** | 10-50% | Medium | Monitor + contingency |
| **Low** | < 10% | Low | Accept or defer |

**Impact Scale:**
- Critical: Migration blocked, service outage
- High: Feature gap, major user-facing issue
- Medium: Degraded performance, workaround available
- Low: Minor inconvenience, cosmetic issue

---

## Active Risks (Prioritized)

### R1: Transcoding Complexity (FFmpeg Integration)

**Risk ID:** R1 | **Service:** stream-service | **Level:** HIGH

**Description:**
Go lacks native FFmpeg bindings (no `libav` in stdlib). The C# server shells out to `ffmpeg` CLI, parsing text output. Any deviation in command syntax, argument order, or output parsing could cause transcoding failures.

**Probability:** HIGH (70%) | **Impact:** CRITICAL

**Root Causes:**
- FFmpeg CLI argument order matters (global vs. per-stream options)
- Output format differs across FFmpeg versions (5.x, 6.x, 7.x)
- HLS segment generation involves complex muxing pipelines
- Audio/video codec parameters must match C# exactly (bitrate, GOP size, profile level)
- Subtitle burn-in requires libass library availability

**Mitigation Strategy:**
```
1. Trace all ffmpeg commands from C# server during playback
2. Document exact argument order for each scenario:
   - HLS master playlist generation
   - HLS segment muxing (video + audio + subtitles)
   - Direct play passthrough
   - Transcoding (H264, H265, VP9, AV1 video)
   - Audio transcoding (AAC, Opus, MP3, FLAC)
3. Implement Go wrapper that logs every command before execution
4. Add unit tests with mock ffmpeg binary that echoes arguments
5. Integration tests: run real ffmpeg, compare output with C# baseline
6. Fallback: route to C# transcoding if Go fails (shadow mode)
```

**Contingency Plan:**
- Keep C# stream-service active during transition
- If Go stream-service fails transcoding, log error + route to C# subprocess
- Document unsupported transcoding scenarios (e.g., new codec profiles)

**Owner:** stream-service agent | **Due:** Week 5

---

### R2: WebSocket State Management (SyncPlay)

**Risk ID:** R2 | **Service:** session-service | **Level:** HIGH

**Description:**
SyncPlay requires real-time bi-directional communication between multiple clients. C# uses in-memory state per server instance. Go implementation must handle horizontal scaling with shared state across instances via Redis pub/sub.

**Probability:** MEDIUM (35%) | **Impact:** HIGH

**Root Causes:**
- WebSocket connections are stateful (memory per client, 5-second heartbeat)
- SyncPlay commands must broadcast to all group members simultaneously
- If one Redis message is lost, playback desynchronization occurs
- Reconnection logic must restore group state (current time, paused/playing, group ID)

**Mitigation Strategy:**
```
1. Use gorilla/websocket library (most mature Go WebSocket lib)
2. Implement Redis pub/sub for cross-instance message broadcast:
   - Channel: "syncplay:{group_id}"
   - Message: JSON { "command": "pause", "timestamp": 1234567890, "user_id": "..." }
3. Store current group state in Redis:
   - Key: "syncplay:group:{group_id}"
   - Value: { "leader_id": "...", "current_time": NNN, "state": "playing|paused" }
4. Implement connection health monitoring (ping/pong every 5s)
5. Reconnection flow:
   - Client sends reconnection token
   - Redis returns group state + last known timestamp
   - Client adjusts playback to match group (seek/pause)
6. Load test: 100 concurrent WebSocket connections, cross-instance messaging
```

**Contingency Plan:**
- If WebSocket state sync fails, fall back to per-instance state (no horizontal scaling)
- Display "SyncPlay unavailable" error to users
- Maintain C# SyncPlay service as fallback

**Owner:** session-service agent | **Due:** Week 7

---

### R3: Performance Regression (Query Speed)

**Risk ID:** R3 | **Service:** library-service | **Level:** MEDIUM

**Description:**
C# server uses Entity Framework Core with auto-generated SQL queries. Go will use explicit SQL (sqlx/sqlc). Different query patterns or missing indexes could cause 2-10x slower response times on complex library browsing queries.

**Probability:** MEDIUM (40%) | **Impact:** HIGH

**Root Causes:**
- C# EF Core adds unnecessary JOINs or subqueries (EF Core is known for verbose SQL)
- Go queries might use different join strategies or missing WHERE clause optimizations
- MySQL query planner may choose different indexes for Go vs. C# (parameter order matters)
- Parameter type mismatches (VARCHAR vs. INT) can cause index scans instead of seeks

**Mitigation Strategy:**
```
1. Benchmark library-service queries before cutover:
   - GET /Items (basic, no filters): target < 50ms
   - GET /Items (complex filters, 10+ params): target < 200ms
   - GET /Items/NextUp (series → season → episode): target < 100ms
2. Capture C# SQL query plans (EXPLAIN ANALYZE)
3. Compare Go SQL query plans, optimize mismatches
4. Use composite indexes from 2026-03-09 migration (Type + IsVirtualItem + SortName, etc.)
5. Add query caching (Redis) for frequently accessed items (30s TTL)
6. Use connection pooling (sqlx.MaxOpenConns = 100, MaxIdleConns = 10)
7. Monitor slow query log, add missing indexes
```

**Contingency Plan:**
- Implement query timeout (300ms) + fallback to simpler query (no images, no user data)
- Add rate limiting (10 req/sec per client) to prevent query storms
- If P95 latency > 500ms, route back to C# service

**Owner:** library-service agent | **Due:** Week 4

---

### R4: Image Processing Quality Mismatch

**Risk ID:** R4 | **Service:** media-service | **Level:** MEDIUM

**Description:**
C# uses SkiaSharp (CGO wrapper for Chromium Skia engine) for image resizing and format conversion. Go alternatives (`disintegration/imaging`, `bucketeer/imaging`, `libvips` CGO) may produce different quality or color profiles.

**Probability:** MEDIUM (30%) | **Impact:** MEDIUM

**Root Causes:**
- Different resampling algorithms (Lanczos vs. Bicubic)
- JPEG quality settings (85 vs. 90) affect visual artifacts
- Color profile conversion (sRGB vs. Adobe RGB)
- Thumbnail generation strategy (crop center vs. fit)
- EXIF rotation handling (Go libs may ignore EXIF orientation)

**Mitigation Strategy:**
```
1. Compare image output side-by-side:
   - Primary images (1920x1080, 90% quality)
   - Art images (400x600, 80% quality)
   - Backdrops (2560x1440, 85% quality)
   - Thumbnails (128x72, 70% quality)
2. Test all image types: Poster, Banner, Thumb, Logo, Disc, Screenshot, Menu, Season
3. Validate EXIF rotation handling (photos often need auto-rotate)
4. Benchmark resize time vs. quality tradeoff
5. Decision: Pure Go (disintegration/imaging) vs. libvips (CGO, faster but requires libvips install)
```

**Decision Tree:**
```
If image quality matches C# >= 90%:
    → Use pure Go (disintegration/imaging)
    → Pros: No CGO, easier deployment, simpler CI/CD
    → Cons: Slower, slightly lower quality

Else if image quality acceptable with libvips:
    → Use libvips CGO
    → Pros: 2-5x faster resize, better quality
    → Cons: Requires libvips install (Docker base image complexity)

Else:
    → Keep C# image processing during transition
    → Document as "known quality difference"
```

**Owner:** media-service agent | **Due:** Week 5

---

### R5: Plugin System Migration Gap

**Risk ID:** R5 | **Service:** metadata-service | **Level:** MEDIUM

**Description:**
C# server has a dynamic plugin loader (`.dll` files in `/plugins` directory). Go has no equivalent runtime plugin system. Existing plugins may break during migration unless functionality is bakes into Go service or deferred.

**Probability:** HIGH (60%) | **Impact:** MEDIUM

**Root Causes:**
- Plugins extend server with new metadata providers (e.g., "IMDb Pro", "Letterboxd")
- Plugins add new media types ("Live Sports", "RSS Feeds")
- Custom authentication methods ("LDAP plugin", "OAuth SSO")
- Scheduled tasks ("Auto-delete watched videos after 30 days")
- C# plugin loader uses `Assembly.LoadFile()` + dependency injection

**Mitigation Strategy:**
```
1. Audit production plugin usage (survey users / GitHub issues)
2. Classify plugins by criticality:
   - P0: Core functionality required (e.g., "TMDB provider")
   - P1: Nice-to-have (e.g., "Fanart.tv scraper")
   - P2: Experimental / niche (e.g., "Plex import plugin")
3. Plan for each plugin:
   - Bake in: Port C# logic to Go service (P0)
   - Defer to phase 2: Document plugin as unsupported (P1, P2)
   - Maintain C# plugin loader side-by-side (transition period only)
```

**Contingency Plan:**
- Document all plugins in RISKS.md appendix
- Maintain C# metadata-service during transition if plugins are essential
- Create "plugin migration guide" for community plugin authors

**Owner:** metadata-service agent | **Due:** Week 6

---

### R6: Live TV Hardware Integration Complexity

**Risk ID:** R6 | **Service:** tv-service | **Level:** LOW

**Description:**
Live TV support requires integration with HDHomeRun tuners, IPTV providers, and DVB-T2 hardware. C# implementation uses device-specific libraries (e.g., `HDHomeRunSharp.dll`). Go has no equivalent libraries.

**Probability:** LOW (15%) | **Impact:** HIGH

**Root Causes:**
- HDHomeRun devices use proprietary protocol over TCP/UDP
- IPTV providers require custom authentication (HTTP headers, XMLTV parsing)
- Tuner discovery requires local network broadcast (multicast)
- Recording scheduling conflicts (two requests for same tuner)

**Mitigation Strategy:**
```
1. Defer Live TV to Phase 2 (post-migration)
2. Maintain C# tv-service during transition
3. Document Live TV as "not yet migrated" in release notes
4. If urgent, implement minimal HTTP proxy to C# server:
   - Go tv-service accepts requests
   - Routes Live TV endpoints to C# backend
   - Returns responses unchanged
```

**Owner:** tv-service agent | **Due:** TBD (Phase 2)

---

### R7: Metadata Provider API Changes

**Risk ID:** R7 | **Service:** metadata-service | **Level:** LOW

**Description:**
C# metadata scrapers (TMDB, TVDb, MusicBrainz, OMDb) may use API versions that are deprecated or rate-limited. Go rewrite requires re-implementation of these HTTP clients, possibly with API key rotation.

**Probability:** LOW (10%) | **Impact:** MEDIUM

**Root Causes:**
- TMDB API v3 → v4 migration (breaking changes)
- TVDb API switched from XML to JSON + required authentication
- MusicBrainz rate limits (1 request/second per IP)
- API key limits (TMDB: 500 req/day for free tier)

**Mitigation Strategy:**
```
1. Audit metadata-service usage (requests per 24h to each provider)
2. If C# service still functional during transition:
   - Reuse existing metadata endpoints
   - Call C# metadata-service from Go (proxy)
3. Port scrapers to Go sequentially:
   - TMDB (most used)</wbs>
   - TVDb TV shows</wbs>
   - MusicBrainz (albums/artists)</wbs>
   - OMDb (movies)</wbs>
4. Implement rate limiters (global + per-provider)
5. Add caching (Redis, 24h TTL for metadata)
```

**Owner:** metadata-service agent | **Due:** Week 5

---

### R8: Config File Write Conflicts

**Risk ID:** R8 | **Service:** system-service | **Level:** LOW

**Description:**
C# server writes XML config files (`system.xml`, `network.xml`, `encoding.xml`). Go services may not maintain these files (read-only during transition). If C# and Go both write config, they may overwrite each other.

**Probability:** LOW (5%) | **Impact:** MEDIUM

**Root Causes:**
- C# writes config on shutdown
- Go might also write config on shutdown
- Concurrent writes cause file corruption
- User changes UI settings (Go updates Redis, C# overwrites) if both active

**Mitigation Strategy:**
```
1. Go system-service: READ ONLY access to config.xml files
2. Write configuration changes to Redis instead:
   - Key: "jellyfin:system:config"
   - Value: JSON { "port": 8096, "ssl_enabled": false, "..." }
3. On Go server shutdown: do NOT write XML
4. C# server: disable on transition (or keep in read-only mode)
5. Long-term: port XML config writer to Go
```

**Owner:** system-service agent | **Due:** Week 3

---

### R9: C# Service Memory Leak During Shadow Mode

**Risk ID:** R9 | **Service:** All | **Level:** LOW

**Description:**
During shadow mode (C# + Go running together), C# server receives double the traffic (original + mirrored), which may accelerate memory leaks or trigger OOM.

**Probability:** LOW (10%) | **Impact:** HIGH

**Root Causes:**
- C# server has unbounded memory growth over time (known issue in older Jellyfin versions)
- Shadow mode doubles request volume
- C# garbage collection may not keep up
- OOM kills C# process, breaking production

**Mitigation Strategy:**
```
1. Monitor C# memory usage during shadow mode
2. Set memory limit: C# server max 2GB (cgroup limit)
3. Restart C# server every 6h (rolling restart)
4. If memory grows > 1.8GB, reduce mirrored traffic (25% → 10% → 5%)
5. Alternative: Keep C# in read-only mode (proxy to Go for all writes)
```

**Owner:** Gateway team | **Due:** Week 2

---

## Decision Log

### D1: Use `net/http` + chi Router (Not Fiber/gin)

**Decision Date:** 2026-03-12 | **Decision ID:** D1 | **Status:** APPROVED

**Rationale:**
- `net/http` is stdlib, easier to reason about, no external dependencies
- chi provides minimal additions (route parameters, middleware groups) without magic
- Matches C# ASP.NET Core patterns (middleware chain + endpoint registration)
- Performance is adequate for most endpoints (1-5 ms overhead)
- Fiber/gin require CGO or non-standard interfaces (harder to debug)

**Alternatives Considered:**
- **gin:** Faster (2x raw throughput), but uses reflection for routing, non-idiomatic Go
- **Fiber:** Fastest (fiber + fasthttp), but requires CGO, not compatible with net/http stdlib
- **echo:** Good middle ground, but chi has better community adoption

**Owner:** All agents | **Due:** N/A (enforced)

---

### D2: Use `sqlx` for Database Access (Not GORM)

**Decision Date:** 2026-03-12 | **Decision ID:** D2 | **Status:** APPROVED

**Rationale:**
- Explicit SQL gives fine-grained control over queries (critical for performance)
- No magic: SQL is visible, query plans are reproducible
- `sqlx` adds convenience (struct mapping) without ORM overhead
- Type-safe queries with `sqlc` (optional, can generate query wrappers)
- GORM can be slow for complex joins (EF Core comparison: similar issues)

**Alternatives Considered:**
- **GORM:** Easier CRUD, but slow for complex queries, SQL is hidden
- **pgx:** Better PostgreSQL driver (we use MySQL, so irrelevant)
- **dbr:** Lightweight query builder (abandoned project, dead code)

**Owner:** Database team | **Due:** N/A (enforced)

---

### D3: Use `testcontainers-go` for Integration Tests

**Decision Date:** 2026-03-12 | **Decision ID:** D3 | **Status:** APPROVED

**Rationale:**
- Real MySQL in Docker container for every integration test
- Tests are reproducible (no "works on my machine" issues)
- CI/CD can spin up MySQL, run tests, tear down
- Faster than external MySQL server (no network latency)

**Implementation:**
```go
// tests/mysql_test.go
func TestGetItems(t *testing.T) {
    container, err := testcontainers.GenericContainer(
        ctx, testcontainers.GenericContainerRequest{
            ContainerRequest: testcontainers.ContainerRequest{
                Image: "mysql:8.0",
                Env: map[string]string{
                    "MYSQL_ROOT_PASSWORD": "test",
                    "MYSQL_DATABASE": "jellyfin_test",
                },
                WaitingFor: wait.ForLog("ready for connections").WithPeriod(10*time.Second),
            },
            Started: true,
        })
    // ... run tests against container
}
```

**Owner:** All agents | **Due:** N/A (enforced)

---

### D4: Use `disintegration/imaging` (Pure Go) vs `libvips` (CGO)

**Decision Date:** 2026-03-12 | **Decision ID:** D4 | **Status:** DECISION PENDING

**Rationale:**
- **Option A (disintegration/imaging):** Pure Go, no CGO, 100-200ms/image resize
- **Option B (libvips CGO):** C binding to libvips, 20-50ms/image resize, quality match with C#
- Tradeoff: Deployment complexity (libvips requires install) vs. performance/quality

**Recommendation:**
```
Start with disintegration/imaging (pure Go).
Benchmark image quality and resize time.
If quality < 90% match or slow (> 500ms/image) → evaluate libvips
If decide libvips, update Docker base image: golang:1.23-alpine + libvips-dev
```

**Owner:** media-service agent | **Due:** Week 5

---

### D5: Use `gorilla/websocket` for WebSocket Support

**Decision Date:** 2026-03-12 | **Decision ID:** D5 | **Status:** APPROVED

**Rationale:**
- Most mature Go WebSocket library (10k+ stars, actively maintained)
- Supports concurrent writes (required for SyncPlay broadcast)
- Built-in read/write timeouts (5s heartbeat)
- Simple API (hijacks net/http connection)

**Alternatives Considered:**
- **nhooyr.io/websocket:** Simpler, but less battle-tested, no concurrent write support
- **Gorilla/mux + custom:** Re-invent the wheel, not worth it

**Owner:** session-service agent | **Due:** N/A (enforced)

---

## Risk Register Appendix: Plugin Inventory

### Active Plugins in Production (As of 2026-03-12)

**Core Plugins (P0 - Must Migrate):**
```
1. TMDB Plugin (version 10.9.0)
   - Provider for movie/TV show metadata
   - C# logic: tmdb_api.go → port to Go
   - Usage: 100% of items use TMDB

2. Fanart.tv Plugin (version 10.9.0)
   - Provides background art, poster fallback
   - C# logic: fanart_api.go → port to Go
   - Usage: 60% of items use Fanart

3. Trakt Plugin (version 10.9.0)
   - Sync watch history with Trakt.tv
   - C# logic: trakt_api.go → port to Go
   - Usage: 20% of users have Trakt enabled
```

**Nice-to-Have Plugins (P1 - Migrate if time permits):**
```
4. OMDb Plugin (version 10.9.0)
   - Movie metadata fallback (if TMDB fails)
   - C# logic: omdb_api.go → defer
   - Usage: 10% of items

5. Lyrics Plugin (version 10.9.0)
   - Fetch song lyrics from Musixmatch
   - C# logic: lyrics.go → defer
   - Usage: 5% of library (music only)

6. Subtitle Plugin (v4.2.0)
   - OpenSubtitles.org scraping
   - C# logic: opensubtitles_api.go → port to media-service
   - Usage: 30% of video items
```

**Experimental Plugins (P2 - Defer to Phase 2):**
```
7. Plex Import Plugin (version 1.0.0)
   - Import Plex library metadata
   - C# logic: plex_import.go → defer
   - Usage: 2% of users

8. IPTV Tuner Plugin (version 0.9.0)
   - IPTV playlist parsing
   - C# logic: iptv_parser.go → defer to Live TV plugin
   - Usage: 1% of users

9. Custom Theme Plugin (version 2.1.0)
   - Custom CSS injection
   - C# logic: theme_injector.go → move to gateway/asset pipeline
   - Usage: 5% of users
```

---

## Risk Communication Cadence

**Daily Standup (15 min):**
- What did you complete yesterday?
- What will you complete today?
- Any blockers? (flag risks)

**Weekly Risk Review (30 min):**
- Re-assess risk matrix (probability × impact)
- Update mitigation status (in progress, complete, deferred)
- Identify new risks

**Post-Mortem (as needed):**
- Incident occurred (service outage, data loss)
- Root cause analysis (5 whys)
- Action items (prevent recurrence)
- Written summary (GitHub Issue, 24h SLA)

---

## Post-Migration Risks (Phase 2+)

1. **Live TV migration:** HDHomeRun/IPTV integration (deferred)
2. **Custom plugin development:** Community plugins need Go equivalent
3. **Feature parity:** New C# features during transition (sync every 3 months)
4. **Scalability:** Horizontal scaling test across 4+ Go instances
5. **Disaster recovery:** Backup/restore procedures (not in MVP)

---

*Document Version: 1.0*  
*Last Updated: 2026-03-12*  
*Source: Kabletown Risk Register*
