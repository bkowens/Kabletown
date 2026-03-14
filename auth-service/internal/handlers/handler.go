package handlers

import (
	"github.com/go-chi/chi/v5"
	"github.com/jellyfinhanced/auth-service/internal/db"
	"github.com/jellyfinhanced/shared/auth"
	"github.com/jmoiron/sqlx"
)

// Handler holds all dependencies for auth-service route handlers.
type Handler struct {
	db            *sqlx.DB
	userRepo      *db.UserRepository
	apiKeyRepo    *db.ApiKeyRepository
	tokenResolver *db.TokenResolver
	serverID      string
}

// New creates a Handler wired to the given database connection and server UUID.
func New(sqlxDB *sqlx.DB, serverID string) *Handler {
	return &Handler{
		db:         sqlxDB,
		userRepo:   db.NewUserRepository(sqlxDB),
		apiKeyRepo: db.NewApiKeyRepository(sqlxDB),
		serverID:   serverID,
	}
}

// SetTokenResolver sets the token resolver after construction.
func (h *Handler) SetTokenResolver(tr *db.TokenResolver) {
	h.tokenResolver = tr
}

// RegisterRoutes wires all auth-service routes onto the given chi router.
func (h *Handler) RegisterRoutes(r chi.Router) {
	// Startup endpoints (public)
	r.Get("/Startup/Configuration", h.StartupConfiguration)
	r.Get("/Startup/RemoteAccessConfiguration", h.StartupConfiguration)
	r.Post("/Startup/Complete", h.StartupComplete)
	r.Get("/Startup/RemoteAccess", h.GetRemoteAccess)
	r.Post("/Startup/RemoteAccess", h.SetRemoteAccess)
	r.Get("/Startup/User", h.GetStartupUser)
	r.Post("/Startup/User", h.SetStartupUser)

	// Auth endpoints
	r.Post("/Users/AuthenticateByName", h.AuthenticateByName)
	r.Get("/Users/AuthenticateWithQuickConnect", h.AuthenticateWithQuickConnect)

	// Password endpoints
	r.Post("/Users/ForgotPassword", h.ForgotPassword)
	r.Post("/Users/ForgotPasswordPin", h.ForgotPasswordPin)

	// QuickConnect endpoints
	r.Get("/QuickConnect/Enabled", h.QuickConnectEnabled)
	r.Post("/QuickConnect/Initiate", h.QuickConnectInitiate)
	r.Post("/QuickConnect/Connect", h.QuickConnectConnect)
	r.Post("/QuickConnect/Authorize", h.QuickConnectAuthorize)
}

// generateToken creates a 256-bit hex random token.
func generateToken() (string, error) {
	return auth.GenerateToken(), nil
}
