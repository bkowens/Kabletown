package db

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

// RunMigrations creates the tables required by the auth-service if they do not exist.
// All column names are PascalCase to match the Jellyfin MySQL schema.
func RunMigrations(database *sqlx.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS Users (
			Id               CHAR(36)     NOT NULL PRIMARY KEY,
			Name             VARCHAR(255) NOT NULL,
			Password         VARCHAR(255) NOT NULL DEFAULT '',
			IsDisabled       TINYINT(1)   NOT NULL DEFAULT 0,
			IsHidden         TINYINT(1)   NOT NULL DEFAULT 0,
			PrimaryImageTag  VARCHAR(255) NOT NULL DEFAULT '',
			UNIQUE KEY uq_users_name (Name)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		`CREATE TABLE IF NOT EXISTS Devices (
			Id                CHAR(36)     NOT NULL PRIMARY KEY,
			UserId            CHAR(36)     NOT NULL,
			DeviceId          VARCHAR(255) NOT NULL,
			AccessToken       VARCHAR(255) NOT NULL,
			FriendlyName      VARCHAR(255) NOT NULL DEFAULT '',
			AppName           VARCHAR(255) NOT NULL DEFAULT '',
			AppVersion        VARCHAR(50)  NOT NULL DEFAULT '',
			Created           DATETIME     NOT NULL,
			DateLastActivity  DATETIME     NOT NULL,
			UNIQUE KEY uq_devices_token   (AccessToken),
			UNIQUE KEY uq_devices_devuser (DeviceId, UserId),
			KEY idx_devices_userid        (UserId)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		`CREATE TABLE IF NOT EXISTS ApiKeys (
			Id           CHAR(36)     NOT NULL PRIMARY KEY,
			AccessToken  VARCHAR(255) NOT NULL,
			Name         VARCHAR(255) NOT NULL DEFAULT '',
			DateCreated  DATETIME     NOT NULL,
			IsAdmin      TINYINT(1)   NOT NULL DEFAULT 1,
			UNIQUE KEY uq_apikeys_token (AccessToken)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,

		`CREATE TABLE IF NOT EXISTS Permissions (
			Id      INT AUTO_INCREMENT NOT NULL PRIMARY KEY,
			UserId  CHAR(36)  NOT NULL,
			Kind    INT       NOT NULL,
			Value   TINYINT(1) NOT NULL DEFAULT 0,
			KEY idx_permissions_user (UserId),
			UNIQUE KEY uq_permissions_user_kind (UserId, Kind)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
	}

	for _, stmt := range stmts {
		if _, err := database.Exec(stmt); err != nil {
			return fmt.Errorf("db.RunMigrations: %w", err)
		}
	}
	return nil
}
