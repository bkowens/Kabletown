package main

import (
    "context"
    "database/sql"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"
    "github.com/go-chi/cors"
    _ "github.com/go-sql-driver/mysql"
)

var (
    dbDSN       = os.Getenv("DB_DSN")
    serverPort  = os.Getenv("STREAMING_SERVICE_PORT")
    serviceName = "streaming-service"
)

type Server struct {
    router       *chi.Mux
    db           *sql.DB
    port         string
}

func NewServer(db *sql.DB, port string) *Server {
    s := &Server{
        router: chi.NewRouter(),
        db:     db,
        port:   port,
    }
    s.setupMiddleware()
    s.setupRoutes()
    return s
}

func (s *Server) setupMiddleware() {
    s.router.Use(middleware.Logger)
    s.router.Use(middleware.Recoverer)
    s.router.Use(middleware.RequestID)
    s.router.Use(cors.Handler(cors.Options{
        AllowedOrigins:  []string{"*"},
        AllowedMethods:  []string{"GET", "HEAD", "OPTIONS"},
        AllowedHeaders:  []string{"Accept", "Authorization", "Content-Type", "X-Emby-Authorization"},
        ExposedHeaders:  []string{"Link", "Content-Length", "Content-Range"},
        AllowCredentials: true,
        MaxAge:           300,
    }))
}

func (s *Server) setupRoutes() {
    s.router.Get("/health", s.healthHandler)
    s.router.Get("/ready", s.readyHandler)

    // Streaming routes
    s.router.Group(func(r chi.Router) {
        // HLS Master Playlist
        r.Get("/Videos/{itemId}/master.m3u8", s.masterPlaylistHandler)
        
        // HLS Variant Playlist
        r.Get("/Videos/{itemId}/stream.m3u8", s.variantPlaylistHandler)
        
        // HLS Segment
        r.Get("/Videos/{itemId}/segment/{segmentId}.ts", s.hlsSegmentHandler)
        
        // Progressive Stream (direct play)
        r.Get("/Videos/{itemId}/stream", s.progressiveStreamHandler)
        
        // HLS from transcode-service
        r.Get("/Items/{id}/Hls/{container}/{segmentId}", s.hlsFromTranscodeHandler)
    })
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{
        "status":   "healthy",
        "service":  serviceName,
        "timestamp": time.Now().UTC().Format(time.RFC3339),
    })
}

func (s *Server) readyHandler(w http.ResponseWriter, r *http.Request) {
    ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
    defer cancel()

    if err := s.db.PingContext(ctx); err != nil {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusServiceUnavailable)
        json.NewEncoder(w).Encode(map[string]string{
            "status":  "unhealthy",
            "database": "connection failed",
        })
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{
        "status":   "ready",
        "database":  "connected",
    })
}

func (s *Server) masterPlaylistHandler(w http.ResponseWriter, r *http.Request) {
    _ = chi.URLParam(r, "itemId")
    
    // Build Master Playlist (see internal/hls/master.go)
    // Ref: docs/hls-streaming-routes.md
    
    // Placeholder: Return basic master playlist
    playlistContent := `#EXTM3U
#EXT-X-VERSION:3
#EXT-X-STREAM-INF:BANDWIDTH=8000000,RESOLUTION=1920x1080,CODECS="avc1.640028,mp4a.40.2"
stream_1080p.m3u8
#EXT-X-STREAM-INF:BANDWIDTH=4000000,RESOLUTION=1280x720,CODECS="avc1.64001f,mp4a.40.2"
stream_720p.m3u8
#EXT-X-STREAM-INF:BANDWIDTH=1500000,RESOLUTION=854x480,CODECS="avc1.64001f,mp4a.40.2"
stream_480p.m3u8
`
    
    w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
    w.WriteHeader(http.StatusOK)
    w.Write([]byte(playlistContent))
}

func (s *Server) variantPlaylistHandler(w http.ResponseWriter, r *http.Request) {
    _ = chi.URLParam(r, "itemId")
    
    // Build Variant Playlist (see internal/hls/variant.go)
    
    playlistContent := `#EXTM3U
#EXT-X-VERSION:3
#EXT-X-TARGETDURATION:6
#EXT-X-MEDIA-SEQUENCE:0
#EXTINF:6.000000,
segment_0.ts
#EXTINF:6.000000,
segment_1.ts
#EXTINF:6.000000,
segment_2.ts
#EXT-X-ENDLIST
`
    
    w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
    w.WriteHeader(http.StatusOK)
    w.Write([]byte(playlistContent))
}

func (s *Server) hlsSegmentHandler(w http.ResponseWriter, r *http.Request) {
    _ = chi.URLParam(r, "itemId")
    _ = chi.URLParam(r, "segmentId")
    
    // Fetch segment from transcode-service or cache (see internal/hls/segment.go)
    // Ref: docs/hls-streaming-routes.md
    
    // Placeholder: Return dummy TS data
    w.Header().Set("Content-Type", "video/MP2T")
    w.Header().Set("Content-Length", "1024")
    w.WriteHeader(http.StatusOK)
    
    // TODO: Actually read from transcoded files or cache
}

func (s *Server) progressiveStreamHandler(w http.ResponseWriter, r *http.Request) {
    _ = chi.URLParam(r, "itemId")
    
    // Progressive download/stream (see internal/handlers/progressive.go)
    // Check if direct play or needs transcoding
    
    http.Error(w, "Not implemented", http.StatusNotImplemented)
}

func (s *Server) hlsFromTranscodeHandler(w http.ResponseWriter, r *http.Request) {
    _ = chi.URLParam(r, "id")
    _ = chi.URLParam(r, "container")
    _ = chi.URLParam(r, "segmentId")
    
    // Proxy to transcode-service for on-demand segments
    // This calls internal transcode-service endpoint
    
    http.Error(w, "Not implemented", http.StatusNotImplemented)
}

func main() {
    if dbDSN == "" {
        log.Fatal("DB_DSN environment variable required")
    }

    db, err := sql.Open("mysql", dbDSN)
    if err != nil {
        log.Fatalf("Failed to open database: %v", err)
    }
    defer db.Close()

    if err := db.Ping(); err != nil {
        log.Fatalf("Failed to connect to database: %v", err)
    }
    log.Println("Database connection established")

    port := fmt.Sprintf(":%s", serverPort)
    server := NewServer(db, serverPort)

    log.Printf("Starting %s on %s", serviceName, port)

    httpServer := &http.Server{
        Addr:         port,
        Handler:      server.router,
        ReadTimeout:  15 * time.Second,
        WriteTimeout: 0, // No timeout for streaming
        IdleTimeout:  120 * time.Second,
    }

    go func() {
        quit := make(chan os.Signal, 1)
        signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
        <-quit
        log.Println("Shutting down server...")

        ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
        defer cancel()

        if err := httpServer.Shutdown(ctx); err != nil {
            log.Printf("Server forced to shutdown: %v", err)
        }
        log.Println("Server exited")
    }()

    if err := httpServer.ListenAndServe(); err != http.ErrServerClosed {
        log.Fatalf("Server error: %v", err)
    }
}
