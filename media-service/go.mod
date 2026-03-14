module kabletown/media-service

go 1.22

require (
    github.com/go-chi/chi/v5 v5.0.11
    github.com/go-chi/cors v1.2.1
    github.com/go-sql-driver/mysql v1.7.1
    golang.org/x/sync v0.7.0
    github.com/jellyfinhanced/shared v0.0.0
)

replace github.com/jellyfinhanced/shared = ../shared
