package db

import (
	"github.com/jmoiron/sqlx"
)

// ItemRepository handles BaseItems CRUD operations with P7 index optimizations
type ItemRepository struct {
	db *sqlx.DB
}

// NewItemRepository creates a new ItemRepository
func NewItemRepository(db *sqlx.DB) *ItemRepository {
	return &ItemRepository{db: db}
}

// GetByParentId retrieves items in a folder (uses IX_BaseItems_ParentId_IsVirtualItem_Type)
func (r *ItemRepository) GetByParentId(parentID string, startIndex, limit int) ([]BaseItemDto, int, error) {
	// Get total count for pagination
	var total int
	err := r.db.Get(&total, `
		SELECT COUNT(*) FROM BaseItems
		WHERE parent_id = ? AND is_virtual_item = FALSE
	`, parentID)
	if err != nil {
		return nil, 0, err
	}

	// Get paginated items (uses IX_BaseItems_ParentId_IsVirtualItem_Type)
	var items []BaseItemDto
	err = r.db.Select(&items, `
		SELECT 
			id, name, sort_name, type, media_type,
			parent_id, season_id, series_id,
			is_folder, is_virtual_item,
			overview, tagline, official_rating,
			production_year, premiere_date,
			community_rating, run_time_ticks,
			can_delete, can_download, location_type,
			date_created, date_last_saved,
			image_tags
		FROM BaseItems
		WHERE parent_id = ? AND is_virtual_item = FALSE
		ORDER BY sort_name ASC
		LIMIT ? OFFSET ?
	`, parentID, limit, startIndex)
	if err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

// GetById retrieves a single item by ID (uses PRIMARY KEY)
func (r *ItemRepository) GetById(id string) (*BaseItemDto, error) {
	item := &BaseItemDto{}
	err := r.db.Get(item, `
		SELECT 
			id, name, sort_name, type, media_type,
			parent_id, season_id, series_id,
			is_folder, is_virtual_item,
			overview, tagline, official_rating,
			production_year, premiere_date,
			community_rating, run_time_ticks,
			can_delete, can_download, location_type,
			path,
			date_created, date_last_saved,
			image_tags
		FROM BaseItems
		WHERE id = ?
	`, id)
	if err != nil {
		return nil, err
	}
	return item, nil
}

// GetRecentlyAddedByType retrieves recently added items (uses IX_BaseItems_Type_IsVirtualItem_DateCreated)
func (r *ItemRepository) GetRecentlyAddedByType(itemType string, limit int) ([]BaseItemDto, error) {
	var items []BaseItemDto
	err := r.db.Select(&items, `
		SELECT 
			id, name, sort_name, type, media_type,
			overview, overview,
			production_year, community_rating, image_tags,
			date_created
		FROM BaseItems
		WHERE type = ? AND is_virtual_item = FALSE
		ORDER BY date_created DESC
		LIMIT ?
	`, itemType, limit)
	if err != nil {
		return nil, err
	}
	return items, nil
}

// GetEpisodesBySeries retrieves all episodes for a series (uses IX_BaseItems_SeriesId_IsVirtualItem)
func (r *ItemRepository) GetEpisodesBySeries(seriesID string, userID *string) ([]BaseItemDto, error) {
	query := `
		SELECT 
			bi.*
		FROM BaseItems bi
		WHERE bi.series_id = ? 
			AND bi.is_virtual_item = FALSE
			AND bi.type = 'Episode'
		ORDER BY bi.sort_name ASC
	`;

	var items []BaseItemDto
	err := r.db.Select(&items, query, seriesID)
	if err != nil {
		return nil, err
	}
	return items, nil
}

// FilterByGenre retrieves items with specific genres (uses IX_ItemValuesMap_ItemValueId)
func (r *ItemRepository) FilterByGenre(genreIDs []string, itemType string, startIndex, limit int) ([]BaseItemDto, int, error) {
	if len(genreIDs) == 0 {
		return nil, 0, nil
	}

	// Get total count
	var total int
	placeholders := make([]string, len(genreIDs))
	args := make([]interface{}, len(genreIDs)+2)
	for i, id := range genreIDs {
		placeholders[i] = "?"
		args[i] = id
	}
	args[len(genreIDs)] = itemType
	args[len(genreIDs)+1] = false

	err := r.db.Get(&total, `
		SELECT COUNT(DISTINCT bi.id) 
		FROM BaseItems bi
		INNER JOIN ItemValuesMap ivm ON bi.id = ivm.library_item_id
		WHERE ivm.item_value_id IN (`+joinPlaceholders(placeholders)+`) 
			AND bi.type = ? 
			AND bi.is_virtual_item = ?
	`, args...)
	if err != nil {
		return nil, 0, err
	}

	// Get paginated items
	args[len(genreIDs)] = itemType
	args[len(genreIDs)+1] = false
	args = append(args, limit, startIndex)

	var items []BaseItemDto
	err = r.db.Select(&items, `
		SELECT DISTINCT b.*
		FROM BaseItems b
		INNER JOIN ItemValuesMap m ON b.id = m.library_item_id
		WHERE m.item_value_id IN (`+joinPlaceholders(placeholders)+`) 
			AND b.type = ? 
			AND b.is_virtual_item = ?
		ORDER BY b.sort_name ASC
		LIMIT ? OFFSET ?
	`, args...)
	if err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

// Helper function to join placeholders
func joinPlaceholders(placeholders []string) string {
	result := ""
	for i := range placeholders {
		if i > 0 {
			result += ","
		}
		result += placeholders[i]
	}
	return result
}

// EnsureBaseItemsTables creates the BaseItems table with P7 indexes
func EnsureBaseItemsTables(db *sqlx.DB) error {
	createSQL := `
		CREATE TABLE IF NOT EXISTS BaseItems (
			id CHAR(36) PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			sort_name VARCHAR(255) NOT NULL,
			original_title VARCHAR(255),
			parent_id CHAR(36),
			season_id CHAR(36),
			series_id CHAR(36),
			type VARCHAR(50) NOT NULL,
			media_type VARCHAR(50),
			is_folder BOOLEAN NOT NULL DEFAULT FALSE,
			is_virtual_item BOOLEAN NOT NULL DEFAULT FALSE,
			path VARCHAR(1024),
			overview TEXT,
			tagline VARCHAR(512),
			production_year INT,
			premiere_date DATETIME(7),
			community_rating FLOAT,
			official_rating VARCHAR(64),
			run_time_ticks BIGINT NOT NULL DEFAULT 0,
			can_delete BOOLEAN NOT NULL DEFAULT FALSE,
			can_download BOOLEAN NOT NULL DEFAULT FALSE,
			location_type VARCHAR(50),
			image_tags TEXT,
			date_created DATETIME(7) NOT NULL,
			date_last_saved DATETIME(7) NOT NULL,
			row_version INT UNSIGNED NOT NULL DEFAULT 0,
			
			-- P7 Indexes matching C# migration: 20260309000000_AddPerformanceIndexes.cs
			INDEX IX_BaseItems_Type_IsVirtualItem_SortName (type, is_virtual_item, sort_name),
			INDEX IX_BaseItems_ParentId_IsVirtualItem_Type (parent_id, is_virtual_item, type),
			INDEX IX_BaseItems_User_Id (user_id),
			INDEX IX_BaseItems_Name (name(100)),
			INDEX IX_BaseItems_SeriesId_IsVirtualItem (series_id, is_virtual_item, type),
			INDEX IX_BaseItems_SeasonId_IsVirtualItem (season_id, is_virtual_item)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
	`;

	_, err := db.Exec(createSQL)
	return err
}
