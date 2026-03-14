package db

import (
	"fmt"
	"log"

	shared_db "github.com/bowens/kabletown/shared/db"
	"github.com/jmoiron/sqlx"
)

// EnsurePlaylistsTables creates the Playlists and PlaylistItems tables with P7 indexes
func EnsurePlaylistsTables(db *sqlx.DB) error {
	// Playlists table
	createPlaylists := shared_db.GetCreatePlaylistsSQL()
	if _, err := db.Exec(createPlaylists); err != nil {
		return fmt.Errorf("failed to create Playlists table: %w", err)
	}
	log.Println("✅ Playlists table ready")

	// PlaylistItems table
	createPlaylistItems := shared_db.GetCreatePlaylistItemsSQL()
	if _, err := db.Exec(createPlaylistItems); err != nil {
		return fmt.Errorf("failed to create PlaylistItems table: %w", err)
	}
	log.Println("✅ PlaylistItems table ready")

	return nil
}