package db

import (
	"database/sql"
	"errors"
	"time"

	"github.com/jellyfinhanced/shared/types"
	"github.com/jmoiron/sqlx"
)

// CollectionRepository handles database operations for collections.
type CollectionRepository struct {
	db *sqlx.DB
}

// NewCollectionRepository creates a new collection repository.
func NewCollectionRepository(db *sqlx.DB) *CollectionRepository {
	return &CollectionRepository{db: db}
}

// CreateCollection creates a new collection.
func (r *CollectionRepository) CreateCollection(name, userID string, imageTag string) (*Collection, error) {
	query := `
		INSERT INTO Collections (name, user_id, image_tag, row_version, date_created)
		VALUES (?, ?, ?, 0, NOW(6))
	`

	result, err := r.db.Exec(query, name, userID, imageTag)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return &Collection{
		ID:          int(id),
		Name:        name,
		UserID:      userID,
		ImageTag:    imageTag,
		RowVersion:  0,
		DateCreated: types.NewJellyfinTime(time.Now()),
	}, nil
}

// GetCollectionByID fetches a collection by ID.
func (r *CollectionRepository) GetCollectionByID(id int) (*Collection, error) {
	query := `
		SELECT id, name, user_id, image_tag, row_version, date_created
		FROM Collections
		WHERE id = ?
	`

	var coll Collection
	err := r.db.QueryRow(query, id).Scan(
		&coll.ID, &coll.Name, &coll.UserID, &coll.ImageTag,
		&coll.RowVersion, &coll.DateCreated,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrCollectionNotFound
		}
		return nil, err
	}

	return &coll, nil
}

// GetCollectionsByUserID fetches all collections for a user.
func (r *CollectionRepository) GetCollectionsByUserID(userID string) ([]*Collection, error) {
	query := `
		SELECT id, name, user_id, image_tag, row_version, date_created
		FROM Collections
		WHERE user_id = ?
		ORDER BY date_created DESC
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var collections []*Collection
	for rows.Next() {
		var coll Collection
		err := rows.Scan(
			&coll.ID, &coll.Name, &coll.UserID, &coll.ImageTag,
			&coll.RowVersion, &coll.DateCreated,
		)
		if err != nil {
			return nil, err
		}
		collections = append(collections, &coll)
	}

	return collections, rows.Err()
}

// GetCollectionDetails fetches a collection with its item count.
func (r *CollectionRepository) GetCollectionDetails(id int) (*CollectionDetails, error) {
	query := `
		SELECT c.id, c.name, c.user_id, c.image_tag, c.row_version, c.date_created,
		       COUNT(ci.id) as item_count
		FROM Collections c
		LEFT JOIN CollectionItems ci ON c.id = ci.collection_id
		WHERE c.id = ?
		GROUP BY c.id, c.name, c.user_id, c.image_tag, c.row_version, c.date_created
	`

	var details CollectionDetails
	err := r.db.QueryRow(query, id).Scan(
		&details.ID, &details.Name, &details.UserID, &details.ImageTag,
		&details.RowVersion, &details.DateCreated, &details.ItemCount,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrCollectionNotFound
		}
		return nil, err
	}

	return &details, nil
}

// UpdateCollection updates a collection name.
func (r *CollectionRepository) UpdateCollection(id int, name string) error {
	query := `
		UPDATE Collections
		SET name = ?, row_version = row_version + 1
		WHERE id = ?
	`

	result, err := r.db.Exec(query, name, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrCollectionNotFound
	}

	return nil
}

// DeleteCollection removes a collection and its items (cascade).
func (r *CollectionRepository) DeleteCollection(id int) error {
	query := `DELETE FROM Collections WHERE id = ?`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrCollectionNotFound
	}

	return nil
}

// AddItemToCollection adds a library item to a collection.
func (r *CollectionRepository) AddItemToCollection(collectionID int, libraryItemID string, nextItemID *int) (*CollectionItem, error) {
	prevID, err := r.getLastItemInCollection(collectionID)
	if err != nil {
		return nil, err
	}

	query := `
		INSERT INTO CollectionItems (collection_id, library_item_id, previous_item_id, next_item_id, row_version)
		VALUES (?, ?, ?, ?, 0)
	`

	result, err := r.db.Exec(query, collectionID, libraryItemID, prevID, nextItemID)
	if err != nil {
		return nil, err
	}

	itemID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return &CollectionItem{
		ID:             int(itemID),
		CollectionID:   collectionID,
		LibraryItemID:  libraryItemID,
		PreviousItemID: prevID,
		NextItemID:     nextItemID,
		RowVersion:     0,
	}, nil
}

// RemoveItemFromCollection removes a library item from a collection.
func (r *CollectionRepository) RemoveItemFromCollection(collectionID int, libraryItemID string) error {
	query := `
		DELETE FROM CollectionItems
		WHERE collection_id = ? AND library_item_id = ?
	`

	result, err := r.db.Exec(query, collectionID, libraryItemID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrCollectionItemNotFound
	}

	return nil
}

// GetCollectionItems retrieves all item IDs in a collection ordered by insertion.
func (r *CollectionRepository) GetCollectionItems(collectionID int) ([]string, error) {
	query := `
		SELECT library_item_id
		FROM CollectionItems
		WHERE collection_id = ?
		ORDER BY id ASC
	`

	rows, err := r.db.Query(query, collectionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []string
	for rows.Next() {
		var itemID string
		if err := rows.Scan(&itemID); err != nil {
			return nil, err
		}
		items = append(items, itemID)
	}

	return items, rows.Err()
}

// getLastItemInCollection returns the id of the last item in the collection,
// or nil if the collection is empty.
func (r *CollectionRepository) getLastItemInCollection(collectionID int) (*int, error) {
	query := `
		SELECT id FROM CollectionItems
		WHERE collection_id = ?
		ORDER BY id DESC LIMIT 1
	`

	var id int
	err := r.db.QueryRow(query, collectionID).Scan(&id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &id, nil
}
