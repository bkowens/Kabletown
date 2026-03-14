-- Kabletown Items Schema v1.0
-- MySQL 8.0+

-- Core items table (P7 indexing: TopParentId + Type)
CREATE TABLE IF NOT EXISTS items (
    id CHAR(36) PRIMARY KEY,
    type VARCHAR(128) NOT NULL,
    name VARCHAR(255) NOT NULL,
    original_title VARCHAR(255),
    overview TEXT,
    path VARCHAR(1024),
    path_hash CHAR(64) UNIQUE,
    
    -- Hierarchy (for P7 indexing)
    parent_id CHAR(36),
    top_parent_id CHAR(36),
    ancestor_ids TEXT, -- Comma-separated UUIDs
    collection_type VARCHAR(64),
    
    -- Media info
    container VARCHAR(255),
    video_codec VARCHAR(64),
    audio_codec VARCHAR(255),

    -- MediaStreams (JSON array stored as TEXT)
    media_streams TEXT,
    media_sources TEXT,
    
    -- Runtime
    run_time_ticks BIGINT,
    
    -- Dates
    premiered_date DATETIME(7),
    date_created DATETIME(7) DEFAULT CURRENT_TIMESTAMP(7),
    date_modified DATETIME(7) DEFAULT CURRENT_TIMESTAMP(7) ON UPDATE CURRENT_TIMESTAMP(7),
  
    -- Status
    is_folder BOOLEAN DEFAULT FALSE,
    is_inmixed BOOLEAN DEFAULT FALSE,
    is_virtual_item BOOLEAN DEFAULT FALSE,
    is_hd BOOLEAN,
    
    -- Metadata
    tmdb_id CHAR(36),
    tvdb_id CHAR(36),
    imdb_id VARCHAR(32),
    tmdb_type VARCHAR(64),
    
    -- Parent info (cached for performance)
    parent_id_display VARCHAR(255),
    studio VARCHAR(255),
    official_rating VARCHAR(64),
    tagline VARCHAR(512),
    
    -- Library linkage
    library_folder_id CHAR(36),
    
    -- P7 composite index target
    CONSTRAINT fk_items_parent FOREIGN KEY (parent_id) REFERENCES items(id) ON DELETE SET NULL,
    CONSTRAINT fk_items_top_parent FOREIGN KEY (top_parent_id) REFERENCES items(id) ON DELETE SET NULL,
    
    INDEX idx_items_topparent_type (top_parent_id, type),  -- P7 index
    INDEX idx_items_parent (parent_id),
    INDEX idx_items_path_hash (path_hash),
    INDEX idx_items_type (type),
    INDEX idx_items_name (name),
    INDEX idx_items_premiered (premiered_date)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Item values (P6 indexing - normalized multi-value fields)
-- Stores: Genres, Studios, People, Tags, CollectionTypes
CREATE TABLE IF NOT EXISTS item_values (
    id CHAR(36) PRIMARY KEY,
    item_id CHAR(36) NOT NULL,
    value_type VARCHAR(64) NOT NULL,  -- Genre, Studio, Person, Tag, CollectionType
    value_name VARCHAR(255) NOT NULL,
    value_id CHAR(36),  -- Reference to item if value is also an item (e.g., Actor, Director)
    
    CONSTRAINT fk_item_values_item FOREIGN KEY (item_id) REFERENCES items(id) ON DELETE CASCADE,
    
    UNIQUE KEY unique_item_value_type (item_id, value_type, value_name),
    INDEX idx_item_values_name_type (value_name, value_type),
    INDEX idx_item_values_value_id (value_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- User data (playback progress, favorites, etc.)
CREATE TABLE IF NOT EXISTS user_data (
    id CHAR(36) PRIMARY KEY,
    user_id CHAR(36) NOT NULL,
    item_id CHAR(36) NOT NULL,
    
    -- Playback
    playback_position_ticks BIGINT DEFAULT 0,
    play_count INT DEFAULT 0,
    is_played BOOLEAN DEFAULT FALSE,
    
    -- Timestamps
    last_play_date DATETIME(7),
    
    -- Rating
    likes BOOLEAN,
    dislike BOOLEAN,
    rating DECIMAL(3,2),
    
    CONSTRAINT fk_user_data_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_user_data_item FOREIGN KEY (item_id) REFERENCES items(id) ON DELETE CASCADE,
    
    UNIQUE KEY unique_user_item (user_id, item_id),
    INDEX idx_user_data_user (user_id),
    INDEX idx_user_data_item (item_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Library folders (media library definitions)
CREATE TABLE IF NOT EXISTS library_folders (
    id CHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    path VARCHAR(1024) NOT NULL,
    library_type VARCHAR(64) NOT NULL,  -- Movies, TVShows, Music, etc.
    
    -- Settings
    default_image_preference VARCHAR(64),
    automatic_refresh_interval_minutes INT DEFAULT 0,
    is_enabled BOOLEAN DEFAULT TRUE,
    is_hidden BOOLEAN DEFAULT FALSE,
    
    -- Metadata
    created_at DATETIME(7) DEFAULT CURRENT_TIMESTAMP(7),
    updated_at DATETIME(7) DEFAULT CURRENT_TIMESTAMP(7) ON UPDATE CURRENT_TIMESTAMP(7),
    
    UNIQUE KEY unique_library_name (name),
    INDEX idx_library_type (library_type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Media paths (physical file locations for library folders)
CREATE TABLE IF NOT EXISTS media_paths (
    id CHAR(36) PRIMARY KEY,
    library_id CHAR(36) NOT NULL,
    path VARCHAR(1024) NOT NULL,
    item_type VARCHAR(64),  -- Movies, TVShows, Music, etc.
    
    CONSTRAINT fk_media_paths FOREIGN KEY (library_id) REFERENCES library_folders(id) ON DELETE CASCADE,
    
    UNIQUE KEY unique_path (path),
    INDEX idx_library_paths (library_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Collections (box sets, manual collections)
CREATE TABLE IF NOT EXISTS collections (
    id CHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    owner_user_id CHAR(36),
    is_managed BOOLEAN DEFAULT FALSE,  -- Auto-generated (box set) vs user-created
    
    -- Metadata
    overview TEXT,
    official_rating VARCHAR(64),
    premiered_date DATETIME(7),
    
    -- Settings
    is_public BOOLEAN DEFAULT TRUE,
    
    CONSTRAINT fk_collections_user FOREIGN KEY (owner_user_id) REFERENCES users(id) ON DELETE SET NULL,
    
    INDEX idx_collections_user (owner_user_id),
    INDEX idx_collections_name (name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Collection members
CREATE TABLE IF NOT EXISTS collection_items (
    id CHAR(36) PRIMARY KEY,
    collection_id CHAR(36) NOT NULL,
    item_id CHAR(36) NOT NULL,
    sort_order INT DEFAULT 0,
    
    CONSTRAINT fk_collection_items_collection FOREIGN KEY (collection_id) REFERENCES collections(id) ON DELETE CASCADE,
    CONSTRAINT fk_collection_items_item FOREIGN KEY (item_id) REFERENCES items(id) ON DELETE CASCADE,
    
    UNIQUE KEY unique_collection_item (collection_id, item_id),
    INDEX idx_collection_items_collection (collection_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Initial data: Insert default library folders
INSERT INTO library_folders (id, name, path, library_type) VALUES
('00000000-0000-0000-0000-000000000101', 'Movies', '/media/movies', 'Movies'),
('00000000-0000-0000-0000-000000000102', 'TV Shows', '/media/tv', 'TvShows'),
('00000000-0000-0000-0000-000000000103', 'Music', '/media/music', 'Music'),
('00000000-0000-0000-0000-000000000104', 'Home Videos', '/media/videos', 'HomeVideos'),
('00000000-0000-0000-0000-000000000105', 'Pictures', '/media/pictures', 'Pictures');
