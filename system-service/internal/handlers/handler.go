package handlers

import (
	"github.com/go-chi/chi/v5"
)

// Handler holds all dependencies for system-service route handlers.
type Handler struct {
	serverID   string
	serverName string
}

// New creates a Handler.
func New(serverID, serverName string) *Handler {
	return &Handler{
		serverID:   serverID,
		serverName: serverName,
	}
}

// RegisterRoutes wires all system-service routes onto the given chi router.
// No auth middleware here — the gateway handles auth for most routes,
// and public endpoints (/System/Info/Public, /System/Ping) need no auth.
func (h *Handler) RegisterRoutes(r chi.Router) {
	// Public (no auth required)
	r.Get("/System/Info/Public", h.GetPublicSystemInfo)
	r.Get("/System/Ping", h.Ping)
	r.Post("/System/Ping", h.Ping)

	// System
	r.Get("/System/Info", h.GetSystemInfo)
	r.Post("/System/Restart", h.Restart)
	r.Post("/System/Shutdown", h.Shutdown)
	r.Get("/System/Logs", h.GetLogFiles)
	r.Get("/System/Endpoint", h.GetEndpointInfo)

	// Configuration
	r.Get("/System/Configuration", h.GetConfiguration)
	r.Post("/System/Configuration", h.UpdateConfiguration)
	r.Get("/System/Configuration/{key}", h.GetNamedConfiguration)
	r.Post("/System/Configuration/{key}", h.UpdateNamedConfiguration)

	// Branding
	r.Get("/Branding/Configuration", h.GetBrandingOptions)
	r.Post("/Branding/Configuration", h.UpdateBrandingOptions)
	r.Get("/Branding/Css", h.GetBrandingCss)

	// Activity log
	r.Get("/System/ActivityLog/Entries", h.GetActivityLog)

	// Environment
	r.Get("/Environment/DefaultDirectoryBrowser", h.GetDefaultDirectoryBrowser)
	r.Get("/Environment/Drives", h.GetDrives)
	r.Get("/Environment/ParentPath", h.GetParentPath)
	r.Post("/Environment/DirectoryContents", h.GetDirectoryContents)
	r.Get("/Environment/NetworkShares", h.GetNetworkShares)

	// Localization
	r.Get("/Localization/Options", h.GetLocalizationOptions)
	r.Get("/Localization/Countries", h.GetCountries)
	r.Get("/Localization/Cultures", h.GetCultures)
	r.Get("/Localization/ParentalRatings", h.GetParentalRatings)

	// Client log
	r.Post("/ClientLog/Document", h.SubmitClientLog)

	// Dashboard
	r.Get("/web/ConfigurationPage", h.GetConfigurationPage)
}
