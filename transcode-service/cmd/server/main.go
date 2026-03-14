package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jellyfinhanced/shared/auth"
	"github.com/jellyfinhanced/shared/db"
	"github.com/jmoiron/sqlx"
)

type TranscodeJob struct {
	ID        string    `json:"id"`
	ItemId    string    `json:"itemId"`
	UserId    string    `json:"userId"`
	DeviceId  string    `json:"deviceId"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
}

type TranscodeManager struct {
	mu         sync.RWMutex
	activeJobs map[string]*TranscodeJob
	jobQueue   []string
}

func NewTranscodeManager() *TranscodeManager {
	return &TranscodeManager{
		activeJobs: make(map[string]*TranscodeJob),
		jobQueue:   make([]string, 0),
	}
}

func (tm *TranscodeManager) CreateJob(itemId, userId, deviceId string) string {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	jobID := itemId + "-" + userId + "-" + deviceId + "-" + time.Now().Format(time.RFC3339)
	job := &TranscodeJob{
		ID:        jobID,
		ItemId:    itemId,
		UserId:    userId,
		DeviceId:  deviceId,
		Status:    "pending",
		CreatedAt: time.Now(),
	}

	tm.activeJobs[jobID] = job
	tm.jobQueue = append(tm.jobQueue, jobID)

	return jobID
}

type Server struct {
	db           *sqlx.DB
	router       *chi.Mux
	transcodeMgr *TranscodeManager
}

func NewServer(database *sqlx.DB) *Server {
	svr := &Server{
		db:           database,
		router:       chi.NewRouter(),
		transcodeMgr: NewTranscodeManager(),
	}

	svr.setupRoutes()
	return svr
}

func (s *Server) setupRoutes() {
	s.router.Use(middleware.Logger)
	s.router.Use(middleware.Recoverer)
	s.router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-Emby-Authorization"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	s.router.Get("/health", s.healthHandler)
	s.router.Get("/ready", s.readyHandler)

	// Protected routes - wrap router with auth middleware
	s.router.Group(func(r chi.Router) {
		r.Use(func(h http.Handler) http.Handler {
			return auth.AuthMiddleware(s.db, h)
		})
		r.Post("/transcodes", s.createTranscodeHandler)
		r.Get("/transcodes", s.listTranscodesHandler)
		r.Delete("/transcodes/{id}", s.cancelTranscodeHandler)
		r.Post("/transcodes/{id}/lock", s.lockTranscodeHandler)
		r.Post("/transcodes/{id}/begin", s.beginRequestHandler)
		r.Post("/transcodes/{id}/end", s.endRequestHandler)
	})
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	s.transcodeMgr.mu.RLock()
	count := len(s.transcodeMgr.activeJobs)
	s.transcodeMgr.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":     "healthy",
		"activeJobs": count,
		"timestamp":  time.Now().Unix(),
	})
}

func (s *Server) readyHandler(w http.ResponseWriter, r *http.Request) {
	// Check DB connection
	if err := s.db.Ping(); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{"status": "not ready", "error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ready"})
}

func (s *Server) createTranscodeHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ItemId   string `json:"itemId"`
		UserId   string `json:"userId"`
		DeviceId string `json:"deviceId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if req.ItemId == "" && req.UserId == "" && req.DeviceId == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	jobID := s.transcodeMgr.CreateJob(req.ItemId, req.UserId, req.DeviceId)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"jobId": jobID, "status": "created"})
}

func (s *Server) listTranscodesHandler(w http.ResponseWriter, r *http.Request) {
	s.transcodeMgr.mu.RLock()
	defer s.transcodeMgr.mu.RUnlock()

	jobs := make([]*TranscodeJob, 0, len(s.transcodeMgr.activeJobs))
	for _, job := range s.transcodeMgr.activeJobs {
		jobs = append(jobs, job)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(jobs)
}

func (s *Server) cancelTranscodeHandler(w http.ResponseWriter, r *http.Request) {
	jobID := chi.URLParam(r, "id")

	s.transcodeMgr.mu.Lock()
	delete(s.transcodeMgr.activeJobs, jobID)
	s.transcodeMgr.mu.Unlock()

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) lockTranscodeHandler(w http.ResponseWriter, r *http.Request) {
	jobID := chi.URLParam(r, "id")

	s.transcodeMgr.mu.Lock()
	if job, ok := s.transcodeMgr.activeJobs[jobID]; ok {
		job.Status = "locked"
		s.transcodeMgr.mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "locked", "jobId": jobID})
		return
	}
	s.transcodeMgr.mu.Unlock()

	http.Error(w, "Job not found", http.StatusNotFound)
}

func (s *Server) beginRequestHandler(w http.ResponseWriter, r *http.Request) {
	jobID := chi.URLParam(r, "id")

	s.transcodeMgr.mu.Lock()
	if job, ok := s.transcodeMgr.activeJobs[jobID]; ok {
		job.Status = "running"
		s.transcodeMgr.activeJobs[jobID] = job
		s.transcodeMgr.mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "running", "jobId": jobID})
		return
	}
	s.transcodeMgr.mu.Unlock()

	http.Error(w, "Job not found", http.StatusNotFound)
}

func (s *Server) endRequestHandler(w http.ResponseWriter, r *http.Request) {
	jobID := chi.URLParam(r, "id")

	s.transcodeMgr.mu.Lock()
	if job, ok := s.transcodeMgr.activeJobs[jobID]; ok {
		job.Status = "completed"
		s.transcodeMgr.activeJobs[jobID] = job

		// Remove from memory after completion
		go func() {
			time.Sleep(5 * time.Minute)
			s.transcodeMgr.mu.Lock()
			delete(s.transcodeMgr.activeJobs, jobID)
			s.transcodeMgr.mu.Unlock()
		}()

		s.transcodeMgr.mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "completed", "jobId": jobID})
		return
	}
	s.transcodeMgr.mu.Unlock()

	http.Error(w, "Job not found", http.StatusNotFound)
}

func main() {
	// Build DSN from environment variables
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "3306")
	dbUser := getEnv("DB_USER", "root")
	dbPass := getEnv("DB_PASS", "")
	dbName := getEnv("DB_NAME", "kabletown")

	dsn := dbUser + ":" + dbPass + "@tcp(" + dbHost + ":" + dbPort + ")/" + dbName + "?parseTime=true"

	dbPool, err := db.NewDB(dsn)
	if err != nil {
		log.Fatalf("Failed to create database connection: %v", err)
	}
	defer dbPool.Close()

	svr := NewServer(dbPool)

	port := getEnv("PORT", "8011")
	log.Printf("Transcode service starting on port %s", port)
	log.Println("Press Ctrl+C to exit")

	if err := http.ListenAndServe(":"+port, svr.router); err != nil {
		log.Fatal(err)
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return strings.TrimSpace(val)
	}
	return defaultVal
}
