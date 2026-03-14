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

	"github.com/bowens/kabletown/auth-service/internal/db"
	"github.com/bowens/kabletown/auth-service/internal/handlers"
	"github.com/bowens/kabletown/shared/auth"
	sharedDB "github.com/bowens/kabletown/shared/db"
	"github.com/bowens/kabletown/shared/response"
)

func main() {
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "3306")
	dbUser := getEnv("DB_USER", "jellyfin")
	dbPassword := getEnv("DB_PASSWORD", "")
	dbName := getEnv("DB_NAME", "jellyfin")
	servicePort := getEnv("SERVICE_PORT", "8001")
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
		log.Fatalf("auth-service: failed to connect to database: %v", err)
	}
	defer sqlxDB.Close()

	if err := db.RunMigrations(sqlxDB); err != nil {
		log.Fatalf("auth-service: migrations failed: %v", err)
	}

	h := handlers.New(sqlxDB, serverID)
	resolver := db.NewTokenResolver(sqlxDB)

	lookup := auth.DeviceLookupFunc(func(token string) (string, bool, error) {
		return resolver.ResolveToken(token)
	})

	r := chi.NewRouter()
	r.Use(chiMiddleware.RealIP)
	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)
	r.Use(response.RequiredHeaders)
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

	h.RegisterRoutes(r, lookup)

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
		log.Printf("auth-service: listening on %s (server-id=%s)", addr, serverID)
		serverErr <- srv.ListenAndServe()
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("auth-service: server error: %v", err)
		}
	case sig := <-quit:
		log.Printf("auth-service: received signal %v — shutting down", sig)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("auth-service: graceful shutdown failed: %v", err)
	}
	fmt.Println("auth-service: shutdown complete")
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
