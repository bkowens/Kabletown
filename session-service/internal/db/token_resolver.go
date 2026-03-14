package db

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// TokenResolver resolves bearer tokens against the Devices and ApiKeys tables.
type TokenResolver struct {
	db *sqlx.DB
}

// NewTokenResolver creates a new TokenResolver backed by the given database connection.
func NewTokenResolver(database *sqlx.DB) *TokenResolver {
	return &TokenResolver{db: database}
}

// ResolveToken attempts to find the user owning the token.
// It first checks the Devices table (session tokens), then the ApiKeys table (API keys).
// Returns the owning userID, whether the user is an admin, and any error.
func (r *TokenResolver) ResolveToken(token string) (userID string, isAdmin bool, err error) {
	// --- Check Devices table ---
	var devUserID string
	err = r.db.QueryRow(
		`SELECT UserId FROM Devices WHERE AccessToken = ? LIMIT 1`,
		token,
	).Scan(&devUserID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return "", false, fmt.Errorf("token_resolver.ResolveToken devices: %w", err)
	}
	if err == nil {
		var isDisabled bool
		userErr := r.db.QueryRow(
			`SELECT IsDisabled FROM Users WHERE Id = ? LIMIT 1`,
			devUserID,
		).Scan(&isDisabled)
		if userErr != nil {
			if errors.Is(userErr, sql.ErrNoRows) {
				return "", false, errors.New("token_resolver: user not found for device token")
			}
			return "", false, fmt.Errorf("token_resolver.ResolveToken user lookup: %w", userErr)
		}
		if isDisabled {
			return "", false, errors.New("token_resolver: user account is disabled")
		}
		admin, adminErr := r.isAdmin(devUserID)
		if adminErr != nil {
			return "", false, adminErr
		}
		return devUserID, admin, nil
	}

	// --- Check ApiKeys table ---
	var keyID string
	var keyIsAdmin bool
	err = r.db.QueryRow(
		`SELECT Id, IsAdmin FROM ApiKeys WHERE AccessToken = ? LIMIT 1`,
		token,
	).Scan(&keyID, &keyIsAdmin)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", false, errors.New("token_resolver: token not found")
		}
		return "", false, fmt.Errorf("token_resolver.ResolveToken apikeys: %w", err)
	}
	return keyID, keyIsAdmin, nil
}

// isAdmin checks the Permissions table to determine if a user has administrator rights.
func (r *TokenResolver) isAdmin(userID string) (bool, error) {
	var value int
	err := r.db.QueryRow(
		`SELECT Value FROM Permissions WHERE UserId = ? AND Kind = 0 LIMIT 1`,
		userID,
	).Scan(&value)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("token_resolver.isAdmin: %w", err)
	}
	return value == 1, nil
}
