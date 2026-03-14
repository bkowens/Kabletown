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

	"github.com/jellyfinhanced/session-service/internal/handlers"
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
	servicePort := getEnv("SERVICE_PORT", "8007")
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
		log.Fatalf("session-service: failed to connect to database: %v", err)
	}
	defer sqlxDB.Close()

	sessionHandler := handlers.NewSessionHandler(sqlxDB.DB)
	deviceHandler := handlers.NewDeviceHandler(sqlxDB.DB)
	syncPlayHandler := handlers.NewSyncPlayHandler(sqlxDB.DB)

	r := chi.NewRouter()
	r.Use(chiMiddleware.RealIP)
	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"*"},
		AllowCredentials: false,
		MaxAge:           300,
	}))
	r.Use(func(next http.Handler) http.Handler {
		return auth.AuthMiddleware(sqlxDB, next)
	})

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"}) //nolint:errcheck
	})

	// Session routes
	r.Route("/Sessions", func(r chi.Router) {
		r.Get("/", sessionHandler.GetSessions)
		r.Post("/{sessionId}/Playing", sessionHandler.ReportPlaying)
		r.Post("/{sessionId}/Playing/Stopped", sessionHandler.ReportStopped)
		r.Post("/{sessionId}/Playing/Progress", sessionHandler.ReportPlaying)
		r.Delete("/Logout", sessionHandler.CloseSession)
		r.Post("/{sessionId}/Message", sessionHandler.SendMessageToSession)
		r.Post("/{sessionId}/Capabilities", sessionHandler.UpdateSessionCapability)
		r.Post("/{sessionId}/Capability", sessionHandler.UpdateSessionCapability)
	})

	// Device routes
	r.Route("/Devices", func(r chi.Router) {
		r.Get("/", deviceHandler.GetDevices)
		r.Delete("/", deviceHandler.DeleteDevice)
	})

	// SyncPlay routes
	r.Route("/SyncPlay", func(r chi.Router) {
		r.Get("/List", syncPlayHandler.GetGroups)
		r.Post("/Create", syncPlayHandler.CreateGroup)
		r.Post("/Join", syncPlayHandler.JoinGroup)
		r.Post("/Leave", syncPlayHandler.LeaveGroup)
		r.Post("/Send", syncPlayHandler.SendCommand)
	})

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
		log.Printf("session-service: listening on %s (server-id=%s)", addr, serverID)
		serverErr <- srv.ListenAndServe()
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("session-service: server error: %v", err)
		}
	case sig := <-quit:
		log.Printf("session-service: received signal %v — shutting down", sig)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("session-service: graceful shutdown failed: %v", err)
	}
	fmt.Println("session-service: shutdown complete")
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
