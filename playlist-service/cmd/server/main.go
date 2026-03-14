package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/jellyfinhanced/playlist-service/internal/db"
	"github.com/jellyfinhanced/playlist-service/internal/handlers"
	sharedDB "github.com/jellyfinhanced/shared/db"
	"github.com/jellyfinhanced/shared/response"
)

func main() {
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "3306")
	dbUser := getEnv("DB_USER", "jellyfin")
	dbPassword := getEnv("DB_PASSWORD", "")
	dbName := getEnv("DB_NAME", "jellyfin")
	servicePort := getEnv("SERVICE_PORT", "8013")
	serverID := getEnv("SERVER_ID", "00000000-0000-0000-0000-000000000000")

	response.SetServerID(serverID)

	sqlxDB, err := sharedDB.Connect(sharedDB.Config{
		Host:     dbHost,
		Port:     dbPort,
		User:     dbUser,
		Password: dbPassword,
		Name:     dbName,
	})
	if err != nil {
		log.Fatalf("playlist-service: failed to connect to database: %v", err)
	}
	defer sqlxDB.Close()

	if err := db.RunMigrations(sqlxDB); err != nil {
		log.Fatalf("playlist-service: migrations failed: %v", err)
	}

	repo := db.NewPlaylistRepository(sqlxDB)

	r := chi.NewRouter()
	r.Use(chiMiddleware.RealIP)
	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)
	r.Use(func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if !response.RequiredHeaders(w, []string{}) { return }
					next.ServeHTTP(w, r)
				})
			})
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"*"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"}) //nolint:errcheck
	})

	handlers.RegisterRoutes(r, repo)

	addr := ":" + servicePort
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	serverErr := make(chan error, 1)
	go func() {
		log.Printf("playlist-service: listening on %s (server-id=%s)", addr, serverID)
		serverErr <- srv.ListenAndServe()
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("playlist-service: server error: %v", err)
		}
	case sig := <-quit:
		log.Printf("playlist-service: received signal %v — shutting down", sig)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("playlist-service: graceful shutdown failed: %v", err)
	}
	fmt.Println("playlist-service: shutdown complete")
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}