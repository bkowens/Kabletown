package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bowens/kabletown/playlist-service/internal/db"
	"github.com/bowens/kabletown/playlist-service/internal/handlers"
	"github.com/bowens/kabletown/playlist-service/internal/middleware"
	shared_db "github.com/bowens/kabletown/shared/db"
	"github.com/go-chi/chi/v5"
	chi_middleware "github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
)

func main() {
	// Server configuration
	serverPort := getEnv("SERVER_PORT", "8003")
	dataDir := getEnv("DATA_DIR", "/var/lib/jellyfin")

	// Initialize server ID
	log.Printf("Server ID: %s", middleware.InitializeServerID(dataDir))

	// Database configuration
	cfg := &shared_db.Config{
		Host:         getEnv("DB_HOST", "localhost"),
		Port:         3306,
		User:         getEnv("DB_USER", "jellyfin"),
		Password:     getEnv("DB_PASSWORD", ""),
		DBName:       getEnv("DB_NAME", "jellyfin"),
		MaxOpenConns: 20,
		MaxIdleConns: 5,
	}

	// Create database connection
	dbConn, err := shared_db.NewPool(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer dbConn.Close()

	// Ensure tables exist
	if err := db.EnsurePlaylistsTables(dbConn); err != nil {
		log.Fatalf("Failed to create tables: %v", err)
	}

	// Create ItemValues tables for P6 filtering (Genre, Studio, Artist, etc.)
	if err := shared_db.RunItemValuesMigrations(dbConn); err != nil {
		log.Fatalf("Failed to create ItemValues tables: %v", err)
	}

	playlistRepo := db.NewPlaylistRepository(dbConn)

	// Create router
	r := chi.NewRouter()

	r.Use(chi_middleware.RequestID)
	r.Use(chi_middleware.RealIP)
	r.Use(chi_middleware.Logger)
	r.Use(chi_middleware.Recoverer)
	r.Use(middleware.ResponseHeadersMiddleware())

	// Health check
	r.Get("/health", healthCheck)

	// Public routes
	r.Get("/api/Playlists", handlers.ListPlaylist)

	// Protected routes
	r.Route("/api/Playlists", func(r chi.Router) {
		r.Get("/", handlers.GetPlaylists)
		r.Post("/", handlers.CreatePlaylist)
		r.Get("/{playlistId}", handlers.GetPlaylist)
		r.Patch("/{playlistId}", handlers.UpdatePlaylist)
		r.Delete("/{playlistId}", handlers.DeletePlaylist)
		
		r.Post("/{playlistId}/AddToPlaylist", handlers.AddToPlaylist)
		r.Get("/{playlistId}/Items", handlers.GetPlaylistItems)
		r.Delete("/{playlistId}/RemoveFromPlaylist", handlers.RemoveFromPlaylist)
	})

	// Inject repository into context for protected routes
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), "playlistRepo", playlistRepo)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	})

	// Start server
	srv := &http.Server{
		Addr:         ":" + serverPort,
		Handler:      r,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	errChan := make(chan error, 1)
	go func() {
		log.Printf("🎵 Playlist Service starting on port %s", serverPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errChan:
		log.Printf("Server error: %v", err)
	case <-done:
		log.Println("Server shutting down...")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	log.Println("Server stopped")
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"Status":"ok"}`))
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// generateServerID creates a UUID for the server
func generateServerID() string {
	return uuid.New().String()
}
