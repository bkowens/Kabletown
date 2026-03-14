// Package dto contains shared data transfer objects used across Kabletown services.
package dto

// PersonDto represents a person (actor, director, writer, etc.) associated with an item.
type PersonDto struct {
	Name      string    `json:"Name"`
	Role      string    `json:"Role,omitempty"`
	Type      string    `json:"Type"`
	Id        string    `json:"Id,omitempty"`
	ImageTags ImageTags `json:"ImageTags,omitempty"`
	Index     *int      `json:"Index,omitempty"`
}

// ImageInfo represents a single image for a media item.
type ImageInfo struct {
	Type      string `json:"Type"`
	SourceUrl string `json:"SourceUrl,omitempty"`
	Url       string `json:"Url,omitempty"`
	Tag       string `json:"Tag,omitempty"`
	Width     int    `json:"Width,omitempty"`
	Height    int    `json:"Height,omitempty"`
	Language  string `json:"Language,omitempty"`
	Index     int    `json:"Index,omitempty"`
}

// BaseItemFields represents which fields should be included in item responses.
type BaseItemFields int

const (
	BaseItemFieldsNone                   BaseItemFields = 0
	BaseItemFieldsBasic                  BaseItemFields = 1 << 0
	BaseItemFieldsChildCount             BaseItemFields = 1 << 1
	BaseItemFieldsCanDelete              BaseItemFields = 1 << 2
	BaseItemFieldsPath                   BaseItemFields = 1 << 3
	BaseItemFieldsOverview               BaseItemFields = 1 << 4
	BaseItemFieldsOfficialRating         BaseItemFields = 1 << 5
	BaseItemFieldsRunTimeTicks           BaseItemFields = 1 << 6
	BaseItemFieldsParentLogoImageTag     BaseItemFields = 1 << 7
	BaseItemFieldsPrimaryImageAspectRatio BaseItemFields = 1 << 8
	BaseItemFieldsSeriesInfo             BaseItemFields = 1 << 9
	BaseItemFieldsTaglines               BaseItemFields = 1 << 10
	BaseItemFieldsGenres                 BaseItemFields = 1 << 11
	BaseItemFieldsStudios                BaseItemFields = 1 << 12
	BaseItemFieldsPremiereDate           BaseItemFields = 1 << 13
	BaseItemFieldsProductionYear         BaseItemFields = 1 << 14
)

// HasFlag checks if a specific field is set in BaseItemFields.
func (b BaseItemFields) HasFlag(flag BaseItemFields) bool {
	return b&flag != 0
}

// QueryResult wraps paginated results with metadata.
type QueryResult[T any] struct {
	Items            []T `json:"Items"`
	TotalRecordCount int `json:"TotalRecordCount"`
	StartIndex       int `json:"StartIndex"`
}

// NewQueryResult creates a paginated result wrapper.
func NewQueryResult[T any](items []T, totalCount, startIndex int) QueryResult[T] {
	if items == nil {
		items = []T{}
	}
	if totalCount < 0 {
		totalCount = 0
	}
	return QueryResult[T]{
		Items:            items,
		TotalRecordCount: totalCount,
		StartIndex:       startIndex,
	}
}
