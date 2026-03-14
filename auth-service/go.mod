module github.com/bowens/kabletown/auth-service

go 1.22

require (
	github.com/bowens/kabletown/shared v0.0.0
	github.com/go-chi/chi/v5 v5.2.5
	github.com/go-chi/cors v1.2.1
	github.com/google/uuid v1.6.0
	github.com/jmoiron/sqlx v1.4.0
	golang.org/x/crypto v0.23.0
)

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/go-sql-driver/mysql v1.8.1 // indirect
)

replace github.com/bowens/kabletown/shared => ../shared
