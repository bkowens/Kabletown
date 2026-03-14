package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"

	"github.com/bowens/kabletown/user-service/internal/db"
	"github.com/bowens/kabletown/shared/auth"
)

// Handler holds all dependencies for user-service route handlers.
type Handler struct {
	db           *sqlx.DB
	userRepo     *db.UserRepository
	userDataRepo *db.UserDataRepository
	displayRepo  *db.DisplayPrefsRepository
	serverID     string
}

// New creates a Handler wired to the given database connection and server UUID.
func New(sqlxDB *sqlx.DB, serverID string) *Handler {
	return &Handler{
		db:           sqlxDB,
		userRepo:     db.NewUserRepository(sqlxDB),
		userDataRepo: db.NewUserDataRepository(sqlxDB),
		displayRepo:  db.NewDisplayPrefsRepository(sqlxDB),
		serverID:     serverID,
	}
}

// RegisterRoutes wires all user-service routes onto the given chi router.
func (h *Handler) RegisterRoutes(r chi.Router, lookup auth.DeviceLookupFunc) {
	anonymousPaths := []string{"/healthz"}
	authMiddleware := auth.NewAuthMiddleware(lookup, anonymousPaths)

	r.Group(func(r chi.Router) {
		r.Use(authMiddleware)

		// User CRUD
		r.Get("/Users", auth.RequireAuth(http.HandlerFunc(h.ListUsers)).ServeHTTP)
		r.Post("/Users/New", auth.RequireAdmin(http.HandlerFunc(h.CreateUser)).ServeHTTP)
		r.Get("/Users/{userId}", auth.RequireAuth(http.HandlerFunc(h.GetUser)).ServeHTTP)
		r.Put("/Users/{userId}", auth.RequireAuth(http.HandlerFunc(h.UpdateUser)).ServeHTTP)
		r.Delete("/Users/{userId}", auth.RequireAdmin(http.HandlerFunc(h.DeleteUser)).ServeHTTP)
		r.Post("/Users/{userId}/Password", auth.RequireAuth(http.HandlerFunc(h.ChangePassword)).ServeHTTP)
		r.Post("/Users/{userId}/Policy", auth.RequireAdmin(http.HandlerFunc(h.UpdatePolicy)).ServeHTTP)
		r.Post("/Users/{userId}/Configuration", auth.RequireAuth(http.HandlerFunc(h.UpdateConfiguration)).ServeHTTP)

		// User library / views
		r.Get("/Users/{userId}/Items/Latest", auth.RequireAuth(http.HandlerFunc(h.GetLatestItems)).ServeHTTP)
		r.Get("/Users/{userId}/Views", auth.RequireAuth(http.HandlerFunc(h.GetUserViews)).ServeHTTP)

		// Display preferences
		r.Get("/DisplayPreferences/{displayPreferencesId}", auth.RequireAuth(http.HandlerFunc(h.GetDisplayPreferences)).ServeHTTP)
		r.Post("/DisplayPreferences/{displayPreferencesId}", auth.RequireAuth(http.HandlerFunc(h.SetDisplayPreferences)).ServeHTTP)

		// Favorites
		r.Post("/Users/{userId}/FavoriteItems/{itemId}", auth.RequireAuth(http.HandlerFunc(h.MarkFavorite)).ServeHTTP)
		r.Delete("/Users/{userId}/FavoriteItems/{itemId}", auth.RequireAuth(http.HandlerFunc(h.UnmarkFavorite)).ServeHTTP)

		// Played state
		r.Post("/Users/{userId}/PlayedItems/{itemId}", auth.RequireAuth(http.HandlerFunc(h.MarkPlayed)).ServeHTTP)
		r.Delete("/Users/{userId}/PlayedItems/{itemId}", auth.RequireAuth(http.HandlerFunc(h.MarkUnplayed)).ServeHTTP)
	})
}
