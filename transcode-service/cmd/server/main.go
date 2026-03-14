package main

import (
    "context"
    "database/sql"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "os"
    "os/exec"
    "os/signal"
    "sync"
    "syscall"
    "time"

    "github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"
    "github.com/go-chi/cors"
    _ "github.com/go-sql-driver/mysql"
    "kabletown/shared/auth"
)

var (
    dbDSN           = os.Getenv("DB_DSN")
    serverPort      = os.Getenv("TRANCODE_SERVICE_PORT")
    serviceName     = "transcode-service"
    ffmpegPath      = os.Getenv("FFMPEG_PATH")
    transcodingPath = os.Getenv("TRANSCODING_PATH")
)

type TranscodeManager struct {
    mu         sync.RWMutex
    activeJobs map[string]*TranscodeJob
    killTimer  map[string]int64
}

type TranscodeJob struct {
    ID                 string
    ItemId             string
    UserId             string
    DeviceId           string
    StartTime          time.Time
    LastAccessTime     time.Time
    SegmentDuration    int
    IsLiveOutput       bool
    FFmpegProcess      *exec.Cmd
    ActiveRequestCount int
}

func NewTranscodeManager() *TranscodeManager {
    return &TranscodeManager{
        activeJobs: make(map[string]*TranscodeJob),
        killTimer:  make(map[string]int64),
    }
}

func (tm *TranscodeManager) CreateJob(itemId, userId, deviceId string) string {
    jobId := fmt.Sprintf("job_%d", time.Now().UnixNano())
    tm.mu.Lock()
    tm.activeJobs[jobId] = &TranscodeJob{
        ID: jobId, ItemId: itemId, UserId: userId, DeviceId: deviceId,
        StartTime:      time.Now(),
        LastAccessTime: time.Now(),
        SegmentDuration: 6,
        ActiveRequestCount: 0,
    }
    tm.killTimer[jobId] = time.Now().Add(60 * time.Second).Unix()
    tm.mu.Unlock()
    return jobId
}

func (tm *TranscodeManager) BeginRequest(jobId string) {
    tm.mu.Lock()
    defer tm.mu.Unlock()
    if job, ok := tm.activeJobs[jobId]; ok {
        job.ActiveRequestCount++
        job.LastAccessTime = time.Now()
        tm.killTimer[jobId] = time.Now().Add(60 * time.Second).Unix()
    }
}

func (tm *TranscodeManager) EndRequest(jobId string) {
    tm.mu.Lock()
    defer tm.mu.Unlock()
    if job, ok := tm.activeJobs[jobId]; ok {
        job.ActiveRequestCount--
    }
}

type Server struct {
    router       *chi.Mux
    db           *sql.DB
    transcodeMgr *TranscodeManager
}

func NewServer(db *sql.DB, port string) *Server {
    s := &Server{router: chi.NewRouter(), db: db, transcodeMgr: NewTranscodeManager()}
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
        r.Post("/transcodes", s.createTranscodeHandler)
        r.Get("/transcodes", s.listTranscodesHandler)
        s.router.Delete("/transcodes/{id}", s.cancelTranscodeHandler)
        s.router.Post("/transcodes/{id}/lock", s.lockTranscodeHandler)
        s.router.Post("/transcodes/{id}/begin", s.beginRequestHandler)
        s.router.Post("/transcodes/{id}/end", s.endRequestHandler)
    })
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
    s.transcodeMgr.mu.RLock()
    count := len(s.transcodeMgr.activeJobs)
    s.transcodeMgr.mu.RUnlock()
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "service": serviceName, "status": "healthy", "activeJobs": count,
    })
}

func (s *Server) readyHandler(w http.ResponseWriter, r *http.Request) {
    ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
    defer cancel()
    if err := s.db.PingContext(ctx); err != nil {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusServiceUnavailable)
        json.NewEncoder(w).Encode(map[string]string{"status": "unhealthy"})
        return
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{"status": "ready"})
}

func (s *Server) createTranscodeHandler(w http.ResponseWriter, r *http.Request) {
    var req struct { ItemId string <json:"itemId"> UserId string <json:"userId"> DeviceId string <json:"deviceId"> }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }
    jobId := s.transcodeMgr.CreateJob(req.ItemId, req.UserId, req.DeviceId)
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{"jobId": jobId, "status": "created"})
}

func (s *Server) listTranscodesHandler(w http.ResponseWriter, r *http.Request) {
    s.transcodeMgr.mu.RLock()
    jobs := []map[string]interface{}{}
    for _, job := range s.transcodeMgr.activeJobs {
        jobs = append(jobs, map[string]interface{}{"jobId": job.ID, "activeRequests": job.ActiveRequestCount})
    }
    s.transcodeMgr.mu.RUnlock()
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(jobs)
}

func (s *Server) cancelTranscodeHandler(w http.ResponseWriter, r *http.Request) {
    jobId := chi.URLParam(r, "id")
    s.transcodeMgr.mu.Lock()
    defer s.transcodeMgr.mu.Unlock()
    if job, ok := s.transcodeMgr.activeJobs[jobId]; ok {
        if job.FFmpegProcess != nil && job.FFmpegProcess.Process != nil {
            job.FFmpegProcess.Process.Kill()
        }
        delete(s.transcodeMgr.activeJobs, jobId)
        delete(s.transcodeMgr.killTimer, jobId)
    }
    w.WriteHeader(http.StatusOK)
}

func (s *Server) lockTranscodeHandler(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
}

func (s *Server) beginRequestHandler(w http.ResponseWriter, r *http.Request) {
    s.transcodeMgr.BeginRequest(chi.URLParam(r, "id"))
    w.WriteHeader(http.StatusOK)
}

func (s *Server) endRequestHandler(w http.ResponseWriter, r *http.Request) {
    s.transcodeMgr.EndRequest(chi.URLParam(r, "id"))
    w.WriteHeader(http.StatusOK)
}

func main() {
    if dbDSN == "" { log.Fatal("DB_DSN required") }
    if ffmpegPath == "" { ffmpegPath = "ffmpeg" }
    if transcodingPath == "" { transcodingPath = "/tmp/transcoding" }
    os.MkdirAll(transcodingPath, 0755)

    db, err := sql.Open("mysql", dbDSN)
    if err != nil { log.Fatalf("DB error: %v", err) }
    defer db.Close()
    if err := db.Ping(); err != nil { log.Fatalf("DB ping: %v", err) }

    port := fmt.Sprintf(":%s", serverPort)
    server := NewServer(db, serverPort)
    log.Printf("Starting %s on %s", serviceName, port)

    httpServer := &http.Server{Addr: port, Handler: server.router}
    go func() {
        quit := make(chan os.Signal, 1)
        signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
        <-quit
        s := httpServer.Shutdown
        ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
        defer cancel()
        s(ctx)
    }()
    httpServer.ListenAndServe()
}