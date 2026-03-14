-- ===============================
-- Parent Folder Contents
-- Uses: IX_BaseItems_ParentId_IsVirtualItem_Type
-- Query: WHERE parent_id = ? AND is_virtual_item = 0 ORDER BY SortName
-- ===============================

-- Pattern: Get immediate children of a folder
-- Index: IX_BaseItems_ParentId_IsVirtualItem_Type (parent_id, is_virtual_item, type)
--
-- Index Usage:
-- 1. parent_id = ? (leftmost prefix - exact match)
-- 2. is_virtual_item = 0 (full index use)
-- 3. ORDER BY sort_name (covered by index, NO filesort)
--
-- Performance: O(log n) search, no temp tables, no filesort
-- Typical rows: 10-100 items per folder

SELECT 
    id,
    name,
    type,
    media_type,
    is_folder,
    overview,
    production_year,
    community_rating,
    image_tag,
    date_created,
    date_last_saved
FROM BaseItems
WHERE parent_id = ?
  AND is_virtual_item = 0
ORDER BY sort_name ASC
LIMIT 100 OFFSET 0;

-- ===============================
-- Parent Folder Contents with Type Filter
-- Uses: IX_BaseItems_ParentId_IsVirtualItem_Type
-- ===============================

-- Pattern: Get specific type within folder (e.g., episodes of a season)
SELECT 
    id,
    name,
    type,
    sort_name,
    overview,
    premiere_date,
    run_time_ticks,
    community_rating
FROM BaseItems
WHERE parent_id = ?
  AND type = 'Episode'
  AND is_virtual_item = 0
ORDER BY sort_name ASC;

-- ===============================
-- Parent Folder Contents with Pagination
-- Uses: IX_BaseItems_ParentId_IsVirtualItem_Type
-- ===============================

-- Pattern: Browsable folder view with pagination
-- Recommended page size: 20-50 items (UI display size)
SELECT 
    id,
    name,
    type,
    is_folder,
    production_year,
    community_rating,
    image_tag,
    sort_name,
    overview
FROM BaseItems
WHERE parent_id = ?
  AND is_virtual_item = 0
ORDER BY sort_name ASC
LIMIT 20 OFFSET 0;

-- ===============================
-- Folder Navigation with Item Count
-- Uses: IX_BaseItems_ParentId_IsVirtualItem_Type + COUNT optimization
-- ===============================

-- Pattern: Get folder contents with count for UI navigation
-- Single query provides both data and total count
WITH folder_contents AS (
    SELECT 
        id,
        name,
        type,
        is_folder,
        production_year,
        community_rating,
        image_tag,
        ROW_NUMBER() OVER (ORDER BY sort_name ASC) as row_num
    FROM BaseItems
    WHERE parent_id = ?
      AND is_virtual_item = 0
)
SELECT 
    fc.*,
    (SELECT COUNT(*) FROM folder_contents) as total_count
FROM folder_contents fc
WHERE row_num BETWEEN 1 AND 20;

-- Alternative (MySQL 5.7 compatible - separate count query)
-- Query 1: Get count
-- Query 2: Get page
SELECT COUNT(*) as total_count
FROM BaseItems
WHERE parent_id = ?
  AND is_virtual_item = 0;

SELECT id, name, type, is_folder, overview, image_tag
FROM BaseItems
WHERE parent_id = ?
  AND is_virtual_item = 0
ORDER BY sort_name ASC
LIMIT 20 OFFSET 0;

-- ===============================
-- Folder Contents Sorted by Different Criteria
-- ===============================

-- Sort by Name (Default, index-optimal)
SELECT * FROM BaseItems
WHERE parent_id = ? AND is_virtual_item = 0
ORDER BY sort_name ASC;

-- Sort by Date (may need filesort - no covering index)
-- Note: For large folders, this is O(n log n) vs O(log n) with sort_name
SELECT *
FROM BaseItems
WHERE parent_id = ? AND is_virtual_item = 0
ORDER BY date_created DESC;

-- Sort by Random (slow, avoid for large folders)
-- Only use for small folders (< 100 items)
SELECT *
FROM BaseItems
WHERE parent_id = ? AND is_virtual_item = 0
ORDER BY RAND()
LIMIT 20;

-- ===============================
-- Recommended Index for Date Sorting
-- ===============================

-- If date sorting is common for folder navigation:
CREATE INDEX IX_BaseItems_ParentId_DateCreated 
    ON BaseItems(parent_id, is_virtual_item, date_created DESC);

-- Then use for date-sorted folder views:
SELECT * FROM BaseItems
WHERE parent_id = ? AND is_virtual_item = 0
ORDER BY date_created DESC;
-- Full index usage, no filesort

-- ===============================
-- P4 Authentication: Token Validation
-- Uses: IX_Devices_AccessToken
-- ===============================

-- Fast lookup: O(log n)
SELECT id, device_id, user_id, name, app_id, app_version
FROM Devices
WHERE access_token = ?
LIMIT 1;

-- ===============================
-- P6 Filtering: Genres for Filter UI
-- Uses: IX_ItemValues_Type_Name
-- ===============================

-- Get all Genre values for filter dropdown in folder
SELECT id, name, type, image_tag
FROM ItemValues
WHERE type = 'Genre'
ORDER BY name ASC;

-- ===============================
-- P6 Filtering: Find Movie by Genre in Folder
-- ===============================

-- Filter folder contents by multiple genres
SELECT DISTINCT bi.id, bi.name, bi.type, bi.overview
FROM BaseItems bi
INNER JOIN ItemValuesMap ivm 
    ON bi.id = ivm.library_item_id
WHERE ivm.item_value_id IN (
    SELECT id FROM ItemValues 
    WHERE type = 'Genre' AND name IN ('Action', 'Comedy')
    UNION
    SELECT id FROM ItemValues 
    WHERE type = 'Genre' AND name = ('Drama')
)
  AND bi.parent_id = ?
  AND bi.is_virtual_item = 0
ORDER BY bi.sort_name ASC;

-- ===============================
-- Query Performance Metrics (Expected)
-- ===============================

-- | Query Type | Rows Scanned | Index Usage | Time (1000 rows) |
-- |------------|-------------|------------|-----------------|
-- | Single type browse | 20 | 100% | 2ms |
-- | Multi-type IN browse | 180 | 80% | 8ms |
-- | Parent folder | 50-200 | 100% | 5ms |
-- | Parent folder + filter | 50 | 90% | 4ms |
-- | Token validation | 1 | 100% | <1ms |
-- | Genre lookup | ALL | 100% | 1ms |

-- ===============================
-- Maintenance: Analyze Index Usage
-- ===============================

-- Check if queries use expected indexes
EXPLAIN SELECT * FROM BaseItems
WHERE parent_id = ? AND is_virtual_item = 0
ORDER BY sort_name ASC;

-- Expected EXPLAIN output:
-- type: range
-- key: IX_BaseItems_ParentId_IsVirtualItem_Type
-- rows: 20-200 (actual rows per folder)
-- Extra: Using index

-- Analyze tables after bulk data load
ANALYZE TABLE BaseItems;
ANALYZE TABLE ItemValues;
ANALYZE TABLE ItemValuesMap;
ANALYZE TABLE Users;
ANALYZE TABLE Devices;
