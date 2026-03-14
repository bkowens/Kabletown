package models

import (
	"encoding/json"
	"time"
)

// Session represents a user device session
type Session struct {
	ID               string    `json:"Id"`
	UserID           string    `json:"UserId,omitempty"`
	DeviceID         string    `json:"DeviceId"`
	AppName          string    `json:"AppName"`
	DeviceName       string    `json:"DeviceName"`
	Client           string    `json:"Client"`
	LastActivityDate time.Time `json:"LastActivityDate"`
	Capabilities     string    `json:"Capabilities,omitempty"` // JSON stored as string
}

// CreateSessionRequest is the request body for creating a session
type CreateSessionRequest struct {
	DeviceID   string `json:"DeviceId"`
	AppName    string `json:"AppName"`
	DeviceName string `json:"DeviceName"`
	Client     string `json:"Client"`
}

// PlaybackState represents playback progress
type PlaybackState struct {
	ItemID            string    `json:"ItemId"`
	PlayPositionTicks int64     `json:"PlayPositionTicks"`
	IsPlaying         bool      `json:"IsPlaying"`
	QueueItemID       string    `json:"QueueItemId,omitempty"`
	PlaylistIndex     int       `json:"PlaylistIndex,omitempty"`
}

// SessionCapability contains session capability information
type SessionCapability struct {
	PlayableMediaTypes       []string `json:"PlayableMediaTypes"`
	SupportedCommands        []string `json:"SupportedCommands"`
	SupportsMediaControl     bool     `json:"SupportsMediaControl"`
	SupportsPersistentId     bool     `json:"SupportsPersistentId"`
}

func (c *SessionCapability) toJSON() string {
	data, _ := json.Marshal(c)
	return string(data)
}

// MessageRequest is a message to send to a session
type MessageRequest struct {
	MessageType string `json:"MessageType"`
	Header      string `json:"Header"`
	Text        string `json:"Text"`
	TimeoutMS   int    `json:"TimeoutMs"`
}

// Device represents a registered device
type Device struct {
	ID                   string    `json:"Id"`
	Name                 string    `json:"Name"`
	UserID               string    `json:"UserId,omitempty"`
	LastUserID           string    `json:"LastUserId,omitempty"`
	DateLastActivityed   time.Time `json:"DateLastActivityed"`
	AppName              string    `json:"AppName"`
	AppVersion           string    `json:"AppVersion"`
	DeviceType           string    `json:"DeviceType"`
	DeviceManufacturer   string    `json:"DeviceManufacturer"`
	DeviceModel          string    `json:"DeviceModel"`
	DeviceID             string    `json:"DeviceId"`
	LastActivityDate     time.Time `json:"LastActivityDate"`
}

// RegisterDeviceRequest is the request body for device registration
type RegisterDeviceRequest struct {
	DeviceID           string `json:"DeviceId"`
	Name               string `json:"Name"`
	AppName            string `json:"AppName"`
	AppVersion         string `json:"AppVersion"`
	DeviceType         string `json:"DeviceType"`
	DeviceManufacturer string `json:"DeviceManufacturer"`
	DeviceModel        string `json:"DeviceModel"`
}

// UpdateDeviceNameRequest is the request body for updating device name
type UpdateDeviceNameRequest struct {
	Name string `json:"Name"`
}
