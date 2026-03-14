package db

import (
	"github.com/jellyfinhanced/shared/types"
)

// Playlist represents a user-created playlist.
type Playlist struct {
	ID          int                `json:"Id"           db:"id"`
	Name        string             `json:"Name"         db:"name"`
	UserID      string             `json:"UserId"       db:"user_id"`
	ImageTag    string             `json:"ImageTag,omitempty" db:"image_tag"`
	RowVersion  uint32             `json:"RowVersion"   db:"row_version"`
	DateCreated types.JellyfinTime `json:"DateCreated"  db:"date_created"`
}

// PlaylistItem represents an item within a playlist.
type PlaylistItem struct {
	ID             int    `json:"Id"               db:"id"`
	PlaylistID     int    `json:"PlaylistId"       db:"playlist_id"`
	LibraryItemID  string `json:"LibraryItemId"    db:"library_item_id"`
	NextItemID     *int   `json:"NextItemId,omitempty"     db:"next_item_id"`
	PreviousItemID *int   `json:"PreviousItemId,omitempty" db:"previous_item_id"`
	RowVersion     uint32 `json:"RowVersion"       db:"row_version"`
}

// PlaylistDetails extends Playlist with metadata.
type PlaylistDetails struct {
	Playlist
	ItemCount    int    `json:"ItemCount"`
	Overview     string `json:"Overview,omitempty"`
	SortName     string `json:"SortName,omitempty"`
	DisplayOrder string `json:"DisplayOrder,omitempty"`
	Duration     int64  `json:"Duration,omitempty"`
}
