// Package auth provides HTTP request authentication utilities
package auth

import (
	"net/http"
	"strings"
)

// HeaderFields represents parsed MediaBrowser authorization header
type HeaderFields struct {
	Token     string
	DeviceID  string
	Client    string
	Version   string
	Device    string
	Platform  string
	Exts      string
	SessionID string
}

// ParseError represents a header parsing error
type ParseError struct {
	Message string
}

func (e *ParseError) Error() string {
	return e.Message
}

// ParseEmbyAuthHeader parses the X-Emby-Authorization / MediaBrowser header
// Format: MediaBrowser Token="abc123", DeviceId="xyz", Client="JellyfinWeb", Version="10.8.0"
// Returns: (token, deviceID, nil) on success or ("", "", error) on failure
func ParseEmbyAuthHeader(header string) (token, deviceID string, err error) {
	if header == "" {
		return "", "", &ParseError{"Authorization header is empty"}
	}

	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "MediaBrowser") {
		return "", "", &ParseError{"Invalid authorization scheme (expected MediaBrowser)"}
	}

	fieldStr := parts[1]

	// Parse comma-separated key=value pairs
	pairs := strings.Split(fieldStr, ",")
	for _, pair := range pairs {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}

		eqIdx := strings.Index(pair, "=")
		if eqIdx < 0 {
			continue
		}

		key := strings.TrimSpace(pair[:eqIdx])
		value := strings.TrimSpace(pair[eqIdx+1:])

		// Remove surrounding quotes if present
		if len(value) >= 2 {
			if value[0] == '"' && value[len(value)-1] == '"' {
				value = value[1 : len(value)-1]
			}
		}

		switch strings.ToLower(key) {
		case "token":
			token = value
		case "deviceid":
			deviceID = value
		case "client":
			_ = value // Could store if needed
		case "version":
			_ = value // Could store if needed
		case "device":
			_ = value // Could store if needed
		}
	}

	if token == "" {
		return "", "", &ParseError{"Token is required"}
	}

	return token, deviceID, nil
}

// ParseMediaBrowserHeader parses the MediaBrowser authorization header
// Returns full HeaderFields struct with all available fields
func ParseMediaBrowserHeader(header string) (*HeaderFields, error) {
	fields := &HeaderFields{}

	if header == "" {
		return fields, nil
	}

	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "MediaBrowser") {
		return fields, &ParseError{"Invalid authorization scheme"}
	}

	fieldStr := parts[1]
	pairs := strings.Split(fieldStr, ",")

	for _, pair := range pairs {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}

		eqIdx := strings.Index(pair, "=")
		if eqIdx < 0 {
			continue
		}

		key := strings.TrimSpace(pair[:eqIdx])
		value := strings.TrimSpace(pair[eqIdx+1:])

		// Remove surrounding quotes if present
		if len(value) >= 2 {
			if value[0] == '"' && value[len(value)-1] == '"' {
				value = value[1 : len(value)-1]
			}
		}

		switch strings.ToLower(key) {
		case "token":
			fields.Token = value
		case "deviceid":
			fields.DeviceID = value
		case "client":
			fields.Client = value
		case "version":
			fields.Version = value
		case "device":
			fields.Device = value
		case "platform":
			fields.Platform = value
		case "exts":
			fields.Exts = value
		case "sessionid":
			fields.SessionID = value
		}
	}

	if fields.Token == "" {
		return fields, &ParseError{"Token is required"}
	}

	return fields, nil
}

// ExtractFullHeader gets all header fields as a map
func ExtractFullHeader(h *HeaderFields) map[string]string {
	result := make(map[string]string)
	if h == nil {
		return result
	}

	if h.Token != "" {
		result["token"] = h.Token
	}
	if h.DeviceID != "" {
		result["deviceid"] = h.DeviceID
	}
	if h.Client != "" {
		result["client"] = h.Client
	}
	if h.Version != "" {
		result["version"] = h.Version
	}
	if h.Device != "" {
		result["device"] = h.Device
	}
	if h.Platform != "" {
		result["platform"] = h.Platform
	}

	return result
}

// HasToken checks if token field is present
func (h *HeaderFields) HasToken() bool {
	return h != nil && h.Token != ""
}

// Validate validates required fields
func (h *HeaderFields) Validate() error {
	if h.Token == "" {
		return &ParseError{"Token is required"}
	}
	return nil
}

// ExtractTokenFromHeader extracts token directly from raw header
func ExtractTokenFromHeader(header string) (string, error) {
	token, _, err := ParseEmbyAuthHeader(header)
	return token, err
}

// HasValidAuth checks if request has valid authorization header
func HasValidAuth(r *http.Request) bool {
	header := r.Header.Get("X-Emby-Authorization")
	if header == "" {
		// Also check Authorization header (alternative)
		auth := r.Header.Get("Authorization")
		if auth != "" {
			header = auth
		} else {
			return false
		}
	}

	h, err := ParseMediaBrowserHeader(header)
	return err == nil && h != nil && h.HasToken()
}

// GetAuthHeader extracts the authorization header from request
func GetAuthHeader(r *http.Request) string {
	header := r.Header.Get("X-Emby-Authorization")
	if header == "" {
		return r.Header.Get("Authorization")
	}
	return header
}

// ParseAuthFromRequest parses authorization header from HTTP request
func ParseAuthFromRequest(r *http.Request) (*HeaderFields, error) {
	header := GetAuthHeader(r)
	if header == "" {
		return nil, &ParseError{"No authorization header found"}
	}
	return ParseMediaBrowserHeader(header)
}