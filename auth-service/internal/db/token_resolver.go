package db

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// Device represents a row in the Devices table.
type Device struct {
	Id               string `db:"Id"`
	UserId           string `db:"UserId"`
	DeviceId         string `db:"DeviceId"`
	FriendlyName     string `db:"FriendlyName"`
	AppName          string `db:"AppName"`
	AppVersion       string `db:"AppVersion"`
	AccessToken      string `db:"AccessToken"`
	Created          string `db:"Created"`
	DateLastActivity string `db:"DateLastActivity"`
}

// ApiKey represents a row in the ApiKeys table.
type ApiKey struct {
	Id          string `db:"Id"`
	AccessToken string `db:"AccessToken"`
	Name        string `db:"Name"`
	DateCreated string `db:"DateCreated"`
	IsAdmin     bool   `db:"IsAdmin"`
}

// userRow is the internal representation of a Users row used during token resolution.
type userRow struct {
	Id         string `db:"Id"`
	Name       string `db:"Name"`
	IsDisabled bool   `db:"IsDisabled"`
}

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
	var dev Device
	err = r.db.Get(&dev,
		`SELECT Id, UserId, DeviceId, FriendlyName, AppName, AppVersion,
		        AccessToken, Created, DateLastActivity
		 FROM   Devices
		 WHERE  AccessToken = ?`,
		token,
	)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return "", false, fmt.Errorf("token_resolver.ResolveToken devices: %w", err)
	}
	if err == nil {
		// Found a device row — look up the user.
		var u userRow
		err = r.db.Get(&u,
			`SELECT Id, Name, IsDisabled FROM Users WHERE Id = ?`,
			dev.UserId,
		)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return "", false, errors.New("token_resolver: user not found for device token")
			}
			return "", false, fmt.Errorf("token_resolver.ResolveToken user lookup: %w", err)
		}
		if u.IsDisabled {
			return "", false, errors.New("token_resolver: user account is disabled")
		}
		admin, adminErr := r.isAdmin(u.Id)
		if adminErr != nil {
			return "", false, adminErr
		}
		return u.Id, admin, nil
	}

	// --- Check ApiKeys table ---
	var key ApiKey
	err = r.db.Get(&key,
		`SELECT Id, AccessToken, Name, DateCreated, IsAdmin
		 FROM   ApiKeys
		 WHERE  AccessToken = ?`,
		token,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", false, errors.New("token_resolver: token not found")
		}
		return "", false, fmt.Errorf("token_resolver.ResolveToken apikeys: %w", err)
	}
	// API keys carry their own isAdmin flag and have no user association.
	return key.Id, key.IsAdmin, nil
}

// isAdmin checks the Permissions table to determine if a user has administrator rights.
// Kind=0 corresponds to IsAdministrator in the Jellyfin permission model.
func (r *TokenResolver) isAdmin(userID string) (bool, error) {
	var value int
	err := r.db.QueryRow(
		`SELECT Value FROM Permissions WHERE UserId = ? AND Kind = 0`,
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
