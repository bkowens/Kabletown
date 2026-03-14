package db

import (
	"fmt"
	"log"

	sharedDB "github.com/jellyfinhanced/shared/db"
	"github.com/jmoiron/sqlx"
)

// RunMigrations creates the Collections and CollectionItems tables with P7 indexes.
func RunMigrations(db *sqlx.DB) error {
	if _, err := db.Exec(sharedDB.GetCreateCollectionsSQL()); err != nil {
		return fmt.Errorf("failed to create Collections table: %w", err)
	}
	log.Println("collection-service: Collections table ready")

	if _, err := db.Exec(sharedDB.GetCreateCollectionItemsSQL()); err != nil {
		return fmt.Errorf("failed to create CollectionItems table: %w", err)
	}
	log.Println("collection-service: CollectionItems table ready")

	return nil
}
