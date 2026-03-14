package auth

import (
	"net/http"
	"testing"
)

// TestParseEmbyAuthHeader tests the basic token/deviceID extraction function
func TestParseEmbyAuthHeader(t *testing.T) {
	tests := []struct {
		name          string
		header        string
		wantToken     string
		wantDeviceID  string
		wantErr       bool
		errContains   string
	}{
		{
			name:         "valid header with token and device ID",
			header:       `MediaBrowser Token="abc123", DeviceId="device456"`,
			wantToken:    "abc123",
			wantDeviceID: "device456",
			wantErr:      false,
		},
		{
			name:         "valid header with only token",
			header:       `MediaBrowser Token="xyz789"`,
			wantToken:    "xyz789",
			wantDeviceID: "",
			wantErr:      false,
		},
		{
			name:        "empty header",
			header:      "",
			wantToken:   "",
			wantErr:     true,
			errContains: "empty",
		},
		{
			name:        "missing MediaBrowser scheme",
			header:      `Bearer Token="abc123"`,
			wantToken:   "",
			wantErr:     true,
			errContains: "Invalid authorization scheme",
		},
		{
			name:        "token missing",
			header:      `MediaBrowser DeviceId="device456"`,
			wantToken:   "",
			wantErr:     true,
			errContains: "Token is required",
		},
		{
			name:         "token with spaces",
			header:       `MediaBrowser Token = "spaced token" , DeviceId = "spaced device"`,
			wantToken:    "spaced token",
			wantDeviceID: "spaced device",
			wantErr:      false,
		},
		{
			name:         "token without quotes",
			header:       `MediaBrowser Token=unquoted_token`,
			wantToken:    "unquoted_token",
			wantDeviceID: "",
			wantErr:      false,
		},
		{
			name:    "malformed - no equals sign",
			header:  `MediaBrowser Token", DeviceId="abc"`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, deviceID, err := ParseEmbyAuthHeader(tt.header)

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseEmbyAuthHeader() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errContains != "" {
				if e, ok := err.(*ParseError); ok {
					if !contains(e.Message, tt.errContains) {
						t.Errorf("ParseEmbyAuthHeader() error message = %v, want to contain %v", e.Message, tt.errContains)
					}
				}
			}

			if !tt.wantErr {
				if token != tt.wantToken {
					t.Errorf("ParseEmbyAuthHeader() token = %v, want %v", token, tt.wantToken)
				}
				if deviceID != tt.wantDeviceID {
					t.Errorf("ParseEmbyAuthHeader() deviceID = %v, want %v", deviceID, tt.wantDeviceID)
				}
			}
		})
	}
}

// TestParseMediaBrowserHeader tests the full header parsing function
func TestParseMediaBrowserHeader(t *testing.T) {
	tests := []struct {
		name       string
		header     string
		wantToken  string
		wantClient string
		wantErr    bool
	}{
		{
			name:       "complete header with all fields",
			header:     `MediaBrowser Token="tok123", DeviceId="dev456", Client="JellyfinWeb", Version="10.8.0", Device="My Device", Platform="Linux"`,
			wantToken:  "tok123",
			wantClient: "JellyfinWeb",
			wantErr:    false,
		},
		{
			name:       "minimal header with only token",
			header:     `MediaBrowser Token="minimal"`,
			wantToken:  "minimal",
			wantClient: "",
			wantErr:    false,
		},
		{
			name:    "no token",
			header:  `MediaBrowser Client="JellyfinWeb"`,
			wantErr: true,
		},
		{
			name:    "empty header",
			header:  "",
			wantErr: false, // Returns empty struct with no error, per implementation
		},
		{
			name:       "header with session ID and extensions",
			header:     `MediaBrowser Token="sess123", SessionId="session456", Exts="custom"`,
			wantToken:  "sess123",
			wantClient: "",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fields, err := ParseMediaBrowserHeader(tt.header)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseMediaBrowserHeader() expected error, got nil")
					return
				}
				return
			}

			if err != nil {
				t.Errorf("ParseMediaBrowserHeader() unexpected error = %v", err)
				return
			}

			if fields == nil {
				t.Fatal("ParseMediaBrowserHeader() returned nil fields")
			}

			if fields.Token != tt.wantToken {
				t.Errorf("ParseMediaBrowserHeader() Token = %v, want %v", fields.Token, tt.wantToken)
			}

			if fields.Client != tt.wantClient {
				t.Errorf("ParseMediaBrowserHeader() Client = %v, want %v", fields.Client, tt.wantClient)
			}
		})
	}
}

// TestExtractFullHeader tests the header extraction to map
func TestExtractFullHeader(t *testing.T) {
	tests := []struct {
		name     string
		headers  *HeaderFields
		wantKeys []string
	}{
		{
			name: "complete header fields",
			headers: &HeaderFields{
				Token:    "token123",
				DeviceID: "device456",
				Client:   "JellyfinWeb",
				Version:  "10.8.0",
			},
			wantKeys: []string{"token", "deviceid", "client", "version"},
		},
		{
			name:     "nil header",
			headers:  nil,
			wantKeys: []string{},
		},
		{
			name:     "empty header",
			headers:  &HeaderFields{},
			wantKeys: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractFullHeader(tt.headers)

			if len(result) != len(tt.wantKeys) {
				t.Errorf("ExtractFullHeader() returned %d keys, want %d", len(result), len(tt.wantKeys))
			}

			for _, key := range tt.wantKeys {
				if _, ok := result[key]; !ok {
					t.Errorf("ExtractFullHeader() missing key: %s", key)
				}
			}
		})
	}
}

// TestHeaderFieldsMethods tests the HeaderFields methods
func TestHeaderFieldsMethods(t *testing.T) {
	t.Run("HasToken returns true with token", func(t *testing.T) {
		h := &HeaderFields{Token: "abc123"}
		if !h.HasToken() {
			t.Error("HasToken() should return true when token is present")
		}
	})

	t.Run("HasToken returns false without token", func(t *testing.T) {
		h := &HeaderFields{}
		if h.HasToken() {
			t.Error("HasToken() should return false when token is empty")
		}
	})

	t.Run("Validate passes with token", func(t *testing.T) {
		h := &HeaderFields{Token: "abc123"}
		if err := h.Validate(); err != nil {
			t.Errorf("Validate() returned error = %v", err)
		}
	})

	t.Run("Validate fails without token", func(t *testing.T) {
		h := &HeaderFields{}
		if err := h.Validate(); err == nil {
			t.Error("Validate() should return error when token is missing")
		}
	})
}

// TestExtractTokenFromHeader tests token extraction
func TestExtractTokenFromHeader(t *testing.T) {
	token, err := ExtractTokenFromHeader(`MediaBrowser Token="extract123"`)
	if err != nil {
		t.Errorf("ExtractTokenFromHeader() unexpected error = %v", err)
	}
	if token != "extract123" {
		t.Errorf("ExtractTokenFromHeader() token = %v, want extract123", token)
	}

	_, err = ExtractTokenFromHeader("invalid")
	if err == nil {
		t.Error("ExtractTokenFromHeader() should return error for invalid header")
	}
}

// TestHasValidAuth tests the request validation helper
func TestHasValidAuth(t *testing.T) {
	t.Run("valid X-Emby-Authorization header", func(t *testing.T) {
		r, _ := http.NewRequest("GET", "/", nil)
		r.Header.Set("X-Emby-Authorization", `MediaBrowser Token="valid"`)
		if !HasValidAuth(r) {
			t.Error("HasValidAuth() should return true for valid header")
		}
	})

	t.Run("valid Authorization header (fallback)", func(t *testing.T) {
		r, _ := http.NewRequest("GET", "/", nil)
		r.Header.Set("Authorization", `MediaBrowser Token="fallback"`)
		if !HasValidAuth(r) {
			t.Error("HasValidAuth() should return true for valid Authorization header")
		}
	})

	t.Run("empty headers", func(t *testing.T) {
		r, _ := http.NewRequest("GET", "/", nil)
		if HasValidAuth(r) {
			t.Error("HasValidAuth() should return false for empty headers")
		}
	})

	t.Run("invalid token", func(t *testing.T) {
		r, _ := http.NewRequest("GET", "/", nil)
		r.Header.Set("X-Emby-Authorization", `MediaBrowser DeviceId="nodata"`)
		if HasValidAuth(r) {
			t.Error("HasValidAuth() should return false for missing token")
		}
	})
}

// TestGetAuthHeader tests header extraction
func TestGetAuthHeader(t *testing.T) {
	t.Run("prefers X-Emby-Authorization", func(t *testing.T) {
		r, _ := http.NewRequest("GET", "/", nil)
		r.Header.Set("X-Emby-Authorization", `MediaBrowser Token="emby"`)
		r.Header.Set("Authorization", `MediaBrowser Token="auth"`)

		header := GetAuthHeader(r)
		if !contains(header, "emby") {
			t.Errorf("GetAuthHeader() should prefer X-Emby-Authorization, got: %s", header)
		}
	})

	t.Run("falls back to Authorization", func(t *testing.T) {
		r, _ := http.NewRequest("GET", "/", nil)
		r.Header.Set("Authorization", `MediaBrowser Token="fallback"`)

		header := GetAuthHeader(r)
		if !contains(header, "fallback") {
			t.Errorf("GetAuthHeader() should fallback to Authorization, got: %s", header)
		}
	})
}

// TestParseAuthFromRequest tests request parsing
func TestParseAuthFromRequest(t *testing.T) {
	t.Run("valid request with X-Emby-Authorization", func(t *testing.T) {
		r, _ := http.NewRequest("GET", "/", nil)
		r.Header.Set("X-Emby-Authorization", `MediaBrowser Token="test123", DeviceId="dev456"`)

		fields, err := ParseAuthFromRequest(r)
		if err != nil {
			t.Errorf("ParseAuthFromRequest() unexpected error = %v", err)
		}
		if fields == nil || fields.Token != "test123" {
			t.Errorf("ParseAuthFromRequest() fields = %v", fields)
		}
	})

	t.Run("invalid request without headers", func(t *testing.T) {
		r, _ := http.NewRequest("GET", "/", nil)

		_, err := ParseAuthFromRequest(r)
		if err == nil {
			t.Error("ParseAuthFromRequest() should return error for missing headers")
		}
	})
}

// TestParseMediaBrowserHeaderCaseInsensitiveKey tests case-insensitive key handling
func TestParseMediaBrowserHeaderCaseInsensitiveKey(t *testing.T) {
	header := `MediaBrowser TOKEN="upper", DEVICEID="mixed", Client="lower"`
	fields, err := ParseMediaBrowserHeader(header)

	if err != nil {
		t.Errorf("ParseMediaBrowserHeader() unexpected error = %v", err)
	}

	if fields.Token != "upper" {
		t.Errorf("Token parsing case-sensitive keys = %v, want upper", fields.Token)
	}
}

// contains is a helper for substring checks
func contains(s, substr string) bool {
	return s != "" && len(s) >= len(substr) && (s == substr || len(s) > len(substr) && 
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || 
		 findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
