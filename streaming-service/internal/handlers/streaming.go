// Package handlers provides HTTP handlers for streaming service operations
package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jellyfinhanced/shared/auth"
	"github.com/jellyfinhanced/shared/dto"
	"github.com/jellyfinhanced/shared/response"
)

// Handler represents the streaming service HTTP handler
type Handler struct {
	db *sql.DB
}

// NewHandler creates a new streaming handler
func NewHandler(db *sql.DB) *Handler {
	return &Handler{db: db}
}

// SetupRoutes sets up all streaming routes
func (h *Handler) SetupRoutes(r chi.Router) {
	// HLS Master Playlist
	r.Get("/Videos/{itemId}/master.m3u8", h.MasterPlaylist)

	// HLS Variant Playlist
	r.Get("/Videos/{itemId}/stream.m3u8", h.VariantPlaylist)

	// HLS Segment
	r.Get("/Videos/{itemId}/segment/{segmentId}.ts", h.HLSSegment)

	// Progressive Stream (direct play)
	r.Get("/Videos/{itemId}/stream", h.ProgressiveStream)

	// HLS from transcode-service
	r.Get("/Items/{id}/Hls/{container}/{segmentId}", h.HLSFromTranscode)

	// Audio Universal Stream
	r.Get("/Audio/{itemId}/universal", h.AudioUniversal)
}

// MasterPlaylist returns HLS master playlist
// GET /Videos/{itemId}/master.m3u8
func (h *Handler) MasterPlaylist(w http.ResponseWriter, r *http.Request) {
	itemID := chi.URLParam(r, "itemId")
	if itemID == "" {
		response.WriteBadRequest(w, "Missing item ID")
		return
	}

	// Verify authentication
	if !auth.HasValidAuth(r) {
		response.WriteUnauthorized(w, "Unauthorized")
		return
	}

	// Build master playlist (placeholder - should query transcoded profiles)
	playlistContent := "#EXTM3U\n" +
		"#EXT-X-VERSION:3\n" +
		"#EXT-X-STREAM-INF:BANDWIDTH=8000000,RESOLUTION=1920x1080,CODECS=\"avc1.640028,mp4a.40.2\"\n" +
		"/Videos/" + itemID + "/stream_1080p.m3u8\n" +
		"#EXT-X-STREAM-INF:BANDWIDTH=4000000,RESOLUTION=1280x720,CODECS=\"avc1.64001f,mp4a.40.2\"\n" +
		"/Videos/" + itemID + "/stream_720p.m3u8\n" +
		"#EXT-X-STREAM-INF:BANDWIDTH=1500000,RESOLUTION=854x480,CODECS=\"avc1.64001f,mp4a.40.2\"\n" +
		"/Videos/" + itemID + "/stream_480p.m3u8\n"

	w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
	response.WriteJSON(w, http.StatusOK, playlistContent)
}

// VariantPlaylist returns HLS variant playlist
// GET /Videos/{itemId}/stream.m3u8
func (h *Handler) VariantPlaylist(w http.ResponseWriter, r *http.Request) {
	itemID := chi.URLParam(r, "itemId")
	if itemID == "" {
		response.WriteBadRequest(w, "Missing item ID")
		return
	}

	if !auth.HasValidAuth(r) {
		response.WriteUnauthorized(w, "Unauthorized")
		return
	}

	// Placeholder variant playlist
	playlistContent := "#EXTM3U\n" +
		"#EXT-X-VERSION:3\n" +
		"#EXT-X-TARGETDURATION:6\n" +
		"#EXT-X-MEDIA-SEQUENCE:0\n" +
		"#EXTINF:6.000000,\n" +
		"/Videos/" + itemID + "/segment_0.ts\n" +
		"#EXTINF:6.000000,\n" +
		"/Videos/" + itemID + "/segment_1.ts\n" +
		"#EXT-X-ENDLIST\n"

	w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
	response.WriteJSON(w, http.StatusOK, playlistContent)
}

// HLSSegment returns HLS segment
// GET /Videos/{itemId}/segment/{segmentId}.ts
func (h *Handler) HLSSegment(w http.ResponseWriter, r *http.Request) {
	itemID := chi.URLParam(r, "itemId")
	segmentID := chi.URLParam(r, "segmentId")

	if itemID == "" || segmentID == "" {
		response.WriteBadRequest(w, "Missing item or segment ID")
		return
	}

	if !auth.HasValidAuth(r) {
		response.WriteUnauthorized(w, "Unauthorized")
		return
	}

	// For now, return dummy TS data
	// In production, read from transcoded segment cache
	w.Header().Set("Content-Type", "video/MP2T")
	w.Header().Set("Content-Length", "1024")
	w.WriteHeader(http.StatusOK)
	w.Write(make([]byte, 1024))
}

// ProgressiveStream serves progressive download stream
// GET /Videos/{itemId}/stream
func (h *Handler) ProgressiveStream(w http.ResponseWriter, r *http.Request) {
	itemID := chi.URLParam(r, "itemId")
	if itemID == "" {
		response.WriteBadRequest(w, "Missing item ID")
		return
	}

	if !auth.HasValidAuth(r) {
		response.WriteUnauthorized(w, "Unauthorized")
		return
	}

	// Placeholder - should stream actual media file
	response.WriteNotImplemented(w, "Progressive streaming not yet implemented")
}

// HLSFromTranscode proxies to transcode-service
// GET /Items/{id}/Hls/{container}/{segmentId}
func (h *Handler) HLSFromTranscode(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	container := chi.URLParam(r, "container")
	segmentIDParam := chi.URLParam(r, "segmentId")

	if id == "" || container == "" || segmentIDParam == "" {
		response.WriteBadRequest(w, "Missing required parameters")
		return
	}

	if !auth.HasValidAuth(r) {
		response.WriteUnauthorized(w, "Unauthorized")
		return
	}

	// Proxy to transcode-service
	response.WriteNotImplemented(w, "Proxy to transcode-service not yet implemented")
}

// AudioUniversal returns universal audio stream
// GET /Audio/{itemId}/universal
func (h *Handler) AudioUniversal(w http.ResponseWriter, r *http.Request) {
	itemID := chi.URLParam(r, "itemId")
	if itemID == "" {
		response.WriteBadRequest(w, "Missing item ID")
		return
	}

	if !auth.HasValidAuth(r) {
		response.WriteUnauthorized(w, "Unauthorized")
		return
	}

	// Get transcoding info from query params
	directPlay := r.URL.Query().Get("static") == "true"

	if directPlay {
		// Direct play - stream original file
		response.WriteNotImplemented(w, "Direct play not yet implemented")
		return
	}

	// Transcode (placeholder)
	response.WriteNotImplemented(w, "Audio transcoding not yet implemented")
}

// GetStreamInfo returns streamable media info for an item
// Used for building master playlists
func (h *Handler) GetStreamInfo(itemID string) (*dto.MediaSourceInfo, error) {
	if itemID == "" {
		return nil, sql.ErrNoRows
	}

	// Placeholder implementation - should query library-service
	return &dto.MediaSourceInfo{
		Id:                    itemID,
		Protocol:              "Http",
		Container:             "ts",
		SupportsDirectPlay:    true,
		SupportsDirectStream:  true,
		MediaStreams:          []dto.MediaStreamInfo{},
	}, nil
}

// GetStreamURL constructs a stream URL for an item
func (h *Handler) GetStreamURL(itemID string, isHLS bool) (string, error) {
	if itemID == "" {
		return "", sql.ErrNoRows
	}

	if isHLS {
		return "/Videos/" + itemID + "/master.m3u8", nil
	}

	return "/Videos/" + itemID + "/stream", nil
}

// CheckTranscodeRequirements checks if transcoding is needed
func (h *Handler) CheckTranscodeRequirements(itemID string) (bool, error) {
	// Placeholder - should check client capabilities vs media properties
	return false, nil
}

// GenerateHLSManifest generates full HLS manifest for an item
func (h *Handler) GenerateHLSManifest(itemID string) (string, error) {
	streamInfo, err := h.GetStreamInfo(itemID)
	if err != nil {
		return "", err
	}

	// Generate playlist based on stream info
	// This is a simplified placeholder
	_ = streamInfo // Use streamInfo in proper implementation
	return "#EXTM3U\n#EXT-X-VERSION:3\n#EXT-X-STREAM-INF:BANDWIDTH=8000000\nstream.m3u8", nil
}

// ProcessTranscodeRequest handles transcoding request parameters
func (h *Handler) ProcessTranscodeRequest(r *http.Request) (map[string]string, error) {
	// Extract transcoding parameters from query
	return map[string]string{
		"container":  r.URL.Query().Get("Container"),
		"videoCodec": r.URL.Query().Get("VideoCodec"),
		"audioCodec": r.URL.Query().Get("AudioCodec"),
		"maxWidth":   r.URL.Query().Get("MaxWidth"),
		"maxHeight":  r.URL.Query().Get("MaxHeight"),
		"maxBitrate": r.URL.Query().Get("MaxBitrate"),
	}, nil
}

// EncodeResponseAsJSON encodes any data as JSON response
func (h *Handler) EncodeResponseAsJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
