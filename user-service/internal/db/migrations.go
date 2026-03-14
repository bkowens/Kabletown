package db

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

// RunMigrations creates all required tables if they do not already exist.
func RunMigrations(database *sqlx.DB) error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS Users (
			Id              CHAR(36)     PRIMARY KEY,
			Name            VARCHAR(255) NOT NULL,
			Password        VARCHAR(255),
			IsDisabled      TINYINT(1)   DEFAULT 0,
			IsHidden        TINYINT(1)   DEFAULT 0,
			PrimaryImageTag VARCHAR(255),
			Configuration   LONGTEXT,
			Policy          LONGTEXT
		)`,
		`CREATE TABLE IF NOT EXISTS Permissions (
			Id     INT AUTO_INCREMENT PRIMARY KEY,
			UserId CHAR(36) NOT NULL,
			Kind   INT      NOT NULL,
			Value  TINYINT(1) NOT NULL DEFAULT 0
		)`,
		`CREATE TABLE IF NOT EXISTS UserData (
			UserId                CHAR(36)  NOT NULL,
			ItemId                CHAR(36)  NOT NULL,
			Played                TINYINT(1) DEFAULT 0,
			PlayCount             INT        DEFAULT 0,
			IsFavorite            TINYINT(1) DEFAULT 0,
			PlaybackPositionTicks BIGINT     DEFAULT 0,
			LastPlayedDate        DATETIME,
			Rating                FLOAT,
			PRIMARY KEY (UserId, ItemId)
		)`,
		`CREATE TABLE IF NOT EXISTS DisplayPreferences (
			Id     VARCHAR(255) NOT NULL,
			UserId CHAR(36)     NOT NULL,
			Client VARCHAR(255) NOT NULL DEFAULT 'emby',
			Data   LONGTEXT,
			PRIMARY KEY (Id, UserId, Client)
		)`,
	}

	for _, stmt := range statements {
		if _, err := database.Exec(stmt); err != nil {
			return fmt.Errorf("migrations: %w", err)
		}
	}
	return nil
}
