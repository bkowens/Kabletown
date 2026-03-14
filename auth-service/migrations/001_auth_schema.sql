-- Kabletown Auth Schema v1.0
-- MySQL 8.0+

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id CHAR(36) PRIMARY KEY,
    username VARCHAR(255) NOT NULL UNIQUE,
    email VARCHAR(255),
    password_hash CHAR(64),
    is_admin BOOLEAN DEFAULT FALSE,
    is_hidden BOOLEAN DEFAULT FALSE,
    is_disabled BOOLEAN DEFAULT FALSE,
    last_login_date DATETIME(7) DEFAULT NULL,
    created_at DATETIME(7) DEFAULT CURRENT_TIMESTAMP(7),
    updated_at DATETIME(7) DEFAULT CURRENT_TIMESTAMP(7) ON UPDATE CURRENT_TIMESTAMP(7)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- API Keys / Sessions table
CREATE TABLE IF NOT EXISTS api_keys (
    id CHAR(36) PRIMARY KEY,
    user_id CHAR(36) NOT NULL,
    api_key_hash CHAR(64) NOT NULL UNIQUE,
    device_id VARCHAR(128) NOT NULL,
    device_name VARCHAR(255),
    last_use_date DATETIME(7) DEFAULT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    expires_at DATETIME(7) DEFAULT NULL,
    created_at DATETIME(7) DEFAULT CURRENT_TIMESTAMP(7),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_api_key_hash (api_key_hash),
    INDEX idx_user_id (user_id),
    INDEX idx_device_id (device_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Devices table
CREATE TABLE IF NOT EXISTS devices (
    id CHAR(36) PRIMARY KEY,
    user_id CHAR(36),
    device_id VARCHAR(128) NOT NULL UNIQUE,
    device_name VARCHAR(255),
    device_type VARCHAR(128),
    last_activity_date DATETIME(7) DEFAULT NULL,
    is_managed BOOLEAN DEFAULT FALSE,
    is_hidden BOOLEAN DEFAULT FALSE,
    created_at DATETIME(7) DEFAULT CURRENT_TIMESTAMP(7),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL,
    INDEX idx_device_id (device_id),
    INDEX idx_user_id (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Sessions table
CREATE TABLE IF NOT EXISTS sessions (
    id CHAR(36) PRIMARY KEY,
    user_id CHAR(36) NOT NULL,
    token_hash CHAR(64) NOT NULL UNIQUE,
    device_id VARCHAR(128) NOT NULL,
    client_type VARCHAR(128),
    ip_address VARCHAR(45),
    user_agent TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    started_at DATETIME(7) DEFAULT CURRENT_TIMESTAMP(7),
    last_activity_date DATETIME(7) DEFAULT CURRENT_TIMESTAMP(7),
    expires_at DATETIME(7) DEFAULT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_token_hash (token_hash),
    INDEX idx_user_id (user_id),
    INDEX idx_device_id (device_id),
    INDEX idx_is_active (is_active)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Quick Connect Codes table
CREATE TABLE IF NOT EXISTS quick_connect_codes (
    id CHAR(36) PRIMARY KEY,
    code VARCHAR(10) NOT NULL UNIQUE,
    secret VARCHAR(64) DEFAULT NULL,
    user_id CHAR(36) DEFAULT NULL,
    device_id VARCHAR(128) DEFAULT NULL,
    expires_at DATETIME(7) NOT NULL,
    secrets_expires_at DATETIME(7) DEFAULT NULL,
    is_used BOOLEAN DEFAULT FALSE,
    created_at DATETIME(7) DEFAULT CURRENT_TIMESTAMP(7),
    used_at DATETIME(7) DEFAULT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_code (code),
    INDEX idx_expires (expires_at),
    INDEX idx_is_used (is_used)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- User preferences table
CREATE TABLE IF NOT EXISTS user_preferences (
    id CHAR(36) PRIMARY KEY,
    user_id CHAR(36) NOT NULL,
    preference_key VARCHAR(255) NOT NULL,
    preference_value TEXT,
    created_at DATETIME(7) DEFAULT CURRENT_TIMESTAMP(7),
    updated_at DATETIME(7) DEFAULT CURRENT_TIMESTAMP(7) ON UPDATE CURRENT_TIMESTAMP(7),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE KEY unique_user_preference (user_id, preference_key),
    INDEX idx_user_id (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Insert default admin user (password: admin123)
INSERT INTO users (id, username, email, password_hash, is_admin) VALUES
('00000000-0000-0000-0000-000000000001', 'admin', 'admin@kabletown.local', '8c6976e5b5410415bde908bd4dee15dfb167a9c873fc4bb8a81f6f2ab448a918', TRUE);

-- Insert initial admin API key
INSERT INTO api_keys (id, user_id, api_key_hash, device_id, device_name, is_active) VALUES
('00000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000001', 
 sha2('admin-token-initial-key-hashed', 256), 'admin-initial-device', 'Admin Device', TRUE);
