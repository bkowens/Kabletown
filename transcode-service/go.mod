module github.com/jellyfinhanced/transcode-service

go 1.22

require (
	github.com/go-chi/chi/v5 v5.2.5
	github.com/go-chi/cors v1.2.1
	github.com/jellyfinhanced/shared v0.0.0
	github.com/jmoiron/sqlx v1.4.0
)

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/go-sql-driver/mysql v1.8.1 // indirect
	github.com/google/uuid v1.6.0 // indirect
)

replace github.com/jellyfinhanced/shared => ../shared
