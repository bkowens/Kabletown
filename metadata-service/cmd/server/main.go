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
	"kabletown/metadata-service/internal/handlers"
)

var appLog = logger.NewLogger("metadata-service")

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
	itemRefreshHandler := handlers.NewItemRefreshHandler(dbPool)
	itemUpdateHandler := handlers.NewItemUpdateHandler(dbPool)
	itemLookupHandler := handlers.NewItemLookupHandler(dbPool)
	remoteImageHandler := handlers.NewRemoteImageHandler(dbPool)
	scheduledTaskHandler := handlers.NewScheduledTaskHandler(dbPool)

	// Setup router
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(render.SetContentType(render.ContentTypeJSON))

	// Auth middleware
	r.Use(handlers.AuthMiddleware(dbPool))

	// Item Refresh routes
	r.Route("/Items/Refresh", func(r chi.Router) {
		r.Post("/", itemRefreshHandler.RefreshAll)
		r.Post("/{itemId}", itemRefreshHandler.RefreshItem)
		r.Post("/{itemId}/Partial", itemRefreshHandler.RefreshPartial)
	})

	// Item Update routes
	r.Route("/Items", func(r chi.Router) {
		r.Post("/Update", itemUpdateHandler.UpdateItem)
		r.Post("/Update/RemoteSearch", itemUpdateHandler.RemoteSearch)
		r.Post("/{itemId}/Update", itemUpdateHandler.UpdateSpecificItem)
	})

	// Item Lookup routes
	r.Route("/Items/Lookup", func(r chi.Router) {
		r.Get("/{itemId}/ProviderInfo", itemLookupHandler.GetProviderInfo)
		r.Get("/{itemId}/RemoteSearchMetadata", itemLookupHandler.SearchMetadata)
	})

	// Remote Image routes
	r.Route("/Items/{itemId}/Images", func(r chi.Router) {
		r.Get("/Remote", remoteImageHandler.GetRemoteImages)
		r.Get("/Remote/Providers", remoteImageHandler.GetImageProviders)
		r.Post("/Remote/Download", remoteImageHandler.DownloadRemoteImage)
		r.Post("/{imageType}/Remote/Upload", remoteImageHandler.UploadRemoteImage)
		r.Delete("/{imageType}/Remote", remoteImageHandler.DeleteRemoteImage)
	})

	// Scheduled Tasks routes
	r.Route("/ScheduledTasks", func(r chi.Router) {
		r.Get("/", scheduledTaskHandler.ListTasks)
		r.Get("/{id}", scheduledTaskHandler.GetTask)
		r.Post("/{id}/Start", scheduledTaskHandler.StartTask)
		r.Post("/{id}/Stop", scheduledTaskHandler.StopTask)
	})

	// Start server
	portStr := getenv("PORT", "8008")

	appLog.Info("Starting metadata service", "port", portStr)
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
