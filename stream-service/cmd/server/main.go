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
    "github.com/jellyfinhanced/shared/auth"
)

var (
    dbDSN       = os.Getenv("DB_DSN")
    serverPort  = os.Getenv("STREAM_SERVICE_PORT")
    serviceName = "stream-service"
)

type Server struct {
    router *chi.Mux
    db     *sql.DB
    port   string
}

func NewServer(db *sql.DB, port string) *Server {
    s := &Server{router: chi.NewRouter(), db: db, port: port}
    s.setupMiddleware()
    s.setupRoutes()
    return s
}

func (s *Server) setupMiddleware() {
    s.router.Use(middleware.Logger)
    s.router.Use(middleware.Recoverer)
    s.router.Use(middleware.RequestID)
    s.router.Use(cors.Handler(cors.Options{
        AllowedOrigins:   []string{"*"},
        AllowedMethods:   []string{"GET", "HEAD", "OPTIONS"},
        AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Emby-Authorization"},
        ExposedHeaders:   []string{"Link", "Content-Length", "Content-Range"},
        AllowCredentials: true,
        MaxAge:           300,
    }))
}

func (s *Server) setupRoutes() {
    s.router.Get("/health", s.healthHandler)
    s.router.Get("/ready", s.readyHandler)

    api := s.router.Group(func(r chi.Router) {
        r.Use(auth.DBTokenValidator(s.db))

        // Video streaming
        s.router.Route("/Videos", func(r chi.Router) {
            r.Get("/{itemId}/master.m3u8", s.masterPlaylistHandler)
            r.Get("/{itemId}/stream.m3u8", s.variantPlaylistHandler)
            r.Get("/{itemId}/segment/{segmentId}.ts", s.hlsSegmentHandler)
            r.Get("/{itemId}/stream", s.progressiveStreamHandler)
            r.Get("/{itemId}/stream.m4s", s.dashSegmentHandler)
        })

        // Audio streaming
        s.router.Route("/Audio", func(r chi.Router) {
            r.Get("/{itemId}/stream", s.audioStreamHandler)
            r.Get("/{itemId}/stream.mp3", s.audioMp3Handler)
            r.Get("/{itemId}/transcode", s.audioTranscodeHandler)
        })

        // Universal audio
        s.router.Route("/UniversalAudio", func(r chi.Router) {
            r.Get("/{itemId}", s.universalAudioHandler)
        })

        // HLS segments from transcode service
        s.router.Route("/Hls", func(r chi.Router) {
            r.Get("/{id}/{container}/{segmentId}", s.hlsFromTranscodeHandler)
        })

        // Trickplay (thumbnail previews)
        s.router.Route("/Trickplay", func(r chi.Router) {
            r.Get("/{itemId}/manifest.m3u8", s.trickplayManifestHandler)
            r.Get("/{itemId}/{resolution}/{time}.jpg", s.trickplayHandler)
        })
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
        json.NewEncoder(w).Encode(map[string]string{"status": "unhealthy", "database": "connection failed"})
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{"status": "ready", "database": "connected"})
}

func (s *Server) masterPlaylistHandler(w http.ResponseWriter, r *http.Request) {
    itemId := chi.URLParam(r, "itemId")
    user := auth.TokenFromContext(r)

    log.Printf("HLS Master Playlist Request: itemId=%s, userId=%s", itemId, user.UserID)

    playlist := `#EXTM3U
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
    w.Write([]byte(playlist))
}

func (s *Server) variantPlaylistHandler(w http.ResponseWriter, r *http.Request) {
    itemId := chi.URLParam(r, "itemId")

    log.Printf("HLS Variant Playlist Request: itemId=%s", itemId)

    playlist := `#EXTM3U
#EXT-X-VERSION:3
#EXT-X-TARGETDURATION:6
#EXT-X-MEDIA-SEQUENCE:0
#EXTINF:6.000000,
segment_0.ts
#EXTINF:6.000000,
segment_1.ts
#EXTINF:6.000000,
segment_2.ts
#EXTINF:6.000000,
segment_3.ts
#EXT-X-ENDLIST
`

    w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
    w.WriteHeader(http.StatusOK)
    w.Write([]byte(playlist))
}

func (s *Server) hlsSegmentHandler(w http.ResponseWriter, r *http.Request) {
    itemId := chi.URLParam(r, "itemId")
    segmentId := chi.URLParam(r, "segmentId")

    log.Printf("HLS Segment Request: itemId=%s, segmentId=%s", itemId, segmentId)

    w.Header().Set("Content-Type", "video/MP2T")
    w.Header().Set("Content-Length", "1024")
    w.WriteHeader(http.StatusOK)

    w.Write(make([]byte, 1024))
}

func (s *Server) progressiveStreamHandler(w http.ResponseWriter, r *http.Request) {
    itemId := chi.URLParam(r, "itemId")

    log.Printf("Progressive Stream Request: itemId=%s", itemId)

    w.Header().Set("Content-Type", "video/mp4")
    w.Header().Set("Accept-Ranges", "bytes")
    w.Header().Set("Content-Length", "1048576")
    w.WriteHeader(http.StatusOK)

    log.Println("Video streaming placeholder response")
}

func (s *Server) dashSegmentHandler(w http.ResponseWriter, r *http.Request) {
    itemId := chi.URLParam(r, "itemId")
    log.Printf("DASH Segment Request: itemId=%s", itemId)
    http.Error(w, "Not implemented", http.StatusNotImplemented)
}

func (s *Server) audioStreamHandler(w http.ResponseWriter, r *http.Request) {
    itemId := chi.URLParam(r, "itemId")
    log.Printf("Audio Stream Request: itemId=%s", itemId)

    w.Header().Set("Content-Type", "audio/mpeg")
    w.Header().Set("Content-Length", "1048576")
    w.WriteHeader(http.StatusOK)
}

func (s *Server) audioMp3Handler(w http.ResponseWriter, r *http.Request) {
    s.audioStreamHandler(w, r)
}

func (s *Server) audioTranscodeHandler(w http.ResponseWriter, r *http.Request) {
    itemId := chi.URLParam(r, "itemId")
    log.Printf("Audio Transcode Request: itemId=%s", itemId)
    http.Error(w, "Not implemented", http.StatusNotImplemented)
}

func (s *Server) universalAudioHandler(w http.ResponseWriter, r *http.Request) {
    itemId := chi.URLParam(r, "itemId")
    log.Printf("Universal Audio Request: itemId=%s", itemId)
    http.Error(w, "Not implemented", http.StatusNotImplemented)
}

func (s *Server) hlsFromTranscodeHandler(w http.ResponseWriter, r *http.Request) {
    itemId := chi.URLParam(r, "id")
    container := chi.URLParam(r, "container")
    segmentId := chi.URLParam(r, "segmentId")

    log.Printf("HLS from Transcode: itemId=%s, container=%s, segmentId=%s", itemId, container, segmentId)

    w.Header().Set("Content-Type", "video/MP2T")
    w.Header().Set("Content-Length", "1024")
    w.WriteHeader(http.StatusOK)
    w.Write(make([]byte, 1024))
}

func (s *Server) trickplayManifestHandler(w http.ResponseWriter, r *http.Request) {
    itemId := chi.URLParam(r, "itemId")
    log.Printf("Trickplay Manifest Request: itemId=%s", itemId)

    manifest := `#EXTM3U
#EXT-X-VERSION:3
#EXT-X-TARGETDURATION:10
#EXT-X-PLAYLIST-TYPE:VOD
#EXTINF:5.0,
00000.jpg
#EXTINF:5.0,
00001.jpg
#EXT-X-ENDLIST
`

    w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
    w.WriteHeader(http.StatusOK)
    w.Write([]byte(manifest))
}

func (s *Server) trickplayHandler(w http.ResponseWriter, r *http.Request) {
    itemId := chi.URLParam(r, "itemId")
    resolution := chi.URLParam(r, "resolution")
    time := chi.URLParam(r, "time")

    log.Printf("Trickplay Thumbnail: itemId=%s, resolution=%s, time=%s", itemId, resolution, time)

    w.Header().Set("Content-Type", "image/jpeg")
    w.WriteHeader(http.StatusOK)
    w.Write(make([]byte, 2048))
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

    port := serverPort
    if port == "" {
        port = "8006"
    }

    server := NewServer(db, port)
    addr := ":" + port

    log.Printf("Starting %s on %s", serviceName, addr)

    httpServer := &http.Server{
        Addr:         addr,
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
