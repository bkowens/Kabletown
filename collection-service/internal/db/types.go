package db

import (
	"github.com/jellyfinhanced/shared/types"
)

// Collection represents a user-created collection (box set).
type Collection struct {
	ID          int              `json:"Id"           db:"id"`
	Name        string           `json:"Name"         db:"name"`
	UserID      string           `json:"UserId"       db:"user_id"`
	ImageTag    string           `json:"ImageTag,omitempty" db:"image_tag"`
	RowVersion  uint32           `json:"RowVersion"   db:"row_version"`
	DateCreated types.JellyfinTime `json:"DateCreated" db:"date_created"`
}

// CollectionItem represents an item within a collection.
type CollectionItem struct {
	ID             int    `json:"Id"              db:"id"`
	CollectionID   int    `json:"CollectionId"    db:"collection_id"`
	LibraryItemID  string `json:"LibraryItemId"   db:"library_item_id"`
	NextItemID     *int   `json:"NextItemId,omitempty"     db:"next_item_id"`
	PreviousItemID *int   `json:"PreviousItemId,omitempty" db:"previous_item_id"`
	RowVersion     uint32 `json:"RowVersion"      db:"row_version"`
}

// CollectionDetails extends Collection with metadata.
type CollectionDetails struct {
	Collection
	ItemCount    int    `json:"ItemCount"`
	Overview     string `json:"Overview,omitempty"`
	SortName     string `json:"SortName,omitempty"`
	DisplayOrder string `json:"DisplayOrder,omitempty"`
}
