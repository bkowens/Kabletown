package db

import (
	shared_db "github.com/bowens/kabletown/shared/db"
)

// GetCreateCollectionsSQL returns the SQL to create Collections table with P7 indexes
func GetCreateCollectionsSQL() string {
	return shared_db.GetCreateCollectionsSQL()
}

// GetCreateCollectionItemsSQL returns the SQL to create CollectionItems table with P7 indexes
func GetCreateCollectionItemsSQL() string {
	return shared_db.GetCreateCollectionItemsSQL()
}