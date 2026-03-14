-- API Keys table (shared across all services)
CREATE TABLE IF NOT EXISTS api_keys (
    Id CHAR(36) PRIMARY KEY,
    UserId CHAR(36) NOT NULL,
    DeviceId CHAR(36) NOT NULL,
    Token CHAR(64) NOT NULL UNIQUE,
    Name VARCHAR(255),
    AppId CHAR(36),
    AppName VARCHAR(255),
    AppVersion VARCHAR(50),
    DateCreated TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    DateLastUsed TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    IsActive BOOLEAN DEFAULT TRUE,
    ExpiresAt TIMESTAMP NULL,
    INDEX idx_token (Token),
    INDEX idx_user_id (UserId),
    FOREIGN KEY (UserId) REFERENCES Users(Id) ON DELETE CASCADE
);

-- Devices table (optional - tracks registered devices)
CREATE TABLE IF NOT EXISTS devices (
    Id CHAR(36) PRIMARY KEY,
    UserId CHAR(36) NOT NULL,
    Name VARCHAR(255),
    DeviceId CHAR(64) UNIQUE,
    AppName VARCHAR(255),
    AppVersion VARCHAR(50),
    DateRegistered TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    LastActivity TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_user_id (UserId),
    INDEX idx_device_id (DeviceId),
    FOREIGN KEY (UserId) REFERENCES Users(Id) ON DELETE CASCADE
);

-- Users table (if not already in item-service)
CREATE TABLE IF NOT EXISTS users (
    Id CHAR(36) PRIMARY KEY,
    Username VARCHAR(100) UNIQUE NOT NULL,
    Email VARCHAR(255),
    PasswordHash CHAR(64) NOT NULL,  -- SHA256 of password
    IsAdmin BOOLEAN DEFAULT FALSE,
    HasPassword BOOLEAN DEFAULT TRUE,
    EnableUser BOOLEAN DEFAULT TRUE,
    DateCreated TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    DateLastLogin TIMESTAMP NULL,
    AuthenticationProviderId VARCHAR(255) DEFAULT 'DefaultAuthenticationProvider',
    PasswordResetProviderId VARCHAR(255) DEFAULT 'DefaultPasswordResetProvider',
    INDEX idx_username (Username),
    INDEX idx_email (Email)
);

-- User Policies (permissions)
CREATE TABLE IF NOT EXISTS user_policies (
    UserId CHAR(36) PRIMARY KEY,
    IsAdmin BOOLEAN DEFAULT FALSE,
    EnableVideoPlaybackTranscoding BOOLEAN DEFAULT TRUE,
    EnableAudioPlaybackTranscoding BOOLEAN DEFAULT TRUE,
    EnableLiveTvManagement BOOLEAN DEFAULT FALSE,
    EnableAudioPlaybackRemuxing BOOLEAN DEFAULT TRUE,
    BlockedTags TEXT,  -- JSON array of blocked tags
    AllowedTags TEXT,  -- JSON array of allowed tags
    MAX(PreferredSubLanguage) VARCHAR(10),  -- Preferred subtitle language
    DEFAULT(AudioNormalization) BOOLEAN DEFAULT FALSE,
    FOREIGN KEY (UserId) REFERENCES Users(Id) ON DELETE CASCADE
);
