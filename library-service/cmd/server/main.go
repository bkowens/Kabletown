package main

import (
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-sql-driver/mysql"
	"github.com/jellyfinhanced/shared/db"
	"github.com/jellyfinhanced/shared/logger"
)

var appLog = logger.NewLogger("library-service")

func main() {
	user := getenv("DB_USER", "kabletown")
	pass := getenv("DB_PASS", "kabletown")
	host := getenv("DB_HOST", "mysql")
	port := getenv("DB_PORT", "3306")
	dbName := getenv("DB_NAME", "kabletown")

	connStr := user + ":" + pass + "@tcp(" + host + ":" + port + ")/" + dbName + "?parseTime=true&loc=Local"

	dbPool, err := db.NewMySQLPool(connStr)
	if err != nil {
		appLog.Fatal("Failed to connect to database", "error", err)
	}
	defer dbPool.Close()

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(render.SetContentType(render.ContentTypeJSON))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		render.JSON(w, r, map[string]string{"status": "library service ready"})
	})

	portStr := getenv("PORT", "8003")
	appLog.Info("Starting library service", "port", portStr)
	err = http.ListenAndServe(":"+portStr, r)
	if err != nil {
		appLog.Fatal("Server failed to start", "error", err)
	}
}

func getenv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
