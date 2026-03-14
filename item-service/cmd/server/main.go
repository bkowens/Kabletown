package main

import (
	"context"
	"database/sql"
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
	serverPort  = os.Getenv("ITEM_SERVICE_PORT")
	serviceName = "item-service"
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

	s.router.Group(func(r chi.Router) {
		// TODO: Add Auth Middleware
		// r.Use(auth.Middleware(s.db))

		s.router.Route("/Items", func(r chi.Router) {
			r.Get("/", s.listItemsHandler)
			r.Get("/{id}", s.getItemHandler)
			r.Get("/{id}/Download", s.downloadHandler)
			s.router.Route("/{id}/Similar", func(r chi.Router) {
				r.Get("/", s.getSimilarItemsHandler)
			})
		})

		s.router.Route("/Users/{userId}/Items", func(r chi.Router) {
			r.Get("/", s.getUserItemsHandler)
			r.Get("/{id}", s.getUserItemHandler)
		})

		s.router.Get("/Users/{userId}/Latest", s.getLatestHandler)
	})
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status":"healthy","service":"%s","timestamp":"%s"}`,
		serviceName, time.Now().UTC().Format(time.RFC3339))
}

func (s *Server) readyHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	if err := s.db.PingContext(ctx); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintf(w, `{"status":"unhealthy","database":"connection failed"}`)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status":"ready","database":"connected"}`)
}

func (s *Server) listItemsHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement InternalItemsQuery (80+ params)
	// See: docs/internal-items-query.md
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

func (s *Server) getItemHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

func (s *Server) downloadHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

func (s *Server) getSimilarItemsHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

func (s *Server) getUserItemsHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

func (s *Server) getUserItemHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

func (s *Server) getLatestHandler(w http.ResponseWriter, r *http.Request) {
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

	port := serverPort
	if port == "" {
		port = "8002"
	}

	server := NewServer(db, port)
	addr := fmt.Sprintf(":%s", port)

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
