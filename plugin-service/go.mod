module github.com/jellyfinhanced/plugin-service

go 1.22

toolchain go1.22.2

replace github.com/jellyfinhanced/shared => ../shared

require (
	github.com/go-chi/chi/v5 v5.0.10
	github.com/go-chi/cors v1.2.2
	github.com/jellyfinhanced/shared v0.0.0
)
