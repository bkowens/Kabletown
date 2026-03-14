-- Kabletown Playstate Schema
-- MySQL 8.0+

-- Playback state tracking (active sessions)
CREATE TABLE IF NOT EXISTS playback_state (
    id CHAR(36) PRIMARY KEY,
    user_id CHAR(36) NOT NULL,
    item_id CHAR(36) NOT NULL,
    
    -- Playback info
    position_ticks BIGINT DEFAULT 0,
    media_source_id CHAR(36),
    audio_stream_index INT,
    subtitle_stream_index INT,
    is_paused BOOLEAN DEFAULT FALSE,
    is_active BOOLEAN DEFAULT TRUE,
    
    -- Timestamps
    started_at DATETIME(7) DEFAULT CURRENT_TIMESTAMP(7),
    updated_at DATETIME(7) DEFAULT CURRENT_TIMESTAMP(7) ON UPDATE CURRENT_TIMESTAMP(7),
    
    CONSTRAINT fk_playback_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_playback_item FOREIGN KEY (item_id) REFERENCES items(id) ON DELETE CASCADE,
    
    UNIQUE KEY unique_active_playback (user_id, item_id, is_active),
    INDEX idx_playback_active (is_active),
    INDEX idx_playback_user_item (user_id, item_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Play history (historical records)
CREATE TABLE IF NOT EXISTS play_history (
    id CHAR(36) PRIMARY KEY,
    user_id CHAR(36) NOT NULL,
    item_id CHAR(36) NOT NULL,
    
    -- Playback details
    position_ticks BIGINT DEFAULT 0,
    duration_ticks BIGINT,
    completed BOOLEAN DEFAULT FALSE,
    play_method VARCHAR(64),
    
    -- Timestamps
    start_date DATETIME(7) NOT NULL,
    end_date DATETIME(7),
    
    CONSTRAINT fk_history_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_history_item FOREIGN KEY (item_id) REFERENCES items(id) ON DELETE CASCADE,
    
    INDEX idx_history_user (user_id),
    INDEX idx_history_item (item_id),
    INDEX idx_history_date (start_date)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
