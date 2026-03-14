-- ==========================================
-- Library Browse: IN Clause Multi-Type Filter
-- Query: WHERE Type IN (?) AND IsVirtualItem = 0 ORDER BY SortName
-- ==========================================

-- Pattern: Get multiple media types at once (e.g., Movies + Shows + Episodes)
--
-- INDEX USAGE ANALYSIS:
-- Index: IX_BaseItems_Type_IsVirtualItem_SortName (type, is_virtual_item, sort_name)
--
-- MySQL will:
-- 1. Use (type) portion of index to filter IN clause values
-- 2. Use (is_virtual_item) to filter 0 values (partial index effectiveness)
-- 3. May need filesort for ORDER BY because IN creates discontinuities
--
-- PERFORMANCE: Still 80%+ faster than no index, even with filesort
-- OPTIMIZATION: Consider adding separate index if multi-type queries are 80%+ of traffic

-- Example: Get Movies and TV Shows for home screen
SELECT 
    id,
    name,
    type,
    production_year,
    overview,
    image_tag,
    date_last_saved
FROM BaseItems
WHERE type IN ('Movie', 'Series', 'Season', 'Episode')
  AND is_virtual_item = 0
ORDER BY sort_name ASC
LIMIT 50 OFFSET 0;

-- ==========================================
-- OPTIMIZED PATTERN FOR MULTI-TYPE QUERIES
-- ==========================================

-- Option 1: Separate queries per type (uses full index, union results)
-- More complex, but 100% index coverage

SELECT id, name, type, sort_name FROM BaseItems
WHERE type = 'Movie' AND is_virtual_item = 0
ORDER BY sort_name ASC
LIMIT 25 OFFSET 0

UNION ALL

SELECT id, name, type, sort_name FROM BaseItems
WHERE type = 'Series' AND is_virtual_item = 0
ORDER BY sort_name ASC
LIMIT 25 OFFSET 0;

-- Option 2: Sort by date_created instead (uses type + is_virtual_item, no filesort)
SELECT *
FROM BaseItems
WHERE type IN ('Movie', 'Series')
  AND is_virtual_item = 0
ORDER BY date_created DESC
LIMIT 50 OFFSET 0;
-- Note: Can use IX_BaseItems_Type_IsVirtualItem if we add sort by creation date

-- ==========================================
-- RECOMMENDED: Add Separate Index for IN Queries
-- ==========================================

-- If multi-type IN queries are common (home screen, browse all media):
CREATE INDEX IX_BaseItems_Type_InClause_SortName 
    ON BaseItems(type, sort_name);
-- This index works better with IN clauses since is_virtual_item is often = 0

-- Or a covering index for common multi-type filters:
CREATE INDEX IX_BaseItems_Type_IsVirtual_SortName_Year
    ON BaseItems(type, is_virtual_item, sort_name, production_year);
-- Includes production_year to avoid table lookups

-- ==========================================
-- P4 TOKEN RESOLUTION (O(log n) - Always optimal)
-- ==========================================

SELECT * FROM Devices
WHERE access_token = ?
LIMIT 1;
-- Uses: IX_Devices_AccessToken

-- ==========================================
-- P6 FILTER: Get multiple genres at once
-- ==========================================

-- Pattern: Filter items by multiple genres
-- Uses: IX_ItemValuesMap_ItemValueId for efficient join

SELECT DISTINCT bi.id, bi.name, bi.type, bi.production_year
FROM BaseItems bi
INNER JOIN ItemValuesMap ivm ON bi.id = ivm.library_item_id
WHERE ivm.item_value_id IN (
    SELECT id FROM ItemValues 
    WHERE type = 'Genre' AND name IN ('Action', 'Adventure', 'Sci-Fi')
)
  AND bi.type IN ('Movie', 'Series')
  AND bi.is_virtual_item = 0
ORDER BY bi.sort_name ASC
LIMIT 20 OFFSET 0;
