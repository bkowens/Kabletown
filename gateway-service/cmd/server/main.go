package main

import (
    "context"
    "fmt"
    "log"
    "net/http"
    "net/http/httputil"
    "net/url"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"
    "github.com/go-chi/cors"
)

var (
    gatewayPort       = os.Getenv("GATEWAY_PORT")
    upstreamAuth      = os.Getenv("UPSTREAM_AUTH")
    upstreamItem      = os.Getenv("UPSTREAM_ITEM")
    upstreamStreaming = os.Getenv("UPSTREAM_STREAMING")
    upstreamTranscode = os.Getenv("UPSTREAM_TRANCODE")
)

type ProxyRouter struct {
    router *chi.Mux
}

func NewProxyRouter() *ProxyRouter {
    return &ProxyRouter{router: chi.NewRouter()}
}

func (pr *ProxyRouter) setupMiddleware() {
    pr.router.Use(middleware.Logger)
    pr.router.Use(middleware.Recoverer)
    pr.router.Use(middleware.RealIP)
    pr.router.Use(cors.Handler(cors.Options{
        AllowedOrigins: []string{"*"},
        AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "HEAD"},
        AllowedHeaders: []string{"*"},
        ExposedHeaders: []string{"Link", "Content-Length"},
        AllowCredentials: false,
        MaxAge: 300,
    }))
}

func (pr *ProxyRouter) setupRoutes() {
    pr.router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("OK"))
    })

    pr.router.Route("/Sessions", func(r chi.Router) {
        r.Get("/*", pr.proxyTo(upstreamAuth))
        r.Post("/*", pr.proxyTo(upstreamAuth))
        r.Delete("/*", pr.proxyTo(upstreamAuth))
    })

    pr.router.Route("/Devices", func(r chi.Router) {
        r.Post("/", pr.proxyTo(upstreamAuth))
        r.Get("/", pr.proxyTo(upstreamAuth))
    })

    pr.router.Route("/Items", func(r chi.Router) {
        r.Get("/{id}", pr.proxyTo(upstreamItem))
        r.Get("/", pr.proxyTo(upstreamItem))
    })

    pr.router.Route("/Videos", func(r chi.Router) {
        r.Get("/{itemId}/master.m3u8", pr.proxyTo(upstreamStreaming))
        r.Get("/{itemId}/stream.m3u8", pr.proxyTo(upstreamStreaming))
        r.Get("/{itemId}/stream", pr.proxyTo(upstreamStreaming))
    })

    pr.router.Route("/transcodes", func(r chi.Router) {
        r.Post("/", pr.proxyTo(upstreamTranscode))
        r.Get("/", pr.proxyTo(upstreamTranscode))
        r.Delete("/{id}", pr.proxyTo(upstreamTranscode))
    })
}

func (pr *ProxyRouter) proxyTo(baseURL string) func(http.ResponseWriter, *http.Request) {
    return func(w http.ResponseWriter, r *http.Request) {
        target, err := url.Parse(baseURL)
        if err != nil {
            http.Error(w, "Service unavailable", http.StatusBadGateway)
            return
        }
        proxy := httputil.NewSingleHostReverseProxy(target)
        proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
            log.Printf("Proxy error: %v", err)
            w.WriteHeader(http.StatusBadGateway)
        }
        proxy.ServeHTTP(w, r)
    }
}

func main() {
    if gatewayPort == "" { gatewayPort = "8013" }
    router := NewProxyRouter()
    router.setupMiddleware()
    router.setupRoutes()
    addr := fmt.Sprintf(":%s", gatewayPort)
    log.Printf("Starting gateway on %s", addr)
    server := &http.Server{
        Addr:        addr,
        Handler:     router.router,
        ReadTimeout: 15 * time.Second,
    }
    go func() {
        quit := make(chan os.Signal, 1)
        signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
        <-quit
        ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
        defer cancel()
        server.Shutdown(ctx)
    }()
    if err := server.ListenAndServe(); err != http.ErrServerClosed {
        log.Fatalf("Gateway error: %v", err)
    }
    log.Println("Gateway stopped")
}
