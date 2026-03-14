// Library Service — content browsing and media library management.
// Port: 8003
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

	libraryHandlers "github.com/jellyfinhanced/library-service/internal/handlers"
	"github.com/jellyfinhanced/shared/auth"
	sharedDB "github.com/jellyfinhanced/shared/db"
	"github.com/jellyfinhanced/shared/response"
)

func main() {
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "3306")
	dbUser := getEnv("DB_USER", "jellyfin")
	dbPassword := getEnv("DB_PASSWORD", "")
	dbName := getEnv("DB_NAME", "jellyfin")
	servicePort := getEnv("SERVICE_PORT", "8003")
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
		log.Fatalf("library-service: failed to connect to database: %v", err)
	}
	defer sqlxDB.Close()

	itemsHandler := libraryHandlers.NewItemsHandler(sqlxDB)
	filterHandler := libraryHandlers.NewFilterHandler(sqlxDB)
	userLibHandler := libraryHandlers.NewUserLibHandler(sqlxDB)

	r := chi.NewRouter()
	r.Use(chiMiddleware.RealIP)
	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)
	r.Use(auth.NewAuthMiddleware(sqlxDB))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
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

	// Item routes
	r.Get("/Items", itemsHandler.GetItems)
	r.Get("/Items/NextUp", itemsHandler.GetNextUp)
	r.Get("/Items/Resume", itemsHandler.GetResume)
	r.Get("/Items/{id}", itemsHandler.GetItem)
	r.Get("/Items/{id}/Ancestors", itemsHandler.GetAncestors)

	// Filter routes
	r.Get("/Genres", filterHandler.GetGenres)
	r.Get("/Studios", filterHandler.GetStudios)
	r.Get("/Persons", filterHandler.GetPersons)
	r.Get("/Years", filterHandler.GetYears)

	// User library routes
	r.Get("/Users/{userId}/Views", userLibHandler.GetUserViews)

	addr := ":" + servicePort
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	serverErr := make(chan error, 1)
	go func() {
		log.Printf("library-service: listening on %s (server-id=%s)", addr, serverID)
		serverErr <- srv.ListenAndServe()
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("library-service: server error: %v", err)
		}
	case sig := <-quit:
		log.Printf("library-service: received signal %v — shutting down", sig)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("library-service: graceful shutdown failed: %v", err)
	}
	fmt.Println("library-service: shutdown complete")
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
