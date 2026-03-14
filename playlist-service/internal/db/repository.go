package db

import (
	"fmt"
	"github.com/jmoiron/sqlx"
)

type PlaylistRepository struct {
	db *sqlx.DB
}

func NewPlaylistRepository(db *sqlx.DB) *PlaylistRepository {
	return &PlaylistRepository{db: db}
}

// CreatePlaylist inserts a new playlist
func (r *PlaylistRepository) CreatePlaylist(p *Playlist) error {
	result, err := r.db.NamedExec(`
		INSERT INTO Playlists (name, user_id, image_tag, row_version, date_created)
		VALUES (:Name, :UserID, :ImageTag, :RowVersion, :DateCreated)
	`, p)
	if err != nil {
		return fmt.Errorf("insert playlist: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("get last insert id: %w", err)
	}
	p.ID = int(id)
	return nil
}

// GetPlaylistByID fetches a playlist by ID
func (r *PlaylistRepository) GetPlaylistByID(id int) (*Playlist, error) {
	var p Playlist
	err := r.db.Get(&p, "SELECT * FROM Playlists WHERE id = ?", id)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// GetUserPlaylists fetches all playlists for a user
func (r *PlaylistRepository) GetUserPlaylists(userID string) ([]Playlist, error) {
	var playlists []Playlist
	err := r.db.Select(&playlists, "SELECT * FROM Playlists WHERE user_id = ? ORDER BY DateCreated DESC", userID)
	if err != nil {
		return nil, err
	}
	return playlists, nil
}

// DeletePlaylist removes a playlist
func (r *PlaylistRepository) DeletePlaylist(id int) error {
	_, err := r.db.Exec("DELETE FROM Playlists WHERE id = ?", id)
	return err
}

// GetPlaylistItemByID fetches a playlist item
func (r *PlaylistRepository) GetPlaylistItemByID(playlistID, itemID int) (*PlaylistItem, error) {
	var item PlaylistItem
	err := r.db.Get(&item,
		"SELECT * FROM PlaylistItems WHERE playlist_id = ? AND id = ?",
		playlistID, itemID,
	)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

// AddItemToPlaylist adds an item to a playlist
func (r *PlaylistRepository) AddItemToPlaylist(playlistID int, libraryItemID string) (*PlaylistItem, error) {
	result, err := r.db.Exec(
		"INSERT INTO PlaylistItems (playlist_id, library_item_id, row_version) VALUES (?, ?, 0)",
		playlistID, libraryItemID,
	)
	if err != nil {
		return nil, fmt.Errorf("insert playlist item: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("get last insert id: %w", err)
	}

	return &PlaylistItem{
		ID:            int(id),
		PlaylistID:    playlistID,
		LibraryItemID: libraryItemID,
	}, nil
}

// RemoveItemFromPlaylist removes an item from a playlist
func (r *PlaylistRepository) RemoveItemFromPlaylist(playlistID, itemID int) error {
	_, err := r.db.Exec("DELETE FROM PlaylistItems WHERE playlist_id = ? AND id = ?", playlistID, itemID)
	return err
}

// GetPlaylistItems fetches all items in a playlist
func (r *PlaylistRepository) GetPlaylistItems(playlistID int) ([]PlaylistItem, error) {
	var items []PlaylistItem
	err := r.db.Select(&items, "SELECT * FROM PlaylistItems WHERE playlist_id = ? ORDER BY id ASC", playlistID)
	if err != nil {
		return nil, err
	}
	return items, nil
}

// IncrementRowVersion updates the row version for optimistic locking
func (r *PlaylistRepository) IncrementRowVersion(table string, id int) error {
	query := ""
	if table == "Playlists" {
		query = "UPDATE Playlists SET row_version = row_version + 1 WHERE id = ?"
	} else if table == "PlaylistItems" {
		query = "UPDATE PlaylistItems SET row_version = row_version + 1 WHERE id = ?"
	} else {
		return fmt.Errorf("unknown table: %s", table)
	}

	_, err := r.db.Exec(query, id)
	return err
}
