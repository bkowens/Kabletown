# Database Schema Documentation

## Overview

Kabletown uses a single **MySQL 8.0** database shared across all services. Each service has clear ownership of specific tables, with foreign key constraints maintaining referential integrity.

### Database Configuration

```yaml
# docker-compose.yml
service: mysql
image: mysql:8.0
environment:
  MYSQL_ROOT_PASSWORD: root_password
  MYSQL_DATABASE: jellyfin
  MYSQL_USER: jellyfin
  MYSQL_PASSWORD: jellyfin_password
port: 3306
volumes:
  - mysql_data:/var/lib/mysql
```

### Connection Strings

**Item Service:**
```
mysql://jellyfin:password@mysql:3306/jellyfin?charset=utf8mb4&collation=utf8mb4_unicode_ci&parseTime=true&loc=Local
```

**Auth Service:**
```
sqlite3 (separate auth.db file) or mysql://jellyfin:password@mysql:3306/jellyfin
```

---

## Entity Relationship Diagram

```mermaid
erDiagram
    USERS ||--o{ API_KEYS : generates
    USERS ||--o{ USER_DATA : has
    USERS ||--o{ DEVICE_REGISTRATIONS : owns
    USERS { boolean IsAdmin
        string Username
        string PasswordHash
    }
    
    API_KEYS { string Token
        string DeviceId
        string AppName
        boolean IsActive
    }
    
    BASE_ITEMS ||--o{ USER_DATA : playstate
    BASE_ITEMS ||--o{ ITEM_VALUES : categorized_as
    BASE_ITEMS { string Type
        string TopParentId
        string Path
    }
    
    ITEM_VALUES { int ValueType
        string NormalizedValue
    }
    
    USER_DATA { boolean IsPlayed
        boolean IsFavorite
        long PlaybackPositionTicks
    }
```

---

## Table Specifications

### Auth Service Tables

#### `users`

**Owner:** Auth Service  
**Purpose:** User account storage with authentication credentials

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| `Id` | CHAR(36) | PRIMARY KEY | GUID identifier (dashed format) |
| `Username` | VARCHAR(100) | UNIQUE NOT NULL | Login username |
| `Email` | VARCHAR(255) | NULLABLE | User email address |
| `PasswordHash` | CHAR(64) | NOT NULL | SHA256 hash (64 hex chars) |
| `IsAdmin` | BOOLEAN | DEFAULT FALSE | Admin permission flag |
| `EnableUser` | BOOLEAN | DEFAULT TRUE | Account active flag |
| `HasPassword` | BOOLEAN | DEFAULT TRUE | Password set flag |
| `DateCreated` | TIMESTAMP | DEFAULT CURRENT_TIMESTAMP | Account creation time |
| `DateLastLogin` | TIMESTAMP | NULLABLE | Last successful login |

**Indexes:**
```sql
CREATE UNIQUE INDEX idx_users_username ON users(Username);
CREATE INDEX idx_users_enableuser ON users(EnableUser);
```

**Sample Data:**
```json
{
  "Id": "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11",
  "Username": "admin",
  "Email": "admin@example.com",
  "PasswordHash": "877b508e615eddf369d86b1e6a0e4e5e0f0e...",
  "IsAdmin": true,
  "EnableUser": true,
  "DateCreated": "2026-01-15T10:30:00Z",
  "DateLastLogin": "2026-03-13T16:45:00Z"
}
```

---

#### `api_keys`

**Owner:** Auth Service  
**Purpose:** Token-to-user mappings for API authentication

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| `Id` | CHAR(36) | PRIMARY KEY | GUID identifier |
| `UserId` | CHAR(36) | FOREIGN KEY → users(Id) | Owner of this token |
| `DeviceId` | CHAR(64) | NOT NULL | Device identifier (from auth header) |
| `Token` | CHAR(64) | UNIQUE NOT NULL | Random 256-bit hex token |
| `Name` | VARCHAR(255) | NULLABLE | User-assigned device name |
| `AppName` | VARCHAR(255) | NULLABLE | Client app name |
| `AppVersion` | VARCHAR(50) | NULLABLE | Client app version |
| `DateCreated` | TIMESTAMP | DEFAULT CURRENT_TIMESTAMP | Token creation time |
| `DateLastUsed` | TIMESTAMP | DEFAULT CURRENT_TIMESTAMP | Last request time |
| `IsActive` | BOOLEAN | DEFAULT TRUE | Revocation flag |

**Indexes:**
```sql
CREATE UNIQUE INDEX idx_api_keys_token ON api_keys(Token);
CREATE INDEX idx_api_keys_userid ON api_keys(UserId);
CREATE INDEX idx_api_keys_deviceid ON api_keys(DeviceId);
CREATE INDEX idx_api_keys_isactive ON api_keys(IsActive);
```

**Foreign Keys:**
```sql
ALTER TABLE api_keys 
  ADD CONSTRAINT fk_api_keys_user 
  FOREIGN KEY (UserId) REFERENCES users(Id) ON DELETE CASCADE;
```

**Sample Data:**
```json
{
  "Id": "b11cd33aa-1234-5678-9abc-def012345678",
  "UserId": "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11",
  "DeviceId": "web-client-desktop-12345",
  "Token": "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
  "Name": "Chrome Browser",
  "AppName": "Kabletown Web",
  "AppVersion": "1.0.0",
  "DateCreated": "2026-03-10T14:22:00Z",
  "DateLastUsed": "2026-03-13T16:50:00Z",
  "IsActive": true
}
```

---

#### `devices`

**Owner:** Auth Service  
**Purpose:** Device registration metadata

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| `Id` | CHAR(36) | PRIMARY KEY | GUID identifier |
| `UserId` | CHAR(36) | FOREIGN KEY → users(Id) | Owner of device (nullable for anonymous) |
| `Name` | VARCHAR(255) | NULLABLE | User-assigned device name |
| `DeviceId` | CHAR(64) | UNIQUE | Device identifier |
| `AppName` | VARCHAR(255) | NULLABLE | Client app name |
| `AppVersion` | VARCHAR(50) | NULLABLE | Client app version |
| `DateRegistered` | TIMESTAMP | DEFAULT CURRENT_TIMESTAMP | Registration time |
| `LastActivity` | TIMESTAMP | DEFAULT CURRENT_TIMESTAMP | Last seen time |

**Indexes:**
```sql
CREATE UNIQUE INDEX idx_devices_deviceid ON devices(DeviceId);
CREATE INDEX idx_devices_userid ON devices(UserId);
```

---

### Item Service Tables

#### `base_items`

**Owner:** Item Service  
**Purpose:** Core media metadata with P7-optimized indexes

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| `Id` | CHAR(36) | PRIMARY KEY | GUID identifier |
| `Name` | VARCHAR(255) | NOT NULL | Display name |
| `Type` | VARCHAR(50) | NOT NULL | BaseItemKind enum value |
| `IsFolder` | BOOLEAN | DEFAULT FALSE | Folder vs file flag |
| `ParentId` | CHAR(36) | NULLABLE | Immediate parent ID |
| `TopParentId` | CHAR(36) | NULLABLE | Root library folder ID |
| `Path` | VARCHAR(500) | NULLABLE | File system path |
| `Container` | VARCHAR(100) | NULLABLE | Media container (mkv, mp4, etc.) |
| `DurationTicks` | BIGINT | NULLABLE | Duration in 100-ns units |
| `Size` | BIGINT | NULLABLE | File size in bytes |
| `Width` | INT | NULLABLE | Video width |
| `Height` | INT | NULLABLE | Video height |
| `ProductionYear` | INT | NULLABLE | Release year |
| `PremiereDate` | TIMESTAMP | NULLABLE | Release date |
| `DateCreated` | TIMESTAMP | DEFAULT CURRENT_TIMESTAMP | Metadata creation |
| `DateModified` | TIMESTAMP | DEFAULT CURRENT_TIMESTAMP | Metadata update |
| `ExtraData` | JSON | NULLABLE | Provider IDs, custom fields |
| `AncestorIds` | TEXT | NULLABLE | Computed CTE result (all ancestors) |

**Indexes (P7 Optimized):**
```sql
-- Core query pattern: TopParentId + Type filtering
CREATE INDEX idx_base_items_topparent_type ON base_items(TopParentId, Type);

-- Parent folder queries
CREATE INDEX idx_base_items_parentid ON base_items(ParentId);

-- Recent additions (DateCreated DESC)
CREATE INDEX idx_base_items_datecreated ON base_items(DateCreated DESC);

-- Full-text search support
CREATE FULLTEXT INDEX idx_base_items_search ON base_items(Name, Overview);

-- Production year filtering
CREATE INDEX idx_base_items_year ON base_items(ProductionYear);

-- Recursive CTE support (self-join on ParentId)
-- AncestorIds column updated via trigger on INSERT/UPDATE
```

**Sample Data:**
```json
{
  "Id": "c22cd33bb-2345-6789-abcd-ef0123456789",
  "Name": "The Matrix",
  "Type": "Movie",
  "IsFolder": false,
  "ParentId": "d33de44cc-3456-789a-bcde-f01234567890",
  "TopParentId": "e44ef55dd-4567-89ab-cdef-012345678901",
  "Path": "/media/movies/The Matrix (1999)/The Matrix.mkv",
  "Container": "matroska",
  "DurationTicks": 8280000000000,  // 23 hours in ticks (23 * 60 * 60 * 10^7)
  "Size": 4294967296,
  "Width": 1920,
  "Height": 1080,
  "ProductionYear": 1999,
  "PremiereDate": "1999-03-31T00:00:00Z",
  "DateCreated": "2026-01-10T08:15:00Z",
  "DateModified": "2026-01-10T08:15:00Z",
  "ExtraData": {
    "ProviderIds": {
      "Tmdb": "603",
      "Imdb": "tt0133093"
    }
  }
}
```

---

#### `item_values`

**Owner:** Item Service  
**Purpose:** P6-normalized filtering for Genre, Studio, Artist (many-to-many)

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| `ItemId` | CHAR(36) | FOREIGN KEY → base_items(Id) | Reference to base item |
| `ValueType` | INT | NOT NULL | 2=Genre, 3=Studio, 4=Artist, 5=Collection, 6=Tag |
| `Value` | VARCHAR(255) | NOT NULL | Original display value |
| `NormalizedValue` | VARCHAR(255) | NOT NULL | Lowercase, trimmed, diacritics removed |

**Composite Primary Key:**
```sql
PRIMARY KEY (ItemId, ValueType, NormalizedValue)
```

**Indexes (P6 Optimization):**
```sql
-- Genre filtering: WHERE ValueType=2 AND NormalizedValue='action'
CREATE INDEX idx_item_values_genre_lookup ON item_values(ValueType, NormalizedValue);

-- Studio filtering: WHERE ValueType=3 AND NormalizedValue='pixar'
CREATE INDEX idx_item_values_studio_lookup ON item_values(ValueType, NormalizedValue);

-- Artist filtering: WHERE ValueType=4 AND NormalizedValue='john doe'
CREATE INDEX idx_item_values_artist_lookup ON item_values(ValueType, NormalizedValue);

-- All items with a specific value
CREATE INDEX idx_item_values_valtype ON item_values(ValueType);
```

**Sample Data:**
```json
{
  "ItemId": "c22cd33bb-2345-6789-abcd-ef0123456789",
  "ValueType": 2,  // Genre
  "Value": "Action",
  "NormalizedValue": "action"
}
```

**Normalization Rules:**
- Lowercase all characters
- Remove leading/trailing whitespace
- Remove diacritics (café → cafe)
- Collapse multiple spaces

---

#### `user_data`

**Owner:** Item Service  
**Purpose:** Per-user playstate, favorites, ratings

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| `UserId` | CHAR(36) | FOREIGN KEY → users(Id) | User identifier |
| `ItemId` | CHAR(36) | FOREIGN KEY → base_items(Id) | Item identifier |
| `IsPlayed` | BOOLEAN | DEFAULT FALSE | Watch completion flag |
| `IsFavorite` | BOOLEAN | DEFAULT FALSE | User favorite flag |
| `PlaybackPositionTicks` | BIGINT | DEFAULT 0 | Resume position (100-ns units) |
| `Rating` | INT | NULLABLE | User rating (1-10 or 1-5) |
| `PlayCount` | INT | DEFAULT 0 | Number of plays |

**Composite Primary Key:**
```sql
PRIMARY KEY (UserId, ItemId)
```

**Indexes:**
```sql
-- Recently played (ORDER BY PlaybackPositionTicks DESC)
CREATE INDEX idx_user_data_position ON user_data(UserId, PlaybackPositionTicks DESC);

-- Recently added (items played within last N days)
CREATE INDEX idx_user_data_played ON user_data(UserId, IsPlayed);

-- Favorite items
CREATE INDEX idx_user_data_favorite ON user_data(UserId, IsFavorite);

-- Next episode lookup (unwatched episodes)
CREATE INDEX idx_user_data_unwatched ON user_data(UserId, IsPlayed) WHERE IsPlayed = false;
```

**Sample Data:**
```json
{
  "UserId": "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11",
  "ItemId": "c22cd33bb-2345-6789-abcd-ef0123456789",
  "IsPlayed": false,
  "IsFavorite": true,
  "PlaybackPositionTicks": 1800000000000,  // 3 hours in
  "Rating": 9,
  "PlayCount": 2
}
```

---

#### `user_policies`

**Owner:** Auth Service  
**Purpose:** User permission and policy settings

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| `UserId` | CHAR(36) | PRIMARY KEY, FOREIGN KEY → users(Id) | User identifier |
| `EnableVideoTranscoding` | BOOLEAN | DEFAULT TRUE | Video transcoding allowed |
| `EnableAudioTranscoding` | BOOLEAN | DEFAULT TRUE | Audio transcoding allowed |
| `MaxStreamingBitrate` | INT | NULLABLE | Max bits/sec for streaming |
| `EnablePlaybackRemuxing` | BOOLEAN | DEFAULT TRUE | Allow remuxing (stream copy) |

**Indexes:**
- Primary key serves as unique lookup

**Sample Data:**
```json
{
  "UserId": "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11",
  "EnableVideoTranscoding": true,
  "EnableAudioTranscoding": true,
  "MaxStreamingBitrate": 20000000,  // 20 Mbps
  "EnablePlaybackRemuxing": true
}
```

---

## Migration Scripts

### Initial Schema Setup

```bash
# Create MySQL database
mysql -h localhost -u root -p -e "CREATE DATABASE IF NOT EXISTS jellyfin CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"

# Import schema
mysql -h localhost -u jellyfin -p jellyfin < Kabletown/migrations/schema.sql
```

### Migration Order

1. **001_users.sql** - User accounts
2. **002_api_keys.sql** - Token management
3. **003_devices.sql** - Device registrations
4. **004_base_items.sql** - Core media metadata with indexes
5. **005_item_values.sql** - P6 normalization table
6. **006_user_data.sql** - Per-user playstate
7. **007_user_policies.sql** - Permission flags
8. **008_triggers.sql** - Automatic AncestorIds computation
9. **009_admin_user.sql** - Default admin account

---

## Data Models

### InternalItemsQuery (Item Repository)

```sql
-- Example: Query filtered by TopParentId + IncludeItemTypes + Pagination
SELECT 
    b.*,
    ud.IsPlayed,
    ud.PlaybackPositionTicks,
    ud.IsFavorite
FROM base_items b
LEFT JOIN user_data ud ON b.Id = ud.ItemId AND ud.UserId = ?
WHERE 
    b.TopParentId = ?
    AND b.Type IN ('Movie', 'Episode')
    AND b.DateCreated >= ?
ORDER BY b.DateCreated DESC
LIMIT ? OFFSET ?;

-- Join with ItemValues for Genre filtering
SELECT b.*
FROM base_items b
JOIN item_values iv ON b.Id = iv.ItemId
WHERE 
    b.TopParentId = ?
    AND iv.ValueType = 2  -- Genre
    AND iv.NormalizedValue IN ('action', 'scifi')
```

### Transcode Job State (In Memory, Not Persisted)

```go
type TranscodingJob struct {
    Id: string               // UUID
    Path: string             // /transcode/{filename}.ts
    Process: *exec.Cmd       // FFmpeg subprocess
    PlaySessionId: string    // Unique session identifier
    DeviceId: string         // Client device identifier
    ActiveRequestCount: int  // Concurrent segment requests (HLS)
    LastPingTime: time.Time  // Last client checkin
    KillTimer: *time.Timer   // Auto-kill after timeout
}
```

---

## Performance Considerations

### Index Usage

**High-Frequency Queries:**

1. **Library Home (TopParentId + Type):**
   ```sql
   -- Uses: idx_base_items_topparent_type
   WHERE TopParentId = ? AND Type = 'Movie'
   ```

2. **Recently Added (DateCreated DESC):**
   ```sql
   -- Uses: idx_base_items_datecreated
   WHERE TopParentId = ? AND DateCreated > ?
   ORDER BY DateCreated DESC LIMIT 20
   ```

3. **Genre Filter (P6):**
   ```sql
   -- Uses: idx_item_values_genre_lookup
   WHERE ValueType = 2 AND NormalizedValue = 'action'
   ```

4. **Playstate Lookup:**
   ```sql
   -- Uses: PRIMARY KEY (UserId, ItemId)
   WHERE UserId = ? AND ItemId = ?
   ```

### Query Optimization Tips

- **Avoid SELECT ***: Use specific columns to reduce memory usage
- **Pagination**: Always use LIMIT + OFFSET for large result sets
- **Explain Plans**: Run EXPLAIN on complex queries to verify index usage
- **Connection Pool**: Item Service uses 20 max connections, 5 idle

---

## Backup Strategy

### Daily Backup Script

```bash
#!/bin/bash
BACKUP_DIR="/backups/mysql"
DATE=$(date +%Y%m%d_%H%M%S)

# Backup to gzipped SQL dump
mysqldump -h localhost -u jellyfin -p"$(JELLYFIN_PWD)" \
  --single-transaction \
  --routines \
  --triggers \
  --events \
  jellyfin | gzip > "$BACKUP_DIR/backup_$DATE.sql.gz"

# Retention: Keep 7 days of backups
find "$BACKUP_DIR" -name "backup_*.sql.gz" -mtime +7 -delete
```

### Recommended Settings

- **Automated**: Cron job at 2:00 AM daily
- **Offsite**: Copy to S3/GCS after backup
- **Test Restores**: Monthly restore verification
