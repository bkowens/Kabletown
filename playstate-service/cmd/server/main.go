package main

import (
    "context"
    "database/sql"
    "encoding/json"
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
    serverPort  = os.Getenv("PLAYSTATE_SERVICE_PORT")
    serviceName = "playstate-service"
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
        AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Emby-Authorization"},
        ExposedHeaders:   []string{"Link"},
        AllowCredentials: true,
        MaxAge:           300,
    }))
}

func (s *Server) setupRoutes() {
    s.router.Get("/health", s.healthHandler)
    s.router.Get("/ready", s.readyHandler)

    api := s.router.Group(func(r chi.Router) {
        r.Use(auth.DBTokenValidator(s.db))

        s.router.Route("/Items/{id}", func(r chi.Router) {
            r.Post("/PlaybackStart", s.playbackStartHandler)
            r.Post("/Playing", s.playingHandler)
            r.Post("/PlaybackStopped", s.playbackStoppedHandler)
            r.Post("/Play", s.playHandler)
        })

        s.router.Route("/Sessions/{id}", func(r chi.Router) {
            r.Post("/Playing", s.sessionPlayingHandler)
            r.Post("/PlaybackStopped", s.sessionPlaybackStoppedHandler)
        })

        s.router.Route("/Users/{userId}/Items/{itemId}", func(r chi.Router) {
            r.Post("/PlaybackStart", s.userPlaybackStartHandler)
            r.Post("/Playing", s.userPlayingHandler)
            r.Post("/PlaybackStopped", s.userPlaybackStoppedHandler)
        })
    })
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{
        "status": "healthy",
        "service": serviceName,
        "timestamp": time.Now().UTC().Format(time.RFC3339),
    })
}

func (s *Server) readyHandler(w http.ResponseWriter, r *http.Request) {
    ctx, cancel := context.WithTimeout(r.Context(), 5 * time.Second)
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

func (s *Server) playbackStartHandler(w http.ResponseWriter, r *http.Request) {
    itemId := chi.URLParam(r, "id")
    user := auth.TokenFromContext(r)
    
    var req struct {
        PositionTicks   int64  `json:"PositionTicks"`
        MediaSourceId    string `json:"MediaSourceId"`
        AudioStreamIndex int    `json:"AudioStreamIndex"`
        SubtitleStreamIndex int `json:"SubtitleStreamIndex"`
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }

    err := s.recordPlaybackStart(user.UserID, itemId, req.PositionTicks, req.MediaSourceId)
    if err != nil {
        log.Printf("Error recording playback start: %v", err)
        http.Error(w, "Failed to record playback start", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
}

func (s *Server) playingHandler(w http.ResponseWriter, r *http.Request) {
    itemId := chi.URLParam(r, "id")
    user := auth.TokenFromContext(r)

    var req struct {
        PositionTicks int64 `json:"PositionTicks"`
        IsPaused      bool  `json:"IsPaused"`
        VolumeLevel   int   `json:"VolumeLevel"`
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }

    err := s.recordPlaying(user.UserID, itemId, req.PositionTicks, req.IsPaused)
    if err != nil {
        log.Printf("Error recording playing: %v", err)
        http.Error(w, "Failed to record playing state", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
}

func (s *Server) playbackStoppedHandler(w http.ResponseWriter, r *http.Request) {
    itemId := chi.URLParam(r, "id")
    user := auth.TokenFromContext(r)

    var req struct {
        PositionTicks int64 `json:"PositionTicks"`
        PlayMethod    string `json:"PlayMethod"`
        PlaySessionId string `json:"PlaySessionId"`
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }

    err := s.recordPlaybackStopped(user.UserID, itemId, req.PositionTicks)
    if err != nil {
        log.Printf("Error recording playback stopped: %v", err)
        http.Error(w, "Failed to record playback stopped", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
}

func (s *Server) playHandler(w http.ResponseWriter, r *http.Request) {
    s.playbackStartHandler(w, r)
}

func (s *Server) sessionPlayingHandler(w http.ResponseWriter, r *http.Request) {
    s.playingHandler(w, r)
}

func (s *Server) sessionPlaybackStoppedHandler(w http.ResponseWriter, r *http.Request) {
    s.playbackStoppedHandler(w, r)
}

func (s *Server) userPlaybackStartHandler(w http.ResponseWriter, r *http.Request) {
    s.playbackStartHandler(w, r)
}

func (s *Server) userPlayingHandler(w http.ResponseWriter, r *http.Request) {
    s.playingHandler(w, r)
}

func (s *Server) userPlaybackStoppedHandler(w http.ResponseWriter, r *http.Request) {
    s.playbackStoppedHandler(w, r)
}

func (s *Server) recordPlaybackStart(userId, itemId string, positionTicks int64, mediaSourceId string) error {
    tx, err := s.db.Begin()
    if err != nil {
        return err
    }
    defer tx.Rollback()

    now := time.Now().UTC()
    
    tx.Exec("DELETE FROM playback_state WHERE user_id = ? AND item_id = ? AND is_active = 1", userId, itemId)

    _, err = tx.Exec(
        "INSERT INTO playback_state (id, user_id, item_id, position_ticks, media_source_id, is_active, started_at, updated_at) VALUES (?, ?, ?, ?, ?, 1, ?, ?)",
        "00000000-0000-0000-0000-000000000000", userId, itemId, positionTicks, mediaSourceId, now, now,
    )
    if err != nil {
        return err
    }

    return tx.Commit()
}

func (s *Server) recordPlaying(userId, itemId string, positionTicks int64, isPaused bool) error {
    now := time.Now().UTC()

    _, err := s.db.Exec(
        "UPDATE playback_state SET position_ticks = ?, is_paused = ?, updated_at = ? WHERE user_id = ? AND item_id = ? AND is_active = 1",
        positionTicks, isPaused, now, userId, itemId,
    )
    if err != nil {
        return err
    }

    _, err = s.db.Exec(
        "UPDATE user_data SET playback_position_ticks = ? WHERE user_id = ? AND item_id = ?",
        positionTicks, userId, itemId,
    )
    return err
}

func (s *Server) recordPlaybackStopped(userId, itemId string, positionTicks int64) error {
    tx, err := s.db.Begin()
    if err != nil {
        return err
    }
    defer tx.Rollback()

    now := time.Now().UTC()

    tx.Exec("UPDATE playback_state SET is_active = 0 WHERE user_id = ? AND item_id = ? AND is_active = 1", userId, itemId)

    _, err = tx.Exec(
        "UPDATE user_data SET play_count = play_count + 1, last_play_date = ?, is_played = ?, playback_position_ticks = 0 WHERE user_id = ? AND item_id = ?",
        now, positionTicks > 0, userId, itemId,
    )
    if err != nil {
        return err
    }

    return tx.Commit()
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
        port = "8004"
    }

    server := NewServer(db, port)
    addr := ":" + port

    log.Printf("Starting %s on %s", serviceName, addr)

    httpServer := &http.Server{
        Addr:         addr,
        Handler:      server.router,
        ReadTimeout:  15 * time.Second,
        WriteTimeout: 60 * time.Second,
        IdleTimeout:  120 * time.Second,
    }

    go func() {
        quit := make(chan os.Signal, 1)
        signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
        <-quit

        log.Println("Shutting down server...")
        ctx, cancel := context.WithTimeout(context.Background(), 30 * time.Second)
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
