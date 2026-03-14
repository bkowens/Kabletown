package db

import (
    "github.com/bowens/kabletown/shared/types"
)

// Playlist represents a user-created playlist
type Playlist struct {
    ID           int              `json:"Id"`
    Name         string           `json:"Name"`
    UserID       string           `json:"UserId"`
    ImageTag     string           `json:"ImageTag,omitempty"`
    RowVersion   uint32           `json:"RowVersion"`
    DateCreated  types.JellyfinTime `json:"DateCreated"`
}

// PlaylistItem represents an item within a playlist
type PlaylistItem struct {
    ID            int     `json:"Id"`
    PlaylistID    int     `json:"PlaylistId"`
    LibraryItemID string  `json:"LibraryItemId"`
    NextItemID    *int    `json:"NextItemId,omitempty"`
    PreviousItemID *int   `json:"PreviousItemId,omitempty"`
    RowVersion    uint32  `json:"RowVersion"`
}

// PlaylistDetails extends Playlist with metadata
type PlaylistDetails struct {
    Playlist
    ItemCount        int     `json:"ItemCount"`
    Overview         string  `json:"Overview,omitempty"`
    SortName         string  `json:"SortName,omitempty"`
    DisplayOrder     string  `json:"DisplayOrder,omitempty"` // DateAdded or Alphabetical
    ItemSortOrder    string  `json:"ItemSortOrder,omitempty"`
    PremiereYear     *int    `json:"PremiereYear,omitempty"`
    Duration         int64   `json:"Duration,omitempty"` // Total run time in ticks
}
