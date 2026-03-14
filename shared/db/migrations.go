package db

import (
	"github.com/jmoiron/sqlx"
)

// RunItemValuesMigrations creates the ItemValues and ItemValuesMap tables
// This should be called during service startup before using ItemValueRepository
func RunItemValuesMigrations(db *sqlx.DB) error {
	// Create ItemValues table
	if _, err := db.Exec(GetCreateItemValuesSQL()); err != nil {
		return err
	}

	// Create ItemValuesMap table
	if _, err := db.Exec(GetCreateItemValuesMapSQL()); err != nil {
		return err
	}

	return nil
}

// IsItemValuesTableExist checks if ItemValues tables have been migrated
func IsItemValuesTableExist(db *sqlx.DB) (bool, error) {
	var count int
	err := db.QueryRow(`
		SELECT COUNT(*) 
		FROM information_schema.tables 
		WHERE table_schema = DATABASE() 
		AND table_name IN ('ItemValues', 'ItemValuesMap')
	`).Scan(&count)
	if err != nil {
		return false, err
	}
	return count == 2, nil
}
