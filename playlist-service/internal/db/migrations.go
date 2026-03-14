package db

import (
	"fmt"
	"log"

	sharedDB "github.com/jellyfinhanced/shared/db"
	"github.com/jmoiron/sqlx"
)

// RunMigrations creates the Playlists and PlaylistItems tables with P7 indexes.
func RunMigrations(db *sqlx.DB) error {
	if _, err := db.Exec(sharedDB.GetCreatePlaylistsSQL()); err != nil {
		return fmt.Errorf("failed to create Playlists table: %w", err)
	}
	log.Println("playlist-service: Playlists table ready")

	if _, err := db.Exec(sharedDB.GetCreatePlaylistItemsSQL()); err != nil {
		return fmt.Errorf("failed to create PlaylistItems table: %w", err)
	}
	log.Println("playlist-service: PlaylistItems table ready")

	return nil
}
