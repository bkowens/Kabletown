package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

// Item represents a row from base_items joined with optional UserData.
type Item struct {
	Id             string          `db:"Id"`
	Name           string          `db:"Name"`
	Type           string          `db:"Type"`
	IsFolder       bool            `db:"IsFolder"`
	ParentId       sql.NullString  `db:"ParentId"`
	TopParentId    sql.NullString  `db:"TopParentId"`
	Path           sql.NullString  `db:"Path"`
	Container      sql.NullString  `db:"Container"`
	DurationTicks  sql.NullInt64   `db:"DurationTicks"`
	Size           sql.NullInt64   `db:"Size"`
	Width          sql.NullInt32   `db:"Width"`
	Height         sql.NullInt32   `db:"Height"`
	ProductionYear sql.NullInt32   `db:"ProductionYear"`
	PremiereDate   sql.NullTime    `db:"PremiereDate"`
	DateCreated    time.Time       `db:"DateCreated"`
	DateModified   time.Time       `db:"DateModified"`
	ExtraData      json.RawMessage `db:"ExtraData"`
	AncestorIds    sql.NullString  `db:"AncestorIds"`

	// UserData fields (from LEFT JOIN; only populated when UserID is set)
	UDPlayed                sql.NullBool    `db:"ud_Played"`
	UDPlayCount             sql.NullInt32   `db:"ud_PlayCount"`
	UDIsFavorite            sql.NullBool    `db:"ud_IsFavorite"`
	UDPlaybackPositionTicks sql.NullInt64   `db:"ud_PlaybackPositionTicks"`
	UDLastPlayedDate        sql.NullTime    `db:"ud_LastPlayedDate"`
	UDRating                sql.NullFloat64 `db:"ud_Rating"`
}

// ItemQuery holds all filter/sort/pagination parameters for querying base_items.
type ItemQuery struct {
	UserID                 string
	ParentID               string
	IncludeItemTypes       []string
	ExcludeItemTypes       []string
	IsFolder               *bool
	Recursive              bool
	SortBy                 []string
	SortOrder              string
	StartIndex             int
	Limit                  int
	SearchTerm             string
	Genres                 []string
	Studios                []string
	Tags                   []string
	Years                  []int
	Filters                []string // IsPlayed, IsUnplayed, IsFavorite
	MediaTypes             []string
	EnableTotalRecordCount bool
	HasTmdbId              *bool
	HasImdbId              *bool
}

// ItemRepository provides read access to base_items and related tables.
type ItemRepository struct {
	db *sqlx.DB
}

// NewItemRepository creates a new ItemRepository.
func NewItemRepository(database *sqlx.DB) *ItemRepository {
	return &ItemRepository{db: database}
}

// QueryItems executes a dynamic query against base_items based on ItemQuery parameters.
// Returns the page of items, the total record count, and any error.
func (r *ItemRepository) QueryItems(q ItemQuery) ([]Item, int, error) {
	if q.Limit <= 0 {
		q.Limit = 50
	}
	if q.Limit > 1000 {
		q.Limit = 1000
	}

	where, args := buildWhereClause(q)

	// Determine ORDER BY clause.
	orderBy := buildOrderBy(q.SortBy, q.SortOrder)

	// Base SELECT columns.
	cols := `b.Id, b.Name, b.Type, b.IsFolder, b.ParentId, b.TopParentId,
		b.Path, b.Container, b.DurationTicks, b.Size, b.Width, b.Height,
		b.ProductionYear, b.PremiereDate, b.DateCreated, b.DateModified,
		b.ExtraData, b.AncestorIds`

	var udCols string
	if q.UserID != "" {
		udCols = `,
		COALESCE(ud.Played, 0)                AS ud_Played,
		COALESCE(ud.PlayCount, 0)             AS ud_PlayCount,
		COALESCE(ud.IsFavorite, 0)            AS ud_IsFavorite,
		COALESCE(ud.PlaybackPositionTicks, 0) AS ud_PlaybackPositionTicks,
		ud.LastPlayedDate                     AS ud_LastPlayedDate,
		ud.Rating                             AS ud_Rating`
	} else {
		udCols = `,
		NULL AS ud_Played,
		NULL AS ud_PlayCount,
		NULL AS ud_IsFavorite,
		NULL AS ud_PlaybackPositionTicks,
		NULL AS ud_LastPlayedDate,
		NULL AS ud_Rating`
	}

	join := ""
	if q.UserID != "" {
		join = fmt.Sprintf(`LEFT JOIN UserData ud ON ud.ItemId = b.Id AND ud.UserId = %s`, placeholder(len(args)+1))
		args = append(args, q.UserID)
	}

	// Subquery joins for genre/studio/tag filtering.
	genreJoin, genreArgs := buildValueJoin(q.Genres, 2, len(args)+1)
	args = append(args, genreArgs...)
	studioJoin, studioArgs := buildValueJoin(q.Studios, 3, len(args)+1)
	args = append(args, studioArgs...)
	tagJoin, tagArgs := buildValueJoin(q.Tags, 6, len(args)+1)
	args = append(args, tagArgs...)

	fromClause := fmt.Sprintf("FROM base_items b %s %s %s %s", join, genreJoin, studioJoin, tagJoin)

	// Count query.
	var total int
	if q.EnableTotalRecordCount {
		countSQL := fmt.Sprintf("SELECT COUNT(*) %s %s", fromClause, where)
		if err := r.db.QueryRow(countSQL, args...).Scan(&total); err != nil {
			return nil, 0, fmt.Errorf("item_repository.QueryItems count: %w", err)
		}
	}

	// Data query.
	dataSQL := fmt.Sprintf(
		"SELECT %s %s %s %s %s LIMIT %d OFFSET %d",
		cols+udCols,
		fromClause,
		where,
		orderBy,
		"",
		q.Limit,
		q.StartIndex,
	)

	rows, err := r.db.Queryx(dataSQL, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("item_repository.QueryItems: %w", err)
	}
	defer rows.Close()

	var items []Item
	for rows.Next() {
		var it Item
		if err := rows.StructScan(&it); err != nil {
			return nil, 0, fmt.Errorf("item_repository.QueryItems scan: %w", err)
		}
		items = append(items, it)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("item_repository.QueryItems rows: %w", err)
	}

	if items == nil {
		items = []Item{}
	}
	return items, total, nil
}

// GetItemByID fetches a single item by its primary key.
func (r *ItemRepository) GetItemByID(id string) (*Item, error) {
	const q = `SELECT b.Id, b.Name, b.Type, b.IsFolder, b.ParentId, b.TopParentId,
		b.Path, b.Container, b.DurationTicks, b.Size, b.Width, b.Height,
		b.ProductionYear, b.PremiereDate, b.DateCreated, b.DateModified,
		b.ExtraData, b.AncestorIds,
		NULL AS ud_Played, NULL AS ud_PlayCount, NULL AS ud_IsFavorite,
		NULL AS ud_PlaybackPositionTicks, NULL AS ud_LastPlayedDate, NULL AS ud_Rating
	FROM base_items b WHERE b.Id = ? LIMIT 1`

	var it Item
	if err := r.db.QueryRowx(q, id).StructScan(&it); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("item_repository.GetItemByID: %w", err)
	}
	return &it, nil
}

// GetItemCounts returns the count of each significant item type.
func (r *ItemRepository) GetItemCounts() (map[string]int, error) {
	const q = `SELECT Type, COUNT(*) AS cnt FROM base_items GROUP BY Type`
	rows, err := r.db.Query(q)
	if err != nil {
		return nil, fmt.Errorf("item_repository.GetItemCounts: %w", err)
	}
	defer rows.Close()

	counts := make(map[string]int)
	for rows.Next() {
		var t string
		var c int
		if err := rows.Scan(&t, &c); err != nil {
			return nil, fmt.Errorf("item_repository.GetItemCounts scan: %w", err)
		}
		counts[t] = c
	}
	return counts, rows.Err()
}

// GetResumeItems returns in-progress items for a user (position > 0 and < duration).
func (r *ItemRepository) GetResumeItems(userID string, startIndex, limit int) ([]Item, int, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 1000 {
		limit = 1000
	}

	const countSQL = `SELECT COUNT(*) FROM base_items b
		INNER JOIN UserData ud ON ud.ItemId = b.Id AND ud.UserId = ?
		WHERE ud.PlaybackPositionTicks > 0
		AND (b.DurationTicks IS NULL OR ud.PlaybackPositionTicks < b.DurationTicks)`

	var total int
	if err := r.db.QueryRow(countSQL, userID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("item_repository.GetResumeItems count: %w", err)
	}

	const dataSQL = `SELECT b.Id, b.Name, b.Type, b.IsFolder, b.ParentId, b.TopParentId,
		b.Path, b.Container, b.DurationTicks, b.Size, b.Width, b.Height,
		b.ProductionYear, b.PremiereDate, b.DateCreated, b.DateModified,
		b.ExtraData, b.AncestorIds,
		COALESCE(ud.Played, 0)                AS ud_Played,
		COALESCE(ud.PlayCount, 0)             AS ud_PlayCount,
		COALESCE(ud.IsFavorite, 0)            AS ud_IsFavorite,
		COALESCE(ud.PlaybackPositionTicks, 0) AS ud_PlaybackPositionTicks,
		ud.LastPlayedDate                     AS ud_LastPlayedDate,
		ud.Rating                             AS ud_Rating
	FROM base_items b
		INNER JOIN UserData ud ON ud.ItemId = b.Id AND ud.UserId = ?
	WHERE ud.PlaybackPositionTicks > 0
	AND (b.DurationTicks IS NULL OR ud.PlaybackPositionTicks < b.DurationTicks)
	ORDER BY ud.LastPlayedDate DESC
	LIMIT ? OFFSET ?`

	rows, err := r.db.Queryx(dataSQL, userID, limit, startIndex)
	if err != nil {
		return nil, 0, fmt.Errorf("item_repository.GetResumeItems: %w", err)
	}
	defer rows.Close()

	var items []Item
	for rows.Next() {
		var it Item
		if err := rows.StructScan(&it); err != nil {
			return nil, 0, fmt.Errorf("item_repository.GetResumeItems scan: %w", err)
		}
		items = append(items, it)
	}
	if items == nil {
		items = []Item{}
	}
	return items, total, rows.Err()
}

// GetItemAncestors returns all ancestor items for the given item using its AncestorIds field.
func (r *ItemRepository) GetItemAncestors(id string) ([]Item, error) {
	// First get the item to read its AncestorIds.
	var ancestorIDs sql.NullString
	if err := r.db.QueryRow(`SELECT AncestorIds FROM base_items WHERE Id = ? LIMIT 1`, id).Scan(&ancestorIDs); err != nil {
		if err == sql.ErrNoRows {
			return []Item{}, nil
		}
		return nil, fmt.Errorf("item_repository.GetItemAncestors: %w", err)
	}
	if !ancestorIDs.Valid || ancestorIDs.String == "" {
		return []Item{}, nil
	}

	ids := strings.Split(ancestorIDs.String, ",")
	if len(ids) == 0 {
		return []Item{}, nil
	}

	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, aid := range ids {
		placeholders[i] = "?"
		args[i] = strings.TrimSpace(aid)
	}

	q := fmt.Sprintf(`SELECT b.Id, b.Name, b.Type, b.IsFolder, b.ParentId, b.TopParentId,
		b.Path, b.Container, b.DurationTicks, b.Size, b.Width, b.Height,
		b.ProductionYear, b.PremiereDate, b.DateCreated, b.DateModified,
		b.ExtraData, b.AncestorIds,
		NULL AS ud_Played, NULL AS ud_PlayCount, NULL AS ud_IsFavorite,
		NULL AS ud_PlaybackPositionTicks, NULL AS ud_LastPlayedDate, NULL AS ud_Rating
	FROM base_items b WHERE b.Id IN (%s)`, strings.Join(placeholders, ","))

	rows, err := r.db.Queryx(q, args...)
	if err != nil {
		return nil, fmt.Errorf("item_repository.GetItemAncestors query: %w", err)
	}
	defer rows.Close()

	var items []Item
	for rows.Next() {
		var it Item
		if err := rows.StructScan(&it); err != nil {
			return nil, fmt.Errorf("item_repository.GetItemAncestors scan: %w", err)
		}
		items = append(items, it)
	}
	if items == nil {
		items = []Item{}
	}
	return items, rows.Err()
}

// ListValuesByType returns distinct name/normalized pairs from item_values for a given ValueType.
type ValueRow struct {
	Value          string `db:"Value"`
	NormalizedValue string `db:"NormalizedValue"`
}

func (r *ItemRepository) ListValuesByType(valueType int) ([]ValueRow, error) {
	const q = `SELECT DISTINCT Value, NormalizedValue FROM item_values WHERE ValueType = ? ORDER BY Value`
	var rows []ValueRow
	if err := r.db.Select(&rows, q, valueType); err != nil {
		return nil, fmt.Errorf("item_repository.ListValuesByType: %w", err)
	}
	return rows, nil
}

// ListDistinctYears returns all distinct non-null ProductionYear values from base_items.
func (r *ItemRepository) ListDistinctYears() ([]int, error) {
	const q = `SELECT DISTINCT ProductionYear FROM base_items WHERE ProductionYear IS NOT NULL ORDER BY ProductionYear DESC`
	rows, err := r.db.Query(q)
	if err != nil {
		return nil, fmt.Errorf("item_repository.ListDistinctYears: %w", err)
	}
	defer rows.Close()

	var years []int
	for rows.Next() {
		var y int
		if err := rows.Scan(&y); err != nil {
			return nil, err
		}
		years = append(years, y)
	}
	if years == nil {
		years = []int{}
	}
	return years, rows.Err()
}

// ---------- query builder helpers ----------

func buildWhereClause(q ItemQuery) (string, []interface{}) {
	var conditions []string
	var args []interface{}

	// ParentId / recursive
	if q.ParentID != "" && !q.Recursive {
		conditions = append(conditions, "b.ParentId = ?")
		args = append(args, q.ParentID)
	} else if q.ParentID != "" && q.Recursive {
		conditions = append(conditions, "(b.ParentId = ? OR b.AncestorIds LIKE ?)")
		args = append(args, q.ParentID, "%"+q.ParentID+"%")
	}

	// IncludeItemTypes
	if len(q.IncludeItemTypes) > 0 {
		ph := makePlaceholders(len(q.IncludeItemTypes))
		conditions = append(conditions, fmt.Sprintf("b.Type IN (%s)", ph))
		for _, t := range q.IncludeItemTypes {
			args = append(args, t)
		}
	}

	// ExcludeItemTypes
	if len(q.ExcludeItemTypes) > 0 {
		ph := makePlaceholders(len(q.ExcludeItemTypes))
		conditions = append(conditions, fmt.Sprintf("b.Type NOT IN (%s)", ph))
		for _, t := range q.ExcludeItemTypes {
			args = append(args, t)
		}
	}

	// IsFolder
	if q.IsFolder != nil {
		if *q.IsFolder {
			conditions = append(conditions, "b.IsFolder = 1")
		} else {
			conditions = append(conditions, "b.IsFolder = 0")
		}
	}

	// SearchTerm — use LIKE (FULLTEXT requires MATCH syntax which varies)
	if q.SearchTerm != "" {
		conditions = append(conditions, "b.Name LIKE ?")
		args = append(args, "%"+q.SearchTerm+"%")
	}

	// Years
	if len(q.Years) > 0 {
		ph := makePlaceholders(len(q.Years))
		conditions = append(conditions, fmt.Sprintf("b.ProductionYear IN (%s)", ph))
		for _, y := range q.Years {
			args = append(args, y)
		}
	}

	// MediaTypes — map to Type column patterns
	if len(q.MediaTypes) > 0 {
		var mtConds []string
		for _, mt := range q.MediaTypes {
			switch strings.ToLower(mt) {
			case "video":
				mtConds = append(mtConds, "b.Type IN ('Movie','Episode','MusicVideo','Trailer','Video')")
			case "audio":
				mtConds = append(mtConds, "b.Type IN ('Audio','MusicAlbum')")
			case "photo":
				mtConds = append(mtConds, "b.Type = 'Photo'")
			case "book":
				mtConds = append(mtConds, "b.Type IN ('Book','AudioBook')")
			}
		}
		if len(mtConds) > 0 {
			conditions = append(conditions, "("+strings.Join(mtConds, " OR ")+")")
		}
	}

	// HasTmdbId / HasImdbId — check ExtraData JSON
	if q.HasTmdbId != nil {
		if *q.HasTmdbId {
			conditions = append(conditions, "JSON_EXTRACT(b.ExtraData, '$.ProviderIds.Tmdb') IS NOT NULL")
		} else {
			conditions = append(conditions, "JSON_EXTRACT(b.ExtraData, '$.ProviderIds.Tmdb') IS NULL")
		}
	}
	if q.HasImdbId != nil {
		if *q.HasImdbId {
			conditions = append(conditions, "JSON_EXTRACT(b.ExtraData, '$.ProviderIds.Imdb') IS NOT NULL")
		} else {
			conditions = append(conditions, "JSON_EXTRACT(b.ExtraData, '$.ProviderIds.Imdb') IS NULL")
		}
	}

	// Filters — require UserData join
	for _, f := range q.Filters {
		switch f {
		case "IsFavorite":
			conditions = append(conditions, "ud.IsFavorite = 1")
		case "IsPlayed":
			conditions = append(conditions, "ud.Played = 1")
		case "IsUnplayed":
			conditions = append(conditions, "(ud.Played IS NULL OR ud.Played = 0)")
		}
	}

	if len(conditions) == 0 {
		return "", args
	}
	return "WHERE " + strings.Join(conditions, " AND "), args
}

func buildOrderBy(sortBy []string, sortOrder string) string {
	if len(sortBy) == 0 {
		return "ORDER BY b.DateCreated DESC"
	}

	dir := "ASC"
	if strings.EqualFold(sortOrder, "Descending") {
		dir = "DESC"
	}

	var cols []string
	for _, s := range sortBy {
		switch s {
		case "SortName", "Name":
			cols = append(cols, "b.Name "+dir)
		case "DateCreated":
			cols = append(cols, "b.DateCreated "+dir)
		case "PremiereDate":
			cols = append(cols, "b.PremiereDate "+dir)
		case "ProductionYear":
			cols = append(cols, "b.ProductionYear "+dir)
		case "Random":
			cols = append(cols, "RAND()")
		default:
			cols = append(cols, "b.DateCreated DESC")
		}
	}
	if len(cols) == 0 {
		return "ORDER BY b.DateCreated DESC"
	}
	return "ORDER BY " + strings.Join(cols, ", ")
}

// buildValueJoin generates a JOIN fragment for filtering by item_values.
// valueType: 2=Genre, 3=Studio, 4=Artist, 5=Collection, 6=Tag
func buildValueJoin(values []string, valueType, argOffset int) (string, []interface{}) {
	if len(values) == 0 {
		return "", nil
	}
	alias := fmt.Sprintf("iv%d", valueType)
	ph := makePlaceholders(len(values))
	join := fmt.Sprintf(
		`INNER JOIN item_values %s ON %s.ItemId = b.Id AND %s.ValueType = %d AND %s.NormalizedValue IN (%s)`,
		alias, alias, alias, valueType, alias, ph,
	)
	args := make([]interface{}, len(values))
	for i, v := range values {
		args[i] = strings.ToLower(v)
	}
	return join, args
}

func makePlaceholders(n int) string {
	if n <= 0 {
		return ""
	}
	parts := make([]string, n)
	for i := range parts {
		parts[i] = "?"
	}
	return strings.Join(parts, ", ")
}

func placeholder(n int) string {
	return "?"
}
