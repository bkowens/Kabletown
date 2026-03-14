// Package query provides internal items query parsing and SQL building
// for the library-service content browsing functionality
package query

import (
	"fmt"
	"strings"
)

// SQLBuilder constructs optimized SQL queries using P7 indices
// The P7 index (items(top_parent_id, type)) enables efficient recursive CTE traversal
type SQLBuilder struct {
	query         *InternalItemsQuery
	params        []interface{}
	whereClauses  []string
	orderClauses  []string
}

// NewSQLBuilder creates a new SQL builder
func NewSQLBuilder(query *InternalItemsQuery) *SQLBuilder {
	return &SQLBuilder{
		query:        query,
		params:       make([]interface{}, 0),
		whereClauses: make([]string, 0),
		orderClauses: make([]string, 0),
	}
}

// Build constructs the complete SQL query
func (b *SQLBuilder) Build() (string, []interface{}) {
	// Base SELECT
	selectSQL := `SELECT 
		i.Id, i.Name, i.Path, i.PremiereDate, i.ProductionYear,
		i.IndexNumber, i.ParentIndexNumber, i.TopParentId,
		i.PresentationUniqueKey, i.OriginalTitle, i.OfficialRating,
		i.ParentalRating, i.ChannelId, i.ChannelType, i.Overview, i.OverviewSource,
		i.ChildCount, i.SeriesId, i.SeriesName, i.AirTime,
		i.AirDays, i.Status, i.Genres, i.Studios, i.Artists,
		i.Container, i.VideoCodec, i.AudioCodec, i.Width, i.Height,
		i.IsHD, i.Is4K, i.Is3D, i.HasSubtitles, i.HasTrailer,
		i.HasSpecialFeature, i.HasThemeSong, i.HasThemeVideo,
		i.DateCreated, i.DateModified, i.OriginatingSource,
		i.Tags, i.LockedFields,
		i.LockData, i.Metasource, i.ExtraType, i.RemoteIdData,
		i.ProviderIds, i.MediaSources, i.MediaStreams, i.Chapters,
		i.Locations, i.Type, i.IsFolder, i.IsVirtualItem,
		i.CanDelete, i.SupportsMediaDeletion, i.SupportsPartialUpdates,
		i.SupportsSync,
		ud.ResumePositionTicks, ud.PlaybackPositionTicks,
		ud.PlayedPercentage, ud.IsPlayed, ud.PlayedUntil,
		ud.Likes, ud.LastPlayedDate, ud.PlaybackCount, ud.CustomData,
		ud.Score, ud.ExternalEpisodeId, ud.ExternalSeriesId
	FROM items i
	LEFT JOIN user_data ud ON i.Id = ud.ItemId AND ud.UserId = ?`

	b.params = append(b.params, b.query.UserId)

	// Build WHERE clause
	b.buildIdFilter()
	b.buildParentFilter()
	b.buildAncestorFilter()
	b.buildItemTypeFilter()
	b.buildYearFilter()
	b.buildRatingFilter()
	b.buildGenreFilter()
	b.buildStudioFilter()
	b.buildPersonFilter()
	b.buildMediaFilter()
	b.buildQualityFilter()
	b.buildHasNotFilter()
	b.buildUserFilter()
	b.buildLocationFilter()
	b.buildSearchFilter()
	b.buildDateFilter()
	b.buildSpecialFilter()

	// Build complete WHERE
	var whereSQL string
	if len(b.whereClauses) > 0 {
		whereSQL = "WHERE " + strings.Join(b.whereClauses, " AND ")
	}

	// Build ORDER BY
	orderSQL := b.buildOrderBy()

	// Build pagination
	pagSQL := b.buildPagination()

	// Construct full query
	fullSQL := selectSQL
	if whereSQL != "" {
		fullSQL += "\n" + whereSQL
	}
	if orderSQL != "" {
		fullSQL += "\n" + orderSQL
	}
	fullSQL += pagSQL

	return fullSQL, b.params
}

// buildIdFilter filters by specific item IDs
func (b *SQLBuilder) buildIdFilter() {
	if len(b.query.Id) > 0 {
		placeholders := make([]string, len(b.query.Id))
		for i, id := range b.query.Id {
			placeholders[i] = "?"
			b.params = append(b.params, id)
		}
		b.whereClauses = append(b.whereClauses,
			fmt.Sprintf("i.Id IN (%s)", strings.Join(placeholders, ", ")))
	}
}

// buildParentFilter filters by parent ID (uses P7 index)
func (b *SQLBuilder) buildParentFilter() {
	if b.query.ParentId != "" {
		if b.query.Recursive {
			// Use P7 index: items(top_parent_id, type)
			b.whereClauses = append(b.whereClauses, "i.TopParentId = ?")
			b.params = append(b.params, b.query.ParentId)
		} else {
			b.whereClauses = append(b.whereClauses, "i.ParentId = ?")
			b.params = append(b.params, b.query.ParentId)
		}
	}
}

// buildAncestorFilter filters by ancestor IDs
func (b *SQLBuilder) buildAncestorFilter() {
	if len(b.query.AncestorIds) == 0 {
		return
	}

	placeholders := make([]string, len(b.query.AncestorIds))
	for i, id := range b.query.AncestorIds {
		placeholders[i] = "?"
		b.params = append(b.params, id)
	}
	b.whereClauses = append(b.whereClauses,
		"i.Id IN (WITH RECURSIVE ancestors AS (SELECT i.Id FROM items WHERE i.Id IN (" +
			strings.Join(placeholders, ", ") + ") UNION ALL SELECT i.Id FROM items i INNER JOIN ancestors a ON i.ParentId = a.Id) SELECT Id FROM ancestors)")
}

// buildYearFilter filters by production year

// buildItemTypeFilter filters by item type (uses P7 index)
func (b *SQLBuilder) buildItemTypeFilter() {
	if len(b.query.IncludeItemTypes) == 0 {
		return
	}

	placeholders := make([]string, len(b.query.IncludeItemTypes))
	for i, t := range b.query.IncludeItemTypes {
		placeholders[i] = "?"
		b.params = append(b.params, t)
	}
	b.whereClauses = append(b.whereClauses, "i.ItemType IN (" + strings.Join(placeholders, ", ") + ")")
}
func (b *SQLBuilder) buildYearFilter() {
	if b.query.MinProductionYear != nil {
		b.whereClauses = append(b.whereClauses, "i.ProductionYear >= ?")
		b.params = append(b.params, *b.query.MinProductionYear)
	}

	if b.query.MaxProductionYear != nil {
		b.whereClauses = append(b.whereClauses, "i.ProductionYear <= ?")
		b.params = append(b.params, *b.query.MaxProductionYear)
	}

	if len(b.query.Years) > 0 {
		placeholders := make([]string, len(b.query.Years))
		for i, y := range b.query.Years {
			placeholders[i] = "?"
			b.params = append(b.params, y)
		}
		b.whereClauses = append(b.whereClauses,
			fmt.Sprintf("i.ProductionYear IN (%s)", strings.Join(placeholders, ", ")))
	}
}

// buildRatingFilter filters by rating
func (b *SQLBuilder) buildRatingFilter() {
	if b.query.ParentalRating != nil {
		b.whereClauses = append(b.whereClauses, "i.ParentalRating = ?")
		b.params = append(b.params, *b.query.ParentalRating)
	}

	if b.query.MaxParentalRating != nil {
		b.whereClauses = append(b.whereClauses, "(i.ParentalRating <= ? OR i.ParentalRating IS NULL)")
		b.params = append(b.params, *b.query.MaxParentalRating)
	}

	if len(b.query.OfficialRatings) > 0 {
		placeholders := make([]string, len(b.query.OfficialRatings))
		for i, r := range b.query.OfficialRatings {
			placeholders[i] = "?"
			b.params = append(b.params, r)
		}
		b.whereClauses = append(b.whereClauses,
			fmt.Sprintf("i.OfficialRating IN (%s)", strings.Join(placeholders, ", ")))
	}
}

// buildGenreFilter filters by genre
func (b *SQLBuilder) buildGenreFilter() {
	// Check both Genres array and WithGenres (backward compatibility)
	genres := append([]string{}, b.query.WithGenres...)

	if len(genres) > 0 {
		placeholders := make([]string, len(genres))
		for i, g := range genres {
			placeholders[i] = "?"
			b.params = append(b.params, g)
		}
		// Use JSON_CONTAINS for array matching
		b.whereClauses = append(b.whereClauses,
			fmt.Sprintf("(JSON_CONTAINS(i.Genres, ?))",
				strings.Join(placeholders, ", ")))
	}
}

// buildStudioFilter filters by studio
func (b *SQLBuilder) buildStudioFilter() {
	// Check both Studios and WithStudios
	studios := append([]string{}, b.query.WithStudios...)

	if len(studios) > 0 {
		placeholders := make([]string, len(studios))
		for i, s := range studios {
			placeholders[i] = "?"
			b.params = append(b.params, s)
		}
		b.whereClauses = append(b.whereClauses,
			fmt.Sprintf("(JSON_CONTAINS(i.Studios, ?))",
				strings.Join(placeholders, ", ")))
	}
}

// buildPersonFilter filters by person/actor
func (b *SQLBuilder) buildPersonFilter() {
	// Check both WithPeople and PeopleIds
	if len(b.query.PeopleIds) > 0 {
		placeholders := make([]string, len(b.query.PeopleIds))
		for i, p := range b.query.PeopleIds {
			placeholders[i] = "?"
			b.params = append(b.params, p)
		}
		b.whereClauses = append(b.whereClauses,
			fmt.Sprintf("EXISTS (SELECT 1 FROM item_people ip WHERE ip.ItemId = items.Id AND ip.PersonId IN (%s))",
				strings.Join(placeholders, ", ")))
	}

	if len(b.query.WithPeople) > 0 {
		placeholders := make([]string, len(b.query.WithPeople))
		for i, p := range b.query.WithPeople {
			placeholders[i] = "?"
			b.params = append(b.params, p)
		}
		b.whereClauses = append(b.whereClauses,
			fmt.Sprintf("EXISTS (SELECT 1 FROM item_people ip WHERE ip.ItemId = items.Id AND ip.PersonName IN (%s))",
				strings.Join(placeholders, ", ")))
	}
}

// buildMediaFilter filters by media type and codec
func (b *SQLBuilder) buildMediaFilter() {
	if len(b.query.MediaTypes) > 0 {
		placeholders := make([]string, len(b.query.MediaTypes))
		for i, t := range b.query.MediaTypes {
			placeholders[i] = "?"
			b.params = append(b.params, t)
		}
		b.whereClauses = append(b.whereClauses,
			fmt.Sprintf("i.Type IN (%s)", strings.Join(placeholders, ", ")))
	}

	if len(b.query.Container) > 0 {
		placeholders := make([]string, len(b.query.Container))
		for i, c := range b.query.Container {
			placeholders[i] = "?"
			b.params = append(b.params, c)
		}
		b.whereClauses = append(b.whereClauses,
			fmt.Sprintf("i.Container IN (%s)", strings.Join(placeholders, ", ")))
	}

	if len(b.query.VideoCodecs) > 0 {
		placeholders := make([]string, len(b.query.VideoCodecs))
		for i, c := range b.query.VideoCodecs {
			placeholders[i] = "?"
			b.params = append(b.params, c)
		}
		b.whereClauses = append(b.whereClauses,
			fmt.Sprintf("i.VideoCodec IN (%s)", strings.Join(placeholders, ", ")))
	}

	if len(b.query.AudioCodecs) > 0 {
		placeholders := make([]string, len(b.query.AudioCodecs))
		for i, c := range b.query.AudioCodecs {
			placeholders[i] = "?"
			b.params = append(b.params, c)
		}
		b.whereClauses = append(b.whereClauses,
			fmt.Sprintf("i.AudioCodec IN (%s)", strings.Join(placeholders, ", ")))
	}
}

// buildQualityFilter filters by video quality
func (b *SQLBuilder) buildQualityFilter() {
	if b.query.Is4K != nil {
		if *b.query.Is4K {
			b.whereClauses = append(b.whereClauses, "i.Width >= 3840")
		} else {
			b.whereClauses = append(b.whereClauses, "i.Width < 3840")
		}
	}

	if b.query.IsHD != nil {
		if *b.query.IsHD && (b.query.Is4K == nil || !*b.query.Is4K) {
			b.whereClauses = append(b.whereClauses, "i.Width >= 1280 AND i.Width < 3840")
		} else if b.query.Is4K == nil {
			b.whereClauses = append(b.whereClauses, "i.Width >= 1280")
		}
	}

	if b.query.Is3D != nil {
		if *b.query.Is3D {
			b.whereClauses = append(b.whereClauses, "i.Is3D = 1")
		} else {
			b.whereClauses = append(b.whereClauses, "i.Is3D = 0")
		}
	}
}

// buildHasNotFilter filters by presence/absence of features
func (b *SQLBuilder) buildHasNotFilter() {
	if b.query.HasSubtitles != nil {
		if *b.query.HasSubtitles {
			b.whereClauses = append(b.whereClauses, "i.HasSubtitles = 1")
		} else {
			b.whereClauses = append(b.whereClauses, "i.HasSubtitles = 0")
		}
	}

	if b.query.HasTrailer != nil {
		if *b.query.HasTrailer {
			b.whereClauses = append(b.whereClauses, "i.HasTrailer = 1")
		} else {
			b.whereClauses = append(b.whereClauses, "i.HasTrailer = 0")
		}
	}

	if b.query.HasSpecialFeature != nil {
		if *b.query.HasSpecialFeature {
			b.whereClauses = append(b.whereClauses, "i.HasSpecialFeature = 1")
		} else {
			b.whereClauses = append(b.whereClauses, "i.HasSpecialFeature = 0")
		}
	}

	if b.query.HasThemeSong != nil {
		if *b.query.HasThemeSong {
			b.whereClauses = append(b.whereClauses, "i.HasThemeSong = 1")
		} else {
			b.whereClauses = append(b.whereClauses, "i.HasThemeSong = 0")
		}
	}

	if b.query.HasThemeVideo != nil {
		if *b.query.HasThemeVideo {
			b.whereClauses = append(b.whereClauses, "i.HasThemeVideo = 1")
		} else {
			b.whereClauses = append(b.whereClauses, "i.HasThemeVideo = 0")
		}
	}

	if b.query.IsMissing != nil {
		if *b.query.IsMissing {
			b.whereClauses = append(b.whereClauses, "i.Path IS NULL")
		} else {
			b.whereClauses = append(b.whereClauses, "i.Path IS NOT NULL")
		}
	}
}

// buildUserFilter filters by user data properties
func (b *SQLBuilder) buildUserFilter() {
	if b.query.IsFavorite != nil {
		if *b.query.IsFavorite {
			b.whereClauses = append(b.whereClauses, "ud.IsFavorite = 1")
		} else {
			b.whereClauses = append(b.whereClauses, "ud.IsFavorite = 0 OR ud.IsFavorite IS NULL")
		}
	}

	if b.query.IsPlayed != nil {
		if *b.query.IsPlayed {
			b.whereClauses = append(b.whereClauses, "ud.IsPlayed = 1")
		} else {
			b.whereClauses = append(b.whereClauses, "ud.IsPlayed = 0 OR ud.IsPlayed IS NULL")
		}
	}

	if b.query.IsUnplayed != nil {
		if *b.query.IsUnplayed {
			b.whereClauses = append(b.whereClauses, "(ud.IsPlayed = 0 OR ud.IsPlayed IS NULL)")
		}
	}

	if b.query.IsResumable != nil {
		if *b.query.IsResumable {
			b.whereClauses = append(b.whereClauses, "ud.ResumePositionTicks > 0 AND ud.PlayedPercentage < 90")
		} else {
			b.whereClauses = append(b.whereClauses, "ud.ResumePositionTicks = 0 OR ud.PlayedPercentage >= 90")
		}
	}
}

// buildLocationFilter filters by item location
func (b *SQLBuilder) buildLocationFilter() {
	if b.query.IsInMixedFolder != nil {
		if *b.query.IsInMixedFolder {
			b.whereClauses = append(b.whereClauses, "i.IsInMixedFolder = 1")
		} else {
			b.whereClauses = append(b.whereClauses, "i.IsInMixedFolder = 0")
		}
	}
}

// buildSearchFilter applies text search
func (b *SQLBuilder) buildSearchFilter() {
	if b.query.SearchQuery != "" {
		// Simple text search on name and overview
		searchParam := "%" + b.query.SearchQuery + "%"
		b.params = append(b.params, searchParam)
		b.whereClauses = append(b.whereClauses,
			"(i.Name LIKE ? OR i.Overview LIKE ?)")
	} else if b.query.NameStartsWithOrContains != "" {
		// Like search but without wildcards
		param := "%" + b.query.NameStartsWithOrContains + "%"
		b.params = append(b.params, param)
		b.whereClauses = append(b.whereClauses,
			"i.Name LIKE ?")
	}
}

// buildDateFilter filters by date fields
func (b *SQLBuilder) buildDateFilter() {
	if b.query.MinDateLastSaved != nil {
		b.whereClauses = append(b.whereClauses, "ud.LastPlayedDate >= ?")
		b.params = append(b.params, b.query.MinDateLastSaved)
	}

	if b.query.MaxDateLastSaved != nil {
		b.whereClauses = append(b.whereClauses, "ud.LastPlayedDate <= ?")
		b.params = append(b.params, b.query.MaxDateLastSaved)
	}

	if b.query.StartDate != nil {
		b.whereClauses = append(b.whereClauses, "i.PremiereDate >= ?")
		b.params = append(b.params, b.query.StartDate)
	}

	if b.query.EndDate != nil {
		b.whereClauses = append(b.whereClauses, "i.PremiereDate <= ?")
		b.params = append(b.params, b.query.EndDate)
	}

	if b.query.MinEndDate != nil {
		b.whereClauses = append(b.whereClauses, "CAST(i.EndDate AS DATETIME) >= ?")
		b.params = append(b.params, b.query.MinEndDate)
	}

	if b.query.MaxEndDate != nil {
		b.whereClauses = append(b.whereClauses, "CAST(i.EndDate AS DATETIME) <= ?")
		b.params = append(b.params, b.query.MaxEndDate)
	}
}

// buildSpecialFilter handles special query conditions
func (b *SQLBuilder) buildSpecialFilter() {
	if b.query.IsFavoriteOrResumable {
		b.whereClauses = append(b.whereClauses, "(ud.IsFavorite = 1 OR (ud.ResumePositionTicks > 0 AND ud.PlayedPercentage < 90))")
	}
}

// buildOrderBy constructs ORDER BY clause
func (b *SQLBuilder) buildOrderBy() string {
	if len(b.query.SortBy) == 0 {
		// Default sort by name
		return "ORDER BY i.Name"
	}

	orderParts := make([]string, len(b.query.SortBy))
	sortOrders := b.query.SortOrder
	if len(sortOrders) == 0 {
		sortOrders = []string{"Ascending"}
	}

	for i, field := range b.query.SortBy {
		direction := "ASC"
		if i < len(sortOrders) && strings.ToLower(sortOrders[i]) == "descending" {
			direction = "DESC"
		}

		// Map sort field to column
		var column string
		switch strings.ToLower(field) {
		case "sortname":
			column = "i.Name"
		case "name":
			column = "i.Name"
		case "dateadded":
			column = "i.DateCreated"
		case "datecreated":
			column = "i.DateCreated"
		case "datemodified":
			column = "i.DateModified"
		case "premiere":
			column = "i.PremiereDate"
		case "productionyear":
			column = "i.ProductionYear"
		case "random":
			return "ORDER BY RAND()"
		case "runtime":
			column = "i.RunTimeTicks"
		case "playbackcount":
			column = "ud.PlaybackCount"
		case "dateplayed":
			column = "ud.LastPlayedDate"
		default:
			column = field
		}

		orderParts[i] = fmt.Sprintf("%s %s", column, direction)
	}

	return "ORDER BY " + strings.Join(orderParts, ", ")
}

// buildPagination adds LIMIT and OFFSET
func (b *SQLBuilder) buildPagination() string {
	page := int64(b.query.StartIndex)
	limit := int64(b.query.Limit)

	return fmt.Sprintf("\nLIMIT %d OFFSET %d", limit, page)
}

// GetCountQuery returns a COUNT query for pagination
func (b *SQLBuilder) GetCountQuery() (string, []interface{}) {
	// Similar to Build() but returns COUNT instead of column list
	baseSQL := `SELECT COUNT(DISTINCT i.Id) FROM items i`
	baseSQL += `LEFT JOIN user_data ud ON i.Id = ud.ItemId AND ud.UserId = ?`

	// Rebuild where clauses for count
	countB := &SQLBuilder{
		query:        b.query,
		params:       make([]interface{}, 0),
		whereClauses: make([]string, 0),
	}

	countB.params = append(countB.params, b.query.UserId)
	countB.buildIdFilter()
	countB.buildParentFilter()
	countB.buildAncestorFilter()
	countB.buildItemTypeFilter()
	countB.buildYearFilter()
	countB.buildRatingFilter()
	countB.buildGenreFilter()
	countB.buildStudioFilter()
	countB.buildPersonFilter()
	countB.buildMediaFilter()
	countB.buildQualityFilter()
	countB.buildHasNotFilter()
	countB.buildUserFilter()
	countB.buildLocationFilter()
	countB.buildSearchFilter()
	countB.buildDateFilter()
	countB.buildSpecialFilter()

	var whereSQL string
	if len(countB.whereClauses) > 0 {
		whereSQL = "WHERE " + strings.Join(countB.whereClauses, " AND ")
	}

	fullSQL := baseSQL
	if whereSQL != "" {
		fullSQL += "\n" + whereSQL
	}

	return fullSQL, countB.params
}