package handlers

import (
	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"

	"github.com/jellyfinhanced/user-service/internal/db"
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
func (h *Handler) RegisterRoutes(r chi.Router) {
	// User CRUD
	r.Get("/Users", h.ListUsers)
	r.Post("/Users/New", h.CreateUser)
	r.Get("/Users/{userId}", h.GetUser)
	r.Put("/Users/{userId}", h.UpdateUser)
	r.Delete("/Users/{userId}", h.DeleteUser)
	r.Post("/Users/{userId}/Password", h.ChangePassword)
	r.Post("/Users/{userId}/Policy", h.UpdatePolicy)
	r.Post("/Users/{userId}/Configuration", h.UpdateConfiguration)

	// User library / views
	r.Get("/Users/{userId}/Items/Latest", h.GetLatestItems)
	r.Get("/Users/{userId}/Views", h.GetUserViews)

	// Display preferences
	r.Get("/DisplayPreferences/{displayPreferencesId}", h.GetDisplayPreferences)
	r.Post("/DisplayPreferences/{displayPreferencesId}", h.SetDisplayPreferences)

	// Favorites
	r.Post("/Users/{userId}/FavoriteItems/{itemId}", h.MarkFavorite)
	r.Delete("/Users/{userId}/FavoriteItems/{itemId}", h.UnmarkFavorite)

	// Played state
	r.Post("/Users/{userId}/PlayedItems/{itemId}", h.MarkPlayed)
	r.Post("/Users/{userId}/PlayedItems/{itemId}/Unplayed", h.MarkUnplayed)
	r.Post("/Users/{userId}/PlayingItems", h.MarkPlaying)
}