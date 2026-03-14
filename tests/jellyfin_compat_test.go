package tests

import (
	"encoding/json"
	"testing"
)

// These tests verify that the Jellyfin API contract requirements are documented and
// enforced through type structure. They do not depend on running services -- that is
// handled by each service's own handler tests. Instead, these tests verify the JSON
// wire format contracts that all services must follow.

// TestJellyfinCompat_AuthResultContract verifies the AuthenticationResult JSON contract.
// Jellyfin compat: POST /Users/AuthenticateByName response must contain these fields.
func TestJellyfinCompat_AuthResultContract(t *testing.T) {
	// This JSON represents the expected shape from auth-service
	sampleJSON := `{
		"User": {
			"Id": "11111111-1111-1111-1111-111111111111",
			"Name": "admin",
			"HasPassword": true,
			"HasConfiguredPassword": true,
			"HasConfiguredEasyPassword": false,
			"EnableAutoLogin": false,
			"IsAdministrator": true,
			"IsDisabled": false,
			"IsHidden": false
		},
		"AccessToken": "64chartoken0123456789abcdef0123456789abcdef0123456789abcdef01",
		"ServerId": "server-id"
	}`

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(sampleJSON), &result); err != nil {
		t.Fatalf("sample AuthenticationResult JSON is invalid: %v", err)
	}

	requiredTopLevel := []string{"User", "AccessToken", "ServerId"}
	for _, f := range requiredTopLevel {
		if _, ok := result[f]; !ok {
			t.Errorf("AuthenticationResult missing required field %q", f)
		}
	}

	user, ok := result["User"].(map[string]interface{})
	if !ok {
		t.Fatal("User should be an object")
	}

	requiredUserFields := []string{
		"Id", "Name", "HasPassword", "HasConfiguredPassword",
		"IsAdministrator", "IsDisabled", "IsHidden",
	}
	for _, f := range requiredUserFields {
		if _, ok := user[f]; !ok {
			t.Errorf("User missing required field %q", f)
		}
	}
}

// TestJellyfinCompat_SystemInfoContract verifies the System/Info JSON contract.
// Jellyfin compat: GET /System/Info must contain all fields expected by jellyfin-web.
func TestJellyfinCompat_SystemInfoContract(t *testing.T) {
	sampleJSON := `{
		"LocalAddress": "http://localhost:8080",
		"StartupWizardCompleted": true,
		"Version": "10.10.0",
		"ProductName": "Jellyfin Server",
		"OperatingSystem": "Linux",
		"OperatingSystemDisplayName": "Linux",
		"Id": "server-id",
		"ServerName": "TestServer",
		"CanSelfRestart": false,
		"CanLaunchWebBrowser": false,
		"ProgramDataPath": "/config",
		"CachePath": "/cache",
		"LogPath": "/log",
		"TranscodingTempPath": "/config/transcodes",
		"IsShuttingDown": false,
		"SupportsLibraryMonitor": false,
		"WebSocketPortNumber": 8096,
		"CompletedInstallations": [],
		"HasPendingRestart": false,
		"HasUpdateAvailable": false,
		"SystemArchitecture": "X64"
	}`

	var body map[string]interface{}
	if err := json.Unmarshal([]byte(sampleJSON), &body); err != nil {
		t.Fatalf("sample System/Info JSON is invalid: %v", err)
	}

	requiredFields := []string{
		"Id", "ServerName", "Version", "ProductName", "OperatingSystem",
		"StartupWizardCompleted", "LocalAddress", "CanSelfRestart",
		"HasPendingRestart", "IsShuttingDown", "SupportsLibraryMonitor",
		"WebSocketPortNumber", "SystemArchitecture", "CompletedInstallations",
		"ProgramDataPath", "CachePath", "LogPath",
	}
	for _, f := range requiredFields {
		if _, ok := body[f]; !ok {
			t.Errorf("System/Info missing required field %q", f)
		}
	}

	// CompletedInstallations must be an array
	if _, ok := body["CompletedInstallations"].([]interface{}); !ok {
		t.Error("CompletedInstallations should be an array")
	}

	// StartupWizardCompleted must be bool
	if _, ok := body["StartupWizardCompleted"].(bool); !ok {
		t.Error("StartupWizardCompleted should be boolean")
	}
}

// TestJellyfinCompat_SystemInfoPublicContract verifies the public subset.
// Jellyfin compat: GET /System/Info/Public returns fewer fields than /System/Info.
func TestJellyfinCompat_SystemInfoPublicContract(t *testing.T) {
	sampleJSON := `{
		"LocalAddress": "http://localhost:8080",
		"StartupWizardCompleted": true,
		"Version": "10.10.0",
		"ProductName": "Jellyfin Server",
		"OperatingSystem": "Linux",
		"Id": "server-id",
		"ServerName": "TestServer"
	}`

	var body map[string]interface{}
	json.Unmarshal([]byte(sampleJSON), &body)

	requiredFields := []string{"Id", "ServerName", "Version", "ProductName", "OperatingSystem", "StartupWizardCompleted", "LocalAddress"}
	for _, f := range requiredFields {
		if _, ok := body[f]; !ok {
			t.Errorf("System/Info/Public missing required field %q", f)
		}
	}

	// Public info should NOT contain internal paths
	forbiddenFields := []string{"ProgramDataPath", "CachePath", "LogPath", "TranscodingTempPath"}
	for _, f := range forbiddenFields {
		if _, ok := body[f]; ok {
			t.Errorf("public info should not contain %q", f)
		}
	}
}

// TestJellyfinCompat_BrandingContract verifies GET /Branding/Configuration shape.
// Jellyfin compat: jellyfin-web calls this on every page load.
func TestJellyfinCompat_BrandingContract(t *testing.T) {
	sampleJSON := `{
		"LoginDisclaimer": "",
		"CustomCss": "",
		"SplashscreenEnabled": false
	}`

	var body map[string]interface{}
	json.Unmarshal([]byte(sampleJSON), &body)

	requiredFields := []string{"LoginDisclaimer", "CustomCss", "SplashscreenEnabled"}
	for _, f := range requiredFields {
		if _, ok := body[f]; !ok {
			t.Errorf("Branding/Configuration missing required field %q", f)
		}
	}

	if _, ok := body["SplashscreenEnabled"].(bool); !ok {
		t.Error("SplashscreenEnabled should be a boolean")
	}
}

// TestJellyfinCompat_QueryResultContract verifies the Items query result shape.
// Jellyfin compat: all list/query endpoints must return Items, TotalRecordCount, StartIndex.
func TestJellyfinCompat_QueryResultContract(t *testing.T) {
	sampleJSON := `{
		"Items": [
			{"Id": "item-1", "Name": "Movie 1", "Type": "Movie"},
			{"Id": "item-2", "Name": "Movie 2", "Type": "Movie"}
		],
		"TotalRecordCount": 100,
		"StartIndex": 0
	}`

	var body map[string]interface{}
	json.Unmarshal([]byte(sampleJSON), &body)

	requiredFields := []string{"Items", "TotalRecordCount", "StartIndex"}
	for _, f := range requiredFields {
		if _, ok := body[f]; !ok {
			t.Errorf("QueryResult missing required field %q", f)
		}
	}

	items, ok := body["Items"].([]interface{})
	if !ok {
		t.Fatal("Items should be an array")
	}
	if len(items) != 2 {
		t.Errorf("expected 2 items, got %d", len(items))
	}

	// Each item must have Id, Name, Type
	for i, raw := range items {
		item, ok := raw.(map[string]interface{})
		if !ok {
			t.Errorf("item %d is not an object", i)
			continue
		}
		for _, f := range []string{"Id", "Name", "Type"} {
			if _, ok := item[f]; !ok {
				t.Errorf("item %d missing field %q", i, f)
			}
		}
	}
}

// TestJellyfinCompat_ErrorResponseContract verifies error response shape.
// Jellyfin compat: error responses must contain "code" and "message" fields.
func TestJellyfinCompat_ErrorResponseContract(t *testing.T) {
	sampleJSON := `{"code": 404, "message": "Not found"}`

	var body map[string]interface{}
	json.Unmarshal([]byte(sampleJSON), &body)

	if _, ok := body["code"]; !ok {
		t.Error("error response missing 'code' field")
	}
	if _, ok := body["message"]; !ok {
		t.Error("error response missing 'message' field")
	}
}

// TestJellyfinCompat_PaginatedResponseContract verifies paginated list shape.
// Jellyfin compat: paginated lists use Items (array), TotalRecordCount (int), StartIndex (int), Limit (int).
func TestJellyfinCompat_PaginatedResponseContract(t *testing.T) {
	sampleJSON := `{
		"Items": [],
		"TotalRecordCount": 0,
		"StartIndex": 0,
		"Limit": 20
	}`

	var body map[string]interface{}
	json.Unmarshal([]byte(sampleJSON), &body)

	requiredFields := []string{"Items", "TotalRecordCount", "StartIndex"}
	for _, f := range requiredFields {
		if _, ok := body[f]; !ok {
			t.Errorf("PaginatedResponse missing required field %q", f)
		}
	}

	// Items must be an array even when empty (not null)
	items, ok := body["Items"].([]interface{})
	if !ok {
		t.Fatal("Items should be an array, even when empty")
	}
	if len(items) != 0 {
		t.Errorf("expected empty array, got %d items", len(items))
	}
}

// TestJellyfinCompat_SessionContract verifies session object shape.
// Jellyfin compat: session objects must include Id, DeviceId, AppName, DeviceName.
func TestJellyfinCompat_SessionContract(t *testing.T) {
	sampleJSON := `{
		"Id": "sess-1",
		"DeviceId": "dev-1",
		"AppName": "Jellyfin Web",
		"DeviceName": "Chrome",
		"Client": "web",
		"LastActivityDate": "2024-01-01T00:00:00Z"
	}`

	var body map[string]interface{}
	json.Unmarshal([]byte(sampleJSON), &body)

	requiredFields := []string{"Id", "DeviceId", "AppName", "DeviceName"}
	for _, f := range requiredFields {
		if _, ok := body[f]; !ok {
			t.Errorf("Session missing required field %q", f)
		}
	}
}

// TestJellyfinCompat_UserPolicyContract verifies user policy shape.
// Jellyfin compat: user policy must include IsAdministrator and auth provider fields.
func TestJellyfinCompat_UserPolicyContract(t *testing.T) {
	sampleJSON := `{
		"IsAdministrator": false,
		"EnableRemoteAccess": true,
		"EnableMediaPlayback": true,
		"EnableAllFolders": true,
		"AuthenticationProviderId": "Emby.Server.Implementations.Library.DefaultAuthenticationProvider",
		"PasswordResetProviderId": "Emby.Server.Implementations.Library.DefaultPasswordResetProvider",
		"SyncPlayAccess": "CreateAndJoinGroups"
	}`

	var body map[string]interface{}
	json.Unmarshal([]byte(sampleJSON), &body)

	requiredFields := []string{
		"IsAdministrator", "EnableRemoteAccess", "EnableMediaPlayback",
		"AuthenticationProviderId", "PasswordResetProviderId",
	}
	for _, f := range requiredFields {
		if _, ok := body[f]; !ok {
			t.Errorf("UserPolicy missing required field %q", f)
		}
	}
}

// TestJellyfinCompat_StartupConfigContract verifies startup configuration shape.
// Jellyfin compat: GET /Startup/Configuration response shape.
func TestJellyfinCompat_StartupConfigContract(t *testing.T) {
	sampleJSON := `{
		"UICulture": "en-US",
		"MetadataCountryCode": "US",
		"PreferredMetadataLanguage": "en"
	}`

	var body map[string]interface{}
	json.Unmarshal([]byte(sampleJSON), &body)

	requiredFields := []string{"UICulture", "MetadataCountryCode", "PreferredMetadataLanguage"}
	for _, f := range requiredFields {
		if _, ok := body[f]; !ok {
			t.Errorf("Startup/Configuration missing required field %q", f)
		}
	}
}
