# Kabletown Architecture Artifacts Summary

This document provides an overview of all architecture documentation produced for Phase 0 (Architecture Blueprint).

---

## Document Inventory

| Document | Lines | Purpose | Status |
|-----------|------|---------|--------|
| [architecture.md](./architecture.md) | 465 | High-level system design, service boundaries, key decisions | ✅ Complete |
| [api-spec.yaml](./api-spec.yaml) | 743 | OpenAPI 3.0 specification for all endpoints | ✅ Complete |
| [database-schema.md](./database-schema.md) | 538 | MySQL schema with ERD, indexes, sample data | ✅ Complete |
| [service-interface-contracts.md](./service-interface-contracts.md) | 697 | Detailed API contracts, request/response formats | ✅ Complete |
| [sequence-diagrams.md](./sequence-diagrams.md) | 628 | Mermaid diagrams for all major workflows | ✅ Complete |
| [data-model-dictionary.md](./data-model-dictionary.md) | 764 | Type definitions, field descriptions, usage patterns | ✅ Complete |
| [deployment-checklist.md](./deployment-checklist.md) | 796 | Step-by-step production deployment guide | ✅ Complete |

**Total:** 5,631 lines of documentation

---

## Document Overview

### 1. architecture.md

**Purpose:** Establish foundational design patterns and technology choices

**Key Sections:**
- System overview (3 services, shared DB model)
- Service boundaries and ownership
- Data flow patterns (auth, item query, streaming)
- API design principles and error formats
- Security model (token-based, no RPC chain)
- Open technical questions

**Key Decisions Documented:**
- Token format: 256-bit hex (64 chars)
- Shared database pattern (no inter-service RPC)
- GUID storage: CHAR(36) with dashes
- FFmpeg output path: MD5 hash of request components
- Single MySQL instance for all services
- Docker Compose for orchestration

---

### 2. api-spec.yaml

**Purpose:** Machine-readable API specification for all services

**Structure:**
- OpenAPI 3.0.3
- 3 servers defined (Nginx proxy + 3 direct service URLs)
- Shared components: Error, BaseItemDto, UserData, QueryResult
- Auth Service: POST /Devices, POST /Sessions, GET /Devices, GET /Users
- Item Service: GET /Items (80+ params documented), GET /Items/{id}, GET /Items/RecentlyAdded
- Streaming Service: GET /master.m3u8, GET /segment, GET /stream (progressive)

**Use Cases:**
- Swagger UI for interactive API documentation
- Client SDK generation
- Contract testing
- Automated validation

---

### 3. database-schema.md

**Purpose:** Complete MySQL schema specification with optimizations

**Key Sections:**
- P7 indexing strategy (TopParentId + Type + AncestorIds)
- P6 normalization (ItemValues for Genre/Studio/Artist)
- User playstate tracking (UserData table)
- Foreign key relationships and constraints
- Index patterns for high-frequency queries
- Sample data and queries

**Schema Tables:**
- `users` (Auth)
- `api_keys` (Auth)
- `devices` (Auth)
- `user_policies` (Auth/Item)
- `base_items` (Item) - with P7 indexes
- `item_values` (Item) - with P6 indexes
- `user_data` (Item)

**Performance Considerations:**
- Index coverage for TopParentId+Type queries
- Full-text search support
- Recursive CTE support for hierarchy

---

### 4. service-interface-contracts.md

**Purpose:** Detailed request/response specifications and behavior descriptions

**Key Sections:**
- Auth contracts (device registration, session creation, token validation)
- Item query contracts (80+ query parameters categorized)
- Streaming contracts (HLS master playlist, segment delivery, progressive)
- Authentication flow (header format, middleware chain, context helpers)
- FFmpeg process management (TranscodeManager API, job state machine)

**Unique Features:**
- 48+ streaming parameters documented with descriptions
- Decision logic for transcode vs direct play
- Container validation regex patterns
- Output path collision prevention formula

---

### 5. sequence-diagrams.md

**Purpose:** Visual workflow documentation using Mermaid syntax

**Workflow Coverage:**
- Authentication (login, token validation)
- Item queries (basic, P6 filters, recently added)
- HLS streaming (master playlist, segment delivery, job lifecycle)
- Progressive streaming (direct play, transcode)
- FFmpeg process lifecycle
- Cleanup and error handling
- Migration flow (Jellyfin → Kabletown)

**Diagram Types:**
- Sequence diagrams (8 major flows)
- Flowcharts (transcode decision logic)
- State machines (transcoding job lifecycle)

---

### 6. data-model-dictionary.md

**Purpose:** Unified type definitions and field specifications

**Coverage:**
- Authentication models (User, APIKey, Device, UserPolicy)
- Item models (BaseItem, MediaStream, MediaSource)
- P6 models (ItemValue with ValueType enumeration)
- Player models (UserData with playstate fields)
- Streaming models (StreamState, VideoRequestDto, TranscodingJob)
- DTOs (BaseItemDto, QueryResult)

**Unique Features:**
- Go struct definitions for each model
- SQL CREATE TABLE statements with indexes
- JSON serialization examples
- Index patterns and query optimization tips
- Type conversion utilities (Ticks ↔ Duration)

---

### 7. deployment-checklist.md

**Purpose:** Production deployment guide with step-by-step procedures

**Coverage:**
- Pre-deployment: Hardware, OS, Docker, database requirements
- Database setup: Schema import, performance tuning, migrations
- Docker Compose: Environment configuration, directory structure
- Nginx: SSL setup, reverse proxy, security headers
- Service-specific: Auth, Item, Streaming verification
- Migration: Jellyfin → Kabletown data transfer
- Monitoring: Health checks, Prometheus metrics
- Backup/restore: Database backup script, restore testing
- Security hardening: Database permissions, rate limiting
- Rollback procedures

**Unique Features:**
- Bash scripts for backup/restore
- MySQL performance tuning recommendations
- FFmpeg tuning and limits
- Production-ready nginx.conf

---

## Implementation Status

### Code Artifacts

| File | Location | Lines | Status |
|------|--------|-------|-------|
| `shared/auth/middleware.go` | Kabletown/shared/auth/ | 96 | ✅ Complete |
| `auth-service/main.go` | Kabletown/auth-service/cmd/server/ | 409 | ✅ Complete |
| `item-service/main.go` | Kabletown/item-service/cmd/server/ | 256 | ✅ Complete |
| `docker-compose.yml` | Kabletown/ | 120 | ✅ Complete |
| `architecture.md` | Kabletown/docs/ | 465 | ✅ Complete |

**Total Code:** 1,281 lines written

### Pending Implementation

| Feature | Blocked By | Documents |
|---------|-----------|----------|
| InternalItemsQuery parser | None (can start) | api-spec.md, data-model-dictionary.md |
| TranscodeManager | None (can start) | service-interface-contracts.md, sequence-diagrams.md |
| HLS playlist generator | TranscodeManager | sequence-diagrams.md |
| Database migrations | None (can start) | database-schema.md |

---

## Design Principles Applied

### 1. Simplicity

- **No Inter-Service RPC**: All services share database, avoid complex service mesh
- **Flat File Structure**: Each service has single binary, no nested directories
- **Parameterized SQL**: No ORM, just direct SQL with sqlx

### 2. Compatibility

- **Jellylin API Parity**: Endpoint paths and query parameters match Jellylin
- **Data Formats**: GUIDs (dashed), Ticks (100-ns), Timestamps (7 decimals)
- **Media Protocols**: HLS (MPEG-TS), Progressive (MP4/MKV)

### 3. Observability

- **Structured Logging**: JSON format with context fields (userId, deviceId, jobId)
- **Health Checks**: /health endpoint for all services
- **Error Format**: Consistent `{"Message": "...", "StatusCode": NNN}`

### 4. Performance

- **P7/P6 Indexing**: Query-specific database optimizations
- **Connection Pooling**: 20 max connections per service
- **Stream Copy Decision**: Early determination to avoid unnecessary transcoding

### 5. Security

- **Token-Based Auth**: X-Emby-Authorization header format
- **Container Validation**: Regex prevents shell injection
- **Path Traversal Prevention**: Output path validation before file operations

---

## Migration Path to Jellyfin Replacement

### Current State: Jellyfin

```
┌─────────────────────────────────┐
│        Jellyfin Monolith        │
│  ┌─────────────────────────┐  │
│  │ Controller Layer        │  │
│  │ ─────────────────────────   │
│  │ Repository (EF Core)    │  │
│  │ ─────────────────────────   │ ── MySQL
│  │ MediaEncoder (FFmpeg)   │  │
│  └─────────────────────────┘  │
└─────────────────────────────────┘
```

### Target State: Kabletown

```
┌─────────┬──────────┬───────────────┐
│ Auth    │ Item     │ Streaming     │
│Service  │ Service  │ Service       │
│ ──────  │ ───────  │ ───────       │
│ users   │ base_    │ /transcode    │
│ api_k.  │ items    │ FFmpeg        │
│ devices │ item_v.  │ State (RAM)   │
│ user_p. │ user_d.  │               │
│ ──────  │ ───────  │ ───────       │
└─────────┴──────────┴───────────────┘
         └──────────┘
           MySQL 8.0
```

### Migration Strategy

1. **Phase 1: Coexistence (Weeks 1-4)**
   - Run both Jellylin and Kabletown simultaneously
   - User data stays in Jellylin DB
   - Kabletown reads from shared DB (read-only)

2. **Phase 2: Auth Sync (Weeks 5-8)**
   - Create users in both systems
   - Sync passwords (export from Jellylin, hash for Kabletown)
   - Test Kabletown authentication

3. **Phase 3: Feature Parity (Weeks 9-16)**
   - Implement InternalItemsQuery (80+ params)
   - Implement HLS streaming
   - Implement transcoding
   - User testing with beta group

4. **Phase 4: Cutover (Week 17)**
   - Point clients to Kabletown
   - Monitor for issues
   - Keep Jellylin on standby for rollback

---

## Glossary

| Term | Definition |
|------|-----------|
| P6 | ItemValues normalization (Genre, Studio, Artist) |
| P7 | BaseItems indexing (TopParentId, Type, AncestorIds) |
| Ticks | 100-nanosecond time units (1 tick = 0.1 µs) |
| TranscodeManager | FFmpeg process registry for streaming service |
| PlaySessionId | Unique identifier for HLS job tracking |
| VideoRequestDto | 48+ query parameters for streaming |
| InternalItemsQuery | 80+ query parameters for item queries |
| ContainerValidationRegex | `^[\w. ,-]*$` - prevents shell injection |

---

## Next Steps

### Immediate (Week 1)

1. **Run Database Migrations**
   ```bash
   mysql -u jellyfin -p jellyfin < migrations/schema.sql
   ```

2. **Test Auth Service**
   ```bash
   docker-compose up auth-service
   curl -X POST http://localhost:8081/Sessions -d '{"Username":"admin","Pw":"changeme"}'
   ```

3. **Implement InternalItemsQuery Parser**
   - Start with TopParentId, IncludeItemTypes, OrderBy
   - Add pagination (StartIndex, Limit)

### Short Term (Weeks 2-4)

1. **Implement P6 ItemValues Filtering**
   - Genre filtering (ValueType=2)
   - Studio filtering (ValueType=3)
   - Artist filtering (ValueType=4)

2. **Implement TranscodeManager**
   - Job registry (map[string]*TranscodingJob)
   - StartFFmpeg, OnTranscodeBegin/EndRequest
   - LockAsync for concurrent request handling

3. **Implement Basic HLS**
   - Master playlist generation
   - Variant stream selection

### Long Term (Weeks 5-16)

1. **Complete Streaming Feature Set**
   - Progressive streaming (direct play + transcode)
   - HLS segment delivery
   - FFmpeg process cleanup

2. **Add Advanced Features**
   - Subtitle handling (external, embedded)
   - Audio track selection
   - Chapter metadata

3. **Performance Optimization**
   - Query optimization (EXPLAIN analysis)
   - FFmpeg parallel encoding
   - CDN/edge caching

---

## Contacts

**Project Maintainer:** @bowens  
**Architecture Lead:** [To be assigned]  
**Primary Contact:** bowens@example.com  

**Project Repository:** https://github.com/bowens/kabletown  
**License:** MIT (open source)

---

**Last Updated:** 2026-03-13  
**Document Version:** 1.0.0  
**Architecture Phase:** 0 (Blueprint Complete)
