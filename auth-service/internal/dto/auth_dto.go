// Package dto defines request/response types for the auth-service HTTP API.
// All JSON field names are PascalCase to maintain Jellyfin wire compatibility.
package dto

// AuthenticateByNameRequest is the body for POST /Users/AuthenticateByName.
// Note: the field carrying the password is "Pw", not "Password".
type AuthenticateByNameRequest struct {
	Name string `json:"Name"`
	Pw   string `json:"Pw"`
}

// AuthenticationResult is the response body returned after a successful login.
type AuthenticationResult struct {
	User        *UserDto    `json:"User"`
	SessionInfo interface{} `json:"SessionInfo,omitempty"`
	AccessToken string      `json:"AccessToken"`
	ServerId    string      `json:"ServerId"`
}

// UserDto is the public representation of a Jellyfin user account.
type UserDto struct {
	Id                       string `json:"Id"`
	Name                     string `json:"Name"`
	ServerId                 string `json:"ServerId,omitempty"`
	PrimaryImageTag          string `json:"PrimaryImageTag,omitempty"`
	HasPassword              bool   `json:"HasPassword"`
	HasConfiguredPassword    bool   `json:"HasConfiguredPassword"`
	HasConfiguredEasyPassword bool  `json:"HasConfiguredEasyPassword"`
	EnableAutoLogin          bool   `json:"EnableAutoLogin"`
	IsAdministrator          bool   `json:"IsAdministrator"`
	IsDisabled               bool   `json:"IsDisabled"`
	IsHidden                 bool   `json:"IsHidden"`
}

// AuthenticationInfo describes a single API key or session token entry.
type AuthenticationInfo struct {
	Id          string `json:"Id"`
	AccessToken string `json:"AccessToken"`
	AppName     string `json:"AppName"`
	DateCreated string `json:"DateCreated"`
	IsActive    bool   `json:"IsActive"`
	UserId      string `json:"UserId,omitempty"`
	UserName    string `json:"UserName,omitempty"`
}
