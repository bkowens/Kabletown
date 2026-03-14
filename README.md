# Kabletown - Go Microservices Media Server

## Architecture Overview

Kabletown is a Go-based replacement for Jellyfin/Emby, implemented as microservices with:
- Shared MySQL database (no inter-service RPC for authentication)
- Token-based authentication (256-bit hex tokens)
- Docker Compose orchestration
- 13 microservices covering all Jellyfin features

## Services

| Service | Port | Description |
|-----------|-----:|-----------|
| auth-service | 8001 | Token validation, user management, sessions |
| item-service | 8002 | Media items, queries (InternalItemsQuery), metadata |
| streaming-service | 8003 | HLS manifests, segment delivery, progressive streams |
| transcode-service | 8004 | FFmpeg process management, TranscodeManager |
| metadata-service | 8005 | External metadata scraping (TMDB, TVDB) |
| search-service | 8006 | Full-text search, fuzzy matching |
| user-service | 8007 | User profiles, preferences, watch history |
| collection-service | 8008 | Box sets, smart collections |
| playlist-service | 8009 | Playlists, playstate, auto-playlists |
| library-service | 8010 | Library scanning, file watching |
| notification-service | 8011 | WebSocket events, SSE streaming |
| plugin-service | 8012 | Plugin lifecycle management |
| nginx | 8080 | API Gateway (Nginx reverse proxy) |

## Quick Start

### Prerequisites

- Docker 20.10+
- Docker Compose v2
- FFmpeg (for transcode service)

### Running

1. Clone and configure
2. cd kabletown
3. Create data volume for media (optional): mkdir -p /data/media
4. Update docker-compose.yml to mount your media folder
5. Start services: docker-compose up -d
6. Check health: curl http://localhost:8080/health
7. Access API via gateway: curl http://localhost:8080/Items

### Default Credentials

- Username: admin
- Password: admin123
- Token: admin-token-initial-key-hashed (hashed SHA256)

## Development

### Build Individual Service

    cd kabletown/auth-service
    go build -o server ./cmd/server
    ./server

### Run Migrations

    docker-compose exec db mysql -u kabletown -pkabletown_password kabletown < /docker-entrypoint-initdb.d/001_auth_schema.sql

## API Documentation

See docs/api-spec.yaml for OpenAPI 3.0 specification.

## Security Notes

1. Token Management: Tokens are stored as SHA256 hashes
2. Path Traversal: All file paths validated with regex
3. CORS: Configured in each service (adjust for production)
4. Database: Use strong passwords, enable TLS connections

## Next Steps

- Implement InternalItemsQuery parser (item-service)
- Implement HLS playlist generation (streaming-service)
- Implement TranscodeManager FFmpeg lifecycle (transcode-service)
- Add user-service endpoint implementations
- Add search-service FTS integration
- Implement notification-service WebSocket hub
- Add plugin-service sandboxing
- Create migration tooling (golang-migrate)
- Add OpenAPI generation from handlers
- Implement healthcheck aggregators
