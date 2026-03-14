package main

import (
    "context"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/bowens/kabletown/collection-service/internal/db"
    "github.com/bowens/kabletown/collection-service/internal/handlers"
    "github.com/bowens/kabletown/collection-service/internal/middleware"
    shared_db "github.com/bowens/kabletown/shared/db"
    "github.com/go-chi/chi/v5"
    chi_middleware "github.com/go-chi/chi/v5/middleware"
    "github.com/joho/godotenv"
)

func main() {
    _ = godotenv.Load()

    // Database configuration from environment
    cfg := &shared_db.Config{
        Host:         getEnv("DB_HOST", "localhost"),
        Port:         3306,
        User:         getEnv("DB_USER", "jellyfin"),
        Password:     getEnv("DB_PASSWORD", ""),
        DBName:       getEnv("DB_NAME", "jellyfin"),
        MaxOpenConns: 20,
        MaxIdleConns: 5,
    }

    // Server configuration
    serverPort := getEnv("SERVER_PORT", "8002")
    dataDir := getEnv("DATA_DIR", "/var/lib/jellyfin")

    // Initialize server ID
    log.Printf("Server ID: %s", middleware.InitializeServerID(dataDir))

    // Create database connection pool
    dbConn, err := shared_db.NewPool(cfg)
    if err != nil {
        log.Fatalf("Failed to initialize database: %v", err)
    }
    defer dbConn.Close()

    collectionRepo := db.NewCollectionRepository(dbConn)

    // Create Collections table if not exists
    createSQL := db.GetCreateCollectionsSQL()
    if _, err := dbConn.Exec(createSQL); err != nil {
        log.Fatalf("Failed to create Collections table: %v", err)
    }

    createItemsSQL := db.GetCreateCollectionItemsSQL()
    if _, err := dbConn.Exec(createItemsSQL); err != nil {
        log.Fatalf("Failed to create CollectionItems table: %v", err)
    }

    // Create ItemValues tables for P6 filtering (Genre, Studio, Artist, etc.)
    if err := shared_db.RunItemValuesMigrations(dbConn); err != nil {
        log.Fatalf("Failed to create ItemValues tables: %v", err)
    }

    r := chi.NewRouter()

    r.Use(chi_middleware.RequestID)
    r.Use(chi_middleware.RealIP)
    r.Use(chi_middleware.Logger)
    r.Use(chi_middleware.Recoverer)
    r.Use(middleware.ResponseHeadersMiddleware())

    // Health check
    r.Get("/health", healthCheck)

    // Public routes
    r.Get("/Collections", handlers.GetCollections)

    // Protected routes
    r.Route("/Collections", func(r chi.Router) {
        r.Post("/", handlers.CreateCollection)
        r.Get("/{collectionId}", handlers.GetCollection)
        r.Patch("/{collectionId}", handlers.UpdateCollection)
        r.Delete("/{collectionId}", handlers.DeleteCollection)

        r.Post("/{collectionId}/AddToCollection", handlers.AddToCollection)
        r.Delete("/{collectionId}/RemoveFromCollection", handlers.RemoveFromCollection)
    })

    // Inject repository into context for protected routes
    r.Use(func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            ctx := context.WithValue(r.Context(), "collectionRepo", collectionRepo)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    })

    srv := &http.Server{
        Addr:         ":" + serverPort,
        Handler:      r,
        ReadTimeout:  30 * time.Second,
        WriteTimeout: 30 * time.Second,
    }

    errChan := make(chan error, 1)

    go func() {
        log.Printf("Starting collection-service on :%s", serverPort)
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            errChan <- err
        }
    }()

    done := make(chan os.Signal, 1)
    signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

    select {
    case err := <-errChan:
        log.Printf("Server error: %v", err)
    case <-done:
        log.Println("Server shutting down...")
    }

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    if err := srv.Shutdown(ctx); err != nil {
        log.Printf("Server shutdown error: %v", err)
    }

    log.Println("Server stopped")
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json; charset=utf-8")
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("{\"status\":\"ok\"}"))
}

func getEnv(key, defaultValue string) string {
    if value, exists := os.LookupEnv(key); exists {
        return value
    }
    return defaultValue
}
