package main

import (
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jellyfinhanced/shared/db"
	"github.com/jellyfinhanced/shared/response"
)

func main() {
	dsn := getEnv("DB_DSN", "root:password@tcp(localhost:3306)/kabletown?parseTime=true")
	dbPool, err := db.NewDB(dsn)
	if err != nil {
		log.Fatalf("DB failed: %v", err)
	}
	defer dbPool.Close()

	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		response.WriteJSON(w, http.StatusOK, map[string]string{"status": "healthy"})
	})

	router.Route("/Videos", func(r chi.Router) {
		r.Get("/{itemId}/master.m3u8", func(w http.ResponseWriter, r *http.Request) {
			response.WriteNotImplemented(w, "Not implemented")
		})
	})

	port := getEnv("SERVICE_PORT", "8006")
	log.Printf("Stream service starting on :%s", port)
	http.ListenAndServe(":"+port, router) //nolint:errcheck
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
