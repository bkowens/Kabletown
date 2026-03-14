// Package db provides database access for search-service.
package db

import (
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

// SearchHint is a single search result row from the base_items table.
type SearchHint struct {
	Id             string  `db:"Id"`
	Name           string  `db:"Name"`
	Type           string  `db:"Type"`
	IsFolder       bool    `db:"IsFolder"`
	ParentId       *string `db:"ParentId"`
	TopParentId    *string `db:"TopParentId"`
	DurationTicks  *int64  `db:"DurationTicks"`
	ProductionYear *int    `db:"ProductionYear"`
}

// SearchRepository provides search queries against the base_items table.
type SearchRepository struct {
	db *sqlx.DB
}

// NewSearchRepository creates a SearchRepository backed by the given connection.
func NewSearchRepository(database *sqlx.DB) *SearchRepository {
	return &SearchRepository{db: database}
}

// Search returns items whose Name matches the search term, optionally filtered
// by item types. Returns the matching rows and total count (before LIMIT/OFFSET).
func (r *SearchRepository) Search(term string, includeTypes []string, limit, offset int) ([]SearchHint, int, error) {
	// Build the optional type filter clause.
	typeFilter := ""
	args := []interface{}{term}

	if len(includeTypes) > 0 {
		placeholders := make([]string, len(includeTypes))
		for i, t := range includeTypes {
			placeholders[i] = "?"
			args = append(args, t)
		}
		typeFilter = fmt.Sprintf(" AND Type IN (%s)", strings.Join(placeholders, ","))
	}

	countQuery := `SELECT COUNT(*) FROM base_items WHERE Name LIKE CONCAT('%', ?, '%')` + typeFilter
	var total int
	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("search_repository.Search count: %w", err)
	}

	selectQuery := `
		SELECT Id, Name, Type, IsFolder, ParentId, TopParentId, DurationTicks, ProductionYear
		FROM base_items
		WHERE Name LIKE CONCAT('%', ?, '%')` + typeFilter + `
		ORDER BY Name ASC
		LIMIT ? OFFSET ?`

	queryArgs := append(args, limit, offset)
	rows, err := r.db.Queryx(selectQuery, queryArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("search_repository.Search query: %w", err)
	}
	defer rows.Close()

	var results []SearchHint
	for rows.Next() {
		var hint SearchHint
		if err := rows.StructScan(&hint); err != nil {
			return nil, 0, fmt.Errorf("search_repository.Search scan: %w", err)
		}
		results = append(results, hint)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("search_repository.Search rows: %w", err)
	}

	return results, total, nil
}
