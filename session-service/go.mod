module kabletown/session-service

go 1.21

require (
	github.com/go-chi/chi/v5 v5.0.11
	github.com/go-chi/render v1.0.3
	github.com/go-sql-driver/mysql v1.7.1
	github.com/google/uuid v1.4.0
	github.com/golang-jwt/jwt/v5 v5.0.0
	golang.org/x/sync v0.5.0
	github.com/jellyfinhanced/shared v0.0.0
)

replace github.com/jellyfinhanced/shared => ../shared
