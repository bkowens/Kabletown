package db

import (
	"github.com/bowens/kabletown/shared/types"
)

// BaseItemDto represents a media item matching Jellyfin's model
type BaseItemDto struct {
	ID                string            `json:"Id"`                        // W2: GUID lowercase
	Name              string            `json:"Name"`                      // W1: PascalCase
	OriginalTitle     string            `json:"OriginalTitle,omitempty"`
	SortName          string            `json:"SortName"`
	Type              string            `json:"Type"`
	MediaType         string            `json:"MediaType,omitempty"`
	ParentID          string            `json:"ParentId,omitempty"`
	SeasonID          string            `json:"SeasonId,omitempty"`
	SeriesID          string            `json:"SeriesId,omitempty"`
	IsFolder          bool              `json:"IsFolder"`
	IsVirtualItem     bool              `json:"IsVirtualItem"`
	Path              string            `json:"Path,omitempty"`
	Overview          string            `json:"Overview,omitempty"`
	Tagline           string            `json:"Tagline,omitempty"`
	ProductionYear    int               `json:"ProductionYear,omitempty"`
	PremiereDate      *types.JellyfinTime `json:"PremiereDate,omitempty"` // W3: 7 decimal places
	CommunityRating   float64           `json:"CommunityRating,omitempty"`
	OfficialRating    string            `json:"OfficialRating,omitempty"`
	RunTimeTicks      int64             `json:"RunTimeTicks"`             // W4: 100-ns ticks
	CanDelete         bool              `json:"CanDelete"`
	CanDownload       bool              `json:"CanDownload"`
	LocationType      string            `json:"LocationType,omitempty"`
	ImageTags         string            `json:"ImageTags,omitempty"`      // W7: pipe-delimited string
	DateCreated       types.JellyfinTime   `json:"DateCreated"`           // W3: 7 decimal places
	DateLastSaved     types.JellyfinTime   `json:"DateLastSaved"`         // W3: 7 decimal places
	RowVersion        uint32            `json:"-"`
}
