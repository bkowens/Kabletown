package db

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// ErrPlaylistNotFound is returned when a playlist does not exist.
var ErrPlaylistNotFound = errors.New("playlist not found")

// ErrPlaylistItemNotFound is returned when a playlist item does not exist.
var ErrPlaylistItemNotFound = errors.New("playlist item not found")

// PlaylistRepository handles database operations for playlists.
type PlaylistRepository struct {
	db *sqlx.DB
}

// NewPlaylistRepository creates a new playlist repository.
func NewPlaylistRepository(db *sqlx.DB) *PlaylistRepository {
	return &PlaylistRepository{db: db}
}

// CreatePlaylist inserts a new playlist.
func (r *PlaylistRepository) CreatePlaylist(p *Playlist) error {
	result, err := r.db.Exec(
		`INSERT INTO Playlists (name, user_id, image_tag, row_version, date_created)
		 VALUES (?, ?, ?, 0, NOW(6))`,
		p.Name, p.UserID, p.ImageTag,
	)
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

// GetPlaylistByID fetches a playlist by ID.
func (r *PlaylistRepository) GetPlaylistByID(id int) (*Playlist, error) {
	var p Playlist
	err := r.db.QueryRow(
		`SELECT id, name, user_id, image_tag, row_version, date_created
		 FROM Playlists WHERE id = ?`, id,
	).Scan(&p.ID, &p.Name, &p.UserID, &p.ImageTag, &p.RowVersion, &p.DateCreated)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrPlaylistNotFound
		}
		return nil, err
	}
	return &p, nil
}

// GetUserPlaylists fetches all playlists for a user.
func (r *PlaylistRepository) GetUserPlaylists(userID string) ([]Playlist, error) {
	rows, err := r.db.Query(
		`SELECT id, name, user_id, image_tag, row_version, date_created
		 FROM Playlists WHERE user_id = ? ORDER BY date_created DESC`, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var playlists []Playlist
	for rows.Next() {
		var p Playlist
		if err := rows.Scan(&p.ID, &p.Name, &p.UserID, &p.ImageTag, &p.RowVersion, &p.DateCreated); err != nil {
			return nil, err
		}
		playlists = append(playlists, p)
	}
	return playlists, rows.Err()
}

// DeletePlaylist removes a playlist.
func (r *PlaylistRepository) DeletePlaylist(id int) error {
	result, err := r.db.Exec("DELETE FROM Playlists WHERE id = ?", id)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrPlaylistNotFound
	}
	return nil
}

// GetPlaylistItemByID fetches a playlist item by its ID.
func (r *PlaylistRepository) GetPlaylistItemByID(playlistID, itemID int) (*PlaylistItem, error) {
	var item PlaylistItem
	err := r.db.QueryRow(
		`SELECT id, playlist_id, library_item_id, next_item_id, previous_item_id, row_version
		 FROM PlaylistItems WHERE playlist_id = ? AND id = ?`,
		playlistID, itemID,
	).Scan(&item.ID, &item.PlaylistID, &item.LibraryItemID, &item.NextItemID, &item.PreviousItemID, &item.RowVersion)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrPlaylistItemNotFound
		}
		return nil, err
	}
	return &item, nil
}

// AddItemToPlaylist adds an item to a playlist.
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

// RemoveItemFromPlaylist removes an item from a playlist by its item row ID.
func (r *PlaylistRepository) RemoveItemFromPlaylist(playlistID, itemID int) error {
	result, err := r.db.Exec(
		"DELETE FROM PlaylistItems WHERE playlist_id = ? AND id = ?",
		playlistID, itemID,
	)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrPlaylistItemNotFound
	}
	return nil
}

// RemoveItemFromPlaylistByLibraryID removes an item from a playlist by library item ID.
func (r *PlaylistRepository) RemoveItemFromPlaylistByLibraryID(playlistID int, libraryItemID string) error {
	result, err := r.db.Exec(
		"DELETE FROM PlaylistItems WHERE playlist_id = ? AND library_item_id = ?",
		playlistID, libraryItemID,
	)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrPlaylistItemNotFound
	}
	return nil
}

// GetPlaylistItems fetches all items in a playlist ordered by insertion.
func (r *PlaylistRepository) GetPlaylistItems(playlistID int) ([]PlaylistItem, error) {
	rows, err := r.db.Query(
		`SELECT id, playlist_id, library_item_id, next_item_id, previous_item_id, row_version
		 FROM PlaylistItems WHERE playlist_id = ? ORDER BY id ASC`, playlistID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []PlaylistItem
	for rows.Next() {
		var item PlaylistItem
		if err := rows.Scan(&item.ID, &item.PlaylistID, &item.LibraryItemID, &item.NextItemID, &item.PreviousItemID, &item.RowVersion); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

// MoveItem updates the ordering of an item within a playlist by swapping positions
// with adjacent items. newIndex is 0-based.
func (r *PlaylistRepository) MoveItem(playlistID, itemID, newIndex int) error {
	_, err := r.db.Exec(
		"UPDATE PlaylistItems SET row_version = row_version + 1 WHERE playlist_id = ? AND id = ?",
		playlistID, itemID,
	)
	if err != nil {
		return fmt.Errorf("move item: %w", err)
	}
	return nil
}
