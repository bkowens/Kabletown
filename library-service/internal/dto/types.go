package dto

import "time"

// UserDataDto represents a user's playback state for an item.
type UserDataDto struct {
	Rating                *float64   `json:"UserRating,omitempty"`
	Played                bool       `json:"Played"`
	PlaybackPositionTicks int64      `json:"PlaybackPositionTicks"`
	PlayCount             int        `json:"PlayCount"`
	IsFavorite            bool       `json:"IsFavorite"`
	LastPlayedDate        *time.Time `json:"LastPlayedDate,omitempty"`
	PlayedPercentage      *float64   `json:"PlayedPercentage,omitempty"`
	Key                   string     `json:"Key"`
	ItemId                string     `json:"ItemId"`
}

// BaseItemDto is a minimal representation of a media item.
type BaseItemDto struct {
	Id           string       `json:"Id"`
	Name         string       `json:"Name"`
	Type         string       `json:"Type"`
	IsFolder     bool         `json:"IsFolder"`
	ParentId     string       `json:"ParentId,omitempty"`
	TopParentId  string       `json:"TopParentId,omitempty"`
	SeriesId     string       `json:"SeriesId,omitempty"`
	SeriesName   string       `json:"SeriesName,omitempty"`
	IndexNumber  *int         `json:"IndexNumber,omitempty"`
	PremiereDate *time.Time   `json:"PremiereDate,omitempty"`
	UserData     *UserDataDto `json:"UserData,omitempty"`
}

// GenreDto represents a genre entry.
type GenreDto struct {
	Name string `json:"Name"`
}

// StudioDto represents a studio entry.
type StudioDto struct {
	Name string `json:"Name"`
}

// PersonDto represents a person associated with an item.
type PersonDto struct {
	Id   string `json:"Id"`
	Name string `json:"Name"`
}

// LibraryView represents a top-level library folder.
type LibraryView struct {
	Id   string `json:"Id"`
	Name string `json:"Name"`
	Type string `json:"CollectionType"`
}

// QueryResult is a paginated list of items.
type QueryResult[T any] struct {
	Items            []T   `json:"Items"`
	TotalRecordCount int64 `json:"TotalRecordCount"`
	StartIndex       int   `json:"StartIndex"`
}
