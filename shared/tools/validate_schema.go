// +build ignore
// go run main.go

package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	// Validate the expected schema exists
	dsn := fmt.Sprintf("%s:%s@tcp(%s:3306)/%s", 
		os.Getenv("DB_USER"), 
		os.Getenv("DB_PASSWORD"), 
		os.Getenv("DB_HOST"),
		os.Getenv("DB_NAME"))

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Check critical tables exist
	tables := []string{"BaseItems", "Users", "Devices", "UserData", "ItemValues", "ItemValuesMap", "Collections", "CollectionItems", "Playlists", "PlaylistItems"}

	for _, table := range tables {
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = ? AND table_name = ?",
			"jellyfin", table).Scan(&count)
		if err != nil {
			log.Fatalf("Error checking table %s: %v", table, err)
		}
		if count == 0 {
			log.Fatalf("Table %s not found in database", table)
		}
		log.Printf("Table %s: OK", table)
	}

	// Verify critical indexes exist (matching C# migration: 20260309000000_AddPerformanceIndexes.cs)
	indexes := []string{
		"IX_BaseItems_Type_IsVirtualItem_SortName",
		"IX_BaseItems_ParentId_IsVirtualItem_Type",
		"IX_UserData_UserId_IsFavorite",
		"IX_UserData_UserId_Played",
		"IX_ItemValues_Type_Name",
		"IX_ItemValuesMap_ItemValueId",
	}

	for _, idx := range indexes {
		var count int
		err := db.QueryRow(`
			SELECT COUNT(*) 
			FROM information_schema.statistics 
			WHERE table_schema = ? AND index_name = ?`,
			"jellyfin", idx).Scan(&count)
		if err != nil {
			log.Fatalf("Error checking index %s: %v", idx, err)
		}
		if count == 0 {
			log.Printf("Warning: Index %s not found", idx)
		} else {
			log.Printf("Index %s: OK", idx)
		}
	}

	log.Println("Database schema validation complete")
}