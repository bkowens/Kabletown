package main

import (
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jellyfinhanced/shared/db"
)

func main() {
	dsn := getEnv("DB_DSN", "root:password@tcp(localhost:3306)/kabletown?parseTime=true")
	dbp, err := db.NewDB(dsn)
	if err != nil {
		log.Fatalf("DB failed: %v", err)
	}
	defer dbp.Close()

	router := chi.NewRouter()
	router.Use(middleware.Logger)

	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	port := getEnv("SERVICE_PORT", "8008")
	log.Printf("Metadata service starting on :%s", port)
	http.ListenAndServe(":"+port, router) //nolint:errcheck
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
