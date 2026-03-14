package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jellyfinhanced/shared/db"
	"github.com/jellyfinhanced/shared/response"
)

func main() {
	dsn := "root:password@tcp(localhost:3306)/kabletown?parseTime=true"
	dbp, err := db.NewDB(dsn)
	if err != nil {
		log.Fatalf("DB failed: %v", err)
	}
	defer dbp.Close()

	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge: 300,
	}))

	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		response.WriteJSON(w, http.StatusOK, map[string]string{"status": "healthy"})
	})

	router.Route("/Artists", func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			response.WriteNotImplemented(w, "Not implemented")
		})
	})

	log.Println("Content service starting on :8010")
	http.ListenAndServe(":8010", router)
}
