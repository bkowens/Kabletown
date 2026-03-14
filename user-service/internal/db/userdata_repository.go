package db

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

// UserData represents a row from the UserData table.
type UserData struct {
	UserId                string    `db:"UserId"`
	ItemId                string    `db:"ItemId"`
	Played                bool      `db:"Played"`
	PlayCount             int       `db:"PlayCount"`
	IsFavorite            bool      `db:"IsFavorite"`
	PlaybackPositionTicks int64     `db:"PlaybackPositionTicks"`
	LastPlayedDate        *time.Time `db:"LastPlayedDate"`
	Rating                *float64  `db:"Rating"`
}

// UserDataRepository provides data access for the UserData table.
type UserDataRepository struct {
	db *sqlx.DB
}

// NewUserDataRepository creates a new UserDataRepository.
func NewUserDataRepository(database *sqlx.DB) *UserDataRepository {
	return &UserDataRepository{db: database}
}

// GetUserData fetches UserData for a specific user/item pair.
// Returns nil if no row exists.
func (r *UserDataRepository) GetUserData(userID, itemID string) (*UserData, error) {
	var ud UserData
	err := r.db.QueryRowx(
		`SELECT UserId, ItemId, Played, PlayCount, IsFavorite,
		        PlaybackPositionTicks, LastPlayedDate, Rating
		 FROM UserData WHERE UserId = ? AND ItemId = ? LIMIT 1`,
		userID, itemID,
	).StructScan(&ud)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("userdata_repository.GetUserData: %w", err)
	}
	return &ud, nil
}

// MarkPlayed marks an item as played and increments the play count.
func (r *UserDataRepository) MarkPlayed(userID, itemID string, datePlayed *time.Time) error {
	if datePlayed == nil {
		now := time.Now().UTC()
		datePlayed = &now
	}
	_, err := r.db.Exec(
		`INSERT INTO UserData (UserId, ItemId, Played, PlayCount, LastPlayedDate)
		 VALUES (?, ?, 1, 1, ?)
		 ON DUPLICATE KEY UPDATE
		   Played = 1,
		   PlayCount = PlayCount + 1,
		   LastPlayedDate = ?`,
		userID, itemID, datePlayed, datePlayed,
	)
	if err != nil {
		return fmt.Errorf("userdata_repository.MarkPlayed: %w", err)
	}
	return nil
}

// MarkUnplayed marks an item as unplayed and resets position.
func (r *UserDataRepository) MarkUnplayed(userID, itemID string) error {
	_, err := r.db.Exec(
		`INSERT INTO UserData (UserId, ItemId, Played, PlayCount, PlaybackPositionTicks)
		 VALUES (?, ?, 0, 0, 0)
		 ON DUPLICATE KEY UPDATE
		   Played = 0,
		   PlaybackPositionTicks = 0`,
		userID, itemID,
	)
	if err != nil {
		return fmt.Errorf("userdata_repository.MarkUnplayed: %w", err)
	}
	return nil
}

// SetFavorite updates the IsFavorite flag for a user/item pair.
func (r *UserDataRepository) SetFavorite(userID, itemID string, favorite bool) error {
	fav := 0
	if favorite {
		fav = 1
	}
	_, err := r.db.Exec(
		`INSERT INTO UserData (UserId, ItemId, IsFavorite)
		 VALUES (?, ?, ?)
		 ON DUPLICATE KEY UPDATE IsFavorite = ?`,
		userID, itemID, fav, fav,
	)
	if err != nil {
		return fmt.Errorf("userdata_repository.SetFavorite: %w", err)
	}
	return nil
}

// SavePosition persists the playback position ticks.
func (r *UserDataRepository) SavePosition(userID, itemID string, ticks int64) error {
	_, err := r.db.Exec(
		`INSERT INTO UserData (UserId, ItemId, PlaybackPositionTicks)
		 VALUES (?, ?, ?)
		 ON DUPLICATE KEY UPDATE PlaybackPositionTicks = ?`,
		userID, itemID, ticks, ticks,
	)
	if err != nil {
		return fmt.Errorf("userdata_repository.SavePosition: %w", err)
	}
	return nil
}
