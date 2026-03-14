package tests

import (
	"bufio"
	"os"
	"regexp"
	"strings"
	"testing"
)

// nginxConfPath is the path to the nginx configuration file under test.
const nginxConfPath = "../nginx/nginx.conf"

// TestNginxConf_FileExists verifies the nginx configuration file is present.
func TestNginxConf_FileExists(t *testing.T) {
	if _, err := os.Stat(nginxConfPath); os.IsNotExist(err) {
		t.Fatal("nginx.conf not found at expected path")
	}
}

// TestNginxConf_RequiredUpstreams verifies all expected upstream blocks exist.
// Jellyfin compat: all microservices must have corresponding upstream definitions.
func TestNginxConf_RequiredUpstreams(t *testing.T) {
	content, err := os.ReadFile(nginxConfPath)
	if err != nil {
		t.Fatalf("failed to read nginx.conf: %v", err)
	}
	conf := string(content)

	requiredUpstreams := []string{
		"auth_service",
		"user_service",
		"library_service",
		"session_service",
		"system_service",
		"media_service",
		"stream_service",
		"metadata_service",
		"content_service",
		"transcode_service",
		"collection_service",
		"playlist_service",
		"notification_service",
		"plugin_service",
		"search_service",
		"playstate_service",
	}

	for _, name := range requiredUpstreams {
		pattern := `upstream\s+` + name + `\s*\{`
		if matched, _ := regexp.MatchString(pattern, conf); !matched {
			t.Errorf("missing upstream block for %q", name)
		}
	}
}

// TestNginxConf_CriticalRoutes verifies all Jellyfin API routes are mapped.
// Jellyfin compat: web frontend calls these paths and expects them to be routed.
func TestNginxConf_CriticalRoutes(t *testing.T) {
	content, err := os.ReadFile(nginxConfPath)
	if err != nil {
		t.Fatalf("failed to read nginx.conf: %v", err)
	}
	conf := string(content)

	routes := []struct {
		path    string
		backend string
	}{
		// Auth service
		{"/Auth", "auth_service"},
		{"/Startup", "auth_service"},
		{"/QuickConnect", "auth_service"},
		{"/Users/AuthenticateByName", "auth_service"},
		{"/Users/ForgotPassword", "auth_service"},

		// User service
		{"/Users", "user_service"},
		{"/DisplayPreferences", "user_service"},

		// Library service
		{"/Items", "library_service"},
		{"/Library", "library_service"},
		{"/Genres", "library_service"},
		{"/Studios", "library_service"},
		{"/Persons", "library_service"},
		{"/Years", "library_service"},
		{"/Artists", "library_service"},

		// Session service
		{"/Sessions", "session_service"},
		{"/Devices", "session_service"},

		// System service
		{"/System", "system_service"},
		{"/Branding", "system_service"},
		{"/Environment", "system_service"},
		{"/Localization", "system_service"},

		// Media service
		{"/Images", "media_service"},
		{"/MediaInfo", "media_service"},

		// Stream service
		{"/Videos", "stream_service"},
		{"/Audio", "stream_service"},

		// Content service
		{"/Movies", "content_service"},
		{"/TvShows", "content_service"},

		// Playstate service
		{"/Playstate", "playstate_service"},
	}

	for _, rt := range routes {
		t.Run(rt.path+" -> "+rt.backend, func(t *testing.T) {
			// Check that location block exists and points to correct upstream
			locationPattern := `location\s+` + regexp.QuoteMeta(rt.path) + `\s*\{.*proxy_pass\s+http://` + rt.backend
			if matched, _ := regexp.MatchString(locationPattern, conf); !matched {
				// Try multiline
				if !strings.Contains(conf, "location "+rt.path) {
					t.Errorf("missing location block for path %q", rt.path)
				} else if !containsRouteToBackend(conf, rt.path, rt.backend) {
					t.Errorf("path %q should route to %q", rt.path, rt.backend)
				}
			}
		})
	}
}

// containsRouteToBackend checks if a location block for the given path routes to the backend.
func containsRouteToBackend(conf, path, backend string) bool {
	scanner := bufio.NewScanner(strings.NewReader(conf))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "location "+path) && strings.Contains(line, "proxy_pass http://"+backend) {
			return true
		}
	}
	return false
}

// TestNginxConf_HealthEndpoint verifies /health returns 200.
// Jellyfin compat: health check endpoint for container orchestration.
func TestNginxConf_HealthEndpoint(t *testing.T) {
	content, err := os.ReadFile(nginxConfPath)
	if err != nil {
		t.Fatalf("failed to read nginx.conf: %v", err)
	}
	conf := string(content)

	if !strings.Contains(conf, "location /health") {
		t.Error("missing /health endpoint")
	}
	if !strings.Contains(conf, "return 200") {
		t.Error("/health should return 200")
	}
}

// TestNginxConf_SecurityHeaders verifies security headers are set.
// Jellyfin compat: X-Content-Type-Options and X-Frame-Options should be set.
func TestNginxConf_SecurityHeaders(t *testing.T) {
	content, err := os.ReadFile(nginxConfPath)
	if err != nil {
		t.Fatalf("failed to read nginx.conf: %v", err)
	}
	conf := string(content)

	if !strings.Contains(conf, "X-Content-Type-Options") {
		t.Error("missing X-Content-Type-Options header")
	}
	if !strings.Contains(conf, "X-Frame-Options") {
		t.Error("missing X-Frame-Options header")
	}
}

// TestNginxConf_ListensOnPort8080 verifies the server listens on the expected port.
// Jellyfin compat: the gateway must listen on port 8080 for docker-compose compatibility.
func TestNginxConf_ListensOnPort8080(t *testing.T) {
	content, err := os.ReadFile(nginxConfPath)
	if err != nil {
		t.Fatalf("failed to read nginx.conf: %v", err)
	}
	if !strings.Contains(string(content), "listen 8080") {
		t.Error("nginx should listen on port 8080")
	}
}

// TestNginxConf_DefaultReturn404 verifies the default location returns 404.
// Jellyfin compat: unmatched paths should return 404 JSON error.
func TestNginxConf_DefaultReturn404(t *testing.T) {
	content, err := os.ReadFile(nginxConfPath)
	if err != nil {
		t.Fatalf("failed to read nginx.conf: %v", err)
	}
	conf := string(content)

	if !strings.Contains(conf, `location / {`) || !strings.Contains(conf, "return 404") {
		t.Error("default location should return 404")
	}
}

// TestNginxConf_StreamingRouteNoBuffering verifies streaming routes disable buffering.
// Jellyfin compat: video/audio streaming must not buffer for real-time playback.
func TestNginxConf_StreamingRouteNoBuffering(t *testing.T) {
	content, err := os.ReadFile(nginxConfPath)
	if err != nil {
		t.Fatalf("failed to read nginx.conf: %v", err)
	}
	conf := string(content)

	streamPaths := []string{"/Videos", "/Audio", "/Hls"}
	for _, path := range streamPaths {
		t.Run(path, func(t *testing.T) {
			scanner := bufio.NewScanner(strings.NewReader(conf))
			for scanner.Scan() {
				line := scanner.Text()
				if strings.Contains(line, "location "+path) {
					if !strings.Contains(line, "proxy_buffering off") {
						t.Errorf("streaming route %s should have proxy_buffering off", path)
					}
					break
				}
			}
		})
	}
}

// TestNginxConf_AuthRoutePrecedence verifies /Users/AuthenticateByName routes to auth_service.
// Jellyfin compat: longest-prefix matching must route auth sub-paths to auth-service, not user-service.
func TestNginxConf_AuthRoutePrecedence(t *testing.T) {
	content, err := os.ReadFile(nginxConfPath)
	if err != nil {
		t.Fatalf("failed to read nginx.conf: %v", err)
	}
	conf := string(content)

	// /Users/AuthenticateByName must go to auth_service, not user_service
	if !containsRouteToBackend(conf, "/Users/AuthenticateByName", "auth_service") {
		t.Error("/Users/AuthenticateByName should route to auth_service")
	}
	// /Users/ForgotPassword must go to auth_service
	if !containsRouteToBackend(conf, "/Users/ForgotPassword", "auth_service") {
		t.Error("/Users/ForgotPassword should route to auth_service")
	}
}

// TestNginxConf_ProxyHeadersSet verifies proxy_set_header directives are present.
// Jellyfin compat: Host and X-Real-IP headers must be forwarded to services.
func TestNginxConf_ProxyHeadersSet(t *testing.T) {
	content, err := os.ReadFile(nginxConfPath)
	if err != nil {
		t.Fatalf("failed to read nginx.conf: %v", err)
	}
	conf := string(content)

	if !strings.Contains(conf, "proxy_set_header Host $host") {
		t.Error("missing proxy_set_header Host")
	}
	if !strings.Contains(conf, "proxy_set_header X-Real-IP $remote_addr") {
		t.Error("missing proxy_set_header X-Real-IP")
	}
}
