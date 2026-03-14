package main

import (
    "context"
    "database/sql"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"
    "github.com/go-chi/cors"
    _ "github.com/go-sql-driver/mysql"
    "kabletown/shared/auth"
)

var (
    dbDSN       = os.Getenv("DB_DSN")
    serverPort  = os.Getenv("ITEM_SERVICE_PORT")
    serviceName = "item-service"
)

type Server struct {
    router *chi.Mux
    db     *sql.DB
    port   string
}

func NewServer(db *sql.DB, port string) *Server {
    s := &Server{
        router: chi.NewRouter(),
        db:     db,
        port:   port,
    }
    s.setupMiddleware()
    s.setupRoutes()
    return s
}

func (s *Server) setupMiddleware() {
    s.router.Use(middleware.Logger)
    s.router.Use(middleware.Recoverer)
    s.router.Use(middleware.RequestID)
    s.router.Timeout(5 * time.Minute) // Item queries can be complex
    
    s.router.Use(cors.Handler(cors.Options{
        AllowedOrigins:  []string{"*"},
        AllowedMethods:  []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowedHeaders:  []string{"Accept", "Authorization", "Content-Type", "X-Emby-Authorization"},
        ExposedHeaders:  []string{"Link"},
        AllowCredentials: true,
        MaxAge:           300,
    }))
}

func (s *Server) setupRoutes() {
    s.router.Get("/health", s.healthHandler)
    s.router.Get("/ready", s.readyHandler)

    // Protected routes
    api := s.router.Group(func(r chi.Router) {
        r.Use(auth.DBTokenValidator(s.db))
        
        // Items endpoints
        r.Get("/Items", s.listItemsHandler)           // InternalItemsQuery with 80+ params
        r.Get("/Items/{id}", s.getItemHandler)
        r.Post("/Items", s.createItemHandler)
        r.Put("/Items/{id}", s.updateItemHandler)
        r.Delete("/Items/{id}", s.deleteItemHandler)
        
        // User items
        r.Get("/Users/{userId}/Items", s.listUserItemsHandler)
        r.Get("/Users/{userId}/Latest", s.getLatestItemsHandler)
        
        // Hierarchy
        r.Get("/Items/{id}/Ancestors", s.getItemAncestorsHandler)
        r.Get("/Items/{id}/Descendants", s.getItemDescendantsHandler)
        
        // Similar items
        r.Get("/Items/{id}/Similar", s.getSimilarItemsHandler)
    })
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{
        "status":   "healthy",
        "service":  serviceName,
        "timestamp": time.Now().UTC().Format(time.RFC3339),
    })
}

func (s *Server) readyHandler(w http.ResponseWriter, r *http.Request) {
    ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
    defer cancel()

    if err := s.db.PingContext(ctx); err != nil {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusServiceUnavailable)
        json.NewEncoder(w).Encode(map[string]string{
            "status":  "unhealthy",
            "database": "connection failed",
        })
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{
        "status":   "ready",
        "database":  "connected",
    })
}

func (s *Server) listItemsHandler(w http.ResponseWriter, r *http.Request) {
    // Parse InternalItemsQuery params (80+ params)
    // Ref: docs/internal-items-query.md
    query := r.URL.Query()
    
    // Common params
    userId := query.Get("userId")
    parentId := query.Get("parentId")
    includeItemTypes := query.Get("IncludeItemTypes")
    excludeItemTypes := query.Get("ExcludeItemTypes")
    sortBy := query.Get("SortBy")
    sortOrder := query.Get("SortOrder")
    limit := query.Get("Limit")
    startIndex := query.Get("StartIndex")
    
    // P6 metadata filters
    genreIds := query.Get("GenreIds")
    studioIds := query.Get("StudioIds")
    artistIds := query.Get("ArtistIds")
    
    // P7 fast filters  
    topParentId := query.Get("TopParentId")
    
    // Build complex query (see internal/query/items.go)
    items, totalRecord, err := s.queryItems(r.Context(), userId, parentId, 
        includeItemTypes, excludeItemTypes, sortBy, sortOrder,
        limit, startIndex, genreIds, studioIds, artistIds, topParentId)
    
    if err != nil {
        log.Printf("Error querying items: %v", err)
        http.Error(w, "Failed to query items", http.StatusInternalServerError)
        return
    }
    
    response := map[string]interface{}{
        "Items":       items,
        "TotalRecord": totalRecord,
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

func (s *Server) queryItems(ctx context.Context, userId, parentId, includeTypes, 
    excludeTypes, sortBy, sortOrder, limit, startIndex, genreIds, studioIds, 
    artistIds, topParentId string) ([]map[string]interface{}, int64, error) {
    
    // TODO: Implement complex SQL query with:
    // - P7 index on (TopParentId, Type)
    // - P6 index on (ItemValues)
    // - Recursive CTE for hierarchy
    // - JOINs for metadata
    // 
    // Ref: docs/internal-items-query.md for full param mapping
    
    // Placeholder: return empty results
    return []map[string]interface{}{}, 0, nil
}

// Placeholder handlers
func (s *Server) getItemHandler(w http.ResponseWriter, r *http.Request) {
    http.Error(w, "Not implemented", http.StatusNotImplemented)
}
func (s *Server) createItemHandler(w http.ResponseWriter, r *http.Request) {
    http.Error(w, "Not implemented", http.StatusNotImplemented)
}
func (s *Server) updateItemHandler(w http.ResponseWriter, r *http.Request) {
    http.Error(w, "Not implemented", http.StatusNotImplemented)
}
func (s *Server) deleteItemHandler(w http.ResponseWriter, r *http.Request) {
    http.Error(w, "Not implemented", http.StatusNotImplemented)
}
func (s *Server) listUserItemsHandler(w http.ResponseWriter, r *http.Request) {
    http.Error(w, "Not implemented", http.StatusNotImplemented)
}
func (s *Server) getLatestItemsHandler(w http.ResponseWriter, r *http.Request) {
    http.Error(w, "Not implemented", http.StatusNotImplemented)
}
func (s *Server) getItemAncestorsHandler(w http.ResponseWriter, r *http.Request) {
    http.Error(w, "Not implemented", http.StatusNotImplemented)
}
func (s *Server) getItemDescendantsHandler(w http.ResponseWriter, r *http.Request) {
    http.Error(w, "Not implemented", http.StatusNotImplemented)
}
func (s *Server) getSimilarItemsHandler(w http.ResponseWriter, r *http.Request) {
    http.Error(w, "Not implemented", http.StatusNotImplemented)
}

func main() {
    if dbDSN == "" {
        log.Fatal("DB_DSN environment variable required")
    }

    db, err := sql.Open("mysql", dbDSN)
    if err != nil {
        log.Fatalf("Failed to open database: %v", err)
    }
    defer db.Close()

    if err := db.Ping(); err != nil {
        log.Fatalf("Failed to connect to database: %v", err)
    }
    log.Println("Database connection established")

    port := fmt.Sprintf(":%s", serverPort)
    server := NewServer(db, serverPort)

    log.Printf("Starting %s on %s", serviceName, port)

    httpServer := &http.Server{
        Addr:         port,
        Handler:      server.router,
        ReadTimeout:  15 * time.Second,
        WriteTimeout: 300 * time.Second, // Longer for complex queries
        IdleTimeout:  120 * time.Second,
    }

    go func() {
        quit := make(chan os.Signal, 1)
        signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
        <-quit
        log.Println("Shutting down server...")

        ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
        defer cancel()

        if err := httpServer.Shutdown(ctx); err != nil {
            log.Printf("Server forced to shutdown: %v", err)
        }
        log.Println("Server exited")
    }()

    if err := httpServer.ListenAndServe(); err != http.ErrServerClosed {
        log.Fatalf("Server error: %v", err)
    }
}
