package db

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// DisplayPrefsRepository handles DisplayPreferences table operations.
type DisplayPrefsRepository struct {
	db *sqlx.DB
}

// NewDisplayPrefsRepository creates a new DisplayPrefsRepository.
func NewDisplayPrefsRepository(database *sqlx.DB) *DisplayPrefsRepository {
	return &DisplayPrefsRepository{db: database}
}

// GetDisplayPreferences fetches the stored JSON data for the given key, user, client triple.
// Returns empty string if not found.
func (r *DisplayPrefsRepository) GetDisplayPreferences(id, userID, client string) (string, error) {
	var data string
	err := r.db.QueryRow(
		`SELECT COALESCE(Data,'') FROM DisplayPreferences
		 WHERE Id = ? AND UserId = ? AND Client = ? LIMIT 1`,
		id, userID, client,
	).Scan(&data)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
		return "", fmt.Errorf("display_prefs_repository.GetDisplayPreferences: %w", err)
	}
	return data, nil
}

// UpsertDisplayPreferences inserts or replaces display preferences data.
func (r *DisplayPrefsRepository) UpsertDisplayPreferences(id, userID, client, data string) error {
	_, err := r.db.Exec(
		`INSERT INTO DisplayPreferences (Id, UserId, Client, Data) VALUES (?, ?, ?, ?)
		 ON DUPLICATE KEY UPDATE Data = ?`,
		id, userID, client, data, data,
	)
	if err != nil {
		return fmt.Errorf("display_prefs_repository.UpsertDisplayPreferences: %w", err)
	}
	return nil
}
