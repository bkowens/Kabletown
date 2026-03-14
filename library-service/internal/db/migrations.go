package db

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

// RunMigrations creates the base_items and item_values tables if they do not exist,
// and applies all required performance indexes.
func RunMigrations(database *sqlx.DB) error {
	stmts := []string{
		// base_items table
		`CREATE TABLE IF NOT EXISTS base_items (
			Id            CHAR(36)      NOT NULL,
			Name          VARCHAR(255)  NOT NULL,
			Type          VARCHAR(50)   NOT NULL,
			IsFolder      TINYINT(1)    DEFAULT 0,
			ParentId      CHAR(36)      NULL,
			TopParentId   CHAR(36)      NULL,
			Path          VARCHAR(500)  NULL,
			Container     VARCHAR(100)  NULL,
			DurationTicks BIGINT        NULL,
			Size          BIGINT        NULL,
			Width         INT           NULL,
			Height        INT           NULL,
			ProductionYear INT          NULL,
			PremiereDate  TIMESTAMP     NULL,
			DateCreated   TIMESTAMP     DEFAULT CURRENT_TIMESTAMP,
			DateModified  TIMESTAMP     DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			ExtraData     JSON          NULL,
			AncestorIds   TEXT          NULL,
			PRIMARY KEY (Id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		// item_values table
		`CREATE TABLE IF NOT EXISTS item_values (
			ItemId          CHAR(36)     NOT NULL,
			ValueType       INT          NOT NULL,
			Value           VARCHAR(255) NOT NULL,
			NormalizedValue VARCHAR(255) NOT NULL,
			PRIMARY KEY (ItemId, ValueType, NormalizedValue)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		// UserData table
		`CREATE TABLE IF NOT EXISTS UserData (
			UserId                CHAR(36)   NOT NULL,
			ItemId                CHAR(36)   NOT NULL,
			Played                TINYINT(1) DEFAULT 0,
			PlayCount             INT        DEFAULT 0,
			IsFavorite            TINYINT(1) DEFAULT 0,
			PlaybackPositionTicks BIGINT     DEFAULT 0,
			LastPlayedDate        DATETIME   NULL,
			Rating                FLOAT      NULL,
			PRIMARY KEY (UserId, ItemId)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		// Indexes on base_items
		`CREATE INDEX IF NOT EXISTS idx_base_items_topparent_type
			ON base_items (TopParentId, Type)`,
		`CREATE INDEX IF NOT EXISTS idx_base_items_parentid
			ON base_items (ParentId)`,
		`CREATE INDEX IF NOT EXISTS idx_base_items_datecreated
			ON base_items (DateCreated DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_base_items_type
			ON base_items (Type)`,
		`CREATE INDEX IF NOT EXISTS idx_base_items_productionyear
			ON base_items (ProductionYear)`,

		// Index on item_values
		`CREATE INDEX IF NOT EXISTS idx_item_values_valuetype_norm
			ON item_values (ValueType, NormalizedValue)`,

		// Index on UserData
		`CREATE INDEX IF NOT EXISTS idx_userdata_userid_itemid
			ON UserData (UserId, ItemId)`,
	}

	// MySQL does not support CREATE INDEX IF NOT EXISTS; use the information_schema check approach.
	// Instead, run the CREATE TABLE stmts and then attempt index creation with error suppression.
	for i, stmt := range stmts {
		if _, err := database.Exec(stmt); err != nil {
			// Ignore duplicate index errors (MySQL error 1061).
			if isDuplicateIndexError(err) {
				continue
			}
			return fmt.Errorf("migrations: statement %d failed: %w", i, err)
		}
	}
	return nil
}

// isDuplicateIndexError returns true when MySQL signals that the index already exists.
func isDuplicateIndexError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	// MySQL error 1061: Duplicate key name
	return contains(msg, "Duplicate key name") || contains(msg, "1061")
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 || findStr(s, sub))
}

func findStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
