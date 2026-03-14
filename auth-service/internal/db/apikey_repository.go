package db

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// ApiKeyRepository handles CRUD operations on the ApiKeys table.
type ApiKeyRepository struct {
	db *sqlx.DB
}

// NewApiKeyRepository creates a new ApiKeyRepository.
func NewApiKeyRepository(database *sqlx.DB) *ApiKeyRepository {
	return &ApiKeyRepository{db: database}
}

// ListApiKeys returns all rows from the ApiKeys table.
func (r *ApiKeyRepository) ListApiKeys() ([]ApiKey, error) {
	var keys []ApiKey
	err := r.db.Select(&keys,
		`SELECT Id, AccessToken, Name, DateCreated, IsAdmin
		 FROM   ApiKeys
		 ORDER BY DateCreated DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("apikey_repository.ListApiKeys: %w", err)
	}
	return keys, nil
}

// CreateApiKey generates a new API key with the given display name and persists it.
// Returns the generated token string.
func (r *ApiKeyRepository) CreateApiKey(name string) (string, error) {
	tokenBytes := make([]byte, 20)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", fmt.Errorf("apikey_repository.CreateApiKey rand: %w", err)
	}
	token := hex.EncodeToString(tokenBytes) // 40-char hex

	id := uuid.New().String()
	now := time.Now().UTC().Format("2006-01-02 15:04:05")

	_, err := r.db.Exec(
		`INSERT INTO ApiKeys (Id, AccessToken, Name, DateCreated, IsAdmin)
		 VALUES (?, ?, ?, ?, 1)`,
		id, token, name, now,
	)
	if err != nil {
		return "", fmt.Errorf("apikey_repository.CreateApiKey insert: %w", err)
	}
	return token, nil
}

// DeleteApiKey removes an API key identified by either its UUID Id or its AccessToken.
func (r *ApiKeyRepository) DeleteApiKey(keyID string) error {
	_, err := r.db.Exec(
		`DELETE FROM ApiKeys WHERE Id = ? OR AccessToken = ?`,
		keyID, keyID,
	)
	if err != nil {
		return fmt.Errorf("apikey_repository.DeleteApiKey: %w", err)
	}
	return nil
}
