package db

import (
    "github.com/bowens/kabletown/shared/types"
)

// Collection represents a user-created collection (box set)
type Collection struct {
    ID           int              `json:"Id"`
    Name         string           `json:"Name"`
    UserID       string           `json:"UserId"`
    ImageTag     string           `json:"ImageTag,omitempty"`
    RowVersion   uint32           `json:"RowVersion"`
    DateCreated  types.JellyfinTime `json:"DateCreated"`
}

// CollectionItem represents an item within a collection
type CollectionItem struct {
    ID             int    `json:"Id"`
    CollectionID   int    `json:"CollectionId"`
    LibraryItemID  string `json:"LibraryItemId"`
    NextItemID     *int   `json:"NextItemId,omitempty"`
    PreviousItemID *int   `json:"PreviousItemId,omitempty"`
    RowVersion     uint32 `json:"RowVersion"`
}

// CollectionDetails extends Collection with metadata
type CollectionDetails struct {
    Collection
    ItemCount     int     `json:"ItemCount"`
    Overview      string  `json:"Overview,omitempty"`
    SortName      string  `json:"SortName,omitempty"`
    DisplayOrder  string  `json:"DisplayOrder,omitempty"` // DateAdded or Alphabetical
    ItemSortOrder string  `json:"ItemSortOrder,omitempty"`
    PremiereYear  *int    `json:"PremiereYear,omitempty"`
}
