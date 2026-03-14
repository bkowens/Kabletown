module github.com/jellyfinhanced/auth-service

go 1.22

toolchain go1.22.2

require (
	github.com/go-chi/chi/v5 v5.0.10
	github.com/go-chi/cors v1.2.2
	github.com/google/uuid v1.6.0
	github.com/jellyfinhanced/shared v0.0.0-00010101000000-000000000000
	github.com/jmoiron/sqlx v1.4.0
	golang.org/x/crypto v0.0.0-20210817164053-32db794688a5
)

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/go-sql-driver/mysql v1.8.1 // indirect
)

replace github.com/jellyfinhanced/shared => ../shared
