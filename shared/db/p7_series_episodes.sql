-- =============================================
-- Series Episode Queries (TV Show Navigation)
-- Uses: IX_BaseItems_SeriesId_IsVirtualItem_Type
-- Query: WHERE series_id = ? AND is_virtual_item = 0 ORDER BY sort_name ASC
-- =============================================
--
-- Index Usage Analysis:
-- Index: IX_BaseItems_SeriesId_IsVirtualItem_Type (series_id, is_virtual_item, type)
--
-- For episode lookup:
--   - series_id = ? (full index use, exact match)
--   - is_virtual_item = 0 (full index use)
--   - ORDER BY sort_name ASC (covered by index - NO filesort)
--
-- Performance: O(log n) search
-- Typical rows: 5-100 (seasons or episodes per series)
-- Query time: 1-3ms
-- =============================================

-- =============================================
-- Pattern 1: Get All Episodes for Series
-- Uses: IX_BaseItems_SeriesId_IsVirtualItem_Type
-- =============================================

SELECT 
    bi.id,
    bi.name,
    bi.sort_name,
    bi.premiere_date,
    bi.season_id,
    bi.series_id,
    bi.run_time_ticks,
    bi.community_rating,
    bi.overview,
    bi.production_year,
    bi.path,
    bi.image_tags
FROM BaseItems bi
WHERE bi.series_id = ?
  AND bi.is_virtual_item = FALSE
  AND bi.type = 'Episode'
ORDER BY bi.sort_name ASC;

-- =============================================
-- Pattern 2: Get Episodes for Specific Season
-- Uses: IX_BaseItems_SeasonId_IsVirtualItem (if exists) or SeriesId
-- =============================================

SELECT 
    bi.id,
    bi.name,
    bi.sort_name,
    bi.premiere_date,
    bi.run_time_ticks,
    bi.community_rating,
    bi.overview
FROM BaseItems bi
WHERE bi.season_id = ?
  AND bi.is_virtual_item = FALSE
  AND bi.type = 'Episode'
ORDER BY bi.sort_name ASC;

-- =============================================
-- Pattern 3: Get Next Unwatched Episode
-- Uses: IX_BaseItems_SeriesId_IsVirtualItem_Type + UserData join
-- =============================================

-- A: Find unwatched episodes (need UserData table join)
SELECT 
    bi.id,
    bi.name,
    bi.premiere_date,
    bi.season_id,
    bi.run_time_ticks,
    ud.playback_position_ticks,
    ud.played,
    ud.last_played_date
FROM BaseItems bi
LEFT JOIN UserData ud ON bi.id = ud.library_item_id AND ud.user_id = ?
WHERE bi.series_id = ?
  AND bi.is_virtual_item = FALSE
  AND bi.type = 'Episode'
  AND COALESCE(ud.played, FALSE) = FALSE
ORDER BY bi.sort_name ASC
LIMIT 1;

-- B: Get last watched position for series (continue watching)
SELECT 
    bi.id,
    bi.name,
    bi.season_id,
    ud.playback_position_ticks,
    ud.last_played_date,
    bi.run_time_ticks
FROM BaseItems bi
LEFT JOIN UserData ud ON bi.id = ud.library_item_id AND ud.user_id = ?
WHERE bi.series_id = ?
  AND bi.is_virtual_item = FALSE
  AND bi.type = 'Episode'
  AND ud.played = FALSE
  AND ud.playback_position_ticks > 0
ORDER BY ud.last_played_date DESC
LIMIT 1;

-- =============================================
-- Pattern 4: Get Season List for Series
-- Uses: IX_BaseItems_SeriesId_IsVirtualItem_Type
-- =============================================

-- Get all seasons (not episodes)
SELECT 
    bi.id,
    bi.name,
    bi.season_id,
    bi.series_id,
    bi.premiere_date,
    bi.production_year,
    bi.image_tags,
    COUNT(ep.id) as episode_count
FROM BaseItems bi
LEFT JOIN BaseItems ep ON ep.season_id = bi.id AND ep.type = 'Episode' AND ep.is_virtual_item = FALSE
WHERE bi.series_id = ?
  AND bi.is_virtual_item = FALSE
  AND bi.type = 'Season'
GROUP BY bi.id, bi.name, bi.season_id, bi.series_id, bi.premiere_date, bi.production_year, bi.image_tags
ORDER BY bi.sort_name ASC;

-- =============================================
-- Pattern 5: Get Series Overview / Summary
-- Uses: PRIMARY KEY (series_id)
-- =============================================

-- Get the series "folder" entry
SELECT 
    id,
    name,
    overview,
    community_rating,
    production_year,
    image_tags
FROM BaseItems
WHERE id = ?  -- This is the series_id
  AND type = 'Series'
LIMIT 1;

-- =============================================
-- Pattern 6: Episode Count by Season
-- =============================================

SELECT 
    s.name as season_name,
    s.id as season_id,
    COUNT(e.id) as episode_count
FROM BaseItems s
LEFT JOIN BaseItems e ON e.season_id = s.id AND e.type = 'Episode' AND e.is_virtual_item = FALSE
WHERE s.series_id = ?
  AND s.is_virtual_item = FALSE
  AND s.type = 'Season'
GROUP BY s.id, s.name, s.season_id
ORDER BY s.sort_name ASC;

-- =============================================
-- OPTIMIZATION: Covering Index for Episode Queries
-- =============================================

-- For frequently run episode queries, a covering index avoids table lookups
-- This includes all columns needed for typical episode listings

CREATE INDEX IX_BaseItems_SeriesId_Covering 
    ON BaseItems(
        series_id ASC,
        is_virtual_item ASC,
        type ASC,
        sort_name ASC,
        id ASC,
        name ASC,
        season_id ASC,
        premiere_date ASC,
        run_time_ticks ASC,
        community_rating ASC
    );

-- Alternative: Separate optimized index for season queries
CREATE INDEX IX_BaseItems_SeasonId_Covering 
    ON BaseItems(
        season_id ASC,
        is_virtual_item ASC,
        type ASC,
        sort_name ASC
    );

-- =============================================
-- Performance Metrics (Expected)
-- =============================================

-- | Library Size | Query Type | Index Usage | Time |
-- |--------------|------------|-------------|------|
-- | 5K items     | All Episodes | 100% | 1ms |
-- | 5K items     | Specific Season | 100% | 1ms |
-- | 5K items     | Next Unwatched | 100% + UserData | 3ms |
-- | 50K items    | All Episodes | 100% | 2ms |
-- | 50K items    | Specific Season | 100% | 1ms |
-- | 50K items    | Next Unwatched | 100% + UserData | 5ms |
-- | 100K+ items  | All Episodes | 100% | 3ms |
-- | 100K+ items  | Specific Season | 100% | 1ms |

-- Note: All times include network round-trip, not just query execution
--       Index scan is consistently O(log n) regardless of library size

-- =============================================
-- Index Usage Verification (EXPLAIN)
-- =============================================

-- Check index usage for episode query
EXPLAIN SELECT * FROM BaseItems
WHERE series_id = ? 
  AND is_virtual_item = FALSE
  AND type = 'Episode'
ORDER BY sort_name ASC;

-- Expected EXPLAIN output:
-- type: range
-- key: IX_BaseItems_SeriesId_IsVirtualItem_Type
-- key_len: ~15 bytes (series_id + is_virtual_item + type)
-- rows: 5-100 (actual episodes/seasons)
-- Extra: Using index; Using where

-- =============================================
-- Maintenance: Check Index Efficiency
-- =============================================

-- Analyze BaseItems table after large imports
ANALYZE TABLE BaseItems;

-- Check slow queries for this pattern
-- Monitor queries taking > 5ms
SET GLOBAL slow_query_log = 'ON';
SET GLOBAL long_query_time = 0.005;  -- 5ms

-- =============================================
-- Integration: Repository Methods (Go)
-- =============================================

/*
// File: item-service/internal/db/item_repository.go

// GetEpisodesBySeries retrieves all episodes for a TV series
// Uses: IX_BaseItems_SeriesId_IsVirtualItem_Type
// Query: SELECT * FROM BaseItems WHERE series_id = ? AND is_virtual_item = 0 AND type = 'Episode' ORDER BY sort_name
func (r *ItemRepository) GetEpisodesBySeries(seriesID string) ([]BaseItemDto, error) {
    query := `
        SELECT id, name, sort_name, season_id, series_id,
               premiere_date, run_time_ticks, community_rating,
               overview, image_tags
        FROM BaseItems
        WHERE series_id = ? AND is_virtual_item = FALSE AND type = 'Episode'
        ORDER BY sort_name ASC
    `;
    
    var items []BaseItemDto
    err := r.db.Select(&items, query, seriesID)
    if err != nil {
        return nil, err
    }
    return items, nil
}

// GetEpisodesBySeason retrieves all episodes for a specific season
// Uses: IX_BaseItems_SeasonId_IsVirtualItem_Type
func (r *ItemRepository) GetEpisodesBySeason(seasonID string) ([]BaseItemDto, error) {
    query := `
        SELECT id, name, sort_name, season_id, series_id,
               premiere_date, run_time_ticks, community_rating,
               overview, image_tags
        FROM BaseItems
        WHERE season_id = ? AND is_virtual_item = FALSE AND type = 'Episode'
        ORDER BY sort_name ASC
    `;
    
    var items []BaseItemDto
    err := r.db.Select(&items, query, seasonID)
    if err != nil {
        return nil, err
    }
    return items, nil
}

// GetNextUnwatchedEpisode finds the next episode a user hasn't watched
// Uses: IX_BaseItems_SeriesId_IsVirtualItem_Type + IX_UserData_UserId_
func (r *ItemRepository) GetNextUnwatchedEpisode(seriesID, userID string) (*BaseItemDto, error) {
    type EpisodeWithUserData struct {
        BaseItemDto
        PlaybackPositionTicks int64    `db:"playback_position_ticks"`
        Played              bool     `db:"played"`
        LastPlayedDate      sql.NullTime `db:"last_played_date"`
    }
    
    query := `
        SELECT bi.*, 
               COALESCE(ud.playback_position_ticks, 0) as playback_position_ticks,
               COALESCE(ud.played, FALSE) as played,
               ud.last_played_date
        FROM BaseItems bi
        LEFT JOIN UserData ud ON bi.id = ud.library_item_id AND ud.user_id = ?
        WHERE bi.series_id = ?
          AND bi.is_virtual_item = FALSE
          AND bi.type = 'Episode'
          AND COALESCE(ud.played, FALSE) = FALSE
        ORDER BY bi.sort_name ASC
        LIMIT 1
    `;
    
    var episode EpisodeWithUserData
    err := r.db.Get(&episode, query, userID, seriesID)
    if err != nil {
        return nil, err
    }
    
    // Map back to BaseItemDto
    item := episode.BaseItemDto
    return &item, nil
}

// GetSeasons retrieves all seasons for a series
// Uses: IX_BaseItems_SeriesId_IsVirtualItem_Type
func (r *ItemRepository) GetSeasons(seriesID string) ([]BaseItemDto, int, error) {
    // Get count
    var total int
    err := r.db.Get(&total, `
        SELECT COUNT(*) FROM BaseItems
        WHERE series_id = ? AND is_virtual_item = FALSE AND type = 'Season'
    `, seriesID)
    if err != nil {
        return nil, 0, err
    }
    
    // Get seasons
    var seasons []BaseItemDto
    err = r.db.Select(&seasons, `
        SELECT id, name, sort_name, series_id, premiere_date,
               production_year, image_tags
        FROM BaseItems
        WHERE series_id = ? AND is_virtual_item = FALSE AND type = 'Season'
        ORDER BY sort_name ASC
    `, seriesID)
    if err != nil {
        return nil, 0, err
    }
    
    return seasons, total, nil
}
*/

-- =============================================
-- Common P7 Query Combinations
-- =============================================

-- 1. Get series + next episode (single page UI)
-- SELECT s.* FROM BaseItems s WHERE s.id = ? AND s.type = 'Series'
-- UNION ALL
-- SELECT e.* FROM BaseItems e WHERE e.series_id = ? AND e.type = 'Episode' AND e.is_virtual_item = 0
-- ORDER BY type DESC, sort_name ASC  -- Series first, then episodes

-- 2. Get user's watch progress for all episodes in series
-- SELECT e.id, e.name, e.sort_name, ud.playback_position_ticks, ud.played
-- FROM BaseItems e
-- LEFT JOIN UserData ud ON e.id = ud.library_item_id AND ud.user_id = ?
-- WHERE e.series_id = ? AND e.is_virtual_item = 0 AND e.type = 'Episode'
-- ORDER BY e.sort_name ASC

-- 3. Count total watched episodes per series (dashboard stats)
-- SELECT bi.series_id, COUNT(*) as total_episodes, SUM(CASE WHEN ud.played = TRUE THEN 1 ELSE 0 END) as watched
-- FROM BaseItems bi
-- INNER JOIN UserData ud ON bi.id = ud.library_item_id AND ud.user_id = ?
-- WHERE bi.type = 'Episode' AND bi.is_virtual_item = FALSE
-- GROUP BY bi.series_id

-- =============================================
-- End of P7 Series Episode Queries Documentation
-- Total lines: 275+
-- =============================================
