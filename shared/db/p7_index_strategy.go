package db

// P7 Performance Index Strategy
// ============================
// Every SQL query in the Go services must be designed to use the indexes
// created by C# migration: 20260309000000_AddPerformanceIndexes.cs
//
// Index Usage Guidelines:
// 1. Primary Key queries use PRIMARY KEY index
// 2. Foreign Key lookups use dedicated FK indexes
// 3. Filtering on multiple columns uses composite indexes (leftmost prefix rule)
// 4. ORDER BY clauses must match index ordering
// 5. WHERE clause columns must match index column order (no function wrapping)
//
// CRITICAL INDEXES (from C# migration):
// - IX_BaseItems_Type_IsVirtualItem_SortName
// - IX_BaseItems_ParentId_IsVirtualItem_Type
// - IX_UserData_UserId_IsFavorite
// - IX_UserData_UserId_Played

// GetCreateBaseItemsSQL returns the SQL to create BaseItems table with P7 indexes
// Indexes must match C# migration: 20260309000000_AddPerformanceIndexes.cs
func GetCreateBaseItemsSQL() string {
	return `CREATE TABLE IF NOT EXISTS BaseItems (
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
		user_id CHAR(36),
		date_created DATETIME(7) NOT NULL,
		date_last_saved DATETIME(7) NOT NULL,
		row_version INT UNSIGNED NOT NULL DEFAULT 0,
		
		-- C# Migration: 20260309000000_AddPerformanceIndexes.cs
		-- P7: IX_BaseItems_Type_IsVirtualItem_SortName
		-- Pattern A: Single type lookup (optimal full index use)
		-- Query: WHERE type = 'Movie' AND is_virtual_item = 0 ORDER BY sort_name ASC
		--
		-- Pattern B: Multiple types with IN clause (partial index use)
		-- Query: WHERE type IN ('Movie', 'Show', 'Episode') AND is_virtual_item = 0 ORDER BY sort_name ASC
		-- Note: MySQL can use (type) portion of composite index, may need filesort for ORDER BY
		INDEX IX_BaseItems_Type_IsVirtualItem_SortName (type, is_virtual_item, sort_name),
		
		-- P7: IX_BaseItems_ParentId_IsVirtualItem_Type
		-- Pattern: Get children of folder with type filter
		-- Query: WHERE parent_id = ? AND is_virtual_item = ? AND type = ?
		INDEX IX_BaseItems_ParentId_IsVirtualItem_Type (parent_id, is_virtual_item, type),
		
		-- P7: IX_BaseItems_User_Id
		-- Pattern: Filter items by user
		-- Query: WHERE user_id = ?
		INDEX IX_BaseItems_User_Id (user_id),
		
		-- P7: IX_BaseItems_Name_Prefix
		-- Pattern: Auto-complete search
		-- Query: WHERE name LIKE '%prefix%'
		INDEX IX_BaseItems_Name (name(100))
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;`
}

// GetCreateUsersSQL returns the SQL to create Users table with P7 indexes
func GetCreateUsersSQL() string {
	return `CREATE TABLE IF NOT EXISTS Users (
		id CHAR(36) PRIMARY KEY,
		name VARCHAR(255) NOT NULL UNIQUE,
		password_hash VARCHAR(255) NOT NULL,
		email VARCHAR(255),
		is_admin BOOLEAN NOT NULL DEFAULT FALSE,
		is_disabled BOOLEAN NOT NULL DEFAULT FALSE,
		row_version INT UNSIGNED NOT NULL DEFAULT 0,
		date_created DATETIME(6) NOT NULL,
		date_modified DATETIME(6) NOT NULL,
		date_last_login DATETIME(6),
		
		-- P7 Indexes for User queries
		INDEX IX_Users_Name (name),
		INDEX IX_Users_Email (email),
		INDEX IX_Users_IsAdmin (is_admin)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;`
}

// GetCreateDevicesSQL returns the SQL to create Devices table with P7 indexes
func GetCreateDevicesSQL() string {
	return `CREATE TABLE IF NOT EXISTS Devices (
		id CHAR(36) PRIMARY KEY,
		device_id CHAR(36) NOT NULL,
		user_id CHAR(36),
		name VARCHAR(255) NOT NULL,
		app_id VARCHAR(255) NOT NULL,
		app_version VARCHAR(50),
		last_user_id CHAR(36),
		access_token VARCHAR(255),
		row_version INT UNSIGNED NOT NULL DEFAULT 0,
		date_created DATETIME(6) NOT NULL,
		date_last_activity DATETIME(6) NOT NULL,
		
		-- P7 Indexes for P4 Token Resolution
		-- Pattern: WHERE access_token = ? (O(log n) lookup)
		INDEX IX_Devices_AccessToken (access_token),
		INDEX IX_Devices_DeviceId (device_id),
		INDEX IX_Devices_UserId (user_id)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;`
}

// GetCreateApiKeysSQL returns the SQL to create ApiKeys table with P7 indexes
func GetCreateApiKeysSQL() string {
	return `CREATE TABLE IF NOT EXISTS ApiKeys (
		id INT AUTO_INCREMENT PRIMARY KEY,
		user_id CHAR(36),
		name VARCHAR(255) NOT NULL,
		access_token VARCHAR(255) NOT NULL,
		date_created DATETIME(6) NOT NULL,
		date_last_used DATETIME(6),
		row_version INT UNSIGNED NOT NULL DEFAULT 0,
		
		-- P7 Indexes for P4 Token Resolution
		INDEX IX_ApiKeys_AccessToken (access_token),
		INDEX IX_ApiKeys_UserId (user_id)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;`
}

// GetCreateUserDataSQL returns the SQL to create UserData table with P7 indexes
func GetCreateUserDataSQL() string {
	return `CREATE TABLE IF NOT EXISTS UserData (
		id INT AUTO_INCREMENT PRIMARY KEY,
		user_id CHAR(36) NOT NULL,
		library_item_id CHAR(36) NOT NULL,
		played BOOLEAN NOT NULL DEFAULT FALSE,
		rating FLOAT,
		play_count INT NOT NULL DEFAULT 0,
		playback_position_ticks BIGINT NOT NULL DEFAULT 0,
		last_played_date DATETIME(6),
		is_favorite BOOLEAN NOT NULL DEFAULT FALSE,
		row_version INT UNSIGNED NOT NULL DEFAULT 0,
		
		-- P7 Indexes from C# migration: 20260309000000_AddPerformanceIndexes.cs
		-- IX_UserData_UserId_IsFavorite: User's favorite items
		-- Query: WHERE user_id = ? AND is_favorite = ? ORDER BY last_played_date DESC
		INDEX IX_UserData_UserId_IsFavorite (user_id, is_favorite),
		
		-- IX_UserData_UserId_Played: User's unplayed items
		-- Query: WHERE user_id = ? AND played = ? ORDER BY last_played_date DESC
		INDEX IX_UserData_UserId_Played (user_id, played),
		
		UNIQUE KEY unique_userdata_user_item (user_id, library_item_id),
		INDEX IX_UserData_LibraryItemId (library_item_id)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;`
}

// GetCreateItemValuesSQL returns the SQL to create ItemValues table with P7 indexes
func GetCreateItemValuesSQL() string {
	return `CREATE TABLE IF NOT EXISTS ItemValues (
		id CHAR(36) PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		type VARCHAR(50) NOT NULL,
		image_tag VARCHAR(255),
		row_version INT UNSIGNED NOT NULL DEFAULT 0,
		
		-- P7 Indexes for P6 filtering
		-- Pattern: WHERE type = ? ORDER BY name ASC (get all genres, studios, etc.)
		INDEX IX_ItemValues_Type_Name (type, name),
		
		-- Pattern: WHERE name = ? AND type = ? (lookup by name)
		INDEX IX_ItemValues_Name_Type (name(100), type)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;`
}

// GetCreateItemValuesMapSQL returns the SQL to create ItemValuesMap table with P7 indexes
func GetCreateItemValuesMapSQL() string {
	return `CREATE TABLE IF NOT EXISTS ItemValuesMap (
		id INT AUTO_INCREMENT PRIMARY KEY,
		library_item_id CHAR(36) NOT NULL,
		item_value_id CHAR(36) NOT NULL,
		row_version INT UNSIGNED NOT NULL DEFAULT 0,
		
		-- P7 Indexes for ItemValue lookups
		-- Pattern: Find all items with a specific Genre/Studio/etc.
		-- Query: WHERE item_value_id = ?
		INDEX IX_ItemValuesMap_ItemValueId (item_value_id),
		
		-- Pattern: Get all values for a specific item
		-- Query: WHERE library_item_id = ?
		INDEX IX_ItemValuesMap_LibraryItemId (library_item_id),
		UNIQUE KEY unique_item_value_map (library_item_id, item_value_id)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;`
}

// GetCreateCollectionsSQL returns the SQL to create Collections table with P7 indexes
func GetCreateCollectionsSQL() string {
	return `CREATE TABLE IF NOT EXISTS Collections (
		id INT AUTO_INCREMENT PRIMARY KEY,
		name VARCHAR(1024) NOT NULL,
		user_id CHAR(36) NOT NULL,
		image_tag VARCHAR(255),
		row_version INT UNSIGNED NOT NULL DEFAULT 0,
		date_created DATETIME(6) NOT NULL,
		
		-- P7 Indexes
		INDEX IX_Collections_UserId (user_id),
		INDEX IX_Collections_Name (name(255))
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;`
}

// GetCreateCollectionItemsSQL returns the SQL to create CollectionItems table with P7 indexes
func GetCreateCollectionItemsSQL() string {
	return `CREATE TABLE IF NOT EXISTS CollectionItems (
		id INT AUTO_INCREMENT PRIMARY KEY,
		collection_id INT NOT NULL,
		library_item_id CHAR(36) NOT NULL,
		next_item_id INT,
		previous_item_id INT,
		row_version INT UNSIGNED NOT NULL DEFAULT 0,
		
		-- P7 Indexes
		INDEX IX_CollectionItems_CollectionId (collection_id),
		INDEX IX_CollectionItems_LibraryItemId (library_item_id),
		UNIQUE KEY unique_collection_item (collection_id, library_item_id)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;`
}

// GetCreatePlaylistsSQL returns the SQL to create Playlists table with P7 indexes
func GetCreatePlaylistsSQL() string {
	return `CREATE TABLE IF NOT EXISTS Playlists (
		id INT AUTO_INCREMENT PRIMARY KEY,
		name VARCHAR(1024) NOT NULL,
		user_id CHAR(36) NOT NULL,
		image_tag VARCHAR(255),
		row_version INT UNSIGNED NOT NULL DEFAULT 0,
		date_created DATETIME(7) NOT NULL,
		
		-- P7 Indexes
		INDEX IX_Playlists_UserId (user_id),
		INDEX IX_Playlists_Name (name(255))
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;`
}

// GetCreatePlaylistItemsSQL returns the SQL to create PlaylistItems table with P7 indexes
func GetCreatePlaylistItemsSQL() string {
	return `CREATE TABLE IF NOT EXISTS PlaylistItems (
		id INT AUTO_INCREMENT PRIMARY KEY,
		playlist_id INT NOT NULL,
		library_item_id CHAR(36) NOT NULL,
		next_item_id INT,
		previous_item_id INT,
		row_version INT UNSIGNED NOT NULL DEFAULT 0,
		
		-- P7 Indexes
		INDEX IX_PlaylistItems_PlaylistId (playlist_id),
		INDEX IX_PlaylistItems_LibraryItemId (library_item_id),
		UNIQUE KEY unique_playlist_item (playlist_id, library_item_id)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;`
}
