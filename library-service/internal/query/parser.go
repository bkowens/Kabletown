// InternalItemsQuery handles the complex content browsing query
// Supports 80+ URL parameters for filtering, sorting, and pagination
package query

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/jellyfinhanced/shared/types"
)

// InternalItemsQuery represents the parsed query parameters for content browsing
type InternalItemsQuery struct {
	// Pagination
	StartIndex int   `json:"start_index"`
	Limit      int   `json:"limit"`
	Skip       *int  `json:"skip,omitempty"` // Alternative to StartIndex

	// Hierarchy
	ParentId           string      `json:"parent_id"`
	UserId             uuid.UUID   `json:"user_id"` // Current user's GUID
	Recursive          bool        `json:"recursive"`
	CollapseSingleItems bool       `json:"collapse_single_items"`

	// Item Types
	IncludeItemTypes []string `json:"include_item_types"` // Comma-separated ItemTypes (Movie, Series, Episode, etc.)
	ExcludeItemTypes []string `json:"exclude_item_types"`

	// Filters
	Filters         []string `json:"filters"`          // Comma-separated special filters
	SearchQuery     string   `json:"search_query"`
	SearchHintLimit int      `json:"search_hint_limit"`

	// Sorting
	SortBy     []string `json:"sort_by"`     // Comma-separated fields
	SortOrder  []string `json:"sort_order"`  // Comma-separated (Ascending/Descending)

	// Item Value Filters (presence/absence)
	IsFavorite       *bool `json:"is_favorite"`
	IsPlayed         *bool `json:"is_played"`
	IsUnplayed       *bool `json:"is_unplayed"`      // Convenience: !IsPlayed
	HasSubtitles     *bool `json:"has_subtitles"`
	HasTrailer       *bool `json:"has_trailer"`
	HasSpecialFeature *bool `json:"has_special_feature"`
	HasThemeSong     *bool `json:"has_theme_song"`
	HasThemeVideo    *bool `json:"has_theme_video"`
	HasActor         *bool `json:"has_actor"`
	HasPrimary       *bool `json:"has_primary"`
	IsHD             *bool `json:"is_hd"`
	Is4K             *bool `json:"is_4k"`
	Is3D             *bool `json:"is_3d"`
	IsMissing        *bool `json:"is_missing"`
	IsResumable      *bool `json:"is_resumable"`     // Has resume point < 90%
	IsFavoriteOrResumable bool `json:"is_favorite_or_resumable"`
	IsInMixedFolder  *bool `json:"is_in_mixed_folder"`
	IsLocked         *bool `json:"is_locked"`
	IsPlaceHolder    *bool `json:"is_placeholder"`
	HasMissingLocators *bool `json:"has_missing_locators"`

	// Type-specific filters
	IsMovie          *bool `json:"is_movie"`
	IsSeries         *bool `json:"is_series"`
	IsNews           *bool `json:"is_news"`
	IsSports         *bool `json:"is_sports"`
	IsLiveTv         *bool `json:"is_live_tv"`
	IsFolder         *bool `json:"is_folder"`
	IsVirtualItem    *bool `json:"is_virtual_item"`

	// Metadata values (IN clause)
	Id               []string `json:"ids"`             // Filter by specific item IDs
	AncestorIds      []string `json:"ancestor_ids"`
	WithGenres       []string `json:"with_genres"`
	WithStudios      []string `json:"with_studios"`
	WithPeople       []string `json:"with_people"`     // Person names
	PeopleIds        []string `json:"people_ids"`      // Person GUIDs
	WithIds          []string `json:"with_ids"`
	AlbumIds         []string `json:"album_ids"`
	ArtistIds        []string `json:"artist_ids"`
	ArtistId         string   `json:"artist_id"`
	AlbumArtist      string   `json:"album_artist"`
	WithArtistIds    []string `json:"with_artist_ids"`

	// Year filter (range)
	MinProductionYear     *int    `json:"min_production_year"`
	MaxProductionYear     *int    `json:"max_production_year"`
	Years                 []int   `json:"years"`       // Specific years

	// Ratings
	ParentalRating     *int    `json:"parental_rating"`      // Exact match
	MaxParentalRating  *int    `json:"max_parental_rating"`  // Max rating threshold
	OfficialRatings    []string `json:"official_ratings"`    // Exact ratings (PG, R, etc.)

	// Container/Codecs
	Container          []string `json:"container"`
	VideoCodecs        []string `json:"video_codecs"`
	AudioCodecs        []string `json:"audio_codecs"`
	MediaTypes         []string `json:"media_types"`          // Video, Audio, Photo

	// Channel/Live TV
	ChannelId          string   `json:"channel_id"`
	StartDate          *types.JellyfinTime `json:"start_date"` // Live TV guide
	EndDate            *types.JellyfinTime `json:"end_date"`
	MinEndDate         *types.JellyfinTime `json:"min_end_date"`
	MaxEndDate         *types.JellyfinTime `json:"max_end_date"`

	// Date/Time filters
	MinDateLastSaved   *types.JellyfinTime `json:"min_date_last_saved"`
	MaxDateLastSaved   *types.JellyfinTime `json:"max_date_last_saved"`

	// Text search
	NameStartsWithOrContains string  `json:"name_startsWithOrContains"`
	SearchTerms             string  `json:"search_terms"`

	// Grouping
	GroupByAlbum  bool   `json:"group_by_album"`
	GroupBySeries bool   `json:"group_by_series"`

	// Virtual folders
	VirtualFolderIds []string `json:"virtual_folder_ids"`

	// Response customization
	EnableUserData    bool   `json:"enable_user_data"`  // Include user_data in response
	EnableImages      bool   `json:"enable_images"`
	ImageTypeLimit    int    `json:"image_type_limit"`
	Fields            []string `json:"fields"`          // Additional fields to request

	// Computed (not query params)
	TotalCount     int64      `json:"total_count"` // Set by COUNT query
	IsGrouped      bool       `json:"is_grouped"`  // Determined by GroupBy* fields
}

// Parse extracts and validates query parameters from the URL
func Parse(queryParams url.Values, userIdStr string) (*InternalItemsQuery, error) {
	q := &InternalItemsQuery{
		StartIndex: parseInt(queryParams.Get("StartIndex"), 0),
		Limit:      parseInt(queryParams.Get("Limit"), 100),
		Recursive:  queryParams.Get("Recursive") == "true",
	}

	// Parse UserId (from auth context or query param)
	if userIdStr != "" {
		id, err := types.ParseGUID(userIdStr)
		if err != nil {
			return nil, fmt.Errorf("invalid user ID: %w", err)
		}
		q.UserId = id
	}

	// Pagination alternative
	if skipStr := queryParams.Get("Skip"); skipStr != "" {
	skip := parseInt(skipStr, 0)
	q.Skip = &skip
	}

	// Hierarchy
	q.ParentId = queryParams.Get("ParentId")
	q.CollapseSingleItems = queryParams.Get("CollapseSingleItems") == "true"

	// Item types
	q.IncludeItemTypes = parseCommaSeparated(queryParams.Get("IncludeItemTypes"))
	q.ExcludeItemTypes = parseCommaSeparated(queryParams.Get("ExcludeItemTypes"))

	// Filters
	q.Filters = parseCommaSeparated(queryParams.Get("Filters"))
	q.SearchQuery = queryParams.Get("SearchQuery")
	q.SearchHintLimit = parseInt(queryParams.Get("SearchHintLimit"), 100)

	// Sorting
	q.SortBy = parseCommaSeparated(queryParams.Get("SortBy"))
	q.SortOrder = parseCommaSeparated(queryParams.Get("SortOrder"))

	// Boolean filters
	q.IsFavorite = parseBool(queryParams.Get("IsFavorite"))
	q.IsPlayed = parseBool(queryParams.Get("IsPlayed"))
	q.IsUnplayed = parseBool(queryParams.Get("IsUnplayed"))
	q.HasSubtitles = parseBool(queryParams.Get("HasSubtitles"))
	q.HasTrailer = parseBool(queryParams.Get("HasTrailer"))
	q.HasSpecialFeature = parseBool(queryParams.Get("HasSpecialFeature"))
	q.HasThemeSong = parseBool(queryParams.Get("HasThemeSong"))
	q.HasThemeVideo = parseBool(queryParams.Get("HasThemeVideo"))
	q.HasActor = parseBool(queryParams.Get("HasActor"))
	q.HasPrimary = parseBool(queryParams.Get("HasPrimary"))
	q.IsHD = parseBool(queryParams.Get("IsHD"))
	q.Is4K = parseBool(queryParams.Get("Is4K"))
	q.Is3D = parseBool(queryParams.Get("Is3D"))
	q.IsMissing = parseBool(queryParams.Get("IsMissing"))
	q.IsResumable = parseBool(queryParams.Get("IsResumable"))
	q.IsFavoriteOrResumable = queryParams.Get("IsFavoriteOrResumable") == "true"
	q.IsInMixedFolder = parseBool(queryParams.Get("IsInMixedFolder"))
	q.IsLocked = parseBool(queryParams.Get("IsLocked"))
	q.IsPlaceHolder = parseBool(queryParams.Get("IsPlaceHolder"))
	q.HasMissingLocators = parseBool(queryParams.Get("HasMissingLocators"))

	// Type-specific
	q.IsMovie = parseBool(queryParams.Get("IsMovie"))
	q.IsSeries = parseBool(queryParams.Get("IsSeries"))
	q.IsNews = parseBool(queryParams.Get("IsNews"))
	q.IsSports = parseBool(queryParams.Get("IsSports"))
	q.IsLiveTv = parseBool(queryParams.Get("IsLiveTv"))
	q.IsFolder = parseBool(queryParams.Get("IsFolder"))
	q.IsVirtualItem = parseBool(queryParams.Get("IsVirtualItem"))

	// ID arrays
	q.Id = parseCommaSeparated(queryParams.Get("Ids"))
	q.AncestorIds = parseCommaSeparated(queryParams.Get("AncestorIds"))
	q.WithGenres = parseCommaSeparated(queryParams.Get("WithGenres"))
	q.WithStudios = parseCommaSeparated(queryParams.Get("WithStudios"))
	q.WithPeople = parseCommaSeparated(queryParams.Get("WithPeople"))
	q.PeopleIds = parseCommaSeparated(queryParams.Get("PeopleIds"))
	q.WithIds = parseCommaSeparated(queryParams.Get("WithIds"))
	q.AlbumIds = parseCommaSeparated(queryParams.Get("AlbumIds"))
	q.ArtistIds = parseCommaSeparated(queryParams.Get("ArtistIds"))
	q.WithArtistIds = parseCommaSeparated(queryParams.Get("WithArtistIds"))

	// Single IDs
	q.ArtistId = queryParams.Get("ArtistId")
	q.AlbumArtist = queryParams.Get("AlbumArtist")
	q.ChannelId = queryParams.Get("ChannelId")

	// Year range
	q.MinProductionYear = parseIntPtr(queryParams.Get("MinProductionYear"))
	q.MaxProductionYear = parseIntPtr(queryParams.Get("MaxProductionYear"))
	q.Years = parseIntArray(queryParams.Get("Years"))

	// Ratings
	q.ParentalRating = parseIntPtr(queryParams.Get("ParentalRating"))
	q.MaxParentalRating = parseIntPtr(queryParams.Get("MaxParentalRating"))
	q.OfficialRatings = parseCommaSeparated(queryParams.Get("OfficialRatings"))

	// Media filters
	q.Container = parseCommaSeparated(queryParams.Get("Container"))
	q.VideoCodecs = parseCommaSeparated(queryParams.Get("VideoCodecs"))
	q.AudioCodecs = parseCommaSeparated(queryParams.Get("AudioCodecs"))
	q.MediaTypes = parseCommaSeparated(queryParams.Get("MediaTypes"))

	// Dates
	if minStart := queryParams.Get("MinStartDate"); minStart != "" {
		jt, err := types.ParseJellyfinTime(minStart)
		if err == nil {
			jtCopy := types.NewJellyfinTime(jt)
			q.StartDate = &jtCopy
		}
	}
	if maxEnd := queryParams.Get("MaxEndDate"); maxEnd != "" {
		jt, err := types.ParseJellyfinTime(maxEnd)
		if err == nil {
			jtCopy := types.NewJellyfinTime(jt)
			q.EndDate = &jtCopy
		}
	}
	if minEnd := queryParams.Get("MinEndDate"); minEnd != "" {
		jt, err := types.ParseJellyfinTime(minEnd)
		if err == nil {
			jtCopy := types.NewJellyfinTime(jt)
			q.MinEndDate = &jtCopy
		}
	}
	if maxEnD := queryParams.Get("MaxEndDate"); maxEnD != "" {
		jt, err := types.ParseJellyfinTime(maxEnD)
		if err == nil {
			jtCopy := types.NewJellyfinTime(jt)
			q.MaxEndDate = &jtCopy
		}
	}
	if minSave := queryParams.Get("MinDateLastSaved"); minSave != "" {
		jt, err := types.ParseJellyfinTime(minSave)
		if err == nil {
			jtCopy := types.NewJellyfinTime(jt)
			q.MinDateLastSaved = &jtCopy
		}
	}
	if maxSave := queryParams.Get("MaxDateLastSaved"); maxSave != "" {
		jt, err := types.ParseJellyfinTime(maxSave)
		if err == nil {
			jtCopy := types.NewJellyfinTime(jt)
			q.MaxDateLastSaved = &jtCopy
		}
	}

	// Text search
	q.NameStartsWithOrContains = queryParams.Get("NameStartsWithOrContains")
	q.SearchTerms = queryParams.Get("SearchTerms")

	// Grouping
	q.GroupByAlbum = queryParams.Get("GroupByAlbum") == "true"
	q.GroupBySeries = queryParams.Get("GroupBySeries") == "true"

	// Virtual folders
	q.VirtualFolderIds = parseCommaSeparated(queryParams.Get("VirtualFolderIds"))

	// Response options
	q.EnableUserData = queryParams.Get("EnableUserData") == "true"
	q.EnableImages = queryParams.Get("EnableImages") == "true"
	q.ImageTypeLimit = parseInt(queryParams.Get("ImageTypeLimit"), 0)
	q.Fields = parseCommaSeparated(queryParams.Get("Fields"))

	// Special flags
	q.IsGrouped = q.GroupByAlbum || q.GroupBySeries

	return q, nil
}

// Validate checks query parameters for consistency
func (q *InternalItemsQuery) Validate() error {
	// Pagination
	if q.StartIndex < 0 {
		return errors.New("StartIndex must be >= 0")
	}
	if q.Limit < 0 {
		return errors.New("Limit must be >= 0")
	}
	if q.Limit == 0 {
		q.Limit = 100 // Default limit
	}
	if q.Limit > 1000 {
		q.Limit = 1000 // Max limit
	}

	// Sorting
	if len(q.SortBy) == 0 && len(q.SortOrder) > 0 {
		q.SortBy = []string{"Name"} // Default sort
		if len(q.SortOrder) > 0 && strings.ToLower(q.SortOrder[0]) == "descending" {
			q.SortBy = []string{"Name"} // Reset for clarity
		}
	}

	// Conflicting filters
	if q.IsPlayed != nil && q.IsUnplayed != nil {
		if *q.IsPlayed != *q.IsUnplayed {
			return errors.New("IsPlayed and IsUnplayed cannot have different values")
		}
	}

	return nil
}

// GetIds returns UUID slice, filtering invalid GUIDs
func (q *InternalItemsQuery) GetIds() []uuid.UUID {
	var ids []uuid.UUID
	for _, idStr := range q.Id {
		if idStr == "" {
			continue
		}
		if id, err := types.ParseGUID(idStr); err == nil && id != uuid.Nil {
			ids = append(ids, id)
		}
	}
	return ids
}

// GetTopParentId returns the P7 index key (top parent from context or query)
func (q *InternalItemsQuery) GetTopParentId() string {
	if q.ParentId == "" {
		// Use UserId or default
		return ""
	}
	return q.ParentId
}

// UseUserFilter determines if user_data JOIN is needed
func (q *InternalItemsQuery) UseUserFilter() bool {
	return q.UserId != uuid.Nil && (
		q.IsFavorite != nil ||
		q.IsPlayed != nil ||
		q.IsResumable != nil ||
		q.EnableUserData ||
		!q.EnableUserData && false)
}

// parseBool converts "true"/"false" string to *bool
func parseBool(s string) *bool {
	if s == "" {
		return nil
	}
	val := s == "true"
	return &val
}

// parseInt converts string to int with default
func parseInt(s string, defaultVal int) int {
	if s == "" {
		return defaultVal
	}
	val, err := strconv.Atoi(s)
	if err != nil {
		return defaultVal
	}
	return val
}

// parseIntPtr converts string to *int (nil if empty/invalid)
func parseIntPtr(s string) *int {
	if s == "" {
		return nil
	}
	val, err := strconv.Atoi(s)
	if err != nil {
		return nil
	}
	return &val
}

// parseIntArray converts comma-separated string to []int
func parseIntArray(s string) []int {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	var result []int
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if val, err := strconv.Atoi(part); err == nil {
			result = append(result, val)
		}
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

// parseCommaSeparated splits a comma-separated string and trims whitespace
func parseCommaSeparated(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}