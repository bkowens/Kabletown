-- ==============================================================================
-- Full-Text Search Queries
-- Uses: FT_BaseItems_Name_OriginalTitle
-- Query: MATCH(name, original_title) AGAINST(? IN BOOLEAN MODE)
-- ==============================================================================
--
-- Index Usage Analysis:
-- Index: FT_BaseItems_Name_OriginalTitle (FULLTEXT on Name, OriginalTitle)
--
-- For full-text search queries:
--   - MATCH(..., ...) AGAINST(? IN BOOLEAN MODE) (full index use)
--   - MySQL full-text search engine (inverted index, word-based)
--   - Relevance scoring built-in (MATCH returns relevance score)
--
-- Performance Characteristics:
--   - Index build: O(n log n) during index creation
--   - Search: O(k) where k = number of matching documents
--   - Relevance scoring: Built-in TF-IDF algorithm
--   - Typical results: 0-100 matches (search term dependent)
--   - Query time: 5-50ms (depends on result set size)
--
-- Boolean Mode Features:
--   - '+' operator: word MUST be present
--   - '-' operator: word MUST NOT be present
--   - '*' wildcard: prefix matching (e.g., 'star*')
--   - '"quoted phrase"': exact phrase matching
--   - '>word' : boost word relevance
--   - '<word' : reduce word relevance
--
-- Limitations:
--   - Minimum word size: 3 characters (ft_min_word_len in MySQL)
--   - Stop words ignored (the, is, at, etc.)
--   - Exact match not guaranteed (word stemming, case insensitive)
--   - Best for broad search, not exact matching
--
-- Alternative Index Strategy:
--   - For exact prefix matching (auto-complete): IX_BaseItems_SortName
--   - For type-filtered search: Composite index on (type, name)
--   - Combine full-text with type filter for better relevance
-- ==============================================================================

-- ==============================================================================
-- Pattern 1: Basic Full-Text Search (All Types)
-- Uses: FT_BaseItems_Name_OriginalTitle
-- ==============================================================================

-- Search across all items by name or original title
-- Returns results ordered by relevance score

SELECT 
    bi.id,
    bi.name,
    bi.original_title,
    bi.type,
    bi.production_year,
    bi.series_id,
    bi.image_tags,
    bi.premiere_date,
    MATCH(bi.name, bi.original_title) AGAINST(? IN BOOLEAN MODE) as relevance_score
FROM BaseItems bi
WHERE bi.is_virtual_item = 0
  AND MATCH(bi.name, bi.original_title) AGAINST(? IN BOOLEAN MODE)
  AND bi.type IN ('Movie', 'Series', 'Episode', 'Trailer', 'MusicVideo')
ORDER BY relevance_score DESC
LIMIT 50;

-- ==============================================================================
-- Pattern 2: Full-Text Search with Type Filter
-- Uses: FT_BaseItems_Name_OriginalTitle + Type filter
-- ==============================================================================

-- Restrict search to specific media type
-- Better relevance for type-specific queries

SELECT 
    bi.id,
    bi.name,
    bi.original_title,
    bi.type,
    bi.production_year,
    bi.series_id,
    bi.season_id,
    bi.index_number,
    bi.image_tags,
    bi.premiere_date,
    MATCH(bi.name, bi.original_title) AGAINST(? IN BOOLEAN MODE) as relevance_score
FROM BaseItems bi
WHERE bi.is_virtual_item = 0
  AND bi.type = 'Movie'
  AND MATCH(bi.name, bi.original_title) AGAINST(? IN BOOLEAN MODE)
ORDER BY relevance_score DESC
LIMIT 20;

-- ==============================================================================
-- Pattern 3: Full-Text Search for Episodes (Series Context)
-- Uses: FT_BaseItems_Name_OriginalTitle + Series filter
-- ==============================================================================

-- Search episodes within a specific series
-- Useful for "search episodes in this show"

SELECT 
    bi.id,
    bi.name,
    bi.original_title,
    bi.type,
    bi.series_id,
    bi.season_id,
    bi.season_index_number,
    bi.index_number as episode_number,
    bi.aired_date,
    bi.image_tags,
    bi.runtime_ticks,
    MATCH(bi.name, bi.original_title) AGAINST(? IN BOOLEAN MODE) as relevance_score
FROM BaseItems bi
WHERE bi.is_virtual_item = 0
  AND bi.type = 'Episode'
  AND bi.series_id = ?
  AND MATCH(bi.name, bi.original_title) AGAINST(? IN BOOLEAN MODE)
ORDER BY bi.season_index_number ASC, bi.index_number ASC
LIMIT 50;

-- ==============================================================================
-- Pattern 4: Boolean Mode Search with Multiple Terms
-- Uses: FT_BaseItems_Name_OriginalTitle (Boolean Mode)
-- ==============================================================================

-- Advanced search with boolean operators
-- Example queries:
--   '+star +wars' (must have both words)
--   'star -wars' (must have 'star', not 'wars')
--   '"star wars"' (exact phrase match)
--   'star*' (prefix match, like 'stars', 'starwars', etc.)

SELECT 
    bi.id,
    bi.name,
    bi.original_title,
    bi.type,
    bi.production_year,
    bi.image_tags,
    bi.premiere_date,
    MATCH(bi.name, bi.original_title) AGAINST(? IN BOOLEAN MODE) as relevance_score
FROM BaseItems bi
WHERE bi.is_virtual_item = 0
  AND MATCH(bi.name, bi.original_title) AGAINST(? IN BOOLEAN MODE)
ORDER BY relevance_score DESC
LIMIT 50;

-- Note: Boolean mode is case-insensitive and ignores stop words
-- Users should be instructed on syntax for advanced searches

-- ==============================================================================
-- Pattern 5: Full-Text Search with Pagination
-- Uses: FT_BaseItems_Name_OriginalTitle + LIMIT/OFFSET
-- ==============================================================================

-- Paginated full-text search results
-- Note: Total count requires separate query

-- Query 1: Get total match count
SELECT COUNT(*) as total_matches
FROM BaseItems bi
WHERE bi.is_virtual_item = 0
  AND MATCH(bi.name, bi.original_title) AGAINST(? IN BOOLEAN MODE);

-- Query 2: Get paginated results
SELECT 
    bi.id,
    bi.name,
    bi.original_title,
    bi.type,
    bi.production_year,
    bi.series_id,
    bi.image_tags,
    MATCH(bi.name, bi.original_title) AGAINST(? IN BOOLEAN MODE) as relevance_score
FROM BaseItems bi
WHERE bi.is_virtual_item = 0
  AND MATCH(bi.name, bi.original_title) AGAINST(? IN BOOLEAN MODE)
ORDER BY relevance_score DESC
LIMIT 20 OFFSET 0;

-- Performance Note: OFFSET with large pages can be slow
-- Consider using keyset pagination (WHERE relevance_score < ? for next page)

-- ==============================================================================
-- Pattern 6: Auto-Complete / Type-Ahead Search (NOT Full-Text)
-- Uses: IX_BaseItems_SortName (Prefix matching, not full-text)
-- ==============================================================================

-- For type-ahead search as user types
-- Full-text has minimum word size (3 chars) and doesn't work well for short inputs
-- Use SortName prefix matching instead

SELECT 
    bi.id,
    bi.name,
    bi.type,
    bi.production_year,
    bi.image_tags,
    'matches_start' as match_type
FROM BaseItems bi
WHERE bi.is_virtual_item = 0
  AND bi.sort_name >= ?
  AND bi.sort_name < CONCAT(?, CHAR(127))
  AND bi.type IN ('Movie', 'Series', 'Episode')
ORDER BY bi.sort_name ASC
LIMIT 20;

-- Note: This uses IX_BaseItems_SortName index for O(log n) lookup
-- Sort_name is the lowercased version of name for case-insensitive matching
-- CHAR(127) is the highest ASCII character, creates range end

-- ==============================================================================
-- Pattern 7: Fuzzy Search Alternative (Like Full-Text but Simpler)
-- Uses: Standard LIKE operators (no full-text index)
-- ==============================================================================

-- Fallback for very short search terms (< 3 chars)
-- or when full-text relevance scoring isn't needed

SELECT 
    bi.id,
    bi.name,
    bi.original_title,
    bi.type,
    bi.production_year,
    bi.image_tags,
    'exact_or_contains' as match_type
FROM BaseItems bi
WHERE bi.is_virtual_item = 0
  AND (bi.name LIKE ? ESCAPE '\\'
       OR bi.original_title LIKE ? ESCAPE '\\')
  AND bi.type IN ('Movie', 'Series', 'Episode', 'Trailer')
ORDER BY LENGTH(bi.name) ASC, bi.name ASC
LIMIT 20;

-- Note: Full-text index is minimum 3 chars by default
-- Use LIKE for 1-2 character searches
-- Performance: O(n) scan, no index use for '%term%' patterns

-- ==============================================================================
-- Pattern 8: Search with User-State Filtering
-- Uses: FT_BaseItems_Name_OriginalTitle + UserData join
-- ==============================================================================

-- Search with user-specific filters (favorites, watched, in-progress)

-- Get searchable items that user has NOT watched yet
SELECT 
    bi.id,
    bi.name,
    bi.original_title,
    bi.type,
    bi.production_year,
    bi.image_tags,
    MATCH(bi.name, bi.original_title) AGAINST(? IN BOOLEAN MODE) as relevance_score,
    ud.played,
    ud.playback_position_ticks,
    ud.is_favorite
FROM BaseItems bi
LEFT JOIN UserData ud ON bi.id = ud.library_item_id AND ud.user_id = ?
WHERE bi.is_virtual_item = 0
  AND MATCH(bi.name, bi.original_title) AGAINST(? IN BOOLEAN MODE)
  AND (ud.played = 0 OR ud.played IS NULL)
  AND bi.type IN ('Movie', 'Series')
ORDER BY relevance_score DESC
LIMIT 50;

-- ==============================================================================
-- Pattern 9: Library-Specific Search
-- Uses: FT_BaseItems_Name_OriginalTitle + Parent/Library hierarchy
-- ==============================================================================

-- Search within a specific collection or library folder

-- Find library root for userId (collections are library roots)
WITH RECURSIVE LibraryTree AS (
    -- Start from user's collections
    SELECT c.id as root_id, c.parent_id, c.library_name
    FROM Collections c
    WHERE c.user_id = ?
    
    UNION ALL
    
    -- Include sub-collections (if nested)
    SELECT lc.id, lc.parent_id, lc.library_name
    FROM BaseItems lc
    INNER JOIN LibraryTree lt ON lc.parent_id = lt.root_id
    WHERE lc.is_virtual_item = 0
      AND lc.type IN ('Collection', 'Folder')
)
SELECT DISTINCT 
    bi.id,
    bi.name,
    bi.original_title,
    bi.type,
    bi.production_year,
    bi.image_tags,
    lt.library_name,
    MATCH(bi.name, bi.original_title) AGAINST(? IN BOOLEAN MODE) as relevance_score
FROM BaseItems bi
INNER JOIN LibraryTree lt ON bi.id = lt.root_id
WHERE bi.is_virtual_item = 0
  AND bi.type IN ('Movie', 'Series')
  AND MATCH(bi.name, bi.original_title) AGAINST(? IN BOOLEAN MODE)
ORDER BY relevance_score DESC
LIMIT 50;

-- ==============================================================================
-- Pattern 10: Smart Search (Auto-Detect Boolean Mode)
-- Uses: Application-side query construction
-- ==============================================================================

-- Application should construct boolean mode query based on input:
-- 
-- Input: "star wars" -> '"star wars"' (phrase match for 2+ words)
-- Input: "super*"    -> 'super*' (prefix match with wildcard)
-- Input: "star -wars" -> 'star -wars' (user provided exclusion)
--
-- SQL pattern (constructed by application):

-- Dynamic query based on searchTerm:
SELECT 
    bi.id,
    bi.name,
    bi.original_title,
    bi.type,
    bi.production_year,
    bi.image_tags,
    bi.premiere_date,
    MATCH(bi.name, bi.original_title) AGAINST(:searchTerm IN BOOLEAN MODE) as relevance_score
FROM BaseItems bi
WHERE bi.is_virtual_item = 0
  AND bi.type IN ('Movie', 'Series', 'Episode', 'Trailer', 'MusicVideo')
  AND MATCH(bi.name, bi.original_title) AGAINST(:searchTerm IN BOOLEAN MODE)
ORDER BY relevance_score DESC
LIMIT 50;

-- Note: Validate user input to prevent injection in boolean mode
-- Escape special characters: * + - > < ( ) ~ ! @ " : \\

-- ==============================================================================
-- Query Performance Summary
-- ==============================================================================
--
-- Query Type                    | Index Used              | Complexity
-- ------------------------------|-------------------------|------------
-- Basic Full-Text Search        | FT_BaseItems_Name_...   | O(k) k=matches
-- Type-Filtered Search          | FT + Type filter        | O(k) + index
-- Episode Search (Series)       | FT + SeriesId + Type    | O(k)
-- Boolean Mode Search           | FT_BaseItems_Name_...   | O(k)
-- Paginated Search              | FT + LIMIT/OFFSET       | O(k + offset)
-- Auto-Complete (short input)   | IX_BaseItems_SortName   | O(log n)
-- Fallback (1-2 chars)          | No index (LIKE)         | O(n)
-- User-State Filtered           | FT + UserData join      | O(k)
--
-- Recommendations:
-- 1. Use full-text for searches > 3 characters
-- 2. Use SortName index auto-complete for short inputs
-- 3. Cache popular search results (e.g., "Marvel", "Star Wars")
-- 4. Implement search suggestions using SortName prefix
-- 5. Combine full-text with filters (type, year, user-state)
-- 6. Consider relevance_score < threshold to filter spam results
-- 7. For very large libraries (> 1M items), consider Elasticsearch/Meilisearch
-- ==============================================================================
