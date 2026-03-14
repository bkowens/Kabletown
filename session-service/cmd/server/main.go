package main

import (
	"context"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/jellyfinhanced/shared/db"
	"github.com/jellyfinhanced/shared/logger"
	"kabletown/session-service/internal/handlers"
)

var appLog = logger.NewLogger("session-service")

func main() {
	// Load DB config
	user := getenv("DB_USER", "kabletown")
	pass := getenv("DB_PASS", "kabletown")
	host := getenv("DB_HOST", "mysql")
	port := getenv("DB_PORT", "3306")
	dbName := getenv("DB_NAME", "kabletown")

	connStr := user + ":" + pass + "@tcp(" + host + ":" + port + ")/" + dbName + "?parseTime=true&loc=Local"

	// Initialize DB
	dbPool, err := db.NewMySQLPool(connStr)
	if err != nil {
		appLog.Fatal("Failed to connect to database", "error", err)
	}
	defer dbPool.Close()

	// Initialize handlers
	sessionHandler := handlers.NewSessionHandler(dbPool)
	deviceHandler := handlers.NewDeviceHandler(dbPool)
	syncPlayHandler := handlers.NewSyncPlayHandler(dbPool)

	// Setup router
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(render.SetContentType(render.ContentTypeJSON))

	// Auth middleware
	r.Use(handlers.AuthMiddleware(dbPool))

	// Session routes
	r.Route("/Sessions", func(r chi.Router) {
		r.Get("/", sessionHandler.GetSessions)
		r.Post("/", sessionHandler.CreateSession)
		r.Get("/{sessionId}", sessionHandler.GetSession)
		r.Post("/{sessionId}/Activity", sessionHandler.ReportSessionActivity)
		r.Post("/{sessionId}/ViewingItem/{itemId}", sessionHandler.ReportViewingItem)
		r.Post("/{sessionId}/Playing", sessionHandler.ReportPlaying)
		r.Post("/{sessionId}/Stopped", sessionHandler.ReportStopped)
		r.Post("/{sessionId}/Capability", sessionHandler.UpdateSessionCapability)
		r.Post("/{sessionId}/Message", sessionHandler.SendMessageToSession)
		r.Delete("/", sessionHandler.CloseSession)
		r.Delete("/{sessionId}", sessionHandler.CloseSpecificSession)
		r.Post("/{sessionId}/KeepAlive", sessionHandler.KeepAlive)
	})

	// Device routes
	r.Route("/Devices", func(r chi.Router) {
		r.Get("/", deviceHandler.GetDevices)
		r.With(middleware.Timeout(30 * time.Second)).Post("/", func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
			defer cancel()
			http.Redirect(w, r.WithContext(ctx), r.URL.Path, http.StatusCreated)
		})
		r.Put("/{id}/name", deviceHandler.UpdateDeviceName)
		r.Delete("/{id}", deviceHandler.DeleteDevice)
		r.Get("/{id}", deviceHandler.GetDevice)
	})

	// SyncPlay routes
	r.Route("/SyncPlay", func(r chi.Router) {
		r.Get("/List", syncPlayHandler.GetGroups)
		r.Post("/Create", syncPlayHandler.CreateGroup)
		r.Post("/Join", syncPlayHandler.JoinGroup)
		r.Post("/Leave", syncPlayHandler.LeaveGroup)
		r.Post("/Send", syncPlayHandler.SendCommand)
	})

	// Start server
	portStr := getenv("PORT", "8007")

	appLog.Info("Starting session service", "port", portStr)
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
