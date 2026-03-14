package query

import (
	"net/url"
	"testing"

	"github.com/google/uuid"
)

// TestParse_Defaults verifies default values when no params are provided.
// Jellyfin compat: default pagination must produce valid SQL.
func TestParse_Defaults(t *testing.T) {
	uid := uuid.New().String()
	q, err := Parse(url.Values{}, uid)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if q.StartIndex != 0 {
		t.Errorf("StartIndex = %d, want 0", q.StartIndex)
	}
	if q.Limit != 100 {
		t.Errorf("Limit = %d, want 100", q.Limit)
	}
	if q.Recursive {
		t.Error("Recursive should default to false")
	}
	if q.UserId.String() != uid {
		t.Errorf("UserId = %q, want %q", q.UserId.String(), uid)
	}
}

// TestParse_Pagination verifies StartIndex and Limit parsing.
// Jellyfin compat: jellyfin-web sends StartIndex and Limit for infinite scroll.
func TestParse_Pagination(t *testing.T) {
	params := url.Values{
		"StartIndex": {"10"},
		"Limit":      {"25"},
	}
	q, err := Parse(params, uuid.New().String())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if q.StartIndex != 10 {
		t.Errorf("StartIndex = %d, want 10", q.StartIndex)
	}
	if q.Limit != 25 {
		t.Errorf("Limit = %d, want 25", q.Limit)
	}
}

// TestParse_InvalidPagination verifies invalid numeric values use defaults.
// Jellyfin compat: malformed query params should not crash.
func TestParse_InvalidPagination(t *testing.T) {
	params := url.Values{
		"StartIndex": {"invalid"},
		"Limit":      {"abc"},
	}
	q, err := Parse(params, uuid.New().String())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if q.StartIndex != 0 {
		t.Errorf("StartIndex = %d, want 0 for invalid input", q.StartIndex)
	}
	if q.Limit != 100 {
		t.Errorf("Limit = %d, want 100 for invalid input", q.Limit)
	}
}

// TestParse_ParentId verifies parent ID filtering.
// Jellyfin compat: browsing a library folder sends ParentId.
func TestParse_ParentId(t *testing.T) {
	parentID := uuid.New().String()
	params := url.Values{
		"ParentId":  {parentID},
		"Recursive": {"true"},
	}
	q, err := Parse(params, uuid.New().String())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if q.ParentId != parentID {
		t.Errorf("ParentId = %q, want %q", q.ParentId, parentID)
	}
	if !q.Recursive {
		t.Error("Recursive should be true")
	}
}

// TestParse_IncludeItemTypes verifies comma-separated item type parsing.
// Jellyfin compat: browsing movies page sends IncludeItemTypes=Movie.
func TestParse_IncludeItemTypes(t *testing.T) {
	params := url.Values{
		"IncludeItemTypes": {"Movie,Series,Episode"},
	}
	q, err := Parse(params, uuid.New().String())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(q.IncludeItemTypes) != 3 {
		t.Fatalf("IncludeItemTypes length = %d, want 3", len(q.IncludeItemTypes))
	}
	if q.IncludeItemTypes[0] != "Movie" {
		t.Errorf("IncludeItemTypes[0] = %q, want Movie", q.IncludeItemTypes[0])
	}
	if q.IncludeItemTypes[1] != "Series" {
		t.Errorf("IncludeItemTypes[1] = %q, want Series", q.IncludeItemTypes[1])
	}
}

// TestParse_SortBy verifies sort field parsing.
// Jellyfin compat: library views support multiple sort fields.
func TestParse_SortBy(t *testing.T) {
	params := url.Values{
		"SortBy":    {"SortName,DateAdded"},
		"SortOrder": {"Ascending,Descending"},
	}
	q, err := Parse(params, uuid.New().String())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(q.SortBy) != 2 {
		t.Fatalf("SortBy length = %d, want 2", len(q.SortBy))
	}
	if q.SortBy[0] != "SortName" {
		t.Errorf("SortBy[0] = %q, want SortName", q.SortBy[0])
	}
	if len(q.SortOrder) != 2 || q.SortOrder[1] != "Descending" {
		t.Errorf("SortOrder = %v, want [Ascending Descending]", q.SortOrder)
	}
}

// TestParse_BooleanFilters verifies boolean filter parsing.
// Jellyfin compat: filter bar sends IsFavorite=true, IsPlayed=false etc.
func TestParse_BooleanFilters(t *testing.T) {
	tests := []struct {
		param string
		value string
		check func(*InternalItemsQuery) *bool
	}{
		{"IsFavorite", "true", func(q *InternalItemsQuery) *bool { return q.IsFavorite }},
		{"IsPlayed", "false", func(q *InternalItemsQuery) *bool { return q.IsPlayed }},
		{"IsUnplayed", "true", func(q *InternalItemsQuery) *bool { return q.IsUnplayed }},
		{"HasSubtitles", "true", func(q *InternalItemsQuery) *bool { return q.HasSubtitles }},
		{"IsMissing", "false", func(q *InternalItemsQuery) *bool { return q.IsMissing }},
		{"IsHD", "true", func(q *InternalItemsQuery) *bool { return q.IsHD }},
		{"Is4K", "true", func(q *InternalItemsQuery) *bool { return q.Is4K }},
	}
	for _, tt := range tests {
		t.Run(tt.param+"="+tt.value, func(t *testing.T) {
			params := url.Values{tt.param: {tt.value}}
			q, err := Parse(params, uuid.New().String())
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			val := tt.check(q)
			if val == nil {
				t.Fatalf("%s should not be nil", tt.param)
			}
			expected := tt.value == "true"
			if *val != expected {
				t.Errorf("%s = %v, want %v", tt.param, *val, expected)
			}
		})
	}
}

// TestParse_BooleanFilters_Empty verifies unset boolean filters are nil.
// Jellyfin compat: omitted filters should not restrict results.
func TestParse_BooleanFilters_Empty(t *testing.T) {
	q, err := Parse(url.Values{}, uuid.New().String())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if q.IsFavorite != nil {
		t.Error("IsFavorite should be nil when not set")
	}
	if q.IsPlayed != nil {
		t.Error("IsPlayed should be nil when not set")
	}
}

// TestParse_YearFilters verifies year range and specific year parsing.
// Jellyfin compat: year filter in the web UI.
func TestParse_YearFilters(t *testing.T) {
	params := url.Values{
		"MinProductionYear": {"2020"},
		"MaxProductionYear": {"2024"},
		"Years":             {"2021,2022,2023"},
	}
	q, err := Parse(params, uuid.New().String())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if q.MinProductionYear == nil || *q.MinProductionYear != 2020 {
		t.Errorf("MinProductionYear = %v, want 2020", q.MinProductionYear)
	}
	if q.MaxProductionYear == nil || *q.MaxProductionYear != 2024 {
		t.Errorf("MaxProductionYear = %v, want 2024", q.MaxProductionYear)
	}
	if len(q.Years) != 3 {
		t.Errorf("Years length = %d, want 3", len(q.Years))
	}
}

// TestParse_GenresAndStudios verifies genre and studio filter parsing.
// Jellyfin compat: genre/studio sidebar filters.
func TestParse_GenresAndStudios(t *testing.T) {
	params := url.Values{
		"WithGenres":  {"Action,Comedy"},
		"WithStudios": {"Marvel Studios,Disney"},
	}
	q, err := Parse(params, uuid.New().String())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(q.WithGenres) != 2 {
		t.Errorf("WithGenres length = %d, want 2", len(q.WithGenres))
	}
	if len(q.WithStudios) != 2 {
		t.Errorf("WithStudios length = %d, want 2", len(q.WithStudios))
	}
}

// TestParse_SearchQuery verifies search text parsing.
// Jellyfin compat: search bar sends SearchQuery parameter.
func TestParse_SearchQuery(t *testing.T) {
	params := url.Values{
		"SearchQuery": {"lord of the rings"},
	}
	q, err := Parse(params, uuid.New().String())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if q.SearchQuery != "lord of the rings" {
		t.Errorf("SearchQuery = %q, want %q", q.SearchQuery, "lord of the rings")
	}
}

// TestParse_Fields verifies extra fields request parsing.
// Jellyfin compat: jellyfin-web requests Fields=Overview,People etc.
func TestParse_Fields(t *testing.T) {
	params := url.Values{
		"Fields": {"Overview,People,MediaSources"},
	}
	q, err := Parse(params, uuid.New().String())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(q.Fields) != 3 {
		t.Errorf("Fields length = %d, want 3", len(q.Fields))
	}
}

// TestParse_EnableUserData verifies user data flag.
// Jellyfin compat: enableUserData=true triggers UserData JOIN in queries.
func TestParse_EnableUserData(t *testing.T) {
	params := url.Values{
		"EnableUserData": {"true"},
	}
	q, err := Parse(params, uuid.New().String())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !q.EnableUserData {
		t.Error("EnableUserData should be true")
	}
}

// TestParse_InvalidUserId verifies invalid user ID is rejected.
// Jellyfin compat: user IDs must be valid UUIDs.
func TestParse_InvalidUserId(t *testing.T) {
	_, err := Parse(url.Values{}, "not-a-uuid")
	if err == nil {
		t.Error("expected error for invalid user ID")
	}
}

// TestParse_EmptyUserId verifies empty user ID is accepted.
// Jellyfin compat: some endpoints may not have user context.
func TestParse_EmptyUserId(t *testing.T) {
	q, err := Parse(url.Values{}, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if q.UserId != uuid.Nil {
		t.Errorf("UserId = %v, want Nil", q.UserId)
	}
}

// TestValidate_NegativeStartIndex verifies validation catches negative values.
// Jellyfin compat: prevents SQL injection via negative OFFSET.
func TestValidate_NegativeStartIndex(t *testing.T) {
	q := &InternalItemsQuery{StartIndex: -1, Limit: 10}
	if err := q.Validate(); err == nil {
		t.Error("expected error for negative StartIndex")
	}
}

// TestValidate_NegativeLimit verifies validation catches negative limit.
// Jellyfin compat: prevents SQL injection via negative LIMIT.
func TestValidate_NegativeLimit(t *testing.T) {
	q := &InternalItemsQuery{StartIndex: 0, Limit: -1}
	if err := q.Validate(); err == nil {
		t.Error("expected error for negative Limit")
	}
}

// TestValidate_ZeroLimitDefaulted verifies zero limit becomes 100.
// Jellyfin compat: omitting Limit should return a reasonable page.
func TestValidate_ZeroLimitDefaulted(t *testing.T) {
	q := &InternalItemsQuery{StartIndex: 0, Limit: 0}
	err := q.Validate()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if q.Limit != 100 {
		t.Errorf("Limit = %d, want 100 (default)", q.Limit)
	}
}

// TestValidate_OverLimitCapped verifies limit > 1000 is capped.
// Jellyfin compat: prevents unbounded queries.
func TestValidate_OverLimitCapped(t *testing.T) {
	q := &InternalItemsQuery{StartIndex: 0, Limit: 5000}
	err := q.Validate()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if q.Limit != 1000 {
		t.Errorf("Limit = %d, want 1000 (max)", q.Limit)
	}
}

// TestValidate_ValidParams verifies good params pass validation.
// Jellyfin compat: normal browsing parameters should always validate.
func TestValidate_ValidParams(t *testing.T) {
	q := &InternalItemsQuery{
		StartIndex:       0,
		Limit:            20,
		IncludeItemTypes: []string{"Movie"},
		SortBy:           []string{"SortName"},
		SortOrder:        []string{"Ascending"},
	}
	if err := q.Validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestParse_CommaSeparatedEmpty verifies empty comma-separated params return nil.
// Jellyfin compat: empty IncludeItemTypes means "all types".
func TestParse_CommaSeparatedEmpty(t *testing.T) {
	params := url.Values{
		"IncludeItemTypes": {""},
	}
	q, err := Parse(params, uuid.New().String())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if q.IncludeItemTypes != nil {
		t.Errorf("IncludeItemTypes = %v, want nil", q.IncludeItemTypes)
	}
}

// TestParse_MediaFilters verifies media type and codec filters.
// Jellyfin compat: advanced filter bar uses Container, VideoCodecs etc.
func TestParse_MediaFilters(t *testing.T) {
	params := url.Values{
		"Container":   {"mkv,mp4"},
		"VideoCodecs": {"h264,hevc"},
		"AudioCodecs": {"aac,ac3"},
		"MediaTypes":  {"Video,Audio"},
	}
	q, err := Parse(params, uuid.New().String())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(q.Container) != 2 {
		t.Errorf("Container length = %d, want 2", len(q.Container))
	}
	if len(q.VideoCodecs) != 2 {
		t.Errorf("VideoCodecs length = %d, want 2", len(q.VideoCodecs))
	}
	if len(q.AudioCodecs) != 2 {
		t.Errorf("AudioCodecs length = %d, want 2", len(q.AudioCodecs))
	}
	if len(q.MediaTypes) != 2 {
		t.Errorf("MediaTypes length = %d, want 2", len(q.MediaTypes))
	}
}

// TestParse_GroupBy verifies grouping flags.
// Jellyfin compat: music views group by album/series.
func TestParse_GroupBy(t *testing.T) {
	params := url.Values{
		"GroupByAlbum":  {"true"},
		"GroupBySeries": {"true"},
	}
	q, err := Parse(params, uuid.New().String())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !q.GroupByAlbum {
		t.Error("GroupByAlbum should be true")
	}
	if !q.GroupBySeries {
		t.Error("GroupBySeries should be true")
	}
	if !q.IsGrouped {
		t.Error("IsGrouped should be true when either group flag is set")
	}
}

// TestParse_Ratings verifies rating filter parsing.
// Jellyfin compat: parental control filters.
func TestParse_Ratings(t *testing.T) {
	params := url.Values{
		"ParentalRating":    {"13"},
		"MaxParentalRating": {"17"},
		"OfficialRatings":   {"PG-13,R"},
	}
	q, err := Parse(params, uuid.New().String())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if q.ParentalRating == nil || *q.ParentalRating != 13 {
		t.Errorf("ParentalRating = %v, want 13", q.ParentalRating)
	}
	if q.MaxParentalRating == nil || *q.MaxParentalRating != 17 {
		t.Errorf("MaxParentalRating = %v, want 17", q.MaxParentalRating)
	}
	if len(q.OfficialRatings) != 2 {
		t.Errorf("OfficialRatings length = %d, want 2", len(q.OfficialRatings))
	}
}

// TestGetIds_ValidAndInvalid verifies GUID filtering in GetIds.
// Jellyfin compat: Ids parameter may contain invalid entries that should be skipped.
func TestGetIds_ValidAndInvalid(t *testing.T) {
	validID := uuid.New().String()
	q := &InternalItemsQuery{
		Id: []string{validID, "invalid", "", uuid.Nil.String()},
	}
	ids := q.GetIds()
	if len(ids) != 1 {
		t.Errorf("GetIds() length = %d, want 1", len(ids))
	}
	if len(ids) > 0 && ids[0].String() != validID {
		t.Errorf("GetIds()[0] = %q, want %q", ids[0].String(), validID)
	}
}

// TestUseUserFilter verifies user data JOIN determination.
// Jellyfin compat: user data JOIN is expensive and should only happen when needed.
func TestUseUserFilter(t *testing.T) {
	uid := uuid.New()
	trueVal := true

	tests := []struct {
		name string
		q    InternalItemsQuery
		want bool
	}{
		{"no user", InternalItemsQuery{}, false},
		{"user with IsFavorite", InternalItemsQuery{UserId: uid, IsFavorite: &trueVal}, true},
		{"user with IsPlayed", InternalItemsQuery{UserId: uid, IsPlayed: &trueVal}, true},
		{"user with EnableUserData", InternalItemsQuery{UserId: uid, EnableUserData: true}, true},
		{"user without filters", InternalItemsQuery{UserId: uid}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.q.UseUserFilter(); got != tt.want {
				t.Errorf("UseUserFilter() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestParseCommaSeparated verifies the comma-separated string parser.
// Jellyfin compat: many query params are comma-separated lists.
func TestParseCommaSeparated(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"", 0},
		{"Movie", 1},
		{"Movie,Series", 2},
		{"Movie, Series, Episode", 3},
		{" , , ", 0},
	}
	for _, tt := range tests {
		result := parseCommaSeparated(tt.input)
		if len(result) != tt.want {
			t.Errorf("parseCommaSeparated(%q) length = %d, want %d", tt.input, len(result), tt.want)
		}
	}
}

// TestParseBool verifies boolean string parsing.
// Jellyfin compat: only "true" maps to true; all else is false or nil.
func TestParseBool(t *testing.T) {
	tests := []struct {
		input string
		isNil bool
		want  bool
	}{
		{"", true, false},
		{"true", false, true},
		{"false", false, false},
		{"TRUE", false, false},
		{"1", false, false},
	}
	for _, tt := range tests {
		result := parseBool(tt.input)
		if tt.isNil {
			if result != nil {
				t.Errorf("parseBool(%q) = %v, want nil", tt.input, *result)
			}
		} else {
			if result == nil {
				t.Errorf("parseBool(%q) = nil, want %v", tt.input, tt.want)
			} else if *result != tt.want {
				t.Errorf("parseBool(%q) = %v, want %v", tt.input, *result, tt.want)
			}
		}
	}
}

// TestParseInt verifies integer parsing with defaults.
// Jellyfin compat: invalid numeric parameters should not crash.
func TestParseInt(t *testing.T) {
	tests := []struct {
		input      string
		defaultVal int
		want       int
	}{
		{"", 42, 42},
		{"10", 0, 10},
		{"abc", 99, 99},
		{"-5", 0, -5},
	}
	for _, tt := range tests {
		got := parseInt(tt.input, tt.defaultVal)
		if got != tt.want {
			t.Errorf("parseInt(%q, %d) = %d, want %d", tt.input, tt.defaultVal, got, tt.want)
		}
	}
}

// TestParseIntPtr verifies nullable integer parsing.
// Jellyfin compat: optional integer filters should be nil when not provided.
func TestParseIntPtr(t *testing.T) {
	result := parseIntPtr("")
	if result != nil {
		t.Error("expected nil for empty string")
	}
	result = parseIntPtr("42")
	if result == nil || *result != 42 {
		t.Error("expected 42")
	}
	result = parseIntPtr("abc")
	if result != nil {
		t.Error("expected nil for invalid string")
	}
}

// TestParseIntArray verifies comma-separated integer array parsing.
// Jellyfin compat: Years parameter contains comma-separated integers.
func TestParseIntArray(t *testing.T) {
	result := parseIntArray("")
	if result != nil {
		t.Error("expected nil for empty string")
	}
	result = parseIntArray("2020,2021,2022")
	if len(result) != 3 {
		t.Errorf("length = %d, want 3", len(result))
	}
	result = parseIntArray("2020,abc,2022")
	if len(result) != 2 {
		t.Errorf("length = %d, want 2 (skipping invalid)", len(result))
	}
}
