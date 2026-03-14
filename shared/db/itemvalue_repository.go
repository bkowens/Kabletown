package db

import (
	"strings"

	"github.com/jmoiron/sqlx"
)

// ItemValueRepository handles ItemValues CRUD operations
type ItemValueRepository struct {
	db *sqlx.DB
}

// NewItemValueRepository creates a new ItemValueRepository
func NewItemValueRepository(db *sqlx.DB) *ItemValueRepository {
	return &ItemValueRepository{db: db}
}

// ItemValue represents a categorized value (Genre, Studio, Artist, etc.)
type ItemValue struct {
	ID   string
	Name string
	Type string // Database stores type as string
	ImageTag *string
}

// CreateOrUpdate creates or updates an ItemValue
// Returns the ID of the created/updated value
func (r *ItemValueRepository) CreateOrUpdate(value *ItemValue) (string, error) {
	query := `
		INSERT INTO ItemValues (id, name, type, image_tag, row_version)
		VALUES (?, ?, ?, ?, 0)
		ON DUPLICATE KEY UPDATE
			name = VALUES(name),
			image_tag = VALUES(image_tag),
			row_version = row_version + 1
	`;

	// Generate ID if not provided
	if value.ID == "" {
		value.ID = generateItemValueID(value.Name)
	}

	_, err := r.db.Exec(query, value.ID, value.Name, value.Type, value.ImageTag)
	if err != nil {
		return "", err
	}

	return value.ID, nil
}

// GetByID retrieves an ItemValue by ID
func (r *ItemValueRepository) GetByID(id string) (*ItemValue, error) {
	value := &ItemValue{}
	err := r.db.Get(value, `SELECT id, name, type, image_tag FROM ItemValues WHERE id = ?`, id)
	if err != nil {
		return nil, err
	}
	return value, nil
}

// GetByNameType retrieves an ItemValue by name and type
func (r *ItemValueRepository) GetByNameType(name, valueType string) (*ItemValue, error) {
	value := &ItemValue{}
	err := r.db.Get(value, `SELECT id, name, type, image_tag FROM ItemValues WHERE name = ? AND type = ?`, name, valueType)
	if err != nil {
		return nil, err
	}
	return value, nil
}

// GetAllByType retrieves all ItemValues of a specific type
func (r *ItemValueRepository) GetAllByType(valueType string) ([]ItemValue, error) {
	var values []ItemValue
	err := r.db.Select(&values, `SELECT id, name, type, image_tag FROM ItemValues WHERE type = ? ORDER BY name`, valueType)
	if err != nil {
		return nil, err
	}
	return values, nil
}

// GetValuesForItem retrieves all ItemValues mapped to a specific library item
func (r *ItemValueRepository) GetValuesForItem(libraryItemID string) ([]ItemValue, error) {
	var values []ItemValue
	err := r.db.Select(&values, `
		SELECT iv.id, iv.name, iv.type, iv.image_tag
		FROM ItemValues iv
		INNER JOIN ItemValuesMap ivm ON iv.id = ivm.item_value_id
		WHERE ivm.library_item_id = ?
		ORDER BY iv.type, iv.name
	`, libraryItemID)
	if err != nil {
		return nil, err
	}
	return values, nil
}

// MapItemValue maps an ItemValue to a library item
func (r *ItemValueRepository) MapItemValue(libraryItemID, itemValueID string) error {
	query := `
		INSERT INTO ItemValuesMap (library_item_id, item_value_id, row_version)
		VALUES (?, ?, 0)
		ON DUPLICATE KEY UPDATE row_version = row_version + 1
	`
	_, err := r.db.Exec(query, libraryItemID, itemValueID)
	return err
}

// UnmapItemValue removes the mapping between an ItemValue and a library item
func (r *ItemValueRepository) UnmapItemValue(libraryItemID, itemValueID string) error {
	_, err := r.db.Exec(
		`DELETE FROM ItemValuesMap WHERE library_item_id = ? AND item_value_id = ?`,
		libraryItemID, itemValueID,
	)
	return err
}

// UnmapAllItemValues removes all mappings for a library item
func (r *ItemValueRepository) UnmapAllItemValues(libraryItemID string) error {
	_, err := r.db.Exec(
		`DELETE FROM ItemValuesMap WHERE library_item_id = ?`,
		libraryItemID,
	)
	return err
}

// FilterItemsByValue returns library item IDs that have a specific ItemValue
func (r *ItemValueRepository) FilterItemsByValue(itemValueID string) ([]string, error) {
	var itemIDs []string
	err := r.db.Select(&itemIDs, `
		SELECT library_item_id FROM ItemValuesMap WHERE item_value_id = ?
	`, itemValueID)
	if err != nil {
		return nil, err
	}
	return itemIDs, nil
}

// FilterItemsByValueNames returns library item IDs that have any of the specified values
func (r *ItemValueRepository) FilterItemsByValueNames(names []string, valueType string) ([]string, error) {
	if len(names) == 0 {
		return []string{}, nil
	}

	// Build IN clause for names
	placeholders := make([]string, len(names))
	args := make([]interface{}, 0, len(names)+1)
	for i, name := range names {
		placeholders[i] = "?"
		args = append(args, name)
	}
	args = append(args, valueType)

	query := `
		SELECT DISTINCT library_item_id 
		FROM ItemValuesMap 
		INNER JOIN ItemValues ON ItemValuesMap.item_value_id = ItemValues.id
		WHERE ItemValues.name IN (` + strings.Join(placeholders, ",") + `) 
		AND ItemValues.type = ?
	`

	var itemIDs []string
	err := r.db.Select(&itemIDs, query, args...)
	if err != nil {
		return nil, err
	}
	return itemIDs, nil
}

// generateItemValueID creates a deterministic ID from a name
func generateItemValueID(name string) string {
	// Simple deterministic ID - replace with proper UUID generation
	// using crypto/sha256 or similar for production
	return name
}
