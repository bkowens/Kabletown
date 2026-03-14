module github.com/bowens/kabletown/system-service

go 1.22

require (
	github.com/bowens/kabletown/shared v0.0.0
	github.com/go-chi/chi/v5 v5.2.5
	github.com/go-chi/cors v1.2.1
)

replace github.com/bowens/kabletown/shared => ../shared
