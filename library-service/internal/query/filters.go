// Package query provides query filter implementations
package query

import (
	"fmt"
	"strings"

	"github.com/jellyfinhanced/shared/types"
)

// FilterType represents different types of item filters
type FilterType int

const (
	FilterIsFavorite FilterType = iota
	FilterIsPlayed
	FilterHasSubtitles
	FilterHasTrailer
	FilterHasSpecialFeature
	FilterGenre
	FilterStudio
	FilterPerson
	FilterYear
	FilterMediaSource
	FilterQuality
	FilterLocation
	FilterUserFilter
	FilterSearch
	FilterDate
	FilterSpecial
)

// Filter represents a single filter condition
type Filter struct {
	Type    FilterType
	Name    string
	Value   interface{}
	Op      string // =, !=, IN, LIKE, >, <
	Builder func(interface{}) (string, []interface{})
}

// FilterFactory creates filters from string parameters
var FilterFactory = map[string]Filter{
	"IsFavorite": {
		Name: "IsFavorite",
		Builder: func(v interface{}) (string, []interface{}) {
			val := v.(bool)
			if val {
				return "ud.IsFavorite = 1", nil
			}
			return "(ud.IsFavorite = 0 OR ud.IsFavorite IS NULL)", nil
		},
	},
	"IsPlayed": {
		Name: "IsPlayed",
		Builder: func(v interface{}) (string, []interface{}) {
			val := v.(bool)
			if val {
				return "ud.IsPlayed = 1", nil
			}
			return "(ud.IsPlayed = 0 OR ud.IsPlayed IS NULL)", nil
		},
	},
	"IsUnplayed": {
		Name: "IsUnplayed",
		Builder: func(v interface{}) (string, []interface{}) {
			val := v.(bool)
			if val {
				return "(ud.IsPlayed = 0 OR ud.IsPlayed IS NULL)", nil
			}
			return "ud.IsPlayed = 1", nil
		},
	},
	"HasSubtitles": {
		Name: "HasSubtitles",
		Builder: func(v interface{}) (string, []interface{}) {
			val := v.(bool)
			if val {
				return "i.HasSubtitles = 1", nil
			}
			return "i.HasSubtitles = 0", nil
		},
	},
	"HasTrailer": {
		Name: "HasTrailer",
		Builder: func(v interface{}) (string, []interface{}) {
			val := v.(bool)
			if val {
				return "i.HasTrailer = 1", nil
			}
			return "i.HasTrailer = 0", nil
		},
	},
	"HasSpecialFeature": {
		Name: "HasSpecialFeature",
		Builder: func(v interface{}) (string, []interface{}) {
			val := v.(bool)
			if val {
				return "i.HasSpecialFeature = 1", nil
			}
			return "i.HasSpecialFeature = 0", nil
		},
	},
	"IsMissing": {
		Name: "IsMissing",
		Builder: func(v interface{}) (string, []interface{}) {
			val := v.(bool)
			if val {
				return "i.Path IS NULL", nil
			}
			return "i.Path IS NOT NULL", nil
		},
	},
	"IsMovie": {
		Name: "IsMovie",
		Builder: func(v interface{}) (string, []interface{}) {
			val := v.(bool)
			if val {
				return "i.IsMovie = 1", nil
			}
			return "i.IsMovie = 0", nil
		},
	},
	"IsSeries": {
		Name: "IsSeries",
		Builder: func(v interface{}) (string, []interface{}) {
			val := v.(bool)
			if val {
				return "i.IsSeries = 1", nil
			}
			return "i.IsSeries = 0", nil
		},
	},
	"IsLiveTv": {
		Name: "IsLiveTv",
		Builder: func(v interface{}) (string, []interface{}) {
			val := v.(bool)
			if val {
				return "i.IsLiveTv = 1", nil
			}
			return "i.IsLiveTv = 0", nil
		},
	},
}

// ParseFilters parses the Filters query string parameter
func ParseFilters(filters string) []string {
	if filters == "" {
		return nil
	}

	var result []string
	parts := strings.Split(filters, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}

	return result
}

// BuildFilterClause builds a SQL WHERE clause from filter name and value
func BuildFilterClause(filterName string, value interface{}) (string, []interface{}) {
	filter, ok := FilterFactory[filterName]
	if !ok {
		return "", nil
	}

	return filter.Builder(value)
}

// BuildMultipleFilters builds SQL WHERE clauses from multiple filter strings
func BuildMultipleFilters(filterStrings []string) (string, []interface{}) {
	if len(filterStrings) == 0 {
		return "", nil
	}

	var clauses []string
	var params []interface{}

	for _, s := range filterStrings {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}

		// Format: "FilterName=Value" or "FilterName"
		parts := strings.Split(s, "=")
		if len(parts) != 2 {
			continue
		}

		filterName := strings.TrimSpace(parts[0])
		valueStr := strings.TrimSpace(parts[1])

		var value interface{}
		switch filterName {
		case "IsFavorite", "IsPlayed", "IsUnplayed", "HasSubtitles", "HasTrailer", "HasSpecialFeature",
			"IsMissing", "IsMovie", "IsSeries", "IsLiveTv":
			value = valueStr == "true"
		default:
			value = valueStr
		}

		clause, p := BuildFilterClause(filterName, value)
		if clause != "" {
			clauses = append(clauses, clause)
			params = append(params, p...)
		}
	}

	if len(clauses) == 0 {
		return "", nil
	}

	return "(" + strings.Join(clauses, " AND ") + ")", params
}

// GenreFilter handles genre-based filtering
type GenreFilter struct {
	Genres []string
	Op     string // "=", "IN", "AND", "OR"
}

// BuildSQL builds SQL for genre filtering
func (gf *GenreFilter) BuildSQL() (string, []interface{}) {
	if len(gf.Genres) == 0 {
		return "", nil
	}

	switch gf.Op {
	case "IN", "":
		placeholders := make([]string, len(gf.Genres))
		var params []interface{}
		for i, g := range gf.Genres {
			placeholders[i] = "?"
			params = append(params, g)
		}
		// Use JSON_CONTAINS for array matching
		return fmt.Sprintf(`(i.Genres IN (%s) OR JSON_CONTAINS(i.Genres, ?))`,
			strings.Join(placeholders, ", ")),
			append(params, gf.Genres[0])

	case "AND":
		// Item must have all genres
		var clauses []string
		var params []interface{}
		for i := range gf.Genres {
			clauses = append(clauses, "JSON_CONTAINS(i.Genres, ?)")
			params = append(params, gf.Genres[i])
		}
		return "(" + strings.Join(clauses, " AND ") + ")", params

	case "OR":
		// Item must have any genre
		var clause string
		for i, _ := range gf.Genres {
			if i == 0 {
				clause = "JSON_CONTAINS(i.Genres, ?)"
			} else {
				clause += " OR JSON_CONTAINS(i.Genres, ?)"
			}
		}
		params := make([]interface{}, len(gf.Genres))
		for i, genre := range gf.Genres {
			params[i] = genre
		}
		return clause, params

	default:
		return "", nil
	}
}

// StudioFilter handles studio-based filtering
type StudioFilter struct {
	Studios []string
	Op      string
}

// BuildSQL builds SQL for studio filtering
func (sf *StudioFilter) BuildSQL() (string, []interface{}) {
	if len(sf.Studios) == 0 {
		return "", nil
	}

	placeholders := make([]string, len(sf.Studios))
	var params []interface{}
	for i, s := range sf.Studios {
		placeholders[i] = "?"
		params = append(params, s)
	}

	return fmt.Sprintf(`(i.Studios IN (%s) OR JSON_CONTAINS(i.Studios, ?))`,
		strings.Join(placeholders, ", ")),
		append(params, sf.Studios[0])
}

// PersonFilter handles person/actor filtering
type PersonFilter struct {
	PersonIds []string
	PeopleNames []string
}

// BuildSQL builds SQL for person filtering
func (pf *PersonFilter) BuildSQL() (string, []interface{}) {
	var clauses []string
	var params []interface{}

	if len(pf.PersonIds) > 0 {
		placeholders := make([]string, len(pf.PersonIds))
		for i, id := range pf.PersonIds {
			placeholders[i] = "?"
			params = append(params, id)
		}
		clauses = append(clauses,
			fmt.Sprintf("EXISTS (SELECT 1 FROM item_people ip WHERE ip.ItemId = items.Id AND ip.PersonId IN (%s))",
				strings.Join(placeholders, ", ")))
	}

	if len(pf.PeopleNames) > 0 {
		placeholders := make([]string, len(pf.PeopleNames))
		for i, name := range pf.PeopleNames {
			placeholders[i] = "?"
			params = append(params, name)
		}
		clauses = append(clauses,
			fmt.Sprintf("EXISTS (SELECT 1 FROM item_people ip WHERE ip.ItemId = items.Id AND ip.PersonName IN (%s))",
				strings.Join(placeholders, ", ")))
	}

	if len(clauses) == 0 {
		return "", nil
	}

	return "(" + strings.Join(clauses, " AND ") + ")", params
}

// YearFilter handles production year filtering
type YearFilter struct {
	MinYear int
	MaxYear int
	Years   []int
}

// BuildSQL builds SQL for year filtering
func (yf *YearFilter) BuildSQL() (string, []interface{}) {
	var clauses []string
	var params []interface{}

	if yf.MinYear != 0 {
		clauses = append(clauses, "i.ProductionYear >= ?")
		params = append(params, yf.MinYear)
	}

	if yf.MaxYear != 0 {
		clauses = append(clauses, "i.ProductionYear <= ?")
		params = append(params, yf.MaxYear)
	}

	if len(yf.Years) > 0 {
		placeholders := make([]string, len(yf.Years))
		for i, year := range yf.Years {
			placeholders[i] = "?"
			params = append(params, year)
		}
		clauses = append(clauses,
			fmt.Sprintf("i.ProductionYear IN (%s)", strings.Join(placeholders, ", ")))
	}

	if len(clauses) == 0 {
		return "", nil
	}

	return "(" + strings.Join(clauses, " AND ") + ")", params
}

// QualityFilter handles quality-based filtering (HD, 4K, 3D)
type QualityFilter struct {
	Is4K *bool
	IsHD *bool
	Is3D *bool
}

// BuildSQL builds SQL for quality filtering
func (qf *QualityFilter) BuildSQL() (string, []interface{}) {
	var clauses []string

	if qf.Is4K != nil {
		if *qf.Is4K {
			clauses = append(clauses, "i.Width >= 3840")
		} else {
			clauses = append(clauses, "i.Width < 3840")
		}
	}

	if qf.IsHD != nil {
		if *qf.IsHD && (qf.Is4K == nil || !*qf.Is4K) {
			clauses = append(clauses, "i.Width >= 1280 AND i.Width < 3840")
		} else if qf.Is4K == nil {
			clauses = append(clauses, "i.Width >= 1280")
		}
	}

	if qf.Is3D != nil {
		if *qf.Is3D {
			clauses = append(clauses, "i.Is3D = 1")
		} else {
			clauses = append(clauses, "i.Is3D = 0")
		}
	}

	if len(clauses) == 0 {
		return "", nil
	}

	return "(" + strings.Join(clauses, " AND ") + ")", nil
}

// UserFilter handles user-specific data filtering (resume, played, favorite)
type UserFilter struct {
	IsFavorite   *bool
	IsPlayed     *bool
	IsUnplayed   *bool
	IsResumable  *bool
}

// BuildSQL builds SQL for user data filtering
func (uf *UserFilter) BuildSQL() (string, []interface{}) {
	var clauses []string

	if uf.IsFavorite != nil {
		if *uf.IsFavorite {
			clauses = append(clauses, "ud.IsFavorite = 1")
		} else {
			clauses = append(clauses, "(ud.IsFavorite = 0 OR ud.IsFavorite IS NULL)")
		}
	}

	if uf.IsPlayed != nil {
		if *uf.IsPlayed {
			clauses = append(clauses, "ud.IsPlayed = 1")
		} else {
			clauses = append(clauses, "(ud.IsPlayed = 0 OR ud.IsPlayed IS NULL)")
		}
	}

	if uf.IsUnplayed != nil && *uf.IsUnplayed {
		clauses = append(clauses, "(ud.IsPlayed = 0 OR ud.IsPlayed IS NULL)")
	}

	if uf.IsResumable != nil && *uf.IsResumable {
		clauses = append(clauses, "ud.ResumePositionTicks > 0 AND ud.PlayedPercentage < 90")
	}

	if len(clauses) == 0 {
		return "", nil
	}

	return "(" + strings.Join(clauses, " AND ") + ")", nil
}

// SearchTextFilter handles full-text search
type SearchTextFilter struct {
	Query string
}

// BuildSQL builds SQL for text search
func (stf *SearchTextFilter) BuildSQL() (string, []interface{}) {
	if stf.Query == "" {
		return "", nil
	}

	searchParam := "%" + strings.ReplaceAll(stf.Query, "\\", "\\\\") + "%"
	return "(i.Name LIKE ? OR i.Overview LIKE ?)", []interface{}{searchParam, searchParam}
}

// DateFilter handles date-based filtering
type DateFilter struct {
	MinDateLastSaved *types.JellyfinTime
	MaxDateLastSaved *types.JellyfinTime
}

// BuildSQL builds SQL for date filtering
func (df *DateFilter) BuildSQL() (string, []interface{}) {
	var clauses []string
	var params []interface{}

	if df.MinDateLastSaved != nil {
		clauses = append(clauses, "ud.LastPlayedDate >= ?")
		params = append(params, df.MinDateLastSaved)
	}

	if df.MaxDateLastSaved != nil {
		clauses = append(clauses, "ud.LastPlayedDate <= ?")
		params = append(params, df.MaxDateLastSaved)
	}

	if len(clauses) == 0 {
		return "", nil
	}

	return "(" + strings.Join(clauses, " AND ") + ")", params
}

// MediaTypeFilter handles media type filtering
type MediaTypeFilter struct {
	MediaTypes []string
	Container  []string
	VideoCodecs []string
	AudioCodecs []string
}

// BuildSQL builds SQL for media type filtering
func (mtf *MediaTypeFilter) BuildSQL() (string, []interface{}) {
	var clauses []string
	var params []interface{}

	if len(mtf.MediaTypes) > 0 {
		placeholders := make([]string, len(mtf.MediaTypes))
		for i, t := range mtf.MediaTypes {
			placeholders[i] = "?"
			params = append(params, t)
		}
		clauses = append(clauses,
			fmt.Sprintf("i.Type IN (%s)", strings.Join(placeholders, ", ")))
	}

	if len(mtf.Container) > 0 {
		placeholders := make([]string, len(mtf.Container))
		for i, c := range mtf.Container {
			placeholders[i] = "?"
			params = append(params, c)
		}
		clauses = append(clauses,
			fmt.Sprintf("i.Container IN (%s)", strings.Join(placeholders, ", ")))
	}

	if len(mtf.VideoCodecs) > 0 {
		placeholders := make([]string, len(mtf.VideoCodecs))
		for i, c := range mtf.VideoCodecs {
			placeholders[i] = "?"
			params = append(params, c)
		}
		clauses = append(clauses,
			fmt.Sprintf("i.VideoCodec IN (%s)", strings.Join(placeholders, ", ")))
	}

	if len(mtf.AudioCodecs) > 0 {
		placeholders := make([]string, len(mtf.AudioCodecs))
		for i, c := range mtf.AudioCodecs {
			placeholders[i] = "?"
			params = append(params, c)
		}
		clauses = append(clauses,
			fmt.Sprintf("i.AudioCodec IN (%s)", strings.Join(placeholders, ", ")))
	}

	if len(clauses) == 0 {
		return "", nil
	}

	return "(" + strings.Join(clauses, " AND ") + ")", params
}