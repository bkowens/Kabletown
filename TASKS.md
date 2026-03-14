# Task List: Kabletown Go Microservices

## 🎯 Phase 1: Core Infrastructure (Complete)
- [x] Database setup (MySQL 8.0, auth schema)
- [x] nginx gateway (8080, routes all services)
- [x] auth-service (8001, sessions, users, devices, API keys)
- [x] docker-compose.yml (all 13 services)
- [x] Shared auth middleware
- [x] Documentation (8,000+ lines)

## 🚧 Phase 2: MVP Implementation (In Progress)

### Priority 1: Library Service (Port 8003) - 80+ Controllers
This is the MAIN browsing/query interface for all content

**Implement handlers for:**
- [ ] `ItemsController` - InternalItemsQuery (80+ params)
  - TopParentId, Type, recursive CTE hierarchy
  - P6/P7 index filters
  - DateRange, Year, Runtime, Genre filters
  - Limit, Offset, SortBy, SortOrder
  - ParentId, AncestorIds, IncludeItemTypes
- [ ] `UserLibraryController` - User-specific queries
- [ ] `UserViewsController` - Home screen, quick views
  - Fetch user views based on library type
  - Group by genres, recent, etc.
- [ ] `LibraryController` - Library metadata
  - GetGenres, GetStudios, GetPersons, GetYears
  - GetArtists, GetAlbums
- [ ] `FilterController` - Filter suggestions
- [ ] `GenresController` - Genre browsing
- [ ] `MusicGenresController` - Music genre browsing
- [ ] `StudiosController` - Studio browsing
- [ ] `PersonsController` - Actor/director browsing
- [ ] `YearsController` - Year browsing
- [ ] `TrailersController` - Trailer queries
- [ ] `SuggestionController` - Browse suggestions

**Database schema needed:**
- [ ] items table (BaseItemKind, Path, PathHash, Name, OriginalTitle)
- [ ] item_values table (P6 index: Genres, Studios, People)
- [ ] library_folders table (Name, Path, LibraryType)
- [ ] media_paths table (FolderPath, ItemType)
- [ ] views (UserViews, CollectionType)

### Priority 2: Streaming Service (Port 8004) - HLS 
- [ ] `MasterPlaylistGenerator` - Master HLS manifest
  - Available bitrates (1080p, 720p, 480p)
  - Codecs, resolution, bandwidth
- [ ] `VariantPlaylistGenerator` - Variant HLS manifest
  - Segment list, durations
  - Segment URLs
- [ ] `SegmentDelivery` - .ts segment file serving
  - Cache segments in /tmp/transcoding
  - Segment cleanup logic
- [ ] `ProgressiveStreamHandler` - Direct play
  - Range header support (byte serving)
- [ ] `SubtitleHandler` - Subtitle delivery
  - Embedded, external, burned-in

**Integration:**
- [ ] Call transcode-service for transcoded segments
- [ ] Cache manifest files
- [ ] Handle segment deletion from cache

### Priority 3: Item Service (Port 8008) - Storage Layer
Internal repository for item storage (called by library-service)

- [ ] `BaseItemRepository` - Items table CRUD
  - GetItemsInternal() - 80+ params query
  - GetItemById(), GetItemByPath()
  - CreateItem(), UpdateItem(), DeleteItem()
- [ ] `P6/P7 Index Management`
  - P7: TopParentId + Type composite index
  - P6: ItemValues (normalization)
- [ ] `ItemValueProcessor` - Handle multi-value fields
- [ ] `RecursiveCTE` - Hierarchy traversal

**Database schema:**
- [ ] items table (full schema from docs)
- [ ] item_values normalization
- [ ] AncestorIds, TopParentId pre-computation

### Priority 4: Update nginx Routing
- [ ] Fix service port assignments:
  - library-service: 8003 (not 8010)
  - streaming-service: 8004 (not 8003)
  - transcode-service: 8005 (not 8004)
  - item-service: 8008 (not 8007)
  - metadata-service: 8006
  - search-service: 8007

## 📋 Phase 3: Supporting Services

### Metadata Service (Port 8006)
- [ ] TMDB integration (movies, TV)
- [ ] TVDB integration (TV)
- [ ] OMDb integration (fallback)
- [ ] Image fetching/downscaling
- [ ] NFO parsing
- [ ] Metadata caching
- [ ] Provider fallback chains

### Search Service (Port 8007)
- [ ] MySQL FTS setup (FULLTEXT indexes)
- [ ] Fuzzy matching (Levenshtein, trigram)
- [ ] Autocomplete endpoint
- [ ] Filter suggestions (genre, year, runtime)
- [ ] Synonyms database
- [ ] Search index sync with item-service

### Collection Service (Port 8009)
- [ ] Manual collections (user-created)
- [ ] Box sets (studio, franchise)
- [ ] Smart collections (rule-based)
- [ ] Collection hierarchy
- [ ] Collection cover art generation
- [ ] Collection membership API

### Playlist Service (Port 8010)
- [ ] User playlists (create, update, delete)
- [ ] Auto-playlists (rule-based)
- [ ] Play queue management
- [ ] Shuffle algorithms
- [ ] Playlist sharing
- [ ] Gapless playback support

### Notification Service (Port 8011)
- [ ] WebSocket hub (SocketHub equivalent)
- [ ] SSE streaming
- [ ] Session broadcasts
- [ ] Transcoder status updates
- [ ] Library scan events

### Plugin Service (Port 8012)
- [ ] Plugin loader (directory scanning)
- [ ] Plugin lifecycle (start/stop)
- [ ] Sandbox isolation
- [ ] API hooks registration
- [ ] Plugin manifest validation
- [ ] Marketplace integration

## 🔧 Phase 4: Infrastructure & Polish

### Database Migrations
- [ ] golang-migrate integration
- [ ] Complete items table schema
- [ ] Library/schema migrations
- [ ] Index optimization (P6, P7)
- [ ] Full-text index setup

### Transcode Service Enhancements
- [ ] Full FFmpeg wrapper (transcode-service internal/ffmpeg/)
- [ ] Process management
- [ ] ActiveRequestCount tracking
- [ ] Kill timer implementation
- [ ] Segment lifecycle
- [ ] Error handling, retry logic

### Testing
- [ ] E2E auth flow test
- [ ] Item query test suite
- [ ] HLS playback test
- [ ] Load testing
- [ ] Integration tests (all services)

### Documentation
- [ ] Swagger/OpenAPI generation
- [ ] Developer setup guide
- [ ] API client SDK generation
- [ ] Troubleshooting guide

---

## 🎯 Immediate Next Steps (This Week)
1. Update nginx.conf with correct port mappings
2. Implement library-service ItemsController (InternalItemsQuery)
3. Create items table migration
4. Test end-to-end: auth → browse items → stream

**Estimated effort:**
- Library service: 4-6 hours (80+ params, complex queries)
- Streaming implementation: 3-4 hours (HLS generation)
- Item service: 2-3 hours (repository layer)
- nginx updates: 30 min
- Testing: 2-3 hours

**Total MVP implementation: 2-3 days**
