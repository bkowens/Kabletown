-- ==============================================
-- User Favorites Queries
-- Uses: IX_UserData_UserId_IsFavorite
-- Query: WHERE user_id = ? AND is_favorite = 1 ORDER BY last_played_date DESC
-- ==============================================
--
-- Index Usage Analysis:
-- Index: IX_UserData_UserId_IsFavorite (user_id, is_favorite, last_played_date)
--
-- For favorites query:
--   - user_id = ? (full index use, exact match)
--   - is_favorite = 1 (full index use)
--   - ORDER BY last_played_date DESC (covered by index - NO filesort)
--
-- Performance: O(log n) search
-- Typical rows: 5-500 (user preference dependent)
-- Query time: 1-5ms
-- ==============================================

-- ==============================================
-- Pattern 1: Get User's Favorite Items
-- Uses: IX_UserData_UserId_IsFavorite
-- ==============================================

SELECT 
    ud.user_id,
    ud.library_item_id,
    ud.is_favorite,
    ud.play_count,
    ud.playback_position_ticks,
    ud.last_played_date,
    ud.rating,
    bi.name,
    bi.type,
    bi.production_year,
    bi.community_rating,
    bi.image_tags,
    bi.run_time_ticks
FROM UserData ud
INNER JOIN BaseItems bi ON ud.library_item_id = bi.id
WHERE ud.user_id = ?
  AND ud.is_favorite = TRUE
ORDER BY ud.last_played_date DESC;

-- ==============================================
-- Pattern 2: Favorite Items with Pagination
-- Uses: IX_UserData_UserId_IsFavorite
-- ==============================================

-- Get paginated favorites with total count

-- Query 1: Get total count
SELECT COUNT(*) as total_count
FROM UserData
WHERE user_id = ?
  AND is_favorite = TRUE;

-- Query 2: Get paginated favorites
-- Uses: IX_UserData_UserId_IsFavorite (full index coverage)
SELECT 
    ud.library_item_id,
    ud.is_favorite,
    ud.last_played_date,
    ud.play_count,
    bi.id,
    bi.name,
    bi.type,
    bi.production_year,
    bi.image_tags
FROM UserData ud
INNER JOIN BaseItems bi ON ud.library_item_id = bi.id
WHERE ud.user_id = ?
  AND ud.is_favorite = TRUE
ORDER BY ud.last_played_date DESC
LIMIT 20 OFFSET 0;

-- ==============================================
-- Pattern 3: Favorite Items by Type (Filter by Movie, Show, etc.)
-- Uses: IX_UserData_UserId_IsFavorite + IX_BaseItems_Type
-- ==============================================

SELECT 
    ud.library_item_id,
    ud.last_played_date,
    bi.id,
    bi.name,
    bi.type,
    bi.production_year,
    bi.image_tags,
    bi.premiere_date
FROM UserData ud
INNER JOIN BaseItems bi ON ud.library_item_id = bi.id
WHERE ud.user_id = ?
  AND ud.is_favorite = TRUE
  AND bi.type IN ('Movie', 'Series')
ORDER BY ud.last_played_date DESC
LIMIT 50;

-- ==============================================
-- Pattern 4: Favorite Items Count by Type
-- ==============================================

-- Use this for dashboard stats showing "Favorites: Movies (12), Shows (8)"

SELECT 
    bi.type,
    COUNT(*) as favorite_count
FROM UserData ud
INNER JOIN BaseItems bi ON ud.library_item_id = bi.id
WHERE ud.user_id = ?
  AND ud.is_favorite = TRUE
GROUP BY bi.type
ORDER BY favorite_count DESC;

-- ==============================================
-- Pattern 5: Recent + Favorite Items (Combined Home Screen)
-- Uses: IX_UserData_UserId_IsFavorite + IX_BaseItems_Type_DateCreated
-- ==============================================

-- Get most recent favorites for "Your Favorites" section

SELECT 
    bi.id,
    bi.name,
    bi.type,
    bi.production_year,
    bi.image_tags,
    ud.last_played_date,
    ud.is_favorite
FROM UserData ud
INNER JOIN BaseItems bi ON ud.library_item_id = bi.id
WHERE ud.user_id = ?
  AND ud.is_favorite = TRUE
ORDER BY ud.last_played_date DESC
LIMIT 20;

-- ==============================================
-- Pattern 6: User's Top 10 Favorites by Play Count
-- ==============================================

-- Show the user's most-watched favorites

SELECT 
    bi.id,
    bi.name,
    bi.type,
    bi.production_year,
    ud.play_count,
    ud.last_played_date,
    bi.image_tags
FROM UserData ud
INNER JOIN BaseItems bi ON ud.library_item_id = bi.id
WHERE ud.user_id = ?
  AND ud.is_favorite = TRUE
ORDER BY ud.play_count DESC
LIMIT 10;

-- ==============================================
-- Pattern 7: Toggle Favorite (Add/Remove)
-- Uses: PRIMARY KEY on (user_id, library_item_id)
-- ==============================================

-- A: Add favorite (upsert)
INSERT INTO UserData (user_id, library_item_id, is_favorite, play_count, row_version, date_created, date_modified)
VALUES (?, ?, TRUE, 0, 0, NOW(6), NOW(6))
ON DUPLICATE KEY UPDATE 
    is_favorite = TRUE,
    row_version = row_version + 1;

-- B: Remove favorite (update)
UPDATE UserData
SET is_favorite = FALSE,
    row_version = row_version + 1,
    date_modified = NOW(6)
WHERE user_id = ?
  AND library_item_id = ?;

-- C: Check if item is favorite (fast lookup)
SELECT is_favorite
FROM UserData
WHERE user_id = ?
  AND library_item_id = ?
LIMIT 1;

-- ==============================================
-- Pattern 8: Get User's Favorite Collection Names
-- ==============================================

-- For dashboard: "You have 5 favorite movie playlists"

SELECT DISTINCT
    c.id,
    c.name,
    c.type,
    COUNT(CASE WHEN ud.is_favorite = TRUE THEN 1 ELSE NULL END) as favorite_count
FROM BaseItems bi
INNER JOIN UserData ud ON bi.id = ud.library_item_id
LEFT JOIN Collections c ON bi.parent_id = c.id
WHERE ud.user_id = ?
  AND ud.is_favorite = TRUE
GROUP BY c.id, c.name, c.type
HAVING favorite_count > 0;

-- ==============================================
-- OPTIMIZATION: Covering Index for Favorites Queries
-- ==============================================

-- For very large libraries with many favorites, a covering index
-- can eliminate table lookups entirely

CREATE INDEX IX_UserData_Favorites_Covering 
    ON UserData(
        user_id ASC,
        is_favorite ASC,
        last_played_date DESC,
        library_item_id ASC,
        play_count ASC,
        playback_position_ticks ASC
    );

-- ==============================================
-- Performance Metrics (Expected)
-- ==============================================

-- | Library Size | Favorites | Index Usage | Time |
-- |--------------|-----------|-------------|------|
-- | 5K items     | 50 | 100% | 1ms |
-- | 5K items     | 200 | 100% | 2ms |
-- | 50K items    | 100 | 100% | 2ms |
-- | 50K items    | 500 | 100% | 3ms |
-- | 100K+ items  | 200 | 100% | 3ms |
-- | 100K+ items  | 1000 | 100% | 5ms |

-- Note: All times include network round-trip
--       Index scan is consistently O(log n) regardless of library size

-- ==============================================
-- Index Usage Verification (EXPLAIN)
-- ==============================================

-- Check index usage for favorites query
EXPLAIN SELECT * FROM UserData ud
INNER JOIN BaseItems bi ON ud.library_item_id = bi.id
WHERE ud.user_id = ?
  AND ud.is_favorite = TRUE
ORDER BY ud.last_played_date DESC;

-- Expected EXPLAIN output:
-- type: index_merge or range
-- key: IX_UserData_UserId_IsFavorite
-- key_len: ~6-10 bytes
-- rows: 5-500 (actual favorites count)
-- Extra: Using index; Using where

-- ==============================================
-- Integration: Repository Methods (Go)
-- ==============================================

/*
// File: item-service/internal/db/user_data_repository.go

// GetUserFavorites retrieves user's favorite items
// Uses: IX_UserData_UserId_IsFavorite
func (r *UserDataRepository) GetUserFavorites(userID string, startIndex, limit int) ([]FavoriteItem, int, error) {
    // Get total count
    var total int
    err := r.db.Get(&total, `
        SELECT COUNT(*) FROM UserData
        WHERE user_id = ? AND is_favorite = TRUE
    `, userID)
    if err != nil {
        return nil, 0, err
    }

    // Get paginated favorites
    var items []FavoriteItem
    err = r.db.Select(&items, `
        SELECT 
            ud.library_item_id,
            ud.is_favorite,
            ud.play_count,
            ud.playback_position_ticks,
            ud.last_played_date,
            bi.id,
            bi.name,
            bi.type,
            bi.production_year,
            bi.image_tags
        FROM UserData ud
        INNER JOIN BaseItems bi ON ud.library_item_id = bi.id
        WHERE ud.user_id = ?
          AND ud.is_favorite = TRUE
        ORDER BY ud.last_played_date DESC
        LIMIT ? OFFSET ?
    `, userID, limit, startIndex)
    if err != nil {
        return nil, 0, err
    }

    return items, total, nil
}

// ToggleFavorite toggles favorite status for an item
// Uses: PRIMARY KEY on (user_id, library_item_id)
func (r *UserDataRepository) ToggleFavorite(userID, itemID string, makeFavorite bool) error {
    _, err := r.db.Exec(`
        INSERT INTO UserData (user_id, library_item_id, is_favorite, row_version, date_created, date_modified)
        VALUES (?, ?, ?, 0, NOW(6), NOW(6))
        ON DUPLICATE KEY UPDATE 
            is_favorite = ?,
            row_version = row_version + 1,
            date_modified = NOW(6)
    `, userID, itemID, makeFavorite, makeFavorite)
    return err
}

// IsFavorite checks if item is marked as favorite by user
// Uses: PRIMARY KEY lookup
func (r *UserDataRepository) IsFavorite(userID, itemID string) (bool, error) {
    var isFavorite bool
    err := r.db.Get(&isFavorite, `
        SELECT is_favorite FROM UserData
        WHERE user_id = ? AND library_item_id = ?
        LIMIT 1
    `, userID, itemID)
    if err == sql.ErrNoRows {
        return false, nil
    }
    if err != nil {
        return false, err
    }
    return isFavorite, nil
}

// CountFavoritesByType returns favorite counts per media type
func (r *UserDataRepository) CountFavoritesByType(userID string) (map[string]int, error) {
    rows, err := r.db.Query(`
        SELECT bi.type, COUNT(*) as count
        FROM UserData ud
        INNER JOIN BaseItems bi ON ud.library_item_id = bi.id
        WHERE ud.user_id = ?
          AND ud.is_favorite = TRUE
        GROUP BY bi.type
        ORDER BY count DESC
    `, userID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    counts := make(map[string]int)
    for rows.Next() {
        var mediaType string
        var count int
        if err := rows.Scan(&mediaType, &count); err != nil {
            return nil, err
        }
        counts[mediaType] = count
    }
    return counts, nil
}
*/

-- ==============================================
-- Common Query Combinations
-- ==============================================

-- 1. Home Screen: Recent + Featured + Favorites
-- SELECT * FROM BaseItems WHERE series_id = ? ... (GetNextUnwatchedEpisode)
-- UNION ALL
-- SELECT * FROM UserData WHERE user_id = ? AND is_favorite = 1 ORDER BY last_played_date DESC LIMIT 5
-- UNION ALL
-- SELECT * FROM BaseItems WHERE production_year = YEAR(NOW()) ORDER BY community_rating DESC LIMIT 5

-- 2. "Continue Watching" + "Your Favorites" in single dashboard
-- SELECT 'continue_watching' as section, id, name, type, image_tags, playback_position_ticks
-- FROM BaseItems WHERE id IN (SELECT library_item_id FROM UserData WHERE user_id = ? AND played = FALSE AND play_count > 0)
-- UNION ALL
-- SELECT 'favorites' as section, bi.id, bi.name, bi.type, bi.image_tags, 0 as playback_position_ticks
-- FROM BaseItems bi
-- INNER JOIN UserData ud ON bi.library_item_id = bi.id
-- WHERE ud.user_id = ? AND ud.is_favorite = TRUE
-- ORDER BY section DESC, last_played_date DESC

-- ==============================================
-- Maintenance: Index Statistics
-- ==============================================

-- Analyze UserData table after large bulk updates
ANALYZE TABLE UserData;

-- Check for missing indexes affecting favorites queries
SELECT * FROM sys.schema_missing_indexes 
WHERE table_name = 'UserData' AND table_schema = 'jellyfin';

-- ==============================================
-- End of P7 User Favorites Queries Documentation
-- Total lines: 380+ (including Go integration code)
-- ==============================================
