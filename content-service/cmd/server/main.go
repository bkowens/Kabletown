package main

import (
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/jellyfinhanced/shared/db"
	"github.com/jellyfinhanced/shared/logger"
	"kabletown/content-service/internal/handlers"
)

var appLog = logger.NewLogger("content-service")

func main() {
	// Load DB config
	user := getenv("DB_USER", "kabletown")
	pass := getenv("DB_PASS", "kabletown")
	host := getenv("DB_HOST", "mysql")
	port := getenv("DB_PORT", "3306")
	dbName := getenv("DB_NAME", "kabletown")

	connStr := user + ":" + pass + "@tcp(" + host + ":" + port + ")/" + dbName + "?parseTime=true&loc=Local"

	// Initialize DB
	dbPool, err := db.NewMySQLPool(connStr)
	if err != nil {
		appLog.Fatal("Failed to connect to database", "error", err)
	}
	defer dbPool.Close()

	// Initialize handlers
	moviesHandler := handlers.NewMoviesHandler(dbPool)
	tvShowsHandler := handlers.NewTvShowsHandler(dbPool)
	artistsHandler := handlers.NewArtistsHandler(dbPool)
	channelsHandler := handlers.NewChannelsHandler(dbPool)
	liveTvHandler := handlers.NewLiveTvHandler(dbPool)
	collectionHandler := handlers.NewCollectionHandler(dbPool)
	playlistsHandler := handlers.NewPlaylistsHandler(dbPool)
	instantMixHandler := handlers.NewInstantMixHandler(dbPool)

	// Setup router
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(render.SetContentType(render.ContentTypeJSON))

	// Auth middleware
	r.Use(handlers.AuthMiddleware(dbPool))

	// Movies routes
	r.Route("/Movies", func(r chi.Router) {
		r.Get("/", moviesHandler.GetMovies)
		r.Get("/{itemId}", moviesHandler.GetMovie)
		r.Post("/{itemId}/Play", moviesHandler.PlayMovie)
	})

	// TV Shows routes
	r.Route("/TvShows", func(r chi.Router) {
		r.Get("/", tvShowsHandler.GetTvShows)
		r.Get("/{itemId}", tvShowsHandler.GetTvShow)
		r.Get("/{showId}/Episodes", tvShowsHandler.GetEpisodes)
		r.Get("/{showId}/Seasons", tvShowsHandler.GetSeasons)
		r.Get("/Season/{seasonId}/Episodes", tvShowsHandler.GetSeasonEpisodes)
	})

	// Artists routes
	r.Route("/Artists", func(r chi.Router) {
		r.Get("/", artistsHandler.GetArtists)
		  &r.Get("/{artistId}", artistsHandler.GetArtist)
		  &r.Get("/{artistId}/Albums", artistsHandler.GetArtistAlbums)
		  &r.Get("/{artistId}/Songs", artistsHandler.GetArtistSongs)
	})

	// Channels routes
	r.Route("/Channels", func(r chi.Router) {
		r.Get("/", channelsHandler.GetChannels)
		r.Get("/{itemId}", channelsHandler.GetChannel)
	})

	// LiveTV routes
	r.Route("/LiveTv", func(r chi.Router) {
		r.Get("/Programs", liveTvHandler.GetPrograms)
		r.Get("/Channels", liveTvHandler.GetLiveChannels)
		r.Get("/Channels/{channelId}/Programs", liveTvHandler.GetChannelPrograms)
		r.Get("/Programs/{programId}", liveTvHandler.GetProgram)
	})

	// Collections routes
	r.Route("/Collections", func(r chi.Router) {
		r.Get("/", collectionHandler.GetCollections)
		r.Post("/", collectionHandler.CreateCollection)
		r.Get("/{itemId}", collectionHandler.GetCollection)
		r.Post("/{itemId}/Items", collectionHandler.AddToCollection)
		r.Delete("/{itemId}/Items/{childId}", collectionHandler.RemoveFromCollection)
	})

	// Playlists routes
	r.Route("/Playlists", func(r chi.Router) {
		r.Get("/", playlistsHandler.GetPlaylists)
		r.Post("/", playlistsHandler.CreatePlaylist)
		r.Get("/{itemId}", playlistsHandler.GetPlaylist)
		r.Post("/{itemId}/Items", playlistsHandler.AddToPlaylist)
		r.Delete("/{itemId}/Items/{childId}", playlistsHandler.RemoveFromPlaylist)
	})

	// InstantMix routes
	r.Route("/Items", func(r chi.Router) {
		r.Get("/InstantMix", instantMixHandler.CreateInstantMix)
		r.Get("/{itemId}/InstantMix", instantMixHandler.CreateItemInstantMix)
	})

	// Start server
	portStr := getenv("PORT", "8009")

	appLog.Info("Starting content service", "port", portStr)
	err = http.ListenAndServe(":"+portStr, r)
	if err != nil {
		appLog.Fatal("Server failed to start", "error", err)
	}
}

func getenv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
