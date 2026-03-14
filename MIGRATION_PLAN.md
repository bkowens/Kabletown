# Jellyfin → Kabletown: API Migration Plan

## Executive Summary

This document outlines the plan for migrating the Jellyfin application from its current C#/ASP.NET Core architecture to a Go-based microservices architecture. The new API collection will be organized under the `Kabletown` subdirectory and will break down the existing monolithic controllers into individual, independent API services.

---

## 1. Current Codebase Analysis

### 1.1 Architecture Overview
- **Framework**: ASP.NET Core (net10.0)
- **Language**: C#
- **Primary Structure**: Monolithic API application with 60+ controllers
- **Location**: `/home/bowens/Code/JellyFinhanced/Jellyfin.Api`

### 1.2 Current Controller Inventory

| Controller | Route | Purpose | Complexity |
|------------|-------|---------|----------|
| ActivityLogController | `/System/ActivityLog` | Activity logging | Low |
| ApiKeyController | `/Auth/ApiKey` | API key management | Low |
| ArtistsController | `/Artists` | Artist metadata | Medium |
| AudioController | `/Audio` | Audio playback | High |
| BackupController | `/System/Backup` | Backup/restore | Medium |
| BrandingController | `/Branding` | Branding configuration | Low |
| ChannelsController | `/Channels` | Live TV channels | Medium |
| ClientLogController | `/Logging/Client` | Client logging | Low |
| CollectionController | `/Collections` | Collection management | Medium |
| ConfigurationController | `/System/Configuration` | System configuration | Medium |
| DashboardController | `/Dashboard` | Dashboard data | Medium |
| DevicesController | `/Devices` | Device management | Medium |
| DisplayPreferencesController | `/DisplayPreferences` | Display settings | Low |
| DynamicHlsController | `/Videos` | HLS streaming | Very High |
| EnvironmentController | `/Environment` | Environment info | Low |
| FilterController | `/Filters` | Filter definitions | Low |
| GenresController | `/Genres` | Genre metadata | Medium |
| HlsSegmentController | `/Videos` | HLS segments | High |
| ImageController | `/Items/{id}/Images` | Image management | Very High |
| InstantMixController | `/InstantMix` | Mix generation | Medium |
| ItemLookupController | `/Items/Lookup` | Item metadata lookup | Medium |
| ItemRefreshController | `/Items/Refresh` | Item refresh triggers | Low |
| ItemsController | `/Items` | Core item CRUD | Very High |
| ItemUpdateController | `/Items` | Item update operations | High |
| LibraryController | `/Library` | Library management | Very High |
| LibraryStructureController | `/Library/VirtualFolders` | Library structure | Medium |
| LiveTvController | `/LiveTv` | Live TV functionality | Very High |
| LocalizationController | `/Localization` | Localization data | Low |
| LyricsController | `/Lyrics` | Lyrics management | Medium |
| MediaInfoController | `/MediaInfo` | Media information | High |
| MediaSegmentsController | `/MediaSegments` | Media segment management | Low |
| MoviesController | `/Movies` | Movie metadata | Medium |
| MusicGenresController | `/MusicGenres` | Music genre metadata | Medium |
| PackageController | `/Packages` | Plugin/package management | Medium |
| PersonsController | `/Persons` | Person metadata | Medium |
| PlaylistsController | `/Playlists` | Playlist management | High |
| PlaystateController | `/PlayedItems` | Playback state | High |
| PluginsController | `/Plugins` | Plugin management | Medium |
| QuickConnectController | `/QuickConnect` | Quick connect auth | Low |
| RemoteImageController | `/Items/{id}/RemoteImages` | Remote image management | Medium |
| ScheduledTasksController | `/ScheduledTasks` | Task scheduling | Medium |
| SearchController | `/Search` | Search functionality | Medium |
| SessionController | `/Sessions` | Session management | High |
| StartupController | `/Startup` | Server startup | Medium |
| StudiosController | `/Studios` | Studio metadata | Medium |
| SubtitleController | `/Videos/{id}/Subtitles` | Subtitle management | High |
| SuggestionsController | `/Suggestions` | Content suggestions | Low |
| SyncPlayController | `/SyncPlay` | Synchronized playback | High |
| SystemController | `/System` | System management | Medium |
| TimeSyncController | `/TimeSync` | Time synchronization | Low |
| TrailersController | `/Trailers` | Trailer content | Low |
| TrickplayController | `/Trickplay` | Trick play images | Medium |
| TvShowsController | `/Shows` | TV show metadata | Medium |
| UniversalAudioController | `/Audio` | Universal audio playback | High |
| UserController | `/Users` | User management | High |
| UserLibraryController | `/Users/{userId}/Views` | User library views | High |
| UserViewsController | `/Users/{userId}/Views` | User view configuration | Medium |
| VideoAttachmentsController | `/Videos/{id}/Attachments` | Video attachments | Low |
| VideosController | `/Videos` | Video playback | Very High |
| YearsController | `/Years` | Year metadata | Low |

### 1.3 Core Services/Interfaces Identified

Controllers depend on the following core service abstractions:
- `IUserManager` - User management
- `ILibraryManager` - Library/Item management
- `ISessionManager` - Session handling
- `IUserDataManager` - User data persistence
- `IDtoService` - DTO generation
- `ILocalizationManager` - Localization
- `IImageProcessor` - Image processing
- `IMediaSourceManager` - Media handling
- `ISubtitleManager` - Subtitle operations
- `IChannelManager` - Channel operations

---

## 2. New Architecture: Kabletown

### 2.1 Architectural Philosophy

The new Go-based API will follow microservices principles:

1. **Single Responsibility**: Each API service handles one domain
2. **Independent Deployability**: Services can be deployed independently
3. **API-First Design**: Well-defined RESTful APIs with OpenAPI/Swagger
4. **Domain-Driven**: Organized by business domain
5. **Modern Go Practices**: Go 1.22+, standard library + minimal dependencies

### 2.2 Service Decomposition

The 60+ controllers will be consolidated into the following API services:

#### Core Services (Essential)
1. **items-api** - Media items CRUD, search, metadata
2. **users-api** - User authentication, profiles, permissions
3. **library-api** - Library structure, folders, collections
4. **playback-api** - Audio/video streaming, playstate
5. **sessions-api** - Device sessions, remote control

#### Content Services
6. **media-api** - Media info, codecs, transcoding
7. **images-api** - Image serving, generation, management
8. **subtitles-api** - Subtitle management, serving
9. **lyrics-api** - Lyrics content

#### Discovery & Browsing
10. **search-api** - Global search, hints, suggestions
11. **metadata-api** - Artists, genres, studios, persons, years
12. **tv-api** - Live TV, EPG, channels, recordings
13. **collections-api** - Collections, playlists, folders

#### System Services
14. **system-api** - System info, configuration, tasks
15. **plugins-api** - Plugin management, updates
16. **devices-api** - Device registration, management
17. **backup-api** - Backup, restore, export

#### Utilities
18. **auth-api** - Authentication, API keys, quick connect
19. **preferences-api** - Display preferences, user settings
20. **websockets-api** - Real-time communication hub

### 2.3 New Project Structure

```
/home/bowens/Code/JellyFinhanced/Kabletown/
├── apis/
│   ├── items/
│   │   ├── cmd/
│   │   │   └── server/
│   │   │       └── main.go
│   │   ├── internal/
│   │   │   ├── handlers/
│   │   │   ├── models/
│   │   │   ├── repository/
│   │   │   ├── services/
│   │   │   └── middleware/
│   │   ├── pkg/
│   │   │   └── api/
│   │   ├── openapi.yaml
│   │   ├── go.mod
│   │   └── README.md
│   ├── users/
│   ├── library/
│   ├── playback/
│   ├── sessions/
│   ├── media/
│   ├── images/
│   ├── subtitles/
│   ├── lyrics/
│   ├── search/
│   ├── metadata/
│   ├── tv/
│   ├── collections/
│   ├── system/
│   ├── plugins/
│   ├── devices/
│   ├── backup/
│   ├── auth/
│   ├── preferences/
│   └── websockets/
├── shared/
│   ├── models/
│   ├── middleware/
│   ├── auth/
│   ├── errors/
│   └── utils/
├── gateway/
│   ├── cmd/
│   │   └── gateway/
│   │       └── main.go
│   ├── internal/
│   │   ├── routing/
│   │   ├── auth/
│   │   └── rate limit/
│   ├── openapi.yaml
│   ├── go.mod
│   └── README.md
├── docker-compose.yml
├── Makefile
├── README.md
└── docs/
    ├── architecture.md
    ├── api-guidelines.md
    └── migration-guide.md
```

---

## 3. Implementation Phases

### Phase 1: Foundation (Weeks 1-4)

**Objective**: Establish project structure and core infrastructure

**Tasks**:
1. Create Kabletown directory structure
2. Set up Go workspace with `go.work`
3. Implement shared packages:
   - `shared/models` - Common data structures
   - `shared/auth` - Authentication utilities
   - `shared/middleware` - HTTP middleware
   - `shared/errors` - Error handling
4. Create API Gateway with basic routing
5. Implement OpenAPI base structure
6. Set up Docker development environment
7. Create Makefile with standard targets

**Deliverables**:
- Working project scaffold
- API Gateway with health checks
- Development container setup
- CI/CD pipeline skeleton

### Phase 2: Core APIs (Weeks 5-10)

**Objective**: Implement essential services for basic functionality

**Prioritized APIs**:
1. **auth-api** (Week 5-6)
   - User authentication
   - API key management
   - Quick Connect
   - JWT handling

2. **users-api** (Week 6-7)
   - User CRUD
   - User profiles
   - Permissions
   - Password management

3. **system-api** (Week 7-8)
   - System information
   - Server configuration
   - Activity logs
   - Environment info

4. **items-api** (Week 8-10)
   - Item listing with filters
   - Single item retrieval
   - Item search
   - Basic metadata

**Deliverables**:
- Functional auth flow
- User management
- Basic media browsing
- All APIs documented with OpenAPI

### Phase 3: Content & Playback (Weeks 11-16)

**Objective**: Implement media playback and content management

**Prioritized APIs**:
1. **playback-api** (Week 11-13)
   - Audio streaming
   - Video streaming
   - HLS segment serving
   - Play state tracking

2. **media-api** (Week 13-14)
   - Media information
   - Transcoding profiles
   - Codec info

3. **images-api** (Week 14-15)
   - Primary images
   - Thumbnail generation
   - Image tags
   - Remote images

4. **subtitles-api** (Week 15-16)
   - Subtitle listing
   - Subtitle serving
   - Subtitle upload

**Deliverables**:
- Full media playback
- Image serving pipeline
- Subtitle support
- Playback state persistence

### Phase 4: Discovery & Organization (Weeks 17-22)

**Objective**: Implement search, metadata, and organization features

**Prioritized APIs**:
1. **metadata-api** (Week 17-18)
   - Artists, albums
   - Genres, studios
   - Persons, years
   - Metadata scraping

2. **search-api** (Week 18-19)
   - Global search
   - Search hints
   - Filters
   - Suggestions

3. **collections-api** (Week 19-20)
   - Collections
   - Playlists
   - User views
   - Folder structure

4. **library-api** (Week 20-22)
   - Virtual folders
   - Library structure
   - Content refresh
   - Item lookup

**Deliverables**:
- Full metadata ecosystem
- Search functionality
- Collection management
- Library organization

### Phase 5: Advanced Features (Weeks 23-28)

**Objective**: Implement remaining features

**Prioritized APIs**:
1. **tv-api** (Week 23-25)
   - Live TV
   - Electronic Program Guide
   - Channel management
   - Recordings

2. **sessions-api** (Week 25-26)
   - Session management
   - Device control
   - Play state sync

3. **plugins-api** (Week 26-27)
   - Plugin system
   - Plugin marketplace
   - Updates

4. **backup-api** (Week 27-28)
   - Backup creation
   - Backup restore
   - Export/import

5. **Lyrics, Preferences, Devices** (Week 28)

**Deliverables**:
- Live TV support
- Plugin system
- Backup/restore
- Completed feature set

### Phase 6: Integration & Testing (Weeks 29-32)

**Objective**: Integrate with front-end and thorough testing

**Tasks**:
1. Update webclient to use new APIs
2. Performance optimization
3. Load testing
4. Security audit
5. Documentation completion
6. Migration tools

**Deliverables**:
- Fully functional application
- Performance benchmarks
- Security certification
- Migration documentation

---

## 4. API Design Guidelines

### 4.1 RESTful Standards

All APIs will follow REST principles:

```yaml
# Example: Get Items
GET /api/items
GET /api/items/{id}
POST /api/items
PATCH /api/items/{id}
DELETE /api/items/{id}

# Query parameters for filtering
?userId={userId}
&parentId={parentId}
&includeItemTypes=Movie,Series
&startIndex={index}
&limit={count}
&sortBy=SortName
&sortOrder=Ascending
```

### 4.2 Response Format

```json
{
  "Items": [...],
  "TotalRecordCount": 100,
  "StartIndex": 0
}
```

### 4.3 Error Handling

```json
{
  "ErrorCode": "NotFound",
  "Message": "Item not found",
  "StatusCode": 404,
  "ItemId": "abc-123"
}
```

### 4.4 Authentication

- API Key via header: `X-EMBY-authorization: MediaBrowser Token="{token}"`
- JWT for session-based auth
- OAuth 2.0 for external integrations

### 4.5 Versioning

URL-based versioning: `/api/v1/items`

---

## 5. Technology Stack

### 5.1 Backend (Go)
- **Language**: Go 1.22+
- **Web Framework**: `gorilla/mux` or `chi`
- **OpenAPI**: `swaggo` for documentation
- **Database**: Access existing Jellyfin database 
- **Caching**: Redis via client library
- **Testing**: `testify`, ` gomock`

### 5.2 Infrastructure
- **Container**: Docker
- **Orchestration**: Docker Compose (dev), Kubernetes (prod)
- **Service Mesh**: Optional (consider Linkerd)
- **Logging**: Structured JSON logging
- **Metrics**: Prometheus compatible

### 5.3 API Gateway
- **Technology**: Go-based custom gateway
- **Features**:
  - Rate limiting
  - Authentication passthrough
  - Request/response transformation
  - Circuit breaking
  - Load balancing

---

## 6. Data Migration Strategy

### 6.1 Database Approach

The new API will **not** create a new database layer. Instead:

1. **Read Operations**: Use existing SQLite/PostgreSQL/Mysql database directly via:
   - Repository pattern with read-only access
   - Query translation from Go to SQL

2. **Write Operations**: 
   - Option A: Continue using existing EF Core via shared library
   - Option B: Implement Go database layer (recommended)

### 6.2 Shared Database Schema

The Go APIs will share the existing Jellyfin database tables:
- `ItemValues` - Media item data
- `BaseItems` - Core item information
- `Users` - User accounts
- `Devices` - Device registry
- `ActivityLogs` - Activity tracking
- `AncestorIds` - Hierarchical relationships
- All other Jellyfin database tables

### 6.3 Migration Strategy

1. Keep existing database intact during transition
2. Read from existing tables via Go ORM (Ent or sqlc)
3. Implement gradual write-over:
   - Phase 1: Read-only APIs
   - Phase 2: Dual-write during transition
   - Phase 3: Full Go write layer

---

## 7. Frontend Integration

### 7.1 API Client Library

Create TypeScript/JavaScript client library:

```typescript
// Usage example
const client = new KabletownClient({
  baseUrl: 'http://localhost:8080',
  token: 'your-auth-token'
});

const items = await client.items.get({
  userId: 'user-id',
  includeItemTypes: ['Movie', 'Series'],
  limit: 20
});
```

### 7.2 Frontend Migration

The webclient will be updated to use the new APIs:

1. Replace direct service calls with API client
2. Update authentication flow
3. Implement WebSocket reconnection
4. Add API error handling
5. Performance monitoring

### 7.3 WebSocket Hub

```typescript
// WebSocket for real-time updates
const ws = new WebSocket('ws://localhost:8080/ws');

ws.onmessage = (event) => {
  const message = JSON.parse(event.data);
  // Handle ServerMessage types
  // GeneralMessage, UserDeleted, PlaybackStart, etc.
};
```

---

## 8. Migration Considerations

### 8.1 Breaking Changes

The new API may introduce breaking changes:

| Old C# API | New Go API | Migration |
|------------|------------|----------|
| `/Emby/Items` | `/api/v1/items` | URL prefix + version |
| `X-Emby-authorization` | `Authorization: MediaBrowser Token="..."` | Header format |
| WebSocket message format | Same format maintained | No change |

### 8.2 Compatibility Layer

During transition period:

```go
// API Gateway can route old paths to new services
map[string]string{
    "/Emby/Items": "/api/v1/items",
    "/Users": "/api/v1/users",
    // ...
}
```

### 8.3 Feature Parity Checklist

- [ ] All endpoints implemented
- [ ] All request/response formats migrated
- [ ] All WebSocket message types supported
- [ ] All authentication methods working
- [ ] Performance meets or exceeds current
- [ ] Security audit passed

---

## 9. Development Guidelines

### 9.1 Go Project Structure

```go
// cmd/server/main.go - Application entry point
package main

import (
    "kabletown/items/internal/handlers"
    "kabletown/items/internal/services"
    ...
)

func main() {
    // Setup dependency injection
    // Start HTTP server
    // Handle graceful shutdown
}

// internal/handlers/http_handler.go
package handlers

type ItemsHandler struct {
    service *services.ItemsService
}

func (h *ItemsHandler) GetItems(w http.ResponseWriter, r *http.Request) {
    // Handle request
}

// internal/services/items_service.go
package services

type ItemsService struct {
    repo repository.ItemsRepository
}

func (s *ItemsService) GetItems(ctx context.Context, params *GetItemsParams) (*QueryResult, error) {
    // Business logic
}
```

### 9.2 OpenAPI First

Generate Go code from OpenAPI spec:

```yaml
# openapi.yaml
openapi: 3.0.0
info:
  title: Items API
  version: 1.0.0

paths:
  /api/v1/items:
    get:
      summary: Get items
      parameters:
        - name: userId
          in: query
          schema:
            type: string
            format: uuid
      responses:
        '200':
          description: Successful response
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/QueryResult'
```

### 9.3 Testing Strategy

```bash
# Unit tests
go test ./internal/...

# Integration tests
go test ./tests/... -tags=integration

# Load tests
# Using go-vegeta or similar tool
```

---

## 10. Risk Mitigation

### 10.1 Technical Risks

| Risk | Impact | Mitigation |
|------|--------|-----------|
| Database compatibility issues | High | Use existing C# database layer via interop or sqlc |
| Performance degradation | High | Extensive load testing, profiling |
| Feature gaps during migration | Medium | Maintain compatibility layer |
| WebSocket complexity | Medium | Careful design of gateway routing |

### 10.2 Project Risks

| Risk | Impact | Mitigation |
|------|--------|-----------|
| Scope creep | High | Strict MVP definition, phased approach |
| Team knowledge gap | Medium | Training on Go best practices |
| Timeline overrun | Medium | Regular milestones, buffer time |

---

## 11. Success Criteria

### 11.1 Functional

- [ ] 100% API coverage matching existing functionality
- [ ] Frontend works without modifications
- [ ] All WebSocket messages functional
- [ ] All authentication methods working
- [ ] Plugin system supported

### 11.2 Non-Functional

- [ ] Performance: Same or better response times
- [ ] Throughput: Equal or greater requests/second
- [ ] Memory usage: <= current usage
- [ ] Startup time: <= current startup time
- [ ] API documentation: Complete and accurate

### 11.3 Quality

- [ ] >80% test coverage
- [ ] Security audit passed
- [ ] No critical bugs
- [ ] Documentation complete

---

## 12. Timeline Summary

| Phase | Duration | Key Deliverables |
|-------|----------|------------------|
| Foundation | 4 weeks | Project structure, gateway, dev env |
| Core APIs | 6 weeks | auth, users, system, items |
| Content & Playback | 6 weeks | playback, media, images, subtitles |
| Discovery & Org | 6 weeks | metadata, search, collections, library |
| Advanced Features | 6 weeks | tv, sessions, plugins, backup |
| Integration & Testing | 4 weeks | Frontend integration, load testing |
| **Total** | **32 weeks** | **Full migration complete** |

---

## 13. Next Steps

1. **Review and approve this plan**
2. **Set up repository structure**
3. **Create initial project scaffold**
4. **Establish development environment**
5. **Begin Phase 1 implementation**

---

## Appendix A: Complete API Mapping

### A.1 Controller to Service Mapping

```
ItemsController → items-api
UserController → users-api  
UserLibraryController → users-api + metadata-api
UserViewsController → collections-api
ApiKeyController → auth-api
SessionController → sessions-api
ActivityLogController → system-api
DevicesController → devices-api
LibraryController → library-api
LibraryStructureController → library-api
CollectionController → collections-api
PlaylistsController → collections-api
PlaystateController → playback-api + sessions-api
ImageController → images-api
DynamicHlsController → playback-api
HlsSegmentController → playback-api
VideosController → playback-api
SubtitlesController → subtitles-api
UniversalAudioController → playback-api
AudioController → playback-api
MediaInfoController → media-api
MediaSegmentsController → playback-api
LiveTvController → tv-api
ChannelsController → tv-api
TvShowsController → metadata-api
MoviesController → metadata-api
ArtistsController → metadata-api
GenresController → metadata-api
MusicGenresController → metadata-api
StudiosController → metadata-api
PersonsController → metadata-api
YearsController → metadata-api
AlbumArtistsController → metadata-api
SearchController → search-api
SuggestionsController → search-api
ItemLookupController → metadata-api
ItemRefreshController → media-api
FilterController → search-api
LyricsController → lyrics-api
ScheduledTasksController → system-api
PluginController → plugins-api
PackageController → plugins-api
LocalizationController → system-api
BrandingController → system-api
DashboardController → system-api
ConfigurationController → system-api
DisplayPreferencesController → preferences-api
EnvironmentController → system-api
QuickConnectController → auth-api
RemoteImageController → images-api
CollectionController → collections-api
TrickplayController → images-api
TimeSyncController → sessions-api
SyncPlayController → sessions-api + playback-api
BackupController → backup-api
ClientLogController → system-api
YearsController → metadata-api
TrailersController → metadata-api
InstantMixController → playback-api
AlbumsController → metadata-api
```

---

## Appendix B: OpenAPI Structure Example

```yaml
# apis/items/openapi.yaml
openapi: 3.0.0
info:
  title: Items API
  description: API for managing media items
  version: 1.0.0
  license:
    name: Apache 2.0

servers:
  - url: http://localhost:8080/api/v1

paths:
  /items:
    get:
      tags:
        - Items
      summary: Gets a list of items
      operationId: getItems
      parameters:
        - name: userId
          in: query
          description: Filter by user ID
          schema:
            type: string
            format: uuid
        - name: parentId
          in: query
          description: Parent folder ID
          schema:
            type: string
            format: uuid
        - name: startIndex
          in: query
          schema:
            type: integer
            minimum: 0
        - name: limit
          in: query
          schema:
            type: integer
            minimum: 1
            maximum: 1000
        - name: includeItemTypes
          in: query
          schema:
            type: array
            items:
              type: string
              enum: [Movie, Series, Episode, Album, Artist, Song, Book, AudioBook]
        - name: sortBy
          in: query
          schema:
            type: string
            enum: [SortName, DateCreated, PremiereDate, Runtime, CommunityRating, Random]
        - name: sortOrder
          in: query
          schema:
            type: string
            enum: [Ascending, Descending]
      responses:
        '200':
          description: Successful operation
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/QueryResultBaseItemDto'
        '400':
          description: Bad request
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'
        '401':
          description: Unauthorized

components:
  securitySchemes:
    BearerAuth:
      type: http
      scheme: bearer
      bearerFormat: MediaBrowser Token

  schemas:
    QueryResultBaseItemDto:
      type: object
      properties:
        Items:
          type: array
          items:
            $ref: '#/components/schemas/BaseItemDto'
        TotalRecordCount:
          type: integer
        StartIndex:
          type: integer

    BaseItemDto:
      type: object
      properties:
        Id:
          type: string
          format: uuid
        Name:
          type: string
        Type:
          type: string
        Overview:
          type: string
        ProductionYear:
          type: integer
        IndexNumber:
          type: integer
        CanDelete:
          type: boolean
        IsFolder:
          type: boolean
        ParentId:
          type: string
          format: uuid
        LocationType:
          type: string
        MediaType:
          type: string
        Path:
          type: string
        OfficialRating:
          type: string

    Error:
      type: object
      properties:
        ErrorCode:
          type: string
        Message:
          type: string
        StatusCode:
          type: integer
```

---

*Document Version: 1.0*
*Last Updated: $(date +%Y-%m-%d)*
*Author: Kabletown Migration Team*
