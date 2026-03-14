package handlers

import (
	"net/http"

	"github.com/bowens/kabletown/shared/response"
)

// GetSystemInfo handles GET /System/Info.
func (h *Handler) GetSystemInfo(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, http.StatusOK, map[string]interface{}{
		"LocalAddress":                   "http://localhost:8080",
		"StartupWizardCompleted":         true,
		"Version":                        "10.10.0",
		"ProductName":                    "Jellyfin Server",
		"OperatingSystem":                "Linux",
		"OperatingSystemDisplayName":     "Linux",
		"Id":                             h.serverID,
		"ServerName":                     h.serverName,
		"CanSelfRestart":                 false,
		"CanLaunchWebBrowser":            false,
		"ProgramDataPath":                "/config",
		"ItemsByNamePath":                "/config/metadata",
		"CachePath":                      "/cache",
		"LogPath":                        "/log",
		"InternalMetadataPath":           "/config/metadata",
		"TranscodingTempPath":            "/config/transcodes",
		"IsShuttingDown":                 false,
		"SupportsLibraryMonitor":         false,
		"WebSocketPortNumber":            8096,
		"CompletedInstallations":         []interface{}{},
		"HasPendingRestart":              false,
		"IsRestarting":                   false,
		"SupportsAutomaticPortMapping":   false,
		"HasUpdateAvailable":             false,
		"EncoderLocation":                "NotFound",
		"SystemArchitecture":             "X64",
	})
}

// GetPublicSystemInfo handles GET /System/Info/Public (no auth required).
func (h *Handler) GetPublicSystemInfo(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, http.StatusOK, map[string]interface{}{
		"LocalAddress":           "http://localhost:8080",
		"StartupWizardCompleted": true,
		"Version":                "10.10.0",
		"ProductName":            "Jellyfin Server",
		"OperatingSystem":        "Linux",
		"Id":                     h.serverID,
		"ServerName":             h.serverName,
	})
}

// Ping handles GET|POST /System/Ping.
func (h *Handler) Ping(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, http.StatusOK, "Jellyfin")
}

// Restart handles POST /System/Restart.
func (h *Handler) Restart(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

// Shutdown handles POST /System/Shutdown.
func (h *Handler) Shutdown(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

// GetLogFiles handles GET /System/Logs.
func (h *Handler) GetLogFiles(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, http.StatusOK, []interface{}{})
}

// GetEndpointInfo handles GET /System/Endpoint.
func (h *Handler) GetEndpointInfo(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, http.StatusOK, map[string]interface{}{
		"IsLocal":           true,
		"IsInNetwork":       true,
		"MaxStreamingBitrate": 140000000,
	})
}
