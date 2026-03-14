package db

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

// User represents a row in the Users table, augmented with the IsAdmin flag
// derived from the Permissions table.
type User struct {
	Id              string `db:"Id"              json:"Id"`
	Name            string `db:"Name"            json:"Name"`
	Password        string `db:"Password"        json:"-"`
	IsAdmin         bool   `json:"IsAdmin"`
	IsDisabled      bool   `db:"IsDisabled"      json:"IsDisabled"`
	IsHidden        bool   `db:"IsHidden"        json:"IsHidden"`
	PrimaryImageTag string `db:"PrimaryImageTag" json:"PrimaryImageTag,omitempty"`
}

// UserRepository handles CRUD operations on the Users table.
type UserRepository struct {
	db       *sqlx.DB
	resolver *TokenResolver
}

// NewUserRepository creates a new UserRepository.
func NewUserRepository(database *sqlx.DB) *UserRepository {
	return &UserRepository{
		db:       database,
		resolver: NewTokenResolver(database),
	}
}

// GetUserByName fetches a user by their display name.
// IsAdmin is populated from the Permissions table.
func (r *UserRepository) GetUserByName(name string) (*User, error) {
	var u User
	err := r.db.Get(&u,
		`SELECT Id, Name, Password, IsDisabled, IsHidden, PrimaryImageTag
		 FROM   Users
		 WHERE  Name = ?`,
		name,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user_repository.GetUserByName: user %q not found", name)
		}
		return nil, fmt.Errorf("user_repository.GetUserByName: %w", err)
	}
	admin, err := r.resolver.isAdmin(u.Id)
	if err != nil {
		return nil, err
	}
	u.IsAdmin = admin
	return &u, nil
}

// GetUserByID fetches a user by their UUID primary key.
// IsAdmin is populated from the Permissions table.
func (r *UserRepository) GetUserByID(id string) (*User, error) {
	var u User
	err := r.db.Get(&u,
		`SELECT Id, Name, Password, IsDisabled, IsHidden, PrimaryImageTag
		 FROM   Users
		 WHERE  Id = ?`,
		id,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user_repository.GetUserByID: user %q not found", id)
		}
		return nil, fmt.Errorf("user_repository.GetUserByID: %w", err)
	}
	admin, err := r.resolver.isAdmin(u.Id)
	if err != nil {
		return nil, err
	}
	u.IsAdmin = admin
	return &u, nil
}

// ListPublicUsers returns all non-hidden, non-disabled users with only public fields.
func (r *UserRepository) ListPublicUsers() ([]User, error) {
	var users []User
	err := r.db.Select(&users,
		`SELECT Id, Name, PrimaryImageTag
		 FROM   Users
		 WHERE  IsHidden = 0 AND IsDisabled = 0
		 ORDER BY Name`,
	)
	if err != nil {
		return nil, fmt.Errorf("user_repository.ListPublicUsers: %w", err)
	}
	return users, nil
}

// ListUsers returns all users.
func (r *UserRepository) ListUsers() ([]User, error) {
	var users []User
	err := r.db.Select(&users,
		`SELECT Id, Name, Password, IsDisabled, IsHidden, PrimaryImageTag
		 FROM   Users
		 ORDER BY Name`,
	)
	if err != nil {
		return nil, fmt.Errorf("user_repository.ListUsers: %w", err)
	}
	// Populate IsAdmin for each user.
	for i := range users {
		admin, err := r.resolver.isAdmin(users[i].Id)
		if err != nil {
			return nil, err
		}
		users[i].IsAdmin = admin
	}
	return users, nil
}

// HashPassword hashes a plain-text password using bcrypt cost 11 (matching C# server).
func HashPassword(plain string) (string, error) {
	// Cost MUST be 11 to match the Jellyfin C# server.
	hash, err := bcrypt.GenerateFromPassword([]byte(plain), 11)
	if err != nil {
		return "", fmt.Errorf("user_repository.HashPassword: %w", err)
	}
	return string(hash), nil
}

// CheckPassword verifies a plain-text password against a bcrypt hash.
func CheckPassword(hash, plain string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain))
}
