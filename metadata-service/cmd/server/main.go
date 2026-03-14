package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jellyfinhanced/shared/db"
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
	
	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	log.Println("Metadata service starting on :8013")
	http.ListenAndServe(":8013", router)
}
