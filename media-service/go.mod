module github.com/jellyfinhanced/media-service

go 1.22

require (
	github.com/go-chi/chi/v5 v5.2.5
	github.com/go-chi/cors v1.2.1
	github.com/google/uuid v1.6.0
	github.com/jellyfinhanced/shared v0.0.0
)

replace github.com/jellyfinhanced/shared => ../shared
