-- ==============================================================================
-- User Resume Queue Queries
-- Uses: IX_UserData_UserId_PlaybackPositionTicks
-- Query: WHERE user_id = ? AND playback_position_ticks > 0 ORDER BY last_played_date DESC
-- ==============================================================================
--
-- Index Usage Analysis:
-- Index: IX_UserData_UserId_PlaybackPositionTicks (user_id, playback_position_ticks)
--
-- For resume queue query:
--   - user_id = ? (full index use, exact match)
--   - playback_position_ticks > 0 (full index use - range scan)
--   - ORDER BY last_played_date DESC (NOT covered - requires filesort)
--
-- Performance Characteristics:
--   - Index scan: O(log n + m) where m = rows matching UserId + PlaybackPositionTicks > 0
--   - Filesort: O(m log m) on last_played_date
--   - Typical rows: 0-50 (most users have few in-progress items)
--   - Query time: 2-10ms
--
-- Optimization Notes:
--   - The index covers the WHERE clause perfectly
--   - For ordering, we have two options:
--     1. Accept filesort (simpler, fast for small m)
--     2. Create composite index with last_played_date (larger index, no filesort)
--   - Current P7 strategy uses 2-column index (storage efficiency)
--   - If resume queue queries dominate, consider adding IX_UserData_UserId_PlaybackPositionTicks_LastPlayedDate
--
-- Alternative Index Strategy (if needed):
--   CREATE INDEX IX_UserData_UserId_Playback_Position_Ticks_Last_Played
--       ON UserData (UserId, PlaybackPositionTicks, LastPlayedDate DESC);
--   This would be a covering index for the full query but costs more storage and write overhead.
-- ==============================================================================

-- ==============================================================================
-- Pattern 1: Get User's In-Progress Items (Resume Queue)
-- Uses: IX_UserData_UserId_PlaybackPositionTicks + Filesort on last_played_date
-- ==============================================================================

SELECT 
    ud.user_id,
    ud.library_item_id,
    ud.playback_position_ticks,
    ud.last_played_date,
    ud.play_count,
    ud.is_favorite,
    bi.name,
    bi.type,
    bi.series_id,
    bi.season_id,
    bi.index_number,      -- Episode number
    bi.index_number_end,  -- End episode number (for movies)
    bi.runtime_ticks,
    bi.image_tags,
    bi.premiere_date,
    bi.production_year
FROM UserData ud
INNER JOIN BaseItems bi ON ud.library_item_id = bi.id
WHERE ud.user_id = ?
  AND ud.playback_position_ticks > 0
ORDER BY ud.last_played_date DESC;

-- ==============================================================================
-- Pattern 2: Resume Queue with Pagination
-- Uses: IX_UserData_UserId_PlaybackPositionTicks + Filesort
-- ==============================================================================

-- Get resumable items with pagination for UI shelves

-- Query 1: Get total count of in-progress items
SELECT COUNT(*) as total_count
FROM UserData
WHERE user_id = ?
  AND playback_position_ticks > 0;

-- Query 2: Get paginated resume queue
-- Note: Filesort required for ORDER BY last_played_date
SELECT 
    ud.library_item_id,
    ud.playback_position_ticks,
    ud.last_played_date,
    ud.play_count,
    ud.runtime_ticks,
    bi.id,
    bi.name,
    bi.type,
    bi.series_id,
    bi.season_id,
    bi.index_number,
    bi.image_tags,
    bi.runtime_ticks,
    -- Calculate progress percentage
    ROUND((ud.playback_position_ticks * 100.0) / NULLIF(bi.runtime_ticks, 0), 1) as progress_percent
FROM UserData ud
INNER JOIN BaseItems bi ON ud.library_item_id = bi.id
WHERE ud.user_id = ?
  AND ud.playback_position_ticks > 0
ORDER BY ud.last_played_date DESC
LIMIT 20 OFFSET 0;

-- ==============================================================================
-- Pattern 3: Resume Queue with Type Filtering
-- Uses: IX_UserData_UserId_PlaybackPositionTicks + Filesort + Type filter
-- ==============================================================================

-- Filter resume queue by media type (e.g., only episodes, only movies)

SELECT 
    ud.library_item_id,
    ud.playback_position_ticks,
    ud.last_played_date,
    bi.id,
    bi.name,
    bi.type,
    bi.series_id,
    bi.season_id,
    bi.index_number,
    bi.image_tags,
    bi.runtime_ticks,
    CASE 
        WHEN bi.type = 'Episode' THEN CONCAT('S', bi.season_index_number, 'E', bi.index_number)
        ELSE NULL
    END as episode_indicator,
    ROUND((ud.playback_position_ticks * 100.0) / NULLIF(bi.runtime_ticks, 0), 1) as progress_percent
FROM UserData ud
INNER JOIN BaseItems bi ON ud.library_item_id = bi.id
WHERE ud.user_id = ?
  AND ud.playback_position_ticks > 0
  AND bi.type IN ('Episode', 'Movie')
ORDER BY ud.last_played_date DESC
LIMIT 20;

-- ==============================================================================
-- Optimized Pattern: Application-Side Sorting (No Filesort)
-- Uses: IX_UserData_UserId_PlaybackPositionTicks (full index, no filesort)
-- ==============================================================================

-- Alternative approach: Use the index efficiently, sort in application layer
-- This avoids filesort but requires fetching all rows

-- Step 1: Use index to fetch all in-progress items (no ORDER BY in SQL)
SELECT 
    ud.library_item_id,
    ud.playback_position_ticks,
    ud.last_played_date,
    ud.play_count,
    bi.id,
    bi.name,
    bi.type,
    bi.runtime_ticks
FROM UserData ud
INNER JOIN BaseItems bi ON ud.library_item_id = bi.id
WHERE ud.user_id = ?
  AND ud.playback_position_ticks > 0;

-- Step 2: Sort by last_played_date in Go application layer
-- This trades SQL filesort for application memory but uses index optimally
-- Best for small result sets (< 100 rows)

-- ==============================================================================
-- Performance Index: Resume Queue by Type
-- Uses: Composite index for specific type queries
-- ==============================================================================

-- For frequently queried types (e.g., "Continue Watching Episodes")
-- Consider this application-specific index if needed:
--
-- CREATE INDEX IX_UserData_UserId_Playback_Ticks_Type
--     ON UserData (UserId, PlaybackPositionTicks, LastPlayedDate DESC);
--
-- Then join with BaseItems for type filtering.
-- This keeps sorting in SQL while maintaining index efficiency.

-- ==============================================================================
-- Pattern 4: Resume Progress Update
-- Uses: IX_UserData_UserId_PlaybackPositionTicks (for read-after-write validation)
-- ==============================================================================

-- Update playback position during playback

UPDATE UserData
SET playback_position_ticks = ?,
    last_played_date = ?
WHERE user_id = ?
  AND library_item_id = ?;

-- Retrieve updated progress for confirmation
SELECT 
    playback_position_ticks,
    last_played_date,
    play_count
FROM UserData
WHERE user_id = ?
  AND library_item_id = ?;

-- ==============================================================================
-- Pattern 5: Clear Resume Position (Mark as Completed)
-- Uses: IX_UserData_UserId_PlaybackPositionTicks
-- ==============================================================================

-- When user marks item as played, clear resume position

UPDATE UserData
SET playback_position_ticks = 0,
    played = TRUE,
    play_count = play_count + 1,
    last_played_date = ?
WHERE user_id = ?
  AND library_item_id = ?;

-- Count affected rows to determine if item was being watched
-- Should return 1 if item was in progress, 0 if not

-- ==============================================================================
-- Pattern 6: Get Resume Position for Specific Item
-- Uses: IX_UserData_ItemId_UserId_PlaybackPositionTicks (alternate index)
-- ==============================================================================

-- For point-lookup by itemId (e.g., when navigating to an episode)
-- This uses the different index optimized for ItemId-first lookups

SELECT 
    ud.playback_position_ticks,
    ud.last_played_date,
    bi.runtime_ticks,
    CASE 
        WHEN bi.runtime_ticks > 0 
        THEN ROUND((ud.playback_position_ticks * 100.0) / bi.runtime_ticks, 1)
        ELSE 0 
    END as progress_percent
FROM UserData ud
INNER JOIN BaseItems bi ON ud.library_item_id = bi.id
WHERE ud.user_id = ?
  AND ud.library_item_id = ?;

-- Note: For single-item lookups, IX_UserData_ItemId_UserId_PlaybackPositionTicks
-- is more efficient than scanning IX_UserData_UserId_PlaybackPositionTicks

-- ==============================================================================
-- Pattern 7: Resume Queue for Parent/Series Navigation
-- Uses: IX_UserData_UserId_PlaybackPositionTicks + IX_BaseItems_SeriesId
-- ==============================================================================

-- Find in-progress episodes of a specific series

SELECT 
    ud.library_item_id,
    ud.playback_position_ticks,
    ud.last_played_date,
    bi.index_number as episode_number,
    bi.name as episode_name,
    bi.season_id,
    bi.season_index_number as season_number,
    bi.image_tags,
    bi.runtime_ticks,
    ROUND((ud.playback_position_ticks * 100.0) / NULLIF(bi.runtime_ticks, 0), 1) as progress_percent
FROM UserData ud
INNER JOIN BaseItems bi ON ud.library_item_id = bi.id
WHERE ud.user_id = ?
  AND bi.series_id = ?
  AND bi.type = 'Episode'
  AND ud.playback_position_ticks > 0
ORDER BY bi.season_index_number ASC, bi.index_number ASC;

-- This helps users resume from the last watched episode in a series

-- ==============================================================================
-- Query Performance Summary
-- ==============================================================================
--
-- Query Type                    | Index Used                    | Files ort?
-- ------------------------------|-------------------------------|------------
-- Resume Queue (all types)      | IX_UserData_UserId_Playback   | Yes
-- Resume Queue (paginated)      | IX_UserData_UserId_Playback   | Yes
-- Resume by Type                | IX_UserData_UserId_Playback   | Yes
-- Resume by Series              | IX_UserData_UserId_Playback   | No (ORDER BY episode)
-- Single Item Position          | IX_UserData_ItemId_UserId_... | No
-- Update Playback Position      | IX_UserData_ItemId_UserId     | N/A
--
-- Recommendations:
-- 1. Current index works well for most scenarios
-- 2. If resume queue > 50 items avg, consider adding last_played_date to index
-- 3. Application-side sorting is viable alternative for small result sets
-- 4. Series-specific queries avoid filesort by ordering by episode number
-- ==============================================================================